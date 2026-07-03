package base

import (
	"log"
	"net/http"
	"time"
	"ujuzi_reloaded/app/backend/auth"
	"ujuzi_reloaded/app/backend/collection"
	"ujuzi_reloaded/app/backend/mid"
	"ujuzi_reloaded/app/backend/web"

	"cloud.google.com/go/firestore"
)

type module struct {
	l            *log.Logger
	sessionStore auth.Store
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
) {
	m := module{
		l:            l,
		sessionStore: sessionStore,
	}

	// Unprotected routes
	app.Handle(http.MethodGet, "/signin", m.signinLoader, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}", m.beginAuthHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}/callback", m.authCallbackHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/logout", m.logoutHandler, stdMid(l, sessionStore.Mid)...)

	// Protected routes (Requires Auth)
	app.Handle(http.MethodGet, "/", m.homeHandler, stdMid(l, sessionStore.Mid)...)

	// API Tracking
	app.Handle(http.MethodPost, "/api/track-click", m.trackClickHandler, stdMid(l, sessionStore.Mid)...)

	// ADMIN ROUTES
	app.Handle(http.MethodGet, "/admin/management", m.adminLoader, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/admin/api/user/{id}", m.adminUserDetails, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/user/{id}/groups", m.adminSaveUserGroups, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/group", m.adminCreateGroup, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/admin/api/group/{id}", m.adminGroupDetails, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/group/{id}", m.adminUpdateGroup, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodDelete, "/admin/api/group/{id}", m.adminDeleteGroup, stdMid(l, sessionStore.Mid)...)

	app.Handle(http.MethodGet, "/admin/dashboards", m.adminDashboardsLoader, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/dashboard", m.adminSaveDashboard, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/admin/api/dashboard/{id}", m.adminDashboardDetails, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodDelete, "/admin/api/dashboard/{id}", m.adminDeleteDashboard, stdMid(l, sessionStore.Mid)...)
}

func (m module) homeHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sessionUser := auth.FromCtx(ctx).User

	// 1. Force fetch latest user data from DB so Groups & Tools update instantly
	doc, err := m.sessionStore.Db().Collection(collection.Users).Doc(sessionUser.ID).Get(ctx)
	var user auth.User
	if err == nil {
		doc.DataTo(&user)
		user.ID = doc.Ref.ID
	} else {
		user = sessionUser // fallback to session cache if DB fails
	}

	// 2. Fetch all groups
	var allGroups []auth.Group
	gIter := m.sessionStore.Db().Collection("groups").Documents(ctx)
	for {
		doc, err := gIter.Next()
		if err != nil {
			break
		}
		var g auth.Group
		doc.DataTo(&g)
		g.ID = doc.Ref.ID
		allGroups = append(allGroups, g)
	}

	// 3. Fetch all dashboards
	var allDashboards []auth.Dashboard
	dIter := m.sessionStore.Db().Collection("dashboards").Documents(ctx)
	for {
		doc, err := dIter.Next()
		if err != nil {
			break
		}
		var d auth.Dashboard
		doc.DataTo(&d)
		d.ID = doc.Ref.ID
		allDashboards = append(allDashboards, d)
	}

	return homePage(user, allGroups, allDashboards).Render(ctx, w)
}

func (m module) trackClickHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sessionUser := auth.FromCtx(ctx).User

	r.ParseForm()
	name := r.FormValue("name")
	urlStr := r.FormValue("url")

	if name == "" || urlStr == "" {
		return nil
	}

	// 1. Log the activity for analytics
	logEntry := map[string]interface{}{
		"UserID":        sessionUser.ID,
		"UserName":      sessionUser.FirstName + " " + sessionUser.LastName,
		"DashboardName": name,
		"DashboardURL":  urlStr,
		"Timestamp":     time.Now().Unix(),
	}
	m.sessionStore.Db().Collection("activity_logs").Add(ctx, logEntry)

	// 2. Safely Update user's click history map for dynamic top 5 apps
	doc, err := m.sessionStore.Db().Collection(collection.Users).Doc(sessionUser.ID).Get(ctx)
	if err == nil {
		var u auth.User
		doc.DataTo(&u)
		if u.DashboardClicks == nil {
			u.DashboardClicks = make(map[string]int)
		}
		key := name + "|#|" + urlStr
		u.DashboardClicks[key]++
		m.sessionStore.Db().Collection(collection.Users).Doc(sessionUser.ID).Update(ctx, []firestore.Update{
			{Path: "dashboardClicks", Value: u.DashboardClicks},
		})
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
