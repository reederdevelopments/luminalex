package core

import "time"

type ContactRecord struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Fields    []string  `json:"fields"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	Synced    bool      `json:"synced"`
}

type UpdateCheckResult struct {
	HasUpdate    bool   `json:"has_update"`
	LatestVer    string `json:"latest_ver"`
	ReleaseNotes string `json:"release_notes"`
	DownloadURL  string `json:"download_url"`
}

type SyncStatus struct {
	IsSyncing bool   `json:"is_syncing"`
	LastSync  string `json:"last_sync"`
	Error     string `json:"error"`
	Details   string `json:"details"`
}

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Enabled      bool   `json:"enabled"`
}

type Client struct {
	ID                    string                `json:"id"`
	FirstName             string                `json:"first_name"`
	MiddleName            string                `json:"middle_name"`
	LastName              string                `json:"last_name"`
	IDNumber              string                `json:"id_number"`
	JurisdictionType      string                `json:"jurisdiction_type"`
	JurisdictionOther     string                `json:"jurisdiction_other"`
	RegistrationNumber    string                `json:"registration_number"`
	Occupation            string                `json:"occupation"`
	MaritalStatus         string                `json:"marital_status"`
	EmployerName          string                `json:"employer_name"`
	EmployerNumber        string                `json:"employer_number"`
	EmployerAddressL1     string                `json:"employer_address_l1"`
	EmployerAddressL2     string                `json:"employer_address_l2"`
	EmployerSuburb        string                `json:"employer_suburb"`
	EmployerCity          string                `json:"employer_city"`
	EmployerPostalCode    string                `json:"employer_postal_code"`
	EmployerCountry       string                `json:"employer_country"`
	EmployerPostalSame    bool                  `json:"employer_postal_same"`
	EmployerPostalL1      string                `json:"employer_postal_l1"`
	EmployerPostalL2      string                `json:"employer_postal_l2"`
	EmployerPostalSuburb  string                `json:"employer_postal_suburb"`
	EmployerPostalCity    string                `json:"employer_postal_city"`
	EmployerPostalCode2   string                `json:"employer_postal_code2"`
	EmployerPostalCountry string                `json:"employer_postal_country"`
	PracticeNumber        string                `json:"practice_number"`
	UpdatedAt             time.Time             `json:"updated_at"`
	Deleted               bool                  `json:"deleted"`
	Synced                bool                  `json:"synced"`
	Addresses             []ClientAddress       `json:"addresses"`
	ContactDetails        []ClientContactDetail `json:"contact_details"`
	Banks                 []ClientBank          `json:"banks"`
}

type ClientSyncDTO struct {
	ID                    string    `json:"id"`
	FirstName             string    `json:"first_name"`
	MiddleName            string    `json:"middle_name"`
	LastName              string    `json:"last_name"`
	IDNumber              string    `json:"id_number"`
	JurisdictionType      string    `json:"jurisdiction_type"`
	JurisdictionOther     string    `json:"jurisdiction_other"`
	RegistrationNumber    string    `json:"registration_number"`
	Occupation            string    `json:"occupation"`
	MaritalStatus         string    `json:"marital_status"`
	EmployerName          string    `json:"employer_name"`
	EmployerNumber        string    `json:"employer_number"`
	EmployerAddressL1     string    `json:"employer_address_l1"`
	EmployerAddressL2     string    `json:"employer_address_l2"`
	EmployerSuburb        string    `json:"employer_suburb"`
	EmployerCity          string    `json:"employer_city"`
	EmployerPostalCode    string    `json:"employer_postal_code"`
	EmployerCountry       string    `json:"employer_country"`
	EmployerPostalSame    bool      `json:"employer_postal_same"`
	EmployerPostalL1      string    `json:"employer_postal_l1"`
	EmployerPostalL2      string    `json:"employer_postal_l2"`
	EmployerPostalSuburb  string    `json:"employer_postal_suburb"`
	EmployerPostalCity    string    `json:"employer_postal_city"`
	EmployerPostalCode2   string    `json:"employer_postal_code2"`
	EmployerPostalCountry string    `json:"employer_postal_country"`
	PracticeNumber        string    `json:"practice_number"`
	UpdatedAt             time.Time `json:"updated_at"`
	Deleted               bool      `json:"deleted"`
}

type ClientAddress struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	AddressType   string    `json:"address_type"`
	IsPrimary     bool      `json:"is_primary"`
	Line1         string    `json:"line1"`
	Line2         string    `json:"line2"`
	Suburb        string    `json:"suburb"`
	City          string    `json:"city"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	PostalSame    bool      `json:"postal_same"`
	PostalLine1   string    `json:"postal_line1"`
	PostalLine2   string    `json:"postal_line2"`
	PostalSuburb  string    `json:"postal_suburb"`
	PostalCity    string    `json:"postal_city"`
	PostalCode2   string    `json:"postal_code2"`
	PostalCountry string    `json:"postal_country"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
	Synced        bool      `json:"synced"`
}

type AddressSyncDTO struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	AddressType   string    `json:"address_type"`
	IsPrimary     bool      `json:"is_primary"`
	Line1         string    `json:"line1"`
	Line2         string    `json:"line2"`
	Suburb        string    `json:"suburb"`
	City          string    `json:"city"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	PostalSame    bool      `json:"postal_same"`
	PostalLine1   string    `json:"postal_line1"`
	PostalLine2   string    `json:"postal_line2"`
	PostalSuburb  string    `json:"postal_suburb"`
	PostalCity    string    `json:"postal_city"`
	PostalCode2   string    `json:"postal_code2"`
	PostalCountry string    `json:"postal_country"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
}

type ClientContactDetail struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	ContactType  string    `json:"contact_type"`
	ContactValue string    `json:"contact_value"`
	IsPrimary    bool      `json:"is_primary"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
	Synced       bool      `json:"synced"`
}

