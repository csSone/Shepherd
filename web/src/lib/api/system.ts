/**
 * 系统 API 客户端
 */

import { apiClient } from './client';

/**
 * 服务器信息响应
 */
export interface ServerInfoResponse {
  success: boolean;
  data?: {
    version: string;
    buildTime: string;
    gitCommit: string;
    name: string;
    status: string;
    mode: string;
    ports: {
      web: number;
      anthropic: number;
      ollama: number;
      lmstudio: number;
    };
  };
  error?: string;
}

/**
 * 系统 API
 */
export const systemApi = {
  /**
   * 获取服务器信息（版本、构建时间等）
   */
  getInfo: () => apiClient.get<ServerInfoResponse>('/info'),
};
