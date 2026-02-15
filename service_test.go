package translatable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslatableService_GetLocales(t *testing.T) {
	tests := []struct {
		name             string
		config           *Config
		expectedCount    int
		expectedDefault  string
		expectedLocales  []string
	}{
		{
			name: "success - multiple locales",
			config: &Config{
				SupportedLocales: []string{"en", "fr", "es"},
				DefaultLocale:    "en",
			},
			expectedCount:   3,
			expectedDefault: "en",
			expectedLocales: []string{"en", "fr", "es"},
		},
		{
			name: "success - single locale",
			config: &Config{
				SupportedLocales: []string{"en"},
				DefaultLocale:    "en",
			},
			expectedCount:   1,
			expectedDefault: "en",
			expectedLocales: []string{"en"},
		},
		{
			name: "success - different default locale",
			config: &Config{
				SupportedLocales: []string{"en", "fr", "de"},
				DefaultLocale:    "fr",
			},
			expectedCount:   3,
			expectedDefault: "fr",
			expectedLocales: []string{"en", "fr", "de"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewTranslatableService(nil, tt.config)
			locales := service.GetLocales()

			assert.Len(t, locales, tt.expectedCount)

			// Verify all locales are present
			localeMap := make(map[string]LocaleInfo)
			for _, locale := range locales {
				localeMap[locale.Locale] = locale
			}

			for _, expectedLocale := range tt.expectedLocales {
				info, found := localeMap[expectedLocale]
				assert.True(t, found, "Expected locale %s not found", expectedLocale)
				assert.Equal(t, expectedLocale, info.Locale)
				assert.Equal(t, expectedLocale == tt.expectedDefault, info.IsDefault)
			}

			// Verify only one default locale
			defaultCount := 0
			for _, locale := range locales {
				if locale.IsDefault {
					defaultCount++
					assert.Equal(t, tt.expectedDefault, locale.Locale)
				}
			}
			assert.Equal(t, 1, defaultCount, "There should be exactly one default locale")
		})
	}
}
