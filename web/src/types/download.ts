/**
 * 下载状态
 */
export type DownloadState =
  | 'idle'
  | 'preparing'
  | 'downloading'
  | 'merging'
  | 'verifying'
  | 'completed'
  | 'failed'
  | 'paused';

/**
 * 下载来源
 */
export type DownloadSource = 'huggingface' | 'modelscope';

/**
 * 下载任务
 */
export interface DownloadTask {
  id: string;
  source: DownloadSource;
  repoId: string;
  fileName: string;
  path: string;
  state: DownloadState;
  downloadedBytes: number;
  totalBytes: number;
  partsCompleted: number;
  partsTotal: number;
  progress: number; // 0-1
  speed: number; // bytes per second
  eta: number; // seconds remaining
  error?: string;
  createdAt: string;
  completedAt?: string;
}

/**
 * 创建下载任务参数
 */
export interface CreateDownloadParams {
  source: DownloadSource;
  repoId: string;
  fileName?: string;
  path?: string;
  maxRetries?: number;
  chunkSize?: number;
}

/**
 * 下载列表响应
 */
export interface DownloadListResponse {
  tasks: DownloadTask[];
  total: number;
  active: number;
}
