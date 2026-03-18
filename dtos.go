package translatable

import (
	"time"

	"github.com/google/uuid"
)

type TranslatableCreateDTO struct {
	TranslatableID string `json:"translatableId"`
	Translatable   string `json:"translatable"`
	Locale         string `json:"locale"`
	Content        string `json:"content"`
}

type TranslatableUpdateDTO struct {
	Locale  string `json:"locale"`
	Content string `json:"content"`
}

type TranslatableResponseDTO struct {
	ID             uuid.UUID  `json:"id"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	TranslatableID uuid.UUID  `json:"translatable_id"`
	Translatable   string     `json:"translatable"`
	Locale         string     `json:"locale"`
	Content        string     `json:"content"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
