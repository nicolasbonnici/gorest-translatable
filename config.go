package translatable

import (
	"errors"
	"fmt"
)

type Config struct {
	AllowedTables    []string `json:"allowed_tables" yaml:"allowed_tables"`
	MaxContentLength int      `json:"max_content_length" yaml:"max_content_length"` // Default: 10KB
}

func (c *Config) Validate() error {
	if len(c.AllowedTables) == 0 {
		return errors.New("allowed_tables cannot be empty")
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, table := range c.AllowedTables {
		if table == "" {
			return errors.New("allowed_tables cannot contain empty strings")
		}
		if seen[table] {
			return fmt.Errorf("duplicate table name in allowed_tables: %s", table)
		}
		seen[table] = true
	}

	if c.MaxContentLength == 0 {
		c.MaxContentLength = 10240
	}

	if c.MaxContentLength < 1 || c.MaxContentLength > 1048576 {
		return errors.New("max_content_length must be between 1 and 1048576 bytes")
	}

	return nil
}

func (c *Config) IsAllowedTable(tableName string) bool {
	for _, allowed := range c.AllowedTables {
		if allowed == tableName {
			return true
		}
	}
	return false
}

func DefaultConfig() *Config {
	return &Config{
		AllowedTables:    []string{"posts", "articles"},
		MaxContentLength: 10240,
	}
}
