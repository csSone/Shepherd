/**
 * WebSocket 类型定义
 * 用于实时通信的类型安全接口
 */

/**
 * WebSocket 连接状态
 */
export type WebSocketConnectionStatus =
  | 'connecting'    // 连接中
  | 'connected'     // 已连接
  | 'disconnected'  // 已断开
  | 'reconnecting'  // 重连中
  | 'error';        // 连接错误

/**
 * WebSocket 消息类型
 */
export type WebSocketMessageType =
  | 'ping'          // 心跳请求
  | 'pong'          // 心跳响应
  | 'event'         // 事件消息
  | 'notification'  // 通知消息
  | 'error';        // 错误消息

/**
 * WebSocket 基础消息接口
 */
export interface WebSocketMessage<T = unknown> {
  type: WebSocketMessageType;
  timestamp: number;
  payload: T;
}

/**
 * 心跳消息
 */
export interface PingMessage extends WebSocketMessage {
  type: 'ping';
}

/**
 * 心跳响应消息
 */
export interface PongMessage extends WebSocketMessage {
  type: 'pong';
}

/**
 * 事件消息
 */
export interface EventMessage<T = unknown> extends Omit<WebSocketMessage<T>, 'payload'> {
  type: 'event';
  payload: {
    eventType: string;
    data: T;
  };
}

/**
 * 通知消息
 */
export interface NotificationMessage extends WebSocketMessage {
  type: 'notification';
  payload: {
    title: string;
    message: string;
    level: 'info' | 'success' | 'warning' | 'error';
  };
}

/**
 * 错误消息
 */
export interface ErrorMessage extends WebSocketMessage {
  type: 'error';
  payload: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

/**
 * WebSocket 客户端配置
 */
export interface WebSocketClientOptions {
  /** WebSocket 服务器 URL */
  url: string;
  /** 心跳间隔（毫秒），默认 30000 */
  heartbeatInterval?: number;
  /** 心跳超时（毫秒），默认 10000 */
  heartbeatTimeout?: number;
  /** 最大重连次数，默认 5 */
  maxReconnectAttempts?: number;
  /** 初始重连延迟（毫秒），默认 1000 */
  initialReconnectDelay?: number;
  /** 最大重连延迟（毫秒），默认 30000 */
  maxReconnectDelay?: number;
  /** 是否自动重连，默认 true */
  autoReconnect?: boolean;
  /** 是否在连接时发送认证信息 */
  authPayload?: Record<string, unknown>;
  /** 调试模式 */
  debug?: boolean;
}

/**
 * WebSocket 客户端事件处理器
 */
export interface WebSocketEventHandlers {
  onOpen?: (event: Event) => void;
  onClose?: (event: CloseEvent) => void;
  onError?: (event: Event) => void;
  onMessage?: (message: WebSocketMessage) => void;
  onReconnecting?: (attempt: number, delay: number) => void;
  onReconnectFailed?: () => void;
  onStatusChange?: (status: WebSocketConnectionStatus) => void;
}

/**
 * WebSocket Hook 返回值
 */
export interface UseWebSocketReturn {
  /** 当前连接状态 */
  connectionStatus: WebSocketConnectionStatus;
  /** 最后接收到的消息 */
  lastMessage: WebSocketMessage | null;
  /** 发送消息 */
  sendMessage: <T = unknown>(message: WebSocketMessage<T>) => void;
  /** 发送原始数据 */
  sendRaw: (data: string | ArrayBuffer | Blob) => void;
  /** 手动连接 */
  connect: () => void;
  /** 手动断开 */
  disconnect: () => void;
  /** 重连次数 */
  reconnectAttempts: number;
  /** 是否已连接 */
  isConnected: boolean;
  /** 是否正在连接 */
  isConnecting: boolean;
}

/**
 * WebSocket Context 值
 */
export interface WebSocketContextValue extends UseWebSocketReturn {
  /** 订阅特定事件类型 */
  subscribe: <T = unknown>(
    eventType: string,
    handler: (data: T) => void
  ) => () => void;
  /** 取消所有订阅 */
  unsubscribeAll: () => void;
}

/**
 * WebSocket Provider Props
 */
export interface WebSocketProviderProps {
  /** 子组件 */
  children: React.ReactNode;
  /** WebSocket 服务器 URL（可选，默认从配置读取） */
  url?: string;
  /** 是否自动连接，默认 true */
  autoConnect?: boolean;
  /** 自定义配置 */
  options?: Partial<WebSocketClientOptions>;
  /** 连接成功回调 */
  onConnect?: () => void;
  /** 断开连接回调 */
  onDisconnect?: () => void;
  /** 错误回调 */
  onError?: (error: Error) => void;
}
