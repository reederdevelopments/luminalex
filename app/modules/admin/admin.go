package admin

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"maoni/app/core/auth"
	"maoni/app/core/mid"
	"maoni/app/core/survey"
	"maoni/app/core/web"
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
	app.Handle(http.MethodGet, "/admin/special", m.specialSurveysLoader, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/results", m.resultsLoader, adminMiddlewares...)

	app.Handle(http.MethodGet, "/admin/surveys/add", m.addSurveyForm, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys", m.createSurvey, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/surveys/{id}/edit", m.editSurveyForm, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/{id}", m.updateSurvey, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/{id}/toggle", m.toggleSurveyStatus, adminMiddlewares...)

	// Special Surveys User Management
	app.Handle(http.MethodGet, "/admin/special/{id}", m.manageSpecialSurveyUsersLoader, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/special/{id}/add", m.addSpecialSurveyUser, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/special/{id}/upload", m.uploadSpecialSurveyUsers, adminMiddlewares...)

	// Partials for HTMX
	app.Handle(http.MethodGet, "/admin/surveys/partials/question", m.questionPartial, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/surveys/partials/group-heading", m.groupHeadingPartial, adminMiddlewares...)
}

func (m module) adminRedirect(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, "/admin/surveys", http.StatusSeeOther)
	return web.ErrHandled
}

func (m module) surveysLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	showInactive := r.URL.Query().Get("show_inactive") == "true"

	allSurveys, err := m.surveyStore.List(r.Context(), showInactive)
	if err != nil {
		return fmt.Errorf("failed to list surveys: %w", err)
	}

	var normalSurveys []survey.Survey
	for _, s := range allSurveys {
		if s.Type == survey.TypeNormal || s.Type == "" {
			normalSurveys = append(normalSurveys, s)
		}
	}

	data := surveysPageData{
		Surveys:      normalSurveys,
		ShowInactive: showInactive,
	}
	return adminPage(r, user, "Surveys", data).Render(r.Context(), w)
}

func (m module) specialSurveysLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	showInactive := r.URL.Query().Get("show_inactive") == "true"

	allSurveys, err := m.surveyStore.List(r.Context(), showInactive)
	if err != nil {
		return fmt.Errorf("failed to list surveys: %w", err)
	}

	var specialSurveys []survey.Survey
	for _, s := range allSurveys {
		if s.Type == survey.TypeSpecial {
			specialSurveys = append(specialSurveys, s)
		}
	}

	data := specialSurveysPageData{
		Surveys:      specialSurveys,
		ShowInactive: showInactive,
	}
	return adminPage(r, user, "Special Surveys", data).Render(r.Context(), w)
}

func (m module) resultsLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	return adminPage(r, user, "Results", "results").Render(r.Context(), w)
}

func (m module) addSurveyForm(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	s := survey.Survey{
		ID:            uuid.NewString(),
		Type:          survey.TypeNormal,
		Questions:     []survey.Question{},
		GroupHeadings: []string{""}, // Start with one empty heading
	}
	return surveyFormPage(user, s, "Create New Survey", "/admin/surveys").Render(r.Context(), w)
}

