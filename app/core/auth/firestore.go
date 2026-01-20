package auth

import (
	"context"
	"errors"
	"log"
	"maoni/app/core/collection"
	"maoni/app/core/web"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Rockup-Consulting/std/x/randx"
	"github.com/go-viper/mapstructure/v2"
	"github.com/markbates/goth"
	"google.golang.org/api/iterator"
)

type FireStore struct {
	db      *firestore.Client
	s       Service
	l       *log.Logger
	devMode bool
}

func NewFireStore(l *log.Logger, client *firestore.Client, authService Service, devMode bool) FireStore {
	return FireStore{
		db:      client,
		s:       authService,
		l:       l,
		devMode: devMode,
	}
}

func (f FireStore) Get(ctx context.Context, now time.Time, authToken string) (User, Session, error) {
	token, err := f.s.UnmarshalToken(authToken)
	if err != nil {
		f.l.Printf("failed to unmarshal token: %s", err)
		return User{}, Session{}, ErrNotAuthenticated
	}

	s, err := f.db.Collection(collection.Sessions).Doc(token.SessionID).Get(ctx)
	if err != nil {
		f.l.Printf("err fetching session: %s", err)
		return User{}, Session{}, err
	}

	var session Session
	if err := mapstructure.Decode(s.Data(), &session); err != nil {
		f.l.Printf("err decoding session: %s", err)
		return User{}, Session{}, ErrNotAuthenticated
	}

	if session.Invalidated {
		f.l.Println("session has been invalidated")
		return User{}, Session{}, ErrNotAuthenticated
	}

	expiresAt := time.Unix(session.ExpiresAt, 0)
	if now.After(expiresAt) {
		f.l.Println("session has expired")
		return User{}, Session{}, ErrNotAuthenticated
	}

	doc, err := f.db.Collection(collection.Users).Doc(session.UserID).Get(ctx)
	if err != nil {
		return User{}, Session{}, err
	}

	var user User
	if err := mapstructure.Decode(doc.Data(), &user); err != nil {
		return User{}, Session{}, err
	}

	return user, session, nil
}

func (f FireStore) HttpGet(ctx context.Context, now time.Time, w http.ResponseWriter, r *http.Request) (User, Session, error) {
	c, err := r.Cookie(Cookie)
	if err != nil {
		deleteCookieAndRedirect(w, r, "/signin", now)
		return User{}, Session{}, web.ErrHandled
	}

	user, session, err := f.Get(ctx, now, c.Value)
	if err != nil {
		deleteCookieAndRedirect(w, r, "/signin", now)
		return User{}, Session{}, web.ErrHandled
	}

	return user, session, nil
}

func (f FireStore) HttpCreate(ctx context.Context, now time.Time, u goth.User, w http.ResponseWriter, r *http.Request) error {
	id := randx.UID()

	iter := f.db.Collection(collection.Users).Where("Email", "==", u.Email).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			f.l.Printf("User %q not found, creating new user.", u.Email)
			newUserID := randx.UID()
			newUser := User{
				ID:           newUserID,
				FirstName:    u.FirstName,
				LastName:     u.LastName,
				Email:        u.Email,
				Name:         u.Name,
				GoogleID:     u.UserID,
				Thumbnail:    u.AvatarURL,
				LastSyncTime: now.Unix(),
				IsAdmin:      false,
			}
			if _, err := f.db.Collection(collection.Users).Doc(newUserID).Set(ctx, newUser); err != nil {
				return err
			}
			doc, err = f.db.Collection(collection.Users).Doc(newUserID).Get(ctx)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var user User
	if err := mapstructure.Decode(doc.Data(), &user); err != nil {
		return err
	}

	s := Session{
		ID:          id,
		UserID:      user.ID,
		CreatedAt:   now.Unix(),
		LastActive:  now.Unix(),
		ExpiresAt:   now.Add(SessionDuration).Unix(),
		Invalidated: false,
	}

	if _, err := f.db.Collection(collection.Sessions).Doc(id).Set(ctx, s); err != nil {
		return err
	}

	token := f.s.CreateToken(s)

	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    token,
		Expires:  now.Add(SessionDuration),
		Secure:   !f.devMode,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (f FireStore) HttpInvalidate(ctx context.Context, now time.Time, w http.ResponseWriter, r *http.Request) error {
	c, err := r.Cookie(Cookie)
	if err != nil {
		deleteCookieAndRedirect(w, r, "/signin", now)
		return nil
	}

	_, session, err := f.Get(ctx, now, c.Value)
	if err != nil {
		deleteCookieAndRedirect(w, r, "/signin", now)
		return err
	}

	if _, err := f.db.Collection(collection.Sessions).Doc(session.ID).Update(ctx, []firestore.Update{
		{Path: "Invalidated", Value: true},
	}); err != nil {
		f.l.Printf("error invalidating session: %s", err)
	}

	deleteCookieAndRedirect(w, r, "/signin", now)
	return nil
}

func (f FireStore) Create(ctx context.Context, now time.Time, u goth.User) (*http.Cookie, error) {
	return nil, errors.New("not implemented")
}

func deleteCookieAndRedirect(w http.ResponseWriter, r *http.Request, path string, now time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    "",
		Path:     "/",
		Expires:  now.Add(-time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, path, http.StatusSeeOther)
}
