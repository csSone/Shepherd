/**
 * SSE 事件类型
 */
export type SSEEventType =
  | 'heartbeat'
  | 'systemStatus'
  | 'modelLoadStart'
  | 'modelLoad'
  | 'modelStop'
  | 'modelSlots'
  | 'console'
  | 'download_progress'
  | 'download_status'
  | 'scan_progress'
  | 'scan_complete'
  | 'clientRegistered'
  | 'clientDisconnected'
  | 'clientResourcesUpdated'
  | 'taskUpdate';

/**
 * SSE 事件基础接口
 */
export interface SSEEvent<T = Record<string, unknown>> {
  type: SSEEventType;
  timestamp: number;
  data: T;
}

/**
 * 任务状态 - 从 task.ts 导入以避免循环依赖
 */
import type { TaskStatus } from './task';
import type { UnifiedNode } from './node';

/**
 * 心跳事件
 */
export interface HeartbeatEvent extends SSEEvent {
  type: 'heartbeat';
}

/**
 * 系统状态事件
 */
export interface SystemStatusEvent extends SSEEvent {
  type: 'systemStatus';
  data: {
    loadedModels: number;
    connections: number;
    confirmedConnections: number;
  };
}

/**
 * 模型加载事件
 */
export interface ModelLoadEvent extends SSEEvent {
  type: 'modelLoad';
  data: {
    modelId: string;
    success: boolean;
    message: string;
    port: number;
  };
}

/**
 * 模型停止事件
 */
export interface ModelStopEvent extends SSEEvent {
  type: 'modelStop';
  data: {
    modelId: string;
    success: boolean;
    message: string;
  };
}

/**
 * 下载进度事件
 */
export interface DownloadProgressEvent extends SSEEvent {
  type: 'download_progress';
  data: {
    taskId: string;
    downloadedBytes: number;
    totalBytes: number;
    partsCompleted: number;
    partsTotal: number;
    progressRatio: number;
  };
}

/**
 * 控制台输出事件
 */
export interface ConsoleEvent extends SSEEvent {
  type: 'console';
  data: {
    modelId: string;
    line64: string; // Base64 编码的日志行
  };
}

/**
 * 客户端注册事件
 */
export interface ClientRegisteredEvent extends SSEEvent {
  type: 'clientRegistered';
  data: {
    clientId: string;
    name: string;
    address: string;
  };
}

/**
 * 客户端断开事件
 */
export interface ClientDisconnectedEvent extends SSEEvent {
  type: 'clientDisconnected';
  data: {
    clientId: string;
    reason?: string;
  };
}

/**
 * 任务更新事件
 */
export interface TaskUpdateEvent extends SSEEvent {
  type: 'taskUpdate';
  data: {
    taskId: string;
    status: TaskStatus;
    result?: Record<string, unknown>;
    error?: string;
  };
}

/**
 * 客户端资源更新事件
 * 后端发送完整节点数据（通过 convertNodeToFrontendFormat）
 */
export interface ClientResourcesUpdatedEvent extends SSEEvent {
  type: 'clientResourcesUpdated';
  data: {
    clientId: string;
    node: UnifiedNode;  // 完整节点数据，与后端 API 格式一致
    timestamp: number;
  };
}
