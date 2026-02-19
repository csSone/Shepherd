# Shepherd Web 前端

现代化的 React + TypeScript 前端应用，用于 Shepherd llama.cpp 模型管理系统。

## 技术栈

- **构建工具**: Vite
- **框架**: React 18
- **语言**: TypeScript 5
- **路由**: React Router v6
- **状态管理**: Zustand + React Query
- **样式**: Tailwind CSS v4
- **图标**: Lucide React

## 快速开始

```bash
# 使用项目脚本启动（推荐）
./scripts/web.sh dev

# 或进入 web 目录
cd web
npm install
npm run dev
```

## 配置

前端配置文件位于 `config/web.config.yaml`（项目根目录）。

### 修改配置后同步

```bash
# 方式 1: 使用启动脚本（自动同步）
./scripts/web.sh dev

# 方式 2: 手动同步
./scripts/sync-web-config.sh

# 方式 3: 自动热重载（开发时推荐）
./scripts/watch-sync-config.sh
```

### 端口配置

修改 `config/web.config.yaml` 中的 `server.port`，然后重启开发服务器。

### 后端地址配置

在 `config/web.config.yaml` 中配置后端地址，支持多后端切换：

```yaml
backend:
  urls:
    - "http://10.0.0.193:9190"  # 局域网地址
    - "http://localhost:9190"   # 本地地址
  currentIndex: 0                # 当前使用的后端索引
```

## 开发命令

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview

# 类型检查
npm run type-check

# 代码检查
npm run lint
```

## 项目结构

```
src/
├── components/          # 可复用组件
│   ├── ui/              # 基础 UI 组件
│   ├── layout/          # 布局组件
├── pages/              # 路由页面
├── features/           # 功能模块
├── stores/             # Zustand 状态
├── hooks/              # 自定义 Hooks
├── lib/                # 工具库
├── types/              # TypeScript 类型
```

## 文档

- [DEPLOYMENT.md](DEPLOYMENT.md) - 部署指南
- [DEVELOPMENT.md](DEVELOPMENT.md) - 开发文档
