package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/ropehapi/wedding-mc/internal/domain"
	"github.com/ropehapi/wedding-mc/internal/middleware"
	"github.com/ropehapi/wedding-mc/internal/service"
)

// weddingServicer is the subset of service.WeddingService used by WeddingHandler.
type weddingServicer interface {
	CreateWedding(ctx context.Context, userID string, req service.CreateWeddingRequest) (*domain.Wedding, error)
	GetWedding(ctx context.Context, userID string) (*domain.Wedding, error)
	UpdateWedding(ctx context.Context, userID string, req service.UpdateWeddingRequest) (*domain.Wedding, error)
	UploadPhoto(ctx context.Context, userID, filename string, r io.Reader, size int64) (*domain.WeddingPhoto, error)
	DeletePhoto(ctx context.Context, userID, photoID string) error
}

// WeddingHandler handles HTTP requests for wedding endpoints.
type WeddingHandler struct {
	svc      weddingServicer
	validate *validator.Validate
}

func NewWeddingHandler(svc weddingServicer) *WeddingHandler {
	return &WeddingHandler{svc: svc, validate: validator.New()}
}

// --- Request / Response types ---

type linkRequest struct {
	Label string `json:"label" validate:"required"`
	URL   string `json:"url"   validate:"required,url"`
}

type createWeddingRequest struct {
	BrideName   string        `json:"bride_name"   validate:"required"`
	GroomName   string        `json:"groom_name"   validate:"required"`
	Date        string        `json:"date"         validate:"required"` // "2006-01-02"
	Time        *string       `json:"time"`
	Location    string        `json:"location"     validate:"required"`
	City        *string       `json:"city"`
	State       *string       `json:"state"`
	Description *string       `json:"description"`
	Links       []linkRequest `json:"links"`
}

// patchWeddingRequest supports true partial updates.
// A field absent from the JSON body will be nil and will not overwrite the stored value.
type patchWeddingRequest struct {
	BrideName   *string        `json:"bride_name"`
	GroomName   *string        `json:"groom_name"`
	Date        *string        `json:"date"` // "2006-01-02"
	Time        *string        `json:"time"`
	Location    *string        `json:"location"`
	City        *string        `json:"city"`
	State       *string        `json:"state"`
	Description *string        `json:"description"`
	Links       *[]linkRequest `json:"links"` // nil = unchanged; [] = clear all
}

type photoResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

type linkResponse struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	Position int    `json:"position"`
}

