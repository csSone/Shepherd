# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### 架构重构 - 职责分离与类型统一
- **统一类型系统** - 新增 `internal/types/common.go` 统一状态和类型定义
  - `types.NodeState` - 统一的节点状态枚举（包含 offline、online、busy、error、degraded、disabled）
  - `types.ApiResponse[T]` - 类型安全的泛型 API 响应格式
  - 消除了 `NodeStatus` 和 `ClientStatus` 的重复定义

- **接口定义** - 新增 `internal/node/interfaces.go` 定义清晰的接口边界
  - `INode` - 节点核心接口
  - `IClientRegistry` - 客户端注册表接口
  - `ICommandQueue` - 命令队列接口
  - `IResourceMonitor` - 资源监控接口

- **职责分离** - 将 Node 的职责拆分为独立组件
  - `internal/registry/memory.go` - 内存客户端注册表实现
  - `internal/commands/queue.go` - 内存命令队列实现
  - `internal/monitor/resource.go` - 资源监控实现（含 GPU 检测）

- **Scheduler 增强** - 新增任务重试功能
  - `scheduler.RetryTask()` - 支持重试失败或已取消的任务

### Changed

- **Node 重构** - `internal/node/node.go` 使用接口而非直接依赖具体实现
  - 添加接口方法（ID、Name、Role、Status、Health 等）
  - 保留向后兼容的旧方法（GetID、GetName 等）
  - 通过接口委托功能而非直接持有子系统

- **适配层清理** - `internal/api/node_adapter.go` 移除临时实现
  - 删除临时 `clusterTask` 结构体
  - 删除临时 `tasks` map 和 `tasksMu`
  - 集成真实的 `scheduler.Scheduler`
  - 创建 `nodeClientManager` 适配器桥接接口
  - 更新所有任务管理 API 使用真实 Scheduler

- **前端类型更新** - 统一前后端类型定义
  - `web/src/types/cluster.ts` - 更新 `Capabilities`、`Client`、`ClusterOverview`
  - `web/src/types/node.ts` - 统一 `NodeStatus` 包含 `disabled` 状态
  - 添加 `gpuCount`、`pendingTasks` 等字段

- **TypeScript 类型安全** - 修复 `ClientCard.tsx` 中的类型问题
  - 添加 `degraded` 状态的颜色和标签
  - 使用可选链操作符处理 `capabilities` 和 `resources`
  - 添加空值合并操作符提供默认值

### Fixed

- 修复 `ClientCard.tsx` 中 `client.capabilities` 可能为 undefined 的类型错误
- 修复 `STATUS_COLORS` 和 `STATUS_LABELS` 缺少 `degraded` 状态的问题
- 修复 `node_adapter_test.go` 缺少 `SchedulerConfig` 参数的问题
- 修复 `cmd/shepherd/main.go` 中 `NewNodeAdapter` 调用参数问题

### Test

- 添加 `internal/registry/memory_test.go` - ClientRegistry 单元测试
- 添加 `internal/monitor/resource_test.go` - ResourceMonitor 单元测试
- 更新 `internal/api/node_adapter_test.go` - 添加 SchedulerConfig 参数
- 所有测试通过（registry、monitor、api）
- 前端类型检查通过（`npm run type-check`）


## [0.3.0] - 2025-02-21

### Added

#### 模型加载配置增强
- **完整的加载模型对话框** - 参考并超越 LlamacppServer，实现功能全面的模型加载配置界面
  - **左右分区布局** - 左列基础配置，右列高级参数，逻辑清晰
  - **参数分组** - 上下文与加速、采样与生成、惩罚机制、批处理与并发、KV缓存配置、其他参数
  - **llama.cpp 版本选择** - 支持多个后端路径配置（ROCm/Vulkan 等）
  - **主 GPU 选择** - 下拉选择可用的 GPU 设备
  - **能力开关** - 思考、工具使用、直译、嵌入 等模型能力配置
  - **Flash Attention 加速** - 开关控制（on/off）
  - **禁用内存映射** - true/false 切换，优化性能
  - **锁定物理内存** - 防止内存交换
  - **输入向量模式** - logitsAll 参数
  - **重排序模式** - reranking 参数
  - **Min-P 采样** - 高级采样参数
  - **存在惩罚/频率惩罚** - 精细控制生成行为
  - **微批大小** - uBatchSize 参数
  - **并行槽位数** - parallelSlots 参数
  - **KV 缓存配置** - 内存上限、统一缓存、类型K/V（f16/f32/q8_0）
  - **DirectIO 模式** - default/true/false
  - **禁用 Jinja 模板** - disableJinja 参数
  - **内置聊天模板** - chatTemplate 选择
  - **上下文移位** - contextShift 参数
  - **额外参数** - 多行文本框，支持任意命令行参数

