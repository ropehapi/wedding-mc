package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/ropehapi/wedding-mc/internal/domain"
	"github.com/ropehapi/wedding-mc/internal/middleware"
	"github.com/ropehapi/wedding-mc/internal/service"
)

// GuestHandler handles HTTP requests for guest endpoints.
type GuestHandler struct {
	svc      service.GuestService
	validate *validator.Validate
}

func NewGuestHandler(svc service.GuestService) *GuestHandler {
	return &GuestHandler{svc: svc, validate: validator.New()}
}

// --- Request / Response types ---

type createGuestRequest struct {
	Name string `json:"name" validate:"required"`
}

type guestResponse struct {
	ID        string     `json:"id"`
	WeddingID string     `json:"wedding_id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	RSVPAt    *time.Time `json:"rsvp_at,omitempty"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

type guestSummaryResponse struct {
	Pending   int `json:"pending"`
	Confirmed int `json:"confirmed"`
	Declined  int `json:"declined"`
}

// --- Handlers ---

// Create godoc
// @Summary Adicionar convidado
// @Tags guests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body createGuestRequest true "Dados do convidado"
// @Success 201 {object} guestResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Failure 422 {object} validationEnvelope
// @Router /v1/guests [post]
func (h *GuestHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req createGuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if errs := h.validate.Struct(req); errs != nil {
		var ve validator.ValidationErrors
		if errors.As(errs, &ve) {
			ValidationError(w, ve)
			return
		}
	}

	guest, err := h.svc.CreateGuest(r.Context(), userID, req.Name)
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusCreated, toGuestResponse(guest))
}

// List godoc
// @Summary Listar convidados
// @Tags guests
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filtro por status (pending, confirmed, declined)"
// @Success 200 {array} guestResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/guests [get]
func (h *GuestHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var statusFilter *domain.RSVPStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.RSVPStatus(s)
		statusFilter = &st
	}

	guests, err := h.svc.ListGuests(r.Context(), userID, statusFilter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := make([]guestResponse, len(guests))
	for i := range guests {
		resp[i] = toGuestResponse(&guests[i])
	}
	JSON(w, http.StatusOK, resp)
}

// Summary godoc
// @Summary Resumo de convidados por status
// @Tags guests
// @Security BearerAuth
// @Produce json
// @Success 200 {object} guestSummaryResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/guests/summary [get]
func (h *GuestHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	counts, err := h.svc.GetSummary(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, guestSummaryResponse{
		Pending:   counts[domain.RSVPPending],
		Confirmed: counts[domain.RSVPConfirmed],
		Declined:  counts[domain.RSVPDeclined],
	})
}

// Update godoc
// @Summary Atualizar convidado
// @Tags guests
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param guestID path string true "ID do convidado"
// @Param body body createGuestRequest true "Campos a atualizar"
// @Success 200 {object} guestResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Failure 422 {object} validationEnvelope
// @Router /v1/guests/{guestID} [patch]
func (h *GuestHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	guestID := chi.URLParam(r, "guestID")

	var req createGuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if errs := h.validate.Struct(req); errs != nil {
		var ve validator.ValidationErrors
		if errors.As(errs, &ve) {
			ValidationError(w, ve)
			return
		}
	}

	guest, err := h.svc.UpdateGuest(r.Context(), userID, guestID, req.Name)
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, toGuestResponse(guest))
}

// Delete godoc
// @Summary Remover convidado
// @Tags guests
// @Security BearerAuth
// @Param guestID path string true "ID do convidado"
// @Success 204
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/guests/{guestID} [delete]
func (h *GuestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	guestID := chi.URLParam(r, "guestID")

	if err := h.svc.DeleteGuest(r.Context(), userID, guestID); err != nil {
		h.handleError(w, err)
		return
	}

	NoContent(w)
}

func (h *GuestHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrConflict):
		Error(w, http.StatusConflict, "conflict", "conflict")
	case errors.Is(err, domain.ErrForbidden):
		Error(w, http.StatusForbidden, "forbidden", "access denied")
	case errors.Is(err, domain.ErrValidation):
		Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	default:
		Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func toGuestResponse(g *domain.Guest) guestResponse {
	return guestResponse{
		ID:        g.ID,
		WeddingID: g.WeddingID,
		Name:      g.Name,
		Status:    string(g.Status),
		RSVPAt:    g.RSVPAt,
		CreatedAt: g.CreatedAt.Format(time.RFC3339),
		UpdatedAt: g.UpdatedAt.Format(time.RFC3339),
	}
}
