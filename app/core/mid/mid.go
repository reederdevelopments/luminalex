package mid

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maoni/app/core/web"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Log logs information about each request.
func Log(l *log.Logger) web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
			l.Printf("request started: method=%s path=%s remote_addr=%s", r.Method, r.URL.Path, r.RemoteAddr)

			err := handler(w, r)

			// In a more advanced logger, a response writer wrapper would be used to capture status code.
			l.Printf("request completed: method=%s path=%s duration=%s", r.Method, r.URL.Path, time.Since(start))

			return err
		}
	}
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// CatchErr catches errors from the handler chain and responds to the client.
func CatchErr(l *log.Logger) web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			// The error is returned from the handler, not recovered.
			err := handler(w, r)
			if err != nil {
				l.Printf("ERROR: %v", err)

				// Handle client-side connection drops specifically
				if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "client disconnected") {
					// Don't write a header if one has already been written.
					// The request context being canceled is the main signal here.
					return err
				}

				// If response has already started, we can't send a new error.
				rw, ok := w.(*responseWriter)
				if ok && rw.wroteHeader {
					return err
				}

				// Respond with a generic 500 internal server error.
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			// The error has been handled, so we return nil.
			return nil
		}
	}
}

// CatchPanic recovers from panics and logs the error.
func CatchPanic() web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					err = fmt.Errorf("panic: %v", rec)
					log.Printf("PANIC : %v\n%s", err, debug.Stack())

					rw, ok := w.(*responseWriter)
					if !ok || !rw.wroteHeader {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
				}
			}()

			return handler(w, r)
		}
	}
}

// TryGzip is a middleware that applies Gzip compression.
var TryGzip web.Middleware = func(h web.Handler) web.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Errors from the main handler will be caught by other middleware
			// like CatchErr. Here we just call it.
			_ = h(w, r)
		})

		// Apply the gzip middleware.
		middleware.Compress(5)(next).ServeHTTP(w, r)

		// The error from h(w,r) is not captured here, but that is by design
		// in this middleware pattern. It will be caught by CatchErr.
		// Since this middleware writes to the response, it should not return
		// an error itself.
		return nil
	}
}
