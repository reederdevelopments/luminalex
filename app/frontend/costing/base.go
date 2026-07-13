package costing

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

type module struct {
	l            *log.Logger
	sessionStore auth.Store
	cache        *CostCache
}

func InitModule(l *log.Logger, app *web.App, sessionStore auth.Store) {
	costCache := NewCostCache(l)

	m := module{
		l:            l,
		sessionStore: sessionStore,
		cache:        costCache,
	}

	// Kick off the comprehensive asynchronous cache preloader
	go m.preloadCacheWorker()

	// Main Layout Route
	app.Handle(http.MethodGet, "/costing", m.costingPageHandler, stdMid(l, sessionStore.Mid)...)

	// Datastream Tab Routes
	app.Handle(http.MethodGet, "/costing/datastream", m.datastreamTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/datastream/metrics", m.datastreamMetricsHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/datastream/details", m.datastreamProjectDetailsHandler, stdMid(l, sessionStore.Mid)...)
}

func (m module) costingPageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	return costingPage().Render(ctx, w)
}
