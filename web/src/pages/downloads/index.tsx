import { useState } from 'react';
import { Plus, Trash2, Filter, CloudDownload } from 'lucide-react';
import {
  useDownloads,
  useCreateDownload,
  usePauseDownload,
  useResumeDownload,
  useCancelDownload,
  useRetryDownload,
  useClearCompletedDownloads,
  useFilteredDownloads,
  useDownloadStats,
} from '@/features/downloads/hooks';
import { DownloadCard } from '@/components/downloads/DownloadCard';
import { CreateDownloadDialog } from '@/components/downloads/CreateDownloadDialog';
import type { DownloadState, DownloadSource } from '@/types';

/**
 * 下载管理页面
 */
export function DownloadsPage() {
  const { data: downloads, isLoading } = useDownloads();
  const createDownload = useCreateDownload();
  const pauseDownload = usePauseDownload();
  const resumeDownload = useResumeDownload();
  const cancelDownload = useCancelDownload();
  const retryDownload = useRetryDownload();
  const clearCompleted = useClearCompletedDownloads();

  // UI 状态
  const [search, setSearch] = useState('');
  const [stateFilter, setStateFilter] = useState<DownloadState | ''>('');
  const [sourceFilter, setSourceFilter] = useState<DownloadSource | ''>('');
  const [dialogOpen, setDialogOpen] = useState(false);

  // 统计
  const stats = useDownloadStats(downloads);

  // 过滤下载任务
  const filteredDownloads = useFilteredDownloads(downloads, {
    search,
    state: stateFilter || undefined,
    source: sourceFilter || undefined,
  });

  // 处理创建下载
  const handleCreateDownload = (params: { source: DownloadSource; repoId: string }) => {
    createDownload.mutate(params as any, {
      onSuccess: () => {
        setDialogOpen(false);
      },
    });
  };

  // 处理清理已完成
  const handleClearCompleted = () => {
    if (confirm('确定要清理所有已完成的下载任务吗？')) {
      clearCompleted.mutate();
    }
  };

  return (
    <div className="space-y-6">
      {/* 标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">下载管理</h1>
          <p className="text-gray-600 dark:text-gray-400">
            共 {stats.total} 个任务，{stats.active} 个进行中
          </p>
        </div>

        <div className="flex items-center gap-2">
          {stats.completed > 0 && (
            <button
              onClick={handleClearCompleted}
              className="flex items-center gap-2 px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            >
              <Trash2 className="w-4 h-4" />
              清理已完成
            </button>
          )}
          <button
            onClick={() => setDialogOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 transition-colors"
          >
            <Plus className="w-4 h-4" />
            新建下载
          </button>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="text-2xl font-bold text-gray-900 dark:text-white">{stats.total}</div>
          <div className="text-sm text-gray-600 dark:text-gray-400">总任务</div>
        </div>
        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{stats.active}</div>
          <div className="text-sm text-gray-600 dark:text-gray-400">进行中</div>
        </div>
        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="text-2xl font-bold text-green-600 dark:text-green-400">{stats.completed}</div>
          <div className="text-sm text-gray-600 dark:text-gray-400">已完成</div>
        </div>
        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="text-2xl font-bold text-gray-900 dark:text-white">
            {((stats.downloadedBytes / (stats.totalBytes || 1)) * 100).toFixed(1)}%
          </div>
          <div className="text-sm text-gray-600 dark:text-gray-400">总进度</div>
        </div>
      </div>

      {/* 搜索和过滤 */}
      <div className="flex flex-wrap items-center gap-3 p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        {/* 搜索框 */}
        <div className="relative flex-1 min-w-[200px]">
          <CloudDownload className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索仓库 ID 或文件名..."
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* 状态过滤 */}
        <select
          value={stateFilter}
          onChange={(e) => setStateFilter(e.target.value as DownloadState | '')}
          className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
        >
          <option value="">所有状态</option>
          <option value="downloading">下载中</option>
          <option value="paused">已暂停</option>
          <option value="completed">已完成</option>
          <option value="failed">失败</option>
        </select>

        {/* 来源过滤 */}
        <select
          value={sourceFilter}
          onChange={(e) => setSourceFilter(e.target.value as DownloadSource | '')}
          className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
        >
          <option value="">所有来源</option>
          <option value="huggingface">HuggingFace</option>
          <option value="modelscope">ModelScope</option>
        </select>
      </div>

      {/* 下载列表 */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-2" />
            <p className="text-gray-600 dark:text-gray-400">加载中...</p>
          </div>
        </div>
      ) : filteredDownloads.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-400">
          <CloudDownload className="w-12 h-12 mb-4" />
          <p className="text-lg mb-2">暂无下载任务</p>
          <p className="text-sm mb-4">创建新任务开始下载模型</p>
          <button
            onClick={() => setDialogOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 transition-colors"
          >
            <Plus className="w-4 h-4" />
            新建下载
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {filteredDownloads.map((task) => (
            <DownloadCard
              key={task.id}
              task={task}
              onPause={() => pauseDownload.mutate(task.id)}
              onResume={() => resumeDownload.mutate(task.id)}
              onCancel={() => cancelDownload.mutate(task.id)}
              onRetry={() => retryDownload.mutate(task.id)}
            />
          ))}
        </div>
      )}

      {/* 创建下载对话框 */}
      <CreateDownloadDialog
        isOpen={dialogOpen}
        onClose={() => setDialogOpen(false)}
        onConfirm={handleCreateDownload}
        isLoading={createDownload.isPending}
      />
    </div>
  );
}
