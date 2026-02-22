# Changelog

All notable changes to Shepherd will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2026-02-22

### Breaking Changes
- **API 路由统一**: `/api/master/clients/*` 和 `/api/master/nodes/*` 统一为 `/api/nodes/*`
  - 旧路由已标记为废弃，将在 v0.4.0 移除
  - 旧路由返回 `X-API-Deprecation` 和 `X-API-Sunset` 响应头
- **前端类型重构**: `Client` 类型迁移到 `UnifiedNode`
  - `web/src/types/cluster.ts` 标记为 `@deprecated`
  - 建议使用 `web/src/types/node.ts` 中的统一类型

### Added
- **统一类型系统**: `internal/types/node.go` 新增统一节点类型定义
  - `NodeCapabilities` - 节点能力（GPU/CPU/内存/软件支持）
  - `NodeResources` - 节点资源使用情况
  - `NodeInfo` - 统一节点信息（兼容 Master/Client/Node）
  - `HeartbeatMessage` - 统一心跳消息
  - `Command` 和 `CommandResult` - 统一命令结构
- **类型别名系统**: 保持向后兼容
  - `internal/node/types.go`: `NodeInfo` → `types.NodeInfo`
  - `internal/cluster/types.go`: `Client` → `types.NodeInfo`
- **前端统一类型**: `web/src/types/node.ts` 新增 `UnifiedNode` 接口
- **API 响应辅助函数**: `internal/api/response.go` 提供统一响应格式
  - `Success()`, `Error()`, `ValidationError()`, `NotFound()` 等
- **任务类型定义**: `web/src/types/task.ts` 新增任务相关类型

### Changed
- **NodeAdapter 路由重构**: 主路由使用 `/api/nodes/*`
  - `RegisterRoutes()` 方法重构
  - 添加 `registerDeprecatedRoutes()` 处理旧路由
  - 添加 `deprecationWarningMiddleware()` 废弃警告中间件
- **前端类型导出**: `web/src/types/index.ts` 重新组织导出
- **客户端组件更新**: 修复可选字段空值处理

### Fixed
- 修复前端 `ClientCard` 组件可选字段空值访问问题
- 统一前后端类型定义，消除重复

### Technical Details
- **后端类型统一**: 新建 `internal/types/node.go` 作为唯一类型定义来源
- **前端类型统一**: `web/src/types/node.ts` 的 `UnifiedNode` 作为推荐类型
- **向后兼容策略**: 使用类型别名保持现有代码无需修改
- **API 废弃策略**: HTTP 响应头 + 日志警告 + 文档标记

### Migration Guide

**后端迁移**:
```go
// 旧代码
import "github.com/shepherd-project/shepherd/internal/node"
nodeInfo := node.NodeInfo{...}

// 新代码（推荐）
import "github.com/shepherd-project/shepherd/internal/types"
nodeInfo := types.NodeInfo{...}

// 旧代码仍可编译（类型别名）
nodeInfo := node.NodeInfo{...} // 等同于 types.NodeInfo
```

**前端迁移**:
```typescript
// 旧代码
import type { Client } from '@/types/cluster';

// 新代码（推荐）
import type { UnifiedNode } from '@/types/node';

// 旧代码仍可编译（类型别名）
import type { Client } from '@/types/cluster'; // 等同于 UnifiedNode
```

**API 路由迁移**:
```bash
# 旧路由（已废弃）
GET /api/master/clients
GET /api/master/nodes

# 新路由（推荐）
GET /api/nodes
```

## [v0.1.4] - 2026-02-22

### Added
- 模型性能测试页面自动加载默认 llama.cpp 路径的设备信息
- 设备列表解析时添加严格的前缀验证 (ROCm/CUDA/Vulkan/Metal)

### Changed
- 优化 BenchmarkDialog 的 useEffect 执行顺序，确保设备检测自动触发

### Fixed
- 修复设备列表解析错误匹配调试信息导致重复设备的问题
  - llama.cpp --list-devices 输出包含 stderr 调试信息
  - 修复前: 解析了 "ggml_cuda_init: found 1 ROCm devices" 导致重复
  - 修复后: 只解析 "Available devices:" 标记后的正式设备列表
- 修复性能测试对话框打开后设备不自动加载的问题
  - 调整 useEffect 顺序，确保设备检测在初始化之前执行

### Technical Details
- **设备检测改进**: `parseDeviceList()` 现在只在找到 "Available devices:" 标记后开始解析
- **设备验证**: 添加 `validDevicePrefix()` 方法验证设备前缀格式
- **前端优化**: BenchmarkDialog useEffect 顺序重构，确保自动加载

## [v0.1.3] - 2026-02-19

### Added
- 配置管理 API (llama.cpp 和模型路径)
- 下载管理器完整实现
- 进程管理 API
- 脚本重组 (linux/macos/windows)

### Changed
- Web 前端完全独立于后端配置
- 路径配置功能从设置页面移到独立配置

## [v0.1.2] - 2026-02-15

### Added
- Web 前端独立架构
- 前端独立配置文件 (web/config.yaml)
- 多后端支持和运行时切换
- SSE 实时事件推送

## [v0.1.1] - 2026-02-10

### Added
- Master-Client 分布式架构
- 统一 Node 模型
- 心跳和资源监控
- 命令分发和调度

## [v0.1.0-alpha] - 2026-02-01

### Added
- 核心功能实现
- GGUF 模型扫描和管理
- OpenAI/Anthropic/Ollama API 兼容
- Web UI 基础功能
