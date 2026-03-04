/**
 * FlyAssistantProvider.tsx
 *
 * Global context provider that manages FlyAssistant state and user session integration.
 * Uses a hybrid React Context + Zustand store pattern for optimal performance.
 */

import React, { createContext, useContext, useCallback, useMemo } from "react";
import { createStore, useStore } from "zustand";
import { subscribeWithSelector } from "zustand/middleware";

// ============================================================================
// Types & Interfaces
// ============================================================================

/**
 * User role tier for feature gating
 */
export type UserRole = "free" | "pro" | "enterprise";

/**
 * Trust tier based on confidence score
 */
export type TrustTier = "low" | "medium" | "high" | "critical";

/**
 * Current route/page context for contextual assistance
 */
export interface RouteContext {
  path: string;
  name: string;
  params?: Record<string, string>;
}

/**
 * Cached conversation entry for memory persistence
 */
export interface ConversationEntry {
  id: string;
  timestamp: number;
  messages: Message[];
  context?: RouteContext;
}

/**
 * Individual message in a conversation
 */
export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: number;
}

/**
 * User session state
 */
export interface UserSession {
  id: string;
  email: string;
  role: UserRole;
  orgId?: string;
}

/**
 * Complete FlyAssistant state
 */
export interface FlyAssistantState {
  // UI State
  isOpen: boolean;
  isMinimized: boolean;
  isFullscreen: boolean;

  // User Context
  userSession: UserSession | null;
  currentRoute: RouteContext | null;

  // Trust & Quality
  trustScore: number;
  trustTier: TrustTier;

  // Error State
  hasError: boolean;
  errorMessage: string | null;

  // Insights & Notifications
  hasInsights: boolean;
  notificationCount: number;

  // Memory Cache
  conversationCache: Map<string, ConversationEntry>;
  currentConversationId: string | null;

  // Actions
  open: () => void;
  close: () => void;
  toggle: () => void;
  minimize: () => void;
  expand: () => void;
  setFullscreen: (value: boolean) => void;
  setUserSession: (session: UserSession | null) => void;
  setCurrentRoute: (route: RouteContext | null) => void;
  setTrustScore: (score: number) => void;
  setError: (error: string | null) => void;
  setHasInsights: (value: boolean) => void;
  setNotificationCount: (count: number) => void;
  addToCache: (entry: ConversationEntry) => void;
  clearCache: () => void;
  setCurrentConversation: (id: string | null) => void;
}

// ============================================================================
// Store Creation
// ============================================================================

/**
 * Calculate trust tier from score
 */
function calculateTrustTier(score: number): TrustTier {
  if (score >= 0.9) return "critical";
  if (score >= 0.7) return "high";
  if (score >= 0.4) return "medium";
  return "low";
}

/**
 * Create the Zustand store with all state and actions
 */
function createFlyAssistantStore() {
  return createStore(
    subscribeWithSelector<FlyAssistantState>((set, get) => ({
      // Initial UI State
      isOpen: false,
      isMinimized: false,
      isFullscreen: false,

      // Initial User Context
      userSession: null,
      currentRoute: null,

      // Initial Trust & Quality
      trustScore: 0.5,
      trustTier: "medium",

      // Initial Error State
      hasError: false,
      errorMessage: null,

      // Initial Insights
      hasInsights: false,
      notificationCount: 0,

      // Initial Memory Cache
      conversationCache: new Map(),
      currentConversationId: null,

      // Actions
      open: () => set({ isOpen: true, isMinimized: false }),

      close: () => set({
        isOpen: false,
        isMinimized: false,
        isFullscreen: false
      }),

      toggle: () => set((state) => ({
        isOpen: !state.isOpen,
        isMinimized: false
      })),

      minimize: () => set({ isMinimized: true }),

      expand: () => set({ isMinimized: false }),

      setFullscreen: (value: boolean) => set({ isFullscreen: value }),

      setUserSession: (session: UserSession | null) => set({ userSession: session }),

      setCurrentRoute: (route: RouteContext | null) => set({ currentRoute: route }),

      setTrustScore: (score: number) => set({
        trustScore: Math.max(0, Math.min(1, score)),
        trustTier: calculateTrustTier(score)
      }),

      setError: (error: string | null) => set({
        hasError: error !== null,
        errorMessage: error
      }),

      setHasInsights: (value: boolean) => set({ hasInsights: value }),

      setNotificationCount: (count: number) => set({
        notificationCount: Math.max(0, count)
      }),

      addToCache: (entry: ConversationEntry) => set((state) => {
        const newCache = new Map(state.conversationCache);
        newCache.set(entry.id, entry);
        // Keep only last 50 conversations
        if (newCache.size > 50) {
          const firstKey = newCache.keys().next().value;
          newCache.delete(firstKey);
        }
        return { conversationCache: newCache };
      }),

      clearCache: () => set({
        conversationCache: new Map(),
        currentConversationId: null
      }),

      setCurrentConversation: (id: string | null) => set({
        currentConversationId: id
      }),
    }))
  );
}

