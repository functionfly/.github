import React from "react";

type Position = "tl" | "tr" | "bl" | "br";

interface CornerBraceProps {
  position: Position;
  className?: string;
}

const STYLES: Record<Position, React.CSSProperties> = {
  tl: {
    position: "absolute",
    top: -1,
    left: -1,
    width: 16,
    height: 16,
    borderTop: "2px solid var(--steel-light)",
    borderLeft: "2px solid var(--steel-light)",
    pointerEvents: "none",
  },
  tr: {
    position: "absolute",
    top: -1,
    right: -1,
    width: 16,
    height: 16,
    borderTop: "2px solid var(--steel-light)",
    borderRight: "2px solid var(--steel-light)",
    pointerEvents: "none",
  },
  bl: {
    position: "absolute",
    bottom: -1,
    left: -1,
    width: 16,
    height: 16,
    borderBottom: "2px solid var(--steel-light)",
    borderLeft: "2px solid var(--steel-light)",
    pointerEvents: "none",
  },
  br: {
    position: "absolute",
    bottom: -1,
    right: -1,
    width: 16,
    height: 16,
    borderBottom: "2px solid var(--steel-light)",
    borderRight: "2px solid var(--steel-light)",
    pointerEvents: "none",
  },
};

export const CornerBrace: React.FC<CornerBraceProps> = ({
  position,
  className = "",
}) => {
  return (
    <div
      className={className}
      style={STYLES[position]}
      aria-hidden="true"
    />
  );
};
