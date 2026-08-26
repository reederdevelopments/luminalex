package views

type LoginData struct {
	Error        string
	LastUsername string
}

type HomeData struct {
	Title string
}

type Contact struct {
	ID        string
	Category  string
	Fields    []string
	IsEditing bool
}

type ClientAddress struct {
	ID            string `json:"id"`
	AddressType   string `json:"address_type"`
	IsPrimary     bool   `json:"is_primary"`
	Line1         string `json:"line1"`
	Line2         string `json:"line2"`
	Suburb        string `json:"suburb"`
	City          string `json:"city"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
	PostalSame    bool   `json:"postal_same"`
	PostalLine1   string `json:"postal_line1"`
	PostalLine2   string `json:"postal_line2"`
	PostalSuburb  string `json:"postal_suburb"`
	PostalCity    string `json:"postal_city"`
	PostalCode2   string `json:"postal_code2"`
	PostalCountry string `json:"postal_country"`
}

type ClientContactDetail struct {
	ID           string `json:"id"`
	ContactType  string `json:"contact_type"`
	ContactValue string `json:"contact_value"`
	IsPrimary    bool   `json:"is_primary"`
}

type ClientBank struct {
	ID            string `json:"id"`
	BankName      string `json:"bank_name"`
	BranchCode    string `json:"branch_code"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	IsPrimary     bool   `json:"is_primary"`
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
	Addresses             []ClientAddress       `json:"addresses"`
	ContactDetails        []ClientContactDetail `json:"contact_details"`
	Banks                 []ClientBank          `json:"banks"`
}

type Service struct {
	ID           string  `json:"id"`
	ServiceType  string  `json:"service_type"`
	Description  string  `json:"description"`
	StandardRate float64 `json:"standard_rate"`
	DurationUnit string  `json:"duration_unit"`
}
