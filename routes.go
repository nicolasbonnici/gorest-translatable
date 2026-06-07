package translatable

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nicolasbonnici/gorest/database"
)

func RegisterRoutes(router fiber.Router, db database.Database, config *Config, translator *Translator, authMiddleware fiber.Handler) {
	RegisterTranslatableRoutes(router, db, config, translator, authMiddleware)
}
