package costing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

func (m module) datastreamTabHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	data := CostingDashboardData{
		StartDate: firstOfMonth.Format("2006-01-02"),
		EndDate:   lastOfMonth.Format("2006-01-02"),
	}
	return datastreamTab(data).Render(ctx, w)
}

func (m module) datastreamMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	projectIDFilter := r.URL.Query().Get("project")
	grouping := r.URL.Query().Get("grouping")

	if grouping == "" {
		grouping = "daily"
	}

	tableData, totals, chartData, err := m.fetchGCPBilling(ctx, startDateStr, endDateStr, projectIDFilter, grouping)
	if err != nil {
		m.l.Printf("ERROR fetching billing data: %v", err)
	}

	chartJSON, _ := json.Marshal(chartData)
	return datastreamMetrics(tableData, totals, string(chartJSON), startDateStr, endDateStr).Render(ctx, w)
}

func (m module) datastreamProjectDetailsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	projectID := r.URL.Query().Get("project")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	details, err := m.fetchProjectDetails(ctx, startDateStr, endDateStr, projectID)
	if err != nil {
		m.l.Printf("ERROR fetching project details: %v", err)
	}

	return projectDetailsRow(details).Render(ctx, w)
}

func executeTableQuery(ctx context.Context, client *bigquery.Client, table, start, end, projectFilterSQL, projectFilter string) (map[string]ProjectCostRaw, ProjectCostRaw, error) {
	tableQueryStr := fmt.Sprintf(`
		SELECT
			IFNULL(project.id, 'Unallocated') as project_id,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as total_cost,
			SUM(CASE WHEN service.description = 'Datastream' THEN cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0) ELSE 0 END) as ds_cost,
			SUM(CASE 
				WHEN service.description = 'Datastream' AND usage.unit IN ('bytes', 'byte') THEN usage.amount / 1073741824 
				WHEN service.description = 'Datastream' AND usage.unit LIKE 'gi%%' THEN usage.amount 
				ELSE 0 END) as ds_gb,
			
			SUM(CASE WHEN service.description = 'Cloud Storage' THEN cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0) ELSE 0 END) as gcs_cost,
			SUM(CASE 
				WHEN service.description = 'Cloud Storage' AND usage.unit = 'byte-seconds' THEN usage.amount / (1073741824 * 2592000)
				WHEN service.description = 'Cloud Storage' AND usage.unit IN ('bytes', 'byte') THEN usage.amount / 1073741824
				ELSE 0 END) as gcs_gb,
			
			SUM(CASE WHEN service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions') THEN cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0) ELSE 0 END) as func_cost,
			SUM(CASE 
				WHEN service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions') AND usage.unit = 'gibibyte-seconds' THEN usage.amount / 3600
				WHEN service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions') AND usage.unit IN ('bytes', 'byte') THEN usage.amount / 1073741824
				ELSE 0 END) as func_gb

		FROM `+"`%s`"+`
		WHERE DATE(usage_start_time) >= DATE(@start) 
		  AND DATE(usage_start_time) <= DATE(@end)
		  AND service.description IN ('Datastream', 'Cloud Storage', 'Cloud Functions', 'Cloud Run', 'Cloud Run Functions')
		  %s
		GROUP BY project_id 
	`, table, projectFilterSQL)

	q := client.Query(tableQueryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
	}
	if projectFilter != "" {
		q.Parameters = append(q.Parameters, bigquery.QueryParameter{Name: "project_id", Value: projectFilter})
	}

	results := make(map[string]ProjectCostRaw)
	var totals ProjectCostRaw

	it, err := q.Read(ctx)
	if err != nil {
		return nil, totals, err
	}

	for {
		var row struct {
			ProjectID string               `bigquery:"project_id"`
			TotalCost bigquery.NullFloat64 `bigquery:"total_cost"`
			DSCost    bigquery.NullFloat64 `bigquery:"ds_cost"`
			DSGB      bigquery.NullFloat64 `bigquery:"ds_gb"`
			GCSCost   bigquery.NullFloat64 `bigquery:"gcs_cost"`
			GCSGB     bigquery.NullFloat64 `bigquery:"gcs_gb"`
			FuncCost  bigquery.NullFloat64 `bigquery:"func_cost"`
			FuncGB    bigquery.NullFloat64 `bigquery:"func_gb"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, totals, err
		}

		results[row.ProjectID] = ProjectCostRaw{
			Total:  row.TotalCost.Float64,
			DS:     row.DSCost.Float64,
			DSGB:   row.DSGB.Float64,
			GCS:    row.GCSCost.Float64,
			GCSGB:  row.GCSGB.Float64,
			Func:   row.FuncCost.Float64,
			FuncGB: row.FuncGB.Float64,
		}

		totals.Total += row.TotalCost.Float64
		totals.DS += row.DSCost.Float64
		totals.DSGB += row.DSGB.Float64
		totals.GCS += row.GCSCost.Float64
		totals.GCSGB += row.GCSGB.Float64
		totals.Func += row.FuncCost.Float64
		totals.FuncGB += row.FuncGB.Float64
	}
	return results, totals, nil
}

func (m module) fetchGCPBilling(ctx context.Context, startStr, endStr, projectFilter, grouping string) ([]ProjectCost, ProjectCost, []ChartPoint, error) {
	cacheKey := fmt.Sprintf("core_%s_%s_%s_%s", startStr, endStr, projectFilter, grouping)
	if entry, found := m.cache.Get(cacheKey); found {
		return entry.TableData, entry.Totals, entry.ChartData, nil
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	client, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, ProjectCost{}, nil, fmt.Errorf("bq client err: %v", err)
	}
	defer client.Close()

	billingTable := "df-ps-staging.GOOGLE_COSTING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D"

	projectFilterSQL := "AND project.id IN ('df-ps-south-africa', 'df-ps-zambia', 'df-ps-kenya', 'df-ps-uganda', 'df-ps-tanzania')"
	if projectFilter != "" {
		projectFilterSQL = "AND project.id = @project_id"
	}

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	prevStart := start.AddDate(0, -1, 0).Format("2006-01-02")
	prevEnd := end.AddDate(0, -1, 0).Format("2006-01-02")

	var currentMap, prevMap map[string]ProjectCostRaw
	var currentTotals, prevTotals ProjectCostRaw
	var errCurr, errPrev error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		currentMap, currentTotals, errCurr = executeTableQuery(ctx, client, billingTable, startStr, endStr, projectFilterSQL, projectFilter)
	}()
	go func() {
		defer wg.Done()
		prevMap, prevTotals, errPrev = executeTableQuery(ctx, client, billingTable, prevStart, prevEnd, projectFilterSQL, projectFilter)
	}()
	wg.Wait()

	if errCurr != nil {
		return nil, ProjectCost{}, nil, errCurr
	}
	if errPrev != nil {
		return nil, ProjectCost{}, nil, errPrev
	}

	var tableRows []ProjectCost
	for projID, curr := range currentMap {
		prev := prevMap[projID]
		tableRows = append(tableRows, ProjectCost{
			Project:         projID,
			Total:           fmt.Sprintf("$%.2f", curr.Total),
			TotalDelta:      calcDelta(curr.Total, prev.Total),
			Datastream:      fmt.Sprintf("$%.2f", curr.DS),
			DatastreamDelta: calcDelta(curr.DS, prev.DS),
			DS_Size:         formatSizeGB(curr.DSGB),
			GCS:             fmt.Sprintf("$%.2f", curr.GCS),
			GCSDelta:        calcDelta(curr.GCS, prev.GCS),
			GCS_Size:        formatSizeGB(curr.GCSGB),
			Functions:       fmt.Sprintf("$%.2f", curr.Func),
			FunctionsDelta:  calcDelta(curr.Func, prev.Func),
			Func_Size:       formatSizeGB(curr.FuncGB),
		})
	}

	totals := ProjectCost{
		Total:           fmt.Sprintf("$%.2f", currentTotals.Total),
		TotalDelta:      calcDelta(currentTotals.Total, prevTotals.Total),
		Datastream:      fmt.Sprintf("$%.2f", currentTotals.DS),
		DatastreamDelta: calcDelta(currentTotals.DS, prevTotals.DS),
		DS_Size:         formatSizeGB(currentTotals.DSGB),
		GCS:             fmt.Sprintf("$%.2f", currentTotals.GCS),
		GCSDelta:        calcDelta(currentTotals.GCS, prevTotals.GCS),
		GCS_Size:        formatSizeGB(currentTotals.GCSGB),
		Functions:       fmt.Sprintf("$%.2f", currentTotals.Func),
		FunctionsDelta:  calcDelta(currentTotals.Func, prevTotals.Func),
		Func_Size:       formatSizeGB(currentTotals.FuncGB),
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
			IFNULL(project.id, 'Unallocated') as project_id,
			CASE 
				WHEN service.description = 'Datastream' THEN 'Datastream'
				WHEN service.description = 'Cloud Storage' THEN 'Cloud Storage'
				WHEN service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions') THEN 'Functions'
				ELSE 'Other' END as service_name,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as daily_cost
		FROM `+"`%s`"+`
		WHERE DATE(usage_start_time) >= DATE(@start) 
		  AND DATE(usage_start_time) <= DATE(@end)
		  AND service.description IN ('Datastream', 'Cloud Storage', 'Cloud Functions', 'Cloud Run', 'Cloud Run Functions')
		  %s
		GROUP BY usage_date, project_id, service_name
		ORDER BY usage_date ASC
	`, dateSelect, billingTable, projectFilterSQL)

	cq := client.Query(chartQueryStr)
	cq.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: startStr},
		{Name: "end", Value: endStr},
	}
	if projectFilter != "" {
		cq.Parameters = append(cq.Parameters, bigquery.QueryParameter{Name: "project_id", Value: projectFilter})
	}

	var chartRows []ChartPoint
	itChart, err := cq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				UsageDate   string               `bigquery:"usage_date"`
				ProjectID   string               `bigquery:"project_id"`
				ServiceName string               `bigquery:"service_name"`
				DailyCost   bigquery.NullFloat64 `bigquery:"daily_cost"`
			}
			if err := itChart.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return nil, ProjectCost{}, nil, err
			}
			chartRows = append(chartRows, ChartPoint{
				Date:    row.UsageDate,
				Project: row.ProjectID,
				Service: row.ServiceName,
				Cost:    row.DailyCost.Float64,
			})
		}
	}

	m.cache.Set(cacheKey, tableRows, totals, chartRows)
	return tableRows, totals, chartRows, nil
}

