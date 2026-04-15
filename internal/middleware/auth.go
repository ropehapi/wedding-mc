package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const userIDKey contextKey = "user_id"

// UserIDFromContext returns the authenticated user ID injected by Auth middleware.
// Returns ("", false) if no user is authenticated.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

// WithUserID injects a userID into ctx, the same way the Auth middleware does.
// Intended for use in handler tests that need an authenticated context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// Auth returns a middleware that validates a Bearer JWT from the Authorization header
// using HS256 + the given secret. On success it injects the `sub` claim as userID
// into the request context (see UserIDFromContext). On failure it responds 401.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, ok := extractBearerToken(r.Header.Get("Authorization"))
			if !ok {
				unauthorized(w)
				return
			}

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})
			if err != nil || !token.Valid || claims.Subject == "" {
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized","message":"invalid or missing token"}`))
}
