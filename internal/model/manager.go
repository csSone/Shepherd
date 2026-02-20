package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/gguf"
	"github.com/shepherd-project/shepherd/Shepherd/internal/process"
)

// Manager manages model scanning and loading
type Manager struct {
	config     *config.Config
	configMgr  *config.Manager
	processMgr *process.Manager

	models     map[string]*Model
	statuses   map[string]*ModelStatus
	scanStatus *ScanStatus

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ScanStatus represents the current scan status
type ScanStatus struct {
	Scanning    bool
	Progress    float64
	CurrentPath string
	StartedAt   time.Time
	Errors      []ScanError
}

// NewManager creates a new model manager
func NewManager(cfg *config.Config, cfgMgr *config.Manager, procMgr *process.Manager) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		config:     cfg,
		configMgr:  cfgMgr,
		processMgr: procMgr,
		models:     make(map[string]*Model),
		statuses:   make(map[string]*ModelStatus),
		scanStatus: &ScanStatus{},
		ctx:        ctx,
		cancel:     cancel,
	}

	// Log initialization info
	paths := m.getScanPaths()
	if len(paths) == 0 {
		fmt.Printf("[WARN] ModelManager: 未配置模型扫描路径\n")
	} else {
		fmt.Printf("[INFO] ModelManager: 初始化完成，配置路径: %v\n", paths)
	}

	// Load saved models
	m.loadModels()
	fmt.Printf("[INFO] ModelManager: 从配置加载了 %d 个模型\n", len(m.models))

	return m
}

// Scan scans for models in configured paths
func (m *Manager) Scan(ctx context.Context) (*ScanResult, error) {
	m.mu.Lock()
	if m.scanStatus.Scanning {
		m.mu.Unlock()
		return nil, fmt.Errorf("scan already in progress")
	}
	m.scanStatus.Scanning = true
	m.scanStatus.StartedAt = time.Now()
	m.scanStatus.Errors = nil
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.scanStatus.Scanning = false
		m.mu.Unlock()
	}()

	result := &ScanResult{
		Models:    []*Model{},
		Errors:    []ScanError{},
		ScannedAt: time.Now(),
	}

	scanPaths := m.getScanPaths()
	fmt.Printf("[INFO] ModelManager: 开始扫描 %d 个路径\n", len(scanPaths))

	// Scan each configured path
	for _, scanPath := range scanPaths {
		fmt.Printf("[INFO] ModelManager: 正在扫描路径: %s\n", scanPath)
		pathModels, pathErrors := m.scanPath(ctx, scanPath)
		fmt.Printf("[INFO] ModelManager: 路径 %s 扫描完成: 找到 %d 个模型, %d 个错误\n",
			scanPath, len(pathModels), len(pathErrors))
		result.Models = append(result.Models, pathModels...)
		result.Errors = append(result.Errors, pathErrors...)
		result.TotalFiles += len(pathModels) + len(pathErrors)
		result.MatchedFiles += len(pathModels)
	}

	result.Duration = time.Since(result.ScannedAt)
	fmt.Printf("[INFO] ModelManager: 扫描完成，总共找到 %d 个模型，耗时 %v\n",
		len(result.Models), result.Duration)

	// Update models map
	m.mu.Lock()
	for _, model := range result.Models {
		m.models[model.ID] = model
	}
	modelCount := len(m.models)
	m.mu.Unlock()
	fmt.Printf("[INFO] ModelManager: 模型缓存已更新，当前共 %d 个模型\n", modelCount)

	// Save to config
	m.saveModels()
	fmt.Printf("[INFO] ModelManager: 已保存 %d 个模型到配置\n", len(result.Models))

	return result, nil
}

