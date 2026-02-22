// Package config provides configuration management for the Shepherd server.
// It handles loading, saving, and validating configuration from YAML files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/shepherd-project/shepherd/Shepherd/internal/storage"
)

const (
	// DefaultConfigDir is the default configuration directory
	DefaultConfigDir = "config"
	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "server.config.yaml"
	// DefaultModelsConfigFile is the default models configuration file
	DefaultModelsConfigFile = "node/models.json"
	// DefaultLaunchConfigFile is the default launch configuration file
	DefaultLaunchConfigFile = "launch_config.json"
)

// ConfigFileNames maps mode to config file name
var ConfigFileNames = map[string]string{
	"hybrid":     "server.config.yaml",
	"master":     "master.config.yaml",
	"client":     "client.config.yaml",
}

// Config represents the complete application configuration
type Config struct {
	Server        ServerConfig          `mapstructure:"server" yaml:"server" json:"server"`
	Model         ModelConfig           `mapstructure:"model" yaml:"model" json:"model"`
	Llamacpp      LlamacppConfig        `mapstructure:"llamacpp" yaml:"llamacpp" json:"llamacpp"`
	Download      DownloadConfig        `mapstructure:"download" yaml:"download" json:"download"`
	ModelRepo     ModelRepoConfig       `mapstructure:"model_repo" yaml:"model_repo" json:"modelRepo"`
	Security      SecurityConfig        `mapstructure:"security" yaml:"security" json:"security"`
	Compatibility CompatibilityConfig   `mapstructure:"compatibility" yaml:"compatibility" json:"compatibility"`
	Log           LogConfig             `mapstructure:"log" yaml:"log" json:"log"`
	Storage       storage.StorageConfig `mapstructure:"storage" yaml:"storage" json:"storage"`
	// Master-Client 分布式配置
	Mode   string       `mapstructure:"mode" yaml:"mode" json:"mode"`
	Master MasterConfig `mapstructure:"master" yaml:"master" json:"master"`
	Client ClientConfig `mapstructure:"client" yaml:"client" json:"client"`
	// Node 节点配置
	Node NodeConfig `mapstructure:"node" yaml:"node" json:"node"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	WebPort       int    `mapstructure:"web_port" yaml:"web_port" json:"webPort"`
	AnthropicPort int    `mapstructure:"anthropic_port" yaml:"anthropic_port" json:"anthropicPort"`
	OllamaPort    int    `mapstructure:"ollama_port" yaml:"ollama_port" json:"ollamaPort"`
	LMStudioPort  int    `mapstructure:"lmstudio_port" yaml:"lmstudio_port" json:"lmstudioPort"`
	Host          string `mapstructure:"host" yaml:"host" json:"host"`
	ReadTimeout   int    `mapstructure:"read_timeout" yaml:"read_timeout" json:"readTimeout"`    // seconds
	WriteTimeout  int    `mapstructure:"write_timeout" yaml:"write_timeout" json:"writeTimeout"` // seconds
}

// ModelConfig contains model scanning and management configuration
type ModelConfig struct {
	Paths        []string    `mapstructure:"paths" yaml:"paths" json:"paths"`                     // 简单路径数组（向后兼容）
	PathConfigs  []ModelPath `mapstructure:"path_configs" yaml:"path_configs" json:"pathConfigs"` // 详细路径配置
	AutoScan     bool        `mapstructure:"auto_scan" yaml:"auto_scan" json:"autoScan"`
	ScanInterval int         `mapstructure:"scan_interval" yaml:"scan_interval" json:"scanInterval"` // seconds, 0 = disable
}

// LlamacppConfig contains llama.cpp binary paths configuration
type LlamacppConfig struct {
	Paths []LlamacppPath `mapstructure:"paths" yaml:"paths" json:"paths"`
}

// LlamacppPath represents a llama.cpp binary path with metadata
type LlamacppPath struct {
	Path        string `mapstructure:"path" yaml:"path" json:"path"`
	Name        string `mapstructure:"name" yaml:"name" json:"name"`
	Description string `mapstructure:"description" yaml:"description" json:"description,omitempty"`
}

// ModelPath represents a model directory path with metadata
type ModelPath struct {
	Path        string `mapstructure:"path" yaml:"path" json:"path"`
	Name        string `mapstructure:"name" yaml:"name" json:"name,omitempty"`
	Description string `mapstructure:"description" yaml:"description" json:"description,omitempty"`
}

// DownloadConfig contains download manager configuration
type DownloadConfig struct {
	Directory     string `mapstructure:"directory" yaml:"directory" json:"directory"`
	MaxConcurrent int    `mapstructure:"max_concurrent" yaml:"max_concurrent" json:"maxConcurrent"`
	ChunkSize     int    `mapstructure:"chunk_size" yaml:"chunk_size" json:"chunkSize"` // bytes
	RetryCount    int    `mapstructure:"retry_count" yaml:"retry_count" json:"retryCount"`
	Timeout       int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"` // seconds
}

