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

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build
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
