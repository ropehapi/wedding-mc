package service

import (
	"context"
	"io"

	"github.com/ropehapi/wedding-mc/internal/domain"
)

// mockWeddingRepo is a test double for domain.WeddingRepository.
type mockWeddingRepo struct {
	createErr error
	created   *domain.Wedding

	findByUserID    *domain.Wedding
	findByUserIDErr error

	findBySlugResults map[string]*domain.Wedding // slug → wedding; nil entry means ErrNotFound
	findBySlugErr     error                      // generic error, overrides results

	updateErr error
	updated   *domain.Wedding

	addPhotoErr error
	addedPhoto  *domain.WeddingPhoto

	findPhotoByID    *domain.WeddingPhoto
	findPhotoByIDErr error

	deletePhotoResult *domain.WeddingPhoto
	deletePhotoErr    error

	setCoverPhotoErr error

	replaceLinksErr error
	replacedLinks   []domain.WeddingLink
}

func (m *mockWeddingRepo) Create(_ context.Context, w *domain.Wedding) error {
	if m.createErr != nil {
		return m.createErr
	}
	w.ID = "wedding-id-1"
	m.created = w
	return nil
}

func (m *mockWeddingRepo) FindByUserID(_ context.Context, _ string) (*domain.Wedding, error) {
	return m.findByUserID, m.findByUserIDErr
}

func (m *mockWeddingRepo) FindBySlug(_ context.Context, slug string) (*domain.Wedding, error) {
	if m.findBySlugErr != nil {
		return nil, m.findBySlugErr
	}
	if m.findBySlugResults != nil {
		w, ok := m.findBySlugResults[slug]
		if !ok || w == nil {
			return nil, domain.ErrNotFound
		}
		return w, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockWeddingRepo) Update(_ context.Context, w *domain.Wedding) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated = w
	return nil
}

func (m *mockWeddingRepo) AddPhoto(_ context.Context, p *domain.WeddingPhoto) error {
	if m.addPhotoErr != nil {
		return m.addPhotoErr
	}
	p.ID = "photo-id-1"
	m.addedPhoto = p
	return nil
}

func (m *mockWeddingRepo) FindPhotoByID(_ context.Context, _ string) (*domain.WeddingPhoto, error) {
	return m.findPhotoByID, m.findPhotoByIDErr
}

func (m *mockWeddingRepo) DeletePhoto(_ context.Context, _ string) (*domain.WeddingPhoto, error) {
	return m.deletePhotoResult, m.deletePhotoErr
}

func (m *mockWeddingRepo) SetCoverPhoto(_ context.Context, _, _ string) error {
	return m.setCoverPhotoErr
}

func (m *mockWeddingRepo) ReplaceLinks(_ context.Context, _ string, links []domain.WeddingLink) error {
	if m.replaceLinksErr != nil {
		return m.replaceLinksErr
	}
	m.replacedLinks = links
	return nil
}

// mockStorage is a test double for StorageService.
type mockStorage struct {
	uploadURL string
	uploadErr error
	deleteErr error
	deletedKey string
}

func (m *mockStorage) Upload(_ context.Context, _ string, _ io.Reader, _ string) (string, error) {
	return m.uploadURL, m.uploadErr
}

func (m *mockStorage) Delete(_ context.Context, key string) error {
	m.deletedKey = key
	return m.deleteErr
}
