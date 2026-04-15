package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/ropehapi/wedding-mc/internal/domain"
)

const maxPhotoSize = 10 * 1024 * 1024 // 10 MB

// CreateWeddingRequest holds the data for creating a new wedding.
type CreateWeddingRequest struct {
	BrideName   string
	GroomName   string
	Date        time.Time
	Time        *string
	Location    string
	City        *string
	State       *string
	Description *string
	Links       []domain.WeddingLink
}

// UpdateWeddingRequest holds the fields that can be changed on an existing wedding.
// A nil pointer means "leave this field unchanged"; a non-nil pointer (even to a zero
// value) means "replace with this value". Links follow the same rule:
// nil → keep existing; pointer to empty slice → clear all links.
type UpdateWeddingRequest struct {
	BrideName   *string
	GroomName   *string
	Date        *time.Time
	Time        *string
	Location    *string
	City        *string
	State       *string
	Description *string
	Links       *[]domain.WeddingLink
}

// WeddingService defines the wedding business logic contract.
type WeddingService interface {
	CreateWedding(ctx context.Context, userID string, req CreateWeddingRequest) (*domain.Wedding, error)
	GetWedding(ctx context.Context, userID string) (*domain.Wedding, error)
	UpdateWedding(ctx context.Context, userID string, req UpdateWeddingRequest) (*domain.Wedding, error)
	UploadPhoto(ctx context.Context, userID, filename string, r io.Reader, size int64) (*domain.WeddingPhoto, error)
	DeletePhoto(ctx context.Context, userID, photoID string) error
}

type weddingService struct {
	weddings domain.WeddingRepository
	storage  StorageService
}

func NewWeddingService(weddings domain.WeddingRepository, storage StorageService) WeddingService {
	return &weddingService{weddings: weddings, storage: storage}
}

func (s *weddingService) CreateWedding(ctx context.Context, userID string, req CreateWeddingRequest) (*domain.Wedding, error) {
	slug, err := s.uniqueSlug(ctx, req.BrideName, req.GroomName)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}

	if req.Date.Before(time.Now()) {
		log.Warn().Str("user_id", userID).Msg("creating wedding with a past date")
	}

	w := &domain.Wedding{
		UserID:      userID,
		Slug:        slug,
		BrideName:   req.BrideName,
		GroomName:   req.GroomName,
		Date:        req.Date,
		Time:        req.Time,
		Location:    req.Location,
		City:        req.City,
		State:       req.State,
		Description: req.Description,
	}
	if err := s.weddings.Create(ctx, w); err != nil {
		return nil, err
	}

	if len(req.Links) > 0 {
		if err := s.weddings.ReplaceLinks(ctx, w.ID, req.Links); err != nil {
			return nil, fmt.Errorf("save links: %w", err)
		}
		w.Links = req.Links
	}

	return w, nil
}

func (s *weddingService) GetWedding(ctx context.Context, userID string) (*domain.Wedding, error) {
	return s.weddings.FindByUserID(ctx, userID)
}

func (s *weddingService) UpdateWedding(ctx context.Context, userID string, req UpdateWeddingRequest) (*domain.Wedding, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.BrideName != nil {
		w.BrideName = *req.BrideName
	}
	if req.GroomName != nil {
		w.GroomName = *req.GroomName
	}
	if req.Date != nil {
		w.Date = *req.Date
	}
	if req.Time != nil {
		w.Time = req.Time
	}
	if req.Location != nil {
		w.Location = *req.Location
	}
	if req.City != nil {
		w.City = req.City
	}
	if req.State != nil {
		w.State = req.State
	}
	if req.Description != nil {
		w.Description = req.Description
	}

	if err := s.weddings.Update(ctx, w); err != nil {
		return nil, err
	}

	if req.Links != nil {
		if err := s.weddings.ReplaceLinks(ctx, w.ID, *req.Links); err != nil {
			return nil, fmt.Errorf("replace links: %w", err)
		}
		w.Links = *req.Links
	}

	return w, nil
}

func (s *weddingService) UploadPhoto(ctx context.Context, userID, filename string, r io.Reader, size int64) (*domain.WeddingPhoto, error) {
	if size > maxPhotoSize {
		return nil, fmt.Errorf("%w: file exceeds 10MB limit", domain.ErrValidation)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	contentType, ok := extToContentType(ext)
	if !ok {
		return nil, fmt.Errorf("%w: unsupported file type %q (accepted: .jpg, .jpeg, .png, .webp)", domain.ErrValidation, ext)
	}

	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("weddings/%s/%s%s", w.ID, uuid.New().String(), ext)
	publicURL, err := s.storage.Upload(ctx, key, r, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload photo: %w", err)
	}

	photo := &domain.WeddingPhoto{
		WeddingID:  w.ID,
		URL:        publicURL,
		StorageKey: key,
	}
	if err := s.weddings.AddPhoto(ctx, photo); err != nil {
		_ = s.storage.Delete(ctx, key) // best-effort cleanup on DB failure
		return nil, fmt.Errorf("save photo record: %w", err)
	}

	return photo, nil
}

func (s *weddingService) DeletePhoto(ctx context.Context, userID, photoID string) error {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	photo, err := s.weddings.FindPhotoByID(ctx, photoID)
	if err != nil {
		return err
	}
	if photo.WeddingID != w.ID {
		return domain.ErrNotFound // don't reveal that the photo exists for another wedding
	}

	if _, err := s.weddings.DeletePhoto(ctx, photoID); err != nil {
		return err
	}

	if err := s.storage.Delete(ctx, photo.StorageKey); err != nil {
		log.Warn().Err(err).Str("key", photo.StorageKey).Msg("failed to delete photo from storage")
	}

	return nil
}

// uniqueSlug generates a URL-safe slug from bride and groom names and ensures
// it is unique in the database, appending a numeric suffix on collision.
func (s *weddingService) uniqueSlug(ctx context.Context, bride, groom string) (string, error) {
	base := slugify(bride) + "-e-" + slugify(groom)
	candidate := base
	for i := 2; i <= 100; i++ {
		_, err := s.weddings.FindBySlug(ctx, candidate)
		if errors.Is(err, domain.ErrNotFound) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("could not generate unique slug after 100 attempts")
}

var (
	// accentReplacer maps common Portuguese/Spanish accented chars to their ASCII equivalents.
	accentReplacer = strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
		"Á", "a", "À", "a", "Ã", "a", "Â", "a", "Ä", "a",
		"É", "e", "È", "e", "Ê", "e", "Ë", "e",
		"Í", "i", "Ì", "i", "Î", "i", "Ï", "i",
		"Ó", "o", "Ò", "o", "Õ", "o", "Ô", "o", "Ö", "o",
		"Ú", "u", "Ù", "u", "Û", "u", "Ü", "u",
		"Ç", "c", "Ñ", "n",
	)

	nonAlphanumericRE = regexp.MustCompile(`[^a-z0-9]+`)
)

// slugify converts a name to a lowercase, hyphen-separated, ASCII-safe slug.
func slugify(s string) string {
	s = accentReplacer.Replace(s)
	s = strings.ToLower(s)
	s = nonAlphanumericRE.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func extToContentType(ext string) (string, bool) {
	m := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
	}
	ct, ok := m[ext]
	return ct, ok
}
