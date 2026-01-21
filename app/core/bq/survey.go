package bq

import (
	"context"
	"fmt"
	"log"
	"maoni/app/core/collection"
	"maoni/app/core/survey"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	surveysTable        = "surveys"
	responsesTable      = "responses"
	specialSurveysTable = "special_surveys"
	cacheDuration       = 5 * time.Minute // Cache surveys for 5 minutes
)

// surveyCacheEntry holds a survey and its expiration time for caching.
type surveyCacheEntry struct {
	survey    survey.Survey
	expiresAt time.Time
}

type SurveyStore struct {
	client    *bigquery.Client
	fs        *firestore.Client // Firestore client for survey definitions
	datasetID string

	// Simple in-memory cache for survey definitions
	mu                    sync.RWMutex
	surveyCache           map[string]surveyCacheEntry // map[surveyID]surveyCacheEntry
	listCache             []survey.Survey
	listCacheExpiresAt    time.Time
	listCacheAll          []survey.Survey
	listCacheAllExpiresAt time.Time
}

func NewSurveyStore(client *bigquery.Client, fs *firestore.Client, datasetID string) *SurveyStore {
	return &SurveyStore{
		client:      client,
		fs:          fs,
		datasetID:   datasetID,
		surveyCache: make(map[string]surveyCacheEntry),
	}
}

// invalidateCache clears the cache for a specific survey and any list caches.
func (s *SurveyStore) invalidateCache(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.surveyCache, id)
	s.listCache = nil
	s.listCacheExpiresAt = time.Time{}
	s.listCacheAll = nil
	s.listCacheAllExpiresAt = time.Time{}
}

