package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type envelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type validationDetail struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

type validationEnvelope struct {
	Error   string             `json:"error"`
	Details []validationDetail `json:"details"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: code, Message: message})
}

func ValidationError(w http.ResponseWriter, errs validator.ValidationErrors) {
	details := make([]validationDetail, 0, len(errs))
	for _, e := range errs {
		details = append(details, validationDetail{
			Field: e.Field(),
			Rule:  e.Tag(),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(validationEnvelope{
		Error:   "validation_error",
		Details: details,
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
