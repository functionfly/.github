import { STATUS_API_BASE_URL } from "@/lib/api";
import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, useState } from "react";

function getStatusWebSocketUrl(): string {
  const base = STATUS_API_BASE_URL.trim();
  if (base.startsWith("/")) {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}/ws/v1/status`;
  }
  const wsOrigin = base.replace(/^http:/, "ws:").replace(/^https:/, "wss:");
  return `${wsOrigin.replace(/\/$/, "")}/ws/v1/status`;
}

interface WebSocketMessage {
  type: string;
  channel?: string;
  timestamp: string;
  data?: unknown;
}

interface UseStatusWebSocketOptions {
  onMessage?: (message: WebSocketMessage) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

// Reconnection configuration - exponential backoff
const RECONNECT_BASE_DELAY = 1000;
const RECONNECT_MAX_DELAY = 30000;

export function useStatusWebSocket(options: UseStatusWebSocketOptions = {}) {
  const {
    onMessage,
    onConnect,
    onDisconnect,
    maxReconnectAttempts = 10,
  } = options;

  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
    null,
  );
  /** Bumps when starting a new socket or disconnecting; stale onclose handlers bail out. */
  const connectionGenRef = useRef(0);
  /** Tracks if disconnect was intentional to prevent auto-reconnect */
  const isIntentionallyClosedRef = useRef(false);
  /**
   * Synchronous guard — isConnecting state inside connect() would be stale (not in useCallback deps).
   */
  const connectInFlightRef = useRef(false);
  const isConnectedRef = useRef(false);
  const onMessageRef = useRef(onMessage);
  const onConnectRef = useRef(onConnect);
  const onDisconnectRef = useRef(onDisconnect);
  const queryClient = useQueryClient();

  useEffect(() => {
    onMessageRef.current = onMessage;
    onConnectRef.current = onConnect;
    onDisconnectRef.current = onDisconnect;
  }, [onMessage, onConnect, onDisconnect]);

  useEffect(() => {
    isConnectedRef.current = isConnected;
  }, [isConnected]);

  // Calculate reconnect delay with true exponential backoff
  const getReconnectDelay = useCallback(() => {
    const delay = Math.min(
      RECONNECT_BASE_DELAY * Math.pow(2, reconnectAttemptsRef.current),
      RECONNECT_MAX_DELAY,
    );
    // Add small random jitter to prevent thundering herd
    return delay + Math.random() * 1000;
  }, []);

  const connect = useCallback(() => {
    if (connectInFlightRef.current) {
      return;
    }

    // Clear any existing reconnect timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    const existing = wsRef.current;
    if (
      existing &&
      (existing.readyState === WebSocket.OPEN ||
        existing.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    if (existing) {
      existing.close();
      wsRef.current = null;
    }

    // Mark as not intentionally closed when attempting to connect
    isIntentionallyClosedRef.current = false;
    connectInFlightRef.current = true;
    setIsConnecting(true);
    setError(null);

    const socketGen = ++connectionGenRef.current;

    try {
      const wsUrl = getStatusWebSocketUrl();
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        if (socketGen !== connectionGenRef.current) return;
        console.log("[StatusWebSocket] Connected");
        connectInFlightRef.current = false;
        setIsConnected(true);
        setIsConnecting(false);
        setError(null);
        reconnectAttemptsRef.current = 0;
        isIntentionallyClosedRef.current = false;

        ws.send(
          JSON.stringify({
            type: "subscribe",
            channels: ["status", "incidents", "components"],
          }),
        );

        onConnectRef.current?.();
      };

      ws.onmessage = (event) => {
        if (socketGen !== connectionGenRef.current) return;
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          if (import.meta.env.DEV) {
            console.log("[StatusWebSocket] Message received:", message);
          }

          switch (message.type) {
            case "status_update":
              queryClient.invalidateQueries({ queryKey: ["platformStatus"] });
              break;
            case "incident_update":
              queryClient.invalidateQueries({ queryKey: ["incidents"] });
              break;
            case "component_update":
              queryClient.invalidateQueries({ queryKey: ["components"] });
              break;
            case "ping":
              ws.send(JSON.stringify({ type: "pong" }));
              break;
            default:
              break;
          }

          onMessageRef.current?.(message);
        } catch (err) {
          console.error("[StatusWebSocket] Error parsing message:", err);
        }
      };

      ws.onerror = () => {
        if (socketGen !== connectionGenRef.current) return;
        if (import.meta.env.DEV) {
          console.error(
            "[StatusWebSocket] Error (see Network tab for details)",
          );
        }
        // Avoid setState here: onclose always follows; duplicate updates caused re-render churn.
      };

      ws.onclose = (event) => {
        if (socketGen !== connectionGenRef.current) return;
        if (import.meta.env.DEV) {
          console.log(
            "[StatusWebSocket] Disconnected:",
            event.code,
            event.reason,
          );
        }
        connectInFlightRef.current = false;
        setIsConnected(false);
        setIsConnecting(false);
        wsRef.current = null;
        if (isIntentionallyClosedRef.current) {
          setError(null);
        }
        onDisconnectRef.current?.();

        // Only reconnect if not intentionally closed and not clean close
        if (
          !isIntentionallyClosedRef.current &&
          !event.wasClean &&
          reconnectAttemptsRef.current < maxReconnectAttempts
        ) {
          reconnectAttemptsRef.current++;
          const delay = getReconnectDelay();

          if (import.meta.env.DEV) {
            console.log(
              `[StatusWebSocket] Reconnecting in ${Math.round(delay)}ms... Attempt ${reconnectAttemptsRef.current}/${maxReconnectAttempts}`,
            );
          }

          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, delay);
        } else if (
          !isIntentionallyClosedRef.current &&
          !event.wasClean &&
          reconnectAttemptsRef.current >= maxReconnectAttempts
        ) {
          console.error("[StatusWebSocket] Max reconnection attempts reached");
          setError(new Error("Max reconnection attempts reached"));
        }
      };
    } catch (err) {
      connectionGenRef.current += 1;
      connectInFlightRef.current = false;
      setIsConnecting(false);
      setError(err instanceof Error ? err : new Error("Failed to connect"));
      console.error("[StatusWebSocket] Connection error:", err);
    }
  }, [maxReconnectAttempts, getReconnectDelay, queryClient]);

  const disconnect = useCallback(() => {
    // Mark as intentionally closed to prevent auto-reconnect
    isIntentionallyClosedRef.current = true;

    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    connectionGenRef.current += 1;
    connectInFlightRef.current = false;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setIsConnected(false);
    setIsConnecting(false);
    setError(null);
  }, [maxReconnectAttempts]);

  const sendMessage = useCallback((message: unknown) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
      return true;
    }
    return false;
  }, []);

  const connectRef = useRef(connect);
  const disconnectRef = useRef(disconnect);
  connectRef.current = connect;
  disconnectRef.current = disconnect;

  // Connect once on mount; omitting connect from deps avoids teardown/reconnect loops when its identity changes.
  useEffect(() => {
    connectRef.current();

    return () => {
      disconnectRef.current();
    };
  }, []);

  // Reconnect when the tab becomes visible again (refs avoid effect churn from isConnected / isConnecting).
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState !== "visible") return;
      if (isConnectedRef.current || connectInFlightRef.current) return;
      reconnectAttemptsRef.current = 0;
      isIntentionallyClosedRef.current = false;
      connectRef.current();
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () =>
      document.removeEventListener("visibilitychange", handleVisibilityChange);
  }, []);

  return {
    isConnected,
    isConnecting,
    error,
    connect,
    disconnect,
    sendMessage,
  };
}

// Hook specifically for the status page that handles all real-time updates
export function useStatusRealtime() {
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  const { isConnected } = useStatusWebSocket({
    onMessage: (message) => {
      setLastUpdate(new Date());

      // Log update for debugging
      if (import.meta.env.DEV) {
        console.log("[StatusRealtime] Update received:", message.type);
      }
    },
  });

  return {
    isConnected,
    lastUpdate,
  };
}