// ModelRepoConfig contains model repository configuration
type ModelRepoConfig struct {
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"` // huggingface.co or hf-mirror.com
	Token    string `mapstructure:"token" yaml:"token" json:"token"`          // HuggingFace API token
	Timeout  int    `mapstructure:"timeout" yaml:"timeout" json:"timeout"`    // seconds
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	APIKeyEnabled  bool     `mapstructure:"api_key_enabled" yaml:"api_key_enabled" json:"apiKeyEnabled"`
	APIKey         string   `mapstructure:"api_key" yaml:"api_key" json:"apiKey"`
	CORSEnabled    bool     `mapstructure:"cors_enabled" yaml:"cors_enabled" json:"corsEnabled"`
	AllowedOrigins []string `mapstructure:"allowed_origins" yaml:"allowed_origins" json:"allowedOrigins"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level      string `mapstructure:"level" yaml:"level" json:"level"`                  // debug, info, warn, error
	Format     string `mapstructure:"format" yaml:"format" json:"format"`               // json, text
	Output     string `mapstructure:"output" yaml:"output" json:"output"`               // stdout, file, both
	Directory  string `mapstructure:"directory" yaml:"directory" json:"directory"`      // log directory
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size" json:"maxSize"`          // MB
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups" json:"maxBackups"` // number of backup files
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age" json:"maxAge"`             // days
	Compress   bool   `mapstructure:"compress" yaml:"compress" json:"compress"`         // compress old logs
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
// Deprecated: Use Node.MasterRole instead. This type is kept for backward compatibility.
type MasterConfig struct {
	Enabled         bool                 `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ClientConfigDir string               `mapstructure:"client_config_dir" yaml:"client_config_dir" json:"clientConfigDir"`
	NetworkScan     NetworkScanConfig    `mapstructure:"network_scan" yaml:"network_scan" json:"networkScan"`
	Scheduler       SchedulerConfig      `mapstructure:"scheduler" yaml:"scheduler" json:"scheduler"` // Deprecated: Use Node-specific scheduler
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
	Strategy       string `mapstructure:"strategy" yaml:"strategy" json:"strategy"` // round_robin, least_loaded, resource_aware
	MaxQueueSize   int    `mapstructure:"max_queue_size" yaml:"max_queue_size" json:"maxQueueSize"`
	TaskTimeout    int    `mapstructure:"task_timeout" yaml:"task_timeout" json:"taskTimeout"` // seconds
	RetryOnFailure bool   `mapstructure:"retry_on_failure" yaml:"retry_on_failure" json:"retryOnFailure"`
	MaxRetries     int    `mapstructure:"max_retries" yaml:"max_retries" json:"maxRetries"`
}

// LogAggregationConfig contains log aggregation settings
type LogAggregationConfig struct {
	Enabled       bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	MaxBufferSize int  `mapstructure:"max_buffer_size" yaml:"max_buffer_size" json:"maxBufferSize"` // bytes per client
	FlushInterval int  `mapstructure:"flush_interval" yaml:"flush_interval" json:"flushInterval"`   // seconds
}

// ClientConfig contains Client node configuration
// Deprecated: Use Node.ClientRole instead. This type is kept for backward compatibility.
type ClientConfig struct {
	Enabled       bool             `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	MasterAddress string           `mapstructure:"master_address" yaml:"master_address" json:"masterAddress"` // Deprecated: Use Node.ClientRole.MasterAddress
	ClientInfo    ClientInfoConfig `mapstructure:"client_info" yaml:"client_info" json:"clientInfo"`
	Heartbeat     HeartbeatConfig  `mapstructure:"heartbeat" yaml:"heartbeat" json:"heartbeat"` // Deprecated: Use Node.ClientRole.HeartbeatInterval/Timeout
	CondaEnv      CondaEnvConfig   `mapstructure:"conda_env" yaml:"conda_env" json:"condaEnv"`
	// 注册和心跳配置
	RegisterRetry     int `mapstructure:"register_retry" yaml:"register_retry" json:"registerRetry"`             // Deprecated: Use Node.ClientRole.RegisterRetry
	HeartbeatInterval int `mapstructure:"heartbeat_interval" yaml:"heartbeat_interval" json:"heartbeatInterval"` // Deprecated: Use Node.ClientRole.HeartbeatInterval
	HeartbeatTimeout  int `mapstructure:"heartbeat_timeout" yaml:"heartbeat_timeout" json:"heartbeatTimeout"`    // Deprecated: Use Node.ClientRole.HeartbeatTimeout
}

// NodeConfig contains node configuration for the new distributed architecture
type NodeConfig struct {
	ID       string            `mapstructure:"id" yaml:"id" json:"id"`             // 节点ID，auto表示自动生成
	Name     string            `mapstructure:"name" yaml:"name" json:"name"`       // 节点名称
	Role     string            `mapstructure:"role" yaml:"role" json:"role"`       // 节点角色: standalone/master/client/hybrid
	Tags     []string          `mapstructure:"tags" yaml:"tags" json:"tags"`       // 节点标签
	Metadata map[string]string `mapstructure:"metadata" yaml:"metadata" json:"metadata"` // 节点元数据
	// 各角色配置
	MasterRole NodeMasterRoleConfig `mapstructure:"master_role" yaml:"master_role" json:"masterRole"` // Master角色配置
	ClientRole NodeClientRoleConfig `mapstructure:"client_role" yaml:"client_role" json:"clientRole"` // Client角色配置
	// 资源和执行器配置
	Resources   NodeResourceConfig    `mapstructure:"resources" yaml:"resources" json:"resources"`       // 资源监控配置
	Executor    NodeExecutorConfig    `mapstructure:"executor" yaml:"executor" json:"executor"`          // 命令执行器配置
	Capabilities NodeCapabilitiesConfig `mapstructure:"capabilities" yaml:"capabilities" json:"capabilities"` // 能力配置
}

// NodeMasterRoleConfig contains Master role specific configuration
type NodeMasterRoleConfig struct {
	Enabled bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"` // 是否启用Master角色
	Port    int           `mapstructure:"port" yaml:"port" json:"port"`          // Master服务端口
	APIKey  string        `mapstructure:"api_key" yaml:"api_key" json:"apiKey"`  // API密钥
	SSL     NodeSSLConfig `mapstructure:"ssl" yaml:"ssl" json:"ssl"`             // SSL配置
}

// NodeClientRoleConfig contains Client role specific configuration
type NodeClientRoleConfig struct {
	Enabled           bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`                                 // 是否启用Client角色
	MasterAddress     string `mapstructure:"master_address" yaml:"master_address" json:"masterAddress"`             // Master地址
	RegisterRetry     int    `mapstructure:"register_retry" yaml:"register_retry" json:"registerRetry"`             // 注册重试次数
	HeartbeatInterval int    `mapstructure:"heartbeat_interval" yaml:"heartbeat_interval" json:"heartbeatInterval"` // 心跳间隔（秒）
	HeartbeatTimeout  int    `mapstructure:"heartbeat_timeout" yaml:"heartbeat_timeout" json:"heartbeatTimeout"`    // 心跳超时（秒）
}

// NodeSSLConfig contains SSL/TLS configuration
type NodeSSLConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`      // 是否启用SSL
	CertPath string `mapstructure:"cert_path" yaml:"cert_path" json:"certPath"` // 证书路径
	KeyPath  string `mapstructure:"key_path" yaml:"key_path" json:"keyPath"`    // 密钥路径
}

// NodeResourceConfig contains resource monitoring configuration
type NodeResourceConfig struct {
	MonitorInterval   int    `mapstructure:"monitor_interval" yaml:"monitor_interval" json:"monitorInterval"`       // 监控间隔（秒）
	ReportGPU         bool   `mapstructure:"report_gpu" yaml:"report_gpu" json:"reportGPU"`                         // 是否报告GPU信息
	ReportTemperature bool   `mapstructure:"report_temperature" yaml:"report_temperature" json:"reportTemperature"` // 是否报告温度
	GPUBackend        string `mapstructure:"gpu_backend" yaml:"gpu_backend" json:"gpuBackend"`                      // GPU后端: auto/nvidia/amd/intel
}

// NodeExecutorConfig contains command executor configuration
type NodeExecutorConfig struct {
	MaxConcurrent   int      `mapstructure:"max_concurrent" yaml:"max_concurrent" json:"maxConcurrent"`         // 最大并发任务数
	TaskTimeout     int      `mapstructure:"task_timeout" yaml:"task_timeout" json:"taskTimeout"`               // 任务超时（秒）
	AllowRemoteStop bool     `mapstructure:"allow_remote_stop" yaml:"allow_remote_stop" json:"allowRemoteStop"` // 是否允许远程停止
	AllowedCommands []string `mapstructure:"allowed_commands" yaml:"allowed_commands" json:"allowedCommands"`   // 允许的命令白名单
}

// NodeCapabilitiesConfig contains node capabilities configuration
type NodeCapabilitiesConfig struct {
	PythonEnabled     bool              `mapstructure:"python_enabled" yaml:"python_enabled" json:"pythonEnabled"`           // 是否启用 Python 支持
	CondaPath         string            `mapstructure:"conda_path" yaml:"conda_path" json:"condaPath"`                       // Conda 可执行文件路径
	CondaEnvironments map[string]string `mapstructure:"conda_environments" yaml:"conda_environments" json:"condaEnvironments"` // Conda 环境列表 (name -> path)
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
	Timeout  int `mapstructure:"timeout" yaml:"timeout" json:"timeout"`    // seconds
}

// CondaEnvConfig contains conda environment configuration
type CondaEnvConfig struct {
	Enabled      bool              `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	CondaPath    string            `mapstructure:"conda_path" yaml:"conda_path" json:"condaPath"`
	Environments map[string]string `mapstructure:"environments" yaml:"environments" json:"environments"` // name -> path
}

// ModelConfigEntry represents a model configuration entry in models.json
type ModelConfigEntry struct {
	ModelID   string `json:"modelId"`
	Path      string `json:"path,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Alias     string `json:"alias,omitempty"`
	Favourite bool   `json:"favourite"`
	// 分卷模型相关字段
	TotalSize    int64             `json:"totalSize,omitempty"`  // 所有分卷的总大小
	ShardCount   int               `json:"shardCount,omitempty"` // 分卷数量
	ShardFiles   []string          `json:"shardFiles,omitempty"` // 所有分卷文件路径
	PrimaryModel *PrimaryModelInfo `json:"primaryModel,omitempty"`
	Mmproj       *MmprojInfo       `json:"mmproj,omitempty"`
}

// PrimaryModelInfo contains information about the primary model
type PrimaryModelInfo struct {
	FileName        string `json:"fileName"`
	Name            string `json:"name,omitempty"`
	Architecture    string `json:"architecture,omitempty"`
	ContextLength   int    `json:"contextLength,omitempty"`
	EmbeddingLength int    `json:"embeddingLength,omitempty"`
}

// MmprojInfo contains information about the multimodal projector
type MmprojInfo struct {
	FileName     string `json:"fileName"`
	Size         int64  `json:"size,omitempty"` // mmproj 文件大小（字节）
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

	// 🔧 FIX: 在测试环境中使用空路径,避免扫描模型文件导致超时
	var modelPaths []string
	autoScan := true
	if testing.Testing() {
		// 测试环境:使用空路径,禁用自动扫描
		modelPaths = []string{}
		autoScan = false
	} else {
		// 生产环境:使用默认路径,启用自动扫描
		modelPaths = []string{
			filepath.Join(cwd, "models"),
			filepath.Join(os.Getenv("HOME"), ".cache/huggingface/hub"),
		}
		autoScan = true
	}

	return &Config{
		Mode: "standalone", // 默认单机模式
		Server: ServerConfig{
			WebPort:       9190,
			AnthropicPort: 9170,
			OllamaPort:    11434,
			LMStudioPort:  1234,
			Host:          "0.0.0.0",
			ReadTimeout:   60,
			WriteTimeout:  60,
		},
		Model: ModelConfig{
			Paths:        modelPaths,
			AutoScan:     autoScan,
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
		Storage: storage.StorageConfig{
			Type: storage.StorageTypeMemory,
			SQLite: &storage.SQLiteConfig{
				Path:      filepath.Join(cwd, "Shepherd", "data", "shepherd.db"),
				EnableWAL: true,
				Pragmas: map[string]string{
					"cache_size":  "-64000", // 64MB cache
					"synchronous": "NORMAL",
				},
			},
		},
		Master: MasterConfig{
			Enabled:         false,
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
				FlushInterval: 10,          // 10 seconds
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
				CondaPath: "",
				Environments: map[string]string{
					"shepherd": "",
				},
			},
			RegisterRetry:     3,
			HeartbeatInterval: 5,
			HeartbeatTimeout:  15,
		},
		// Node 节点配置
		Node: NodeConfig{
			ID:       "auto",
			Name:     "",
			Role:     "standalone",
			Tags:     []string{},
			Metadata: map[string]string{
				"os":   "linux",
				"arch": "amd64",
			},
			MasterRole: NodeMasterRoleConfig{
				Enabled: false,
				Port:    9190,
				APIKey:  "",
				SSL: NodeSSLConfig{
					Enabled:  false,
					CertPath: "",
					KeyPath:  "",
				},
			},
			ClientRole: NodeClientRoleConfig{
				Enabled:           false,
				MasterAddress:     "",
				RegisterRetry:     3,
				HeartbeatInterval: 5,
				HeartbeatTimeout:  15,
			},
			Resources: NodeResourceConfig{
				MonitorInterval:   5,
				ReportGPU:         true,
				ReportTemperature: true,
				GPUBackend:        "auto",
			},
			Executor: NodeExecutorConfig{
				MaxConcurrent:   4,
				TaskTimeout:     3600,
				AllowRemoteStop: true,
				AllowedCommands: []string{
					"load_model",
					"unload_model",
					"run_llamacpp",
					"stop_process",
					"scan_models",
					"collect_logs",
				},
			},
			Capabilities: NodeCapabilitiesConfig{
				PythonEnabled:     false,
				CondaPath:         "",
				CondaEnvironments: map[string]string{
					"shepherd": "",
				},
			},
		},
		ModelRepo: ModelRepoConfig{
			Endpoint: "huggingface.co",
			Token:    "",
			Timeout:  30,
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

// syncLegacyConfig 将旧的 Client/Master 配置同步到新的 Node 配置
// 此方法确保向后兼容，旧配置会自动同步到新配置
func (c *Config) syncLegacyConfig() {
	// 如果 Node.Role 为空，尝试从 mode 字段同步，或默认为 standalone
	if c.Node.Role == "" {
		if c.Mode != "" {
			c.Node.Role = c.Mode
		} else {
			// 如果 mode 也是空，默认为 standalone
			c.Node.Role = "standalone"
			c.Mode = "standalone"
		}
	}

	// 同步 Client 配置 -> Node.ClientRole 和 Node.Capabilities
	if c.Client.Enabled {
		// 同步 ClientRole 配置
		if !c.Node.ClientRole.Enabled {
			c.Node.ClientRole.Enabled = true
		}
		if c.Client.MasterAddress != "" && c.Node.ClientRole.MasterAddress == "" {
			c.Node.ClientRole.MasterAddress = c.Client.MasterAddress
		}
		if c.Client.RegisterRetry != 0 && c.Node.ClientRole.RegisterRetry == 0 {
			c.Node.ClientRole.RegisterRetry = c.Client.RegisterRetry
		}
		// 优先使用较新的心跳间隔配置（heartbeat_interval 字段）
		if c.Client.HeartbeatInterval != 0 && c.Node.ClientRole.HeartbeatInterval == 0 {
			c.Node.ClientRole.HeartbeatInterval = c.Client.HeartbeatInterval
		}
		if c.Client.HeartbeatTimeout != 0 && c.Node.ClientRole.HeartbeatTimeout == 0 {
			c.Node.ClientRole.HeartbeatTimeout = c.Client.HeartbeatTimeout
		}
		// 如果 heartbeat 块中的值更大，优先使用
		if c.Client.Heartbeat.Interval > c.Node.ClientRole.HeartbeatInterval {
			c.Node.ClientRole.HeartbeatInterval = c.Client.Heartbeat.Interval
		}
		if c.Client.Heartbeat.Timeout > c.Node.ClientRole.HeartbeatTimeout {
			c.Node.ClientRole.HeartbeatTimeout = c.Client.Heartbeat.Timeout
		}

		// 同步 ClientInfo -> Node
		if len(c.Client.ClientInfo.Tags) > 0 && len(c.Node.Tags) == 0 {
			c.Node.Tags = c.Client.ClientInfo.Tags
		}
		if len(c.Client.ClientInfo.Metadata) > 0 && len(c.Node.Metadata) == 0 {
			// 合并 metadata，保留系统信息
			if c.Node.Metadata == nil {
				c.Node.Metadata = make(map[string]string)
			}
			for k, v := range c.Client.ClientInfo.Metadata {
				c.Node.Metadata[k] = v
			}
		}

		// 同步 CondaEnv -> Node.Capabilities
		if c.Client.CondaEnv.Enabled && !c.Node.Capabilities.PythonEnabled {
			c.Node.Capabilities.PythonEnabled = true
			c.Node.Capabilities.CondaPath = c.Client.CondaEnv.CondaPath
			if c.Node.Capabilities.CondaEnvironments == nil {
				c.Node.Capabilities.CondaEnvironments = make(map[string]string)
			}
			for name, path := range c.Client.CondaEnv.Environments {
				c.Node.Capabilities.CondaEnvironments[name] = path
			}
		}
	}

	// 同步 Master 配置 -> Node.MasterRole
	if c.Master.Enabled && !c.Node.MasterRole.Enabled {
		c.Node.MasterRole.Enabled = true
		// Scheduler 配置保留，但不再直接映射到 Node 配置
		// Scheduler 是 Master 专用配置，将在运行时从 Master 读取
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// 同步旧配置到新配置（向后兼容）
	c.syncLegacyConfig()

	// Validate mode
	if c.Mode == "" {
		c.Mode = "standalone"
	}
	validModes := map[string]bool{"standalone": true, "hybrid": true, "master": true, "client": true}
	if !validModes[c.Mode] {
		return fmt.Errorf("invalid mode: %s (must be standalone, hybrid, master, or client)", c.Mode)
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

	// 验证 Node 配置
	if err := c.validateNodeConfig(); err != nil {
		return err
	}

	return nil
}

// validateNodeConfig validates the Node configuration
func (c *Config) validateNodeConfig() error {
	// 验证节点角色
	validRoles := map[string]bool{"standalone": true, "master": true, "client": true, "hybrid": true}
	if !validRoles[c.Node.Role] {
		return fmt.Errorf("invalid node role: %s (must be standalone, master, client, or hybrid)", c.Node.Role)
	}

	// 验证 MasterRole 配置
	if c.Node.MasterRole.Enabled {
		if c.Node.MasterRole.Port < 1 || c.Node.MasterRole.Port > 65535 {
			return fmt.Errorf("invalid master role port: %d", c.Node.MasterRole.Port)
		}
		if c.Node.MasterRole.SSL.Enabled {
			if c.Node.MasterRole.SSL.CertPath == "" {
				return fmt.Errorf("SSL enabled but cert path is empty")
			}
			if c.Node.MasterRole.SSL.KeyPath == "" {
				return fmt.Errorf("SSL enabled but key path is empty")
			}
		}
	}

	// 验证 ClientRole 配置
	if c.Node.ClientRole.Enabled {
		if c.Node.ClientRole.MasterAddress == "" {
			return fmt.Errorf("client role enabled but master address is empty")
		}
		if c.Node.ClientRole.HeartbeatInterval < 1 {
			return fmt.Errorf("heartbeat interval must be at least 1 second")
		}
		if c.Node.ClientRole.HeartbeatTimeout < c.Node.ClientRole.HeartbeatInterval {
			return fmt.Errorf("heartbeat timeout must be greater than heartbeat interval")
		}
	}

	// 验证 Resource 配置
	if c.Node.Resources.MonitorInterval < 1 {
		return fmt.Errorf("resource monitor interval must be at least 1 second")
	}
	validGPUBackends := map[string]bool{"auto": true, "nvidia": true, "amd": true, "intel": true, "": true}
	if !validGPUBackends[c.Node.Resources.GPUBackend] {
		return fmt.Errorf("invalid GPU backend: %s (must be auto, nvidia, amd, or intel)", c.Node.Resources.GPUBackend)
	}

	// 验证 Executor 配置
	if c.Node.Executor.MaxConcurrent < 1 {
		return fmt.Errorf("executor max concurrent must be at least 1")
	}
	if c.Node.Executor.TaskTimeout < 1 {
		return fmt.Errorf("executor task timeout must be at least 1 second")
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
	config           *Config
	configPath       string
	modelsConfigPath string
	launchConfigPath string
	mu               sync.RWMutex
	cachedModels     []ModelConfigEntry
	cachedModelsTime int64
	mode             string // 运行模式
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

// NewManagerWithPath creates a new configuration manager with a custom config path
func NewManagerWithPath(mode, configPath string) *Manager {
	configDir := filepath.Dir(configPath)
	// models.json 始终存储在 config/node/ 目录下，与主配置文件位置无关
	modelsDir := filepath.Join(GetConfigDir(), "node")
	return &Manager{
		configPath:       configPath,
		modelsConfigPath: filepath.Join(modelsDir, "models.json"),
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