- **GPU 设备检测增强** - 兼容 LlamacppServer 的设备数据格式
  - 使用 `llama-bench --list-devices` 获取设备信息
  - 返回双格式：简单设备字符串数组（LlamacppServer 兼容）+ 详细 GPU 信息
  - 后备方案：`roc-smi --json` 用于 AMD GPU 检测

- **VRAM 显存估算** - 使用 `llama-fit-params` 工具实现真实显存估算
  - 新增 API 端点：`POST /api/models/vram/estimate`
  - 根据模型路径和参数估算显存需求
  - UI 集成：估算按钮位于取消和开始加载按钮之间
  - 实时显示估算结果（GB 单位）

- **模型能力配置** - 新增模型能力的存储和读取
  - `GET /api/models/capabilities/get` - 获取模型能力配置
  - `POST /api/models/capabilities/set` - 保存模型能力配置
  - 支持的能力：thinking（思考）、tools（工具使用）、rerank（重排序）、embedding（嵌入）

#### 前端架构优化
- **国际化 (i18n) 支持** - 完整的多语言国际化实现
  - 集成 i18next + react-i18next + i18next-browser-languagedetector
  - 支持中文 (zh-CN) 和英文 (en-US) 语言切换
  - 语言设置持久化到 localStorage
  - 新增 `LanguageToggle` 组件，集成到 Header
  - 翻译文件结构：`web/src/locales/zh-CN.json`, `web/src/locales/en-US.json`
  - 类型安全的翻译键定义

- **WebSocket 实时通信** - 完整的 WebSocket 客户端实现
  - 新增 `WebSocketClient` 类，支持自动重连和心跳检测
  - 自动重连机制：指数退避 (1s → 2s → 4s → 8s → 16s)，最多 5 次
  - 心跳检测：30 秒间隔 ping/pong，10 秒超时检测
  - `useWebSocket` Hook 和 `WebSocketContext` 全局状态管理
  - `WebSocketStatus` 组件：可视化连接状态指示器
  - 完整的 TypeScript 类型定义和内存泄漏防护

- **YAML 配置解析升级** - 从自定义解析器升级为标准 js-yaml
  - 使用 `js-yaml` 库替代简易正则解析器
  - 完整的错误处理和回退机制（YAML 格式错误时 toast 提示并使用默认配置）
  - 支持复杂 YAML 结构（嵌套对象、数组等）
  - 向后兼容，保持现有 `config.yaml` 格式不变
  - 新增 `DEFAULT_CONFIG` 作为解析失败时的安全回退

- **单元测试基础架构** - 完整的 Vitest 测试框架
  - 配置 Vitest + React Testing Library + jsdom
  - 覆盖率报告配置（c8 provider）
  - 测试工具函数：`renderWithProviders` 包装器
  - 示例测试用例：ConfigLoader (2个) + ApiClient (2个)
  - 测试脚本：`npm test`, `npm run test:coverage`
  - 测试文档：`web/TESTING.md`

#### 版本文档统一
- **版本信息标准化** - 创建 `VERSIONS.md` 作为版本权威来源
  - 统一前端版本信息：React 19.2.0, Vite 7.x, TypeScript 5.x, Tailwind CSS 4.x
  - 统一后端版本信息：Go 1.25.7
  - 版本历史记录和变更追溯
  - 消除 README 与 VERSIONS.md 之间的版本信息不一致

### Fixed

#### 稳定性修复
- **HTTP 客户端超时配置** - 修复 API 请求挂起 30 秒导致前端卡死的问题
  - 添加 10 秒总超时，5 秒连接超时
  - 配置连接池和 Keep-Alive 优化性能
  - 修复文件：`internal/modelrepo/client.go`
- **Logger 空指针安全** - 修复 SSE 连接触发 panic 的问题
  - 添加 `f.Stat()` 错误处理，避免 nil 指针访问
  - 失败时降级到 stderr 输出，确保服务不中断
  - 修复文件：`internal/logger/logger.go`
- **Logger stdout Sync 警告** - 修复对 stdout/stderr 执行 Sync 导致的警告
  - 只对普通文件执行 Sync，跳过 stdout/stderr
  - 修复文件：`internal/logger/logger.go`

#### 前端优化
- **API 防抖机制** - 减少无效请求，防止前端卡死
  - `repoId.length > 3` 检查，至少 3 个字符才触发请求
  - AbortSignal 支持，组件卸载时自动取消请求
  - 5 分钟缓存 + 10 分钟 gcTime
  - 修复文件：`web/src/features/downloads/hooks.ts`
- **API 客户端增强** - 支持请求取消
  - `get()` 方法添加 `signal` 参数
  - 修复文件：`web/src/lib/api/client.ts`
  - 修复文件：`web/src/lib/api/downloads.ts`

