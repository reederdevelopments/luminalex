package auth

const Cookie = "UJUZI_RELOADED_AUTH"

// Add explicit firestore tags to guarantee correct mapping
type User struct {
	ID              string         `firestore:"-"`
	FirstName       string         `firestore:"FirstName"`
	LastName        string         `firestore:"LastName"`
	Email           string         `firestore:"Email"`
	Name            string         `firestore:"Name"`
	GoogleID        string         `firestore:"GoogleID"`
	LastSyncTime    int64          `firestore:"LastSyncTime"`
	Thumbnail       string         `firestore:"Thumbnail"`
	IsAdmin         bool           `firestore:"IsAdmin"`
	DefaultCountry  string         `firestore:"DefaultCountry"`
	Countries       []string       `firestore:"Countries"`
	Groups          []string       `firestore:"Groups"`
	Tools           []string       `firestore:"Tools"`
	DashboardClicks map[string]int `firestore:"dashboardClicks,omitempty"`
}

type Group struct {
	ID         string   `firestore:"-"`
	Name       string   `firestore:"name"`
	Level      string   `firestore:"level"`
	Dashboards []string `firestore:"dashboards"`
}

type DashParams struct {
	CountryCode string `firestore:"countryCode"`
	Branch      string `firestore:"branch"`
	Consultant  string `firestore:"consultant"`
	Cycle       string `firestore:"cycle"`
	PrevCycle   string `firestore:"prevCycle"`
	Date        string `firestore:"date"`
	Period      string `firestore:"period"`
}

type DashPage struct {
	Name string `firestore:"name"`
	Code string `firestore:"code"`
}

type Dashboard struct {
	ID          string     `firestore:"-"`
	Name        string     `firestore:"name"`
	Description string     `firestore:"description"`
	ReportURL   string     `firestore:"reportUrl"`
	Area        string     `firestore:"area"`
	Parameters  DashParams `firestore:"parameters"`
	Pages       []DashPage `firestore:"pages"`
}
