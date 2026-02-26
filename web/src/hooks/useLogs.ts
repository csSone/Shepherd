import { useState, useEffect, useCallback, useRef } from 'react';
import type { LogEntry, LogLevel, LogSource } from '@/types/logs';
import { apiClient } from '@/lib/api/client';

/**
 * SSE 日志流返回的原始数据格式
 */
interface SSELogEntry {
  timestamp: string;
  level: string;
  message: string;
  fields?: Record<string, unknown>;
}

/**
 * useLogs Hook 配置选项
 */
interface UseLogsOptions {
  maxLogs?: number;
  autoConnect?: boolean;
  fromBeginning?: boolean; // 是否从头开始加载历史日志
}

/**
 * 使用日志流 Hook
 *
 * 自动连接 SSE 日志流，实时接收后端日志
 *
 * @example
 * const { logs, isConnected, error, reconnect } = useLogs({
 *   maxLogs: 1000,
 *   fromBeginning: true  // 页面刷新后加载本次运行的所有历史日志
 * });
 */
export function useLogs(options: UseLogsOptions = {}) {
  const { maxLogs = 1000, autoConnect = true, fromBeginning = true } = options;
  
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const eventSourceRef = useRef<EventSource | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 10;

  /**
   * 将 SSE 日志条目转换为 LogEntry
   */
  const convertSSEToLogEntry = useCallback((data: SSELogEntry): LogEntry => {
    // 解析时间戳
    const timestamp = data.timestamp 
      ? new Date(data.timestamp).getTime() 
      : Date.now();
    
    // 映射日志级别
    const levelMap: Record<string, LogLevel> = {
      'DEBUG': 'debug',
      'INFO': 'info',
      'WARN': 'warn',
      'ERROR': 'error',
      'FATAL': 'fatal',
    };
    const level = levelMap[data.level?.toUpperCase()] || 'info';
    
    // 从 fields 中提取来源信息
    const fields = data.fields || {};
    const source = (fields['source'] as LogSource) || 'system';
    const modelId = fields['modelId'] as string | undefined;
    const clientId = fields['clientId'] as string | undefined;
    
    return {
      timestamp,
      level,
      source,
      message: data.message || '',
      modelId,
      clientId,
      metadata: fields,
    };
  }, []);

  /**
   * 处理 SSE 消息
   */
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data: SSELogEntry = JSON.parse(event.data);
      const entry = convertSSEToLogEntry(data);
      
      setLogs(prev => {
        const newLogs = [...prev, entry];
        // 限制日志数量
        if (newLogs.length > maxLogs) {
          return newLogs.slice(-maxLogs);
        }
        return newLogs;
      });
    } catch (err) {
      console.error('Failed to parse log entry:', err);
    }
  }, [convertSSEToLogEntry, maxLogs]);

  /**
   * 处理连接打开
   */
  const handleOpen = useCallback(() => {
    console.log('Log stream connected');
    setIsConnected(true);
    setError(null);
    reconnectAttemptsRef.current = 0;
  }, []);

  /**
   * 处理连接错误
   */
  const handleError = useCallback(() => {
    console.error('Log stream connection error');
    setIsConnected(false);
    setError('连接日志流失败');
    
    // 尝试重连
    if (reconnectAttemptsRef.current < maxReconnectAttempts) {
      const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
      reconnectTimeoutRef.current = setTimeout(() => {
        reconnectAttemptsRef.current++;
        connect();
      }, delay);
    }
  }, []);

  /**
   * 连接到日志流
   */
  const connect = useCallback(() => {
    // 关闭旧连接
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }

    try {
      const baseUrl = apiClient.getBaseUrl();
      // 添加 fromBeginning 参数以加载历史日志
      const sseUrl = `${baseUrl}/logs/stream?fromBeginning=${fromBeginning ? 'true' : 'false'}`;

      console.log('Connecting to log stream:', sseUrl);

      const es = new EventSource(sseUrl);
      es.addEventListener('message', handleMessage);
      es.addEventListener('open', handleOpen);
      es.addEventListener('error', handleError);

      eventSourceRef.current = es;
    } catch (err) {
      console.error('Failed to create EventSource:', err);
      setError('创建日志流连接失败');
    }
  }, [handleMessage, handleOpen, handleError, fromBeginning]);

  /**
   * 断开连接
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
    
    setIsConnected(false);
  }, []);

  /**
   * 重新连接
   */
  const reconnect = useCallback(() => {
    reconnectAttemptsRef.current = 0;
    connect();
  }, [connect]);

  /**
   * 清空日志
   */
  const clearLogs = useCallback(() => {
    setLogs([]);
  }, []);

  /**
   * 自动连接
   */
  useEffect(() => {
    if (autoConnect) {
      connect();
    }
    
    return () => {
      disconnect();
    };
  }, [autoConnect, connect, disconnect]);

  return {
    logs,
    isConnected,
    error,
    reconnect,
    disconnect,
    clearLogs,
  };
}
