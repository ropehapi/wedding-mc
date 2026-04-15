CREATE TABLE wedding_links (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wedding_id  UUID          NOT NULL REFERENCES weddings(id) ON DELETE CASCADE,
    label       VARCHAR(255)  NOT NULL,
    url         VARCHAR(1000) NOT NULL,
    position    INTEGER       NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wedding_links_wedding_id ON wedding_links(wedding_id);
