package translatable

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/database/postgres"
	"github.com/nicolasbonnici/gorest/database/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eagerFakeRows scans a fixed set of Translatable records into the destinations produced by
// scanTranslatable, letting the eager-load grouping logic be exercised without a real driver.
type eagerFakeRows struct {
	data []Translatable
	pos  int
}

func (r *eagerFakeRows) Next() bool {
	if r.pos >= len(r.data) {
		return false
	}
	r.pos++
	return true
}

func (r *eagerFakeRows) Scan(dest ...any) error {
	row := r.data[r.pos-1]
	*(dest[0].(*uuid.UUID)) = row.ID
	*(dest[1].(**uuid.UUID)) = row.UserID
	*(dest[2].(*uuid.UUID)) = row.TranslatableID
	*(dest[3].(*string)) = row.Translatable
	*(dest[4].(*string)) = row.Locale
	*(dest[5].(*string)) = row.Content
	*(dest[6].(**time.Time)) = row.UpdatedAt
	*(dest[7].(*time.Time)) = row.CreatedAt
	return nil
}

func (r *eagerFakeRows) Close() error { return nil }
func (r *eagerFakeRows) Err() error   { return nil }

type eagerFakeDB struct {
	dialect  database.Dialect
	rows     []Translatable
	queryErr error

	lastQuery string
	lastArgs  []any
	queryN    int
}

func (d *eagerFakeDB) Query(ctx context.Context, q string, args ...any) (database.Rows, error) {
	d.queryN++
	d.lastQuery = q
	d.lastArgs = args
	if d.queryErr != nil {
		return nil, d.queryErr
	}
	return &eagerFakeRows{data: d.rows}, nil
}

func (d *eagerFakeDB) QueryRow(ctx context.Context, q string, args ...any) database.Row { return nil }
func (d *eagerFakeDB) Exec(ctx context.Context, q string, args ...any) (database.Result, error) {
	return nil, nil
}
func (d *eagerFakeDB) Connect(ctx context.Context, dsn string) error   { return nil }
func (d *eagerFakeDB) Close() error                                    { return nil }
func (d *eagerFakeDB) Ping(ctx context.Context) error                  { return nil }
func (d *eagerFakeDB) Begin(ctx context.Context) (database.Tx, error)  { return nil, nil }
func (d *eagerFakeDB) Dialect() database.Dialect                       { return d.dialect }
func (d *eagerFakeDB) DriverName() string                              { return "fake" }
func (d *eagerFakeDB) Introspector() database.SchemaIntrospector       { return nil }

func newTranslation(resType string, resID uuid.UUID, locale string) Translatable {
	return Translatable{
		ID:             uuid.New(),
		TranslatableID: resID,
		Translatable:   resType,
		Locale:         locale,
		Content:        locale + "-content",
		CreatedAt:      time.Now(),
	}
}

func TestGetTranslationsForResources_GroupsInSingleQuery(t *testing.T) {
	postID1 := uuid.New()
	postID2 := uuid.New()

	rows := []Translatable{
		newTranslation("post", postID1, "en"),
		newTranslation("post", postID1, "fr"),
		newTranslation("post", postID1, "es"),
		newTranslation("post", postID1, "de"),
		newTranslation("post", postID2, "en"),
	}

	db := &eagerFakeDB{dialect: &sqlite.SQLiteDialect{}, rows: rows}
	svc := NewTranslatableService(db, &Config{SupportedLocales: []string{"en", "fr", "es", "de"}, DefaultLocale: "en"})

	grouped, err := svc.GetTranslationsForResources(context.Background(), "post", []uuid.UUID{postID1, postID2})
	require.NoError(t, err)

	assert.Equal(t, 1, db.queryN, "eager-load must issue exactly one query")
	assert.Len(t, grouped[postID1], 4)
	assert.Len(t, grouped[postID2], 1)

	locales := make([]string, 0, 4)
	for _, tr := range grouped[postID1] {
		locales = append(locales, tr.Locale)
	}
	assert.Equal(t, []string{"en", "fr", "es", "de"}, locales, "per-resource locales preserved in row order")

	assert.Contains(t, db.lastQuery, "IN", "resource ids filtered with a single IN clause")
	assert.Equal(t, []any{"post", postID1, postID2}, db.lastArgs)
}

func TestGetTranslationsForResources_DeduplicatesIDs(t *testing.T) {
	postID := uuid.New()
	db := &eagerFakeDB{dialect: &postgres.PostgresDialect{}, rows: nil}
	svc := NewTranslatableService(db, &Config{SupportedLocales: []string{"en"}, DefaultLocale: "en"})

	_, err := svc.GetTranslationsForResources(context.Background(), "post", []uuid.UUID{postID, postID, postID})
	require.NoError(t, err)

	assert.Equal(t, []any{"post", postID}, db.lastArgs, "duplicate ids collapse to a single placeholder")
	assert.True(t, strings.Contains(db.lastQuery, "$1"), "postgres dialect uses numbered placeholders")
}

func TestGetTranslationsForResources_EmptyIDsSkipsQuery(t *testing.T) {
	db := &eagerFakeDB{dialect: &sqlite.SQLiteDialect{}}
	svc := NewTranslatableService(db, &Config{SupportedLocales: []string{"en"}, DefaultLocale: "en"})

	grouped, err := svc.GetTranslationsForResources(context.Background(), "post", nil)
	require.NoError(t, err)
	assert.Empty(t, grouped)
	assert.Equal(t, 0, db.queryN, "no ids means no round-trip")
}

func TestGetTranslations_SingleResource(t *testing.T) {
	postID := uuid.New()
	rows := []Translatable{
		newTranslation("post", postID, "en"),
		newTranslation("post", postID, "fr"),
	}
	db := &eagerFakeDB{dialect: &sqlite.SQLiteDialect{}, rows: rows}
	svc := NewTranslatableService(db, &Config{SupportedLocales: []string{"en", "fr"}, DefaultLocale: "en"})

	got, err := svc.GetTranslations(context.Background(), "post", postID)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, 1, db.queryN)
}

func TestGetTranslationsForResources_PropagatesQueryError(t *testing.T) {
	postID := uuid.New()
	db := &eagerFakeDB{dialect: &sqlite.SQLiteDialect{}, queryErr: errors.New("boom")}
	svc := NewTranslatableService(db, &Config{SupportedLocales: []string{"en"}, DefaultLocale: "en"})

	_, err := svc.GetTranslationsForResources(context.Background(), "post", []uuid.UUID{postID})
	require.Error(t, err)
}
