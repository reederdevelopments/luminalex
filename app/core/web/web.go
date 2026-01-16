package web

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const DefaultShutdownTimeout = time.Second * 15

var ErrHandled = errors.New("handler has already written a response")

type Handler func(w http.ResponseWriter, r *http.Request) error

type Middleware func(h Handler) Handler

func wrapMiddleware(mw []Middleware, handler Handler) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}
	return handler
}

type App struct {
	*chi.Mux
}

func NewApp() *App {
	return &App{
		Mux: chi.NewMux(),
	}
}

func (a *App) Handle(method string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			// Error handling can be enhanced here
			// For now, it's handled by middleware like mid.CatchErr
		}
	}

	// a.Method(method, path, h) // OLD
	a.Method(method, path, http.HandlerFunc(h)) // NEW
}

func (a *App) HandleStd(method string, path string, handler http.HandlerFunc, mw ...Middleware) {
	wrappedHandler := func(w http.ResponseWriter, r *http.Request) error {
		handler(w, r)
		return nil
	}
	a.Handle(method, path, wrappedHandler, mw...)
}

func Set(ctx context.Context, key any, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

func Get[T any](ctx context.Context, key any) T {
	v, ok := ctx.Value(key).(T)
	if !ok {
		panic("value not found in context")
	}
	return v
}
