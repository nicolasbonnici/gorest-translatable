package translatable

import (
	"github.com/nicolasbonnici/gorest/plugin"
)

// Compile-time check that TranslatablePlugin implements required interfaces
var (
	_ plugin.Plugin            = (*TranslatablePlugin)(nil)
	_ plugin.MigrationProvider = (*TranslatablePlugin)(nil)
	_ plugin.EndpointSetup     = (*TranslatablePlugin)(nil)
)
