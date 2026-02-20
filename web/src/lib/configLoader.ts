import { useState, useEffect } from 'react';
import type { AppConfig } from './configTypes';

/**
 * 默认配置
 */
const DEFAULT_CONFIG: AppConfig = {
  api: {
    baseUrl: 'http://localhost:9190',
    basePath: '/api',
    timeout: 30000,
    connectTimeout: 5000,
    retryCount: 3,
    retryDelay: 1000,
  },
  sse: {
    endpoint: '/events',
    reconnect: true,
    reconnectDelay: 3000,
    maxReconnectAttempts: 10,
    connectionTimeout: 30000,
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
    pageSize: 10,
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
    remoteEndpoint: '',
    batchSize: 100,
    flushInterval: 5000,
  },
  cache: {
    modelsTTL: 300000,
    clientsTTL: 60000,
    downloadsTTL: 10000,
    configTTL: 300000,
    persistent: true,
    prefix: 'shepherd:',
    versioning: true,
  },
};

export class ConfigLoader {
  // Naive YAML parser for simple key: value + nesting by indentation
  static parseYaml(yaml: string): any {
    const lines = yaml.split(/\r?\n/).filter((l) => l.trim().length > 0)
    const root: any = {}
    const stack: any[] = [root]
    const indentStack: number[] = [-1]

    for (const rawLine of lines) {
      const line = rawLine
      const indent = line.match(/^\s*/)?.[0].length ?? 0
      const content = line.trim()
      const colonIndex = content.indexOf(':')
      if (colonIndex === -1) continue
      const key = content.substring(0, colonIndex).trim()
      const value = content.substring(colonIndex + 1).trim()

      // Adjust current context according to indentation
      while (indent <= indentStack[indentStack.length - 1] && stack.length > 0) {
        stack.pop()
        indentStack.pop()
      }

      if (value === '' || value === undefined) {
        const obj: any = {}
        stack[stack.length - 1][key] = obj
        stack.push(obj)
        indentStack.push(indent)
      } else {
        let val: any = value
        if (/^-?\d+$/.test(value)) val = Number(value)
        else if (value === 'true') val = true
        else if (value === 'false') val = false
        stack[stack.length - 1][key] = val
      }
    }

    return root
  }

  /**
   * 加载配置文件
   */
  async load(): Promise<AppConfig> {
    try {
      const response = await fetch('/config.yaml')
      if (!response.ok) {
        throw new Error(`Failed to load config: ${response.status}`)
      }
      const yamlText = await response.text()
      const parsed = ConfigLoader.parseYaml(yamlText)
      
      // 合并默认配置和解析的配置
      return this.mergeConfig(DEFAULT_CONFIG, parsed)
    } catch (error) {
      console.warn('Failed to load config.yaml, using default config:', error)
      return DEFAULT_CONFIG
    }
  }

  /**
   * 合并配置
   */
  private mergeConfig(defaults: AppConfig, loaded: any): AppConfig {
    return {
      api: { ...defaults.api, ...loaded?.api },
      sse: { ...defaults.sse, ...loaded?.sse },
      features: { ...defaults.features, ...loaded?.features },
      ui: { ...defaults.ui, ...loaded?.ui },
      logging: { ...defaults.logging, ...loaded?.logging },
      cache: { ...defaults.cache, ...loaded?.cache },
      openai: loaded?.openai ? { ...defaults.openai, ...loaded.openai } : defaults.openai,
      performance: loaded?.performance ? { ...defaults.performance, ...loaded.performance } : defaults.performance,
    }
  }
}

/**
 * 配置加载器单例实例
 */
export const configLoader = new ConfigLoader();

/**
 * React Hook: 获取配置
 * 
 * @example
 * const config = useConfig();
 * console.log(config.api.baseUrl);
 */
export function useConfig(): AppConfig {
  const [config, setConfig] = useState<AppConfig>(DEFAULT_CONFIG);

  useEffect(() => {
    let mounted = true;

    configLoader.load().then((loadedConfig) => {
      if (mounted) {
        setConfig(loadedConfig);
      }
    });

    return () => {
      mounted = false;
    };
  }, []);

  return config;
}
