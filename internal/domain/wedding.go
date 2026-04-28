package domain

import (
	"context"
	"time"
)

type Wedding struct {
	ID          string     `db:"id" json:"id"`
	UserID      string     `db:"user_id" json:"user_id"`
	Slug        string     `db:"slug" json:"slug"`
	BrideName   string     `db:"bride_name" json:"bride_name"`
	GroomName   string     `db:"groom_name" json:"groom_name"`
	Date        time.Time  `db:"date" json:"date"`
	Time        *string    `db:"time" json:"time,omitempty"`
	Location    string     `db:"location" json:"location"`
	City        *string    `db:"city" json:"city,omitempty"`
	State       *string    `db:"state" json:"state,omitempty"`
	Description *string    `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`

	Photos []WeddingPhoto `db:"-" json:"photos,omitempty"`
	Links  []WeddingLink  `db:"-" json:"links,omitempty"`
}

type WeddingPhoto struct {
	ID         string    `db:"id" json:"id"`
	WeddingID  string    `db:"wedding_id" json:"wedding_id"`
	URL        string    `db:"url" json:"url"`
	StorageKey string    `db:"storage_key" json:"-"`
	IsCover    bool      `db:"is_cover" json:"is_cover"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type WeddingLink struct {
	ID        string    `db:"id" json:"id"`
	WeddingID string    `db:"wedding_id" json:"wedding_id"`
	Label     string    `db:"label" json:"label"`
	URL       string    `db:"url" json:"url"`
	Position  int       `db:"position" json:"position"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type WeddingRepository interface {
	Create(ctx context.Context, w *Wedding) error
	FindByUserID(ctx context.Context, userID string) (*Wedding, error)
	FindBySlug(ctx context.Context, slug string) (*Wedding, error)
	Update(ctx context.Context, w *Wedding) error
	AddPhoto(ctx context.Context, p *WeddingPhoto) error
	DeletePhoto(ctx context.Context, photoID string) (*WeddingPhoto, error)
	FindPhotoByID(ctx context.Context, photoID string) (*WeddingPhoto, error)
	SetCoverPhoto(ctx context.Context, photoID, weddingID string) error
	ReplaceLinks(ctx context.Context, weddingID string, links []WeddingLink) error
}
