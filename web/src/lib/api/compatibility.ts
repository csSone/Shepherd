/**
 * API 兼容性配置 API 客户端
 */

import { apiClient } from './client';

/**
 * API 兼容性配置
 */
export interface CompatibilityConfig {
  ollama: {
    enabled: boolean;
    port: number;
  };
  lmstudio: {
    enabled: boolean;
    port: number;
  };
}

/**
 * 兼容性配置响应
 */
interface CompatibilityResponse {
  success: boolean;
  data: CompatibilityConfig;
  error?: string;
}

/**
 * 更新配置响应
 */
interface UpdateResponse {
  success: boolean;
  message?: string;
  error?: string;
  errorType?: 'in_use' | 'permission' | 'invalid' | 'unknown';
  service?: 'ollama' | 'lmstudio';
  autoDisabled?: boolean;
  data?: CompatibilityConfig;
}

/**
 * 连接测试响应
 */
interface TestConnectionResponse {
  success: boolean;
  valid: boolean;
  message?: string;
  error?: string;
}

/**
 * API 兼容性配置管理 API
 */
export const compatibilityApi = {
  /**
   * 获取兼容性配置
   */
  get: (): Promise<CompatibilityResponse> =>
    apiClient.get<CompatibilityResponse>('/config/compatibility'),

  /**
   * 更新兼容性配置
   */
  update: (config: CompatibilityConfig): Promise<UpdateResponse> =>
    apiClient.put<UpdateResponse>('/config/compatibility', config),

  /**
   * 测试端口连接
   */
  testConnection: (port: number, type: 'ollama' | 'lmstudio'): Promise<TestConnectionResponse> =>
    apiClient.post<TestConnectionResponse>('/config/compatibility/test', { port, type }),
};
