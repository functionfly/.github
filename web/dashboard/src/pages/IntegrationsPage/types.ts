import { Integration, IntegrationCategory } from './data/integrations';

export interface CategoryColors {
  [key: string]: string;
}

export const categoryColors: CategoryColors = {
  "Cloud Providers": "#3b82f6",
  "Frameworks": "#10b981",
  "Deployment Platforms": "#8b5cf6",
  "Databases": "#f59e0b",
  "APIs & Services": "#ef4444",
  "Monitoring & Analytics": "#06b6d4",
  "AI Agent Frameworks": "#a855f7",
};

export interface IntegrationFilterProps {
  categories: IntegrationCategory[];
  integrations: Integration[];
  activeFilter: string | null;
  onFilterChange: (category: string | null) => void;
  categoryColors: CategoryColors;
}

export interface IntegrationCardProps {
  integration: Integration;
  isExpanded: boolean;
  onToggleExpansion: (integrationId: string) => void;
}

export interface IntegrationGridProps {
  integrations: Integration[];
  expandedIntegrations: Set<string>;
  onToggleExpansion: (integrationId: string) => void;
}

export interface CategorySectionProps {
  category: string;
  categoryIndex: number;
  integrations: Integration[];
  expandedIntegrations: Set<string>;
  onToggleExpansion: (integrationId: string) => void;
}

export interface PageHeaderProps {
  activeFilter: string | null;
  categoryColors: CategoryColors;
}

export interface CTASectionProps {
  activeFilter: string | null;
}