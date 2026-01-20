package bq

import (
	"context"
	"fmt"
	"maoni/app/core/survey"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	surveysTable        = "surveys"
	responsesTable      = "responses"
	specialSurveysTable = "special_surveys"
)

type SurveyStore struct {
	client    *bigquery.Client
	datasetID string
}

func NewSurveyStore(client *bigquery.Client, datasetID string) *SurveyStore {
	return &SurveyStore{client: client, datasetID: datasetID}
}

func (s *SurveyStore) Create(ctx context.Context, su survey.Survey) error {
	// Use a DML INSERT statement to avoid mixing streaming inserts and DML,
	// which can cause errors about affecting rows in the streaming buffer.
	q := s.client.Query(fmt.Sprintf(`
               INSERT INTO %s.%s (id, name, description, type, is_enabled, allow_multiple_submissions, created_at, updated_at, questions, group_headings, banner)
               VALUES (@id, @name, @description, @type, @is_enabled, @allow_multiple_submissions, @created_at, @updated_at, @questions, @group_headings, @banner)
       `, s.datasetID, surveysTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "id", Value: su.ID},
		{Name: "name", Value: su.Name},
		{Name: "description", Value: su.Description},
		{Name: "type", Value: su.Type},
		{Name: "is_enabled", Value: su.IsEnabled},
		{Name: "allow_multiple_submissions", Value: su.AllowMultipleSubmissions},
		{Name: "created_at", Value: su.CreatedAt},
		{Name: "updated_at", Value: su.UpdatedAt},
		{Name: "questions", Value: su.Questions}, // The client library handles struct slices
		{Name: "group_headings", Value: su.GroupHeadings},
		{Name: "banner", Value: su.Banner},
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
	// BQ doesn't have a simple "update row" command like SQL.
	// We need to run a MERGE statement.
	q := s.client.Query(fmt.Sprintf(`
		MERGE %s.%s T
		USING (SELECT @id as id) S
		ON T.id = S.id
		WHEN MATCHED THEN
			UPDATE SET name = @name, description = @description, type = @type, is_enabled = @is_enabled, allow_multiple_submissions = @allow_multiple_submissions, updated_at = @updated_at, questions = @questions, group_headings = @group_headings, banner = @banner
	`, s.datasetID, surveysTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "id", Value: su.ID},
		{Name: "name", Value: su.Name},
		{Name: "description", Value: su.Description},
		{Name: "type", Value: su.Type},
		{Name: "is_enabled", Value: su.IsEnabled},
		{Name: "allow_multiple_submissions", Value: su.AllowMultipleSubmissions},
		{Name: "updated_at", Value: su.UpdatedAt},
		{Name: "questions", Value: su.Questions},
		{Name: "group_headings", Value: su.GroupHeadings},
		{Name: "banner", Value: su.Banner},
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

// ListForUser retrieves all surveys a user is eligible to take.
// This includes all enabled "normal" surveys and any "special" surveys
// they have been assigned to and have not yet completed.
func (s *SurveyStore) ListForUser(ctx context.Context, userEmail string) ([]survey.Survey, error) {
	// 1. Get all enabled normal surveys
	normalQuery := s.client.Query(fmt.Sprintf("SELECT * FROM `%s.%s` WHERE is_enabled = TRUE AND (type IS NULL OR type = 'normal')", s.datasetID, surveysTable))
	normalIt, err := normalQuery.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list normal surveys: %w", err)
	}

	var result []survey.Survey
	for {
		var su survey.Survey
		err := normalIt.Next(&su)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate normal survey list: %w", err)
		}
		result = append(result, su)
	}

	// 2. Get all special survey assignments for the user that are enabled and not yet submitted
	type specialSurveyRow struct {
		survey.Survey
		AssignmentID string `bigquery:"assignment_id"`
		Variable1    string `bigquery:"variable_1"`
		Variable2    string `bigquery:"variable_2"`
		Variable3    string `bigquery:"variable_3"`
		Variable4    string `bigquery:"variable_4"`
		Variable5    string `bigquery:"variable_5"`
	}

	specialQuery := s.client.Query(fmt.Sprintf(`
		SELECT s.*, ss.assignment_id, ss.variable_1, ss.variable_2, ss.variable_3, ss.variable_4, ss.variable_5
		FROM %[1]s.%[2]s s
		JOIN %[1]s.%[3]s ss ON s.id = ss.survey_id
		WHERE s.is_enabled = TRUE
		  AND s.type = 'special'
		  AND ss.user_email = @email
		  AND ss.response_id IS NULL
	`, s.datasetID, surveysTable, specialSurveysTable))
	specialQuery.Parameters = []bigquery.QueryParameter{{Name: "email", Value: userEmail}}

	specialIt, err := specialQuery.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list special surveys for user: %w", err)
	}

	for {
		var row specialSurveyRow
		err := specialIt.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate special survey list: %w", err)
		}

		su := row.Survey
		su.AssignmentID = row.AssignmentID
		su.PrefillData = map[string]string{
			"variable_1": row.Variable1,
			"variable_2": row.Variable2,
			"variable_3": row.Variable3,
			"variable_4": row.Variable4,
			"variable_5": row.Variable5,
		}
		result = append(result, su)
	}

	return result, nil
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

