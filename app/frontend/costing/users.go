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

type UserCost struct {
	UserEmail    string
	SafeID       string // Safe for CSS Selectors
	URLSafeEmail string // Safe for URL Query Params
	TotalSize    string
	TotalCost    string
	TotalNum     float64
}

type UserTableData struct {
	Users      []UserCost
	GrandTotal string
	GrandSize  string
}

type UserResourceCost struct {
	QueryType string
	TotalSize string
	TotalCost string
	TotalNum  float64
	Months    []MonthData
}

type UserDetailsData struct {
	MonthHeaders []string
	Details      []UserResourceCost
}

func formatBytes(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.0f B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", b/float64(div), "KMGTPE"[exp])
}

// Generate a completely clean ID for HTMX targeting
func makeSafeID(email string) string {
	safe := strings.ReplaceAll(email, "@", "-")
	safe = strings.ReplaceAll(safe, ".", "-")
	safe = strings.ReplaceAll(safe, "+", "-")
	safe = strings.ReplaceAll(safe, "_", "-")
	return "usr-" + safe
}

func getMultiProjectFrom(projectFilter, selectClause, whereClause string) string {
	allProjects := []string{
		"df-fs-insights",
		"df-ps-south-africa",
		"df-ps-kenya",
		"df-ps-uganda",
		"df-ps-zambia",
		"df-ps-tanzania",
		"df-ps-staging",
		"df-fs-rt-insights",
	}

	var targetProjects []string
	if projectFilter != "" {
		targetProjects = []string{projectFilter}
	} else {
		targetProjects = allProjects
	}

	var clauses []string
	for _, p := range targetProjects {
		clauses = append(clauses, fmt.Sprintf("SELECT %s FROM `%s.region-europe-west9.INFORMATION_SCHEMA.JOBS_BY_PROJECT` WHERE %s", selectClause, p, whereClause))
	}

	if len(clauses) == 1 {
		return clauses[0]
	}
	return "(" + strings.Join(clauses, " UNION ALL ") + ")"
}

func (m module) usersTabHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()

	daysSinceMonday := int(now.Weekday()) - 1
	if daysSinceMonday < 0 {
		daysSinceMonday = 6
	}
	startOfWeek := now.AddDate(0, 0, -daysSinceMonday)

	startStr := startOfWeek.Format("2006-01-02")
	endStr := now.Format("2006-01-02")

	usersList, _ := m.fetchUserEmails(ctx)

	data := CostingDashboardData{
		StartDate: startStr,
		EndDate:   endStr,
		Labels:    usersList,
	}
	return usersTab(data).Render(ctx, w)
}

func (m module) usersMetricsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	grouping := r.URL.Query().Get("grouping")
	userFilter := r.URL.Query().Get("user_email")
	projectFilter := r.URL.Query().Get("project")

	if grouping == "" {
		grouping = "daily"
	}

	tableData, chartData, err := m.fetchUsersBilling(ctx, startDateStr, endDateStr, projectFilter, grouping, userFilter)
	if err != nil {
		m.l.Printf("ERROR fetching Users billing data: %v", err)
	}

	chartJSON, _ := json.Marshal(chartData)
	return usersMetrics(tableData, string(chartJSON), startDateStr, endDateStr, projectFilter).Render(ctx, w)
}

func (m module) usersDetailsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userEmail := r.URL.Query().Get("user_email")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	projectFilter := r.URL.Query().Get("project")

	details, err := m.fetchUserDetails(ctx, startDateStr, endDateStr, projectFilter, userEmail)
	if err != nil {
		m.l.Printf("ERROR fetching User details: %v", err)
	}

	return usersDetailsRow(details).Render(ctx, w)
}

