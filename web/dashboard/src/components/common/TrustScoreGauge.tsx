import React from 'react';
import { getTrustColor } from './TrustScoreBadge';

interface TrustScoreGaugeProps {
  score: number;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  showLabel?: boolean;
  animated?: boolean;
}

const sizeConfig = {
  sm: { width: 60, strokeWidth: 6, fontSize: 'text-sm' },
  md: { width: 100, strokeWidth: 8, fontSize: 'text-lg' },
  lg: { width: 140, strokeWidth: 10, fontSize: 'text-2xl' },
  xl: { width: 180, strokeWidth: 12, fontSize: 'text-3xl' },
};

export function TrustScoreGauge({
  score,
  size = 'md',
  showLabel = true,
  animated = true
}: TrustScoreGaugeProps) {
  const config = sizeConfig[size];
  const normalizedScore = Math.max(0, Math.min(100, score));
  const color = getTrustColor(normalizedScore);

  // Calculate the circumference and stroke dash
  const radius = (config.width - config.strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (normalizedScore / 100) * circumference;

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg
        width={config.width}
        height={config.width}
        className={`transform -rotate-90 ${animated ? 'animate-pulse-slow' : ''}`}
      >
        {/* Background circle */}
        <circle
          cx={config.width / 2}
          cy={config.width / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={config.strokeWidth}
          className="text-gray-200"
        />
        {/* Progress circle */}
        <circle
          cx={config.width / 2}
          cy={config.width / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={config.strokeWidth}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          className={`transition-all duration-1000 ease-out ${animated ? 'animate-progress' : ''}`}
          style={{
            filter: `drop-shadow(0 0 8px ${color}40)`,
          }}
        />
      </svg>

      {/* Center text */}
      <div className="absolute inset-0 flex items-center justify-center">
        <span className={`font-bold ${config.fontSize}`} style={{ color }}>
          {Math.round(normalizedScore)}
        </span>
      </div>

      {/* Label below */}
      {showLabel && (
        <div className="absolute -bottom-6 w-full text-center">
          <span className="text-xs text-gray-500 font-medium">Trust Score</span>
        </div>
      )}
    </div>
  );
}

interface TrustScoreBarProps {
  score: number;
  showLabel?: boolean;
  height?: 'sm' | 'md' | 'lg';
}

const barHeightConfig = {
  sm: 'h-1.5',
  md: 'h-2.5',
  lg: 'h-4',
};

export function TrustScoreBar({ score, showLabel = true, height = 'md' }: TrustScoreBarProps) {
  const normalizedScore = Math.max(0, Math.min(100, score));
  const color = getTrustColor(normalizedScore);

  return (
    <div className="w-full">
      <div className={`w-full bg-gray-200 rounded-full overflow-hidden ${barHeightConfig[height]}`}>
        <div
          className="h-full rounded-full transition-all duration-500 ease-out"
          style={{
            width: `${normalizedScore}%`,
            backgroundColor: color,
            boxShadow: `0 0 10px ${color}60`,
          }}
        />
      </div>
      {showLabel && (
        <div className="flex justify-between mt-1 text-xs text-gray-500">
          <span>0</span>
          <span>100</span>
        </div>
      )}
    </div>
  );
}
