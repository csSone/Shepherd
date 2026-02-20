package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/gguf"
	"github.com/shepherd-project/shepherd/Shepherd/internal/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadStateString(t *testing.T) {
	tests := []struct {
		state    LoadState
		expected string
	}{
		{StateUnloaded, "unloaded"},
		{StateLoading, "loading"},
		{StateLoaded, "loaded"},
		{StateUnloading, "unloading"},
		{StateError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestNewManager(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.models)
	assert.NotNil(t, manager.statuses)
	assert.NotNil(t, manager.scanStatus)
}

func TestManagerIsGGUFFile(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"GGUF file", "/path/to/model.gguf", true},
		{"Uppercase GGUF", "/path/to/model.GGUF", true},
		{"gguf- prefix", "/path/to/gguf-model.bin", true},
		{"No GGUF", "/path/to/model.bin", false},
		{"HuggingFace cache", "/cache/models--org--model/snapshots/abc123/model.gguf", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.isGGUFFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManagerGetSetModel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	t.Run("Get non-existent model", func(t *testing.T) {
		_, exists := manager.GetModel("non-existent")
		assert.False(t, exists)
	})

	t.Run("List models initially empty", func(t *testing.T) {
		models := manager.ListModels()
		assert.Empty(t, models)
	})
}

func TestManagerSetAlias(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// Create a temp directory with a mock model
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model.gguf")

	// Create a minimal GGUF file for testing
	err := createMinimalGGUF(modelPath)
	require.NoError(t, err)

	// Load model
	model, err := manager.loadModel(modelPath)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.models[model.ID] = model
	manager.mu.Unlock()

	// Set alias
	err = manager.SetAlias(model.ID, "Test Model")
	assert.NoError(t, err)

	// Verify
	retrieved, _ := manager.GetModel(model.ID)
	assert.Equal(t, "Test Model", retrieved.Alias)
}

func TestManagerSetFavourite(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// Create a temp directory with a mock model
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model.gguf")

	// Create a minimal GGUF file for testing
	err := createMinimalGGUF(modelPath)
	require.NoError(t, err)

	// Load model
	model, err := manager.loadModel(modelPath)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.models[model.ID] = model
	manager.mu.Unlock()

	// Set favourite
	err = manager.SetFavourite(model.ID, true)
	assert.NoError(t, err)

	// Verify
	retrieved, _ := manager.GetModel(model.ID)
	assert.True(t, retrieved.Favourite)
}

func TestManagerGetStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	t.Run("Get non-existent status", func(t *testing.T) {
		_, exists := manager.GetStatus("non-existent")
		assert.False(t, exists)
	})

	t.Run("List status initially empty", func(t *testing.T) {
		statuses := manager.ListStatus()
		assert.Empty(t, statuses)
	})
}

func TestManagerGetScanStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	status := manager.GetScanStatus()
	assert.NotNil(t, status)
	assert.False(t, status.Scanning)
}

func TestFindAvailablePort(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	port := manager.findAvailablePort()
	assert.GreaterOrEqual(t, port, 8081)
	assert.Less(t, port, 8181)
}

func TestFindMmproj(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	t.Run("No mmproj found", func(t *testing.T) {
		mmproj := manager.findMmproj("/path/to/model.gguf")
		assert.Empty(t, mmproj)
	})
}

func TestGenerateModelID(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	metadata := &gguf.Metadata{
		Name: "test-model",
	}

	id := manager.generateModelID("/path/to/model.gguf", metadata)
	// ID 现在是 hash-based 格式: model-{hash}
	assert.Contains(t, id, "model-")
	// 验证格式: base-hash (两部分)
	parts := strings.Split(id, "-")
	assert.Len(t, parts, 2)
	// hash 部分应该是 16 字符 (SHA256 的前 8 字节，hex 编码)
	assert.Len(t, parts[1], 16)
}

func TestScanStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// Initial status
	status := manager.GetScanStatus()
	assert.False(t, status.Scanning)
	assert.Equal(t, 0.0, status.Progress)
}

