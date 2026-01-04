CREATE TABLE translatable (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    translatable_id UUID NOT NULL,
    translatable TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT 'en',
    content TEXT NOT NULL,
    updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(translatable_id, translatable, locale)
);

CREATE INDEX idx_translatable_lookup ON translatable(translatable_id, translatable, locale);
CREATE INDEX idx_translatable_user ON translatable(user_id);
CREATE INDEX idx_translatable_created ON translatable(created_at DESC);

COMMENT ON TABLE translatable IS 'Stores multi-language translations for various resource types (posts, articles, products, etc.)';
COMMENT ON COLUMN translatable.translatable_id IS 'UUID of the parent resource';
COMMENT ON COLUMN translatable.translatable IS 'Type of parent resource (e.g., post, article, product)';
COMMENT ON COLUMN translatable.locale IS 'Language code (e.g., en, fr, es)';
COMMENT ON COLUMN translatable.content IS 'The translated content';
