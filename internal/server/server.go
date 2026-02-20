// Package server provides the HTTP server for the Shepherd application.
// It handles HTTP requests, routing, middleware, and serves the web UI.
package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/anthropic"
	api "github.com/shepherd-project/shepherd/Shepherd/internal/api"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/ollama"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/openai"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/paths"
	storageapi "github.com/shepherd-project/shepherd/Shepherd/internal/api/storage"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/model"
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
	Favourite   bool                   `json:"favourite"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      string                 `json:"status"`
	IsLoaded    bool                   `json:"isLoaded"`
}

// Server represents the HTTP server
type Server struct {
	engine       *gin.Engine
	httpServer   *http.Server
	config       *Config
	handlers     *Handlers
	wsMgr        *websocket.Manager
	modelMgr     *model.Manager
	storageMgr   *storage.Manager
	downloadMgr  *DownloadManager // 下载管理器
	nodeAdapter  *api.NodeAdapter // Node API 适配器

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
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
	OpenAI    *openai.Handler
	Ollama    *ollama.Handler
	Anthropic *anthropic.Handler
	Paths     *paths.Handler
	Storage   *storageapi.Handler
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

	// Create API handlers
	s.handlers.OpenAI = openai.NewHandler(modelMgr)
	s.handlers.Ollama = ollama.NewHandler(modelMgr)
	s.handlers.Anthropic = anthropic.NewHandler(modelMgr)
	s.handlers.Paths = paths.NewHandler(config.ConfigMgr)
	s.handlers.Storage = storageapi.NewHandler(config.ConfigMgr, storageMgr)

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

	// API routes
	api := s.engine.Group("/api")
	{
		// Server info
		api.GET("/info", s.handleServerInfo)

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
			models.GET("/:id", s.handleGetModel)
			models.POST("/:id/load", s.handleLoadModel)
			models.POST("/:id/unload", s.handleUnloadModel)
			models.PUT("/:id/alias", s.handleSetAlias)
			models.PUT("/:id/favourite", s.handleSetFavourite)
		}

		// Scan routes
		scan := api.Group("/scan")
		{
			scan.POST("", s.handleScanModels)
			scan.GET("/status", s.handleGetScanStatus)
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

	var dtos []ModelDTO
	for _, m := range models {
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

		// Convert metadata
		if m.Metadata != nil {
			dto.Metadata = map[string]interface{}{
				"name":            m.Metadata.Name,
				"architecture":    m.Metadata.Architecture,
				"quantization":    m.Metadata.Quantization,
				"contextLength":   m.Metadata.ContextLength,
				"embeddingLength": m.Metadata.EmbeddingLength,
				"layerCount":      m.Metadata.BlockSize,
				"headCount":       m.Metadata.HeadCount,
			}
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

	result, err := s.modelMgr.Load(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	var req struct {
		URL        string `json:"url" binding:"required"`
		TargetPath string `json:"target_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求格式",
			"message": err.Error(),
		})
		return
	}

	task, err := s.downloadMgr.CreateDownload(req.URL, req.TargetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建下载失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "下载任务已创建",
		"task":    task,
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

func (s *Server) handleListProcesses(c *gin.Context) {
	processMgr := s.modelMgr.GetProcessManager()
	if processMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "进程管理器未初始化"})
		return
	}

	running, loading := processMgr.ListAll()

	// 转换为切片格式
	type ProcessInfo struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		PID      int    `json:"pid"`
		Port     int    `json:"port"`
		CtxSize  int    `json:"ctx_size"`
		Running  bool   `json:"running"`
		Loading  bool   `json:"loading"`
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
		for _, entry := range entries {
			s.sendSSE(c, &entry)
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