// createMinimalGGUF creates a minimal valid GGUF file for testing
// File size is at least 2048 bytes to pass model validation (requires >= 1024 bytes)
func createMinimalGGUF(path string) error {
	// GGUF header (24 bytes)
	header := []byte{
		// Magic: "GGUF" (0x47475546)
		0x47, 0x47, 0x55, 0x46,
		// Version: 3 (little endian uint32)
		0x03, 0x00, 0x00, 0x00,
		// Tensor count: 0 (little endian uint64)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Metadata KV count: 0 (little endian uint64)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	// Pad to 2048 bytes to meet minimum file size requirement
	padding := make([]byte, 2048-24)
	data := append(header, padding...)

	return os.WriteFile(path, data, 0644)
}

func TestLoadModel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	t.Run("Load minimal GGUF", func(t *testing.T) {
		tmpDir := t.TempDir()
		modelPath := filepath.Join(tmpDir, "test-model.gguf")

		err := createMinimalGGUF(modelPath)
		require.NoError(t, err)

		model, err := manager.loadModel(modelPath)
		require.NoError(t, err)

		assert.NotNil(t, model)
		assert.NotEmpty(t, model.ID)
		assert.NotEmpty(t, model.Name) // Should default to file name
		assert.Equal(t, modelPath, model.Path)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		_, err := manager.loadModel("/nonexistent/file.gguf")
		assert.Error(t, err)
	})
}

func TestScanPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	t.Run("Scan directory with GGUF files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create some test files
		modelPath1 := filepath.Join(tmpDir, "model1.gguf")
		modelPath2 := filepath.Join(tmpDir, "model2.GGUF")
		nonModelPath := filepath.Join(tmpDir, "readme.txt")

		createMinimalGGUF(modelPath1)
		createMinimalGGUF(modelPath2)
		os.WriteFile(nonModelPath, []byte("test"), 0644)

		// Scan
		models, errors := manager.scanPath(context.Background(), tmpDir)

		assert.Len(t, models, 2) // Should find 2 GGUF files
		assert.Len(t, errors, 0)
	})

	t.Run("Scan non-existent directory", func(t *testing.T) {
		_, errors := manager.scanPath(context.Background(), "/nonexistent/path")

		assert.NotEmpty(t, errors)
	})
}

func TestLoadUnload(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	t.Run("Load non-existent model", func(t *testing.T) {
		req := &LoadRequest{
			ModelID: "non-existent",
			CtxSize: 4096,
		}

		result, err := manager.Load(req)
		assert.Error(t, err)
		// Result may be nil if load fails early, just check error
		_ = result
	})

	t.Run("Unload non-existent model", func(t *testing.T) {
		err := manager.Unload("non-existent")
		assert.Error(t, err)
	})
}

func TestScanConfigDefaults(t *testing.T) {
	config := ScanConfig{
		Paths: []string{"/models"},
	}

	assert.NotEmpty(t, config.Paths)
	assert.False(t, config.Recursive) // Default is false
}

func TestLoadRequestDefaults(t *testing.T) {
	req := &LoadRequest{
		ModelID: "test-model",
		CtxSize: 4096,
	}

	assert.Equal(t, "test-model", req.ModelID)
	assert.Equal(t, 4096, req.CtxSize)
}

func TestLoadResultDefaults(t *testing.T) {
	result := &LoadResult{
		Success: true,
		ModelID: "test-model",
		Port:    8081,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "test-model", result.ModelID)
	assert.Equal(t, 8081, result.Port)
}

func TestModelDefaults(t *testing.T) {
	model := &Model{
		ID:   "test-id",
		Name: "test-model",
		Path: "/path/to/model.gguf",
	}

	assert.Equal(t, "test-id", model.ID)
	assert.Equal(t, "test-model", model.Name)
	assert.Equal(t, "/path/to/model.gguf", model.Path)
	assert.False(t, model.Favourite)
}

