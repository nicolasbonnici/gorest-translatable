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
	Content        string     `json:"content" db:"content"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type CreateTranslatableRequest struct {
	TranslatableID string `json:"translatable_id" validate:"required,uuid"`
	Translatable   string `json:"translatable" validate:"required"`
	Content        string `json:"content" validate:"required"`
}

type UpdateTranslatableRequest struct {
	Content string `json:"content" validate:"required"`
}

func (r *CreateTranslatableRequest) Validate(config *Config) error {
	if _, err := uuid.Parse(r.TranslatableID); err != nil {
		return errors.New("translatable_id must be a valid UUID")
	}

	if !config.IsAllowedTable(r.Translatable) {
		return errors.New("translatable type is not allowed")
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
		Content:        r.Content,
		CreatedAt:      time.Now(),
	}, nil
}

type QueryParams struct {
	TranslatableID *uuid.UUID
	Translatable   *string
	UserID         *uuid.UUID
	Limit          int
	Offset         int
}

func (q *QueryParams) Validate() error {
	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.Limit < 1 || q.Limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	if q.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	return nil
}
