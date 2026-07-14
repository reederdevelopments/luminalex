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

// --- Shared Structs for Dataform Tab ---

type DiscoveredTarget struct {
	Project string
	Schema  string
	Runs    int64
}

type DFJobCost struct {
	JobLabel  string
	TotalRuns int64
	TotalCost string
	TotalNum  float64
}

type DFTableData struct {
	Jobs       []DFJobCost
	GrandTotal string
	TotalRuns  int64
}

type DFResourceCostEx struct {
	Database  string
	Schema    string
	Name      string
	TotalRuns int64
	TotalCost string
	TotalNum  float64
	Months    []MonthData
}

type DFDetailsData struct {
	MonthHeaders []string
	Details      []DFResourceCostEx
}

type monthlyStats struct {
	cost float64
}

// --- Route Handlers ---

func (m module) dataformTabHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	nowStr := time.Now().Format("2006-01-02")

	labels, _ := m.fetchDataformLabels(ctx)

	data := CostingDashboardData{
		StartDate: nowStr,
		EndDate:   nowStr,
		Labels:    labels,
	}
	return dataformTab(data).Render(ctx, w)
}

func (m module) dataformMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	grouping := r.URL.Query().Get("grouping")
	queryLabelFilter := r.URL.Query().Get("query_label")

	if grouping == "" {
		grouping = "daily"
	}

	tableData, chartData, err := m.fetchDataformBilling(ctx, startDateStr, endDateStr, grouping, queryLabelFilter)
	if err != nil {
		m.l.Printf("ERROR fetching Dataform billing data: %v", err)
	}

	chartJSON, _ := json.Marshal(chartData)
	return dataformMetrics(tableData, string(chartJSON), startDateStr, endDateStr).Render(ctx, w)
}

func (m module) dataformProjectDetailsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	jobLabel := r.URL.Query().Get("job_label")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	details, err := m.fetchDataformDetails(ctx, startDateStr, endDateStr, jobLabel)
	if err != nil {
		m.l.Printf("ERROR fetching Dataform details: %v", err)
	}

	return dataformDetailsRow(details).Render(ctx, w)
}

// --- Data Pipelines ---

