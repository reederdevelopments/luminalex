package base

import (
	"controlroom/app/backend/auth"
	"controlroom/app/backend/mid"
	"controlroom/app/backend/web"
	"log"
	"net/http"
)

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
) {
	// Unprotected routes
	m := module{
		l:            l,
		sessionStore: sessionStore,
	}
	app.Handle(http.MethodGet, "/signin", m.signinLoader, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}", m.beginAuthHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}/callback", m.authCallbackHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/logout", m.logoutHandler, stdMid(l, sessionStore.Mid)...)

	// Protected routes (Requires Auth)
	app.Handle(http.MethodGet, "/", m.homeHandler, stdMid(l, sessionStore.Mid)...)

}

type module struct {
	l            *log.Logger
	sessionStore auth.Store
}

func (m module) homeHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User

	return homePage(user).Render(ctx, w)
}
