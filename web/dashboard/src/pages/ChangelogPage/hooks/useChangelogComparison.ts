import { useState } from 'react';
import { ChangelogEntry } from '@/api/content';
import { findDifferences, getAvailableVersions } from '../utils/changelogUtils';

export const useChangelogComparison = (changelogEntries: ChangelogEntry[]) => {
  // Version comparison states
  const [compareVersion1, setCompareVersion1] = useState<string>('');
  const [compareVersion2, setCompareVersion2] = useState<string>('');

  // Get entries for comparison
  const getComparisonData = () => {
    if (!compareVersion1 || !compareVersion2) return null;

    const entry1 = changelogEntries.find(e => e.version === compareVersion1);
    const entry2 = changelogEntries.find(e => e.version === compareVersion2);

    if (!entry1 || !entry2) return null;

    return { entry1, entry2 };
  };

  const comparisonData = getComparisonData();
  const differences = comparisonData ? findDifferences(comparisonData.entry1, comparisonData.entry2) : null;

  return {
    compareVersion1,
    setCompareVersion1,
    compareVersion2,
    setCompareVersion2,
    availableVersions: getAvailableVersions(changelogEntries),
    comparisonData,
    differences
  };
};