/**
 * 客户端状态
 */
export type ClientStatus = 'offline' | 'online' | 'busy' | 'error' | 'disabled';

/**
 * 客户端能力
 */
export interface Capabilities {
  cpuCount: number;
  memory: number; // bytes
  gpuCount: number;
  gpuMemory: number; // bytes
  supportsLlama: boolean;
  supportsPython: boolean;
}

/**
 * 资源使用情况
 */
export interface ResourceUsage {
  cpuPercent: number;
  memoryUsed: number; // bytes
  memoryTotal: number; // bytes
  gpuPercent: number;
  gpuMemoryUsed: number; // bytes
  gpuMemoryTotal: number; // bytes
  diskUsed: number; // bytes
  diskTotal: number; // bytes
}

/**
 * 客户端节点
 */
export interface Client {
  id: string;
  name: string;
  address: string;
  port: number;
  tags: string[];
  capabilities: Capabilities;
  resources?: ResourceUsage;
  status: ClientStatus;
  lastSeen: string;
  connected: boolean;
  metadata: Record<string, string>;
}

/**
 * 任务类型
 */
export type TaskType = 'load_model' | 'unload_model' | 'run_python' | 'run_llamacpp' | 'custom';

/**
 * 任务状态
 */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

/**
 * 调度策略
 */
export type ScheduleStrategy = 'round_robin' | 'least_loaded' | 'resource_aware';

/**
 * 调度任务
 */
export interface ClusterTask {
  id: string;
  type: TaskType;
  payload: Record<string, unknown>;
  assignedTo?: string;
  status: TaskStatus;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
  result?: Record<string, unknown>;
  error?: string;
  retryCount: number;
  maxRetries: number;
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
 * 集群概览
 */
export interface ClusterOverview {
  totalClients: number;
  onlineClients: number;
  offlineClients: number;
  busyClients: number;
  totalTasks: number;
  runningTasks: number;
  completedTasks: number;
  failedTasks: number;
}

/**
 * 客户端列表响应
 */
export interface ClientListResponse {
  clients: Client[];
  total: number;
}

/**
 * 任务列表响应
 */
export interface TaskListResponse {
  tasks: ClusterTask[];
  total: number;
}
