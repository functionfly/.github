import { Globe } from 'lucide-react';
import { PROVIDER_REGION_META, getProviderRegionGroup } from '../constants/providerMeta';

interface RegionBadgesProps {
  regions: readonly string[];
  providerId: string;
}

const GROUP_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  americas: { bg: 'bg-blue-50 dark:bg-blue-950/30', text: 'text-blue-700 dark:text-blue-400', border: 'border-blue-200 dark:border-blue-800' },
  europe: { bg: 'bg-purple-50 dark:bg-purple-950/30', text: 'text-purple-700 dark:text-purple-400', border: 'border-purple-200 dark:border-purple-800' },
  asia: { bg: 'bg-amber-50 dark:bg-amber-950/30', text: 'text-amber-700 dark:text-amber-400', border: 'border-amber-200 dark:border-amber-800' },
  oceania: { bg: 'bg-teal-50 dark:bg-teal-950/30', text: 'text-teal-700 dark:text-teal-400', border: 'border-teal-200 dark:border-teal-800' },
};

const GROUP_ICONS: Record<string, string> = {
  americas: '🌎',
  europe: '🌍',
  asia: '🌏',
  oceania: '🌏',
};

const GROUP_LABELS: Record<string, string> = {
  americas: 'Americas',
  europe: 'Europe',
  asia: 'Asia Pacific',
  oceania: 'Oceania',
};

export function RegionBadges({ regions, providerId }: RegionBadgesProps) {
  // Group regions by continent
  const groups = regions.reduce<Record<string, string[]>>((acc, region) => {
    const group = getProviderRegionGroup(providerId, region);
    if (!acc[group]) acc[group] = [];
    acc[group].push(region);
    return acc;
  }, {});

  const groupKeys = Object.keys(groups);

  return (
    <div className="flex flex-wrap items-center gap-2 text-sm">
      <Globe className="w-4 h-4 text-text-muted shrink-0" />
      <span className="text-text-secondary font-medium">{regions.length} regions</span>
      <span className="text-text-muted">•</span>
      <div className="flex flex-wrap gap-1.5">
        {groupKeys.map((group) => {
          const colors = GROUP_COLORS[group] || GROUP_COLORS.americas;
          const count = groups[group].length;
          return (
            <span
              key={group}
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium border ${colors.bg} ${colors.text} ${colors.border}`}
              title={groups[group].join(', ')}
            >
              <span>{GROUP_ICONS[group]}</span>
              <span>{GROUP_LABELS[group]}</span>
              <span className="opacity-60">({count})</span>
            </span>
          );
        })}
      </div>
    </div>
  );
}

interface RegionListProps {
  regions: readonly string[];
  providerId: string;
  maxDisplay?: number;
}

export function RegionList({ regions, providerId, maxDisplay = 3 }: RegionListProps) {
  const displayRegions = regions.slice(0, maxDisplay);
  const remaining = regions.length - maxDisplay;

  return (
    <div className="flex items-center gap-2 text-sm text-text-secondary">
      <Globe className="w-4 h-4 text-text-muted" />
      <span className="font-medium">{regions.length} regions</span>
      <span className="text-text-muted">•</span>
      <span className="text-text-muted truncate">
        {displayRegions.join(', ')}
        {remaining > 0 && ` +${remaining} more`}
      </span>
    </div>
  );
}
