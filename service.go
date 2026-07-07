package translatable

import (
	"context"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/query"
)

// translationColumns is the canonical SELECT column order shared by every read path so
// that a single scan helper can hydrate a Translatable regardless of the caller.
var translationColumns = []string{
	"id", "user_id", "translatable_id", "translatable", "locale", "content", "updated_at", "created_at",
}

// rowScanner is satisfied by both database.Row (single row) and database.Rows (result set),
// letting one scan routine serve QueryRow and Query callers alike.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanTranslatable(s rowScanner) (Translatable, error) {
	var t Translatable
	err := s.Scan(
		&t.ID,
		&t.UserID,
		&t.TranslatableID,
		&t.Translatable,
		&t.Locale,
		&t.Content,
		&t.UpdatedAt,
		&t.CreatedAt,
	)
	return t, err
}

type TranslatableService struct {
	db     database.Database
	config *Config
}

func NewTranslatableService(db database.Database, config *Config) *TranslatableService {
	return &TranslatableService{
		db:     db,
		config: config,
	}
}

func (s *TranslatableService) GetLocales() LocalesResponse {
	locales := make([]LocaleInfo, 0, len(s.config.SupportedLocales))
	for _, locale := range s.config.SupportedLocales {
		locales = append(locales, LocaleInfo{
			Locale:    locale,
			IsDefault: locale == s.config.DefaultLocale,
		})
	}
	return LocalesResponse{Default: s.config.DefaultLocale, Locales: locales}
}

func (s *TranslatableService) DefaultLocale() string {
	return s.config.DefaultLocale
}

func (s *TranslatableService) TargetLocales() []string {
	targets := make([]string, 0, len(s.config.SupportedLocales))
	for _, locale := range s.config.SupportedLocales {
		if locale != s.config.DefaultLocale {
			targets = append(targets, locale)
		}
	}
	return targets
}

// GetTranslations eager-loads every locale of a single resource in one query.
func (s *TranslatableService) GetTranslations(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]Translatable, error) {
	byResource, err := s.GetTranslationsForResources(ctx, resourceType, []uuid.UUID{resourceID})
	if err != nil {
		return nil, err
	}
	return byResource[resourceID], nil
}

// GetTranslationsForResources eager-loads every locale for a set of same-typed resources in
// a single query and returns them grouped by resource id. Hydrating a page of translated
// resources previously issued one query per resource (and sometimes per locale), an N+1 that
// this collapses into a single `translatable = ? AND translatable_id IN (...)` round-trip.
// Rows are ordered by (translatable_id, locale) so the per-resource slices are deterministic.
func (s *TranslatableService) GetTranslationsForResources(ctx context.Context, resourceType string, resourceIDs []uuid.UUID) (map[uuid.UUID][]Translatable, error) {
	grouped := make(map[uuid.UUID][]Translatable, len(resourceIDs))
	if len(resourceIDs) == 0 {
		return grouped, nil
	}

	ids := make([]any, 0, len(resourceIDs))
	seen := make(map[uuid.UUID]struct{}, len(resourceIDs))
	for _, id := range resourceIDs {
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	sql, args, err := query.New(s.db.Dialect()).
		Select(translationColumns...).
		From((Translatable{}).TableName()).
		Where(query.Eq("translatable", resourceType)).
		Where(query.In("translatable_id", ids...)).
		OrderBy("translatable_id", query.ASC).
		OrderBy("locale", query.ASC).
		Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		t, scanErr := scanTranslatable(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		grouped[t.TranslatableID] = append(grouped[t.TranslatableID], t)
	}

	return grouped, rows.Err()
}
