// Package server provides the HTTP server for the Shepherd application.
// It handles HTTP requests, routing, middleware, and serves the web UI.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	api "github.com/shepherd-project/shepherd/Shepherd/internal/api"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/anthropic"
	compatibilityapi "github.com/shepherd-project/shepherd/Shepherd/internal/api/compatibility"
	filesystemapi "github.com/shepherd-project/shepherd/Shepherd/internal/api/filesystem"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/ollama"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/openai"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/paths"
	storageapi "github.com/shepherd-project/shepherd/Shepherd/internal/api/storage"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/model"
	modelrepoclient "github.com/shepherd-project/shepherd/Shepherd/internal/modelrepo"
	"github.com/shepherd-project/shepherd/Shepherd/internal/port"
	"github.com/shepherd-project/shepherd/Shepherd/internal/storage"
	"github.com/shepherd-project/shepherd/Shepherd/internal/websocket"
)

// ModelDTO represents a model for API responses
type ModelDTO struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Alias       string                 `json:"alias"`
	Path        string                 `json:"path"`
	PathPrefix  string                 `json:"pathPrefix"`
	Size        int64                  `json:"size"`
	TotalSize   int64                  `json:"totalSize,omitempty"`  // 包含所有分卷的总大小
	ShardCount  int                    `json:"shardCount,omitempty"` // 分卷数量
	ShardFiles  []string               `json:"shardFiles,omitempty"` // 所有分卷文件路径
	MmprojPath  string                 `json:"mmprojPath,omitempty"` // mmproj 文件路径
	Favourite   bool                   `json:"favourite"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      string                 `json:"status"`
	IsLoaded    bool                   `json:"isLoaded"`
	ScannedAt   string                 `json:"scannedAt,omitempty"` // 扫描时间（ISO 8601 格式）
}

// nonEmptyString 返回非空字符串，如果是空字符串则返回 nil
// 这样 JSON 序列化时会省略该字段而不是返回空字符串
func nonEmptyString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// Server represents the HTTP server
type Server struct {
	engine      *gin.Engine
	httpServer  *http.Server
	config      *Config
	handlers    *Handlers
	wsMgr       *websocket.Manager
	modelMgr    *model.Manager
	storageMgr  *storage.Manager
	downloadMgr *DownloadManager        // 下载管理器
	nodeAdapter *api.NodeAdapter        // Node API 适配器
	repoClient  *modelrepoclient.Client // 模型仓库客户端

	// 新增字段：WebSocket Hub 和端口管理器
	wsHub         *WebSocketHub
	portAllocator *port.PortAllocator

	// 模型能力存储
	capabilities   map[string]*ModelCapabilities // modelId -> capabilities
	capabilitiesMu sync.RWMutex

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ModelCapabilities 表示模型能力配置
type ModelCapabilities struct {
	Thinking  bool `json:"thinking"`  // 思考能力（如 DeepSeek-R1）
	Tools     bool `json:"tools"`     // 工具使用/函数调用
	Rerank    bool `json:"rerank"`    // 重排序能力
	Embedding bool `json:"embedding"` // 嵌入向量生成
}

// Config contains server configuration
type Config struct {
	WebPort       int
	AnthropicPort int
	OllamaPort    int
	LMStudioPort  int
	Host          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	WebUIPath     string
	// Mode and ServerMode for runtime configuration
	Mode      string // standalone|master|client
	ServerCfg *config.Config
	ConfigMgr *config.Manager // 配置管理器
}

// Handlers contains handler instances
type Handlers struct {
	OpenAI        *openai.Handler
	Ollama        *ollama.Handler
	Anthropic     *anthropic.Handler
	Paths         *paths.Handler
	Storage       *storageapi.Handler
	Compatibility *compatibilityapi.Handler
	Filesystem    *filesystemapi.Handler
}

// NewServer creates a new HTTP server
func NewServer(config *Config, modelMgr *model.Manager) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		handlers: &Handlers{},
		modelMgr: modelMgr,
	}

	// Initialize storage manager
	storageMgr, err := storage.NewManager(&config.ServerCfg.Storage)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize storage manager: %w", err)
	}
	s.storageMgr = storageMgr

	// Create download manager
	s.downloadMgr = NewDownloadManager(3) // 最多3个并发下载

	// Create WebSocket manager
	s.wsMgr = websocket.NewManager(modelMgr)

	// Create WebSocket Hub (新增)
	s.wsHub = NewWebSocketHub()

	// Create port allocator (新增)
	s.portAllocator = port.NewPortAllocator(8081, 9000)

	// Create model repository client with config
	cfg := config.ConfigMgr.Get()
	timeout := time.Duration(cfg.ModelRepo.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	s.repoClient = modelrepoclient.NewClientWithConfig(cfg.ModelRepo.Endpoint, cfg.ModelRepo.Token, timeout)

	// Create compatibility server manager
	compatServerManager := compatibilityapi.NewServerManager(modelMgr)

	// Create API handlers
	s.handlers.OpenAI = openai.NewHandler(modelMgr)
	s.handlers.Ollama = ollama.NewHandler(modelMgr)
	s.handlers.Anthropic = anthropic.NewHandler(modelMgr)
	s.handlers.Paths = paths.NewHandler(config.ConfigMgr)
	s.handlers.Storage = storageapi.NewHandler(config.ConfigMgr, storageMgr)
	s.handlers.Compatibility = compatibilityapi.NewHandler(config.ConfigMgr, compatServerManager)
	s.handlers.Filesystem = filesystemapi.NewHandler()

	// Setup Gin engine
	if config.WebUIPath == "" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	s.engine = gin.New()
	s.setupMiddleware()
	s.setupRoutes()

	return s, nil
}

// setupMiddleware configures server middleware
func (s *Server) setupMiddleware() {
	s.engine.Use(
		gin.Recovery(),
		s.corsMiddleware(),
		s.loggerMiddleware(),
	)
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// WebSocket endpoint (for SSE)
	s.engine.GET("/api/events", s.handleEvents)

	// WebSocket endpoint (新增)
	s.engine.GET("/ws", s.handleWebSocket)

	// API routes
	api := s.engine.Group("/api")
	{
		// Server info
		api.GET("/info", s.handleServerInfo)

		// System info
		api.GET("/system/gpus", s.handleGetGPUs)
		api.GET("/system/llamacpp-backends", s.handleGetLlamacppBackends)

		// Configuration routes
		config := api.Group("/config")
		{
			config.GET("", s.handleGetConfig)
			config.PUT("", s.handleUpdateConfig)

			// Llama.cpp paths
			llamacpp := config.Group("/llamacpp/paths")
			{
				llamacpp.GET("", s.handlers.Paths.GetLlamaCppPaths)
				llamacpp.POST("", s.handlers.Paths.AddLlamaCppPath)
				llamacpp.DELETE("", s.handlers.Paths.RemoveLlamaCppPath)
				llamacpp.POST("/test", s.handlers.Paths.TestLlamaCppPath)
			}

			// Model paths
			models := config.Group("/models/paths")
			{
				models.GET("", s.handlers.Paths.GetModelPaths)
				models.POST("", s.handlers.Paths.AddModelPath)
				models.PUT("", s.handlers.Paths.UpdateModelPath)
				models.DELETE("", s.handlers.Paths.RemoveModelPath)
			}

			// Storage configuration
			storage := config.Group("/storage")
			{
				storage.GET("", s.handlers.Storage.GetStorageConfig)
				storage.PUT("", s.handlers.Storage.UpdateStorageConfig)
				storage.GET("/stats", s.handlers.Storage.GetStats)
			}

			// Compatibility configuration
			compatibility := config.Group("/compatibility")
			{
				compatibility.GET("", s.handlers.Compatibility.GetCompatibility)
				compatibility.PUT("", s.handlers.Compatibility.UpdateCompatibility)
				compatibility.POST("/test", s.handlers.Compatibility.TestConnection)
			}
		}

		// Chat/Conversation routes
		conversations := api.Group("/conversations")
		{
			conversations.GET("", s.handlers.Storage.GetConversations)
			conversations.GET("/:id", s.handlers.Storage.GetConversation)
			conversations.DELETE("/:id", s.handlers.Storage.DeleteConversation)
		}

		// Model routes
		models := api.Group("/models")
		{
			models.GET("", s.handleListModels)

			// 模型能力管理（必须在 :id 路由之前）
			models.GET("/capabilities/get", s.handleGetModelCapabilities)
			models.POST("/capabilities/set", s.handleSetModelCapabilities)

			// 显存估算（必须在 :id 路由之前）
			models.POST("/vram/estimate", s.handleEstimateVRAM)

			// 模型具体操作（包含 :id 参数的路由必须在最后）
			models.GET("/:id", s.handleGetModel)
			models.POST("/:id/load", s.handleLoadModel)
			models.POST("/:id/unload", s.handleUnloadModel)
			models.PUT("/:id/alias", s.handleSetAlias)
			models.PUT("/:id/favourite", s.handleSetFavourite)
		}

		// Model scan routes
		modelScan := api.Group("/model/scan")
		{
			modelScan.POST("", s.handleScanModels)
			modelScan.GET("/status", s.handleGetScanStatus)
		}

		// Download routes
		downloads := api.Group("/downloads")
		{
			downloads.GET("", s.handleListDownloads)
			downloads.POST("", s.handleCreateDownload)
			downloads.GET("/:id", s.handleGetDownload)
			downloads.POST("/:id/pause", s.handlePauseDownload)
			downloads.POST("/:id/resume", s.handleResumeDownload)
			downloads.DELETE("/:id", s.handleDeleteDownload)
		}

		// Model repository routes (远程模型仓库文件浏览)
		// 路由格式: /api/repo/files?source=huggingface&repoId=Qwen/Qwen2-7B-Instruct
		// 使用查询参数以支持 repoId 中包含斜杠
		repo := api.Group("/repo")
		{
			repo.GET("/files", s.handleListModelFiles)
			repo.GET("/search", s.handleSearchModels)
			repo.GET("/config", s.handleGetModelRepoConfig)
			repo.PUT("/config", s.handleUpdateModelRepoConfig)
			repo.GET("/endpoints", s.handleGetAvailableEndpoints)
		}

		// Process routes
		processes := api.Group("/processes")
		{
			processes.GET("", s.handleListProcesses)
			processes.GET("/:id", s.handleGetProcess)
			processes.POST("/:id/stop", s.handleStopProcess)
		}

		// Log routes
		logs := api.Group("/logs")
		{
			logs.GET("/stream", s.handleLogStream)
			logs.GET("/entries", s.handleLogEntries)
		}

		// System routes
		system := api.Group("/system")
		{
			system.GET("/filesystem", s.handlers.Filesystem.ListDirectory)
			system.POST("/filesystem/validate", s.handlers.Filesystem.ValidatePath)
		}
	}

	// OpenAI compatible API
	openai := s.engine.Group("/v1")
	{
		openai.POST("/chat/completions", s.handleOpenAIChat)
		openai.POST("/completions", s.handleOpenAIComplete)
		openai.GET("/models", s.handleOpenAIModels)
	}

	// Anthropic compatible API
	anthropic := s.engine.Group("/v1")
	{
		anthropic.POST("/messages", s.handleAnthropicMessages)
	}

	// Ollama compatible API
	ollama := s.engine.Group("/api")
	{
		ollama.POST("/chat", s.handleOllamaChat)
		ollama.POST("/tags", s.handleOllamaTags)
	}

	// Static files for Web UI
	s.engine.Static("/assets", s.config.WebUIPath+"/assets")
	s.engine.Static("/favicon.svg", s.config.WebUIPath+"/favicon.svg")
	s.engine.StaticFile("/", s.config.WebUIPath+"/index.html")
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already started
	if s.httpServer != nil {
		return fmt.Errorf("server already started")
	}

	// Start WebSocket manager
	s.wsMgr.Start()

	// Start WebSocket Hub (新增)
	go s.wsHub.Run()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.WebPort)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	// Start server in background
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		logger.Infof("启动 HTTP 服务器，监听 %s", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("HTTP 服务器错误: %v", err)
		}
		logger.Info("HTTP 服务器已停止")
	}()

	return nil
}

// Stop stops the HTTP server gracefully
func (s *Server) Stop() error {
	s.mu.Lock()
	if s.httpServer == nil {
		s.mu.Unlock()
		return fmt.Errorf("server not started")
	}
	s.mu.Unlock()

	logger.Info("开始停止 HTTP 服务器...")

	// Step 1: Cancel context to signal all goroutines
	s.cancel()

	// Step 2: Stop accepting new connections (but don't close existing ones)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 3: Shutdown HTTP server gracefully
	s.mu.Lock()
	if s.httpServer != nil {
		logger.Info("关闭 HTTP 服务器...")
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("HTTP 服务器关闭失败: %v", err)
			// Force close if graceful shutdown fails
			s.httpServer.Close()
		} else {
			logger.Info("HTTP 服务器已优雅关闭")
		}
		s.httpServer = nil
	}
	s.mu.Unlock()

	// Step 4: Stop WebSocket manager
	logger.Info("停止 WebSocket 管理器...")
	s.wsMgr.Stop()
	logger.Info("WebSocket 管理器已停止")

	// Step 4.5: Stop download manager
	logger.Info("停止下载管理器...")
	if s.downloadMgr != nil {
		s.downloadMgr.Stop()
		logger.Info("下载管理器已停止")
	}

	// Step 5: Close storage manager
	logger.Info("关闭存储管理器...")
	if s.storageMgr != nil {
		if err := s.storageMgr.Close(); err != nil {
			logger.Errorf("存储管理器关闭失败: %v", err)
		} else {
			logger.Info("存储管理器已关闭")
		}
	}

	// Step 6: Wait for all goroutines to finish
	logger.Info("等待所有协程完成...")
	s.wg.Wait()
	logger.Info("所有协程已完成")

	return nil
}

// Shutdown performs graceful shutdown with context
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("开始优雅关闭...")

	// Create a channel for shutdown completion
	done := make(chan error, 1)

	go func() {
		done <- s.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.Errorf("优雅关闭失败: %v", err)
			return err
		}
		logger.Info("优雅关闭完成")
		return nil
	case <-ctx.Done():
		logger.Warn("优雅关闭超时，强制退出")
		// Force stop
		s.mu.Lock()
		if s.httpServer != nil {
			s.httpServer.Close()
			s.httpServer = nil
		}
		s.mu.Unlock()
		return ctx.Err()
	}
}

// GetEngine returns the Gin engine (for testing)
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// GetWebSocketManager returns the WebSocket manager
func (s *Server) GetWebSocketManager() *websocket.Manager {
	return s.wsMgr
}

// RegisterMasterHandler 注册 Master Handler（已废弃）
// Deprecated: 请使用 RegisterNodeAdapter 代替
func (s *Server) RegisterMasterHandler(handler interface{}) {
	logger.Warn("RegisterMasterHandler 已废弃，请使用 RegisterNodeAdapter 代替")
}

// RegisterNodeAdapter 注册 Node API 适配器
func (s *Server) RegisterNodeAdapter(nodeAdapter *api.NodeAdapter) {
	s.nodeAdapter = nodeAdapter

	// 注册 Node API 路由
	api := s.engine.Group("/api")
	nodeAdapter.RegisterRoutes(api)
	logger.Info("Node API 适配器路由已注册")
}

// Middleware

// corsMiddleware handles CORS
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// loggerMiddleware logs requests
func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logFields := map[string]interface{}{
			"method":  c.Request.Method,
			"path":    path,
			"status":  status,
			"latency": latency.String(),
		}

		if query != "" {
			logFields["query"] = query
		}

		// Log based on status code
		if status >= 500 {
			logger.WithFields(logFields).Error("请求处理失败")
		} else if status >= 400 {
			logger.WithFields(logFields).Warn("客户端错误")
		} else {
			logger.WithFields(logFields).Info("请求处理成功")
		}
	}
}

// Health check handler
func (s *Server) handleServerInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": "1.0.0",
		"name":    "Shepherd",
		"status":  "running",
		"ports": gin.H{
			"web":       s.config.WebPort,
			"anthropic": s.config.AnthropicPort,
			"ollama":    s.config.OllamaPort,
			"lmstudio":  s.config.LMStudioPort,
		},
	})
}

// handleGetGPUs 返回系统可用的 GPU 列表
// 返回格式兼容 LlamacppServer 的设备列表格式
func (s *Server) handleGetGPUs(c *gin.Context) {
	// 首先尝试使用 llama-bench 获取设备列表（与 LlamacppServer 一致）
	llamacppBinPath := ""
	if s.config != nil && s.config.ServerCfg != nil && len(s.config.ServerCfg.Llamacpp.Paths) > 0 {
		// 使用第一个可用的 llama.cpp 路径
		for _, p := range s.config.ServerCfg.Llamacpp.Paths {
			if fileInfo, err := os.Stat(p.Path); err == nil && fileInfo.IsDir() {
				llamacppBinPath = p.Path
				break
			}
		}
	}

	deviceStrings := []string{} // 简单设备描述字符串（兼容 LlamacppServer）
	gpus := []gin.H{}           // 详细 GPU 信息（Shepherd 扩展）

	if llamacppBinPath != "" {
		// 尝试使用 llama-bench 获取设备列表
		benchPath := llamacppBinPath + "/llama-bench"
		cmd := exec.Command(benchPath, "--list-devices")
		output, err := cmd.CombinedOutput()
		if err == nil {
			// 解析 llama-bench 输出
			lines := strings.Split(string(output), "\n")
			inDeviceList := false
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if strings.Contains(line, "Available devices") {
					inDeviceList = true
					continue
				}
				if inDeviceList {
					// 解析设备行，例如: "ROCm0: AMD Radeon Graphics (122880 MiB, 115050 MiB free)"
					deviceStrings = append(deviceStrings, line)

					// 同时提取详细信息
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						deviceID := strings.TrimSpace(parts[0])
						description := strings.TrimSpace(parts[1])

						// 提取内存信息
						var totalMemory, freeMemory string
						// 使用正则表达式提取内存信息
						memRe := regexp.MustCompile(`\((\d+) MiB(?:, (\d+) MiB free)?\)`)
						if memMatches := memRe.FindStringSubmatch(description); len(memMatches) > 0 {
							totalMemory = memMatches[1] + " MiB"
							if len(memMatches) > 2 && memMatches[2] != "" {
								freeMemory = memMatches[2] + " MiB"
							}
						}

						gpus = append(gpus, gin.H{
							"id":          deviceID,
							"name":        description,
							"totalMemory": totalMemory,
							"freeMemory":  freeMemory,
							"available":   true,
						})
					}
				}
			}
		}
	}

	// 如果 llama-bench 失败，回退到 roc-smi
	if len(deviceStrings) == 0 {
		cmd := exec.Command("roc-smi", "--json")
		output, err := cmd.Output()
		if err == nil {
			var rocmOutput []map[string]interface{}
			if err := json.Unmarshal(output, &rocmOutput); err == nil {
				for i, gpu := range rocmOutput {
					// 提取 GPU 关键信息
					gpuName := ""
					if name, ok := gpu["card_name"].(string); ok {
						gpuName = name
					} else if name, ok := gpu["Card name"].(string); ok {
						gpuName = name
					}

					// 获取架构信息
					architecture := ""
					if arch, ok := gpu["card_series"].(string); ok {
						architecture = arch
					} else if arch, ok := gpu["Card series"].(string); ok {
						architecture = arch
					}

					// 获取 VRAM
					var totalMemory, freeMemory string
					if vram, ok := gpu["vram_total_memory"].(float64); ok {
						totalMemory = fmt.Sprintf("%.0f MiB", vram/(1024*1024))
					}
					if vramFree, ok := gpu["vram_total_free_memory"].(float64); ok {
						freeMemory = fmt.Sprintf("%.0f MiB", vramFree/(1024*1024))
					}

					// 构建 device string（类似 llama-bench 格式）
					deviceID := fmt.Sprintf("ROCm%d", i)
					deviceString := fmt.Sprintf("%s: %s", deviceID, gpuName)
					if totalMemory != "" {
						deviceString += fmt.Sprintf(" (%s", totalMemory)
						if freeMemory != "" {
							deviceString += fmt.Sprintf(", %s free", freeMemory)
						}
						deviceString += ")"
					}
					deviceStrings = append(deviceStrings, deviceString)

					gpus = append(gpus, gin.H{
						"id":           deviceID,
						"name":         gpuName,
						"architecture": architecture,
						"totalMemory":  totalMemory,
						"freeMemory":   freeMemory,
						"available":    true,
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": deviceStrings, // 简单设备字符串列表（兼容 LlamacppServer）
		"gpus":    gpus,          // 详细 GPU 信息（Shepherd 扩展）
		"count":   len(gpus),
	})
}

// handleGetLlamacppBackends 返回可用的 llama.cpp 后端列表
func (s *Server) handleGetLlamacppBackends(c *gin.Context) {
	backends := []gin.H{}

	// 从配置中获取 llama.cpp 路径
	if s.config != nil && s.config.ServerCfg != nil {
		paths := s.config.ServerCfg.Llamacpp.Paths
		for _, p := range paths {
			// 检查路径是否存在
			available := false
			if fileInfo, err := os.Stat(p.Path); err == nil {
				// 检查是否是目录
				available = fileInfo.IsDir()
			}

			backends = append(backends, gin.H{
				"path":        p.Path,
				"name":        p.Name,
				"description": p.Description,
				"available":   available,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"backends": backends,
		"count":    len(backends),
	})
}

// handleEstimateVRAM 估算模型显存需求
func (s *Server) handleEstimateVRAM(c *gin.Context) {
	var req struct {
		ModelID        string `json:"modelId"`
		LlamaBinPath   string `json:"llamaBinPath"`
		CtxSize        int    `json:"ctxSize"`
		BatchSize      int    `json:"batchSize"`
		UBatchSize     int    `json:"uBatchSize"`
		Parallel       int    `json:"parallel"`
		FlashAttention bool   `json:"flashAttention"`
		KVUnified      bool   `json:"kvUnified"`
		CacheTypeK     string `json:"cacheTypeK"`
		CacheTypeV     string `json:"cacheTypeV"`
		ExtraParams    string `json:"extraParams"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求: " + err.Error()})
		return
	}

	// 验证必需参数
	if req.ModelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 modelId 参数"})
		return
	}
	if req.LlamaBinPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 llamaBinPath 参数"})
		return
	}

	// 从模型管理器获取模型信息
	model, exists := s.modelMgr.GetModel(req.ModelID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在: " + req.ModelID})
		return
	}

	// 构建模型文件路径
	var modelPath string
	if model.ShardCount > 0 && len(model.ShardFiles) > 0 {
		// 分卷模型，使用主模型文件
		modelPath = model.ShardFiles[0]
	} else {
		modelPath = model.Path
	}

	// 构建 llama-fit-params 命令
	args := []string{
		"--model", modelPath,
	}

	// 添加支持的参数
	if req.CtxSize > 0 {
		args = append(args, "--ctx-size", fmt.Sprintf("%d", req.CtxSize))
	}
	if req.BatchSize > 0 {
		args = append(args, "--batch-size", fmt.Sprintf("%d", req.BatchSize))
	}
	if req.UBatchSize > 0 {
		args = append(args, "--ubatch-size", fmt.Sprintf("%d", req.UBatchSize))
	}
	if req.Parallel > 0 {
		args = append(args, "--parallel", fmt.Sprintf("%d", req.Parallel))
	}
	if req.FlashAttention {
		args = append(args, "--flash-attn", "1")
	}
	if req.KVUnified {
		args = append(args, "--kv-unified", "1")
	}
	if req.CacheTypeK != "" {
		args = append(args, "--cache-type-k", req.CacheTypeK)
	}
	if req.CacheTypeV != "" {
		args = append(args, "--cache-type-v", req.CacheTypeV)
	}

	// 构建完整命令
	cmdPath := filepath.Join(req.LlamaBinPath, "llama-fit-params")
	cmd := exec.Command(cmdPath, args...)

	// 执行命令（设置30秒超时）
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// 检查是否有部分输出
		logger.Error("llama-fit-params 执行失败", "error", err.Error(), "output", outputStr)

		// 尝试从错误输出中提取错误信息
		errorMsg := "估算失败"
		if strings.Contains(outputStr, "llama_model_load") || strings.Contains(outputStr, "failed to load model") {
			errorMsg = "模型加载失败，请检查模型文件是否正确"
		} else if strings.Contains(outputStr, "llama_params_fit") {
			errorMsg = "参数拟合失败，请检查参数是否有效"
		}

		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   errorMsg,
			"details": outputStr,
		})
		return
	}

	// 解析输出，提取显存估算值
	vramMB := 0

	// 匹配格式: "llama_params_fit_impl: projected to use XXX MiB of device memory"
	vramRe := regexp.MustCompile(`llama_params_fit_impl: projected to use (\d+) MiB`)
	if matches := vramRe.FindStringSubmatch(outputStr); len(matches) > 1 {
		vramMB, _ = strconv.Atoi(matches[1])
	}

	// 构建响应
	result := gin.H{
		"success": vramMB > 0,
	}

	if vramMB > 0 {
		result["vram"] = fmt.Sprintf("%d", vramMB)
		result["vramMB"] = vramMB
		result["vramGB"] = fmt.Sprintf("%.2f", float64(vramMB)/1024)
	} else {
		// 如果没有找到显存值，检查是否有错误信息
		errorRe := regexp.MustCompile(`llama_init_from_model.*`)
		if errorMatch := errorRe.FindString(outputStr); errorMatch != "" {
			result["error"] = strings.TrimSpace(errorMatch)
		} else {
			result["error"] = "无法解析显存估算结果"
		}
		result["details"] = outputStr
	}

	c.JSON(http.StatusOK, gin.H{
		"success": vramMB > 0,
		"data":    result,
	})
}

