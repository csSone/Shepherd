// Package config provides configuration management for the Shepherd server.
// It handles loading, saving, and validating configuration from YAML files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// DefaultConfigDir is the default configuration directory
	DefaultConfigDir = "config"
	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "server.config.yaml"
	// DefaultModelsConfigFile is the default models configuration file
	DefaultModelsConfigFile = "models.json"
	// DefaultLaunchConfigFile is the default launch configuration file
	DefaultLaunchConfigFile = "launch_config.json"
)

// ConfigFileNames maps mode to config file name
var ConfigFileNames = map[string]string{
	"standalone": "server.config.yaml",
	"master":     "master.config.yaml",
	"client":     "client.config.yaml",
}

// Config represents the complete application configuration
type Config struct {
	Server        ServerConfig        `mapstructure:"server" yaml:"server" json:"server"`
	Model         ModelConfig         `mapstructure:"model" yaml:"model" json:"model"`
	Llamacpp      LlamacppConfig      `mapstructure:"llamacpp" yaml:"llamacpp" json:"llamacpp"`
	Download      DownloadConfig      `mapstructure:"download" yaml:"download" json:"download"`
	Security      SecurityConfig      `mapstructure:"security" yaml:"security" json:"security"`
	Compatibility CompatibilityConfig `mapstructure:"compatibility" yaml:"compatibility" json:"compatibility"`
	Log           LogConfig           `mapstructure:"log" yaml:"log" json:"log"`
	// Master-Client 分布式配置
	Mode          string              `mapstructure:"mode" yaml:"mode" json:"mode"` // master|client|standalone
	Master        MasterConfig        `mapstructure:"master" yaml:"master" json:"master"`
	Client        ClientConfig        `mapstructure:"client" yaml:"client" json:"client"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	WebPort       int  `mapstructure:"web_port" yaml:"web_port" json:"webPort"`
	AnthropicPort int  `mapstructure:"anthropic_port" yaml:"anthropic_port" json:"anthropicPort"`
	OllamaPort    int  `mapstructure:"ollama_port" yaml:"ollama_port" json:"ollamaPort"`
	LMStudioPort  int  `mapstructure:"lmstudio_port" yaml:"lmstudio_port" json:"lmstudioPort"`
	Host          string `mapstructure:"host" yaml:"host" json:"host"`
	ReadTimeout   int    `mapstructure:"read_timeout" yaml:"read_timeout" json:"readTimeout"`   // seconds
	WriteTimeout  int    `mapstructure:"write_timeout" yaml:"write_timeout" json:"writeTimeout"` // seconds
}

// ModelConfig contains model scanning and management configuration
type ModelConfig struct {
	Paths    []string `mapstructure:"paths" yaml:"paths" json:"paths"`
	AutoScan bool     `mapstructure:"auto_scan" yaml:"auto_scan" json:"autoScan"`
	ScanInterval int  `mapstructure:"scan_interval" yaml:"scan_interval" json:"scanInterval"` // seconds, 0 = disable
}

// LlamacppConfig contains llama.cpp binary paths configuration
type LlamacppConfig struct {
	Paths []LlamacppPath `mapstructure:"paths" yaml:"paths" json:"paths"`
}

// LlamacppPath represents a llama.cpp binary path with metadata
type LlamacppPath struct {
	Path string `mapstructure:"path" yaml:"path" json:"path"`
	Name string `mapstructure:"name" yaml:"name" json:"name"`
}

// DownloadConfig contains download manager configuration
type DownloadConfig struct {
	Directory     string `mapstructure:"directory" yaml:"directory" json:"directory"`
	MaxConcurrent int    `mapstructure:"max_concurrent" yaml:"max_concurrent" json:"maxConcurrent"`
	ChunkSize     int    `mapstructure:"chunk_size" yaml:"chunk_size" json:"chunkSize"` // bytes
	RetryCount    int    `mapstructure:"retry_count" yaml:"retry_count" json:"retryCount"`
	Timeout       int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"` // seconds
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	APIKeyEnabled bool   `mapstructure:"api_key_enabled" yaml:"api_key_enabled" json:"apiKeyEnabled"`
	APIKey        string `mapstructure:"api_key" yaml:"api_key" json:"apiKey"`
	CORSEnabled   bool   `mapstructure:"cors_enabled" yaml:"cors_enabled" json:"corsEnabled"`
	AllowedOrigins []string `mapstructure:"allowed_origins" yaml:"allowed_origins" json:"allowedOrigins"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level      string `mapstructure:"level" yaml:"level" json:"level"` // debug, info, warn, error
	Format     string `mapstructure:"format" yaml:"format" json:"format"` // json, text
	Output     string `mapstructure:"output" yaml:"output" json:"output"` // stdout, file, both
	Directory  string `mapstructure:"directory" yaml:"directory" json:"directory"` // log directory
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size" json:"maxSize"` // MB
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups" json:"maxBackups"` // number of backup files
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age" json:"maxAge"` // days
	Compress   bool   `mapstructure:"compress" yaml:"compress" json:"compress"` // compress old logs
}

