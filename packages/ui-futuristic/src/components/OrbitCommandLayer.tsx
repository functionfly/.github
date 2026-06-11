/**
 * @functionfly/ui-futuristic
 * OrbitCommandLayer - Orbiting command palette with radial navigation
 */

import React, { useState, useEffect, useRef } from "react";
import { cn } from "@functionfly/ui-core";
import {
  Orbit,
  Circle,
  Play,
  Eye,
  Layers,
  Network,
  FileText,
  Activity,
} from "lucide-react";
import { Rocket, Bug, Shield, Users, CreditCard, ShieldCheck } from "./icons";
import { Settings } from "lucide-react";
import type {
  OrbitCommandLayerProps,
  OrbitalLayer,
  OrbitalItem,
} from "../types";

export const OrbitCommandLayer: React.FC<OrbitCommandLayerProps> = ({
  layers = [
    {
      radius: 60,
      speed: 0.5,
      items: [
        { id: "1", label: "Execute", angle: 0 },
        { id: "2", label: "Monitor", angle: 72 },
        { id: "3", label: "Deploy", angle: 144 },
        { id: "4", label: "Scale", angle: 216 },
        { id: "5", label: "Debug", angle: 288 },
      ],
    },
    {
      radius: 100,
      speed: 0.3,
      items: [
        { id: "6", label: "API", angle: 30 },
        { id: "7", label: "Logs", angle: 90 },
        { id: "8", label: "Metrics", angle: 150 },
        { id: "9", label: "Config", angle: 210 },
        { id: "10", label: "Secrets", angle: 270 },
        { id: "11", label: "Domains", angle: 330 },
      ],
    },
    {
      radius: 140,
      speed: 0.15,
      items: [
        { id: "12", label: "Users", angle: 0 },
        { id: "13", label: "Billing", angle: 60 },
        { id: "14", label: "Settings", angle: 120 },
        { id: "15", label: "Audit", angle: 180 },
        { id: "16", label: "Help", angle: 240 },
        { id: "17", label: "Docs", angle: 300 },
      ],
    },
  ],
  activeItemId = null,
  centerLabel = "CMD",
  isOpen = true,
  onItemSelect,
  onToggle,
  className,
}) => {
  const [rotation, setRotation] = useState(0);
  const [hoveredItem, setHoveredItem] = useState<string | null>(null);
  const animationRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    if (!isOpen) return;

    const animate = (_timestamp?: number) => {
      setRotation((prev) => (prev + 0.2) % 360);
      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isOpen]);

  const getItemIcon = (label: string): React.ReactNode => {
    const icons: Record<string, React.ReactNode> = {
      Execute: <Play className="w-3 h-3" />,
      Monitor: <Eye className="w-3 h-3" />,
      Deploy: <Rocket className="w-3 h-3" />,
      Scale: <Layers className="w-3 h-3" />,
      Debug: <Bug className="w-3 h-3" />,
      API: <Network className="w-3 h-3" />,
      Logs: <FileText className="w-3 h-3" />,
      Metrics: <Activity className="w-3 h-3" />,
      Config: <Settings className="w-3 h-3" />,
      Secrets: <Shield className="w-3 h-3" />,
      Users: <Users className="w-3 h-3" />,
      Billing: <CreditCard className="w-3 h-3" />,
      Settings: <Settings className="w-3 h-3" />,
      Audit: <ShieldCheck className="w-3 h-3" />,
    };
    return icons[label] || <Circle className="w-3 h-3" />;
  };

  return (
    <div
      className={cn(
        "relative flex items-center justify-center",
        isOpen ? "w-80 h-80" : "w-16 h-16",
        "transition-all duration-500 ease-out",
        className,
      )}
    >
      {/* Outer glow ring */}
      <div className="absolute inset-0 rounded-full bg-gradient-to-r from-cyan-500/20 via-purple-500/20 to-cyan-500/20 blur-xl animate-spin-slow" />

      {/* Orbital rings */}
      {isOpen &&
        layers.map((layer: OrbitalLayer, layerIndex: number) => (
          <div
            key={layerIndex}
            className="absolute rounded-full border border-cyan-500/30"
            style={{
              width: layer.radius * 2 + 40,
              height: layer.radius * 2 + 40,
              animation: `pulse-ring ${2 + layerIndex * 0.5}s ease-in-out infinite`,
            }}
          >
            {/* Items on this layer */}
            {layer.items.map((item: OrbitalItem) => {
              const angleRad =
                ((item.angle + rotation * layer.speed) * Math.PI) / 180;
              const x = Math.cos(angleRad) * layer.radius;
              const y = Math.sin(angleRad) * layer.radius;
              const isHovered = hoveredItem === item.id;
              const isActive = activeItemId === item.id;

              return (
                <div
                  key={item.id}
                  className={cn(
                    "absolute flex flex-col items-center justify-center cursor-pointer transition-all duration-300",
                    "group",
                    isActive && "z-10",
                  )}
                  style={{
                    left: "50%",
                    top: "50%",
                    transform: `translate(${x}px, ${y}px) translate(-50%, -50%)`,
                  }}
                  onClick={() => onItemSelect?.(item, layer)}
                  onMouseEnter={() => setHoveredItem(item.id)}
                  onMouseLeave={() => setHoveredItem(null)}
                >
                  {/* Item node */}
                  <div
                    className={cn(
                      "relative flex items-center justify-center w-10 h-10 rounded-full",
                      "bg-gradient-to-br from-slate-900/90 to-slate-800/90",
                      "border border-cyan-500/50",
                      "shadow-[0_0_15px_rgba(6,182,212,0.3)]",
                      "transition-all duration-300",
                      isHovered &&
                        "scale-125 bg-gradient-to-br from-cyan-500/30 to-purple-500/30",
                      isActive &&
                        "bg-cyan-500/40 border-cyan-400 shadow-[0_0_25px_rgba(6,182,212,0.6)]",
                    )}
                  >
                    <span
                      className={cn(
                        "text-cyan-400 transition-colors",
                        isHovered && "text-white",
                        isActive && "text-cyan-200",
                      )}
                    >
                      {getItemIcon(item.label)}
                    </span>

                    {/* Glow effect */}
                    {isHovered && (
                      <div className="absolute inset-0 rounded-full bg-cyan-500/20 animate-ping" />
                    )}
                  </div>

                  {/* Label */}
                  <div
                    className={cn(
                      "absolute top-full mt-2 px-2 py-1 rounded-md",
                      "bg-slate-900/95 border border-cyan-500/30",
                      "text-[10px] text-cyan-300 font-medium whitespace-nowrap",
                      "opacity-0 group-hover:opacity-100 transition-opacity duration-200",
                      "shadow-lg shadow-cyan-500/20",
                    )}
                  >
                    {item.label}
                  </div>
                </div>
              );
            })}
          </div>
        ))}

      {/* Center command button */}
      <button
        onClick={onToggle}
        className={cn(
          "relative z-20 flex items-center justify-center rounded-full",
          "bg-gradient-to-br from-cyan-500/20 to-purple-500/20",
          "border-2 border-cyan-500/60",
          "shadow-[0_0_30px_rgba(6,182,212,0.4),inset_0_0_20px_rgba(6,182,212,0.2)]",
          "transition-all duration-300",
          "hover:shadow-[0_0_40px_rgba(6,182,212,0.6),inset_0_0_30px_rgba(6,182,212,0.3)]",
          "hover:border-cyan-400 hover:scale-110",
          isOpen ? "w-16 h-16" : "w-16 h-16",
        )}
      >
        <div className="relative">
          <Orbit
            className={cn(
              "w-6 h-6 text-cyan-400 transition-transform",
              isOpen && "rotate-180",
            )}
          />
          <div className="absolute inset-0 rounded-full bg-cyan-400/20 animate-pulse" />
        </div>

        {/* Center label */}
        <span className="absolute -bottom-6 text-[10px] text-cyan-300/80 font-mono">
          {centerLabel}
        </span>
      </button>

      {/* Decorative particles */}
      {isOpen &&
        Array.from({ length: 12 }).map((_, i) => (
          <div
            key={i}
            className="absolute w-1 h-1 rounded-full bg-cyan-400/60"
            style={{
              left: "50%",
              top: "50%",
              animation: `orbit-particle ${3 + i * 0.3}s linear infinite`,
              animationDelay: `${i * 0.2}s`,
            }}
          />
        ))}

      <style>{`
        @keyframes pulse-ring {
          0%, 100% { opacity: 0.3; transform: scale(1); }
          50% { opacity: 0.6; transform: scale(1.02); }
        }
        @keyframes orbit-particle {
          0% { transform: rotate(0deg) translateX(160px) rotate(0deg); opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { transform: rotate(360deg) translateX(160px) rotate(-360deg); opacity: 0; }
        }
        @keyframes spin-slow {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        .animate-spin-slow { animation: spin-slow 20s linear infinite; }
      `}</style>
    </div>
  );
};
