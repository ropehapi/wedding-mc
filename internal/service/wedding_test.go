package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ropehapi/wedding-mc/internal/domain"
)

func newTestWeddingService(repo *mockWeddingRepo, storage *mockStorage) WeddingService {
	return NewWeddingService(repo, storage)
}

func baseCreateReq() CreateWeddingRequest {
	return CreateWeddingRequest{
		BrideName: "Ana",
		GroomName: "João",
		Date:      time.Now().Add(30 * 24 * time.Hour), // future date
		Location:  "Buffet Royal",
	}
}

// ---- Slug generation ----

func TestSlugify_BasicNames(t *testing.T) {
	cases := []struct {
		bride, groom string
		want         string
	}{
		{"Ana", "João", "ana-e-joao"},
		{"Maria Luíza", "Carlos", "maria-luiza-e-carlos"},
		{"Fernanda", "André", "fernanda-e-andre"},
		{"BRUNA", "TIAGO", "bruna-e-tiago"},
		{"Ção", "Açaí", "cao-e-acai"},
	}
	for _, tc := range cases {
		got := slugify(tc.bride) + "-e-" + slugify(tc.groom)
		if got != tc.want {
			t.Errorf("slugify(%q, %q): got %q, want %q", tc.bride, tc.groom, got, tc.want)
		}
	}
}

func TestSlugify_OnlyAllowsAlphanumericAndHyphens(t *testing.T) {
	result := slugify("Ana & Maria-José!")
	for _, ch := range result {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			t.Errorf("slugify produced invalid character %q in %q", ch, result)
		}
	}
}

func TestSlugify_NoLeadingOrTrailingHyphens(t *testing.T) {
	result := slugify("   Ana   ")
	if strings.HasPrefix(result, "-") || strings.HasSuffix(result, "-") {
		t.Errorf("slug should not have leading/trailing hyphens: %q", result)
	}
}

