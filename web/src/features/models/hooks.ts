import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import type {
  Model,
  ModelListResponse,
  LoadModelParams,
  ModelStatus,
  ModelCapabilities,
} from '@/types';

/**
 * 模型列表 Hook
 */
export function useModels() {
  return useQuery<Model[]>({
    queryKey: ['models'],
    queryFn: async (): Promise<Model[]> => {
      const response = await apiClient.get('/models') as ModelListResponse;
      // 调试日志：检查返回的数据
      console.log('[useModels] 获取到模型列表:', response.models.length, '个模型');
      const qwenModel = response.models.find((m: Model) => m.name.includes('Qwen3.5-397B'));
      if (qwenModel) {
        console.log('[useModels] Qwen3.5-397B 模型数据:', {
          name: qwenModel.name,
          size: qwenModel.size,
          totalSize: qwenModel.totalSize,
          shardCount: qwenModel.shardCount,
          mmprojPath: qwenModel.mmprojPath,
        });
      }
      return response.models;
    },
    staleTime: 5 * 1000, // 5 秒后数据视为过期，会更频繁刷新
    refetchOnWindowFocus: true, // 窗口获得焦点时刷新
  });
}

/**
 * 单个模型 Hook
 */
export function useModel(modelId: string) {
  return useQuery({
    queryKey: ['models', modelId],
    queryFn: async () => {
      const response = await apiClient.get<{ model: Model }>(`/models/${modelId}`);
      return response.model;
    },
    enabled: !!modelId,
  });
}

/**
 * 加载模型 Hook
 */
export function useLoadModel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params: LoadModelParams) => {
      const response = await apiClient.post<{ success: boolean }>(
        `/models/${params.modelId}/load`,
        params
      );
      return response;
    },
    onSuccess: () => {
      // 使模型列表查询失效
      queryClient.invalidateQueries({ queryKey: ['models'] });
    },
  });
}

/**
 * 卸载模型 Hook
 */
export function useUnloadModel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (modelId: string) => {
      const response = await apiClient.post<{ success: boolean }>(
        `/models/${modelId}/unload`
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['models'] });
    },
  });
}

/**
 * 更新模型别名 Hook
 */
export function useUpdateModelAlias() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ modelId, alias }: { modelId: string; alias: string }) => {
      const response = await apiClient.put<{ success: boolean }>(
        `/models/${modelId}/alias`,
        { alias }
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['models'] });
    },
  });
}

/**
 * 设置模型收藏 Hook
 */
export function useSetModelFavourite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ modelId, favourite }: { modelId: string; favourite: boolean }) => {
      const response = await apiClient.put<{ success: boolean }>(
        `/models/${modelId}/favourite`,
        { favourite }
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['models'] });
    },
  });
}

/**
 * 扫描模型 Hook
 */
export function useScanModels() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.post<{
        message: string;
        models_found: number;
        errors: number;
        duration_ms: number;
        models: Model[];
        scan_errors: Array<{ path: string; error: string }>;
      }>('/model/scan');
      return response;
    },
    onSuccess: (data) => {
      // 扫描完成后强制刷新模型列表，清除缓存
      queryClient.invalidateQueries({ queryKey: ['models'], refetchType: 'all' });
      // 显示扫描结果
      console.log(`[useScanModels] 扫描完成: 找到 ${data.models_found} 个模型`);
      // 检查 Qwen3.5-397B 模型数据
      const qwenModel = data.models?.find((m: Model) => m.name.includes('Qwen3.5-397B'));
      if (qwenModel) {
        console.log('[useScanModels] Qwen3.5-397B 扫描结果:', {
          name: qwenModel.name,
          totalSize: qwenModel.totalSize,
          shardCount: qwenModel.shardCount,
          mmprojPath: qwenModel.mmprojPath,
        });
      }
    },
  });
}

/**
 * 扫描状态 Hook
 */
export function useScanStatus() {
  return useQuery({
    queryKey: ['scan', 'status'],
    queryFn: async () => {
      const response = await apiClient.get<{
        scanning: boolean;
        progress?: number;
        currentPath?: string;
      }>('/model/scan/status');
      return response;
    },
    refetchInterval: (query) => {
      // 如果正在扫描，每秒刷新一次；否则不刷新
      const data = query.state.data;
      return data?.scanning ? 1000 : false;
    },
  });
}

/**
 * 过滤模型 Hook（包含排序）
 * 使用 useMemo 确保排序稳定且只依赖项变化时重新计算
 */
