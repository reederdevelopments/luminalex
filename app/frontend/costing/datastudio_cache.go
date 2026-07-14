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

type DSRecord struct {
	DocID        string  `firestore:"doc_id"`
	AssetID      string  `firestore:"asset_id"`
	ProjectName  string  `firestore:"project_name"`
	UsageDate    string  `firestore:"usage_date"`
	QueryType    string  `firestore:"query_type"`
	ResourceName string  `firestore:"resource_name"`
	Credentials  string  `firestore:"credentials"`
	Runs         int64   `firestore:"runs"`
	RunTimeSec   float64 `firestore:"runtime_sec"`
	TotalBytes   float64 `firestore:"total_bytes"`
	Cost         float64 `firestore:"cost"`
}

type DSUsageManager struct {
	bqClient *bigquery.Client
	fsClient *firestore.Client
}

func NewDSUsageManager(ctx context.Context, projectID string) (*DSUsageManager, error) {
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_PROJECT_ID")
		if projectID == "" {
			projectID = "df-frontend"
		}
	}
	bq, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	fs, err := firestore.NewClientWithDatabase(ctx, projectID, "controlroom")
	if err != nil {
		bq.Close()
		return nil, err
	}
	return &DSUsageManager{bqClient: bq, fsClient: fs}, nil
}

func (m *DSUsageManager) Close() {
	if m.bqClient != nil {
		m.bqClient.Close()
	}
	if m.fsClient != nil {
		m.fsClient.Close()
	}
}

func genDSKey(assetID, projectName, resourceName, dateStr string) string {
	input := fmt.Sprintf("%s:%s:%s:%s", assetID, projectName, resourceName, dateStr)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}

func (m *DSUsageManager) FetchDSUsage(ctx context.Context, assetID, startStr, endStr string) ([]DSRecord, error) {
	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)
	todayStr := time.Now().Format("2006-01-02")

	cachedMap, err := m.loadFromFirestore(ctx, assetID, startStr, endStr)
	if err != nil {
		return nil, err
	}

	var finalRecords []DSRecord
	curr := start
	var missingIntervalStart time.Time
	var insideMissingWindow bool

	for !curr.After(end) {
		dateStr := curr.Format("2006-01-02")
		hasCache := false

		if recs, found := cachedMap[dateStr]; found && len(recs) > 0 && dateStr != todayStr {
			hasCache = true
			finalRecords = append(finalRecords, recs...)
		}

		if !hasCache {
			if !insideMissingWindow {
				missingIntervalStart = curr
				insideMissingWindow = true
			}
		} else if insideMissingWindow {
			m.syncMissingWindow(ctx, assetID, missingIntervalStart, curr.AddDate(0, 0, -1), &finalRecords)
			insideMissingWindow = false
		}
		curr = curr.AddDate(0, 0, 1)
	}

	if insideMissingWindow {
		m.syncMissingWindow(ctx, assetID, missingIntervalStart, end, &finalRecords)
	}
	return finalRecords, nil
}

func (m *DSUsageManager) syncMissingWindow(ctx context.Context, assetID string, start, end time.Time, output *[]DSRecord) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	bqRecords, err := m.fetchFromBigQuery(ctxTimeout, assetID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err == nil && len(bqRecords) > 0 {
		_ = m.saveToFirestore(ctx, bqRecords)
		*output = append(*output, bqRecords...)
	}
}

func (m *DSUsageManager) loadFromFirestore(ctx context.Context, assetID, start, end string) (map[string][]DSRecord, error) {
	cachedMap := make(map[string][]DSRecord)

	// FIX: We only filter by asset_id to utilize the default single-field index.
	// We handle the Date boundaries strictly in memory to avoid the need for a custom composite index.
	iter := m.fsClient.Collection("ds_usage_cache").
		Where("asset_id", "==", assetID).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var r DSRecord
		if err := doc.DataTo(&r); err == nil {
			if r.UsageDate >= start && r.UsageDate <= end {
				cachedMap[r.UsageDate] = append(cachedMap[r.UsageDate], r)
			}
		}
	}
	return cachedMap, nil
}

