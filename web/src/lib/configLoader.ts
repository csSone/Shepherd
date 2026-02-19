/**
 * 前端独立配置加载器
 * 从 web/config.yaml 加载配置，不依赖后端
 */

import { AppConfig } from './config';

/**
 * 配置文件路径
 */
const CONFIG_PATH = '/config.yaml';

/**
 * 配置加载器类
 */
class ConfigLoader {
  private config: AppConfig | null = null;
  private loadPromise: Promise<AppConfig> | null = null;

  /**
   * 加载配置文件
   */
  async load(): Promise<AppConfig> {
    // 如果已经加载过，直接返回
    if (this.config) {
      return this.config;
    }

    // 如果正在加载，等待加载完成
    if (this.loadPromise) {
      return this.loadPromise;
    }

    // 开始加载
    this.loadPromise = this.fetchConfig();

    try {
      this.config = await this.loadPromise;
      return this.config;
    } finally {
      this.loadPromise = null;
    }
  }

  /**
   * 从服务器获取配置文件
   */
  private async fetchConfig(): Promise<AppConfig> {
    try {
      const response = await fetch(CONFIG_PATH, {
        headers: {
          'Accept': 'text/yaml, application/yaml',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to load config: ${response.status} ${response.statusText}`);
      }

      const yamlText = await response.text();

      // 解析 YAML（简单的键值对解析）
      const config = this.parseYaml(yamlText);

      return config;
    } catch (error) {
      console.error('Failed to load frontend config, using defaults:', error);
      return this.getDefaultConfig();
    }
  }

  /**
   * 解析 YAML 配置文件
   * 注意：这是一个简化的 YAML 解析器，仅处理我们的配置文件格式
   */
  private parseYaml(yamlText: string): AppConfig {
    const lines = yamlText.split('\n');
    const config: any = {};
    const stack: Array<{ obj: any; level: number }> = [{ obj: config, level: -1 }];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const trimmed = line.trim();

      // 跳过空行和注释
      if (!trimmed || trimmed.startsWith('#')) {
        continue;
      }

      // 计算缩进级别
      const indent = line.search(/\S|$//);
      const level = Math.floor(indent / 2);

      // 弹出栈到正确的层级
      while (stack.length > 1 && stack[stack.length - 1].level >= level) {
        stack.pop();
      }

      const current = stack[stack.length - 1].obj;

      // 解析键值对
      if (trimmed.includes(':')) {
        const [key, ...valueParts] = trimmed.split(':');
        const value = valueParts.join(':').trim();

        if (value === '' || value === '|') {
          // 这是一个对象或数组的开始
          const newObj: any = Array.isArray(current[key]) ? [] : {};
          current[key] = newObj;
          stack.push({ obj: newObj, level });
        } else if (value.startsWith('"') || value.startsWith("'")) {
          // 字符串值
          current[key] = value.slice(1, -1);
        } else if (value === 'true' || value === 'false') {
          // 布尔值
          current[key] = value === 'true';
        } else if (value === 'null' || value === '~') {
          // null 值
          current[key] = null;
        } else if (!isNaN(Number(value))) {
          // 数字值
          current[key] = Number(value);
        } else if (value.startsWith('- ')) {
          // 数组
          current[key] = [];
          stack.push({ obj: current[key], level });
        } else {
          // 其他值作为字符串
          current[key] = value;
        }
      } else if (trimmed.startsWith('- ')) {
        // 数组项
        const itemValue = trimmed.slice(2).trim();
        if (Array.isArray(current)) {
          if (itemValue.startsWith('"') || itemValue.startsWith("'")) {
            current.push(itemValue.slice(1, -1));
          } else if (itemValue === 'true' || itemValue === 'false') {
            current.push(itemValue === 'true');
          } else if (!isNaN(Number(itemValue))) {
            current.push(Number(itemValue));
          } else {
            current.push(itemValue);
          }
        }
      }
    }

    return this.normalizeConfig(config);
  }

  /**
   * 规范化配置，确保类型正确
   */
  private normalizeConfig(config: any): AppConfig {
    return {
      ...config,
      api: {
        baseUrl: config.backend?.urls?.[config.backend?.currentIndex || 0] || 'http://localhost:9190',
        basePath: '/api',
        timeout: config.backend?.timeout || 30000,
        connectTimeout: config.backend?.timeout || 10000,
        retryCount: config.backend?.retry?.count || 3,
        retryDelay: config.backend?.retry?.delay || 1000,
      },
      sse: {
        endpoint: config.sse?.endpoint || '/api/events',
        reconnect: config.sse?.reconnect ?? true,
        reconnectDelay: config.sse?.reconnectDelay || 2000,
        maxReconnectAttempts: config.sse?.maxReconnectAttempts ?? -1,
        connectionTimeout: config.sse?.connectionTimeout || 60000,
        heartbeatEnabled: config.sse?.heartbeatEnabled ?? true,
        heartbeatInterval: config.sse?.heartbeatInterval || 30000,
      },
      features: {
        models: config.features?.models ?? true,
        downloads: config.features?.downloads ?? true,
        cluster: config.features?.cluster ?? true,
        logs: config.features?.logs ?? true,
        chat: config.features?.chat ?? true,
        settings: config.features?.settings ?? true,
        dashboard: config.features?.dashboard ?? true,
      },
      ui: {
        theme: config.ui?.theme || 'auto',
        language: config.ui?.language || 'zh-CN',
        pageSize: config.ui?.pageSize || 20,
        pageSizeOptions: config.ui?.pageSizeOptions || [10, 20, 50, 100],
        virtualScrollThreshold: config.ui?.virtualScrollThreshold || 100,
        animations: config.ui?.animations ?? true,
        skeleton: config.ui?.skeleton ?? true,
        breadcrumb: config.ui?.breadcrumb ?? true,
        sidebarExpanded: config.ui?.sidebarExpanded ?? true,
        compactMode: config.ui?.compactMode ?? false,
      },
      logging: {
        level: config.logging?.level || 'info',
        console: config.logging?.console ?? true,
        remote: config.logging?.remote ?? false,
        remoteEndpoint: config.logging?.remoteEndpoint || '/api/logs',
        batchSize: config.logging?.batchSize || 50,
        flushInterval: config.logging?.flushInterval || 5000,
      },
      cache: {
        modelsTTL: config.cache?.modelsTTL || 60000,
        clientsTTL: config.cache?.clientsTTL || 30000,
        downloadsTTL: config.cache?.downloadsTTL || 5000,
        configTTL: config.cache?.configTTL || 300000,
        persistent: config.cache?.persistent ?? true,
        prefix: config.cache?.prefix || 'shepherd_web_',
        versioning: config.cache?.versioning ?? true,
      },
      openai: config.openai ? {
        endpoint: config.openai.endpoint || '/v1/chat/completions',
        defaultModel: config.openai.defaultModel || '',
        temperature: config.openai.temperature ?? 0.7,
        maxTokens: config.openai.maxTokens || 4096,
        topP: config.openai.topP ?? 0.9,
        frequencyPenalty: config.openai.frequencyPenalty ?? 0,
        presencePenalty: config.openai.presencePenalty ?? 0,
        streamTimeout: config.openai.streamTimeout || 120000,
      } : undefined,
      performance: config.performance ? {
        monitoring: config.performance.monitoring ?? false,
        sampleRate: config.performance.sampleRate ?? 0.1,
        preloading: config.performance.preloading ?? true,
        virtualScroll: config.performance.virtualScroll ?? true,
        codeSplitting: config.performance.codeSplitting ?? true,
        lazyImageThreshold: config.performance.lazyImageThreshold,
        preloadResources: config.performance.preloadResources,
      } : undefined,
    };
  }

  /**
   * 获取默认配置
   */
  private getDefaultConfig(): AppConfig {
    return {
      api: {
        baseUrl: 'http://localhost:9190',
        basePath: '/api',
        timeout: 30000,
        connectTimeout: 10000,
        retryCount: 3,
        retryDelay: 1000,
      },
      sse: {
        endpoint: '/api/events',
        reconnect: true,
        reconnectDelay: 2000,
        maxReconnectAttempts: -1,
        connectionTimeout: 60000,
        heartbeatEnabled: true,
        heartbeatInterval: 30000,
      },
      features: {
        models: true,
        downloads: true,
        cluster: true,
        logs: true,
        chat: true,
        settings: true,
        dashboard: true,
      },
      ui: {
        theme: 'auto',
        language: 'zh-CN',
        pageSize: 20,
        pageSizeOptions: [10, 20, 50, 100],
        virtualScrollThreshold: 100,
        animations: true,
        skeleton: true,
        breadcrumb: true,
        sidebarExpanded: true,
        compactMode: false,
      },
      logging: {
        level: 'info',
        console: true,
        remote: false,
        remoteEndpoint: '/api/logs',
        batchSize: 50,
        flushInterval: 5000,
      },
      cache: {
        modelsTTL: 60000,
        clientsTTL: 30000,
        downloadsTTL: 5000,
        configTTL: 300000,
        persistent: true,
        prefix: 'shepherd_web_',
        versioning: true,
      },
    };
  }

  /**
   * 获取当前配置
   */
  getConfig(): AppConfig {
    if (!this.config) {
      throw new Error('Config not loaded. Call load() first.');
    }
    return this.config;
  }

  /**
   * 获取后端 URL
   */
  getBackendUrl(): string {
    const config = this.getConfig();
    return config.api.baseUrl;
  }

  /**
   * 切换到不同的后端
   */
  async switchBackend(index: number): Promise<void> {
    // 重新加载配置
    this.config = null;
    const newConfig = await this.load();

    // 如果指定了索引，使用它
    if (index >= 0 && index < (newConfig as any).backendUrls?.length) {
      (newConfig as any).backend.currentIndex = index;
    }

    this.config = newConfig;
  }
}

// 导出单例
export const configLoader = new ConfigLoader();

/**
 * 导出便捷 Hook
 */
export function useConfig() {
  if (!configLoader.config) {
    console.warn('Config not loaded yet. Using default values.');
    return configLoader['getDefaultConfig']();
  }
  return configLoader.getConfig();
}
