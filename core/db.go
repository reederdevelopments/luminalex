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
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
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

	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		id TEXT PRIMARY KEY,
		first_name TEXT NOT NULL,
		middle_name TEXT,
		last_name TEXT NOT NULL,
		id_number TEXT,
		jurisdiction_type TEXT NOT NULL,
		jurisdiction_other TEXT,
		registration_number TEXT,
		occupation TEXT,
		marital_status TEXT,
		employer_name TEXT,
		employer_number TEXT,
		employer_address_l1 TEXT,
		employer_address_l2 TEXT,
		employer_suburb TEXT,
		employer_city TEXT,
		employer_postal_code TEXT,
		employer_country TEXT,
		employer_postal_same INTEGER NOT NULL DEFAULT 1,
		employer_postal_l1 TEXT,
		employer_postal_l2 TEXT,
		employer_postal_suburb TEXT,
		employer_postal_city TEXT,
		employer_postal_code2 TEXT,
		employer_postal_country TEXT,
		practice_number TEXT,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS client_addresses (
		id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		address_type TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 0,
		line1 TEXT NOT NULL,
		line2 TEXT,
		suburb TEXT,
		city TEXT NOT NULL,
		postal_code TEXT NOT NULL,
		country TEXT NOT NULL,
		postal_same INTEGER NOT NULL DEFAULT 1,
		postal_line1 TEXT,
		postal_line2 TEXT,
		postal_suburb TEXT,
		postal_city TEXT,
		postal_code2 TEXT,
		postal_country TEXT,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS client_contact_details (
		id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		contact_type TEXT NOT NULL,
		contact_value TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS client_banks (
		id TEXT PRIMARY KEY,
		client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		bank_name TEXT NOT NULL,
		branch_code TEXT NOT NULL,
		account_number TEXT NOT NULL,
		account_type TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);`

	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return err
	}

	servicesSchema := `
	CREATE TABLE IF NOT EXISTS services (
		id TEXT PRIMARY KEY,
		service_type TEXT NOT NULL,
		description TEXT NOT NULL,
		standard_rate REAL NOT NULL,
		duration_unit TEXT NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);`

	if _, err := s.db.ExecContext(ctx, servicesSchema); err != nil {
		return fmt.Errorf("init table services: %w", err)
	}

	mattersSchema := `
	CREATE TABLE IF NOT EXISTS matters (
		id TEXT PRIMARY KEY,
		reference TEXT NOT NULL,
		client_id TEXT NOT NULL,
		status TEXT NOT NULL,
		matter_type TEXT NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);
	
	CREATE TABLE IF NOT EXISTS matter_notes (
		id TEXT PRIMARY KEY,
		matter_id TEXT NOT NULL,
		author TEXT NOT NULL,
		content TEXT NOT NULL,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);
	
	CREATE TABLE IF NOT EXISTS matter_services (
		id TEXT PRIMARY KEY,
		matter_id TEXT NOT NULL,
		service_id TEXT NOT NULL,
		snapshot_desc TEXT NOT NULL,
		snapshot_rate REAL NOT NULL,
		snapshot_unit TEXT NOT NULL,
		qty REAL NOT NULL,
		add_tax INTEGER NOT NULL DEFAULT 1,
		updated_at DATETIME NOT NULL,
		deleted INTEGER NOT NULL DEFAULT 0,
		synced INTEGER NOT NULL DEFAULT 0
	);`

	if _, err := s.db.ExecContext(ctx, mattersSchema); err != nil {
		return fmt.Errorf("init table matters: %w", err)
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
			continue
		}

		if err := json.Unmarshal([]byte(fieldsJSON), &rec.Fields); err != nil {
			rec.Fields = []string{"[Data Formatting Error]"}
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
	return err
}

func (s *LocalStore) SaveRecordsBatch(ctx context.Context, category string, records []ContactRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
	INSERT INTO %s (id, fields, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		fields = excluded.fields,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = excluded.synced
	WHERE excluded.updated_at >= %s.updated_at;`, category, category)

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		fieldsJSON, _ := json.Marshal(record.Fields)
		del := 0
		if record.Deleted {
			del = 1
		}
		syn := 0
		if record.Synced {
			syn = 1
		}

		if _, err = stmt.ExecContext(ctx, record.ID, string(fieldsJSON), record.UpdatedAt.UTC(), del, syn); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *LocalStore) GetUnsyncedRecords(ctx context.Context, category string) ([]ContactRecord, error) {
	query := fmt.Sprintf(`SELECT id, fields, updated_at, deleted FROM %s WHERE synced = 0`, category)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
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
			continue
		}

		_ = json.Unmarshal([]byte(fieldsJSON), &rec.Fields)
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
