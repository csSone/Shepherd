/**
 * 模型元数据 - 完全匹配后端 server.go 返回的字段
 */
export interface ModelMetadata {
  // 基本信息
  name?: string;
  architecture: string;
  quantization?: string;
  type?: string; // "model", "adapter", "projector", "imatrix"
  author?: string | null;
  url?: string | null;
  description?: string | null;
  license?: string | null;

  // 文件类型信息
  fileType?: number;
  fileTypeDescriptor?: string; // 更详细的文件类型描述（如 Q4_K_M, Q5_0_L 等）
  quantizationVersion?: number;

  // 模型参数
  parameters?: number;
  bitsPerWeight?: number;

  // 文件信息
  alignment?: number;
  fileSize?: number;
  modelSize?: number;

  // 模型架构参数（可能为 0）
  contextLength?: number;
  embeddingLength?: number;
  layerCount?: number;
  headCount?: number;

  // 以下字段保留以兼容旧代码（后端不再返回）
  blockSize?: number;
  feedForwardLength?: number;
  attentionHeadCount?: number;
  attentionHeadCountKeyValue?: number;
  ropeDimensionCount?: number;
  ggmlFileType?: string;
  tokenizer?: string;
}

/**
 * 模型信息
 */
export interface Model {
  id: string;
  name: string;
  displayName: string;
  alias?: string;
  path: string;
  pathPrefix: string;
  size: number;
  // 分卷模型相关字段
  totalSize?: number;    // 所有分卷的总大小
  shardCount?: number;   // 分卷数量
  shardFiles?: string[]; // 所有分卷文件路径
  favourite: boolean;
  isLoaded: boolean;
  isLoading: boolean;
  isMultimodal: boolean;
  status: ModelStatus;
  port?: number;
  slots?: Slot[];
  metadata: ModelMetadata;
  tags?: string[];
  mmprojPath?: string;
  scannedAt: string;
}

/**
 * 模型状态
 */
export type ModelStatus = 'stopped' | 'loading' | 'loaded' | 'running' | 'unloading' | 'error';

/**
 * 处理槽位
 */
export interface Slot {
  id: number;
  isProcessing: boolean;
  isSpeculative?: boolean;
  taskId?: string;
}

/**
 * 加载模型参数
 */
export interface LoadModelParams {
  // 基础参数
  modelId: string;
  ctxSize?: number;
  batchSize?: number;
  threads?: number;
  gpuLayers?: number;
  temperature?: number;
  topP?: number;
  topK?: number;
  repeatPenalty?: number;
  seed?: number;
  nPredict?: number;

  // 后端配置
  llamaCppPath?: string;      // llama.cpp 可执行文件路径
  mainGpu?: number | string;  // 主GPU选择

  // 能力开关
  capabilities?: {
    thinking?: boolean;    // 思考能力
    tools?: boolean;       // 工具使用
    translation?: boolean; // 直译
    embedding?: boolean;   // 嵌入
  };

  // 上下文与加速
  flashAttention?: boolean;       // Flash Attention 加速
  noMmap?: boolean;               // 禁用内存映射
  lockMemory?: boolean;           // 锁定物理内存

  // 采样参数
  logitsAll?: boolean;            // 输入向量模式
  reranking?: boolean;            // 重排序模式
  minP?: number;                  // Min-P 采样

  // 惩罚参数
  presencePenalty?: number;       // 存在惩罚
  frequencyPenalty?: number;      // 频率惩罚

  // 批处理参数
  uBatchSize?: number;            // 微批大小
  parallelSlots?: number;         // 并发槽位数

  // KV缓存
  kvCacheSize?: number;           // KV缓存内存上限
  kvCacheUnified?: boolean;       // 统一KV缓存区
  kvCacheTypeK?: string;          // KV缓存类型K (f16, f32, q8_0)
  kvCacheTypeV?: string;          // KV缓存类型V

  // 其他参数
  directIo?: string;              // DirectIO 模式
  disableJinja?: boolean;         // 禁用 Jinja 模板
  chatTemplate?: string;          // 内置聊天模板
  contextShift?: boolean;         // 上下文移位
  extraArgs?: string;             // 额外命令行参数
}

/**
 * 模型列表响应
 */
export interface ModelListResponse {
  models: Model[];
  total: number;
  loaded: number;
}

/**
 * 模型能力配置
 */
export interface ModelCapabilities {
  thinking?: boolean;    // 思考能力（如 DeepSeek-R1 等）
  tools?: boolean;       // 工具使用/函数调用
  rerank?: boolean;      // 重排序能力
  embedding?: boolean;   // 嵌入向量生成
}

/**
 * 模型能力响应
 */
export interface ModelCapabilitiesResponse {
  modelId: string;
  capabilities: ModelCapabilities;
  success?: boolean;
  error?: string;
}
