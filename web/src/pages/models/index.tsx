import { useState } from 'react';
import { Search, RefreshCw, Filter, Grid3X3, List, Pencil, MessageSquare } from 'lucide-react';
import { useModels, useLoadModel, useUnloadModel, useSetModelFavourite, useUpdateModelAlias, useScanModels, useFilteredModels } from '@/features/models/hooks';
import { ModelCard } from '@/components/models/ModelCard';
import { LoadModelDialog } from '@/components/models/LoadModelDialog';
import { EditAliasDialog } from '@/components/models/EditAliasDialog';
import { TestModelDialog } from '@/components/models/TestModelDialog';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { Model, ModelStatus } from '@/types';
import { useAlertDialog } from '@/hooks/useAlertDialog';

/**
 * 模型管理页面
 */
export function ModelsPage() {
  const alertDialog = useAlertDialog();
  const { data: models = [], isLoading } = useModels();
  const loadModel = useLoadModel();
  const unloadModel = useUnloadModel();
  const setFavourite = useSetModelFavourite();
  const updateAlias = useUpdateModelAlias();
  const scanModels = useScanModels();

  // UI 状态
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<ModelStatus | ''>('');
  const [favouriteFilter, setFavouriteFilter] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  // 加载对话框状态
  const [dialogModel, setDialogModel] = useState<Model | null>(null);

  // 编辑别名对话框状态
  const [editAliasModel, setEditAliasModel] = useState<Model | null>(null);

  // 测试模型对话框状态
  const [testModel, setTestModel] = useState<Model | null>(null);

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
  const handleUnloadClick = async (modelId: string) => {
    const confirmed = await alertDialog.confirm({
      title: '卸载模型',
      description: '确定要卸载此模型吗？',
    });
    if (confirmed) {
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

  // 处理编辑别名
  const handleEditAlias = (model: Model) => {
    setEditAliasModel(model);
  };

  const handleAliasConfirm = (alias: string) => {
    if (editAliasModel) {
      updateAlias.mutate(
        { modelId: editAliasModel.id, alias },
        {
          onSuccess: () => {
            setEditAliasModel(null);
          },
        }
      );
    }
  };

  // 处理测试模型
  const handleTestModel = (model: Model) => {
    setTestModel(model);
  };

  return (
    <div className="space-y-6">
      {/* 标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">模型管理</h1>
          <p className="text-muted-foreground">
            共 {models?.length || 0} 个模型，已加载 {models?.filter((m) => m.isLoaded).length || 0} 个
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            onClick={handleScan}
            disabled={scanModels.isPending}
            variant="default"
            size="sm"
          >
            <RefreshCw className={cn('w-4 h-4', scanModels.isPending && 'animate-spin')} />
            扫描模型
          </Button>
        </div>
      </div>

      {/* 搜索和过滤 */}
      <div className="flex flex-wrap items-center gap-3 p-4 bg-card rounded-lg border border-border">
        {/* 搜索框 */}
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索模型名称、架构..."
            className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-input text-foreground placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* 状态过滤 */}
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as ModelStatus | '')}
          className="px-3 py-2 border border-border rounded-lg bg-input text-foreground"
        >
          <option value="">所有状态</option>
          <option value="loaded">已加载</option>
          <option value="running">运行中</option>
          <option value="stopped">已停止</option>
          <option value="loading">加载中</option>
          <option value="error">错误</option>
        </select>

        {/* 收藏过滤 */}
        <Button
          onClick={() => setFavouriteFilter(!favouriteFilter)}
          variant={favouriteFilter ? 'default' : 'outline'}
          size="sm"
          className={cn(
            favouriteFilter && 'border-yellow-500 bg-yellow-500 text-white hover:bg-yellow-600 dark:bg-yellow-600 dark:hover:bg-yellow-700 dark:border-yellow-700'
          )}
        >
          <Filter className="w-4 h-4" />
          只看收藏
        </Button>

        {/* 视图切换 */}
        <div className="flex items-center rounded-lg overflow-hidden border border-border/50">
          <Button
            onClick={() => setViewMode('grid')}
            variant={viewMode === 'grid' ? 'default' : 'ghost'}
            size="sm"
            className="rounded-none border-0"
          >
            <Grid3X3 className="w-4 h-4" />
          </Button>
          <Button
            onClick={() => setViewMode('list')}
            variant={viewMode === 'list' ? 'default' : 'ghost'}
            size="sm"
            className="rounded-none border-0"
          >
            <List className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* 模型列表 */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-center">
            <RefreshCw className="w-8 h-8 animate-spin text-blue-600 mx-auto mb-2" />
            <p className="text-muted-foreground">加载中...</p>
          </div>
        </div>
      ) : filteredModels.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
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
              actions={
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleEditAlias(model)}
                    title="编辑别名"
                  >
                    <Pencil className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleTestModel(model)}
                    title="测试模型"
                  >
                    <MessageSquare className="w-4 h-4" />
                  </Button>
                </div>
              }
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

      {/* Edit Alias Dialog */}
      {editAliasModel && (
        <EditAliasDialog
          isOpen={!!editAliasModel}
          onClose={() => setEditAliasModel(null)}
          onConfirm={handleAliasConfirm}
          modelId={editAliasModel.id}
          modelName={editAliasModel.name}
          currentAlias={editAliasModel.alias}
          isLoading={updateAlias.isPending}
        />
      )}

      {/* Test Model Dialog */}
      {testModel && (
        <TestModelDialog
          isOpen={!!testModel}
          onClose={() => setTestModel(null)}
          modelId={testModel.id}
          modelName={testModel.alias || testModel.name}
          isModelLoaded={testModel.isLoaded}
          onLoadModel={() => {
            setTestModel(null);
            handleLoadClick(testModel);
          }}
        />
      )}
    </div>
  );
}
