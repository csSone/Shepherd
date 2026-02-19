import { Sun, Moon, Monitor, ChevronDown } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { useUIStore, type Theme } from '@/stores/uiStore';
import { cn } from '@/lib/utils';

/**
 * 主题选项
 */
interface ThemeOption {
  value: Theme;
  label: string;
  icon: typeof Sun;
}

const themeOptions: ThemeOption[] = [
  { value: 'light', label: '浅色模式', icon: Sun },
  { value: 'dark', label: '深色模式', icon: Moon },
  { value: 'system', label: '跟随系统', icon: Monitor },
];

/**
 * 主题切换下拉框组件
 */
export function ThemeToggle() {
  const { theme, setTheme } = useUIStore();
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // 点击外部关闭下拉框
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const currentTheme = themeOptions.find((option) => option.value === theme);
  const CurrentIcon = currentTheme?.icon || Monitor;

  return (
    <div ref={containerRef} className="relative w-[140px]">
      {/* 触发按钮 - 带半透明反色边框 */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm font-medium',
          'transition-all duration-200',
          // 半透明边框 - 根据主题自动反色
          'border border-border/40 hover:border-border/60',
          // 半透明背景 - 确保按钮可见
          'bg-background/50 hover:bg-background/80',
          // 状态变化
          'hover:bg-accent',
          'focus:outline-none focus:ring-2 focus:ring-ring focus:border-primary/50',
          // 展开状态
          isOpen && 'bg-accent border-border/60'
        )}
        aria-label={`选择主题（当前：${currentTheme?.label}）`}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        title={currentTheme?.label}
      >
        <div className="flex items-center gap-2">
          <CurrentIcon size={16} />
          <span className="text-xs">{currentTheme?.label}</span>
        </div>
        <ChevronDown
          size={12}
          className={cn(
            'transition-transform duration-200',
            isOpen && 'rotate-180'
          )}
        />
      </button>

      {/* 下拉菜单 */}
      {isOpen && (
        <div
          className={cn(
            'absolute left-0 right-0 top-full z-50 mt-2',
            'rounded-lg border bg-popover shadow-md',
            'animate-in fade-in-0 zoom-in-95'
          )}
          role="listbox"
          aria-label="主题选项"
        >
          <div className="py-1">
            {themeOptions.map((option) => {
              const Icon = option.icon;
              const isSelected = option.value === theme;

              return (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => {
                    setTheme(option.value);
                    setIsOpen(false);
                  }}
                  className={cn(
                    'flex w-full items-center justify-between gap-2 px-3 py-2 text-sm',
                    'transition-colors duration-150',
                    'hover:bg-accent hover:text-accent-foreground',
                    isSelected && 'bg-accent',
                    'focus:outline-none focus:bg-accent'
                  )}
                  role="option"
                  aria-selected={isSelected}
                >
                  <div className="flex items-center gap-2">
                    {/* 图标容器 - 带半透明反色边框 */}
                    <div className={cn(
                      'flex items-center justify-center rounded-md p-1',
                      'border transition-all duration-200',
                      isSelected
                        ? 'border-primary/30 bg-primary/10'
                        : 'border-border/40 hover:border-border/60',
                      'bg-background/50'
                    )}>
                      <Icon size={14} className={cn(
                        'transition-colors duration-200',
                        isSelected ? 'text-primary' : 'text-foreground'
                      )} />
                    </div>
                    <span className={cn(
                      'text-xs',
                      isSelected && 'font-medium'
                    )}>
                      {option.label}
                    </span>
                  </div>
                  {isSelected && (
                    <span className="text-xs">✓</span>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