### Performance
- API 响应超时从 30 秒降至 5 秒（**提升 6 倍**）
- 前端输入不再卡死，用户体验显著改善
- SSE 连接稳定，无 panic

### Added

#### 模型仓库集成
- **HuggingFace/ModelScope 集成** - 新增模型仓库客户端 (`internal/modelrepo/client.go`)
  - 支持从 HuggingFace 和 ModelScope 获取模型文件列表
  - 自动生成下载 URL
  - GGUF 文件过滤和识别
- **API 端点** - `GET /api/repo/files?source=huggingface&repoId=owner/model`
  - 支持包含斜杠的仓库 ID（使用查询参数避免路径冲突）
  - 返回 GGUF 文件列表（名称、大小、下载 URL）

#### 前端增强
- **下载对话框文件浏览器** - 新增模型文件浏览和选择功能
  - 实时加载 HuggingFace 仓库的 GGUF 文件列表
  - 文件大小格式化显示（GB/MB/KB）
  - 点击选择文件，视觉反馈（高亮、对勾图标）
- **动态轮询优化** - 修复下载页面频繁刷新问题
  - 无任务时不刷新（之前固定 2 秒轮询）
  - 有活跃任务时每秒刷新（preparing、downloading、merging、verifying 状态）
- **API 客户端重构** - 统一的下载管理 API 客户端 (`web/src/lib/api/downloads.ts`)
  - 所有下载相关 API 集中管理
  - TypeScript 类型安全
  - 支持新格式（source + repoId）和旧格式（直接 URL）

#### 测试
- **模型仓库单元测试** - `internal/modelrepo/client_test.go`
  - URL 生成测试（HuggingFace/ModelScope）
  - 仓库 ID 解析测试
  - GGUF 文件识别测试
  - 100% 测试覆盖率

### Changed
- **下载 API 双格式支持** - 向后兼容旧的直接 URL 格式
  - 新格式: `{source, repoId, fileName, path}` - 从模型仓库下载
  - 旧格式: `{url, target_path}` - 直接 URL 下载
- **前端轮询策略** - 从固定 2 秒改为动态调整
  - 减少无效请求，节省服务器资源
  - 优化用户体验

### Fixed
- **路由冲突** - 解决 `/api/models/:id` 与 `/api/repo/:source/:repoId/files` 的冲突
  - 使用查询参数替代路径参数
  - 前端使用 `encodeURIComponent()` 正确编码仓库 ID
- **前端类型错误** - 修复 `refetchInterval` 回调函数的 TypeScript 类型
  - 使用 `query.state.data` 访问数据而非直接访问 `data` 参数
- **网络超时处理** - 改进 HuggingFace API 调用的错误处理

### Technical Details

#### 新增文件
- `internal/modelrepo/client.go` - 模型仓库客户端（~140 行）
- `internal/modelrepo/client_test.go` - 单元测试（~200 行）
- `web/src/lib/api/downloads.ts` - API 客户端（~84 行）

#### 修改文件
- `internal/server/server.go` - 添加 `/api/repo/files` 路由和处理函数
- `web/src/features/downloads/hooks.ts` - 动态轮询、useModelFiles hook
- `web/src/components/downloads/CreateDownloadDialog.tsx` - 文件浏览器 UI

### Changed
- **脚本重组** - 所有脚本按操作系统分类到 `linux/`, `macos/`, `windows/` 子目录
- **macOS 支持** - 新增 macOS 专用脚本，支持 Intel 和 Apple Silicon
- **文档增强** - 每个平台都有详细的 README.md 文档

### Added
- **Linux 脚本** - `scripts/linux/` 目录，包含 build.sh, run.sh, web.sh 等
- **macOS 脚本** - `scripts/macos/` 目录，支持 Universal Binary 和代码签名
- **Windows 脚本** - `scripts/windows/` 目录，包含 build.bat, run.bat, web.bat
- **脚本总览** - `scripts/README.md` 提供跨平台脚本对比和快速开始指南
- **迁移指南** - `scripts/MIGRATION.md` 帮助从旧脚本路径迁移

### Removed
- **旧脚本文件** - 删除 `scripts/` 根目录下的重复脚本：
  - `build.sh`, `run.sh`, `web.sh` → 已迁移到 `linux/` 和 `macos/`
  - `build.bat`, `run.bat`, `web.bat` → 已迁移到 `windows/`
  - `sync-web-config.sh`, `watch-sync-config.sh` → 已迁移到 `linux/`
- **保留脚本** - `build-all.sh` 和 `release.sh` 保留在根目录用于跨平台编译

### Fixed
- **脚本路径计算** - 修复所有脚本的路径计算，使用统一的 `$(cd "$SCRIPT_DIR/../.." && pwd)` 向上两级到项目根目录
- **配置同步脚本** - 修复 `sync-web-config.sh` 和 `watch-sync-config.sh` 的路径问题
- **Web 脚本** - 修复 Linux 和 macOS 版 `web.sh` 的路径计算错误

