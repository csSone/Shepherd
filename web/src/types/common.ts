/**
 * Unified type definitions for Shepherd frontend
 * Shepherd 前端统一类型定义
 */

// Unified node state - matches backend types.NodeState
// 统一的节点状态 - 与后端 types.NodeState 匹配
export type NodeState =
  | 'offline'
  | 'online'
  | 'busy'
  | 'error'
  | 'degraded'
  | 'disabled';

// Unified error codes - matches backend types.ErrorCode
// 统一的错误码 - 与后端 types.ErrorCode 匹配
export type ErrorCode =
  | 'NODE_NOT_FOUND'
  | 'INVALID_REQUEST'
  | 'TIMEOUT'
  | 'COMMAND_FAILED'
  | 'NOT_AUTHENTICATED'
  | 'PERMISSION_DENIED'
  | 'RESOURCE_EXHAUSTED'
  | 'INTERNAL_ERROR';

// Error information - matches backend types.ErrorInfo
// 错误信息 - 与后端 types.ErrorInfo 匹配
export interface ErrorInfo {
  code: ErrorCode;
  message: string;
  details?: string;
}

// Response metadata - matches backend types.ResponseMeta
// 响应元数据 - 与后端 types.ResponseMeta 匹配
export interface ResponseMeta {
  timestamp: string;
  requestId: string;
  latency?: number; // milliseconds
}

// Unified API response format - matches backend types.ApiResponse
// 统一的 API 响应格式 - 与后端 types.ApiResponse 匹配
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: ErrorInfo;
  metadata?: ResponseMeta;
}

// Paginated response - matches backend types.PaginatedResponse
// 分页响应 - 与后端 types.PaginatedResponse 匹配
export interface PaginatedResponse<T = unknown> {
  success: boolean;
  data?: T[];
  total: number;
  page: number;
  pageSize: number;
  error?: ErrorInfo;
  metadata?: ResponseMeta;
}

// Utility functions for working with API responses

/**
 * Checks if an API response was successful
 */
export function isSuccess<T>(response: ApiResponse<T>): response is ApiResponse<T> & { success: true } {
  return response.success === true;
}

/**
 * Checks if an API response was an error
 */
export function isError<T>(response: ApiResponse<T>): response is ApiResponse<T> & { success: false } {
  return response.success === false;
}

/**
 * Extracts the error message from an API response
 */
export function getErrorMessage(response: ApiResponse): string {
  if (response.error) {
    if (response.error.details) {
      return `[${response.error.code}] ${response.error.message}: ${response.error.details}`;
    }
    return `[${response.error.code}] ${response.error.message}`;
  }
  return 'Unknown error';
}

/**
 * Type guard for checking if a value is a valid NodeState
 */
export function isValidNodeState(value: string): value is NodeState {
  return ['offline', 'online', 'busy', 'error', 'degraded', 'disabled'].includes(value);
}

/**
 * Type guard for checking if a value is a valid ErrorCode
 */
export function isValidErrorCode(value: string): value is ErrorCode {
  return [
    'NODE_NOT_FOUND',
    'INVALID_REQUEST',
    'TIMEOUT',
    'COMMAND_FAILED',
    'NOT_AUTHENTICATED',
    'PERMISSION_DENIED',
    'RESOURCE_EXHAUSTED',
    'INTERNAL_ERROR',
  ].includes(value);
}

// Common error responses
export const ErrorResponses = {
  nodeNotFound: (nodeId: string): ApiResponse => ({
    success: false,
    error: {
      code: 'NODE_NOT_FOUND',
      message: `Node ${nodeId} not found`,
    },
  }),

  invalidRequest: (message: string): ApiResponse => ({
    success: false,
    error: {
      code: 'INVALID_REQUEST',
      message,
    },
  }),

  timeout: (operation: string): ApiResponse => ({
    success: false,
    error: {
      code: 'TIMEOUT',
      message: `Operation ${operation} timed out`,
    },
  }),

  commandFailed: (command: string, details?: string): ApiResponse => ({
    success: false,
    error: {
      code: 'COMMAND_FAILED',
      message: `Command ${command} failed`,
      details,
    },
  }),

  notAuthenticated: (): ApiResponse => ({
    success: false,
    error: {
      code: 'NOT_AUTHENTICATED',
      message: 'Authentication required',
    },
  }),

  permissionDenied: (resource?: string): ApiResponse => ({
    success: false,
    error: {
      code: 'PERMISSION_DENIED',
      message: resource ? `Permission denied for ${resource}` : 'Permission denied',
    },
  }),

  internalError: (details?: string): ApiResponse => ({
    success: false,
    error: {
      code: 'INTERNAL_ERROR',
      message: 'Internal server error',
      details,
    },
  }),
};
