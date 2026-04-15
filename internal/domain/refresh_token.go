package domain

import (
	"context"
	"time"
)

type RefreshToken struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	TokenHash string    `db:"token_hash" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	Revoked   bool      `db:"revoked" json:"revoked"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID string) error
}
