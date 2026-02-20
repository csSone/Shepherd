/**
 * API 错误类
 */
export class ApiError extends Error {
  status: number;
  data: unknown;

  constructor(
    status: number,
    data: unknown,
    message?: string
  ) {
    super(message || `API Error: ${status}`);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
  }
}

/**
 * API 客户端类
 * 支持动态配置后端 URL，前端完全独立运行
 */
class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = '/api') {
    this.baseUrl = baseUrl;
  }

  /**
   * 设置后端 URL
   */
  setBaseUrl(url: string): void {
    this.baseUrl = url;
  }

  /**
   * 获取当前后端 URL
   */
  getBaseUrl(): string {
    return this.baseUrl;
  }

  /**
   * 构建完整 URL
   */
  private buildUrl(endpoint: string, params?: Record<string, string>): string {
    // 如果 endpoint 是相对路径，添加 baseUrl
    let url = endpoint;
    if (!url.startsWith('http://') && !url.startsWith('https://')) {
      url = this.baseUrl + endpoint;
    }

    // 添加查询参数
    if (params) {
      const urlObj = new URL(url);
      Object.entries(params).forEach(([k, v]) => urlObj.searchParams.set(k, v));
      url = urlObj.toString();
    }

    return url;
  }

  /**
   * GET 请求
   */
  async get<T>(endpoint: string, params?: Record<string, string>, signal?: AbortSignal): Promise<T> {
    const url = this.buildUrl(endpoint, params);
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      signal, // 支持 AbortSignal，用于取消请求
    });

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new ApiError(response.status, data);
    }

    return response.json();
  }

  /**
   * POST 请求
   */
  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    const url = this.buildUrl(endpoint);
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: data ? JSON.stringify(data) : undefined,
    });

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new ApiError(response.status, data);
    }

    return response.json();
  }

  /**
   * PUT 请求
   */
  async put<T>(endpoint: string, data?: unknown): Promise<T> {
    const url = this.buildUrl(endpoint);
    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: data ? JSON.stringify(data) : undefined,
    });

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new ApiError(response.status, data);
    }

    return response.json();
  }

  /**
   * DELETE 请求
   */
  async delete<T>(endpoint: string): Promise<T> {
    const url = this.buildUrl(endpoint);
    const response = await fetch(url, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new ApiError(response.status, data);
    }

    return response.json();
  }
}

/**
 * 导出单例实例（默认值，会在配置加载后更新）
 */
export const apiClient = new ApiClient();

/**
 * 更新 API 客户端的后端 URL
 * 此函数在配置加载后调用
 */
export function updateApiClientUrl(baseUrl: string): void {
  apiClient.setBaseUrl(baseUrl);
}
