package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ropehapi/wedding-mc/internal/domain"
)

// mockGiftRepo is a test double for domain.GiftRepository.
type mockGiftRepo struct {
	createErr error
	created   *domain.Gift

	findAllResult []domain.Gift
	findAllErr    error

	findByIDResult *domain.Gift
	findByIDErr    error
	// findByIDSecond is returned on the second call (used by CancelReserve to re-fetch).
	findByIDSecond    *domain.Gift
	findByIDCallCount int

	updateErr error

	deleteErr error

	reserveErr error

	cancelReserveErr error

	countResult map[domain.GiftStatus]int
	countErr    error
}

func (m *mockGiftRepo) Create(_ context.Context, g *domain.Gift) error {
	if m.createErr != nil {
		return m.createErr
	}
	g.ID = "gift-id-1"
	m.created = g
	return nil
}

func (m *mockGiftRepo) FindAll(_ context.Context, _ string, _ *domain.GiftStatus) ([]domain.Gift, error) {
	return m.findAllResult, m.findAllErr
}

func (m *mockGiftRepo) FindByID(_ context.Context, _ string) (*domain.Gift, error) {
	m.findByIDCallCount++
	if m.findByIDCallCount > 1 && m.findByIDSecond != nil {
		return m.findByIDSecond, nil
	}
	return m.findByIDResult, m.findByIDErr
}

func (m *mockGiftRepo) Update(_ context.Context, _ *domain.Gift) error {
	return m.updateErr
}

func (m *mockGiftRepo) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockGiftRepo) Reserve(_ context.Context, _, _ string) error {
	return m.reserveErr
}

func (m *mockGiftRepo) CancelReserve(_ context.Context, _ string) error {
	return m.cancelReserveErr
}

func (m *mockGiftRepo) CountByStatus(_ context.Context, _ string) (map[domain.GiftStatus]int, error) {
	return m.countResult, m.countErr
}

func newTestGiftService(gifts *mockGiftRepo, weddings *mockWeddingRepo) GiftService {
	return NewGiftService(gifts, weddings)
}

// ---- CreateGift ----

func TestCreateGift_Success(t *testing.T) {
	w := baseWedding()
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{}
	svc := newTestGiftService(gifts, weddings)

	g, err := svc.CreateGift(context.Background(), "user-1", CreateGiftRequest{Name: "Panela"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.WeddingID != w.ID {
		t.Errorf("wrong wedding_id: %q", g.WeddingID)
	}
}

func TestCreateGift_WeddingNotFound(t *testing.T) {
	weddings := &mockWeddingRepo{findByUserIDErr: domain.ErrNotFound}
	svc := newTestGiftService(&mockGiftRepo{}, weddings)

	_, err := svc.CreateGift(context.Background(), "user-1", CreateGiftRequest{Name: "Panela"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---- ReserveGift ----

func TestReserveGift_Available(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: w.ID, Status: domain.GiftAvailable}
	weddings := &mockWeddingRepo{findBySlugResults: map[string]*domain.Wedding{w.Slug: w}}
	gifts := &mockGiftRepo{findByIDResult: gift}
	svc := newTestGiftService(gifts, weddings)

	err := svc.ReserveGift(context.Background(), w.Slug, "gift-1", "João")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReserveGift_AlreadyReserved(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: w.ID, Status: domain.GiftReserved}
	weddings := &mockWeddingRepo{findBySlugResults: map[string]*domain.Wedding{w.Slug: w}}
	gifts := &mockGiftRepo{
		findByIDResult: gift,
		reserveErr:     domain.ErrConflict,
	}
	svc := newTestGiftService(gifts, weddings)

	err := svc.ReserveGift(context.Background(), w.Slug, "gift-1", "João")
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestReserveGift_GiftFromAnotherWedding(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: "other-wedding"}
	weddings := &mockWeddingRepo{findBySlugResults: map[string]*domain.Wedding{w.Slug: w}}
	gifts := &mockGiftRepo{findByIDResult: gift}
	svc := newTestGiftService(gifts, weddings)

	err := svc.ReserveGift(context.Background(), w.Slug, "gift-1", "João")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestReserveGift_SlugNotFound(t *testing.T) {
	svc := newTestGiftService(&mockGiftRepo{}, &mockWeddingRepo{})

	err := svc.ReserveGift(context.Background(), "unknown-slug", "gift-1", "João")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---- CancelReserve ----

func TestCancelReserve_Success(t *testing.T) {
	w := baseWedding()
	reserved := &domain.Gift{ID: "gift-1", WeddingID: w.ID, Status: domain.GiftReserved}
	available := &domain.Gift{ID: "gift-1", WeddingID: w.ID, Status: domain.GiftAvailable}
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{findByIDResult: reserved, findByIDSecond: available}
	svc := newTestGiftService(gifts, weddings)

	result, err := svc.CancelReserve(context.Background(), "user-1", "gift-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.GiftAvailable {
		t.Errorf("expected available after cancel, got %q", result.Status)
	}
}

func TestCancelReserve_GiftFromAnotherWedding(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: "other-id"}
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{findByIDResult: gift}
	svc := newTestGiftService(gifts, weddings)

	_, err := svc.CancelReserve(context.Background(), "user-1", "gift-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---- GetSummary ----

func TestGiftGetSummary_CorrectCounts(t *testing.T) {
	w := baseWedding()
	counts := map[domain.GiftStatus]int{
		domain.GiftAvailable: 4,
		domain.GiftReserved:  2,
	}
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{countResult: counts}
	svc := newTestGiftService(gifts, weddings)

	result, err := svc.GetSummary(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[domain.GiftAvailable] != 4 {
		t.Errorf("available: want 4, got %d", result[domain.GiftAvailable])
	}
	if result[domain.GiftReserved] != 2 {
		t.Errorf("reserved: want 2, got %d", result[domain.GiftReserved])
	}
}

// ---- UpdateGift ----

func TestUpdateGift_PartialUpdate(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: w.ID, Name: "Old Name"}
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{findByIDResult: gift}
	svc := newTestGiftService(gifts, weddings)

	newName := "New Name"
	updated, err := svc.UpdateGift(context.Background(), "user-1", "gift-1", UpdateGiftRequest{Name: &newName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", updated.Name)
	}
}

func TestUpdateGift_GiftFromAnotherWedding(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: "other-id"}
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{findByIDResult: gift}
	svc := newTestGiftService(gifts, weddings)

	name := "Name"
	_, err := svc.UpdateGift(context.Background(), "user-1", "gift-1", UpdateGiftRequest{Name: &name})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---- DeleteGift ----

func TestDeleteGift_Success(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: w.ID}
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{findByIDResult: gift}
	svc := newTestGiftService(gifts, weddings)

	err := svc.DeleteGift(context.Background(), "user-1", "gift-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteGift_GiftFromAnotherWedding(t *testing.T) {
	w := baseWedding()
	gift := &domain.Gift{ID: "gift-1", WeddingID: "other-id"}
	weddings := &mockWeddingRepo{findByUserID: w}
	gifts := &mockGiftRepo{findByIDResult: gift}
	svc := newTestGiftService(gifts, weddings)

	err := svc.DeleteGift(context.Background(), "user-1", "gift-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
