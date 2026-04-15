package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

// Recoverer catches panics from downstream handlers, logs the stack trace
// via zerolog, and responds with 500 {"error":"internal_error"}.
func Recoverer(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error().
						Interface("panic", rec).
						Str("stack", string(debug.Stack())).
						Str("path", r.URL.Path).
						Msg("panic recovered")

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal_error","message":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
