CREATE TABLE weddings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    slug        VARCHAR(255)  NOT NULL UNIQUE,
    bride_name  VARCHAR(255)  NOT NULL,
    groom_name  VARCHAR(255)  NOT NULL,
    date        DATE          NOT NULL,
    time        TIME,
    location    VARCHAR(500)  NOT NULL,
    city        VARCHAR(255),
    state       VARCHAR(2),
    description TEXT,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_weddings_slug ON weddings(slug);
CREATE INDEX idx_weddings_user_id ON weddings(user_id);
