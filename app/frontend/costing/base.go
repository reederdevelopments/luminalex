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

	// 1. Kick off the asynchronous cache preloaders
	go m.preloadCacheWorker()
	go m.preloadFastCacheWorker() // 10 min preloader for 90-day

	app.Handle(http.MethodGet, "/costing", m.costingPageHandler, stdMid(l, sessionStore.Mid)...)

	// Overall Tab Routes (Executive Summary)
	app.Handle(http.MethodGet, "/costing/overall", m.overallTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/overall/metrics", m.overallMetricsHandler, stdMid(l, sessionStore.Mid)...)

	// Datastream Tab Routes
	app.Handle(http.MethodGet, "/costing/datastream", m.datastreamTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/datastream/metrics", m.datastreamMetricsHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/datastream/details", m.datastreamProjectDetailsHandler, stdMid(l, sessionStore.Mid)...)

	// 3rd Party Tab Routes
	app.Handle(http.MethodGet, "/costing/thirdparty", m.thirdpartyTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/thirdparty/metrics", m.thirdpartyMetricsHandler, stdMid(l, sessionStore.Mid)...)

	// Dataform Tab Routes
	app.Handle(http.MethodGet, "/costing/dataform", m.dataformTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/dataform/metrics", m.dataformMetricsHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/dataform/details", m.dataformProjectDetailsHandler, stdMid(l, sessionStore.Mid)...)

	// Users Tab Routes
	app.Handle(http.MethodGet, "/costing/users", m.usersTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/users/metrics", m.usersMetricsHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/users/details", m.usersDetailsHandler, stdMid(l, sessionStore.Mid)...)

	// DataStudio Tab Routes
	app.Handle(http.MethodGet, "/costing/datastudio", m.datastudioTabHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/datastudio/metrics", m.datastudioMetricsHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/costing/datastudio/details", m.datastudioDetailsHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/costing/datastudio/mapping", m.saveDSMappingHandler, stdMid(l, sessionStore.Mid)...)
}

func (m module) costingPageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	return costingPage().Render(ctx, w)
}
