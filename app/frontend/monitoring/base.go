package monitoring

import (
	"controlroom/app/backend/auth"
	"controlroom/app/backend/mid"
	"controlroom/app/backend/web"
	"log"
	"net/http"
	"time"
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

type module struct {
	l            *log.Logger
	sessionStore auth.Store
	cache        *MonCache
}

func InitModule(l *log.Logger, app *web.App, sessionStore auth.Store) {
	monCache := NewMonCache(l)

	m := module{
		l:            l,
		sessionStore: sessionStore,
		cache:        monCache,
	}

	app.Handle(http.MethodGet, "/monitoring", m.monitoringPageHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/monitoring/thirdparty", m.thirdpartyTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/monitoring/thirdparty/list", m.thirdpartyListHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/monitoring/thirdparty/detail", m.thirdpartyDetailHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/monitoring/thirdparty/trigger", m.triggerJobHandler, stdMid(l, sessionStore.Mid)...)
}

func (m module) monitoringPageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	data := MonitoringDashboardData{
		StartDate: firstOfMonth.Format("2006-01-02"),
		EndDate:   now.Format("2006-01-02"),
	}

	return monitoringPage(data).Render(ctx, w)
}