export function useFilteredModels(
  models: Model[] | undefined,
  filters: {
    search?: string;
    status?: ModelStatus;
    favourite?: boolean;
  }
): Model[] {
  if (!models) return [];

  // 过滤模型
  const filtered = models.filter((model) => {
    // 搜索过滤
    if (filters.search) {
      const search = filters.search.toLowerCase();
      const matchName = model.name.toLowerCase().includes(search);
      const matchAlias = model.alias?.toLowerCase().includes(search);
      const matchArch = model.metadata.architecture?.toLowerCase().includes(search);
      if (!matchName && !matchAlias && !matchArch) return false;
    }

    // 状态过滤
    if (filters.status && model.status !== filters.status) return false;

    // 收藏过滤
    if (filters.favourite && !model.favourite) return false;

    return true;
  });

  // 排序模型：稳定的排序，确保每次刷新后顺序一致
  // 排序优先级：名称（字母）> 扫描时间 > 路径
  return [...filtered].sort((a: Model, b: Model) => {
    // 优先按显示名称（别名或模型名）排序
    const aName = (a.alias || a.displayName || a.name).toLowerCase();
    const bName = (b.alias || b.displayName || b.name).toLowerCase();

    const nameCompare = aName.localeCompare(bName, 'zh-CN');
    if (nameCompare !== 0) return nameCompare;

    // 名称相同时，按扫描时间降序排序（最新的在前）
    const aTime = new Date(a.scannedAt).getTime();
    const bTime = new Date(b.scannedAt).getTime();
    if (aTime !== bTime) return bTime - aTime;

    // 扫描时间也相同时，按路径排序
    return a.path.localeCompare(b.path);
  });
}

/**
 * 系统端点 GPU 信息类型（Shepherd 扩展格式）
 * 用于 /system/gpus 端点返回的 GPU 信息
 */
export interface SystemGPUInfo {
  id: string;          // 设备 ID，如 "ROCm0"
  name: string;        // GPU 名称
  totalMemory?: string; // 总内存，如 "122880 MiB"
  freeMemory?: string;  // 可用内存，如 "115050 MiB"
  architecture?: string; // 架构信息
  available: boolean;  // 是否可用
}

/**
 * 系统 GPU 列表响应
 */
export interface SystemGPUListResponse {
  gpus: SystemGPUInfo[];      // 详细 GPU 信息（Shepherd 扩展）
  devices: string[];    // 简单设备字符串列表（兼容 LlamacppServer 格式）
  count: number;
}

/**
 * 获取系统 GPU 列表 Hook
 */
export function useGPUs() {
  return useQuery<SystemGPUListResponse>({
    queryKey: ['system', 'gpus'],
    queryFn: async () => {
      const response = await apiClient.get<SystemGPUListResponse>('/system/gpus');
      return response;
    },
    staleTime: 60 * 1000, // GPU 信息缓存 1 分钟
    refetchOnWindowFocus: false,
  });
}

/**
 * llama.cpp 后端信息类型
 */
export interface LlamacppBackend {
  path: string;
  name: string;
  description: string;
  available: boolean;
}

export interface LlamacppBackendListResponse {
  backends: LlamacppBackend[];
  count: number;
}

/**
 * 获取可用的 llama.cpp 后端列表 Hook
 */
export function useLlamacppBackends() {
  return useQuery({
    queryKey: ['system', 'llamacpp-backends'],
    queryFn: async () => {
      const response = await apiClient.get<LlamacppBackendListResponse>('/system/llamacpp-backends');
      return response.backends;
    },
    staleTime: 60 * 1000, // 后端列表缓存 1 分钟
    refetchOnWindowFocus: false,
  });
}

/**
 * 模型能力配置 Hook
 */
export function useModelCapabilities(modelId: string) {
  return useQuery({
    queryKey: ['models', 'capabilities', modelId],
    queryFn: async () => {
      const response = await apiClient.get<{ capabilities: ModelCapabilities }>('/models/capabilities/get', { modelId });
      return response.capabilities;
    },
    enabled: !!modelId,
    staleTime: 10 * 60 * 1000, // 能力配置缓存 10 分钟
    refetchOnWindowFocus: false,
  });
}

/**
 * 设置模型能力 Hook
 */
export function useSetModelCapabilities() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ modelId, capabilities }: {
      modelId: string;
      capabilities: ModelCapabilities;
    }) => {
      const response = await apiClient.post<{ success: boolean; message?: string }>(
        '/models/capabilities/set',
        { modelId, capabilities }
      );
      return response;
    },
    onSuccess: (data, variables) => {
      // 使能力查询失效
      queryClient.invalidateQueries({
        queryKey: ['models', 'capabilities', variables.modelId]
      });
    },
  });
}

/**
 * 显存估算请求参数
 */
export interface EstimateVRAMParams {
  modelId: string;
  llamaBinPath: string;
  ctxSize?: number;
  batchSize?: number;
  uBatchSize?: number;
  parallel?: number;
  flashAttention?: boolean;
  kvUnified?: boolean;
  cacheTypeK?: string;
  cacheTypeV?: string;
  extraParams?: string;
}

/**
 * 显存估算响应
 */
export interface EstimateVRAMResponse {
  success: boolean;
  data?: {
    vram?: string;      // "60565"
    vramMB?: number;    // 60565
    vramGB?: string;    // "59.15"
    error?: string;
    details?: string;
  };
  error?: string;
  details?: string;
}

/**
 * 估算显存 Hook
 */
export function useEstimateVRAM() {
  return useMutation({
    mutationFn: async (params: EstimateVRAMParams) => {
      const response = await apiClient.post<EstimateVRAMResponse>(
        '/models/vram/estimate',
        params
      );
      return response;
    },
  });
}
