package translatable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslatableService_GetLocales(t *testing.T) {
	tests := []struct {
		name            string
		config          *Config
		expectedCount   int
		expectedDefault string
		expectedLocales []string
	}{
		{
			name: "multiple locales",
			config: &Config{
				SupportedLocales: []string{"en", "fr", "es"},
				DefaultLocale:    "en",
			},
			expectedCount:   3,
			expectedDefault: "en",
			expectedLocales: []string{"en", "fr", "es"},
		},
		{
			name: "single locale",
			config: &Config{
				SupportedLocales: []string{"en"},
				DefaultLocale:    "en",
			},
			expectedCount:   1,
			expectedDefault: "en",
			expectedLocales: []string{"en"},
		},
		{
			name: "non-english default",
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
			resp := service.GetLocales()

			assert.Equal(t, tt.expectedDefault, resp.Default)
			assert.Len(t, resp.Locales, tt.expectedCount)

			localeMap := make(map[string]LocaleInfo)
			for _, locale := range resp.Locales {
				localeMap[locale.Locale] = locale
			}

			for _, expectedLocale := range tt.expectedLocales {
				info, found := localeMap[expectedLocale]
				assert.True(t, found, "locale %s not found", expectedLocale)
				assert.Equal(t, expectedLocale == tt.expectedDefault, info.IsDefault)
			}

			defaultCount := 0
			for _, locale := range resp.Locales {
				if locale.IsDefault {
					defaultCount++
					assert.Equal(t, tt.expectedDefault, locale.Locale)
				}
			}
			assert.Equal(t, 1, defaultCount)
		})
	}
}

func TestTranslatableService_DefaultLocale(t *testing.T) {
	service := NewTranslatableService(nil, &Config{
		SupportedLocales: []string{"en", "fr"},
		DefaultLocale:    "fr",
	})
	assert.Equal(t, "fr", service.DefaultLocale())
}

func TestTranslatableService_TargetLocales(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name:     "excludes default",
			config:   &Config{SupportedLocales: []string{"en", "fr", "es"}, DefaultLocale: "en"},
			expected: []string{"fr", "es"},
		},
		{
			name:     "only default locale",
			config:   &Config{SupportedLocales: []string{"en"}, DefaultLocale: "en"},
			expected: []string{},
		},
		{
			name:     "non-english default",
			config:   &Config{SupportedLocales: []string{"en", "fr", "de"}, DefaultLocale: "fr"},
			expected: []string{"en", "de"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewTranslatableService(nil, tt.config)
			targets := service.TargetLocales()
			assert.ElementsMatch(t, tt.expected, targets)
		})
	}
}
