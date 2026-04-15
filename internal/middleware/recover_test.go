package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverer_CatchesPanic(t *testing.T) {
	logger, buf := newTestLogger()

	h := Recoverer(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "internal_error" {
		t.Errorf("error: got %q, want internal_error", body["error"])
	}
	if !strings.Contains(buf.String(), "panic recovered") {
		t.Errorf("expected log entry, got %q", buf.String())
	}
}

func TestRecoverer_PassesThroughOK(t *testing.T) {
	logger, _ := newTestLogger()

	h := Recoverer(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body: got %q", rec.Body.String())
	}
}
