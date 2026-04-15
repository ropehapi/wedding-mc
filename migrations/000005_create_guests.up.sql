CREATE TYPE rsvp_status AS ENUM ('pending', 'confirmed', 'declined');

CREATE TABLE guests (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id  UUID        NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    status      rsvp_status NOT NULL DEFAULT 'pending',
    rsvp_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guests_wedding_id ON guests(wedding_id);
CREATE INDEX idx_guests_status ON guests(wedding_id, status);
