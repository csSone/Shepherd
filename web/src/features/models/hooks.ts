import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import type {
  Model,
  ModelListResponse,
  LoadModelParams,
  ModelStatus,
} from '@/types';

/**
 * 模型列表 Hook
 */
export function useModels() {
  return useQuery({
    queryKey: ['models'],
    queryFn: async () => {
      const response = await apiClient.get<ModelListResponse>('/models');
      return response.models;
    },
    staleTime: 30 * 1000, // 30 秒
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
      const response = await apiClient.post<{ success: boolean }>('/scan');
      return response;
    },
    onSuccess: () => {
      // 扫描完成后刷新模型列表
      queryClient.invalidateQueries({ queryKey: ['models'] });
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
      }>('/scan/status');
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
 * 过滤模型 Hook
 */
export function useFilteredModels(
  models: Model[] | undefined,
  filters: {
    search?: string;
    status?: ModelStatus;
    favourite?: boolean;
  }
) {
  if (!models) return [];

  return models.filter((model) => {
    // 搜索过滤
    if (filters.search) {
      const search = filters.search.toLowerCase();
      const matchName = model.name.toLowerCase().includes(search);
      const matchAlias = model.alias?.toLowerCase().includes(search);
      const matchArch = model.metadata.architecture.toLowerCase().includes(search);
      if (!matchName && !matchAlias && !matchArch) return false;
    }

    // 状态过滤
    if (filters.status && model.status !== filters.status) return false;

    // 收藏过滤
    if (filters.favourite && !model.favourite) return false;

    return true;
  });
}