// handleGetConfig 返回当前配置（不包含敏感信息）
func (s *Server) handleGetConfig(c *gin.Context) {
	if s.config == nil || s.config.ServerCfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "配置未初始化",
		})
		return
	}

	cfg := s.config.ServerCfg

	c.JSON(http.StatusOK, gin.H{
		"mode": s.config.Mode,
		"server": gin.H{
			"host":           s.config.Host,
			"web_port":       s.config.WebPort,
			"anthropic_port": s.config.AnthropicPort,
			"ollama_port":    s.config.OllamaPort,
			"lm_studio_port": s.config.LMStudioPort,
		},
		"storage": gin.H{
			"type":   cfg.Storage.Type,
			"sqlite": cfg.Storage.SQLite,
		},
		"models": gin.H{
			"paths":     cfg.Model.Paths,
			"auto_scan": cfg.Model.AutoScan,
		},
		"node": gin.H{
			"role": cfg.Node.Role,
			"id":   cfg.Node.ID,
			"name": cfg.Node.Name,
		},
		"llamacpp": gin.H{
			"paths": cfg.Llamacpp.Paths,
		},
	})
}

// handleUpdateConfig 更新配置
func (s *Server) handleUpdateConfig(c *gin.Context) {
	var req struct {
		Mode      string   `json:"mode"`
		WebPort   int      `json:"web_port"`
		AutoScan  bool     `json:"auto_scan"`
		ScanPaths []string `json:"scan_paths"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	restartRequired := false

	// 更新模式
	if req.Mode != "" {
		s.config.Mode = req.Mode
	}

	// 更新端口（需要重启）
	if req.WebPort > 0 && req.WebPort != s.config.WebPort {
		s.config.WebPort = req.WebPort
		restartRequired = true
	}

	// 更新扫描路径
	if req.ScanPaths != nil && len(req.ScanPaths) > 0 {
		s.config.ServerCfg.Model.Paths = req.ScanPaths

		// 触发重新扫描
		if req.AutoScan {
			go s.modelMgr.Scan(c.Request.Context())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "配置更新成功",
		"restart_required": restartRequired,
	})
}

func (s *Server) handleListModels(c *gin.Context) {
	models := s.modelMgr.ListModels()
	statuses := s.modelMgr.ListStatus()

	fmt.Printf("[DEBUG] handleListModels: 总共 %d 个模型\n", len(models))

	var dtos []ModelDTO
	for _, m := range models {
		// 调试日志：检查分卷信息
		if m.ShardCount > 0 {
			fmt.Printf("[DEBUG] 分卷模型: %s, ShardCount=%d, TotalSize=%d, Files=%d\n",
				m.Name, m.ShardCount, m.TotalSize, len(m.ShardFiles))
		}

		dto := ModelDTO{
			ID:          m.ID,
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Alias:       m.Alias,
			Path:        m.Path,
			PathPrefix:  m.PathPrefix,
			Size:        m.Size,
			Favourite:   m.Favourite,
			Status:      "stopped",
			IsLoaded:    false,
		}

		// 添加分卷信息
		if m.ShardCount > 0 {
			dto.ShardCount = m.ShardCount
			dto.TotalSize = m.TotalSize
			dto.ShardFiles = m.ShardFiles
		}

		// 添加 mmproj 路径
		if m.MmprojPath != "" {
			dto.MmprojPath = m.MmprojPath
		}

		// 添加扫描时间（处理零值情况）
		if !m.ScannedAt.IsZero() {
			dto.ScannedAt = m.ScannedAt.Format(time.RFC3339)
		}

		// Convert metadata - 包含所有 gguf-parser-go 提供的字段
		if m.Metadata != nil {
			metadata := map[string]interface{}{
				"name":            m.Metadata.Name,
				"architecture":    m.Metadata.Architecture,
				"quantization":    m.Metadata.Quantization,
				"contextLength":   m.Metadata.ContextLength,
				"embeddingLength": m.Metadata.EmbeddingLength,
				"layerCount":      m.Metadata.BlockSize,
				"headCount":       m.Metadata.HeadCount,
				// 新增的 gguf-parser-go 字段
				"type":                nonEmptyString(m.Metadata.Type),
				"author":              nonEmptyString(m.Metadata.Author),
				"url":                 nonEmptyString(m.Metadata.URL),
				"description":         nonEmptyString(m.Metadata.Description),
				"license":             nonEmptyString(m.Metadata.License),
				"fileType":            m.Metadata.FileType,
				"fileTypeDescriptor":  nonEmptyString(m.Metadata.FileTypeDescriptor),
				"quantizationVersion": m.Metadata.QuantizationVersion,
				"parameters":          m.Metadata.Parameters,
				"bitsPerWeight":       m.Metadata.BitsPerWeight,
				"alignment":           m.Metadata.Alignment,
				"fileSize":            m.Metadata.FileSize,
				"modelSize":           m.Metadata.ModelSize,
			}
			dto.Metadata = metadata
		}

		// Add status info
		if status, ok := statuses[m.ID]; ok {
			dto.Status = status.State.String()
			dto.IsLoaded = status.State == model.StateLoaded
		}

		dtos = append(dtos, dto)
	}

	c.JSON(http.StatusOK, gin.H{"models": dtos, "total": len(dtos)})
}

func (s *Server) handleGetModel(c *gin.Context) {
	id := c.Param("id")
	model, exists := s.modelMgr.GetModel(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"model": model})
}

func (s *Server) handleLoadModel(c *gin.Context) {
	id := c.Param("id")

	// 检查是否异步加载（查询参数 async=true）
	asyncMode := c.DefaultQuery("async", "true") == "true"

	var req model.LoadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用默认值
		req = model.LoadRequest{
			ModelID: id,
			CtxSize: 4096,
		}
	} else {
		req.ModelID = id
	}

	var result *model.LoadResult
	var err error

	if asyncMode {
		// 异步加载
		result, err = s.modelMgr.LoadAsync(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 立即返回异步响应
		if result.Loading {
			c.JSON(http.StatusOK, gin.H{
				"message":  "模型正在加载中",
				"model_id": result.ModelID,
				"async":    true,
				"status":   "loading",
			})
			return
		}

		if result.AlreadyLoaded {
			// 模型已加载
			c.JSON(http.StatusOK, gin.H{
				"message":  "模型已加载",
				"model_id": result.ModelID,
				"port":     result.Port,
				"status":   "loaded",
			})
			return
		}
	} else {
		// 同步加载（旧模式，保持兼容性）
		result, err = s.modelMgr.Load(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if !result.Success {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "模型加载成功",
		"model_id": result.ModelID,
		"port":     result.Port,
		"ctx_size": result.CtxSize,
	})
}

func (s *Server) handleUnloadModel(c *gin.Context) {
	id := c.Param("id")
	if err := s.modelMgr.Unload(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "模型卸载成功"})
}

func (s *Server) handleSetAlias(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Alias string `json:"alias"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	if err := s.modelMgr.SetAlias(id, req.Alias); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "别名设置成功"})
}

