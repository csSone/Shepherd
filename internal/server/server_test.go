package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/model"
	"github.com/shepherd-project/shepherd/Shepherd/internal/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test server with model manager
func createTestServer(t *testing.T) *Server {
	cfg := config.DefaultConfig()
	configMgr := config.NewManager("standalone")
	_, _ = configMgr.Load() // 加载默认配置，忽略错误

	procMgr := process.NewManager()
	modelMgr := model.NewManager(cfg, configMgr, procMgr)

	serverConfig := &Config{
		WebPort:       8080,
		AnthropicPort: 8070,
		OllamaPort:    11434,
		LMStudioPort:  1234,
		Host:          "0.0.0.0",
		ReadTimeout:   60 * time.Second,
		WriteTimeout:  60 * time.Second,
		ServerCfg:     cfg,
		ConfigMgr:     configMgr,
	}

	server, err := NewServer(serverConfig, modelMgr)
	require.NoError(t, err)
	return server
}

func TestNewServer(t *testing.T) {
	server := createTestServer(t)

	assert.NotNil(t, server)
	assert.NotNil(t, server.engine)
	assert.NotNil(t, server.handlers)
	assert.NotNil(t, server.wsMgr)
	assert.NotNil(t, server.modelMgr)
	assert.NotNil(t, server.config)
}

func TestServerHandleServerInfo(t *testing.T) {
	server := createTestServer(t)
	router := server.GetEngine()

	req := httptest.NewRequest("GET", "/api/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "Shepherd", response["name"])
	assert.Equal(t, "running", response["status"])

	ports, ok := response["ports"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(8080), ports["web"])
	assert.Equal(t, float64(8070), ports["anthropic"])
}

func TestServerCORSMiddleware(t *testing.T) {
	server := createTestServer(t)
	router := server.GetEngine()

	req := httptest.NewRequest("OPTIONS", "/api/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestServerRoutes(t *testing.T) {
	server := createTestServer(t)
	router := server.GetEngine()

	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
	}{
		// Server info
		{"Server info", "GET", "/api/info", http.StatusOK},

		// Config routes
		{"Get config", "GET", "/api/config", http.StatusOK},
		{"Update config", "PUT", "/api/config", http.StatusOK},

		// Model routes - 注意：测试中没有预加载模型，所以 test-id 返回错误状态码是正确的
		{"List models", "GET", "/api/models", http.StatusOK},
		{"Get model", "GET", "/api/models/test-id", http.StatusNotFound}, // 模型不存在
		{"Load model", "POST", "/api/models/test-id/load", http.StatusInternalServerError}, // 模型不存在时加载失败
		{"Unload model", "POST", "/api/models/test-id/unload", http.StatusInternalServerError}, // 模型不存在时卸载失败
		{"Set alias", "PUT", "/api/models/test-id/alias", http.StatusBadRequest}, // 缺少请求体
		{"Set favourite", "PUT", "/api/models/test-id/favourite", http.StatusBadRequest}, // 缺少请求体

		// Scan routes
		{"Scan models", "POST", "/api/scan", http.StatusOK},
		{"Scan status", "GET", "/api/scan/status", http.StatusOK},

		// Download routes
		{"List downloads", "GET", "/api/downloads", http.StatusOK},
		{"Create download", "POST", "/api/downloads", http.StatusOK},
		{"Get download", "GET", "/api/downloads/test-id", http.StatusOK},
		{"Pause download", "POST", "/api/downloads/test-id/pause", http.StatusOK},
		{"Resume download", "POST", "/api/downloads/test-id/resume", http.StatusOK},
		{"Delete download", "DELETE", "/api/downloads/test-id", http.StatusOK},

		// Process routes
		{"List processes", "GET", "/api/processes", http.StatusOK},
		{"Get process", "GET", "/api/processes/test-id", http.StatusOK},
		{"Stop process", "POST", "/api/processes/test-id/stop", http.StatusOK},

		// Note: SSE endpoint /api/events is tested separately

		// OpenAI API
		{"OpenAI models", "GET", "/v1/models", http.StatusOK},

		// Ollama API
		{"Ollama tags", "POST", "/api/tags", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestServerSSEEndpoint(t *testing.T) {
	server := createTestServer(t)
	router := server.GetEngine()

	// Create a request with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/api/events", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	// Run in goroutine since SSE will hang until timeout
	done := make(chan bool)
	go func() {
		router.ServeHTTP(w, req)
		done <- true
	}()

	// Wait for request to complete (or timeout)
	select {
	case <-done:
		// Check that we got the proper SSE headers
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/event-stream")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("SSE request did not complete within timeout")
	}
}

func TestServerStartStop(t *testing.T) {
	cfg := config.DefaultConfig()
	configMgr := config.NewManager("standalone")
	_, _ = configMgr.Load()

	procMgr := process.NewManager()
	modelMgr := model.NewManager(cfg, configMgr, procMgr)

	serverConfig := &Config{
		WebPort:   18080, // Use non-standard port for testing
		Host:      "127.0.0.1",
		ServerCfg: cfg,
		ConfigMgr: configMgr,
	}

	server, err := NewServer(serverConfig, modelMgr)
	require.NoError(t, err)

	// Start server
	err = server.Start()
	assert.NoError(t, err)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Try to connect
	resp, err := http.Get("http://127.0.0.1:18080/api/info")
	if err == nil {
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Stop server
	err = server.Stop()
	assert.NoError(t, err)
}

func TestServerMiddleware(t *testing.T) {
	server := createTestServer(t)
	router := server.GetEngine()

	// Test that recovery middleware works
	req := httptest.NewRequest("GET", "/api/invalid-route", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 for invalid routes
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServerConfigDefaults(t *testing.T) {
	server := createTestServer(t)

	assert.NotNil(t, server)
	assert.NotNil(t, server.engine)
}

func BenchmarkServerRequest(b *testing.B) {
	cfg := config.DefaultConfig()
	configMgr := config.NewManager("standalone")
	_, _ = configMgr.Load()

	procMgr := process.NewManager()
	modelMgr := model.NewManager(cfg, configMgr, procMgr)

	serverConfig := &Config{
		WebPort:   8080,
		Host:      "0.0.0.0",
		ServerCfg: cfg,
		ConfigMgr: configMgr,
	}

	server, err := NewServer(serverConfig, modelMgr)
	if err != nil {
		b.Fatal(err)
	}
	router := server.GetEngine()

	req := httptest.NewRequest("GET", "/api/info", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
