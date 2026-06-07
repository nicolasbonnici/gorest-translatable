package translatable

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
)

type TranslatableResource struct {
	processor      processor.Processor[Translatable, TranslatableCreateDTO, TranslatableUpdateDTO, TranslatableResponseDTO]
	service        *TranslatableService
	translator     *Translator
	authMiddleware fiber.Handler
}

func RegisterTranslatableRoutes(router fiber.Router, db database.Database, config *Config, translator *Translator, authMiddleware fiber.Handler) {
	service := NewTranslatableService(db, config)

	translatableCRUD := crud.New[Translatable](db)
	hooks := NewTranslatableHooks(db, config)
	converter := &TranslatableConverter{}

	fieldMapping := map[string]string{
		"id":              "id",
		"user_id":         "user_id",
		"translatable_id": "translatable_id",
		"translatable":    "translatable",
		"locale":          "locale",
		"content":         "content",
		"updated_at":      "updated_at",
		"created_at":      "created_at",
	}

	proc := processor.New(processor.ProcessorConfig[Translatable, TranslatableCreateDTO, TranslatableUpdateDTO, TranslatableResponseDTO]{
		DB:                 db,
		CRUD:               translatableCRUD,
		Converter:          converter,
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
		FieldMap:           fieldMapping,
		AllowedFields:      []string{"id", "user_id", "translatable_id", "translatable", "locale", "content", "updated_at", "created_at"},
	}).
		WithCreateHook(hooks.CreateHook).
		WithUpdateHook(hooks.UpdateHook).
		WithDeleteHook(hooks.DeleteHook).
		WithGetByIDHook(hooks.GetByIDHook).
		WithGetAllHook(hooks.GetAllHook)

	resource := &TranslatableResource{
		processor:      proc,
		service:        service,
		translator:     translator,
		authMiddleware: authMiddleware,
	}

	router.Post("/translations", resource.Create)
	router.Get("/translations/:id", resource.GetByID)
	router.Get("/translations", resource.GetAll)
	router.Put("/translations/:id", resource.Update)
	router.Delete("/translations/:id", resource.Delete)
	router.Get("/locales", resource.GetLocales)

	if authMiddleware != nil {
		router.Post("/translations/:type/:id/translate", authMiddleware, resource.Translate)
	} else {
		router.Post("/translations/:type/:id/translate", resource.Translate)
	}
}

func (r *TranslatableResource) Create(c fiber.Ctx) error {
	return r.processor.Create(c)
}

func (r *TranslatableResource) GetByID(c fiber.Ctx) error {
	return r.processor.GetByID(c)
}

func (r *TranslatableResource) GetAll(c fiber.Ctx) error {
	return r.processor.GetAll(c)
}

func (r *TranslatableResource) Update(c fiber.Ctx) error {
	return r.processor.Update(c)
}

func (r *TranslatableResource) Delete(c fiber.Ctx) error {
	return r.processor.Delete(c)
}

func (r *TranslatableResource) GetLocales(c fiber.Ctx) error {
	return c.JSON(r.service.GetLocales())
}

func (r *TranslatableResource) Translate(c fiber.Ctx) error {
	if r.translator == nil || *r.translator == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "auto-translation is not configured")
	}

	resourceType := c.Params("type")
	resourceID := c.Params("id")

	var userID *uuid.UUID
	if raw, ok := c.Locals("user_id").(string); ok && raw != "" {
		parsed, err := uuid.Parse(raw)
		if err == nil {
			userID = &parsed
		}
	}

	result, err := (*r.translator).Translate(c.Context(), resourceType, resourceID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(result)
}
