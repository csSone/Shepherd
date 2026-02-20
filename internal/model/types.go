// Package model provides model scanning and management functionality.
// It handles discovering GGUF models, reading their metadata, and managing model loading.
package model

import (
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/gguf"
)

// Model represents a discovered GGUF model with HuggingFace-style management
type Model struct {
	// Basic information
	ID          string   // Unique model ID (path hash)
	Name        string   // Model name
	DisplayName string   // Display name for UI (handles duplicates)
	Alias       string   // Display alias
	Description string   // Model description/card (HuggingFace-style)
	Path        string   // File path to the GGUF file
	PathPrefix  string   // Path prefix for duplicate identification (e.g., "models/A", "cache/B")
	Size        int64    // File size in bytes
	Favourite   bool     // User's favorite flag
	Tags        []string // Model tags for categorization (e.g., "chat", "code", "multilingual")
	License     string   // Model license
	Author      string   // Model author/organization
	Downloads   int      // Download count for downloaded models

	// GGUF metadata
	Metadata *gguf.Metadata

	// Additional files (e.g., mmproj)
	MmprojPath string
	MmprojMeta *gguf.Metadata

	// Scanning info
	ScannedAt  time.Time
	SourcePath string // Original scan path
	SourceType string // "local", "huggingface", "modelscope"

	// Usage statistics
	LoadCount   int       // Number of times loaded
	LastLoaded  time.Time // Last load time
	TotalTokens int64     // Total tokens generated (if tracked)
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

// ModelFilter represents filter criteria for model search
type ModelFilter struct {
	Tags         []string
	Architecture string
	MinContext   int
	MaxSize      int64
	LoadedOnly   bool
	Favourites   bool
	SearchQuery  string
	SourceType   string
	License      string
}

// ModelSort represents sort options for model listing
type ModelSort struct {
	Field     string // "name", "size", "scanned_at", "load_count"
	Direction string // "asc", "desc"
}

// ModelSearchResult represents the result of a model search
type ModelSearchResult struct {
	Models        []*Model
	Total         int
	Filtered      int
	Tags          map[string]int // Tag frequency
	Architectures map[string]int // Architecture frequency
}
