package bq

import (
	"context"
	"fmt"
	"log"
	"maoni/app/core/collection"
	"maoni/app/core/events"
	"maoni/app/core/survey"
	"maoni/app/core/web"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	surveysTable        = "surveys"
	responsesTable      = "responses"
	specialSurveysTable = "special_surveys"
	categoriesTable     = "categories"
	cacheDuration       = 5 * time.Minute // Cache surveys for 5 minutes
)

// surveyCacheEntry holds a survey and its expiration time for caching.
type surveyCacheEntry struct {
	survey    survey.Survey
	expiresAt time.Time
}

// SurveyProgress represents a user's saved progress for a survey.
type SurveyProgress struct {
	UserID       string         `firestore:"user_id"`
	SurveyID     string         `firestore:"survey_id"`
	AssignmentID string         `firestore:"assignment_id"`
	Answers      map[string]any `firestore:"answers"`
	UpdatedAt    time.Time      `firestore:"updated_at"`
}

type SurveyStore struct {
	client    *bigquery.Client
	fs        *firestore.Client // Firestore client for survey definitions
	datasetID string
	broker    events.Publisher

	// Simple in-memory cache for survey definitions
	mu                    sync.RWMutex
	surveyCache           map[string]surveyCacheEntry // map[surveyID]surveyCacheEntry
	listCache             []survey.Survey
	listCacheExpiresAt    time.Time
	listCacheAll          []survey.Survey
	listCacheAllExpiresAt time.Time
}