// scanPath scans a single path for models with enhanced robustness
func (m *Manager) scanPath(ctx context.Context, scanPath string) ([]*Model, []ScanError) {
	var models []*Model
	var errors []ScanError
	var mu sync.Mutex
	var fileCount int
	var matchedCount int

	// Update scan status
	m.mu.Lock()
	m.scanStatus.CurrentPath = scanPath
	m.mu.Unlock()

	// Check if path exists
	info, err := os.Stat(scanPath)
	if err != nil {
		fmt.Printf("[ERROR] ModelManager: 路径访问失败: %s - %v\n", scanPath, err)
		return nil, []ScanError{{Path: scanPath, Error: fmt.Sprintf("路径访问失败: %v", err)}}
	}

	fmt.Printf("[DEBUG] ModelManager: 开始扫描路径: %s (类型: %s)\n", scanPath, func() string {
		if info.IsDir() {
			return "目录"
		}
		return "文件"
	}())

	// Check if path is readable
	if info.IsDir() {
		// Test read permission by trying to open the directory
		f, err := os.Open(scanPath)
		if err != nil {
			return nil, []ScanError{{Path: scanPath, Error: fmt.Sprintf("目录读取失败: %v", err)}}
		}
		f.Close()
	}

	// Use concurrent processing for directories
	if info.IsDir() {
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 10)

		err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err != nil {
				mu.Lock()
				errors = append(errors, ScanError{
					Path:  path,
					Error: fmt.Sprintf("文件访问错误: %v", err),
				})
				mu.Unlock()
				return nil
			}

			if info.IsDir() {
				return nil
			}

			fileCount++

			// 检查是否为模型文件
			if m.isModelFile(path) {
				matchedCount++
				fmt.Printf("[INFO] ModelManager: 找到模型文件: %s\n", path)

				wg.Add(1)
				semaphore <- struct{}{}

				go func(filePath string) {
					defer wg.Done()
					defer func() { <-semaphore }()

					model, err := m.loadModelWithValidation(filePath)
					if err != nil {
						fmt.Printf("[WARN] ModelManager: 加载模型失败: %s - %v\n", filePath, err)
						mu.Lock()
						errors = append(errors, ScanError{
							Path:  filePath,
							Error: err.Error(),
						})
						mu.Unlock()
					} else {
						fmt.Printf("[INFO] ModelManager: 成功加载模型: %s (ID: %s)\n", model.Name, model.ID)
						mu.Lock()
						models = append(models, model)
						mu.Unlock()
					}
				}(path)
			}

			return nil
		})

		wg.Wait()

		fmt.Printf("[DEBUG] ModelManager: 路径 %s 扫描完成: 共检查 %d 个文件，匹配 %d 个模型文件\n",
			scanPath, fileCount, matchedCount)

		if err != nil && err != ctx.Err() {
			errors = append(errors, ScanError{
				Path:  scanPath,
				Error: fmt.Sprintf("扫描中断: %v", err),
			})
		}
	} else if m.isModelFile(scanPath) {
		fmt.Printf("[INFO] ModelManager: 单文件模型: %s\n", scanPath)
		model, err := m.loadModelWithValidation(scanPath)
		if err != nil {
			errors = append(errors, ScanError{
				Path:  scanPath,
				Error: err.Error(),
			})
		} else {
			models = append(models, model)
		}
	} else {
		fmt.Printf("[WARN] ModelManager: 路径不是模型文件: %s\n", scanPath)
	}

	return models, errors
}

// isModelFile checks if a file is a supported model file (GGUF, SafeTensors, etc.)
func (m *Manager) isModelFile(path string) bool {
	// Check file extension
	base := filepath.Base(path)

	// 支持的模型格式
	// GGUF 格式 (主要支持，可被 llama.cpp 加载)
	patterns := []string{
		".gguf", // GGUF 格式
		".GGUF",
		"gguf-", // 分卷 GGUF
	}

	// 检查文件扩展名
	for _, pattern := range patterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}

	// HuggingFace 缓存目录模式检查
	// 模式1: models--org--model/snapshots/hash/*.gguf
	if matched, _ := regexp.MatchString(`models--.+--.+\.gguf$`, path); matched {
		return true
	}

	// 模式2: HuggingFace 缓存中的 snapshots 目录
	if strings.Contains(path, "snapshots") {
		// 检查是否包含模型文件扩展名
		if strings.HasSuffix(base, ".safetensors") ||
			strings.HasSuffix(base, ".bin") ||
			strings.HasSuffix(base, ".safetensors") {
			fmt.Printf("[DEBUG] 找到 HuggingFace 格式模型: %s\n", path)
			return true
		}
	}

	return false
}

