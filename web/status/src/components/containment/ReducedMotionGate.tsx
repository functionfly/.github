import React, { createContext, useContext, useState, useEffect } from 'react';

interface ReducedMotionContextValue {
  prefersReducedMotion: boolean;
}

const ReducedMotionContext = createContext<ReducedMotionContextValue>({
  prefersReducedMotion: false,
});

export function useReducedMotion(): boolean {
  const context = useContext(ReducedMotionContext);
  return context.prefersReducedMotion;
}

interface ReducedMotionGateProps {
  children: React.ReactNode;
}

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
