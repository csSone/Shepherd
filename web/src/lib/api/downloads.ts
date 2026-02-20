/**
 * 下载管理 API 客户端
 */

import { apiClient } from './client';
import type { CreateDownloadParams, DownloadTask, DownloadListResponse } from '@/types';

/**
 * 模型文件信息
 */
export interface ModelFileInfo {
  name: string;
  size: number;
  download_url: string;
}

/**
 * 模型文件列表响应
 */
export interface ModelFilesResponse {
  success: boolean;
  data: ModelFileInfo[];
  error?: string;
}

/**
 * 下载管理 API
 */
export const downloadsApi = {
  /**
   * 获取下载任务列表
   */
  list: (): Promise<DownloadListResponse> =>
    apiClient.get<DownloadListResponse>('/downloads'),

  /**
   * 创建下载任务
   */
  create: (params: CreateDownloadParams): Promise<{ success: boolean; message?: string; data?: DownloadTask; error?: string }> =>
    apiClient.post('/downloads', params),

  /**
   * 获取单个下载任务
   */
  get: (id: string): Promise<DownloadTask> =>
    apiClient.get<DownloadTask>(`/downloads/${id}`),

  /**
   * 暂停下载
   */
  pause: (id: string): Promise<{ success: boolean; message?: string; error?: string }> =>
    apiClient.post(`/downloads/${id}/pause`),

  /**
   * 恢复下载
   */
  resume: (id: string): Promise<{ success: boolean; message?: string; error?: string }> =>
    apiClient.post(`/downloads/${id}/resume`),

  /**
   * 取消下载
   */
  cancel: (id: string): Promise<{ success: boolean; message?: string; error?: string }> =>
    apiClient.delete<{ success: boolean; message?: string; error?: string }>(`/downloads/${id}`),

  /**
   * 重试下载
   */
  retry: (id: string): Promise<{ success: boolean; message?: string; error?: string }> =>
    apiClient.post(`/downloads/${id}/retry`),

  /**
   * 清理已完成的下载
   */
  clearCompleted: (): Promise<{ success: boolean; message?: string; error?: string }> =>
    apiClient.delete<{ success: boolean; message?: string; error?: string }>('/downloads/completed'),

  /**
   * 获取模型文件列表
   * 使用查询参数以支持 repoId 中包含斜杠 (如 Qwen/Qwen2-7B-Instruct)
   */
  listModelFiles: (source: 'huggingface' | 'modelscope', repoId: string): Promise<ModelFilesResponse> =>
    apiClient.get<ModelFilesResponse>(`/repo/files?source=${source}&repoId=${encodeURIComponent(repoId)}`),
};