func (m module) fetchUserEmails(ctx context.Context) ([]string, error) {
	cacheKey := "users_dropdown_emails"
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

	selects := "user_email"
	wheres := "creation_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY) AND user_email IS NOT NULL AND job_type = 'QUERY' AND user_email NOT LIKE '%gserviceaccount.com%'"
	multiProjectFrom := getMultiProjectFrom("", selects, wheres)

	queryStr := fmt.Sprintf(`SELECT DISTINCT user_email FROM %s`, multiProjectFrom)
	q := client.Query(queryStr)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var users []string
	for {
		var row struct {
			UserEmail string `bigquery:"user_email"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		if row.UserEmail != "" {
			users = append(users, row.UserEmail)
		}
	}

	sort.Strings(users)
	m.cache.SetLabels(cacheKey, users)
	return users, nil
}

func (m module) fetchUsersBilling(ctx context.Context, startStr, endStr, projectFilter, grouping, userFilter string) (UserTableData, []ChartPoint, error) {
	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "df-frontend"
	}

	client, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return UserTableData{}, nil, err
	}
	defer client.Close()

	selects := "user_email, total_bytes_billed, creation_time"
	wheres := "creation_time >= TIMESTAMP(@start) AND creation_time < TIMESTAMP_ADD(TIMESTAMP(@end), INTERVAL 1 DAY)  AND total_bytes_billed > 0 AND job_type = 'QUERY' AND user_email NOT LIKE '%gserviceaccount.com%'"
	if userFilter != "" {
		wheres += " AND user_email = @user_email"
	}

	multiProjectFrom := getMultiProjectFrom(projectFilter, selects, wheres)

	tableQueryStr := fmt.Sprintf(`
		SELECT 
			user_email,
			SUM(total_bytes_billed) as total_bytes,
			(SUM(total_bytes_billed) / POWER(1024, 4)) * 6.25 AS total_cost
		FROM %s
		GROUP BY user_email
		ORDER BY total_cost DESC
	`, multiProjectFrom)

	tq := client.Query(tableQueryStr)
	tq.Parameters = []bigquery.QueryParameter{{Name: "start", Value: startStr}, {Name: "end", Value: endStr}}
	if userFilter != "" {
		tq.Parameters = append(tq.Parameters, bigquery.QueryParameter{Name: "user_email", Value: userFilter})
	}

	var users []UserCost
	var grandTotalCost float64
	var grandTotalBytes float64

	it, err := tq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				UserEmail  string               `bigquery:"user_email"`
				TotalBytes bigquery.NullInt64   `bigquery:"total_bytes"`
				TotalCost  bigquery.NullFloat64 `bigquery:"total_cost"`
			}
			if err := it.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return UserTableData{}, nil, err
			}

			// Generate safe properties for DOM routing and URL queries
			users = append(users, UserCost{
				UserEmail:    row.UserEmail,
				SafeID:       makeSafeID(row.UserEmail),
				URLSafeEmail: url.QueryEscape(row.UserEmail),
				TotalSize:    formatBytes(float64(row.TotalBytes.Int64)),
				TotalCost:    fmt.Sprintf("$%.2f", row.TotalCost.Float64),
				TotalNum:     row.TotalCost.Float64,
			})
			grandTotalCost += row.TotalCost.Float64
			grandTotalBytes += float64(row.TotalBytes.Int64)
		}
	}

	dateSelect := "CAST(DATE(creation_time) AS STRING)"
	if grouping == "monthly" {
		dateSelect = "CAST(FORMAT_DATE('%Y-%m', creation_time) AS STRING)"
	} else if grouping == "yearly" {
		dateSelect = "CAST(FORMAT_DATE('%Y', creation_time) AS STRING)"
	}

	chartQueryStr := fmt.Sprintf(`
		SELECT 
			%s as usage_date,
			user_email,
			(SUM(total_bytes_billed) / POWER(1024, 4)) * 6.25 AS daily_cost
		FROM %s
		GROUP BY usage_date, user_email
		ORDER BY usage_date ASC
	`, dateSelect, multiProjectFrom)

	cq := client.Query(chartQueryStr)
	cq.Parameters = []bigquery.QueryParameter{{Name: "start", Value: startStr}, {Name: "end", Value: endStr}}
	if userFilter != "" {
		cq.Parameters = append(cq.Parameters, bigquery.QueryParameter{Name: "user_email", Value: userFilter})
	}

	var chartRows []ChartPoint
	itChart, err := cq.Read(ctx)
	if err == nil {
		for {
			var row struct {
				UsageDate string               `bigquery:"usage_date"`
				UserEmail string               `bigquery:"user_email"`
				DailyCost bigquery.NullFloat64 `bigquery:"daily_cost"`
			}
			if err := itChart.Next(&row); err != nil {
				if err == iterator.Done {
					break
				}
				return UserTableData{}, nil, err
			}
			chartRows = append(chartRows, ChartPoint{
				Date:    row.UsageDate,
				Service: row.UserEmail,
				Cost:    row.DailyCost.Float64,
			})
		}
	}

	return UserTableData{
		Users:      users,
		GrandTotal: fmt.Sprintf("$%.2f", grandTotalCost),
		GrandSize:  formatBytes(grandTotalBytes),
	}, chartRows, nil
}

func (m module) fetchUserDetails(ctx context.Context, startStr, endStr, projectFilter, userEmail string) (UserDetailsData, error) {
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

	// Fetch via optimized Read-Through Manager
	var auditRecords []UserUsageRecord
	usageMgr, err := NewUserUsageManager(ctx, gcpProjectID)
	if err == nil {
		defer usageMgr.Close()
		records, err := usageMgr.FetchUserUsage(ctx, userEmail, startStr, endStr)
		if err == nil {
			auditRecords = records
		} else {
			m.l.Printf("FIRESTORE/BQ FETCH ERROR: %v", err)
		}
	} else {
		m.l.Printf("CACHE MANAGER INIT ERROR: %v", err)
	}

	// Pivot into memory Map
	type monthAggr struct {
		cost  float64
		bytes float64
	}
	rawMap := make(map[string]map[string]monthAggr)

	for _, rec := range auditRecords {
		// Pure in-memory filtering for absolute maximum performance
		if projectFilter != "" && rec.ProjectName != projectFilter {
			continue
		}

		t, err := time.Parse("2006-01-02", rec.UsageDate)
		if err != nil {
			continue
		}
		mStr := t.Format("2006-01")

		if _, exists := rawMap[rec.QueryType]; !exists {
			rawMap[rec.QueryType] = make(map[string]monthAggr)
		}

		aggr := rawMap[rec.QueryType][mStr]
		aggr.cost += rec.Cost
		aggr.bytes += rec.TotalBytes
		rawMap[rec.QueryType][mStr] = aggr
	}

	// Translate to Final UI Model
	var details []UserResourceCost
	for qType, monthMap := range rawMap {
		var monthsData []MonthData
		var totalCostForRange float64
		var totalBytesForRange float64

		for _, m := range requestedMonths {
			aggr := monthMap[m]
			totalCostForRange += aggr.cost
			totalBytesForRange += aggr.bytes

			prevMDate, _ := time.Parse("2006-01", m)
			prevMStr := prevMDate.AddDate(0, -1, 0).Format("2006-01")
			prevCost := monthMap[prevMStr].cost

			monthsData = append(monthsData, MonthData{
				MonthName: formatMonthHeader(m),
				Cost:      fmt.Sprintf("$%.2f", aggr.cost),
				Delta:     calcDelta(aggr.cost, prevCost),
			})
		}

		if totalBytesForRange > 0 {
			details = append(details, UserResourceCost{
				QueryType: qType,
				TotalSize: formatBytes(totalBytesForRange),
				TotalCost: fmt.Sprintf("$%.2f", totalCostForRange),
				TotalNum:  totalCostForRange,
				Months:    monthsData,
			})
		}
	}

	sort.Slice(details, func(i, j int) bool { return details[i].TotalNum > details[j].TotalNum })
	return UserDetailsData{MonthHeaders: formattedHeaders, Details: details}, nil
}
