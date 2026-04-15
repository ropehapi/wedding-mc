package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/ropehapi/wedding-mc/internal/domain"
)

type refreshTokenRepo struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, rt *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query, rt.UserID, rt.TokenHash, rt.ExpiresAt).
		Scan(&rt.ID, &rt.CreatedAt)
}

func (r *refreshTokenRepo) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.GetContext(ctx, &rt, `SELECT * FROM refresh_tokens WHERE token_hash = $1`, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepo) RevokeByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`, userID)
	return err
}
