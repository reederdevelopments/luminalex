package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"controlroom/app/backend/web"
)

type key int8

const ctxKey key = 0

type AuthCtx struct {
	User    User
	Session Session
}

func Set(ctx context.Context, u User, s Session) context.Context {
	c := AuthCtx{u, s}
	return context.WithValue(ctx, ctxKey, c)
}

func FromCtx(ctx context.Context) AuthCtx {
	c, ok := ctx.Value(ctxKey).(AuthCtx)
	if !ok {
		panic("AuthCtx not on context")
	}
	return c
}

func (f FireStore) Mid(h web.Handler) web.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		now := web.Now()

		u, s, err := f.HttpGet(ctx, now, w, r)
		if err != nil {
			return err
		}

		ctx = Set(ctx, u, s)
		r = r.WithContext(ctx)

		return h(w, r)
	}
}

func CatchErr(l *log.Logger) web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			// The error is returned from the handler, not recovered.
			err := handler(w, r)
			if err != nil {
				// If the error is our sentinel, it means the response has been handled.
				// We return nil to stop further processing gracefully.
				if errors.Is(err, web.ErrHandled) {
					return nil
				}

				l.Printf("ERROR: %v", err)

				// Handle client-side connection drops specifically
				if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "client disconnected") {
					// The client has gone away, so we can't write a response.
					return err
				}

				// The original superfluous write error happened here. Now it's protected
				// by the check for web.ErrHandled above.
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			// The error has been handled, so we return nil.
			return nil
		}
	}
}
