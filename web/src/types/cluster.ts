/**
 * 客户端状态 - 统一与 NodeStatus 保持一致
 */
export type ClientStatus = 'offline' | 'online' | 'busy' | 'error' | 'degraded' | 'disabled';

/**
 * 客户端能力 - 匹配后端 cluster.Capabilities
 * 注意：后端 cluster.Capabilities 没有 GPUCount，但 node.NodeCapabilities 有
 * 这里添加 gpuCount 以兼容前端组件显示需求
 */
export interface Capabilities {
  gpu: boolean;
  gpuName?: string;
  gpuMemory?: number;  // bytes
  gpuCount?: number;   // GPU 数量，来自 node.NodeCapabilities
  cpuCount: number;
  memory: number;  // bytes
  supportsLlama: boolean;
  supportsPython: boolean;
  condaEnvs?: string[];
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
 * 客户端节点 - 匹配后端 cluster.Client
 */
export interface Client {
  id: string;
  name: string;
  address: string;
  port: number;
  tags: string[];
  capabilities?: Capabilities;
  resources?: ResourceUsage;
  status: ClientStatus;
  lastSeen: string;  // ISO 8601 格式
  connected: boolean;
  metadata: Record<string, string>;
}

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
  pendingTasks: number;  // 新增：待处理任务数
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
  clients: Client[];
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