func TestScanResultDefaults(t *testing.T) {
	result := &ScanResult{
		Models:    []*Model{},
		Errors:    []ScanError{},
		ScannedAt: time.Now(),
	}

	assert.NotNil(t, result.Models)
	assert.NotNil(t, result.Errors)
}

func TestScanErrorDefaults(t *testing.T) {
	err := ScanError{
		Path:  "/path/to/file",
		Error: "test error",
	}

	assert.Equal(t, "/path/to/file", err.Path)
	assert.Equal(t, "test error", err.Error)
}

func BenchmarkListModels(b *testing.B) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// Add some models
	for i := 0; i < 100; i++ {
		model := &Model{
			ID:   fmt.Sprintf("model-%d", i),
			Name: fmt.Sprintf("Model %d", i),
			Path: fmt.Sprintf("/path/to/model-%d.gguf", i),
		}
		manager.models[model.ID] = model
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ListModels()
	}
}

// TestLoadAsync tests asynchronous model loading
func TestLoadAsync(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// 创建测试模型
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model.gguf")
	err := createMinimalGGUF(modelPath)
	require.NoError(t, err)

	// 加载模型
	model, err := manager.loadModel(modelPath)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.models[model.ID] = model
	manager.mu.Unlock()

	// 测试异步加载
	t.Run("异步加载返回立即响应", func(t *testing.T) {
		req := &LoadRequest{
			ModelID: model.ID,
			CtxSize: 4096,
		}

		result, err := manager.LoadAsync(req)

		// 应该立即返回而不阻塞
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Async)
		assert.True(t, result.Loading)
	})

	// 等待异步加载完成或失败
	time.Sleep(2 * time.Second)

	// 检查状态
	status, exists := manager.GetStatus(model.ID)
	if exists {
		// 状态应该被设置（可能是 loading, loaded 或 error）
		assert.NotEqual(t, StateUnloaded, status.State)
	}
}

// TestLoadAsyncAlreadyLoaded tests loading an already loaded model
func TestLoadAsyncAlreadyLoaded(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// 创建测试模型
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model.gguf")
	err := createMinimalGGUF(modelPath)
	require.NoError(t, err)

	model, err := manager.loadModel(modelPath)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.models[model.ID] = model
	manager.mu.Unlock()

	// 手动设置为已加载状态
	manager.mu.Lock()
	manager.statuses[model.ID] = &ModelStatus{
		ID:       model.ID,
		Name:     model.Name,
		State:    StateLoaded,
		Port:     8081,
		LoadedAt: time.Now(),
	}
	manager.mu.Unlock()

	// 尝试再次加载
	req := &LoadRequest{
		ModelID: model.ID,
		CtxSize: 4096,
	}

	result, err := manager.LoadAsync(req)

	assert.NoError(t, err)
	assert.True(t, result.Async)
	assert.True(t, result.AlreadyLoaded)
	assert.False(t, result.Loading)
}

// TestLoadAsyncAlreadyLoading tests loading while already loading
func TestLoadAsyncAlreadyLoading(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// 创建测试模型
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test-model.gguf")
	err := createMinimalGGUF(modelPath)
	require.NoError(t, err)

	model, err := manager.loadModel(modelPath)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.models[model.ID] = model
	manager.mu.Unlock()

	// 手动设置为加载中状态
	manager.mu.Lock()
	manager.statuses[model.ID] = &ModelStatus{
		ID:    model.ID,
		Name:  model.Name,
		State: StateLoading,
	}
	manager.mu.Unlock()

	// 尝试再次加载
	req := &LoadRequest{
		ModelID: model.ID,
		CtxSize: 4096,
	}

	result, err := manager.LoadAsync(req)

	assert.NoError(t, err)
	assert.True(t, result.Async)
	assert.True(t, result.Loading)
	assert.False(t, result.AlreadyLoaded)
}

