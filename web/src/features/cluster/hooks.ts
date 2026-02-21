import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import { useConfig } from '@/lib/config';
import type {
  Client,
  ClusterTask,
  ClientListResponse,
  TaskListResponse,
  ClusterOverview,
  ScanStatus,
  ClientStatus,
  ScheduleStrategy,
} from '@/types';

/**
 * 获取当前运行模式
 * 集群相关功能仅在 Master 模式下可用
 */
function useClusterMode() {
  const { data: config } = useConfig();
  return config?.server?.mode || 'standalone';
}

/**
 * 集群概览 Hook
 */
export function useClusterOverview() {
  const mode = useClusterMode();

  return useQuery({
    queryKey: ['cluster', 'overview'],
    queryFn: async () => {
      const response = await apiClient.get<ClusterOverview>('/master/overview');
      return response;
    },
    staleTime: 10 * 1000, // 10 秒
    refetchInterval: 5000, // 每 5 秒刷新
    enabled: mode === 'master', // 仅在 Master 模式下启用
  });
}

/**
 * 客户端列表 Hook
 */
export function useClients() {
  const mode = useClusterMode();

  return useQuery({
    queryKey: ['cluster', 'clients'],
    queryFn: async () => {
      const response = await apiClient.get<ClientListResponse>('/master/clients');
      return response.clients;
    },
    staleTime: 10 * 1000,
    refetchInterval: 5000,
    enabled: mode === 'master', // 仅在 Master 模式下启用
  });
}

/**
 * 单个客户端 Hook
 */
export function useClient(clientId: string) {
  return useQuery({
    queryKey: ['cluster', 'clients', clientId],
    queryFn: async () => {
      const response = await apiClient.get<{ client: Client }>(`/master/clients/${clientId}`);
      return response.client;
    },
    enabled: !!clientId,
    refetchInterval: 3000,
  });
}

/**
 * 断开客户端 Hook
 */
export function useDisconnectClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (clientId: string) => {
      const response = await apiClient.delete<{ success: boolean }>(`/master/clients/${clientId}`);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cluster'] });
    },
  });
}

/**
 * 任务列表 Hook
 */
export function useClusterTasks() {
  const mode = useClusterMode();

  return useQuery({
    queryKey: ['cluster', 'tasks'],
    queryFn: async () => {
      const response = await apiClient.get<TaskListResponse>('/master/tasks');
      return response.tasks;
    },
    staleTime: 5 * 1000,
    refetchInterval: 2000,
    enabled: mode === 'master', // 仅在 Master 模式下启用
  });
}

/**
 * 创建任务 Hook
 */
export function useCreateClusterTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (task: {
      type: string;
      payload: Record<string, unknown>;
      assignTo?: string;
    }) => {
      const response = await apiClient.post<{ task: ClusterTask }>('/master/tasks', task);
      return response.task;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cluster', 'tasks'] });
    },
  });
}

/**
 * 取消任务 Hook
 */
export function useCancelClusterTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (taskId: string) => {
      const response = await apiClient.delete<{ success: boolean }>(`/master/tasks/${taskId}`);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cluster', 'tasks'] });
    },
  });
}

/**
 * 重试任务 Hook
 */
export function useRetryClusterTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (taskId: string) => {
      const response = await apiClient.post<{ success: boolean }>(`/master/tasks/${taskId}/retry`);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cluster', 'tasks'] });
    },
  });
}

/**
 * 扫描网络 Hook
 */
export function useNetworkScan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params?: {
      cidr?: string;
      portRange?: string;
      timeout?: number;
    }) => {
      const response = await apiClient.post<{ success: boolean }>('/master/scan', params || {});
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cluster', 'scan'] });
    },
  });
}

/**
 * 扫描状态 Hook
 */
export function useScanStatus() {
  return useQuery({
    queryKey: ['cluster', 'scan'],
    queryFn: async () => {
      const response = await apiClient.get<ScanStatus>('/master/scan/status');
      return response;
    },
    refetchInterval: (query) => {
      const data = query.state.data;
      return data?.running ? 1000 : false;
    },
  });
}

/**
 * 设置调度策略 Hook
 */
export function useSetScheduleStrategy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (strategy: ScheduleStrategy) => {
      const response = await apiClient.put<{ success: boolean }>('/master/schedule/strategy', {
        strategy,
      });
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cluster'] });
    },
  });
}

/**
 * 过滤客户端 Hook
 */
export function useFilteredClients(
  clients: Client[] | undefined,
  filters: {
    search?: string;
    status?: ClientStatus;
    hasTag?: string;
  }
) {
  if (!clients) return [];

  return clients.filter((client) => {
    // 搜索过滤
    if (filters.search) {
      const search = filters.search.toLowerCase();
      const matchName = client.name.toLowerCase().includes(search);
      const matchAddress = client.address.toLowerCase().includes(search);
      if (!matchName && !matchAddress) return false;
    }

    // 状态过滤
    if (filters.status && client.status !== filters.status) return false;

    // 标签过滤
    if (filters.hasTag && !client.tags.includes(filters.hasTag)) return false;

    return true;
  });
}

/**
 * 过滤任务 Hook
 */
export function useFilteredTasks(
  tasks: ClusterTask[] | undefined,
  filters: {
    search?: string;
    status?: string;
    type?: string;
    assignedTo?: string;
  }
) {
  if (!tasks) return [];

  return tasks.filter((task) => {
    // 搜索过滤
    if (filters.search) {
      const search = filters.search.toLowerCase();
      const matchId = task.id.toLowerCase().includes(search);
      if (!matchId) return false;
    }

    // 状态过滤
    if (filters.status && task.status !== filters.status) return false;

    // 类型过滤
    if (filters.type && task.type !== filters.type) return false;

    // 分配过滤
    if (filters.assignedTo && task.assignedTo !== filters.assignedTo) return false;

    return true;
  });
}
