package translatable

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/stretchr/testify/assert"

	"github.com/nicolasbonnici/gorest-translatable/mocks"
)

func setupTestApp(db database.Database, config *Config) (*fiber.App, *TranslatableResource) {
	app := fiber.New()
	resource := &TranslatableResource{
		db:     db,
		config: config,
	}
	return app, resource
}

func TestTranslatableResource_Create(t *testing.T) {
	config := &Config{
		AllowedTypes:     []string{"products", "articles"},
		SupportedLocales: []string{"en", "fr"},
		DefaultLocale:    "en",
		MaxContentLength: 10000,
	}

	tests := []struct {
		name           string
		body           interface{}
		userID         *uuid.UUID
		mockExecFunc   func(ctx context.Context, query string, args ...interface{}) (database.Result, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "success - valid request",
			body: map[string]interface{}{
				"translatableId": uuid.New().String(),
				"translatable":   "products",
				"locale":         "en",
				"content":        "Test content",
			},
			userID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return mocks.NewMockResult(1), nil
			},
			expectedStatus: 201,
			checkResponse: func(t *testing.T, body []byte) {
				var response Translatable
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "products", response.Translatable)
				assert.Equal(t, "en", response.Locale)
			},
		},
		{
			name:           "error - invalid JSON",
			body:           "invalid json",
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid request body")
			},
		},
		{
			name: "error - invalid translatable_id",
			body: map[string]interface{}{
				"translatableId":"not-a-uuid",
				"translatable":    "products",
				"locale":          "en",
				"content":         "Test content",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "translatable_id")
			},
		},
		{
			name: "error - invalid translatable type",
			body: map[string]interface{}{
				"translatableId":uuid.New().String(),
				"translatable":    "invalid_type",
				"locale":          "en",
				"content":         "Test content",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "translatable")
			},
		},
		{
			name: "error - unsupported locale",
			body: map[string]interface{}{
				"translatableId":uuid.New().String(),
				"translatable":    "products",
				"locale":          "de",
				"content":         "Test content",
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "locale")
			},
		},
		{
			name: "error - database error",
			body: map[string]interface{}{
				"translatableId":uuid.New().String(),
				"translatable":    "products",
				"locale":          "en",
				"content":         "Test content",
			},
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to create translation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mocks.MockDatabase{
				ExecFunc: tt.mockExecFunc,
			}

			app, resource := setupTestApp(mockDB, config)

			if tt.userID != nil {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals("user_id", *tt.userID)
					return c.Next()
				})
			}

			app.Post("/translations", resource.Create)

			var bodyReader io.Reader
			if str, ok := tt.body.(string); ok {
				bodyReader = bytes.NewReader([]byte(str))
			} else {
				bodyBytes, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			req := httptest.NewRequest("POST", "/translations", bodyReader)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			tt.checkResponse(t, body)
		})
	}
}

func TestTranslatableResource_GetByID(t *testing.T) {
	config := DefaultConfig()

	validID := uuid.New()

	tests := []struct {
		name            string
		id              string
		mockQueryRowFunc func(ctx context.Context, query string, args ...interface{}) database.Row
		expectedStatus  int
		checkResponse   func(t *testing.T, body []byte)
	}{
		{
			name: "success - translation found",
			id:   validID.String(),
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						if len(dest) >= 8 {
							*dest[0].(*uuid.UUID) = validID
							*dest[1].(**uuid.UUID) = nil
							*dest[2].(*uuid.UUID) = uuid.New()
							*dest[3].(*string) = "products"
							*dest[4].(*string) = "en"
							*dest[5].(*string) = "Test content"
							// dest[6] is **time.Time for updated_at
							// dest[7] is *time.Time for created_at
						}
						return nil
					},
				}
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response Translatable
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, validID, response.ID)
				assert.Equal(t, "products", response.Translatable)
			},
		},
		{
			name:           "error - invalid ID",
			id:             "not-a-uuid",
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid ID")
			},
		},
		{
			name: "error - translation not found",
			id:   validID.String(),
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						return pgx.ErrNoRows
					},
				}
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Translation not found")
			},
		},
		{
			name: "error - database error",
			id:   validID.String(),
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						return errors.New("database error")
					},
				}
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to get translation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mocks.MockDatabase{
				QueryRowFunc: tt.mockQueryRowFunc,
			}

			app, resource := setupTestApp(mockDB, &config)
			app.Get("/translations/:id", resource.GetByID)

			req := httptest.NewRequest("GET", "/translations/"+tt.id, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			tt.checkResponse(t, body)
		})
	}
}

