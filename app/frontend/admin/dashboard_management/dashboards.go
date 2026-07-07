package dashboard_management

import (
	"net/http"
	"ujuzi_reloaded/app/backend/auth"
	"log"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	l            *log.Logger
	sessionStore auth.Store
}

func NewModule(l *log.Logger, sessionStore auth.Store) *Module {
	return &Module{
		l:            l,
		sessionStore: sessionStore,
	}
}

func (m *Module) AdminDashboardsLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	dashboards := m.getAllDashboards(r)
	return adminDashboardsPage(dashboards).Render(ctx, w)
}

func (m *Module) AdminDashboardDetails(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	dashID := chi.URLParam(r, "id")

	doc, err := m.sessionStore.Db().Collection("dashboards").Doc(dashID).Get(ctx)
	if err != nil {
		return err
	}
	var d auth.Dashboard
	doc.DataTo(&d)
	d.ID = doc.Ref.ID

	return dashDetailsPanel(d).Render(ctx, w)
}

func (m *Module) AdminSaveDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()

	dashID := r.FormValue("id")

	params := auth.DashParams{
		CountryCode: r.FormValue("param_country_code"),
		Branch:      r.FormValue("param_branch"),
		Consultant:  r.FormValue("param_consultant"),
		Cycle:       r.FormValue("param_cycle"),
		PrevCycle:   r.FormValue("param_prev_cycle"),
		Date:        r.FormValue("param_date"),
		Period:      r.FormValue("param_period"),
	}

	pageNames := r.Form["page_name"]
	pageCodes := r.Form["page_code"]
	var pages []auth.DashPage
	for i := 0; i < len(pageNames); i++ {
		if pageNames[i] != "" && pageCodes[i] != "" {
			pages = append(pages, auth.DashPage{Name: pageNames[i], Code: pageCodes[i]})
		}
	}

	dashboard := auth.Dashboard{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		ReportURL:   r.FormValue("reportUrl"),
		Area:        r.FormValue("area"),
		Parameters:  params,
		Pages:       pages,
	}

	if dashID == "" {
		_, _, err := m.sessionStore.Db().Collection("dashboards").Add(ctx, dashboard)
		if err != nil {
			return err
		}
	} else {
		_, err := m.sessionStore.Db().Collection("dashboards").Doc(dashID).Set(ctx, dashboard)
		if err != nil {
			return err
		}
	}

	w.Header().Set("HX-Refresh", "true")
	return nil
}

func (m *Module) AdminDeleteDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	dashID := chi.URLParam(r, "id")
	_, err := m.sessionStore.Db().Collection("dashboards").Doc(dashID).Delete(ctx)
	if err != nil {
		return err
	}

	w.Header().Set("HX-Refresh", "true")
	return nil
}

func (m *Module) getAllDashboards(r *http.Request) []auth.Dashboard {
	ctx := r.Context()
	var dashes []auth.Dashboard
	iter := m.sessionStore.Db().Collection("dashboards").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var d auth.Dashboard
		doc.DataTo(&d)
		d.ID = doc.Ref.ID
		dashes = append(dashes, d)
	}
	return dashes
}
