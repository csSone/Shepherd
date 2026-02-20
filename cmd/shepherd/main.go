// Shepherd - llama.cpp 模型管理系统
// 这是主程序入口文件
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/client"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/master"
	"github.com/shepherd-project/shepherd/Shepherd/internal/model"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
	"github.com/shepherd-project/shepherd/Shepherd/internal/process"
	"github.com/shepherd-project/shepherd/Shepherd/internal/server"
	"github.com/shepherd-project/shepherd/Shepherd/internal/shutdown"
)

// 版本信息（编译时注入）
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// App 表示应用程序实例，包含所有组件
type App struct {
	// 基础组件
	cfg         *config.Config
	configMgr   *config.Manager
	procMgr     *process.Manager
	modelMgr    *model.Manager
	shutdownMgr *shutdown.Manager
	srv         *server.Server

	// 分布式节点组件
	node            *node.Node
	nodeManager     *master.NodeManager
	masterHandler   *master.MasterHandler
	masterConnector *client.MasterConnector

	// 运行模式
	role string
}

func main() {
	// 命令行参数
	mode := flag.String("mode", "", "运行模式: standalone, master, client")
	version := flag.Bool("version", false, "显示版本信息")
	masterAddr := flag.String("master-address", "", "Master 地址 (client 模式)")
	flag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("Shepherd v%s\n", Version)
		fmt.Printf("构建时间: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// 打印启动信息
	printBanner()

	// 确定运行模式
	// 优先使用位置参数，然后是命令行参数，否则默认为 standalone
	runMode := "standalone"

	// 检查位置参数 (shepherd standalone, shepherd master, shepherd client)
	args := flag.Args()
	if len(args) > 0 {
		runMode = args[0]
	} else if *mode != "" {
		runMode = *mode
	}

	// 验证运行模式（向后兼容）
	if runMode != "standalone" && runMode != "master" && runMode != "client" {
		fmt.Fprintf(os.Stderr, "错误: 无效的运行模式 '%s'，必须是 standalone、master 或 client\n", runMode)
		os.Exit(1)
	}

	// 创建应用程序实例
	app := &App{}

	// 初始化应用程序
	if err := app.Initialize(runMode, *masterAddr); err != nil {
		fmt.Fprintf(os.Stderr, "初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 启动应用程序
	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}

	// 等待关闭
	app.Wait()

	fmt.Println("✓ 服务器已关闭")
	fmt.Println("再见！")
}

// printBanner 打印启动横幅
func printBanner() {
	fmt.Print(`
╔═════════════════════════════════════════════════════╗
║                                                       ║
║   Shepherd - llama.cpp 模型管理系统                  ║
║   (C) 2026 Shepherd Project                             ║
║                                                       ║
║   Go 语言重构版本 - 更快、更轻、更简单               ║
║                                                       ║
╚═════════════════════════════════════════════════════╝
`)
	fmt.Printf("版本: %s\n", Version)
	fmt.Printf("Commit: %s\n\n", GitCommit)
}

// Initialize 初始化应用程序
func (app *App) Initialize(runMode, masterAddr string) error {
	// 创建配置管理器（根据运行模式）
	app.configMgr = config.NewManager(runMode)

	// 加载配置
	cfg, err := app.configMgr.Load()
	if err != nil {
		fmt.Printf("警告: 无法加载配置文件，使用默认配置: %v\n", err)
		cfg = config.DefaultConfig()
	}
	app.cfg = cfg

	// 确保配置中的 mode 与运行时一致（向后兼容）
	cfg.Mode = runMode

	// 命令行参数覆盖配置
	if masterAddr != "" && runMode == "client" {
		cfg.Client.MasterAddress = masterAddr
		cfg.Node.ClientRole.MasterAddress = masterAddr
	}

	// 确定节点角色（新架构优先使用 cfg.Node.Role）
	app.role = app.determineRole(runMode)

	// 初始化日志系统
	if err := logger.InitLogger(&cfg.Log, app.role); err != nil {
		fmt.Printf("警告: 无法初始化日志系统: %v\n", err)
	}

	// 初始化日志流用于实时查看
	logger.InitLogStream(1000)

	logger.Info("Shepherd 正在启动...")
	logger.Infof("版本: %s", Version)
	logger.Infof("运行模式: %s", cfg.Mode)
	logger.Infof("节点角色: %s", app.role)
	logger.Infof("配置文件: %s", app.configMgr.GetConfigPath())

	// 创建进程管理器
	app.procMgr = process.NewManager()

	// 创建模型管理器
	app.modelMgr = model.NewManager(cfg, app.configMgr, app.procMgr)

	// 根据角色初始化分布式组件
	if err := app.initDistributedComponents(); err != nil {
		return fmt.Errorf("初始化分布式组件失败: %w", err)
	}

	// 创建 HTTP 服务器
	serverCfg := &server.Config{
		WebPort:       cfg.Server.WebPort,
		AnthropicPort: cfg.Server.AnthropicPort,
		OllamaPort:    cfg.Server.OllamaPort,
		LMStudioPort:  cfg.Server.LMStudioPort,
		Host:          cfg.Server.Host,
		ReadTimeout:   time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:  time.Duration(cfg.Server.WriteTimeout) * time.Second,
		WebUIPath:     "./web/dist",
		Mode:          cfg.Mode,
		ServerCfg:     cfg,
		ConfigMgr:     app.configMgr,
	}

	app.srv, err = server.NewServer(serverCfg, app.modelMgr)
	if err != nil {
		return fmt.Errorf("无法创建服务器: %w", err)
	}

	// 注册 Master API 路由（如果是 master 或 hybrid 模式）
	if app.role == "master" || app.role == "hybrid" {
		if app.masterHandler != nil {
			app.srv.RegisterMasterHandler(app.masterHandler)
		}
	}

	// 创建优雅关闭管理器
	app.shutdownMgr = shutdown.NewManager(10 * time.Second)

	// 注册关闭钩子
	app.registerShutdownHooks()

	return nil
}

// determineRole 根据运行模式和配置确定节点角色
func (app *App) determineRole(runMode string) string {
	// 如果配置了 Node.Role，优先使用（新架构）
	if app.cfg.Node.Role != "" && app.cfg.Node.Role != "standalone" {
		return app.cfg.Node.Role
	}

	// 否则根据运行模式映射（向后兼容）
	switch runMode {
	case "master":
		// 如果 master 角色被显式启用，返回 master
		if app.cfg.Node.MasterRole.Enabled {
			return "master"
		}
		// 否则使用旧架构的 master 模式
		return "master"
	case "client":
		// 如果 client 角色被显式启用，返回 client
		if app.cfg.Node.ClientRole.Enabled {
			return "client"
		}
		// 否则使用旧架构的 client 模式
		return "client"
	default:
		// 默认 standalone 模式
		return "standalone"
	}
}

// initDistributedComponents 根据角色初始化分布式组件
func (app *App) initDistributedComponents() error {
	logger.Infof("初始化分布式组件，角色: %s", app.role)

	switch app.role {
	case "standalone":
		// Standalone 模式：创建本地 Node（可选）
		if err := app.initStandaloneNode(); err != nil {
			logger.Warnf("初始化 standalone 节点失败: %v", err)
		}

	case "master":
		// Master 模式：创建 Node + NodeManager
		if err := app.initMasterNode(); err != nil {
			return fmt.Errorf("初始化 master 节点失败: %w", err)
		}
		if err := app.initNodeManager(); err != nil {
			return fmt.Errorf("初始化节点管理器失败: %w", err)
		}

	case "client":
		// Client 模式：创建 Node + MasterConnector
		if err := app.initClientNode(); err != nil {
			return fmt.Errorf("初始化 client 节点失败: %w", err)
		}
		if err := app.initMasterConnector(); err != nil {
			return fmt.Errorf("初始化 Master 连接器失败: %w", err)
		}

	case "hybrid":
		// Hybrid 模式：创建 Node + NodeManager + MasterConnector
		if err := app.initHybridNode(); err != nil {
			return fmt.Errorf("初始化 hybrid 节点失败: %w", err)
		}
		if err := app.initNodeManager(); err != nil {
			return fmt.Errorf("初始化节点管理器失败: %w", err)
		}
		if err := app.initMasterConnector(); err != nil {
			return fmt.Errorf("初始化 Master 连接器失败: %w", err)
		}

	default:
		return fmt.Errorf("未知的节点角色: %s", app.role)
	}

	return nil
}

// initStandaloneNode 初始化 Standalone 模式的 Node
func (app *App) initStandaloneNode() error {
	nodeCfg := app.buildNodeConfig()
	nodeCfg.Role = node.NodeRoleStandalone

	n, err := node.NewNode(nodeCfg)
	if err != nil {
		return err
	}

	app.node = n
	logger.Info("Standalone 节点已创建")
	return nil
}

// initMasterNode 初始化 Master 模式的 Node
func (app *App) initMasterNode() error {
	nodeCfg := app.buildNodeConfig()
	nodeCfg.Role = node.NodeRoleMaster
	nodeCfg.Port = app.cfg.Node.MasterRole.Port

	n, err := node.NewNode(nodeCfg)
	if err != nil {
		return err
	}

	app.node = n
	logger.Info("Master 节点已创建")
	return nil
}

// initClientNode 初始化 Client 模式的 Node
func (app *App) initClientNode() error {
	nodeCfg := app.buildNodeConfig()
	nodeCfg.Role = node.NodeRoleClient
	nodeCfg.MasterAddress = app.cfg.Node.ClientRole.MasterAddress

	n, err := node.NewNode(nodeCfg)
	if err != nil {
		return err
	}

	app.node = n
	logger.Info("Client 节点已创建")
	return nil
}

// initHybridNode 初始化 Hybrid 模式的 Node
func (app *App) initHybridNode() error {
	nodeCfg := app.buildNodeConfig()
	nodeCfg.Role = node.NodeRoleHybrid
	nodeCfg.Port = app.cfg.Node.MasterRole.Port
	nodeCfg.MasterAddress = app.cfg.Node.ClientRole.MasterAddress

	n, err := node.NewNode(nodeCfg)
	if err != nil {
		return err
	}

	app.node = n
	logger.Info("Hybrid 节点已创建")
	return nil
}

// buildNodeConfig 从应用配置构建 NodeConfig
func (app *App) buildNodeConfig() *node.NodeConfig {
	cfg := app.cfg

	nodeID := cfg.Node.ID
	if nodeID == "auto" || nodeID == "" {
		// 生成默认节点 ID
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "shepherd"
		}
		nodeID = fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
	}

	nodeName := cfg.Node.Name
	if nodeName == "" {
		nodeName = nodeID
	}

	return &node.NodeConfig{
		ID:                nodeID,
		Name:              nodeName,
		Address:           cfg.Server.Host,
		Port:              cfg.Server.WebPort,
		HeartbeatInterval: time.Duration(cfg.Node.ClientRole.HeartbeatInterval) * time.Second,
		Timeout:           time.Duration(cfg.Node.ClientRole.HeartbeatTimeout) * time.Second,
		MaxRetries:        cfg.Node.ClientRole.RegisterRetry,
		LogLevel:          cfg.Log.Level,
		EnableMetrics:     true,
	}
}

// initNodeManager 初始化 NodeManager
func (app *App) initNodeManager() error {
	app.nodeManager = master.NewNodeManager(nil) // 使用默认日志
	logger.Info("节点管理器已创建")
	return nil
}

// initMasterConnector 初始化 MasterConnector
func (app *App) initMasterConnector() error {
	if app.node == nil {
		return fmt.Errorf("节点未初始化")
	}

	masterAddr := app.cfg.Node.ClientRole.MasterAddress
	if masterAddr == "" {
		return fmt.Errorf("Master 地址未配置")
	}

	// 创建资源监控器
	resourceMonitor := node.NewResourceMonitor(
		&node.ResourceMonitorConfig{
			Interval: time.Duration(app.cfg.Node.Resources.MonitorInterval) * time.Second,
		})

	// 创建心跳管理器
	heartbeatMgr := node.NewHeartbeatManager(
		&node.HeartbeatConfig{
			NodeID:          app.node.GetID(),
			MasterAddr:      masterAddr,
			Interval:        time.Duration(app.cfg.Node.ClientRole.HeartbeatInterval) * time.Second,
			Timeout:         time.Duration(app.cfg.Node.ClientRole.HeartbeatTimeout) * time.Second,
			MaxRetries:      app.cfg.Node.ClientRole.RegisterRetry,
			ResourceMonitor: resourceMonitor,
			OnConnect: func() {
				logger.Info("已连接到 Master")
			},
			OnDisconnect: func(err error) {
				logger.Warnf("与 Master 断开连接: %v", err)
			},
		})

	// 创建命令执行器
	executor := node.NewCommandExecutor(
		&node.CommandExecutorConfig{
			MaxConcurrent:   app.cfg.Node.Executor.MaxConcurrent,
			Timeout:         time.Duration(app.cfg.Node.Executor.TaskTimeout) * time.Second,
			AllowedCommands: []node.CommandType{}, // 使用默认值
		})

	// 创建 MasterConnector
	connectorCfg := &client.MasterConnectorConfig{
		NodeID:               app.node.GetID(),
		MasterAddr:           masterAddr,
		NodeInfo:             app.node.ToInfo(),
		HeartbeatMgr:         heartbeatMgr,
		Executor:             executor,
		ResourceMonitor:      resourceMonitor,
		MaxReconnectAttempts: app.cfg.Node.ClientRole.RegisterRetry,
		CommandPollInterval:  2 * time.Second,
	}

	connector, err := client.NewMasterConnector(connectorCfg)
	if err != nil {
		return err
	}

	app.masterConnector = connector
	logger.Infof("Master 连接器已创建，目标: %s", masterAddr)
	return nil
}

// registerShutdownHooks 注册优雅关闭钩子
func (app *App) registerShutdownHooks() {
	// 1. 优先级最高：停止接受新连接（HTTP服务器）
	app.shutdownMgr.Register("http-server", func(ctx context.Context) error {
		if app.srv != nil {
			return app.srv.Shutdown(ctx)
		}
		return nil
	}, shutdown.PriorityCritical)

	// 2. 优先级高：停止 MasterConnector（如果是 client 或 hybrid 模式）
	if app.masterConnector != nil {
		app.shutdownMgr.Register("master-connector", func(ctx context.Context) error {
			return app.masterConnector.Disconnect()
		}, shutdown.PriorityCritical)
	}

	// 3. 优先级高：停止 NodeManager（如果是 master 或 hybrid 模式）
	if app.nodeManager != nil {
		app.shutdownMgr.Register("node-manager", func(ctx context.Context) error {
			app.nodeManager.Stop()
			return nil
		}, shutdown.PriorityHigh)
	}

	// 4. 优先级高：停止 Node
	if app.node != nil {
		app.shutdownMgr.Register("node", func(ctx context.Context) error {
			return app.node.Stop()
		}, shutdown.PriorityHigh)
	}

	// 5. 优先级高：停止所有模型加载和处理
	if app.modelMgr != nil {
		app.shutdownMgr.Register("models", func(ctx context.Context) error {
			app.modelMgr.Close()
			return nil
		}, shutdown.PriorityHigh)
	}

	// 6. 优先级中：停止所有进程
	if app.procMgr != nil {
		app.shutdownMgr.Register("processes", func(ctx context.Context) error {
			app.procMgr.StopAll()
			return nil
		}, shutdown.PriorityNormal)
	}

	// 7. 优先级低：关闭日志系统
	app.shutdownMgr.Register("logger", func(ctx context.Context) error {
		logger.Info("日志系统已关闭")
		return nil
	}, shutdown.PriorityLow)
}

// Start 启动应用程序
func (app *App) Start() error {
	// 启动 Node（如果已创建）
	if app.node != nil {
		if err := app.node.Start(); err != nil {
			return fmt.Errorf("启动节点失败: %w", err)
		}
		logger.Info("节点已启动")
	}

	// 启动 NodeManager（如果是 master 或 hybrid 模式）
	if app.nodeManager != nil {
		app.nodeManager.Start()
		logger.Info("节点管理器已启动")
	}

	// 启动 MasterHandler（如果是 master 或 hybrid 模式）
	if app.role == "master" || app.role == "hybrid" {
		if app.nodeManager != nil {
			app.masterHandler = master.NewMasterHandler(app.nodeManager, nil)
			logger.Info("Master 处理器已创建")
		}
	}

	// 启动 MasterConnector（如果是 client 或 hybrid 模式）
	if app.masterConnector != nil {
		if err := app.masterConnector.Connect(); err != nil {
			// Client 模式连接失败是致命错误
			if app.role == "client" {
				return fmt.Errorf("连接到 Master 失败: %w", err)
			}
			// Hybrid 模式记录警告但继续运行
			logger.Warnf("连接到 Master 失败: %v", err)
		} else {
			logger.Info("已连接到 Master")
		}
	}

	// 启动 HTTP 服务器
	if err := app.srv.Start(); err != nil {
		return fmt.Errorf("无法启动服务器: %w", err)
	}

	// 启动优雅关闭管理器
	app.shutdownMgr.Start()

	// 打印启动信息
	app.printStartupInfo()

	return nil
}

// printStartupInfo 打印启动信息
func (app *App) printStartupInfo() {
	logger.Infof("HTTP 服务器已启动，监听 %s:%d", app.cfg.Server.Host, app.cfg.Server.WebPort)
	fmt.Printf("✓ 运行模式: %s\n", app.cfg.Mode)
	fmt.Printf("✓ 节点角色: %s\n", app.role)
	fmt.Printf("✓ HTTP 服务器已启动，监听 %s:%d\n", app.cfg.Server.Host, app.cfg.Server.WebPort)
	fmt.Printf("✓ Web UI: http://localhost:%d\n", app.cfg.Server.WebPort)
	fmt.Printf("✓ OpenAI API: http://localhost:%d/v1\n", app.cfg.Server.WebPort)

	if app.cfg.Compatibility.Ollama.Enabled {
		fmt.Printf("✓ Ollama API: http://localhost:%d\n", app.cfg.Server.OllamaPort)
	}

	if app.role == "master" || app.role == "hybrid" {
		fmt.Printf("✓ Master API: http://localhost:%d/api/master\n", app.cfg.Server.WebPort)
	}

	if app.role == "client" && app.masterConnector != nil {
		fmt.Printf("✓ 连接到 Master: %s\n", app.masterConnector.GetMasterAddr())
	}

	fmt.Println("\n按 Ctrl+C 停止服务器...")
}

// Wait 等待应用程序关闭
func (app *App) Wait() {
	// 等待关闭信号或上下文取消
	select {
	case <-app.shutdownMgr.Context().Done():
		// Shutdown initiated
	case <-app.shutdownMgr.Done():
		// Shutdown complete
	}

	// 等待所有关闭钩子完成
	app.shutdownMgr.Wait()

	logger.Info("服务器已关闭")
}
