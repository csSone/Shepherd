// Package tester provides llama.cpp availability testing functionality.
package tester

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/types"
)

// Tester tests llama.cpp availability
type Tester struct {
	timeout time.Duration
}

// NewTester creates a new llama.cpp tester
func NewTester(timeout time.Duration) *Tester {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Tester{timeout: timeout}
}

// TestResult contains the result of testing a specific binary
type TestResult struct {
	Path     string
	Exists   bool
	Runnable bool
	Version  string
	Error    string
}

// TestAll tests all available llama.cpp binaries and returns the first working one
func (t *Tester) TestAll() *types.LlamacppTestResult {
	start := time.Now()

	// Find all potential binaries
	binaries := t.findBinaries()

	for _, binary := range binaries {
		result := t.TestBinary(binary)
		if result.Runnable {
			return &types.LlamacppTestResult{
				Success:    true,
				BinaryPath: result.Path,
				Version:    result.Version,
				Output:     fmt.Sprintf("Found working binary: %s", result.Path),
				TestedAt:   time.Now(),
				Duration:   time.Since(start).Milliseconds(),
			}
		}
	}

	// None worked
	return &types.LlamacppTestResult{
		Success:  false,
		Error:    "No working llama.cpp binary found",
		TestedAt: time.Now(),
		Duration: time.Since(start).Milliseconds(),
	}
}

// TestBinary tests a specific llama.cpp binary
func (t *Tester) TestBinary(path string) *TestResult {
	result := &TestResult{Path: path}

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		result.Error = fmt.Sprintf("Binary not found: %v", err)
		return result
	}
	result.Exists = true

	// Test if binary is executable by running --help
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--help")
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		result.Error = "Timeout while testing binary"
		return result
	}

	if err != nil {
		result.Error = fmt.Sprintf("Failed to run binary: %v", err)
		return result
	}

	result.Runnable = true
	result.Version = t.extractVersion(string(output))

	return result
}

// TestSpecific tests a specific binary path
func (t *Tester) TestSpecific(path string) *types.LlamacppTestResult {
	start := time.Now()

	result := t.TestBinary(path)

	if result.Runnable {
		return &types.LlamacppTestResult{
			Success:    true,
			BinaryPath: result.Path,
			Version:    result.Version,
			Output:     fmt.Sprintf("Binary is working: %s", result.Path),
			TestedAt:   time.Now(),
			Duration:   time.Since(start).Milliseconds(),
		}
	}

	return &types.LlamacppTestResult{
		Success:    false,
		BinaryPath: path,
		Error:      result.Error,
		TestedAt:   time.Now(),
		Duration:   time.Since(start).Milliseconds(),
	}
}

// findBinaries finds potential llama.cpp binaries
func (t *Tester) findBinaries() []string {
	var binaries []string

	// Check environment variable
	if envPath := os.Getenv("LLAMACPP_SERVER_PATH"); envPath != "" {
		binaries = append(binaries, envPath)
	}

	// Common paths
	commonPaths := []string{
		"/home/user/workspace/llama.cpp/build/bin/server",
		"/usr/local/bin/llama-server",
		"/usr/bin/llama-server",
		filepath.Join(os.Getenv("HOME"), "llama.cpp/build/bin/server"),
		"/home/user/miniconda3/envs/rocm7.2/bin/llama-server",
		"/opt/llama.cpp/server",
	}

	for _, path := range commonPaths {
		if t.isUnique(path, binaries) {
			binaries = append(binaries, path)
		}
	}

	// Check PATH
	if path, err := exec.LookPath("llama-server"); err == nil {
		if t.isUnique(path, binaries) {
			binaries = append(binaries, path)
		}
	}

	return binaries
}

// isUnique checks if a path is not already in the list
func (t *Tester) isUnique(path string, list []string) bool {
	for _, p := range list {
		if p == path {
			return false
		}
	}
	return true
}

// extractVersion tries to extract version from help output
func (t *Tester) extractVersion(output string) string {
	// Common version patterns
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		if strings.Contains(line, "version") {
			// Try to extract version number
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.Contains(part, "version") && i+1 < len(parts) {
					return strings.TrimSpace(parts[i+1])
				}
			}
		}
	}

	// Return first non-empty line as fallback
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}

	return "unknown"
}

// GetSystemInfo returns system information relevant to llama.cpp
func (t *Tester) GetSystemInfo() map[string]interface{} {
	info := map[string]interface{}{
		"os":           runtime.GOOS,
		"architecture": runtime.GOARCH,
		"cpus":         runtime.NumCPU(),
	}

	// Check for GPU support
	info["cuda_available"] = t.checkCUDA()
	info["rocm_available"] = t.checkROCm()

	return info
}

// checkCUDA checks if CUDA is available
func (t *Tester) checkCUDA() bool {
	_, err := exec.LookPath("nvidia-smi")
	return err == nil
}

// checkROCm checks if ROCm is available
func (t *Tester) checkROCm() bool {
	_, err := exec.LookPath("rocm-smi")
	return err == nil
}
