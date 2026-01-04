package translatable

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCreateTranslatableRequest_Validate(t *testing.T) {
	config := &Config{
		AllowedTypes:     []string{"posts", "articles"},
		SupportedLocales: []string{"en", "fr"},
		DefaultLocale:    "en",
		MaxContentLength: 100,
	}

	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		req     *CreateTranslatableRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &CreateTranslatableRequest{
				TranslatableID: validUUID,
				Translatable:   "posts",
				Locale:         "en",
				Content:        "Valid content",
			},
			wantErr: false,
		},
		{
			name: "invalid UUID",
			req: &CreateTranslatableRequest{
				TranslatableID: "not-a-uuid",
				Translatable:   "posts",
				Locale:         "en",
				Content:        "Content",
			},
			wantErr: true,
			errMsg:  "translatable_id must be a valid UUID",
		},
		{
			name: "translatable not in allowed list",
			req: &CreateTranslatableRequest{
				TranslatableID: validUUID,
				Translatable:   "categories",
				Locale:         "en",
				Content:        "Content",
			},
			wantErr: true,
			errMsg:  "translatable type is not allowed",
		},
		{
			name: "locale not supported",
			req: &CreateTranslatableRequest{
				TranslatableID: validUUID,
				Translatable:   "posts",
				Locale:         "de",
				Content:        "Content",
			},
			wantErr: true,
			errMsg:  "locale is not supported",
		},
		{
			name: "empty content",
			req: &CreateTranslatableRequest{
				TranslatableID: validUUID,
				Translatable:   "posts",
				Locale:         "en",
				Content:        "   ",
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "content exceeds max length",
			req: &CreateTranslatableRequest{
				TranslatableID: validUUID,
				Translatable:   "posts",
				Locale:         "en",
				Content:        strings.Repeat("a", 101),
			},
			wantErr: true,
			errMsg:  "content exceeds maximum length",
		},
		{
			name: "XSS content is escaped",
			req: &CreateTranslatableRequest{
				TranslatableID: validUUID,
				Translatable:   "posts",
				Locale:         "en",
				Content:        "<script>alert('xss')</script>",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(config)
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
				if strings.Contains(tt.req.Content, "<script>") {
					t.Error("Content should be HTML escaped, but contains <script>")
				}
			}
		})
	}
}

func TestUpdateTranslatableRequest_Validate(t *testing.T) {
	config := &Config{
		AllowedTypes:     []string{"posts"},
		SupportedLocales: []string{"en", "fr"},
		MaxContentLength: 50,
	}

	tests := []struct {
		name    string
		req     *UpdateTranslatableRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid update",
			req: &UpdateTranslatableRequest{
				Locale:  "en",
				Content: "Updated content",
			},
			wantErr: false,
		},
		{
			name: "locale not supported",
			req: &UpdateTranslatableRequest{
				Locale:  "de",
				Content: "Content",
			},
			wantErr: true,
			errMsg:  "locale is not supported",
		},
		{
			name: "empty content",
			req: &UpdateTranslatableRequest{
				Locale:  "en",
				Content: "  ",
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "content too long",
			req: &UpdateTranslatableRequest{
				Locale:  "en",
				Content: strings.Repeat("a", 51),
			},
			wantErr: true,
			errMsg:  "content exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate(config)
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

func TestCreateTranslatableRequest_ToTranslatable(t *testing.T) {
	validUUID := uuid.New().String()
	userID := uuid.New()

	req := &CreateTranslatableRequest{
		TranslatableID: validUUID,
		Translatable:   "posts",
		Locale:         "en",
		Content:        "Test content",
	}

	translatable, err := req.ToTranslatable(&userID)
	if err != nil {
		t.Fatalf("ToTranslatable() unexpected error: %v", err)
	}

	if translatable == nil {
		t.Fatal("ToTranslatable() returned nil")
	}

	if translatable.ID == uuid.Nil {
		t.Error("ToTranslatable() should generate ID")
	}

	if translatable.UserID == nil || *translatable.UserID != userID {
		t.Errorf("ToTranslatable() UserID = %v, want %v", translatable.UserID, userID)
	}

	expectedTranslatableID, _ := uuid.Parse(validUUID)
	if translatable.TranslatableID != expectedTranslatableID {
		t.Errorf("ToTranslatable() TranslatableID = %v, want %v", translatable.TranslatableID, expectedTranslatableID)
	}

	if translatable.Translatable != "posts" {
		t.Errorf("ToTranslatable() Translatable = %v, want posts", translatable.Translatable)
	}

	if translatable.Locale != "en" {
		t.Errorf("ToTranslatable() Locale = %v, want en", translatable.Locale)
	}

	if translatable.Content != "Test content" {
		t.Errorf("ToTranslatable() Content = %v, want 'Test content'", translatable.Content)
	}

	if translatable.CreatedAt.IsZero() {
		t.Error("ToTranslatable() should set CreatedAt")
	}
}

func TestQueryParams_Validate(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"posts"},
		SupportedLocales:   []string{"en", "fr"},
		PaginationLimit:    20,
		MaxPaginationLimit: 100,
	}

	locale := "en"
	invalidLocale := "de"

	tests := []struct {
		name    string
		params  *QueryParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with defaults",
			params: &QueryParams{
				Limit:  0,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "valid with custom values",
			params: &QueryParams{
				Limit:  50,
				Offset: 10,
				Locale: &locale,
			},
			wantErr: false,
		},
		{
			name: "limit too large",
			params: &QueryParams{
				Limit:  101,
				Offset: 0,
			},
			wantErr: true,
			errMsg:  "limit out of range",
		},
		{
			name: "negative offset",
			params: &QueryParams{
				Limit:  20,
				Offset: -1,
			},
			wantErr: true,
			errMsg:  "offset must be non-negative",
		},
		{
			name: "unsupported locale",
			params: &QueryParams{
				Limit:  20,
				Offset: 0,
				Locale: &invalidLocale,
			},
			wantErr: true,
			errMsg:  "locale is not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate(config)
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
				if tt.params.Limit == 0 {
					t.Error("Validate() should set default limit")
				}
			}
		})
	}
}