type weddingResponse struct {
	ID          string         `json:"id"`
	Slug        string         `json:"slug"`
	BrideName   string         `json:"bride_name"`
	GroomName   string         `json:"groom_name"`
	Date        string         `json:"date"`
	Time        *string        `json:"time,omitempty"`
	Location    string         `json:"location"`
	City        *string        `json:"city,omitempty"`
	State       *string        `json:"state,omitempty"`
	Description *string        `json:"description,omitempty"`
	Photos      []photoResponse `json:"photos"`
	Links       []linkResponse  `json:"links"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// --- Handlers ---

// Get godoc
// @Summary Consultar casamento
// @Tags wedding
// @Security BearerAuth
// @Produce json
// @Success 200 {object} weddingResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/wedding [get]
func (h *WeddingHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	wedding, err := h.svc.GetWedding(r.Context(), userID)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, toWeddingResponse(wedding))
}

// Create godoc
// @Summary Criar casamento
// @Tags wedding
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body createWeddingRequest true "Dados do casamento"
// @Success 201 {object} weddingResponse
// @Failure 401 {object} errorEnvelope
// @Failure 409 {object} errorEnvelope
// @Failure 422 {object} validationEnvelope
// @Router /v1/wedding [post]
func (h *WeddingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req createWeddingRequest
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
	for _, l := range req.Links {
		if errs := h.validate.Struct(l); errs != nil {
			var ve validator.ValidationErrors
			if errors.As(errs, &ve) {
				ValidationError(w, ve)
				return
			}
		}
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "validation_error", "date must be in YYYY-MM-DD format")
		return
	}

	links := toServiceLinks(req.Links)
	wedding, err := h.svc.CreateWedding(r.Context(), userID, service.CreateWeddingRequest{
		BrideName:   req.BrideName,
		GroomName:   req.GroomName,
		Date:        date,
		Time:        req.Time,
		Location:    req.Location,
		City:        req.City,
		State:       req.State,
		Description: req.Description,
		Links:       links,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSON(w, http.StatusCreated, toWeddingResponse(wedding))
}

// Update godoc
// @Summary Editar casamento
// @Tags wedding
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body patchWeddingRequest true "Campos a atualizar"
// @Success 200 {object} weddingResponse
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Failure 422 {object} validationEnvelope
// @Router /v1/wedding [patch]
func (h *WeddingHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req patchWeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	if req.Links != nil {
		for _, l := range *req.Links {
			if errs := h.validate.Struct(l); errs != nil {
				var ve validator.ValidationErrors
				if errors.As(errs, &ve) {
					ValidationError(w, ve)
					return
				}
			}
		}
	}

	svcReq := service.UpdateWeddingRequest{
		BrideName:   req.BrideName,
		GroomName:   req.GroomName,
		Time:        req.Time,
		Location:    req.Location,
		City:        req.City,
		State:       req.State,
		Description: req.Description,
	}

	if req.Date != nil {
		d, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			Error(w, http.StatusUnprocessableEntity, "validation_error", "date must be in YYYY-MM-DD format")
			return
		}
		svcReq.Date = &d
	}

	if req.Links != nil {
		domainLinks := toServiceLinks(*req.Links)
		svcReq.Links = &domainLinks
	}

	wedding, err := h.svc.UpdateWedding(r.Context(), userID, svcReq)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSON(w, http.StatusOK, toWeddingResponse(wedding))
}

// UploadPhoto godoc
// @Summary Upload de foto
// @Tags wedding
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Arquivo de imagem (JPEG, PNG, WebP, max 10MB)"
// @Success 201 {object} photoResponse
// @Failure 401 {object} errorEnvelope
// @Failure 422 {object} errorEnvelope
// @Router /v1/wedding/photos [post]
func (h *WeddingHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		Error(w, http.StatusBadRequest, "bad_request", "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, "validation_error", "field 'photo' is required")
		return
	}
	defer file.Close()

	photo, err := h.svc.UploadPhoto(r.Context(), userID, header.Filename, file, header.Size)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	JSON(w, http.StatusCreated, photoResponse{
		ID:        photo.ID,
		URL:       photo.URL,
		CreatedAt: photo.CreatedAt.Format(time.RFC3339),
	})
}

// DeletePhoto godoc
// @Summary Remover foto
// @Tags wedding
// @Security BearerAuth
// @Param photoID path string true "ID da foto"
// @Success 204
// @Failure 401 {object} errorEnvelope
// @Failure 404 {object} errorEnvelope
// @Router /v1/wedding/photos/{photoID} [delete]
func (h *WeddingHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	photoID := chi.URLParam(r, "photoID")
	if photoID == "" {
		Error(w, http.StatusBadRequest, "bad_request", "photoID is required")
		return
	}

	if err := h.svc.DeletePhoto(r.Context(), userID, photoID); err != nil {
		h.handleError(w, r, err)
		return
	}

	NoContent(w)
}

func (h *WeddingHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(w, http.StatusNotFound, "not_found", "wedding not found")
	case errors.Is(err, domain.ErrConflict):
		Error(w, http.StatusConflict, "conflict", "wedding already exists for this account")
	case errors.Is(err, domain.ErrForbidden):
		Error(w, http.StatusForbidden, "forbidden", "access denied")
	case errors.Is(err, domain.ErrValidation):
		Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	default:
		log.Error().Err(err).Str("request_id", middleware.RequestIDFromContext(r.Context())).Msg("unhandled wedding error")
		Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

// --- Converters ---

func toWeddingResponse(w *domain.Wedding) weddingResponse {
	photos := make([]photoResponse, len(w.Photos))
	for i, p := range w.Photos {
		photos[i] = photoResponse{
			ID:        p.ID,
			URL:       p.URL,
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

	return weddingResponse{
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
		CreatedAt:   w.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   w.UpdatedAt.Format(time.RFC3339),
	}
}

func toServiceLinks(reqs []linkRequest) []domain.WeddingLink {
	links := make([]domain.WeddingLink, len(reqs))
	for i, l := range reqs {
		links[i] = domain.WeddingLink{Label: l.Label, URL: l.URL}
	}
	return links
}
