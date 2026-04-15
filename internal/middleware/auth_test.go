package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func signToken(t *testing.T, secret, sub string, exp time.Duration) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   sub,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func runAuth(t *testing.T, header string) (*httptest.ResponseRecorder, string, bool) {
	t.Helper()
	var capturedID string
	var capturedOK bool

	h := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID, capturedOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, capturedID, capturedOK
}

func TestAuth_NoHeader(t *testing.T) {
	rec, _, ok := runAuth(t, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rec.Code)
	}
	if ok {
		t.Errorf("handler should not have been reached")
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	cases := []string{
		"not-bearer",
		"Bearer ",
		"Bearer",
		"Basic abc",
	}
	for _, h := range cases {
		t.Run(h, func(t *testing.T) {
			rec, _, _ := runAuth(t, h)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("header %q: got %d, want 401", h, rec.Code)
			}
		})
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	rec, _, _ := runAuth(t, "Bearer not-a-jwt")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rec.Code)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	token := signToken(t, "other-secret", "user-1", time.Hour)
	rec, _, _ := runAuth(t, "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rec.Code)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	token := signToken(t, testSecret, "user-1", -time.Hour)
	rec, _, _ := runAuth(t, "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rec.Code)
	}
}

func TestAuth_EmptySubject(t *testing.T) {
	token := signToken(t, testSecret, "", time.Hour)
	rec, _, _ := runAuth(t, "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rec.Code)
	}
}

func TestAuth_ValidToken_InjectsUserID(t *testing.T) {
	token := signToken(t, testSecret, "user-123", time.Hour)
	rec, id, ok := runAuth(t, "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
	if !ok || id != "user-123" {
		t.Errorf("userID: got (%q, %v), want (user-123, true)", id, ok)
	}
}

func TestUserIDFromContext_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id, ok := UserIDFromContext(req.Context())
	if ok || id != "" {
		t.Errorf("got (%q, %v), want (\"\", false)", id, ok)
	}
}
