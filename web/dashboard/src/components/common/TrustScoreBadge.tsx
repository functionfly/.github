import React from 'react';
import { Shield, CheckCircle, AlertTriangle, XCircle, Info } from 'lucide-react';

export type TrustLevel = 'excellent' | 'good' | 'fair' | 'poor' | 'very_poor' | 'insufficient_data';

interface TrustScoreBadgeProps {
  trustScore: number;
  trustLevel?: TrustLevel;
  showScore?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

const levelConfig: Record<TrustLevel, { label: string; color: string; bgColor: string; icon: React.ReactNode }> = {
  excellent: {
    label: 'Excellent',
    color: 'text-green-700',
    bgColor: 'bg-green-50',
    icon: <CheckCircle className="w-4 h-4" />,
  },
  good: {
    label: 'Good',
    color: 'text-emerald-700',
    bgColor: 'bg-emerald-50',
    icon: <Shield className="w-4 h-4" />,
  },
  fair: {
    label: 'Fair',
    color: 'text-yellow-700',
    bgColor: 'bg-yellow-50',
    icon: <AlertTriangle className="w-4 h-4" />,
  },
  poor: {
    label: 'Poor',
    color: 'text-orange-700',
    bgColor: 'bg-orange-50',
    icon: <AlertTriangle className="w-4 h-4" />,
  },
  very_poor: {
    label: 'Very Poor',
    color: 'text-red-700',
    bgColor: 'bg-red-50',
    icon: <XCircle className="w-4 h-4" />,
  },
  insufficient_data: {
    label: 'Insufficient Data',
    color: 'text-gray-700',
    bgColor: 'bg-gray-50',
    icon: <Info className="w-4 h-4" />,
  },
};

const sizeClasses = {
  sm: 'text-xs px-2 py-0.5 gap-1',
  md: 'text-sm px-3 py-1 gap-1.5',
  lg: 'text-base px-4 py-1.5 gap-2',
};

export function TrustScoreBadge({ trustScore, trustLevel, showScore = true, size = 'md' }: TrustScoreBadgeProps) {
  // Determine trust level from score if not provided
  const level: TrustLevel = trustLevel || getTrustLevel(trustScore);
  const config = levelConfig[level];

  return (
    <span
      className={`
        inline-flex items-center font-medium rounded-full
        ${config.color} ${config.bgColor}
        ${sizeClasses[size]}
      `}
    >
      {config.icon}
      {showScore && <span>{Math.round(trustScore)}</span>}
      <span>{config.label}</span>
    </span>
  );
}

export function getTrustLevel(score: number): TrustLevel {
  if (score >= 80) return 'excellent';
  if (score >= 60) return 'good';
  if (score >= 40) return 'fair';
  if (score >= 20) return 'poor';
  if (score > 0) return 'very_poor';
  return 'insufficient_data';
}

export function getTrustColor(score: number): string {
  if (score >= 80) return '#22c55e'; // green-500
  if (score >= 60) return '#10b981'; // emerald-500
  if (score >= 40) return '#eab308'; // yellow-500
  if (score >= 20) return '#f97316'; // orange-500
  if (score > 0) return '#ef4444'; // red-500
  return '#6b7280'; // gray-500
}