// isGGUFFile checks if a file is specifically a GGUF model file (deprecated, use isModelFile)
// 保留此方法以兼容性，内部调用 isModelFile
func (m *Manager) isGGUFFile(path string) bool {
	return m.isModelFile(path)
}

// loadModel loads a model from a file path
func (m *Manager) loadModel(path string) (*Model, error) {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Read GGUF metadata
	metadata, err := gguf.ReadMetadata(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// Generate model ID
	modelID := m.generateModelID(path, metadata)

	// Calculate path prefix for duplicate identification
	pathPrefix := m.calculatePathPrefix(path)

	// Get model name
	modelName := metadata.Name
	if modelName == "" {
		modelName = filepath.Base(path)
		modelName = strings.TrimSuffix(modelName, ".gguf")
		modelName = strings.TrimSuffix(modelName, ".GGUF")
	}

	// Create display name with path prefix for duplicates
	displayName := modelName
	if pathPrefix != "" && pathPrefix != "models" {
		displayName = fmt.Sprintf("%s [%s]", modelName, pathPrefix)
	}

	// Create model
	model := &Model{
		ID:          modelID,
		Name:        modelName,
		DisplayName: displayName,
		Path:        path,
		PathPrefix:  pathPrefix,
		Size:        info.Size(),
		Metadata:    metadata,
		ScannedAt:   time.Now(),
		SourcePath:  filepath.Dir(path),
	}

	// Check for mmproj
	mmprojPath := m.findMmproj(path)
	if mmprojPath != "" {
		mmprojMeta, err := gguf.ReadMetadata(mmprojPath)
		if err == nil {
			model.MmprojPath = mmprojPath
			model.MmprojMeta = mmprojMeta
		}
	}

	return model, nil
}

// loadModelWithValidation loads a model with additional validation
func (m *Manager) loadModelWithValidation(path string) (*Model, error) {
	// Validate file exists and is readable
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("无法访问模型文件: %w", err)
	}

	// Check file size (must be at least 1KB to be valid)
	if info.Size() < 1024 {
		return nil, fmt.Errorf("模型文件太小 (%d bytes), 可能已损坏", info.Size())
	}

	// Check file is readable
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("无法读取模型文件: %w", err)
	}
	f.Close()

	// Load model
	model, err := m.loadModel(path)
	if err != nil {
		return nil, err
	}

	// Validate metadata
	if model.Metadata == nil {
		return nil, fmt.Errorf("无法读取模型元数据")
	}

	return model, nil
}

// generateModelID generates a unique model ID using path hash
func (m *Manager) generateModelID(path string, metadata *gguf.Metadata) string {
	// Use hash of full path for uniqueness
	hash := sha256.Sum256([]byte(path))
	hashStr := hex.EncodeToString(hash[:8])

	// Get base name without extension
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".gguf")
	base = strings.TrimSuffix(base, ".GGUF")

	return fmt.Sprintf("%s-%s", base, hashStr)
}

// findMmproj looks for a multimodal projector file
func (m *Manager) findMmproj(modelPath string) string {
	dir := filepath.Dir(modelPath)
	base := filepath.Base(modelPath)

	// Remove .gguf extension
	base = strings.TrimSuffix(base, ".gguf")
	base = strings.TrimSuffix(base, ".GGUF")

	// Common mmproj patterns
	patterns := []string{
		filepath.Join(dir, base+"-mmproj.gguf"),
		filepath.Join(dir, base+"-mmproj-f16.gguf"),
		filepath.Join(dir, "mmproj.gguf"),
		filepath.Join(dir, "mmproj-model.gguf"),
	}

	for _, pattern := range patterns {
		if _, err := os.Stat(pattern); err == nil {
			return pattern
		}
	}

	return ""
}

