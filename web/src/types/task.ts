/**
 * 任务相关类型定义
 * Task-related type definitions
 * 
 * 从 cluster.ts 迁移至此，避免循环依赖
 * Migrated from cluster.ts to avoid circular dependencies
 */

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
 * 任务列表响应
 */
export interface TaskListResponse {
  tasks: ClusterTask[];
  total: number;
}
