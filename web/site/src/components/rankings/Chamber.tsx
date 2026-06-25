import React, { useRef, useState, useEffect, useCallback, type ReactNode } from 'react';

/* =============================================
   CHAMBER - Base containment panel
   ============================================= */
interface ChamberProps {
  children: ReactNode;
  ribs?: boolean;
  animate?: boolean;
  stagger?: boolean;
  className?: string;
}

export function Chamber({ 
  children, 
  ribs = false, 
  animate = false,
  stagger = false,
  className = '' 
}: ChamberProps) {
  const classes = [
    'chamber',
    ribs ? 'chamber--ribbed' : '',
    animate ? 'chamber--animate' : '',
    stagger ? 'chamber--stagger' : '',
    className
  ].filter(Boolean).join(' ');

  return (
    <div className={classes}>
      {children}
    </div>
  );
}

/* =============================================
   CORNER BRACE - L-shaped bracket detail
   ============================================= */
type CornerPosition = 'tl' | 'tr' | 'bl' | 'br';

interface CornerBraceProps {
  position: CornerPosition;
}

export function CornerBrace({ position }: CornerBraceProps) {
  return (
    <div 
      className={`corner-brace corner-brace--${position}`}
      aria-hidden="true"
    />
  );
}

/* =============================================
   STATUS PILL - Live/pending/revoked indicator
   ============================================= */
type StatusType = 'live' | 'pending' | 'revoked';

interface StatusPillProps {
  status: StatusType;
  label?: string;
}

const STATUS_LABELS: Record<StatusType, string> = {
  live: 'Live',
  pending: 'Pending',
  revoked: 'Revoked',
};

export function StatusPill({ status, label }: StatusPillProps) {
  const displayLabel = label || STATUS_LABELS[status];
  
  return (
    <span className={`status-pill status-pill--${status}`}>
      <span className="status-pill__dot" aria-hidden="true" />
      <span>{displayLabel}</span>
    </span>
  );
}

/* =============================================
   GAUGE STRIP / GAUGE - Instrumentation readouts
   ============================================= */
interface GaugeData {
  value: string | number;
  label: string;
}

interface GaugeStripProps {
  gauges: GaugeData[];
  className?: string;
}

export function GaugeStrip({ gauges, className = '' }: GaugeStripProps) {
  return (
    <div className={`gauge-strip ${className}`}>
      {gauges.map((gauge, i) => (
        <div key={i} className="gauge">
          <span className="gauge__tick" aria-hidden="true" />
          <span className="gauge__value">{gauge.value}</span>
          <span className="gauge__label">{gauge.label}</span>
        </div>
      ))}
    </div>
  );
}

interface GaugeProps {
  value: string | number;
  label: string;
  className?: string;
}

export function Gauge({ value, label, className = '' }: GaugeProps) {
  return (
    <div className={`gauge ${className}`}>
      <span className="gauge__tick" aria-hidden="true" />
      <span className="gauge__value">{value}</span>
      <span className="gauge__label">{label}</span>
    </div>
  );
}

/* =============================================
   ANNOTATION TAG - Engineering dimension callout
   ============================================= */
type TagPosition = 'tl' | 'tr' | 'bl' | 'br';

interface AnnotationTagProps {
  label: string;
  position?: TagPosition;
}

export function AnnotationTag({ label, position = 'tl' }: AnnotationTagProps) {
  return (
    <span className={`annotation-tag annotation-tag--${position}`}>
      {label}
    </span>
  );
}

/* =============================================
   SEALED BUTTON / FRAME BUTTON
   ============================================= */
interface SealedButtonProps {
  children: ReactNode;
  variant?: 'primary' | 'secondary';
  onClick?: () => void;
  className?: string;
  disabled?: boolean;
  type?: 'button' | 'submit';
}

export function SealedButton({ 
  children, 
  variant = 'primary',
  onClick,
  className = '',
  disabled = false,
  type = 'button'
}: SealedButtonProps) {
  const classes = [
    'sealed-button',
    `sealed-button--${variant}`,
    className
  ].filter(Boolean).join(' ');

  return (
    <button 
      type={type}
      className={classes}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  );
}

/* =============================================
   REDUCED MOTION GATE - Accessibility wrapper
   ============================================= */
interface ReducedMotionGateProps {
  children: ReactNode;
  fallback: ReactNode;
}

function usePrefersReducedMotion() {
  const [prefersReduced, setPrefersReduced] = useState(false);
  
  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReduced(mq.matches);
    
    const handler = (e: MediaQueryListEvent) => setPrefersReduced(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);
  
  return prefersReduced;
}