// getScanPaths returns the list of scan paths (from PathConfigs or Paths)
func (m *Manager) getScanPaths() []string {
	if len(m.config.Model.PathConfigs) > 0 {
		paths := make([]string, 0, len(m.config.Model.PathConfigs))
		for _, pc := range m.config.Model.PathConfigs {
			paths = append(paths, pc.Path)
		}
		fmt.Printf("[DEBUG] getScanPaths: 从 PathConfigs 返回 %d 个路径: %v\n", len(paths), paths)
		return paths
	}
	fmt.Printf("[DEBUG] getScanPaths: 从 Paths 返回 %d 个路径: %v\n", len(m.config.Model.Paths), m.config.Model.Paths)
	return m.config.Model.Paths
}

// calculatePathPrefix calculates a short path prefix for display
func (m *Manager) calculatePathPrefix(path string) string {
	// Get directory of the model file
	dir := filepath.Dir(path)

	// Check against configured scan paths
	for _, scanPath := range m.getScanPaths() {
		// Clean paths for comparison
		cleanScanPath := filepath.Clean(scanPath)
		cleanDir := filepath.Clean(dir)

		// Check if this path is under the scan path
		if strings.HasPrefix(cleanDir, cleanScanPath) {
			// Get relative path from scan root
			rel, err := filepath.Rel(cleanScanPath, cleanDir)
			if err != nil {
				continue
			}

			// Get scan path base name as root
			scanBase := filepath.Base(cleanScanPath)
			if scanBase == "." || scanBase == "/" {
				scanBase = "models"
			}

			// If relative path is "." (same dir), just return scan base
			if rel == "." {
				return scanBase
			}

			// Return scan base + first level subdir
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) > 0 && parts[0] != "" {
				return filepath.Join(scanBase, parts[0])
			}
			return scanBase
		}
	}

	// Fallback: use parent directory name
	return filepath.Base(dir)
}

// GetModel returns a model by ID
func (m *Manager) GetModel(id string) (*Model, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	model, exists := m.models[id]
	if !exists {
		return nil, false
	}

	// Return a copy
	modelCopy := *model
	return &modelCopy, true
}

