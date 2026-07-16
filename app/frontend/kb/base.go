package kb

import (
	"context"
	"controlroom/app/backend/auth"
	"controlroom/app/backend/mid"
	"controlroom/app/backend/web"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Rockup-Consulting/std/x/randx"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/iterator"
)

type Project struct {
	ID        string `firestore:"id"`
	Name      string `firestore:"name"`
	Order     int    `firestore:"order"`
	CreatedAt int64  `firestore:"created_at"`
}

type Page struct {
	ID        string `firestore:"id"`
	ProjectID string `firestore:"project_id"`
	ParentID  string `firestore:"parent_id"`
	Title     string `firestore:"title"`
	Content   string `firestore:"content"`
	Order     int    `firestore:"order"`
	CreatedAt int64  `firestore:"created_at"`
	UpdatedAt int64  `firestore:"updated_at"`
	UpdatedBy string `firestore:"updated_by"`
}

type PageHistory struct {
	ID        string `firestore:"id"`
	PageID    string `firestore:"page_id"`
	Content   string `firestore:"content"`
	UpdatedAt int64  `firestore:"updated_at"`
	UpdatedBy string `firestore:"updated_by"`
}

type module struct {
	l            *log.Logger
	sessionStore auth.Store
}

func InitModule(l *log.Logger, app *web.App, sessionStore auth.Store) {
	m := module{l: l, sessionStore: sessionStore}
	mw := []web.Middleware{mid.Log(l), mid.CatchErr(l), mid.CatchPanic(), sessionStore.Mid}

	app.Handle(http.MethodGet, "/kb", m.kbIndexHandler, mw...)
	app.Handle(http.MethodPost, "/kb/project/create", m.createProjectHandler, mw...)
	app.Handle(http.MethodPost, "/kb/project/{id}/move", m.moveProjectHandler, mw...)
	app.Handle(http.MethodPost, "/kb/page/create", m.createPageHandler, mw...)
	app.Handle(http.MethodGet, "/kb/page/{id}", m.viewPageHandler, mw...)
	app.Handle(http.MethodGet, "/kb/page/{id}/edit", m.editPageHandler, mw...)
	app.Handle(http.MethodPost, "/kb/page/{id}/save", m.savePageHandler, mw...)
	app.Handle(http.MethodPost, "/kb/page/{id}/delete", m.deletePageHandler, mw...)
	app.Handle(http.MethodPost, "/kb/page/{id}/rename", m.renamePageHandler, mw...)
	app.Handle(http.MethodPost, "/kb/page/{id}/move", m.movePageHandler, mw...)
}

func hasChildren(pages []Page, parentID string) bool {
	for _, p := range pages {
		if p.ParentID == parentID {
			return true
		}
	}
	return false
}

func (m module) getProjectsAndPages(ctx context.Context) ([]Project, map[string][]Page, error) {
	db := m.sessionStore.Db()

	var projects []Project
	pIter := db.Collection("kb_projects").OrderBy("order", firestore.Asc).OrderBy("created_at", firestore.Asc).Documents(ctx)
	for {
		doc, err := pIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		var p Project
		if err := doc.DataTo(&p); err == nil {
			projects = append(projects, p)
		}
	}

	pagesByProject := make(map[string][]Page)
	pgIter := db.Collection("kb_pages").OrderBy("order", firestore.Asc).Documents(ctx)
	for {
		doc, err := pgIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		var p Page
		if err := doc.DataTo(&p); err == nil {
			pagesByProject[p.ProjectID] = append(pagesByProject[p.ProjectID], p)
		}
	}

	return projects, pagesByProject, nil
}

func (m module) kbIndexHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	projects, pagesByProject, err := m.getProjectsAndPages(ctx)
	if err != nil {
		m.l.Printf("%v", err)
	}
	return kbLayout(projects, pagesByProject, nil, nil, false).Render(ctx, w)
}

func (m module) createProjectHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		return err
	}
	name := r.FormValue("name")
	if name != "" {
		id := randx.UID()
		_, err := m.sessionStore.Db().Collection("kb_projects").Doc(id).Set(ctx, Project{
			ID:        id,
			Name:      name,
			Order:     0,
			CreatedAt: time.Now().Unix(),
		})
		if err != nil {
			return err
		}
	}
	http.Redirect(w, r, "/kb", http.StatusSeeOther)
	return nil
}

