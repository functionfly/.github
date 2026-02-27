import { apiClient } from "./client";

export interface ServiceStatus {
  name: string;
  status: 'operational' | 'degraded' | 'outage';
  uptime: string;
  responseTime: string;
}

export interface SSLCertificate {
  domain: string;
  issuer: string;
  expiryDate: string; // ISO date string
  status: 'valid' | 'expiring' | 'expired';
  autoRenewal: boolean;
}

export interface SecurityIncident {
  id: string;
  title: string;
  severity: 'info' | 'warning' | 'critical';
  status: 'open' | 'resolved' | 'investigating';
  timestamp: string; // ISO date string
  description: string;
  impact: string;
  duration: string;
}

export interface SecurityMetrics {
  overallScore: number;
  lastUpdated: string; // ISO date string
  services: ServiceStatus[];
  certificates: SSLCertificate[];
  recentIncidents: SecurityIncident[];
}

export interface ComplianceFramework {
  name: string;
  status: 'Certified' | 'Compliant' | 'In Progress' | 'Not Applicable';
  description: string;
  auditor: string;
  lastAudit: string;
  nextAudit: string;
}

export interface SecurityMeasure {
  category: string;
  icon: string; // Icon name
  measures: string[];
}

export interface IncidentResponse {
  detection: string;
  response: string;
  communication: string;
  recovery: string;
  learning: string;
}

export interface SecurityFAQ {
  id: string;
  question: string;
  answer: string;
}

export interface SecurityResource {
  title: string;
  description: string;
  href: string;
}

export interface ContactInfo {
  type: 'security' | 'compliance';
  title: string;
  email: string;
  notes: string;
  icon: string; // Icon name
}

export const securityApi = {
  // Get comprehensive security metrics and status
  getSecurityMetrics: () =>
    apiClient.get<SecurityMetrics>("/v1/metrics/security"),

  // Get service status information
  getServiceStatus: () =>
    apiClient.get<{ services: ServiceStatus[] }>("/v1/metrics/security/services"),

  // Get SSL certificate information
  getSSLCertificates: () =>
    apiClient.get<{ certificates: SSLCertificate[] }>("/v1/metrics/security/certificates"),

  // Get recent security incidents
  getRecentIncidents: (limit: number = 10) =>
    apiClient.get<{ incidents: SecurityIncident[] }>(`/v1/metrics/security/incidents?limit=${limit}`),

  // Get compliance frameworks
  getComplianceFrameworks: () =>
    apiClient.get<{ frameworks: ComplianceFramework[] }>("/v1/metrics/security/compliance"),

  // Get security measures (might be static)
  getSecurityMeasures: () =>
    apiClient.get<{ measures: SecurityMeasure[] }>("/v1/metrics/security/measures"),

  // Get incident response procedures
  getIncidentResponse: () =>
    apiClient.get<IncidentResponse>("/v1/metrics/security/incident-response"),

  // Get security FAQ
  getSecurityFAQ: () =>
    apiClient.get<{ faqs: SecurityFAQ[] }>("/v1/metrics/security/faq"),

  // Get security resources
  getSecurityResources: () =>
    apiClient.get<{ resources: SecurityResource[] }>("/v1/metrics/security/resources"),

  // Get contact information
  getContactInfo: () =>
    apiClient.get<{ contacts: ContactInfo[] }>("/v1/metrics/security/contacts"),
};