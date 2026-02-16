package translatable

import (
	"errors"
	"fmt"

	"github.com/nicolasbonnici/gorest/database"
)

type Config struct {
	Database           database.Database
	AllowedTypes       []string `json:"allowed_types" yaml:"allowed_types"`
	SupportedLocales   []string `json:"supported_locales" yaml:"supported_locales"`
	DefaultLocale      string   `json:"default_locale" yaml:"default_locale"`
	PaginationLimit    int      `json:"pagination_limit" yaml:"pagination_limit"`
	MaxPaginationLimit int      `json:"max_pagination_limit" yaml:"max_pagination_limit"`
	MaxContentLength   int      `json:"max_content_length" yaml:"max_content_length"`
}

func (c *Config) Validate() error {
	if err := c.validateAllowedTypes(); err != nil {
		return err
	}

	if err := c.validateSupportedLocales(); err != nil {
		return err
	}

	if err := c.validateDefaultLocale(); err != nil {
		return err
	}

	c.applyDefaults()

	if c.MaxContentLength < 1 || c.MaxContentLength > 1048576 {
		return errors.New("max_content_length must be between 1 and 1048576 bytes")
	}

	return nil
}

func (c *Config) validateAllowedTypes() error {
	if len(c.AllowedTypes) == 0 {
		return errors.New("allowed_types cannot be empty")
	}

	seen := make(map[string]bool)
	for _, t := range c.AllowedTypes {
		if t == "" {
			return errors.New("allowed_types cannot contain empty strings")
		}
		if seen[t] {
			return fmt.Errorf("duplicate type in allowed_types: %s", t)
		}
		seen[t] = true
	}

	return nil
}

func (c *Config) validateSupportedLocales() error {
	if len(c.SupportedLocales) == 0 {
		return errors.New("supported_locales cannot be empty")
	}

	seen := make(map[string]bool)
	for _, locale := range c.SupportedLocales {
		if locale == "" {
			return errors.New("supported_locales cannot contain empty strings")
		}
		if seen[locale] {
			return fmt.Errorf("duplicate locale in supported_locales: %s", locale)
		}
		seen[locale] = true
	}

	return nil
}

func (c *Config) validateDefaultLocale() error {
	if c.DefaultLocale == "" {
		return errors.New("default_locale cannot be empty")
	}

	for _, locale := range c.SupportedLocales {
		if locale == c.DefaultLocale {
			return nil
		}
	}

	return errors.New("default_locale must be one of the supported_locales")
}

func (c *Config) applyDefaults() {
	if c.PaginationLimit <= 0 {
		c.PaginationLimit = 20
	}

	if c.MaxPaginationLimit <= 0 {
		c.MaxPaginationLimit = 100
	}

	if c.MaxContentLength <= 0 {
		c.MaxContentLength = 10240
	}
}

func (c *Config) IsAllowedType(typeName string) bool {
	for _, allowed := range c.AllowedTypes {
		if allowed == typeName {
			return true
		}
	}
	return false
}

func (c *Config) IsSupportedLocale(locale string) bool {
	for _, supported := range c.SupportedLocales {
		if supported == locale {
			return true
		}
	}
	return false
}

func DefaultConfig() Config {
	return Config{
		AllowedTypes:       []string{"post"},
		SupportedLocales:   []string{"en", "fr", "es"},
		DefaultLocale:      "en",
		PaginationLimit:    20,
		MaxPaginationLimit: 100,
		MaxContentLength:   10240,
	}
}