// --- Special Survey Methods ---

func (s *SurveyStore) AddSpecialSurveyUsers(ctx context.Context, users []survey.SpecialSurveyUser) error {
	if len(users) == 0 {
		return nil
	}
	// Use DML INSERT with UNNEST for batch insertion. This avoids issues with
	// mixing streaming inserts and other DML operations. It does not perform
	// upsert logic, so re-uploading a CSV will create duplicate assignments.
	q := s.client.Query(fmt.Sprintf(`
		INSERT INTO %s.%s (assignment_id, survey_id, user_email, variable_1, variable_2, variable_3, variable_4, variable_5, response_id)
		SELECT assignment_id, survey_id, user_email, variable_1, variable_2, variable_3, variable_4, variable_5, NULL
		FROM UNNEST(@users)
	`, s.datasetID, specialSurveysTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "users", Value: users},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run special survey user insert job: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for special survey user insert job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("special survey user insert job failed: %w", err)
	}
	return nil
}

func (s *SurveyStore) ListSpecialSurveyUsers(ctx context.Context, surveyID string) ([]survey.SpecialSurveyUser, error) {
	q := s.client.Query(fmt.Sprintf("SELECT * FROM `%s.%s` WHERE survey_id = @survey_id", s.datasetID, specialSurveysTable))
	q.Parameters = []bigquery.QueryParameter{{Name: "survey_id", Value: surveyID}}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list special survey users: %w", err)
	}

	var users []survey.SpecialSurveyUser
	for {
		var user survey.SpecialSurveyUser
		err := it.Next(&user)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate special survey user list: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *SurveyStore) GetSpecialSurveyAssignment(ctx context.Context, assignmentID string) (survey.SpecialSurveyUser, bool, error) {
	q := s.client.Query(fmt.Sprintf("SELECT * FROM `%s.%s` WHERE assignment_id = @assignment_id LIMIT 1", s.datasetID, specialSurveysTable))
	q.Parameters = []bigquery.QueryParameter{
		{Name: "assignment_id", Value: assignmentID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return survey.SpecialSurveyUser{}, false, fmt.Errorf("failed to read special survey assignment: %w", err)
	}

	var user survey.SpecialSurveyUser
	err = it.Next(&user)
	if err == iterator.Done {
		return survey.SpecialSurveyUser{}, false, nil // Not found
	}
	if err != nil {
		return survey.SpecialSurveyUser{}, false, fmt.Errorf("failed to iterate special survey assignment: %w", err)
	}

	return user, true, nil
}

func (s *SurveyStore) UpdateSpecialSurveyUserResponse(ctx context.Context, assignmentID, responseID string) error {
	q := s.client.Query(fmt.Sprintf(`
		UPDATE %s.%s
		SET response_id = @response_id
		WHERE assignment_id = @assignment_id
	`, s.datasetID, specialSurveysTable))
	q.Parameters = []bigquery.QueryParameter{
		{Name: "response_id", Value: responseID},
		{Name: "assignment_id", Value: assignmentID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run update special survey user job: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for update special survey user job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("update special survey user job failed: %w", err)
	}
	return nil
}

func (s *SurveyStore) GetSpecialSurveyUserCount(ctx context.Context, surveyID string) (int, error) {
	q := s.client.Query(fmt.Sprintf("SELECT COUNT(assignment_id) FROM `%s.%s` WHERE survey_id = @survey_id", s.datasetID, specialSurveysTable))
	q.Parameters = []bigquery.QueryParameter{
		{Name: "survey_id", Value: surveyID},
	}
	it, err := q.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to query special survey user count: %w", err)
	}
	var row []bigquery.Value
	err = it.Next(&row)
	if err == iterator.Done {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read special survey user count: %w", err)
	}
	if len(row) > 0 {
		if count, ok := row[0].(int64); ok {
			return int(count), nil
		}
	}
	return 0, nil
}
