# Shepherd 脚本说明

本目录包含 Shepherd 项目的编译和运行脚本。

## 📁 脚本列表

### 编译脚本

### 单平台编译

#### Linux/macOS

```bash
./scripts/build.sh [version]
```

参数：
- `version`: 版本号（可选，默认为 `dev`）

示例：
```bash
./scripts/build.sh              # 使用默认版本号
./scripts/build.sh 1.0.0        # 指定版本号
```

输出：`build/shepherd` (或 `build/shepherd-linux-amd64`)

#### Windows

```batch
scripts\build.bat [version]
```

### 跨平台编译

一次性编译所有支持的平台：

```bash
./scripts/build-all.sh [version]
```

支持的平台：
- linux/amd64
- linux/arm64
- darwin/amd64 (macOS Intel)
- darwin/arm64 (macOS Apple Silicon)
- windows/amd64
- windows/386

输出文件位于 `build/` 目录：
```
build/
├── shepherd-linux-amd64
├── shepherd-linux-arm64
├── shepherd-darwin-amd64
├── shepherd-darwin-arm64
├── shepherd-windows-amd64.exe
├── shepherd-windows-386.exe
└── SHA256SUMS
```

### 发布打包

将编译好的二进制文件打包成发布包：

```bash
./scripts/release.sh [version]
```

生成的发布包位于 `release/` 目录：
```
release/
├── shepherd-1.0.0-linux-amd64.tar.gz
├── shepherd-1.0.0-darwin-amd64.tar.gz
├── shepherd-1.0.0-windows-amd64.zip
└── SHA256SUMS
```

每个发布包包含：
- 可执行文件
- 启动脚本 (`start.sh` 或 `start.bat`)
- 配置文件示例 (`config/config.yaml`)
- README 文档

## 环境要求

### 必需
- Go 1.25+
- Git

### 可选
- upx (用于进一步压缩二进制文件)
- docker (用于容器化部署)

## 编译参数

编译脚本会注入以下信息到二进制文件中：

- `Version`: 版本号
- `BuildTime`: 构建时间 (UTC)
- `GitCommit`: Git 提交哈希

这些信息可以通过 `--version` 参数查看：

```bash
./shepherd --version
```

## 高级用法

### 使用 UPX 压缩

进一步减小二进制文件大小：

```bash
upx --best --lzma build/shepherd
```

### 交叉编译

如果需要交叉编译到其他平台，可以使用 Go 的交叉编译功能：

```bash
# 编译 Windows 版本 (在 Linux 上)
GOOS=windows GOARCH=amd64 go build -o build/shepherd-windows.exe cmd/shepherd/main.go

# 编译 macOS 版本 (在 Linux 上)
GOOS=darwin GOARCH=amd64 go build -o build/shepherd-macos cmd/shepherd/main.go

# 编译 ARM 版本
GOARCH=arm64 GOARM=7 go build -o build/shepherd-arm64 cmd/shepherd/main.go
```

### 调试编译

编译带调试信息的版本：

```bash
go build -gcflags="all=-N -l" -o build/shepherd-debug cmd/shepherd/main.go
```

### 性能优化编译

启用更多优化：

```bash
go build -ldflags="-s -w" -gcflags="-l=4" cmd/shepherd/main.go
```

## Docker 构建

如果需要构建 Docker 镜像，请参考项目根目录的 `Dockerfile`。

```bash
docker build -t shepherd:latest .
```

## 故障排除

### 编译错误：找不到 go 命令

确保 Go 已安装并在 PATH 中：

```bash
which go
go version
```

### 编译错误：权限不足

确保脚本有执行权限：

```bash
chmod +x scripts/*.sh
```

### 测试失败

跳过测试：

```bash
RUN_TESTS=false ./scripts/build.sh
```

### Windows 下的编码问题

确保终端使用 UTF-8 编码：

```batch
chcp 65001
```

## 🚀 运行脚本

### Linux/macOS (run.sh)

```bash
# 查看帮助
./scripts/run.sh --help

# 单机模式
./scripts/run.sh standalone

# Master 模式
./scripts/run.sh master --port 9190 --scan

# Client 模式
./scripts/run.sh client --master http://192.168.1.100:9190 --name client-1

# 运行前先编译
./scripts/run.sh master -b
```

### Windows (run.bat)

```batch
REM 查看帮助
scripts\run.bat --help

REM 单机模式
scripts\run.bat standalone

REM Master 模式
scripts\run.bat master --port 9190 --scan

REM Client 模式
scripts\run.bat client --master http://192.168.1.100:9190 --name client-1
```

### 运行脚本选项

| 选项 | 说明 |
|------|------|
| `-h, --help` | 显示帮助信息 |
| `-b, --build` | 运行前先编译 |
| `-v, --version` | 显示版本信息 |
| `--mode MODE` | 运行模式：standalone, master, client |
| `--master URL` | Master 地址（Client 模式必需） |
| `--name NAME` | Client 名称 |
| `--tags TAGS` | Client 标签（逗号分隔） |
| `--port PORT` | Web 服务器端口（Master 模式） |
| `--scan` | 启动时自动扫描网络（Master 模式） |

## 📖 使用场景

### 开发环境

每次运行前自动编译：

```bash
./scripts/run.sh standalone -b
```

### 生产部署

**Master 节点：**

```bash
./scripts/run.sh master --scan
```

**Client 节点：**

```bash
./scripts/run.sh client \
  --master http://master.example.com:9190 \
  --name gpu-server-1 \
  --tags "gpu,rocm,high-memory"
```

### 本地测试

```bash
# 终端 1: 启动 Master
./scripts/run.sh master --port 9190

# 终端 2: 启动 Client
./scripts/run.sh client --master http://localhost:9190 --name test-client
```

## 贡献

如需添加新的编译目标或改进脚本，请提交 Pull Request。
