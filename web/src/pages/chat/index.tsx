import { useState, useRef, useEffect } from 'react';
import { MessageSquare, Plus, Trash2 } from 'lucide-react';
import { ChatMessage } from '@/components/chat/ChatMessage';
import { ChatInput } from '@/components/chat/ChatInput';
import { useStreamingChat, getLoadedModels, createChatSession, saveChatHistory, loadChatHistory, deleteChatSession } from '@/features/chat/hooks';
import type { ChatMessage as ChatMessageType } from '@/features/chat';

/**
 * 聊天页面
 */
export function ChatPage() {
  const [messages, setMessages] = useState<ChatMessageType[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [currentResponse, setCurrentResponse] = useState('');
  const [models, setModels] = useState<string[]>([]);
  const [selectedModel, setSelectedModel] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const streamingChat = useStreamingChat();

  // 加载可用模型
  useEffect(() => {
    getLoadedModels().then(setModels).catch(console.error);
  }, []);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, currentResponse]);

  // 处理发送消息
  const handleSend = (content: string) => {
    if (!selectedModel) {
      alert('请先选择一个已加载的模型');
      return;
    }

    const userMessage: ChatMessageType = {
      role: 'user',
      content,
      timestamp: Date.now(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setIsStreaming(true);
    setCurrentResponse('');

    streamingChat.mutate(
      {
        model: selectedModel,
        messages: [...messages, userMessage],
        temperature: 0.7,
        topP: 0.95,
        onChunk: (chunk) => {
          setCurrentResponse((prev) => prev + chunk);
        },
        onComplete: (message) => {
          const assistantMessage: ChatMessageType = {
            role: 'assistant',
            content: message,
            timestamp: Date.now(),
          };
          setMessages((prev) => [...prev, assistantMessage]);
          setCurrentResponse('');
          setIsStreaming(false);
        },
        onError: (error) => {
          console.error('Chat error:', error);
          setCurrentResponse('');
          setIsStreaming(false);
        },
      },
      {
        onError: (error) => {
          console.error('Mutation error:', error);
          setIsStreaming(false);
        },
      }
    );
  };

  // 处理停止生成
  const handleStop = () => {
    setIsStreaming(false);
    setCurrentResponse('');
  };

  // 处理新建对话
  const handleNewChat = () => {
    if (messages.length > 0 && !confirm('确定要开始新对话吗？当前对话将被清空。')) {
      return;
    }
    setMessages([]);
    setCurrentResponse('');
  };

  // 处理清空对话
  const handleClearHistory = () => {
    if (confirm('确定要清空对话历史吗？')) {
      setMessages([]);
      setCurrentResponse('');
    }
  };

  return (
    <div className="h-full flex flex-col bg-white dark:bg-gray-800">
      {/* 标题栏 */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50">
        <div className="flex items-center gap-3">
          <MessageSquare className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          <h1 className="text-lg font-semibold text-gray-900 dark:text-white">AI 聊天</h1>
        </div>

        <div className="flex items-center gap-2">
          {/* 模型选择 */}
          <select
            value={selectedModel}
            onChange={(e) => setSelectedModel(e.target.value)}
            className="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-sm text-gray-900 dark:text-white"
          >
            <option value="">选择模型...</option>
            {models.map((model) => (
              <option key={model} value={model}>
                {model}
              </option>
            ))}
          </select>

          <button
            onClick={handleNewChat}
            className="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors"
            title="新建对话"
          >
            <Plus className="w-5 h-5" />
          </button>

          {messages.length > 0 && (
            <button
              onClick={handleClearHistory}
              className="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors"
              title="清空对话"
            >
              <Trash2 className="w-5 h-5" />
            </button>
          )}
        </div>
      </div>

      {/* 消息列表 */}
      <div className="flex-1 overflow-y-auto">
        {messages.length === 0 && !currentResponse ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-500 dark:text-gray-400">
            <MessageSquare className="w-16 h-16 mb-4 opacity-50" />
            <p className="text-lg mb-2">开始对话</p>
            <p className="text-sm">选择一个模型，然后输入消息开始聊天</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {messages.map((message, index) => (
              <ChatMessage key={index} message={message} />
            ))}

            {/* 流式响应 */}
            {currentResponse && (
              <ChatMessage
                message={{
                  role: 'assistant',
                  content: currentResponse,
                  timestamp: Date.now(),
                }}
                isStreaming
              />
            )}
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 输入框 */}
      <ChatInput
        onSend={handleSend}
        onStop={handleStop}
        disabled={!selectedModel || isStreaming}
        isStreaming={isStreaming}
        placeholder={selectedModel ? '输入消息... (按 Enter 发送)' : '请先选择一个模型'}
      />
    </div>
  );
}