export function ReducedMotionGate({ children, fallback }: ReducedMotionGateProps) {
  const prefersReduced = usePrefersReducedMotion();
  return prefersReduced ? <>{fallback}</> : <>{children}</>;
}

/* =============================================
   TRUST SEAL - Security foil holographic badge
   ============================================= */
interface TrustSealProps {
  size?: 'sm' | 'md' | 'lg';
  showShimmer?: boolean;
  onHover?: boolean;
}

export function TrustSeal({ 
  size = 'md', 
  showShimmer = true,
  onHover = true 
}: TrustSealProps) {
  const sealRef = useRef<HTMLDivElement>(null);
  const [angle, setAngle] = useState(0);
  const [hasAnimated, setHasAnimated] = useState(false);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!onHover || !sealRef.current) return;
    
    const rect = sealRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left - rect.width / 2;
    const y = e.clientY - rect.top - rect.height / 2;
    const newAngle = Math.atan2(y, x) * (180 / Math.PI);
    setAngle(newAngle);
  }, [onHover]);

  const handleMouseEnter = useCallback(() => {
    if (showShimmer && !hasAnimated) {
      setHasAnimated(true);
    }
  }, [showShimmer, hasAnimated]);

  const gradientStyle = onHover 
    ? { background: `conic-gradient(from ${angle}deg, var(--chamber-foil-a), var(--chamber-foil-b), var(--chamber-foil-c), var(--chamber-foil-a))` }
    : undefined;

  return (
    <div 
      ref={sealRef}
      className={`trust-seal trust-seal--${size}`}
      onMouseMove={handleMouseMove}
      onMouseEnter={handleMouseEnter}
      role="img"
      aria-label="Verified trust seal"
    >
      <div 
        className={`trust-seal__outer ${hasAnimated ? 'trust-seal__outer--shimmer' : ''}`}
        style={gradientStyle}
      >
        <div className="trust-seal__inner">
          <svg 
            className="trust-seal__icon" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2" 
            strokeLinecap="round" 
            strokeLinejoin="round"
          >
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            <path d="m9 12 2 2 4-4" />
          </svg>
        </div>
      </div>
    </div>
  );
}

/* =============================================
   TRUST BADGE - Full badge with label
   ============================================= */
interface TrustBadgeProps {
  status: StatusType;
  label?: string;
  size?: 'sm' | 'md' | 'lg';
  showSeal?: boolean;
}

export function TrustBadge({ 
  status, 
  label,
  size = 'md',
  showSeal = true 
}: TrustBadgeProps) {
  return (
    <ReducedMotionGate
      fallback={
        <span className={`status-pill status-pill--${status}`}>
          {showSeal && (
            <span className="status-pill__dot" aria-hidden="true" />
          )}
          <span>{label || STATUS_LABELS[status]}</span>
        </span>
      }
    >
      <span style={{ display: 'inline-flex', alignItems: 'center', gap: '8px' }}>
        {showSeal && <TrustSeal size={size === 'sm' ? 'sm' : 'md'} showShimmer={true} onHover={true} />}
        <StatusPill status={status} label={label} />
      </span>
    </ReducedMotionGate>
  );
}

/* =============================================
   MOVEMENT BADGE - Rank change indicator
   ============================================= */
interface MovementBadgeProps {
  delta: number;
}

export function MovementBadge({ delta }: MovementBadgeProps) {
  if (delta === 0) return null;
  
  const isUp = delta > 0;
  
  return (
    <span className={`movement-badge movement-badge--${isUp ? 'up' : 'down'}`}>
      {isUp ? (
        <svg className="movement-badge__arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="m18 15-6-6-6 6" />
        </svg>
      ) : (
        <svg className="movement-badge__arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="m6 9 6 6 6-6" />
        </svg>
      )}
      {Math.abs(delta)}
    </span>
  );
}

/* =============================================
   RANK DISPLAY - With top-3 highlighting
   ============================================= */
interface RankDisplayProps {
  rank: number;
  delta?: number;
}

export function RankDisplay({ rank, delta }: RankDisplayProps) {
  const isTop3 = rank <= 3;
  
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '8px' }}>
      <span className={`rankings-table__rank ${isTop3 ? 'rankings-table__rank--top3' : ''}`}>
        #{rank}
      </span>
      {delta !== undefined && delta !== 0 && <MovementBadge delta={delta} />}
    </span>
  );
}
