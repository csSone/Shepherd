# 脚本目录迁移指南

## 📋 概述

Shepherd 脚本已按操作系统重新组织到子目录中，提供更好的跨平台支持和文档管理。

## 📁 新目录结构

```
scripts/
├── linux/              # Linux 专用脚本
│   ├── build.sh
│   ├── run.sh
│   ├── web.sh
│   ├── sync-web-config.sh
│   ├── watch-sync-config.sh
│   └── README.md
├── macos/              # macOS 专用脚本
│   ├── build.sh
│   ├── run.sh
│   ├── web.sh
│   └── README.md
├── windows/            # Windows 专用脚本
│   ├── build.bat
│   ├── run.bat
│   ├── web.bat
│   └── README.md
├── build-all.sh        # 跨平台编译脚本
├── release.sh          # 发布打包脚本
└── README.md           # 脚本总览
```

## 🔄 迁移步骤

### 从旧脚本路径迁移

如果您在文档、脚本或自动化中使用了旧的脚本路径，请按以下方式更新：

#### 1. 编译脚本

| 旧路径 | 新路径 (Linux) | 新路径 (macOS) | 新路径 (Windows) |
|--------|---------------|---------------|-----------------|
| `./scripts/build.sh` | `./scripts/linux/build.sh` | `./scripts/macos/build.sh` | `scripts\windows\build.bat` |

#### 2. 运行脚本

| 旧路径 | 新路径 (Linux) | 新路径 (macOS) | 新路径 (Windows) |
|--------|---------------|---------------|-----------------|
| `./scripts/run.sh` | `./scripts/linux/run.sh` | `./scripts/macos/run.sh` | `scripts\windows\run.bat` |

#### 3. Web 脚本

| 旧路径 | 新路径 (Linux) | 新路径 (macOS) | 新路径 (Windows) |
|--------|---------------|---------------|-----------------|
| `./scripts/web.sh` | `./scripts/linux/web.sh` | `./scripts/macos/web.sh` | `scripts\windows\web.bat` |

### 示例迁移

#### Shell/Bash 脚本

```bash
# 旧代码
./scripts/build.sh
./scripts/run.sh standalone
./scripts/web.sh dev

# 新代码 (Linux)
./scripts/linux/build.sh
./scripts/linux/run.sh standalone
./scripts/linux/web.sh dev
```

#### Windows 批处理

```batch
REM 旧代码
scripts\build.bat
scripts\run.bat standalone
scripts\web.bat dev

REM 新代码
scripts\windows\build.bat
scripts\windows\run.bat standalone
scripts\windows\web.bat dev
```

#### PowerShell

```powershell
# 旧代码
& ".\scripts\build.sh"
& ".\scripts\run.sh" "standalone"

# 新代码
& ".\scripts\windows\build.bat"
& ".\scripts\windows\run.bat" "standalone"
```

#### Makefile

```makefile
# 旧代码
build:
	./scripts/build.sh

run:
	./scripts/run.sh standalone

# 新代码 (使用 OS 检测)
UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
    SCRIPT_DIR := scripts/linux
endif
ifeq ($(UNAME_S),Darwin)
    SCRIPT_DIR := scripts/macos
endif

build:
	./$(SCRIPT_DIR)/build.sh

run:
	./$(SCRIPT_DIR)/run.sh standalone
```

## 📝 文档更新

### Markdown 文档

如果您在 README、指南或其他文档中引用了脚本，请更新路径：

```markdown
<!-- 旧 -->
运行: `./scripts/run.sh standalone`

<!-- 新 (Linux) -->
运行: `./scripts/linux/run.sh standalone`

<!-- 新 (macOS) -->
运行: `./scripts/macos/run.sh standalone`

<!-- 新 (Windows) -->
运行: `scripts\windows\run.bat standalone`
```

### 代码注释

```go
// 旧注释
// 使用: ./scripts/build.sh 编译项目

// 新注释
// 使用: ./scripts/linux/build.sh (Linux) 编译项目
//       ./scripts/macos/build.sh (macOS)
//       scripts\windows\build.bat (Windows)
```

## 🔄 自动化迁移工具

如果您需要批量更新多个文件中的脚本路径，可以使用以下命令：

### Linux/macOS

```bash
# 更新所有 Markdown 文件
find . -name "*.md" -type f -exec sed -i 's|./scripts/build.sh|./scripts/linux/build.sh|g' {} +

# 更新所有 Shell 脚本
find . -name "*.sh" -type f -exec sed -i 's|./scripts/build.sh|./scripts/linux/build.sh|g' {} +
```

### Windows PowerShell

```powershell
# 更新所有 Markdown 文件
Get-ChildItem -Recurse -Filter "*.md" | ForEach-Object {
    (Get-Content $_.FullName) -replace '\./scripts/build\.sh','./scripts/linux/build.sh' | Set-Content $_.FullName
}
```

## ⚠️ 注意事项

### 1. 兼容性

- **旧脚本暂时保留**: 为确保平滑过渡，旧脚本仍在 `scripts/` 根目录
- **逐步淘汰**: 未来版本将移除旧脚本
- **立即迁移**: 建议尽快迁移到新路径

### 2. 跨平台脚本

如果您的脚本需要跨平台运行，建议使用条件判断：

```bash
#!/bin/bash
# 跨平台编译脚本

OS=$(uname -s)

case "${OS}" in
    Linux*)
        ./scripts/linux/build.sh "$@"
        ;;
    Darwin*)
        ./scripts/macos/build.sh "$@"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        ./scripts/windows/build.bat "$@"
        ;;
    *)
        echo "不支持的操作系统: ${OS}"
        exit 1
        ;;
esac
```

### 3. CI/CD 管道

更新 CI/CD 配置文件（如 `.github/workflows`, `.gitlab-ci.yml`, `Jenkinsfile`）：

```yaml
# .github/workflows/build.yml (示例)
- name: Build (Linux)
  run: ./scripts/linux/build.sh

- name: Build (macOS)
  run: ./scripts/macos/build.sh

- name: Build (Windows)
  run: scripts\windows\build.bat
```

## 📚 参考资源

- [Linux 脚本文档](./linux/README.md)
- [macOS 脚本文档](./macos/README.md)
- [Windows 脚本文档](./windows/README.md)
- [脚本总览](./README.md)

## 🆘 需要帮助?

如果您在迁移过程中遇到问题：

1. 查阅对应操作系统的 README.md
2. 检查脚本的帮助信息 (`--help` 参数)
3. 提交 Issue 到 GitHub 仓库

---

*最后更新: 2026-02-20*