func (m module) moveProjectHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	projectID := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		return err
	}

	orderStr := r.FormValue("order")
	if orderStr != "" {
		if order, err := strconv.Atoi(orderStr); err == nil {
			_, err := m.sessionStore.Db().Collection("kb_projects").Doc(projectID).Update(ctx, []firestore.Update{
				{Path: "order", Value: order},
			})
			if err != nil {
				return err
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (m module) createPageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User
	if err := r.ParseForm(); err != nil {
		return err
	}
	projectID := r.FormValue("project_id")
	parentID := r.FormValue("parent_id")
	title := r.FormValue("title")

	if projectID != "" && title != "" {
		id := randx.UID()
		now := time.Now().Unix()
		page := Page{
			ID:        id,
			ProjectID: projectID,
			ParentID:  parentID,
			Title:     title,
			Content:   "<h1>" + title + "</h1><p>Start writing...</p>",
			Order:     0,
			CreatedAt: now,
			UpdatedAt: now,
			UpdatedBy: user.Name,
		}
		_, err := m.sessionStore.Db().Collection("kb_pages").Doc(id).Set(ctx, page)
		if err != nil {
			return err
		}
		http.Redirect(w, r, "/kb/page/"+id+"/edit", http.StatusSeeOther)
		return nil
	}
	http.Redirect(w, r, "/kb", http.StatusSeeOther)
	return nil
}

func (m module) viewPageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	pageID := chi.URLParam(r, "id")

	doc, err := m.sessionStore.Db().Collection("kb_pages").Doc(pageID).Get(ctx)
	if err != nil || !doc.Exists() {
		http.Redirect(w, r, "/kb", http.StatusSeeOther)
		return nil
	}

	projects, pagesByProject, err := m.getProjectsAndPages(ctx)
	if err != nil {
		return err
	}

	var p Page
	if err := doc.DataTo(&p); err != nil {
		return err
	}

	if strings.Contains(p.Content, "{{TOC}}") {
		iter := m.sessionStore.Db().Collection("kb_pages").Where("parent_id", "==", p.ID).OrderBy("order", firestore.Asc).Documents(ctx)
		var children []Page
		for {
			cDoc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				break
			}
			var child Page
			if err := cDoc.DataTo(&child); err == nil {
				children = append(children, child)
			}
		}

		tocHTML := `<div class="kb-toc bg-gray-50 border border-gray-200 p-5 rounded-xl my-6 shadow-sm">
						<h3 class="text-sm font-bold text-gray-800 mb-3 uppercase tracking-wider flex items-center gap-2">
							<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-4 h-4 text-blue-600"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 010 3.75H5.625a1.875 1.875 0 010-3.75z" /></svg>
							Table of Contents
						</h3>
						<ul class="list-none p-0 m-0 flex flex-col gap-2">`

		if len(children) > 0 {
			for _, c := range children {
				tocHTML += fmt.Sprintf(`<li><a href="/kb/page/%s" class="text-blue-600 hover:text-blue-800 hover:underline font-medium text-sm flex items-center gap-2"><svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-3 h-3 text-gray-400"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" /></svg>%s</a></li>`, c.ID, c.Title)
			}
		} else {
			tocHTML += `<li class="text-gray-500 text-sm italic">No sub-pages found.</li>`
		}
		tocHTML += `</ul></div>`

		p.Content = strings.ReplaceAll(p.Content, "<p>{{TOC}}</p>", tocHTML)
		p.Content = strings.ReplaceAll(p.Content, "{{TOC}}", tocHTML)
	}

	var history []PageHistory
	hIter := m.sessionStore.Db().Collection("kb_history").Where("page_id", "==", pageID).OrderBy("updated_at", firestore.Desc).Documents(ctx)
	for {
		hDoc, err := hIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		var h PageHistory
		if err := hDoc.DataTo(&h); err == nil {
			history = append(history, h)
		}
	}

	return kbLayout(projects, pagesByProject, &p, history, false).Render(ctx, w)
}

