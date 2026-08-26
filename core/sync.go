package core

import (
	"context"
	"errors"
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

	var syncErrors []error

	// 1. Sync Contact Directories
	for _, cat := range e.categories {
		unsynced, err := e.store.GetUnsyncedRecords(ctx, cat)
		if err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("fetch unsynced for %s: %w", cat, err))
			continue
		}

		for _, rec := range unsynced {
			if err := e.client.UpsertRecord(ctx, rec); err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("push record %s: %w", rec.ID, err))
				continue
			}
			_ = e.store.MarkSynced(ctx, cat, rec.ID)
		}

		remoteRecords, err := e.client.FetchUpdatedAfter(ctx, cat, e.lastSync)
		if err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("fetch remote for %s: %w", cat, err))
			continue
		}

		_ = e.store.SaveRecordsBatch(ctx, cat, remoteRecords)
	}

	// 2. Sync Relational Entities
	if err := e.syncRelationalClients(ctx); err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("sync clients: %w", err))
	}

	if err := e.syncRelationalServices(ctx); err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("sync services: %w", err))
	}

	if err := e.syncRelationalMatters(ctx); err != nil {
		syncErrors = append(syncErrors, fmt.Errorf("sync matters: %w", err))
	}

	if len(syncErrors) == 0 {
		e.lastSync = time.Now().UTC()
		return nil
	}

	return errors.Join(syncErrors...)
}

