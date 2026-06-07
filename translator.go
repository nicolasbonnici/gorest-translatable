package translatable

import (
	"context"

	"github.com/google/uuid"
)

// TranslationResult holds the per-locale outcome of a translate request.
type TranslationResult struct {
	Translated []string `json:"translated"`
	Skipped    []string `json:"skipped"`
	Failed     []string `json:"failed"`
}

// Translator is satisfied by *ai.AutoTranslator (via an adapter in the host app).
type Translator interface {
	Translate(ctx context.Context, resourceType, resourceID string, userID *uuid.UUID) (*TranslationResult, error)
}