func (m module) editPageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	pageID := chi.URLParam(r, "id")

	doc, err := m.sessionStore.Db().Collection("kb_pages").Doc(pageID).Get(ctx)
	if err != nil || !doc.Exists() {
		http.Redirect(w, r, "/kb", http.StatusSeeOther)
		return nil
	}

	projects, pagesByProject, err := m.getProjectsAndPages(ctx)
	if err != nil {
		return err
	}

	var p Page
	if err := doc.DataTo(&p); err != nil {
		return err
	}

	var history []PageHistory
	hIter := m.sessionStore.Db().Collection("kb_history").Where("page_id", "==", pageID).OrderBy("updated_at", firestore.Desc).Documents(ctx)
	for {
		hDoc, err := hIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		var h PageHistory
		if err := hDoc.DataTo(&h); err == nil {
			history = append(history, h)
		}
	}

	return kbLayout(projects, pagesByProject, &p, history, true).Render(ctx, w)
}

func (m module) savePageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	user := auth.FromCtx(ctx).User
	pageID := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		return err
	}
	content := r.FormValue("content")
	title := r.FormValue("title")

	now := time.Now().Unix()
	db := m.sessionStore.Db()

	doc, err := db.Collection("kb_pages").Doc(pageID).Get(ctx)
	if err != nil || !doc.Exists() {
		http.Redirect(w, r, "/kb", http.StatusSeeOther)
		return nil
	}

	var oldPage Page
	if err := doc.DataTo(&oldPage); err != nil {
		return err
	}

	if oldPage.Content != content || oldPage.Title != title {
		_, err := db.Collection("kb_history").Doc(randx.UID()).Set(ctx, PageHistory{
			ID:        randx.UID(),
			PageID:    pageID,
			Content:   oldPage.Content,
			UpdatedAt: now,
			UpdatedBy: user.Name,
		})
		if err != nil {
			m.l.Printf("%v", err)
		}
	}

	_, err = db.Collection("kb_pages").Doc(pageID).Update(ctx, []firestore.Update{
		{Path: "title", Value: title},
		{Path: "content", Value: content},
		{Path: "updated_at", Value: now},
		{Path: "updated_by", Value: user.Name},
	})
	if err != nil {
		return err
	}

	http.Redirect(w, r, "/kb/page/"+pageID, http.StatusSeeOther)
	return nil
}

func (m module) deletePageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	pageID := chi.URLParam(r, "id")
	_, err := m.sessionStore.Db().Collection("kb_pages").Doc(pageID).Delete(ctx)
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/kb", http.StatusSeeOther)
	return nil
}

func (m module) renamePageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	pageID := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		return err
	}
	title := r.FormValue("title")
	if title != "" {
		_, err := m.sessionStore.Db().Collection("kb_pages").Doc(pageID).Update(ctx, []firestore.Update{
			{Path: "title", Value: title},
		})
		if err != nil {
			return err
		}
	}
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	return nil
}

func (m module) movePageHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	pageID := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		return err
	}

	newParentID := r.FormValue("parent_id")
	newProjectID := r.FormValue("project_id")
	orderStr := r.FormValue("order")

	updates := []firestore.Update{
		{Path: "parent_id", Value: newParentID},
	}

	if newProjectID != "" {
		updates = append(updates, firestore.Update{Path: "project_id", Value: newProjectID})
	}

	if orderStr != "" {
		if order, err := strconv.Atoi(orderStr); err == nil {
			updates = append(updates, firestore.Update{Path: "order", Value: order})
		}
	}

	_, err := m.sessionStore.Db().Collection("kb_pages").Doc(pageID).Update(ctx, updates)
	if err != nil {
		return err
	}

	if newProjectID != "" {
		m.updateProjectRecursive(ctx, pageID, newProjectID)
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.WriteHeader(http.StatusOK)
		return nil
	}

	http.Redirect(w, r, "/kb/page/"+pageID, http.StatusSeeOther)
	return nil
}

func (m module) updateProjectRecursive(ctx context.Context, parentID string, newProjectID string) {
	db := m.sessionStore.Db()
	iter := db.Collection("kb_pages").Where("parent_id", "==", parentID).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		childID := doc.Ref.ID
		_, _ = db.Collection("kb_pages").Doc(childID).Update(ctx, []firestore.Update{
			{Path: "project_id", Value: newProjectID},
		})
		m.updateProjectRecursive(ctx, childID, newProjectID)
	}
}
