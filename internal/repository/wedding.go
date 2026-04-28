package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ropehapi/wedding-mc/internal/domain"
)

type weddingRepo struct {
	db *sqlx.DB
}

func NewWeddingRepository(db *sqlx.DB) domain.WeddingRepository {
	return &weddingRepo{db: db}
}

func (r *weddingRepo) Create(ctx context.Context, w *domain.Wedding) error {
	query := `
		INSERT INTO weddings
			(user_id, slug, bride_name, groom_name, date, time, location, city, state, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		w.UserID, w.Slug, w.BrideName, w.GroomName,
		w.Date, w.Time, w.Location, w.City, w.State, w.Description,
	).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (r *weddingRepo) FindByUserID(ctx context.Context, userID string) (*domain.Wedding, error) {
	var w domain.Wedding
	err := r.db.GetContext(ctx, &w, `SELECT * FROM weddings WHERE user_id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadPhotosAndLinks(ctx, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *weddingRepo) FindBySlug(ctx context.Context, slug string) (*domain.Wedding, error) {
	var w domain.Wedding
	err := r.db.GetContext(ctx, &w, `SELECT * FROM weddings WHERE slug = $1`, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadPhotosAndLinks(ctx, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *weddingRepo) Update(ctx context.Context, w *domain.Wedding) error {
	query := `
		UPDATE weddings SET
			bride_name  = $1,
			groom_name  = $2,
			date        = $3,
			time        = $4,
			location    = $5,
			city        = $6,
			state       = $7,
			description = $8,
			updated_at  = NOW()
		WHERE id = $9
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, query,
		w.BrideName, w.GroomName, w.Date, w.Time,
		w.Location, w.City, w.State, w.Description, w.ID,
	).Scan(&w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (r *weddingRepo) AddPhoto(ctx context.Context, p *domain.WeddingPhoto) error {
	query := `
		INSERT INTO wedding_photos (wedding_id, url, storage_key)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, p.WeddingID, p.URL, p.StorageKey).
		Scan(&p.ID, &p.CreatedAt)
}

func (r *weddingRepo) FindPhotoByID(ctx context.Context, photoID string) (*domain.WeddingPhoto, error) {
	var p domain.WeddingPhoto
	err := r.db.GetContext(ctx, &p, `SELECT * FROM wedding_photos WHERE id = $1`, photoID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *weddingRepo) DeletePhoto(ctx context.Context, photoID string) (*domain.WeddingPhoto, error) {
	var p domain.WeddingPhoto
	err := r.db.QueryRowContext(
		ctx, `DELETE FROM wedding_photos WHERE id = $1 RETURNING id, wedding_id, url, storage_key, created_at`, photoID,
	).Scan(&p.ID, &p.WeddingID, &p.URL, &p.StorageKey, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *weddingRepo) SetCoverPhoto(ctx context.Context, photoID, weddingID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx,
		`UPDATE wedding_photos SET is_cover = FALSE WHERE wedding_id = $1`, weddingID,
	); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE wedding_photos SET is_cover = TRUE WHERE id = $1 AND wedding_id = $2`, photoID, weddingID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit()
}

func (r *weddingRepo) ReplaceLinks(ctx context.Context, weddingID string, links []domain.WeddingLink) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM wedding_links WHERE wedding_id = $1`, weddingID); err != nil {
		return fmt.Errorf("delete links: %w", err)
	}

	for i := range links {
		links[i].WeddingID = weddingID
		links[i].Position = i
		err := tx.QueryRowContext(ctx,
			`INSERT INTO wedding_links (wedding_id, label, url, position)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, created_at`,
			links[i].WeddingID, links[i].Label, links[i].URL, links[i].Position,
		).Scan(&links[i].ID, &links[i].CreatedAt)
		if err != nil {
			return fmt.Errorf("insert link %d: %w", i, err)
		}
	}

	return tx.Commit()
}

func (r *weddingRepo) loadPhotosAndLinks(ctx context.Context, w *domain.Wedding) error {
	photos := []domain.WeddingPhoto{}
	if err := r.db.SelectContext(ctx, &photos,
		`SELECT * FROM wedding_photos WHERE wedding_id = $1 ORDER BY created_at`, w.ID,
	); err != nil {
		return err
	}
	w.Photos = photos

	links := []domain.WeddingLink{}
	if err := r.db.SelectContext(ctx, &links,
		`SELECT * FROM wedding_links WHERE wedding_id = $1 ORDER BY position`, w.ID,
	); err != nil {
		return err
	}
	w.Links = links
	return nil
}
