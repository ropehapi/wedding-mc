package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ropehapi/wedding-mc/internal/domain"
)

type giftRepo struct {
	db *sqlx.DB
}

func NewGiftRepository(db *sqlx.DB) domain.GiftRepository {
	return &giftRepo{db: db}
}

func (r *giftRepo) Create(ctx context.Context, g *domain.Gift) error {
	query := `
		INSERT INTO gifts (wedding_id, name, description, image_url, store_url, price)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		g.WeddingID, g.Name, g.Description, g.ImageURL, g.StoreURL, g.Price,
	).Scan(&g.ID, &g.Status, &g.CreatedAt, &g.UpdatedAt)
}

func (r *giftRepo) FindAll(ctx context.Context, weddingID string, status *domain.GiftStatus) ([]domain.Gift, error) {
	gifts := []domain.Gift{}
	if status != nil {
		err := r.db.SelectContext(ctx, &gifts,
			`SELECT * FROM gifts WHERE wedding_id = $1 AND status = $2 ORDER BY created_at`,
			weddingID, *status,
		)
		return gifts, err
	}
	err := r.db.SelectContext(ctx, &gifts,
		`SELECT * FROM gifts WHERE wedding_id = $1 ORDER BY created_at`,
		weddingID,
	)
	return gifts, err
}

func (r *giftRepo) FindByID(ctx context.Context, id string) (*domain.Gift, error) {
	var g domain.Gift
	err := r.db.GetContext(ctx, &g, `SELECT * FROM gifts WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *giftRepo) Update(ctx context.Context, g *domain.Gift) error {
	query := `
		UPDATE gifts SET
			name        = $1,
			description = $2,
			image_url   = $3,
			store_url   = $4,
			price       = $5,
			updated_at  = NOW()
		WHERE id = $6
		RETURNING updated_at`
	err := r.db.QueryRowContext(ctx, query,
		g.Name, g.Description, g.ImageURL, g.StoreURL, g.Price, g.ID,
	).Scan(&g.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (r *giftRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM gifts WHERE id = $1`, id)
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
	return nil
}

// Reserve atomically reserves a gift using SELECT FOR UPDATE to prevent double-booking.
func (r *giftRepo) Reserve(ctx context.Context, giftID, guestName string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var id string
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM gifts WHERE id = $1 AND status = 'available' FOR UPDATE`,
		giftID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrConflict
	}
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE gifts SET status = 'reserved', reserved_by_name = $1, reserved_at = NOW(), updated_at = NOW() WHERE id = $2`,
		guestName, giftID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *giftRepo) CancelReserve(ctx context.Context, giftID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE gifts SET status = 'available', reserved_by_name = NULL, reserved_at = NULL, updated_at = NOW() WHERE id = $1`,
		giftID,
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
	return nil
}

func (r *giftRepo) CountByStatus(ctx context.Context, weddingID string) (map[domain.GiftStatus]int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM gifts WHERE wedding_id = $1 GROUP BY status`,
		weddingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[domain.GiftStatus]int{
		domain.GiftAvailable: 0,
		domain.GiftReserved:  0,
	}
	for rows.Next() {
		var status domain.GiftStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}
