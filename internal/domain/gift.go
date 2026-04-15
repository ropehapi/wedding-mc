package domain

import (
	"context"
	"time"
)

type GiftStatus string

const (
	GiftAvailable GiftStatus = "available"
	GiftReserved  GiftStatus = "reserved"
)

type Gift struct {
	ID             string     `db:"id" json:"id"`
	WeddingID      string     `db:"wedding_id" json:"wedding_id"`
	Name           string     `db:"name" json:"name"`
	Description    *string    `db:"description" json:"description,omitempty"`
	ImageURL       *string    `db:"image_url" json:"image_url,omitempty"`
	StoreURL       *string    `db:"store_url" json:"store_url,omitempty"`
	Price          *float64   `db:"price" json:"price,omitempty"`
	Status         GiftStatus `db:"status" json:"status"`
	ReservedByName *string    `db:"reserved_by_name" json:"reserved_by_name,omitempty"`
	ReservedAt     *time.Time `db:"reserved_at" json:"reserved_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

type GiftRepository interface {
	Create(ctx context.Context, g *Gift) error
	FindAll(ctx context.Context, weddingID string, status *GiftStatus) ([]Gift, error)
	FindByID(ctx context.Context, id string) (*Gift, error)
	Update(ctx context.Context, g *Gift) error
	Delete(ctx context.Context, id string) error
	Reserve(ctx context.Context, giftID, guestName string) error
	CancelReserve(ctx context.Context, giftID string) error
	CountByStatus(ctx context.Context, weddingID string) (map[GiftStatus]int, error)
}
