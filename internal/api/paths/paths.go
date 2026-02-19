// Package paths provides API handlers for path configuration management.
package paths

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": paths,
			"count": len(paths),
		},
	})
}

// AddLlamaCppPath adds a new llama.cpp path
func (h *Handler) AddLlamaCppPath(c *gin.Context) {
	var req config.LlamacppPath
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate path
	if req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Path is required",
		})
		return
	}

	// Normalize and validate path
	normalizedPath, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid path: %v", err),
		})
		return
	}
	req.Path = normalizedPath

	// Load current config
	cfg := h.configManager.Get()

	// Check for duplicate
	for _, p := range cfg.Llamacpp.Paths {
		if p.Path == req.Path {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Path already exists",
			})
			return
		}
	}

	// Add path
	cfg.Llamacpp.Paths = append(cfg.Llamacpp.Paths, req)

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Llama.cpp path added successfully",
			"added":   req,
			"count":   len(cfg.Llamacpp.Paths),
		},
	})
}

// RemoveLlamaCppPath removes a llama.cpp path
func (h *Handler) RemoveLlamaCppPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Path query parameter is required",
		})
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
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Path not found",
		})
		return
	}

	cfg.Llamacpp.Paths = newPaths

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Llama.cpp path removed successfully",
			"removed": path,
			"count":   len(cfg.Llamacpp.Paths),
		},
	})
}

// TestLlamaCppPath tests if a llama.cpp path is valid
func (h *Handler) TestLlamaCppPath(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate path
	_, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"valid":   false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   true,
		"message": "Path is valid",
	})
}

// GetModelPaths returns all configured model paths
func (h *Handler) GetModelPaths(c *gin.Context) {
	cfg := h.configManager.Get()
	paths := cfg.Model.PathConfigs

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": paths,
			"count": len(paths),
		},
	})
}

// AddModelPath adds a new model path
func (h *Handler) AddModelPath(c *gin.Context) {
	var req config.ModelPath
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate path
	if req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Path is required",
		})
		return
	}

	// Normalize and validate path
	normalizedPath, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid path: %v", err),
		})
		return
	}
	req.Path = normalizedPath

	// Load current config
	cfg := h.configManager.Get()

	// Check for duplicate
	for _, p := range cfg.Model.PathConfigs {
		if p.Path == req.Path {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Path already exists",
			})
			return
		}
	}

	// Add path
	cfg.Model.PathConfigs = append(cfg.Model.PathConfigs, req)

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Model path added successfully",
			"added":   req,
			"count":   len(cfg.Model.PathConfigs),
		},
	})
}

// UpdateModelPath updates an existing model path
func (h *Handler) UpdateModelPath(c *gin.Context) {
	var req config.ModelPath
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Validate path
	if req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Path is required",
		})
		return
	}

	// Normalize and validate path
	normalizedPath, err := h.validateAndNormalizePath(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid path: %v", err),
		})
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
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Path not found",
		})
		return
	}

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Model path updated successfully",
			"updated": req,
		},
	})
}

// RemoveModelPath removes a model path
func (h *Handler) RemoveModelPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Path query parameter is required",
		})
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
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Path not found",
		})
		return
	}

	cfg.Model.PathConfigs = newPaths

	// Save config
	if err := h.configManager.Save(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to save config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Model path removed successfully",
			"removed": path,
			"count":   len(cfg.Model.PathConfigs),
		},
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