// CompatibilityConfig contains API compatibility layer settings
type CompatibilityConfig struct {
	Ollama   OllamaConfig   `mapstructure:"ollama" yaml:"ollama" json:"ollama"`
	LMStudio LMStudioConfig `mapstructure:"lmstudio" yaml:"lmstudio" json:"lmstudio"`
}

// OllamaConfig contains Ollama API compatibility settings
type OllamaConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Port    int  `mapstructure:"port" yaml:"port" json:"port"`
}

// LMStudioConfig contains LM Studio API compatibility settings
type LMStudioConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Port    int  `mapstructure:"port" yaml:"port" json:"port"`
}

// MasterConfig contains Master node configuration
type MasterConfig struct {
	Enabled         bool               `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ClientConfigDir string             `mapstructure:"client_config_dir" yaml:"client_config_dir" json:"clientConfigDir"`
	NetworkScan     NetworkScanConfig  `mapstructure:"network_scan" yaml:"network_scan" json:"networkScan"`
	Scheduler       SchedulerConfig    `mapstructure:"scheduler" yaml:"scheduler" json:"scheduler"`
	LogAggregation  LogAggregationConfig `mapstructure:"log_aggregation" yaml:"log_aggregation" json:"logAggregation"`
}

// NetworkScanConfig contains network scanner configuration
type NetworkScanConfig struct {
	Enabled      bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Subnets      []string `mapstructure:"subnets" yaml:"subnets" json:"subnets"`
	PortRange    string   `mapstructure:"port_range" yaml:"port_range" json:"portRange"`
	Timeout      int      `mapstructure:"timeout" yaml:"timeout" json:"timeout"` // seconds
	AutoDiscover bool     `mapstructure:"auto_discover" yaml:"auto_discover" json:"autoDiscover"`
	Interval     int      `mapstructure:"interval" yaml:"interval" json:"interval"` // seconds, 0 = manual
}

// SchedulerConfig contains task scheduler configuration
type SchedulerConfig struct {
	Strategy          string `mapstructure:"strategy" yaml:"strategy" json:"strategy"` // round_robin, least_loaded, resource_aware
	MaxQueueSize      int    `mapstructure:"max_queue_size" yaml:"max_queue_size" json:"maxQueueSize"`
	TaskTimeout       int    `mapstructure:"task_timeout" yaml:"task_timeout" json:"taskTimeout"` // seconds
	RetryOnFailure    bool   `mapstructure:"retry_on_failure" yaml:"retry_on_failure" json:"retryOnFailure"`
	MaxRetries        int    `mapstructure:"max_retries" yaml:"max_retries" json:"maxRetries"`
}

// LogAggregationConfig contains log aggregation settings
type LogAggregationConfig struct {
	Enabled      bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	MaxBufferSize int   `mapstructure:"max_buffer_size" yaml:"max_buffer_size" json:"maxBufferSize"` // bytes per client
	FlushInterval int   `mapstructure:"flush_interval" yaml:"flush_interval" json:"flushInterval"` // seconds
}

// ClientConfig contains Client node configuration
type ClientConfig struct {
	Enabled         bool              `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	MasterAddress   string            `mapstructure:"master_address" yaml:"master_address" json:"masterAddress"`
	ClientInfo      ClientInfoConfig  `mapstructure:"client_info" yaml:"client_info" json:"clientInfo"`
	Heartbeat       HeartbeatConfig   `mapstructure:"heartbeat" yaml:"heartbeat" json:"heartbeat"`
	CondaEnv        CondaEnvConfig    `mapstructure:"conda_env" yaml:"conda_env" json:"condaEnv"`
}

