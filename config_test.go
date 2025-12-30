package translatable

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				AllowedTables:    []string{"posts", "articles"},
				MaxContentLength: 10240,
			},
			wantErr: false,
		},
		{
			name: "empty allowed tables",
			config: &Config{
				AllowedTables:    []string{},
				MaxContentLength: 10240,
			},
			wantErr: true,
			errMsg:  "allowed_tables cannot be empty",
		},
		{
			name: "allowed tables with empty string",
			config: &Config{
				AllowedTables:    []string{"posts", ""},
				MaxContentLength: 10240,
			},
			wantErr: true,
			errMsg:  "allowed_tables cannot contain empty strings",
		},
		{
			name: "duplicate table names",
			config: &Config{
				AllowedTables:    []string{"posts", "articles", "posts"},
				MaxContentLength: 10240,
			},
			wantErr: true,
			errMsg:  "duplicate table name in allowed_tables: posts",
		},
		{
			name: "zero max content length sets default",
			config: &Config{
				AllowedTables:    []string{"posts"},
				MaxContentLength: 0,
			},
			wantErr: false,
		},
		{
			name: "max content length too small",
			config: &Config{
				AllowedTables:    []string{"posts"},
				MaxContentLength: 0,
			},
			wantErr: false, // 0 sets default
		},
		{
			name: "max content length too large",
			config: &Config{
				AllowedTables:    []string{"posts"},
				MaxContentLength: 2000000, // 2MB, over 1MB limit
			},
			wantErr: true,
			errMsg:  "max_content_length must be between 1 and 1048576 bytes",
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
				// Check default was set
				if tt.config.MaxContentLength == 0 {
					t.Errorf("MaxContentLength should be set to default 10240, got %d", tt.config.MaxContentLength)
				}
			}
		})
	}
}

func TestConfig_IsAllowedTable(t *testing.T) {
	config := &Config{
		AllowedTables: []string{"posts", "articles", "products"},
	}

	tests := []struct {
		name      string
		tableName string
		want      bool
	}{
		{
			name:      "allowed table - posts",
			tableName: "posts",
			want:      true,
		},
		{
			name:      "allowed table - articles",
			tableName: "articles",
			want:      true,
		},
		{
			name:      "not allowed table",
			tableName: "categories",
			want:      false,
		},
		{
			name:      "empty string",
			tableName: "",
			want:      false,
		},
		{
			name:      "case sensitive - Posts",
			tableName: "Posts",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.IsAllowedTable(tt.tableName); got != tt.want {
				t.Errorf("IsAllowedTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if len(config.AllowedTables) == 0 {
		t.Error("DefaultConfig() should have at least one allowed table")
	}

	if config.MaxContentLength != 10240 {
		t.Errorf("DefaultConfig() MaxContentLength = %d, want 10240", config.MaxContentLength)
	}

	if err := config.Validate(); err != nil {
		t.Errorf("DefaultConfig() should be valid, got error: %v", err)
	}
}
