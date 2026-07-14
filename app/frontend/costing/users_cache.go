package costing

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type UserUsageRecord struct {
	DocID        string    `firestore:"doc_id"`
	UserEmail    string    `firestore:"user_email"`
	UsageDate    string    `firestore:"usage_date"` // YYYY-MM-DD
	ProjectName  string    `firestore:"project_name"`
	QueryType    string    `firestore:"query_type"`
	TotalBytes   float64   `firestore:"total_bytes"`
	Cost         float64   `firestore:"cost"`
	LastCachedAt time.Time `firestore:"last_cached_at"`
}

type UserUsageManager struct {
	bqClient *bigquery.Client
	fsClient *firestore.Client
	project  string
}

func NewUserUsageManager(ctx context.Context, projectID string) (*UserUsageManager, error) {
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_PROJECT_ID")
		if projectID == "" {
			projectID = "df-frontend"
		}
	}

	bq, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to init bq: %v", err)
	}

	fs, err := firestore.NewClientWithDatabase(ctx, projectID, "controlroom")
	if err != nil {
		bq.Close()
		return nil, fmt.Errorf("failed to init firestore: %v", err)
	}

	return &UserUsageManager{
		bqClient: bq,
		fsClient: fs,
		project:  projectID,
	}, nil
}

func (m *UserUsageManager) Close() {
	if m.bqClient != nil {
		m.bqClient.Close()
	}
	if m.fsClient != nil {
		m.fsClient.Close()
	}
}

func genUserUsageKey(email, projectName, qType, dateStr string) string {
	input := fmt.Sprintf("%s:%s:%s:%s", email, projectName, qType, dateStr)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}

func (m *UserUsageManager) FetchUserUsage(ctx context.Context, email, startStr, endStr string) ([]UserUsageRecord, error) {
	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)
	todayStr := time.Now().Format("2006-01-02")

	cachedMap, err := m.loadFromFirestore(ctx, email, startStr, endStr)
	if err != nil {
		return nil, err
	}

	var finalRecords []UserUsageRecord
	curr := start

	var missingIntervalStart time.Time
	var insideMissingWindow bool

	for !curr.After(end) {
		dateStr := curr.Format("2006-01-02")
		hasCache := false

		// Skip cache for "today" to ensure we get live intra-day updates
		if recs, found := cachedMap[dateStr]; found && len(recs) > 0 && dateStr != todayStr {
			hasCache = true
			finalRecords = append(finalRecords, recs...)
		}

		if !hasCache {
			if !insideMissingWindow {
				missingIntervalStart = curr
				insideMissingWindow = true
			}
		} else {
			if insideMissingWindow {
				m.syncMissingWindow(ctx, email, missingIntervalStart, curr.AddDate(0, 0, -1), &finalRecords)
				insideMissingWindow = false
			}
		}
		curr = curr.AddDate(0, 0, 1)
	}

	if insideMissingWindow {
		m.syncMissingWindow(ctx, email, missingIntervalStart, end, &finalRecords)
	}

	return finalRecords, nil
}

func (m *UserUsageManager) syncMissingWindow(ctx context.Context, email string, start, end time.Time, output *[]UserUsageRecord) {
	// Add a 60-second safeguard timeout for BQ extraction
	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	bqRecords, err := m.fetchFromBigQuery(ctxTimeout, email, start.Format("2006-01-02"), end.Format("2006-01-02"))

	if err != nil {
		fmt.Printf("ERROR fetching from BQ for user %s: %v\n", email, err)
		return
	}

	if len(bqRecords) > 0 {
		_ = m.saveToFirestore(ctx, bqRecords)
		*output = append(*output, bqRecords...)
	}
}

