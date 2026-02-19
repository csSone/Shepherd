/**
 * 运行时配置类型定义
 */
export interface AppConfig {
  app: {
    name: string;
    version: string;
    description: string;
  };
  api: {
    baseUrl: string;
    timeout: number;
    retryCount: number;
    retryDelay: number;
  };
  sse: {
    endpoint: string;
    reconnect: boolean;
    reconnectDelay: number;
    maxReconnectAttempts: number;
  };
  features: {
    models: boolean;
    downloads: boolean;
    cluster: boolean;
    logs: boolean;
    chat: boolean;
  };
  ui: {
    theme: 'light' | 'dark' | 'auto';
    language: string;
    pageSize: number;
    virtualScrollThreshold: number;
  };
  logging: {
    level: 'debug' | 'info' | 'warn' | 'error';
    console: boolean;
    remote: boolean;
  };
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
  app: {
    name: 'Shepherd',
    version: '1.0.0',
    description: '分布式 AI 模型管理系统',
  },
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
      // 尝试从服务器加载配置
      const response = await fetch('/api/config', {
        timeout: 5000,
      });

      if (response.ok) {
        const serverConfig = await response.json();
        this.config = this.mergeConfig(defaultConfig, serverConfig);
      }
    } catch (error) {
      console.warn('Failed to load server config, using defaults:', error);
      // 使用默认配置
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
