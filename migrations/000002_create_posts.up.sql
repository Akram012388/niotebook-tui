CREATE TABLE posts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT posts_content_not_empty CHECK (char_length(TRIM(content)) > 0),
    CONSTRAINT posts_content_max_length CHECK (char_length(content) <= 140)
);

CREATE INDEX idx_posts_created_at ON posts (created_at DESC);
CREATE INDEX idx_posts_author_created ON posts (author_id, created_at DESC);
