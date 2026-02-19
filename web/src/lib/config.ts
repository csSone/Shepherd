/**
 * Shepherd Web 前端运行时配置
 * 配置文件位置: config/web.config.yaml
 */

/**
 * 运行时配置类型定义
 */
export interface AppConfig {
  // API 配置
  api: {
    baseUrl: string;
    timeout: number;
    retryCount: number;
    retryDelay: number;
  };
  // SSE 实时事件配置
  sse: {
    endpoint: string;
    reconnect: boolean;
    reconnectDelay: number;
    maxReconnectAttempts: number;
  };
  // 功能开关
  features: {
    models: boolean;
    downloads: boolean;
    cluster: boolean;
    logs: boolean;
    chat: boolean;
  };
  // UI 配置
  ui: {
    theme: 'light' | 'dark' | 'auto';
    language: string;
    pageSize: number;
    virtualScrollThreshold: number;
  };
  // 日志配置
  logging: {
    level: 'debug' | 'info' | 'warn' | 'error';
    console: boolean;
    remote: boolean;
  };
  // 缓存配置
  cache: {
    modelsTTL: number;
    clientsTTL: number;
    persistent: boolean;
  };
}

/**
 * 默认配置
 */
const defaultConfig: AppConfig = {
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:9190',
    timeout: 30000,
    retryCount: 3,
    retryDelay: 1000,
  },
  sse: {
    endpoint: '/api/events',
    reconnect: true,
    reconnectDelay: 2000,
    maxReconnectAttempts: -1,
  },
  features: {
    models: true,
    downloads: true,
    cluster: true,
    logs: true,
    chat: true,
  },
  ui: {
    theme: 'auto',
    language: 'zh-CN',
    pageSize: 20,
    virtualScrollThreshold: 100,
  },
  logging: {
    level: 'info',
    console: true,
    remote: false,
  },
  cache: {
    modelsTTL: 60000,
    clientsTTL: 30000,
    persistent: true,
  },
};

/**
 * 配置管理器
 */
class ConfigManager {
  private config: AppConfig;

  constructor() {
    this.config = defaultConfig;
  }

  /**
   * 初始化配置（从服务器加载）
   */
  async init(): Promise<void> {
    try {
      // 从后端 API 加载配置
      const response = await fetch('/api/config/web', {
        headers: {
          'Accept': 'application/json',
        },
      });

      if (response.ok) {
        const serverConfig = await response.json();
        this.config = this.mergeConfig(defaultConfig, serverConfig);
      } else {
        console.warn('Failed to load config from server, using defaults');
      }
    } catch (error) {
      console.warn('Failed to load server config, using defaults:', error);
    }
  }

  /**
   * 获取配置
   */
  getConfig(): AppConfig {
    return this.config;
  }

  /**
   * 获取特定配置项
   */
  get<K extends keyof AppConfig>(key: K): AppConfig[K] {
    return this.config[key];
  }

  /**
   * 合并配置
   */
  private mergeConfig(base: AppConfig, override: Partial<AppConfig>): AppConfig {
    return {
      ...base,
      ...override,
      api: { ...base.api, ...override.api },
      sse: { ...base.sse, ...override.sse },
      features: { ...base.features, ...override.features },
      ui: { ...base.ui, ...override.ui },
      logging: { ...base.logging, ...override.logging },
      cache: { ...base.cache, ...override.cache },
    };
  }
}

// 导出单例
export const config = new ConfigManager();

// 导出便捷 Hook
export function useConfig() {
  return config.getConfig();
}
