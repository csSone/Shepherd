# Shepherd 脚本总览

Shepherd 项目提供跨平台的构建和运行脚本，支持 Linux、macOS 和 Windows 三大操作系统。

## 📁 目录结构

```
scripts/
├── linux/              # Linux 脚本
│   ├── build.sh        # 编译脚本
│   ├── run.sh          # 运行脚本
│   ├── web.sh          # Web 前端脚本
│   ├── sync-web-config.sh
│   ├── watch-sync-config.sh
│   └── README.md       # Linux 详细文档
├── macos/              # macOS 脚本
│   ├── build.sh        # 编译脚本 (支持 Intel/Apple Silicon)
│   ├── run.sh          # 运行脚本
│   ├── web.sh          # Web 前端脚本
│   └── README.md       # macOS 详细文档
├── windows/            # Windows 脚本
│   ├── build.bat       # 编译脚本
│   ├── run.bat         # 运行脚本
│   ├── web.bat         # Web 前端脚本
│   └── README.md       # Windows 详细文档
├── build-all.sh        # 跨平台编译脚本
├── release.sh          # 发布打包脚本
└── README.md           # 本文档
```

## 🚀 快速开始

### Linux

```bash
# 编译
./scripts/linux/build.sh

# 运行
./scripts/linux/run.sh standalone

# Web 前端
./scripts/linux/web.sh dev
```

### macOS

```bash
# 编译 (自动检测架构)
./scripts/macos/build.sh

# 运行
./scripts/macos/run.sh standalone

# Web 前端
./scripts/macos/web.sh dev
```

### Windows

```batch
REM 编译
scripts\windows\build.bat

REM 运行
scripts\windows\run.bat standalone

REM Web 前端
scripts\windows\web.bat dev
```

## 📋 脚本功能对比

| 功能 | Linux | macOS | Windows |
|------|-------|-------|---------|
| 编译 | ✅ build.sh | ✅ build.sh | ✅ build.bat |
| 运行 | ✅ run.sh | ✅ run.sh | ✅ run.bat |
| Web 前端 | ✅ web.sh | ✅ web.sh | ✅ web.bat |
| 配置同步 | ✅ sync-web-config.sh | ❌ | ❌ |
| 配置监视 | ✅ watch-sync-config.sh | ❌ | ❌ |
| Universal Binary | ❌ | ✅ | ❌ |
| 代码签名 | ❌ | ✅ | ❌ |

## 🔧 构建脚本功能

### Linux (build.sh)

- **自动架构检测**: x86_64, ARM64, RISC-V
- **版本注入**: 通过 ldflags 注入版本信息
- **Go 代理**: 自动设置 GOPROXY
- **符号链接**: 为非 amd64 架构创建链接

### macOS (build.sh)

- **自动架构检测**: x86_64 (Intel), ARM64 (Apple Silicon)
- **Universal Binary**: 可构建同时支持 Intel 和 Apple Silicon 的版本
- **代码签名**: 支持代码签名证书
- **Gatekeeper**: 自动处理隔离属性问题

### Windows (build.bat)

- **自动架构检测**: x86_64, ARM64
- **版本注入**: 注入版本和构建时间
- **Go 代理**: 自动设置 GOPROXY

## 🏃 运行脚本功能

### 共同功能

所有平台的运行脚本都支持：

- **三种模式**: standalone, master, client
- **自动编译**: `-b/--build` 选项
- **版本显示**: `-v/--version` 选项
- **帮助信息**: `-h/--help` 选项
- **Client 配置**: `--master`, `--name`, `--tags` 选项

### 平台特定功能

**Linux**:
- 标准 systemd 服务支持

**macOS**:
- Gatekeeper 隔离问题自动修复 (`--no-gatekeeper`)
- Launch Agent 支持

**Windows**:
- 服务管理 (NSSM/sc)
- 防火墙规则配置

## 🌐 Web 前端脚本

所有平台的 Web 脚本支持：

- `dev` - 启动开发服务器
- `build` - 构建生产版本
- `preview` - 预览构建结果
- `install` - 安装依赖
- `clean` - 清理构建文件
- `fix` - 修复依赖问题
- `check` - 检查依赖状态

## 🔨 跨平台构建

### build-all.sh

构建所有平台的二进制文件：

```bash
./scripts/build-all.sh v0.1.3
```

输出目录：
```
build/
├── shepherd-linux-amd64
├── shepherd-linux-arm64
├── shepherd-darwin-amd64
├── shepherd-darwin-arm64
├── shepherd-windows-amd64.exe
└── shepherd-windows-arm64.exe
```

### release.sh

创建发布包：

```bash
./scripts/release.sh v0.1.3
```

输出：
```
release/
├── shepherd-v0.1.3-linux-amd64.tar.gz
├── shepherd-v0.1.3-linux-arm64.tar.gz
├── shepherd-v0.1.3-darwin-amd64.tar.gz
├── shepherd-v0.1.3-darwin-arm64.tar.gz
├── shepherd-v0.1.3-windows-amd64.zip
└── CHECKSUMS.txt
```

## 📝 环境变量

### 通用环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `GOPROXY` | Go 模块代理 | https://goproxy.cn,direct |
| `RUN_TESTS` | 编译后运行测试 | (未设置) |
| `SHEPHERD_CLIENT_NAME` | Client 节点名称 | (未设置) |
| `SHEPHERD_CLIENT_TAGS` | Client 节点标签 | (未设置) |

### macOS 特定

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `BUILD_UNIVERSAL` | 构建 Universal Binary | (未设置) |
| `CODESIGN_IDENTITY` | 代码签名证书 | (未设置) |

## 🛠️ 依赖要求

### Linux

- **Go**: 1.21+ (通过包管理器或官方安装包)
- **Git**: 任意版本
- **Bash**: 4.0+

### macOS

- **Go**: 1.21+ (Homebrew 或官方安装包)
- **Git**: 任意版本 (Xcode Command Line Tools)
- **Bash**: 3.2+ (系统自带)

### Windows

- **Go**: 1.21+ (Chocolatey 或官方安装包)
- **Git**: 任意版本 (Git for Windows)
- **PowerShell**: 5.1+ (系统自带)
- **CMD**: 系统自带

## 📚 详细文档

- [Linux 脚本详细文档](./scripts/linux.md)
- [macOS 脚本详细文档](./scripts/macos.md)
- [Windows 脚本详细文档](./scripts/windows.md)
- [迁移指南](./migration.md)

## 🔍 故障排查

### 通用问题

**编译失败**:
```bash
# 检查 Go 版本
go version

# 清理缓存
go clean -modcache

# 更新依赖
go mod tidy
```

**权限问题**:
```bash
# Linux/macOS
chmod +x ./scripts/*/*.sh

# Windows: 以管理员身份运行
```

### 平台特定问题

- **Linux**: [详见故障排查](./scripts/linux.md#故障排查)
- **macOS**: [详见故障排查](./scripts/macos.md#故障排查)
- **Windows**: [详见故障排查](./scripts/windows.md#故障排查)

## 🤝 贡献指南

添加新脚本时，请确保：

1. **跨平台一致性**: 所有平台的脚本应提供相似的功能
2. **错误处理**: 提供清晰的错误信息
3. **文档更新**: 更新对应平台的 README.md
4. **可执行权限**: Linux/macOS 脚本需要可执行权限
5. **测试**: 在目标平台上测试脚本

## 📄 许可证

本项目的脚本遵循相同的开源许可证。

---

*Shepherd - 分布式 AI 模型管理系统*
