# Web 前端独立配置 - 更新说明

## 重要变更

前端配置已从**后端驱动**改为**前端独立配置**。

### 之前的架构（耦合）

```
前端 (3000) → 代理 → 后端 (9190)
    ↓
从后端获取配置
```

### 新的架构（解耦）

```
前端 (任意端口) → 直接连接 → 后端 (任意地址)
    ↓
从本地 config.yaml 读取配置
```

## 配置文件位置变更

| 旧配置文件 | 状态 | 说明 |
|-----------|------|------|
| `config/web.config.yaml` | ❌ 删除 | 后端控制的配置，已废弃 |
| `web/config.yaml` | ✅ 使用 | 前端独立配置，完全控制 |

## 新的配置结构

`web/config.yaml` 是前端的主配置文件：

```yaml
# 后端服务器配置（可配置多个）
backend:
  urls:
    - "http://localhost:9190"       # 主后端
    - "http://backup:9190"          # 备用后端
  currentIndex: 0                   # 当前使用的后端

# 前端服务器配置
server:
  port: 3000                       # 前端运行端口

# 功能开关
features:
  models: true
  downloads: true
  cluster: true

# UI 配置
ui:
  theme: "auto"
  language: "zh-CN"
  pageSize: 20
```

## 主要改进

1. **前端完全独立** - 可以部署到任何服务器，连接任意后端
2. **多后端支持** - 配置多个后端地址，运行时切换
3. **无代理依赖** - 开发模式不再需要 Vite 代理
4. **CORS 友好** - 后端已配置允许所有源访问

## 开发命令变更

### 启动前端

```bash
# 旧方式（仍然可用，但需要后端运行）
npm run dev

# 新方式（完全独立）
npm run dev
```

### 配置同步

修改 `web/config.yaml` 后，运行：

```bash
./scripts/sync-web-config.sh
```

或者在 `web/` 目录运行：

```bash
cp config.yaml public/config.yaml
```

### 验证配置

启动前端后，检查浏览器控制台：

```
Frontend config loaded: {
  backendUrl: "http://localhost:9190",
  features: { models: true, downloads: true, ... }
}
```

## API 调用

前端现在使用完整的后端 URL：

```typescript
// 旧方式（相对路径，依赖代理）
fetch('/api/models')

// 新方式（完整 URL，独立连接）
fetch('http://localhost:9190/api/models')
```

## 迁移指南

### 如果您之前使用 `config/web.config.yaml`

1. **备份旧配置**：
   ```bash
   cp config/web.config.yaml config/web.config.yaml.bak
   ```

2. **创建新配置**：
   - 创建 `web/config.yaml`（使用新的格式）
   - 复制后端 URL 配置
   - 调整其他配置项

3. **更新部署**：
   - 开发环境：使用 `npm run dev`
   - 生产环境：重新构建并部署

### 配置映射表

| 旧配置路径 | 新配置路径 | 说明 |
|-----------|-----------|------|
| `config/web.config.yaml.server` | `web/config.yaml.server` | 服务器配置 |
| `config/web.config.yaml.api.baseUrl` | `web/config.yaml.backend.urls[0]` | 后端 URL |
| `config/web.config.yaml.features` | `web/config.yaml.features` | 功能开关 |
| `config/web.config.yaml.ui` | `web/config.yaml.ui` | UI 配置 |

## 注意事项

1. **`public/config.yaml` 是自动生成的**
   - 源文件：`web/config.yaml`
   - 运行 `./scripts/sync-web-config.sh` 同步

2. **开发模式端口由 Vite 控制**
   - `web/config.yaml` 中的 `server.port` 仅为将来独立部署准备
   - 当前开发端口：`vite.config.ts` 中的 `port: 3000`

3. **后端 CORS 必须启用**
   - 后端已默认启用 CORS（允许所有源）
   - 如需限制，修改后端配置中的 `security.cors` 部分

## 获取帮助

如有问题，请查看：

- [DEPLOYMENT.md](./DEPLOYMENT.md) - 详细部署指南
- [README.md](../README.md) - 项目文档
- [config.yaml](./config.yaml) - 配置示例