// ClientInfoConfig contains client identification information
type ClientInfoConfig struct {
	ID       string            `mapstructure:"id" yaml:"id" json:"id"` // Auto-generated if empty
	Name     string            `mapstructure:"name" yaml:"name" json:"name"`
	Tags     []string          `mapstructure:"tags" yaml:"tags" json:"tags"`
	Metadata map[string]string `mapstructure:"metadata" yaml:"metadata" json:"metadata"`
}

// HeartbeatConfig contains heartbeat settings
type HeartbeatConfig struct {
	Interval int `mapstructure:"interval" yaml:"interval" json:"interval"` // seconds
	Timeout  int `mapstructure:"timeout" yaml:"timeout" json:"timeout"`  // seconds
}

// CondaEnvConfig contains conda environment configuration
type CondaEnvConfig struct {
	Enabled      bool              `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	CondaPath    string            `mapstructure:"conda_path" yaml:"conda_path" json:"condaPath"`
	Environments map[string]string `mapstructure:"environments" yaml:"environments" json:"environments"` // name -> path
}

// ModelConfigEntry represents a model configuration entry in models.json
type ModelConfigEntry struct {
	ModelID     string              `json:"modelId"`
	Path        string              `json:"path,omitempty"`
	Size        int64               `json:"size,omitempty"`
	Alias       string              `json:"alias,omitempty"`
	Favourite   bool                `json:"favourite"`
	PrimaryModel *PrimaryModelInfo  `json:"primaryModel,omitempty"`
	Mmproj      *MmprojInfo         `json:"mmproj,omitempty"`
}

// PrimaryModelInfo contains information about the primary model
type PrimaryModelInfo struct {
	FileName         string `json:"fileName"`
	Name             string `json:"name,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
	ContextLength    int    `json:"contextLength,omitempty"`
	EmbeddingLength  int    `json:"embeddingLength,omitempty"`
}

