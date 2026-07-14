package costing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type DSAssetCost struct {
	AssetID   string
	AssetName string
	SafeID    string
	URLSafeID string
	TotalSize string
	TotalRuns int64
	TotalCost string
	TotalNum  float64
}

type DSTableData struct {
	Assets     []DSAssetCost
	GrandTotal string
	GrandSize  string
	GrandRuns  int64
}

type DSQueryDetail struct {
	ResourceName string
	Credentials  string
	QueryType    string
	TotalRuns    int64
	AvgRunTime   string
	TotalCost    string
	TotalSize    string
	TotalNum     float64
}

type DSDetailsData struct {
	Details []DSQueryDetail
}

func (m module) fetchDSMappings(ctx context.Context) map[string]string {
	mappings := make(map[string]string)
	iter := m.sessionStore.Db().Collection("datastudio_meta").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		if name, ok := doc.Data()["Name"].(string); ok {
			mappings[doc.Ref.ID] = name
		}
	}
	return mappings
}

func (m module) saveDSMappingHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	assetID := strings.TrimSpace(r.FormValue("asset_id"))
	assetName := strings.TrimSpace(r.FormValue("asset_name"))

	if assetID != "" && assetName != "" {
		_, err := m.sessionStore.Db().Collection("datastudio_meta").Doc(assetID).Set(ctx, map[string]interface{}{
			"Name": assetName,
		})
		if err != nil {
			m.l.Printf("Failed to save DS mapping: %v", err)
		}
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func (m module) datastudioTabHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()

	daysSinceMonday := int(now.Weekday()) - 1
	if daysSinceMonday < 0 {
		daysSinceMonday = 6
	}
	startOfWeek := now.AddDate(0, 0, -daysSinceMonday)

	startStr := startOfWeek.Format("2006-01-02")
	endStr := now.Format("2006-01-02")

	assetsList, _ := m.fetchDSAseets(ctx)
	mappings := m.fetchDSMappings(ctx)

	data := CostingDashboardData{
		StartDate: startStr,
		EndDate:   endStr,
		Labels:    assetsList,
		LabelMap:  mappings, // Send the mappings dictionary forward
	}
	return datastudioTab(data).Render(ctx, w)
}

func (m module) datastudioMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	grouping := r.URL.Query().Get("grouping")
	assetFilter := r.URL.Query().Get("asset_id")
	projectFilter := r.URL.Query().Get("project")

	if grouping == "" {
		grouping = "daily"
	}

	tableData, chartData, err := m.fetchDSBilling(ctx, startDateStr, endDateStr, projectFilter, grouping, assetFilter)
	if err != nil {
		m.l.Printf("ERROR fetching DS billing data: %v", err)
	}

	chartJSON, _ := json.Marshal(chartData)
	return datastudioMetrics(tableData, string(chartJSON), startDateStr, endDateStr, projectFilter).Render(ctx, w)
}

func (m module) datastudioDetailsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	assetID := r.URL.Query().Get("asset_id")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	projectFilter := r.URL.Query().Get("project")

	details, err := m.fetchDSDetails(ctx, startDateStr, endDateStr, projectFilter, assetID)
	if err != nil {
		m.l.Printf("ERROR fetching DS details: %v", err)
	}

	return datastudioDetailsRow(details).Render(ctx, w)
}

