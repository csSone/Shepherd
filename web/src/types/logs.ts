/**
 * 日志级别
 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'fatal';

/**
 * 日志来源
 */
export type LogSource = 'system' | 'model' | 'download' | 'cluster' | 'client';

/**
 * 日志条目
 */
export interface LogEntry {
  timestamp: number;
  level: LogLevel;
  source: LogSource;
  message: string;
  modelId?: string;
  clientId?: string;
  metadata?: Record<string, unknown>;
}

/**
 * 日志过滤参数
 */
export interface LogFilters {
  level?: LogLevel;
  source?: LogSource;
  search?: string;
  modelId?: string;
  clientId?: string;
  startTime?: number;
  endTime?: number;
}

/**
 * 日志统计
 */
export interface LogStats {
  total: number;
  byLevel: Record<LogLevel, number>;
  bySource: Record<LogSource, number>;
}
