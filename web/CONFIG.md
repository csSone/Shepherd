# 前端配置热重载

## 问题

修改 `config/web.config.yaml` 后，前端不会自动重新加载配置。

## 原因

前端从 `web/public/config.yaml` 读取配置，而不是直接从 `config/web.config.yaml`。

## 解决方案

### 方式 1：手动同步（推荐用于单次修改）

```bash
# 在项目根目录执行
./scripts/sync-web-config.sh
```

然后刷新浏览器（Ctrl+Shift+R 强制刷新）。

### 方式 2：自动热重载（推荐用于开发）

**在终端 1：启动配置监视器**
```bash
# 在项目根目录执行
./scripts/watch-sync-config.sh
```

或者在 web 目录：
```bash
cd web
npm run dev:watch
```

**在终端 2：启动前端开发服务器**
```bash
cd web
npm run dev
```

现在当你修改 `config/web.config.yaml` 时，配置会自动同步到 `web/public/config.yaml`。

### 方式 3：使用符号链接（不推荐）

```bash
# 删除 web/public/config.yaml
rm web/public/config.yaml

# 创建符号链接
ln -s ../../config/web.config.yaml web/public/config.yaml
```

**注意**：这种方式不推荐，因为：
1. Windows 对符号链接支持较差
2. 可能在某些构建环境中出问题
3. .gitignore 可能无法正确忽略

## 工作原理

```
config/web.config.yaml (源文件)
         ↓
  手动同步 / 自动监视
         ↓
web/public/config.yaml (前端读取)
         ↓
     浏览器加载
```

## 配置文件说明

- **`config/web.config.yaml`** - 主配置文件（源文件，纳入版本控制）
  - 修改这个文件来改变前端配置
  - Git 跟踪这个文件

- **`web/public/config.yaml`** - 自动生成的副本（不纳入版本控制）
  - 由同步脚本自动生成
  - .gitignore 忽略这个文件
  - 不要手动编辑这个文件

## 依赖

自动热重载需要安装 `inotify-tools`：

```bash
# Ubuntu/Debian
sudo apt-get install inotify-tools

# CentOS/RHEL
sudo yum install inotify-tools

# macOS
brew install fswatch
```

## 故障排查

### 配置修改后没有生效

1. 检查是否运行了同步脚本
   ```bash
   ./scripts/sync-web-config.sh
   ```

2. 检查 `web/public/config.yaml` 是否已更新
   ```bash
   diff config/web.config.yaml web/public/config.yaml
   ```

3. 强制刷新浏览器（Ctrl+Shift+R）

### 监视脚本不工作

1. 检查是否安装了 inotify-tools
   ```bash
   which inotifywait
   ```

2. 查看脚本输出是否有错误信息

3. 确保文件路径正确
