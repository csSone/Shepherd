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

	// Load saved models
	m.loadModels()

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

	// Scan each configured path
	for _, scanPath := range m.config.Model.Paths {
		pathModels, pathErrors := m.scanPath(ctx, scanPath)
		result.Models = append(result.Models, pathModels...)
		result.Errors = append(result.Errors, pathErrors...)
		result.TotalFiles += len(pathModels) + len(pathErrors)
		result.MatchedFiles += len(pathModels)
	}

	result.Duration = time.Since(result.ScannedAt)

	// Update models map
	m.mu.Lock()
	for _, model := range result.Models {
		m.models[model.ID] = model
	}
	m.mu.Unlock()

	// Save to config
	m.saveModels()

	return result, nil
}

// scanPath scans a single path for models
func (m *Manager) scanPath(ctx context.Context, scanPath string) ([]*Model, []ScanError) {
	var models []*Model
	var errors []ScanError

	// Update scan status
	m.mu.Lock()
	m.scanStatus.CurrentPath = scanPath
	m.mu.Unlock()

	// Check if path exists
	info, err := os.Stat(scanPath)
	if err != nil {
		return nil, []ScanError{{Path: scanPath, Error: err.Error()}}
	}

	// Walk directory
	if info.IsDir() {
		err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
			// Check context
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err != nil {
				errors = append(errors, ScanError{Path: path, Error: err.Error()})
				return nil
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check if file is a GGUF model
			if m.isGGUFFile(path) {
				model, err := m.loadModel(path)
				if err != nil {
					errors = append(errors, ScanError{Path: path, Error: err.Error()})
				} else {
					models = append(models, model)
				}
			}

			return nil
		})

		if err != nil {
			errors = append(errors, ScanError{Path: scanPath, Error: err.Error()})
		}
	} else if m.isGGUFFile(scanPath) {
		// Single file
		model, err := m.loadModel(scanPath)
		if err != nil {
			errors = append(errors, ScanError{Path: scanPath, Error: err.Error()})
		} else {
			models = append(models, model)
		}
	}

	return models, errors
}

// isGGUFFile checks if a file is a GGUF model file
func (m *Manager) isGGUFFile(path string) bool {
	// Check file extension
	base := filepath.Base(path)

	// Common GGUF patterns
	patterns := []string{
		".gguf",
		".GGUF",
		"gguf-",
	}

	for _, pattern := range patterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}

	// Check for models--*.gguf pattern (HuggingFace cache)
	if matched, _ := regexp.MatchString(`models--.+--.+\.gguf$`, path); matched {
		return true
	}

	return false
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

// calculatePathPrefix calculates a short path prefix for display
func (m *Manager) calculatePathPrefix(path string) string {
	// Get directory of the model file
	dir := filepath.Dir(path)

	// Check against configured scan paths
	for _, scanPath := range m.config.Model.Paths {
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
		return
	}

	configModels, err := m.configMgr.LoadModelsConfig()
	if err != nil {
		return
	}

	// Load aliases and favourites
	aliases, _ := m.configMgr.LoadAliasMap()
	favourites, _ := m.configMgr.LoadFavouriteMap()

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
			}
		}
	}
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

// Close closes the manager
func (m *Manager) Close() error {
	m.cancel()
	m.wg.Wait()
	return nil
}