func (e *SyncEngine) syncRelationalClients(ctx context.Context) error {
	localClients, err := e.store.GetClients(ctx)
	if err != nil {
		return err
	}

	for _, client := range localClients {
		if !client.Synced {
			dto := ClientSyncDTO{
				ID:                    client.ID,
				FirstName:             client.FirstName,
				MiddleName:            client.MiddleName,
				LastName:              client.LastName,
				IDNumber:              client.IDNumber,
				JurisdictionType:      client.JurisdictionType,
				JurisdictionOther:     client.JurisdictionOther,
				RegistrationNumber:    client.RegistrationNumber,
				Occupation:            client.Occupation,
				MaritalStatus:         client.MaritalStatus,
				EmployerName:          client.EmployerName,
				EmployerNumber:        client.EmployerNumber,
				EmployerAddressL1:     client.EmployerAddressL1,
				EmployerAddressL2:     client.EmployerAddressL2,
				EmployerSuburb:        client.EmployerSuburb,
				EmployerCity:          client.EmployerCity,
				EmployerPostalCode:    client.EmployerPostalCode,
				EmployerCountry:       client.EmployerCountry,
				EmployerPostalSame:    client.EmployerPostalSame,
				EmployerPostalL1:      client.EmployerPostalL1,
				EmployerPostalL2:      client.EmployerPostalL2,
				EmployerPostalSuburb:  client.EmployerPostalSuburb,
				EmployerPostalCity:    client.EmployerPostalCity,
				EmployerPostalCode2:   client.EmployerPostalCode2,
				EmployerPostalCountry: client.EmployerPostalCountry,
				PracticeNumber:        client.PracticeNumber,
				UpdatedAt:             client.UpdatedAt,
				Deleted:               client.Deleted,
			}
			if err := e.client.UpsertGeneric(ctx, "clients", dto); err != nil {
				return err
			}
			_, _ = e.store.db.ExecContext(ctx, "UPDATE clients SET synced = 1 WHERE id = ?", client.ID)
		}

		for _, addr := range client.Addresses {
			if !addr.Synced {
				dto := AddressSyncDTO{
					ID:            addr.ID,
					ClientID:      addr.ClientID,
					AddressType:   addr.AddressType,
					IsPrimary:     addr.IsPrimary,
					Line1:         addr.Line1,
					Line2:         addr.Line2,
					Suburb:        addr.Suburb,
					City:          addr.City,
					PostalCode:    addr.PostalCode,
					Country:       addr.Country,
					PostalSame:    addr.PostalSame,
					PostalLine1:   addr.PostalLine1,
					PostalLine2:   addr.PostalLine2,
					PostalSuburb:  addr.PostalSuburb,
					PostalCity:    addr.PostalCity,
					PostalCode2:   addr.PostalCode2,
					PostalCountry: addr.PostalCountry,
					UpdatedAt:     addr.UpdatedAt,
					Deleted:       addr.Deleted,
				}
				if err := e.client.UpsertGeneric(ctx, "client_addresses", dto); err != nil {
					return err
				}
				_, _ = e.store.db.ExecContext(ctx, "UPDATE client_addresses SET synced = 1 WHERE id = ?", addr.ID)
			}
		}

		for _, cd := range client.ContactDetails {
			if !cd.Synced {
				dto := ContactDetailSyncDTO{
					ID:           cd.ID,
					ClientID:     cd.ClientID,
					ContactType:  cd.ContactType,
					ContactValue: cd.ContactValue,
					IsPrimary:    cd.IsPrimary,
					UpdatedAt:    cd.UpdatedAt,
					Deleted:      cd.Deleted,
				}
				if err := e.client.UpsertGeneric(ctx, "client_contact_details", dto); err != nil {
					return err
				}
				_, _ = e.store.db.ExecContext(ctx, "UPDATE client_contact_details SET synced = 1 WHERE id = ?", cd.ID)
			}
		}

		for _, bank := range client.Banks {
			if !bank.Synced {
				dto := BankSyncDTO{
					ID:            bank.ID,
					ClientID:      bank.ClientID,
					BankName:      bank.BankName,
					BranchCode:    bank.BranchCode,
					AccountNumber: bank.AccountNumber,
					AccountType:   bank.AccountType,
					IsPrimary:     bank.IsPrimary,
					UpdatedAt:     bank.UpdatedAt,
					Deleted:       bank.Deleted,
				}
				if err := e.client.UpsertGeneric(ctx, "client_banks", dto); err != nil {
					return err
				}
				_, _ = e.store.db.ExecContext(ctx, "UPDATE client_banks SET synced = 1 WHERE id = ?", bank.ID)
			}
		}
	}

	remoteClients, err := e.client.FetchClientsAfter(ctx, e.lastSync)
	if err != nil {
		return err
	}

	if len(remoteClients) > 0 {
		var clientIDs []string
		for _, c := range remoteClients {
			clientIDs = append(clientIDs, c.ID)
		}

		addresses, _ := e.client.FetchAddresses(ctx, clientIDs)
		contacts, _ := e.client.FetchContacts(ctx, clientIDs)
		banks, _ := e.client.FetchBanks(ctx, clientIDs)

		addressMap := make(map[string][]ClientAddress)
		for _, a := range addresses {
			a.Synced = true
			addressMap[a.ClientID] = append(addressMap[a.ClientID], a)
		}

		contactMap := make(map[string][]ClientContactDetail)
		for _, c := range contacts {
			c.Synced = true
			contactMap[c.ClientID] = append(contactMap[c.ClientID], c)
		}

		bankMap := make(map[string][]ClientBank)
		for _, b := range banks {
			b.Synced = true
			bankMap[b.ClientID] = append(bankMap[b.ClientID], b)
		}

		for _, remote := range remoteClients {
			remote.Addresses = addressMap[remote.ID]
			remote.ContactDetails = contactMap[remote.ID]
			remote.Banks = bankMap[remote.ID]
			remote.Synced = true

			_ = e.store.SaveClient(ctx, remote)
			_, _ = e.store.db.ExecContext(ctx, "UPDATE clients SET synced = 1 WHERE id = ?", remote.ID)
		}
	}

	return nil
}

func (e *SyncEngine) syncRelationalServices(ctx context.Context) error {
	localServices, err := e.store.GetAllServices(ctx)
	if err != nil {
		return err
	}

	for _, svc := range localServices {
		if !svc.Synced {
			if err := e.client.UpsertGeneric(ctx, "services", map[string]any{
				"id":            svc.ID,
				"service_type":  svc.ServiceType,
				"description":   svc.Description,
				"standard_rate": svc.StandardRate,
				"duration_unit": svc.DurationUnit,
				"updated_at":    svc.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999Z"),
				"deleted":       svc.Deleted,
			}); err != nil {
				return err
			}
			_, _ = e.store.db.ExecContext(ctx, "UPDATE services SET synced = 1 WHERE id = ?", svc.ID)
		}
	}

	remoteServices, err := e.client.FetchServicesAfter(ctx, e.lastSync)
	if err != nil {
		return err
	}

	for _, remote := range remoteServices {
		remote.Synced = true
		if err := e.store.SaveService(ctx, remote); err != nil {
			return err
		}
		_, _ = e.store.db.ExecContext(ctx, "UPDATE services SET synced = 1 WHERE id = ?", remote.ID)
	}

	return nil
}