// ListModels returns all models
func (m *Manager) ListModels() []*Model {
	m.mu.RLock()
	modelCount := len(m.models)
	m.mu.RUnlock()

	// 如果内存中没有模型，自动触发一次扫描
	if modelCount == 0 {
		fmt.Printf("[INFO] ListModels: 内存中没有模型，触发自动扫描\n")
		// 使用 background context 进行自动扫描
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// 在 goroutine 中执行扫描，避免阻塞当前调用
		done := make(chan bool, 1)
		go func() {
			if _, err := m.Scan(ctx); err != nil {
				fmt.Printf("[WARN] ListModels: 自动扫描失败: %v\n", err)
			}
			done <- true
		}()

		// 等待扫描完成，最多等待 10 秒
		select {
		case <-done:
			fmt.Printf("[INFO] ListModels: 自动扫描完成\n")
		case <-time.After(10 * time.Second):
			fmt.Printf("[WARN] ListModels: 自动扫描超时，返回当前模型列表\n")
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	models := make([]*Model, 0, len(m.models))
	for _, model := range m.models {
		modelCopy := *model
		models = append(models, &modelCopy)
	}

	return models
}

// Load loads a model
func (m *Manager) Load(req *LoadRequest) (*LoadResult, error) {
	// Get model
	model, exists := m.GetModel(req.ModelID)
	if !exists {
		return nil, fmt.Errorf("model not found: %s", req.ModelID)
	}

	// Check if already loading
	m.mu.Lock()
	if status, exists := m.statuses[req.ModelID]; exists && status.State == StateLoading {
		m.mu.Unlock()
		return nil, fmt.Errorf("model already loading: %s", req.ModelID)
	}

	// Create status
	status := &ModelStatus{
		ID:    req.ModelID,
		Name:  model.Name,
		State: StateLoading,
	}
	m.statuses[req.ModelID] = status
	m.mu.Unlock()

	startTime := time.Now()

	// Find llama.cpp binary
	binPath := m.findLlamaCppBinary()
	if binPath == "" {
		m.mu.Lock()
		status.State = StateError
		status.Error = fmt.Errorf("llama.cpp binary not found")
		m.mu.Unlock()
		return &LoadResult{
			Success: false,
			ModelID: req.ModelID,
			Error:   status.Error,
		}, status.Error
	}

	// Build command
	opts := map[string]interface{}{
		"ctx_size":       req.CtxSize,
		"batch_size":     req.BatchSize,
		"threads":        req.Threads,
		"gpu_layers":     req.GPULayers,
		"temperature":    req.Temperature,
		"top_p":          req.TopP,
		"top_k":          req.TopK,
		"repeat_penalty": req.RepeatPenalty,
		"n_predict":      req.NPredict,
	}

	// Find available port
	port := m.findAvailablePort()

	cmd, err := process.BuildCommand(binPath, model.Path, port, opts)
	if err != nil {
		m.mu.Lock()
		status.State = StateError
		status.Error = err
		m.mu.Unlock()
		return &LoadResult{
			Success: false,
			ModelID: req.ModelID,
			Error:   err,
		}, err
	}

	// Start process
	proc, err := m.processMgr.Start(req.ModelID, model.Name, cmd, binPath)
	if err != nil {
		m.mu.Lock()
		status.State = StateError
		status.Error = err
		m.mu.Unlock()
		return &LoadResult{
			Success: false,
			ModelID: req.ModelID,
			Error:   err,
		}, err
	}

	// Update status
	m.mu.Lock()
	status.State = StateLoaded
	status.ProcessID = proc.ID
	status.Port = port
	status.LoadedAt = time.Now()
	m.mu.Unlock()

	duration := time.Since(startTime)

	return &LoadResult{
		Success:  true,
		ModelID:  req.ModelID,
		Port:     port,
		CtxSize:  req.CtxSize,
		Duration: duration,
	}, nil
}

// Unload unloads a model
func (m *Manager) Unload(modelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.statuses[modelID]
	if !exists {
		return fmt.Errorf("model not loaded: %s", modelID)
	}

	if status.State != StateLoaded {
		return fmt.Errorf("model not in loaded state: %s", modelID)
	}

	// Stop process
	if err := m.processMgr.Stop(modelID); err != nil {
		return err
	}

	// Update status
	status.State = StateUnloaded
	status.ProcessID = ""
	status.Port = 0

	return nil
}

// GetStatus returns the status of a model
func (m *Manager) GetStatus(modelID string) (*ModelStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.statuses[modelID]
	if !exists {
		return nil, false
	}

	// Return a copy
	statusCopy := *status
	return &statusCopy, true
}

// ListStatus returns all model statuses
func (m *Manager) ListStatus() map[string]*ModelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]*ModelStatus, len(m.statuses))
	for k, v := range m.statuses {
		statusCopy := *v
		statuses[k] = &statusCopy
	}

	return statuses
}

// GetScanStatus returns the current scan status
func (m *Manager) GetScanStatus() *ScanStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statusCopy := *m.scanStatus
	return &statusCopy
}

// SetAlias sets the alias for a model
func (m *Manager) SetAlias(modelID, alias string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	model, exists := m.models[modelID]
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	model.Alias = alias

	// Save to config
	if m.configMgr != nil {
		if err := m.configMgr.SaveModelAlias(modelID, alias); err != nil {
			return err
		}
	}

	return nil
}

