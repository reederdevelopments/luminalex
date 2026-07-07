package engineering

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	executions "cloud.google.com/go/workflows/executions/apiv1"
	"cloud.google.com/go/workflows/executions/apiv1/executionspb"
	"google.golang.org/api/option"
)

type ReloadCacheData struct {
	Collections []string
	Entities    []CoreReload
	LastUpdated time.Time
}

var (
	reloadCacheMu sync.RWMutex
	reloadCache   = make(map[string]*ReloadCacheData)
)

func (m *Module) StartReloadBackgroundCache() {
	m.l.Println("Starting background cache refresh for /reload page...")

	refreshAll := func() {
		for _, cc := range []string{"za", "ke", "ug", "tz", "zm"} {
			m.refreshReloadCacheForCountry(cc)
		}
	}

	go refreshAll()
	go m.PollWorkflowStatuses()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			m.l.Println("Starting periodic cache refresh for /reload page...")
			refreshAll()
		}
	}()
}

func (m *Module) PollWorkflowStatuses() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		ctx := context.Background()
		for _, cc := range []string{"za", "ke", "ug", "tz", "zm"} {
			projectID := getProjectForCountry(cc)
			dsClient, err := getDatastoreClient(ctx, projectID)
			if err != nil {
				continue
			}

			q := datastore.NewQuery("CoreReload").Filter("refresh_in_progress =", true)
			var refreshing []CoreReload
			keys, err := dsClient.GetAll(ctx, q, &refreshing)
			if err != nil {
				continue
			}

			if len(refreshing) > 0 {
				m.l.Printf("Found %d refreshing tables in %s", len(refreshing), cc)
			}

			for i, entity := range refreshing {
				if entity.ExecutionID == "" {
					continue
				}

				execClient, err := executions.NewClient(ctx, option.WithEndpoint("workflowexecutions.googleapis.com:443"))
				if err != nil {
					continue
				}

				req := &executionspb.GetExecutionRequest{Name: entity.ExecutionID}
				execution, err := execClient.GetExecution(ctx, req)
				if err != nil {
					m.l.Printf("Error getting execution status for %s: %v", entity.ExecutionID, err)
					continue
				}

				if execution.State == executionspb.Execution_SUCCEEDED || execution.State == executionspb.Execution_FAILED || execution.State == executionspb.Execution_CANCELLED {
					m.l.Printf("Execution %s for table %s finished with state %s", entity.ExecutionID, entity.Name, execution.State.String())
					entity.RefreshInProgress = false
					entity.ExecutionID = ""
					if _, err := dsClient.Put(ctx, keys[i], &entity); err != nil {
						m.l.Printf("Error updating entity %s: %v", keys[i].Name, err)
					}
				}
				execClient.Close()
			}
		}
	}
}

