# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Shepherd 是一个高性能轻量级分布式 llama.cpp 模型管理系统，采用 Go 语言开发。

**核心特性：**
- 单一二进制部署，无运行时依赖
- 支持 Master-Client 分布式架构
- 统一 Node 模型，节点角色可灵活切换（Standalone/Master/Client/Hybrid）
- 多 API 兼容（OpenAI/Anthropic/Ollama/LM Studio）
- Web 前端与后端完全解耦，支持运行时切换后端
- **统一类型系统（v0.2.0+）**：`internal/types/` 定义跨模块类型

## 环境要求

**后端开发：**
- Go 1.25+ (GOROOT: `/home/user/sdk/go`)
- Git
- GOPROXY: `https://goproxy.cn,direct`

**前端开发：**
- Node.js 18+
- npm

**运行时依赖：**
- llama.cpp（用于模型推理）
- ROCm 7.x（AMD GPU）或 CUDA（NVIDIA GPU）

## 常用命令

### 编译和运行

```bash
# 编译当前平台（自动注入版本信息）
make build

# 跨平台编译
make build-all

# 运行（使用 Makefile）
make run

# 直接运行编译后的二进制
./build/shepherd standalone    # 单机模式
./build/shepherd master        # Master 模式
./build/shepherd client --master-address http://master:9190  # Client 模式
./build/shepherd hybrid        # Hybrid 模式（默认）

# 查看版本
./build/shepherd --version

# 清理构建文件
make clean
```

### 测试和代码质量

```bash
# 运行所有测试
make test
# 或
go test ./... -v

# 运行特定包的测试
make test TEST=./internal/config
go test ./internal/model -run TestScanModels -v

# 运行测试并生成覆盖率报告
make test-coverage

# 代码检查（需要安装 golangci-lint）
make lint

# 代码格式化
make fmt

# 整理依赖
make tidy
```

### 前端开发

```bash
cd web

# 安装依赖
npm install

# 启动开发服务器（端口 3000）
npm run dev

# 类型检查
npm run type-check

# 代码检查
npm run lint
npm run lint:fix

# 运行测试
npm run test
npm run test:coverage

# 构建生产版本
npm run build

# 预览生产构建
npm run preview

# 同步配置文件（修改 web/config.yaml 后）
npm run sync-config
```

## 架构概览

### 核心组件关系

```
┌─────────────────────────────────────────────────────────────────┐
│                        Shepherd 架构                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐  │
│  │   cmd/       │      │  internal/   │      │    web/      │  │
│  │ shepherd/    │──────│   server/    │──────│   (前端)      │  │
│  │ main.go      │      │              │      │              │  │
│  └──────────────┘      └──────┬───────┘      └──────────────┘  │
│                               │                                 │
│              ┌──────────────────┼──────────────────┐            │
│              ▼                  ▼                  ▼            │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    │
│  │   node/      │    │   model/     │    │   config/    │    │
│  │ (分布式节点)  │    │ (模型管理)   │    │  (配置管理)  │    │
│  └──────────────┘    └──────────────┘    └──────────────┘    │
│         │                    │                                │
│         └────────────────────┼────────────────┐              │
│                             ▼                ▼              │
│                   ┌──────────────┐  ┌──────────────┐         │
│                   │  storage/    │  │    api/      │         │
│                   │ (持久化层)   │  │ (API 处理)   │         │
│                   └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

### 目录结构

| 目录 | 说明 |
|------|------|
| `cmd/shepherd/` | 主程序入口，处理命令行参数和应用程序初始化 |
| `internal/api/` | API 处理器，包括 OpenAI、Anthropic、Ollama 等兼容层 |
| `internal/types/` | **统一类型定义**（NodeState、ErrorCode、ApiResponse、Command 等） |
| `internal/node/` | 统一节点模型，支持 Standalone/Master/Client/Hybrid 角色 |
| `internal/model/` | 模型管理器，负责 GGUF 模型扫描、加载、卸载 |
| `internal/storage/` | 持久化层，支持内存和 SQLite 后端 |
| `internal/config/` | 配置管理，支持多模式配置文件 |
| `internal/server/` | HTTP 服务器， Gin 路由和中间件 |
| `internal/logger/` | 日志系统，按运行模式分类 |
| `internal/client/` | Client 节点组件，Master 连接和心跳 |
| `internal/client/tester/` | llama.cpp 可用性测试（新增） |
| `internal/client/configreport/` | Client 配置信息收集（新增） |
| `internal/master/` | Master 节点组件，调度器和节点管理 |
| `internal/process/` | 进程管理，启动和管理 llama.cpp 进程 |
| `internal/cluster/` | 集群通信，心跳、注册、命令分发 |
| `web/src/types/` | **前端类型定义**（node.ts, cluster.ts, events.ts 等） |
| `web/` | React 前端，独立配置和部署 |
| `config/example/` | 示例配置文件 |

### 组件详解

#### 1. 统一类型系统 (`internal/types/`)

**重要（v0.2.0+）**：所有跨模块使用的类型定义在 `internal/types/` 中：

```go
// internal/types/common.go - 通用类型
type NodeState string         // 统一节点状态
type ErrorCode string         // 统一错误码
type ApiResponse[T any]       // 统一 API 响应格式

