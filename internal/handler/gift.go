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

// GiftHandler handles HTTP requests for gift endpoints.
type GiftHandler struct {
	svc      service.GiftService
	validate *validator.Validate
}

func NewGiftHandler(svc service.GiftService) *GiftHandler {
	return &GiftHandler{svc: svc, validate: validator.New()}
}

// --- Request / Response types ---

type createGiftRequest struct {
	Name        string   `json:"name"        validate:"required"`
	Description *string  `json:"description"`
	ImageURL    *string  `json:"image_url"`
	StoreURL    *string  `json:"store_url"`
	Price       *float64 `json:"price"`
}

type patchGiftRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	ImageURL    *string  `json:"image_url"`
	StoreURL    *string  `json:"store_url"`
	Price       *float64 `json:"price"`
}

type giftResponse struct {
	ID             string     `json:"id"`
	WeddingID      string     `json:"wedding_id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	ImageURL       *string    `json:"image_url,omitempty"`
	StoreURL       *string    `json:"store_url,omitempty"`
	Price          *float64   `json:"price,omitempty"`
	Status         string     `json:"status"`
	ReservedByName *string    `json:"reserved_by_name,omitempty"`
	ReservedAt     *time.Time `json:"reserved_at,omitempty"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
}

type giftSummaryResponse struct {
	Available int `json:"available"`
	Reserved  int `json:"reserved"`
}

// --- Handlers ---

// Create godoc
// @Summary Adicionar presente
// @Tags gifts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body createGiftRequest true "Dados do presente"
// @Success 201 {object} giftResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Failure 422 {object} validationEnvelope
// @Router /v1/gifts [post]
func (h *GiftHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req createGiftRequest
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

	gift, err := h.svc.CreateGift(r.Context(), userID, service.CreateGiftRequest{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		StoreURL:    req.StoreURL,
		Price:       req.Price,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusCreated, toGiftResponse(gift))
}

// List godoc
// @Summary Listar presentes
// @Tags gifts
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filtro por status (available, reserved)"
// @Success 200 {array} giftResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/gifts [get]
func (h *GiftHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var statusFilter *domain.GiftStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.GiftStatus(s)
		statusFilter = &st
	}

	gifts, err := h.svc.ListGifts(r.Context(), userID, statusFilter)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := make([]giftResponse, len(gifts))
	for i := range gifts {
		resp[i] = toGiftResponse(&gifts[i])
	}
	JSON(w, http.StatusOK, resp)
}

// Summary godoc
// @Summary Resumo de presentes por status
// @Tags gifts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} giftSummaryResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/gifts/summary [get]
func (h *GiftHandler) Summary(w http.ResponseWriter, r *http.Request) {
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

	JSON(w, http.StatusOK, giftSummaryResponse{
		Available: counts[domain.GiftAvailable],
		Reserved:  counts[domain.GiftReserved],
	})
}

// Update godoc
// @Summary Atualizar presente
// @Tags gifts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param giftID path string true "ID do presente"
// @Param body body patchGiftRequest true "Campos a atualizar"
// @Success 200 {object} giftResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/gifts/{giftID} [patch]
func (h *GiftHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	giftID := chi.URLParam(r, "giftID")

	var req patchGiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	gift, err := h.svc.UpdateGift(r.Context(), userID, giftID, service.UpdateGiftRequest{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		StoreURL:    req.StoreURL,
		Price:       req.Price,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, toGiftResponse(gift))
}

// Delete godoc
// @Summary Remover presente
// @Tags gifts
// @Security BearerAuth
// @Param giftID path string true "ID do presente"
// @Success 204
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/gifts/{giftID} [delete]
func (h *GiftHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	giftID := chi.URLParam(r, "giftID")

	if err := h.svc.DeleteGift(r.Context(), userID, giftID); err != nil {
		h.handleError(w, err)
		return
	}

	NoContent(w)
}

// CancelReserve godoc
// @Summary Cancelar reserva de presente
// @Tags gifts
// @Security BearerAuth
// @Param giftID path string true "ID do presente"
// @Success 200 {object} giftResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/gifts/{giftID}/reserve [delete]
func (h *GiftHandler) CancelReserve(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	giftID := chi.URLParam(r, "giftID")

	gift, err := h.svc.CancelReserve(r.Context(), userID, giftID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, toGiftResponse(gift))
}

func (h *GiftHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrConflict):
		Error(w, http.StatusConflict, "conflict", "gift already reserved")
	case errors.Is(err, domain.ErrForbidden):
		Error(w, http.StatusForbidden, "forbidden", "access denied")
	case errors.Is(err, domain.ErrValidation):
		Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	default:
		Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func toGiftResponse(g *domain.Gift) giftResponse {
	return giftResponse{
		ID:             g.ID,
		WeddingID:      g.WeddingID,
		Name:           g.Name,
		Description:    g.Description,
		ImageURL:       g.ImageURL,
		StoreURL:       g.StoreURL,
		Price:          g.Price,
		Status:         string(g.Status),
		ReservedByName: g.ReservedByName,
		ReservedAt:     g.ReservedAt,
		CreatedAt:      g.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      g.UpdatedAt.Format(time.RFC3339),
	}
}