// SetFavourite sets the favourite flag for a model
func (m *Manager) SetFavourite(modelID string, favourite bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	model, exists := m.models[modelID]
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	model.Favourite = favourite

	// Save to config
	if m.configMgr != nil {
		if err := m.configMgr.SaveModelFavourite(modelID, favourite); err != nil {
			return err
		}
	}

	return nil
}

// loadModels loads models from config
func (m *Manager) loadModels() {
	if m.configMgr == nil {
		fmt.Printf("[WARN] loadModels: configMgr is nil, skipping load\n")
		return
	}

	configModels, err := m.configMgr.LoadModelsConfig()
	if err != nil {
		fmt.Printf("[ERROR] loadModels: failed to load models config: %v\n", err)
		return
	}

	fmt.Printf("[INFO] loadModels: loaded %d models from config\n", len(configModels))

	// Load aliases and favourites
	aliases, _ := m.configMgr.LoadAliasMap()
	favourites, _ := m.configMgr.LoadFavouriteMap()

	loadedCount := 0
	for _, cfgModel := range configModels {
		// Try to load the model from disk
		if info, err := os.Stat(cfgModel.Path); err == nil && !info.IsDir() {
			model, err := m.loadModel(cfgModel.Path)
			if err == nil {
				model.ID = cfgModel.ModelID
				if alias, ok := aliases[model.ID]; ok {
					model.Alias = alias
				}
				if fav, ok := favourites[model.ID]; ok {
					model.Favourite = fav
				}
				m.models[model.ID] = model
				loadedCount++
			} else {
				fmt.Printf("[WARN] loadModels: failed to load model %s: %v\n", cfgModel.Path, err)
			}
		} else {
			fmt.Printf("[WARN] loadModels: model file not found: %s\n", cfgModel.Path)
		}
	}
	fmt.Printf("[INFO] loadModels: successfully loaded %d models into cache\n", loadedCount)
}

// saveModels saves models to config
func (m *Manager) saveModels() {
	if m.configMgr == nil {
		return
	}

	// Convert models to config entries
	var configModels []config.ModelConfigEntry
	for _, model := range m.models {
		entry := config.ModelConfigEntry{
			ModelID:   model.ID,
			Path:      model.Path,
			Size:      model.Size,
			Alias:     model.Alias,
			Favourite: model.Favourite,
		}

		// Add primary model info if available
		if model.Metadata != nil {
			entry.PrimaryModel = &config.PrimaryModelInfo{
				FileName:        filepath.Base(model.Path),
				Name:            model.Metadata.Name,
				Architecture:    model.Metadata.Architecture,
				ContextLength:   model.Metadata.ContextLength,
				EmbeddingLength: model.Metadata.EmbeddingLength,
			}
		}

		// Add mmproj info if available
		if model.MmprojMeta != nil {
			entry.Mmproj = &config.MmprojInfo{
				FileName:     filepath.Base(model.MmprojPath),
				Name:         model.MmprojMeta.Name,
				Architecture: model.MmprojMeta.Architecture,
			}
		}

		configModels = append(configModels, entry)
	}

	m.configMgr.SaveModelsConfig(configModels)
}

// findLlamaCppBinary finds the llama.cpp binary
func (m *Manager) findLlamaCppBinary() string {
	// Check configured paths
	for _, llamacppPath := range m.config.Llamacpp.Paths {
		binaryPath := filepath.Join(llamacppPath.Path, "llama-server")
		if _, err := os.Stat(binaryPath); err == nil {
			return llamacppPath.Path
		}
	}

	// Check common locations
	commonPaths := []string{
		"/usr/local/bin",
		"/usr/bin",
		"./llama.cpp",
	}

	for _, path := range commonPaths {
		binaryPath := filepath.Join(path, "llama-server")
		if _, err := os.Stat(binaryPath); err == nil {
			return path
		}
	}

	return ""
}

