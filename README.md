# GoREST Translatable Plugin


[![CI](https://github.com/nicolasbonnici/gorest-translatable/actions/workflows/ci.yml/badge.svg)](https://github.com/nicolasbonnici/gorest-translatable/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-translatable)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-translatable)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A production-ready plugin for GoREST framework that provides multi-language content support through a polymorphic `translations` table.

## Features

- **CRUD Operations**: Create, Read, Update, Delete translatable content
- **Multi-Resource Support**: Attach translatable content to any resource
- **Security**: Built-in XSS protection, ownership validation, and configurable content limits
- **Flexible Querying**: Filter by resource ID, type, or user

## Installation

```bash
go get github.com/nicolas/gorest-translatable
```

## Database Setup

The plugin includes migrations to create the `translations` table. The migrations will be run automatically when you initialize the plugin with GoREST's migration system.

The table structure includes:
- Multi-database support (PostgreSQL, MySQL, SQLite)
- Polymorphic relationship via `translatable_id` and `translatable` columns
- Locale support for multi-language content
- User ownership tracking
- Automatic timestamps

## Usage

### 1. Initialize the Plugin

```go
package main

import (
    "database/sql"
    "log"
    "net/http"

    "github.com/nicolas/gorest-translatable"
)

func main() {
    // Your database connection
    db, err := sql.Open("postgres", "your-connection-string")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Configure the plugin
    config := &translatable.Config{
        AllowedTables:    []string{"posts", "articles", "products"},
        MaxContentLength: 10240, // 10KB
    }

    // Create plugin instance
    plugin, err := translatable.NewPlugin(config)
    if err != nil {
        log.Fatal(err)
    }

    // Initialize with database
    if err := plugin.Initialize(db); err != nil {
        log.Fatal(err)
    }

    // Register routes
    mux := http.NewServeMux()
    plugin.RegisterRoutes(mux)

    // Start server
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### 2. Configuration

```go
type Config struct {
    // List of allowed table names for the translatable column
    AllowedTables []string

    // Maximum content length in bytes (default: 10KB, max: 1MB)
    MaxContentLength int
}
```

**Example:**

```go
config := &translatable.Config{
    AllowedTables:    []string{"posts", "articles", "products", "categories"},
    MaxContentLength: 20480, // 20KB
}
```

## API Endpoints

### Create Translation

```http
POST /api/translations
Content-Type: application/json

{
  "translatable_id": "550e8400-e29b-41d4-a716-446655440000",
  "translatable": "posts",
  "locale": "en",
  "content": "This is the translatable content"
}
```

**Response:**

```json
{
  "id": "650e8400-e29b-41d4-a716-446655440000",
  "user_id": "750e8400-e29b-41d4-a716-446655440000",
  "translatable_id": "550e8400-e29b-41d4-a716-446655440000",
  "translatable": "posts",
  "content": "This is the translatable content",
  "created_at": "2025-12-30T10:00:00Z"
}
```

### Get Translation by ID

```http
GET /api/translations/{id}
```

**Response:**

```json
{
  "id": "650e8400-e29b-41d4-a716-446655440000",
  "user_id": "750e8400-e29b-41d4-a716-446655440000",
  "translatable_id": "550e8400-e29b-41d4-a716-446655440000",
  "translatable": "posts",
  "content": "This is the translatable content",
  "created_at": "2025-12-30T10:00:00Z",
  "updated_at": "2025-12-30T11:00:00Z"
}
```

### Query Translations

```http
GET /api/translations?translatable_id={uuid}&translatable=posts&locale=en&limit=20&offset=0
```

**Query Parameters:**

- `translatable_id` (optional): Filter by parent resource UUID
- `translatable` (optional): Filter by resource type
- `user_id` (optional): Filter by user UUID
- `limit` (optional): Results per page (default: 20, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response:**

```json
{
  "data": [
    {
      "id": "650e8400-e29b-41d4-a716-446655440000",
      "translatable_id": "550e8400-e29b-41d4-a716-446655440000",
      "translatable": "posts",
      "content": "Content here",
      "created_at": "2025-12-30T10:00:00Z"
    }
  ],
  "limit": 20,
  "offset": 0
}
```

### Update Translation

```http
PUT /api/translations/{id}
Content-Type: application/json

{
  "locale": "en",
  "content": "Updated content"
}
```

**Note:** Users can only update their own translation entries (validated via `user_id` from auth middleware).

### Delete Translation

```http
DELETE /api/translations/{id}
```

**Note:** Users can only delete their own translation entries.

## Security Features

### 1. XSS Protection

All content is automatically HTML-escaped to prevent XSS attacks:

```go
// Input
content := "<script>alert('xss')</script>"

// Stored
content := "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
```

### 2. Ownership Validation

The plugin uses GoREST's auth middleware to extract `user_id` from the request context. Users can only update/delete their own entries.

### 3. Input Validation

- `translatable`: Must be in the allowed list
- `translatable_id`: Must be a valid UUID
- `content`: Required, trimmed, max length enforced

### 4. Content Length Limits

Configurable maximum content length (default: 10KB, max: 1MB) prevents abuse.

## Integration with GoREST Middleware

The plugin relies on GoREST's existing middleware:

```go
// GoREST auth middleware should set user_id in context
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract user from JWT/session
        userID := extractUserID(r)

        // Set in context
        ctx := context.WithValue(r.Context(), "user_id", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Error Handling

The plugin returns standard HTTP status codes:

- `200 OK`: Success
- `201 Created`: Resource created
- `400 Bad Request`: Validation error
- `404 Not Found`: Resource not found or no permission
- `500 Internal Server Error`: Server error

**Error Response Format:**

```json
{
  "error": "Error message here"
}
```

## Examples

### Example 1: Add Translation Content to a Post

```bash
curl -X POST http://localhost:8080/api/translations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "translatable_id": "550e8400-e29b-41d4-a716-446655440000",
    "translatable": "posts",
    "locale": "en",
    "content": "This is a comment on the post"
  }'
```

### Example 2: Get All Translations for a Post

```bash
curl "http://localhost:8080/api/translations?translatable_id=550e8400-e29b-41d4-a716-446655440000&translatable=posts"
```

### Example 3: Update a Translation

```bash
curl -X PUT http://localhost:8080/api/translations/650e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "locale": "en",
    "content": "Updated content"
  }'
```

### Example 4: Delete a Translation

```bash
curl -X DELETE http://localhost:8080/api/translations/650e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Testing

```bash
# Run tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

## Production Checklist

- ✅ Run database migration
- ✅ Configure allowed tables in config
- ✅ Set appropriate `max_content_length`
- ✅ Ensure GoREST auth middleware is active
- ✅ Set up monitoring and logging
- ✅ Configure rate limiting (via GoREST)
- ✅ Enable HTTPS in production
- ✅ Regular database backups

---

## Git Hooks

This directory contains git hooks for the GoREST plugin to maintain code quality.

### Available Hooks

#### pre-commit

Runs before each commit to ensure code quality:
- **Linting**: Runs `make lint` to check code style and potential issues
- **Tests**: Runs `make test` to verify all tests pass

### Installation

#### Automatic Installation

Run the install script from the project root:

```bash
./.githooks/install.sh
```

#### Manual Installation

Copy the hooks to your `.git/hooks` directory:

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---


## License

MIT License

## Contributing

Pull requests are welcome! Please ensure:

1. Tests pass
2. Code is formatted with `gofmt`
3. Changes are documented

## Support

For issues and questions, please open an issue on GitHub.
