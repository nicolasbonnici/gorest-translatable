package migrations

import (
	"context"

	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/migrations"
)

func GetMigrations() migrations.MigrationSource {
	builder := migrations.NewMigrationBuilder("gorest-translatable")

	builder.Add(
		"20260104000001000",
		"create_translations_table",
		func(ctx context.Context, db database.Database) error {
			if err := migrations.SQL(ctx, db, migrations.DialectSQL{
				Postgres: `CREATE TABLE IF NOT EXISTS translations (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					user_id UUID REFERENCES users(id) ON DELETE SET NULL,
					translatable_id UUID NOT NULL,
					translatable TEXT NOT NULL,
					locale TEXT NOT NULL DEFAULT 'en',
					content JSONB NOT NULL,
					updated_at TIMESTAMP(0) WITH TIME ZONE,
					created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
					UNIQUE(translatable_id, translatable, locale)
				)`,
				MySQL: `CREATE TABLE IF NOT EXISTS translations (
					id CHAR(36) PRIMARY KEY,
					user_id CHAR(36),
					translatable_id CHAR(36) NOT NULL,
					translatable VARCHAR(255) NOT NULL,
					locale VARCHAR(10) NOT NULL DEFAULT 'en',
					content JSON NOT NULL,
					updated_at TIMESTAMP NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
					UNIQUE KEY unique_translation (translatable_id, translatable, locale),
					INDEX idx_translatable_lookup (translatable_id, translatable, locale),
					INDEX idx_translatable_user (user_id),
					INDEX idx_translatable_created (created_at DESC)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
				SQLite: `CREATE TABLE IF NOT EXISTS translations (
					id TEXT PRIMARY KEY,
					user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
					translatable_id TEXT NOT NULL,
					translatable TEXT NOT NULL,
					locale TEXT NOT NULL DEFAULT 'en',
					content TEXT NOT NULL,
					updated_at TEXT,
					created_at TEXT NOT NULL DEFAULT (datetime('now')),
					UNIQUE(translatable_id, translatable, locale)
				)`,
			}); err != nil {
				return err
			}

			// Create indexes for Postgres and SQLite
			if db.DriverName() == "postgres" {
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					Postgres: `CREATE INDEX IF NOT EXISTS idx_translations_lookup ON translations(translatable_id, translatable, locale)`,
				}); err != nil {
					return err
				}
				if err := migrations.CreateIndex(ctx, db, "idx_translations_user", "translations", "user_id"); err != nil {
					return err
				}
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					Postgres: `CREATE INDEX IF NOT EXISTS idx_translations_created ON translations(created_at DESC)`,
				}); err != nil {
					return err
				}
			}

			if db.DriverName() == "sqlite" {
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					SQLite: `CREATE INDEX IF NOT EXISTS idx_translations_lookup ON translations(translatable_id, translatable, locale)`,
				}); err != nil {
					return err
				}
				if err := migrations.CreateIndex(ctx, db, "idx_translations_user", "translations", "user_id"); err != nil {
					return err
				}
				if err := migrations.SQL(ctx, db, migrations.DialectSQL{
					SQLite: `CREATE INDEX IF NOT EXISTS idx_translations_created ON translations(created_at DESC)`,
				}); err != nil {
					return err
				}
			}

			return nil
		},
		func(ctx context.Context, db database.Database) error {
			// Drop indexes first
			if db.DriverName() == "postgres" || db.DriverName() == "sqlite" {
				_ = migrations.DropIndex(ctx, db, "idx_translations_lookup", "translations")
				_ = migrations.DropIndex(ctx, db, "idx_translations_user", "translations")
				_ = migrations.DropIndex(ctx, db, "idx_translations_created", "translations")
			}

			return migrations.DropTableIfExists(ctx, db, "translations")
		},
	)

	return builder.Build()
}
