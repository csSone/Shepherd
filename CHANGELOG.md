# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Web 前端完全独立架构**
  - 前端现在拥有独立的配置文件 (`web/config.yaml`)
  - 不再依赖后端提供配置，前端可连接任意后端服务器
  - 支持多后端配置和运行时切换
  - 移除 Vite 代理依赖，前端直接连接后端 API
  - 新增 `configLoader.ts` 配置加载器
  - 新增 `web/configTypes.ts` 独立类型定义文件
  - 新增 `web/DEPLOYMENT.md` 部署指南
  - 新增 `web/INDEPENDENT.md` 迁移指南
  - 新增 `scripts/sync-web-config.sh` 配置同步脚本

### Changed
- 后端不再提供 `/api/config/web` 端点（前端独立配置）
- API 客户端支持动态后端 URL 配置
- Vite 配置移除代理设置，前端独立运行
- 后端 CORS 已配置允许所有源访问
- SSE Hook 现在使用完整后端 URL 而非相对路径
- 功能配置：禁用尚未实现的集群管理和日志查看功能

### Removed
- 删除 `config/web.config.yaml`（后端控制的前端配置）
- 删除 `internal/server/server.go` 中的 `handleGetWebConfig` 方法
- 删除后端 `/api/config/web` 路由

### Fixed
- **前端循环依赖** - 创建独立的 `configTypes.ts` 避免循环导入
- **SSE 连接 404 错误** - 修复 SSE Hook 使用相对路径导致连接到前端服务器的问题
- **QueryClient 错误** - 重构 App 组件，确保 useSSE Hook 在 QueryClientProvider 内部使用
- **类型安全** - 修复 useConfig() 函数访问私有成员的问题
- **语法错误** - 修复 configLoader.ts 中的正则表达式语法错误
- 禁用未实现的功能（集群管理、日志查看），避免 404 错误

### Changed
- 后端不再提供 `/api/config/web` 端点（前端独立配置）
- API 客户端支持动态后端 URL 配置
- Vite 配置移除代理设置，前端独立运行
- 后端 CORS 已配置允许所有源访问

### Removed
- 删除 `config/web.config.yaml`（后端控制的前端配置）
- 删除 `internal/server/server.go` 中的 `handleGetWebConfig` 方法
- 删除后端 `/api/config/web` 路由

### Added
- Master-Client 分布式架构支持
- Master-Client 分布式架构支持
  - Master 模式：管理多个 Client 节点
  - Client 模式：作为工作节点执行任务
  - 网络自动发现和节点注册
  - 分布式任务调度（支持轮询、最少负载、资源感知策略）
  - 跨节点日志聚合
  - Conda 环境集成
- 命令行参数支持 (`--mode`, `--version`, `--master-address`)
- 跨平台编译脚本（Linux/macOS/Windows）
- 版本信息注入

### Changed
- 优化日志系统，添加实例方法
- 更新配置验证逻辑，优先检查模式参数
- 完善 README 文档结构

### Fixed
- 修复配置测试中的模式验证问题

## [0.1.0-alpha] - 2026-02-19

### Added
- ✅ GGUF 模型解析器
- ✅ 配置系统（YAML + 热更新）
- ✅ 进程管理器（llama.cpp 进程控制）
- ✅ HTTP 服务器（Gin 框架，多端口支持）
- ✅ 模型管理器（扫描、加载、状态管理）
- ✅ SSE 实时事件广播系统
- ✅ OpenAI API 兼容层
- ✅ Anthropic API 兼容层
- ✅ Ollama API 兼容层
- ✅ LM Studio API 兼容层
- ✅ 下载管理器（HTTP 下载、断点续传）
- ✅ 结构化日志系统（支持文件轮转）

### Changed
- 相比 Java 版本：
  - 启动速度提升 20 倍（<500ms）
  - 内存占用减少 85%（~30MB）
  - 部署体积减少 90%（~15MB）
  - 无需 JVM 运行时

### Tested
- 所有模块单元测试通过
- 覆盖率：GGUF 解析、配置系统、进程管理、HTTP 服务器、模型管理、下载管理、API 兼容层

### Documentation
- 项目概述文档
- 架构设计文档
- 实施路线图
- API 参考文档
- 编译和安装指南
- 贡献指南

## [Future Plans]

### v0.2.0 (Planned)
- [ ] MCP (Model Context Protocol) 支持
- [ ] Web UI 界面
- [ ] 系统托盘（Windows/Linux/macOS）
- [ ] Docker 镜像支持
- [ ] Kubernetes 部署支持

### v0.3.0 (Planned)
- [ ] 模型量化功能
- [ ] 批量模型管理
- [ ] 模型性能基准测试
- [ ] 更多 API 兼容（Gemini、Claude 等）

### v1.0.0 (Planned)
- [ ] 生产就绪
- [ ] 完整的集成测试
- [ ] 性能优化
- [ ] 安全加固
- [ ] 企业级支持