// internal/types/node.go - 节点相关类型
type NodeCapabilities         // 节点能力
type NodeResources            // 节点资源
type NodeInfo                 // 统一节点信息
type HeartbeatMessage         // 统一心跳消息
type CommandType string       // 命令类型
type Command                  // 命令结构
type CommandResult            // 命令执行结果
```

**类型别名（向后兼容）**：
- `internal/node/types.go` 中的类型使用别名指向 `internal/types/`
- `internal/cluster/types.go` 中的 `Client` 是 `types.NodeInfo` 的别名

#### 2. Node 统一节点模型 (`internal/node/`)

Shepherd 采用统一的 Node 架构，每个节点可以扮演不同角色：

```go
// 节点角色
type NodeRole string
const (
    NodeRoleStandalone NodeRole = "standalone"  // 单机模式
    NodeRoleMaster     NodeRole = "master"      // 主节点
    NodeRoleClient     NodeRole = "client"      // 工作节点
    NodeRoleHybrid     NodeRole = "hybrid"      // 混合节点
)
```

**关键设计：**
- Node 包含子系统管理器（ResourceMonitor、SubsystemManager）
- Client 角色：自动注册、心跳、命令接收
- Master/Hybrid 角色：客户端注册表、命令队列、结果存储
- 所有角色共享资源监控和能力报告

**初始化流程：**
```go
// cmd/shepherd/main.go
app.Initialize(mode, masterAddr, configPath)
  ├─ configMgr.Load()                    // 加载配置
  ├─ logger.InitLogger()                  // 初始化日志
  ├─ app.initDistributedComponents()      // 初始化分布式组件
  │   ├─ app.initStandaloneNode()        // 单机节点
  │   ├─ app.initMasterNode()            // Master 节点 + NodeAdapter
  │   ├─ app.initClientNode()            // Client 节点
  │   └─ app.initHybridNode()            // Hybrid 节点 + NodeAdapter
  └─ server.NewServer()                   // 创建 HTTP 服务器
```

#### 3. 模型管理器 (`internal/model/`)

负责 GGUF 格式模型的扫描、加载、卸载：

**核心功能：**
- 自动扫描配置路径中的 GGUF 模型
- 使用 gguf-parser-go 解析模型元数据
- 支持分卷模型（自动识别 split-00001-of-00005 格式）
- 支持视觉模型（mmproj 文件检测）
- 端口自动分配（8081-9000 范围）
- 模型能力配置（thinking/tools/rerank/embedding）
- **丰富的加载参数**（详见 `doc/api/model-loading.md`）：
  - 基础参数: 上下文大小、批次大小、线程数、GPU 层数
  - 采样参数: 温度、Top-P、Top-K、重复惩罚、Min-P、Presence/Frequency 惩罚
  - 性能优化: Flash Attention、内存锁定、UBatch、并行槽位
  - KV 缓存: 类型配置 (K/V)、统一缓存、缓存大小
  - 模板系统: Jinja 禁用、自定义模板、上下文切换
  - GPU 配置: 多设备支持、主 GPU 选择、设备列表

**扫描流程：**
```go
modelMgr.Scan(ctx)
  ├─ 遍历配置路径
  ├─ 检测 .gguf 文件
  ├─ 解析 GGUF 元数据
  ├─ 检测分卷文件
  ├─ 检测 mmproj 文件
  └─ 更新模型存储
