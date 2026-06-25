import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';

interface ReducedMotionContextValue {
  prefersReducedMotion: boolean;
}

const ReducedMotionContext = createContext<ReducedMotionContextValue>({
  prefersReducedMotion: false,
});

/**
 * Hook to check if user prefers reduced motion.
 */
export function useReducedMotion(): boolean {
  const context = useContext(ReducedMotionContext);
  return context.prefersReducedMotion;
}

interface ReducedMotionGateProps {
  children: React.ReactNode;
}

/**
 * Provider component that detects user's prefers-reduced-motion preference
 * and provides it to children via context.
 * 
 * Children can use the `useReducedMotion` hook to adapt their behavior.
 * 
 * Usage:
 * ```tsx
 * <ReducedMotionGate>
 *   <TrustSeal />
 * </ReducedMotionGate>
 * ```
 * 
 * Or in a child component:
 * ```tsx
 * const prefersReducedMotion = useReducedMotion();
 * ```
 */
export const ReducedMotionGate: React.FC<ReducedMotionGateProps> = ({
  children,
}) => {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(mq.matches);
    
    const handler = (e: MediaQueryListEvent) => setPrefersReducedMotion(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  return (
    <ReducedMotionContext.Provider value={{ prefersReducedMotion }}>
      {children}
    </ReducedMotionContext.Provider>
  );
};
