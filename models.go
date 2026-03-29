package translatable

import (
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

func (Translatable) TableName() string {
	return "translations"
}

type LocaleInfo struct {
	Locale    string `json:"locale"`
	IsDefault bool   `json:"isDefault"`
}