func (m module) fetchProjectDetails(ctx context.Context, startStr, endStr, projectID string) (ProjectDetailsData, error) {
	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	client, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return ProjectDetailsData{}, fmt.Errorf("bq client err: %v", err)
	}
	defer client.Close()

	billingTable := "df-ps-staging.GOOGLE_COSTING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D"

	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)

	curr := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	var requestedMonths []string
	for !curr.After(endMonth) {
		requestedMonths = append(requestedMonths, curr.Format("2006-01"))
		curr = curr.AddDate(0, 1, 0)
	}

	prevStartStr := start.AddDate(0, -1, 0).Format("2006-01-02")

	queryStr := fmt.Sprintf(`
		SELECT
			service.description as service_name,
			COALESCE(resource.name, sku.description, 'Unknown Resource') as resource_name,
			FORMAT_DATE('%%Y-%%m', usage_start_time) as usage_month,
			SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as total_cost
		FROM `+"`%s`"+`
		WHERE IFNULL(project.id, 'Unallocated') = @project_id
		  AND DATE(usage_start_time) >= DATE(@prevStart) 
		  AND DATE(usage_start_time) <= DATE(@end)
		  AND service.description IN ('Datastream', 'Cloud Storage', 'Cloud Functions', 'Cloud Run', 'Cloud Run Functions')
		GROUP BY service_name, resource_name, usage_month
	`, billingTable)

	q := client.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "prevStart", Value: prevStartStr},
		{Name: "end", Value: endStr},
		{Name: "project_id", Value: projectID},
	}

	rawCosts := make(map[string]map[string]float64)
	resourceMeta := make(map[string][2]string)

	it, err := q.Read(ctx)
	if err != nil {
		return ProjectDetailsData{}, err
	}

	for {
		var row struct {
			ServiceName  string               `bigquery:"service_name"`
			ResourceName string               `bigquery:"resource_name"`
			UsageMonth   string               `bigquery:"usage_month"`
			TotalCost    bigquery.NullFloat64 `bigquery:"total_cost"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return ProjectDetailsData{}, err
		}

		key := row.ServiceName + "|" + row.ResourceName
		if _, exists := rawCosts[key]; !exists {
			rawCosts[key] = make(map[string]float64)
			resourceMeta[key] = [2]string{row.ServiceName, row.ResourceName}
		}
		rawCosts[key][row.UsageMonth] += row.TotalCost.Float64
	}

	var details []ResourceCostEx
	for key, monthMap := range rawCosts {
		svc := resourceMeta[key][0]
		res := resourceMeta[key][1]

		var monthsData []MonthData
		var totalForRange float64

		for _, m := range requestedMonths {
			cost := monthMap[m]
			totalForRange += cost

			prevMDate, _ := time.Parse("2006-01", m)
			prevMStr := prevMDate.AddDate(0, -1, 0).Format("2006-01")
			prevCost := monthMap[prevMStr]

			delta := calcDelta(cost, prevCost)

			monthsData = append(monthsData, MonthData{
				MonthName: formatMonthHeader(m),
				Cost:      fmt.Sprintf("$%.2f", cost),
				Delta:     delta,
			})
		}

		if totalForRange > 0.005 {
			cleanSKU := res
			parts := strings.Split(cleanSKU, "/")
			if len(parts) > 0 {
				cleanSKU = parts[len(parts)-1]
			}

			details = append(details, ResourceCostEx{
				Service:     svc,
				SKU:         cleanSKU,
				Description: getSKUDescription(res),
				TotalCost:   fmt.Sprintf("$%.2f", totalForRange),
				TotalNum:    totalForRange,
				Months:      monthsData,
			})
		}
	}

	sort.Slice(details, func(i, j int) bool {
		return details[i].TotalNum > details[j].TotalNum
	})

	var formattedHeaders []string
	for _, m := range requestedMonths {
		formattedHeaders = append(formattedHeaders, formatMonthHeader(m))
	}

	return ProjectDetailsData{
		MonthHeaders: formattedHeaders,
		Details:      details,
	}, nil
}