func (m module) fetchDSAseets(ctx context.Context) ([]string, error) {
	cacheKey := "ds_dropdown_assets"
	if entry, found := m.cache.Get(cacheKey); found && len(entry.Labels) > 0 {
		return entry.Labels, nil
	}

	client, err := bigquery.NewClient(ctx, "df-frontend")
	if err != nil {
		return nil, err
	}
	defer client.Close()

	selects := "labels"
	wheres := "creation_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) AND EXISTS (SELECT 1 FROM UNNEST(labels) WHERE key LIKE '%studio_report_id')"
	multiProjectFrom := getMultiProjectFrom("", selects, wheres)

	queryStr := fmt.Sprintf(`
		SELECT DISTINCT 
			(SELECT value FROM UNNEST(labels) WHERE key IN ('looker_studio_report_id', 'datastudio_report_id') LIMIT 1) as asset_id
		FROM %s
	`, multiProjectFrom)
	q := client.Query(queryStr)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var assets []string
	for {
		var row struct {
			AssetID bigquery.NullString `bigquery:"asset_id"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		if row.AssetID.Valid && row.AssetID.StringVal != "" {
			assets = append(assets, row.AssetID.StringVal)
		}
	}

	sort.Strings(assets)
	m.cache.SetLabels(cacheKey, assets)
	return assets, nil
}

func (m module) fetchDSBilling(ctx context.Context, startStr, endStr, projectFilter, grouping, assetFilter string) (DSTableData, []ChartPoint, error) {
	client, err := bigquery.NewClient(ctx, "df-frontend")
	if err != nil {
		return DSTableData{}, nil, err
	}
	defer client.Close()

	mappings := m.fetchDSMappings(ctx)

	selects := "labels, total_bytes_billed, creation_time"
	wheres := "creation_time >= TIMESTAMP(@start) AND creation_time < TIMESTAMP_ADD(TIMESTAMP(@end), INTERVAL 1 DAY) AND total_bytes_billed > 0 AND EXISTS (SELECT 1 FROM UNNEST(labels) WHERE key LIKE '%studio_report_id')"

	if assetFilter != "" {
		wheres += " AND EXISTS (SELECT 1 FROM UNNEST(labels) WHERE (key = 'looker_studio_report_id' OR key = 'datastudio_report_id') AND value = @asset_id)"
	}

	multiProjectFrom := getMultiProjectFrom(projectFilter, selects, wheres)

	tableQueryStr := fmt.Sprintf(`
		SELECT 
			COALESCE((SELECT value FROM UNNEST(labels) WHERE key IN ('looker_studio_report_id', 'datastudio_report_id') LIMIT 1), 'Unknown Dashboard') as asset_id,
			COUNT(*) as total_runs,
			SUM(total_bytes_billed) as total_bytes,
			(SUM(total_bytes_billed) / POWER(1024, 4)) * 6.25 AS total_cost
		FROM %s
		GROUP BY asset_id
		ORDER BY total_cost DESC
	`, multiProjectFrom)

	tq := client.Query(tableQueryStr)
	tq.Parameters = []bigquery.QueryParameter{{Name: "start", Value: startStr}, {Name: "end", Value: endStr}}
	if assetFilter != "" {
		tq.Parameters = append(tq.Parameters, bigquery.QueryParameter{Name: "asset_id", Value: assetFilter})
	}

	var assets []DSAssetCost
	var grandTotalCost, grandTotalBytes float64
	var grandRuns int64

	it, err := tq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				AssetID    string               `bigquery:"asset_id"`
				TotalRuns  bigquery.NullInt64   `bigquery:"total_runs"`
				TotalBytes bigquery.NullInt64   `bigquery:"total_bytes"`
				TotalCost  bigquery.NullFloat64 `bigquery:"total_cost"`
			}
			if err := it.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return DSTableData{}, nil, err
			}

			assetName := row.AssetID
			if mapped, ok := mappings[row.AssetID]; ok {
				assetName = mapped
			}

			assets = append(assets, DSAssetCost{
				AssetID:   row.AssetID,
				AssetName: assetName,
				SafeID:    makeSafeID(row.AssetID),
				URLSafeID: url.QueryEscape(row.AssetID),
				TotalRuns: row.TotalRuns.Int64,
				TotalSize: formatBytes(float64(row.TotalBytes.Int64)),
				TotalCost: fmt.Sprintf("$%.2f", row.TotalCost.Float64),
				TotalNum:  row.TotalCost.Float64,
			})
			grandTotalCost += row.TotalCost.Float64
			grandTotalBytes += float64(row.TotalBytes.Int64)
			grandRuns += row.TotalRuns.Int64
		}
	}

	dateSelect := "CAST(DATE(creation_time) AS STRING)"
	if grouping == "monthly" {
		dateSelect = "CAST(FORMAT_DATE('%Y-%m', creation_time) AS STRING)"
	}

	chartQueryStr := fmt.Sprintf(`
		SELECT 
			%s as usage_date,
			COALESCE((SELECT value FROM UNNEST(labels) WHERE key IN ('looker_studio_report_id', 'datastudio_report_id') LIMIT 1), 'Unknown Dashboard') as asset_id,
			(SUM(total_bytes_billed) / POWER(1024, 4)) * 6.25 AS daily_cost
		FROM %s
		GROUP BY usage_date, asset_id
		ORDER BY usage_date ASC
	`, dateSelect, multiProjectFrom)

	cq := client.Query(chartQueryStr)
	cq.Parameters = tq.Parameters

	var chartRows []ChartPoint
	itChart, err := cq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				UsageDate string               `bigquery:"usage_date"`
				AssetID   string               `bigquery:"asset_id"`
				DailyCost bigquery.NullFloat64 `bigquery:"daily_cost"`
			}
			if err := itChart.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return DSTableData{}, nil, err
			}

			assetName := row.AssetID
			if mapped, ok := mappings[row.AssetID]; ok {
				assetName = mapped
			}

			chartRows = append(chartRows, ChartPoint{
				Date:    row.UsageDate,
				Service: assetName,
				Cost:    row.DailyCost.Float64,
			})
		}
	}

	return DSTableData{
		Assets:     assets,
		GrandTotal: fmt.Sprintf("$%.2f", grandTotalCost),
		GrandSize:  formatBytes(grandTotalBytes),
		GrandRuns:  grandRuns,
	}, chartRows, nil
}

func (m module) fetchDSDetails(ctx context.Context, startStr, endStr, projectFilter, assetID string) (DSDetailsData, error) {
	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	usageMgr, err := NewDSUsageManager(ctx, gcpProjectID)
	var auditRecords []DSRecord
	if err == nil {
		defer usageMgr.Close()
		records, err := usageMgr.FetchDSUsage(ctx, assetID, startStr, endStr)
		if err == nil {
			auditRecords = records
		} else {
			m.l.Printf("DS USAGE MANAGER FETCH ERROR: %v", err)
		}
	}

	type keyAggr struct {
		runs  int64
		time  float64
		cost  float64
		bytes float64
	}

	rawMap := make(map[string]keyAggr)
	metaMap := make(map[string][3]string)

	for _, rec := range auditRecords {
		if projectFilter != "" && rec.ProjectName != projectFilter {
			continue
		}

		key := fmt.Sprintf("%s|%s|%s", rec.ResourceName, rec.Credentials, rec.QueryType)

		aggr := rawMap[key]
		aggr.runs += rec.Runs
		aggr.time += rec.RunTimeSec
		aggr.cost += rec.Cost
		aggr.bytes += rec.TotalBytes

		rawMap[key] = aggr
		metaMap[key] = [3]string{rec.ResourceName, rec.Credentials, rec.QueryType}
	}

	var details []DSQueryDetail
	for key, aggr := range rawMap {
		meta := metaMap[key]

		avgTime := 0.0
		if aggr.runs > 0 {
			avgTime = aggr.time / float64(aggr.runs)
		}

		details = append(details, DSQueryDetail{
			ResourceName: meta[0],
			Credentials:  meta[1],
			QueryType:    meta[2],
			TotalRuns:    aggr.runs,
			AvgRunTime:   fmt.Sprintf("%.2fs", avgTime),
			TotalSize:    formatBytes(aggr.bytes),
			TotalCost:    fmt.Sprintf("$%.2f", aggr.cost),
			TotalNum:     aggr.cost,
		})
	}

	sort.Slice(details, func(i, j int) bool { return details[i].TotalNum > details[j].TotalNum })
	return DSDetailsData{Details: details}, nil
}
