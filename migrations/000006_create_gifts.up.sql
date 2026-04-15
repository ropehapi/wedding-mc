CREATE TYPE gift_status AS ENUM ('available', 'reserved');

CREATE TABLE gifts (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id       UUID         NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    description      TEXT,
    image_url        VARCHAR(1000),
    store_url        VARCHAR(1000),
    price            NUMERIC(10,2),
    status           gift_status  NOT NULL DEFAULT 'available',
    reserved_by_name VARCHAR(255),
    reserved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gifts_wedding_id ON gifts(wedding_id);
CREATE INDEX idx_gifts_status ON gifts(wedding_id, status);
