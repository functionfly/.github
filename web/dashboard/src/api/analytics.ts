import { apiClient } from './client';
import type { AnalyticsSettings, UpdateAnalyticsRequest, UpdateAnalyticsResponse } from '@/types';
import { analyticsSettingsSchema, updateAnalyticsRequestSchema, updateAnalyticsResponseSchema } from '../lib/api-validation';

// Get current analytics settings
export async function getAnalyticsSettings(): Promise<AnalyticsSettings> {
  return await apiClient.getValidatedData(analyticsSettingsSchema, '/v1/admin/analytics') as AnalyticsSettings;
}

// Update analytics settings
export async function updateAnalyticsSettings(
  settings: UpdateAnalyticsRequest
): Promise<UpdateAnalyticsResponse> {
  // Validate input data
  const validation = updateAnalyticsRequestSchema.safeParse(settings);
  if (!validation.success) {
    throw new Error(`Invalid analytics settings: ${validation.error.message}`);
  }
  return await apiClient.patchValidatedData(updateAnalyticsResponseSchema, '/v1/admin/analytics', settings) as UpdateAnalyticsResponse;
}