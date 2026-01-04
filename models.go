package translatable

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Translatable struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	TranslatableID uuid.UUID  `json:"translatable_id" db:"translatable_id"`
	Translatable   string     `json:"translatable" db:"translatable"`
	Locale         string     `json:"locale" db:"locale"`
	Content        string     `json:"content" db:"content"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type CreateTranslatableRequest struct {
	TranslatableID string `json:"translatableId" validate:"required,uuid"`
	Translatable   string `json:"translatable" validate:"required"`
	Locale         string `json:"locale" validate:"required"`
	Content        string `json:"content" validate:"required"`
}

type UpdateTranslatableRequest struct {
	Locale  string `json:"locale" validate:"required"`
	Content string `json:"content" validate:"required"`
}

func (r *CreateTranslatableRequest) Validate(config *Config) error {
	if _, err := uuid.Parse(r.TranslatableID); err != nil {
		return errors.New("translatable_id must be a valid UUID")
	}

	if !config.IsAllowedType(r.Translatable) {
		return errors.New("translatable type is not allowed")
	}

	if !config.IsSupportedLocale(r.Locale) {
		return errors.New("locale is not supported")
	}

	r.Content = strings.TrimSpace(r.Content)
	if r.Content == "" {
		return errors.New("content cannot be empty")
	}

	if len(r.Content) > config.MaxContentLength {
		return errors.New("content exceeds maximum length")
	}

	// Sanitize HTML to prevent XSS
	r.Content = html.EscapeString(r.Content)

	return nil
}

func (r *UpdateTranslatableRequest) Validate(config *Config) error {
	if !config.IsSupportedLocale(r.Locale) {
		return errors.New("locale is not supported")
	}

	r.Content = strings.TrimSpace(r.Content)
	if r.Content == "" {
		return errors.New("content cannot be empty")
	}

	if len(r.Content) > config.MaxContentLength {
		return errors.New("content exceeds maximum length")
	}

	// Sanitize HTML to prevent XSS
	r.Content = html.EscapeString(r.Content)

	return nil
}

func (r *CreateTranslatableRequest) ToTranslatable(userID *uuid.UUID) (*Translatable, error) {
	translatableID, err := uuid.Parse(r.TranslatableID)
	if err != nil {
		return nil, err
	}

	return &Translatable{
		ID:             uuid.New(),
		UserID:         userID,
		TranslatableID: translatableID,
		Translatable:   r.Translatable,
		Locale:         r.Locale,
		Content:        r.Content,
		CreatedAt:      time.Now(),
	}, nil
}

type QueryParams struct {
	TranslatableID *uuid.UUID
	Translatable   *string
	Locale         *string
	UserID         *uuid.UUID
	Limit          int
	Offset         int
}

func (q *QueryParams) Validate(config *Config) error {
	if q.Limit == 0 {
		q.Limit = config.PaginationLimit
	}
	if q.Limit < 1 || q.Limit > config.MaxPaginationLimit {
		return errors.New("limit out of range")
	}
	if q.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	if q.Locale != nil && !config.IsSupportedLocale(*q.Locale) {
		return errors.New("locale is not supported")
	}
	return nil
}