// MmprojInfo contains information about the multimodal projector
type MmprojInfo struct {
	FileName     string `json:"fileName"`
	Name         string `json:"name,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

// LaunchConfig represents model launch parameters
type LaunchConfig struct {
	CtxSize       int     `mapstructure:"ctx_size" yaml:"ctx_size" json:"ctxSize"`
	BatchSize     int     `mapstructure:"batch_size" yaml:"batch_size" json:"batchSize"`
	Threads       int     `mapstructure:"threads" yaml:"threads" json:"threads"`
	GPULayers     int     `mapstructure:"gpu_layers" yaml:"gpu_layers" json:"gpuLayers"`
	Temperature   float64 `mapstructure:"temperature" yaml:"temperature" json:"temperature"`
	TopP          float64 `mapstructure:"top_p" yaml:"top_p" json:"topP"`
	TopK          int     `mapstructure:"top_k" yaml:"top_k" json:"topK"`
	RepeatPenalty float64 `mapstructure:"repeat_penalty" yaml:"repeat_penalty" json:"repeatPenalty"`
	Seed          int     `mapstructure:"seed" yaml:"seed" json:"seed"`
	NPredict      int     `mapstructure:"n_predict" yaml:"n_predict" json:"nPredict"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	// Get current working directory or use default
	cwd, _ := os.Getwd()
	downloadDir := filepath.Join(cwd, "downloads")
	logDir := filepath.Join(cwd, "logs")
	modelPaths := []string{
		filepath.Join(cwd, "models"),
		filepath.Join(os.Getenv("HOME"), ".cache/huggingface/hub"),
	}

	return &Config{
		Mode: "standalone", // 默认单机模式
		Server: ServerConfig{
			WebPort:       9190,
			AnthropicPort: 9170,
			OllamaPort:    11434,
			LMStudioPort:  1234,
			Host:         "0.0.0.0",
			ReadTimeout:  60,
			WriteTimeout: 60,
		},
		Model: ModelConfig{
			Paths:        modelPaths,
			AutoScan:     true,
			ScanInterval: 0,
		},
		Llamacpp: LlamacppConfig{
			Paths: []LlamacppPath{
				{
					Path: filepath.Join(cwd, "llama.cpp"),
					Name: "Default",
				},
			},
		},
		Download: DownloadConfig{
			Directory:     downloadDir,
			MaxConcurrent: 4,
			ChunkSize:     1024 * 1024, // 1MB
			RetryCount:    3,
			Timeout:       300, // 5 minutes
		},
		Security: SecurityConfig{
			APIKeyEnabled:  false,
			APIKey:         "",
			CORSEnabled:    true,
			AllowedOrigins: []string{"*"},
		},
		Compatibility: CompatibilityConfig{
			Ollama: OllamaConfig{
				Enabled: true,
				Port:    11434,
			},
			LMStudio: LMStudioConfig{
				Enabled: false,
				Port:    1234,
			},
		},
		Log: LogConfig{
			Level:      "info",
			Format:     "json",
			Output:     "both", // stdout + file
			Directory:  logDir,
			MaxSize:    100, // 100MB
			MaxBackups: 3,
			MaxAge:     7, // 7 days
			Compress:   true,
		},
		Master: MasterConfig{
			Enabled:        false,
			ClientConfigDir: filepath.Join(cwd, "config", "clients"),
			NetworkScan: NetworkScanConfig{
				Enabled:      false,
				Subnets:      []string{"192.168.1.0/24", "10.0.0.0/8"},
				PortRange:    "9191-9200",
				Timeout:      5,
				AutoDiscover: false,
				Interval:     0,
			},
			Scheduler: SchedulerConfig{
				Strategy:       "round_robin",
				MaxQueueSize:   100,
				TaskTimeout:    300, // 5 minutes
				RetryOnFailure: true,
				MaxRetries:     3,
			},
			LogAggregation: LogAggregationConfig{
				Enabled:       true,
				MaxBufferSize: 1024 * 1024, // 1MB per client
				FlushInterval: 10,           // 10 seconds
			},
		},
		Client: ClientConfig{
			Enabled:       false,
			MasterAddress: "",
			ClientInfo: ClientInfoConfig{
				ID:   "", // Auto-generated
				Name: "", // Will use hostname
				Tags: []string{},
				Metadata: map[string]string{
					"os":   "linux",
					"arch": "amd64",
				},
			},
			Heartbeat: HeartbeatConfig{
				Interval: 30, // 30 seconds
				Timeout:  90, // 90 seconds
			},
			CondaEnv: CondaEnvConfig{
				Enabled:   false,
				CondaPath: "/home/user/miniconda3",
				Environments: map[string]string{
					"rocm7.2": "/home/user/miniconda3/envs/rocm7.2",
				},
			},
		},
	}
}

