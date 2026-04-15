package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/cors"
)

// CORS returns a middleware configured with the given origins list.
// allowedOrigins is a comma-separated string (e.g. "https://app.example.com,https://admin.example.com").
// Use "*" to allow any origin (dev only — credentials will be disabled in that case).
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := parseOrigins(allowedOrigins)
	allowCreds := !containsWildcard(origins)

	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: allowCreds,
		MaxAge:           300,
	})
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if o == "*" {
			return true
		}
	}
	return false
}