// ============================================================================
// Context Setup
// ============================================================================

type FlyAssistantStore = ReturnType<typeof createFlyAssistantStore>;

interface FlyAssistantContextValue {
  store: FlyAssistantStore;
}

const FlyAssistantContext = createContext<FlyAssistantContextValue | null>(null);

// ============================================================================
// Hook Exports
// ============================================================================

/**
 * Hook to access the entire FlyAssistant store
 * @throws Error if used outside of FlyAssistantProvider
 */
export function useFlyAssistantStore(): FlyAssistantStore {
  const context = useContext(FlyAssistantContext);
  if (!context) {
    throw new Error(
      "useFlyAssistantStore must be used within a FlyAssistantProvider"
    );
  }
  return context.store;
}

/**
 * Hook to select specific state from the FlyAssistant store
 * Uses Zustand's selector pattern for optimal re-renders
 *
 * @example
 * const isOpen = useFlyAssistant((state) => state.isOpen);
 * const { open, close } = useFlyAssistant((state) => ({ open: state.open, close: state.close }));
 */
export function useFlyAssistant<T>(selector: (state: FlyAssistantState) => T): T {
  const store = useFlyAssistantStore();
  return useStore(store, selector);
}

/**
 * Hook for FlyAssistant UI actions
 */
export function useFlyAssistantActions() {
  const store = useFlyAssistantStore();
  return useStore(store, (state) => ({
    open: state.open,
    close: state.close,
    toggle: state.toggle,
    minimize: state.minimize,
    expand: state.expand,
    setFullscreen: state.setFullscreen,
  }));
}

/**
 * Hook for FlyAssistant user/session state
 */
export function useFlyAssistantUser() {
  const store = useFlyAssistantStore();
  return useStore(store, (state) => ({
    userSession: state.userSession,
    currentRoute: state.currentRoute,
    setUserSession: state.setUserSession,
    setCurrentRoute: state.setCurrentRoute,
  }));
}

/**
 * Hook for FlyAssistant trust/error state
 */
export function useFlyAssistantStatus() {
  const store = useFlyAssistantStore();
  return useStore(store, (state) => ({
    trustScore: state.trustScore,
    trustTier: state.trustTier,
    hasError: state.hasError,
    errorMessage: state.errorMessage,
    hasInsights: state.hasInsights,
    notificationCount: state.notificationCount,
    setTrustScore: state.setTrustScore,
    setError: state.setError,
    setHasInsights: state.setHasInsights,
    setNotificationCount: state.setNotificationCount,
  }));
}

/**
 * Hook for FlyAssistant conversation cache
 */
export function useFlyAssistantCache() {
  const store = useFlyAssistantStore();
  return useStore(store, (state) => ({
    conversationCache: state.conversationCache,
    currentConversationId: state.currentConversationId,
    addToCache: state.addToCache,
    clearCache: state.clearCache,
    setCurrentConversation: state.setCurrentConversation,
  }));
}

// ============================================================================
// Provider Component
// ============================================================================

export interface FlyAssistantProviderProps {
  /** Child components */
  children: React.ReactNode;
  /** Initial user session (optional) */
  initialSession?: UserSession | null;
  /** Initial route context (optional) */
  initialRoute?: RouteContext | null;
}

/**
 * FlyAssistantProvider - Global context provider for FlyAssistant
 *
 * Wraps the application to provide FlyAssistant state and actions
 * to all child components. Uses Zustand for state management with
 * React Context for store instance sharing.
 *
 * @example
 * ```tsx
 * <FlyAssistantProvider initialSession={user}>
 *   <App />
 * </FlyAssistantProvider>
 * ```
 */
export function FlyAssistantProvider({
  children,
  initialSession = null,
  initialRoute = null,
}: FlyAssistantProviderProps) {
  const storeRef = React.useRef<FlyAssistantStore | null>(null);

  // Create store instance only once
  if (!storeRef.current) {
    storeRef.current = createFlyAssistantStore();

    // Set initial values
    if (initialSession) {
      storeRef.current.getState().setUserSession(initialSession);
    }
    if (initialRoute) {
      storeRef.current.getState().setCurrentRoute(initialRoute);
    }
  }

  const contextValue = useMemo(
    () => ({ store: storeRef.current! }),
    []
  );

  return (
    <FlyAssistantContext.Provider value={contextValue}>
      {children}
    </FlyAssistantContext.Provider>
  );
}

export default FlyAssistantProvider;
