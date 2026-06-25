import type { ReactNode, ReactElement } from 'react';
import { useState } from 'react';

interface TooltipProps {
  children: ReactElement;
  content: ReactNode;
}

export function Tooltip({ children, content }: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [timeoutId, setTimeoutId] = useState<ReturnType<typeof setTimeout> | null>(null);

  const show = () => {
    if (timeoutId) clearTimeout(timeoutId);
    const id = setTimeout(() => setIsVisible(true), 400);
    setTimeoutId(id);
  };

  const hide = () => {
    setIsVisible(false);
    if (timeoutId) {
      clearTimeout(timeoutId);
      setTimeoutId(null);
    }
  };

  return (
    <div
      className="relative inline-flex"
      onMouseEnter={show}
      onMouseLeave={hide}
    >
      {children}
      {isVisible && (
        <div className="tooltip">
          {content}
          <div className="tooltip__caret" />
        </div>
      )}
    </div>
  );
}
