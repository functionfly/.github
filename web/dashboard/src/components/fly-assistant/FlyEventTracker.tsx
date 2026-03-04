/**
 * FlyEventTracker.tsx
 *
 * Tracks errors, deployments, trust changes, marketplace interactions,
 * and assistant usage with batching and privacy support.
 *
 * @module fly-assistant
 */

import {
  createContext,
  useContext,
  useCallback,
  useRef,
  useEffect,
  useState,
  ReactNode,
} from "react";
import {
  useFlyAssistant,
  UserRole,
} from "./FlyAssistantProvider";

// ============================================================================
// Types & Interfaces
// ============================================================================

/**
 * Event types that can be tracked
 */
export type TrackedEventType =
  | "error"
  | "deployment"
  | "trust_change"
  | "marketplace_view"
  | "marketplace_install"
  | "marketplace_rate"
  | "assistant_message"
  | "assistant_action"
  | "assistant_open"
  | "assistant_close"
  | "execution_preview"
  | "execution_confirm"
  | "execution_cancel"
  | "page_view"
  | "custom";

/**
 * Risk level for execution events
 */
export type RiskLevel = "low" | "medium" | "high" | "critical";

/**
 * Tracked event structure
 */
export interface TrackedEvent {
  /** Unique event ID */
  id: string;
  /** Event type */
  type: TrackedEventType;
  /** Event timestamp */
  timestamp: number;
  /** Event data payload */
  data: Record<string, unknown>;
  /** Event context */
  context: {
    /** Current route path */
    route: string;
    /** Current function ID if applicable */
    function?: string;
    /** User role */
    userRole: UserRole;
    /** Session ID */
    sessionId?: string;
  };
  /** Session sequence number */
  sequenceNumber: number;
}

/**
 * Props for FlyEventTracker provider
 */
export interface FlyEventTrackerProps {
  /** Child components */
  children: ReactNode;
  /** Whether tracking is enabled */
  enabled?: boolean;
  /** Number of events to batch before flushing */
  batchSize?: number;
  /** Interval to flush events in ms */
  flushIntervalMs?: number;
  /** Callback for each tracked event */
  onEvent?: (event: TrackedEvent) => void;
  /** Callback for batched events */
  onBatch?: (events: TrackedEvent[]) => void;
  /** Respect Do Not Track header */
  respectDNT?: boolean;
  /** Enable debug overlay */
  debug?: boolean;
  /** Session ID for tracking */
  sessionId?: string;
}

/**
 * Error event data
 */
export interface ErrorEventData {
  message: string;
  stack?: string;
  source?: string;
  lineno?: number;
  colno?: number;
}

/**
 * Deployment event data
 */
export interface DeploymentEventData {
  functionId: string;
  version: string;
  environment: string;
  success: boolean;
  duration?: number;
  error?: string;
}

/**
 * Trust change event data
 */
export interface TrustChangeEventData {
  previousScore: number;
  newScore: number;
  reason?: string;
}

/**
 * Marketplace event data
 */
export interface MarketplaceEventData {
  functionId: string;
  functionName: string;
  category?: string;
  rating?: number;
}

/**
 * Assistant usage event data
 */
export interface AssistantUsageEventData {
  messageCount?: number;
  actionType?: string;
  mode?: string;
}

// ============================================================================
// Context Types
// ============================================================================

interface EventTrackerContextValue {
  /** Track a generic event */
  track: (type: TrackedEventType, data?: Record<string, unknown>) => void;
  /** Track an error */
  trackError: (error: Error, context?: Record<string, unknown>) => void;
  /** Track a deployment */
  trackDeployment: (data: DeploymentEventData) => void;
  /** Track trust score change */
  trackTrustChange: (data: TrustChangeEventData) => void;
  /** Track marketplace interaction */
  trackMarketplace: (
    subtype: "view" | "install" | "rate",
    data: MarketplaceEventData
  ) => void;
  /** Track assistant usage */
  trackAssistant: (
    subtype: "message" | "action" | "open" | "close",
    data?: AssistantUsageEventData
  ) => void;
  /** Manually flush events */
  flush: () => void;
  /** Whether tracking is enabled */
  isEnabled: boolean;
  /** Pending events count */
  pendingCount: number;
}

// ============================================================================
// Context
// ============================================================================

