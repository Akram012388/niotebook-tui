CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username     VARCHAR(15) NOT NULL,
    email        VARCHAR(255) NOT NULL,
    password     VARCHAR(255) NOT NULL,
    display_name VARCHAR(50) NOT NULL DEFAULT '',
    bio          TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT users_username_format CHECK (username ~ '^[a-z0-9]([a-z0-9_]*[a-z0-9])?$'),
    CONSTRAINT users_username_no_consecutive_underscores CHECK (username NOT LIKE '%__%'),
    CONSTRAINT users_email_format CHECK (email ~ '^[^@]+@[^@]+\.[^@]+$'),
    CONSTRAINT users_bio_max_length CHECK (char_length(bio) <= 160),
    CONSTRAINT users_display_name_max_length CHECK (char_length(display_name) <= 50)
);

CREATE UNIQUE INDEX idx_users_username_lower ON users (LOWER(username));
CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));
