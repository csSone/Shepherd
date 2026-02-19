<div align="center">

# 🐏 Shepherd

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build](https://img.shields.io/badge/Build-passing-brightgreen.svg)]()

**高性能轻量级 llama.cpp 模型管理系统**

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [文档](#-文档) • [贡献](#-贡献)

</div>

---

## ✨ 功能特性

### 🚀 核心能力
- **极快启动** - <500ms 启动时间，相比 Java 版本快 20 倍
- **低内存占用** - 仅 ~30MB 内存，相比 Java 版本减少 85%
- **单一二进制** - 无需运行时依赖，开箱即用
- **分布式架构** - 支持 Master-Client 多节点部署

### 📦 模型管理
- 自动扫描 GGUF 格式模型
- 一键加载/卸载，支持多目录管理
- 模型收藏、别名、分卷自动识别
- 视觉模型 (mmproj) 支持

### 🔌 多 API 兼容
| API | 端口 | 状态 |
|-----|------|------|
| OpenAI | `:9190/v1` | ✅ |
| Anthropic | `:9170/v1` | ✅ |
| Ollama | `:11434` | ✅ |
| LM Studio | `:1234` | ✅ |

### 🌐 分布式管理
- **Master 模式** - 管理多个 Client 节点
- **Client 模式** - 作为工作节点执行任务
- **自动发现** - 内网自动扫描和注册 Client
- **任务调度** - 支持轮询、最少负载、资源感知策略
- **Conda 集成** - 使用 Client 端 Python 环境

### 📥 下载管理
- HuggingFace / ModelScope 模型下载
- 断点续传，并发下载（最多 4 任务）
- 实时进度监控

### 🎨 Web 前端
- **React + TypeScript** - 现代化前端技术栈
- **YAML 配置驱动** - 单一配置文件生成所有前端配置
- **实时 UI 更新** - SSE 实时事件推送
- **响应式设计** - 支持桌面和移动端

---

## 📦 快速开始

### 安装

<details>
<summary><b>从源码编译</b></summary>

```bash
# 克隆仓库
git clone https://github.com/shepherd-project/shepherd.git
cd shepherd

# 编译 (支持 Linux/macOS/Windows)
make build
# 或
./scripts/build.sh
```

</details>

<details>
<summary><b>使用 Makefile</b></summary>

```bash
make build        # 编译当前平台
make build-all    # 跨平台编译所有平台
make release      # 打包发布版本
make install      # 安装到系统
```

</details>

<details>
<summary><b>下载预编译版本</b></summary>

前往 [Releases](https://github.com/shepherd-project/shepherd/releases) 下载对应平台的二进制文件。

</details>

### 配置

创建 `config/config.yaml`：

```yaml
# 运行模式: standalone, master, client
mode: standalone

server:
  web_port: 9190

model:
  paths:
    - "./models"
    - "~/.cache/huggingface/hub"
  auto_scan: true
```

### 运行

<details>
<summary><b>使用运行脚本 (推荐)</b></summary>

**Linux/macOS:**

```bash
# 单机模式
./scripts/run.sh standalone

# Master 模式
./scripts/run.sh master --port 9190 --scan

# Client 模式
./scripts/run.sh client --master http://192.168.1.100:9190 --name client-1

# 运行前先编译
./scripts/run.sh master -b

# 查看帮助
./scripts/run.sh --help
```

**Windows:**

```batch
REM 单机模式
scripts\run.bat standalone

REM Master 模式
scripts\run.bat master --port 9190 --scan

REM Client 模式
scripts\run.bat client --master http://192.168.1.100:9190 --name client-1

REM 运行前先编译
scripts\run.bat master -b
```

</details>

<details>
<summary><b>直接使用二进制文件</b></summary>

```bash
# 单机模式 (默认)
./build/shepherd

# Master 模式
./build/shepherd --mode=master

# Client 模式
./build/shepherd --mode=client --master-address=http://master:9190

# 查看版本
./build/shepherd --version
```

</details>

访问 Web UI: http://localhost:9190

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
│   ├── logger/            # 日志系统
│   ├── model/             # 模型管理器
│   ├── process/           # 进程管理
│   ├── server/            # HTTP 服务器
│   └── websocket/         # SSE 实时通信
├── config/                # 配置文件目录
│   ├── config.yaml        # 后端服务器配置
│   └── web/               # Web 前端配置
│       └── config.yaml    # 前端 YAML 配置（单一数据源）
├── scripts/               # 编译和部署脚本
│   └── generate-web-configs.ts  # 前端配置生成器
├── web/                   # Web 前端
│   ├── src/               # React + TypeScript 源码
│   └── [配置文件自动生成] # 由 config.yaml 生成
└── docs/                  # 项目文档
```

---

## 📚 文档

| 文档 | 描述 |
|------|------|
| [编译和安装](docs/06-编译和安装.md) | 详细编译指南 |
| [项目概述](docs/01-项目概述.md) | 项目背景和目标 |
| [架构设计](docs/03-架构设计.md) | 系统架构说明 |
| [实施路线图](docs/04-实施路线图.md) | 开发进度和计划 |
| [API 参考](docs/05-API参考.md) | API 接口文档 |

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

Shepherd Web 前端采用 **YAML 配置驱动** 的方式管理所有配置文件：

```bash
cd web

# 1. 安装依赖
npm install

# 2. 修改配置文件
# 编辑 config/web/config.yaml

# 3. 生成前端配置（自动运行）
npm run gen:config

# 4. 启动开发服务器
npm run dev

# 5. 构建生产版本
npm run build

# 6. 类型检查
npm run type-check

# 7. 代码检查
npm run lint
```

**配置生成流程：**

```
config/web/config.yaml (单一数据源)
         ↓
scripts/generate-web-configs.ts
         ↓
web/ (自动生成的配置文件)
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
└── eslint.config.js
```

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

欢迎贡献！请查看 [贡献指南](CONTRIBUTING.md)。

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
- [x] Master-Client 分布式管理
- [x] Web UI - React + TypeScript 前端
- [ ] MCP (Model Context Protocol) 支持
- [ ] 系统托盘
- [ ] v1.0.0 - 生产就绪

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

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