```

#### 4. API 兼容层 (`internal/api/`)

提供多种 API 格式的兼容支持：

| API | 端口 | 路径 | 状态 |
|-----|------|------|------|
| OpenAI | :9190 | `/v1/*` | ✅ |
| Anthropic | :9170 | `/v1/messages` | ✅ |
| Ollama | :11434 | `/api/*` | ✅ |

**关键处理器：**
- `openai.Handler` - OpenAI API 兼容
- `anthropic.Handler` - Anthropic API 兼容
- `ollama.Handler` - Ollama API 兼容
- `benchmark.Handler` - 模型压测和性能测试
- `paths.Handler` - llama.cpp 和模型路径配置
- `NodeAdapter` - 统一节点 API（`/api/nodes/*`）

**v0.2.0+ API 路由变更：**
- 旧路由：`/api/master/clients/*` → 新路由：`/api/nodes/*`
- 旧路由标记为废弃，将在 v0.4.0 移除

#### 5. 存储层 (`internal/storage/`)

支持多种存储后端：

```go
type StorageType string
const (
    StorageTypeMemory     StorageType = "memory"     // 内存存储（默认）
    StorageTypeSQLite     StorageType = "sqlite"     // SQLite 文件
    StorageTypePostgreSQL StorageType = "postgresql" // 未来支持
)
```

**数据模型：**
- `Conversation` - 聊天对话
- `Message` - 聊天消息
- `Benchmark` - 压测任务
- `BenchmarkConfig` - 压测配置模板

#### 6. 分布式架构 (`internal/cluster/`, `internal/client/`, `internal/master/`)

**Master-Client 通信流程：**
```
Client 节点                    Master 节点
    │                              │
    ├────── 注册 ────────────────►│
    │◄───── 注册确认 ─────────────┤
    │                              │
    ├────── 心跳（含资源状态） ───►│
    │◄───── 心跳响应 ─────────────┤
    │                              │
    ├────── 命令执行结果 ─────────►│
    │                              │
    │◄───── 下发命令 ─────────────┤
    │                              │
```

**关键组件：**
- `HeartbeatMessage` - 心跳消息，包含 CPU/GPU/内存状态
- `Command` - 命令结构，支持签名验证
- `Scheduler` - 调度器（round_robin/least_loaded/resource_aware）
- `internal/client/tester` - llama.cpp 可用性测试（v0.3.0+）
- `internal/client/configreport` - Client 配置信息收集（v0.3.0+）
- `internal/download/` - HuggingFace 模型下载管理器（v0.3.0+）

### 配置文件

**配置文件位置：**
- `config/example/server.config.yaml` - 单机模式
- `config/example/master.config.yaml` - Master 模式
- `config/example/client.config.yaml` - Client 模式

**前端独立配置：**
- `web/config.yaml` - 前端配置，支持多后端和运行时切换

**配置加载优先级：**
1. 命令行参数 `--config`
2. 运行模式对应的配置文件
3. 默认配置（`config.DefaultConfig()`）

**模式变更（v0.2.0+）：**
- 默认模式从 `hybrid` 改为 `standalone`
- 支持所有四种模式：standalone, hybrid, master, client

### 日志系统

日志按运行模式分类，文件名格式：`shepherd-{mode}-{date}.log`

**日志配置：**
```yaml
log:
  level: "info"        # debug, info, warn, error
  format: "json"       # text, json
  output: "both"       # stdout, file, both
  directory: "logs"
  max_size: 100        # MB
  max_age: 7           # days
```

**日志格式（v0.3.1+）：**
- **文本格式**: `[时间] [文件:行号] 级别 消息 字段...`
  - 示例: `[2026-02-26 20:17:51] [manager.go:72] INFO 模型加载完成 modelCount=1`
- **JSON 格式**: `{"time":"...","caller":"...","level":"INFO","msg":"...",...}`

**日志查看：**
- 实时流：`GET /api/logs/stream`
- 历史记录：`GET /api/logs/entries?limit=100`

**使用 Logger：**
```go
import "github.com/shepherd-project/shepherd/Shepherd/internal/logger"

// 基础日志
logger.Info("消息内容")
logger.Warn("警告消息", "key1", "value1")
logger.Error("错误消息", "error", err)

// 格式化日志
logger.Infof("模型 %s 加载成功", modelID)
logger.Debugf("调试信息: %+v", data)
```

### 前端架构

**技术栈：**
- React 19.2.0 + TypeScript 5.x
- Vite 7.x 构建工具
- React Router v7 路由
- Zustand 状态管理
- React Query 数据获取
- Tailwind CSS 4.x + shadcn/ui

**关键特性：**
- 完全独立于后端，可连接任意后端服务器
- 支持多后端配置和运行时切换
- SSE 实时事件推送
- 国际化支持（中英文）

**前端配置：**
```yaml
# web/config.yaml
backend:
  urls:
    - "http://localhost:9190"
    - "http://backup:9190"
  currentIndex: 0

features:
  models: true
  downloads: true
  cluster: false      # 开发中
  logs: false         # 开发中
  chat: true
  settings: true
  dashboard: true
```

**前端类型系统**：
```
web/src/types/
├── node.ts       # 统一节点类型（UnifiedNode, NodeRole, NodeStatus）
├── cluster.ts    # 集群任务类型（ClusterTask, TaskStatus）- @deprecated
├── events.ts     # SSE 事件类型
├── model.ts      # 模型相关类型
├── download.ts   # 下载管理类型
├── logs.ts       # 日志类型
├── websocket.ts  # WebSocket 类型
├── task.ts       # 任务相关类型
└── common.ts     # 通用类型
```

**类型同步**：前端类型与后端 `internal/types/` 中的定义保持一致

**v0.2.0+ 前端类型迁移：**
- `web/src/types/cluster.ts` 中的 `Client` 类型标记为 `@deprecated`
- 建议使用 `web/src/types/node.ts` 中的 `UnifiedNode` 类型

## 开发指南

## API 文档

详细的 API 文档位于 `doc/api/` 目录：

| 文档 | 描述 |
|------|------|
| `model-loading.md` | 模型加载参数详解（所有支持的 llama.cpp 参数） |

### 添加新的 API 端点

1. 在 `internal/api/` 中创建新的处理器
2. 在 `internal/server/server.go` 的 `setupRoutes()` 中注册路由
3. 添加对应的中间件和错误处理

示例：
```go
// internal/server/server.go
func (s *Server) setupRoutes() {
    api := s.engine.Group("/api")
    {
        api.GET("/new-endpoint", s.handleNewEndpoint)
    }
}

func (s *Server) handleNewEndpoint(c *gin.Context) {
    api.Success(c, gin.H{"message": "Hello"})
}
```

### 添加新的存储实体

1. 在 `internal/storage/storage.go` 中定义数据结构
2. 在 `internal/storage/sqlite.go` 中实现数据库操作
3. 在 `internal/storage/memory.go` 中实现内存存储操作
4. 更新 `Store` 接口

### 修改 Node 行为

Node 子系统在 `internal/node/subsystem.go` 中定义：

- `RegistrationSubsystem` - Client 自动注册
- `HeartbeatSubsystem` - 心跳发送
- `CommandSubsystem` - 命令处理

添加新子系统：
```go
type MyCustomSubsystem struct {
    node *Node
}

func (s *MyCustomSubsystem) Start() error { /* ... */ }
func (s *MyCustomSubsystem) Stop() error { /* ... */ }
func (s *MyCustomSubsystem) Name() string { return "custom" }
```

### 前端开发流程

```bash
# 1. 修改配置后同步
cd web && npm run sync-config

