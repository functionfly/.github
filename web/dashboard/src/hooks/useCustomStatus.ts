import { useState, useCallback, useEffect } from 'react';
import { usersApi } from '@/api/users';

export type CustomStatusValue = 'online' | 'away' | 'busy' | 'offline' | '';

export interface CustomStatus {
  customStatus: CustomStatusValue;
  customStatusEmoji: string;
}

export interface UseCustomStatusReturn {
  status: CustomStatus;
  isLoading: boolean;
  isUpdating: boolean;
  error: string | null;
  setStatus: (status: CustomStatusValue, emoji?: string) => Promise<void>;
  clearStatus: () => Promise<void>;
}

export function useCustomStatus(): UseCustomStatusReturn {
  const [status, setStatusState] = useState<CustomStatus>({
    customStatus: '',
    customStatusEmoji: '',
  });
  const [isLoading, setIsLoading] = useState(true);
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = useCallback(async () => {
    try {
      const response = await usersApi.getCustomStatus();
      setStatusState({
        customStatus: (response.customStatus || '') as CustomStatusValue,
        customStatusEmoji: response.customStatusEmoji || '',
      });
      setError(null);
    } catch (err) {
      console.error('Failed to fetch custom status:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch status');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchStatus();
  }, [fetchStatus]);

  const setStatus = useCallback(async (customStatus: CustomStatusValue, customStatusEmoji?: string) => {
    setIsUpdating(true);
    setError(null);
    try {
      const response = await usersApi.updateCustomStatus({ customStatus, customStatusEmoji });
      setStatusState({
        customStatus: (response.customStatus || '') as CustomStatusValue,
        customStatusEmoji: response.customStatusEmoji || '',
      });
    } catch (err) {
      console.error('Failed to update custom status:', err);
      setError(err instanceof Error ? err.message : 'Failed to update status');
      throw err;
    } finally {
      setIsUpdating(false);
    }
  }, []);

  const clearStatus = useCallback(async () => {
    await setStatus('');
  }, [setStatus]);

  return {
    status,
    isLoading,
    isUpdating,
    error,
    setStatus,
    clearStatus,
  };
}

export const CUSTOM_STATUS_OPTIONS: Array<{
  value: CustomStatusValue;
  label: string;
  emoji: string;
  description: string;
}> = [
  { value: '', label: 'Auto', emoji: '⚡', description: 'Follows your activity' },
  { value: 'online', label: 'Online', emoji: '🟢', description: 'Available for work' },
  { value: 'away', label: 'Away', emoji: '🟡', description: 'Temporarily away' },
  { value: 'busy', label: 'Busy', emoji: '🔴', description: 'Do not disturb' },
  { value: 'offline', label: 'Appear Offline', emoji: '⚫', description: 'Hidden from others' },
];
