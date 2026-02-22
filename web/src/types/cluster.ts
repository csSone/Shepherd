/**
 * @deprecated 请使用 web/src/types/node.ts 中的统一类型定义
 *
 * 迁移指南：
 * - Client -> UnifiedNode
 * - Capabilities -> NodeCapabilities
 * - ClientStatus -> NodeStatus
 * - ResourceUsage -> NodeResources
 *
 * 这个文件将在 v0.4.0 版本完全移除
 */

import type { UnifiedNode } from './node';

// 重新导出统一类型以保持向后兼容
export type {
  UnifiedNode as Client,
  NodeCapabilities as Capabilities,
  NodeStatus as ClientStatus,
  NodeResources as ResourceUsage,
  GPUInfo,
  NodeRole,
} from './node';

/**
 * 任务类型 - 匹配后端 cluster.TaskType
 */
export type TaskType = 'load_model' | 'unload_model' | 'run_python' | 'run_llamacpp' | 'custom';

/**
 * 任务状态 - 匹配后端 cluster.TaskStatus
 */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

/**
 * 调度策略 - 匹配后端 scheduler.SchedulingStrategy
 */
export type ScheduleStrategy = 'round_robin' | 'least_loaded' | 'resource_aware';

/**
 * 调度任务 - 匹配后端 cluster.Task 结构
 */
export interface ClusterTask {
  id: string;
  type: TaskType;
  payload: Record<string, unknown>;
  assignedTo?: string;  // 分配的客户端 ID
  status: TaskStatus;
  createdAt: string;  // ISO 8601 格式
  startedAt?: string;  // ISO 8601 格式
  completedAt?: string;  // ISO 8601 格式
  result?: Record<string, unknown>;
  error?: string;
  retryCount?: number;  // 可选，后端 Task 结构体未包含此字段
  maxRetries?: number;  // 可选，后端 Task 结构体未包含此字段
}

/**
 * 扫描状态
 */
export interface ScanStatus {
  running: boolean;
  found: DiscoveredClient[];
}

/**
 * 发现的客户端
 */
export interface DiscoveredClient {
  address: string;
  port: number;
  respondedAt: string;
}

/**
 * 集群概览 - 匹配后端 GET /api/master/overview 响应格式
 */
export interface ClusterOverview {
  totalClients: number;
  onlineClients: number;
  offlineClients: number;
  busyClients: number;
  totalTasks: number;
  pendingTasks: number;  // 待处理任务数
  runningTasks: number;
  completedTasks: number;
  failedTasks: number;
  nodes?: {
    stats: {
      total: number;
      online: number;
      offline: number;
      busy: number;
    };
  };
}

/**
 * 客户端列表响应 - 匹配后端 GET /api/master/clients 响应格式
 */
export interface ClientListResponse {
  clients: UnifiedNode[];
  total: number;
  stats?: {
    total: number;
    online: number;
    offline: number;
    busy: number;
  };
}

/**
 * 任务列表响应
 */
export interface TaskListResponse {
  tasks: ClusterTask[];
  total: number;
}
