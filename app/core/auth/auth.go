package auth

const Cookie = "MAONI_AUTH"

// User represents an authenticated user in the system.
type User struct {
	ID           string
	FirstName    string
	LastName     string
	Email        string
	Name         string
	GoogleID     string
	LastSyncTime int64
	Thumbnail    string
	IsAdmin      bool
}
