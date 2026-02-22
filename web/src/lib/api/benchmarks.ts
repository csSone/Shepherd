import { apiClient } from './client';
import type {
  Benchmark,
  BenchmarkConfig,
  BenchmarkResult,
  BenchmarkResultFile,
  BenchmarkListDataResponse,
  BenchmarkListResponse,
  BenchmarkResultResponse,
  CreateBenchmarkRequest,
  CreateBenchmarkResponse,
  SaveBenchmarkConfigRequest,
  SaveBenchmarkConfigResponse,
  LoadBenchmarkConfigResponse,
  BenchmarkParamsResponse,
  LlamaCppVersion,
  ComputeDevice,
} from '@/types';

/**
 * 压测 API
 */
export const benchmarksApi = {
  /**
   * 获取压测参数列表
   */
  async getParams(): Promise<BenchmarkParamsResponse> {
    return apiClient.get<BenchmarkParamsResponse>('/models/param/benchmark/list');
  },

  /**
   * 获取可用计算设备
   */
  async getDevices(llamaBinPath: string): Promise<{ success: boolean; data?: { devices: string[] }; error?: string }> {
    return apiClient.get<{ success: boolean; data?: { devices: string[] }; error?: string }>(
      `/model/device/list?llamaBinPath=${encodeURIComponent(llamaBinPath)}`
    );
  },

  /**
   * 获取 Llama.cpp 版本列表
   */
  async getLlamaCppVersions(): Promise<{ success: boolean; data?: { items: Array<{ path: string; name?: string; description?: string }> } }> {
    return apiClient.get('/llamacpp/list');
  },

  /**
   * 创建压测任务
   */
  async create(params: CreateBenchmarkRequest): Promise<CreateBenchmarkResponse> {
    return apiClient.post<CreateBenchmarkResponse>('/models/benchmark', params);
  },

  /**
   * 获取压测结果列表
   */
  async listResults(modelId: string): Promise<BenchmarkListResponse> {
    return apiClient.get<BenchmarkListResponse>(`/models/benchmark/list?modelId=${encodeURIComponent(modelId)}`);
  },

  /**
   * 获取压测结果详情
   */
  async getResult(fileName: string): Promise<BenchmarkResultResponse> {
    return apiClient.get<BenchmarkResultResponse>(`/models/benchmark/get?fileName=${encodeURIComponent(fileName)}`);
  },

  /**
   * 删除压测结果
   */
  async deleteResult(fileName: string): Promise<{ success: boolean; error?: string }> {
    return apiClient.post<{ success: boolean; error?: string }>(`/models/benchmark/delete?fileName=${encodeURIComponent(fileName)}`);
  },

  /**
   * 获取压测任务列表
   */
  async list(): Promise<BenchmarkListDataResponse> {
    return apiClient.get<BenchmarkListDataResponse>('/models/benchmark/tasks');
  },

  /**
   * 获取单个压测任务
   */
  async get(benchmarkId: string): Promise<{ success: boolean; data?: Benchmark; error?: string }> {
    return apiClient.get<{ success: boolean; data?: Benchmark; error?: string }>(`/models/benchmark/tasks/${benchmarkId}`);
  },

  /**
   * 取消压测任务
   */
  async cancel(benchmarkId: string): Promise<{ success: boolean; error?: string }> {
    return apiClient.post<{ success: boolean; error?: string }>(`/models/benchmark/tasks/${benchmarkId}/cancel`);
  },

  /**
   * 保存压测配置
   */
  async saveConfig(params: SaveBenchmarkConfigRequest): Promise<SaveBenchmarkConfigResponse> {
    return apiClient.post<SaveBenchmarkConfigResponse>('/models/benchmark/configs', params);
  },

  /**
   * 获取压测配置列表
   */
  async listConfigs(): Promise<LoadBenchmarkConfigResponse> {
    return apiClient.get<LoadBenchmarkConfigResponse>('/models/benchmark/configs');
  },

  /**
   * 获取单个压测配置
   */
  async getConfig(name: string): Promise<{ success: boolean; data?: BenchmarkConfig; error?: string }> {
    return apiClient.get<{ success: boolean; data?: BenchmarkConfig; error?: string }>(
      `/models/benchmark/configs/${encodeURIComponent(name)}`
    );
  },

  /**
   * 删除压测配置
   */
  async deleteConfig(name: string): Promise<{ success: boolean; error?: string }> {
    return apiClient.delete<{ success: boolean; error?: string }>(
      `/models/benchmark/configs/${encodeURIComponent(name)}`
    );
  },
};
