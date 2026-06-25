import React from 'react';
import { cn } from '@/lib/utils';

interface GaugeItemProps {
  value: string | number;
  label: string;
  variant?: 'default' | 'accent' | 'warning';
  tick?: boolean;
}

export function GaugeItem({ value, label, variant = 'default', tick = false }: GaugeItemProps) {
  return (
    <div className="gauge-strip__item">
      {tick && <div className="gauge-strip__tick" />}
      <div
        className={cn(
          'gauge-strip__value',
          variant === 'accent' && 'gauge-strip__value--accent',
          variant === 'warning' && 'gauge-strip__value--warning'
        )}
      >
        {value}
      </div>
      <div className="gauge-strip__label">{label}</div>
    </div>
  );
}

interface GaugeStripProps {
  items: GaugeItemProps[];
  className?: string;
}

export function GaugeStrip({ items, className = '' }: GaugeStripProps) {
  return (
    <div className={cn('gauge-strip', className)}>
      {items.map((item, index) => (
        <GaugeItem key={index} {...item} />
      ))}
    </div>
  );
}
