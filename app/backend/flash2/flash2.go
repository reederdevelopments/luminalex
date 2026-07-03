package flash2

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

type Type int

const (
	Info Type = iota + 1
	Warn
	Err

	flashKey  = "flash"
	separator = "$"

	flashExpTime = time.Minute
)

type Flash struct {
	Type Type
	Msg  string
}

func (f Flash) Marshal() string {
	flash, err := json.Marshal(f)
	if err != nil {
		panic(err.Error())
	}

	return base64.URLEncoding.EncodeToString(flash)
}

func Unmarshal(flash string) (Flash, error) {
	f, err := base64.URLEncoding.DecodeString(flash)
	if err != nil {
		return Flash{}, err
	}

	out := Flash{}
	err = json.Unmarshal(f, &out)
	if err != nil {
		return Flash{}, err
	}

	return out, nil
}

func Create(w http.ResponseWriter, flashType Type, msg string, now time.Time) {
	flash := Flash{
		Type: flashType,
		Msg:  msg,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     flashKey,
		Value:    flash.Marshal(),
		Path:     "/",
		Expires:  now.Add(flashExpTime),
		HttpOnly: true,
		Secure:   true,                  // added
		SameSite: http.SameSiteNoneMode, // changed from Lax
	})
}

// Get attempts to find and unmarshal a flash message. If it is successful, the Flash is returned
// with true, else false is returned.
func Get(w http.ResponseWriter, r *http.Request, now time.Time) (Flash, bool) {
	// get cookie
	c, err := r.Cookie(flashKey)
	if err != nil {
		return Flash{}, false
	}

	flash, err := Unmarshal(c.Value)
	if err != nil {
		return Flash{}, false
	}

	// delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     flashKey,
		Value:    "",
		Path:     "/",
		Expires:  now.Add(-time.Hour * 24),
		HttpOnly: true,
		Secure:   true,                  // added
		SameSite: http.SameSiteNoneMode, // changed from Lax
	})

	return flash, true
}

func (t Type) String() string {
	switch t {
	case Info:
		return "Info"
	case Err:
		return "Error"
	case Warn:
		return "Warning"
	default:
		panic("unimplemented flash type")
	}
}