func (m module) fetchDataformLabels(ctx context.Context) ([]string, error) {
	cacheKey := "df_dropdown_labels"
	if entry, found := m.cache.Get(cacheKey); found && len(entry.Labels) > 0 {
		return entry.Labels, nil
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	client, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	billingTable := "df-ps-staging.GOOGLE_COSTING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D"

	queryStr := fmt.Sprintf(`
		SELECT DISTINCT
			REGEXP_REPLACE(l.value, '^dataform:', '') as clean_label
		FROM 
			`+"`%s`"+`,
			UNNEST(labels) l
		WHERE project.id = 'df-fs-insights'
		  AND DATE(usage_start_time) >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)
		  AND service.description LIKE '%%BigQuery%%'
		  AND l.key IN ('query_label', 'dataform')
	`, billingTable)

	q := client.Query(queryStr)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var labels []string
	for {
		var row struct {
			CleanLabel string `bigquery:"clean_label"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		lbl := strings.TrimSpace(row.CleanLabel)
		if lbl != "" && lbl != "Unknown Job" && lbl != "Unknown" && !strings.Contains(lbl, "_") {
			labels = append(labels, lbl)
		}
	}

	sort.Strings(labels)
	m.cache.SetLabels(cacheKey, labels)
	return labels, nil
}

func (m module) fetchDataformBilling(ctx context.Context, startStr, endStr, grouping, queryLabelFilter string) (DFTableData, []ChartPoint, error) {
	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	client, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return DFTableData{}, nil, fmt.Errorf("bq client err: %v", err)
	}
	defer client.Close()

	billingTable := "df-ps-staging.GOOGLE_COSTING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D"

	filterSQL := ""
	if queryLabelFilter != "" {
		filterSQL = "AND (l.value = @query_label OR l.value = CONCAT('dataform:', @query_label))"
	}

	tableQueryStr := fmt.Sprintf(`
		SELECT
			REGEXP_REPLACE(l.value, '^dataform:', '') as job_label,
			COUNT(DISTINCT COALESCE(resource.name, (SELECT value FROM UNNEST(system_labels) WHERE key = 'compute.googleapis.com/job_id'), CAST(usage_start_time AS STRING))) as total_runs,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as total_cost
		FROM 
			`+"`%s`"+`,
			UNNEST(labels) l
		WHERE project.id = 'df-fs-insights'
		  AND DATE(usage_start_time) >= DATE(@start) 
		  AND DATE(usage_start_time) <= DATE(@end)
		  AND service.description LIKE '%%BigQuery%%'
		  AND l.key IN ('query_label', 'dataform')
		  %s
		GROUP BY job_label
		HAVING total_cost != 0
		ORDER BY total_cost DESC
	`, billingTable, filterSQL)

	tq := client.Query(tableQueryStr)
	tq.Parameters = []bigquery.QueryParameter{{Name: "start", Value: startStr}, {Name: "end", Value: endStr}}
	if queryLabelFilter != "" {
		tq.Parameters = append(tq.Parameters, bigquery.QueryParameter{Name: "query_label", Value: queryLabelFilter})
	}

	var jobs []DFJobCost
	var grandTotalNum float64
	var totalRuns int64

	it, err := tq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				JobLabel  string               `bigquery:"job_label"`
				TotalRuns int64                `bigquery:"total_runs"`
				TotalCost bigquery.NullFloat64 `bigquery:"total_cost"`
			}
			if err := it.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return DFTableData{}, nil, err
			}

			jobs = append(jobs, DFJobCost{
				JobLabel:  row.JobLabel,
				TotalRuns: row.TotalRuns,
				TotalCost: fmt.Sprintf("$%.2f", row.TotalCost.Float64),
				TotalNum:  row.TotalCost.Float64,
			})
			grandTotalNum += row.TotalCost.Float64
			totalRuns += row.TotalRuns
		}
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
			REGEXP_REPLACE(l.value, '^dataform:', '') as job_label,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as daily_cost
		FROM 
			`+"`%s`"+`,
			UNNEST(labels) l
		WHERE project.id = 'df-fs-insights'
		  AND DATE(usage_start_time) >= DATE(@start) 
		  AND DATE(usage_start_time) <= DATE(@end)
		  AND service.description LIKE '%%BigQuery%%'
		  AND l.key IN ('query_label', 'dataform')
		  %s
		GROUP BY usage_date, job_label
		HAVING daily_cost != 0
		ORDER BY usage_date ASC
	`, dateSelect, billingTable, filterSQL)

	cq := client.Query(chartQueryStr)
	cq.Parameters = []bigquery.QueryParameter{{Name: "start", Value: startStr}, {Name: "end", Value: endStr}}
	if queryLabelFilter != "" {
		cq.Parameters = append(cq.Parameters, bigquery.QueryParameter{Name: "query_label", Value: queryLabelFilter})
	}

	var chartRows []ChartPoint
	itChart, err := cq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				UsageDate string               `bigquery:"usage_date"`
				JobLabel  string               `bigquery:"job_label"`
				DailyCost bigquery.NullFloat64 `bigquery:"daily_cost"`
			}
			if err := itChart.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return DFTableData{}, nil, err
			}
			chartRows = append(chartRows, ChartPoint{
				Date:    row.UsageDate,
				Service: row.JobLabel,
				Cost:    row.DailyCost.Float64,
			})
		}
	}

	return DFTableData{Jobs: jobs, GrandTotal: fmt.Sprintf("$%.2f", grandTotalNum), TotalRuns: totalRuns}, chartRows, nil
}

