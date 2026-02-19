import { useState } from 'react';
import { X, Cloud, Database } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { DownloadSource, CreateDownloadParams } from '@/types';

interface CreateDownloadDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (params: CreateDownloadParams) => void;
  isLoading?: boolean;
}

export function CreateDownloadDialog({
  isOpen,
  onClose,
  onConfirm,
  isLoading = false,
}: CreateDownloadDialogProps) {
  const [source, setSource] = useState<DownloadSource>('huggingface');
  const [repoId, setRepoId] = useState('');
  const [fileName, setFileName] = useState('');
  const [path, setPath] = useState('');
  const [maxRetries, setMaxRetries] = useState('3');
  const [chunkSize, setChunkSize] = useState('');

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const params: CreateDownloadParams = {
      source,
      repoId: repoId.trim(),
      path: path || undefined,
      fileName: fileName || undefined,
      maxRetries: Number(maxRetries) || 3,
      chunkSize: chunkSize ? Number(chunkSize) : undefined,
    };

    onConfirm(params);
  };

  const setExample = (src: DownloadSource, exampleRepo: string, exampleFile?: string) => {
    setSource(src);
    setRepoId(exampleRepo);
    setFileName(exampleFile || '');
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
        {/* 标题栏 */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            创建下载任务
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
          {/* 来源选择 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              下载来源
            </label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setSource('huggingface')}
                className={cn(
                  'flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-md border-2 transition-colors',
                  source === 'huggingface'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-300 dark:border-gray-600 hover:border-gray-400'
                )}
              >
                <Cloud className="w-5 h-5" />
                <span className="font-medium">HuggingFace</span>
              </button>
              <button
                type="button"
                onClick={() => setSource('modelscope')}
                className={cn(
                  'flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-md border-2 transition-colors',
                  source === 'modelscope'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-300 dark:border-gray-600 hover:border-gray-400'
                )}
              >
                <Database className="w-5 h-5" />
                <span className="font-medium">ModelScope</span>
              </button>
            </div>
          </div>

          {/* 仓库 ID */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              仓库 ID
            </label>
            <input
              type="text"
              value={repoId}
              onChange={(e) => setRepoId(e.target.value)}
              placeholder={source === 'huggingface' ? 'Qwen/Qwen2-7B-Instruct' : 'Qwen/Qwen2-7B-Instruct'}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              required
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              例如: {source === 'huggingface' ? 'Qwen/Qwen2-7B-Instruct' : 'Qwen/Qwen2-7B-Instruct'}
            </p>
          </div>

          {/* 文件名（可选） */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              文件名（可选）
            </label>
            <input
              type="text"
              value={fileName}
              onChange={(e) => setFileName(e.target.value)}
              placeholder="qwen2-7b-instruct-q4_k_m.gguf"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              留空则下载所有文件
            </p>
          </div>

          {/* 保存路径 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              保存路径（可选）
            </label>
            <input
              type="text"
              value={path}
              onChange={(e) => setPath(e.target.value)}
              placeholder="/path/to/models"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              留空则使用默认路径
            </p>
          </div>

          {/* 高级选项 */}
          <details className="group">
            <summary className="cursor-pointer text-sm font-medium text-gray-700 dark:text-gray-300 list-none flex items-center gap-2">
              <span className="transform group-open:rotate-90 transition-transform">▶</span>
              高级选项
            </summary>

            <div className="mt-3 space-y-3 pl-5">
              {/* 最大重试次数 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  最大重试次数
                </label>
                <input
                  type="number"
                  value={maxRetries}
                  onChange={(e) => setMaxRetries(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  min={0}
                  max={10}
                />
              </div>

              {/* 分块大小 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  分块大小（字节）
                </label>
                <input
                  type="number"
                  value={chunkSize}
                  onChange={(e) => setChunkSize(e.target.value)}
                  placeholder="默认: 10485760 (10MB)"
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  min={1024}
                  step={1024}
                />
              </div>
            </div>
          </details>

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
              disabled={isLoading || !repoId.trim()}
              className={cn(
                'px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 transition-colors',
                (!repoId.trim() || isLoading) && 'opacity-50 cursor-not-allowed'
              )}
            >
              {isLoading ? '创建中...' : '创建任务'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
