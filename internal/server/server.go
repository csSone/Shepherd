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
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/openai"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/ollama"
	"github.com/shepherd-project/shepherd/Shepherd/internal/api/paths"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/model"
	"github.com/shepherd-project/shepherd/Shepherd/internal/websocket"
)

// Server represents the HTTP server
type Server struct {
	engine      *gin.Engine
	httpServer  *http.Server
	config      *Config
	handlers    *Handlers
	wsMgr       *websocket.Manager
	modelMgr    *model.Manager

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
	Mode        string // standalone|master|client
	ServerCfg   *config.Config
	ConfigMgr   *config.Manager // 配置管理器
}

// Handlers contains handler instances
type Handlers struct {
	OpenAI    *openai.Handler
	Ollama    *ollama.Handler
	Anthropic *anthropic.Handler
	Paths     *paths.Handler
}

// NewServer creates a new HTTP server
func NewServer(config *Config, modelMgr *model.Manager) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		handlers: &Handlers{},
		modelMgr: modelMgr,
	}

	// Create WebSocket manager
	s.wsMgr = websocket.NewManager(modelMgr)

	// Create API handlers
	s.handlers.OpenAI = openai.NewHandler(modelMgr)
	s.handlers.Ollama = ollama.NewHandler(modelMgr)
	s.handlers.Anthropic = anthropic.NewHandler(modelMgr)
	s.handlers.Paths = paths.NewHandler(config.ConfigMgr)

	// Setup Gin engine
	if config.WebUIPath == "" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	s.engine = gin.New()
	s.setupMiddleware()
	s.setupRoutes()

	return s
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

	// Step 5: Wait for all goroutines to finish
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

// Placeholder handlers (to be implemented)
func (s *Server) handleGetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleUpdateConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleListModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"models": []interface{}{}})
}

func (s *Server) handleGetModel(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (s *Server) handleLoadModel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleUnloadModel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleSetAlias(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleSetFavourite(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleScanModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleGetScanStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleListDownloads(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"downloads": []interface{}{}})
}

func (s *Server) handleCreateDownload(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleGetDownload(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handlePauseDownload(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleResumeDownload(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleDeleteDownload(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleListProcesses(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"processes": []interface{}{}})
}

func (s *Server) handleGetProcess(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
}

func (s *Server) handleStopProcess(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TODO: implement"})
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