func TestTranslatableResource_Query(t *testing.T) {
	config := &Config{
		AllowedTypes:       []string{"products", "articles"},
		SupportedLocales:   []string{"en", "fr"},
		DefaultLocale:      "en",
		PaginationLimit:    10,
		MaxPaginationLimit: 100,
	}

	tests := []struct {
		name             string
		queryString      string
		mockQueryRowFunc func(ctx context.Context, query string, args ...interface{}) database.Row
		mockQueryFunc    func(ctx context.Context, query string, args ...interface{}) (database.Rows, error)
		expectedStatus   int
		checkResponse    func(t *testing.T, body []byte)
	}{
		{
			name:        "success - no filters",
			queryString: "",
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						*dest[0].(*int) = 5
						return nil
					},
				}
			},
			mockQueryFunc: func(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
				return mocks.NewMockRows(2), nil
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(5), response["hydra:totalItems"])
				assert.NotNil(t, response["hydra:member"])
			},
		},
		{
			name:        "success - with translatable_id filter",
			queryString: "?translatable_id=" + uuid.New().String(),
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						*dest[0].(*int) = 1
						return nil
					},
				}
			},
			mockQueryFunc: func(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
				return mocks.NewMockRows(1), nil
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(1), response["hydra:totalItems"])
			},
		},
		{
			name:           "error - invalid translatable_id",
			queryString:    "?translatable_id=not-a-uuid",
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid translatable_id")
			},
		},
		{
			name:        "success - with locale filter",
			queryString: "?locale=en",
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						*dest[0].(*int) = 3
						return nil
					},
				}
			},
			mockQueryFunc: func(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
				return mocks.NewMockRows(3), nil
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(3), response["hydra:totalItems"])
			},
		},
		{
			name:        "success - with pagination",
			queryString: "?limit=5&offset=10",
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						*dest[0].(*int) = 50
						return nil
					},
				}
			},
			mockQueryFunc: func(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
				return mocks.NewMockRows(5), nil
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(50), response["hydra:totalItems"])
			},
		},
		{
			name:        "error - database error on count",
			queryString: "",
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						return errors.New("database error")
					},
				}
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to query translations")
			},
		},
		{
			name:        "error - database error on query",
			queryString: "",
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						*dest[0].(*int) = 5
						return nil
					},
				}
			},
			mockQueryFunc: func(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to query translations")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mocks.MockDatabase{
				QueryRowFunc: tt.mockQueryRowFunc,
				QueryFunc:    tt.mockQueryFunc,
			}

			app, resource := setupTestApp(mockDB, config)
			app.Get("/translations", resource.Query)

			req := httptest.NewRequest("GET", "/translations"+tt.queryString, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			tt.checkResponse(t, body)
		})
	}
}

