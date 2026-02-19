import { useState } from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { LoadModelParams } from '@/types';

interface LoadModelDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (params: LoadModelParams) => void;
  modelId: string;
  modelName: string;
  isLoading?: boolean;
}

export function LoadModelDialog({
  isOpen,
  onClose,
  onConfirm,
  modelId,
  modelName,
  isLoading = false,
}: LoadModelDialogProps) {
  const [params, setParams] = useState<LoadModelParams>({
    modelId,
    ctxSize: 8192,
    batchSize: 512,
    threads: 4,
    gpuLayers: 99,
    temperature: 0.7,
    topP: 0.95,
    topK: 40,
    repeatPenalty: 1.1,
    seed: -1,
    nPredict: -1,
  });

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onConfirm(params);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        {/* 标题栏 */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            加载模型配置
          </h2>
          <button
            onClick={onClose}
            disabled={isLoading}
            className="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-50"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 表单内容 */}
        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          {/* 模型名称 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              模型
            </label>
            <div className="px-3 py-2 bg-gray-100 dark:bg-gray-700 rounded-md text-gray-900 dark:text-white">
              {modelName}
            </div>
          </div>

          {/* 两列布局 */}
          <div className="grid grid-cols-2 gap-4">
            {/* 上下文大小 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                上下文大小
              </label>
              <input
                type="number"
                value={params.ctxSize}
                onChange={(e) => setParams({ ...params, ctxSize: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={512}
                max={131072}
                step={512}
              />
            </div>

            {/* 批次大小 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                批次大小
              </label>
              <input
                type="number"
                value={params.batchSize}
                onChange={(e) => setParams({ ...params, batchSize: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={1}
                max={8192}
              />
            </div>

            {/* 线程数 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                线程数
              </label>
              <input
                type="number"
                value={params.threads}
                onChange={(e) => setParams({ ...params, threads: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={1}
                max={256}
              />
            </div>

            {/* GPU 层数 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                GPU 层数
              </label>
              <input
                type="number"
                value={params.gpuLayers}
                onChange={(e) => setParams({ ...params, gpuLayers: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={0}
                max={999}
              />
            </div>

            {/* Temperature */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Temperature
              </label>
              <input
                type="number"
                value={params.temperature}
                onChange={(e) => setParams({ ...params, temperature: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={0}
                max={2}
                step={0.1}
              />
            </div>

            {/* Top P */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Top P
              </label>
              <input
                type="number"
                value={params.topP}
                onChange={(e) => setParams({ ...params, topP: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={0}
                max={1}
                step={0.05}
              />
            </div>

            {/* Top K */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Top K
              </label>
              <input
                type="number"
                value={params.topK}
                onChange={(e) => setParams({ ...params, topK: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={0}
                max={100}
              />
            </div>

            {/* Repeat Penalty */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Repeat Penalty
              </label>
              <input
                type="number"
                value={params.repeatPenalty}
                onChange={(e) => setParams({ ...params, repeatPenalty: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={0}
                max={2}
                step={0.1}
              />
            </div>

            {/* Seed */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Seed (-1 随机)
              </label>
              <input
                type="number"
                value={params.seed}
                onChange={(e) => setParams({ ...params, seed: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={-1}
              />
            </div>

            {/* Max Tokens */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Max Tokens (-1 无限)
              </label>
              <input
                type="number"
                value={params.nPredict}
                onChange={(e) => setParams({ ...params, nPredict: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                min={-1}
              />
            </div>
          </div>

          {/* 按钮 */}
          <div className="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
            <button
              type="button"
              onClick={onClose}
              disabled={isLoading}
              className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md transition-colors disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className={cn(
                'px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 transition-colors',
                isLoading && 'opacity-50 cursor-not-allowed'
              )}
            >
              {isLoading ? '加载中...' : '开始加载'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
