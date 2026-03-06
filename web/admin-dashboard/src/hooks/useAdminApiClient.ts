/**
 * Admin API Client Hook
 */

import { useCallback } from 'react';
import { adminApiClient } from '@/lib/api/adminClient';

export function useAdminApiClient() {
  return useCallback(() => adminApiClient, []);
}
