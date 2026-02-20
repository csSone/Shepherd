import { useState } from 'react';
import {
  Settings as SettingsIcon,
  Zap,
  Toolbox,
  Info,
  Plug,
  Save,
  FolderOpen,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { PathConfigPanel } from '@/components/settings/PathConfigPanel';

/**
 * 设置标签类型
 */
type SettingsTab = 'general' | 'paths' | 'benchmark' | 'mcp' | 'about';

/**
 * 设置菜单项
 */
interface SettingsMenuItem {
  id: SettingsTab;
  icon: typeof SettingsIcon;
  label: string;
}

const settingsMenuItems: SettingsMenuItem[] = [
  { id: 'general', icon: SettingsIcon, label: '通用设置' },
  { id: 'paths', icon: FolderOpen, label: '路径配置' },
  { id: 'benchmark', icon: Zap, label: '性能压测' },
  { id: 'mcp', icon: Toolbox, label: 'MCP 管理' },
  { id: 'about', icon: Info, label: '关于' },
];

/**
 * 设置页面组件
 */
export function SettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('general');

  return (
    <div className="h-full">
      {/* 顶部标题栏 */}
      <div className="border-b px-5 py-3">
        <h1 className="text-xl font-semibold">设置</h1>
      </div>

      {/* 设置内容区域 */}
      <div className="flex h-[calc(100%-53px)]">
        {/* 左侧菜单 */}
        <div className="w-48 border-r bg-background p-3">
          <nav className="space-y-1" role="tablist" aria-label="设置菜单">
            {settingsMenuItems.map((item) => {
              const Icon = item.icon;
              const isActive = activeTab === item.id;

              return (
                <button
                  key={item.id}
                  type="button"
                  role="tab"
                  aria-selected={isActive}
                  onClick={() => setActiveTab(item.id)}
                  className={cn(
                    'flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-xs font-medium transition-all duration-200',
                    isActive
                      ? 'bg-primary text-primary-foreground shadow-sm'
                      : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                  )}
                >
                  <Icon size={16} />
                  <span>{item.label}</span>
                </button>
              );
            })}
          </nav>
        </div>

        {/* 右侧内容 */}
        <div className="flex-1 overflow-y-auto p-5">
          {activeTab === 'general' && <GeneralSettingsPanel />}
          {activeTab === 'paths' && <PathsSettingsPanel />}
          {activeTab === 'benchmark' && <BenchmarkPanel />}
          {activeTab === 'mcp' && <McpPanel />}
          {activeTab === 'about' && <AboutPanel />}
        </div>
      </div>
    </div>
  );
}

/**
 * 通用设置面板
 */
function GeneralSettingsPanel() {
  const [ollamaEnabled, setOllamaEnabled] = useState(false);
  const [ollamaPort, setOllamaPort] = useState('11434');
  const [lmstudioEnabled, setLmstudioEnabled] = useState(false);
  const [lmstudioPort, setLmstudioPort] = useState('1234');
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'success' | 'error'>('idle');

  const handleSave = async () => {
    setSaveStatus('saving');
    try {
      // TODO: 调用 API 保存配置
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setSaveStatus('success');
      setTimeout(() => setSaveStatus('idle'), 2000);
    } catch {
      setSaveStatus('error');
      setTimeout(() => setSaveStatus('idle'), 2000);
    }
  };

  return (
    <div className="max-w-2xl space-y-4">
      <div>
        <h2 className="text-lg font-semibold">API 兼容性设置</h2>
        <p className="text-xs text-muted-foreground">
          配置 Ollama 和 LM Studio API 兼容层端口
        </p>
      </div>

      {/* Ollama 配置 */}
      <div className="rounded-lg border bg-card p-4">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <Plug size={16} className="text-primary" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">Ollama API</h3>
              <p className="text-xs text-muted-foreground">
                启用 Ollama 兼容的 API 端点
              </p>
            </div>
          </div>
          <label className="relative inline-flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={ollamaEnabled}
              onChange={(e) => setOllamaEnabled(e.target.checked)}
              className="peer sr-only"
            />
            <div className="h-5 w-9 rounded-full bg-muted peer-checked:bg-primary transition-colors duration-200 after:content-[''] after:absolute after:left-[2px] after:top-[2px] after:h-4 after:w-4 after:rounded-full after:bg-background after:transition-transform after:duration-200 peer-checked:after:translate-x-full" />
            <span className="text-xs text-muted-foreground">启用</span>
          </label>
        </div>

        {ollamaEnabled && (
          <div className="mt-3">
            <label className="block text-xs font-medium mb-1.5">端口</label>
            <input
              type="number"
              min="1"
              max="65535"
              value={ollamaPort}
              onChange={(e) => setOllamaPort(e.target.value)}
              className="w-full max-w-[160px] rounded-md border bg-background px-2.5 py-1.5 text-xs"
              placeholder="11434"
            />
          </div>
        )}
      </div>

      {/* LM Studio 配置 */}
      <div className="rounded-lg border bg-card p-4">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <Plug size={16} className="text-primary" />
            </div>
            <div>
              <h3 className="text-sm font-semibold">LM Studio API</h3>
              <p className="text-xs text-muted-foreground">
                启用 LM Studio 兼容的 API 端点
              </p>
            </div>
          </div>
          <label className="relative inline-flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={lmstudioEnabled}
              onChange={(e) => setLmstudioEnabled(e.target.checked)}
              className="peer sr-only"
            />
            <div className="h-5 w-9 rounded-full bg-muted peer-checked:bg-primary transition-colors duration-200 after:content-[''] after:absolute after:left-[2px] after:top-[2px] after:h-4 after:w-4 after:rounded-full after:bg-background after:transition-transform after:duration-200 peer-checked:after:translate-x-full" />
            <span className="text-xs text-muted-foreground">启用</span>
          </label>
        </div>

        {lmstudioEnabled && (
          <div className="mt-3">
            <label className="block text-xs font-medium mb-1.5">端口</label>
            <input
              type="number"
              min="1"
              max="65535"
              value={lmstudioPort}
              onChange={(e) => setLmstudioPort(e.target.value)}
              className="w-full max-w-[160px] rounded-md border bg-background px-2.5 py-1.5 text-xs"
              placeholder="1234"
            />
          </div>
        )}
      </div>

      {/* 保存按钮 */}
      <div className="flex items-center gap-2">
        <button
          onClick={handleSave}
          disabled={saveStatus === 'saving'}
          className="flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Save size={14} />
          {saveStatus === 'saving' ? '保存中...' :
           saveStatus === 'success' ? '已保存 ✓' :
           saveStatus === 'error' ? '保存失败' :
           '保存设置'}
        </button>
      </div>
    </div>
  );
}