func (m *Module) refreshReloadCacheForCountry(cc string) {
	ctx := context.Background()
	m.l.Printf("Starting background refresh for /reload page data for country: %s", cc)

	projectID := getProjectForCountry(cc)
	client, err := getDatastoreClient(ctx, projectID)
	if err != nil {
		m.l.Printf("Failed to connect to %s: %v", projectID, err)
		return
	}

	query := datastore.NewQuery("__kind__").KeysOnly()
	keys, err := client.GetAll(ctx, query, nil)
	if err != nil {
		m.l.Printf("Permission/Read error on %s: %v", projectID, err)
		return
	}

	var collections []string
	allEntitiesMap := make(map[string]CoreReload)

	for _, k := range keys {
		if strings.HasPrefix(k.Name, "__") {
			continue
		}
		collections = append(collections, k.Name)

		var entities []CoreReload
		q := datastore.NewQuery(k.Name)
		eKeys, err := client.GetAll(ctx, q, &entities)

		if err != nil {
			if len(entities) == 0 {
				m.l.Printf("Warning: failed to get entities for collection %s in %s: %v", k.Name, cc, err)
				continue
			}
		}

		for i, entity := range entities {
			entity.EncodedKey = eKeys[i].Encode()
			jobKey := fmt.Sprintf("%s-%s", entity.Database, entity.Name)
			if existing, ok := allEntitiesMap[jobKey]; ok {
				if entity.PreviousTableRefreshSAST > existing.PreviousTableRefreshSAST {
					allEntitiesMap[jobKey] = entity
				}
			} else {
				allEntitiesMap[jobKey] = entity
			}
		}
	}

	var uniqueEntities []CoreReload
	for _, entity := range allEntitiesMap {
		uniqueEntities = append(uniqueEntities, entity)
	}

	sort.Slice(uniqueEntities, func(i, j int) bool {
		return uniqueEntities[i].Name < uniqueEntities[j].Name
	})

	reloadCacheMu.Lock()
	reloadCache[cc] = &ReloadCacheData{
		Collections: collections,
		Entities:    uniqueEntities,
		LastUpdated: time.Now(),
	}
	reloadCacheMu.Unlock()
	m.l.Printf("Finished background refresh for /reload page for country: %s. Fetched %d entities across %d collections.", cc, len(uniqueEntities), len(collections))
}

func (m *Module) EngineeringReloadList(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	cc := r.URL.Query().Get("cc")
	search := strings.ToLower(r.URL.Query().Get("search"))
	collectionFilter := r.URL.Query().Get("collection")
	canRefresh := r.URL.Query().Get("can_refresh") == "true"
	isRefreshing := r.URL.Query().Get("is_refreshing") == "true"

	if cc == "" {
		cc = "za"
	}

	reloadCacheMu.RLock()
	cached, exists := reloadCache[cc]
	reloadCacheMu.RUnlock()

	if !exists {
		return reloadLoadingPanel(cc).Render(ctx, w)
	}

	if collectionFilter == "" && len(cached.Collections) > 0 {
		collectionFilter = cached.Collections[0]
	}

	var filtered []CoreReload
	for _, entity := range cached.Entities {
		if collectionFilter != "" && entity.Database != collectionFilter {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(entity.Name), search) {
			continue
		}
		if canRefresh && !entity.HasGcsData {
			continue
		}
		if isRefreshing && !entity.RefreshInProgress {
			continue
		}
		filtered = append(filtered, entity)
	}

	return reloadTablesPanel(filtered, cached.Collections, collectionFilter, "").Render(ctx, w)
}

