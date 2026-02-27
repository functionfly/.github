export interface ServiceStatus {
  name: string;
  status: string; // "operational", "degraded", "outage"
  uptime: string;
  responseTime: string;
}

export interface SSLCertificate {
  domain: string;
  issuer: string;
  expiryDate: string; // ISO date string
  status: string; // "valid", "expiring", "expired"
  autoRenewal: boolean;
}

export interface SecurityIncident {
  id: string;
  title: string;
  severity: string; // "info", "warning", "critical"
  status: string; // "open", "resolved", "investigating"
  timestamp: string; // ISO date string
  description: string;
  impact: string;
  duration: string;
}

export interface ComplianceFramework {
  name: string;
  status: string; // "Certified", "Compliant", "In Progress", "Not Applicable"
  description: string;
  auditor: string;
  lastAudit: string;
  nextAudit: string;
}

export interface SecurityMeasure {
  category: string;
  icon: string; // Icon name as string
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
  type: string; // "security", "compliance"
  title: string;
  email: string;
  notes: string;
  icon: string; // Icon name as string
}