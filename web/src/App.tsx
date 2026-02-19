import { QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { queryClient } from './lib/query-client';
import { MainLayout } from './components/layout/MainLayout';
import { DashboardPage } from './pages/dashboard';
import { ModelsPage } from './pages/models';
import { DownloadsPage } from './pages/downloads';
import { ChatPage } from './pages/chat';
import { ClusterPage } from './pages/cluster';
import { LogsPage } from './pages/logs';
import { useSSE } from './hooks/useSSE';

// 导入代码高亮样式
import 'highlight.js/styles/github-dark.css';

/**
 * 应用内容组件
 * 在 QueryClientProvider 内部使用 SSE
 */
function AppContent() {
  // 启用 SSE 实时事件
  useSSE({
    onMessage: (event) => {
      console.log('SSE Event:', event);
    },
  });

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<MainLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="models" element={<ModelsPage />} />
          <Route path="downloads" element={<DownloadsPage />} />
          <Route path="chat" element={<ChatPage />} />
          <Route path="cluster" element={<ClusterPage />} />
          <Route path="logs" element={<LogsPage />} />
          <Route path="settings" element={<div>设置页面 - 开发中</div>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

/**
 * 应用根组件
 */
function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppContent />
    </QueryClientProvider>
  );
}

export default App;
