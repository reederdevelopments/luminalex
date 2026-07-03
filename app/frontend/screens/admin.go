package base

import (
	"net/http"
	"sort"
	"strconv"
	"ujuzi_reloaded/app/backend/auth"
	"ujuzi_reloaded/app/backend/collection"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/iterator"
)

func (m module) adminLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var users []auth.User
	iter := m.sessionStore.Db().Collection(collection.Users).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		var u auth.User
		doc.DataTo(&u)

		// ADD THIS LINE: Manually map the ID for HTMX to use in the frontend
		u.ID = doc.Ref.ID

		users = append(users, u)
	}

	groups := m.getAllGroupsSorted(r)
	dashboards := m.getAllDashboards(r)

	return adminManagementPage(users, groups, dashboards).Render(ctx, w)
}

func (m module) adminUserDetails(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	doc, err := m.sessionStore.Db().Collection(collection.Users).Doc(userID).Get(ctx)
	if err != nil {
		return err
	}
	var u auth.User
	doc.DataTo(&u)

	// ADD THIS LINE: Manually map the ID so the form knows where to POST
	u.ID = doc.Ref.ID

	groups := m.getAllGroupsSorted(r)
	return userDetailsPanel(u, groups).Render(ctx, w)
}

func (m module) adminSaveUserGroups(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	r.ParseForm()
	selectedGroups := r.Form["groups"]
	selectedTools := r.Form["tools"] // New tools array

	_, err := m.sessionStore.Db().Collection(collection.Users).Doc(userID).Update(ctx, []firestore.Update{
		{Path: "Groups", Value: selectedGroups},
		{Path: "Tools", Value: selectedTools},
	})
	if err != nil {
		return err
	}

	// Invalidate specific user cache so they get the fresh tools/groups on next request
	m.sessionStore.InvalidateUserCache(userID)

	w.Header().Set("HX-Trigger", "user-saved")
	w.Write([]byte(`<div class="p-4 bg-green-50 text-green-700 rounded mb-4 font-bold text-sm">Assignments updated successfully!</div>`))
	return nil
}

// --- NEW GROUP MANAGEMENT ENDPOINTS ---

func (m module) adminCreateGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()

	level := r.FormValue("level")
	if !m.isLevelUnique(r, level, "") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="text-red-500 text-xs font-bold mt-2">Error: Level must be unique!</div>`))
		return nil
	}

	newGroup := auth.Group{
		Name:  r.FormValue("name"),
		Level: level,
	}

	ref, _, err := m.sessionStore.Db().Collection("groups").Add(ctx, newGroup)
	if err != nil {
		return err
	}
	newGroup.ID = ref.ID

	// Refresh the whole group list via HTMX trigger or return newly sorted list
	w.Header().Set("HX-Refresh", "true")
	return nil
}

func (m module) adminGroupDetails(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	groupID := chi.URLParam(r, "id")

	doc, err := m.sessionStore.Db().Collection("groups").Doc(groupID).Get(ctx)
	if err != nil {
		return err
	}
	var g auth.Group
	doc.DataTo(&g)
	g.ID = doc.Ref.ID

	dashboards := m.getAllDashboards(r)
	return groupDetailsPanel(g, dashboards).Render(ctx, w)
}

func (m module) adminUpdateGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	groupID := chi.URLParam(r, "id")
	r.ParseForm()

	level := r.FormValue("level")
	if !m.isLevelUnique(r, level, groupID) {
		w.Write([]byte(`<div class="p-4 bg-red-50 text-red-700 rounded mb-4 font-bold text-sm">Error: Level must be unique!</div>`))
		return nil
	}

	updates := []firestore.Update{
		{Path: "name", Value: r.FormValue("name")},
		{Path: "level", Value: level},
		{Path: "dashboards", Value: r.Form["dashboards"]},
	}

	_, err := m.sessionStore.Db().Collection("groups").Doc(groupID).Update(ctx, updates)
	if err != nil {
		return err
	}

	// Group changes affect multiple users' access scopes, wipe the cache to force fresh DB reads
	m.sessionStore.ClearAllUserCache()

	// Trigger a full iframe reload to update the left-hand table automatically
	w.Header().Set("HX-Refresh", "true")
	return nil
}

func (m module) adminDeleteGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	groupID := chi.URLParam(r, "id")
	_, err := m.sessionStore.Db().Collection("groups").Doc(groupID).Delete(ctx)
	if err != nil {
		return err
	}

	// Wipe user cache to ensure removed groups are purged from memory evaluations
	m.sessionStore.ClearAllUserCache()

	w.Header().Set("HX-Refresh", "true")
	return nil
}

// --- HELPERS ---

func (m module) getAllGroupsSorted(r *http.Request) []auth.Group {
	ctx := r.Context()
	var groups []auth.Group
	gIter := m.sessionStore.Db().Collection("groups").Documents(ctx)
	for {
		doc, err := gIter.Next()
		if err != nil {
			break
		}
		var g auth.Group
		doc.DataTo(&g)
		g.ID = doc.Ref.ID
		groups = append(groups, g)
	}

	sort.Slice(groups, func(i, j int) bool {
		li, _ := strconv.Atoi(groups[i].Level)
		lj, _ := strconv.Atoi(groups[j].Level)
		return li < lj // ASC
	})
	return groups
}

func (m module) isLevelUnique(r *http.Request, level string, ignoreID string) bool {
	ctx := r.Context()
	iter := m.sessionStore.Db().Collection("groups").Where("level", "==", level).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		if doc.Ref.ID != ignoreID {
			return false // found another group with same level
		}
	}
	return true
}

func (m module) getAllDashboards(r *http.Request) []auth.Dashboard {
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
