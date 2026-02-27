export const RISK_LEVELS = {
  excellent: {
    color: '#10b981', // green-500
    bgColor: '#dcfce7', // green-100
    textColor: '#166534', // green-800
    label: 'Excellent'
  },
  good: {
    color: '#f59e0b', // yellow-500
    bgColor: '#fef3c7', // yellow-100
    textColor: '#92400e', // yellow-800
    label: 'Good'
  },
  warning: {
    color: '#f97316', // orange-500
    bgColor: '#fed7aa', // orange-100
    textColor: '#9a3412', // orange-800
    label: 'Warning'
  },
  critical: {
    color: '#ef4444', // red-500
    bgColor: '#fecaca', // red-100
    textColor: '#991b1b', // red-800
    label: 'Critical'
  },
  info: {
    color: '#3b82f6', // blue-500
    bgColor: '#dbeafe', // blue-100
    textColor: '#1e40af', // blue-800
    label: 'Info'
  }
} as const;

export type RiskLevel = keyof typeof RISK_LEVELS;

export function getRiskLevel(score: number): RiskLevel {
  if (score >= 98) return 'excellent';
  if (score >= 95) return 'good';
  if (score >= 85) return 'warning';
  return 'critical';
}

export function getStatusRiskLevel(status: string): RiskLevel {
  switch (status.toLowerCase()) {
    case 'operational':
    case 'valid':
    case 'certified':
    case 'resolved':
      return 'excellent';
    case 'degraded':
    case 'expiring':
    case 'compliant':
    case 'investigating':
      return 'warning';
    case 'outage':
    case 'expired':
    case 'not applicable':
    case 'open':
      return 'critical';
    default:
      return 'info';
  }
}

export function getSeverityRiskLevel(severity: string): RiskLevel {
  switch (severity.toLowerCase()) {
    case 'info':
      return 'info';
    case 'warning':
      return 'warning';
    case 'critical':
      return 'critical';
    default:
      return 'info';
  }
}