import React from 'react';

interface FounderBadgeProps {
  founderNumber: number | null;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
  tier?: 'founder' | 'founder_pro' | 'founder_elite';
}

export function FounderBadge({
  founderNumber,
  size = 'md',
  showLabel = true,
  tier = 'founder',
}: FounderBadgeProps) {
  if (founderNumber === null || founderNumber === undefined) {
    return null;
  }

  const sizeClasses = {
    sm: 'text-xs px-1.5 py-0.5',
    md: 'text-sm px-2 py-1',
    lg: 'text-base px-3 py-1.5',
  };

  const iconSizes = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  };

  const tierStyles = {
    founder: {
      bg: 'linear-gradient(135deg, #f59e0b 0%, #d97706 50%, #b45309 100%)',
      border: '#fbbf24',
      shadow: '0 2px 8px rgba(245, 158, 11, 0.4)',
      hoverShadow: '0 4px 16px rgba(245, 158, 11, 0.6)',
    },
    founder_pro: {
      bg: 'linear-gradient(135deg, #e5e7eb 0%, #9ca3af 50%, #6b7280 100%)',
      border: '#f3f4f6',
      shadow: '0 2px 8px rgba(156, 163, 175, 0.4)',
      hoverShadow: '0 4px 16px rgba(156, 163, 175, 0.6)',
    },
    founder_elite: {
      bg: 'linear-gradient(135deg, #fef3c7 0%, #fcd34d 25%, #f59e0b 50%, #d97706 75%, #92400e 100%)',
      border: '#fcd34d',
      shadow: '0 2px 8px rgba(251, 191, 36, 0.4), inset 0 1px 0 rgba(255,255,255,0.4)',
      hoverShadow: '0 6px 20px rgba(251, 191, 36, 0.7), inset 0 1px 0 rgba(255,255,255,0.4)',
    },
  };

  const style = tierStyles[tier];

  return (
    <span
      className={`sc-founder-badge transition-all duration-300 hover:scale-105 cursor-default ${sizeClasses[size]}`}
      title={`Founder #${founderNumber} — permanent member`}
      style={{
        background: style.bg,
        color: tier === 'founder_pro' ? '#1f2937' : '#ffffff',
        border: `2px solid ${style.border}`,
        display: 'inline-flex',
        alignItems: 'center',
        gap: '0.35rem',
        padding: '0.25rem 0.75rem',
        borderRadius: '6px',
        fontWeight: 700,
        boxShadow: style.shadow,
        textShadow: tier === 'founder_elite' ? '0 1px 2px rgba(0,0,0,0.3)' : 'none',
      }}
    >
      <svg
        className={iconSizes[size]}
        viewBox="0 0 24 24"
        fill="currentColor"
        style={{
          filter: tier === 'founder_elite' ? 'drop-shadow(0 1px 1px rgba(0,0,0,0.3))' : 'none',
        }}
      >
        <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
      </svg>
      {showLabel && (
        <div className="flex items-center gap-1.5">
          {tier === 'founder_pro' && (
            <span className="text-xs font-bold uppercase tracking-wider opacity-80">PRO</span>
          )}
          {tier === 'founder_elite' && (
            <span className="text-xs font-bold uppercase tracking-wider opacity-90">ELITE</span>
          )}
          <span>Founders #{founderNumber}</span>
          <span className="text-xs opacity-75">• Permanent</span>
        </div>
      )}
    </span>
  );
}

interface FounderBadgeInlineProps {
  founderNumber: number | null;
}

export function FounderBadgeInline({ founderNumber }: FounderBadgeInlineProps) {
  if (founderNumber === null || founderNumber === undefined) {
    return null;
  }

  return (
    <span className="text-amber-600 font-bold">
      #{founderNumber}
    </span>
  );
}
