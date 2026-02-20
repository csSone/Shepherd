import { useState } from 'react';
import { Search, RefreshCw, Filter, Grid3X3, List } from 'lucide-react';
import { useModels, useLoadModel, useUnloadModel, useSetModelFavourite, useScanModels, useFilteredModels } from '@/features/models/hooks';
import { ModelCard } from '@/components/models/ModelCard';
import { LoadModelDialog } from '@/components/models/LoadModelDialog';
import { cn } from '@/lib/utils';
import type { Model, ModelStatus } from '@/types';

/**
 * 模型管理页面
 */
export function ModelsPage() {
  const { data: models, isLoading } = useModels();
  const loadModel = useLoadModel();
  const unloadModel = useUnloadModel();
  const setFavourite = useSetModelFavourite();
  const scanModels = useScanModels();

  // UI 状态
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<ModelStatus | ''>('');
  const [favouriteFilter, setFavouriteFilter] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  // 加载对话框状态
  const [dialogModel, setDialogModel] = useState<Model | null>(null);

  // 过滤模型
  const filteredModels = useFilteredModels(models, {
    search,
    status: statusFilter || undefined,
    favourite: favouriteFilter || undefined,
  });

  // 处理加载模型
  const handleLoadClick = (model: Model) => {
    setDialogModel(model);
  };

  const handleLoadConfirm = (params: { modelId: string }) => {
    loadModel.mutate(params as any, {
      onSuccess: () => {
        setDialogModel(null);
      },
    });
  };

  // 处理卸载模型
  const handleUnloadClick = (modelId: string) => {
    if (confirm('确定要卸载此模型吗？')) {
      unloadModel.mutate(modelId);
    }
  };

  // 处理收藏切换
  const handleToggleFavourite = (modelId: string, favourite: boolean) => {
    setFavourite.mutate({ modelId, favourite: !favourite });
  };

  // 处理扫描
  const handleScan = () => {
    scanModels.mutate();
  };

  return (
    <div className="space-y-6">
      {/* 标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">模型管理</h1>
          <p className="text-gray-600 dark:text-gray-400">
            共 {models?.length || 0} 个模型，已加载 {models?.filter((m) => m.isLoaded).length || 0} 个
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleScan}
            disabled={scanModels.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={cn('w-4 h-4', scanModels.isPending && 'animate-spin')} />
            扫描模型
          </button>
        </div>
      </div>

      {/* 搜索和过滤 */}
      <div className="flex flex-wrap items-center gap-3 p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        {/* 搜索框 */}
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索模型名称、架构..."
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* 状态过滤 */}
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as ModelStatus | '')}
          className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
        >
          <option value="">所有状态</option>
          <option value="loaded">已加载</option>
          <option value="running">运行中</option>
          <option value="stopped">已停止</option>
          <option value="loading">加载中</option>
          <option value="error">错误</option>
        </select>

        {/* 收藏过滤 */}
        <button
          onClick={() => setFavouriteFilter(!favouriteFilter)}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-lg border transition-colors',
            favouriteFilter
              ? 'border-yellow-500 bg-yellow-50 text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400'
              : 'border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
          )}
        >
          <Filter className="w-4 h-4" />
          只看收藏
        </button>

        {/* 视图切换 */}
        <div className="flex items-center border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden">
          <button
            onClick={() => setViewMode('grid')}
            className={cn(
              'p-2 transition-colors',
              viewMode === 'grid'
                ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white'
                : 'bg-white dark:bg-gray-800 text-gray-500 hover:text-gray-700'
            )}
          >
            <Grid3X3 className="w-4 h-4" />
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={cn(
              'p-2 transition-colors',
              viewMode === 'list'
                ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white'
                : 'bg-white dark:bg-gray-800 text-gray-500 hover:text-gray-700'
            )}
          >
            <List className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* 模型列表 */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <RefreshCw className="w-8 h-8 animate-spin text-blue-600 mx-auto mb-2" />
            <p className="text-gray-600 dark:text-gray-400">加载中...</p>
          </div>
        </div>
      ) : filteredModels.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-400">
          <p className="text-lg mb-2">没有找到模型</p>
          <p className="text-sm">尝试调整搜索条件或扫描模型目录</p>
        </div>
      ) : (
        <div
          className={cn(
            'gap-4',
            viewMode === 'grid'
              ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'
              : 'flex flex-col'
          )}
        >
          {filteredModels.map((model) => (
            <ModelCard
              key={model.id}
              model={model}
              onLoad={() => handleLoadClick(model)}
              onUnload={() => handleUnloadClick(model.id)}
              onToggleFavourite={() => handleToggleFavourite(model.id, model.favourite)}
            />
          ))}
        </div>
      )}

      {/* 加载对话框 */}
      {dialogModel && (
        <LoadModelDialog
          isOpen={!!dialogModel}
          onClose={() => setDialogModel(null)}
          onConfirm={handleLoadConfirm}
          modelId={dialogModel.id}
          modelName={dialogModel.alias || dialogModel.name}
          isLoading={loadModel.isPending}
        />
      )}
    </div>
  );
}
