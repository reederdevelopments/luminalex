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

type LineageRecord struct {
	DocID        string    `firestore:"doc_id"`
	JobLabel     string    `firestore:"job_label"`
	UsageDate    string    `firestore:"usage_date"`
	ProjectName  string    `firestore:"project_name"`
	SchemaName   string    `firestore:"schema_name"`
	TableName    string    `firestore:"table_name"`
	Runs         int64     `firestore:"runs"`
	Cost         float64   `firestore:"cost"`
	LastCachedAt time.Time `firestore:"last_cached_at"`
}

type LineageManager struct {
	bqClient *bigquery.Client
	fsClient *firestore.Client
	project  string
}

func NewLineageManager(ctx context.Context, projectID string) (*LineageManager, error) {
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

	// FIX: Explicitly target the 'controlroom' Firestore database instance instead of (default)
	fs, err := firestore.NewClientWithDatabase(ctx, projectID, "controlroom")
	if err != nil {
		bq.Close()
		return nil, fmt.Errorf("failed to init firestore on controlroom db: %v", err)
	}

	return &LineageManager{
		bqClient: bq,
		fsClient: fs,
		project:  projectID,
	}, nil
}

func (lm *LineageManager) Close() {
	if lm.bqClient != nil {
		lm.bqClient.Close()
	}
	if lm.fsClient != nil {
		lm.fsClient.Close()
	}
}

func genLineageKey(jobLabel, schema, table, dateStr string) string {
	input := fmt.Sprintf("%s:%s:%s:%s", jobLabel, schema, table, dateStr)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
}

func (lm *LineageManager) FetchLineage(ctx context.Context, jobLabel, startStr, endStr string) ([]LineageRecord, error) {
	start, _ := time.Parse("2006-01-02", startStr)
	end, _ := time.Parse("2006-01-02", endStr)
	todayStr := time.Now().Format("2006-01-02")

	cachedMap, err := lm.loadFromFirestore(ctx, jobLabel, startStr, endStr)
	if err != nil {
		return nil, err
	}

	var finalRecords []LineageRecord
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
		} else {
			if insideMissingWindow {
				lm.syncMissingWindow(ctx, jobLabel, missingIntervalStart, curr.AddDate(0, 0, -1), &finalRecords)
				insideMissingWindow = false
			}
		}
		curr = curr.AddDate(0, 0, 1)
	}

	if insideMissingWindow {
		lm.syncMissingWindow(ctx, jobLabel, missingIntervalStart, end, &finalRecords)
	}

	return finalRecords, nil
}

func (lm *LineageManager) syncMissingWindow(ctx context.Context, jobLabel string, start, end time.Time, output *[]LineageRecord) {
	bqRecords, err := lm.fetchFromBigQuery(ctx, jobLabel, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err == nil && len(bqRecords) > 0 {
		_ = lm.saveToFirestore(ctx, bqRecords)
		*output = append(*output, bqRecords...)
	}
}

func (lm *LineageManager) loadFromFirestore(ctx context.Context, jobLabel, start, end string) (map[string][]LineageRecord, error) {
	cachedMap := make(map[string][]LineageRecord)

	iter := lm.fsClient.Collection("dataform_lineage").
		Where("job_label", "==", jobLabel).
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

		var r LineageRecord
		if err := doc.DataTo(&r); err != nil {
			return nil, err
		}
		cachedMap[r.UsageDate] = append(cachedMap[r.UsageDate], r)
	}
	return cachedMap, nil
}

func (lm *LineageManager) fetchFromBigQuery(ctx context.Context, jobLabel, start, end string) ([]LineageRecord, error) {
	queryStr := fmt.Sprintf(`
		SELECT 
		  CAST(DATE(creation_time) AS STRING) as usage_date,
		  destination_table.project_id as project_name,
		  destination_table.dataset_id as schema_name,
		  destination_table.table_id as table_name,
		  COUNT(*) as runs,
		  (SUM(total_bytes_billed) / POWER(1024, 4)) * 6.25 AS cost
		FROM 
		  ` + "`df-fs-insights.region-europe-west9.INFORMATION_SCHEMA.JOBS`" + `
		WHERE 
		  DATE(creation_time) >= DATE(@start)
		  AND DATE(creation_time) <= DATE(@end)
		  AND EXISTS (
			SELECT 1 
			FROM UNNEST(labels) 
			WHERE value = @job_label 
		  )
		  AND destination_table.table_id IS NOT NULL
		  AND LEFT(destination_table.dataset_id, 1) != '_'
		GROUP BY ALL
	`)

	q := lm.bqClient.Query(queryStr)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start", Value: start},
		{Name: "end", Value: end},
		{Name: "job_label", Value: jobLabel},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var records []LineageRecord
	for {
		var row struct {
			UsageDate   string  `bigquery:"usage_date"`
			ProjectName string  `bigquery:"project_name"`
			SchemaName  string  `bigquery:"schema_name"`
			TableName   string  `bigquery:"table_name"`
			Runs        int64   `bigquery:"runs"`
			Cost        float64 `bigquery:"cost"`
		}
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		docID := genLineageKey(jobLabel, row.SchemaName, row.TableName, row.UsageDate)
		records = append(records, LineageRecord{
			DocID:        docID,
			JobLabel:     jobLabel,
			UsageDate:    row.UsageDate,
			ProjectName:  row.ProjectName,
			SchemaName:   row.SchemaName,
			TableName:    row.TableName,
			Runs:         row.Runs,
			Cost:         row.Cost,
			LastCachedAt: time.Now(),
		})
	}
	return records, nil
}

func (lm *LineageManager) saveToFirestore(ctx context.Context, records []LineageRecord) error {
	bulkWriter := lm.fsClient.BulkWriter(ctx)
	for _, r := range records {
		docRef := lm.fsClient.Collection("dataform_lineage").Doc(r.DocID)
		_, _ = bulkWriter.Set(docRef, r)
	}
	bulkWriter.Flush()
	return nil
}
