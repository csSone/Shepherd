import { useState, useEffect, useCallback, useRef } from 'react';
import type {
  WebSocketMessage,
  WebSocketConnectionStatus,
  UseWebSocketReturn,
  WebSocketClientOptions,
} from '@/types/websocket';
import { WebSocketClient } from '@/lib/websocket';
import { apiClient } from '@/lib/api/client';

interface UseWebSocketOptions extends Partial<WebSocketClientOptions> {
  url?: string;
  autoConnect?: boolean;
}

export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const {
    url,
    autoConnect = true,
    ...clientOptions
  } = options;

  const [connectionStatus, setConnectionStatus] = useState<WebSocketConnectionStatus>('disconnected');
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);

  const clientRef = useRef<WebSocketClient | null>(null);
  const mountedRef = useRef(true);

  const getWebSocketUrl = useCallback(() => {
    if (url) return url;
    
    const baseUrl = apiClient.getBaseUrl();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return baseUrl.replace(/^https?:/, protocol).replace(/\/api$/, '/ws');
  }, [url]);

  const createClient = useCallback(() => {
    const wsUrl = getWebSocketUrl();
    
    const client = new WebSocketClient({
      url: wsUrl,
      ...clientOptions,
    });

    client.setHandlers({
      onStatusChange: (status) => {
        if (mountedRef.current) {
          setConnectionStatus(status);
        }
      },
      onMessage: (message) => {
        if (mountedRef.current) {
          setLastMessage(message);
        }
      },
      onReconnecting: (attempt) => {
        if (mountedRef.current) {
          setReconnectAttempts(attempt);
        }
      },
      onReconnectFailed: () => {
        if (mountedRef.current) {
          setReconnectAttempts(clientOptions.maxReconnectAttempts ?? 5);
        }
      },
    });

    return client;
  }, [getWebSocketUrl, clientOptions]);

  const connect = useCallback(() => {
    if (!clientRef.current) {
      clientRef.current = createClient();
    }
    clientRef.current.connect();
  }, [createClient]);

  const disconnect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.disconnect();
    }
  }, []);

  const sendMessage = useCallback(<T = unknown,>(message: WebSocketMessage<T>) => {
    if (clientRef.current) {
      clientRef.current.send(message);
    }
  }, []);

  const sendRaw = useCallback((data: string | ArrayBuffer | Blob) => {
    if (clientRef.current) {
      clientRef.current.sendRaw(data);
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;

    if (autoConnect) {
      connect();
    }

    return () => {
      mountedRef.current = false;
      if (clientRef.current) {
        clientRef.current.disconnect();
        clientRef.current = null;
      }
    };
  }, [autoConnect, connect]);

  return {
    connectionStatus,
    lastMessage,
    sendMessage,
    sendRaw,
    connect,
    disconnect,
    reconnectAttempts,
    isConnected: connectionStatus === 'connected',
    isConnecting: connectionStatus === 'connecting',
  };
}
