package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

func (m module) thirdpartyTabHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	data := MonitoringDashboardData{
		StartDate: firstOfMonth.Format("2006-01-02"),
		EndDate:   now.Format("2006-01-02"),
	}

	if r.Header.Get("HX-Request") == "true" {
		return thirdpartyTab(data).Render(ctx, w)
	}

	return monitoringPage(data).Render(ctx, w)
}

func (m module) thirdpartyListHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	if startStr == "" || endStr == "" {
		now := time.Now()
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		startStr = firstOfMonth.Format("2006-01-02")
		endStr = now.Format("2006-01-02")
	}
	summaries, err := m.fetchJobSummaries(ctx, startStr, endStr)
	if err != nil {
		m.l.Printf("ERROR fetching job summaries: %v", err)
	}
	return thirdpartyList(summaries, startStr, endStr).Render(ctx, w)
}

func (m module) thirdpartyDetailHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	jobName := r.URL.Query().Get("job")
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	now := time.Now()
	if startStr == "" || endStr == "" {
		startStr = now.AddDate(0, 0, -7).Format("2006-01-02")
		endStr = now.Format("2006-01-02")
	}

	logs, err := m.fetchJobLogs(ctx, jobName, startStr, endStr)
	if err != nil {
		m.l.Printf("ERROR fetching job logs: %v", err)
	}

	var latest LogEntry
	var parsedMetrics ParsedMetrics
	if len(logs) > 0 {
		latest = logs[0]
		if latest.MetricsJSON != "" {
			_ = json.Unmarshal([]byte(latest.MetricsJSON), &parsedMetrics)
		}
	}

	data := MonitoringDashboardData{
		StartDate: startStr,
		EndDate:   endStr,
	}

	return thirdpartyDetail(jobName, data, logs, latest, parsedMetrics).Render(ctx, w)
}

func (m module) triggerJobHandler(w http.ResponseWriter, r *http.Request) error {
	jobName := r.URL.Query().Get("job")
	m.l.Printf("Triggering job: %s", jobName)
	w.WriteHeader(http.StatusOK)
	return nil
}