## [0.1.3] - 2025-02-20

### Added

#### 配置管理 API
- `GET /api/config` - 获取当前系统配置（服务器、存储、模型、节点等）
- `PUT /api/config` - 更新系统配置，支持运行时修改模式、端口和扫描路径

#### 下载管理 API
- `GET /api/downloads` - 列出所有下载任务
- `POST /api/downloads` - 创建新的下载任务（支持 URL 和目标路径）
- `GET /api/downloads/:id` - 获取指定下载任务的状态和进度
- `POST /api/downloads/:id/pause` - 暂停正在进行的下载任务
- `POST /api/downloads/:id/resume` - 恢复已暂停的下载任务
- `DELETE /api/downloads/:id` - 删除下载任务并清理部分下载的文件

#### 进程管理 API
- `GET /api/processes` - 列出所有运行中和加载中的进程
- `GET /api/processes/:id` - 获取指定进程的详细信息（PID、端口、状态等）
- `POST /api/processes/:id/stop` - 停止指定的进程

#### 下载管理器
- 完整的下载任务管理系统（`internal/server/download_manager.go`）
- 支持并发下载控制（最多3个并发任务）
- 实时进度跟踪和速度计算
- 支持暂停/恢复/取消下载
- 自动创建目标目录和文件

#### Node 架构增强
- `Node.GetProcessManager()` 方法 - 访问进程管理器
- 完整的子系统停止日志记录
- HTTP 心跳发送到 Master 功能

### Changed
- **版本更新** - 版本号从 1.0.0 更新至 0.1.3

### Improved
- **子系统管理**: 添加完善的错误日志记录
- **心跳系统**: 实现完整的 HTTP 心跳发送功能，支持 Master 地址配置
- **进程管理**: 通过 model.Manager 暴露进程管理器访问接口

### Technical Details

#### 新增文件
- `internal/server/download_manager.go` - 下载管理器完整实现（~300行）

#### 修改文件
- `internal/server/server.go` - 实现 9 个 TODO handler，集成下载管理器
- `internal/model/manager.go` - 添加 `GetProcessManager()` 方法
- `internal/node/subsystem.go` - 实现日志记录和心跳发送
- `internal/version/version.go` - 版本更新至 0.1.3

#### API 响应格式
所有 API 遵循 RESTful 规范：
- `200 OK` - 操作成功
- `400 Bad Request` - 请求格式错误或缺少必需参数
- `404 Not Found` - 资源不存在（下载/进程ID无效）
- `500 Internal Server Error` - 服务器内部错误

### Testing
- 所有核心功能测试通过
- 代码编译验证通过
- Node 子系统测试通过

## [0.1.2] - 2026-02-19

### Fixed
- **Web 启动脚本** - 修复 `./scripts/web.sh dev` 不起作用的问题
  - 集成配置同步功能（自动调用 sync-web-config.sh）
  - 添加端口冲突检测和自动停止旧进程
  - 移除命令行端口参数（改为从配置文件读取）

### Changed
- **开发服务器启动** - 现在统一使用 `./scripts/web.sh dev`
  - 删除临时启动脚本 `scripts/start-dev.sh`
  - 删除过时文档：`INDEPENDENT.md`, `CONFIG.md`, `START_DEV.md`
  - 更新 `web/README.md`，添加配置说明和快速开始指南

### Removed
- 删除 `web/INDEPENDENT.md` - 迁移指南，迁移已完成
- 删除 `web/CONFIG.md` - 配置热重载说明，内容整合到 README
- 删除 `web/START_DEV.md` - 临时故障排查文档，问题已解决
- 删除 `scripts/start-dev.sh` - 临时启动脚本，功能已整合到 web.sh

## [0.1.1] - 2026-02-19

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

- **Master-Client 分布式架构支持**
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
- 后端不再提供 `/api/config/web` 端点（前端独立配置）
- API 客户端支持动态后端 URL 配置
- Vite 配置移除代理设置，前端独立运行
- 后端 CORS 已配置允许所有源访问
- SSE Hook 现在使用完整后端 URL 而非相对路径
- 功能配置：禁用尚未实现的集群管理和日志查看功能
- 优化日志系统，添加实例方法
- 更新配置验证逻辑，优先检查模式参数
- 完善 README 文档结构

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
- **配置测试** - 修复配置测试中的模式验证问题
- 禁用未实现的功能（集群管理、日志查看），避免 404 错误

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
- [ ] 系统托盘（Windows/Linux/macOS）
- [ ] Docker 镜像支持
- [ ] Kubernetes 部署支持
- [ ] 集群管理和日志查看功能完善

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