type ContactDetailSyncDTO struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	ContactType  string    `json:"contact_type"`
	ContactValue string    `json:"contact_value"`
	IsPrimary    bool      `json:"is_primary"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
}

type ClientBank struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	BankName      string    `json:"bank_name"`
	BranchCode    string    `json:"branch_code"`
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"`
	IsPrimary     bool      `json:"is_primary"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
	Synced        bool      `json:"synced"`
}

type BankSyncDTO struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	BankName      string    `json:"bank_name"`
	BranchCode    string    `json:"branch_code"`
	AccountNumber string    `json:"account_number"`
	AccountType   string    `json:"account_type"`
	IsPrimary     bool      `json:"is_primary"`
	UpdatedAt     time.Time `json:"updated_at"`
	Deleted       bool      `json:"deleted"`
}

type Service struct {
	ID           string    `json:"id"`
	ServiceType  string    `json:"service_type"`
	Description  string    `json:"description"`
	StandardRate float64   `json:"standard_rate"`
	DurationUnit string    `json:"duration_unit"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
	Synced       bool      `json:"synced"`
}

type ServiceSyncDTO struct {
	ID           string    `json:"id"`
	ServiceType  string    `json:"service_type"`
	Description  string    `json:"description"`
	StandardRate float64   `json:"standard_rate"`
	DurationUnit string    `json:"duration_unit"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
}

type AppConfig struct {
	LastUsername string `json:"last_username"`
}

type ExportPayload struct {
	Filename string     `json:"filename"`
	Headers  []string   `json:"headers"`
	Rows     [][]string `json:"rows"`
}

type Matter struct {
	ID         string    `json:"id"`
	Reference  string    `json:"reference"`
	ClientID   string    `json:"client_id"`
	Status     string    `json:"status"`
	MatterType string    `json:"matter_type"`
	UpdatedAt  time.Time `json:"updated_at"`
	Deleted    bool      `json:"deleted"`
	Synced     bool      `json:"synced"`
}

type MatterNote struct {
	ID        string    `json:"id"`
	MatterID  string    `json:"matter_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	Synced    bool      `json:"synced"`
}

type MatterService struct {
	ID           string    `json:"id"`
	MatterID     string    `json:"matter_id"`
	ServiceID    string    `json:"service_id"`
	SnapshotDesc string    `json:"snapshot_desc"`
	SnapshotRate float64   `json:"snapshot_rate"`
	SnapshotUnit string    `json:"snapshot_unit"`
	Qty          float64   `json:"qty"`
	AddTax       bool      `json:"add_tax"`
	UpdatedAt    time.Time `json:"updated_at"`
	Deleted      bool      `json:"deleted"`
	Synced       bool      `json:"synced"`
}
