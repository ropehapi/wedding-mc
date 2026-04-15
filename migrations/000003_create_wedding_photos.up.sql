CREATE TABLE wedding_photos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id  UUID          NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    url         VARCHAR(1000) NOT NULL,
    storage_key VARCHAR(1000) NOT NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wedding_photos_wedding_id ON wedding_photos(wedding_id);