const EventTrackerContext = createContext<EventTrackerContextValue | null>(null);

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Check if Do Not Track is enabled
 */
function isDNTEnabled(): boolean {
  if (typeof window === "undefined") return false;
  return (
    navigator.doNotTrack === "1" ||
    navigator.doNotTrack === "yes" ||
    (window as { doNotTrack?: string }).doNotTrack === "1"
  );
}

/**
 * Generate unique event ID
 */
function generateEventId(): string {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Generate session ID
 */
function generateSessionId(): string {
  return `session-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

// ============================================================================
// Hook
// ============================================================================

/**
 * Hook for tracking events
 *
 * @example
 * ```tsx
 * const { track, trackError } = useEventTracking();
 * track("custom", { button: "clicked" });
 * trackError(new Error("Something failed"));
 * ```
 */
export function useEventTracking(): EventTrackerContextValue {
  const context = useContext(EventTrackerContext);
  if (!context) {
    throw new Error(
      "useEventTracking must be used within FlyEventTracker"
    );
  }
  return context;
}

// ============================================================================
// Component
// ============================================================================

/**
 * FlyEventTracker - Event tracking provider component
 *
 * Batches events and sends them periodically. Respects DNT headers.
 * Provides hooks for tracking various event types.
 *
 * @example
 * ```tsx
 * <FlyEventTracker
 *   enabled={true}
 *   batchSize={10}
 *   flushIntervalMs={5000}
 *   onBatch={(events) => sendToAnalytics(events)}
 *   respectDNT={true}
 * >
 *   <App />
 * </FlyEventTracker>
 * ```
 */
export function FlyEventTracker({
  children,
  enabled = true,
  batchSize = 10,
  flushIntervalMs = 5000,
  onEvent,
  onBatch,
  respectDNT = true,
  debug = false,
  sessionId: providedSessionId,
}: FlyEventTrackerProps) {
  const currentRoute = useFlyAssistant((state) => state.currentRoute);
  const userSession = useFlyAssistant((state) => state.userSession);

  const eventBufferRef = useRef<TrackedEvent[]>([]);
  const sequenceRef = useRef(0);
  const flushIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const sessionIdRef = useRef(providedSessionId || generateSessionId());

  const [pendingCount, setPendingCount] = useState(0);

  // Check if tracking should be enabled
  const isEnabled = enabled && !(respectDNT && isDNTEnabled());

  /**
   * Create a tracked event
   */
  const createEvent = useCallback(
    (type: TrackedEventType, data: Record<string, unknown> = {}): TrackedEvent => {
      sequenceRef.current += 1;

      return {
        id: generateEventId(),
        type,
        timestamp: Date.now(),
        data,
        context: {
          route: currentRoute?.path || window.location.pathname,
          function: currentRoute?.params?.id as string | undefined,
          userRole: userSession?.role || "free",
          sessionId: sessionIdRef.current,
        },
        sequenceNumber: sequenceRef.current,
      };
    },
    [currentRoute, userSession]
  );

  /**
   * Flush events to callback
   */
  const flush = useCallback(() => {
    if (eventBufferRef.current.length === 0) return;

    const events = [...eventBufferRef.current];
    eventBufferRef.current = [];
    setPendingCount(0);

    if (onBatch) {
      onBatch(events);
    }

    events.forEach((event) => {
      if (onEvent) {
        onEvent(event);
      }
    });
  }, [onBatch, onEvent]);

  /**
   * Add event to buffer
   */
  const addToBuffer = useCallback(
    (event: TrackedEvent) => {
      if (!isEnabled) return;

      eventBufferRef.current.push(event);
      setPendingCount(eventBufferRef.current.length);

      // Flush if batch size reached
      if (eventBufferRef.current.length >= batchSize) {
        flush();
      }
    },
    [isEnabled, batchSize, flush]
  );

  /**
   * Track a generic event
   */
  const track = useCallback(
    (type: TrackedEventType, data?: Record<string, unknown>) => {
      if (!isEnabled) return;
      const event = createEvent(type, data);
      addToBuffer(event);
    },
    [isEnabled, createEvent, addToBuffer]
  );

  /**
   * Track an error
   */
  const trackError = useCallback(
    (error: Error, context?: Record<string, unknown>) => {
      if (!isEnabled) return;

      const errorData: Record<string, unknown> = {
        message: error.message,
        stack: error.stack,
        ...context,
      };

      const event = createEvent("error", errorData);
      addToBuffer(event);
    },
    [isEnabled, createEvent, addToBuffer]
  );

  /**
   * Track a deployment
   */
  const trackDeployment = useCallback(
    (data: DeploymentEventData) => {
      if (!isEnabled) return;
      const event = createEvent("deployment", { ...data });
      addToBuffer(event);
    },
    [isEnabled, createEvent, addToBuffer]
  );

  /**
   * Track trust score change
   */
  const trackTrustChange = useCallback(
    (data: TrustChangeEventData) => {
      if (!isEnabled) return;
      const event = createEvent("trust_change", { ...data });
      addToBuffer(event);
    },
    [isEnabled, createEvent, addToBuffer]
  );

  /**
   * Track marketplace interaction
   */
  const trackMarketplace = useCallback(
    (subtype: "view" | "install" | "rate", data: MarketplaceEventData) => {
      if (!isEnabled) return;

      const typeMap: Record<string, TrackedEventType> = {
        view: "marketplace_view",
        install: "marketplace_install",
        rate: "marketplace_rate",
      };

      const event = createEvent(typeMap[subtype], { ...data });
      addToBuffer(event);
    },
    [isEnabled, createEvent, addToBuffer]
  );

  /**
   * Track assistant usage
   */
  const trackAssistant = useCallback(
    (subtype: "message" | "action" | "open" | "close", data?: AssistantUsageEventData) => {
      if (!isEnabled) return;

      const typeMap: Record<string, TrackedEventType> = {
        message: "assistant_message",
        action: "assistant_action",
        open: "assistant_open",
        close: "assistant_close",
      };

      const event = createEvent(typeMap[subtype], data as Record<string, unknown>);
      addToBuffer(event);
    },
    [isEnabled, createEvent, addToBuffer]
  );

  // Setup flush interval
  useEffect(() => {
    if (!isEnabled) return;

    flushIntervalRef.current = setInterval(() => {
      flush();
    }, flushIntervalMs);

    return () => {
      if (flushIntervalRef.current) {
        clearInterval(flushIntervalRef.current);
      }
    };
  }, [isEnabled, flushIntervalMs, flush]);

  // Flush on unmount
  useEffect(() => {
    return () => {
      flush();
    };
  }, [flush]);

  // Global error handler
  useEffect(() => {
    if (!isEnabled) return;

    const handleError = (event: ErrorEvent) => {
      trackError(event.error || new Error(event.message), {
        source: event.filename,
        lineno: event.lineno,
        colno: event.colno,
      });
    };

    window.addEventListener("error", handleError);
    return () => window.removeEventListener("error", handleError);
  }, [isEnabled, trackError]);

  // Unhandled promise rejection handler
  useEffect(() => {
    if (!isEnabled) return;

    const handleRejection = (event: PromiseRejectionEvent) => {
      const error = event.reason instanceof Error
        ? event.reason
        : new Error(String(event.reason));
      trackError(error, { type: "unhandled_rejection" });
    };

    window.addEventListener("unhandledrejection", handleRejection);
    return () => window.removeEventListener("unhandledrejection", handleRejection);
  }, [isEnabled, trackError]);

  const contextValue: EventTrackerContextValue = {
    track,
    trackError,
    trackDeployment,
    trackTrustChange,
    trackMarketplace,
    trackAssistant,
    flush,
    isEnabled,
    pendingCount,
  };

  return (
    <EventTrackerContext.Provider value={contextValue}>
      {children}
      {debug && isEnabled && <EventDebugOverlay pendingCount={pendingCount} />}
    </EventTrackerContext.Provider>
  );
}

// ============================================================================
// Debug Overlay Component
// ============================================================================

interface EventDebugOverlayProps {
  pendingCount: number;
}

function EventDebugOverlay({ pendingCount }: EventDebugOverlayProps) {
  return (
    <div className="fixed bottom-4 left-4 z-[9999] rounded-lg bg-slate-900/90 px-3 py-2 text-xs text-white shadow-lg backdrop-blur-sm">
      <div className="flex items-center gap-2">
        <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
        <span>Events: {pendingCount}</span>
      </div>
    </div>
  );
}

export default FlyEventTracker;
