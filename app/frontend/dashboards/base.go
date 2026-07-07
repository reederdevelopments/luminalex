package base

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"ujuzi_reloaded/app/backend/auth"
	"ujuzi_reloaded/app/backend/collection"
	"ujuzi_reloaded/app/backend/mid"
	"ujuzi_reloaded/app/backend/web"
	"ujuzi_reloaded/app/frontend/admin/dashboard_management"
	"ujuzi_reloaded/app/frontend/admin/engineering"
	"ujuzi_reloaded/app/frontend/admin/user_management"
	"ujuzi_reloaded/app/frontend/tools/dataform"

	"cloud.google.com/go/firestore"
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
	coreDBs map[string]*sql.DB, // <-- ADDED: Inject MySQL DBs
) {
	// The new modules
	userMgmt := user_management.NewModule(l, sessionStore, coreDBs) // <-- FIXED
	dashMgmt := dashboard_management.NewModule(l, sessionStore)
	engMgmt := engineering.NewModule(l, sessionStore)
	dfTool := dataform.NewModule(l, sessionStore)

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

	// API Tracking
	app.Handle(http.MethodPost, "/api/track-click", m.trackClickHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/api/favorites", m.updateFavoritesHandler, stdMid(l, sessionStore.Mid)...)

	// ADMIN ROUTES
	app.Handle(http.MethodGet, "/admin/management", userMgmt.AdminLoader, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/admin/api/user/{id}", userMgmt.AdminUserDetails, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/user/{id}/groups", userMgmt.AdminSaveUserGroups, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/group", userMgmt.AdminCreateGroup, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/admin/api/group/{id}", userMgmt.AdminGroupDetails, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/group/{id}", userMgmt.AdminUpdateGroup, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodDelete, "/admin/api/group/{id}", userMgmt.AdminDeleteGroup, stdMid(l, sessionStore.Mid)...)

	// Manual System Sync Triggers
	app.Handle(http.MethodPost, "/admin/api/sync/users", userMgmt.AdminSyncUsers, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/sync/usergroups", userMgmt.AdminSyncUserGroups, stdMid(l, sessionStore.Mid)...)

	app.Handle(http.MethodGet, "/admin/dashboards", dashMgmt.AdminDashboardsLoader, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/dashboard", dashMgmt.AdminSaveDashboard, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/admin/api/dashboard/{id}", dashMgmt.AdminDashboardDetails, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodDelete, "/admin/api/dashboard/{id}", dashMgmt.AdminDeleteDashboard, stdMid(l, sessionStore.Mid)...)

	// ENGINEERING ROUTES
	app.Handle(http.MethodGet, "/admin/engineering", engMgmt.AdminEngineeringLoader, stdMid(l, sessionStore.Mid)...)

	// Reload
	app.Handle(http.MethodGet, "/admin/api/engineering/reload/list", engMgmt.EngineeringReloadList, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/engineering/reload/action", engMgmt.EngineeringReloadAction, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/engineering/reload/toggle-enable", engMgmt.EngineeringReloadToggleEnable, stdMid(l, sessionStore.Mid)...)

	// Job Config
	app.Handle(http.MethodGet, "/admin/api/engineering/jobconfig/list", engMgmt.EngineeringJobConfigList, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/admin/api/engineering/jobconfig/{id}", engMgmt.EngineeringJobConfigDetails, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/admin/api/engineering/jobconfig", engMgmt.EngineeringJobConfigSave, stdMid(l, sessionStore.Mid)...)

	// DATAFORM ROUTES
	app.Handle(http.MethodGet, "/tools/dataform", dfTool.Loader, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/tools/api/dataform/batch/status", dfTool.BatchStatus, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/tools/api/dataform/batch/run", dfTool.BatchRun, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/tools/api/dataform/batch/cancel", dfTool.BatchCancel, stdMid(l, sessionStore.Mid)...)

	// START BACKGROUND CACHE
	engMgmt.StartReloadBackgroundCache()
}

type module struct {
	l            *log.Logger
	sessionStore auth.Store
}

func (m module) homeHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User

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

func (m module) updateFavoritesHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sessionUser := auth.FromCtx(ctx).User

	var payload struct {
		Favorites []string `json:"favorites"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	if payload.Favorites == nil {
		payload.Favorites = []string{}
	}

	_, err := m.sessionStore.Db().Collection(collection.Users).Doc(sessionUser.ID).Update(ctx, []firestore.Update{
		{Path: "Favorites", Value: payload.Favorites},
	})
	if err != nil {
		return err
	}

	m.sessionStore.InvalidateUserCache(sessionUser.ID)
	w.WriteHeader(http.StatusOK)
	return nil
}

func (m module) trackClickHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	sessionUser := auth.FromCtx(ctx).User

	var payload struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		m.l.Printf("Track Click: Decode failed - %v", err)
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	if payload.Name == "" || payload.Url == "" {
		return nil
	}

	// 1. Log Activity
	logEntry := map[string]interface{}{
		"UserID":        sessionUser.ID,
		"UserName":      sessionUser.FirstName + " " + sessionUser.LastName,
		"DashboardName": payload.Name,
		"DashboardURL":  payload.Url,
		"Timestamp":     time.Now().Unix(),
	}
	m.sessionStore.Db().Collection("activity_logs").Add(ctx, logEntry)

	// 2. Update User Profile Safely
	doc, err := m.sessionStore.Db().Collection(collection.Users).Doc(sessionUser.ID).Get(ctx)
	if err == nil {
		var u auth.User
		doc.DataTo(&u)

		if u.DashboardClicks == nil {
			u.DashboardClicks = make(map[string]int)
		}

		rawKey := payload.Name + "|#|" + payload.Url

		// Use RawURLEncoding to guarantee NO slashes, dots, OR equals signs
		safeKey := base64.RawURLEncoding.EncodeToString([]byte(rawKey))

		u.DashboardClicks[safeKey]++

		// Bypass firestore.Update path limitations using Set + MergeAll
		_, err = m.sessionStore.Db().Collection(collection.Users).Doc(sessionUser.ID).Set(ctx, map[string]interface{}{
			"dashboardClicks": u.DashboardClicks,
		}, firestore.MergeAll)

		if err != nil {
			m.l.Printf("Track Click: Firestore Set failed! Error: %v", err)
		}
	} else {
		m.l.Printf("Track Click: Failed to get user document - %v", err)
	}

	// 3. Clear Cache
	m.sessionStore.InvalidateUserCache(sessionUser.ID)

	w.WriteHeader(http.StatusOK)
	return nil
}