func (m module) createSurvey(w http.ResponseWriter, r *http.Request) error {
	s, err := parseSurveyForm(r)
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	now := time.Now()
	s.ID = uuid.NewString()
	s.CreatedAt = now
	s.UpdatedAt = now
	if s.Type == "" {
		s.Type = survey.TypeNormal
	}
	for i := range s.Questions {
		if s.Questions[i].ID == "" {
			s.Questions[i].ID = uuid.NewString()
		}
	}

	if err := m.surveyStore.Create(r.Context(), s); err != nil {
		return fmt.Errorf("failed to create survey: %w", err)
	}

	if s.Type == survey.TypeSpecial {
		http.Redirect(w, r, "/admin/special", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/surveys", http.StatusSeeOther)
	}
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

	formAction := fmt.Sprintf("/admin/surveys/%s", id)
	return surveyFormPage(user, s, "Edit Survey", formAction).Render(r.Context(), w)
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
	s.UpdatedAt = time.Now()
	if s.Type == "" {
		s.Type = survey.TypeNormal
	}

	for i := range s.Questions {
		if s.Questions[i].ID == "" {
			s.Questions[i].ID = uuid.NewString()
		}
	}

	if err := m.surveyStore.Update(r.Context(), s); err != nil {
		return fmt.Errorf("failed to update survey: %w", err)
	}

	if s.Type == survey.TypeSpecial {
		http.Redirect(w, r, "/admin/special", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/surveys", http.StatusSeeOther)
	}
	return web.ErrHandled
}

func (m module) toggleSurveyStatus(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	s, err := m.surveyStore.Get(r.Context(), id)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found"), http.StatusNotFound)
	}

	s.IsEnabled = !s.IsEnabled
	s.UpdatedAt = time.Now()

	if err := m.surveyStore.Update(r.Context(), s); err != nil {
		return fmt.Errorf("failed to toggle survey status: %w", err)
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

func (m module) manageSpecialSurveyUsersLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	surveyID := chi.URLParam(r, "id")

	s, err := m.surveyStore.Get(r.Context(), surveyID)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found"), http.StatusNotFound)
	}
	if s.Type != survey.TypeSpecial {
		return web.NewRequestError(fmt.Errorf("not a special survey"), http.StatusBadRequest)
	}

	users, err := m.surveyStore.ListSpecialSurveyUsers(r.Context(), surveyID)
	if err != nil {
		return fmt.Errorf("failed to list special survey users: %w", err)
	}

	return specialSurveyDetailsPage(user, s, users).Render(r.Context(), w)
}

func (m module) addSpecialSurveyUser(w http.ResponseWriter, r *http.Request) error {
	surveyID := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	userEmail := r.FormValue("user_email")
	if userEmail == "" {
		return web.NewRequestError(fmt.Errorf("email is required"), http.StatusBadRequest)
	}

	user := survey.SpecialSurveyUser{
		AssignmentID: uuid.NewString(),
		SurveyID:     surveyID,
		UserEmail:    userEmail,
		Variable1:    r.FormValue("variable_1"),
		Variable2:    r.FormValue("variable_2"),
		Variable3:    r.FormValue("variable_3"),
		Variable4:    r.FormValue("variable_4"),
		Variable5:    r.FormValue("variable_5"),
	}

	if err := m.surveyStore.AddSpecialSurveyUsers(r.Context(), []survey.SpecialSurveyUser{user}); err != nil {
		return fmt.Errorf("failed to add special survey user: %w", err)
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return web.ErrHandled
}

func (m module) uploadSpecialSurveyUsers(w http.ResponseWriter, r *http.Request) error {
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

func adminTabs(r *http.Request) []Tab {
	tabs := []Tab{
		{
			Title:       "Surveys",
			Href:        "/admin/surveys",
			ActiveLinks: []string{"/admin/surveys"},
		},
		{
			Title:       "Special Surveys",
			Href:        "/admin/special",
			ActiveLinks: []string{"/admin/special"},
		},
		{
			Title:       "Results",
			Href:        "/admin/results",
			ActiveLinks: []string{"/admin/results"},
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

	log.Println("--- Post Form Values ---")
	for k, v := range r.PostForm {
		log.Printf("%s: %v\n", k, v)
	}
	log.Println("------------------------")

	var s survey.Survey
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	// This will decode top-level fields like Name and Description.
	// It might fail on Questions, but we'll parse them manually for robustness.
	if err := decoder.Decode(&s, r.PostForm); err != nil {
		log.Printf("Error decoding form with gorilla/schema (this may be ok if only questions failed): %v", err)
		// We don't return an error here, as we are handling questions manually.
	}

	// For debugging, let's see what gorilla/schema managed to do.
	initialDecodeJSON, _ := json.MarshalIndent(s, "", "  ")
	log.Printf("--- Decoded Survey by gorilla/schema (before manual question parsing) ---\n%s", string(initialDecodeJSON))

	// Manually set fields that might not be decoded correctly
	s.Type = r.FormValue("Type")
	if s.Type == "" {
		s.Type = survey.TypeNormal
	}

	// Because unchecked checkboxes don't appear in form data, gorilla/schema won't update
	// a field from true to false on an edit. We must manually handle them based on form value presence.
	s.IsEnabled = r.FormValue("IsEnabled") == "true"
	s.AllowMultipleSubmissions = r.FormValue("AllowMultipleSubmissions") == "true"

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

	finalJsonData, _ := json.MarshalIndent(s, "", "  ")
	log.Printf("--- Final Parsed Survey ---\n%s", string(finalJsonData))

	return s, nil
}
