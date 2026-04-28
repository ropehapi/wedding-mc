package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/ropehapi/wedding-mc/internal/domain"
)

// publicWeddingServicer is the subset of WeddingService used by PublicHandler.
type publicWeddingServicer interface {
	GetWeddingBySlug(ctx context.Context, slug string) (*domain.Wedding, error)
}

// publicGuestServicer is the subset of GuestService used by PublicHandler.
type publicGuestServicer interface {
	ListGuestsByWeddingID(ctx context.Context, weddingID string, status *domain.RSVPStatus) ([]domain.Guest, error)
	RSVP(ctx context.Context, slug, guestID, accessCode string, status domain.RSVPStatus) (*domain.Guest, error)
	ValidateGuestCode(ctx context.Context, slug, guestID, accessCode string) (*domain.Guest, error)
	LookupGuestByCode(ctx context.Context, slug, accessCode string) (*domain.Guest, error)
}

// publicGiftServicer is the subset of GiftService used by PublicHandler.
type publicGiftServicer interface {
	ListGiftsByWeddingID(ctx context.Context, weddingID string, status *domain.GiftStatus) ([]domain.Gift, error)
	ReserveGift(ctx context.Context, slug, giftID, guestName string) error
}

// PublicHandler handles public (unauthenticated) HTTP endpoints.
type PublicHandler struct {
	weddings publicWeddingServicer
	guests   publicGuestServicer
	gifts    publicGiftServicer
	validate *validator.Validate
}

func NewPublicHandler(weddings publicWeddingServicer, guests publicGuestServicer, gifts publicGiftServicer) *PublicHandler {
	return &PublicHandler{
		weddings: weddings,
		guests:   guests,
		gifts:    gifts,
		validate: validator.New(),
	}
}

// --- Response types (safe for public exposure — no sensitive fields) ---

type publicWeddingResponse struct {
	ID          string          `json:"id"`
	Slug        string          `json:"slug"`
	BrideName   string          `json:"bride_name"`
	GroomName   string          `json:"groom_name"`
	Date        string          `json:"date"`
	Time        *string         `json:"time,omitempty"`
	Location    string          `json:"location"`
	City        *string         `json:"city,omitempty"`
	State       *string         `json:"state,omitempty"`
	Description *string         `json:"description,omitempty"`
	Photos      []photoResponse `json:"photos"`
	Links       []linkResponse  `json:"links"`
}

type publicGuestResponse struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Status string     `json:"status"`
	RSVPAt *time.Time `json:"rsvp_at,omitempty"`
}

