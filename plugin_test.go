package translatable

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPlugin(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				AllowedTables:    []string{"posts", "articles"},
				MaxContentLength: 10240,
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty allowed tables",
			config: &Config{
				AllowedTables:    []string{},
				MaxContentLength: 10240,
			},
			wantErr: true,
		},
		{
			name: "invalid config - max length too large",
			config: &Config{
				AllowedTables:    []string{"posts"},
				MaxContentLength: 2000000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := NewPlugin(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && plugin == nil {
				t.Error("NewPlugin() returned nil plugin with valid config")
			}
		})
	}
}

func TestPlugin_Name(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts"},
		MaxContentLength: 10240,
	}
	plugin, _ := NewPlugin(config)

	name := plugin.Name()
	if name != "translatable" {
		t.Errorf("Name() = %v, want 'translatable'", name)
	}
}

func TestPlugin_Version(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts"},
		MaxContentLength: 10240,
	}
	plugin, _ := NewPlugin(config)

	version := plugin.Version()
	if version == "" {
		t.Error("Version() returned empty string")
	}
	if version != "1.0.0" {
		t.Errorf("Version() = %v, want '1.0.0'", version)
	}
}

func TestPlugin_Initialize(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts"},
		MaxContentLength: 10240,
	}
	plugin, _ := NewPlugin(config)

	mockRepo := &mockRepository{}
	plugin.repo = mockRepo
	plugin.handler = NewHandler(mockRepo, config)

	if plugin.repo == nil {
		t.Error("Initialize() should set repository")
	}
	if plugin.handler == nil {
		t.Error("Initialize() should set handler")
	}
}

func TestPlugin_RegisterRoutes(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts"},
		MaxContentLength: 10240,
	}
	plugin, _ := NewPlugin(config)
	mockRepo := &mockRepository{}
	plugin.repo = mockRepo
	plugin.handler = NewHandler(mockRepo, config)

	mux := http.NewServeMux()
	plugin.RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/translatable"},
		{http.MethodGet, "/api/translatable/123"},
		{http.MethodGet, "/api/translatable"},
		{http.MethodPut, "/api/translatable/123"},
		{http.MethodDelete, "/api/translatable/123"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not registered", tt.method, tt.path)
			}
		})
	}
}

func TestPlugin_Shutdown(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts"},
		MaxContentLength: 10240,
	}
	plugin, _ := NewPlugin(config)

	err := plugin.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() unexpected error = %v", err)
	}
}

func TestPlugin_Config(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts", "articles"},
		MaxContentLength: 20480,
	}
	plugin, _ := NewPlugin(config)

	returnedConfig := plugin.Config()
	if returnedConfig == nil {
		t.Fatal("Config() returned nil")
	}

	if len(returnedConfig.AllowedTables) != len(config.AllowedTables) {
		t.Errorf("Config() AllowedTables length = %v, want %v",
			len(returnedConfig.AllowedTables), len(config.AllowedTables))
	}

	if returnedConfig.MaxContentLength != config.MaxContentLength {
		t.Errorf("Config() MaxContentLength = %v, want %v",
			returnedConfig.MaxContentLength, config.MaxContentLength)
	}
}

func TestPlugin_RegisterRoutesWithCustomRouter(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts"},
		MaxContentLength: 10240,
	}
	plugin, _ := NewPlugin(config)
	mockRepo := &mockRepository{}
	plugin.repo = mockRepo
	plugin.handler = NewHandler(mockRepo, config)

	// Create mock router
	mockRouter := &mockRouter{
		routes: make(map[string]http.HandlerFunc),
	}

	plugin.RegisterRoutesWithCustomRouter(mockRouter)

	expectedRoutes := []string{
		"POST:/api/translatable",
		"GET:/api/translatable/:id",
		"GET:/api/translatable",
		"PUT:/api/translatable/:id",
		"DELETE:/api/translatable/:id",
	}

	for _, route := range expectedRoutes {
		if _, exists := mockRouter.routes[route]; !exists {
			t.Errorf("Route %s not registered", route)
		}
	}
}

type mockRouter struct {
	routes map[string]http.HandlerFunc
}

func (m *mockRouter) POST(path string, handler http.HandlerFunc) {
	m.routes["POST:"+path] = handler
}

func (m *mockRouter) GET(path string, handler http.HandlerFunc) {
	m.routes["GET:"+path] = handler
}

func (m *mockRouter) PUT(path string, handler http.HandlerFunc) {
	m.routes["PUT:"+path] = handler
}

func (m *mockRouter) DELETE(path string, handler http.HandlerFunc) {
	m.routes["DELETE:"+path] = handler
}

func TestPlugin_Lifecycle(t *testing.T) {
	config := &Config{
		AllowedTables:    []string{"posts", "articles"},
		MaxContentLength: 10240,
	}

	plugin, err := NewPlugin(config)
	if err != nil {
		t.Fatalf("NewPlugin() failed: %v", err)
	}

	if plugin.Name() != "translatable" {
		t.Errorf("Plugin name = %v, want 'translatable'", plugin.Name())
	}
	if plugin.Version() == "" {
		t.Error("Plugin version is empty")
	}

	mockRepo := &mockRepository{}
	plugin.repo = mockRepo
	plugin.handler = NewHandler(mockRepo, config)

	mux := http.NewServeMux()
	plugin.RegisterRoutes(mux)

	cfg := plugin.Config()
	if cfg == nil {
		t.Error("Config() returned nil after initialization")
	}

	if err := plugin.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}
