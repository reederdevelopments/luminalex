package base

import (
	"context"
	"net/http"
	"controlroom/app/backend/auth"
	"controlroom/app/backend/web"

	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

func (m module) signinLoader(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := web.Now()

	cookie, err := r.Cookie(auth.Cookie)
	if err == nil {
		// Validate the cookie to prevent loops
		_, _, err := m.sessionStore.Get(ctx, now, cookie.Value)
		if err == nil {
			// Preserve query string when redirecting to the dashboard
			target := "/"
			if raw := r.URL.RawQuery; raw != "" {
				target = "/?" + raw
			}
			http.Redirect(w, r, target, http.StatusSeeOther)
			return web.ErrHandled
		}
	}

	return signinPage().Render(ctx, w)
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
	now := web.Now()

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		m.l.Printf("ERROR completing user auth: %s", err)
		return err
	}

	if err := m.sessionStore.HttpCreate(ctx, now, user, w, r); err != nil {
		m.l.Printf("ERROR creating session: %s", err)
		return err
	}

	return callback().Render(ctx, w)
}

func (m module) logoutHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	now := web.Now()

	gothic.Logout(w, r)
	if err := m.sessionStore.HttpInvalidate(ctx, now, w, r); err != nil {
		m.l.Printf("ERROR invalidating session: %s", err)
		return err
	}
	return nil
}
