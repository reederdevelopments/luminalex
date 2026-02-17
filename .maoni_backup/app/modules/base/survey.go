package base

import (
	"fmt"
	"maoni/app/core/auth"
	"maoni/app/core/survey"
	"maoni/app/core/web"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

/*// hasQuestionsInGroup checks if there are any questions for a given group index.
func hasQuestionsInGroup(questions []survey.Question, groupIndex int) bool {
	for _, q := range questions {
		if q.GroupNumber == groupIndex+1 {
			return true
		}
	}
	return false
}*/

func (m module) surveyLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User

	surveys, err := m.surveyStore.ListForUser(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("loading surveys for user: %w", err)
	}

	return surveyListPage(user, surveys).Render(ctx, w)
}

func (m module) viewSurveyHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User
	surveyID := chi.URLParam(r, "id")
	assignmentID := r.URL.Query().Get("assignment_id")

	s, err := m.surveyStore.Get(ctx, surveyID)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found: %w", err), http.StatusNotFound)
	}

	now := web.Now()
	if !s.IsEnabled {
		return web.NewRequestError(fmt.Errorf("survey is not active"), http.StatusForbidden)
	}
	if !s.SurveyOpen.IsZero() && now.Before(s.SurveyOpen) {
		return web.NewRequestError(fmt.Errorf("survey is not yet open"), http.StatusForbidden)
	}
	if !s.SurveyClosed.IsZero() && now.After(s.SurveyClosed) {
		return web.NewRequestError(fmt.Errorf("survey has closed"), http.StatusForbidden)
	}

	prefills := make(map[string]string)
	if assignmentID == "" {
		return web.NewRequestError(fmt.Errorf("assignment ID is required for this survey"), http.StatusBadRequest)
	}
	specialData, found, err := m.surveyStore.GetSpecialSurveyAssignment(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("checking special survey status for user: %w", err)
	}
	if !found || specialData.UserEmail != user.Email {
		return web.NewRequestError(fmt.Errorf("you are not assigned to this survey"), http.StatusForbidden)
	}
	if specialData.ResponseID != "" {
		return surveyAlreadyCompletedPage(user).Render(ctx, w)
	}

	s.AssignmentID = assignmentID // Pass assignment ID to the template
	prefills["variable_1"] = specialData.Variable1
	prefills["variable_2"] = specialData.Variable2
	prefills["variable_3"] = specialData.Variable3
	prefills["variable_4"] = specialData.Variable4
	prefills["variable_5"] = specialData.Variable5

	return TakeSurveyPage(user, s, prefills).Render(ctx, w)
}

func (m module) submitSurveyHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User
	surveyID := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}
	assignmentID := r.FormValue("assignment_id")

	s, err := m.surveyStore.Get(ctx, surveyID)
	if err != nil {
		return web.NewRequestError(fmt.Errorf("survey not found: %w", err), http.StatusNotFound)
	}

	now := web.Now()
	if !s.IsEnabled {
		return web.NewRequestError(fmt.Errorf("survey is not active"), http.StatusForbidden)
	}
	if !s.SurveyOpen.IsZero() && now.Before(s.SurveyOpen) {
		return web.NewRequestError(fmt.Errorf("survey is not yet open"), http.StatusForbidden)
	}
	if !s.SurveyClosed.IsZero() && now.After(s.SurveyClosed) {
		return web.NewRequestError(fmt.Errorf("survey has closed"), http.StatusForbidden)
	}

	// --- Validation & Response creation ---
	var answers []survey.Answer
	for _, q := range s.Questions {
		var values []string
		var err error

		switch survey.QuestionType(q.Type) {
		case survey.MultiGridRadio:
			allRowsAnswered := true
			for i, rowText := range q.Rows {
				formKey := fmt.Sprintf("%s_%d", q.ID, i)
				selectedValue := r.FormValue(formKey)
				if selectedValue != "" {
					values = append(values, fmt.Sprintf("%s:%s", rowText, selectedValue))
				} else {
					allRowsAnswered = false
				}
			}
			if q.IsRequired && !allRowsAnswered {
				err = web.NewRequestError(fmt.Errorf("all rows are required for question: '%s'", q.Text), http.StatusBadRequest)
			}
		default:
			values = r.Form[q.ID]
			if q.IsRequired && len(values) == 0 {
				err = web.NewRequestError(fmt.Errorf("question '%s' is required", q.Text), http.StatusBadRequest)
			}
		}

		if err != nil {
			return err
		}

		if len(values) > 0 {
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
		SubmittedAt: now,
		Answers:     answers,
	}

	if err := m.surveyStore.SaveResponse(ctx, response); err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	if assignmentID == "" {
		return web.NewRequestError(fmt.Errorf("missing assignment ID"), http.StatusBadRequest)
	}

	// Ensure the current user is the one assigned to this specific survey instance.
	assignment, found, err := m.surveyStore.GetSpecialSurveyAssignment(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("validating assignment: %w", err)
	}
	if !found || assignment.UserEmail != user.Email {
		return web.NewRequestError(fmt.Errorf("invalid assignment for user"), http.StatusForbidden)
	}

	if err := m.surveyStore.UpdateSpecialSurveyUserResponse(ctx, assignmentID, response.ID); err != nil {
		return fmt.Errorf("failed to update special survey assignment: %w", err)
	}

	return surveyThankYouPage(user).Render(ctx, w)
}
