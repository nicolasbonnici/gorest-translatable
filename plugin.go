package translatable

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest-translatable/migrations"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/plugin"
)

type TranslatablePlugin struct {
	config Config
	db     database.Database
}

func NewPlugin() plugin.Plugin {
	return &TranslatablePlugin{}
}

func (p *TranslatablePlugin) Name() string {
	return "translatable"
}

func (p *TranslatablePlugin) Initialize(config map[string]interface{}) error {
	p.config = DefaultConfig()

	if db, ok := config["database"].(database.Database); ok {
		p.db = db
		p.config.Database = db
	}

	if allowedTypes, ok := config["allowed_types"].([]interface{}); ok {
		types := make([]string, 0, len(allowedTypes))
		for _, t := range allowedTypes {
			if str, ok := t.(string); ok {
				types = append(types, str)
			}
		}
		if len(types) > 0 {
			p.config.AllowedTypes = types
		}
	}

	if supportedLocales, ok := config["supported_locales"].([]interface{}); ok {
		locales := make([]string, 0, len(supportedLocales))
		for _, l := range supportedLocales {
			if str, ok := l.(string); ok {
				locales = append(locales, str)
			}
		}
		if len(locales) > 0 {
			p.config.SupportedLocales = locales
		}
	}

	if defaultLocale, ok := config["default_locale"].(string); ok {
		p.config.DefaultLocale = defaultLocale
	}

	if paginationLimit, ok := config["pagination_limit"].(int); ok {
		p.config.PaginationLimit = paginationLimit
	}

	if maxPaginationLimit, ok := config["max_pagination_limit"].(int); ok {
		p.config.MaxPaginationLimit = maxPaginationLimit
	}

	if maxContentLength, ok := config["max_content_length"].(int); ok {
		p.config.MaxContentLength = maxContentLength
	}

	return p.config.Validate()
}

func (p *TranslatablePlugin) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func (p *TranslatablePlugin) SetupEndpoints(app *fiber.App) error {
	if p.db == nil {
		return nil
	}

	RegisterRoutes(app, p.db, &p.config)
	return nil
}

func (p *TranslatablePlugin) MigrationSource() interface{} {
	return migrations.GetMigrations()
}

func (p *TranslatablePlugin) MigrationDependencies() []string {
	return []string{"auth"}
}
