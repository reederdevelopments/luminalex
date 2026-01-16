package auth

import (
	"context"
	"maoni/app/core/web"
	"net/http"
	"time"

	"github.com/markbates/goth"
)

const (
	SessionDuration = time.Hour * 24 * 365
)

type Session struct {
	ID          string
	UserID      string
	CreatedAt   int64
	ExpiresAt   int64
	LastActive  int64
	Invalidated bool
}

type Store interface {
	Get(ctx context.Context, now time.Time, authToken string) (User, Session, error)
	HttpGet(ctx context.Context, now time.Time, w http.ResponseWriter, r *http.Request) (User, Session, error)
	Create(ctx context.Context, now time.Time, u goth.User) (*http.Cookie, error)
	HttpCreate(ctx context.Context, now time.Time, u goth.User, w http.ResponseWriter, r *http.Request) error
	HttpInvalidate(ctx context.Context, now time.Time, w http.ResponseWriter, r *http.Request) error
	Mid(h web.Handler) web.Handler
}
