import React, { type ReactNode } from 'react';
import { LightSourceProvider } from './LightSourceProvider';
import { GlassPanel } from './GlassPanel';
import { ChromaticEdge } from './ChromaticEdge';
import { HolographicBackdrop } from './HolographicBackdrop';
import { ScrollParallaxCamera } from './ScrollParallaxCamera';

export interface HSHeroProps {
  children: ReactNode;
  /** Optional Spline scene URL/object for the ambient backdrop. */
  scene?: string | object;
  /** Enable subtle scroll-linked parallax for refractive children. Default false. */
  parallax?: boolean;
  /** Optional className for the outer wrapper */
  className?: string;
  /** Optional inline style */
  style?: React.CSSProperties;
}

/**
 * HSHero - the canonical hero wrapper for the marketing site.
 * Composes:
 *  - LightSourceProvider (the single light source for the hero)
 *  - HolographicBackdrop (Spline or static fallback; low-GPU/reduced-motion
 *    users get the static version via ReducedMotionGate inside the backdrop)
 *  - Optional ScrollParallaxCamera wrapper for refractive content
 *
 * Usage: <HSHero><YourHeroContent /></HSHero>
 */
export function HSHero({
  children,
  scene,
  parallax = false,
  className,
  style,
}: HSHeroProps): React.ReactElement {
  return (
    <LightSourceProvider>
      <div
        className={className}
        style={{
          position: 'relative',
          isolation: 'isolate',
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          ...style,
        }}
      >
        <HolographicBackdrop
          scene={scene}
          intensity="subtle"
          style={{ position: 'absolute', inset: 0, zIndex: 'var(--hs-z-backdrop)' }}
        />
        {parallax ? (
          <ScrollParallaxCamera style={{ position: 'relative', zIndex: 0, width: '100%' }}>
            {children}
          </ScrollParallaxCamera>
        ) : (
          children
        )}
      </div>
    </LightSourceProvider>
  );
}

export interface HSFeatureCardProps {
  children: ReactNode;
  accent?: 'flame' | 'cyan' | 'strat' | 'taxiway' | 'beacon' | 'afterburner' | 'default';
  interactive?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

const accentMap: Record<NonNullable<HSFeatureCardProps['accent']>, string> = {
  flame: 'var(--hs-accent-warm)',
  cyan: 'var(--hs-accent)',
  strat: '#8b5cf6',
  taxiway: '#f59e0b',
  beacon: '#10b981',
  afterburner: '#ef4444',
  default: 'var(--hs-accent)',
};

/**
 * HSFeatureCard - GlassPanel wrapped card with chromatic edge and
 * optional hover lift. Replaces ff-card-feature in the homepage.
 */
export function HSFeatureCard({
  children,
  accent = 'default',
  interactive = true,
  className,
  style,
}: HSFeatureCardProps): React.ReactElement {
  return (
    <GlassPanel
      noise
      radius="lg"
      className={className}
      style={{
        padding: '2rem',
        cursor: interactive ? 'pointer' : undefined,
        transition: 'transform 300ms var(--hs-ease-spring), box-shadow 300ms var(--hs-ease-out)',
        ...style,
      }}
    >
      <ChromaticEdge
        style={{
          // Per-card accent on the chromatic edge
          ['--hs-chromatic-accent' as string]: accentMap[accent],
        }}
      >
        {children}
      </ChromaticEdge>
    </GlassPanel>
  );
}

export interface HSStatProps {
  end: number;
  suffix?: string;
  label: string;
  duration?: number;
}

/**
 * HSStat - glass-wrapped animated stat (numbers + label) for hero/stats rows.
 * Numbers use mono font for "tech/refractive" feel.
 */
export function HSStat({ end, suffix = '', label, duration = 2000 }: HSStatProps): React.ReactElement {
  const [count, setCount] = React.useState(0);
  const ref = React.useRef<HTMLDivElement>(null);
  const started = React.useRef(false);

  React.useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && !started.current) {
          started.current = true;
          const startTime = performance.now();
          const animate = (now: number) => {
            const elapsed = now - startTime;
            const progress = Math.min(elapsed / duration, 1);
            const eased = 1 - Math.pow(1 - progress, 3);
            setCount(Math.floor(eased * end));
            if (progress < 1) requestAnimationFrame(animate);
          };
          requestAnimationFrame(animate);
        }
      },
      { threshold: 0.5 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [end, duration]);

  return (
    <GlassPanel
      depth={0}
      radius="md"
      style={{
        padding: '1.25rem 1.5rem',
        textAlign: 'center',
        minWidth: 160,
      }}
    >
      <div
        ref={ref}
        style={{
          fontFamily: 'var(--font-mono, ui-monospace, monospace)',
          fontSize: '1.75rem',
          fontWeight: 700,
          color: 'var(--hs-text)',
          lineHeight: 1.1,
          letterSpacing: '-0.02em',
        }}
      >
        {count.toLocaleString()}
        {suffix}
      </div>
      <div
        style={{
          fontSize: '0.75rem',
          color: 'var(--hs-text-dim)',
          textTransform: 'uppercase',
          letterSpacing: '0.08em',
          marginTop: '0.5rem',
        }}
      >
        {label}
      </div>
    </GlassPanel>
  );
}
