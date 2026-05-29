import type { LucideIcon } from 'lucide-react';
import { Bot, CreditCard, FunctionSquare, Key } from 'lucide-react';

export const PAGE_SIZE = 20;

export type PurchaseKind = 'function' | 'agent' | 'license' | 'subscription';
export type PurchaseTab = 'all' | PurchaseKind | 'timeline';
export type ViewMode = 'cards' | 'table';
export type StatusFilter = 'all' | 'active' | 'inactive' | 'revoked';
export type DateRangeFilter = 'all' | '7d' | '30d' | '90d';

export const DATE_RANGE_OPTIONS: { value: DateRangeFilter; labelKey: string }[] = [
  { value: 'all', labelKey: 'purchasesPage.dateAll' },
  { value: '7d', labelKey: 'purchasesPage.date7d' },
  { value: '30d', labelKey: 'purchasesPage.date30d' },
  { value: '90d', labelKey: 'purchasesPage.date90d' },
];

export const STATUS_FILTER_OPTIONS: { value: StatusFilter; labelKey: string }[] = [
  { value: 'all', labelKey: 'purchasesPage.statusAll' },
  { value: 'active', labelKey: 'purchasesPage.statusActive' },
  { value: 'inactive', labelKey: 'purchasesPage.statusInactive' },
  { value: 'revoked', labelKey: 'purchasesPage.statusRevoked' },
];

export const KIND_META: Record<
  PurchaseKind,
  {
    accent: string;
    iconBg: string;
    icon: LucideIcon;
    tabKey: string;
  }
> = {
  function: {
    accent: 'border-l-aviation-cyan',
    iconBg: 'bg-aviation-cyan/10 text-aviation-cyan',
    icon: FunctionSquare,
    tabKey: 'purchasesPage.tabFunctions',
  },
  agent: {
    accent: 'border-l-aviation-amber',
    iconBg: 'bg-aviation-amber/10 text-aviation-amber',
    icon: Bot,
    tabKey: 'purchasesPage.tabAgents',
  },
  license: {
    accent: 'border-l-emerald-500',
    iconBg: 'bg-emerald-500/10 text-emerald-400',
    icon: Key,
    tabKey: 'purchasesPage.tabLicenses',
  },
  subscription: {
    accent: 'border-l-rose-500',
    iconBg: 'bg-rose-500/10 text-rose-400',
    icon: CreditCard,
    tabKey: 'purchasesPage.tabSubscriptions',
  },
};

export const TAB_ORDER: PurchaseTab[] = [
  'all',
  'timeline',
  'function',
  'agent',
  'license',
  'subscription',
];