func (s *SurveyStore) Create(ctx context.Context, su survey.Survey) error {
	// 1. Write to Firestore (primary source of truth for reads)
	_, err := s.fs.Collection(collection.Surveys).Doc(su.ID).Set(ctx, su)
	if err != nil {
		return fmt.Errorf("failed to create survey in firestore: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(su.ID)

	// 2. Write a copy to BigQuery for analytics/backup.
	// This is done asynchronously in a goroutine to not block the request.
	go func() {
		// Create a new context for the background task
		bgCtx := context.Background()
		inserter := s.client.Dataset(s.datasetID).Table(surveysTable).Inserter()
		if err := inserter.Put(bgCtx, &su); err != nil {
			log.Printf("WARNING: failed to insert survey into BigQuery (firestore succeeded): %v", err)
		}
	}()

	return nil
}

func (s *SurveyStore) Update(ctx context.Context, su survey.Survey) error {
	// 1. Update in Firestore
	_, err := s.fs.Collection(collection.Surveys).Doc(su.ID).Set(ctx, su)
	if err != nil {
		return fmt.Errorf("failed to update survey in firestore: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(su.ID)

	// 2. Update copy in BigQuery asynchronously.
	go func() {
		bgCtx := context.Background()
		q := s.client.Query(fmt.Sprintf(`
			UPDATE %s.%s
			SET name = @name, description = @description, type = @type, is_enabled = @is_enabled, allow_multiple_submissions = @allow_multiple_submissions, updated_at = @updated_at, questions = @questions, group_headings = @group_headings, banner = @banner
			WHERE id = @id
		`, s.datasetID, surveysTable))

		// Need to convert []survey.Question to something BQ client understands for query params
		var bqQuestions []bigquery.Value
		for _, q := range su.Questions {
			bqQuestions = append(bqQuestions, map[string]bigquery.Value{
				"id":               q.ID,
				"text":             q.Text,
				"type":             q.Type,
				"options":          q.Options,
				"is_required":      q.IsRequired,
				"group_number":     q.GroupNumber,
				"prefill_variable": q.PrefillVariable,
			})
		}

		q.Parameters = []bigquery.QueryParameter{
			{Name: "id", Value: su.ID},
			{Name: "name", Value: su.Name},
			{Name: "description", Value: su.Description},
			{Name: "type", Value: su.Type},
			{Name: "is_enabled", Value: su.IsEnabled},
			{Name: "allow_multiple_submissions", Value: su.AllowMultipleSubmissions},
			{Name: "updated_at", Value: su.UpdatedAt},
			{Name: "questions", Value: bqQuestions},
			{Name: "group_headings", Value: su.GroupHeadings},
			{Name: "banner", Value: su.Banner},
		}

		job, err := q.Run(bgCtx)
		if err != nil {
			log.Printf("WARNING: failed to run BQ update job (firestore succeeded): %v", err)
			return
		}
		status, err := job.Wait(bgCtx)
		if err != nil {
			log.Printf("WARNING: failed to wait for BQ update job: %v", err)
			return
		}
		if err := status.Err(); err != nil {
			log.Printf("WARNING: BQ update job failed: %v", err)
		}
	}()

	return nil
}

func (s *SurveyStore) Get(ctx context.Context, id string) (survey.Survey, error) {
	// 1. Check cache
	s.mu.RLock()
	entry, found := s.surveyCache[id]
	s.mu.RUnlock()
	if found && time.Now().Before(entry.expiresAt) {
		return entry.survey, nil
	}

	// 2. Fetch from Firestore
	doc, err := s.fs.Collection(collection.Surveys).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return survey.Survey{}, fmt.Errorf("survey not found")
		}
		return survey.Survey{}, fmt.Errorf("failed to get survey from firestore: %w", err)
	}

	var su survey.Survey
	if err := doc.DataTo(&su); err != nil {
		return survey.Survey{}, fmt.Errorf("failed to decode survey data: %w", err)
	}

	// 3. Update cache
	s.mu.Lock()
	s.surveyCache[id] = surveyCacheEntry{
		survey:    su,
		expiresAt: time.Now().Add(cacheDuration),
	}
	s.mu.Unlock()

	return su, nil
}

// addResponseCounts is a helper to decouple BQ count queries from Firestore reads
func (s *SurveyStore) addResponseCounts(ctx context.Context, surveys []survey.Survey) ([]survey.Survey, error) {
	// Create a copy to avoid modifying the cached slice
	surveysCopy := make([]survey.Survey, len(surveys))
	copy(surveysCopy, surveys)

	for i := range surveysCopy {
		count, err := s.GetResponseCount(ctx, surveysCopy[i].ID)
		if err != nil {
			// Log the error but don't fail the whole list
			log.Printf("could not get response count for survey %s: %v", surveysCopy[i].ID, err)
		}
		surveysCopy[i].ResponseCount = count
	}
	return surveysCopy, nil
}

func (s *SurveyStore) List(ctx context.Context, showInactive bool) ([]survey.Survey, error) {
	// Check list cache first
	s.mu.RLock()
	if showInactive && s.listCacheAll != nil && time.Now().Before(s.listCacheAllExpiresAt) {
		cachedList := s.listCacheAll
		s.mu.RUnlock()
		return s.addResponseCounts(ctx, cachedList)
	}
	if !showInactive && s.listCache != nil && time.Now().Before(s.listCacheExpiresAt) {
		cachedList := s.listCache
		s.mu.RUnlock()
		return s.addResponseCounts(ctx, cachedList)
	}
	s.mu.RUnlock()

	// Fetch from Firestore
	iter := s.fs.Collection(collection.Surveys).OrderBy("created_at", firestore.Desc).Documents(ctx)
	var allSurveys []survey.Survey
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate survey list from firestore: %w", err)
		}
		var su survey.Survey
		if err := doc.DataTo(&su); err != nil {
			return nil, fmt.Errorf("failed to decode survey from firestore: %w", err)
		}
		allSurveys = append(allSurveys, su)
	}

	var filteredSurveys []survey.Survey
	if showInactive {
		filteredSurveys = allSurveys
	} else {
		for _, su := range allSurveys {
			if su.IsEnabled {
				filteredSurveys = append(filteredSurveys, su)
			}
		}
	}

	// Update cache
	s.mu.Lock()
	if showInactive {
		s.listCacheAll = make([]survey.Survey, len(allSurveys))
		copy(s.listCacheAll, allSurveys)
		s.listCacheAllExpiresAt = time.Now().Add(cacheDuration)
	} else {
		// Cache only the enabled surveys if that's what was requested
		// to serve future identical requests faster.
		s.listCache = make([]survey.Survey, len(filteredSurveys))
		copy(s.listCache, filteredSurveys)
		s.listCacheExpiresAt = time.Now().Add(cacheDuration)
	}
	s.mu.Unlock()

	return s.addResponseCounts(ctx, filteredSurveys)
}

