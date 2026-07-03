package base

import (
	"net/http"
	"ujuzi_reloaded/app/backend/auth"

	"github.com/go-chi/chi/v5"
)

func (m module) adminDashboardsLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	dashboards := m.getAllDashboards(r)
	return adminDashboardsPage(dashboards).Render(ctx, w)
}

func (m module) adminDashboardDetails(w http.ResponseWriter, r *http.Request) error {
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

func (m module) adminSaveDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()

	dashID := r.FormValue("id")

	// Parse parameters as strings for EditBoxes
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

func (m module) adminDeleteDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	dashID := chi.URLParam(r, "id")
	_, err := m.sessionStore.Db().Collection("dashboards").Doc(dashID).Delete(ctx)
	if err != nil {
		return err
	}

	w.Header().Set("HX-Refresh", "true")
	return nil
}
