/**
 * FlyAssistantPortal.tsx
 *
 * React Portal component that renders the FlyAssistant UI above all other layers.
 * Creates a dedicated DOM element for z-index isolation and accessibility.
 */

import React, { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

// ============================================================================
// Constants
// ============================================================================

/** DOM element ID for the FlyAssistant portal container */
export const PORTAL_CONTAINER_ID = "fly-assistant-portal";

/** Default z-index for the portal (above most UI elements) */
export const PORTAL_Z_INDEX = 9999;

// ============================================================================
// Types & Interfaces
// ============================================================================

export interface FlyAssistantPortalProps {
  /** Child components to render inside the portal */
  children: React.ReactNode;
  /** Optional custom z-index (defaults to 9999) */
  zIndex?: number;
  /** Optional ARIA label for the portal container */
  ariaLabel?: string;
  /** Optional ARIA role for the portal container */
  ariaRole?: React.AriaRole;
  /** Optional test ID for testing */
  testId?: string;
}

// ============================================================================
// Portal Container Management
// ============================================================================

/**
 * Get or create the portal container element
 * Ensures a single container exists in the DOM
 */
function getOrCreatePortalContainer(): HTMLElement {
  let container = document.getElementById(PORTAL_CONTAINER_ID);

  if (!container) {
    container = document.createElement("div");
    container.id = PORTAL_CONTAINER_ID;
    container.setAttribute("data-fly-assistant", "portal");

    // Ensure container is positioned and has high z-index
    container.style.position = "fixed";
    container.style.top = "0";
    container.style.left = "0";
    container.style.width = "100%";
    container.style.height = "100%";
    container.style.pointerEvents = "none"; // Let clicks pass through by default
    container.style.zIndex = String(PORTAL_Z_INDEX);

    // Append to body
    document.body.appendChild(container);
  }

  return container;
}

/**
 * Remove the portal container if it exists and is empty
 */
function removePortalContainer(): void {
  const container = document.getElementById(PORTAL_CONTAINER_ID);
  if (container && container.childElementCount === 0) {
    document.body.removeChild(container);
  }
}

// ============================================================================
// Portal Component
// ============================================================================

/**
 * FlyAssistantPortal - Renders children into a dedicated DOM element via React Portal
 *
 * This component ensures the FlyAssistant UI is rendered above all other layers
 * with proper z-index isolation. It creates the portal container if it doesn't exist
 * and manages its lifecycle.
 *
 * @example
 * ```tsx
 * <FlyAssistantPortal>
 *   <FlyAssistantPanel isOpen={isOpen}>
 *     <ChatContent />
 *   </FlyAssistantPanel>
 * </FlyAssistantPortal>
 * ```
 */
export function FlyAssistantPortal({
  children,
  zIndex = PORTAL_Z_INDEX,
  ariaLabel = "FlyAssistant AI Assistant",
  ariaRole = "complementary",
  testId = "fly-assistant-portal",
}: FlyAssistantPortalProps) {
  const [container, setContainer] = useState<HTMLElement | null>(null);
  const containerRef = useRef<HTMLElement | null>(null);

  // Initialize container on mount
  useEffect(() => {
    const el = getOrCreatePortalContainer();
    containerRef.current = el;
    setContainer(el);

    // Cleanup on unmount only if no other portals are using it
    return () => {
      // Small delay to allow other portals to unmount first
      setTimeout(() => {
        removePortalContainer();
      }, 0);
    };
  }, []);

  // Update z-index when prop changes
  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.style.zIndex = String(zIndex);
    }
  }, [zIndex]);

  // Don't render until container is ready
  if (!container) {
    return null;
  }

  return createPortal(
    <div
      data-testid={testId}
      role={ariaRole}
      aria-label={ariaLabel}
      className="fly-assistant-portal-content"
      style={{
        width: "100%",
        height: "100%",
        pointerEvents: "none", // Container doesn't capture clicks
      }}
    >
      {children}
    </div>,
    container
  );
}

// ============================================================================
// Helper Hook
// ============================================================================

/**
 * Hook to check if the FlyAssistant portal container exists
 * Useful for determining if FlyAssistant is mounted
 */
export function useFlyAssistantPortal(): {
  isMounted: boolean;
  container: HTMLElement | null;
} {
  const [isMounted, setIsMounted] = useState(false);
  const [container, setContainer] = useState<HTMLElement | null>(null);

  useEffect(() => {
    const el = document.getElementById(PORTAL_CONTAINER_ID);
    setIsMounted(!!el);
    setContainer(el);
  }, []);

  return { isMounted, container };
}

export default FlyAssistantPortal;
