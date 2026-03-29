package translatable

import (
	"context"
	"errors"
	"html"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	auth "github.com/nicolasbonnici/gorest-auth"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/query"
)

type TranslatableHooks struct {
	db     database.Database
	config *Config
}

func NewTranslatableHooks(db database.Database, config *Config) *TranslatableHooks {
	return &TranslatableHooks{
		db:     db,
		config: config,
	}
}

func (h *TranslatableHooks) CreateHook(c *fiber.Ctx, dto TranslatableCreateDTO, model *Translatable) error {
	if _, err := uuid.Parse(dto.TranslatableID); err != nil {
		return fiber.NewError(400, "translatable_id must be a valid UUID")
	}

	if !h.config.IsAllowedType(dto.Translatable) {
		return fiber.NewError(400, "translatable type is not allowed")
	}

	if !h.config.IsSupportedLocale(dto.Locale) {
		return fiber.NewError(400, "locale is not supported")
	}

	content := strings.TrimSpace(dto.Content)
	if content == "" {
		return fiber.NewError(400, "content cannot be empty")
	}

	if len(content) > h.config.MaxContentLength {
		return fiber.NewError(400, "content exceeds maximum length")
	}

	model.Content = html.EscapeString(content)

	userID := getUserIDFromFiberContext(c)
	if userID != nil {
		model.UserID = userID
	}

	return nil
}

func (h *TranslatableHooks) UpdateHook(c *fiber.Ctx, dto TranslatableUpdateDTO, model *Translatable) error {
	if !h.config.IsSupportedLocale(dto.Locale) {
		return fiber.NewError(400, "locale is not supported")
	}

	content := strings.TrimSpace(dto.Content)
	if content == "" {
		return fiber.NewError(400, "content cannot be empty")
	}

	if len(content) > h.config.MaxContentLength {
		return fiber.NewError(400, "content exceeds maximum length")
	}

	model.Content = html.EscapeString(content)

	id := c.Params("id")
	ctx := auth.Context(c)
	userID := getUserIDFromFiberContext(c)

	existing, err := h.getTranslatable(ctx, id)
	if err != nil {
		return fiber.NewError(404, "Translation not found")
	}

	if userID != nil && existing.UserID != nil && *existing.UserID != *userID {
		return fiber.NewError(403, "You can only update your own translations")
	}

	return nil
}

func (h *TranslatableHooks) DeleteHook(c *fiber.Ctx, id any) error {
	ctx := auth.Context(c)
	userID := getUserIDFromFiberContext(c)

	existing, err := h.getTranslatable(ctx, id)
	if err != nil {
		return fiber.NewError(404, "Translation not found")
	}

	if userID != nil && existing.UserID != nil && *existing.UserID != *userID {
		return fiber.NewError(403, "You can only delete your own translations")
	}

	return nil
}

func (h *TranslatableHooks) GetByIDHook(c *fiber.Ctx, id any) error {
	return nil
}

func (h *TranslatableHooks) GetAllHook(c *fiber.Ctx, conditions *[]query.Condition, orderBy *[]crud.OrderByClause) error {
	return nil
}

func (h *TranslatableHooks) getTranslatable(ctx context.Context, id any) (*Translatable, error) {
	var t Translatable
	idStr, ok := id.(string)
	if !ok {
		return nil, errors.New("invalid ID type")
	}

	idUUID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	sql := "SELECT * FROM translations WHERE id = " + h.db.Dialect().Placeholder(1)
	err = h.db.QueryRow(ctx, sql, idUUID).Scan(
		&t.ID,
		&t.UserID,
		&t.TranslatableID,
		&t.Translatable,
		&t.Locale,
		&t.Content,
		&t.UpdatedAt,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func getUserIDFromFiberContext(c *fiber.Ctx) *uuid.UUID {
	user := auth.GetAuthenticatedUser(c)
	if user == nil {
		return nil
	}

	userUUID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil
	}

	return &userUUID
}
