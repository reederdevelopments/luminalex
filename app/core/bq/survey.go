package bq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maoni/app/core/survey"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	surveysTable   = "surveys"
	responsesTable = "responses"
)

type SurveyStore struct {
	client    *bigquery.Client
	datasetID string
}

func NewSurveyStore(client *bigquery.Client, datasetID string) *SurveyStore {
	return &SurveyStore{client: client, datasetID: datasetID}
}

func (s *SurveyStore) Create(ctx context.Context, su survey.Survey) error {
	jsonData, _ := json.MarshalIndent(su, "", "  ")
	log.Printf("BIGQUERY Create: Saving survey data:\n%s", string(jsonData))

	// Use a DML INSERT statement to avoid mixing streaming inserts and DML,
	// which can cause errors about affecting rows in the streaming buffer.
	q := s.client.Query(fmt.Sprintf(`
               INSERT INTO %s.%s (id, name, description, is_enabled, allow_multiple_submissions, created_at, updated_at, questions, group_headings)
               VALUES (@id, @name, @description, @is_enabled, @allow_multiple_submissions, @created_at, @updated_at, @questions, @group_headings)
       `, s.datasetID, surveysTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "id", Value: su.ID},
		{Name: "name", Value: su.Name},
		{Name: "description", Value: su.Description},
		{Name: "is_enabled", Value: su.IsEnabled},
		{Name: "allow_multiple_submissions", Value: su.AllowMultipleSubmissions},
		{Name: "created_at", Value: su.CreatedAt},
		{Name: "updated_at", Value: su.UpdatedAt},
		{Name: "questions", Value: su.Questions}, // The client library handles struct slices
		{Name: "group_headings", Value: su.GroupHeadings},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run survey insert job: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for survey insert job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("survey insert job failed: %w", err)
	}

	return nil
}

func (s *SurveyStore) Update(ctx context.Context, su survey.Survey) error {
	jsonData, _ := json.MarshalIndent(su, "", "  ")
	log.Printf("BIGQUERY Update: Saving survey data for ID %s:\n%s", su.ID, string(jsonData))
	// BQ doesn't have a simple "update row" command like SQL.
	// We need to run a MERGE statement.
	q := s.client.Query(fmt.Sprintf(`
		MERGE %s.%s T
		USING (SELECT @id as id) S
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET name = @name, description = @description, is_enabled = @is_enabled, allow_multiple_submissions = @allow_multiple_submissions, updated_at = @updated_at, questions = @questions, group_headings = @group_headings
	`, s.datasetID, surveysTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "id", Value: su.ID},
		{Name: "name", Value: su.Name},
		{Name: "description", Value: su.Description},
		{Name: "is_enabled", Value: su.IsEnabled},
		{Name: "allow_multiple_submissions", Value: su.AllowMultipleSubmissions},
		{Name: "updated_at", Value: su.UpdatedAt},
		{Name: "questions", Value: su.Questions},
		{Name: "group_headings", Value: su.GroupHeadings},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run update job: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for update job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("update job failed: %w", err)
	}

	return nil
}

func (s *SurveyStore) Get(ctx context.Context, id string) (survey.Survey, error) {
	q := s.client.Query(fmt.Sprintf("SELECT * FROM `%s.%s` WHERE id = @id LIMIT 1", s.datasetID, surveysTable))
	q.Parameters = []bigquery.QueryParameter{{Name: "id", Value: id}}

	it, err := q.Read(ctx)
	if err != nil {
		return survey.Survey{}, fmt.Errorf("failed to read survey: %w", err)
	}

	var su survey.Survey
	err = it.Next(&su)
	if err == iterator.Done {
		return survey.Survey{}, fmt.Errorf("survey not found")
	}
	if err != nil {
		return survey.Survey{}, fmt.Errorf("failed to iterate survey results: %w", err)
	}
	return su, nil
}

func (s *SurveyStore) List(ctx context.Context, showInactive bool) ([]survey.Survey, error) {
	queryStr := fmt.Sprintf("SELECT * FROM `%s.%s`", s.datasetID, surveysTable)
	if !showInactive {
		queryStr += " WHERE is_enabled = TRUE"
	}
	queryStr += " ORDER BY created_at DESC"

	q := s.client.Query(queryStr)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list surveys: %w", err)
	}

	var surveys []survey.Survey
	for {
		var su survey.Survey
		err := it.Next(&su)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate survey list: %w", err)
		}

		count, err := s.GetResponseCount(ctx, su.ID)
		if err != nil {
			// Log the error but don't fail the whole list
			fmt.Printf("could not get response count for survey %s: %v", su.ID, err)
		}
		su.ResponseCount = count

		surveys = append(surveys, su)
	}

	return surveys, nil
}

func (s *SurveyStore) SaveResponse(ctx context.Context, r survey.Response) error {
	inserter := s.client.Dataset(s.datasetID).Table(responsesTable).Inserter()
	if err := inserter.Put(ctx, r); err != nil {
		return fmt.Errorf("failed to insert response: %w", err)
	}
	return nil
}

func (s *SurveyStore) HasUserResponded(ctx context.Context, surveyID, userID string) (bool, error) {
	q := s.client.Query(fmt.Sprintf("SELECT COUNT(id) FROM `%s.%s` WHERE survey_id = @survey_id AND user_id = @user_id", s.datasetID, responsesTable))
	q.Parameters = []bigquery.QueryParameter{
		{Name: "survey_id", Value: surveyID},
		{Name: "user_id", Value: userID},
	}
	it, err := q.Read(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to query responses: %w", err)
	}
	var row []bigquery.Value
	err = it.Next(&row)
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to read response count: %w", err)
	}
	count, ok := row[0].(int64)
	return ok && count > 0, nil
}

func (s *SurveyStore) GetResponseCount(ctx context.Context, surveyID string) (int, error) {
	q := s.client.Query(fmt.Sprintf("SELECT COUNT(id) FROM `%s.%s` WHERE survey_id = @survey_id", s.datasetID, responsesTable))
	q.Parameters = []bigquery.QueryParameter{
		{Name: "survey_id", Value: surveyID},
	}
	it, err := q.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query response count: %w", err)
	}
	var row []bigquery.Value
	err = it.Next(&row)
	if err == iterator.Done {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read response count: %w", err)
	}
	if len(row) > 0 {
		if count, ok := row[0].(int64); ok {
			return int(count), nil
		}
	}
	return 0, nil
}
