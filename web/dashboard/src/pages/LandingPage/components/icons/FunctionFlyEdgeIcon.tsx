import React from "react";

/**
 * FunctionFly Edge – custom icon for our managed edge provider.
 * Combines a node (circle) with a forward chevron to suggest "edge" and speed.
 * Brand color: #6366f1 (indigo).
 */
export const FunctionFlyEdgeIcon: React.FC<{ className?: string }> = ({
  className = "w-8 h-8",
}) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="FunctionFly Edge"
  >
    <title>FunctionFly Edge</title>
    <defs>
      <linearGradient
        id="functionfly-edge-gradient"
        x1="0%"
        y1="0%"
        x2="100%"
        y2="100%"
      >
        <stop offset="0%" stopColor="#6366f1" />
        <stop offset="100%" stopColor="#8b5cf6" />
      </linearGradient>
    </defs>
    {/* Outer ring (edge node) */}
    <circle
      cx="12"
      cy="12"
      r="10"
      stroke="url(#functionfly-edge-gradient)"
      strokeWidth="2"
      fill="none"
    />
    {/* Forward chevron (edge / direction) */}
    <path
      d="M10 8l4 4-4 4"
      stroke="url(#functionfly-edge-gradient)"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);
