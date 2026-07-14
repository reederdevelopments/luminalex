package costing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type TPJobCost struct {
	JobName   string
	TotalCost string
	TotalNum  float64
	Months    []MonthData
}

type TPTableData struct {
	MonthHeaders []string
	Jobs         []TPJobCost
	MonthTotals  []string
	GrandTotal   string
}

type TPChartPoint struct {
	Date    string  `json:"date"`
	JobName string  `json:"job_name"`
	Cost    float64 `json:"cost"`
}

func (m module) thirdpartyTabHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	data := CostingDashboardData{
		StartDate: firstOfMonth.Format("2006-01-02"),
		EndDate:   lastOfMonth.Format("2006-01-02"),
	}
	return thirdpartyTab(data).Render(ctx, w)
}

func (m module) thirdpartyMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	grouping := r.URL.Query().Get("grouping")

	if grouping == "" {
		grouping = "daily"
	}

	tableData, chartData, err := m.fetchThirdPartyBilling(ctx, startDateStr, endDateStr, grouping)
	if err != nil {
		m.l.Printf("ERROR fetching 3rd party billing data: %v", err)
	}

	chartJSON, _ := json.Marshal(chartData)
	return thirdpartyMetrics(tableData, string(chartJSON)).Render(ctx, w)
}

func (m module) fetchThirdPartyBilling(ctx context.Context, startStr, endStr, grouping string) (TPTableData, []TPChartPoint, error) {
	cacheKey := fmt.Sprintf("tp_core_%s_%s_%s", startStr, endStr, grouping)

	// FIX: Leveraging the specific TPTable cache entry
	if entry, found := m.cache.Get(cacheKey); found && len(entry.TPTable.Jobs) > 0 {
		return entry.TPTable, entry.TPChart, nil
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	client, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return TPTableData{}, nil, fmt.Errorf("bq client err: %v", err)
	}
	defer client.Close()

	billingTable := "df-ps-staging.GOOGLE_COSTING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D"

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	curr := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	var requestedMonths []string
	var formattedHeaders []string
	for !curr.After(endMonth) {
		requestedMonths = append(requestedMonths, curr.Format("2006-01"))
		formattedHeaders = append(formattedHeaders, formatMonthHeader(curr.Format("2006-01")))
		curr = curr.AddDate(0, 1, 0)
	}

	prevStartStr := start.AddDate(0, -1, 0).Format("2006-01-02")

	tableQueryStr := fmt.Sprintf(`
		SELECT
			COALESCE(resource.name, sku.description, 'Unknown Job') as job_name,
			FORMAT_DATE('%%Y-%%m', usage_start_time) as usage_month,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as total_cost
		FROM `+"`%s`"+`
		WHERE project.id = 'df-ps-staging'
		  AND DATE(usage_start_time) >= DATE(@prevStart) 
		  AND DATE(usage_start_time) <= DATE(@end)
		  AND service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions')
		GROUP BY job_name, usage_month
	`, billingTable)

	tq := client.Query(tableQueryStr)
	tq.Parameters = []bigquery.QueryParameter{{Name: "prevStart", Value: prevStartStr}, {Name: "end", Value: endStr}}

	rawCosts := make(map[string]map[string]float64)
	it, err := tq.Read(ctx)
	if err != nil {
		return TPTableData{}, nil, err
	}

	for {
		var row struct {
			JobName    string               `bigquery:"job_name"`
			UsageMonth string               `bigquery:"usage_month"`
			TotalCost  bigquery.NullFloat64 `bigquery:"total_cost"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return TPTableData{}, nil, err
		}

		cleanJob := row.JobName
		parts := strings.Split(cleanJob, "/")
		if len(parts) > 0 {
			cleanJob = parts[len(parts)-1]
		}

		if _, exists := rawCosts[cleanJob]; !exists {
			rawCosts[cleanJob] = make(map[string]float64)
		}
		rawCosts[cleanJob][row.UsageMonth] += row.TotalCost.Float64
	}

	var jobs []TPJobCost
	var grandTotalNum float64
	monthlyTotalsNum := make(map[string]float64)

	for jobName, monthMap := range rawCosts {
		var monthsData []MonthData
		var totalForRange float64

		for _, m := range requestedMonths {
			cost := monthMap[m]
			totalForRange += cost
			monthlyTotalsNum[m] += cost

			prevMDate, _ := time.Parse("2006-01", m)
			prevMStr := prevMDate.AddDate(0, -1, 0).Format("2006-01")
			prevCost := monthMap[prevMStr]

			monthsData = append(monthsData, MonthData{
				MonthName: formatMonthHeader(m),
				Cost:      fmt.Sprintf("$%.2f", cost),
				Delta:     calcDelta(cost, prevCost),
			})
		}

		if totalForRange > 0 {
			jobs = append(jobs, TPJobCost{
				JobName:   jobName,
				TotalCost: fmt.Sprintf("$%.2f", totalForRange),
				TotalNum:  totalForRange,
				Months:    monthsData,
			})
			grandTotalNum += totalForRange
		}
	}

	sort.Slice(jobs, func(i, j int) bool { return jobs[i].TotalNum > jobs[j].TotalNum })

	var formattedMonthTotals []string
	for _, m := range requestedMonths {
		formattedMonthTotals = append(formattedMonthTotals, fmt.Sprintf("$%.2f", monthlyTotalsNum[m]))
	}

	tableOutput := TPTableData{
		MonthHeaders: formattedHeaders,
		Jobs:         jobs,
		MonthTotals:  formattedMonthTotals,
		GrandTotal:   fmt.Sprintf("$%.2f", grandTotalNum),
	}

	dateSelect := "CAST(DATE(usage_start_time) AS STRING)"
	if grouping == "monthly" {
		dateSelect = "CAST(FORMAT_DATE('%Y-%m', usage_start_time) AS STRING)"
	} else if grouping == "yearly" {
		dateSelect = "CAST(FORMAT_DATE('%Y', usage_start_time) AS STRING)"
	}

	chartQueryStr := fmt.Sprintf(`
		SELECT
			%s as usage_date,
			COALESCE(resource.name, sku.description, 'Unknown Job') as job_name,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as daily_cost
		FROM `+"`%s`"+`
		WHERE project.id = 'df-ps-staging'
		  AND DATE(usage_start_time) >= DATE(@start) 
		  AND DATE(usage_start_time) <= DATE(@end)
		  AND service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions')
		GROUP BY usage_date, job_name
		HAVING daily_cost > 0
		ORDER BY usage_date ASC
	`, dateSelect, billingTable)

	cq := client.Query(chartQueryStr)
	cq.Parameters = []bigquery.QueryParameter{{Name: "start", Value: startStr}, {Name: "end", Value: endStr}}

	var chartRows []TPChartPoint
	itChart, err := cq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				UsageDate string               `bigquery:"usage_date"`
				JobName   string               `bigquery:"job_name"`
				DailyCost bigquery.NullFloat64 `bigquery:"daily_cost"`
			}
			if err := itChart.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return tableOutput, nil, err
			}

			cleanJob := row.JobName
			parts := strings.Split(cleanJob, "/")
			if len(parts) > 0 {
				cleanJob = parts[len(parts)-1]
			}

			chartRows = append(chartRows, TPChartPoint{
				Date:    row.UsageDate,
				JobName: cleanJob,
				Cost:    row.DailyCost.Float64,
			})
		}
	}

	// FIX: Save strictly to the TP cache
	m.cache.SetTP(cacheKey, tableOutput, chartRows)

	return tableOutput, chartRows, nil
}
