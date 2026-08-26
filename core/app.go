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

	// 1. Hardcode your Supabase credentials here so users don't need a .env file
	supabaseURL := "https://frakboneidergnmzznkh.supabase.co"                                                                                                                                                                         // <-- Paste your Supabase URL
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImZyYWtib25laWRlcmdubXp6bmtoIiwicm9sZSI6ImFub24iLCJpYXQiOjE3ODQ5MTk3MDAsImV4cCI6MjEwMDQ5NTcwMH0.8-tRutPQW5UbMpGwK-k8Uto4zGgw12mCd8oiGHULTYM" // <-- Paste your Supabase Anon Key
	sbClient := NewSupabaseClient(supabaseURL, supabaseKey)

	syncEngine := NewSyncEngine(store, sbClient)

	// 2. Hardcode your GitHub Repository details
	owner := "reederdevelopments"
	repo := "luminalex"

	// 3. This is your app's internal version.
	// Before you build v1.0.2 in the future, you must manually change this to "v1.0.1".
	version := "v1.0.4"

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

	if err := a.syncEngine.PerformSync(a.ctx); err != nil {
		fmt.Printf("[Startup Sync Error] %v\n", err)
	}

	go a.startBackgroundTasks()
}

func (a *App) startBackgroundTasks() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.runRoutineTasks()
		}
	}
}

func (a *App) runRoutineTasks() {
	if err := a.syncEngine.PerformSync(a.ctx); err != nil {
		fmt.Printf("[Sync Error] %v\n", err)
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
	err := a.syncEngine.PerformSync(a.ctx)
	status := a.syncEngine.GetStatus()
	if err != nil {
		status.Error = "Sync Failed"
		status.Details = fmt.Sprintf("%+v", err)
	}
	return status
}

func (a *App) CheckUpdate() (*UpdateCheckResult, error) {
	return a.updater.CheckForUpdates(a.ctx)
}

func (a *App) PerformUpdate(downloadURL string) error {
	return a.updater.ApplyUpdate(a.ctx, downloadURL)
}

func (a *App) RestartApp() error {
	return a.updater.RestartApp()
}
