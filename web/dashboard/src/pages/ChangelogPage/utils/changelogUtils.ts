import { Shield, Bug, Sparkles, Zap, CheckCircle, AlertTriangle } from 'lucide-react';
import { ChangelogEntry } from '@/api/content';

export const getBadgeVariant = (type: string) => {
  switch (type) {
    case 'major':
      return 'default';
    case 'minor':
      return 'secondary';
    case 'patch':
      return 'outline';
    default:
      return 'outline';
  }
};

export const getTypeColor = (type: string) => {
  switch (type) {
    case 'major':
      return 'bg-gradient-to-r from-green-500 to-emerald-500 text-white border-green-500/20';
    case 'minor':
      return 'bg-gradient-to-r from-blue-500 to-cyan-500 text-white border-blue-500/20';
    case 'patch':
      return 'bg-gradient-to-r from-yellow-500 to-orange-500 text-white border-yellow-500/20';
    default:
      return 'bg-gradient-to-r from-gray-500 to-slate-500 text-white border-gray-500/20';
  }
};

export const getIconComponent = (iconName: string) => {
  switch (iconName) {
    case 'Shield':
      return Shield;
    case 'Bug':
      return Bug;
    case 'Sparkles':
      return Sparkles;
    case 'Zap':
      return Zap;
    case 'CheckCircle':
      return CheckCircle;
    case 'AlertTriangle':
      return AlertTriangle;
    default:
      return CheckCircle;
  }
};

export const findDifferences = (entry1: ChangelogEntry, entry2: ChangelogEntry) => {
  const differences = {
    newCategories: [] as string[],
    removedCategories: [] as string[],
    changedCategories: [] as string[],
    categoryDetails: {} as Record<string, { added: string[], removed: string[], unchanged: string[] }>
  };

  const categories1 = entry1.changes.map(c => c.category);
  const categories2 = entry2.changes.map(c => c.category);

  // Find new and removed categories
  differences.newCategories = categories2.filter(cat => !categories1.includes(cat));
  differences.removedCategories = categories1.filter(cat => !categories2.includes(cat));

  // Find categories that exist in both
  const commonCategories = categories1.filter(cat => categories2.includes(cat));
  differences.changedCategories = commonCategories;

  // Analyze changes within common categories
  commonCategories.forEach(category => {
    const change1 = entry1.changes.find(c => c.category === category);
    const change2 = entry2.changes.find(c => c.category === category);

    if (change1 && change2) {
      const items1 = change1.items;
      const items2 = change2.items;

      differences.categoryDetails[category] = {
        added: items2.filter(item => !items1.includes(item)),
        removed: items1.filter(item => !items2.includes(item)),
        unchanged: items1.filter(item => items2.includes(item))
      };
    }
  });

  return differences;
};

export const getAvailableVersions = (changelogEntries: ChangelogEntry[]) => {
  return changelogEntries.map(entry => ({
    value: entry.version,
    label: `${entry.version} (${new Date(entry.date).toLocaleDateString()})`
  }));
};