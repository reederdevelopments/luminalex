package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SyncEngine struct {
	store      *LocalStore
	client     *SupabaseClient
	categories []string
	lastSync   time.Time
	mu         sync.Mutex
	isSyncing  bool
}

func NewSyncEngine(store *LocalStore, client *SupabaseClient) *SyncEngine {
	return &SyncEngine{
		store:      store,
		client:     client,
		categories: []string{"banks", "masters", "sheriffs", "magistrates", "highcourts", "lawfirms"},
		lastSync:   time.Time{},
	}
}

func (e *SyncEngine) PerformSync(ctx context.Context) error {
	e.mu.Lock()
	if e.isSyncing {
		e.mu.Unlock()
		return nil
	}
	e.isSyncing = true
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.isSyncing = false
		e.mu.Unlock()
	}()

	for _, cat := range e.categories {
		unsynced, err := e.store.GetUnsyncedRecords(ctx, cat)
		if err != nil {
			return fmt.Errorf("fetch unsynced for %s: %w", cat, err)
		}

		for _, rec := range unsynced {
			if err := e.client.UpsertRecord(ctx, rec); err != nil {
				return fmt.Errorf("push record %s: %w", rec.ID, err)
			}
			if err := e.store.MarkSynced(ctx, cat, rec.ID); err != nil {
				return fmt.Errorf("mark synced %s: %w", rec.ID, err)
			}
		}

		remoteRecords, err := e.client.FetchUpdatedAfter(ctx, cat, e.lastSync)
		if err != nil {
			return fmt.Errorf("fetch remote for %s: %w", cat, err)
		}

		for _, remoteRec := range remoteRecords {
			if err := e.store.SaveRecord(ctx, remoteRec); err != nil {
				return fmt.Errorf("save remote record %s: %w", remoteRec.ID, err)
			}
		}
	}

	e.lastSync = time.Now().UTC()
	return nil
}

func (e *SyncEngine) GetStatus() SyncStatus {
	e.mu.Lock()
	defer e.mu.Unlock()

	lastSyncStr := "Never"
	if !e.lastSync.IsZero() {
		lastSyncStr = e.lastSync.Format("2006-01-02 15:04:05")
	}

	return SyncStatus{
		IsSyncing: e.isSyncing,
		LastSync:  lastSyncStr,
	}
}
