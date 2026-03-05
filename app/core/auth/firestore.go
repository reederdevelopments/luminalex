package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maoni/app/core/collection"
	"maoni/app/core/web"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/Rockup-Consulting/std/x/randx"
	"github.com/markbates/goth"
	"google.golang.org/api/iterator"
)

type FireStore struct {
	db           *firestore.Client
	s            Service
	l            *log.Logger
	devMode      bool
	sessionCache sync.Map // map[sessionID]Session
	userCache    sync.Map // map[userID]User
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

	// 1. Check session cache
	var session Session
	cached, found := f.sessionCache.Load(token.SessionID)
	if found {
		session = cached.(Session)
	} else {
		// Cache miss, fetch from Firestore
		s, err := f.db.Collection(collection.Sessions).Doc(token.SessionID).Get(ctx)
		if err != nil {
			f.l.Printf("err fetching session: %s", err)
			return User{}, Session{}, err
		}
		if err := s.DataTo(&session); err != nil {
			f.l.Printf("err decoding session: %s", err)
			return User{}, Session{}, ErrNotAuthenticated
		}
		// Store in cache
		f.sessionCache.Store(session.ID, session)
	}

	if session.Invalidated {
		f.l.Println("session has been invalidated")
		f.sessionCache.Delete(session.ID) // Evict from cache
		return User{}, Session{}, ErrNotAuthenticated
	}

	expiresAt := time.Unix(session.ExpiresAt, 0)
	if now.After(expiresAt) {
		f.l.Println("session has expired")
		f.sessionCache.Delete(session.ID) // Evict from cache
		return User{}, Session{}, ErrNotAuthenticated
	}

	// 2. Check user cache
	var user User
	cached, found = f.userCache.Load(session.UserID)
	if found {
		user = cached.(User)
	} else {
		// Cache miss, fetch from Firestore
		doc, err := f.db.Collection(collection.Users).Doc(session.UserID).Get(ctx)
		if err != nil {
			return User{}, Session{}, err
		}
		if err := doc.DataTo(&user); err != nil {
			return User{}, Session{}, err
		}
		// Store in cache
		f.userCache.Store(user.ID, user)
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
	normalizedEmail := strings.ToLower(strings.TrimSpace(u.Email))

	iter := f.db.Collection(collection.Users).Where("Email", "==", normalizedEmail).Documents(ctx)
	doc, err := iter.Next()

	var user User
	if err != nil {
		if err == iterator.Done {
			f.l.Printf("User %q not found, creating new user.", u.Email)
			user = User{
				ID:           randx.UID(),
				FirstName:    u.FirstName,
				LastName:     u.LastName,
				Email:        normalizedEmail,
				Name:         u.Name,
				GoogleID:     u.UserID,
				Thumbnail:    u.AvatarURL,
				LastSyncTime: now.Unix(),
				IsAdmin:      false,
			}
			if _, err := f.db.Collection(collection.Users).Doc(user.ID).Set(ctx, user); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Existing user, update their info
		if err := doc.DataTo(&user); err != nil {
			return fmt.Errorf("decoding existing user: %w", err)
		}

		user.FirstName = u.FirstName
		user.LastName = u.LastName
		user.Name = u.Name
		user.Thumbnail = u.AvatarURL
		user.LastSyncTime = now.Unix()

		if _, err := f.db.Collection(collection.Users).Doc(user.ID).Set(ctx, user); err != nil {
			return fmt.Errorf("updating existing user: %w", err)
		}
	}

	// Cache the user object
	f.userCache.Store(user.ID, user)

	// Create a new session
	id := randx.UID()
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

	// Cache the new session
	f.sessionCache.Store(s.ID, s)

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

	// Delete from cache
	f.sessionCache.Delete(session.ID)

	deleteCookieAndRedirect(w, r, "/signin", now)
	return nil
}

func (f FireStore) Create(ctx context.Context, now time.Time, u goth.User) (*http.Cookie, error) {
	return nil, errors.New("not implemented")
}

func (f FireStore) GetUserByEmail(ctx context.Context, email string) (User, error) {
	// Note: Caching by email isn't implemented for simplicity, but could be added.
	// We still benefit from caching by user ID if a user is fetched multiple times via ID.

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	iter := f.db.Collection(collection.Users).Where("Email", "==", normalizedEmail).Limit(1).Documents(ctx)
	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return User{}, fmt.Errorf("user with email %s not found", email)
		}
		return User{}, fmt.Errorf("querying user by email: %w", err)
	}

	var user User
	if err := doc.DataTo(&user); err != nil {
		return User{}, fmt.Errorf("decoding user data: %w", err)
	}

	// Cache the user object by its ID for subsequent lookups
	f.userCache.Store(user.ID, user)

	return user, nil
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