// ListForUser retrieves all surveys a user is eligible to take.
// This includes all enabled "normal" surveys and any "special" surveys
// they have been assigned to and have not yet completed.
func (s *SurveyStore) ListForUser(ctx context.Context, userEmail string) ([]survey.Survey, error) {
	var result []survey.Survey

	// 1. Get all enabled surveys from Firestore (uses cache).
	allEnabledSurveys, err := s.List(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list active surveys: %w", err)
	}

	for _, su := range allEnabledSurveys {
		if su.Type == survey.TypeNormal || su.Type == "" {
			result = append(result, su)
		}
	}

	// 2. Get all special survey assignments for the user that are not yet submitted from Firestore.
	// This query requires a composite index on (user_email, response_id). Firestore will provide a link to create it if it doesn't exist.
	iter := s.fs.Collection(collection.SpecialSurveyAssignments).
		Where("user_email", "==", userEmail).
		Where("response_id", "==", "").
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate special survey assignments: %w", err)
		}
		var assignment survey.SpecialSurveyUser
		if err := doc.DataTo(&assignment); err != nil {
			log.Printf("WARNING: failed to decode special survey assignment: %v", err)
			continue
		}

		// 3. For each assignment, get the survey definition from Firestore (uses cache).
		su, err := s.Get(ctx, assignment.SurveyID)
		if err != nil {
			log.Printf("WARNING: could not get special survey %s for user %s: %v", assignment.SurveyID, userEmail, err)
			continue // Skip if survey is not found or there's an error
		}

		// Only add if it's enabled.
		if su.IsEnabled {
			su.AssignmentID = assignment.AssignmentID
			su.PrefillData = map[string]string{
				"variable_1": assignment.Variable1,
				"variable_2": assignment.Variable2,
				"variable_3": assignment.Variable3,
				"variable_4": assignment.Variable4,
				"variable_5": assignment.Variable5,
			}
			result = append(result, su)
		}
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

	// 1. Asynchronously stream to BigQuery for analytics.
	go func() {
		bgCtx := context.Background()
		inserter := s.client.Dataset(s.datasetID).Table(specialSurveysTable).Inserter()

		type bqSpecialSurveyUser struct {
			AssignmentID string `bigquery:"assignment_id"`
			SurveyID     string `bigquery:"survey_id"`
			UserEmail    string `bigquery:"user_email"`
			Variable1    string `bigquery:"variable_1"`
			Variable2    string `bigquery:"variable_2"`
			Variable3    string `bigquery:"variable_3"`
			Variable4    string `bigquery:"variable_4"`
			Variable5    string `bigquery:"variable_5"`
		}
		var bqUsers []bqSpecialSurveyUser
		for _, u := range users {
			bqUsers = append(bqUsers, bqSpecialSurveyUser{
				AssignmentID: u.AssignmentID,
				SurveyID:     u.SurveyID,
				UserEmail:    u.UserEmail,
				Variable1:    u.Variable1,
				Variable2:    u.Variable2,
				Variable3:    u.Variable3,
				Variable4:    u.Variable4,
				Variable5:    u.Variable5,
			})
		}

		if err := inserter.Put(bgCtx, bqUsers); err != nil {
			log.Printf("WARNING: failed to insert special survey users into BigQuery: %v", err)
		}
	}()

	// 2. Write to Firestore for real-time state tracking.
	bw := s.fs.BulkWriter(ctx)
	for i := range users {
		docRef := s.fs.Collection(collection.SpecialSurveyAssignments).Doc(users[i].AssignmentID)
		if _, err := bw.Set(docRef, &users[i]); err != nil {
			bw.Flush() // Attempt to flush pending writes before returning
			return fmt.Errorf("failed to add special survey user %s to firestore batch: %w", users[i].UserEmail, err)
		}
	}
	bw.Flush()

	return nil
}

func (s *SurveyStore) ListSpecialSurveyUsers(ctx context.Context, surveyID string) ([]survey.SpecialSurveyUser, error) {
	iter := s.fs.Collection(collection.SpecialSurveyAssignments).Where("survey_id", "==", surveyID).Documents(ctx)
	var users []survey.SpecialSurveyUser
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate special survey user list from firestore: %w", err)
		}
		var user survey.SpecialSurveyUser
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to decode special survey user from firestore: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *SurveyStore) GetSpecialSurveyAssignment(ctx context.Context, assignmentID string) (survey.SpecialSurveyUser, bool, error) {
	doc, err := s.fs.Collection(collection.SpecialSurveyAssignments).Doc(assignmentID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return survey.SpecialSurveyUser{}, false, nil
		}
		return survey.SpecialSurveyUser{}, false, fmt.Errorf("failed to get special survey assignment from firestore: %w", err)
	}

	var user survey.SpecialSurveyUser
	if err := doc.DataTo(&user); err != nil {
		return survey.SpecialSurveyUser{}, false, fmt.Errorf("failed to decode special survey assignment from firestore: %w", err)
	}

	return user, true, nil
}

func (s *SurveyStore) UpdateSpecialSurveyUserResponse(ctx context.Context, assignmentID, responseID string) error {
	docRef := s.fs.Collection(collection.SpecialSurveyAssignments).Doc(assignmentID)
	_, err := docRef.Update(ctx, []firestore.Update{
		{Path: "response_id", Value: responseID},
	})
	if err != nil {
		return fmt.Errorf("failed to update special survey assignment in firestore: %w", err)
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
