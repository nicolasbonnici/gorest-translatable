package translatable

import (
	"database/sql"
	"fmt"
	"net/http"
)

type Plugin struct {
	config  *Config
	handler *Handler
	repo    Repository
}

func NewPlugin(config *Config) (*Plugin, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Plugin{
		config: config,
	}, nil
}

func (p *Plugin) Name() string {
	return "translatable"
}

func (p *Plugin) Version() string {
	return "1.0.0"
}

func (p *Plugin) Initialize(db *sql.DB) error {
	p.repo = NewRepository(db)
	p.handler = NewHandler(p.repo, p.config)
	return nil
}

func (p *Plugin) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/translatable", p.handler.Create)
	mux.HandleFunc("GET /api/translatable/{id}", p.handler.GetByID)
	mux.HandleFunc("GET /api/translatable", p.handler.Query)
	mux.HandleFunc("PUT /api/translatable/{id}", p.handler.Update)
	mux.HandleFunc("DELETE /api/translatable/{id}", p.handler.Delete)
}

func (p *Plugin) Shutdown() error {
	return nil
}

func (p *Plugin) Config() *Config {
	return p.config
}

type Router interface {
	POST(path string, handler http.HandlerFunc)
	GET(path string, handler http.HandlerFunc)
	PUT(path string, handler http.HandlerFunc)
	DELETE(path string, handler http.HandlerFunc)
}

func (p *Plugin) RegisterRoutesWithCustomRouter(router Router) {
	router.POST("/api/translatable", p.handler.Create)
	router.GET("/api/translatable/:id", p.handler.GetByID)
	router.GET("/api/translatable", p.handler.Query)
	router.PUT("/api/translatable/:id", p.handler.Update)
	router.DELETE("/api/translatable/:id", p.handler.Delete)
}
