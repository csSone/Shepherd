/**
 * 前端配置类型定义
 * 独立的类型文件，避免循环依赖
 */

/**
 * API 配置
 */
export interface ApiConfig {
  baseUrl: string;
  basePath: string;
  timeout: number;
  connectTimeout: number;
  retryCount: number;
  retryDelay: number;
}

/**
 * SSE 配置
 */
export interface SseConfig {
  endpoint: string;
  reconnect: boolean;
  reconnectDelay: number;
  maxReconnectAttempts: number;
  connectionTimeout: number;
  heartbeatEnabled: boolean;
  heartbeatInterval: number;
}

/**
 * 功能开关配置
 */
export interface FeaturesConfig {
  models: boolean;
  downloads: boolean;
  cluster: boolean;
  logs: boolean;
  chat: boolean;
  settings: boolean;
  dashboard: boolean;
}

/**
 * UI 配置
 */
export interface UiConfig {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  pageSize: number;
  pageSizeOptions: number[];
  virtualScrollThreshold: number;
  animations: boolean;
  skeleton: boolean;
  breadcrumb: boolean;
  sidebarExpanded: boolean;
  compactMode: boolean;
}

/**
 * 日志配置
 */
export interface LoggingConfig {
  level: 'debug' | 'info' | 'warn' | 'error';
  console: boolean;
  remote: boolean;
  remoteEndpoint: string;
  batchSize: number;
  flushInterval: number;
}

/**
 * 缓存配置
 */
export interface CacheConfig {
  modelsTTL: number;
  clientsTTL: number;
  downloadsTTL: number;
  configTTL: number;
  persistent: boolean;
  prefix: string;
  versioning: boolean;
}

/**
 * OpenAI 配置
 */
export interface OpenAIConfig {
  endpoint: string;
  defaultModel: string;
  temperature: number;
  maxTokens: number;
  topP: number;
  frequencyPenalty: number;
  presencePenalty: number;
  streamTimeout: number;
}

/**
 * 性能配置
 */
export interface PerformanceConfig {
  monitoring: boolean;
  sampleRate: number;
  preloading: boolean;
  virtualScroll: boolean;
  codeSplitting: boolean;
  lazyImageThreshold?: number;
  preloadResources?: string[];
}

/**
 * 应用配置接口
 */
export interface AppConfig {
  api: ApiConfig;
  sse: SseConfig;
  features: FeaturesConfig;
  ui: UiConfig;
  logging: LoggingConfig;
  cache: CacheConfig;
  openai?: OpenAIConfig;
  performance?: PerformanceConfig;
}

/**
 * 服务器配置接口（向后兼容）
 */
export interface ServerConfig {
  host: string;
  port: number;
  https: boolean;
  cors: {
    enabled: boolean;
    origin: string;
    methods: string;
    headers: string;
    credentials: boolean;
  };
}