// TestIsLoading tests the isLoading helper method
func TestIsLoading(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	modelID := "test-model"

	// 初始状态：未加载
	assert.False(t, manager.isLoading(modelID))

	// 设置为加载中
	manager.mu.Lock()
	manager.statuses[modelID] = &ModelStatus{
		ID:    modelID,
		State: StateLoading,
	}
	manager.mu.Unlock()

	assert.True(t, manager.isLoading(modelID))

	// 设置为已加载
	manager.mu.Lock()
	manager.statuses[modelID].State = StateLoaded
	manager.mu.Unlock()

	assert.False(t, manager.isLoading(modelID))
}

// TestLoadAsyncNonExistentModel tests loading a non-existent model
func TestLoadAsyncNonExistentModel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	req := &LoadRequest{
		ModelID: "non-existent-model",
		CtxSize: 4096,
	}

	result, err := manager.LoadAsync(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "model not found")
}

// TestLoadResultAsyncFields tests LoadResult async fields
func TestLoadResultAsyncFields(t *testing.T) {
	result := &LoadResult{
		Success:       true,
		ModelID:       "test-model",
		Port:          8081,
		CtxSize:       4096,
		Async:         true,
		Loading:       true,
		AlreadyLoaded: false,
	}

	assert.True(t, result.Async)
	assert.True(t, result.Loading)
	assert.False(t, result.AlreadyLoaded)
}

// BenchmarkLoadAsync benchmarks asynchronous loading
func BenchmarkLoadAsync(b *testing.B) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// 创建测试模型
	modelPath := "/tmp/bench-model.gguf"
	_ = createMinimalGGUF(modelPath)
	defer os.Remove(modelPath)

	model, _ := manager.loadModel(modelPath)
	manager.mu.Lock()
	manager.models[model.ID] = model
	manager.mu.Unlock()

	req := &LoadRequest{
		ModelID: model.ID,
		CtxSize: 4096,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.LoadAsync(req)
		// 清理状态以便下次测试
		manager.mu.Lock()
		delete(manager.statuses, model.ID)
		manager.mu.Unlock()
	}
}

// TestModelStatusTransitions tests model status state transitions
func TestModelStatusTransitions(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	modelID := "test-model"

	// 初始状态：unloaded
	status, exists := manager.GetStatus(modelID)
	assert.False(t, exists)

	// 设置为 loading
	manager.mu.Lock()
	manager.statuses[modelID] = &ModelStatus{
		ID:    modelID,
		State: StateLoading,
	}
	manager.mu.Unlock()

	status, exists = manager.GetStatus(modelID)
	assert.True(t, exists)
	assert.Equal(t, StateLoading, status.State)

	// 转换到 loaded
	manager.mu.Lock()
	manager.statuses[modelID].State = StateLoaded
	manager.statuses[modelID].Port = 8081
	manager.statuses[modelID].LoadedAt = time.Now()
	manager.mu.Unlock()

	status, _ = manager.GetStatus(modelID)
	assert.Equal(t, StateLoaded, status.State)
	assert.Equal(t, 8081, status.Port)

	// 转换到 error
	manager.mu.Lock()
	manager.statuses[modelID].State = StateError
	manager.statuses[modelID].Error = fmt.Errorf("test error")
	manager.mu.Unlock()

	status, _ = manager.GetStatus(modelID)
	assert.Equal(t, StateError, status.State)
	assert.NotNil(t, status.Error)
}

// TestListStatus tests listing all model statuses
func TestListStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	cfgMgr := config.NewManager("standalone")
	procMgr := process.NewManager()

	manager := NewManager(cfg, cfgMgr, procMgr)

	// 添加多个状态
	modelIDs := []string{"model-1", "model-2", "model-3"}

	states := []LoadState{StateLoaded, StateLoading, StateError}

	for i, modelID := range modelIDs {
		manager.mu.Lock()
		manager.statuses[modelID] = &ModelStatus{
			ID:    modelID,
			State: states[i],
		}
		manager.mu.Unlock()
	}

	// 列出所有状态
	allStatuses := manager.ListStatus()

	assert.Len(t, allStatuses, 3)

	for _, modelID := range modelIDs {
		status, exists := allStatuses[modelID]
		assert.True(t, exists)
		assert.NotEmpty(t, status.ID)
	}
}
