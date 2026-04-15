package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCORS_AllowsOrigin(t *testing.T) {
	h := CORS("https://example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("Allow-Origin: got %q, want https://example.com", got)
	}
}

func TestCORS_PreflightReturnsOK(t *testing.T) {
	h := CORS("*")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called on preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Errorf("preflight status: got %d, want 200 or 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("Allow-Methods header missing")
	}
}

func TestParseOrigins(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", []string{"*"}},
		{"*", []string{"*"}},
		{"https://a.com", []string{"https://a.com"}},
		{"https://a.com,https://b.com", []string{"https://a.com", "https://b.com"}},
		{" https://a.com , https://b.com ", []string{"https://a.com", "https://b.com"}},
		{",,", []string{"*"}},
	}
	for _, tc := range cases {
		got := parseOrigins(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseOrigins(%q): got %v, want %v", tc.in, got, tc.want)
		}
	}
}
