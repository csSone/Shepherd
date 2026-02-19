/**
 * 模型元数据
 */
export interface ModelMetadata {
  architecture: string;
  quantization?: string;
  contextLength?: number;
  embeddingLength?: number;
  blockSize?: number;
  feedForwardLength?: number;
  attentionHeadCount?: number;
  attentionHeadCountKeyValue?: number;
  ropeDimensionCount?: number;
  layerCount?: number;
  ggmlFileType?: string;
  tokenizer?: string;
}

/**
 * 模型信息
 */
export interface Model {
  id: string;
  name: string;
  alias?: string;
  path: string;
  size: number;
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
}

/**
 * 模型列表响应
 */
export interface ModelListResponse {
  models: Model[];
  total: number;
  loaded: number;
}
