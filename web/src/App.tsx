import { QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { queryClient } from './lib/query-client';
import { MainLayout } from './components/layout/MainLayout';
import { DashboardPage } from './pages/dashboard';
import { useSSE } from './hooks/useSSE';

/**
 * 应用根组件
 */
function App() {
  // 启用 SSE 实时事件
  useSSE({
    onMessage: (event) => {
      console.log('SSE Event:', event);
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<MainLayout />}>
            <Route index element={<DashboardPage />} />
            {/* 其他路由将在后续实现 */}
            <Route path="models" element={<div>模型管理页面 - 开发中</div>} />
            <Route path="downloads" element={<div>下载管理页面 - 开发中</div>} />
            <Route path="chat" element={<div>聊天页面 - 开发中</div>} />
            <Route path="cluster" element={<div>集群管理页面 - 开发中</div>} />
            <Route path="logs" element={<div>日志页面 - 开发中</div>} />
            <Route path="settings" element={<div>设置页面 - 开发中</div>} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
