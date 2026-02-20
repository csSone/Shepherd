import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api/client';
import type {
  DistributedNode,
  NodeListResponse,
  NodeRegisterRequest,
  NodeRegisterResponse,
  SendCommandRequest,
  SendCommandResponse,
} from '@/types/node';

/**
 * 获取所有节点列表
 */
export function useNodes() {
  return useQuery({
    queryKey: ['nodes'],
    queryFn: async () => {
      const response = await apiClient.get<NodeListResponse>('/master/nodes');
      return response.nodes;
    },
    refetchInterval: 5000,
  });
}

/**
 * 获取单个节点详情
 */
export function useNode(nodeId: string) {
  return useQuery({
    queryKey: ['nodes', nodeId],
    queryFn: async () => {
      const response = await apiClient.get<{ node: DistributedNode }>(`/master/nodes/${nodeId}`);
      return response.node;
    },
    enabled: !!nodeId,
  });
}

/**
 * 注册新节点
 */
export function useRegisterNode() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (nodeInfo: NodeRegisterRequest) => {
      const response = await apiClient.post<NodeRegisterResponse>('/master/nodes/register', nodeInfo);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
    },
  });
}

/**
 * 注销节点
 */
export function useUnregisterNode() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (nodeId: string) => {
      const response = await apiClient.delete<{ success: boolean }>(`/master/nodes/${nodeId}`);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
    },
  });
}

/**
 * 向节点发送命令
 */
export function useSendCommand() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ nodeId, command }: { nodeId: string; command: SendCommandRequest }) => {
      const response = await apiClient.post<SendCommandResponse>(`/master/nodes/${nodeId}/command`, command);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
    },
  });
}

/**
 * 获取在线节点数
 */
export function useOnlineNodeCount() {
  return useQuery({
    queryKey: ['nodes', 'count', 'online'],
    queryFn: async () => {
      const response = await apiClient.get<NodeListResponse>('/master/nodes');
      return response.online;
    },
    refetchInterval: 5000,
  });
}

/**
 * 按角色筛选节点
 */
export function useNodesByRole(role: string) {
  return useQuery({
    queryKey: ['nodes', 'role', role],
    queryFn: async () => {
      const response = await apiClient.get<NodeListResponse>('/master/nodes');
      return response.nodes.filter((node) => node.role === role);
    },
    enabled: !!role,
  });
}

/**
 * 按状态筛选节点
 */
export function useNodesByStatus(status: string) {
  return useQuery({
    queryKey: ['nodes', 'status', status],
    queryFn: async () => {
      const response = await apiClient.get<NodeListResponse>('/master/nodes');
      return response.nodes.filter((node) => node.status === status);
    },
    enabled: !!status,
  });
}
