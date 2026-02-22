# 🐏 Shepherd

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**高性能轻量级分布式 llama.cpp 模型管理系统**

---

## ✨ 核心特性

- **极快启动** - <500ms 启动时间，相比 Java 版本快 20 倍
- **低内存占用** - 仅 ~30MB 内存，相比 Java 版本减少 85%
- **单一二进制** - 无需运行时依赖，开箱即用
- **分布式架构** - 支持 Master-Client 多节点部署
- **多 API 兼容** - OpenAI / Anthropic / Ollama / LM Studio

### 📦 模型管理
- 自动扫描 GGUF 格式模型
- 一键加载/卸载，支持多目录管理
- 模型收藏、别名、分卷自动识别
- 视觉模型 (mmproj) 支持

### 🌐 分布式架构

| 角色 | 说明 | 使用场景 |
|------|------|----------|
| **Standalone** | 单机模式 | 单用户本地部署 |
| **Master** | 主节点，管理其他 Client | 中心化管理集群 |
| **Client** | 工作节点，向 Master 注册 | GPU 工作节点 |
| **Hybrid** | 既是 Master 又是 Client | 分层管理 |

**核心特性：**
- 统一 Node 模型，节点可随时切换角色
- 智能心跳（5秒间隔，自动故障检测）
- 资源上报（CPU/GPU/内存/显存实时监控）
- 智能调度（资源感知、负载均衡）

### 🎨 Web 前端
- React 19 + TypeScript + Vite 7 + Tailwind CSS 4
- 前端独立配置，支持多后端和运行时切换
- SSE 实时事件推送

---

## 📦 快速开始

### 安装

**从源码编译：**
```bash
git clone https://github.com/shepherd-project/shepherd.git
cd shepherd
make build
```

**下载预编译版本：**
前往 [Releases](https://github.com/shepherd-project/shepherd/releases) 下载对应平台的二进制文件。

### 配置

配置文件位置：`config/*.config.yaml`

| 配置文件 | 运行模式 |
|---------|---------|
| `server.config.yaml` | standalone |
| `master.config.yaml` | master |
| `client.config.yaml` | client |

**前端独立配置：** `web/config.yaml` - 支持多后端配置和运行时切换

### 运行

```bash
# 单机模式（默认）
./build/shepherd standalone

# Master 模式
./build/shepherd master

# Client 模式
./build/shepherd client --master-address=http://master:9190

# 查看版本
./build/shepherd --version
```

访问 Web UI: http://localhost:9190

---

## 🌐 分布式部署

### 架构示例

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

### 快速部署

**1. 启动 Master：**
```bash
./build/shepherd master
```

**2. 启动 Client：**
```bash
./build/shepherd client --master-address=http://master:9190
```

**3. 查看集群状态：**
```bash
curl http://master:9190/api/nodes
```

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

### 分布式任务调度

```bash
# 创建任务
curl -X POST http://master:9190/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "run_python",
    "payload": {
      "script": "/path/to/script.py",
      "conda_env": "rocm7.2"
    }
  }'
```

---

## 🛠️ 开发

### 后端

```bash
make test        # 运行测试
make lint        # 代码检查
make fmt         # 代码格式化
make build-all   # 跨平台编译
```

### 前端

```bash
cd web
npm install      # 安装依赖
npm run dev      # 开发服务器（端口 3000）
npm run build    # 构建生产版本
```

---

## 📊 性能对比

| 特性 | Java 版本 | Go 版本 | 改进 |
|------|---------|---------|------|
| 启动时间 | 5-10 秒 | <500ms | **20x** |
| 内存占用 | ~200MB | ~30MB | **-85%** |
| 部署体积 | ~150MB | ~15MB | **-90%** |

---

## 🗺️ 路线图

- [x] v0.1.0 - 核心功能
- [x] v0.1.1 - Master-Client 分布式管理
- [x] v0.1.2 - Web UI 前端独立架构
- [x] v0.1.3 - 配置/下载/进程管理
- [x] v0.1.4 - 模型压测 UI 优化
- [x] **v0.2.0** - **类型系统统一重构**
- [ ] v0.3.0 - 系统托盘和桌面应用
- [ ] v0.4.0 - 移除废弃 API
- [ ] v1.0.0 - 生产就绪

**v0.2.0 更新：** API 路由统一为 `/api/nodes/*`，类型迁移到 `UnifiedNode`。详见 [CHANGELOG.md](CHANGELOG.md)。

---

## 📚 文档

| 文档 | 描述 |
|------|------|
| [CHANGELOG.md](CHANGELOG.md) | 变更日志 |
| [贡献指南](doc/contributing.md) | 贡献指南 |
| [脚本总览](doc/scripts.md) | 脚本文档 |
| [Web 前端部署](doc/web/deployment.md) | 前端部署指南 |
| [Web 前端开发](doc/web/development.md) | 前端开发文档 |

---

## 📄 许可证

Apache License 2.0 - 详见 [LICENSE](LICENSE) 文件

---

## 🙏 致谢

- [llama.cpp](https://github.com/ggerganov/llama.cpp) - 核心推理引擎
- [LlamacppServer](https://github.com/markpublish/LlamacppServer) - 原始 Java 版本

---

## 📞 联系方式

- **问题反馈**: [GitHub Issues](https://github.com/shepherd-project/shepherd/issues)
- **功能建议**: [GitHub Discussions](https://github.com/shepherd-project/shepherd/discussions)

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请点个 Star！**

</div>
