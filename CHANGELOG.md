# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
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
