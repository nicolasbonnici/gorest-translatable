package translatable

import (
	"testing"
)

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin()

	if plugin == nil {
		t.Fatal("NewPlugin() returned nil")
	}

	if plugin.Name() != "translatable" {
		t.Errorf("Name() = %v, want 'translatable'", plugin.Name())
	}
}

func TestTranslatablePlugin_Initialize(t *testing.T) {
	plugin := &TranslatablePlugin{}

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "valid config",
			config: map[string]interface{}{
				"allowed_types":        []interface{}{"posts", "articles"},
				"supported_locales":    []interface{}{"en", "fr"},
				"default_locale":       "en",
				"pagination_limit":     20,
				"max_pagination_limit": 100,
				"max_content_length":   10240,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.Initialize(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTranslatablePlugin_MigrationDependencies(t *testing.T) {
	plugin := &TranslatablePlugin{}
	deps := plugin.MigrationDependencies()

	if len(deps) == 0 {
		t.Error("MigrationDependencies() should return at least one dependency")
	}

	found := false
	for _, dep := range deps {
		if dep == "auth" {
			found = true
			break
		}
	}

	if !found {
		t.Error("MigrationDependencies() should include 'auth'")
	}
}
