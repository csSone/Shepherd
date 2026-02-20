import { type ReactNode } from 'react';
import { Brain, Cpu, HardDrive, Star, Loader2, Play, Square, ToggleLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { Model, ModelStatus } from '@/types';

interface ModelCardProps {
  model: Model;
  onLoad?: () => void;
  onUnload?: () => void;
  onToggleFavourite?: () => void;
  actions?: ReactNode;
}

/**
 * 模型状态颜色映射
 */
const STATUS_COLORS: Record<ModelStatus, string> = {
  stopped: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  loading: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  loaded: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  running: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300',
  unloading: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  error: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
};

/**
 * 模型状态标签
 */
const STATUS_LABELS: Record<ModelStatus, string> = {
  stopped: '已停止',
  loading: '加载中',
  loaded: '已加载',
  running: '运行中',
  unloading: '卸载中',
  error: '错误',
};

/**
 * 格式化文件大小
 */
function formatSize(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(2)} ${units[unitIndex]}`;
}

/**
 * 获取模型图标
 */
function getModelIcon(architecture: string): ReactNode {
  const arch = architecture.toLowerCase();

  if (arch.includes('qwen')) return <Brain className="w-5 h-5 text-blue-500" />;
  if (arch.includes('llama')) return <Brain className="w-5 h-5 text-amber-500" />;
  if (arch.includes('mistral')) return <Brain className="w-5 h-5 text-purple-500" />;
  if (arch.includes('gemma')) return <Brain className="w-5 h-5 text-pink-500" />;
  if (arch.includes('deepseek')) return <Brain className="w-5 h-5 text-cyan-500" />;

  return <Brain className="w-5 h-5 text-gray-500" />;
}

export function ModelCard({ model, onLoad, onUnload, onToggleFavourite, actions }: ModelCardProps) {
  const statusColor = STATUS_COLORS[model.status];
  const statusLabel = STATUS_LABELS[model.status];
  const isLoading = model.status === 'loading' || model.isLoading;
  const isLoaded = model.status === 'loaded' || model.status === 'running' || model.isLoaded;

  // 调试日志：检查 Qwen3.5-397B 模型的数据
  if (model.name.includes('Qwen3.5-397B')) {
    console.log('[ModelCard] Qwen3.5-397B 渲染数据:', {
      name: model.name,
      size: model.size,
      totalSize: model.totalSize,
      shardCount: model.shardCount,
      displaySize: formatSize(model.totalSize ?? model.size),
    });
  }

  return (
    <div className="group relative bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 hover:shadow-lg transition-all duration-200">
      {/* 收藏按钮 */}
      <Button
        onClick={onToggleFavourite}
        variant="ghost"
        size="icon"
        className={cn(
          'absolute top-3 right-3 h-8 w-8 rounded-full',
          model.favourite && 'text-yellow-500 hover:text-yellow-600 hover:bg-yellow-50 dark:hover:bg-yellow-900/20'
        )}
      >
        <Star className={cn('w-5 h-5', model.favourite && 'fill-current')} />
      </Button>

      {/* 模型图标和名称 */}
      <div className="flex items-start gap-3 mb-3">
        <div className="p-2 bg-gray-100 dark:bg-gray-700 rounded-lg">
          {getModelIcon(model.metadata.architecture)}
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-gray-900 dark:text-white truncate">
            {model.alias || model.displayName || model.name}
          </h3>
          {model.pathPrefix && (
            <p className="text-xs text-gray-400 dark:text-gray-300 truncate">
              来自: {model.pathPrefix}
            </p>
          )}
        </div>
      </div>

      {/* 模型元数据 */}
      <div className="space-y-2 mb-3">
        {/* 架构 */}
        <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
          <Cpu className="w-4 h-4" />
          <span className="truncate">{model.metadata.architecture}</span>
        </div>

        {/* 大小 - 优先使用 totalSize（分卷模型的总大小） */}
        <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
          <HardDrive className="w-4 h-4" />
          <span>
            {formatSize(model.totalSize ?? model.size)}
            {model.shardCount && model.shardCount > 1 && (
              <span className="ml-1 text-xs text-gray-500 dark:text-gray-400">
                ({model.shardCount} 分卷)
              </span>
            )}
          </span>
        </div>

        {/* 量化 - 优先使用 fileTypeDescriptor（更详细的描述） */}
        {(model.metadata.fileTypeDescriptor || model.metadata.quantization) && (
          <div className="text-sm text-gray-600 dark:text-gray-300">
            量化: {model.metadata.fileTypeDescriptor || model.metadata.quantization}
          </div>
        )}

        {/* 上下文长度 - 只有大于 0 时才显示 */}
        {model.metadata.contextLength && model.metadata.contextLength > 0 && (
          <div className="text-sm text-gray-600 dark:text-gray-300">
            上下文: {model.metadata.contextLength?.toString()}
          </div>
        )}
      </div>

      {/* 状态和标签 */}
      <div className="flex items-center gap-2 mb-3">
        <span className={cn('px-2 py-1 rounded-md text-xs font-medium', statusColor)}>
          {isLoading ? (
            <>
              <Loader2 className="w-3 h-3 inline mr-1 animate-spin" />
              {statusLabel}
            </>
          ) : (
            statusLabel
          )}
        </span>

        {model.isMultimodal && (
          <span className="px-2 py-1 bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300 rounded-md text-xs font-medium">
            多模态
          </span>
        )}

        {isLoaded && model.port && (
          <span className="px-2 py-1 bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300 rounded-md text-xs font-medium">
            端口: {model.port}
          </span>
        )}
      </div>

      {/* 槽位信息 */}
      {isLoaded && model.slots && model.slots.length > 0 && (
        <div className="mb-3 p-2 bg-gray-50 dark:bg-gray-900 rounded-md">
          <div className="text-xs text-gray-600 dark:text-gray-300 mb-1">处理槽位</div>
          <div className="flex gap-1 flex-wrap">
            {model.slots.map((slot) => (
              <div
                key={slot.id}
                className={cn(
                  'w-6 h-6 rounded flex items-center justify-center text-xs',
                  slot.isProcessing
                    ? 'bg-green-500 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
                )}
              >
                {slot.id}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 操作按钮 */}
      <div className="flex items-center gap-2">
        {!isLoaded ? (
          <Button
            onClick={onLoad}
            disabled={isLoading}
            variant="default"
            className="flex-1"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                加载中...
              </>
            ) : (
              <>
                <Play className="w-4 h-4" />
                加载模型
              </>
            )}
          </Button>
        ) : (
          <Button
            onClick={onUnload}
            disabled={model.status === 'unloading'}
            variant="destructive"
            className="flex-1"
          >
            {model.status === 'unloading' ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                卸载中...
              </>
            ) : (
              <>
                <Square className="w-4 h-4" />
                卸载模型
              </>
            )}
          </Button>
        )}

        {actions}
      </div>

      {/* 扫描时间 */}
      <div className="mt-2 text-xs text-gray-400 dark:text-gray-400">
        扫描于: {new Date(model.scannedAt).toLocaleString('zh-CN')}
      </div>
    </div>
  );
}
