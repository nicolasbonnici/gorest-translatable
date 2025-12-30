CREATE TABLE translatable (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    translatable_id UUID NOT NULL,
    translatable TEXT NOT NULL,
    content TEXT NOT NULL,
    updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_translatable_lookup ON translatable(translatable_id, translatable);
CREATE INDEX idx_translatable_user ON translatable(user_id);
CREATE INDEX idx_translatable_created ON translatable(created_at DESC);

COMMENT ON TABLE translatable IS 'Stores translatable content for various resource types (posts, articles, products, etc.)';
COMMENT ON COLUMN translatable.translatable_id IS 'UUID of the parent resource';
COMMENT ON COLUMN translatable.translatable IS 'Type of parent resource (e.g., posts, articles, products)';
COMMENT ON COLUMN translatable.content IS 'The translatable content';
