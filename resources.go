package translatable

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
)

type TranslatableResource struct {
	processor processor.Processor[Translatable, TranslatableCreateDTO, TranslatableUpdateDTO, TranslatableResponseDTO]
	service   *TranslatableService
}

func RegisterTranslatableRoutes(app *fiber.App, db database.Database, config *Config) {
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
		processor: proc,
		service:   service,
	}

	app.Post("/translations", resource.Create)
	app.Get("/translations/:id", resource.GetByID)
	app.Get("/translations", resource.GetAll)
	app.Put("/translations/:id", resource.Update)
	app.Delete("/translations/:id", resource.Delete)
	app.Get("/locales", resource.GetLocales)
}

func (r *TranslatableResource) Create(c *fiber.Ctx) error {
	return r.processor.Create(c)
}

func (r *TranslatableResource) GetByID(c *fiber.Ctx) error {
	return r.processor.GetByID(c)
}

func (r *TranslatableResource) GetAll(c *fiber.Ctx) error {
	return r.processor.GetAll(c)
}

func (r *TranslatableResource) Update(c *fiber.Ctx) error {
	return r.processor.Update(c)
}

func (r *TranslatableResource) Delete(c *fiber.Ctx) error {
	return r.processor.Delete(c)
}

func (r *TranslatableResource) GetLocales(c *fiber.Ctx) error {
	locales := r.service.GetLocales()
	return c.JSON(locales)
}
