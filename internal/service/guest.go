package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ropehapi/wedding-mc/internal/domain"
)

// GuestService defines the guest business logic contract.
type GuestService interface {
	CreateGuest(ctx context.Context, userID string, name string) (*domain.Guest, error)
	ListGuests(ctx context.Context, userID string, status *domain.RSVPStatus) ([]domain.Guest, error)
	ListGuestsByWeddingID(ctx context.Context, weddingID string, status *domain.RSVPStatus) ([]domain.Guest, error)
	UpdateGuest(ctx context.Context, userID, guestID, name string) (*domain.Guest, error)
	DeleteGuest(ctx context.Context, userID, guestID string) error
	RSVP(ctx context.Context, slug, guestID, accessCode string, status domain.RSVPStatus) (*domain.Guest, error)
	ValidateGuestCode(ctx context.Context, slug, guestID, accessCode string) (*domain.Guest, error)
	LookupGuestByCode(ctx context.Context, slug, accessCode string) (*domain.Guest, error)
	GetSummary(ctx context.Context, userID string) (map[domain.RSVPStatus]int, error)
}

type guestService struct {
	guests   domain.GuestRepository
	weddings domain.WeddingRepository
}

func NewGuestService(guests domain.GuestRepository, weddings domain.WeddingRepository) GuestService {
	return &guestService{guests: guests, weddings: weddings}
}

func (s *guestService) CreateGuest(ctx context.Context, userID string, name string) (*domain.Guest, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	g := &domain.Guest{
		WeddingID: w.ID,
		Name:      name,
		Status:    domain.RSVPPending,
	}
	if err := s.guests.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *guestService) ListGuests(ctx context.Context, userID string, status *domain.RSVPStatus) ([]domain.Guest, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.guests.FindAll(ctx, w.ID, status)
}

func (s *guestService) ListGuestsByWeddingID(ctx context.Context, weddingID string, status *domain.RSVPStatus) ([]domain.Guest, error) {
	return s.guests.FindAll(ctx, weddingID, status)
}

func (s *guestService) UpdateGuest(ctx context.Context, userID, guestID, name string) (*domain.Guest, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	g, err := s.guests.FindByID(ctx, guestID)
	if err != nil {
		return nil, err
	}
	if g.WeddingID != w.ID {
		return nil, domain.ErrNotFound
	}

	g.Name = name
	if err := s.guests.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *guestService) DeleteGuest(ctx context.Context, userID, guestID string) error {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	g, err := s.guests.FindByID(ctx, guestID)
	if err != nil {
		return err
	}
	if g.WeddingID != w.ID {
		return domain.ErrNotFound
	}

	return s.guests.Delete(ctx, guestID)
}

// RSVP updates a guest's attendance status via the public slug endpoint.
func (s *guestService) RSVP(ctx context.Context, slug, guestID, accessCode string, status domain.RSVPStatus) (*domain.Guest, error) {
	if status != domain.RSVPConfirmed && status != domain.RSVPDeclined {
		return nil, fmt.Errorf("%w: status must be 'confirmed' or 'declined'", domain.ErrValidation)
	}

	w, err := s.weddings.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	g, err := s.guests.FindByID(ctx, guestID)
	if err != nil {
		return nil, err
	}
	if g.WeddingID != w.ID {
		return nil, domain.ErrNotFound
	}
	if g.AccessCode != accessCode {
		return nil, domain.ErrForbidden
	}

	now := time.Now()
	g.Status = status
	g.RSVPAt = &now
	if err := s.guests.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

// ValidateGuestCode finds the guest and validates its access code.
func (s *guestService) ValidateGuestCode(ctx context.Context, slug, guestID, accessCode string) (*domain.Guest, error) {
	w, err := s.weddings.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	g, err := s.guests.FindByID(ctx, guestID)
	if err != nil {
		return nil, err
	}
	if g.WeddingID != w.ID {
		return nil, domain.ErrNotFound
	}
	if g.AccessCode != accessCode {
		return nil, domain.ErrForbidden
	}
	return g, nil
}

// LookupGuestByCode finds a guest by access code within a wedding, without requiring a known guest ID.
func (s *guestService) LookupGuestByCode(ctx context.Context, slug, accessCode string) (*domain.Guest, error) {
	w, err := s.weddings.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	g, err := s.guests.FindByAccessCode(ctx, w.ID, accessCode)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *guestService) GetSummary(ctx context.Context, userID string) (map[domain.RSVPStatus]int, error) {
	w, err := s.weddings.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.guests.CountByStatus(ctx, w.ID)
}
