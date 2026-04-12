import type { LucideIcon } from 'lucide-react';

export interface TabDefinition {
  id: string;
  label: string;
  icon: LucideIcon;
}

export interface UsageLimits {
  requestLimit: number;
  isUnlimited: boolean;
  isOverLimit: boolean;
  usagePercent: number;
  remaining: number | null;
}
