package admin

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maoni/app/core/auth"
	"maoni/app/core/mid"
	"maoni/app/core/survey"
	"maoni/app/core/web"
	"maoni/app/modules/base"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
)

type module struct {
	l            *log.Logger
	sessionStore auth.Store
	surveyStore  survey.Store
}

func stdMid(l *log.Logger, additionalMid ...web.Middleware) []web.Middleware {
	middlewares := []web.Middleware{
		mid.Log(l),
		mid.CatchErr(l),
		mid.CatchPanic(),
	}
	middlewares = append(middlewares, additionalMid...)
	return middlewares
}

// adminOnly is a middleware that ensures the user is an admin.
func (m module) adminOnly(h web.Handler) web.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		user := auth.FromCtx(r.Context()).User
		if !user.IsAdmin {
			http.Redirect(w, r, "/", http.StatusForbidden)
			return web.ErrHandled
		}
		return h(w, r)
	}
}

func InitModule(l *log.Logger, app *web.App, sessionStore auth.Store, surveyStore survey.Store) {
	m := module{
		l:            l,
		sessionStore: sessionStore,
		surveyStore:  surveyStore,
	}

	adminMiddlewares := stdMid(l, sessionStore.Mid, m.adminOnly)

	app.Handle(http.MethodGet, "/admin", m.adminRedirect, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/surveys", m.surveysLoader, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/results", m.resultsLoader, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/config", m.configLoader, adminMiddlewares...)

	app.Handle(http.MethodGet, "/admin/surveys/add", m.addSurveyForm, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys", m.createSurvey, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/surveys/{id}/edit", m.editSurveyForm, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/{id}", m.updateSurvey, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/{id}/toggle", m.toggleSurveyStatus, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/preview", m.previewSurvey, adminMiddlewares...)

	// Special Surveys User Management
	app.Handle(http.MethodGet, "/admin/surveys/{id}/users", m.manageSurveyUsersLoader, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/{id}/users/upload", m.uploadSurveyUsers, adminMiddlewares...)

	// Partials for HTMX
	app.Handle(http.MethodGet, "/admin/surveys/partials/question", m.questionPartial, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/surveys/partials/group-heading", m.groupHeadingPartial, adminMiddlewares...)

	// Category Management
	app.Handle(http.MethodGet, "/admin/config/categories-modal", m.categoriesModal, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/config/categories", m.createCategory, adminMiddlewares...)
	app.Handle(http.MethodDelete, "/admin/config/categories/{id}", m.deleteCategory, adminMiddlewares...)
}

func (m module) adminRedirect(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, "/admin/surveys", http.StatusSeeOther)
	return web.ErrHandled
}

func (m module) surveysLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	showInactive := r.URL.Query().Get("show_inactive") == "true"
	ctx := r.Context()

	allSurveys, err := m.surveyStore.List(ctx, showInactive)
	if err != nil {
		return fmt.Errorf("failed to list surveys: %w", err)
	}

	categories, err := m.surveyStore.ListCategories(ctx)
	if err != nil {
		return fmt.Errorf("failed to list categories: %w", err)
	}
	categoryMap := make(map[string]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}

	var surveyViews []surveyView
	for _, s := range allSurveys {
		surveyViews = append(surveyViews, surveyView{
			Survey:       s,
			CategoryName: categoryMap[s.CategoryID],
		})
	}

	data := surveysPageData{
		Surveys:      surveyViews,
		ShowInactive: showInactive,
	}
	return adminPage(r, user, "Surveys", data).Render(r.Context(), w)
}

func (m module) resultsLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	ctx := r.Context()

	// Filtering params
	nameFilter := r.URL.Query().Get("name")

	allSurveys, err := m.surveyStore.List(ctx, true) // Get all surveys
	if err != nil {
		return fmt.Errorf("failed to list surveys for results: %w", err)
	}

	// Fetch all counts in batch from BigQuery for efficiency
	responseCounts, err := m.surveyStore.GetAllResponseCounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all response counts: %w", err)
	}
	assignedCounts, err := m.surveyStore.GetAllAssignedUserCounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all assigned user counts: %w", err)
	}

	var filteredSurveys []survey.Survey
	for _, s := range allSurveys {
		// Apply name filter
		if nameFilter != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(nameFilter)) {
			continue
		}

		// Update counts from the maps
		if count, ok := responseCounts[s.ID]; ok {
			s.ResponseCount = count
		} else {
			s.ResponseCount = 0 // Default to 0 if not in map
		}
		if count, ok := assignedCounts[s.ID]; ok {
			s.AssignedUserCount = count
		} else {
			s.AssignedUserCount = 0 // Default to 0 if not in map
		}

		filteredSurveys = append(filteredSurveys, s)
	}

	data := resultsPageData{
		Surveys:    filteredSurveys,
		NameFilter: nameFilter,
	}
	return adminPage(r, user, "Results", data).Render(ctx, w)
}

