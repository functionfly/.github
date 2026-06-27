import { apiClient } from '@/api/client';
import { enterpriseAuditApi } from '@/api/enterprise';
import { tokenVault } from '@/utils/token-vault';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useCallback } from 'react';

export const enterpriseAuditKeys = {
  all: ['enterpriseAudit'] as const,
  logs: (params?: Record<string, unknown>) => [...enterpriseAuditKeys.all, 'logs', params] as const,
  filters: () => [...enterpriseAuditKeys.all, 'filters'] as const,
};

export interface AuditLogParams {
  limit?: number;
  offset?: number;
  service_area?: string;
  action?: string;
  resource_type?: string;
  resource_id?: string;
  actor_type?: string;
  actor_id?: string;
  success?: boolean;
  start_time?: string;
  end_time?: string;
  search?: string;
  [key: string]: unknown;
}

export function useEnterpriseAuditLogs(params?: AuditLogParams) {
  return useQuery({
    queryKey: enterpriseAuditKeys.logs(params),
    queryFn: () => enterpriseAuditApi.listLogs(params),
    staleTime: 1000 * 30,
  });
}

export function useEnterpriseAuditFilters() {
  return useQuery({
    queryKey: enterpriseAuditKeys.filters(),
    queryFn: () => enterpriseAuditApi.getFilters(),
    staleTime: 1000 * 60 * 5,
  });
}

async function getAuthToken(): Promise<string> {
  await tokenVault.initialize();
  const token = await tokenVault.getAccessToken();
  if (token) return token;
  return localStorage.getItem('ff-access-token') || '';
}

export interface AuditExportParams {
  from?: string;
  to?: string;
  format?: 'json' | 'csv' | 'cef';
  service_area?: string;
  action?: string;
}

export interface AuditExportResult {
  format: 'json' | 'csv' | 'cef';
  row_count: number;
  generated_at: string;
  hmac_sha256: string;
  body: Blob;
}

export function useExportEnterpriseAudit() {
  return useMutation({
    mutationFn: async ({ from, to, format, service_area, action }: AuditExportParams) => {
      const params = new URLSearchParams();
      if (from) params.set('from', from);
      if (to) params.set('to', to);
      if (format) params.set('format', format);
      if (service_area) params.set('service_area', service_area);
      if (action) params.set('action', action);
      const url = `/v1/enterprise/audit/export?${params.toString()}`;
      const token = await getAuthToken();
      const response = await fetch(apiClient.getBaseUrl() + url, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const result: AuditExportResult = {
        format: (format ?? 'json') as 'json' | 'csv' | 'cef',
        row_count: parseInt(response.headers.get('X-Audit-Row-Count') ?? '0', 10),
        generated_at: response.headers.get('X-Audit-Generated-At') ?? new Date().toISOString(),
        hmac_sha256: response.headers.get('X-Audit-Signature') ?? '',
        body: await response.blob(),
      };
      return result;
    },
  });
}

export function useDownloadEnterpriseAuditExport() {
  const mutation = useExportEnterpriseAudit();
  return useCallback(
    async (params: AuditExportParams, filename: string): Promise<AuditExportResult | null> => {
      const result = await mutation.mutateAsync(params);
      const url = URL.createObjectURL(result.body);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
      return result;
    },
    [mutation]
  );
}
