import React, { useCallback, useRef, useState } from "react";

interface TooltipProps {
  content: string;
  children: React.ReactElement;
  delay?: number;
}

export const Tooltip: React.FC<TooltipProps> = ({
  content,
  children,
  delay = 400,
}) => {
  const [visible, setVisible] = useState(false);
  const showTimer = useRef<number>(0);

  const show = useCallback(() => {
    showTimer.current = window.setTimeout(() => setVisible(true), delay);
  }, [delay]);

  const hide = useCallback(() => {
    clearTimeout(showTimer.current);
    setVisible(false);
  }, []);

  return (
    <div
      style={{ position: "relative", display: "inline-flex" }}
      onMouseEnter={show}
      onMouseLeave={hide}
      onFocus={show}
      onBlur={hide}
    >
      {children}
      {visible && (
        <div
          role="tooltip"
          style={{
            position: "absolute",
            bottom: "100%",
            left: "50%",
            transform: "translateX(-50%)",
            marginBottom: "6px",
            background: "var(--panel-raised)",
            border: "1px solid var(--steel)",
            borderRadius: "var(--radius-sm)",
            padding: "var(--space-2) var(--space-3)",
            fontFamily: "var(--font-body)",
            fontSize: "13px",
            lineHeight: 1.5,
            color: "var(--text-dim)",
            whiteSpace: "nowrap",
            zIndex: "var(--z-toast)",
            animation: "tooltipFadeIn var(--duration-fast) var(--ease-out)",
            pointerEvents: "none",
          }}
        >
          {content}
          {/* Caret */}
          <div
            style={{
              position: "absolute",
              top: "100%",
              left: "50%",
              transform: "translateX(-50%)",
              width: 0,
              height: 0,
              borderLeft: "6px solid transparent",
              borderRight: "6px solid transparent",
              borderTop: "6px solid var(--steel)",
            }}
          />
        </div>
      )}
    </div>
  );
};
