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

	if !s.IsEnabled {
		return web.NewRequestError(fmt.Errorf("survey is not active"), http.StatusForbidden)
	}

	prefills := make(map[string]string)
	if s.Type == survey.TypeSpecial {
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

	} else if !s.AllowMultipleSubmissions {
		hasResponded, err := m.surveyStore.HasUserResponded(ctx, surveyID, user.ID)
		if err != nil {
			return fmt.Errorf("checking user response status: %w", err)
		}
		if hasResponded {
			return surveyAlreadyCompletedPage(user).Render(ctx, w)
		}
	}

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

	if !s.IsEnabled {
		return web.NewRequestError(fmt.Errorf("survey is not active"), http.StatusForbidden)
	}

	// --- Validation & Response creation ---
	var answers []survey.Answer
	for _, q := range s.Questions {
		values := r.Form[q.ID]
		if q.IsRequired && len(values) == 0 {
			// This is basic validation. More robust validation might be needed.
			return web.NewRequestError(fmt.Errorf("question '%s' is required", q.Text), http.StatusBadRequest)
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
		SubmittedAt: time.Now(),
		Answers:     answers,
	}

	if err := m.surveyStore.SaveResponse(ctx, response); err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	if s.Type == survey.TypeSpecial {
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
	}

	return surveyThankYouPage(user).Render(ctx, w)
}