# 2. 启动开发服务器
npm run dev

# 3. 前端会自动连接到 web/config.yaml 中配置的后端
# 4. 修改代码后会自动热更新
```

**添加新页面：**
1. 在 `web/src/pages/` 创建页面组件
2. 在 `web/src/router/index.tsx` 中添加路由
3. 在 `web/config.yaml` 中启用功能开关

### 调试技巧

**查看日志：**
```bash
# 实时查看日志
tail -f logs/shepherd-standalone-$(date +%Y-%m-%d).log

# 或使用 API
curl http://localhost:9190/api/logs/stream
```

**检查模型状态：**
```bash
curl http://localhost:9190/api/models
curl http://localhost:9190/api/processes
```

**查看节点状态（Master 模式）：**
```bash
curl http://localhost:9190/api/nodes
```

## 重要注意事项

### GPU 检测

GPU 检测使用以下工具：
- NVIDIA: `nvidia-smi`
- AMD: `rocm-smi`
- llama.cpp: `llama-bench --list-devices`

如果 GPU 检测失败，检查：
1. ROCm/CUDA 驱动是否正确安装
2. llama.cpp 是否使用正确的后端编译
3. 环境变量 `HIP_VISIBLE_DEVICES` 或 `CUDA_VISIBLE_DEVICES`

### 模型路径

模型路径支持两种配置方式：
```yaml
model:
  paths:                      # 简单数组（向后兼容）
    - "./models"
  path_configs:               # 详细配置
    - path: "~/.cache/huggingface/hub"
      name: "HuggingFace Cache"
      description: "HuggingFace 模型缓存"