func (m module) fetchJobSummaries(ctx context.Context, startStr, endStr string) ([]JobSummary, error) {
	key := startStr + endStr
	if summaries, ok := m.cache.GetSummaries(key); ok {
		return summaries, nil
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

	queryStr := `
		WITH job_logs AS (
			SELECT 
				JOB_NAME,
				STATUS,
				START_TIME,
				TOTAL_ROWS,
				TOTAL_BYTES
			FROM ` + "`df-ps-staging.LOAD_LOGS.LOGS`" + `
			WHERE DATE(START_TIME) >= DATE(@start) 
			  AND DATE(START_TIME) <= DATE(@end)
		),
		job_stats AS (
			SELECT 
				JOB_NAME,
				SUM(TOTAL_ROWS) as sum_rows,
				SUM(TOTAL_BYTES) as sum_bytes,
				MAX(START_TIME) as last_run_time,
				ARRAY_AGG(STATUS ORDER BY START_TIME DESC LIMIT 1)[OFFSET(0)] as last_status
			FROM job_logs
			GROUP BY JOB_NAME
		),
		billing AS (
			SELECT
				SPLIT(COALESCE(resource.name, sku.description, ''), '/')[SAFE_OFFSET(ARRAY_LENGTH(SPLIT(COALESCE(resource.name, sku.description, ''), '/')) - 1)] as clean_job_name,
				SUM(cost + IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as total_cost
			FROM ` + "`df-ps-staging.EXT_GCP_BILLING.gcp_billing_export_resource_v1_01FF43_BAACE5_55390D`" + `
			WHERE DATE(usage_start_time) >= DATE(@start) 
			  AND DATE(usage_start_time) <= DATE(@end)
			  AND service.description IN ('Cloud Functions', 'Cloud Run', 'Cloud Run Functions')
			GROUP BY clean_job_name
		)
		SELECT 
			s.JOB_NAME,
			s.last_status,
			s.last_run_time,
			s.sum_rows,
			s.sum_bytes,
			COALESCE(b.total_cost, 0.0) as total_cost
		FROM job_stats s
		LEFT JOIN billing b 
		ON b.clean_job_name = s.JOB_NAME OR b.clean_job_name LIKE CONCAT('%', s.JOB_NAME, '%')
		ORDER BY s.last_run_time DESC
	`
	q := client.Query(queryStr)
	q.Location = "europe-west9" // Explicitly define location to prevent 404s
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: startStr},
		{Name: "end", Value: endStr},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var summaries []JobSummary
	for {
		var row struct {
			JobName     string                 `bigquery:"JOB_NAME"`
			Status      string                 `bigquery:"last_status"`
			LastRunTime bigquery.NullTimestamp `bigquery:"last_run_time"`
			SumRows     bigquery.NullInt64     `bigquery:"sum_rows"`
			SumBytes    bigquery.NullInt64     `bigquery:"sum_bytes"`
			TotalCost   bigquery.NullFloat64   `bigquery:"total_cost"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		summaries = append(summaries, JobSummary{
			JobName:     row.JobName,
			Status:      row.Status,
			LastRunTime: row.LastRunTime.Timestamp,
			TotalRows:   row.SumRows.Int64,
			TotalBytes:  row.SumBytes.Int64,
			TotalCost:   row.TotalCost.Float64,
		})
	}

	m.cache.SetSummaries(key, summaries)
	return summaries, nil
}

func (m module) fetchJobLogs(ctx context.Context, jobName, startStr, endStr string) ([]LogEntry, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return nil, err
	}

	if logs, ok := m.cache.GetLogs(jobName, start, end); ok {
		return logs, nil
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

	queryStr := `
		SELECT 
			JOB_NAME, START_TIME, END_TIME, DURATION_SECONDS, STATUS, 
			TOTAL_ROWS, TOTAL_BYTES, METRICS_JSON, ERROR_MESSAGE, STACK_TRACE
		FROM ` + "`df-ps-staging.LOAD_LOGS.LOGS`" + `
		WHERE DATE(START_TIME) >= DATE(@start) 
		  AND DATE(START_TIME) <= DATE(@end)
		ORDER BY START_TIME DESC
	`
	q := client.Query(queryStr)
	q.Location = "europe-west9" // Explicitly define location to prevent 404s
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: startStr},
		{Name: "end", Value: endStr},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	fetchedData := make(map[string][]LogEntry)
	for {
		var row struct {
			JobName         bigquery.NullString    `bigquery:"JOB_NAME"`
			StartTime       bigquery.NullTimestamp `bigquery:"START_TIME"`
			EndTime         bigquery.NullTimestamp `bigquery:"END_TIME"`
			DurationSeconds bigquery.NullFloat64   `bigquery:"DURATION_SECONDS"`
			Status          bigquery.NullString    `bigquery:"STATUS"`
			TotalRows       bigquery.NullInt64     `bigquery:"TOTAL_ROWS"`
			TotalBytes      bigquery.NullInt64     `bigquery:"TOTAL_BYTES"`
			MetricsJSON     bigquery.NullString    `bigquery:"METRICS_JSON"`
			ErrorMessage    bigquery.NullString    `bigquery:"ERROR_MESSAGE"`
			StackTrace      bigquery.NullString    `bigquery:"STACK_TRACE"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		entry := LogEntry{
			JobName:         row.JobName.StringVal,
			StartTime:       row.StartTime.Timestamp,
			EndTime:         row.EndTime.Timestamp,
			DurationSeconds: row.DurationSeconds.Float64,
			Status:          row.Status.StringVal,
			TotalRows:       row.TotalRows.Int64,
			TotalBytes:      row.TotalBytes.Int64,
			MetricsJSON:     row.MetricsJSON.StringVal,
			ErrorMessage:    row.ErrorMessage.StringVal,
			StackTrace:      row.StackTrace.StringVal,
		}
		fetchedData[entry.JobName] = append(fetchedData[entry.JobName], entry)
	}

	m.cache.SetLogs(start, end, fetchedData)
	return fetchedData[jobName], nil
}
