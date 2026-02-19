// Shepherd - llama.cpp 模型管理系统
// 这是主程序入口文件
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/model"
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

	// 验证运行模式
	if runMode != "standalone" && runMode != "master" && runMode != "client" {
		fmt.Fprintf(os.Stderr, "错误: 无效的运行模式 '%s'，必须是 standalone、master 或 client\n", runMode)
		os.Exit(1)
	}

	// 创建配置管理器（根据运行模式）
	configMgr := config.NewManager(runMode)

	// 加载配置
	cfg, err := configMgr.Load()
	if err != nil {
		fmt.Printf("警告: 无法加载配置文件，使用默认配置: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// 确保配置中的 mode 与运行时一致
	cfg.Mode = runMode

	// 命令行参数覆盖配置
	if *masterAddr != "" && runMode == "client" {
		cfg.Client.MasterAddress = *masterAddr
	}

	// 初始化日志系统（传递运行模式）
	if err := logger.InitLogger(&cfg.Log, runMode); err != nil {
		fmt.Printf("警告: 无法初始化日志系统: %v\n", err)
	}

	logger.Info("Shepherd 正在启动...")
	logger.Infof("版本: %s", Version)
	logger.Infof("运行模式: %s", cfg.Mode)
	logger.Infof("配置文件: %s", configMgr.GetConfigPath())

	// 根据模式启动不同的组件
	switch cfg.Mode {
	case "master":
		logger.Info("启动 Master 模式...")
		// TODO: 初始化 Master 组件
	case "client":
		logger.Info("启动 Client 模式...")
		if cfg.Client.MasterAddress == "" {
			logger.Fatal("Client 模式需要指定 master-address")
		}
		logger.Infof("连接到 Master: %s", cfg.Client.MasterAddress)
		// TODO: 初始化 Client 组件
	case "standalone", "":
		cfg.Mode = "standalone"
		logger.Info("启动单机模式...")
	default:
		logger.Fatalf("未知的运行模式: %s (支持: standalone, master, client)", cfg.Mode)
	}

	// 创建进程管理器
	procMgr := process.NewManager()

	// 创建模型管理器
	modelMgr := model.NewManager(cfg, configMgr, procMgr)

	// 创建 HTTP 服务器
	serverCfg := &server.Config{
		WebPort:       cfg.Server.WebPort,
		AnthropicPort: cfg.Server.AnthropicPort,
		OllamaPort:    cfg.Server.OllamaPort,
		LMStudioPort:  cfg.Server.LMStudioPort,
		Host:          cfg.Server.Host,
		ReadTimeout:   time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:  time.Duration(cfg.Server.WriteTimeout) * time.Second,
		WebUIPath:     "./web",
		Mode:          cfg.Mode,
		ServerCfg:     cfg,
	}

	srv := server.NewServer(serverCfg, modelMgr)

	// 创建优雅关闭管理器
	shutdownMgr := shutdown.NewManager(10 * time.Second)

	// 注册关闭钩子（按优先级顺序执行）
	// 1. 优先级最高：停止接受新连接（HTTP服务器）
	shutdownMgr.Register("http-server", func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	}, shutdown.PriorityCritical)

	// 2. 优先级高：停止所有模型加载和处理
	shutdownMgr.Register("models", func(ctx context.Context) error {
		modelMgr.Close()
		return nil
	}, shutdown.PriorityHigh)

	// 3. 优先级中：停止所有进程
	shutdownMgr.Register("processes", func(ctx context.Context) error {
		procMgr.StopAll()
		return nil
	}, shutdown.PriorityNormal)

	// 4. 优先级低：关闭日志系统
	shutdownMgr.Register("logger", func(ctx context.Context) error {
		logger.Info("日志系统已关闭")
		return nil
	}, shutdown.PriorityLow)

	// 启动服务器
	if err := srv.Start(); err != nil {
		logger.Fatalf("无法启动服务器: %v", err)
	}

	// 启动优雅关闭管理器
	shutdownMgr.Start()

	logger.Infof("HTTP 服务器已启动，监听 %s:%d", cfg.Server.Host, cfg.Server.WebPort)
	fmt.Printf("✓ 运行模式: %s\n", cfg.Mode)
	fmt.Printf("✓ HTTP 服务器已启动，监听 %s:%d\n", cfg.Server.Host, cfg.Server.WebPort)
	fmt.Printf("✓ Web UI: http://localhost:%d\n", cfg.Server.WebPort)
	fmt.Printf("✓ OpenAI API: http://localhost:%d/v1\n", cfg.Server.WebPort)
	if cfg.Compatibility.Ollama.Enabled {
		fmt.Printf("✓ Ollama API: http://localhost:%d\n", cfg.Server.OllamaPort)
	}

	fmt.Println("\n按 Ctrl+C 停止服务器...")

	// 等待关闭信号或上下文取消
	select {
	case <-shutdownMgr.Context().Done():
		// Shutdown initiated
	case <-shutdownMgr.Done():
		// Shutdown complete
	}

	// 等待所有关闭钩子完成
	shutdownMgr.Wait()

	logger.Info("服务器已关闭")
	fmt.Println("✓ 服务器已关闭")
	fmt.Println("再见！")
}
