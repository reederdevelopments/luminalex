package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *LocalStore) GetClients(ctx context.Context) ([]Client, error) {
	query := `SELECT id, first_name, middle_name, last_name, id_number, jurisdiction_type, jurisdiction_other, registration_number, occupation, marital_status, employer_name, employer_number, employer_address_l1, employer_address_l2, employer_suburb, employer_city, employer_postal_code, employer_country, employer_postal_same, employer_postal_l1, employer_postal_l2, employer_postal_suburb, employer_postal_city, employer_postal_code2, employer_postal_country, practice_number, updated_at, deleted, synced FROM clients WHERE deleted = 0 ORDER BY last_name, first_name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query clients: %w", err)
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var c Client
		var empPostalSame, del, syn int
		if err := rows.Scan(&c.ID, &c.FirstName, &c.MiddleName, &c.LastName, &c.IDNumber, &c.JurisdictionType, &c.JurisdictionOther, &c.RegistrationNumber, &c.Occupation, &c.MaritalStatus, &c.EmployerName, &c.EmployerNumber, &c.EmployerAddressL1, &c.EmployerAddressL2, &c.EmployerSuburb, &c.EmployerCity, &c.EmployerPostalCode, &c.EmployerCountry, &empPostalSame, &c.EmployerPostalL1, &c.EmployerPostalL2, &c.EmployerPostalSuburb, &c.EmployerPostalCity, &c.EmployerPostalCode2, &c.EmployerPostalCountry, &c.PracticeNumber, &c.UpdatedAt, &del, &syn); err != nil {
			continue
		}
		c.EmployerPostalSame = empPostalSame == 1
		c.Deleted = del == 1
		c.Synced = syn == 1

		c.Addresses, _ = s.getClientAddresses(ctx, c.ID)
		c.ContactDetails, _ = s.getClientContactDetails(ctx, c.ID)
		c.Banks, _ = s.getClientBanks(ctx, c.ID)

		clients = append(clients, c)
	}

	return clients, rows.Err()
}

func (s *LocalStore) SaveClient(ctx context.Context, client Client) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if client.ID == "" {
		client.ID = uuid.New().String()
	}
	client.UpdatedAt = time.Now().UTC()

	del := 0
	if client.Deleted {
		del = 1
	}
	empPostalSame := 0
	if client.EmployerPostalSame {
		empPostalSame = 1
	}

	clientQuery := `
	INSERT INTO clients (id, first_name, middle_name, last_name, id_number, jurisdiction_type, jurisdiction_other, registration_number, occupation, marital_status, employer_name, employer_number, employer_address_l1, employer_address_l2, employer_suburb, employer_city, employer_postal_code, employer_country, employer_postal_same, employer_postal_l1, employer_postal_l2, employer_postal_suburb, employer_postal_city, employer_postal_code2, employer_postal_country, practice_number, updated_at, deleted, synced)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	ON CONFLICT(id) DO UPDATE SET
		first_name = excluded.first_name,
		middle_name = excluded.middle_name,
		last_name = excluded.last_name,
		id_number = excluded.id_number,
		jurisdiction_type = excluded.jurisdiction_type,
		jurisdiction_other = excluded.jurisdiction_other,
		registration_number = excluded.registration_number,
		occupation = excluded.occupation,
		marital_status = excluded.marital_status,
		employer_name = excluded.employer_name,
		employer_number = excluded.employer_number,
		employer_address_l1 = excluded.employer_address_l1,
		employer_address_l2 = excluded.employer_address_l2,
		employer_suburb = excluded.employer_suburb,
		employer_city = excluded.employer_city,
		employer_postal_code = excluded.employer_postal_code,
		employer_country = excluded.employer_country,
		employer_postal_same = excluded.employer_postal_same,
		employer_postal_l1 = excluded.employer_postal_l1,
		employer_postal_l2 = excluded.employer_postal_l2,
		employer_postal_suburb = excluded.employer_postal_suburb,
		employer_postal_city = excluded.employer_postal_city,
		employer_postal_code2 = excluded.employer_postal_code2,
		employer_postal_country = excluded.employer_postal_country,
		practice_number = excluded.practice_number,
		updated_at = excluded.updated_at,
		deleted = excluded.deleted,
		synced = 0;`

	_, err = tx.ExecContext(ctx, clientQuery, client.ID, client.FirstName, client.MiddleName, client.LastName, client.IDNumber, client.JurisdictionType, client.JurisdictionOther, client.RegistrationNumber, client.Occupation, client.MaritalStatus, client.EmployerName, client.EmployerNumber, client.EmployerAddressL1, client.EmployerAddressL2, client.EmployerSuburb, client.EmployerCity, client.EmployerPostalCode, client.EmployerCountry, empPostalSame, client.EmployerPostalL1, client.EmployerPostalL2, client.EmployerPostalSuburb, client.EmployerPostalCity, client.EmployerPostalCode2, client.EmployerPostalCountry, client.PracticeNumber, client.UpdatedAt, del)
	if err != nil {
		return fmt.Errorf("upsert client: %w", err)
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM client_addresses WHERE client_id = ?", client.ID)
	for _, addr := range client.Addresses {
		if addr.ID == "" {
			addr.ID = uuid.New().String()
		}
		isPrim := 0
		if addr.IsPrimary {
			isPrim = 1
		}
		posSame := 0
		if addr.PostalSame {
			posSame = 1
		}
		addrQuery := `
		INSERT INTO client_addresses (id, client_id, address_type, is_primary, line1, line2, suburb, city, postal_code, country, postal_same, postal_line1, postal_line2, postal_suburb, postal_city, postal_code2, postal_country, updated_at, deleted, synced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0)`
		if _, err := tx.ExecContext(ctx, addrQuery, addr.ID, client.ID, addr.AddressType, isPrim, addr.Line1, addr.Line2, addr.Suburb, addr.City, addr.PostalCode, addr.Country, posSame, addr.PostalLine1, addr.PostalLine2, addr.PostalSuburb, addr.PostalCity, addr.PostalCode2, addr.PostalCountry, client.UpdatedAt); err != nil {
			return fmt.Errorf("insert client address: %w", err)
		}
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM client_contact_details WHERE client_id = ?", client.ID)
	for _, cd := range client.ContactDetails {
		if cd.ID == "" {
			cd.ID = uuid.New().String()
		}
		isPrim := 0
		if cd.IsPrimary {
			isPrim = 1
		}
		contactQuery := `
		INSERT INTO client_contact_details (id, client_id, contact_type, contact_value, is_primary, updated_at, deleted, synced)
		VALUES (?, ?, ?, ?, ?, ?, 0, 0)`
		if _, err := tx.ExecContext(ctx, contactQuery, cd.ID, client.ID, cd.ContactType, cd.ContactValue, isPrim, client.UpdatedAt); err != nil {
			return fmt.Errorf("insert client contact: %w", err)
		}
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM client_banks WHERE client_id = ?", client.ID)
	for _, bank := range client.Banks {
		if bank.ID == "" {
			bank.ID = uuid.New().String()
		}
		isPrim := 0
		if bank.IsPrimary {
			isPrim = 1
		}
		bankQuery := `
		INSERT INTO client_banks (id, client_id, bank_name, branch_code, account_number, account_type, is_primary, updated_at, deleted, synced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, 0)`
		if _, err := tx.ExecContext(ctx, bankQuery, bank.ID, client.ID, bank.BankName, bank.BranchCode, bank.AccountNumber, bank.AccountType, isPrim, client.UpdatedAt); err != nil {
			return fmt.Errorf("insert client bank: %w", err)
		}
	}

	return tx.Commit()
}

func (s *LocalStore) getClientAddresses(ctx context.Context, clientID string) ([]ClientAddress, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, client_id, address_type, is_primary, line1, line2, suburb, city, postal_code, country, postal_same, postal_line1, postal_line2, postal_suburb, postal_city, postal_code2, postal_country, updated_at, deleted, synced FROM client_addresses WHERE client_id = ? AND deleted = 0`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ClientAddress
	for rows.Next() {
		var a ClientAddress
		var isPrim, posSame, del, syn int
		_ = rows.Scan(&a.ID, &a.ClientID, &a.AddressType, &isPrim, &a.Line1, &a.Line2, &a.Suburb, &a.City, &a.PostalCode, &a.Country, &posSame, &a.PostalLine1, &a.PostalLine2, &a.PostalSuburb, &a.PostalCity, &a.PostalCode2, &a.PostalCountry, &a.UpdatedAt, &del, &syn)
		a.IsPrimary = isPrim == 1
		a.PostalSame = posSame == 1
		a.Deleted = del == 1
		a.Synced = syn == 1
		res = append(res, a)
	}
	return res, nil
}

