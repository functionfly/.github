import { useState, useEffect } from 'react';
import { contentApi, ChangelogEntry } from '@/api/content';

export const useChangelogData = () => {
  const [changelogEntries, setChangelogEntries] = useState<ChangelogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchChangelog = async () => {
      try {
        setLoading(true);
        const entries = await contentApi.getPublishedChangelogEntries(50); // Fetch more entries for better filtering
        setChangelogEntries(entries);
      } catch (err) {
        console.error('Failed to fetch changelog:', err);
        setError('Failed to load changelog entries');
      } finally {
        setLoading(false);
      }
    };

    fetchChangelog();
  }, []);

  return {
    changelogEntries,
    loading,
    error
  };
};