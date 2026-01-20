package base

import (
	"log"
	"maoni/app/core/auth"
	"maoni/app/core/mid"
	"maoni/app/core/survey"
	"maoni/app/core/web"
	"net/http"
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

func InitModule(
	l *log.Logger,
	app *web.App,
	sessionStore auth.Store,
	surveyStore survey.Store,
) {
	m := module{
		l:            l,
		sessionStore: sessionStore,
		surveyStore:  surveyStore,
	}

	// Unprotected routes
	app.Handle(http.MethodGet, "/signin", m.signinLoader, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}", m.beginAuthHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}/callback", m.authCallbackHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/logout", m.logoutHandler, stdMid(l, sessionStore.Mid)...)

	// Protected routes
	app.Handle(http.MethodGet, "/", m.surveyLoader, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/surveys/{id}", m.viewSurveyHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/surveys/{id}", m.submitSurveyHandler, stdMid(l, sessionStore.Mid)...)
}
