package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *LocalStore) GetServices(ctx context.Context, serviceType string) ([]Service, error) {
	query := `SELECT id, service_type, description, standard_rate, duration_unit, updated_at, deleted, synced FROM services WHERE deleted = 0 AND service_type = ? ORDER BY description`
	rows, err := s.db.QueryContext(ctx, query, serviceType)
	if err != nil {
		return nil, fmt.Errorf("query services: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var svc Service
		var del, syn int
		if err := rows.Scan(&svc.ID, &svc.ServiceType, &svc.Description, &svc.StandardRate, &svc.DurationUnit, &svc.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		svc.Deleted = del == 1
		svc.Synced = syn == 1
		services = append(services, svc)
	}
	return services, rows.Err()
}

func (s *LocalStore) GetAllServices(ctx context.Context) ([]Service, error) {
	query := `SELECT id, service_type, description, standard_rate, duration_unit, updated_at, deleted, synced FROM services WHERE deleted = 0`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query all services: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var svc Service
		var del, syn int
		if err := rows.Scan(&svc.ID, &svc.ServiceType, &svc.Description, &svc.StandardRate, &svc.DurationUnit, &svc.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		svc.Deleted = del == 1
		svc.Synced = syn == 1
		services = append(services, svc)
	}
	return services, rows.Err()
}

func (s *LocalStore) SaveService(ctx context.Context, svc Service) error {
	if svc.ID == "" {
		svc.ID = uuid.New().String()
	}
	svc.UpdatedAt = time.Now().UTC()

	del := 0
	if svc.Deleted {
		del = 1
	}

	query := `
	INSERT INTO services (id, service_type, description, standard_rate, duration_unit, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		service_type = excluded.service_type,
		description = excluded.description,
		standard_rate = excluded.standard_rate,
		duration_unit = excluded.duration_unit,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err := s.db.ExecContext(ctx, query, svc.ID, svc.ServiceType, svc.Description, svc.StandardRate, svc.DurationUnit, svc.UpdatedAt, del)
	return err
}

func (s *LocalStore) DeleteService(ctx context.Context, id string) error {
	query := `UPDATE services SET deleted = 1, updated_at = ?, synced = 0 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}
