package base

import (
	"maoni/app/core/auth"
	"net/http"
)

func (m module) surveyLoader(w http.ResponseWriter, r *http.Request) error {
	user := auth.FromCtx(r.Context()).User
	return surveyPage(user.Name).Render(r.Context(), w)
}