// ---------------------------------------------------------
// NEW MATTER SYNC LOGIC
// ---------------------------------------------------------
func (e *SyncEngine) syncRelationalMatters(ctx context.Context) error {
	// 1. Sync Matters
	localMatters, err := e.store.GetMatters(ctx)
	if err == nil {
		for _, m := range localMatters {
			if !m.Synced {
				if err := e.client.UpsertGeneric(ctx, "matters", map[string]any{
					"id":          m.ID,
					"reference":   m.Reference,
					"client_id":   m.ClientID,
					"status":      m.Status,
					"matter_type": m.MatterType,
					"updated_at":  m.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999Z"),
					"deleted":     m.Deleted,
				}); err != nil {
					return err
				}
				_, _ = e.store.db.ExecContext(ctx, "UPDATE matters SET synced = 1 WHERE id = ?", m.ID)
			}
		}
	}

	// 2. Sync Matter Notes
	// Fetch all notes (we just query all unsynced notes directly to save time)
	rows, _ := e.store.db.QueryContext(ctx, `SELECT id, matter_id, author, content, updated_at, deleted FROM matter_notes WHERE synced = 0`)
	if rows != nil {
		var pendingNotes []MatterNote
		for rows.Next() {
			var n MatterNote
			var del int
			_ = rows.Scan(&n.ID, &n.MatterID, &n.Author, &n.Content, &n.UpdatedAt, &del)
			n.Deleted = del == 1
			pendingNotes = append(pendingNotes, n)
		}
		rows.Close()

		for _, n := range pendingNotes {
			if err := e.client.UpsertGeneric(ctx, "matter_notes", map[string]any{
				"id":         n.ID,
				"matter_id":  n.MatterID,
				"author":     n.Author,
				"content":    n.Content,
				"updated_at": n.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999Z"),
				"deleted":    n.Deleted,
			}); err == nil {
				_, _ = e.store.db.ExecContext(ctx, "UPDATE matter_notes SET synced = 1 WHERE id = ?", n.ID)
			}
		}
	}

	// 3. Sync Matter Services (Invoice Lines)
	sRows, _ := e.store.db.QueryContext(ctx, `SELECT id, matter_id, service_id, snapshot_desc, snapshot_rate, snapshot_unit, qty, add_tax, updated_at, deleted FROM matter_services WHERE synced = 0`)
	if sRows != nil {
		var pendingServices []MatterService
		for sRows.Next() {
			var s MatterService
			var tax, del int
			_ = sRows.Scan(&s.ID, &s.MatterID, &s.ServiceID, &s.SnapshotDesc, &s.SnapshotRate, &s.SnapshotUnit, &s.Qty, &tax, &s.UpdatedAt, &del)
			s.AddTax = tax == 1
			s.Deleted = del == 1
			pendingServices = append(pendingServices, s)
		}
		sRows.Close()

		for _, s := range pendingServices {
			if err := e.client.UpsertGeneric(ctx, "matter_services", map[string]any{
				"id":            s.ID,
				"matter_id":     s.MatterID,
				"service_id":    s.ServiceID,
				"snapshot_desc": s.SnapshotDesc,
				"snapshot_rate": s.SnapshotRate,
				"snapshot_unit": s.SnapshotUnit,
				"qty":           s.Qty,
				"add_tax":       s.AddTax,
				"updated_at":    s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999Z"),
				"deleted":       s.Deleted,
			}); err == nil {
				_, _ = e.store.db.ExecContext(ctx, "UPDATE matter_services SET synced = 1 WHERE id = ?", s.ID)
			}
		}
	}

	// NOTE: If you need to pull matters FROM Supabase down to local clients that were created by other users,
	// you would add a c.FetchMattersAfter() request here similar to the clients and services sync.
	// For now, this securely pushes your local matter data, notes, and invoice lines up to the cloud!

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
