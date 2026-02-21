import { useState } from 'react';
import { Search, Download, ExternalLink, Loader2, Settings, Key, Globe } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useHuggingFaceSearch, useModelRepoConfig, useAvailableEndpoints, useUpdateModelRepoConfig } from '@/features/downloads/hooks';
import type { HuggingFaceModel } from '@/lib/api/downloads';
import { cn } from '@/lib/utils';

interface HuggingFaceSearchPanelProps {
  onDownload: (model: HuggingFaceModel) => void;
}

export function HuggingFaceSearchPanel({ onDownload }: HuggingFaceSearchPanelProps) {
  const [query, setQuery] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [showSettings, setShowSettings] = useState(false);
  const [tokenInput, setTokenInput] = useState('');
  
  const { data: searchResult, isLoading, error } = useHuggingFaceSearch(query, 20);
  const { data: config, isLoading: configLoading } = useModelRepoConfig();
  const { data: endpoints } = useAvailableEndpoints();
  const updateConfig = useUpdateModelRepoConfig();

  const handleSaveSettings = () => {
    const updates: { endpoint?: string; token?: string } = {};
    if (config && endpoints) {
      const selectedEndpoint = (document.getElementById('endpoint-select') as HTMLSelectElement)?.value;
      if (selectedEndpoint && selectedEndpoint !== config.endpoint) {
        updates.endpoint = selectedEndpoint;
      }
    }
    if (tokenInput) {
      updates.token = tokenInput;
    }
    if (Object.keys(updates).length > 0) {
      updateConfig.mutate(updates, {
        onSuccess: () => {
          setTokenInput('');
          setShowSettings(false);
        }
      });
    } else {
      setShowSettings(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchInput.trim()) {
      setQuery(searchInput.trim());
    }
  };

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        {/* 搜索框 */}
        <form onSubmit={handleSearch} className="flex gap-2 flex-1">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <input
              type="text"
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              placeholder="搜索 HuggingFace 模型，如：llama、qwen、mistral..."
              className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-input text-foreground placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <Button
            type="submit"
            disabled={isLoading || !searchInput.trim()}
            className="whitespace-nowrap"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                搜索中...
              </>
            ) : (
              <>
                <Search className="w-4 h-4 mr-2" />
                搜索
              </>
            )}
          </Button>
        </form>

        {/* 设置按钮 */}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowSettings(!showSettings)}
          className="ml-2"
        >
          <Settings className="w-4 h-4" />
        </Button>
      </div>

      {/* 设置面板 */}
      {showSettings && (
        <div className="p-4 bg-card rounded-lg border border-border space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="font-medium text-foreground flex items-center gap-2">
              <Settings className="w-4 h-4" />
              搜索设置
            </h3>
            <button
              onClick={() => setShowSettings(false)}
              className="text-muted-foreground hover:text-foreground"
            >
              ✕
            </button>
          </div>
          
          {/* 端点选择 */}
          <div>
            <label className="flex items-center gap-2 text-sm font-medium text-foreground mb-1">
              <Globe className="w-4 h-4" />
              API 端点
            </label>
            <select
              id="endpoint-select"
              defaultValue={config?.endpoint || 'huggingface.co'}
              disabled={configLoading}
              className="w-full px-3 py-2 border border-border rounded-md bg-input text-foreground"
            >
              {endpoints && Object.entries(endpoints).map(([value, label]) => (
                <option key={value} value={value}>
                  {label} ({value})
                </option>
              ))}
              {!endpoints && (
                <>
                  <option value="huggingface.co">HuggingFace 官方 (huggingface.co)</option>
                  <option value="hf-mirror.com">HuggingFace 镜像 (hf-mirror.com)</option>
                </>
              )}
            </select>
            <p className="text-xs text-muted-foreground mt-1">
              如果官方站点访问缓慢，可尝试使用镜像站点
            </p>
          </div>

          {/* Token 配置 */}
          <div>
            <label className="flex items-center gap-2 text-sm font-medium text-foreground mb-1">
              <Key className="w-4 h-4" />
              Access Token (可选)
            </label>
            <input
              type="password"
              value={tokenInput}
              onChange={(e) => setTokenInput(e.target.value)}
              placeholder={config?.token ? '已配置 (输入新值替换)' : 'hf_...'}
              className="w-full px-3 py-2 border border-border rounded-md bg-input text-foreground"
            />
            <p className="text-xs text-muted-foreground mt-1">
              用于访问私有模型或提高速率限制，可在 HuggingFace 设置页面获取
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-border">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowSettings(false)}
            >
              取消
            </Button>
            <Button
              size="sm"
              onClick={handleSaveSettings}
              disabled={updateConfig.isPending}
            >
              {updateConfig.isPending ? '保存中...' : '保存'}
            </Button>
          </div>
        </div>
      )}

      {/* 搜索结果 */}
      {error && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400">
          搜索失败: {error.message}
          {error.message?.includes('timeout') && (
            <p className="text-sm mt-1">
              建议尝试切换至镜像站点，点击上方的设置按钮修改端点
            </p>
          )}
        </div>
      )}

      {searchResult && searchResult.items.length === 0 && (
        <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
          <Search className="w-12 h-12 mb-4" />
          <p className="text-lg">未找到模型</p>
          <p className="text-sm">尝试使用不同的关键词搜索</p>
        </div>
      )}

      {searchResult && searchResult.items.length > 0 && (
        <div className="space-y-3">
          <div className="text-sm text-muted-foreground">
            找到 {searchResult.total} 个模型，显示前 {searchResult.items.length} 个
          </div>
          
          {searchResult.items.map((model) => (
            <div
              key={model.id}
              className="p-4 bg-card rounded-lg border border-border hover:border-blue-300 dark:hover:border-blue-700 transition-colors"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <h3 className="font-medium text-foreground truncate">
                    {model.modelId}
                  </h3>
                  <p className="text-sm text-muted-foreground">
                    作者: {model.author}
                  </p>
                  
                  {/* 标签 */}
                  {model.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {model.tags.slice(0, 5).map((tag) => (
                        <span
                          key={tag}
                          className="px-2 py-0.5 text-xs bg-muted text-muted-foreground rounded"
                        >
                          {tag}
                        </span>
                      ))}
                      {model.tags.length > 5 && (
                        <span className="px-2 py-0.5 text-xs text-muted-foreground">
                          +{model.tags.length - 5}
                        </span>
                      )}
                    </div>
                  )}

                  {/* 统计 */}
                  <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
                    <span>下载: {formatNumber(model.downloads)}</span>
                    <span>点赞: {formatNumber(model.likes)}</span>
                  </div>
                </div>

                {/* 操作按钮 */}
                <div className="flex flex-col gap-2">
                  <Button
                    onClick={() => onDownload(model)}
                    size="sm"
                    className="whitespace-nowrap"
                  >
                    <Download className="w-4 h-4 mr-1" />
                    下载
                  </Button>
                  <a
                    href={`https://${config?.endpoint || 'huggingface.co'}/${model.modelId}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={cn(
                      'inline-flex items-center justify-center px-3 py-1.5 text-sm font-medium rounded-md',
                      'bg-muted text-foreground',
                      'hover:bg-muted transition-colors'
                    )}
                  >
                    <ExternalLink className="w-4 h-4 mr-1" />
                    查看
                  </a>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
