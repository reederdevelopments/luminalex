package auth

const Cookie = "CONTROLROOM_AUTH"

// Add explicit firestore tags to guarantee correct mapping
type User struct {
	ID           string `firestore:"-"`
	FirstName    string `firestore:"FirstName"`
	LastName     string `firestore:"LastName"`
	Email        string `firestore:"Email"`
	Name         string `firestore:"Name"`
	GoogleID     string `firestore:"GoogleID"`
	LastSyncTime int64  `firestore:"LastSyncTime"`
	Thumbnail    string `firestore:"Thumbnail"`
	IsAdmin      bool   `firestore:"IsAdmin"`
}
