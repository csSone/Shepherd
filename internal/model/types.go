// Package model provides model scanning and management functionality.
// It handles discovering GGUF models, reading their metadata, and managing model loading.
package model

import (
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/gguf"
)

// Model represents a discovered GGUF model
type Model struct {
	// Basic information
	ID          string // Unique model ID (path hash)
	Name        string // Model name
	DisplayName string // Display name for UI (handles duplicates)
	Alias       string // Display alias
	Path        string // File path to the GGUF file
	PathPrefix  string // Path prefix for duplicate identification (e.g., "models/A", "cache/B")
	Size        int64  // File size in bytes
	Favourite   bool   // User's favorite flag

	// GGUF metadata
	Metadata *gguf.Metadata

	// Additional files (e.g., mmproj)
	MmprojPath string
	MmprojMeta *gguf.Metadata

	// Scanning info
	ScannedAt  time.Time
	SourcePath string // Original scan path
}

// ModelStatus represents the loading status of a model
type ModelStatus struct {
	ID        string
	Name      string
	State     LoadState
	ProcessID string
	Port      int
	CtxSize   int
	LoadedAt  time.Time
	Error     error
}

// LoadState represents the loading state
type LoadState int

const (
	StateUnloaded LoadState = iota
	StateLoading
	StateLoaded
	StateUnloading
	StateError
)

func (s LoadState) String() string {
	switch s {
	case StateUnloaded:
		return "unloaded"
	case StateLoading:
		return "loading"
	case StateLoaded:
		return "loaded"
	case StateUnloading:
		return "unloading"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// ScanConfig contains configuration for model scanning
type ScanConfig struct {
	Paths          []string
	Recursive      bool
	FollowSymlinks bool
	MaxDepth       int
	IncludePattern string // Regex pattern for files to include
	ExcludePattern string // Regex pattern for files to exclude
}

// ScanResult represents the result of a scan operation
type ScanResult struct {
	Models       []*Model
	Errors       []ScanError
	ScannedAt    time.Time
	Duration     time.Duration
	TotalFiles   int
	MatchedFiles int
}

// ScanError represents an error during scanning
type ScanError struct {
	Path  string
	Error string
}

// LoadRequest contains parameters for loading a model
type LoadRequest struct {
	ModelID       string
	CtxSize       int
	BatchSize     int
	Threads       int
	GPULayers     int
	Temperature   float64
	TopP          float64
	TopK          int
	RepeatPenalty float64
	Seed          int
	NPredict      int
}

// LoadResult represents the result of a load operation
type LoadResult struct {
	Success  bool
	ModelID  string
	Port     int
	CtxSize  int
	Error    error
	Duration time.Duration
}
