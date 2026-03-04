/**
 * useEventTracking.ts
 *
 * Hook for tracking events, errors, and analytics.
 * Must be used within FlyEventTracker provider.
 *
 * @module fly-assistant/hooks
 */

import { useCallback } from "react";
import {
  useEventTracking as useBaseEventTracking,
  DeploymentEventData,
  TrustChangeEventData,
  MarketplaceEventData,
  AssistantUsageEventData,
} from "../FlyEventTracker";

// ============================================================================
// Types
// ============================================================================

/**
 * Return type for useEventTracking hook
 */
export interface UseEventTrackingReturn {
  /** Track a generic event */
  track: (eventName: string, data?: Record<string, unknown>) => void;
  /** Track an error */
  trackError: (error: Error, context?: Record<string, unknown>) => void;
  /** Track a deployment event */
  trackDeployment: (data: DeploymentEventData) => void;
  /** Track a trust score change */
  trackTrustChange: (data: TrustChangeEventData) => void;
  /** Track marketplace interactions */
  trackMarketplaceView: (data: MarketplaceEventData) => void;
  trackMarketplaceInstall: (data: MarketplaceEventData) => void;
  trackMarketplaceRate: (data: MarketplaceEventData) => void;
  /** Track assistant usage */
  trackAssistantMessage: (data?: AssistantUsageEventData) => void;
  trackAssistantAction: (actionType: string, data?: Record<string, unknown>) => void;
  trackAssistantOpen: (mode?: string) => void;
  trackAssistantClose: () => void;
  /** Track page views */
  trackPageView: (pageName: string, data?: Record<string, unknown>) => void;
  /** Manually flush events */
  flush: () => void;
  /** Whether tracking is enabled */
  isEnabled: boolean;
  /** Number of pending events */
  pendingCount: number;
}

// ============================================================================
// Hook
// ============================================================================

/**
 * useEventTracking - Hook for event tracking
 *
 * Provides methods for tracking various events. Must be used
 * within a FlyEventTracker provider.
 *
 * @example
 * ```tsx
 * const { track, trackError, trackAssistantAction } = useEventTracking();
 *
 * // Track a custom event
 * track("button_click", { button: "save" });
 *
 * // Track an error
 * trackError(new Error("Failed to save"));
 *
 * // Track assistant action
 * trackAssistantAction("deploy", { functionId: "my-func" });
 * ```
 */
export function useEventTracking(): UseEventTrackingReturn {
  const baseTracking = useBaseEventTracking();

  const {
    track: baseTrack,
    trackError,
    trackDeployment,
    trackTrustChange,
    trackMarketplace,
    trackAssistant,
    flush,
    isEnabled,
    pendingCount,
  } = baseTracking;

  /**
   * Track a custom event
   */
  const track = useCallback(
    (eventName: string, data?: Record<string, unknown>) => {
      baseTrack("custom", { eventName, ...data });
    },
    [baseTrack]
  );

  /**
   * Track marketplace view
   */
  const trackMarketplaceView = useCallback(
    (data: MarketplaceEventData) => {
      trackMarketplace("view", data);
    },
    [trackMarketplace]
  );

  /**
   * Track marketplace install
   */
  const trackMarketplaceInstall = useCallback(
    (data: MarketplaceEventData) => {
      trackMarketplace("install", data);
    },
    [trackMarketplace]
  );

  /**
   * Track marketplace rating
   */
  const trackMarketplaceRate = useCallback(
    (data: MarketplaceEventData) => {
      trackMarketplace("rate", data);
    },
    [trackMarketplace]
  );

  /**
   * Track assistant message
   */
  const trackAssistantMessage = useCallback(
    (data?: AssistantUsageEventData) => {
      trackAssistant("message", data);
    },
    [trackAssistant]
  );

  /**
   * Track assistant action
   */
  const trackAssistantAction = useCallback(
    (actionType: string, data?: Record<string, unknown>) => {
      trackAssistant("action", { actionType, ...data });
    },
    [trackAssistant]
  );

  /**
   * Track assistant open
   */
  const trackAssistantOpen = useCallback(
    (mode?: string) => {
      trackAssistant("open", mode ? { mode } : undefined);
    },
    [trackAssistant]
  );

  /**
   * Track assistant close
   */
  const trackAssistantClose = useCallback(() => {
    trackAssistant("close");
  }, [trackAssistant]);

  /**
   * Track page view
   */
  const trackPageView = useCallback(
    (pageName: string, data?: Record<string, unknown>) => {
      baseTrack("page_view", { pageName, ...data });
    },
    [baseTrack]
  );

  return {
    track,
    trackError,
    trackDeployment,
    trackTrustChange,
    trackMarketplaceView,
    trackMarketplaceInstall,
    trackMarketplaceRate,
    trackAssistantMessage,
    trackAssistantAction,
    trackAssistantOpen,
    trackAssistantClose,
    trackPageView,
    flush,
    isEnabled,
    pendingCount,
  };
}

export default useEventTracking;
