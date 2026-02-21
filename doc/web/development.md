# Shepherd Web 开发文档

## 快速开始

### 开发模式

```bash
# 使用运行脚本（推荐）
./run.sh dev
# 或指定端口
./run.sh dev -p 4000

# 或直接使用 npm
npm run dev
```

### 生产构建

```bash
# 构建生产版本
./run.sh build

# 预览构建结果
./run.sh preview
```

## 项目结构

```
web/
├── src/
│   ├── components/          # 可复用组件
│   │   ├── ui/              # 基础 UI 组件 (shadcn/ui 风格)
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   └── badge.tsx
│   │   ├── layout/          # 布局组件
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Header.tsx
│   │   │   └── MainLayout.tsx
│   │   ├── models/          # 模型相关组件
│   │   ├── downloads/       # 下载相关组件
│   │   ├── chat/            # 聊天相关组件
│   │   └── cluster/         # 集群相关组件
│   │
│   ├── pages/              # 路由页面
│   │   ├── dashboard/       # 仪表盘
│   │   ├── models/          # 模型管理
│   │   ├── downloads/       # 下载管理
│   │   ├── chat/            # 聊天界面
│   │   ├── cluster/         # 集群管理
│   │   ├── logs/            # 日志查看
│   │   └── settings/        # 设置
│   │
│   ├── features/           # 功能模块 (业务逻辑)
│   │   ├── models/          # 模型功能
│   │   │   ├── hooks.ts     # React Query hooks
│   │   │   └── types.ts     # 类型定义
│   │   ├── downloads/
│   │   ├── chat/
│   │   └── cluster/
│   │
│   ├── stores/             # Zustand 状态管理
│   │   └── uiStore.ts       # UI 状态 (侧边栏、主题等)
│   │
│   ├── hooks/              # 自定义 Hooks
│   │   ├── useSSE.ts        # SSE 事件监听
│   │   └── ...
│   │
│   ├── lib/                # 工具库
│   │   ├── api/             # API 客户端
│   │   │   └── client.ts    # HTTP 客户端封装
│   │   └── utils.ts        # 工具函数
│   │
│   ├── types/              # TypeScript 类型定义
│   │   ├── model.ts        # 模型类型
│   │   ├── download.ts     # 下载类型
│   │   ├── cluster.ts      # 集群类型
│   │   ├── events.ts       # SSE 事件类型
│   │   └── index.ts        # 统一导出
│   │
│   └── i18n/               # 国际化
│       └── locales/
│           ├── en.json      # 英文
│           └── zh.json      # 中文
│
├── public/                 # 静态资源
├── dist/                   # 生产构建输出
├── run.sh                  # Linux/macOS 运行脚本
├── run.bat                 # Windows 运行脚本
└── package.json           # 项目配置
```

## 核心架构

### 1. API 层

**API 客户端** (`src/lib/api/client.ts`):
- 类型安全的 fetch 封装
- 统一错误处理
- 支持 GET/POST/PUT/DELETE

**React Query Hooks** (`src/features/*/hooks.ts`):
- 自动缓存和重新验证
- 乐观更新
- 并行请求处理

### 2. 实时通信

**SSE Hook** (`src/hooks/useSSE.ts`):
- 自动连接到 `/api/events`
- 指数退避重连
- 自动使相关查询失效

**事件处理**:
```typescript
useSSE({
  onMessage: (event) => {
    // 处理 SSE 事件
    switch (event.type) {
      case 'modelLoad':
        // 模型加载事件
        break;
    }
  }
});
```

### 3. 状态管理

**Zustand** (`src/stores/uiStore.ts`):
- UI 状态（侧边栏、主题、模态框）
- 持久化到 localStorage
- 简洁的 API

**React Query**:
- 服务器状态（模型、下载、集群）
- 自动缓存和同步
- 加载和错误状态

### 4. 组件设计

**基础组件** (`src/components/ui/`):
- 使用 Tailwind CSS
- 完全类型安全
- 支持主题定制

**布局组件** (`src/components/layout/`):
- Sidebar - 可折叠侧边栏
- Header - 顶部栏和搜索
- MainLayout - 主布局容器

## 开发规范

### 文件命名

- 组件文件: PascalCase (例: `ModelCard.tsx`)
- 工具文件: camelCase (例: `apiClient.ts`)
- Hooks 文件: camelCase with `use` 前缀 (例: `useSSE.ts`)
- 类型文件: camelCase (例: `model.ts`)

### 导入顺序

```typescript
// 1. React 核心
import { useState, useEffect } from 'react';

// 2. 第三方库
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';

// 3. 相对导入
import { useModels } from '@/features/models/hooks';
```

### TypeScript 规范

- 优先使用 `type` 导入类型
- 避免使用 `any`
- 接口优先于类型别名（用于对象）
- 类型别名用于联合类型和工具类型

### 组件规范

```typescript
// ✅ 推荐
interface ComponentProps {
  title: string;
  onSubmit: () => void;
}

export function Component({ title, onSubmit }: ComponentProps) {
  // ...
}

// ❌ 避免
export function Component({ title, onSubmit }: any) {
  // ...
}
```

## API 集成

### 端点配置

开发环境下，Vite 自动代理 API 请求：

```javascript
// vite.config.ts
server: {
  proxy: {
    '/api': 'http://localhost:9190',
    '/v1': 'http://localhost:9190'
  }
}
```

### 数据获取

```typescript
// 使用 React Query Hook
function MyComponent() {
  const { data, isLoading, error } = useModels();

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;
  
  return <div>{data?.map(...)}</div>;
}
```

### 数据更新

```typescript
// 使用 Mutation
function LoadButton({ modelId }: { modelId: string }) {
  const { mutate: loadModel, isPending } = useLoadModel();

  return (
    <Button onClick={() => loadModel({ modelId })} disabled={isPending}>
      {isPending ? '加载中...' : '加载模型'}
    </Button>
  );
}
```

## 样式系统

### Tailwind CSS v4

项目使用 Tailwind CSS v4 (基于 @import):

```css
/* src/index.css */
@import "tailwindcss";
```

### 主题定制

CSS 变量在 `src/index.css` 中定义：

```css
:root {
  --primary: 221.2 83.2% 53.3%;
  --radius: 0.5rem;
}
```

### 组件样式

```typescript
// 使用 cn() 工具函数合并类名
import { cn } from '@/lib/utils';

className={cn(
  'base-class',
  isActive && 'active-class',
  className
)}
```

## 测试

### 类型检查

```bash
npm run type-check
```

### Linting

```bash
npm run lint
npm run lint:fix
```

## 构建和部署

### 生产构建

```bash
./run.sh build
```

输出到 `dist/` 目录，可直接部署到静态文件服务器。

### 环境变量

创建 `.env.local` 文件（不要提交）：

```env
VITE_API_BASE_URL=http://localhost:9190
VITE_WS_URL=ws://localhost:9190
```

## 故障排除

### 端口已被占用

```bash
# 使用不同端口
./run.sh dev -p 4000
```

### 类型错误

```bash
# 清理缓存并重新构建
rm -rf node_modules/.vite dist
npm run build
```

### 依赖问题

```bash
# 重新安装依赖
rm -rf node_modules package-lock.json
npm install
```

## 下一步

- [ ] 实现模型管理页面
- [ ] 实现下载管理页面
- [ ] 实现聊天界面
- [ ] 实现集群管理页面
- [ ] 添加单元测试
- [ ] 添加 E2E 测试