func TestTranslatableResource_Update(t *testing.T) {
	config := &Config{
		AllowedTypes:     []string{"products"},
		SupportedLocales: []string{"en", "fr"},
		DefaultLocale:    "en",
		MaxContentLength: 10000,
	}

	validID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name             string
		id               string
		body             interface{}
		userID           *uuid.UUID
		mockExecFunc     func(ctx context.Context, query string, args ...interface{}) (database.Result, error)
		mockQueryRowFunc func(ctx context.Context, query string, args ...interface{}) database.Row
		expectedStatus   int
		checkResponse    func(t *testing.T, body []byte)
	}{
		{
			name: "success - update translation",
			id:   validID.String(),
			body: map[string]interface{}{
				"content": "Updated content",
				"locale":  "en",
			},
			userID: &userID,
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return mocks.NewMockResult(1), nil
			},
			mockQueryRowFunc: func(ctx context.Context, query string, args ...interface{}) database.Row {
				return &mocks.MockRow{
					ScanFunc: func(dest ...interface{}) error {
						if len(dest) >= 8 {
							*dest[0].(*uuid.UUID) = validID
							*dest[1].(**uuid.UUID) = &userID
							*dest[2].(*uuid.UUID) = uuid.New()
							*dest[3].(*string) = "products"
							*dest[4].(*string) = "en"
							*dest[5].(*string) = "Updated content"
							// dest[6] is **time.Time for updated_at
							// dest[7] is *time.Time for created_at
						}
						return nil
					},
				}
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response Translatable
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, validID, response.ID)
				assert.Equal(t, "Updated content", response.Content)
			},
		},
		{
			name:           "error - invalid ID",
			id:             "not-a-uuid",
			body:           map[string]interface{}{"content": "test", "locale": "en"},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid ID")
			},
		},
		{
			name:           "error - invalid JSON",
			id:             validID.String(),
			body:           "invalid json",
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid request body")
			},
		},
		{
			name: "error - translation not found",
			id:   validID.String(),
			body: map[string]interface{}{
				"content": "Updated content",
				"locale":  "en",
			},
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return mocks.NewMockResult(0), nil
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "translation not found")
			},
		},
		{
			name: "error - database error",
			id:   validID.String(),
			body: map[string]interface{}{
				"content": "Updated content",
				"locale":  "en",
			},
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to update translation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mocks.MockDatabase{
				ExecFunc:     tt.mockExecFunc,
				QueryRowFunc: tt.mockQueryRowFunc,
			}

			app, resource := setupTestApp(mockDB, config)

			if tt.userID != nil {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals("user_id", *tt.userID)
					return c.Next()
				})
			}

			app.Put("/translations/:id", resource.Update)

			var bodyReader io.Reader
			if str, ok := tt.body.(string); ok {
				bodyReader = bytes.NewReader([]byte(str))
			} else {
				bodyBytes, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			req := httptest.NewRequest("PUT", "/translations/"+tt.id, bodyReader)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			tt.checkResponse(t, body)
		})
	}
}

func TestTranslatableResource_Delete(t *testing.T) {
	config := DefaultConfig()

	validID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		id             string
		userID         *uuid.UUID
		mockExecFunc   func(ctx context.Context, query string, args ...interface{}) (database.Result, error)
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:   "success - delete translation",
			id:     validID.String(),
			userID: &userID,
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return mocks.NewMockResult(1), nil
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], "deleted successfully")
			},
		},
		{
			name:           "error - invalid ID",
			id:             "not-a-uuid",
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid ID")
			},
		},
		{
			name: "error - translation not found",
			id:   validID.String(),
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return mocks.NewMockResult(0), nil
			},
			expectedStatus: 404,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "translation not found")
			},
		},
		{
			name: "error - database error",
			id:   validID.String(),
			mockExecFunc: func(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: 500,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to delete translation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mocks.MockDatabase{
				ExecFunc: tt.mockExecFunc,
			}

			app, resource := setupTestApp(mockDB, &config)

			if tt.userID != nil {
				app.Use(func(c *fiber.Ctx) error {
					c.Locals("user_id", *tt.userID)
					return c.Next()
				})
			}

			app.Delete("/translations/:id", resource.Delete)

			req := httptest.NewRequest("DELETE", "/translations/"+tt.id, nil)

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			tt.checkResponse(t, body)
		})
	}
}

func TestGetUserIDFromFiberContext(t *testing.T) {
	tests := []struct {
		name     string
		locals   map[string]interface{}
		expected *uuid.UUID
	}{
		{
			name: "success - UUID type",
			locals: map[string]interface{}{
				"user_id": uuid.New(),
			},
			expected: func() *uuid.UUID {
				id := uuid.New()
				return &id
			}(),
		},
		{
			name: "success - string type",
			locals: map[string]interface{}{
				"user_id": uuid.New().String(),
			},
			expected: func() *uuid.UUID {
				id := uuid.New()
				return &id
			}(),
		},
		{
			name:     "nil - no user_id",
			locals:   map[string]interface{}{},
			expected: nil,
		},
		{
			name: "nil - invalid UUID string",
			locals: map[string]interface{}{
				"user_id": "not-a-uuid",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				for key, value := range tt.locals {
					c.Locals(key, value)
				}
				result := getUserIDFromFiberContext(c)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
				}
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			_, err := app.Test(req)
			assert.NoError(t, err)
		})
	}
}
