import { useState } from 'react';
import { ChangelogEntry } from '@/api/content';
import { getAvailableVersions } from '../utils/changelogUtils';

export const useWhatsNew = (changelogEntries: ChangelogEntry[]) => {
  const [whatsNewVersion, setWhatsNewVersion] = useState<string>('');

  // Get "what's new since" data
  const getWhatsNewData = () => {
    if (!whatsNewVersion) return [];

    const baseVersion = changelogEntries.find(e => e.version === whatsNewVersion);
    if (!baseVersion) return [];

    const baseIndex = changelogEntries.findIndex(e => e.version === whatsNewVersion);
    return changelogEntries.slice(0, baseIndex); // All entries newer than the base version
  };

  const availableVersions = getAvailableVersions(changelogEntries).slice(1); // Skip the latest version

  return {
    whatsNewVersion,
    setWhatsNewVersion,
    whatsNewData: getWhatsNewData(),
    availableVersions
  };
};