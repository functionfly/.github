import { Shield, AlertTriangle, AlertCircle, Info, CheckCircle, XCircle, Clock, FileCode, Container, Server, Lock } from 'lucide-react';

export type SeverityLevel = 'critical' | 'high' | 'medium' | 'low' | 'info';
export type VulnerabilityStatus = 'open' | 'fixed' | 'accepted' | 'false_positive';
export type ScanStatus = 'running' | 'completed' | 'failed';
export type ScanType = 'container' | 'dockerfile' | 'compose' | 'runtime' | 'image';

export interface Vulnerability {
  id: string;
  title: string;
  description: string;
  severity: SeverityLevel;
  cvss_score?: number;
  cve?: string;
  category: string;
  component: string;
  location?: string;
  status: VulnerabilityStatus;
  remediation?: string;
  reference_urls?: string[];
  metadata?: Record<string, unknown>;
  discovered: string;
  updated: string;
}

export interface ScanSummary {
  total_vulnerabilities: number;
  critical_count: number;
  high_count: number;
  medium_count: number;
  low_count: number;
  info_count: number;
  risk_score: number;
  coverage_percentage: number;
  compliance_score?: number;
}

export interface SecurityScan {
  id: string;
  type: string;
  status: ScanStatus;
  target: string;
  started_at: string;
  completed_at?: string;
  duration?: number;
  vulnerabilities: Vulnerability[];
  summary: ScanSummary;
}

export const severityConfig: Record<SeverityLevel, {
  label: string;
  color: string;
  bgColor: string;
  borderColor: string;
  icon: typeof Shield;
  score: number;
}> = {
  critical: {
    label: 'Critical',
    color: 'text-red-500',
    bgColor: 'bg-red-500/10',
    borderColor: 'border-red-500/30',
    icon: XCircle,
    score: 5,
  },
  high: {
    label: 'High',
    color: 'text-orange-500',
    bgColor: 'bg-orange-500/10',
    borderColor: 'border-orange-500/30',
    icon: AlertTriangle,
    score: 4,
  },
  medium: {
    label: 'Medium',
    color: 'text-yellow-500',
    bgColor: 'bg-yellow-500/10',
    borderColor: 'border-yellow-500/30',
    icon: AlertCircle,
    score: 3,
  },
  low: {
    label: 'Low',
    color: 'text-blue-500',
    bgColor: 'bg-blue-500/10',
    borderColor: 'border-blue-500/30',
    icon: Info,
    score: 2,
  },
  info: {
    label: 'Info',
    color: 'text-green-500',
    bgColor: 'bg-green-500/10',
    borderColor: 'border-green-500/30',
    icon: CheckCircle,
    score: 1,
  },
};

export const statusConfig: Record<VulnerabilityStatus, {
  label: string;
  color: string;
  bgColor: string;
}> = {
  open: {
    label: 'Open',
    color: 'text-red-400',
    bgColor: 'bg-red-500/10',
  },
  fixed: {
    label: 'Fixed',
    color: 'text-green-400',
    bgColor: 'bg-green-500/10',
  },
  accepted: {
    label: 'Accepted',
    color: 'text-amber-400',
    bgColor: 'bg-amber-500/10',
  },
  false_positive: {
    label: 'False Positive',
    color: 'text-gray-400',
    bgColor: 'bg-gray-500/10',
  },
};

export const scanTypeConfig: Record<string, { label: string; icon: typeof Container }> = {
  container: { label: 'Container Security', icon: Container },
  dockerfile: { label: 'Dockerfile', icon: FileCode },
  compose: { label: 'Docker Compose', icon: Server },
  runtime: { label: 'Runtime Security', icon: Shield },
  image: { label: 'Image Vulnerabilities', icon: Lock },
};

export const getSeverityFromScore = (score: number): SeverityLevel => {
  if (score >= 9.0) return 'critical';
  if (score >= 7.0) return 'high';
  if (score >= 4.0) return 'medium';
  if (score >= 0.1) return 'low';
  return 'info';
};

export const calculateRiskScore = (summary: ScanSummary): number => {
  const weights = { critical: 10, high: 5, medium: 2, low: 1, info: 0 };
  const maxScore = 100;
  const weightedSum = 
    summary.critical_count * weights.critical +
    summary.high_count * weights.high +
    summary.medium_count * weights.medium +
    summary.low_count * weights.low;
  return Math.min(100, Math.round((weightedSum / maxScore) * 100));
};

export const sortVulnerabilities = (vulns: Vulnerability[]): Vulnerability[] => {
  const severityOrder: Record<SeverityLevel, number> = {
    critical: 0, high: 1, medium: 2, low: 3, info: 4,
  };
  return [...vulns].sort((a, b) => {
    const severityDiff = severityOrder[a.severity] - severityOrder[b.severity];
    if (severityDiff !== 0) return severityDiff;
    return new Date(b.discovered).getTime() - new Date(a.discovered).getTime();
  });
};

export const groupVulnerabilitiesBySeverity = (vulns: Vulnerability[]) => {
  return {
    critical: vulns.filter(v => v.severity === 'critical'),
    high: vulns.filter(v => v.severity === 'high'),
    medium: vulns.filter(v => v.severity === 'medium'),
    low: vulns.filter(v => v.severity === 'low'),
    info: vulns.filter(v => v.severity === 'info'),
  };
};

export const formatDuration = (ms: number): string => {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${Math.round(ms / 1000)}s`;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.round((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
};

export const formatDate = (dateString: string): string => {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(dateString));
};
