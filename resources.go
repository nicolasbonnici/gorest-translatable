package translatable

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nicolasbonnici/gorest/database"
)

type TranslatableResource struct {
	db      database.Database
	config  *Config
	service *TranslatableService
}

func RegisterTranslatableRoutes(app *fiber.App, db database.Database, config *Config) {
	service := NewTranslatableService(db, config)

	resource := &TranslatableResource{
		db:      db,
		config:  config,
		service: service,
	}

	app.Post("/translations", resource.Create)
	app.Get("/translations/:id", resource.GetByID)
	app.Get("/translations", resource.Query)
	app.Put("/translations/:id", resource.Update)
	app.Delete("/translations/:id", resource.Delete)
	app.Get("/locales", resource.GetLocales)
}

func (r *TranslatableResource) Create(c *fiber.Ctx) error {
	var req CreateTranslatableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := req.Validate(r.config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userID := getUserIDFromFiberContext(c)

	translatable, err := req.ToTranslatable(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := r.createInDB(c.Context(), translatable); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create translation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(translatable)
}

func (r *TranslatableResource) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	translatable, err := r.getByIDFromDB(c.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Translation not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get translation",
		})
	}

	return c.JSON(translatable)
}

func (r *TranslatableResource) Query(c *fiber.Ctx) error {
	params := QueryParams{
		Limit:  r.config.PaginationLimit,
		Offset: 0,
	}

	if translatableIDStr := c.Query("translatable_id"); translatableIDStr != "" {
		if id, err := uuid.Parse(translatableIDStr); err == nil {
			params.TranslatableID = &id
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid translatable_id",
			})
		}
	}

	if translatable := c.Query("translatable"); translatable != "" {
		params.Translatable = &translatable
	}

	if locale := c.Query("locale"); locale != "" {
		params.Locale = &locale
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			params.UserID = &id
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid user_id",
			})
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	if err := params.Validate(r.config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	results, total, err := r.queryFromDB(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query translations",
		})
	}

	return c.JSON(fiber.Map{
		"@context":         "http://www.w3.org/ns/hydra/context.jsonld",
		"@id":              "/translations",
		"@type":            "hydra:Collection",
		"hydra:totalItems": total,
		"hydra:member":     results,
	})
}

func (r *TranslatableResource) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	var req UpdateTranslatableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := req.Validate(r.config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userID := getUserIDFromFiberContext(c)

	if err := r.updateInDB(c.Context(), id, req.Content, req.Locale, userID); err != nil {
		if err.Error() == "translation not found" ||
			err.Error() == "translation not found or you don't have permission to update it" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update translation",
		})
	}

	translatable, err := r.getByIDFromDB(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get updated translation",
		})
	}

	return c.JSON(translatable)
}

func (r *TranslatableResource) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	userID := getUserIDFromFiberContext(c)

	if err := r.deleteFromDB(c.Context(), id, userID); err != nil {
		if err.Error() == "translation not found" ||
			err.Error() == "translation not found or you don't have permission to delete it" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete translation",
		})
	}

	return c.JSON(fiber.Map{"message": "Translation deleted successfully"})
}

// GetLocales returns all available locales with default flag
func (r *TranslatableResource) GetLocales(c *fiber.Ctx) error {
	locales := r.service.GetLocales()
	return c.JSON(locales)
}

func getUserIDFromFiberContext(c *fiber.Ctx) *uuid.UUID {
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(uuid.UUID); ok {
			return &uid
		}
		if uidStr, ok := userID.(string); ok {
			if uid, err := uuid.Parse(uidStr); err == nil {
				return &uid
			}
		}
	}
	return nil
}

// Database methods

func (r *TranslatableResource) createInDB(ctx context.Context, t *Translatable) error {
	query := `
		INSERT INTO translations (id, user_id, translatable_id, translatable, locale, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		t.ID,
		t.UserID,
		t.TranslatableID,
		t.Translatable,
		t.Locale,
		t.Content,
		t.CreatedAt,
	)

	return err
}

func (r *TranslatableResource) getByIDFromDB(ctx context.Context, id uuid.UUID) (*Translatable, error) {
	query := `
		SELECT id, user_id, translatable_id, translatable, locale, content, updated_at, created_at
		FROM translations
		WHERE id = $1
	`

	var t Translatable
	err := r.db.QueryRow(ctx, query, id).Scan(
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

func (r *TranslatableResource) queryFromDB(ctx context.Context, params QueryParams) ([]*Translatable, int, error) {
	// Build query
	query := `
		SELECT id, user_id, translatable_id, translatable, locale, content, updated_at, created_at
		FROM translations
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM translations WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	if params.TranslatableID != nil {
		clause := " AND translatable_id = $" + strconv.Itoa(argCount)
		query += clause
		countQuery += clause
		args = append(args, *params.TranslatableID)
		argCount++
	}

	if params.Translatable != nil {
		clause := " AND translatable = $" + strconv.Itoa(argCount)
		query += clause
		countQuery += clause
		args = append(args, *params.Translatable)
		argCount++
	}

	if params.Locale != nil {
		clause := " AND locale = $" + strconv.Itoa(argCount)
		query += clause
		countQuery += clause
		args = append(args, *params.Locale)
		argCount++
	}

	if params.UserID != nil {
		clause := " AND user_id = $" + strconv.Itoa(argCount)
		query += clause
		countQuery += clause
		args = append(args, *params.UserID)
		argCount++
	}

	// Get total count
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Add pagination
	query += " ORDER BY created_at DESC"
	query += " LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []*Translatable
	for rows.Next() {
		var t Translatable
		err := rows.Scan(
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
			return nil, 0, err
		}
		results = append(results, &t)
	}

	return results, total, rows.Err()
}

func (r *TranslatableResource) updateInDB(ctx context.Context, id uuid.UUID, content, locale string, userID *uuid.UUID) error {
	query := `
		UPDATE translations
		SET content = $1, locale = $2, updated_at = $3
		WHERE id = $4
	`

	args := []interface{}{content, locale, time.Now(), id}
	if userID != nil {
		query += " AND user_id = $5"
		args = append(args, *userID)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		if userID != nil {
			return errors.New("translation not found or you don't have permission to update it")
		}
		return errors.New("translation not found")
	}

	return nil
}

func (r *TranslatableResource) deleteFromDB(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	query := `DELETE FROM translations WHERE id = $1`

	args := []interface{}{id}
	if userID != nil {
		query += " AND user_id = $2"
		args = append(args, *userID)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		if userID != nil {
			return errors.New("translation not found or you don't have permission to delete it")
		}
		return errors.New("translation not found")
	}

	return nil
}