func (s *LocalStore) getClientContactDetails(ctx context.Context, clientID string) ([]ClientContactDetail, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, client_id, contact_type, contact_value, is_primary, updated_at, deleted, synced FROM client_contact_details WHERE client_id = ? AND deleted = 0`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ClientContactDetail
	for rows.Next() {
		var cd ClientContactDetail
		var isPrim, del, syn int
		_ = rows.Scan(&cd.ID, &cd.ClientID, &cd.ContactType, &cd.ContactValue, &isPrim, &cd.UpdatedAt, &del, &syn)
		cd.IsPrimary = isPrim == 1
		cd.Deleted = del == 1
		cd.Synced = syn == 1
		res = append(res, cd)
	}
	return res, nil
}

func (s *LocalStore) getClientBanks(ctx context.Context, clientID string) ([]ClientBank, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, client_id, bank_name, branch_code, account_number, account_type, is_primary, updated_at, deleted, synced FROM client_banks WHERE client_id = ? AND deleted = 0`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ClientBank
	for rows.Next() {
		var b ClientBank
		var isPrim, del, syn int
		_ = rows.Scan(&b.ID, &b.ClientID, &b.BankName, &b.BranchCode, &b.AccountNumber, &b.AccountType, &isPrim, &b.UpdatedAt, &del, &syn)
		b.IsPrimary = isPrim == 1
		b.Deleted = del == 1
		b.Synced = syn == 1
		res = append(res, b)
	}
	return res, nil
}
