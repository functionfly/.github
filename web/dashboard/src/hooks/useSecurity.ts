import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { securityApi, type SecurityMetrics, type SecurityIncident } from '@/api/security';

// Query keys
export const securityKeys = {
  all: ['security'] as const,
  metrics: () => [...securityKeys.all, 'metrics'] as const,
  services: () => [...securityKeys.all, 'services'] as const,
  certificates: () => [...securityKeys.all, 'certificates'] as const,
  incidents: (limit?: number) => [...securityKeys.all, 'incidents', limit] as const,
  compliance: () => [...securityKeys.all, 'compliance'] as const,
  measures: () => [...securityKeys.all, 'measures'] as const,
  incidentResponse: () => [...securityKeys.all, 'incident-response'] as const,
  faq: () => [...securityKeys.all, 'faq'] as const,
  resources: () => [...securityKeys.all, 'resources'] as const,
  contacts: () => [...securityKeys.all, 'contacts'] as const,
};

// Get comprehensive security metrics
export function useSecurityMetrics() {
  return useQuery({
    queryKey: securityKeys.metrics(),
    queryFn: () => securityApi.getSecurityMetrics(),
    staleTime: 1000 * 60,
  });
}

// Get service status
export function useServiceStatus() {
  return useQuery({
    queryKey: securityKeys.services(),
    queryFn: () => securityApi.getServiceStatus(),
    staleTime: 1000 * 60,
    refetchInterval: 60000, // Refresh every minute
  });
}

// Get SSL certificates
export function useSSLCertificates() {
  return useQuery({
    queryKey: securityKeys.certificates(),
    queryFn: () => securityApi.getSSLCertificates(),
    staleTime: 1000 * 60 * 5,
  });
}

// Get recent incidents
export function useSecurityIncidents(limit?: number) {
  return useQuery({
    queryKey: securityKeys.incidents(limit),
    queryFn: () => securityApi.getRecentIncidents(limit),
    staleTime: 1000 * 60,
  });
}

// Get compliance frameworks
export function useComplianceFrameworks() {
  return useQuery({
    queryKey: securityKeys.compliance(),
    queryFn: () => securityApi.getComplianceFrameworks(),
    staleTime: 1000 * 60 * 10,
  });
}

// Get security measures
export function useSecurityMeasures() {
  return useQuery({
    queryKey: securityKeys.measures(),
    queryFn: () => securityApi.getSecurityMeasures(),
    staleTime: 1000 * 60 * 10,
  });
}

// Get incident response procedures
export function useIncidentResponse() {
  return useQuery({
    queryKey: securityKeys.incidentResponse(),
    queryFn: () => securityApi.getIncidentResponse(),
    staleTime: 1000 * 60 * 60,
  });
}

// Get security FAQ
export function useSecurityFAQ() {
  return useQuery({
    queryKey: securityKeys.faq(),
    queryFn: () => securityApi.getSecurityFAQ(),
    staleTime: 1000 * 60 * 10,
  });
}

// Get security resources
export function useSecurityResources() {
  return useQuery({
    queryKey: securityKeys.resources(),
    queryFn: () => securityApi.getSecurityResources(),
    staleTime: 1000 * 60 * 60,
  });
}

// Get contact information
export function useSecurityContacts() {
  return useQuery({
    queryKey: securityKeys.contacts(),
    queryFn: () => securityApi.getContactInfo(),
    staleTime: 1000 * 60 * 60,
  });
}
