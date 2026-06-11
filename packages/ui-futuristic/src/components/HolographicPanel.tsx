/**
 * @functionfly/ui-futuristic
 * HolographicPanel - Holographic display effect panel
 */

import React, { useState, useEffect } from "react";
import { cn } from "@functionfly/ui-core";
import { Hexagon } from "lucide-react";
import type { HolographicPanelProps, HolographicEffect } from "../types";

export const HolographicPanel: React.FC<HolographicPanelProps> = ({
  effect = "cyan",
  intensity = 0.8,
  children,
  className,
}) => {
  const [scanlineOffset, setScanlineOffset] = useState(0);
  const [glitchOffset, setGlitchOffset] = useState({ x: 0, y: 0 });

  useEffect(() => {
    const interval = setInterval(() => {
      setScanlineOffset((prev) => (prev + 1) % 4);
      if (Math.random() > 0.95) {
        setGlitchOffset({
          x: (Math.random() - 0.5) * 4,
          y: (Math.random() - 0.5) * 4,
        });
      } else {
        setGlitchOffset({ x: 0, y: 0 });
      }
    }, 50);
    return () => clearInterval(interval);
  }, []);

  const getEffectColors = (): {
    primary: string;
    glow: string;
    scanline: string;
    border: string;
  } => {
    switch (effect) {
      case "rainbow":
        return {
          primary: "from-purple-500 via-cyan-500 to-pink-500",
          glow: "rgba(6, 182, 212, 0.4)",
          scanline: "rgba(6, 182, 212, 0.1)",
          border: "cyan-500",
        };
      case "cyan":
        return {
          primary: "from-cyan-500/20 to-cyan-600/20",
          glow: "rgba(6, 182, 212, 0.4)",
          scanline: "rgba(6, 182, 212, 0.1)",
          border: "cyan-400",
        };
      case "magenta":
        return {
          primary: "from-pink-500/20 to-purple-600/20",
          glow: "rgba(236, 72, 153, 0.4)",
          scanline: "rgba(236, 72, 153, 0.1)",
          border: "pink-400",
        };
      case "white":
        return {
          primary: "from-slate-200/20 to-slate-400/20",
          glow: "rgba(255, 255, 255, 0.3)",
          scanline: "rgba(255, 255, 255, 0.1)",
          border: "slate-300",
        };
    }
  };

  const colors = getEffectColors();

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-xl",
        "backdrop-blur-md bg-slate-900/60",
        "border border-" + colors.border + "/50",
        "shadow-[" + colors.glow + "]",
        className,
      )}
      style={{
        transform: `translate(${glitchOffset.x}px, ${glitchOffset.y}px)`,
        opacity: intensity,
      }}
    >
      {/* Scanline effect */}
      <div
        className="absolute inset-0 pointer-events-none z-10 overflow-hidden"
        style={{
          background: `repeating-linear-gradient(
            0deg,
            transparent,
            transparent 2px,
            ${colors.scanline} 2px,
            ${colors.scanline} 4px
          )`,
          backgroundPosition: `0 ${scanlineOffset * 4}px`,
        }}
      />

      {/* Holographic shimmer */}
      <div
        className={cn(
          "absolute inset-0 pointer-events-none z-20",
          "bg-gradient-to-br from-white/5 via-transparent to-transparent",
        )}
      />

      {/* Content */}
      <div className="relative z-30 p-6">
        {children || (
          <div className="flex flex-col items-center gap-3">
            <Hexagon
              className={cn("w-12 h-12 text-" + colors.border + "-400")}
            />
            <span className={cn("text-sm text-" + colors.border + "-300")}>
              HOLOGRAPHIC DISPLAY ACTIVE
            </span>
          </div>
        )}
      </div>

      {/* Bottom scan line */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/30 to-transparent" />

      {/* Corner accents */}
      <div
        className={cn(
          "absolute top-2 left-2 w-4 h-4 border-t-2 border-l-2 border-" +
            colors.border +
            "-400/50 rounded-tl",
        )}
      />
      <div
        className={cn(
          "absolute top-2 right-2 w-4 h-4 border-t-2 border-r-2 border-" +
            colors.border +
            "-400/50 rounded-tr",
        )}
      />
      <div
        className={cn(
          "absolute bottom-2 left-2 w-4 h-4 border-b-2 border-l-2 border-" +
            colors.border +
            "-400/50 rounded-bl",
        )}
      />
      <div
        className={cn(
          "absolute bottom-2 right-2 w-4 h-4 border-b-2 border-r-2 border-" +
            colors.border +
            "-400/50 rounded-br",
        )}
      />
    </div>
  );
};
