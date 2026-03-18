package translatable

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
)

func setupTestApp(db database.Database, config *Config) (*fiber.App, *TranslatableResource) {
	app := fiber.New()
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
	return app, resource
}

func TestTranslatableResourceInitialization(t *testing.T) {
	config := DefaultConfig()
	_, resource := setupTestApp(nil, &config)

	if resource == nil {
		t.Fatal("TranslatableResource should not be nil")
	}

	if resource.service == nil {
		t.Fatal("TranslatableResource service should not be nil")
	}
}