func TestCreateWedding_SlugCollision_AppendsNumericSuffix(t *testing.T) {
	// "ana-e-joao" is already taken; "ana-e-joao-2" is also taken; "ana-e-joao-3" is free.
	repo := &mockWeddingRepo{
		findBySlugResults: map[string]*domain.Wedding{
			"ana-e-joao":   {ID: "existing-1"},
			"ana-e-joao-2": {ID: "existing-2"},
		},
	}
	svc := newTestWeddingService(repo, &mockStorage{uploadURL: "http://x"})

	req := baseCreateReq()
	req.BrideName = "Ana"
	req.GroomName = "João"
	w, err := svc.CreateWedding(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Slug != "ana-e-joao-3" {
		t.Errorf("slug: got %q, want ana-e-joao-3", w.Slug)
	}
}

// ---- CreateWedding ----

func TestCreateWedding_Success(t *testing.T) {
	repo := &mockWeddingRepo{}
	svc := newTestWeddingService(repo, &mockStorage{})

	w, err := svc.CreateWedding(context.Background(), "user-1", baseCreateReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID == "" {
		t.Error("ID should be set after creation")
	}
	if w.Slug == "" {
		t.Error("Slug should be set")
	}
	if repo.created == nil {
		t.Error("Create was not called on repository")
	}
}

func TestCreateWedding_Conflict(t *testing.T) {
	repo := &mockWeddingRepo{createErr: domain.ErrConflict}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.CreateWedding(context.Background(), "user-1", baseCreateReq())
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestCreateWedding_WithLinks_ReplacesCalled(t *testing.T) {
	repo := &mockWeddingRepo{}
	svc := newTestWeddingService(repo, &mockStorage{})

	req := baseCreateReq()
	req.Links = []domain.WeddingLink{
		{Label: "Buffet", URL: "https://buffet.com"},
		{Label: "Fotógrafo", URL: "https://foto.com"},
	}
	_, err := svc.CreateWedding(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.replacedLinks) != 2 {
		t.Errorf("ReplaceLinks: got %d links, want 2", len(repo.replacedLinks))
	}
}

// ---- GetWedding ----

func TestGetWedding_ReturnsWeddingForUser(t *testing.T) {
	expected := &domain.Wedding{ID: "w-1", UserID: "user-1", BrideName: "Ana"}
	repo := &mockWeddingRepo{findByUserID: expected}
	svc := newTestWeddingService(repo, &mockStorage{})

	w, err := svc.GetWedding(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.ID != "w-1" {
		t.Errorf("ID: got %q, want w-1", w.ID)
	}
}

func TestGetWedding_NotFound(t *testing.T) {
	repo := &mockWeddingRepo{findByUserIDErr: domain.ErrNotFound}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.GetWedding(context.Background(), "user-999")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---- UpdateWedding ----

func TestUpdateWedding_PartialUpdate_OnlyChangedFields(t *testing.T) {
	newCity := "São Paulo"
	existing := &domain.Wedding{
		ID: "w-1", UserID: "user-1",
		BrideName: "Ana", GroomName: "João",
		Location: "Buffet A", City: ptr("Campinas"),
	}
	repo := &mockWeddingRepo{findByUserID: existing}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.UpdateWedding(context.Background(), "user-1", UpdateWeddingRequest{
		City: &newCity,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updated.City == nil || *repo.updated.City != "São Paulo" {
		t.Errorf("City: got %v, want São Paulo", repo.updated.City)
	}
	if repo.updated.BrideName != "Ana" {
		t.Errorf("BrideName should be unchanged: got %q", repo.updated.BrideName)
	}
}

func TestUpdateWedding_NotFound(t *testing.T) {
	repo := &mockWeddingRepo{findByUserIDErr: domain.ErrNotFound}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.UpdateWedding(context.Background(), "user-999", UpdateWeddingRequest{})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateWedding_ClearsLinksWhenEmptySliceSent(t *testing.T) {
	existing := &domain.Wedding{
		ID: "w-1",
		Links: []domain.WeddingLink{{Label: "X", URL: "https://x.com"}},
	}
	repo := &mockWeddingRepo{findByUserID: existing}
	svc := newTestWeddingService(repo, &mockStorage{})

	emptyLinks := []domain.WeddingLink{}
	_, err := svc.UpdateWedding(context.Background(), "user-1", UpdateWeddingRequest{
		Links: &emptyLinks,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.replacedLinks == nil || len(repo.replacedLinks) != 0 {
		t.Errorf("expected ReplaceLinks called with empty slice, got %v", repo.replacedLinks)
	}
}

func TestUpdateWedding_LinksNil_DoesNotCallReplaceLinks(t *testing.T) {
	existing := &domain.Wedding{ID: "w-1"}
	repo := &mockWeddingRepo{findByUserID: existing}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.UpdateWedding(context.Background(), "user-1", UpdateWeddingRequest{}) // Links == nil
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.replacedLinks != nil {
		t.Error("ReplaceLinks should not be called when Links is nil in request")
	}
}

// ---- UploadPhoto ----

func TestUploadPhoto_Success(t *testing.T) {
	repo := &mockWeddingRepo{findByUserID: &domain.Wedding{ID: "w-1"}}
	storage := &mockStorage{uploadURL: "http://localhost/uploads/weddings/w-1/abc.jpg"}
	svc := newTestWeddingService(repo, storage)

	photo, err := svc.UploadPhoto(context.Background(), "user-1", "photo.jpg",
		bytes.NewReader([]byte("fake")), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if photo.URL == "" {
		t.Error("photo URL should be set")
	}
	if photo.StorageKey == "" {
		t.Error("storage key should be set")
	}
	if repo.addedPhoto == nil {
		t.Error("AddPhoto should have been called")
	}
}

func TestUploadPhoto_ExceedsMaxSize(t *testing.T) {
	repo := &mockWeddingRepo{findByUserID: &domain.Wedding{ID: "w-1"}}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.UploadPhoto(context.Background(), "user-1", "big.jpg",
		bytes.NewReader([]byte("data")), maxPhotoSize+1)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for oversized file, got %v", err)
	}
}

func TestUploadPhoto_UnsupportedExtension(t *testing.T) {
	repo := &mockWeddingRepo{findByUserID: &domain.Wedding{ID: "w-1"}}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.UploadPhoto(context.Background(), "user-1", "document.pdf",
		bytes.NewReader([]byte("data")), 1024)
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("expected ErrValidation for unsupported ext, got %v", err)
	}
}

func TestUploadPhoto_StorageFailure_CleansUp(t *testing.T) {
	repo := &mockWeddingRepo{
		findByUserID: &domain.Wedding{ID: "w-1"},
		addPhotoErr:  errors.New("db error"),
	}
	storage := &mockStorage{uploadURL: "http://localhost/uploads/key.jpg"}
	svc := newTestWeddingService(repo, storage)

	_, err := svc.UploadPhoto(context.Background(), "user-1", "photo.jpg",
		bytes.NewReader([]byte("data")), 4)
	if err == nil {
		t.Fatal("expected error when AddPhoto fails")
	}
	if storage.deletedKey == "" {
		t.Error("storage cleanup (Delete) should be called when AddPhoto fails")
	}
}

func TestUploadPhoto_AcceptedExtensions(t *testing.T) {
	for _, ext := range []string{"photo.jpg", "photo.jpeg", "photo.png", "photo.webp"} {
		t.Run(ext, func(t *testing.T) {
			repo := &mockWeddingRepo{findByUserID: &domain.Wedding{ID: "w-1"}}
			storage := &mockStorage{uploadURL: "http://localhost/key"}
			svc := newTestWeddingService(repo, storage)

			_, err := svc.UploadPhoto(context.Background(), "user-1", ext,
				bytes.NewReader([]byte("x")), 1)
			if err != nil {
				t.Errorf("should accept %q: %v", ext, err)
			}
		})
	}
}

// ---- DeletePhoto ----

func TestDeletePhoto_Success(t *testing.T) {
	photo := &domain.WeddingPhoto{ID: "p-1", WeddingID: "w-1", StorageKey: "weddings/w-1/abc.jpg"}
	repo := &mockWeddingRepo{
		findByUserID:      &domain.Wedding{ID: "w-1"},
		findPhotoByID:     photo,
		deletePhotoResult: photo,
	}
	storage := &mockStorage{}
	svc := newTestWeddingService(repo, storage)

	if err := svc.DeletePhoto(context.Background(), "user-1", "p-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.deletedKey != photo.StorageKey {
		t.Errorf("storage key deleted: got %q, want %q", storage.deletedKey, photo.StorageKey)
	}
}

func TestDeletePhoto_PhotoBelongsToAnotherWedding(t *testing.T) {
	photo := &domain.WeddingPhoto{ID: "p-1", WeddingID: "w-OTHER"}
	repo := &mockWeddingRepo{
		findByUserID:  &domain.Wedding{ID: "w-1"},
		findPhotoByID: photo,
	}
	svc := newTestWeddingService(repo, &mockStorage{})

	err := svc.DeletePhoto(context.Background(), "user-1", "p-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeletePhoto_WeddingNotFound(t *testing.T) {
	repo := &mockWeddingRepo{findByUserIDErr: domain.ErrNotFound}
	svc := newTestWeddingService(repo, &mockStorage{})

	err := svc.DeletePhoto(context.Background(), "user-999", "p-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateWedding_AllFieldsSet(t *testing.T) {
	existing := &domain.Wedding{ID: "w-1"}
	repo := &mockWeddingRepo{findByUserID: existing}
	svc := newTestWeddingService(repo, &mockStorage{})

	newName := "Bruna"
	newGroom := "Tiago"
	newDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newTime := "18:00"
	newLoc := "Salão X"
	newCity := "SP"
	newState := "SP"
	newDesc := "Nossa história"

	w, err := svc.UpdateWedding(context.Background(), "user-1", UpdateWeddingRequest{
		BrideName:   &newName,
		GroomName:   &newGroom,
		Date:        &newDate,
		Time:        &newTime,
		Location:    &newLoc,
		City:        &newCity,
		State:       &newState,
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.BrideName != "Bruna" {
		t.Errorf("BrideName: got %q", w.BrideName)
	}
	if w.GroomName != "Tiago" {
		t.Errorf("GroomName: got %q", w.GroomName)
	}
	if w.Location != "Salão X" {
		t.Errorf("Location: got %q", w.Location)
	}
	if w.Time == nil || *w.Time != "18:00" {
		t.Errorf("Time: got %v", w.Time)
	}
}

func TestUpdateWedding_ReplaceLinksError(t *testing.T) {
	existing := &domain.Wedding{ID: "w-1"}
	repo := &mockWeddingRepo{
		findByUserID:    existing,
		replaceLinksErr: errors.New("db error"),
	}
	svc := newTestWeddingService(repo, &mockStorage{})

	links := []domain.WeddingLink{{Label: "X", URL: "https://x.com"}}
	_, err := svc.UpdateWedding(context.Background(), "user-1", UpdateWeddingRequest{
		Links: &links,
	})
	if err == nil {
		t.Fatal("expected error when ReplaceLinks fails")
	}
}

func TestCreateWedding_PastDate_StillCreates(t *testing.T) {
	repo := &mockWeddingRepo{}
	svc := newTestWeddingService(repo, &mockStorage{})

	req := baseCreateReq()
	req.Date = time.Now().Add(-24 * time.Hour) // past date

	w, err := svc.CreateWedding(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Error("should return created wedding even for past date")
	}
}

func TestCreateWedding_ReplaceLinksError(t *testing.T) {
	repo := &mockWeddingRepo{replaceLinksErr: errors.New("db error")}
	svc := newTestWeddingService(repo, &mockStorage{})

	req := baseCreateReq()
	req.Links = []domain.WeddingLink{{Label: "X", URL: "https://x.com"}}

	_, err := svc.CreateWedding(context.Background(), "user-1", req)
	if err == nil {
		t.Fatal("expected error when ReplaceLinks fails")
	}
}

func TestUniqueSlug_RepoError(t *testing.T) {
	repo := &mockWeddingRepo{findBySlugErr: errors.New("db error")}
	svc := newTestWeddingService(repo, &mockStorage{})

	_, err := svc.CreateWedding(context.Background(), "user-1", baseCreateReq())
	if err == nil {
		t.Fatal("expected error when FindBySlug returns a non-ErrNotFound error")
	}
}

func TestDeletePhoto_PhotoNotFound(t *testing.T) {
	repo := &mockWeddingRepo{
		findByUserID:     &domain.Wedding{ID: "w-1"},
		findPhotoByIDErr: domain.ErrNotFound,
	}
	svc := newTestWeddingService(repo, &mockStorage{})

	err := svc.DeletePhoto(context.Background(), "user-1", "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ---- helpers ----

func ptr[T any](v T) *T { return &v }
