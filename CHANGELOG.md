# Changelog

All notable changes to Shepherd will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