func (m *Module) EngineeringReloadToggleEnable(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()

	encodedKey := r.FormValue("key")
	enabled := r.FormValue("enabled") == "true"
	cc := r.FormValue("cc")
	if cc == "" {
		cc = "za" // Default fallback
	}

	key, err := datastore.DecodeKey(encodedKey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	projectID := getProjectForCountry(cc)
	client, err := getDatastoreClient(ctx, projectID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	tx, err := client.NewTransaction(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	var entity CoreReload
	if err := tx.Get(key, &entity); err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	entity.HasGcsData = enabled // Set the updated toggle value

	if _, err := tx.Put(key, &entity); err != nil {
		tx.Rollback()
		m.l.Printf("Error updating HasGcsData: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	if _, err := tx.Commit(); err != nil {
		m.l.Printf("Error committing HasGcsData update: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	// Sync local cache
	reloadCacheMu.Lock()
	if c, ok := reloadCache[cc]; ok {
		for i, e := range c.Entities {
			if e.EncodedKey == encodedKey {
				c.Entities[i].HasGcsData = enabled
				break
			}
		}
	}
	reloadCacheMu.Unlock()

	w.WriteHeader(http.StatusOK)
	return nil
}

func (m *Module) EngineeringReloadAction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	r.ParseForm()
	action := r.FormValue("action")
	table := r.FormValue("table")
	database := r.FormValue("database")
	cc := r.FormValue("cc")
	if cc == "" {
		cc = "za"
	}

	m.l.Printf("Reload Action triggered: %s on %s.%s (CC: %s)", action, database, table, cc)

	if action == "refresh_table" {
		execClient, err := executions.NewClient(ctx, option.WithEndpoint("workflowexecutions.googleapis.com:443"))
		if err != nil {
			return err
		}
		defer execClient.Close()

		workflowName := fmt.Sprintf("projects/%s/locations/%s/workflows/%s", getProjectForCountry(cc), "europe-west9", "manual-refresh-single-table")
		createReq := &executionspb.CreateExecutionRequest{
			Parent: workflowName,
			Execution: &executionspb.Execution{
				Argument: fmt.Sprintf(`{"table": "%s", "database": "%s"}`, table, database),
			},
		}

		execution, err := execClient.CreateExecution(ctx, createReq)
		if err != nil {
			m.l.Printf("ERROR: Failed to trigger workflow: %v", err)
			w.Write([]byte(`<div class="p-4 bg-red-50 text-red-700 rounded font-bold text-sm">Failed to trigger refresh. Check logs.</div>`))
			return nil
		}
		m.l.Printf("Created execution %q", execution.Name)

		reloadCacheMu.Lock()
		if cached, ok := reloadCache[cc]; ok {
			for i, entity := range cached.Entities {
				if entity.Name == table && entity.Database == database {
					m.l.Printf("Setting RefreshInProgress for %s.%s", database, table)
					cached.Entities[i].RefreshInProgress = true
					cached.Entities[i].ExecutionID = execution.Name

					encodedKey := cached.Entities[i].EncodedKey
					key, err := datastore.DecodeKey(encodedKey)
					if err == nil {
						projectID := getProjectForCountry(cc)
						dsClient, _ := getDatastoreClient(ctx, projectID)

						tx, err := dsClient.NewTransaction(ctx)
						if err != nil {
							m.l.Printf("Error starting transaction: %v", err)
							break
						}

						var dsEntity CoreReload
						if err := tx.Get(key, &dsEntity); err == nil {
							dsEntity.RefreshInProgress = true
							dsEntity.ExecutionID = execution.Name
							if _, err := tx.Put(key, &dsEntity); err != nil {
								m.l.Printf("Error putting entity: %v", err)
								tx.Rollback()
							}
							if _, err := tx.Commit(); err != nil {
								m.l.Printf("Error committing transaction: %v", err)
							}
						} else {
							tx.Rollback()
						}
					}
					break
				}
			}
		}
		reloadCacheMu.Unlock()
	}

	if action == "update_cron" {
		cronVal := r.FormValue("cron")
		encodedKey := r.FormValue("key")
		key, err := datastore.DecodeKey(encodedKey)
		if err == nil {
			projectID := getProjectForCountry(cc)
			dsClient, _ := getDatastoreClient(ctx, projectID)

			tx, err := dsClient.NewTransaction(ctx)
			if err != nil {
				m.l.Printf("Error starting transaction: %v", err)
			} else {
				var entity CoreReload
				if err := tx.Get(key, &entity); err == nil {
					entity.CronSchedule = cronVal
					if _, err := tx.Put(key, &entity); err != nil {
						tx.Rollback()
					}
					if _, err := tx.Commit(); err != nil {
						m.l.Printf("Error committing transaction: %v", err)
					}

					reloadCacheMu.Lock()
					if c, ok := reloadCache[cc]; ok {
						for i, e := range c.Entities {
							if e.EncodedKey == encodedKey {
								c.Entities[i].CronSchedule = cronVal
								break
							}
						}
					}
					reloadCacheMu.Unlock()
				} else {
					tx.Rollback()
				}
			}
		}
	}

	w.Header().Set("HX-Trigger", "reload-list")
	w.Write([]byte(fmt.Sprintf(`<div class="p-4 bg-green-50 text-green-700 rounded font-bold text-sm mb-4">Action %s triggered successfully!</div>`, action)))
	return nil
}
