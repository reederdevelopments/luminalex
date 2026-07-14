package costing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type OverallKPIs struct {
	TotalCost    string
	TotalDelta   string
	DataCost     string
	DataDelta    string
	ComputeCost  string
	ComputeDelta string
	StorageCost  string
	StorageDelta string
}

type ProjectSummary struct {
	ProjectID   string
	TotalCost   string
	DataCost    string
	ComputeCost string
	StorageCost string
	OtherCost   string
	TotalNum    float64
	DataNum     float64
	ComputeNum  float64
	StorageNum  float64
	OtherNum    float64
}

type OverallTableData struct {
	KPIs     OverallKPIs
	Projects []ProjectSummary
}

type OverallChartPoint struct {
	Date     string  `json:"date"`
	Category string  `json:"category"`
	Cost     float64 `json:"cost"`
}

type OverallCachePayload struct {
	TableData OverallTableData
	ChartData []OverallChartPoint
	ExpiresAt time.Time
}

// Local cache to keep overall tab lightning fast
var ovCache sync.Map

func (m module) overallTabHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	data := CostingDashboardData{
		StartDate: firstOfMonth.Format("2006-01-02"),
		EndDate:   lastOfMonth.Format("2006-01-02"),
	}
	return overallTab(data).Render(ctx, w)
}

func (m module) overallMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	grouping := r.URL.Query().Get("grouping")

	if grouping == "" {
		grouping = "daily"
	}

	tableData, chartData, err := m.fetchOverallBilling(ctx, startDateStr, endDateStr, grouping)
	if err != nil {
		m.l.Printf("ERROR fetching Overall billing data: %v", err)
	}

	chartJSON, _ := json.Marshal(chartData)
	return overallMetrics(tableData, string(chartJSON)).Render(ctx, w)
}

