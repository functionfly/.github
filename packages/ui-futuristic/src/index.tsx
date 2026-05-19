/**
 * @functionfly/ui-futuristic
 * Futuristic Signature Components - Branding-Level UI
 */

import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { cn } from '@functionfly/ui-core';
import {
  Orbit,
  Sparkles,
  Hexagon,
  Circle,
  Diamond,
  Square,
  Triangle,
  Zap,
  Brain,
  Cpu,
  Database,
  Network,
  Bot,
  Star,
  ChevronRight,
  ChevronLeft,
  Play,
  Pause,
  RefreshCw,
  Activity,
  TrendingUp,
  TrendingDown,
  Minus,
  Info,
  X,
  Maximize2,
  Minimize2,
  Eye,
  Layers,
  Radio,
  Waves,
  Atom,
  FluxCapacitor,
  Binary,
  Code2,
  FileText,
  MessageSquare,
  Lightbulb,
  Target,
  Crosshair,
} from 'lucide-react';

// ============================================================================
// OrbitCommandLayer - Orbiting command palette with radial navigation
// ============================================================================

export const OrbitCommandLayer: React.FC<OrbitCommandLayerProps> = ({
  layers = [
    { radius: 60, speed: 0.5, items: [
      { id: '1', label: 'Execute', angle: 0 },
      { id: '2', label: 'Monitor', angle: 72 },
      { id: '3', label: 'Deploy', angle: 144 },
      { id: '4', label: 'Scale', angle: 216 },
      { id: '5', label: 'Debug', angle: 288 },
    ]},
    { radius: 100, speed: 0.3, items: [
      { id: '6', label: 'API', angle: 30 },
      { id: '7', label: 'Logs', angle: 90 },
      { id: '8', label: 'Metrics', angle: 150 },
      { id: '9', label: 'Config', angle: 210 },
      { id: '10', label: 'Secrets', angle: 270 },
      { id: '11', label: 'Domains', angle: 330 },
    ]},
    { radius: 140, speed: 0.15, items: [
      { id: '12', label: 'Users', angle: 0 },
      { id: '13', label: 'Billing', angle: 60 },
      { id: '14', label: 'Settings', angle: 120 },
      { id: '15', label: 'Audit', angle: 180 },
      { id: '16', label: 'Help', angle: 240 },
      { id: '17', label: 'Docs', angle: 300 },
    ]},
  ],
  activeItemId = null,
  centerLabel = 'CMD',
  isOpen = true,
  onItemSelect,
  onToggle,
  className,
}) => {
  const [rotation, setRotation] = useState(0);
  const [hoveredItem, setHoveredItem] = useState<string | null>(null);
  const [selectedLayer, setSelectedLayer] = useState<number>(0);
  const animationRef = useRef<number>();

  useEffect(() => {
    if (!isOpen) return;
    
    const animate = () => {
      setRotation(prev => (prev + 0.2) % 360);
      animationRef.current = requestAnimationFrame(animate);
    };
    
    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isOpen]);

  const getItemIcon = (label: string) => {
    const icons: Record<string, React.ReactNode> = {
      'Execute': <Play className="w-3 h-3" />,
      'Monitor': <Eye className="w-3 h-3" />,
      'Deploy': <Rocket className="w-3 h-3" />,
      'Scale': <Layers className="w-3 h-3" />,
      'Debug': <Bug className="w-3 h-3" />,
      'API': <Network className="w-3 h-3" />,
      'Logs': <FileText className="w-3 h-3" />,
      'Metrics': <Activity className="w-3 h-3" />,
      'Config': <Settings className="w-3 h-3" />,
      'Secrets': <Shield className="w-3 h-3" />,
      'Users': <Users className="w-3 h-3" />,
      'Billing': <CreditCard className="w-3 h-3" />,
      'Settings': <Settings className="w-3 h-3" />,
      'Audit': <ShieldCheck className="w-3 h-3" />,
    };
    return icons[label] || <Circle className="w-3 h-3" />;
  };

  return (
    <div className={cn(
      'relative flex items-center justify-center',
      isOpen ? 'w-80 h-80' : 'w-16 h-16',
      'transition-all duration-500 ease-out',
      className
    )}>
      {/* Outer glow ring */}
      <div className="absolute inset-0 rounded-full bg-gradient-to-r from-cyan-500/20 via-purple-500/20 to-cyan-500/20 blur-xl animate-spin-slow" />
      
      {/* Orbital rings */}
      {isOpen && layers.map((layer, layerIndex) => (
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
          {layer.items.map((item) => {
            const angleRad = ((item.angle + rotation * layer.speed) * Math.PI) / 180;
            const x = Math.cos(angleRad) * layer.radius;
            const y = Math.sin(angleRad) * layer.radius;
            const isHovered = hoveredItem === item.id;
            const isActive = activeItemId === item.id;

            return (
              <div
                key={item.id}
                className={cn(
                  'absolute flex flex-col items-center justify-center cursor-pointer transition-all duration-300',
                  'group',
                  isActive && 'z-10'
                )}
                style={{
                  left: '50%',
                  top: '50%',
                  transform: `translate(${x}px, ${y}px) translate(-50%, -50%)`,
                }}
                onClick={() => onItemSelect?.(item, layer)}
                onMouseEnter={() => setHoveredItem(item.id)}
                onMouseLeave={() => setHoveredItem(null)}
              >
                {/* Item node */}
                <div className={cn(
                  'relative flex items-center justify-center w-10 h-10 rounded-full',
                  'bg-gradient-to-br from-slate-900/90 to-slate-800/90',
                  'border border-cyan-500/50',
                  'shadow-[0_0_15px_rgba(6,182,212,0.3)]',
                  'transition-all duration-300',
                  isHovered && 'scale-125 bg-gradient-to-br from-cyan-500/30 to-purple-500/30',
                  isActive && 'bg-cyan-500/40 border-cyan-400 shadow-[0_0_25px_rgba(6,182,212,0.6)]',
                )}>
                  <span className={cn(
                    'text-cyan-400 transition-colors',
                    isHovered && 'text-white',
                    isActive && 'text-cyan-200'
                  )}>
                    {getItemIcon(item.label)}
                  </span>
                  
                  {/* Glow effect */}
                  {isHovered && (
                    <div className="absolute inset-0 rounded-full bg-cyan-500/20 animate-ping" />
                  )}
                </div>
                
                {/* Label */}
                <div className={cn(
                  'absolute top-full mt-2 px-2 py-1 rounded-md',
                  'bg-slate-900/95 border border-cyan-500/30',
                  'text-[10px] text-cyan-300 font-medium whitespace-nowrap',
                  'opacity-0 group-hover:opacity-100 transition-opacity duration-200',
                  'shadow-lg shadow-cyan-500/20'
                )}>
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
          'relative z-20 flex items-center justify-center rounded-full',
          'bg-gradient-to-br from-cyan-500/20 to-purple-500/20',
          'border-2 border-cyan-500/60',
          'shadow-[0_0_30px_rgba(6,182,212,0.4),inset_0_0_20px_rgba(6,182,212,0.2)]',
          'transition-all duration-300',
          'hover:shadow-[0_0_40px_rgba(6,182,212,0.6),inset_0_0_30px_rgba(6,182,212,0.3)]',
          'hover:border-cyan-400 hover:scale-110',
          isOpen ? 'w-16 h-16' : 'w-16 h-16'
        )}
      >
        <div className="relative">
          <Orbit className={cn('w-6 h-6 text-cyan-400 transition-transform', isOpen && 'rotate-180')} />
          <div className="absolute inset-0 rounded-full bg-cyan-400/20 animate-pulse" />
        </div>
        
        {/* Center label */}
        <span className="absolute -bottom-6 text-[10px] text-cyan-300/80 font-mono">
          {centerLabel}
        </span>
      </button>

      {/* Decorative particles */}
      {isOpen && Array.from({ length: 12 }).map((_, i) => (
        <div
          key={i}
          className="absolute w-1 h-1 rounded-full bg-cyan-400/60"
          style={{
            left: '50%',
            top: '50%',
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

// ============================================================================
// QuantumWorkspaceTransition - Quantum-inspired workspace switching
// ============================================================================

export const QuantumWorkspaceTransition: React.FC<QuantumWorkspaceTransitionProps> = ({
  phase = 'collapse',
  fromWorkspace = 'Workspace A',
  toWorkspace = 'Workspace B',
  progress = 0,
  onPhaseComplete,
  className,
}) => {
  const [localPhase, setLocalPhase] = useState<TransitionPhase>(phase);
  const [particlePositions, setParticlePositions] = useState<Array<{ x: number; y: number; vx: number; vy: number }>>([]);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationRef = useRef<number>();

  useEffect(() => {
    setLocalPhase(phase);
  }, [phase]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const particles = Array.from({ length: 50 }, () => ({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      vx: (Math.random() - 0.5) * 4,
      vy: (Math.random() - 0.5) * 4,
      size: Math.random() * 3 + 1,
    }));

    const animate = () => {
      ctx.fillStyle = 'rgba(6, 16, 32, 0.2)';
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      particles.forEach(p => {
        p.x += p.vx;
        p.y += p.vy;

        if (p.x < 0 || p.x > canvas.width) p.vx *= -1;
        if (p.y < 0 || p.y > canvas.height) p.vy *= -1;

        ctx.beginPath();
        ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(6, 182, 212, ${0.3 + Math.random() * 0.5})`;
        ctx.fill();
      });

      animationRef.current = requestAnimationFrame(animate);
    };

    animate();

    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [localPhase]);

  const getPhaseColor = () => {
    switch (localPhase) {
      case 'collapse': return 'from-red-500/50 to-orange-500/50';
      case 'teleport': return 'from-purple-500/50 to-cyan-500/50';
      case 'expand': return 'from-cyan-500/50 to-green-500/50';
    }
  };

  const getPhaseLabel = () => {
    switch (localPhase) {
      case 'collapse': return 'COLLAPSING WAVE FUNCTION';
      case 'teleport': return 'QUANTUM TUNNELING';
      case 'expand': return 'WAVE FUNCTION COLLAPSE';
    }
  };

  return (
    <div className={cn(
      'relative overflow-hidden rounded-xl bg-slate-900/95',
      'border border-cyan-500/30',
      'shadow-[0_0_40px_rgba(6,182,212,0.2)]',
      className
    )}>
      {/* Quantum particle canvas */}
      <canvas
        ref={canvasRef}
        className="absolute inset-0 w-full h-full"
        width={400}
        height={200}
      />

      {/* Phase indicator */}
      <div className={cn(
        'relative z-10 flex flex-col items-center justify-center p-8',
        'bg-gradient-to-b ' + getPhaseColor(),
        'transition-all duration-500'
      )}>
        {/* Phase name */}
        <div className="mb-4 px-4 py-2 rounded-full bg-slate-900/80 border border-cyan-500/50">
          <span className="text-sm font-mono text-cyan-300 tracking-widest">
            {getPhaseLabel()}
          </span>
        </div>

        {/* Workspace names */}
        <div className="flex items-center gap-4 mb-6">
          <div className={cn(
            'px-4 py-2 rounded-lg bg-slate-800/80 border border-slate-600/50',
            'text-sm text-slate-300',
            localPhase === 'collapse' && 'opacity-50 scale-95'
          )}>
            {fromWorkspace}
          </div>
          
          <div className="flex items-center">
            <div className={cn(
              'w-8 h-0.5 bg-gradient-to-r from-cyan-500 to-purple-500',
              'transition-all duration-300',
              localPhase === 'teleport' && 'w-16'
            )} />
            <Atom className="w-5 h-5 text-cyan-400 animate-spin-slow mx-2" />
            <div className={cn(
              'w-8 h-0.5 bg-gradient-to-l from-cyan-500 to-purple-500',
              'transition-all duration-300',
              localPhase === 'teleport' && 'w-16'
            )} />
          </div>
          
          <div className={cn(
            'px-4 py-2 rounded-lg bg-slate-800/80 border border-slate-600/50',
            'text-sm text-slate-300',
            localPhase === 'expand' && 'opacity-50 scale-95'
          )}>
            {toWorkspace}
          </div>
        </div>

        {/* Progress bar */}
        <div className="w-64 h-2 bg-slate-800 rounded-full overflow-hidden">
          <div
            className={cn(
              'h-full transition-all duration-300 ease-out',
              'bg-gradient-to-r from-cyan-500 via-purple-500 to-cyan-400',
              'shadow-[0_0_10px_rgba(6,182,212,0.5)]'
            )}
            style={{ width: `${progress * 100}%` }}
          />
        </div>

        {/* Phase status */}
        <div className="mt-4 text-xs text-cyan-400/80 font-mono">
          PHASE {localPhase.toUpperCase()} • {Math.round(progress * 100)}%
        </div>

        {/* Glitch effect overlay */}
        {localPhase === 'teleport' && (
          <div className="absolute inset-0 pointer-events-none">
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-cyan-500/10 to-transparent animate-glitch" />
          </div>
        )}
      </div>

      <style>{`
        @keyframes spin-slow {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        .animate-spin-slow { animation: spin-slow 3s linear infinite; }
        @keyframes glitch {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(100%); }
        }
        .animate-glitch { animation: glitch 0.5s ease-in-out infinite; }
      `}</style>
    </div>
  );
};

// ============================================================================
// HolographicPanel - Holographic display effect panel
// ============================================================================

export const HolographicPanel: React.FC<HolographicPanelProps> = ({
  effect = 'cyan',
  intensity = 0.8,
  children,
  className,
}) => {
  const [scanlineOffset, setScanlineOffset] = useState(0);
  const [glitchOffset, setGlitchOffset] = useState({ x: 0, y: 0 });

  useEffect(() => {
    const interval = setInterval(() => {
      setScanlineOffset(prev => (prev + 1) % 4);
      if (Math.random() > 0.95) {
        setGlitchOffset({
          x: (Math.random() - 0.5) * 4,
          y: (Math.random() - 0.5) * 2,
        });
      } else {
        setGlitchOffset({ x: 0, y: 0 });
      }
    }, 50);
    return () => clearInterval(interval);
  }, []);

  const getEffectColors = () => {
    switch (effect) {
      case 'rainbow':
        return {
          primary: 'from-purple-500 via-cyan-500 to-pink-500',
          glow: 'rgba(6, 182, 212, 0.4)',
          scanline: 'rgba(6, 182, 212, 0.1)',
          border: 'cyan-500',
        };
      case 'cyan':
        return {
          primary: 'from-cyan-500/20 to-cyan-600/20',
          glow: 'rgba(6, 182, 212, 0.4)',
          scanline: 'rgba(6, 182, 212, 0.1)',
          border: 'cyan-400',
        };
      case 'magenta':
        return {
          primary: 'from-pink-500/20 to-purple-600/20',
          glow: 'rgba(236, 72, 153, 0.4)',
          scanline: 'rgba(236, 72, 153, 0.1)',
          border: 'pink-400',
        };
      case 'white':
        return {
          primary: 'from-slate-200/20 to-slate-400/20',
          glow: 'rgba(255, 255, 255, 0.3)',
          scanline: 'rgba(255, 255, 255, 0.1)',
          border: 'slate-300',
        };
    }
  };

  const colors = getEffectColors();

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-xl',
        'backdrop-blur-md bg-slate-900/60',
        'border border-' + colors.border + '/50',
        'shadow-[' + colors.glow + ']',
        className
      )}
      style={{
        transform: `translate(${glitchOffset.x}px, ${glitchOffset.y}px)`,
        boxShadow: `0 0 ${20 * intensity}px ${colors.glow}`,
      }}
    >
      {/* Scanlines overlay */}
      <div
        className="absolute inset-0 pointer-events-none z-10"
        style={{
          background: `repeating-linear-gradient(
            0deg,
            transparent,
            transparent 2px,
            ${colors.scanline} 2px,
            ${colors.scanline} 4px
          )`,
          backgroundPosition: `0 ${scanlineOffset}px`,
        }}
      />

      {/* Holographic shimmer */}
      <div
        className={cn(
          'absolute inset-0 pointer-events-none z-20',
          'bg-gradient-to-br from-white/5 via-transparent to-transparent'
        )}
      />

      {/* Content */}
      <div className="relative z-30 p-6">
        {children || (
          <div className="flex flex-col items-center gap-3">
            <Hexagon className={cn('w-12 h-12 text-' + colors.border + '-400')} />
            <span className={cn('text-sm text-' + colors.border + '-300')}>
              HOLOGRAPHIC DISPLAY ACTIVE
            </span>
          </div>
        )}
      </div>

      {/* Edge glow effects */}
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/50 to-transparent" />
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/30 to-transparent" />

      {/* Corner accents */}
      <div className={cn('absolute top-2 left-2 w-4 h-4 border-t-2 border-l-2 border-' + colors.border + '-400/50 rounded-tl')} />
      <div className={cn('absolute top-2 right-2 w-4 h-4 border-t-2 border-r-2 border-' + colors.border + '-400/50 rounded-tr')} />
      <div className={cn('absolute bottom-2 left-2 w-4 h-4 border-b-2 border-l-2 border-' + colors.border + '-400/50 rounded-bl')} />
      <div className={cn('absolute bottom-2 right-2 w-4 h-4 border-b-2 border-r-2 border-' + colors.border + '-400/50 rounded-br')} />
    </div>
  );
};

// ============================================================================
// CinematicFocusMode - Cinematic focus mode for immersive viewing
// ============================================================================

export const CinematicFocusMode: React.FC<CinematicFocusModeProps> = ({
  mode = 'theater',
  isActive = false,
  content,
  onActivate,
  onDeactivate,
  className,
}) => {
  const [vignetteOpacity, setVignetteOpacity] = useState(0.3);
  const [curtainPosition, setCurtainPosition] = useState(0);

  useEffect(() => {
    if (isActive) {
      const vignetteTarget = mode === 'theater' ? 0.8 : mode === 'spotlight' ? 0.5 : 0.2;
      const curtainTarget = mode === 'theater' ? 100 : mode === 'spotlight' ? 50 : 0;
      
      let frame = 0;
      const animate = () => {
        frame++;
        const progress = Math.min(frame / 60, 1);
        const eased = 1 - Math.pow(1 - progress, 3);
        
        setVignetteOpacity(vignetteOpacity * (1 - eased) + vignetteTarget * eased);
        setCurtainPosition(curtainPosition * (1 - eased) + curtainTarget * eased);
        
        if (progress < 1) requestAnimationFrame(animate);
      };
      requestAnimationFrame(animate);
    } else {
      setVignetteOpacity(0.3);
      setCurtainPosition(0);
    }
  }, [isActive, mode]);

  const getCurtainColor = () => {
    switch (mode) {
      case 'theater': return 'from-black via-slate-900 to-black';
      case 'spotlight': return 'from-slate-950 via-slate-900/80 to-slate-950';
      case 'zen': return 'from-slate-900 via-slate-800/60 to-slate-900';
    }
  };

  return (
    <div className={cn('relative overflow-hidden rounded-xl bg-slate-900', className)}>
      {/* Content area */}
      <div className={cn(
        'transition-all duration-700 ease-out',
        isActive && mode === 'spotlight' && 'scale-95'
      )}>
        {content || (
          <div className="aspect-video flex items-center justify-center">
            <div className="text-center">
              <Crosshair className={cn(
                'w-16 h-16 text-cyan-400 mx-auto mb-4',
                'transition-transform duration-500',
                isActive && 'scale-110'
              )} />
              <p className="text-cyan-300/80 text-sm">
                {isActive ? `FOCUS MODE: ${mode.toUpperCase()}` : 'CLICK TO ACTIVATE FOCUS MODE'}
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Cinematic overlays */}
      {isActive && (
        <>
          {/* Vignette */}
          <div
            className="absolute inset-0 pointer-events-none z-20"
            style={{
              background: `radial-gradient(ellipse at center, transparent 20%, rgba(0,0,0,${vignetteOpacity}) 100%)`,
            }}
          />

          {/* Theater curtains */}
          {mode === 'theater' && (
            <>
              <div
                className={cn(
                  'absolute top-0 bottom-0 left-0 z-30',
                  'bg-gradient-to-r from-black to-transparent',
                  'transition-all duration-700'
                )}
                style={{ width: `${curtainPosition}%` }}
              />
              <div
                className={cn(
                  'absolute top-0 bottom-0 right-0 z-30',
                  'bg-gradient-to-l from-black to-transparent',
                  'transition-all duration-700'
                )}
                style={{ width: `${curtainPosition}%` }}
              />
            </>
          )}

          {/* Spotlight effect */}
          {mode === 'spotlight' && (
            <div
              className="absolute inset-0 pointer-events-none z-20"
              style={{
                background: 'radial-gradient(circle at 50% 50%, transparent 10%, rgba(0,0,0,0.7) 60%)',
              }}
            />
          )}

          {/* Zen ambient glow */}
          {mode === 'zen' && (
            <div className="absolute inset-0 pointer-events-none z-20">
              <div className="absolute inset-0 bg-gradient-to-b from-cyan-900/10 via-transparent to-cyan-900/10" />
              <div className="absolute inset-0 animate-zen-pulse" />
            </div>
          )}

          {/* Focus indicator */}
          <div className="absolute top-4 left-4 z-40 flex items-center gap-2 px-3 py-1.5 rounded-full bg-black/60 backdrop-blur-sm border border-cyan-500/30">
            <div className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse" />
            <span className="text-xs text-cyan-300 font-mono tracking-wider">
              {mode.toUpperCase()} FOCUS
            </span>
          </div>
        </>
      )}

      {/* Toggle button */}
      <button
        onClick={isActive ? onDeactivate : onActivate}
        className={cn(
          'absolute bottom-4 right-4 z-40',
          'flex items-center gap-2 px-4 py-2 rounded-lg',
          'bg-slate-800/80 backdrop-blur-sm',
          'border border-cyan-500/30',
          'text-sm text-cyan-300',
          'hover:bg-slate-700/80 hover:border-cyan-400',
          'transition-all duration-200'
        )}
      >
        {isActive ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
        {isActive ? 'Exit Focus' : 'Enter Focus'}
      </button>

      <style>{`
        @keyframes zen-pulse {
          0%, 100% { opacity: 0.3; }
          50% { opacity: 0.6; }
        }
        .animate-zen-pulse { animation: zen-pulse 4s ease-in-out infinite; }
      `}</style>
    </div>
  );
};

// ============================================================================
// AIThoughtWave - AI thought wave visualization
// ============================================================================

export const AIThoughtWave: React.FC<AIThoughtWaveProps> = ({
  points = Array.from({ length: 60 }, (_, i) => ({
    timestamp: Date.now() - (60 - i) * 100,
    amplitude: Math.sin(i * 0.3) * 50 + Math.random() * 20 + 30,
    frequency: 0.5 + Math.random() * 0.5,
  })),
  isActive = true,
  color = '#06b6d4',
  showGrid = true,
  onPointHover,
  className,
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const [animatedPoints, setAnimatedPoints] = useState(points);
  const animationRef = useRef<number>();

  useEffect(() => {
    if (!isActive) return;

    const animate = () => {
      setAnimatedPoints(prev => {
        const newPoints = [...prev];
        newPoints.shift();
        newPoints.push({
          timestamp: Date.now(),
          amplitude: Math.sin(Date.now() * 0.005) * 50 + Math.random() * 20 + 30,
          frequency: 0.5 + Math.random() * 0.5,
        });
        return newPoints;
      });
      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isActive]);

  const maxAmplitude = Math.max(...animatedPoints.map(p => p.amplitude), 100);
  const svgHeight = 120;
  const svgWidth = 100;

  const pathD = animatedPoints.map((p, i) => {
    const x = (i / (animatedPoints.length - 1)) * svgWidth;
    const y = svgHeight - (p.amplitude / maxAmplitude) * svgHeight;
    return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
  }).join(' ');

  const areaD = pathD + ` L ${svgWidth} ${svgHeight} L 0 ${svgHeight} Z`;

  return (
    <div className={cn(
      'relative p-4 rounded-xl bg-slate-900/95',
      'border border-cyan-500/30',
      'shadow-[0_0_20px_rgba(6,182,212,0.2)]',
      className
    )}>
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Brain className="w-4 h-4 text-cyan-400" />
          <span className="text-xs text-cyan-300 font-mono">NEURAL ACTIVITY</span>
        </div>
        <div className={cn(
          'flex items-center gap-1.5 px-2 py-1 rounded-full bg-slate-800/80 border border-cyan-500/30'
        )}>
          <div className={cn(
            'w-2 h-2 rounded-full',
            isActive ? 'bg-green-400 animate-pulse' : 'bg-slate-500'
          )} />
          <span className="text-[10px] text-cyan-300">{isActive ? 'ACTIVE' : 'IDLE'}</span>
        </div>
      </div>

      {/* Wave visualization */}
      <div className="relative">
        {/* Grid */}
        {showGrid && (
          <svg className="absolute inset-0 w-full h-full" viewBox={`0 0 ${svgWidth} ${svgHeight}`} preserveAspectRatio="none">
            {[0, 25, 50, 75, 100].map(y => (
              <line
                key={y}
                x1="0"
                y1={(y / 100) * svgHeight}
                x2={svgWidth}
                y2={(y / 100) * svgHeight}
                className="stroke-cyan-500/20"
                strokeWidth="0.5"
                strokeDasharray="2 2"
              />
            ))}
            {[0, 25, 50, 75, 100].map(x => (
              <line
                key={x}
                x1={(x / 100) * svgWidth}
                y1="0"
                x2={(x / 100) * svgWidth}
                y2={svgHeight}
                className="stroke-cyan-500/10"
                strokeWidth="0.5"
                strokeDasharray="2 2"
              />
            ))}
          </svg>
        )}

        {/* Wave area */}
        <svg className="w-full h-28" viewBox={`0 0 ${svgWidth} ${svgHeight}`} preserveAspectRatio="none">
          <defs>
            <linearGradient id="waveGradient" x1="0%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" stopColor={color} stopOpacity="0.4" />
              <stop offset="100%" stopColor={color} stopOpacity="0" />
            </linearGradient>
          </defs>
          
          {/* Area fill */}
          <path d={areaD} fill="url(#waveGradient)" />
          
          {/* Wave line */}
          <path
            d={pathD}
            fill="none"
            stroke={color}
            strokeWidth="2"
            className="drop-shadow-[0_0_4px_rgba(6,182,212,0.8)]"
          />
          
          {/* Active point indicator */}
          {hoveredIndex !== null && animatedPoints[hoveredIndex] && (
            <g>
              <circle
                cx={(hoveredIndex / (animatedPoints.length - 1)) * svgWidth}
                cy={svgHeight - (animatedPoints[hoveredIndex].amplitude / maxAmplitude) * svgHeight}
                r="3"
                fill={color}
                className="animate-ping"
              />
              <circle
                cx={(hoveredIndex / (animatedPoints.length - 1)) * svgWidth}
                cy={svgHeight - (animatedPoints[hoveredIndex].amplitude / maxAmplitude) * svgHeight}
                r="2"
                fill={color}
              />
            </g>
          )}
        </svg>

        {/* Hovered point info */}
        {hoveredIndex !== null && animatedPoints[hoveredIndex] && (
          <div className="absolute top-2 right-2 p-2 rounded-lg bg-slate-800/95 border border-cyan-500/30">
            <div className="text-[10px] text-slate-400">Amplitude</div>
            <div className="text-sm text-cyan-300 font-mono">
              {animatedPoints[hoveredIndex].amplitude.toFixed(1)}
            </div>
          </div>
        )}
      </div>

      {/* Stats row */}
      <div className="flex items-center justify-between mt-3 pt-3 border-t border-cyan-500/20">
        <div className="flex items-center gap-4">
          <div>
            <div className="text-[10px] text-slate-500">AVG AMP</div>
            <div className="text-xs text-cyan-300">
              {(animatedPoints.reduce((a, b) => a + b.amplitude, 0) / animatedPoints.length).toFixed(1)}
            </div>
          </div>
          <div>
            <div className="text-[10px] text-slate-500">PEAK</div>
            <div className="text-xs text-cyan-300">
              {Math.max(...animatedPoints.map(p => p.amplitude)).toFixed(1)}
            </div>
          </div>
          <div>
            <div className="text-[10px] text-slate-500">FREQ</div>
            <div className="text-xs text-cyan-300">
              {(animatedPoints.reduce((a, b) => a + b.frequency, 0) / animatedPoints.length).toFixed(2)}Hz
            </div>
          </div>
        </div>
        <Activity className="w-4 h-4 text-cyan-400/60 animate-pulse" />
      </div>
    </div>
  );
};

// ============================================================================
// GlassExecutionCard - Glass-morphism execution card
// ============================================================================

export const GlassExecutionCard: React.FC<GlassExecutionCardProps> = ({
  execution,
  onClick,
  onCancel,
  className,
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const [timeLeft, setTimeLeft] = useState(execution.duration || 0);

  useEffect(() => {
    if (execution.status !== 'running') return;
    
    const interval = setInterval(() => {
      setTimeLeft(prev => Math.max(0, prev - 1));
    }, 1000);
    
    return () => clearInterval(interval);
  }, [execution.status]);

  const getStatusConfig = () => {
    switch (execution.status) {
      case 'running':
        return {
          bg: 'bg-cyan-500/10',
          border: 'border-cyan-500/50',
          text: 'text-cyan-400',
          icon: <Play className="w-3 h-3" />,
          glow: 'shadow-[0_0_15px_rgba(6,182,212,0.3)]',
        };
      case 'completed':
        return {
          bg: 'bg-green-500/10',
          border: 'border-green-500/50',
          text: 'text-green-400',
          icon: <CheckCircle className="w-3 h-3" />,
          glow: 'shadow-[0_0_10px_rgba(34,197,94,0.2)]',
        };
      case 'failed':
        return {
          bg: 'bg-red-500/10',
          border: 'border-red-500/50',
          text: 'text-red-400',
          icon: <XCircle className="w-3 h-3" />,
          glow: 'shadow-[0_0_10px_rgba(239,68,68,0.3)]',
        };
      case 'pending':
        return {
          bg: 'bg-slate-500/10',
          border: 'border-slate-500/50',
          text: 'text-slate-400',
          icon: <Clock className="w-3 h-3" />,
          glow: '',
        };
    }
  };

  const statusConfig = getStatusConfig();
  const progress = execution.progress || 0;

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div
      className={cn(
        'relative group cursor-pointer',
        'rounded-xl overflow-hidden',
        'bg-gradient-to-br from-slate-800/60 to-slate-900/80',
        'backdrop-blur-xl',
        'border border-slate-700/50',
        'transition-all duration-300',
        'hover:border-cyan-500/50 hover:shadow-[0_0_30px_rgba(6,182,212,0.2)]',
        isHovered && 'scale-[1.02]',
        className
      )}
      onClick={() => onClick?.(execution)}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Glass reflection */}
      <div className="absolute inset-0 bg-gradient-to-br from-white/5 via-transparent to-transparent pointer-events-none" />
      
      {/* Status glow */}
      <div className={cn(
        'absolute top-0 left-0 right-0 h-1',
        statusConfig.bg.replace('/10', '/30'),
        'transition-all duration-300'
      )} />

      <div className="p-4">
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-3">
            {/* Status indicator */}
            <div className={cn(
              'relative flex items-center justify-center w-10 h-10 rounded-lg',
              'bg-slate-800/80 border border-slate-600/50',
              statusConfig.glow
            )}>
              <span className={statusConfig.text}>{statusConfig.icon}</span>
              
              {/* Running animation */}
              {execution.status === 'running' && (
                <div className="absolute inset-0 rounded-lg border-2 border-cyan-400/50 animate-ping" />
              )}
            </div>
            
            <div>
              <h4 className="text-sm font-medium text-slate-100">{execution.name}</h4>
              <div className="flex items-center gap-2 mt-0.5">
                <span className={cn('text-[10px] uppercase', statusConfig.text)}>
                  {execution.status}
                </span>
                {execution.startTime && (
                  <span className="text-[10px] text-slate-500">
                    Started {new Date(execution.startTime).toLocaleTimeString()}
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Actions */}
          {execution.status === 'running' && (
            <button
              onClick={(e) => { e.stopPropagation(); onCancel?.(execution.id); }}
              className={cn(
                'p-1.5 rounded-lg',
                'bg-red-500/20 border border-red-500/50',
                'text-red-400',
                'hover:bg-red-500/30',
                'transition-colors'
              )}
            >
              <Square className="w-3 h-3" />
            </button>
          )}
        </div>

        {/* Progress bar */}
        {execution.status === 'running' && (
          <div className="mb-3">
            <div className="flex items-center justify-between mb-1">
              <span className="text-[10px] text-slate-400">Progress</span>
              <span className="text-[10px] text-cyan-400">{progress}%</span>
            </div>
            <div className="h-1.5 bg-slate-800 rounded-full overflow-hidden">
              <div
                className={cn(
                  'h-full rounded-full transition-all duration-300',
                  'bg-gradient-to-r from-cyan-500 to-cyan-400',
                  'shadow-[0_0_10px_rgba(6,182,212,0.5)]'
                )}
                style={{ width: `${progress}%` }}
              />
            </div>
          </div>
        )}

        {/* Duration */}
        {execution.duration && (
          <div className="flex items-center justify-between">
            <span className="text-[10px] text-slate-500">Duration</span>
            <span className={cn('text-xs font-mono', statusConfig.text)}>
              {formatDuration(timeLeft)}
            </span>
          </div>
        )}

        {/* Metadata */}
        {execution.metadata && Object.keys(execution.metadata).length > 0 && (
          <div className="mt-3 pt-3 border-t border-slate-700/50 flex flex-wrap gap-2">
            {Object.entries(execution.metadata).slice(0, 3).map(([key, value]) => (
              <span
                key={key}
                className="px-2 py-0.5 rounded text-[10px] bg-slate-700/50 text-slate-400"
              >
                {key}: {String(value).slice(0, 20)}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Hover glow effect */}
      {isHovered && (
        <div className="absolute inset-0 rounded-xl bg-cyan-500/5 pointer-events-none" />
      )}
    </div>
  );
};

// ============================================================================
// TokenStormRenderer - Token stream visualization
// ============================================================================

export const TokenStormRenderer: React.FC<TokenStormRendererProps> = ({
  events = [
    { id: '1', type: 'input', content: 'What is the status of', timestamp: Date.now() - 5000 },
    { id: '2', type: 'thought', content: 'Retrieving system status...', timestamp: Date.now() - 4000 },
    { id: '3', type: 'output', content: 'All systems operational', timestamp: Date.now() - 3000 },
    { id: '4', type: 'input', content: 'Show me the metrics', timestamp: Date.now() - 2000 },
    { id: '5', type: 'thought', content: 'Aggregating metrics...', timestamp: Date.now() - 1000 },
    { id: '6', type: 'output', content: '99.9% uptime, 45ms avg latency', timestamp: Date.now() },
  ],
  isStreaming = true,
  speed = 1,
  onEventClick,
  className,
}) => {
  const [visibleEvents, setVisibleEvents] = useState(events.slice(0, 4));
  const [currentEventIndex, setCurrentEventIndex] = useState(4);

  useEffect(() => {
    if (!isStreaming || currentEventIndex >= events.length) return;

    const interval = setInterval(() => {
      setVisibleEvents(prev => {
        const newEvents = [...prev, events[currentEventIndex]];
        if (newEvents.length > 6) newEvents.shift();
        return newEvents;
      });
      setCurrentEventIndex(prev => prev + 1);
    }, 2000 / speed);

    return () => clearInterval(interval);
  }, [isStreaming, currentEventIndex, events, speed]);

  const getEventColors = (type: TokenEventType) => {
    switch (type) {
      case 'input':
        return {
          bg: 'bg-cyan-500/20',
          border: 'border-cyan-500/50',
          text: 'text-cyan-300',
          icon: <ChevronRight className="w-3 h-3" />,
        };
      case 'output':
        return {
          bg: 'bg-green-500/20',
          border: 'border-green-500/50',
          text: 'text-green-300',
          icon: <ChevronLeft className="w-3 h-3" />,
        };
      case 'thought':
        return {
          bg: 'bg-purple-500/20',
          border: 'border-purple-500/50',
          text: 'text-purple-300',
          icon: <Lightbulb className="w-3 h-3" />,
        };
    }
  };

  return (
    <div className={cn(
      'flex flex-col rounded-xl bg-slate-900/95',
      'border border-slate-700/50',
      'shadow-[0_0_20px_rgba(0,0,0,0.5)]',
      className
    )}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700/50">
        <div className="flex items-center gap-2">
          <Waves className="w-4 h-4 text-cyan-400" />
          <span className="text-xs text-cyan-300 font-mono tracking-wider">
            TOKEN STREAM
          </span>
        </div>
        <div className="flex items-center gap-2">
          <div className={cn(
            'w-2 h-2 rounded-full',
            isStreaming ? 'bg-green-400 animate-pulse' : 'bg-slate-500'
          )} />
          <span className="text-[10px] text-slate-400">
            {isStreaming ? 'STREAMING' : 'PAUSED'}
          </span>
        </div>
      </div>

      {/* Token stream */}
      <div className="flex-1 overflow-hidden p-4">
        <div className="relative h-full">
          {/* Streaming line */}
          {isStreaming && (
            <div className="absolute left-0 top-0 bottom-0 w-0.5 bg-gradient-to-b from-cyan-400 via-purple-400 to-transparent animate-stream" />
          )}

          {/* Events */}
          <div className="space-y-3">
            {visibleEvents.map((event, index) => {
              const colors = getEventColors(event.type);
              const isNew = index === visibleEvents.length - 1 && isStreaming;

              return (
                <div
                  key={event.id}
                  onClick={() => onEventClick?.(event)}
                  className={cn(
                    'relative p-3 rounded-lg',
                    'bg-slate-800/50 border',
                    colors.border,
                    'transition-all duration-300',
                    isNew && 'animate-fade-in',
                    'cursor-pointer hover:bg-slate-800/80'
                  )}
                >
                  {/* Event type badge */}
                  <div className={cn(
                    'absolute -top-2 left-3 flex items-center gap-1 px-2 py-0.5 rounded-full',
                    colors.bg,
                    'border border-inherit'
                  )}>
                    <span className={colors.text}>{colors.icon}</span>
                    <span className={cn('text-[10px] uppercase font-medium', colors.text)}>
                      {event.type}
                    </span>
                  </div>

                  {/* Content */}
                  <div className="mt-2">
                    <p className={cn('text-sm', colors.text)}>
                      {event.content}
                    </p>
                    <span className="text-[10px] text-slate-500 mt-1 block">
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="px-4 py-2 border-t border-slate-700/50 bg-slate-800/50">
        <div className="flex items-center justify-between text-[10px] text-slate-500">
          <span>{visibleEvents.length} tokens rendered</span>
          <span>Speed: {speed}x</span>
        </div>
      </div>

      <style>{`
        @keyframes stream {
          0% { transform: scaleY(0); opacity: 0; }
          50% { opacity: 1; }
          100% { transform: scaleY(1); opacity: 0; }
        }
        .animate-stream { animation: stream 1s ease-out infinite; }
        @keyframes fade-in {
          0% { opacity: 0; transform: translateY(10px); }
          100% { opacity: 1; transform: translateY(0); }
        }
        .animate-fade-in { animation: fade-in 0.3s ease-out; }
      `}</style>
    </div>
  );
};

// ============================================================================
// SwarmMindVisualizer - Swarm intelligence visualization
// ============================================================================

export const SwarmMindVisualizer: React.FC<SwarmMindVisualizerProps> = ({
  agents = Array.from({ length: 30 }, (_, i) => ({
    id: `agent-${i}`,
    x: Math.random() * 400,
    y: Math.random() * 300,
    velocity: { dx: (Math.random() - 0.5) * 4, dy: (Math.random() - 0.5) * 4 },
    state: ['exploring', 'exploiting', 'returning'][Math.floor(Math.random() * 3)] as SwarmAgentState,
  })),
  targetX = 200,
  targetY = 150,
  isActive = true,
  onAgentClick,
  className,
}) => {
  const [swarmAgents, setSwarmAgents] = useState(agents);
  const [hoveredAgent, setHoveredAgent] = useState<string | null>(null);
  const animationRef = useRef<number>();

  useEffect(() => {
    if (!isActive) return;

    const animate = () => {
      setSwarmAgents(prev => prev.map(agent => {
        const dx = targetX - agent.x;
        const dy = targetY - agent.y;
        const dist = Math.sqrt(dx * dx + dy * dy);
        
        let newState: SwarmAgentState = 'exploring';
        let vx = agent.velocity.dx;
        let vy = agent.velocity.dy;

        if (dist < 100) {
          newState = 'exploiting';
          vx += dx * 0.002 + (Math.random() - 0.5) * 0.5;
          vy += dy * 0.002 + (Math.random() - 0.5) * 0.5;
        } else if (dist > 200) {
          newState = 'returning';
          vx += dx * 0.001;
          vy += dy * 0.001;
        } else {
          newState = 'exploring';
          vx += (Math.random() - 0.5) * 1;
          vy += (Math.random() - 0.5) * 1;
        }

        // Apply velocity with damping
        vx *= 0.98;
        vy *= 0.98;

        // Clamp velocity
        const maxVel = 4;
        vx = Math.max(-maxVel, Math.min(maxVel, vx));
        vy = Math.max(-maxVel, Math.min(maxVel, vy));

        return {
          ...agent,
          x: agent.x + vx,
          y: agent.y + vy,
          velocity: { dx: vx, dy: vy },
          state: newState,
        };
      }));

      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isActive, targetX, targetY]);

  const getAgentColor = (state: SwarmAgentState) => {
    switch (state) {
      case 'exploring': return { fill: '#06b6d4', glow: 'rgba(6,182,212,0.5)' };
      case 'exploiting': return { fill: '#a855f7', glow: 'rgba(168,85,247,0.5)' };
      case 'returning': return { fill: '#22c55e', glow: 'rgba(34,197,94,0.5)' };
    }
  };

  return (
    <div className={cn(
      'relative rounded-xl overflow-hidden bg-slate-900/95',
      'border border-slate-700/50',
      'shadow-[0_0_30px_rgba(0,0,0,0.5)]',
      className
    )}>
      {/* Header */}
      <div className="absolute top-3 left-3 z-10 flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
        <Bot className="w-4 h-4 text-cyan-400" />
        <span className="text-xs text-cyan-300 font-mono tracking-wider">
          SWARM MIND
        </span>
      </div>

      {/* Legend */}
      <div className="absolute top-3 right-3 z-10 flex items-center gap-3 px-3 py-1.5 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
        <div className="flex items-center gap-1">
          <div className="w-2 h-2 rounded-full bg-cyan-400" />
          <span className="text-[10px] text-slate-400">Exploring</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-2 h-2 rounded-full bg-purple-400" />
          <span className="text-[10px] text-slate-400">Exploiting</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-2 h-2 rounded-full bg-green-400" />
          <span className="text-[10px] text-slate-400">Returning</span>
        </div>
      </div>

      {/* SVG Canvas */}
      <svg className="w-full h-64" viewBox="0 0 400 300">
        {/* Grid */}
        <defs>
          <pattern id="swarm-grid" width="20" height="20" patternUnits="userSpaceOnUse">
            <path d="M 20 0 L 0 0 0 20" fill="none" stroke="rgba(100,116,139,0.2)" strokeWidth="0.5" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#swarm-grid)" />

        {/* Target zone */}
        <circle
          cx={targetX}
          cy={targetY}
          r="40"
          fill="rgba(6,182,212,0.1)"
          stroke="rgba(6,182,212,0.3)"
          strokeWidth="1"
          strokeDasharray="4 4"
        >
          <animate attributeName="r" values="35;45;35" dur="2s" repeatCount="indefinite" />
        </circle>
        <circle cx={targetX} cy={targetY} r="8" fill="rgba(6,182,212,0.3)" />

        {/* Agent trails */}
        {swarmAgents.map(agent => {
          const colors = getAgentColor(agent.state);
          return (
            <line
              key={`trail-${agent.id}`}
              x1={agent.x}
              y1={agent.y}
              x2={agent.x - agent.velocity.dx * 3}
              y2={agent.y - agent.velocity.dy * 3}
              stroke={colors.fill}
              strokeWidth="1"
              strokeOpacity="0.3"
            />
          );
        })}

        {/* Agents */}
        {swarmAgents.map(agent => {
          const colors = getAgentColor(agent.state);
          const isHovered = hoveredAgent === agent.id;

          return (
            <g
              key={agent.id}
              onClick={() => onAgentClick?.(agent)}
              onMouseEnter={() => setHoveredAgent(agent.id)}
              onMouseLeave={() => setHoveredAgent(null)}
              className="cursor-pointer"
            >
              {/* Glow */}
              <circle
                cx={agent.x}
                cy={agent.y}
                r={isHovered ? 8 : 5}
                fill={colors.fill}
                fillOpacity="0.2"
                filter="blur(2px)"
              />
              
              {/* Main body */}
              <circle
                cx={agent.x}
                cy={agent.y}
                r={isHovered ? 6 : 4}
                fill={colors.fill}
                stroke={isHovered ? '#fff' : 'none'}
                strokeWidth="1"
              >
                {agent.state === 'exploiting' && (
                  <animate attributeName="r" values="4;6;4" dur="0.5s" repeatCount="indefinite" />
                )}
              </circle>

              {/* Velocity vector */}
              {isHovered && (
                <line
                  x1={agent.x}
                  y1={agent.y}
                  x2={agent.x + agent.velocity.dx * 5}
                  y2={agent.y + agent.velocity.dy * 5}
                  stroke={colors.fill}
                  strokeWidth="1"
                  markerEnd="url(#arrowhead)"
                />
              )}
            </g>
          );
        })}

        {/* Arrow marker definition */}
        <defs>
          <marker id="arrowhead" markerWidth="6" markerHeight="6" refX="5" refY="3" orient="auto">
            <polygon points="0 0, 6 3, 0 6" fill="currentColor" />
          </marker>
        </defs>
      </svg>

      {/* Stats footer */}
      <div className="absolute bottom-0 left-0 right-0 px-4 py-2 bg-slate-800/90 border-t border-slate-700/50">
        <div className="flex items-center justify-between text-[10px] text-slate-400">
          <span>{swarmAgents.length} Agents Active</span>
          <span>Target: ({targetX}, {targetY})</span>
        </div>
      </div>
    </div>
  );
};

// ============================================================================
// AmbientTelemetryLayer - Ambient background telemetry display
// ============================================================================

export const AmbientTelemetryLayer: React.FC<AmbientTelemetryLayerProps> = ({
  metrics = [
    { label: 'CPU', value: 67, unit: '%', trend: 'up' },
    { label: 'Memory', value: 45, unit: '%', trend: 'stable' },
    { label: 'Network', value: 128, unit: 'MB/s', trend: 'down' },
    { label: 'Requests', value: 2847, unit: '/min', trend: 'up' },
    { label: 'Latency', value: 23, unit: 'ms', trend: 'down' },
    { label: 'Error Rate', value: 0.12, unit: '%', trend: 'stable' },
  ],
  isActive = true,
  opacity = 0.7,
  onMetricClick,
  className,
}) => {
  const [pulseState, setPulseState] = useState(0);

  useEffect(() => {
    if (!isActive) return;
    const interval = setInterval(() => {
      setPulseState(prev => (prev + 1) % 360);
    }, 50);
    return () => clearInterval(interval);
  }, [isActive]);

  const getTrendIcon = (trend: TelemetryTrend) => {
    switch (trend) {
      case 'up': return <TrendingUp className="w-3 h-3 text-green-400" />;
      case 'down': return <TrendingDown className="w-3 h-3 text-cyan-400" />;
      case 'stable': return <Minus className="w-3 h-3 text-slate-400" />;
    }
  };

  const getTrendColor = (value: number, trend: TelemetryTrend) => {
    if (trend === 'stable') return 'text-slate-300';
    if (value > 80) return trend === 'up' ? 'text-red-400' : 'text-green-400';
    if (value > 60) return trend === 'up' ? 'text-amber-400' : 'text-green-400';
    return 'text-cyan-400';
  };

  return (
    <div
      className={cn(
        'relative p-6 rounded-2xl',
        'bg-gradient-to-br from-slate-900/95 via-slate-800/90 to-slate-900/95',
        'backdrop-blur-xl',
        'border border-slate-700/30',
        'overflow-hidden',
        className
      )}
      style={{ opacity }}
    >
      {/* Ambient background effects */}
      {isActive && (
        <>
          {/* Pulsing gradient orbs */}
          <div
            className="absolute -top-20 -right-20 w-40 h-40 rounded-full bg-cyan-500/10 blur-3xl"
            style={{
              transform: `scale(${1 + Math.sin(pulseState * 0.02) * 0.3})`,
              transition: 'transform 0.5s ease-out',
            }}
          />
          <div
            className="absolute -bottom-20 -left-20 w-40 h-40 rounded-full bg-purple-500/10 blur-3xl"
            style={{
              transform: `scale(${1 + Math.cos(pulseState * 0.02) * 0.3})`,
              transition: 'transform 0.5s ease-out',
            }}
          />
          
          {/* Scan line effect */}
          <div
            className="absolute inset-0 pointer-events-none overflow-hidden"
          >
            <div
              className="absolute left-0 right-0 h-px bg-gradient-to-r from-transparent via-cyan-500/50 to-transparent"
              style={{
                top: `${(pulseState % 100)}%`,
                animation: 'scan-line 3s linear infinite',
              }}
            />
          </div>
        </>
      )}

      {/* Header */}
      <div className="relative z-10 flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-cyan-400" />
          <span className="text-sm text-cyan-300 font-mono tracking-wider">
            AMBIENT TELEMETRY
          </span>
        </div>
        <div className={cn(
          'flex items-center gap-1.5 px-2 py-1 rounded-full',
          'bg-slate-800/80 border border-cyan-500/30'
        )}>
          <Radio className={cn(
            'w-3 h-3 text-cyan-400',
            isActive && 'animate-pulse'
          )} />
          <span className="text-[10px] text-cyan-300">
            {isActive ? 'LIVE' : 'OFFLINE'}
          </span>
        </div>
      </div>

      {/* Metrics grid */}
      <div className="relative z-10 grid grid-cols-2 md:grid-cols-3 gap-4">
        {metrics.map((metric, index) => (
          <div
            key={metric.label}
            onClick={() => onMetricClick?.(metric)}
            className={cn(
              'group relative p-4 rounded-xl',
              'bg-slate-800/50 backdrop-blur-sm',
              'border border-slate-700/50',
              'hover:border-cyan-500/50 hover:bg-slate-800/80',
              'transition-all duration-300 cursor-pointer',
              'hover:shadow-[0_0_20px_rgba(6,182,212,0.2)]'
            )}
            style={{
              animationDelay: `${index * 100}ms`,
            }}
          >
            {/* Metric glow on hover */}
            <div className="absolute inset-0 rounded-xl bg-cyan-500/0 group-hover:bg-cyan-500/5 transition-colors" />

            <div className="relative">
              {/* Label row */}
              <div className="flex items-center justify-between mb-2">
                <span className="text-[10px] text-slate-500 uppercase tracking-wider">
                  {metric.label}
                </span>
                {getTrendIcon(metric.trend)}
              </div>

              {/* Value */}
              <div className="flex items-baseline gap-1">
                <span className={cn(
                  'text-2xl font-bold font-mono',
                  getTrendColor(metric.value, metric.trend)
                )}>
                  {metric.value.toLocaleString()}
                </span>
                <span className="text-xs text-slate-500">{metric.unit}</span>
              </div>

              {/* Subtle bar */}
              <div className="mt-2 h-1 bg-slate-700/50 rounded-full overflow-hidden">
                <div
                  className={cn(
                    'h-full rounded-full transition-all duration-500',
                    metric.value > 80 ? 'bg-red-500' :
                    metric.value > 60 ? 'bg-amber-500' : 'bg-cyan-500'
                  )}
                  style={{
                    width: `${Math.min(metric.value, 100)}%`,
                    boxShadow: `0 0 8px ${
                      metric.value > 80 ? 'rgba(239,68,68,0.5)' :
                      metric.value > 60 ? 'rgba(245,158,11,0.5)' : 'rgba(6,182,212,0.5)'
                    }`,
                  }}
                />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Opacity control */}
      <div className="relative z-10 mt-6 pt-4 border-t border-slate-700/50">
        <div className="flex items-center justify-between text-[10px] text-slate-500">
          <span>Background Layer</span>
          <span>{Math.round(opacity * 100)}% Opacity</span>
        </div>
      </div>

      <style>{`
        @keyframes scan-line {
          0% { transform: translateY(0); }
          100% { transform: translateY(100px); }
        }
      `}</style>
    </div>
  );
};

// ============================================================================
// DigitalTwinViewport - Digital twin viewport
// ============================================================================

export const DigitalTwinViewport: React.FC<DigitalTwinViewportProps> = ({
  state = {
    entities: [
      { id: '1', type: 'Server', position: { x: 100, y: 100, z: 0 }, status: 'healthy' },
      { id: '2', type: 'Database', position: { x: 200, y: 80, z: 0 }, status: 'healthy' },
      { id: '3', type: 'Cache', position: { x: 150, y: 150, z: 0 }, status: 'warning' },
      { id: '4', type: 'Queue', position: { x: 250, y: 120, z: 0 }, status: 'healthy' },
      { id: '5', type: 'API', position: { x: 175, y: 200, z: 0 }, status: 'healthy' },
      { id: '6', type: 'Worker', position: { x: 300, y: 100, z: 0 }, status: 'healthy' },
    ],
    connections: [
      { source: '1', target: '2', strength: 0.9 },
      { source: '1', target: '3', strength: 0.7 },
      { source: '2', target: '4', strength: 0.8 },
      { source: '3', target: '5', strength: 0.6 },
      { source: '1', target: '5', strength: 0.85 },
      { source: '5', target: '6', strength: 0.75 },
      { source: '4', target: '6', strength: 0.65 },
    ],
    timestamp: Date.now(),
  },
  selectedEntityId = null,
  isAnimating = true,
  cameraRotation = { x: 0, y: 0 },
  onEntitySelect,
  onConnectionHover,
  className,
}) => {
  const [rotation, setRotation] = useState(cameraRotation);
  const [hoveredEntity, setHoveredEntity] = useState<string | null>(null);
  const [hoveredConnection, setHoveredConnection] = useState<string | null>(null);
  const animationRef = useRef<number>();

  useEffect(() => {
    if (!isAnimating) return;

    const animate = () => {
      setRotation(prev => ({
        x: prev.x + 0.1,
        y: prev.y + 0.15,
      }));
      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current);
    };
  }, [isAnimating]);

  const getEntityColor = (status: string) => {
    switch (status) {
      case 'healthy': return { primary: '#22c55e', glow: 'rgba(34,197,94,0.5)' };
      case 'warning': return { primary: '#f59e0b', glow: 'rgba(245,158,11,0.5)' };
      case 'critical': return { primary: '#ef4444', glow: 'rgba(239,68,68,0.5)' };
      default: return { primary: '#6b7280', glow: 'rgba(107,114,128,0.5)' };
    }
  };

  const getEntityIcon = (type: string) => {
    switch (type) {
      case 'Server': return <Server className="w-4 h-4" />;
      case 'Database': return <Database className="w-4 h-4" />;
      case 'Cache': return <Cpu className="w-4 h-4" />;
      case 'Queue': return <Layers className="w-4 h-4" />;
      case 'API': return <Network className="w-4 h-4" />;
      case 'Worker': return <Bot className="w-4 h-4" />;
      default: return <Hexagon className="w-4 h-4" />;
    }
  };

  const getConnectionId = (source: string, target: string) => `${source}-${target}`;

  return (
    <div className={cn(
      'relative rounded-xl overflow-hidden bg-slate-900/95',
      'border border-slate-700/50',
      'shadow-[0_0_40px_rgba(0,0,0,0.5)]',
      className
    )}>
      {/* Header */}
      <div className="absolute top-3 left-3 z-20 flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
        <Atom className="w-4 h-4 text-cyan-400" />
        <span className="text-xs text-cyan-300 font-mono tracking-wider">
          DIGITAL TWIN
        </span>
      </div>

      {/* Status badges */}
      <div className="absolute top-3 right-3 z-20 flex items-center gap-2">
        <div className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-slate-600/50">
          <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          <span className="text-[10px] text-slate-300">
            {state.entities.filter(e => e.status === 'healthy').length} Healthy
          </span>
        </div>
        {state.entities.some(e => e.status === 'warning') && (
          <div className="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-slate-800/90 backdrop-blur-sm border border-amber-500/50">
            <div className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
            <span className="text-[10px] text-amber-300">
              {state.entities.filter(e => e.status === 'warning').length} Warning
            </span>
          </div>
        )}
      </div>

      {/* 3D Canvas */}
      <svg
        className="w-full h-80"
        viewBox="0 0 400 300"
        style={{
          transform: `perspective(800px) rotateX(${rotation.x}deg) rotateY(${rotation.y}deg)`,
          transformStyle: 'preserve-3d',
        }}
      >
        <defs>
          {/* Grid pattern */}
          <pattern id="twin-grid" width="20" height="20" patternUnits="userSpaceOnUse">
            <path d="M 20 0 L 0 0 0 20" fill="none" stroke="rgba(100,116,139,0.15)" strokeWidth="0.5" />
          </pattern>
          
          {/* Glow filter */}
          <filter id="twin-glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="3" result="coloredBlur" />
            <feMerge>
              <feMergeNode in="coloredBlur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        {/* Background grid */}
        <rect width="100%" height="100%" fill="url(#twin-grid)" />

        {/* Connections */}
        {state.connections.map(conn => {
          const source = state.entities.find(e => e.id === conn.source);
          const target = state.entities.find(e => e.id === conn.target);
          if (!source || !target) return null;

          const connId = getConnectionId(conn.source, conn.target);
          const isHovered = hoveredConnection === connId;
          const isSelected = selectedEntityId === conn.source || selectedEntityId === conn.target;

          return (
            <g
              key={connId}
              onMouseEnter={() => {
                setHoveredConnection(connId);
                onConnectionHover?.(conn);
              }}
              onMouseLeave={() => {
                setHoveredConnection(null);
                onConnectionHover?.(null);
              }}
              className="cursor-pointer"
            >
              {/* Connection line */}
              <line
                x1={source.position.x}
                y1={source.position.y}
                x2={target.position.x}
                y2={target.position.y}
                stroke={isHovered || isSelected ? '#06b6d4' : 'rgba(100,116,139,0.5)'}
                strokeWidth={isHovered ? 3 : conn.strength * 2}
                strokeOpacity={conn.strength}
                filter={isHovered ? 'url(#twin-glow)' : undefined}
              />
              
              {/* Data flow particles */}
              {isAnimating && (
                <circle r="3" fill="#06b6d4">
                  <animateMotion
                    path={`M${source.position.x},${source.position.y} L${target.position.x},${target.position.y}`}
                    dur={`${2 / conn.strength}s`}
                    repeatCount="indefinite"
                  />
                </circle>
              )}
            </g>
          );
        })}

        {/* Entities */}
        {state.entities.map(entity => {
          const colors = getEntityColor(entity.status);
          const isSelected = selectedEntityId === entity.id;
          const isHovered = hoveredEntity === entity.id;

          return (
            <g
              key={entity.id}
              onClick={() => onEntitySelect?.(entity)}
              onMouseEnter={() => setHoveredEntity(entity.id)}
              onMouseLeave={() => setHoveredEntity(null)}
              className="cursor-pointer"
            >
              {/* Glow effect */}
              <circle
                cx={entity.position.x}
                cy={entity.position.y}
                r={isHovered || isSelected ? 25 : 20}
                fill={colors.glow}
                filter="url(#twin-glow)"
                className="transition-all duration-300"
              />

              {/* Main shape */}
              <g transform={`translate(${entity.position.x - 12}, ${entity.position.y - 12})`}>
                {/* Hexagonal shape */}
                <polygon
                  points="12,0 24,6 24,18 12,24 0,18 0,6"
                  fill="rgba(15,23,42,0.9)"
                  stroke={isSelected ? '#06b6d4' : colors.primary}
                  strokeWidth={isSelected || isHovered ? 2 : 1}
                  className="transition-all duration-200"
                />
                
                {/* Icon */}
                <g transform="translate(4, 4)" className="text-slate-300" style={{ color: colors.primary }}>
                  {getEntityIcon(entity.type)}
                </g>
              </g>

              {/* Label on hover */}
              {(isHovered || isSelected) && (
                <g transform={`translate(${entity.position.x - 30}, ${entity.position.y + 30})`}>
                  <rect
                    x="0"
                    y="0"
                    width="60"
                    height="24"
                    rx="4"
                    fill="rgba(15,23,42,0.95)"
                    stroke={colors.primary}
                    strokeWidth="1"
                  />
                  <text
                    x="30"
                    y="15"
                    textAnchor="middle"
                    className="text-[10px] fill-slate-300"
                    style={{ color: colors.primary }}
                  >
                    {entity.type}
                  </text>
                </g>
              )}

              {/* Pulsing ring for warning/critical */}
              {(entity.status === 'warning' || entity.status === 'critical') && isAnimating && (
                <circle
                  cx={entity.position.x}
                  cy={entity.position.y}
                  r="25"
                  fill="none"
                  stroke={colors.primary}
                  strokeWidth="2"
                  strokeOpacity="0.5"
                >
                  <animate attributeName="r" values="20;35;20" dur="2s" repeatCount="indefinite" />
                  <animate attributeName="strokeOpacity" values="0.5;0;0.5" dur="2s" repeatCount="indefinite" />
                </circle>
              )}
            </g>
          );
        })}
      </svg>

      {/* Footer info */}
      <div className="absolute bottom-0 left-0 right-0 px-4 py-3 bg-slate-800/90 border-t border-slate-700/50">
        <div className="flex items-center justify-between text-[10px] text-slate-500">
          <div className="flex items-center gap-4">
            <span>{state.entities.length} Entities</span>
            <span>{state.connections.length} Connections</span>
          </div>
          <span>Last Update: {new Date(state.timestamp).toLocaleTimeString()}</span>
        </div>
      </div>

      {/* Selected entity detail panel */}
      {selectedEntityId && state.entities.find(e => e.id === selectedEntityId) && (
        <div className="absolute bottom-12 left-4 right-4 p-4 rounded-lg bg-slate-800/95 border border-cyan-500/30 backdrop-blur-sm">
          {(() => {
            const entity = state.entities.find(e => e.id === selectedEntityId)!;
            const colors = getEntityColor(entity.status);
            return (
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ backgroundColor: colors.glow }}>
                    <span style={{ color: colors.primary }}>{getEntityIcon(entity.type)}</span>
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-slate-100">{entity.type}</h4>
                    <p className="text-[10px] text-slate-400">ID: {entity.id}</p>
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-sm font-medium" style={{ color: colors.primary }}>
                    {entity.status.toUpperCase()}
                  </div>
                  <p className="text-[10px] text-slate-500">
                    Pos: ({entity.position.x.toFixed(0)}, {entity.position.y.toFixed(0)}, {entity.position.z.toFixed(0)})
                  </p>
                </div>
              </div>
            );
          })()}
        </div>
      )}
    </div>
  );
};

// ============================================================================
// Missing Lucide Icons
// ============================================================================

// Add missing icons that weren't imported at the top
const Rocket: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z" />
    <path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z" />
    <path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0" />
    <path d="M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" />
  </svg>
);

const Bug: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="m8 2 1.88 1.88" />
    <path d="M14.12 3.88 16 2" />
    <path d="M9 7.13v-1a3.003 3.003 0 1 1 6 0v1" />
    <path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v3c0 3.3-2.7 6-6 6" />
    <path d="M12 20v-9" />
    <path d="M6.53 9C4.6 8.8 3 7.1 3 5" />
    <path d="M6 13H2" />
    <path d="M3 21c0-2.1 1.7-3.9 3.8-4" />
    <path d="M20.97 5c0 2.1-1.6 3.8-3.5 4" />
    <path d="M22 13h-4" />
    <path d="M17.2 17c2.1.1 3.8 1.9 3.8 4" />
  </svg>
);

const Shield: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" />
  </svg>
);

const CreditCard: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="20" height="14" x="2" y="5" rx="2" />
    <line x1="2" x2="22" y1="10" y2="10" />
  </svg>
);

const ShieldCheck: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" />
    <path d="m9 12 2 2 4-4" />
  </svg>
);

const Server: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="20" height="8" x="2" y="2" rx="2" ry="2" />
    <rect width="20" height="8" x="2" y="14" rx="2" ry="2" />
    <line x1="6" x2="6.01" y1="6" y2="6" />
    <line x1="6" x2="6.01" y1="18" y2="18" />
  </svg>
);

const Users: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
    <circle cx="9" cy="7" r="4" />
    <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
    <path d="M16 3.13a4 4 0 0 1 0 7.75" />
  </svg>
);

const CheckCircle: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <path d="m9 12 2 2 4-4" />
  </svg>
);

const XCircle: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <path d="m15 9-6 6" />
    <path d="m9 9 6 6" />
  </svg>
);

const Clock: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="12" cy="12" r="10" />
    <polyline points="12 6 12 12 16 14" />
  </svg>
);

const Square: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="18" height="18" x="3" y="3" rx="2" />
  </svg>
);

const Settings: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
    <circle cx="12" cy="12" r="3" />
  </svg>
);

const DollarSign: React.FC<{ className?: string }> = ({ className }) => (
  <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="12" x2="12" y1="2" y2="22" />
    <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
  </svg>
);
