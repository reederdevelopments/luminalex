package core

import (
	"context"
	"fmt"
	"os"
	"time"
)

type App struct {
	ctx        context.Context
	store      *LocalStore
	supabase   *SupabaseClient
	syncEngine *SyncEngine
	updater    *AutoUpdater
}

func NewApp() *App {
	store, err := NewLocalStore()
	if err != nil {
		fmt.Printf("Error initializing local store: %v\n", err)
		os.Exit(1)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	sbClient := NewSupabaseClient(supabaseURL, supabaseKey)

	syncEngine := NewSyncEngine(store, sbClient)

	owner := os.Getenv("GITHUB_REPO_OWNER")
	repo := os.Getenv("GITHUB_REPO_NAME")
	version := os.Getenv("APP_CURRENT_VERSION")
	if version == "" {
		version = "v1.0.0"
	}
	updater := NewAutoUpdater(owner, repo, version)

	return &App{
		store:      store,
		supabase:   sbClient,
		syncEngine: syncEngine,
		updater:    updater,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	go a.startBackgroundSync()
}

func (a *App) startBackgroundSync() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	_ = a.syncEngine.PerformSync(a.ctx)

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			_ = a.syncEngine.PerformSync(a.ctx)
		}
	}
}

func (a *App) GetContacts(category string) ([]ContactRecord, error) {
	return a.store.GetCategoryRecords(a.ctx, category)
}

func (a *App) SaveContact(record ContactRecord) error {
	record.UpdatedAt = time.Now().UTC()
	record.Synced = false
	if err := a.store.SaveRecord(a.ctx, record); err != nil {
		return err
	}
	go func() {
		_ = a.syncEngine.PerformSync(context.Background())
	}()
	return nil
}

func (a *App) DeleteContact(category string, id string) error {
	records, err := a.store.GetCategoryRecords(a.ctx, category)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.ID == id {
			rec.Deleted = true
			rec.UpdatedAt = time.Now().UTC()
			rec.Synced = false
			if err := a.store.SaveRecord(a.ctx, rec); err != nil {
				return err
			}
			break
		}
	}

	go func() {
		_ = a.syncEngine.PerformSync(context.Background())
	}()
	return nil
}

func (a *App) TriggerSync() SyncStatus {
	_ = a.syncEngine.PerformSync(a.ctx)
	return a.syncEngine.GetStatus()
}

func (a *App) CheckUpdate() (*UpdateCheckResult, error) {
	return a.updater.CheckForUpdates(a.ctx)
}

func (a *App) PerformUpdate(downloadURL string) error {
	return a.updater.ApplyUpdate(a.ctx, downloadURL)
}
