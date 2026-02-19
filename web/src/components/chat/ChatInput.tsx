import { useState, useRef, useEffect } from 'react';
import { Send, Square } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ChatInputProps {
  onSend: (content: string) => void;
  onStop?: () => void;
  disabled?: boolean;
  isStreaming?: boolean;
  placeholder?: string;
}

export function ChatInput({
  onSend,
  onStop,
  disabled = false,
  isStreaming = false,
  placeholder = '输入消息...',
}: ChatInputProps) {
  const [content, setContent] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    // 自动调整高度
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
    }
  }, [content]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = content.trim();
    if (!trimmed || disabled || isStreaming) return;

    onSend(trimmed);
    setContent('');

    // 重置高度
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="border-t border-gray-200 dark:border-gray-700 p-4">
      <div className="flex items-end gap-3">
        <div className="flex-1 relative">
          <textarea
            ref={textareaRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            className={cn(
              'w-full px-4 py-3 pr-12 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400',
              disabled && 'opacity-50 cursor-not-allowed'
            )}
            rows={1}
            style={{ minHeight: '48px', maxHeight: '200px' }}
          />

          {/* 字符计数 */}
          {content.length > 0 && (
            <div className="absolute bottom-2 right-2 text-xs text-gray-400 dark:text-gray-500">
              {content.length.toLocaleString()}
            </div>
          )}
        </div>

        {/* 发送/停止按钮 */}
        {isStreaming ? (
          <button
            type="button"
            onClick={onStop}
            className={cn(
              'flex items-center justify-center w-12 h-12 rounded-lg transition-colors',
              'bg-red-600 hover:bg-red-700 text-white dark:bg-red-500 dark:hover:bg-red-600'
            )}
            title="停止生成"
          >
            <Square className="w-5 h-5" />
          </button>
        ) : (
          <button
            type="submit"
            disabled={!content.trim() || disabled}
            className={cn(
              'flex items-center justify-center w-12 h-12 rounded-lg transition-colors',
              !content.trim() || disabled
                ? 'bg-gray-300 dark:bg-gray-700 text-gray-500 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700 text-white dark:bg-blue-500 dark:hover:bg-blue-600'
            )}
            title="发送 (Enter)"
          >
            <Send className="w-5 h-5" />
          </button>
        )}
      </div>

      {/* 提示信息 */}
      <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
        按 Enter 发送，Shift + Enter 换行
      </div>
    </form>
  );
}
