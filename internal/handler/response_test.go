package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
)

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(out); err != nil {
		t.Fatalf("decode body: %v (raw=%q)", err, rec.Body.String())
	}
}

func TestJSON_WritesEnvelopeAndStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, map[string]string{"hello": "world"})

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	var body struct {
		Data map[string]string `json:"data"`
	}
	decodeBody(t, rec, &body)
	if body.Data["hello"] != "world" {
		t.Errorf("data: got %+v", body.Data)
	}
}

func TestJSON_CustomStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusCreated, map[string]int{"id": 1})

	if rec.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201", rec.Code)
	}
}

func TestError_WritesErrorEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, http.StatusNotFound, "not_found", "guest not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rec.Code)
	}
	var body struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	decodeBody(t, rec, &body)
	if body.Error != "not_found" {
		t.Errorf("error: got %q", body.Error)
	}
	if body.Message != "guest not found" {
		t.Errorf("message: got %q", body.Message)
	}
}

func TestError_EmptyMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, http.StatusUnauthorized, "unauthorized", "")

	// omitempty significa que "message" não deve aparecer no JSON
	var raw map[string]any
	decodeBody(t, rec, &raw)
	if _, ok := raw["message"]; ok {
		t.Errorf("message field should be omitted, got %+v", raw)
	}
	if raw["error"] != "unauthorized" {
		t.Errorf("error: got %v", raw["error"])
	}
}

func TestValidationError(t *testing.T) {
	type payload struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"required,min=18"`
	}
	v := validator.New()
	err := v.Struct(payload{Email: "not-an-email", Age: 10})
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	rec := httptest.NewRecorder()
	ValidationError(rec, errs)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status: got %d, want 422", rec.Code)
	}
	var body struct {
		Error   string `json:"error"`
		Details []struct {
			Field string `json:"field"`
			Rule  string `json:"rule"`
		} `json:"details"`
	}
	decodeBody(t, rec, &body)
	if body.Error != "validation_error" {
		t.Errorf("error: got %q", body.Error)
	}
	if len(body.Details) != 2 {
		t.Fatalf("details: got %d entries, want 2 (%+v)", len(body.Details), body.Details)
	}
	fields := map[string]string{}
	for _, d := range body.Details {
		fields[d.Field] = d.Rule
	}
	if fields["Email"] != "email" {
		t.Errorf("Email rule: got %q, want email", fields["Email"])
	}
	if fields["Age"] != "min" {
		t.Errorf("Age rule: got %q, want min", fields["Age"])
	}
}

func TestNoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	NoContent(rec)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body should be empty, got %q", rec.Body.String())
	}
}
