package base

import (
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type SearchResult struct {
	Title string
	URL   string
	Icon  string
	Type  string
}

func (m module) searchHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	var results []SearchResult

	if query == "" {
		return searchDropdown(results, "").Render(ctx, w)
	}

	// 1. Search Static App Modules
	staticRoutes := []SearchResult{
		{Title: "Costing: Datastream", URL: "/costing/datastream", Type: "Module", Icon: "📊"},
		{Title: "Costing: Dataform", URL: "/costing/dataform", Type: "Module", Icon: "⚙️"},
		{Title: "Costing: 3rd Party", URL: "/costing/thirdparty", Type: "Module", Icon: "🔌"},
		{Title: "Costing: Users", URL: "/costing/users", Type: "Module", Icon: "👥"},
		{Title: "Costing: DataStudio", URL: "/costing/datastudio", Type: "Module", Icon: "📈"},
		{Title: "Knowledge Base", URL: "/kb", Type: "Module", Icon: "📚"},
	}

	for _, route := range staticRoutes {
		if strings.Contains(strings.ToLower(route.Title), query) {
			results = append(results, route)
		}
	}

	// 2. Search KB Pages in Firestore
	iter := m.sessionStore.Db().Collection("kb_pages").OrderBy("updated_at", firestore.Desc).Limit(50).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}

		title := ""
		if t, ok := doc.Data()["title"].(string); ok {
			title = t
		}
		content := ""
		if c, ok := doc.Data()["content"].(string); ok {
			content = c
		}

		if strings.Contains(strings.ToLower(title), query) || strings.Contains(strings.ToLower(content), query) {
			results = append(results, SearchResult{
				Title: title,
				URL:   "/kb/page/" + doc.Ref.ID,
				Type:  "KB Page",
				Icon:  "📄",
			})
		}
	}

	return searchDropdown(results, query).Render(ctx, w)
}