func NewSurveyStore(client *bigquery.Client, fs *firestore.Client, datasetID string, broker events.Publisher) *SurveyStore {
	return &SurveyStore{
		client:      client,
		fs:          fs,
		datasetID:   datasetID,
		broker:      broker,
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
			SET name = @name, description = @description, type = @type, is_enabled = @is_enabled, updated_at = @updated_at, questions = @questions, group_headings = @group_headings, banner = @banner, category_id = @category_id, survey_open = @survey_open, survey_closed = @survey_closed
			WHERE id = @id
		`, s.datasetID, surveysTable))

		// The BigQuery client can use the struct slice directly as it has `bigquery` tags.
		q.Parameters = []bigquery.QueryParameter{
			{Name: "id", Value: su.ID},
			{Name: "name", Value: su.Name},
			{Name: "description", Value: su.Description},
			{Name: "type", Value: su.Type},
			{Name: "is_enabled", Value: su.IsEnabled},
			{Name: "updated_at", Value: su.UpdatedAt},
			{Name: "questions", Value: su.Questions},
			{Name: "group_headings", Value: su.GroupHeadings},
			{Name: "banner", Value: su.Banner},
			{Name: "category_id", Value: su.CategoryID},
			{Name: "survey_open", Value: su.SurveyOpen},
			{Name: "survey_closed", Value: su.SurveyClosed},
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

func (s *SurveyStore) List(ctx context.Context, showInactive bool) ([]survey.Survey, error) {
	// Check list cache first
	s.mu.RLock()
	if showInactive && s.listCacheAll != nil && time.Now().Before(s.listCacheAllExpiresAt) {
		surveysCopy := make([]survey.Survey, len(s.listCacheAll))
		copy(surveysCopy, s.listCacheAll)
		s.mu.RUnlock()
		return surveysCopy, nil
	}
	if !showInactive && s.listCache != nil && time.Now().Before(s.listCacheExpiresAt) {
		surveysCopy := make([]survey.Survey, len(s.listCache))
		copy(surveysCopy, s.listCache)
		s.mu.RUnlock()
		return surveysCopy, nil
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

	return filteredSurveys, nil
}

// ListForUser retrieves all surveys a user is eligible to take.
// This includes all "special" surveys they have been assigned to and have not yet completed.
func (s *SurveyStore) ListForUser(ctx context.Context, userEmail string) ([]survey.Survey, error) {
	var result []survey.Survey

	// Get all special survey assignments for the user that are not yet submitted from Firestore.
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

		// For each assignment, get the survey definition from Firestore (uses cache).
		su, err := s.Get(ctx, assignment.SurveyID)
		if err != nil {
			log.Printf("WARNING: could not get special survey %s for user %s: %v", assignment.SurveyID, userEmail, err)
			continue // Skip if survey is not found or there's an error
		}

		now := web.Now()
		isActive := su.IsEnabled
		if !su.SurveyOpen.IsZero() && now.Before(su.SurveyOpen) {
			isActive = false
		}
		if !su.SurveyClosed.IsZero() && now.After(su.SurveyClosed) {
			isActive = false
		}

		// Only add if it's effectively active.
		if isActive {
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
	// Asynchronously write a copy to BigQuery for analytics/backup.
	go func() {
		// Use a background context because the original request's context might be cancelled
		bgCtx := context.Background()
		inserter := s.client.Dataset(s.datasetID).Table(responsesTable).Inserter()
		if err := inserter.Put(bgCtx, r); err != nil {
			log.Printf("WARNING: failed to insert response %s into BigQuery: %v", r.ID, err)
		}
	}()

	// Increment the response count in Firestore.
	surveyRef := s.fs.Collection(collection.Surveys).Doc(r.SurveyID)
	_, err := surveyRef.Update(ctx, []firestore.Update{
		{Path: "response_count", Value: firestore.Increment(1)},
	})
	if err != nil {
		// The response is in BQ but the count is not updated. This is a state inconsistency.
		// Log a warning. A more robust system might have a background job to reconcile counts.
		log.Printf("WARNING: failed to increment response count for survey %s: %v", r.SurveyID, err)
	}
	s.invalidateCache(r.SurveyID) // Invalidate cache to reflect new count

	return nil
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

	// Increment the assigned user count in Firestore.
	surveyID := users[0].SurveyID
	surveyRef := s.fs.Collection(collection.Surveys).Doc(surveyID)
	_, err := surveyRef.Update(ctx, []firestore.Update{
		{Path: "assigned_user_count", Value: firestore.Increment(len(users))},
	})
	if err != nil {
		// Log and continue, as writing the user assignments is the primary goal.
		log.Printf("WARNING: failed to increment assigned user count for survey %s: %v", surveyID, err)
	}
	s.invalidateCache(surveyID)

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

// --- Category Methods ---

func (s *SurveyStore) ListCategories(ctx context.Context) ([]survey.Category, error) {
	iter := s.fs.Collection(collection.Categories).OrderBy("name", firestore.Asc).Documents(ctx)
	var categories []survey.Category
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate categories: %w", err)
		}
		var cat survey.Category
		if err := doc.DataTo(&cat); err != nil {
			return nil, fmt.Errorf("failed to decode category: %w", err)
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (s *SurveyStore) CreateCategory(ctx context.Context, name string) (survey.Category, error) {
	id := uuid.NewString()
	cat := survey.Category{
		ID:   id,
		Name: name,
	}
	// 1. Write to Firestore
	_, err := s.fs.Collection(collection.Categories).Doc(id).Set(ctx, cat)
	if err != nil {
		return survey.Category{}, fmt.Errorf("failed to create category in firestore: %w", err)
	}

	// 2. Write a copy to BigQuery for analytics/backup.
	go func() {
		// Create a new context for the background task
		bgCtx := context.Background()
		inserter := s.client.Dataset(s.datasetID).Table(categoriesTable).Inserter()
		if err := inserter.Put(bgCtx, &cat); err != nil {
			log.Printf("WARNING: failed to insert category into BigQuery (firestore succeeded): %v", err)
		}
	}()

	return cat, nil
}

func (s *SurveyStore) DeleteCategory(ctx context.Context, id string) error {
	// 1. Delete from Firestore
	_, err := s.fs.Collection(collection.Categories).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete category from firestore: %w", err)
	}

	// 2. Asynchronously delete from BigQuery.
	go func() {
		bgCtx := context.Background()
		q := s.client.Query(fmt.Sprintf(`
			DELETE FROM %s.%s
			WHERE id = @id
		`, s.datasetID, categoriesTable))
		q.Parameters = []bigquery.QueryParameter{
			{Name: "id", Value: id},
		}

		job, err := q.Run(bgCtx)
		if err != nil {
			log.Printf("WARNING: failed to run BQ delete job for category (firestore succeeded): %v", err)
			return
		}
		status, err := job.Wait(bgCtx)
		if err != nil {
			log.Printf("WARNING: failed to wait for BQ delete job for category: %v", err)
			return
		}
		if err := status.Err(); err != nil {
			log.Printf("WARNING: BQ delete job for category failed: %v", err)
		}
	}()

	return nil
}

func (s *SurveyStore) GetAllResponseCounts(ctx context.Context) (map[string]int, error) {
	counts := make(map[string]int)
	q := s.client.Query(fmt.Sprintf("SELECT survey_id, COUNT(id) FROM `%s.%s` GROUP BY survey_id", s.datasetID, responsesTable))
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query response counts: %w", err)
	}

	type countRow struct {
		SurveyID string `bigquery:"survey_id"`
		Count    int64  `bigquery:"f0_"`
	}

	for {
		var row countRow
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read response count row: %w", err)
		}
		counts[row.SurveyID] = int(row.Count)
	}
	return counts, nil
}

func (s *SurveyStore) GetAllAssignedUserCounts(ctx context.Context) (map[string]int, error) {
	counts := make(map[string]int)
	q := s.client.Query(fmt.Sprintf("SELECT survey_id, COUNT(assignment_id) FROM `%s.%s` GROUP BY survey_id", s.datasetID, specialSurveysTable))
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query assigned user counts: %w", err)
	}

	type countRow struct {
		SurveyID string `bigquery:"survey_id"`
		Count    int64  `bigquery:"f0_"`
	}

	for {
		var row countRow
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read assigned user count row: %w", err)
		}
		counts[row.SurveyID] = int(row.Count)
	}
	return counts, nil
}

func (s *SurveyStore) CheckAndManageSurveyStatus(ctx context.Context) error {
	log.Println("Running scheduled job: CheckAndManageSurveyStatus")
	now := web.Now()
	wasUpdated := false

	iter := s.fs.Collection(collection.Surveys).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate surveys for status check: %w", err)
		}

		var su survey.Survey
		if err := doc.DataTo(&su); err != nil {
			log.Printf("WARNING: could not decode survey %s for status check: %v", doc.Ref.ID, err)
			continue
		}

		// Only manage status automatically if at least one date is set.
		// If no dates are set, the IsEnabled flag is considered manually controlled.
		if su.SurveyOpen.IsZero() && su.SurveyClosed.IsZero() {
			continue
		}

		// Determine the desired state based on the schedule.
		shouldBeEnabled := true
		if !su.SurveyOpen.IsZero() && now.Before(su.SurveyOpen) {
			shouldBeEnabled = false
		}
		if !su.SurveyClosed.IsZero() && now.After(su.SurveyClosed) {
			shouldBeEnabled = false
		}

		// If the current state doesn't match the desired state, update it.
		if su.IsEnabled != shouldBeEnabled {
			wasUpdated = true
			log.Printf("Survey '%s' (%s) is currently IsEnabled=%v, but schedule requires IsEnabled=%v. Updating.", su.Name, su.ID, su.IsEnabled, shouldBeEnabled)
			su.IsEnabled = shouldBeEnabled
			su.UpdatedAt = now

			// Update both Firestore and BigQuery
			if err := s.Update(ctx, su); err != nil {
				log.Printf("ERROR: failed to update survey status for %s: %v", su.ID, err)
				// Continue to the next one
			} else {
				log.Printf("Successfully updated survey '%s' status to IsEnabled=%v.", su.Name, su.IsEnabled)
			}
		}
	}

	if wasUpdated {
		s.broker.Publish("surveys-updated")
		log.Println("Published surveys-updated event because survey statuses changed.")
	}

	log.Println("Finished scheduled job: CheckAndManageSurveyStatus")
	return nil
}

// --- Survey Progress ---

func (s *SurveyStore) SaveProgress(ctx context.Context, userID, surveyID, assignmentID string, answers map[string]any) error {
	progress := SurveyProgress{
		UserID:       userID,
		SurveyID:     surveyID,
		AssignmentID: assignmentID,
		Answers:      answers,
		UpdatedAt:    web.Now(),
	}

	_, err := s.fs.Collection(collection.SurveyProgress).Doc(assignmentID).Set(ctx, progress)
	if err != nil {
		return fmt.Errorf("failed to save progress to firestore: %w", err)
	}
	return nil
}

func (s *SurveyStore) GetProgress(ctx context.Context, assignmentID string) (map[string]any, error) {
	doc, err := s.fs.Collection(collection.SurveyProgress).Doc(assignmentID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil // No progress saved yet, which is not an error.
		}
		return nil, fmt.Errorf("failed to get progress from firestore: %w", err)
	}

	var progress SurveyProgress
	if err := doc.DataTo(&progress); err != nil {
		return nil, fmt.Errorf("failed to decode progress data: %w", err)
	}

	return progress.Answers, nil
}

func (s *SurveyStore) DeleteProgress(ctx context.Context, assignmentID string) error {
	_, err := s.fs.Collection(collection.SurveyProgress).Doc(assignmentID).Delete(ctx)
	if err != nil && status.Code(err) != codes.NotFound {
		return fmt.Errorf("failed to delete progress from firestore: %w", err)
	}
	return nil
}