func (m module) configLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	return adminPage(r, user, "Config", nil).Render(r.Context(), w)
}

func (m module) addSurveyForm(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	cats, err := m.surveyStore.ListCategories(r.Context())
	if err != nil {
		return fmt.Errorf("listing categories for survey form: %w", err)
	}
	s := survey.Survey{
		ID:            uuid.NewString(),
		Type:          survey.TypeSpecial,
		IsEnabled:     true,
		Questions:     []survey.Question{},
		GroupHeadings: []string{""}, // Start with one empty heading
	}
	return surveyFormPage(user, s, "Create New Survey", "/admin/surveys", cats).Render(r.Context(), w)
}

func (m module) createSurvey(w http.ResponseWriter, r *http.Request) error {
	s, err := parseSurveyForm(r)
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	now := web.Now()
	s.ID = uuid.NewString()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.Type = survey.TypeSpecial
	for i := range s.Questions {
		if s.Questions[i].ID == "" {
			s.Questions[i].ID = uuid.NewString()
		}
	}

	if err := m.surveyStore.Create(r.Context(), s); err != nil {
		return fmt.Errorf("failed to create survey: %w", err)
	}

	http.Redirect(w, r, "/admin/surveys", http.StatusSeeOther)
	return web.ErrHandled
}

func (m module) editSurveyForm(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	id := chi.URLParam(r, "id")
	s, err := m.surveyStore.Get(r.Context(), id)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found"), http.StatusNotFound)
	}
	if len(s.GroupHeadings) == 0 {
		s.GroupHeadings = []string{""}
	}

	cats, err := m.surveyStore.ListCategories(r.Context())
	if err != nil {
		return fmt.Errorf("listing categories for survey form: %w", err)
	}

	formAction := fmt.Sprintf("/admin/surveys/%s", id)
	return surveyFormPage(user, s, "Edit Survey", formAction, cats).Render(r.Context(), w)
}

func (m module) updateSurvey(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	s, err := parseSurveyForm(r)
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	// Preserve existing IDs for questions if they exist, generate for new ones
	existingSurvey, err := m.surveyStore.Get(r.Context(), id)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found"), http.StatusNotFound)
	}
	s.CreatedAt = existingSurvey.CreatedAt
	s.ID = id
	s.UpdatedAt = web.Now()
	s.Type = survey.TypeSpecial

	for i := range s.Questions {
		if s.Questions[i].ID == "" {
			s.Questions[i].ID = uuid.NewString()
		}
	}

	if err := m.surveyStore.Update(r.Context(), s); err != nil {
		return fmt.Errorf("failed to update survey: %w", err)
	}

	http.Redirect(w, r, "/admin/surveys", http.StatusSeeOther)
	return web.ErrHandled
}

func (m module) previewSurvey(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	s, err := parseSurveyForm(r)
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}
	// The survey is not saved, so it doesn't have a real ID.
	// We pass the parsed data directly to the survey page template.
	return base.TakeSurveyPage(user, s, nil).Render(r.Context(), w)
}

