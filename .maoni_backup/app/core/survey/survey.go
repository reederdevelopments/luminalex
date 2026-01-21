package survey

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
)

// QuestionType defines the type of a question.
type QuestionType string

const (
	Text           QuestionType = "text"
	Textarea       QuestionType = "textarea"
	MultipleChoice QuestionType = "multiple-choice"
	Checkbox       QuestionType = "checkbox"
	Dropdown       QuestionType = "dropdown"
)

const (
	TypeNormal  = "normal"
	TypeSpecial = "special"
)

// Question represents a single question in a survey.
type Question struct {
	ID              string   `json:"id" bigquery:"id" schema:"ID" firestore:"id"`
	Text            string   `json:"text" bigquery:"text" schema:"Text" firestore:"text"`
	Type            string   `json:"type" bigquery:"type" schema:"Type" firestore:"type"`
	Options         []string `json:"options" bigquery:"options" schema:"Options" firestore:"options"`
	IsRequired      bool     `json:"is_required" bigquery:"is_required" schema:"IsRequired" firestore:"is_required"`
	GroupNumber     int      `json:"group_number" bigquery:"group_number" schema:"GroupNumber" firestore:"group_number"`
	PrefillVariable string   `json:"prefill_variable" bigquery:"prefill_variable" schema:"PrefillVariable" firestore:"prefill_variable"`
}

// Saveable satisfies the bigquery.ValueSaver interface for the Question struct.
func (q *Question) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"id":               q.ID,
		"text":             q.Text,
		"type":             q.Type,
		"options":          q.Options,
		"is_required":      q.IsRequired,
		"group_number":     q.GroupNumber,
		"prefill_variable": q.PrefillVariable,
	}, "", nil
}

// Survey represents a survey with its questions.
type Survey struct {
	ID                       string     `json:"id" bigquery:"id" firestore:"id"`
	Name                     string     `json:"name" bigquery:"name" firestore:"name"`
	Description              string     `json:"description" bigquery:"description" firestore:"description"`
	Banner                   string     `json:"banner,omitempty" bigquery:"banner" schema:"Banner" firestore:"banner,omitempty"`
	Type                     string     `json:"type" bigquery:"type" schema:"Type" firestore:"type"`
	IsEnabled                bool       `json:"is_enabled" bigquery:"is_enabled" firestore:"is_enabled"`
	AllowMultipleSubmissions bool       `json:"allow_multiple_submissions" bigquery:"allow_multiple_submissions" firestore:"allow_multiple_submissions"`
	CreatedAt                time.Time  `json:"created_at" schema:"-" bigquery:"created_at" firestore:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" schema:"-" bigquery:"updated_at" firestore:"updated_at"`
	Questions                []Question `json:"questions" schema:"questions" bigquery:"questions" firestore:"questions"`
	GroupHeadings            []string   `json:"group_headings" bigquery:"group_headings" schema:"GroupHeadings" firestore:"group_headings"`
	// This field is for display purposes only, not stored in BQ.
	ResponseCount     int `json:"response_count,omitempty" bigquery:"-" firestore:"-"`
	AssignedUserCount int `json:"assigned_user_count,omitempty" bigquery:"-" firestore:"-"`
	// AssignmentID is used for special surveys to identify a unique assignment for a user.
	AssignmentID string `json:"assignment_id,omitempty" bigquery:"-" firestore:"-"`
	// PrefillData holds variables for a special survey assignment.
	PrefillData map[string]string `json:"prefill_data,omitempty" bigquery:"-" firestore:"-"`
}

// Answer represents a user's answer to a single question.
type Answer struct {
	QuestionID string   `bigquery:"question_id"`
	Values     []string `bigquery:"values"`
}

// Response represents a single submission of a survey.
type Response struct {
	ID          string    `bigquery:"id"`
	SurveyID    string    `bigquery:"survey_id"`
	UserID      string    `bigquery:"user_id"`
	SubmittedAt time.Time `bigquery:"submitted_at"`
	Answers     []Answer  `bigquery:"answers"`
}

// SpecialSurveyUser represents a user assigned to a special survey with prefill data.
type SpecialSurveyUser struct {
	AssignmentID string `json:"assignment_id" firestore:"assignment_id"`
	SurveyID     string `json:"survey_id" firestore:"survey_id"`
	UserEmail    string `json:"user_email" firestore:"user_email"`
	Variable1    string `json:"variable_1" firestore:"variable_1,omitempty"`
	Variable2    string `json:"variable_2" firestore:"variable_2,omitempty"`
	Variable3    string `json:"variable_3" firestore:"variable_3,omitempty"`
	Variable4    string `json:"variable_4" firestore:"variable_4,omitempty"`
	Variable5    string `json:"variable_5" firestore:"variable_5,omitempty"`
	ResponseID   string `json:"response_id" firestore:"response_id"`
}

type Store interface {
	Create(ctx context.Context, s Survey) error
	Update(ctx context.Context, s Survey) error
	Get(ctx context.Context, id string) (Survey, error)
	List(ctx context.Context, showInactive bool) ([]Survey, error)
	ListForUser(ctx context.Context, userEmail string) ([]Survey, error)
	SaveResponse(ctx context.Context, r Response) error
	HasUserResponded(ctx context.Context, surveyID, userID string) (bool, error)
	GetResponseCount(ctx context.Context, surveyID string) (int, error)

	// Special Surveys
	AddSpecialSurveyUsers(ctx context.Context, users []SpecialSurveyUser) error
	ListSpecialSurveyUsers(ctx context.Context, surveyID string) ([]SpecialSurveyUser, error)
	GetSpecialSurveyAssignment(ctx context.Context, assignmentID string) (SpecialSurveyUser, bool, error)
	UpdateSpecialSurveyUserResponse(ctx context.Context, assignmentID, responseID string) error
	GetSpecialSurveyUserCount(ctx context.Context, surveyID string) (int, error)
}
