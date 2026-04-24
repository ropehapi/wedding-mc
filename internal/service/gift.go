package service

import (
	"context"

	"github.com/ropehapi/wedding-mc/internal/domain"
)

// CreateGiftRequest holds the data for creating a new gift.
type CreateGiftRequest struct {
	Name        string
	Description *string
	ImageURL    *string
	StoreURL    *string
	Price       *float64
}

// UpdateGiftRequest holds the fields that can be changed on a gift.
type UpdateGiftRequest struct {
	Name        *string
	Description *string
	ImageURL    *string
	StoreURL    *string
	Price       *float64
}

// GiftService defines the gift business logic contract.
type GiftService interface {
	CreateGift(ctx context.Context, userID string, req CreateGiftRequest) (*domain.Gift, error)
	ListGifts(ctx context.Context, userID string, status *domain.GiftStatus) ([]domain.Gift, error)
	ListGiftsByWeddingID(ctx context.Context, weddingID string, status *domain.GiftStatus) ([]domain.Gift, error)
	UpdateGift(ctx context.Context, userID, giftID string, req UpdateGiftRequest) (*domain.Gift, error)
	DeleteGift(ctx context.Context, userID, giftID string) error
	ReserveGift(ctx context.Context, slug, giftID, guestName string) error
	CancelReserve(ctx context.Context, userID, giftID string) (*domain.Gift, error)
	GetSummary(ctx context.Context, userID string) (map[domain.GiftStatus]int, error)
}

type giftService struct {
	gifts    domain.GiftRepository
	weddings domain.WeddingRepository
}

func NewGiftService(gifts domain.GiftRepository, weddings domain.WeddingRepository) GiftService {
	return &giftService{gifts: gifts, weddings: weddings}
}

func (s *giftService) CreateGift(ctx context.Context, userID string, req CreateGiftRequest) (*domain.Gift, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	g := &domain.Gift{
		WeddingID:   w.ID,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		StoreURL:    req.StoreURL,
		Price:       req.Price,
	}
	if err := s.gifts.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *giftService) ListGifts(ctx context.Context, userID string, status *domain.GiftStatus) ([]domain.Gift, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.gifts.FindAll(ctx, w.ID, status)
}

func (s *giftService) ListGiftsByWeddingID(ctx context.Context, weddingID string, status *domain.GiftStatus) ([]domain.Gift, error) {
	return s.gifts.FindAll(ctx, weddingID, status)
}

func (s *giftService) UpdateGift(ctx context.Context, userID, giftID string, req UpdateGiftRequest) (*domain.Gift, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	g, err := s.gifts.FindByID(ctx, giftID)
	if err != nil {
		return nil, err
	}
	if g.WeddingID != w.ID {
		return nil, domain.ErrNotFound
	}

	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.Description != nil {
		g.Description = req.Description
	}
	if req.ImageURL != nil {
		g.ImageURL = req.ImageURL
	}
	if req.StoreURL != nil {
		g.StoreURL = req.StoreURL
	}
	if req.Price != nil {
		g.Price = req.Price
	}

	if err := s.gifts.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *giftService) DeleteGift(ctx context.Context, userID, giftID string) error {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	g, err := s.gifts.FindByID(ctx, giftID)
	if err != nil {
		return err
	}
	if g.WeddingID != w.ID {
		return domain.ErrNotFound
	}

	return s.gifts.Delete(ctx, giftID)
}

// ReserveGift reserves a gift atomically via the public slug endpoint.
func (s *giftService) ReserveGift(ctx context.Context, slug, giftID, guestName string) error {
	w, err := s.weddings.FindBySlug(ctx, slug)
	if err != nil {
		return err
	}

	g, err := s.gifts.FindByID(ctx, giftID)
	if err != nil {
		return err
	}
	if g.WeddingID != w.ID {
		return domain.ErrNotFound
	}

	return s.gifts.Reserve(ctx, giftID, guestName)
}

// CancelReserve cancels a gift reservation. Only the wedding owner can cancel.
func (s *giftService) CancelReserve(ctx context.Context, userID, giftID string) (*domain.Gift, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	g, err := s.gifts.FindByID(ctx, giftID)
	if err != nil {
		return nil, err
	}
	if g.WeddingID != w.ID {
		return nil, domain.ErrNotFound
	}

	if err := s.gifts.CancelReserve(ctx, giftID); err != nil {
		return nil, err
	}

	// Return updated state
	return s.gifts.FindByID(ctx, giftID)
}

func (s *giftService) GetSummary(ctx context.Context, userID string) (map[domain.GiftStatus]int, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.gifts.CountByStatus(ctx, w.ID)
}
