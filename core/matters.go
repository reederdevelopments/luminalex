package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *LocalStore) FetchLatestMatterReference(ctx context.Context, initials string) (string, error) {
	prefix := fmt.Sprintf("WW-%s-%%", initials)
	// We added AND deleted = 0 AND LENGTH(reference) < 15 to completely bypass the old timestamp anomaly
	query := `SELECT reference FROM matters WHERE reference LIKE ? AND deleted = 0 AND LENGTH(reference) < 15 ORDER BY reference DESC LIMIT 1`
	var ref string
	err := s.db.QueryRowContext(ctx, query, prefix).Scan(&ref)
	if err != nil {
		return "", err
	}
	return ref, nil
}

func (a *App) GenerateMatterReference(ctx context.Context) string {
	username := a.GetLastUsername()
	initials := "USR"
	if len(username) >= 2 {
		initials = strings.ToUpper(username[:2])
	}

	var lastRef string

	sbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	ref, err := a.supabase.FetchLatestMatterReference(sbCtx, initials)
	if err == nil && ref != "" {
		lastRef = ref
	} else {
		ref, _ = a.store.FetchLatestMatterReference(ctx, initials)
		if ref != "" {
			lastRef = ref
		}
	}

	seq := 1
	if lastRef != "" {
		parts := strings.Split(lastRef, "-")
		if len(parts) == 3 {
			var parsedSeq int
			fmt.Sscanf(parts[2], "%d", &parsedSeq)
			if parsedSeq < 1000000 {
				seq = parsedSeq + 1
			}
		}
	}

	return fmt.Sprintf("WW-%s-%05d", initials, seq)
}

func (s *LocalStore) GetMatters(ctx context.Context) ([]Matter, error) {
	query := `SELECT id, reference, client_id, status, matter_type, updated_at, deleted, synced FROM matters WHERE deleted = 0 ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query matters: %w", err)
	}
	defer rows.Close()

	var matters []Matter
	for rows.Next() {
		var m Matter
		var del, syn int
		if err := rows.Scan(&m.ID, &m.Reference, &m.ClientID, &m.Status, &m.MatterType, &m.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		m.Deleted = del == 1
		m.Synced = syn == 1
		matters = append(matters, m)
	}
	return matters, rows.Err()
}

func (s *LocalStore) SaveMatter(ctx context.Context, m Matter) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	m.UpdatedAt = time.Now().UTC()
	del := 0
	if m.Deleted {
		del = 1
	}

	query := `
	INSERT INTO matters (id, reference, client_id, status, matter_type, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		status = excluded.status,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, m.ID, m.Reference, m.ClientID, m.Status, m.MatterType, m.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteMatter(ctx context.Context, id string) error {
	query := `UPDATE matters SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (s *LocalStore) GetMatterNotes(ctx context.Context, matterID string) ([]MatterNote, error) {
	query := `SELECT id, matter_id, author, content, updated_at, deleted, synced FROM matter_notes WHERE matter_id = ? AND deleted = 0 ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query, matterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []MatterNote
	for rows.Next() {
		var n MatterNote
		var del, syn int
		if err := rows.Scan(&n.ID, &n.MatterID, &n.Author, &n.Content, &n.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		n.Deleted = del == 1
		n.Synced = syn == 1
		notes = append(notes, n)
	}
	return notes, nil
}

func (s *LocalStore) SaveMatterNote(ctx context.Context, n MatterNote) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	n.UpdatedAt = time.Now().UTC()
	del := 0
	if n.Deleted {
		del = 1
	}

	query := `
	INSERT INTO matter_notes (id, matter_id, author, content, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		content = excluded.content,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, n.ID, n.MatterID, n.Author, n.Content, n.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteMatterNote(ctx context.Context, id string) error {
	query := `UPDATE matter_notes SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (s *LocalStore) GetMatterServices(ctx context.Context, matterID string) ([]MatterService, error) {
	query := `SELECT id, matter_id, service_id, snapshot_desc, snapshot_rate, snapshot_unit, qty, add_tax, updated_at, deleted, synced FROM matter_services WHERE matter_id = ? AND deleted = 0 ORDER BY updated_at ASC`
	rows, err := s.db.QueryContext(ctx, query, matterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var svcs []MatterService
	for rows.Next() {
		var svc MatterService
		var addTax, del, syn int
		if err := rows.Scan(&svc.ID, &svc.MatterID, &svc.ServiceID, &svc.SnapshotDesc, &svc.SnapshotRate, &svc.SnapshotUnit, &svc.Qty, &addTax, &svc.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		svc.AddTax = addTax == 1
		svc.Deleted = del == 1
		svc.Synced = syn == 1
		svcs = append(svcs, svc)
	}
	return svcs, nil
}

func (s *LocalStore) SaveMatterService(ctx context.Context, svc MatterService) error {
	if svc.ID == "" {
		svc.ID = uuid.New().String()
	}
	svc.UpdatedAt = time.Now().UTC()
	addTax := 0
	if svc.AddTax {
		addTax = 1
	}
	del := 0
	if svc.Deleted {
		del = 1
	}

	query := `
	INSERT INTO matter_services (id, matter_id, service_id, snapshot_desc, snapshot_rate, snapshot_unit, qty, add_tax, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		snapshot_desc = excluded.snapshot_desc,
		snapshot_rate = excluded.snapshot_rate,
		snapshot_unit = excluded.snapshot_unit,
		qty = excluded.qty,
		add_tax = excluded.add_tax,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, svc.ID, svc.MatterID, svc.ServiceID, svc.SnapshotDesc, svc.SnapshotRate, svc.SnapshotUnit, svc.Qty, addTax, svc.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteMatterService(ctx context.Context, id string) error {
	query := `UPDATE matter_services SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (s *LocalStore) GetMattersByClient(ctx context.Context, clientID string) ([]Matter, error) {
	query := `SELECT id, reference, client_id, status, matter_type, updated_at, deleted, synced FROM matters WHERE deleted = 0 AND client_id = ? ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("query matters by client: %w", err)
	}
	defer rows.Close()

	var matters []Matter
	for rows.Next() {
		var m Matter
		var del, syn int
		if err := rows.Scan(&m.ID, &m.Reference, &m.ClientID, &m.Status, &m.MatterType, &m.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		m.Deleted = del == 1
		m.Synced = syn == 1
		matters = append(matters, m)
	}
	return matters, rows.Err()
}