```

### 端口分配

模型服务器端口自动分配范围：8081-9000

如需修改：
```go
// internal/port/allocator.go
NewPortAllocator(8081, 9000)  // 修改范围
```

### 优雅关闭

Shepherd 支持优雅关闭，按以下顺序清理资源：
1. 停止接受新连接（HTTP 服务器）
2. 停止 Node（所有角色）
3. 停止所有模型加载和处理
4. 停止所有子进程
5. 关闭日志系统

总超时时间：10 秒

### 安全注意事项

1. **API Key**: 生产环境务必配置 `security.api_key`
2. **CORS**: 默认允许所有来源，生产环境应限制 `security.allowed_origins`
3. **Master API**: 使用 Master 角色时配置 `node.master_role.api_key`
4. **命令白名单**: Client 角色配置 `node.executor.allowed_commands`

### 常见问题

**问题：模型加载失败**
- 检查 llama.cpp 路径配置
- 检查模型文件完整性
- 查看 `/api/models/{id}` 返回的详细信息
- 检查 `/api/processes` 进程状态

**问题：Client 无法连接 Master**
- 检查 `client.master_address` 配置
- 检查 Master 防火墙规则
- 查看日志中的注册错误

**问题：前端无法连接后端**
- 检查 `web/config.yaml` 中的 `backend.urls`
- 使用浏览器开发者工具查看网络请求
- 确认后端 `/api/info` 可访问

**问题：编译错误 undefined: node.CommandTypeTestLlamacpp**
- 这是类型迁移期间的临时问题
- 确保导入 `github.com/shepherd-project/shepherd/Shepherd/internal/types`
- 使用 `types.CommandTypeTestLlamacpp` 而不是 `node.CommandTypeTestLlamacpp`

## 测试

### 测试文件分布

```
internal/
├── api/*_test.go              # API 处理器测试
├── client/*_test.go           # Client 组件测试
├── master/*_test.go           # Master 组件测试
├── node/*_test.go             # Node 核心测试
├── model/*_test.go            # 模型管理测试
├── storage/*_test.go          # 存储层测试
├── config/*_test.go           # 配置管理测试
└── ...
```

**总计**：40+ 测试文件，覆盖核心功能

### 运行单个测试

```bash
# 运行特定包的测试
go test ./internal/model -v

# 运行单个测试函数
go test ./internal/model -run TestScanModels -v

# 运行带覆盖率的测试
go test ./internal/storage -coverprofile=coverage.out
go tool cover -html=coverage.out

# 运行特定包的测试（使用 make）
make test TEST=./internal/config
```

### 前端测试

```bash
cd web

# 运行所有测试（vitest）
npm run test

# 运行带覆盖率的测试
npm run test:coverage

# 监视模式
npm run test -- --watch

# UI 模式（如支持）
npm run test -- --ui
```

**测试框架**：Vitest + Testing Library

### 测试风格

**Go 测试**（table-driven + testify）：
```go
func TestNodeStartStop(t *testing.T) {
    config := &NodeConfig{
        ID:      "test-node-1",
        Role:    NodeRoleStandalone,
        // ...
    }

    node, err := NewNode(config)
    require.NoError(t, err)  // 使用 require 处理设置错误

    assert.Equal(t, "test-node-1", node.GetID())  // 使用 assert 验证结果
}
```

## 版本控制

编译时注入版本信息：
```go
// cmd/shepherd/main.go
var (
    Version   = "dev"       // 通过 `-ldflags` 注入
    BuildTime = "unknown"   // 自动生成
    GitCommit = "unknown"   // Git commit hash
)
```

**Makefile 自动处理**：
```bash
# 使用 Makefile 编译（自动注入版本信息）
make build VERSION=v0.2.0

# 手动编译
go build -ldflags "-X main.Version=v0.2.0 -X main.GitCommit=$(git rev-parse --short HEAD)"
```

**当前版本**：v0.3.1（见 `cmd/shepherd/main.go` 和 `CHANGELOG.md`）

### v0.3.0+ 更新

**HuggingFace 集成**（v0.3.0+）：
- 集成 `go-huggingface/hub` 和 `HuggingFaceModelDownloader` SDK
- 支持基础下载模式（Basic）和高级下载模式（Advanced，支持分块/可恢复）
- 前端添加 HuggingFace 模型搜索和下载 UI
- 支持 HF-Mirror 镜像源加速下载

**llama.cpp 测试**（v0.3.0+）：
- `internal/client/tester` 包：自动检测和测试 llama.cpp 二进制
- 系统信息收集（GPU/CPU/ROCm 版本）
- 前端添加 llama.cpp 可用性测试 UI
- API: `POST /api/nodes/:id/test-llamacpp`

**配置报告**（v0.3.0+）：
- `internal/client/configreport` 包：收集节点配置信息
- llama.cpp 路径配置和可用性
- 模型路径和模型数量统计
- 环境信息（OS/Kernel/Python/Go 版本）
- API: `GET /api/nodes/:id/config`

**日志增强**（v0.3.1+）：
- 所有日志自动包含调用位置（文件名:行号）
- 增强模型加载/卸载过程的日志记录
- 文本格式: `[时间] 级别 [文件:行号] 消息`
- JSON 格式: `{"caller":"文件:行号",...}`

### v0.3.1 模型加载参数

**新增参数**（详见 `doc/api/model-loading.md`）：
- **采样参数**: `reranking`, `minP`, `presencePenalty`, `frequencyPenalty`
- **模板系统**: `disableJinja`, `chatTemplate`, `contextShift`
- **KV 缓存配置**: `kvCacheTypeK`, `kvCacheTypeV`, `kvCacheUnified`, `kvCacheSize`

## 贡献指南

详细的贡献指南请参考：`doc/contributing.md`

**快速流程**：
1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

**代码规范：**
- 遵循 Go 标准代码风格（见 `AGENTS.md`）
- 运行 `make fmt` 格式化代码
- 运行 `make lint` 检查代码
- 添加测试覆盖新功能
- 使用 Conventional Commits 格式：`feat:`, `fix:`, `docs:`, `refactor:`, `test:`

## v0.2.0 迁移说明

**类型统一**：所有跨模块类型已迁移到 `internal/types/`
- 后端使用 `types.NodeInfo` 而不是 `node.NodeInfo` 或 `cluster.Client`
- 前端使用 `UnifiedNode` 而不是 `Client`

**API 路由变更**：
- `/api/master/clients/*` → `/api/nodes/*`
- 旧路由仍可用但已标记为废弃

**默认模式变更**：
- 默认运行模式从 `hybrid` 改为 `standalone`
