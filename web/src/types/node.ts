/**
 * 节点角色
 */
export type NodeRole = 'standalone' | 'master' | 'client' | 'hybrid';

/**
 * 节点状态 - 统一状态定义，与后端 types.NodeState 保持一致
 */
export type NodeStatus = 'offline' | 'online' | 'busy' | 'error' | 'degraded' | 'disabled';

/**
 * GPU 信息
 */
export interface GPUInfo {
  index: number;
  name: string;
  vendor: string;
  totalMemory: number;
  usedMemory: number;
  temperature: number;
  utilization: number;
  driverVersion?: string;
}

/**
 * llama.cpp 信息
 */
export interface LlamacppInfo {
  path: string;
  version: string;
  buildType: string;
  gpuBackend?: string;
  supportsGPU: boolean;
  available: boolean;
}

/**
 * 模型信息
 */
export interface NodeModelInfo {
  path: string;
  name: string;
  size: number;
  format: string;
  loaded: boolean;
  loadedAt?: string;
  loadedBy?: string;
}

/**
 * 节点资源
 */
export interface NodeResources {
  cpuUsed: number;
  cpuTotal: number;
  memoryUsed: number;
  memoryTotal: number;
  diskUsed: number;
  diskTotal: number;
  gpuInfo: GPUInfo[];
  networkRx: number;
  networkTx: number;
  uptime: number;
  loadAverage: number[];
}

/**
 * 节点能力
 */
export interface NodeCapabilities {
  gpu: boolean;
  gpuCount: number;
  gpuNames: string[];
  cpuCount: number;
  memory: number;
  supportsLlama: boolean;
  supportsPython: boolean;
  condaEnvs: string[];
  dockerEnabled: boolean;
}

/**
 * 分布式节点 - 匹配后端 internal/node/types.go NodeInfo
 */
export interface DistributedNode {
  id: string;
  name: string;
  address: string;
  port: number;
  role: NodeRole;
  status: NodeStatus;
  version: string;
  tags: string[];
  capabilities?: NodeCapabilities;
  resources?: NodeResources;
  metadata: Record<string, string>;
  createdAt: string;
  updatedAt: string;
  lastSeen: string;
}

/**
 * 心跳消息
 */
export interface HeartbeatMessage {
  nodeId: string;
  timestamp: string;
  status: NodeStatus;
  role: NodeRole;
  resources?: NodeResources;
  capabilities?: NodeCapabilities;
  metadata?: Record<string, unknown>;
  sequence: number;
}

/**
 * 命令类型
 */
export type CommandType = 
  | 'load_model' 
  | 'unload_model' 
  | 'run_llamacpp' 
  | 'stop_process' 
  | 'update_config' 
  | 'collect_logs' 
  | 'scan_models';

/**
 * 命令
 */
export interface NodeCommand {
  id: string;
  type: CommandType;
  fromNodeId: string;
  toNodeId?: string;
  payload: Record<string, unknown>;
  createdAt: string;
  timeout?: number;
  priority: number;
  retryCount: number;
  maxRetries: number;
}

/**
 * 命令结果
 */
export interface CommandResult {
  commandId: string;
  fromNodeId: string;
  toNodeId: string;
  success: boolean;
  result?: Record<string, unknown>;
  error?: string;
  completedAt: string;
  duration: number;
  metadata?: Record<string, string>;
}

/**
 * 节点统计信息
 */
export interface NodeStats {
  total: number;
  online: number;
  offline: number;
  busy: number;
}

/**
 * 节点列表响应 - 匹配后端 GET /api/master/nodes 响应格式
 */
export interface NodeListResponse {
  nodes: DistributedNode[];
  stats: NodeStats;
}

/**
 * 节点注册请求
 */
export interface NodeRegisterRequest {
  id: string;
  name: string;
  address: string;
  port: number;
  role: NodeRole;
  capabilities: NodeCapabilities;
  metadata?: Record<string, string>;
}

/**
 * 节点注册响应
 */
export interface NodeRegisterResponse {
  message: string;
  node: DistributedNode;
}

/**
 * 发送命令请求
 */
export interface SendCommandRequest {
  type: CommandType;
  payload: Record<string, unknown>;
  timeout?: number;
  priority?: number;
}

/**
 * 发送命令响应
 */
export interface SendCommandResponse {
  message: string;
  command: NodeCommand;
}
