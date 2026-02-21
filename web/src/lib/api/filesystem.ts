/**
 * 文件系统 API 客户端
 */

import { apiClient } from './client';

/**
 * 目录项
 */
export interface DirectoryItem {
  name: string;
  path: string;
  size?: number; // 文件大小(字节),目录为空
}

/**
 * 目录列表响应
 */
export interface DirectoryListResponse {
  currentPath: string;
  parentPath: string;
  folders: DirectoryItem[];
  files: DirectoryItem[];
  roots?: DirectoryItem[];
  error?: string;
}

/**
 * 路径验证结果
 */
export interface PathValidationResult {
  success: boolean;
  valid: boolean;
  exists?: boolean;
  isDirectory?: boolean;
  isReadable?: boolean;
  normalizedPath?: string;
  error?: string;
}

/**
 * 文件系统 API
 */
export const filesystemApi = {
  /**
   * 列出目录内容
   * @param path 目录路径,空值表示根目录
   */
  listDirectory: (path?: string): Promise<{ success: boolean; data: DirectoryListResponse }> =>
    apiClient.get<{ success: boolean; data: DirectoryListResponse }>('/system/filesystem', path ? { path } : undefined),

  /**
   * 验证路径
   * @param path 要验证的路径
   */
  validatePath: (path: string): Promise<PathValidationResult> =>
    apiClient.post<PathValidationResult>('/system/filesystem/validate', { path }),
};
