/**
 * Shepherd Web 前端运行时配置
 *
 * 配置加载优先级：
 * 1. web/config.yaml - 前端独立配置文件（推荐）
 * 2. 默认配置 - 代码中的默认值
 *
 * 前端完全独立运行，可以连接任意后端服务器
 */

// 导出配置类型（避免循环依赖）
export type {
  AppConfig,
  ServerConfig,
  ApiConfig,
  SseConfig,
  FeaturesConfig,
  UiConfig,
  LoggingConfig,
  CacheConfig,
  OpenAIConfig,
  PerformanceConfig
} from './configTypes';

// 导出配置加载器
export { configLoader, useConfig } from './configLoader';

// 导出 API 客户端的更新函数
export { updateApiClientUrl } from './api/client';