/**
 * 性能压测面板
 */
function BenchmarkPanel() {
  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-center">
        <Zap size={48} className="mx-auto mb-4 text-muted-foreground" />
        <h3 className="text-lg font-semibold">性能压测</h3>
        <p className="text-sm text-muted-foreground mt-2">
          性能压测功能开发中...
        </p>
      </div>
    </div>
  );
}

/**
 * MCP 管理面板
 */
function McpPanel() {
  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-center">
        <Toolbox size={48} className="mx-auto mb-4 text-muted-foreground" />
        <h3 className="text-lg font-semibold">MCP 管理</h3>
        <p className="text-sm text-muted-foreground mt-2">
          MCP (Model Context Protocol) 管理功能开发中...
        </p>
      </div>
    </div>
  );
}

/**
 * 关于面板
 */
function AboutPanel() {
  return (
    <div className="max-w-2xl mx-auto">
      <div className="text-center mb-6">
        <div className="flex h-16 w-16 items-center justify-center rounded-xl bg-primary mx-auto mb-3 text-2xl">
          🐏
        </div>
        <h2 className="text-xl font-bold">Shepherd</h2>
        <p className="text-sm text-muted-foreground">高性能轻量级 llama.cpp 模型管理系统</p>
      </div>

      <div className="rounded-lg border bg-card p-4 space-y-2">
        <div className="flex items-center justify-between py-1.5 border-b">
          <span className="text-sm text-muted-foreground">版本</span>
          <span className="font-mono text-sm font-medium">v0.1.3</span>
        </div>
        <div className="flex items-center justify-between py-1.5 border-b">
          <span className="text-sm text-muted-foreground">构建时间</span>
          <span className="font-mono text-xs">2026-02-19</span>
        </div>
        <div className="flex items-center justify-between py-1.5 border-b">
          <span className="text-sm text-muted-foreground">Go 版本</span>
          <span className="font-mono text-xs">1.25+</span>
        </div>
        <div className="flex items-center justify-between py-1.5 border-b">
          <span className="text-sm text-muted-foreground">React 版本</span>
          <span className="font-mono text-xs">19.x</span>
        </div>
        <div className="flex items-center justify-between py-1.5">
          <span className="text-sm text-muted-foreground">许可证</span>
          <span className="text-xs">Apache 2.0</span>
        </div>
      </div>

      <div className="mt-4 text-center text-xs text-muted-foreground">
        <p>© 2026 Shepherd Project. Licensed under Apache 2.0</p>
        <p className="mt-1">
          <a
            href="https://github.com/shepherd-project/shepherd"
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:underline"
          >
            GitHub Repository
          </a>
        </p>
      </div>
    </div>
  );
}

/**
 * 路径配置面板
 */
function PathsSettingsPanel() {
  return (
    <div className="max-w-3xl space-y-5">
      {/* llama.cpp 路径配置 */}
      <div className="rounded-lg border bg-card p-4">
        <PathConfigPanel type="llamacpp" />
      </div>

      {/* 分隔线 */}
      <div className="border-t" />

      {/* 模型路径配置 */}
      <div className="rounded-lg border bg-card p-4">
        <PathConfigPanel type="models" />
      </div>
    </div>
  );
}
