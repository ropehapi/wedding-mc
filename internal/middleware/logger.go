package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// RequestIDFromContext returns the request ID injected by Logger middleware.
// Empty string if none is present.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.status = http.StatusOK
		r.wroteHeader = true
	}
	return r.ResponseWriter.Write(b)
}

// Logger returns a middleware that logs each HTTP request as JSON via zerolog.
// It injects a generated request ID into the request context (see RequestIDFromContext).
// It does NOT log the request body for security reasons.
func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := uuid.NewString()
			ctx := context.WithValue(r.Context(), requestIDKey, reqID)

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			next.ServeHTTP(rec, r.WithContext(ctx))

			logger.Info().
				Str("request_id", reqID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rec.status).
				Int64("duration_ms", time.Since(start).Milliseconds()).
				Str("remote_addr", r.RemoteAddr).
				Msg("http request")
		})
	}
}
