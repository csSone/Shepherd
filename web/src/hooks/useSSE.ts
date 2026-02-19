import { useEffect, useRef, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { SSEEvent } from '@/types';
import { apiClient } from '@/lib/api/client';

/**
 * useSSE Hook 配置选项
 */
interface UseSSEOptions {
  onMessage?: (event: SSEEvent) => void;
  onError?: (error: Event) => void;
  onOpen?: () => void;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

/**
 * SSE 事件监听 Hook
 *
 * 自动连接到 SSE 端点，处理重连，并根据事件类型使相关查询失效
 */
export function useSSE(options: UseSSEOptions = {}) {
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const queryClient = useQueryClient();

  const {
    onMessage,
    onError,
    onOpen,
    reconnectInterval = 1000,
    maxReconnectAttempts = 10,
  } = options;

  /**
   * 处理消息事件
   */
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const data: SSEEvent = JSON.parse(event.data);

        // 根据事件类型使相关查询失效
        switch (data.type) {
          case 'modelLoad':
          case 'modelLoadStart':
          case 'modelStop':
            queryClient.invalidateQueries({ queryKey: ['models'] });
            break;
          case 'download_progress':
          case 'download_status':
            queryClient.invalidateQueries({ queryKey: ['downloads'] });
            break;
          case 'clientRegistered':
          case 'clientDisconnected':
            queryClient.invalidateQueries({ queryKey: ['clients'] });
            queryClient.invalidateQueries({ queryKey: ['cluster'] });
            break;
          case 'taskUpdate':
            queryClient.invalidateQueries({ queryKey: ['tasks'] });
            queryClient.invalidateQueries({ queryKey: ['cluster'] });
            break;
          case 'systemStatus':
            queryClient.invalidateQueries({ queryKey: ['system'] });
            break;
        }

        onMessage?.(data);
      } catch (error) {
        console.error('Failed to parse SSE event:', error);
      }
    },
    [onMessage, queryClient]
  );

  /**
   * 处理错误事件（实现指数退避重连）
   */
  const handleError = useCallback(
    (error: Event) => {
      console.error('SSE connection error:', error);

      if (reconnectAttemptsRef.current < maxReconnectAttempts) {
        const delay = Math.min(
          reconnectInterval * Math.pow(2, reconnectAttemptsRef.current),
          30000
        );

        reconnectTimeoutRef.current = setTimeout(() => {
          reconnectAttemptsRef.current++;
          connect();
        }, delay);
      }

      onError?.(error);
    },
    [maxReconnectAttempts, onError, reconnectInterval]
  );

  /**
   * 处理连接打开事件
   */
  const handleOpen = useCallback(() => {
    console.log('SSE connection opened');
    reconnectAttemptsRef.current = 0;
    onOpen?.();
  }, [onOpen]);

  /**
   * 连接到 SSE 端点
   */
  const connect = useCallback(() => {
    // 如果已经连接，不重复连接
    if (
      eventSourceRef.current?.readyState === EventSource.OPEN ||
      eventSourceRef.current?.readyState === EventSource.CONNECTING
    ) {
      return;
    }

    // 清理旧连接
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    // 清除待处理的重连
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    try {
      // 从 API 客户端获取后端 URL
      const baseUrl = apiClient.getBaseUrl();
      const sseUrl = `${baseUrl}/events`;

      console.log('Connecting to SSE endpoint:', sseUrl);

      const es = new EventSource(sseUrl);
      es.addEventListener('message', handleMessage);
      es.addEventListener('error', handleError);
      es.addEventListener('open', handleOpen);

      eventSourceRef.current = es;
    } catch (error) {
      console.error('Failed to create EventSource:', error);
      handleError(error as Event);
    }
  }, [handleMessage, handleError, handleOpen]);

  /**
   * 断开 SSE 连接
   */
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
  }, []);

  /**
   * 组件挂载时连接，卸载时断开
   */
  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  return { connect, disconnect };
}