func (m *UserUsageManager) loadFromFirestore(ctx context.Context, email, start, end string) (map[string][]UserUsageRecord, error) {
	cachedMap := make(map[string][]UserUsageRecord)

	iter := m.fsClient.Collection("user_usage_cache").
		Where("user_email", "==", email).
		Where("usage_date", ">=", start).
		Where("usage_date", "<=", end).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var r UserUsageRecord
		if err := doc.DataTo(&r); err != nil {
			return nil, err
		}
		cachedMap[r.UsageDate] = append(cachedMap[r.UsageDate], r)
	}
	return cachedMap, nil
}

func (m *UserUsageManager) fetchFromBigQuery(ctx context.Context, email, start, end string) ([]UserUsageRecord, error) {
	// Extract across ALL projects simultaneously to prevent fragmented cache states
	selects := "project_id as project_name, labels, job_id, statement_type, job_type, creation_time, total_bytes_billed"
	wheres := "user_email = @user_email AND creation_time >= TIMESTAMP(@start) AND creation_time < TIMESTAMP_ADD(TIMESTAMP(@end), INTERVAL 1 DAY) AND total_bytes_billed > 0 AND job_type = 'QUERY'"

	// Note: getMultiProjectFrom is defined in users.go (part of the 'costing' package)
	multiProjectFrom := getMultiProjectFrom("", selects, wheres)

	queryStr := fmt.Sprintf(`
		SELECT 
			project_name,
			CAST(DATE(creation_time) AS STRING) as usage_date,
			CASE 
				WHEN EXISTS (SELECT 1 FROM UNNEST(labels) WHERE key = 'data_source_id' AND value = 'scheduled_query') THEN 'System: Scheduled Query'
				WHEN job_id LIKE '%%scheduled_query%%' THEN 'System: Scheduled Query'
				WHEN statement_type = 'CREATE_VIEW' THEN 'Manual: View Creation'
				WHEN statement_type = 'CREATE_TABLE_AS_SELECT' THEN 'Manual: Table Creation (CTAS)'
				WHEN statement_type = 'MERGE' THEN 'Manual: Merge (Upsert)'
				WHEN statement_type = 'SELECT' THEN 'Manual: Normal Query'
				WHEN statement_type = 'SCRIPT' THEN 'Manual: Multi-statement Script'
				WHEN statement_type IS NOT NULL THEN CONCAT('Manual: ', statement_type)
				ELSE CONCAT('Manual: ', COALESCE(job_type, 'Unknown'))
			END as query_type,
			SUM(total_bytes_billed) as total_bytes,
			(SUM(total_bytes_billed) / POWER(1024, 4)) * 6.25 AS cost
		FROM %s
		GROUP BY project_name, usage_date, query_type
	`, multiProjectFrom)

	q := m.bqClient.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
		{Name: "user_email", Value: email},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var records []UserUsageRecord
	for {
		var row struct {
			ProjectName string               `bigquery:"project_name"`
			UsageDate   string               `bigquery:"usage_date"`
			QueryType   string               `bigquery:"query_type"`
			TotalBytes  bigquery.NullInt64   `bigquery:"total_bytes"` // Safely capture BQ INT64
			Cost        bigquery.NullFloat64 `bigquery:"cost"`        // Safely capture BQ FLOAT64
		}

		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		docID := genUserUsageKey(email, row.ProjectName, row.QueryType, row.UsageDate)
		records = append(records, UserUsageRecord{
			DocID:        docID,
			UserEmail:    email,
			ProjectName:  row.ProjectName,
			UsageDate:    row.UsageDate,
			QueryType:    row.QueryType,
			TotalBytes:   float64(row.TotalBytes.Int64), // Cast to float64 for our struct & Firestore
			Cost:         row.Cost.Float64,
			LastCachedAt: time.Now(),
		})
	}
	return records, nil
}

func (m *UserUsageManager) saveToFirestore(ctx context.Context, records []UserUsageRecord) error {
	bulkWriter := m.fsClient.BulkWriter(ctx)
	for _, r := range records {
		docRef := m.fsClient.Collection("user_usage_cache").Doc(r.DocID)
		_, _ = bulkWriter.Set(docRef, r)
	}
	bulkWriter.Flush()
	return nil
}
