package bq

import (
	"context"
	"fmt"
	"log"
	"strings"

	"cloud.google.com/go/bigquery"
)

func NewClient(ctx context.Context, projectID string) (*bigquery.Client, error) {
	return bigquery.NewClient(ctx, projectID)
}

func EnsureSchema(ctx context.Context, client *bigquery.Client, datasetID string) error {
	log.Printf("Ensuring BigQuery dataset '%s' exists...", datasetID)
	md := &bigquery.DatasetMetadata{
		Location: "US", // You can change this to your preferred location
	}
	dataset := client.Dataset(datasetID)
	if err := dataset.Create(ctx, md); err != nil {
		if !isAlreadyExistsError(err) {
			return fmt.Errorf("failed to create dataset: %w", err)
		}
		log.Printf("Dataset '%s' already exists.", datasetID)
	} else {
		log.Printf("Dataset '%s' created.", datasetID)
	}

	if err := ensureSurveysTable(ctx, dataset); err != nil {
		return err
	}

	if err := ensureResponsesTable(ctx, dataset); err != nil {
		return err
	}

	if err := ensureSpecialSurveysTable(ctx, dataset); err != nil {
		return err
	}

	return nil
}

func ensureSurveysTable(ctx context.Context, dataset *bigquery.Dataset) error {
	log.Println("Ensuring 'surveys' table exists and schema is up to date...")
	table := dataset.Table("surveys")

	questionSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType, Required: true},
		{Name: "text", Type: bigquery.StringFieldType, Required: true},
		{Name: "type", Type: bigquery.StringFieldType, Required: true},
		{Name: "options", Type: bigquery.StringFieldType, Repeated: true},
		{Name: "is_required", Type: bigquery.BooleanFieldType, Required: true},
		{Name: "group_number", Type: bigquery.IntegerFieldType},
		{Name: "prefill_variable", Type: bigquery.StringFieldType},
	}

	schema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType, Required: true},
		{Name: "name", Type: bigquery.StringFieldType, Required: true},
		{Name: "description", Type: bigquery.StringFieldType},
		{Name: "banner", Type: bigquery.StringFieldType},
		{Name: "is_enabled", Type: bigquery.BooleanFieldType, Required: true},
		{Name: "allow_multiple_submissions", Type: bigquery.BooleanFieldType, Required: true},
		{Name: "created_at", Type: bigquery.TimestampFieldType, Required: true},
		{Name: "updated_at", Type: bigquery.TimestampFieldType, Required: true},
		{Name: "questions", Type: bigquery.RecordFieldType, Repeated: true, Schema: questionSchema},
		{Name: "group_headings", Type: bigquery.StringFieldType, Repeated: true},
		{Name: "type", Type: bigquery.StringFieldType},
	}

	if err := createOrUpdateTable(ctx, table, schema); err != nil {
		return fmt.Errorf("surveys table: %w", err)
	}
	return nil
}

func ensureResponsesTable(ctx context.Context, dataset *bigquery.Dataset) error {
	log.Println("Ensuring 'responses' table exists and schema is up to date...")
	table := dataset.Table("responses")

	answerSchema := bigquery.Schema{
		{Name: "question_id", Type: bigquery.StringFieldType, Required: true},
		{Name: "values", Type: bigquery.StringFieldType, Repeated: true},
	}

	schema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType, Required: true},
		{Name: "survey_id", Type: bigquery.StringFieldType, Required: true},
		{Name: "user_id", Type: bigquery.StringFieldType, Required: true},
		{Name: "submitted_at", Type: bigquery.TimestampFieldType, Required: true},
		{Name: "answers", Type: bigquery.RecordFieldType, Repeated: true, Schema: answerSchema},
	}

	if err := createOrUpdateTable(ctx, table, schema); err != nil {
		return fmt.Errorf("responses table: %w", err)
	}
	return nil
}

func ensureSpecialSurveysTable(ctx context.Context, dataset *bigquery.Dataset) error {
	log.Println("Ensuring 'special_surveys' table exists and schema is up to date...")
	table := dataset.Table("special_surveys")

	schema := bigquery.Schema{
		{Name: "assignment_id", Type: bigquery.StringFieldType, Required: true},
		{Name: "survey_id", Type: bigquery.StringFieldType, Required: true},
		{Name: "user_email", Type: bigquery.StringFieldType, Required: true},
		{Name: "variable_1", Type: bigquery.StringFieldType},
		{Name: "variable_2", Type: bigquery.StringFieldType},
		{Name: "variable_3", Type: bigquery.StringFieldType},
		{Name: "variable_4", Type: bigquery.StringFieldType},
		{Name: "variable_5", Type: bigquery.StringFieldType},
		{Name: "response_id", Type: bigquery.StringFieldType},
	}

	if err := createOrUpdateTable(ctx, table, schema); err != nil {
		return fmt.Errorf("special_surveys table: %w", err)
	}
	return nil
}

func createOrUpdateTable(ctx context.Context, table *bigquery.Table, schema bigquery.Schema) error {
	meta, err := table.Metadata(ctx)
	if err != nil {
		if !isNotFoundError(err) {
			return fmt.Errorf("getting table metadata: %w", err)
		}
		// Table does not exist, create it.
		log.Printf("Table '%s' not found, creating...", table.TableID)
		if err := table.Create(ctx, &bigquery.TableMetadata{Schema: schema}); err != nil {
			return fmt.Errorf("creating table: %w", err)
		}
		log.Printf("Table '%s' created.", table.TableID)
		return nil
	}

	// Table exists, check for schema update.
	// update := bigquery.SchemaUpdateOption{Schema: schema}
	if _, err := table.Update(ctx, bigquery.TableMetadataToUpdate{Schema: schema}, meta.ETag); err != nil {
		// Attempting to update with the same schema might cause an error, which we can ignore.
		if !strings.Contains(err.Error(), "no changes detected") {
			// This might fail if there are breaking changes. Manual intervention might be needed.
			log.Printf("Warning: failed to update schema for table '%s'. This might be expected if no changes are needed, or it could indicate a problem: %v", table.TableID, err)
		}
	} else {
		log.Printf("Schema for table '%s' is up to date.", table.TableID)
	}

	return nil
}

func isAlreadyExistsError(err error) bool {
	return strings.Contains(err.Error(), "Already Exists")
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "Not found")
}