func (s *Server) handleSetFavourite(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Favourite bool `json:"favourite"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	if err := s.modelMgr.SetFavourite(id, req.Favourite); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "收藏设置成功"})
}

// handleGetModelCapabilities 获取模型能力配置
func (s *Server) handleGetModelCapabilities(c *gin.Context) {
	modelID := c.Query("modelId")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 modelId 参数"})
		return
	}

	s.capabilitiesMu.RLock()
	caps, exists := s.capabilities[modelID]
	s.capabilitiesMu.RUnlock()

	if !exists {
		// 如果没有配置过，返回默认值（全部为 false）
		c.JSON(http.StatusOK, gin.H{
			"modelId":      modelID,
			"capabilities": &ModelCapabilities{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modelId":      modelID,
		"capabilities": caps,
		"success":      true,
	})
}

// handleSetModelCapabilities 设置模型能力配置
func (s *Server) handleSetModelCapabilities(c *gin.Context) {
	var req struct {
		ModelID      string             `json:"modelId"`
		Capabilities *ModelCapabilities `json:"capabilities"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求: " + err.Error()})
		return
	}

	if req.ModelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 modelId"})
		return
	}

	if req.Capabilities == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 capabilities"})
		return
	}

	// 应用约束规则：rerank 和 embedding 互斥
	if req.Capabilities.Rerank && req.Capabilities.Embedding {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rerank 和 embedding 不能同时启用"})
		return
	}

	// 如果启用了 rerank 或 embedding，则禁用 thinking 和 tools
	if req.Capabilities.Rerank || req.Capabilities.Embedding {
		req.Capabilities.Thinking = false
		req.Capabilities.Tools = false
	}

	// 保存到内存存储
	s.capabilitiesMu.Lock()
	s.capabilities[req.ModelID] = req.Capabilities
	s.capabilitiesMu.Unlock()

	logger.Info("模型能力已更新", "modelId", req.ModelID,
		"thinking", req.Capabilities.Thinking,
		"tools", req.Capabilities.Tools,
		"rerank", req.Capabilities.Rerank,
		"embedding", req.Capabilities.Embedding)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模型能力已保存",
	})
}

