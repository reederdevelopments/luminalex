package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type LocalStore struct {
	db *sql.DB
}

func NewLocalStore() (*LocalStore, error) {
	appDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("user config dir: %w", err)
	}

	dataDir := filepath.Join(appDir, "LuminaLex")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "luminalex.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	store := &LocalStore{db: db}
	if err := store.initTables(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *LocalStore) initTables(ctx context.Context) error {
	categories := []string{"banks", "masters", "sheriffs", "magistrates", "highcourts", "lawfirms"}
	for _, cat := range categories {
		query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			fields TEXT NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted INTEGER NOT NULL DEFAULT 0,
			synced INTEGER NOT NULL DEFAULT 0
		);`, cat)
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("init table %s: %w", cat, err)
		}
	}
	return nil
}

func (s *LocalStore) GetCategoryRecords(ctx context.Context, category string) ([]ContactRecord, error) {
	query := fmt.Sprintf(`SELECT id, fields, updated_at, deleted, synced FROM %s WHERE deleted = 0`, category)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", category, err)
	}
	defer rows.Close()

	var records []ContactRecord
	for rows.Next() {
		var rec ContactRecord
		rec.Category = category
		var fieldsJSON string
		var updatedAtRaw time.Time
		var del, syn int

		if err := rows.Scan(&rec.ID, &fieldsJSON, &updatedAtRaw, &del, &syn); err != nil {
			return nil, fmt.Errorf("scan record: %w", err)
		}

		if err := json.Unmarshal([]byte(fieldsJSON), &rec.Fields); err != nil {
			return nil, fmt.Errorf("unmarshal fields: %w", err)
		}

		rec.UpdatedAt = updatedAtRaw
		rec.Deleted = del == 1
		rec.Synced = syn == 1
		records = append(records, rec)
	}

	return records, rows.Err()
}

func (s *LocalStore) SaveRecord(ctx context.Context, record ContactRecord) error {
	fieldsJSON, err := json.Marshal(record.Fields)
	if err != nil {
		return fmt.Errorf("marshal fields: %w", err)
	}

	del := 0
	if record.Deleted {
		del = 1
	}
	syn := 0
	if record.Synced {
		syn = 1
	}

	query := fmt.Sprintf(`
	INSERT INTO %s (id, fields, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		fields = excluded.fields,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = excluded.synced
	WHERE excluded.updated_at >= %s.updated_at;`, record.Category, record.Category)

	_, err = s.db.ExecContext(ctx, query, record.ID, string(fieldsJSON), record.UpdatedAt.UTC(), del, syn)
	if err != nil {
		return fmt.Errorf("upsert record in %s: %w", record.Category, err)
	}
	return nil
}

func (s *LocalStore) GetUnsyncedRecords(ctx context.Context, category string) ([]ContactRecord, error) {
	query := fmt.Sprintf(`SELECT id, fields, updated_at, deleted FROM %s WHERE synced = 0`, category)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query unsynced %s: %w", category, err)
	}
	defer rows.Close()

	var records []ContactRecord
	for rows.Next() {
		var rec ContactRecord
		rec.Category = category
		var fieldsJSON string
		var updatedAtRaw time.Time
		var del int

		if err := rows.Scan(&rec.ID, &fieldsJSON, &updatedAtRaw, &del); err != nil {
			return nil, fmt.Errorf("scan record: %w", err)
		}

		if err := json.Unmarshal([]byte(fieldsJSON), &rec.Fields); err != nil {
			return nil, fmt.Errorf("unmarshal fields: %w", err)
		}

		rec.UpdatedAt = updatedAtRaw
		rec.Deleted = del == 1
		rec.Synced = false
		records = append(records, rec)
	}

	return records, rows.Err()
}

func (s *LocalStore) MarkSynced(ctx context.Context, category string, id string) error {
	query := fmt.Sprintf(`UPDATE %s SET synced = 1 WHERE id = ?`, category)
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *LocalStore) Close() error {
	return s.db.Close()
}
