package migrations

import _ "embed"

//go:embed 001_create_translatable.sql
var migration001 string

func GetMigrations() map[string]string {
	return map[string]string{
		"001_create_translatable.sql": migration001,
	}
}
