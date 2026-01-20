package base

import (
	"fmt"
	"maoni/app/core/auth"
	"maoni/app/core/survey"
	"maoni/app/core/web"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (m module) surveyLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User

	surveys, err := m.surveyStore.List(ctx, false) // Only show active surveys
	if err != nil {
		return fmt.Errorf("loading surveys: %w", err)
	}

	return surveyListPage(user, surveys).Render(ctx, w)
}

func (m module) viewSurveyHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User
	surveyID := chi.URLParam(r, "id")

	s, err := m.surveyStore.Get(ctx, surveyID)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found: %w", err), http.StatusNotFound)
	}

	if !s.IsEnabled {
		return web.NewRequestError(fmt.Errorf("survey is not active"), http.StatusForbidden)
	}

	if !s.AllowMultipleSubmissions {
		hasResponded, err := m.surveyStore.HasUserResponded(ctx, surveyID, user.ID)
		if err != nil {
			return fmt.Errorf("checking user response status: %w", err)
		}
		if hasResponded {
			// You might want a nicer "already completed" page
			return surveyAlreadyCompletedPage(user).Render(ctx, w)
		}
	}

	return takeSurveyPage(user, s).Render(ctx, w)
}

func (m module) submitSurveyHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User
	surveyID := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	s, err := m.surveyStore.Get(ctx, surveyID)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found: %w", err), http.StatusNotFound)
	}

	if !s.IsEnabled {
		return web.NewRequestError(fmt.Errorf("survey is not active"), http.StatusForbidden)
	}

	if !s.AllowMultipleSubmissions {
		hasResponded, err := m.surveyStore.HasUserResponded(ctx, surveyID, user.ID)
		if err != nil {
			return fmt.Errorf("checking user response status: %w", err)
		}
		if hasResponded {
			return web.NewRequestError(fmt.Errorf("already responded"), http.StatusForbidden)
		}
	}

	var answers []survey.Answer
	for _, q := range s.Questions {
		values, ok := r.Form[q.ID]
		if q.IsRequired && (!ok || len(values) == 0 || values[0] == "") {
			return web.NewRequestError(fmt.Errorf("question '%s' is required", q.Text), http.StatusBadRequest)
		}
		if ok && len(values) > 0 {
			answers = append(answers, survey.Answer{
				QuestionID: q.ID,
				Values:     values,
			})
		}
	}

	response := survey.Response{
		ID:          uuid.NewString(),
		SurveyID:    surveyID,
		UserID:      user.ID,
		SubmittedAt: time.Now(),
		Answers:     answers,
	}

	if err := m.surveyStore.SaveResponse(ctx, response); err != nil {
		return fmt.Errorf("saving response: %w", err)
	}

	return surveyThankYouPage(user).Render(ctx, w)
}
