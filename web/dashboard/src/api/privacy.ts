import { apiClient } from './client';

export interface PrivacySettings {
  id?: string;
  privacy_level: 'standard' | 'enhanced' | 'maximum' | 'gdpr';
  anonymize_ip: boolean;
  anonymize_user_agent: boolean;
  log_geo_data: boolean;
  log_embed_origin: boolean;
  store_input_output: boolean;
  retention_days: number;
  gdpr_mode: boolean;
  auto_delete_enabled: boolean;
  consent_required: boolean;
  consent_given_at?: string;
  consent_version?: string;
  ip_mask_type: 'none' | 'hash' | 'partial' | 'redact';
  user_agent_mask_type: 'none' | 'hash' | 'partial' | 'redact';
  created_at?: string;
  updated_at?: string;
}

export interface DataExportRequest {
  id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  request_type: string;
  requested_at: string;
  completed_at?: string;
  expires_at?: string;
  download_url?: string;
  download_token?: string;
  file_size?: number;
  record_count?: number;
  error_message?: string;
}

export interface DataDeletionRequest {
  id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed' | 'partial';
  request_type: string;
  requested_at: string;
  completed_at?: string;
  records_deleted?: number;
  records_anonymized?: number;
  error_message?: string;
}

export const privacyApi = {
  getSettings: () =>
    apiClient.get<PrivacySettings>('/v1/privacy/settings'),

  updateSettings: (data: Partial<PrivacySettings>) =>
    apiClient.put<PrivacySettings>('/v1/privacy/settings', data),

  requestDataExport: (requestType = 'full') =>
    apiClient.post<DataExportRequest>('/v1/privacy/export', { request_type: requestType }),

  getExportStatus: (id: string) =>
    apiClient.get<DataExportRequest>(`/v1/privacy/export/${id}`),

  requestDataDeletion: (requestType = 'full') =>
    apiClient.post<DataDeletionRequest>('/v1/privacy/deletion', { request_type: requestType }),

  getDeletionStatus: (id: string) =>
    apiClient.get<DataDeletionRequest>(`/v1/privacy/deletion/${id}`),
};