func (lm *LineageManager) DiscoverTablesByLabel(ctx context.Context, start, end, jobLabel string) (map[string]DiscoveredTarget, error) {
	fullJobLabel := "dataform:" + jobLabel

	queryStr := fmt.Sprintf(`
		SELECT 
		  project_id as project_name,
		  destination_table.dataset_id as schema_name,
		  destination_table.table_id as table_name,
		  COUNT(*) as runs
		FROM 
		  ` + "`df-fs-insights.region-europe-west9.INFORMATION_SCHEMA.JOBS`" + `
		WHERE 
		  DATE(creation_time) >= DATE(@start)
		  AND DATE(creation_time) <= DATE(@end)
		  AND EXISTS (
			SELECT 1 
			FROM UNNEST(labels) 
			WHERE value = @job_label OR value = REGEXP_REPLACE(@job_label, '^dataform:', '')
		  )
		  AND destination_table.table_id IS NOT NULL
		  AND LEFT(destination_table.dataset_id, 1) != '_'
		GROUP BY ALL
	`)

	q := lm.bqClient.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
		{Name: "job_label", Value: fullJobLabel},
	}

	tablesMap := make(map[string]DiscoveredTarget)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	for {
		var row struct {
			Project string `bigquery:"project_name"`
			Schema  string `bigquery:"schema_name"`
			Table   string `bigquery:"table_name"`
			Runs    int64  `bigquery:"runs"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		tablesMap[row.Table] = DiscoveredTarget{
			Project: row.Project,
			Schema:  row.Schema,
			Runs:    row.Runs,
		}
	}
	return tablesMap, nil
}

func (m module) fetchDataformDetails(ctx context.Context, startStr, endStr, jobLabel string) (DFDetailsData, error) {
	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	curr := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	var requestedMonths []string
	var formattedHeaders []string
	for !curr.After(endMonth) {
		mStr := curr.Format("2006-01")
		requestedMonths = append(requestedMonths, mStr)
		formattedHeaders = append(formattedHeaders, formatMonthHeader(mStr))
		curr = curr.AddDate(0, 1, 0)
	}

	// 1. Fetch unified data directly from your verified read-through Firestore/BQ manager
	var auditRecords []LineageRecord
	lineageMgr, err := NewLineageManager(ctx, gcpProjectID)
	if err == nil {
		defer lineageMgr.Close()
		records, err := lineageMgr.FetchLineage(ctx, jobLabel, startStr, endStr)
		if err == nil {
			auditRecords = records
		} else {
			m.l.Printf("FIRESTORE/BQ FETCH ERROR: %v", err)
		}
	}

	// 2. Pivot the unified operational data into our dynamic month grid map
	type tableStats struct {
		runs map[string]int64
		cost map[string]float64
	}

	// Key format: project_name|schema_name|table_name
	pivotMap := make(map[string]*tableStats)

	for _, rec := range auditRecords {
		key := fmt.Sprintf("%s|%s|%s", rec.ProjectName, rec.SchemaName, rec.TableName)

		t, err := time.Parse("2006-01-02", rec.UsageDate)
		if err != nil {
			continue
		}
		mStr := t.Format("2006-01")

		if _, exists := pivotMap[key]; !exists {
			pivotMap[key] = &tableStats{
				runs: make(map[string]int64),
				cost: make(map[string]float64),
			}
		}

		pivotMap[key].runs[mStr] += rec.Runs
		pivotMap[key].cost[mStr] += rec.Cost
	}

	// 3. Construct your final UI model
	var details []DFResourceCostEx
	for key, stats := range pivotMap {
		parts := strings.Split(key, "|")
		proj := parts[0]
		schema := parts[1]
		tbl := parts[2]

		var monthsData []MonthData
		var totalCostForRange float64
		var totalRunsForRange int64

		for _, m := range requestedMonths {
			cVal := stats.cost[m]
			totalCostForRange += cVal
			totalRunsForRange += stats.runs[m]

			// Compute month-over-month comparisons dynamically
			prevMDate, _ := time.Parse("2006-01", m)
			prevMStr := prevMDate.AddDate(0, -1, 0).Format("2006-01")
			prevCVal := stats.cost[prevMStr]

			monthsData = append(monthsData, MonthData{
				MonthName: formatMonthHeader(m),
				Cost:      fmt.Sprintf("$%.2f", cVal),
				Delta:     calcDelta(cVal, prevCVal),
			})
		}

		details = append(details, DFResourceCostEx{
			Database:  proj,
			Schema:    schema,
			Name:      tbl,
			TotalRuns: totalRunsForRange,
			TotalCost: fmt.Sprintf("$%.2f", totalCostForRange),
			TotalNum:  totalCostForRange,
			Months:    monthsData,
		})
	}

	// Order entries so your highest impact tables sit at the top
	sort.Slice(details, func(i, j int) bool {
		return details[i].TotalNum > details[j].TotalNum
	})

	return DFDetailsData{MonthHeaders: formattedHeaders, Details: details}, nil
}
