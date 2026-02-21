// Package paths provides API handlers for path configuration management.
package paths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/types"
)

// Handler handles path configuration requests
type Handler struct {
	configManager *config.Manager
}

// NewHandler creates a new paths handler
func NewHandler(configManager *config.Manager) *Handler {
	return &Handler{
		configManager: configManager,
	}
}

// GetLlamaCppPaths returns all configured llama.cpp paths
func (h *Handler) GetLlamaCppPaths(c *gin.Context) {
	cfg := h.configManager.Get()
	paths := cfg.Llamacpp.Paths

	api.Success(c, gin.H{
		"items": paths,
		"count": len(paths),
	})
}

// AddLlamaCppPath adds a new llama.cpp path
func (h *Handler) AddLlamaCppPath(c *gin.Context) {
	var req config.LlamacppPath
	if err := c.ShouldBindJSON(&req); err != nil {
		api.BadRequest(c, "Invalid request body")
		return
	}

	// Validate path
	if req.Path == "" {
		api.BadRequest(c, "Path is required")
		return
	}

	// Normalize and validate path
	normalizedPath, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		api.BadRequest(c, fmt.Sprintf("Invalid path: %v", err))
		return
	}
	req.Path = normalizedPath

	// Load current config
	cfg := h.configManager.Get()

	// Check for duplicate
	for _, p := range cfg.Llamacpp.Paths {
		if p.Path == req.Path {
			api.Error(c, types.ErrConflict, "Path already exists")
			return
		}
	}

	// Add path
	cfg.Llamacpp.Paths = append(cfg.Llamacpp.Paths, req)

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		api.ErrorWithDetails(c, types.ErrInternalError, "Failed to save config", err.Error())
		return
	}

	api.Success(c, gin.H{
		"message": "Llama.cpp path added successfully",
		"added":   req,
		"count":   len(cfg.Llamacpp.Paths),
	})
}

// RemoveLlamaCppPath removes a llama.cpp path
func (h *Handler) RemoveLlamaCppPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		api.BadRequest(c, "Path query parameter is required")
		return
	}

	// Load current config
	cfg := h.configManager.Get()

	// Find and remove
	found := false
	newPaths := make([]config.LlamacppPath, 0, len(cfg.Llamacpp.Paths))
	for _, p := range cfg.Llamacpp.Paths {
		if p.Path != path {
			newPaths = append(newPaths, p)
		} else {
			found = true
		}
	}

	if !found {
		api.NotFound(c, "Path")
		return
	}

	cfg.Llamacpp.Paths = newPaths

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		api.ErrorWithDetails(c, types.ErrInternalError, "Failed to save config", err.Error())
		return
	}

	api.Success(c, gin.H{
		"message": "Llama.cpp path removed successfully",
		"removed": path,
		"count":   len(cfg.Llamacpp.Paths),
	})
}

// TestLlamaCppPath tests if a llama.cpp path is valid
func (h *Handler) TestLlamaCppPath(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		api.BadRequest(c, "Invalid request body")
		return
	}

	// Validate path
	_, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		api.Success(c, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	api.Success(c, gin.H{
		"valid":   true,
		"message": "Path is valid",
	})
}

// GetModelPaths returns all configured model paths
func (h *Handler) GetModelPaths(c *gin.Context) {
	cfg := h.configManager.Get()
	paths := cfg.Model.PathConfigs

	api.Success(c, gin.H{
		"items": paths,
		"count": len(paths),
	})
}

// AddModelPath adds a new model path
func (h *Handler) AddModelPath(c *gin.Context) {
	var req config.ModelPath
	if err := c.ShouldBindJSON(&req); err != nil {
		api.BadRequest(c, "Invalid request body")
		return
	}

	// Validate path
	if req.Path == "" {
		api.BadRequest(c, "Path is required")
		return
	}

	// Normalize and validate path
	normalizedPath, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		api.BadRequest(c, fmt.Sprintf("Invalid path: %v", err))
		return
	}
	req.Path = normalizedPath

	// Load current config
	cfg := h.configManager.Get()

	// Check for duplicate
	for _, p := range cfg.Model.PathConfigs {
		if p.Path == req.Path {
			api.Error(c, types.ErrConflict, "Path already exists")
			return
		}
	}

	// Add path
	cfg.Model.PathConfigs = append(cfg.Model.PathConfigs, req)

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		api.ErrorWithDetails(c, types.ErrInternalError, "Failed to save config", err.Error())
		return
	}

	api.Success(c, gin.H{
		"message": "Model path added successfully",
		"added":   req,
		"count":   len(cfg.Model.PathConfigs),
	})
}

// UpdateModelPath updates an existing model path
func (h *Handler) UpdateModelPath(c *gin.Context) {
	var req config.ModelPath
	if err := c.ShouldBindJSON(&req); err != nil {
		api.BadRequest(c, "Invalid request body")
		return
	}

	// Validate path
	if req.Path == "" {
		api.BadRequest(c, "Path is required")
		return
	}

	// Normalize and validate path
	normalizedPath, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		api.BadRequest(c, fmt.Sprintf("Invalid path: %v", err))
		return
	}
	req.Path = normalizedPath

	// Load current config
	cfg := h.configManager.Get()

	// Find and update
	found := false
	for i, p := range cfg.Model.PathConfigs {
		if p.Path == normalizedPath || (req.Name != "" && p.Name == req.Name) {
			cfg.Model.PathConfigs[i] = req
			found = true
			break
		}
	}

	if !found {
		api.NotFound(c, "Path")
		return
	}

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		api.ErrorWithDetails(c, types.ErrInternalError, "Failed to save config", err.Error())
		return
	}

	api.Success(c, gin.H{
		"message": "Model path updated successfully",
		"updated": req,
	})
}

// RemoveModelPath removes a model path
func (h *Handler) RemoveModelPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		api.BadRequest(c, "Path query parameter is required")
		return
	}

	// Load current config
	cfg := h.configManager.Get()

	// Find and remove
	found := false
	newPaths := make([]config.ModelPath, 0, len(cfg.Model.PathConfigs))
	for _, p := range cfg.Model.PathConfigs {
		if p.Path != path {
			newPaths = append(newPaths, p)
		} else {
			found = true
		}
	}

	if !found {
		api.NotFound(c, "Path")
		return
	}

	cfg.Model.PathConfigs = newPaths

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		api.ErrorWithDetails(c, types.ErrInternalError, "Failed to save config", err.Error())
		return
	}

	api.Success(c, gin.H{
		"message": "Model path removed successfully",
		"removed": path,
		"count":   len(cfg.Model.PathConfigs),
	})
}

// validateAndNormalizePath validates and normalizes a directory path
func (h *Handler) validateAndNormalizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", absPath)
		}
		return "", fmt.Errorf("failed to access path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Check for symlinks (security consideration)
	if info.Mode()&os.ModeSymlink != 0 {
		// Resolve symlink
		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve symlink: %w", err)
		}
		absPath = realPath

		// Check if real path is a directory
		info, err = os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to access resolved path: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("resolved path is not a directory: %s", absPath)
		}
	}

	// Clean the path
	return filepath.Clean(absPath), nil
}
