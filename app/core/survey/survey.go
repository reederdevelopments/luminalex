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
	ID              string   `json:"id" bigquery:"id" schema:"ID"`
	Text            string   `json:"text" bigquery:"text" schema:"Text"`
	Type            string   `json:"type" bigquery:"type" schema:"Type"`
	Options         []string `json:"options" bigquery:"options" schema:"Options"`
	IsRequired      bool     `json:"is_required" bigquery:"is_required" schema:"IsRequired"`
	GroupNumber     int      `json:"group_number" bigquery:"group_number" schema:"GroupNumber"`
	PrefillVariable string   `json:"prefill_variable" bigquery:"prefill_variable" schema:"PrefillVariable"`
}

// Survey represents a survey with its questions.
type Survey struct {
	ID                       string     `json:"id" bigquery:"id"`
	Name                     string     `json:"name" bigquery:"name"`
	Description              string     `json:"description" bigquery:"description"`
	Type                     string     `json:"type" bigquery:"type" schema:"Type"`
	IsEnabled                bool       `json:"is_enabled" bigquery:"is_enabled"`
	AllowMultipleSubmissions bool       `json:"allow_multiple_submissions" bigquery:"allow_multiple_submissions"`
	CreatedAt                time.Time  `json:"created_at" schema:"-" bigquery:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" schema:"-" bigquery:"updated_at"`
	Questions                []Question `json:"questions" schema:"questions" bigquery:"questions"`
	GroupHeadings            []string   `json:"group_headings" bigquery:"group_headings" schema:"GroupHeadings"`
	// This field is for display purposes only, not stored in BQ.
	ResponseCount int `json:"response_count,omitempty" bigquery:"-"`
	// AssignmentID is used for special surveys to identify a unique assignment for a user.
	AssignmentID string `json:"assignment_id,omitempty" bigquery:"-"`
	// PrefillData holds variables for a special survey assignment.
	PrefillData map[string]string `json:"prefill_data,omitempty" bigquery:"-"`
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
	AssignmentID string              `bigquery:"assignment_id" json:"assignment_id"`
	SurveyID     string              `bigquery:"survey_id" json:"survey_id"`
	UserEmail    string              `bigquery:"user_email" json:"user_email"`
	Variable1    string              `bigquery:"variable_1" json:"variable_1"`
	Variable2    string              `bigquery:"variable_2" json:"variable_2"`
	Variable3    string              `bigquery:"variable_3" json:"variable_3"`
	Variable4    string              `bigquery:"variable_4" json:"variable_4"`
	Variable5    string              `bigquery:"variable_5" json:"variable_5"`
	ResponseID   bigquery.NullString `bigquery:"response_id" json:"response_id"`
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
}