func (m module) toggleSurveyStatus(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	s, err := m.surveyStore.Get(ctx, id)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found"), http.StatusNotFound)
	}

	// Support simple POST toggle for existing form, and JSON for new JS client
	if r.Header.Get("Content-Type") == "application/json" {
		var payload struct {
			IsEnabled bool `json:"isEnabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			return web.NewRequestError(fmt.Errorf("invalid request body: %w", err), http.StatusBadRequest)
		}
		s.IsEnabled = payload.IsEnabled
	} else {
		// Fallback to simple toggle for form-based submissions
		s.IsEnabled = !s.IsEnabled
	}

	s.UpdatedAt = web.Now()

	if err := m.surveyStore.Update(ctx, s); err != nil {
		return fmt.Errorf("failed to toggle survey status: %w", err)
	}

	// For JS client, return OK. For form, redirect.
	if r.Header.Get("Content-Type") == "application/json" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return web.ErrHandled
}

func (m module) questionPartial(w http.ResponseWriter, r *http.Request) error {
	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid index"), http.StatusBadRequest)
	}

	q := survey.Question{} // Empty question
	return questionForm(q, index).Render(r.Context(), w)
}

func (m module) groupHeadingPartial(w http.ResponseWriter, r *http.Request) error {
	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid index"), http.StatusBadRequest)
	}

	return groupHeadingInput("", index).Render(r.Context(), w)
}

func (m module) manageSurveyUsersLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	surveyID := chi.URLParam(r, "id")

	s, err := m.surveyStore.Get(r.Context(), surveyID)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found"), http.StatusNotFound)
	}

	users, err := m.surveyStore.ListSpecialSurveyUsers(r.Context(), surveyID)
	if err != nil {
		return fmt.Errorf("failed to list special survey users: %w", err)
	}

	return specialSurveyDetailsPage(user, s, users).Render(r.Context(), w)
}

func (m module) uploadSurveyUsers(w http.ResponseWriter, r *http.Request) error {
	surveyID := chi.URLParam(r, "id")
	ctx := r.Context()

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		return web.NewRequestError(fmt.Errorf("could not parse multipart form: %w", err), http.StatusBadRequest)
	}

	file, _, err := r.FormFile("csv_file")
	if err != nil {
		return web.NewRequestError(fmt.Errorf("could not get csv_file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return web.NewRequestError(fmt.Errorf("could not read csv file: %w", err), http.StatusBadRequest)
	}

	if len(records) < 2 {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return web.ErrHandled
	}
	records = records[1:] // Skip header

	var users []survey.SpecialSurveyUser
	for i, record := range records {
		if len(record) == 0 || record[0] == "" {
			continue
		}
		user := survey.SpecialSurveyUser{
			AssignmentID: uuid.NewString(),
			SurveyID:     surveyID,
			UserEmail:    record[0],
		}
		if len(record) > 1 {
			user.Variable1 = record[1]
		}
		if len(record) > 2 {
			user.Variable2 = record[2]
		}
		if len(record) > 3 {
			user.Variable3 = record[3]
		}
		if len(record) > 4 {
			user.Variable4 = record[4]
		}
		if len(record) > 5 {
			user.Variable5 = record[5]
		}
		users = append(users, user)

		if (i+1)%500 == 0 { // Process in batches of 500
			if err := m.surveyStore.AddSpecialSurveyUsers(ctx, users); err != nil {
				return fmt.Errorf("failed to add special survey users batch: %w", err)
			}
			users = []survey.SpecialSurveyUser{} // Reset batch
		}
	}
	// Process any remaining users
	if len(users) > 0 {
		if err := m.surveyStore.AddSpecialSurveyUsers(ctx, users); err != nil {
			return fmt.Errorf("failed to add final special survey users batch: %w", err)
		}
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return web.ErrHandled
}

func (m module) categoriesModal(w http.ResponseWriter, r *http.Request) error {
	cats, err := m.surveyStore.ListCategories(r.Context())
	if err != nil {
		return fmt.Errorf("listing categories: %w", err)
	}
	return categoriesModal(cats).Render(r.Context(), w)
}

func (m module) createCategory(w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	if name == "" {
		return web.NewRequestError(errors.New("category name is required"), http.StatusBadRequest)
	}

	if _, err := m.surveyStore.CreateCategory(r.Context(), name); err != nil {
		return fmt.Errorf("creating category: %w", err)
	}

	cats, err := m.surveyStore.ListCategories(r.Context())
	if err != nil {
		return fmt.Errorf("listing categories after create: %w", err)
	}

	return categoriesModalContent(cats).Render(r.Context(), w)
}

func (m module) deleteCategory(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	if err := m.surveyStore.DeleteCategory(r.Context(), id); err != nil {
		return fmt.Errorf("deleting category: %w", err)
	}

	cats, err := m.surveyStore.ListCategories(r.Context())
	if err != nil {
		return fmt.Errorf("listing categories after delete: %w", err)
	}
	return categoriesModalContent(cats).Render(r.Context(), w)
}

func adminTabs(r *http.Request) []Tab {
	tabs := []Tab{
		{
			Title:       "Surveys",
			Href:        "/admin/surveys",
			ActiveLinks: []string{"/admin/surveys"},
		},
		{
			Title:       "Results",
			Href:        "/admin/results",
			ActiveLinks: []string{"/admin/results"},
		},
		{
			Title:       "Config",
			Href:        "/admin/config",
			ActiveLinks: []string{"/admin/config"},
		},
	}

	for i, t := range tabs {
		// Use exact match for top-level tabs, prefix for sub-pages
		if r.URL.Path == t.Href || strings.HasPrefix(r.URL.Path, t.Href+"/") {
			tabs[i].Active = true
		}
	}
	return tabs
}

func parseSurveyForm(r *http.Request) (survey.Survey, error) {
	if err := r.ParseForm(); err != nil {
		return survey.Survey{}, err
	}

	var s survey.Survey
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	// This will decode top-level fields like Name and Description.
	// It might fail on Questions, but we'll parse them manually for robustness.
	if err := decoder.Decode(&s, r.PostForm); err != nil {
		// We can ignore this error for now as we manually parse questions which often causes a schema mismatch
	}

	// Manually set fields that might not be decoded correctly
	s.Type = r.FormValue("Type")
	if s.Type == "" {
		s.Type = survey.TypeSpecial
	}
	s.Banner = r.FormValue("Banner")

	// Because unchecked checkboxes don't appear in form data, gorilla/schema won't update
	// a field from true to false on an edit. We must manually handle them based on form value presence.
	s.IsEnabled = r.FormValue("IsEnabled") == "true"

	// gorilla/schema doesn't handle removal of all items from a slice field correctly.
	// Manually re-assign from the form post to ensure it's cleared if empty.
	s.GroupHeadings = r.PostForm["GroupHeadings"]

	var validQuestions []survey.Question
	// Manually iterate through form values to build the questions slice.
	for i := 0; ; i++ {
		textKey := fmt.Sprintf("questions[%d].Text", i)
		// If Text for index i doesn't exist, we assume there are no more questions.
		if _, ok := r.PostForm[textKey]; !ok {
			break
		}
		// If Text is submitted but empty, it's an empty question from the client that was likely deleted, so we skip it.
		if r.FormValue(textKey) == "" {
			continue
		}

		groupNumberStr := r.FormValue(fmt.Sprintf("questions[%d].GroupNumber", i))
		groupNumber, err := strconv.Atoi(groupNumberStr)
		if err != nil || groupNumber < 1 {
			groupNumber = 1 // Default to group 1
		}

		q := survey.Question{
			ID:              r.FormValue(fmt.Sprintf("questions[%d].ID", i)),
			Text:            r.FormValue(textKey),
			Type:            r.FormValue(fmt.Sprintf("questions[%d].Type", i)),
			Options:         r.PostForm[fmt.Sprintf("questions[%d].Options", i)],
			IsRequired:      r.FormValue(fmt.Sprintf("questions[%d].IsRequired", i)) == "true",
			GroupNumber:     groupNumber,
			PrefillVariable: r.FormValue(fmt.Sprintf("questions[%d].PrefillVariable", i)),
		}

		// Filter out empty options submitted by the form.
		var validOptions []string
		for _, opt := range q.Options {
			if opt != "" {
				validOptions = append(validOptions, opt)
			}
		}
		q.Options = validOptions

		validQuestions = append(validQuestions, q)
	}
	s.Questions = validQuestions

	// Parse date-time fields in GMT+2
	const layout = "2006-01-02T15:04"
	openStr := r.FormValue("SurveyOpen")
	if openStr != "" {
		t, err := time.ParseInLocation(layout, openStr, web.GMTPlus2)
		if err == nil {
			s.SurveyOpen = t
		}
	}
	closeStr := r.FormValue("SurveyClosed")
	if closeStr != "" {
		t, err := time.ParseInLocation(layout, closeStr, web.GMTPlus2)
		if err == nil {
			s.SurveyClosed = t
		}
	}

	return s, nil
}
