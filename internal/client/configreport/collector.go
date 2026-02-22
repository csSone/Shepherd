// Package configreport provides configuration reporting functionality for client nodes.
package configreport

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/types"
)

// Collector collects client configuration information
type Collector struct {
	cfg *config.Config
}

// NewCollector creates a new configuration collector
func NewCollector(cfg *config.Config) *Collector {
	return &Collector{cfg: cfg}
}

// Collect collects all configuration information
func (c *Collector) Collect() *types.NodeConfigInfo {
	info := &types.NodeConfigInfo{
		ReportedAt: time.Now(),
	}

	// Collect llama.cpp paths
	info.LlamaCppPaths = c.collectLlamaCppPaths()

	// Collect model paths
	info.ModelPaths = c.collectModelPaths()

	// Collect environment info
	info.Environment = c.collectEnvironment()

	// Collect conda config
	info.Conda = c.collectCondaConfig()

	// Collect executor config
	info.Executor = c.collectExecutorConfig()

	return info
}

// collectLlamaCppPaths collects llama.cpp binary paths
func (c *Collector) collectLlamaCppPaths() []types.LlamaCppPathInfo {
	var paths []types.LlamaCppPathInfo

	// From configuration
	for _, lp := range c.cfg.Llamacpp.Paths {
		info := types.LlamaCppPathInfo{
			Path:        lp.Path,
			Name:        lp.Name,
			Description: lp.Description,
			Exists:      fileExists(lp.Path),
		}
		// Try to get version
		if info.Exists {
			info.Version = c.getLlamaCppVersion(lp.Path)
		}
		paths = append(paths, info)
	}

	// Check common paths
	commonPaths := []string{
		"/home/user/workspace/llama.cpp/build/bin/server",
		"/usr/local/bin/llama-server",
		filepath.Join(os.Getenv("HOME"), "llama.cpp/build/bin/server"),
		"/home/user/miniconda3/envs/rocm7.2/bin/llama-server",
		"/opt/llama.cpp/server",
	}

	// Check environment variable
	if envPath := os.Getenv("LLAMACPP_SERVER_PATH"); envPath != "" {
		commonPaths = append([]string{envPath}, commonPaths...)
	}

	for _, path := range commonPaths {
		// Skip if already in config
		if c.pathInList(path, paths) {
			continue
		}

		if fileExists(path) {
			info := types.LlamaCppPathInfo{
				Path:   path,
				Name:   filepath.Base(path),
				Exists: true,
			}
			info.Version = c.getLlamaCppVersion(path)
			paths = append(paths, info)
		}
	}

	// Check PATH
	if path, err := exec.LookPath("llama-server"); err == nil {
		if !c.pathInList(path, paths) {
			info := types.LlamaCppPathInfo{
				Path:   path,
				Name:   "llama-server (PATH)",
				Exists: true,
			}
			info.Version = c.getLlamaCppVersion(path)
			paths = append(paths, info)
		}
	}

	return paths
}

// collectModelPaths collects model directory paths
func (c *Collector) collectModelPaths() []types.ModelPathInfo {
	var paths []types.ModelPathInfo

	// From configuration
	for _, mp := range c.cfg.Model.PathConfigs {
		info := types.ModelPathInfo{
			Path:        mp.Path,
			Name:        mp.Name,
			Description: mp.Description,
			Exists:      dirExists(mp.Path),
		}
		if info.Exists {
			info.ModelCount = c.countModels(mp.Path)
		}
		paths = append(paths, info)
	}

	// Also check simple paths
	for _, path := range c.cfg.Model.Paths {
		if c.modelPathInList(path, paths) {
			continue
		}
		info := types.ModelPathInfo{
			Path:   path,
			Exists: dirExists(path),
		}
		if info.Exists {
			info.ModelCount = c.countModels(path)
		}
		paths = append(paths, info)
	}

	return paths
}

// collectEnvironment collects environment information
func (c *Collector) collectEnvironment() *types.EnvironmentInfo {
	info := &types.EnvironmentInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	// Get kernel version
	if runtime.GOOS == "linux" {
		if data, err := exec.Command("uname", "-r").Output(); err == nil {
			info.KernelVersion = strings.TrimSpace(string(data))
		}
	}

	// Get ROCm version
	if data, err := exec.Command("rocm-smi", "--showdriverversion").Output(); err == nil {
		info.ROCmVersion = strings.TrimSpace(string(data))
	}

	// Get Python version
	if data, err := exec.Command("python", "--version").Output(); err == nil {
		info.PythonVersion = strings.TrimSpace(string(data))
	} else if data, err := exec.Command("python3", "--version").Output(); err == nil {
		info.PythonVersion = strings.TrimSpace(string(data))
	}

	// Get Go version
	info.GoVersion = runtime.Version()

	return info
}

// collectCondaConfig collects conda configuration
func (c *Collector) collectCondaConfig() *types.CondaConfigInfo {
	// Check new NodeCapabilitiesConfig first
	if c.cfg.Node.Capabilities.PythonEnabled {
		return &types.CondaConfigInfo{
			Enabled:      true,
			CondaPath:    c.cfg.Node.Capabilities.CondaPath,
			Environments: c.cfg.Node.Capabilities.CondaEnvironments,
		}
	}

	// Check deprecated ClientConfig.CondaEnv
	if c.cfg.Client.CondaEnv.Enabled {
		return &types.CondaConfigInfo{
			Enabled:      true,
			CondaPath:    c.cfg.Client.CondaEnv.CondaPath,
			Environments: c.cfg.Client.CondaEnv.Environments,
		}
	}

	return &types.CondaConfigInfo{Enabled: false}
}

// collectExecutorConfig collects executor configuration
func (c *Collector) collectExecutorConfig() *types.ExecutorConfigInfo {
	return &types.ExecutorConfigInfo{
		MaxConcurrent:   c.cfg.Node.Executor.MaxConcurrent,
		TaskTimeout:     c.cfg.Node.Executor.TaskTimeout,
		AllowRemoteStop: c.cfg.Node.Executor.AllowRemoteStop,
		AllowedCommands: c.cfg.Node.Executor.AllowedCommands,
	}
}

// getLlamaCppVersion tries to get llama.cpp version
func (c *Collector) getLlamaCppVersion(binaryPath string) string {
	// Try --version flag
	if data, err := exec.Command(binaryPath, "--version").Output(); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Try -v flag
	if data, err := exec.Command(binaryPath, "-v").Output(); err == nil {
		return strings.TrimSpace(string(data))
	}

	return ""
}

// pathInList checks if a path is already in the list
func (c *Collector) pathInList(path string, list []types.LlamaCppPathInfo) bool {
	for _, item := range list {
		if item.Path == path {
			return true
		}
	}
	return false
}

// modelPathInList checks if a model path is already in the list
func (c *Collector) modelPathInList(path string, list []types.ModelPathInfo) bool {
	for _, item := range list {
		if item.Path == path {
			return true
		}
	}
	return false
}

// countModels counts GGUF models in a directory
func (c *Collector) countModels(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".gguf") {
			count++
		}
	}
	return count
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