// findAvailablePort finds an available port for the model server
func (m *Manager) findAvailablePort() int {
	// Start from base port and find available
	basePort := 8081

	statuses := m.ListStatus()
	usedPorts := make(map[int]bool)
	for _, status := range statuses {
		if status.Port > 0 {
			usedPorts[status.Port] = true
		}
	}

	for port := basePort; port < basePort+100; port++ {
		if !usedPorts[port] {
			return port
		}
	}

	return basePort
}

// GetLoadedModelCount returns the number of currently loaded models
func (m *Manager) GetLoadedModelCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, status := range m.statuses {
		if status.State == StateLoaded {
			count++
		}
	}
	return count
}

// ValidationResult represents the result of model validation
type ValidationResult struct {
	Valid         bool
	Errors        []string
	Warnings      []string
	ModelCount    int
	InvalidModels []string
	MissingFiles  []string
}

// ValidateModels validates all models in the cache
func (m *Manager) ValidateModels() *ValidationResult {
	result := &ValidationResult{
		Valid:         true,
		Errors:        []string{},
		Warnings:      []string{},
		InvalidModels: []string{},
		MissingFiles:  []string{},
	}

	m.mu.RLock()
	models := make(map[string]*Model)
	for id, model := range m.models {
		models[id] = model
	}
	m.mu.RUnlock()

	result.ModelCount = len(models)

	for id, model := range models {
		// Check if file exists
		if _, err := os.Stat(model.Path); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("模型 %s 文件不存在: %v", id, err))
			result.MissingFiles = append(result.MissingFiles, model.Path)
			result.InvalidModels = append(result.InvalidModels, id)
			continue
		}

		// Check file size consistency
		if info, err := os.Stat(model.Path); err == nil {
			if info.Size() != model.Size {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("模型 %s 文件大小已变更: 缓存 %d, 实际 %d", id, model.Size, info.Size()))
			}
		}

		// Validate metadata
		if model.Metadata == nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("模型 %s 缺少元数据", id))
			result.InvalidModels = append(result.InvalidModels, id)
		}
	}

	return result
}

// CleanInvalidModels removes invalid models from cache
func (m *Manager) CleanInvalidModels() int {
	result := m.ValidateModels()
	if !result.Valid {
		m.mu.Lock()
		for _, id := range result.InvalidModels {
			delete(m.models, id)
			delete(m.statuses, id)
		}
		m.mu.Unlock()

		// Save updated models
		m.saveModels()
	}

	return len(result.InvalidModels)
}

// SearchModels searches and filters models based on criteria
// 如果内存中没有模型，会自动触发一次扫描
func (m *Manager) SearchModels(filter *ModelFilter, sort *ModelSort) *ModelSearchResult {
	m.mu.RLock()
	modelCount := len(m.models)
	m.mu.RUnlock()

	// 如果内存中没有模型，自动触发一次扫描
	if modelCount == 0 {
		fmt.Printf("[INFO] SearchModels: 内存中没有模型，触发自动扫描\n")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		done := make(chan bool, 1)
		go func() {
			if _, err := m.Scan(ctx); err != nil {
				fmt.Printf("[WARN] SearchModels: 自动扫描失败: %v\n", err)
			}
			done <- true
		}()

		select {
		case <-done:
			fmt.Printf("[INFO] SearchModels: 自动扫描完成\n")
		case <-time.After(10 * time.Second):
			fmt.Printf("[WARN] SearchModels: 自动扫描超时\n")
		}
	}

	m.mu.RLock()
	allModels := make([]*Model, 0, len(m.models))
	for _, model := range m.models {
		modelCopy := *model
		allModels = append(allModels, &modelCopy)
	}
	m.mu.RUnlock()

	result := &ModelSearchResult{
		Models:        []*Model{},
		Total:         len(allModels),
		Tags:          make(map[string]int),
		Architectures: make(map[string]int),
	}

	// Collect statistics
	for _, model := range allModels {
		if model.Metadata != nil && model.Metadata.Architecture != "" {
			result.Architectures[model.Metadata.Architecture]++
		}
		for _, tag := range model.Tags {
			result.Tags[tag]++
		}
	}

	// Apply filters
	filtered := make([]*Model, 0)
	for _, model := range allModels {
		if m.matchesFilter(model, filter) {
			filtered = append(filtered, model)
		}
	}

	// Apply sorting
	if sort != nil {
		m.sortModels(filtered, sort)
	}

	result.Models = filtered
	result.Filtered = len(filtered)

	return result
}

