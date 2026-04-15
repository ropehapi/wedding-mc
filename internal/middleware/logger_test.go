package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func newTestLogger() (zerolog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	return logger, &buf
}

func TestLogger_GeneratesRequestID(t *testing.T) {
	logger, _ := newTestLogger()
	var captured string

	h := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured == "" {
		t.Fatal("expected request ID in context, got empty string")
	}
}

func TestLogger_LogsFields(t *testing.T) {
	logger, buf := newTestLogger()

	h := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/foo/bar", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log not valid json: %v (raw=%q)", err, buf.String())
	}
	if entry["method"] != "POST" {
		t.Errorf("method: got %v, want POST", entry["method"])
	}
	if entry["path"] != "/foo/bar" {
		t.Errorf("path: got %v", entry["path"])
	}
	if entry["status"].(float64) != 201 {
		t.Errorf("status: got %v, want 201", entry["status"])
	}
	if entry["request_id"] == "" {
		t.Errorf("request_id missing")
	}
}

func TestLogger_CapturesStatusWhenHandlerWritesDirectly(t *testing.T) {
	logger, buf := newTestLogger()

	h := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var entry map[string]any
	_ = json.Unmarshal(buf.Bytes(), &entry)
	if entry["status"].(float64) != 418 {
		t.Errorf("status: got %v, want 418", entry["status"])
	}
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := RequestIDFromContext(req.Context()); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