// DefaultLaunchConfig returns default launch parameters
func DefaultLaunchConfig() *LaunchConfig {
	return &LaunchConfig{
		CtxSize:       4096,
		BatchSize:     512,
		Threads:       8,
		GPULayers:     99,
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          40,
		RepeatPenalty: 1.1,
		Seed:          -1, // Random
		NPredict:      -1, // Unlimited
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate mode
	validModes := map[string]bool{"standalone": true, "master": true, "client": true}
	if !validModes[c.Mode] {
		return fmt.Errorf("invalid mode: %s (must be standalone, master, or client)", c.Mode)
	}

	// Validate server ports
	if c.Server.WebPort < 1 || c.Server.WebPort > 65535 {
		return fmt.Errorf("invalid web port: %d", c.Server.WebPort)
	}
	if c.Server.AnthropicPort < 1 || c.Server.AnthropicPort > 65535 {
		return fmt.Errorf("invalid anthropic port: %d", c.Server.AnthropicPort)
	}
	if c.Server.OllamaPort < 1 || c.Server.OllamaPort > 65535 {
		return fmt.Errorf("invalid ollama port: %d", c.Server.OllamaPort)
	}
	if c.Server.LMStudioPort < 1 || c.Server.LMStudioPort > 65535 {
		return fmt.Errorf("invalid lmstudio port: %d", c.Server.LMStudioPort)
	}

	// Check for port conflicts
	ports := map[int]string{
		c.Server.WebPort:       "web",
		c.Server.AnthropicPort: "anthropic",
	}
	if c.Compatibility.Ollama.Enabled {
		if _, exists := ports[c.Server.OllamaPort]; exists {
			return fmt.Errorf("port conflict: ollama port %d conflicts with another service", c.Server.OllamaPort)
		}
		ports[c.Server.OllamaPort] = "ollama"
	}
	if c.Compatibility.LMStudio.Enabled {
		if _, exists := ports[c.Server.LMStudioPort]; exists {
			return fmt.Errorf("port conflict: lmstudio port %d conflicts with another service", c.Server.LMStudioPort)
		}
		ports[c.Server.LMStudioPort] = "lmstudio"
	}

	// Validate download settings
	if c.Download.MaxConcurrent < 1 {
		return fmt.Errorf("max concurrent downloads must be at least 1")
	}
	if c.Download.ChunkSize < 1024 {
		return fmt.Errorf("chunk size too small (minimum 1024 bytes)")
	}

	// Validate model paths
	for _, path := range c.Model.Paths {
		if path == "" {
			return fmt.Errorf("model path cannot be empty")
		}
	}

	// Validate Master mode specific settings
	if c.Mode == "master" && c.Master.Enabled {
		if c.Master.ClientConfigDir == "" {
			return fmt.Errorf("master client config directory cannot be empty")
		}
		if c.Master.NetworkScan.Enabled && len(c.Master.NetworkScan.Subnets) == 0 {
			return fmt.Errorf("network scan enabled but no subnets configured")
		}
	}

	// Validate Client mode specific settings
	if c.Mode == "client" && c.Client.Enabled {
		if c.Client.MasterAddress == "" {
			return fmt.Errorf("client mode requires master address")
		}
		if c.Client.Heartbeat.Interval < 1 {
			return fmt.Errorf("heartbeat interval must be at least 1 second")
		}
		if c.Client.CondaEnv.Enabled && c.Client.CondaEnv.CondaPath == "" {
			return fmt.Errorf("conda enabled but conda path is empty")
		}
	}

	return nil
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	// Allow override via environment variable
	if dir := os.Getenv("SHEPHERD_CONFIG_DIR"); dir != "" {
		return dir
	}
	return DefaultConfigDir
}

// EnsureConfigDir ensures the configuration directory exists
func EnsureConfigDir() error {
	configDir := GetConfigDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}
	return nil
}

// Manager manages configuration loading and saving
type Manager struct {
	config            *Config
	configPath        string
	modelsConfigPath  string
	launchConfigPath  string
	mu                sync.RWMutex
	cachedModels      []ModelConfigEntry
	cachedModelsTime  int64
	mode              string // 运行模式
}

// NewManager creates a new configuration manager
func NewManager(mode string) *Manager {
	configDir := GetConfigDir()
	configFile := DefaultConfigFile
	if f, ok := ConfigFileNames[mode]; ok {
		configFile = f
	}
	return &Manager{
		configPath:       filepath.Join(configDir, configFile),
		modelsConfigPath: filepath.Join(configDir, DefaultModelsConfigFile),
		launchConfigPath: filepath.Join(configDir, DefaultLaunchConfigFile),
		mode:             mode,
	}
}

// GetConfigPath returns the main configuration file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// GetMode returns the current mode
func (m *Manager) GetMode() string {
	return m.mode
}

// GetModelsConfigPath returns the models configuration file path
func (m *Manager) GetModelsConfigPath() string {
	return m.modelsConfigPath
}

// GetLaunchConfigPath returns the launch configuration file path
func (m *Manager) GetLaunchConfigPath() string {
	return m.launchConfigPath
}