// publicGiftResponse intentionally omits reserved_by_name — uses a bool instead.
type publicGiftResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
	StoreURL    *string  `json:"store_url,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Reserved    bool     `json:"reserved"`
}

// --- Request types ---

type rsvpRequest struct {
	Status     string `json:"status" validate:"required,oneof=confirmed declined"`
	AccessCode string `json:"access_code" validate:"required,len=6"`
}

type reserveGiftRequest struct {
	GuestID    string `json:"guest_id" validate:"required"`
	AccessCode string `json:"access_code" validate:"required,len=6"`
}

type validateCodeRequest struct {
	AccessCode string `json:"access_code" validate:"required,len=6"`
}

// --- Handlers ---

// GetWedding godoc
// @Summary Página pública do casamento
// @Tags public
// @Produce json
// @Param slug path string true "Slug do casamento"
// @Success 200 {object} publicWeddingResponse
// @Failure 404 {object} errorEnvelope
// @Router /v1/public/{slug} [get]
func (h *PublicHandler) GetWedding(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	wedding, err := h.weddings.GetWeddingBySlug(r.Context(), slug)
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, toPublicWeddingResponse(wedding))
}

// ListGuests godoc
// @Summary Lista de convidados pública
// @Tags public
// @Produce json
// @Param slug path string true "Slug do casamento"
// @Success 200 {array} publicGuestResponse
// @Failure 404 {object} errorEnvelope
// @Router /v1/public/{slug}/guests [get]
func (h *PublicHandler) ListGuests(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	wedding, err := h.weddings.GetWeddingBySlug(r.Context(), slug)
	if err != nil {
		h.handleError(w, err)
		return
	}

	guests, err := h.guests.ListGuestsByWeddingID(r.Context(), wedding.ID, nil)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := make([]publicGuestResponse, len(guests))
	for i, g := range guests {
		resp[i] = publicGuestResponse{
			ID:     g.ID,
			Name:   g.Name,
			Status: string(g.Status),
			RSVPAt: g.RSVPAt,
		}
	}
	JSON(w, http.StatusOK, resp)
}

// RSVP godoc
// @Summary Confirmar/recusar presença
// @Tags public
// @Accept json
// @Produce json
// @Param slug path string true "Slug do casamento"
// @Param guestID path string true "ID do convidado"
// @Param body body rsvpRequest true "Status RSVP"
// @Success 200 {object} publicGuestResponse
// @Failure 404 {object} errorEnvelope
// @Failure 422 {object} validationEnvelope
// @Router /v1/public/{slug}/guests/{guestID}/rsvp [post]
func (h *PublicHandler) RSVP(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	guestID := chi.URLParam(r, "guestID")

	var req rsvpRequest
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

	guest, err := h.guests.RSVP(r.Context(), slug, guestID, req.AccessCode, domain.RSVPStatus(req.Status))
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, publicGuestResponse{
		ID:     guest.ID,
		Name:   guest.Name,
		Status: string(guest.Status),
		RSVPAt: guest.RSVPAt,
	})
}

// ValidateCode godoc
// @Summary Valida o código de acesso de um convidado
// @Tags public
// @Accept json
// @Produce json
// @Param slug path string true "Slug do casamento"
// @Param body body validateCodeRequest true "Código de acesso"
// @Success 200 {object} publicGuestResponse
// @Failure 403 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/public/{slug}/guests/validate-code [post]
func (h *PublicHandler) ValidateCode(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req validateCodeRequest
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

	guest, err := h.guests.LookupGuestByCode(r.Context(), slug, req.AccessCode)
	if err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, publicGuestResponse{
		ID:     guest.ID,
		Name:   guest.Name,
		Status: string(guest.Status),
		RSVPAt: guest.RSVPAt,
	})
}

// ListGifts godoc
// @Summary Lista pública de presentes
// @Tags public
// @Produce json
// @Param slug path string true "Slug do casamento"
// @Success 200 {array} publicGiftResponse
// @Failure 404 {object} errorEnvelope
// @Router /v1/public/{slug}/gifts [get]
func (h *PublicHandler) ListGifts(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	wedding, err := h.weddings.GetWeddingBySlug(r.Context(), slug)
	if err != nil {
		h.handleError(w, err)
		return
	}

	gifts, err := h.gifts.ListGiftsByWeddingID(r.Context(), wedding.ID, nil)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := make([]publicGiftResponse, len(gifts))
	for i, g := range gifts {
		resp[i] = publicGiftResponse{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			ImageURL:    g.ImageURL,
			StoreURL:    g.StoreURL,
			Price:       g.Price,
			Reserved:    g.Status == domain.GiftReserved,
		}
	}
	JSON(w, http.StatusOK, resp)
}

// ReserveGift godoc
// @Summary Reservar presente
// @Tags public
// @Accept json
// @Produce json
// @Param slug path string true "Slug do casamento"
// @Param giftID path string true "ID do presente"
// @Param body body reserveGiftRequest true "Nome do convidado"
// @Success 200
// @Failure 404 {object} errorEnvelope
// @Failure 409 {object} errorEnvelope
// @Failure 422 {object} validationEnvelope
// @Router /v1/public/{slug}/gifts/{giftID}/reserve [post]
func (h *PublicHandler) ReserveGift(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	giftID := chi.URLParam(r, "giftID")

	var req reserveGiftRequest
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

	guest, err := h.guests.ValidateGuestCode(r.Context(), slug, req.GuestID, req.AccessCode)
	if err != nil {
		h.handleError(w, err)
		return
	}

	if err := h.gifts.ReserveGift(r.Context(), slug, giftID, guest.Name); err != nil {
		h.handleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "reserved"})
}

func (h *PublicHandler) handleError(w http.ResponseWriter, err error) {
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

func toPublicWeddingResponse(w *domain.Wedding) publicWeddingResponse {
	photos := make([]photoResponse, len(w.Photos))
	for i, p := range w.Photos {
		photos[i] = photoResponse{
			ID:        p.ID,
			URL:       p.URL,
			IsCover:   p.IsCover,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
		}
	}
	links := make([]linkResponse, len(w.Links))
	for i, l := range w.Links {
		links[i] = linkResponse{
			ID:       l.ID,
			Label:    l.Label,
			URL:      l.URL,
			Position: l.Position,
		}
	}
	return publicWeddingResponse{
		ID:          w.ID,
		Slug:        w.Slug,
		BrideName:   w.BrideName,
		GroomName:   w.GroomName,
		Date:        w.Date.Format("2006-01-02"),
		Time:        w.Time,
		Location:    w.Location,
		City:        w.City,
		State:       w.State,
		Description: w.Description,
		Photos:      photos,
		Links:       links,
	}
}