func (m *DSUsageManager) fetchFromBigQuery(ctx context.Context, assetID, start, end string) ([]DSRecord, error) {
	selects := "project_id as project_name, labels, user_email, query, referenced_tables, start_time, end_time, creation_time, total_bytes_billed"
	wheres := "creation_time >= TIMESTAMP(@start) AND creation_time < TIMESTAMP_ADD(TIMESTAMP(@end), INTERVAL 1 DAY) AND total_bytes_billed > 0 AND EXISTS (SELECT 1 FROM UNNEST(labels) WHERE (key = 'looker_studio_report_id' OR key = 'datastudio_report_id') AND value = @asset_id)"
	multiProjectFrom := getMultiProjectFrom("", selects, wheres)

	queryStr := fmt.Sprintf(`
		SELECT 
			project_name,
			CAST(DATE(creation_time) AS STRING) as usage_date,
			COALESCE(user_email, 'Service Account / Viewer') as credentials,
			CASE WHEN query LIKE '%%(SELECT%%' OR query LIKE '%%JOIN%%' THEN 'Custom Query' ELSE 'Standard Table' END as query_type,
			COALESCE(
				(SELECT STRING_AGG(CONCAT(dataset_id, '.', table_id), ' | ') FROM UNNEST(referenced_tables) WHERE SUBSTR(dataset_id, 1, 1) != '_'),
				'Computed / Complex Query'
			) as resource_name,
			COUNT(*) as runs,
			SUM(TIMESTAMP_DIFF(end_time, start_time, MILLISECOND)) / 1000.0 as runtime_sec,
			SUM(total_bytes_billed) as total_bytes,
			(SUM(total_bytes_billed) / POWER(1024, 4)) * 6.25 AS cost
		FROM %s
		GROUP BY project_name, usage_date, credentials, query_type, resource_name
	`, multiProjectFrom)

	q := m.bqClient.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
		{Name: "asset_id", Value: assetID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var records []DSRecord
	for {
		var row struct {
			ProjectName  string               `bigquery:"project_name"`
			UsageDate    string               `bigquery:"usage_date"`
			Credentials  bigquery.NullString  `bigquery:"credentials"`
			QueryType    string               `bigquery:"query_type"`
			ResourceName bigquery.NullString  `bigquery:"resource_name"`
			Runs         bigquery.NullInt64   `bigquery:"runs"`
			RuntimeSec   bigquery.NullFloat64 `bigquery:"runtime_sec"`
			TotalBytes   bigquery.NullInt64   `bigquery:"total_bytes"`
			Cost         bigquery.NullFloat64 `bigquery:"cost"`
		}

		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		docID := genDSKey(assetID, row.ProjectName, row.ResourceName.StringVal, row.UsageDate)
		records = append(records, DSRecord{
			DocID:        docID,
			AssetID:      assetID,
			ProjectName:  row.ProjectName,
			UsageDate:    row.UsageDate,
			QueryType:    row.QueryType,
			ResourceName: row.ResourceName.StringVal,
			Credentials:  row.Credentials.StringVal,
			Runs:         row.Runs.Int64,
			RunTimeSec:   row.RuntimeSec.Float64,
			TotalBytes:   float64(row.TotalBytes.Int64),
			Cost:         row.Cost.Float64,
		})
	}
	return records, nil
}

func (m *DSUsageManager) saveToFirestore(ctx context.Context, records []DSRecord) error {
	bulkWriter := m.fsClient.BulkWriter(ctx)
	for _, r := range records {
		docRef := m.fsClient.Collection("ds_usage_cache").Doc(r.DocID)
		_, _ = bulkWriter.Set(docRef, r)
	}
	bulkWriter.Flush()
	return nil
}
