package base

import (
	"fmt"
	"log"
	"maoni/app/core/auth"
	"maoni/app/core/events"
	"maoni/app/core/mid"
	"maoni/app/core/survey"
	"maoni/app/core/web"
	"net/http"
)

type module struct {
	l            *log.Logger
	sessionStore auth.Store
	surveyStore  survey.Store
	eventBroker  events.Subscriber
}

func stdMid(l *log.Logger, additionalMid ...web.Middleware) []web.Middleware {
	middlewares := []web.Middleware{
		mid.Log(l),
		mid.CatchErr(l),
		mid.CatchPanic(),
	}
	middlewares = append(middlewares, additionalMid...)
	return middlewares
}

func InitModule(
	l *log.Logger,
	app *web.App,
	sessionStore auth.Store,
	surveyStore survey.Store,
	eventBroker events.Subscriber,
) {
	m := module{
		l:            l,
		sessionStore: sessionStore,
		surveyStore:  surveyStore,
		eventBroker:  eventBroker,
	}

	// Unprotected routes
	app.Handle(http.MethodGet, "/signin", m.signinLoader, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}", m.beginAuthHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/auth/{provider}/callback", m.authCallbackHandler, stdMid(l)...)
	app.Handle(http.MethodGet, "/logout", m.logoutHandler, stdMid(l, sessionStore.Mid)...)

	// Protected routes
	app.Handle(http.MethodGet, "/", m.surveyLoader, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/surveys/{id}", m.viewSurveyHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodPost, "/surveys/{id}", m.submitSurveyHandler, stdMid(l, sessionStore.Mid)...)
	app.Handle(http.MethodGet, "/events", m.eventStreamHandler, stdMid(l, sessionStore.Mid)...)
}

func (m module) eventStreamHandler(w http.ResponseWriter, r *http.Request) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return web.ErrHandled
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := m.eventBroker.Subscribe()
	defer m.eventBroker.Unsubscribe(clientChan)

	ctx := r.Context()

	for {
		select {
		case event := <-clientChan:
			if _, err := fmt.Fprintf(w, "event: %s\ndata: {}\n\n", event); err != nil {
				m.l.Printf("SSE error writing to client: %v", err)
				return nil // Client has disconnected, so we just stop.
			}
			flusher.Flush()
		case <-ctx.Done():
			m.l.Println("SSE client disconnected")
			return nil
		}
	}
}
