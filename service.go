package translatable

import (
	"github.com/nicolasbonnici/gorest/database"
)

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

func (s *TranslatableService) GetLocales() []LocaleInfo {
	locales := make([]LocaleInfo, 0, len(s.config.SupportedLocales))

	for _, locale := range s.config.SupportedLocales {
		locales = append(locales, LocaleInfo{
			Locale:    locale,
			IsDefault: locale == s.config.DefaultLocale,
		})
	}

	return locales
}