// Executes the aggregation query and returns totals + project mapped breakdown
func executeOverallTableQuery(ctx context.Context, client *bigquery.Client, start, end string) (float64, float64, float64, float64, map[string]*ProjectSummary, error) {
	queryStr := `
		SELECT
			IFNULL(project.id, 'Unallocated') as project_id,
			CASE
				WHEN service.description LIKE '%BigQuery%' THEN 'Data Analytics'
				WHEN service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions', 'Compute Engine', 'App Engine') THEN 'Compute'
				WHEN service.description IN ('Cloud Storage', 'Datastream', 'Cloud SQL', 'Cloud Bigtable') THEN 'Storage & DBs'
				ELSE 'Other'
			END as category,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as cost
		FROM ` + "`df-ps-staging.GOOGLE_COSTING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D`" + `
		WHERE DATE(usage_start_time) >= DATE(@start)
		  AND DATE(usage_start_time) <= DATE(@end)
		GROUP BY project_id, category
	`

	q := client.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
	}

	var grandTotal, dataTotal, computeTotal, storageTotal float64
	projectMap := make(map[string]*ProjectSummary)

	it, err := q.Read(ctx)
	if err != nil {
		return 0, 0, 0, 0, nil, err
	}

	for {
		var row struct {
			ProjectID string               `bigquery:"project_id"`
			Category  string               `bigquery:"category"`
			Cost      bigquery.NullFloat64 `bigquery:"cost"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return 0, 0, 0, 0, nil, err
		}

		cost := row.Cost.Float64
		if cost < 0.01 && cost > -0.01 {
			continue // Skip negligible zero-balance noise
		}

		p, exists := projectMap[row.ProjectID]
		if !exists {
			p = &ProjectSummary{ProjectID: row.ProjectID}
			projectMap[row.ProjectID] = p
		}

		switch row.Category {
		case "Data Analytics":
			p.DataNum += cost
			dataTotal += cost
		case "Compute":
			p.ComputeNum += cost
			computeTotal += cost
		case "Storage & DBs":
			p.StorageNum += cost
			storageTotal += cost
		default:
			p.OtherNum += cost
		}
		p.TotalNum += cost
		grandTotal += cost
	}

	return grandTotal, dataTotal, computeTotal, storageTotal, projectMap, nil
}

func (m module) fetchOverallBilling(ctx context.Context, startStr, endStr, grouping string) (OverallTableData, []OverallChartPoint, error) {
	cacheKey := fmt.Sprintf("ov_%s_%s_%s", startStr, endStr, grouping)
	if cached, ok := ovCache.Load(cacheKey); ok {
		payload := cached.(OverallCachePayload)
		if time.Now().Before(payload.ExpiresAt) {
			return payload.TableData, payload.ChartData, nil
		}
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	client, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return OverallTableData{}, nil, err
	}
	defer client.Close()

	// Calculate strict Delta Window sizes based on user input
	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)
	days := int(end.Sub(start).Hours() / 24)

	prevEnd := start.AddDate(0, 0, -1)
	prevStart := prevEnd.AddDate(0, 0, -days)

	prevStartStr := prevStart.Format("2006-01-02")
	prevEndStr := prevEnd.Format("2006-01-02")

	var currTotal, currData, currCompute, currStorage float64
	var prevTotal, prevData, prevCompute, prevStorage float64
	var currProjectMap map[string]*ProjectSummary
	var errCurr, errPrev error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		currTotal, currData, currCompute, currStorage, currProjectMap, errCurr = executeOverallTableQuery(ctx, client, startStr, endStr)
	}()
	go func() {
		defer wg.Done()
		prevTotal, prevData, prevCompute, prevStorage, _, errPrev = executeOverallTableQuery(ctx, client, prevStartStr, prevEndStr)
	}()
	wg.Wait()

	if errCurr != nil {
		return OverallTableData{}, nil, errCurr
	}
	if errPrev != nil {
		m.l.Printf("Warning: Failed to fetch previous period delta: %v", errPrev)
	}

	kpis := OverallKPIs{
		TotalCost:    fmt.Sprintf("$%.2f", currTotal),
		TotalDelta:   calcDelta(currTotal, prevTotal),
		DataCost:     fmt.Sprintf("$%.2f", currData),
		DataDelta:    calcDelta(currData, prevData),
		ComputeCost:  fmt.Sprintf("$%.2f", currCompute),
		ComputeDelta: calcDelta(currCompute, prevCompute),
		StorageCost:  fmt.Sprintf("$%.2f", currStorage),
		StorageDelta: calcDelta(currStorage, prevStorage),
	}

	var projects []ProjectSummary
	for _, p := range currProjectMap {
		if p.TotalNum > 0.05 {
			p.TotalCost = fmt.Sprintf("$%.2f", p.TotalNum)
			p.DataCost = fmt.Sprintf("$%.2f", p.DataNum)
			p.ComputeCost = fmt.Sprintf("$%.2f", p.ComputeNum)
			p.StorageCost = fmt.Sprintf("$%.2f", p.StorageNum)
			p.OtherCost = fmt.Sprintf("$%.2f", p.OtherNum)
			projects = append(projects, *p)
		}
	}

	// Sort Projects by Highest Impact
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].TotalNum > projects[j].TotalNum
	})

	dateSelect := "CAST(DATE(usage_start_time) AS STRING)"
	if grouping == "monthly" {
		dateSelect = "CAST(FORMAT_DATE('%Y-%m', usage_start_time) AS STRING)"
	} else if grouping == "yearly" {
		dateSelect = "CAST(FORMAT_DATE('%Y', usage_start_time) AS STRING)"
	}

	chartQueryStr := fmt.Sprintf(`
		SELECT
			%s as usage_date,
			CASE
				WHEN service.description LIKE '%%BigQuery%%' THEN 'Data Analytics'
				WHEN service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions', 'Compute Engine', 'App Engine') THEN 'Compute'
				WHEN service.description IN ('Cloud Storage', 'Datastream', 'Cloud SQL', 'Cloud Bigtable') THEN 'Storage & DBs'
				ELSE 'Other'
			END as category,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as daily_cost
		FROM `+"`df-ps-staging.GOOGLE_COSTING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D`"+`
		WHERE DATE(usage_start_time) >= DATE(@start) 
		  AND DATE(usage_start_time) <= DATE(@end)
		GROUP BY usage_date, category
		HAVING daily_cost > 0
		ORDER BY usage_date ASC
	`, dateSelect)

	cq := client.Query(chartQueryStr)
	cq.Parameters = []bigquery.QueryParameter{{Name: "start", Value: startStr}, {Name: "end", Value: endStr}}

	var chartRows []OverallChartPoint
	itChart, err := cq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				UsageDate string               `bigquery:"usage_date"`
				Category  string               `bigquery:"category"`
				DailyCost bigquery.NullFloat64 `bigquery:"daily_cost"`
			}
			if err := itChart.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return OverallTableData{}, nil, err
			}
			chartRows = append(chartRows, OverallChartPoint{
				Date:     row.UsageDate,
				Category: row.Category,
				Cost:     row.DailyCost.Float64,
			})
		}
	}

	tableData := OverallTableData{KPIs: kpis, Projects: projects}

	// Cache for 2 hours
	ovCache.Store(cacheKey, OverallCachePayload{
		TableData: tableData,
		ChartData: chartRows,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	})

	return tableData, chartRows, nil
}
