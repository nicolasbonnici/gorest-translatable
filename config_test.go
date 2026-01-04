package translatable

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				AllowedTypes:       []string{"posts", "articles"},
				SupportedLocales:   []string{"en", "fr"},
				DefaultLocale:      "en",
				MaxContentLength:   10240,
				PaginationLimit:    20,
				MaxPaginationLimit: 100,
			},
			wantErr: false,
		},
		{
			name: "empty allowed types",
			config: Config{
				AllowedTypes:     []string{},
				SupportedLocales: []string{"en"},
				DefaultLocale:    "en",
			},
			wantErr: true,
			errMsg:  "allowed_types cannot be empty",
		},
		{
			name: "allowed types with empty string",
			config: Config{
				AllowedTypes:     []string{"posts", ""},
				SupportedLocales: []string{"en"},
				DefaultLocale:    "en",
			},
			wantErr: true,
			errMsg:  "allowed_types cannot contain empty strings",
		},
		{
			name: "duplicate type names",
			config: Config{
				AllowedTypes:     []string{"posts", "articles", "posts"},
				SupportedLocales: []string{"en"},
				DefaultLocale:    "en",
			},
			wantErr: true,
			errMsg:  "duplicate type in allowed_types: posts",
		},
		{
			name: "empty supported locales",
			config: Config{
				AllowedTypes:     []string{"posts"},
				SupportedLocales: []string{},
				DefaultLocale:    "en",
			},
			wantErr: true,
			errMsg:  "supported_locales cannot be empty",
		},
		{
			name: "empty default locale",
			config: Config{
				AllowedTypes:     []string{"posts"},
				SupportedLocales: []string{"en"},
				DefaultLocale:    "",
			},
			wantErr: true,
			errMsg:  "default_locale cannot be empty",
		},
		{
			name: "default locale not in supported",
			config: Config{
				AllowedTypes:     []string{"posts"},
				SupportedLocales: []string{"en", "fr"},
				DefaultLocale:    "es",
			},
			wantErr: true,
			errMsg:  "default_locale must be one of the supported_locales",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfig_IsAllowedType(t *testing.T) {
	config := Config{
		AllowedTypes: []string{"posts", "articles", "products"},
	}

	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{
			name:     "allowed type - posts",
			typeName: "posts",
			want:     true,
		},
		{
			name:     "allowed type - articles",
			typeName: "articles",
			want:     true,
		},
		{
			name:     "not allowed type",
			typeName: "categories",
			want:     false,
		},
		{
			name:     "empty string",
			typeName: "",
			want:     false,
		},
		{
			name:     "case sensitive - Posts",
			typeName: "Posts",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.IsAllowedType(tt.typeName); got != tt.want {
				t.Errorf("IsAllowedType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IsSupportedLocale(t *testing.T) {
	config := Config{
		SupportedLocales: []string{"en", "fr", "es"},
	}

	tests := []struct {
		name   string
		locale string
		want   bool
	}{
		{
			name:   "supported locale - en",
			locale: "en",
			want:   true,
		},
		{
			name:   "supported locale - fr",
			locale: "fr",
			want:   true,
		},
		{
			name:   "not supported locale",
			locale: "de",
			want:   false,
		},
		{
			name:   "empty string",
			locale: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.IsSupportedLocale(tt.locale); got != tt.want {
				t.Errorf("IsSupportedLocale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if len(config.AllowedTypes) == 0 {
		t.Error("DefaultConfig() should have at least one allowed type")
	}

	if len(config.SupportedLocales) == 0 {
		t.Error("DefaultConfig() should have at least one supported locale")
	}

	if config.DefaultLocale == "" {
		t.Error("DefaultConfig() should have a default locale")
	}

	if config.PaginationLimit != 20 {
		t.Errorf("DefaultConfig() PaginationLimit = %d, want 20", config.PaginationLimit)
	}

	if config.MaxPaginationLimit != 100 {
		t.Errorf("DefaultConfig() MaxPaginationLimit = %d, want 100", config.MaxPaginationLimit)
	}

	if config.MaxContentLength != 10240 {
		t.Errorf("DefaultConfig() MaxContentLength = %d, want 10240", config.MaxContentLength)
	}

	if err := config.Validate(); err != nil {
		t.Errorf("DefaultConfig() should be valid, got error: %v", err)
	}
}
