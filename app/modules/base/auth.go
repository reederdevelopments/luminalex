package base

import (
	"context"
	"maoni/app/core/auth"
	"maoni/app/core/web"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

func (m module) signinLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()

	cookie, err := r.Cookie(auth.Cookie)
	if err == nil {
		// A cookie exists. We must validate it before redirecting to prevent loops.
		_, _, err := m.sessionStore.Get(ctx, now, cookie.Value)
		if err == nil {
			// The session is valid, redirect to the main page.
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return web.ErrHandled
		}
		// If there was an error (e.g., expired/invalid session),
		// we fall through to render the sign-in page again.
		// The invalid cookie will be overwritten upon a new successful login.
	}

	// No cookie, or the existing one was invalid. Render the sign-in page.
	// Note: flash messages are not implemented yet, so passing empty values.
	return signinPage(false, "").Render(ctx, w)
}

func (m module) beginAuthHandler(w http.ResponseWriter, r *http.Request) error {
	provider := chi.URLParam(r, "provider")
	ctx := context.WithValue(r.Context(), gothic.ProviderParamKey, provider)
	r = r.WithContext(ctx)

	gothic.BeginAuthHandler(w, r)
	return nil
}

func (m module) authCallbackHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		m.l.Printf("ERROR completing user auth: %s", err)
		// Return the error directly. The middleware will handle the response.
		return err
	}

	if err := m.sessionStore.HttpCreate(ctx, now, user, w, r); err != nil {
		m.l.Printf("ERROR creating session: %s", err)
		// Return the error directly. The middleware will handle the response.
		return err
	}

	return callback().Render(ctx, w)
}

func (m module) logoutHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := time.Now()

	gothic.Logout(w, r)
	if err := m.sessionStore.HttpInvalidate(ctx, now, w, r); err != nil {
		m.l.Printf("ERROR invalidating session: %s", err)
		// The HttpInvalidate function already handles the redirect, but we return the error
		// so the middleware chain is aware of it.
		return err
	}
	// Normally, HttpInvalidate will redirect and this won't be reached.
	return nil
}
