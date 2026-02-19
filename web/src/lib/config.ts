/**
 * Shepherd Web 前端运行时配置
 *
 * 配置加载优先级：
 * 1. web/config.yaml - 前端独立配置文件（推荐）
 * 2. 默认配置 - 代码中的默认值
 *
 * 前端完全独立运行，可以连接任意后端服务器
 */

// 重新导出配置加载器和接口
export { configLoader, AppConfig, ServerConfig } from './configLoader';
export type { AppConfig } from './configLoader';

// 保留旧的 ConfigManager 作为备用（向后兼容）
// 但推荐使用 configLoader
export { useConfig } from './configLoader';

// 重新导出 API 客户端的更新函数
export { updateApiClientUrl } from './api/client';