// matchesFilter checks if a model matches the filter criteria
func (m *Manager) matchesFilter(model *Model, filter *ModelFilter) bool {
	if filter == nil {
		return true
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, tag := range filter.Tags {
			for _, modelTag := range model.Tags {
				if strings.EqualFold(tag, modelTag) {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	// Architecture filter
	if filter.Architecture != "" && model.Metadata != nil {
		if !strings.EqualFold(model.Metadata.Architecture, filter.Architecture) {
			return false
		}
	}

	// Min context filter
	if filter.MinContext > 0 && model.Metadata != nil {
		if model.Metadata.ContextLength < filter.MinContext {
			return false
		}
	}

	// Max size filter
	if filter.MaxSize > 0 && model.Size > filter.MaxSize {
		return false
	}

	// Loaded only filter
	if filter.LoadedOnly {
		m.mu.RLock()
		status, exists := m.statuses[model.ID]
		m.mu.RUnlock()
		if !exists || status.State != StateLoaded {
			return false
		}
	}

	// Favourites filter
	if filter.Favourites && !model.Favourite {
		return false
	}

	// Search query
	if filter.SearchQuery != "" {
		query := strings.ToLower(filter.SearchQuery)
		match := false
		if strings.Contains(strings.ToLower(model.Name), query) {
			match = true
		}
		if strings.Contains(strings.ToLower(model.Alias), query) {
			match = true
		}
		if strings.Contains(strings.ToLower(model.Description), query) {
			match = true
		}
		if model.Metadata != nil {
			if strings.Contains(strings.ToLower(model.Metadata.Architecture), query) {
				match = true
			}
		}
		if !match {
			return false
		}
	}

	// Source type filter
	if filter.SourceType != "" && model.SourceType != filter.SourceType {
		return false
	}

	// License filter
	if filter.License != "" && !strings.EqualFold(model.License, filter.License) {
		return false
	}

	return true
}

// sortModels sorts models based on sort criteria
func (m *Manager) sortModels(models []*Model, sort *ModelSort) {
	if sort == nil || sort.Field == "" {
		return
	}

	less := func(i, j int) bool {
		switch sort.Field {
		case "name":
			if sort.Direction == "desc" {
				return models[i].Name > models[j].Name
			}
			return models[i].Name < models[j].Name
		case "size":
			if sort.Direction == "desc" {
				return models[i].Size > models[j].Size
			}
			return models[i].Size < models[j].Size
		case "scanned_at":
			if sort.Direction == "desc" {
				return models[i].ScannedAt.After(models[j].ScannedAt)
			}
			return models[i].ScannedAt.Before(models[j].ScannedAt)
		case "load_count":
			if sort.Direction == "desc" {
				return models[i].LoadCount > models[j].LoadCount
			}
			return models[i].LoadCount < models[j].LoadCount
		default:
			return models[i].Name < models[j].Name
		}
	}

	// Simple bubble sort for demonstration
	for i := 0; i < len(models); i++ {
		for j := i + 1; j < len(models); j++ {
			if !less(i, j) {
				models[i], models[j] = models[j], models[i]
			}
		}
	}
}

// Close closes the manager
func (m *Manager) Close() error {
	m.cancel()
	m.wg.Wait()
	return nil
}

// GetProcessManager returns the process manager
func (m *Manager) GetProcessManager() *process.Manager {
	return m.processMgr
}
