import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import { type SecurityScan, type Vulnerability, type VulnerabilityStatus } from '@/components/ContainerScan/types';
import { z } from 'zod';

// API Response schemas
const vulnerabilitySchema = z.object({
  id: z.string(),
  title: z.string(),
  description: z.string(),
  severity: z.enum(['critical', 'high', 'medium', 'low', 'info']),
  cvss_score: z.number().optional(),
  cve: z.string().optional(),
  category: z.string(),
  component: z.string(),
  location: z.string().optional(),
  status: z.enum(['open', 'fixed', 'accepted', 'false_positive']),
  remediation: z.string().optional(),
  reference_urls: z.array(z.string()).optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  discovered: z.string(),
  updated: z.string(),
});

const scanSummarySchema = z.object({
  total_vulnerabilities: z.number(),
  critical_count: z.number(),
  high_count: z.number(),
  medium_count: z.number(),
  low_count: z.number(),
  info_count: z.number(),
  risk_score: z.number(),
  coverage_percentage: z.number(),
  compliance_score: z.number().optional(),
});

const securityScanSchema = z.object({
  id: z.string(),
  type: z.string(),
  status: z.enum(['running', 'completed', 'failed']),
  target: z.string(),
  started_at: z.string(),
  completed_at: z.string().optional(),
  duration: z.number().optional(),
  vulnerabilities: z.array(vulnerabilitySchema),
  summary: scanSummarySchema,
});

// API Types
export interface TriggerScanRequest {
  target?: string;
  scan_types?: Array<'container' | 'dockerfile' | 'compose' | 'runtime' | 'image'>;
}

export interface UpdateVulnerabilityRequest {
  status: VulnerabilityStatus;
  notes?: string;
}

// API Functions
export const containerScanApi = {
  // Get the latest container security scan
  getLatestScan: async (): Promise<SecurityScan | null> => {
    try {
      const response = await apiClient.get<{ scan: SecurityScan | null }>('/v1/admin/security/scans/container/latest');
      return response.scan;
    } catch (error) {
      console.warn('Failed to fetch latest scan:', error);
      return null;
    }
  },

  // Get a specific scan by ID
  getScan: async (scanId: string): Promise<SecurityScan> => {
    const response = await apiClient.get<{ scan: SecurityScan }>(`/v1/admin/security/scans/${scanId}`);
    return securityScanSchema.parse(response.scan);
  },

  // List all container scans
  listScans: async (limit: number = 10): Promise<SecurityScan[]> => {
    const response = await apiClient.get<{ scans: SecurityScan[] }>(`/v1/admin/security/scans?type=container&limit=${limit}`);
    return z.array(securityScanSchema).parse(response.scans);
  },

  // Trigger a new container scan
  triggerScan: async (request: TriggerScanRequest = {}): Promise<{ scan_id: string }> => {
    const response = await apiClient.post<{ scan_id: string }>('/v1/admin/security/scans', {
      type: 'container',
      target: request.target || 'all',
      scan_types: request.scan_types || ['container', 'dockerfile', 'compose', 'runtime'],
    });
    return response;
  },

  // Update vulnerability status
  updateVulnerability: async (vulnId: string, request: UpdateVulnerabilityRequest): Promise<Vulnerability> => {
    const response = await apiClient.patch<{ vulnerability: Vulnerability }>(
      `/v1/admin/security/vulnerabilities/${vulnId}`,
      request
    );
    return vulnerabilitySchema.parse(response.vulnerability);
  },

  // Export scan results
  exportScan: async (scanId: string, format: 'json' | 'csv' | 'sarif' = 'json'): Promise<string> => {
    const response = await apiClient.get<{ download_url: string }>(
      `/v1/admin/security/scans/${scanId}/export?format=${format}`
    );
    return response.download_url;
  },
};

// React Query Hooks
export function useContainerScan(scanId?: string) {
  return useQuery({
    queryKey: ['container-scan', scanId],
    queryFn: () => scanId ? containerScanApi.getScan(scanId) : containerScanApi.getLatestScan(),
    refetchInterval: (query) => {
      const data = query.state.data as SecurityScan | null;
      // Poll every 5 seconds while scan is running
      if (data?.status === 'running') return 5000;
      return false;
    },
  });
}

export function useLatestContainerScan() {
  return useQuery({
    queryKey: ['container-scan', 'latest'],
    queryFn: () => containerScanApi.getLatestScan(),
    refetchInterval: (query) => {
      const data = query.state.data as SecurityScan | null;
      // Poll every 5 seconds while scan is running
      if (data?.status === 'running') return 5000;
      return false;
    },
  });
}

export function useContainerScans(limit: number = 10) {
  return useQuery({
    queryKey: ['container-scans', limit],
    queryFn: () => containerScanApi.listScans(limit),
  });
}

export function useTriggerScan() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: containerScanApi.triggerScan,
    onSuccess: () => {
      // Invalidate latest scan to trigger polling
      queryClient.invalidateQueries({ queryKey: ['container-scan', 'latest'] });
    },
  });
}

export function useUpdateVulnerability() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ vulnId, status, notes }: { vulnId: string; status: VulnerabilityStatus; notes?: string }) =>
      containerScanApi.updateVulnerability(vulnId, { status, notes }),
    onSuccess: (data) => {
      // Invalidate affected queries
      queryClient.invalidateQueries({ queryKey: ['container-scan'] });
    },
  });
}

export function useExportScan() {
  return useMutation({
    mutationFn: ({ scanId, format }: { scanId: string; format: 'json' | 'csv' | 'sarif' }) =>
      containerScanApi.exportScan(scanId, format),
  });
}
