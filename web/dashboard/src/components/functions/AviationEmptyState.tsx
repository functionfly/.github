import { Button } from '@/components/ui/button';
import { Radar, Satellite, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';

interface AviationEmptyStateProps {
  onDeploy: () => void;
  searchQuery?: string;
}

/**
 * Aviation-themed empty state with animated radar and flight metaphor
 * Theme-aware: adapts to light/dark mode with enhanced visibility
 */
export function AviationEmptyState({ onDeploy, searchQuery }: AviationEmptyStateProps) {
  const [mounted, setMounted] = useState(false);
  const [radarRotation, setRadarRotation] = useState(0);
  const [blips, setBlips] = useState<{ id: number; x: number; y: number; delay: number }[]>([]);

  useEffect(() => {
    setMounted(true);

    // Generate random radar blips
    const newBlips = Array.from({ length: 5 }, (_, i) => ({
      id: i,
      x: Math.random() * 60 + 20, // 20-80%
      y: Math.random() * 60 + 20,
      delay: Math.random() * 3,
    }));
    setBlips(newBlips);

    // Animate radar sweep
    const interval = setInterval(() => {
      setRadarRotation((prev) => (prev + 1) % 360);
    }, 16); // ~60fps

    return () => clearInterval(interval);
  }, []);

  const isSearch = !!searchQuery;

  return (
    <div className="aviation-panel aviation-panel-glow p-8 md:p-12 relative overflow-hidden">
      {/* Background grid pattern - theme aware */}
      <div className="absolute inset-0 aviation-grid-bg opacity-50" />

      {/* Animated radar display */}
      <div className="relative z-10 flex flex-col items-center">
        {/* Radar circle */}
        <div
          className={`relative w-48 h-48 md:w-64 md:h-64 mb-8 transition-all duration-700 ${
            mounted ? 'opacity-100 scale-100' : 'opacity-0 scale-90'
          }`}
        >
          {/* Outer ring - stronger in light mode for visibility */}
          <div
            className="absolute inset-0 rounded-full border-2 dark:border-(--color-aviation-amber-dim)"
            style={{
              borderColor: 'var(--color-aviation-amber)',
              opacity: 0.6,
            }}
          />

          {/* Middle ring - more visible */}
          <div
            className="absolute inset-4 rounded-full border-2 dark:border-(--color-aviation-amber-dim)"
            style={{
              borderColor: 'var(--color-aviation-amber)',
              opacity: 0.4,
            }}
          />

          {/* Inner ring */}
          <div
            className="absolute inset-8 rounded-full border-2 dark:border-(--color-aviation-amber-dim)"
            style={{
              borderColor: 'var(--color-aviation-amber)',
              opacity: 0.25,
            }}
          />

          {/* Center point - amber glow with strong shadow */}
          <div
            className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-3 h-3 rounded-full"
            style={{
              background: 'var(--color-aviation-amber)',
              boxShadow:
                '0 0 15px var(--color-aviation-amber-glow), 0 0 30px var(--color-aviation-amber-glow)',
            }}
          />

          {/* Crosshairs - thicker for visibility */}
          <div
            className="absolute top-0 left-1/2 -translate-x-1/2 w-0.5 h-full"
            style={{
              background: 'var(--color-aviation-amber)',
              opacity: 0.5,
            }}
          />
          <div
            className="absolute top-1/2 left-0 -translate-y-1/2 w-full h-0.5"
            style={{
              background: 'var(--color-aviation-amber)',
              opacity: 0.5,
            }}
          />

          {/* Radar sweep - stronger cyan gradient */}
          <div
            className="absolute top-1/2 left-1/2 w-1/2 h-1.5 origin-left rounded-full"
            style={{
              transform: `rotate(${radarRotation}deg)`,
              background:
                'linear-gradient(90deg, var(--color-aviation-cyan) 0%, var(--color-aviation-cyan-glow) 50%, transparent 100%)',
              boxShadow: '0 0 10px var(--color-aviation-cyan-glow)',
            }}
          />

          {/* Radar blips - cyan color with stronger glow */}
          {blips.map((blip) => (
            <div
              key={blip.id}
              className="absolute w-2.5 h-2.5 rounded-full"
              style={{
                left: `${blip.x}%`,
                top: `${blip.y}%`,
                background: 'var(--color-aviation-cyan)',
                boxShadow:
                  '0 0 12px var(--color-aviation-cyan-glow), 0 0 20px var(--color-aviation-cyan)',
                animation: `aviation-blip 3s ease-in-out infinite`,
                animationDelay: `${blip.delay}s`,
              }}
            />
          ))}

          {/* Glow effect - more pronounced */}
          <div
            className="absolute inset-0 rounded-full"
            style={{
              background:
                'radial-gradient(circle at center, var(--color-aviation-cyan-glow) 0%, transparent 60%)',
              opacity: 0.3,
            }}
          />
        </div>

        {/* Status indicators - theme aware */}
        <div
          className={`flex items-center gap-6 mb-6 transition-all duration-700 delay-100 ${
            mounted ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}
        >
          <div className="flex items-center gap-2">
            <div className="aviation-status-led warning" />
            <span className="aviation-label">RADAR: ACTIVE</span>
          </div>
          <div className="flex items-center gap-2">
            <div
              className="aviation-status-led"
              style={{
                background: 'var(--color-aviation-text-muted)',
                boxShadow: 'none',
              }}
            />
            <span className="aviation-label">FLEET: STANDBY</span>
          </div>
          <div className="flex items-center gap-2">
            <Satellite className="w-3 h-3" style={{ color: 'var(--color-aviation-text-muted)' }} />
            <span className="aviation-label">LINK: ONLINE</span>
          </div>
        </div>

        {/* Main message - higher contrast text */}
        <h3
          className={`text-xl md:text-2xl font-bold mb-3 font-mono tracking-tight transition-all duration-700 delay-200 ${
            mounted ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}
          style={{ color: 'var(--color-aviation-text-primary)' }}
        >
          {isSearch ? (
            <span className="flex items-center gap-2">
              <Radar className="w-5 h-5" style={{ color: 'var(--color-aviation-amber)' }} />
              NO SIGNAL MATCHING &ldquo;{searchQuery}&rdquo;
            </span>
          ) : (
            <span className="flex items-center gap-2">
              <Radar className="w-5 h-5" style={{ color: 'var(--color-aviation-amber)' }} />
              AIRSPACE CLEAR - NO ACTIVE FUNCTIONS
            </span>
          )}
        </h3>

        <p
          className={`text-sm md:text-base max-w-md text-center mb-8 font-mono transition-all duration-700 delay-300 ${
            mounted ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}
          style={{ color: 'var(--color-aviation-text-secondary)' }}
        >
          {isSearch
            ? 'Adjust search parameters or clear filters to expand scan range'
            : 'Your fleet is ready for takeoff. Sign in to see your functions or deploy your first edge function.'}
        </p>

        {/* CTA Button */}
        {!isSearch && (
          <div
            className={`transition-all duration-700 delay-400 ${
              mounted ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
            }`}
          >
            <Button
              onClick={onDeploy}
              className="aviation-button aviation-button-primary gap-2 text-sm"
            >
              <Zap className="w-4 h-4" />
              INITIATE DEPLOYMENT
            </Button>
          </div>
        )}

        {/* Technical readout - theme aware */}
        <div
          className={`mt-8 grid grid-cols-3 gap-8 text-center transition-all duration-700 delay-500 ${
            mounted ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}
        >
          <div>
            <div className="aviation-label mb-1">STATUS</div>
            <div
              className="font-mono text-xs font-semibold"
              style={{ color: 'var(--color-aviation-amber)' }}
            >
              READY
            </div>
          </div>
          <div>
            <div className="aviation-label mb-1">REGION</div>
            <div className="aviation-value-cyan font-mono text-xs font-semibold">GLOBAL</div>
          </div>
          <div>
            <div className="aviation-label mb-1">LATENCY</div>
            <div
              className="font-mono text-xs font-semibold"
              style={{ color: 'var(--color-aviation-amber)' }}
            >
              &lt;50ms
            </div>
          </div>
        </div>
      </div>

      {/* Corner decorations */}
      <div className="aviation-instrument-corner aviation-instrument-corner-tl" />
      <div className="aviation-instrument-corner aviation-instrument-corner-tr" />
      <div className="aviation-instrument-corner aviation-instrument-corner-bl" />
      <div className="aviation-instrument-corner aviation-instrument-corner-br" />

      {/* Decorative scanline - theme aware cyan */}
      <div
        className="absolute inset-0 pointer-events-none overflow-hidden"
        style={{
          background:
            'linear-gradient(180deg, transparent 0%, var(--color-aviation-cyan-subtle) 50%, transparent 100%)',
          animation: 'aviation-scan 4s linear infinite',
        }}
      />
    </div>
  );
}
