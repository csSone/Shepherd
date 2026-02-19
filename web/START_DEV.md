# 启动前端开发服务器

## 问题：配置了 3030 端口，但无法访问

## 原因

1. **旧的 Vite 进程仍在运行** - 使用旧的配置（端口 3000）
2. **需要重新启动** - vite.config.ts 已更新，但需要重启服务器

## 解决方案

### 方式 1：使用启动脚本（推荐）

```bash
cd /home/user/workspace/Shepherd
./scripts/start-dev.sh
```

这个脚本会：
- ✅ 自动同步最新配置
- ✅ 检查并停止占用端口的进程
- ✅ 使用正确的端口（从 config.yaml 读取）
- ✅ 显示访问地址

### 方式 2：手动启动

```bash
# 1. 停止旧的 Vite 进程
pkill -f "vite.*3000"

# 2. 确认配置已同步
./scripts/sync-web-config.sh

# 3. 进入 web 目录
cd web

# 4. 启动开发服务器（会自动使用配置的端口）
npm run dev
```

### 方式 3：使用 npm script

```bash
cd /home/user/workspace/Shepherd/web
npm run dev:start
```

## 验证

启动成功后，你应该看到：

```
VITE v7.3.1  ready in 69 ms
➜  Local:   http://localhost:3030/
➜  Network: http://10.0.0.193:3030/
```

然后访问：
- **本地**: http://localhost:3030
- **局域网**: http://10.0.0.193:3030

## 工作原理

```
config/web.config.yaml (server.port: 3030)
         ↓
  vite.config.ts 读取 public/config.yaml
         ↓
  Vite 使用 3030 端口启动
         ↓
  浏览器访问 http://10.0.0.193:3030
```

## 故障排查

### 端口仍被占用

```bash
# 查找占用端口的进程
lsof -i :3030

# 停止进程
kill -9 <PID>
```

### 配置没有生效

1. 确认配置已同步：
   ```bash
   cat web/public/config.yaml | grep port
   ```

2. 重新启动服务器（配置在启动时读取）

### 3000 端口仍然在使用

旧的进程可能没有完全停止。强制停止：
```bash
pkill -9 -f vite
```