func (s *Server) handleScanModels(c *gin.Context) {
	result, err := s.modelMgr.Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "扫描完成",
		"models_found": len(result.Models),
		"errors":       len(result.Errors),
		"duration_ms":  result.Duration.Milliseconds(),
		"models":       result.Models,
		"scan_errors":  result.Errors,
	})
}

func (s *Server) handleGetScanStatus(c *gin.Context) {
	status := s.modelMgr.GetScanStatus()
	c.JSON(http.StatusOK, gin.H{
		"scanning":     status.Scanning,
		"progress":     status.Progress,
		"current_path": status.CurrentPath,
		"started_at":   status.StartedAt,
		"errors":       status.Errors,
	})
}

func (s *Server) handleListDownloads(c *gin.Context) {
	downloads := s.downloadMgr.ListDownloads()
	c.JSON(http.StatusOK, gin.H{
		"downloads": downloads,
		"total":     len(downloads),
	})
}

func (s *Server) handleCreateDownload(c *gin.Context) {
	// 支持两种请求格式:
	// 1. 新格式: { source, repoId, fileName, path } - 用于从模型仓库下载
	// 2. 旧格式: { url, target_path } - 直接URL下载(向后兼容)

	var req struct {
		Source   modelrepoclient.Source `json:"source"`
		RepoID   string                 `json:"repoId"`
		FileName string                 `json:"fileName"`
		Path     string                 `json:"path"`

		// 旧格式参数(向后兼容)
		URL        string `json:"url"`
		TargetPath string `json:"target_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	var downloadURL string
	var targetPath string

	// 使用新格式(source + repoId)
	if req.Source != "" && req.RepoID != "" {
		// 生成下载 URL
		url, err := s.repoClient.GenerateDownloadURL(req.Source, req.RepoID, req.FileName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "生成下载URL失败",
				"message": err.Error(),
			})
			return
		}
		downloadURL = url
		targetPath = req.Path
	} else if req.URL != "" {
		// 使用旧格式(直接URL)
		downloadURL = req.URL
		targetPath = req.TargetPath
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少必要参数",
			"message": "请提供 source/repoId 或 url",
		})
		return
	}

	task, err := s.downloadMgr.CreateDownload(downloadURL, targetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建下载失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "下载任务已创建",
		"data":    task,
	})
}

func (s *Server) handleGetDownload(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "下载ID不能为空"})
		return
	}

	task, exists := s.downloadMgr.GetDownload(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "下载任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (s *Server) handlePauseDownload(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "下载ID不能为空"})
		return
	}

	if err := s.downloadMgr.PauseDownload(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "下载已暂停"})
}

func (s *Server) handleResumeDownload(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "下载ID不能为空"})
		return
	}

	if err := s.downloadMgr.ResumeDownload(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "下载已恢复"})
}

func (s *Server) handleDeleteDownload(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "下载ID不能为空"})
		return
	}

	if err := s.downloadMgr.DeleteDownload(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "下载任务已删除"})
}

// handleListModelFiles handles requests to list model files from a repository
func (s *Server) handleListModelFiles(c *gin.Context) {
	// 使用查询参数而不是路径参数，以支持 repoId 中包含斜杠
	source := c.Query("source")
	repoID := c.Query("repoId")

	if source == "" || repoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少必要参数: 需要 source 和 repoId 查询参数",
		})
		return
	}

	// 目前只支持 HuggingFace
	if source != "huggingface" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "目前只支持 HuggingFace 源",
		})
		return
	}

	files, err := s.repoClient.ListGGUFFiles(repoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取文件列表失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    files,
	})
}

// handleSearchModels handles requests to search for models on HuggingFace
func (s *Server) handleSearchModels(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少必要参数: 需要 q 查询参数",
		})
		return
	}

	// Parse limit parameter (default 20)
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	result, err := s.repoClient.SearchHuggingFaceModels(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "搜索模型失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// handleGetModelRepoConfig returns the current model repository configuration
func (s *Server) handleGetModelRepoConfig(c *gin.Context) {
	cfg := s.config.ConfigMgr.Get()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"endpoint": cfg.ModelRepo.Endpoint,
			"token":    maskToken(cfg.ModelRepo.Token),
			"timeout":  cfg.ModelRepo.Timeout,
		},
	})
}

// maskToken masks the token for security
func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// handleUpdateModelRepoConfig updates the model repository configuration
func (s *Server) handleUpdateModelRepoConfig(c *gin.Context) {
	var req struct {
		Endpoint string `json:"endpoint"`
		Token    string `json:"token"`
		Timeout  int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	cfg := s.config.ConfigMgr.Get()

	// Update endpoint if provided
	if req.Endpoint != "" {
		cfg.ModelRepo.Endpoint = req.Endpoint
	}

	// Update token if provided (allow empty string to clear token)
	if req.Token != "" {
		cfg.ModelRepo.Token = req.Token
	}

	// Update timeout if provided
	if req.Timeout > 0 {
		cfg.ModelRepo.Timeout = req.Timeout
	}

	// Save config
	if err := s.config.ConfigMgr.Save(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存配置失败",
			"message": err.Error(),
		})
		return
	}

	// Update the repo client with new settings
	timeout := time.Duration(cfg.ModelRepo.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	s.repoClient = modelrepoclient.NewClientWithConfig(cfg.ModelRepo.Endpoint, cfg.ModelRepo.Token, timeout)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"endpoint": cfg.ModelRepo.Endpoint,
			"token":    maskToken(cfg.ModelRepo.Token),
			"timeout":  cfg.ModelRepo.Timeout,
		},
	})
}

// handleGetAvailableEndpoints returns available HuggingFace endpoints
func (s *Server) handleGetAvailableEndpoints(c *gin.Context) {
	endpoints := modelrepoclient.GetAvailableEndpoints()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    endpoints,
	})
}

func (s *Server) handleListProcesses(c *gin.Context) {
	processMgr := s.modelMgr.GetProcessManager()
	if processMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "进程管理器未初始化"})
		return
	}

	running, loading := processMgr.ListAll()

	// 转换为切片格式
	type ProcessInfo struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		PID     int    `json:"pid"`
		Port    int    `json:"port"`
		CtxSize int    `json:"ctx_size"`
		Running bool   `json:"running"`
		Loading bool   `json:"loading"`
	}

	var processes []ProcessInfo
	for _, p := range running {
		processes = append(processes, ProcessInfo{
			ID:      p.ID,
			Name:    p.Name,
			PID:     p.GetPID(),
			Port:    p.GetPort(),
			CtxSize: p.GetCtxSize(),
			Running: p.IsRunning(),
			Loading: false,
		})
	}
	for _, p := range loading {
		processes = append(processes, ProcessInfo{
			ID:      p.ID,
			Name:    p.Name,
			PID:     p.GetPID(),
			Port:    p.GetPort(),
			CtxSize: p.GetCtxSize(),
			Running: p.IsRunning(),
			Loading: true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"processes": processes,
		"stats": gin.H{
			"running": len(running),
			"loading": len(loading),
			"total":   len(running) + len(loading),
		},
	})
}

func (s *Server) handleGetProcess(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "进程ID不能为空"})
		return
	}

	processMgr := s.modelMgr.GetProcessManager()
	if processMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "进程管理器未初始化"})
		return
	}

	proc, exists := processMgr.Get(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "进程不存在",
			"process": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"process": gin.H{
			"id":       proc.ID,
			"name":     proc.Name,
			"cmd":      proc.Cmd,
			"bin_path": proc.BinPath,
			"pid":      proc.GetPID(),
			"port":     proc.GetPort(),
			"ctx_size": proc.GetCtxSize(),
			"running":  proc.IsRunning(),
		},
	})
}

func (s *Server) handleStopProcess(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "进程ID不能为空"})
		return
	}

	processMgr := s.modelMgr.GetProcessManager()
	if processMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "进程管理器未初始化"})
		return
	}

	if err := processMgr.Stop(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "停止进程失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "进程已停止",
		"id":      id,
	})
}

// handleLogStream streams log entries using Server-Sent Events
func (s *Server) handleLogStream(c *gin.Context) {
	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Get parameters
	fromBeginning := c.DefaultQuery("fromBeginning", "false") == "true"
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	// Get log stream
	logStream := logger.GetLogStream()

	// Flush headers
	c.Writer.Flush()

	// Create channel for log entries
	logCh := logStream.Subscribe()
	defer logStream.Unsubscribe(logCh)

	// Send existing entries if requested
	if fromBeginning {
		entries := logStream.GetEntries(limit)
		if len(entries) > 0 {
			// Batch send historical logs to reduce network I/O
			var buf strings.Builder
			for _, entry := range entries {
				data := fmt.Sprintf("data: {\"timestamp\":\"%s\",\"level\":\"%s\",\"message\":\"%s\"}\n\n",
					entry.Timestamp.Format(time.RFC3339),
					entry.Level,
					entry.Message)
				buf.WriteString(data)
			}
			c.Writer.WriteString(buf.String())
			c.Writer.Flush()
		}
	}

	// Keep connection alive and send new entries
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := c.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-logCh:
			if !ok {
				return
			}
			s.sendSSE(c, &entry)
			c.Writer.Flush()
		case <-ticker.C:
			// Send keepalive comment
			c.SSEvent("keepalive", "")
			c.Writer.Flush()
		}
	}
}

// sendSSE sends a log entry as Server-Sent Event
func (s *Server) sendSSE(c *gin.Context, entry *logger.StreamLogEntry) {
	data := fmt.Sprintf("data: {\"timestamp\":\"%s\",\"level\":\"%s\",\"message\":\"%s\"}\n\n",
		entry.Timestamp.Format(time.RFC3339),
		entry.Level,
		entry.Message)
	c.Writer.WriteString(data)
}

// handleLogEntries returns recent log entries
func (s *Server) handleLogEntries(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	logStream := logger.GetLogStream()
	entries := logStream.GetEntries(limit)

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"count":   len(entries),
	})
}

func (s *Server) handleEvents(c *gin.Context) {
	s.wsMgr.HandleWebSocket(c)
}

func (s *Server) handleOpenAIChat(c *gin.Context) {
	s.handlers.OpenAI.HandleChatCompletions(c)
}

func (s *Server) handleOpenAIComplete(c *gin.Context) {
	s.handlers.OpenAI.HandleCompletions(c)
}

func (s *Server) handleOpenAIModels(c *gin.Context) {
	s.handlers.OpenAI.HandleModels(c)
}

func (s *Server) handleAnthropicMessages(c *gin.Context) {
	s.handlers.Anthropic.HandleMessages(c)
}

func (s *Server) handleOllamaChat(c *gin.Context) {
	s.handlers.Ollama.HandleChat(c)
}

func (s *Server) handleOllamaTags(c *gin.Context) {
	s.handlers.Ollama.HandleTags(c)
}
