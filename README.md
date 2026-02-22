# 🐏 Shepherd

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Build](https://img.shields.io/badge/Build-passing-brightgreen.svg)]()

**高性能轻量级分布式 llama.cpp ·管理系统**

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [文档](#-文档) • [贡献](#-贡献)

---

## ✨ 功能特性

### 🚀 核心能力
- **极快启动** - <500ms 启动时间，相比 Java 版本快 20 倍
- **低内存占用** - 仅 ~30MB 内存，相比 Java 版本减少 85%
- **单一二进制** - 无需运行时依赖，开箱即用
- **分布式架构** - 支持 Master-Client 多节点部署
- **优雅启停** - 完善的信号处理和资源清理机制
- **智能日志** - 按运行模式分类的日志系统 (shepherd-{mode}-{date}.log)
- **稳定可靠** - 完善的错误处理和降级策略，无 panic 设计

### 📦 模型管理
- 自动扫描 GGUF 格式模型
- 一键加载/卸载，支持多目录管理
- 模型收藏、别名、分卷自动识别
- 视觉模型 (mmproj) 支持
- **llama.cpp 路径配置** - Web UI 配置多个 llama.cpp 路径
- **模型路径配置** - 灵活管理模型扫描路径

### 🔌 多 API 兼容
| API | 端口 | 状态 |
|-----|------|------|
| OpenAI | `:9190/v1` | ✅ |
| Anthropic | `:9170/v1` | ✅ |
| Ollama | `:11434` | ✅ |
| LM Studio | `:1234` | ✅ |

### 🌐 分布式架构 (新)

Shepherd 现在支持统一的 **Node 架构**，每个节点可以灵活地扮演不同角色：

| 角色 | 说明 | 使用场景 |
|------|------|----------|
| **Standalone** | 单机模式，所有功能本地执行 | 单用户本地部署 |
| **Master** | 主节点，管理其他 Client 节点 | 中心化管理集群 |
| **Client** | 工作节点，向 Master 注册并执行命令 | GPU 工作节点 |
| **Hybrid** | 既是 Master 又是 Client | 分层管理，可接入上层 Master |

**核心特性：**
- **统一 Node 模型** - 每个节点可随时切换角色
- **智能心跳** - 5秒间隔，指数退避重连，自动故障检测
- **资源上报** - CPU、GPU、内存、显存、llama.cpp 版本实时上报
- **安全命令** - 白名单验证、签名防篡改、资源限流
- **智能调度** - 资源感知、负载均衡、模型本地性优化
- **多 GPU 支持** - 自动检测 NVIDIA/AMD/Intel GPU

**Master 功能：**
- 节点注册与心跳管理
- 实时资源监控（CPU/GPU/内存/显存）
- 智能任务调度（3种策略）
- 命令下发与结果收集

**Client 功能：**
- 自动向 Master 注册
- 定期心跳上报资源状态
- 接收并执行 Master 命令
- 断线自动重连

### 📝 日志系统
- **按模式分类** - 日志文件名包含运行模式 (shepherd-{mode}-{date}.log)
- **自动轮转** - 支持按日期和文件大小自动轮转
- **备份管理** - 自动清理过期日志文件
- **优雅关闭** - 确保日志在关闭前正确写入

### 🎛️ 运行时配置
- **前端独立配置** - Web 前端拥有独立配置文件 (`web/config.yaml`)
- **多后端支持** - 前端可连接任意后端服务器，支持运行时切换
- **CORS 控制** - 可配置跨域访问策略
- **SSE 支持** - 服务器推送事件实时更新

### 📥 下载管理
- **模型仓库集成** - 支持 HuggingFace 和 ModelScope 模型仓库
  - 浏览仓库中的 GGUF 文件列表
  - 查看文件大小和详细信息
  - 一键下载选定的模型文件
- **智能下载** - 断点续传，并发下载（最多 4 任务）
- **实时进度** - 下载速度、ETA、分块进度显示
- **动态刷新** - 优化轮询策略，仅活跃任务时刷新

- ### 🎨 Web 前端
- - **React + TypeScript** - 现代化前端技术栈
- - **前端版本**: React 19.2.0、Vite 7.x、TypeScript 5.x、Tailwind CSS 4.x
- **独立配置** - 前端拥有独立配置文件，可连接任意后端
- **多后端支持** - 支持配置多个后端地址，运行时切换
- **实时 UI 更新** - SSE 实时事件推送
- **响应式设计** - 支持桌面和移动端

---

## 📦 快速开始

### 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/shepherd-project/shepherd.git
cd shepherd

# 编译 (根据操作系统选择对应脚本)

# Linux
./scripts/linux/build.sh

# macOS
./scripts/macos/build.sh

# Windows
scripts\windows\build.bat

# 或使用 Makefile
make build
```

**更多脚本信息请查看：** [doc/scripts.md](doc/scripts.md)

### 使用 Makefile

```bash
make build        # 编译当前平台
make build-all    # 跨平台编译所有平台
make release      # 打包发布版本
make install      # 安装到系统
```

### 下载预编译版本

前往 [Releases](https://github.com/shepherd-project/shepherd/releases) 下载对应平台的二进制文件。

### 配置

Shepherd 使用 YAML 配置文件，支持三种运行模式：

| 配置文件 | 运行模式 | 说明 |
|---------|---------|------|
| `config/server.config.yaml` | standalone | 单机模式配置 |
| `config/master.config.yaml` | master | Master 节点配置 |
| `config/client.config.yaml` | client | Client 节点配置 |

**示例配置 (server.config.yaml):**

```yaml
# 运行模式
mode: standalone

# 服务器配置
server:
  host: "0.0.0.0"
  web_port: 9190
  read_timeout: 30
  write_timeout: 30

# 模型扫描路径
model:
  paths:
    - "./models"
    - "~/.cache/huggingface/hub"
  auto_scan: true

# 日志配置
log:
  level: "info"         # debug, info, warn, error
  format: "json"        # text, json
  output: "both"        # stdout, file, both
  directory: "logs"
  max_size: 100         # MB
  max_age: 7            # days
```

**Web 前端独立配置 (web/config.yaml):**

```yaml
# 后端服务器配置（可配置多个）
backend:
  urls:
    - "http://localhost:9190"       # 主后端
    - "http://backup:9190"          # 备用后端
  currentIndex: 0                   # 当前使用的后端索引

# 功能开关
features:
  models: true        # 模型管理（已实现）
  downloads: true     # 下载管理（已实现）
  cluster: false      # 集群管理（开发中）
  logs: false         # 日志查看（开发中）
  chat: true          # 聊天功能（已实现）
  settings: true      # 设置页面（已实现）
  dashboard: true     # 仪表盘（已实现）

# UI 配置
ui:
  theme: "auto"
  language: "zh-CN"
  pageSize: 20
```

前端现在完全独立运行，不依赖后端配置。详见 [doc/web/deployment.md](doc/web/deployment.md) 和 [doc/web/development.md](doc/web/development.md)。

**功能状态说明：**
- ✅ 已实现：模型管理、下载管理、聊天、设置、仪表盘、路径配置
- 🔜 开发中：集群管理、日志查看（将在后续版本实现）

### 路径配置功能

Shepherd 支持通过 Web UI 灵活配置 llama.cpp 和模型路径：

**llama.cpp 路径配置:**
- 在设置页面配置多个 llama.cpp 安装路径
- 支持自定义名称和描述
- 路径有效性自动验证
- 适用于多 llama.cpp 环境管理

**模型路径配置:**
- 配置多个模型扫描目录
- 支持自定义名称和描述
- 自动扫描和发现 GGUF 模型
- 便于组织和管理分散的模型文件

**API 端点:**
```bash
# llama.cpp 路径管理
GET    /api/config/llamacpp/paths          # 获取所有路径
POST   /api/config/llamacpp/paths          # 添加路径
DELETE /api/config/llamacpp/paths          # 删除路径
POST   /api/config/llamacpp/test           # 测试路径有效性

# 模型路径管理
GET    /api/config/models/paths            # 获取所有路径
POST   /api/config/models/paths            # 添加路径
PUT    /api/config/models/paths            # 更新路径
DELETE /api/config/models/paths            # 删除路径
```

### 运行

<details>
<summary><b>位置参数方式 (推荐)</b></summary>

```bash
# 单机模式 (默认)
./build/shepherd standalone

# Master 模式
./build/shepherd master

# Client 模式
./build/shepherd client --master-address=http://master:9190

# 查看版本
./build/shepherd --version
```

</details>

<details>
<summary><b>使用运行脚本</b></summary>

**Linux:**

```bash
# 单机模式
./scripts/linux/run.sh standalone

# Master 模式
./scripts/linux/run.sh master

# Client 模式
./scripts/linux/run.sh client --master http://192.168.1.100:9190 --name client-1

# 运行前先编译
./scripts/linux/run.sh standalone -b

# 查看帮助
./scripts/linux/run.sh --help
```

**macOS:**

```bash
# 单机模式
./scripts/macos/run.sh standalone

# Master 模式
./scripts/macos/run.sh master

# Client 模式
./scripts/macos/run.sh client --master http://192.168.1.100:9190 --name client-1

# 运行前先编译
./scripts/macos/run.sh standalone -b

# 跳过 Gatekeeper 验证
./scripts/macos/run.sh standalone --no-gatekeeper
```

**Windows:**

```batch
REM 单机模式
scripts\windows\run.bat standalone

REM Master 模式
scripts\windows\run.bat master

REM Client 模式
scripts\windows\run.bat client --master http://192.168.1.100:9190 --name client-1

REM 运行前先编译
scripts\windows\run.bat standalone -b
```

**详细文档:** [doc/scripts.md](doc/scripts.md)

</details>

<details>
<summary><b>优雅关闭</b></summary>

Shepherd 支持优雅关闭，按正确顺序清理资源：

```bash
# 发送 SIGTERM (Ctrl+C)
kill -TERM <pid>

# 或发送 SIGINT
kill -INT <pid>

# 系统会按以下顺序关闭：
# 1. 停止接受新连接 (HTTP 服务器)
# 2. 停止所有模型加载和处理
# 3. 停止所有子进程
# 4. 关闭日志系统
# 总超时时间: 10 秒
```

</details>

<details>
<summary><b>前端开发服务器（独立模式）</b></summary>

```bash
# 启动前端开发服务器 (端口 3000)
cd web
npm run dev

# 或使用脚本 (根据操作系统选择)

# Linux
./scripts/linux/web.sh dev

# macOS
./scripts/macos/web.sh dev

# Windows
scripts\windows\web.bat dev

# 前端会从 web/config.yaml 读取后端配置
# 可连接到任意后端服务器
```

**前端独立运行的优势：**
- 前端完全独立，可部署到任意服务器
- 支持连接任意后端服务器（无需代理）
- 多后端配置，运行时切换
- 开发模式更简单，无需等待后端启动

</details>

访问 Web UI: http://localhost:3000 (开发模式) 或 http://localhost:9190 (后端托管)

**日志文件位置：**
```
logs/shepherd-standalone-2026-02-19.log    # 单机模式
logs/shepherd-master-2026-02-19.log       # Master 模式
logs/shepherd-client-2026-02-19.log       # Client 模式
```

---

## 🌐 分布式部署

### 架构概述

Shepherd 支持灵活的分布式部署，每个节点可以独立运行或组成集群：

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Master Node   │◄────┤  Hybrid Node    │◄────┤  Client Node    │
│   (Port 9190)   │     │ (Port 9190+9191)│     │                 │
└────────┬────────┘     └────────┬────────┘     └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐     ┌─────────────────┐
│  Client Node 1  │     │  Client Node 2  │
│   (GPU Server)  │     │   (GPU Server)  │
└─────────────────┘     └─────────────────┘
```

### 部署模式

#### 1. Standalone 模式（默认）

单机运行，所有功能本地执行：

```bash
./shepherd standalone
# 或
./shepherd  # 默认为 standalone
```

#### 2. Master 模式

作为集群管理中心：

```bash
./shepherd master --port 9190
```

Master 节点提供：
- Web UI 管理界面
- RESTful API (`/api/master/*`)
- 节点注册和心跳管理
- 任务调度和分发

#### 3. Client 模式

作为工作节点连接 Master：

```bash
./shepherd client --master http://192.168.1.100:9190
```

Client 节点功能：
- 向 Master 注册并定期心跳
- 上报资源信息（CPU/GPU/内存）
- 接收并执行 Master 下发的命令
- 断线自动重连

#### 4. Hybrid 模式（高级）

同时作为 Master 和 Client，支持分层管理：

```bash
./shepherd hybrid \
  --port 9190 \
  --upstream-master http://10.0.0.1:9190
```

适用场景：
- 多层级集群管理
- 区域 Master 汇聚到中心 Master
- 复杂网络拓扑

### 配置文件示例

#### Master 配置 (`config/master.config.yaml`)

```yaml
node:
  id: "master-01"
  name: "Central Master"
  role: "master"
  
  master:
    enabled: true
    port: 9190
    api_key: "your-secret-key"
    
  resources:
    monitor_interval: 5
    report_gpu: true
```

#### Client 配置 (`config/client.config.yaml`)

```yaml
node:
  id: "auto"  # 自动生成
  name: "GPU Server 1"
  role: "client"
  
  client:
    enabled: true
    master_address: "http://192.168.1.100:9190"
    heartbeat_interval: 5
    heartbeat_timeout: 15
    
  executor:
    max_concurrent: 4
    allowed_commands:
      - load_model
      - unload_model
      - run_llamacpp
```

### 完整部署示例

**场景：3 节点 GPU 集群**

1. **启动 Master** (管理节点):
```bash
# Node: 192.168.1.100
./shepherd master --port 9190
```

2. **启动 Client 1** (GPU 服务器 1):
```bash
# Node: 192.168.1.101
./shepherd client \
  --master http://192.168.1.100:9190 \
  --name "GPU-Server-1"
```

3. **启动 Client 2** (GPU 服务器 2):
```bash
# Node: 192.168.1.102
./shepherd client \
  --master http://192.168.1.100:9190 \
  --name "GPU-Server-2"
```

4. **验证集群状态**:
```bash
curl http://192.168.1.100:9190/api/master/nodes
```

### 安全建议

1. **API Key 认证**: 生产环境务必配置 `api_key`
2. **TLS 加密**: 使用 HTTPS 保护通信
3. **防火墙**: 仅开放必要的端口
4. **资源限制**: 配置 `max_concurrent` 防止过载

---

## 💡 使用示例

### OpenAI API

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:9190/v1",
    api_key="dummy"
)

response = client.chat.completions.create(
    model="llama-2-7b-chat",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

### Master-Client 分布式部署

```bash
# 1. 启动 Master 节点
./shepherd --mode=master

# 2. 在其他机器启动 Client 节点
./shepherd --mode=client --master-address=http://master:9190

# 3. 查看集群状态
curl http://master:9190/api/master/clients

# 4. 创建调度任务
curl -X POST http://master:9190/api/master/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "run_python",
    "payload": {
      "script": "/path/to/script.py",
      "conda_env": "rocm7.2"
    }
  }'
```

### SSE 实时事件监听

```javascript
const eventSource = new EventSource('http://localhost:9190/api/events');

eventSource.addEventListener('log', (e) => {
  const data = JSON.parse(e.data);
  console.log(`[LOG] ${data.message}`);
});
```

---

## 🏗️ 项目结构

```
Shepherd/
├── cmd/shepherd/          # 主程序入口
├── internal/
│   ├── api/               # OpenAI/Anthropic/Ollama API
│   ├── cluster/           # Master-Client 分布式管理
│   ├── client/            # Client 端组件
│   ├── config/            # 配置管理
│   ├── download/          # 下载管理器
│   ├── gguf/              # GGUF 模型解析
│   ├── logger/            # 日志系统 (按模式分类)
│   ├── modelrepo/         # 模型仓库客户端 (HuggingFace/ModelScope)
│   ├── model/             # 模型管理器
│   ├── process/           # 进程管理
│   ├── server/            # HTTP 服务器 (优雅关闭)
│   ├── shutdown/          # 优雅关闭管理器
│   └── websocket/         # SSE 实时通信
├── config/                # 配置文件目录
│   ├── server.config.yaml    # 单机模式配置
│   ├── master.config.yaml    # Master 模式配置
│   └── client.config.yaml    # Client 模式配置
├── scripts/               # 编译和部署脚本
├── web/                   # Web 前端（独立配置）
│   ├── config.yaml          # 前端独立配置文件
│   ├── public/
│   │   └── config.yaml      # 配置副本（自动同步）
│   ├── src/               # React + TypeScript 源码
│   │   └── lib/
│   │       ├── configLoader.ts  # 配置加载器
│   │       ├── api/
│   │       │   └── downloads.ts  # 下载 API 客户端
│   │       └── features/
│   │           └── downloads/    # 下载功能
│   │               └── hooks.ts  # 下载 hooks（动态轮询）
│   ├── components/
│   │   └── downloads/
│   │       └── CreateDownloadDialog.tsx  # 文件浏览器 UI
│   ├── DEPLOYMENT.md          # 部署指南
│   ├── DEVELOPMENT.md         # 开发文档
│   └── [开发工具配置]         # TypeScript/Vite/ESLint 等
├── logs/                  # 日志目录 (自动创建)
│   ├── shepherd-standalone-*.log
│   ├── shepherd-master-*.log
│   └── shepherd-client-*.log
└── doc/                   # 项目文档
```

---

## 🔒 稳定性和性能

### 稳定性保障

Shepherd 采用了多层防护确保系统稳定运行：

| 问题类型 | 防护措施 | 状态 |
|---------|---------|------|
| **API 超时** | HTTP 客户端 5 秒连接超时，10 秒总超时 | ✅ |
| **Logger Panic** | 空指针检查 + 错误降级到 stderr | ✅ |
| **资源泄漏** | 完善的 defer 清理和优雅关闭 | ✅ |
| **并发竞态** | Mutex 保护共享状态 | ✅ |
| **前端卡死** | AbortSignal 自动取消 + 防抖机制 | ✅ |

### 性能优化

- **动态轮询** - 仅在有活跃任务时轮询，节省资源
- **请求缓存** - 5 分钟 staleTime + 10 分钟 gcTime
- **连接池** - HTTP 客户端复用连接，100 最大空闲连接
- **Keep-Alive** - 30 秒保持连接，减少握手开销

### 监控和日志

- **按模式分类** - 日志文件名包含运行模式，便于排查
- **自动轮转** - 按日期和大小自动轮转日志
- **SSE 实时推送** - 事件实时通知，无需轮询

### 已知问题修复

```diff
- 修复前: API 请求挂起 30 秒，前端卡死
+ 修复后: 5 秒快速超时，前端响应流畅

- 修复前: Logger 空指针导致 panic，SSE 连接崩溃
+ 修复后: 安全降级，服务持续运行

- 修复前: 每次按键触发 API 请求
+ 修复后: repoId > 3 字符才触发，支持取消
```

---

## 📚 文档

| 文档 | 描述 |
|------|------|
| [贡献指南](doc/contributing.md) | 贡献指南 |
| [安全策略](doc/security.md) | 安全策略 |
| [AI 代理指南](doc/agents.md) | AI 编码代理指南 |
| [脚本总览](doc/scripts.md) | 脚本文档总览 |
| [Web 前端部署](doc/web/deployment.md) | 前端部署指南 |
| [Web 前端开发](doc/web/development.md) | 前端开发文档 |

---

## 🛠️ 开发

### 环境要求

**后端开发:**
- Go 1.25+
- Git

**前端开发:**
- Node.js 18+
- npm 或 yarn

### Web 前端开发

```bash
cd web

# 1. 安装依赖
npm install

# 2. 配置前端（可选）
# 编辑 web/config.yaml 指定后端地址

# 3. 启动开发服务器
npm run dev

# 4. 同步配置（修改 config.yaml 后）
./scripts/sync-web-config.sh

# 5. 构建生产版本
npm run build

# 6. 类型检查
npm run type-check

# 7. 代码检查
npm run lint
```

**前端独立配置：**

前端现在使用独立的配置文件 `web/config.yaml`，不依赖后端：

```yaml
# 后端服务器配置（支持多个）
backend:
  urls:
    - "http://localhost:9190"
    - "http://backup:9190"
  currentIndex: 0

# 功能开关
features:
  models: true        # 模型管理（已实现）
  downloads: true     # 下载管理（已实现）
  cluster: false      # 集群管理（开发中）
  logs: false         # 日志查看（开发中）
  chat: true          # 聊天功能（已实现）
  settings: true      # 设置页面（已实现）
  dashboard: true     # 仪表盘（已实现）

# UI 配置
ui:
  theme: "auto"
  language: "zh-CN"
```

**功能状态说明：**
- ✅ 已实现：模型管理、下载管理、聊天、设置、仪表盘
- 🔜 开发中：集群管理、日志查看（将在后续版本实现）

**架构优势：**
- ✅ 前端完全独立，可连接任意后端
- ✅ 无需 Vite 代理，开发更简单
- ✅ 支持多后端配置和运行时切换
- ✅ 后端仅提供数据 API

**开发工具配置：**
├── postcss.config.js
└── eslint.config.js
```

**开发工具配置：**

- **TypeScript:** `tsconfig.json`, `tsconfig.app.json`, `tsconfig.node.json`
- **Vite:** `vite.config.ts`
- **Tailwind:** `tailwind.config.js`
- **PostCSS:** `postcss.config.js`
- **ESLint:** `eslint.config.js`

这些配置文件直接写在 `web/` 目录中，不需要额外生成。

**技术栈:**
- **构建工具:** Vite 7.x
- **框架:** React 19 + TypeScript 5.x
- **路由:** React Router v7
- **状态管理:** Zustand + React Query
- **UI 组件:** Tailwind CSS 4.x + shadcn/ui
- **Markdown:** react-markdown + remark-gfm + rehype-highlight

### 后端开发命令

```bash
# 运行测试
make test

# 代码检查
make lint

# 代码格式化
make fmt

# 跨平台编译
make build-all

# 清理构建文件
make clean
```

### 贡献指南

欢迎贡献！请查看 [贡献指南](doc/contributing.md)。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📊 性能对比

| 特性 | Java 版本 | Go 版本 | 改进 |
|------|---------|---------|------|
| 启动时间 | 5-10 秒 | <500ms | **20x** |
| 内存占用 | ~200MB | ~30MB | **-85%** |
| 部署体积 | ~150MB | ~15MB | **-90%** |
| 部署方式 | 需要 JVM | 单一二进制 | 更简单 |

---

## 🗺️ 路线图

- [x] v0.1.0-alpha - 核心功能 (M1-M9)
- [x] v0.1.1 - Master-Client 分布式管理
- [x] v0.1.2 - Web UI 前端独立架构
- [x] v0.1.3 - 配置/下载/进程管理 API + 脚本重组
- [x] v0.1.4 - 模型压测 UI 优化和设备检测修复
- [ ] v0.2.0 - MCP (Model Context Protocol) 支持
- [ ] v0.3.0 - 系统托盘和桌面应用
- [ ] v1.0.0 - 生产就绪

---

## 版本对照表

详细版本信息请参见 [VERSIONS.md](./VERSIONS.md)。

**当前版本概览:**
- **前端**: React 19.2.0, Vite 7.x, TypeScript 5.x, Tailwind CSS 4.x
- **后端**: Go 1.25.7
- **Node.js**: 18+

**最近更新 (Unreleased):**
- ✅ 国际化 (i18n) 支持 - 中英文切换
- ✅ WebSocket 实时通信 - 自动重连和心跳检测
- ✅ YAML 解析升级 - 使用 js-yaml 标准库
- ✅ 单元测试架构 - Vitest + React Testing Library
- ✅ 版本文档统一 - 标准化版本管理

## 📄 许可证

本项目采用 Apache License 2.0 - 详见 [LICENSE](LICENSE) 文件。

---

## 🙏 致谢

- [llama.cpp](https://github.com/ggerganov/llama.cpp) - 核心推理引擎
- [LlamacppServer](https://github.com/markpublish/LlamacppServer) - 原始 Java 版本
- 所有第三方库的贡献者

---

## 📞 联系方式

- **问题反馈**: [GitHub Issues](https://github.com/shepherd-project/shepherd/issues)
- **功能建议**: [GitHub Discussions](https://github.com/shepherd-project/shepherd/discussions)

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请点个 Star！**

</div>
