package admin

import (
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
	app.Handle(http.MethodGet, "/admin/results", m.resultsLoader, adminMiddlewares...)

	app.Handle(http.MethodGet, "/admin/surveys/add", m.addSurveyForm, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys", m.createSurvey, adminMiddlewares...)
	app.Handle(http.MethodGet, "/admin/surveys/{id}/edit", m.editSurveyForm, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/{id}", m.updateSurvey, adminMiddlewares...)
	app.Handle(http.MethodPost, "/admin/surveys/{id}/toggle", m.toggleSurveyStatus, adminMiddlewares...)

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

	surveys, err := m.surveyStore.List(r.Context(), showInactive)
	if err != nil {
		return fmt.Errorf("failed to list surveys: %w", err)
	}

	data := surveysPageData{
		Surveys:      surveys,
		ShowInactive: showInactive,
	}
	return adminPage(r, user, "Surveys", data).Render(r.Context(), w)
}

func (m module) resultsLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	return adminPage(r, user, "Results", "results").Render(r.Context(), w)
}

func (m module) addSurveyForm(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	s := survey.Survey{
		ID:            uuid.NewString(),
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
			ID:          r.FormValue(fmt.Sprintf("questions[%d].ID", i)),
			Text:        r.FormValue(textKey),
			Type:        r.FormValue(fmt.Sprintf("questions[%d].Type", i)),
			Options:     r.PostForm[fmt.Sprintf("questions[%d].Options", i)],
			IsRequired:  r.FormValue(fmt.Sprintf("questions[%d].IsRequired", i)) == "true",
			GroupNumber: groupNumber,
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
