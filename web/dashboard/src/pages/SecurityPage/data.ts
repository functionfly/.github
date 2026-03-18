import type {
  ComplianceFramework,
  ContactInfo,
  IncidentResponse,
  SecurityFAQ,
  SecurityIncident,
  SecurityMeasure,
  SecurityResource,
  ServiceStatus,
  SSLCertificate,
} from './types';

export const initialServiceStatus: ServiceStatus[] = [
  { name: 'API Gateway', status: 'operational', uptime: '99.98%', responseTime: '45ms' },
  { name: 'Database', status: 'operational', uptime: '99.99%', responseTime: '12ms' },
  { name: 'CDN', status: 'operational', uptime: '99.95%', responseTime: '28ms' },
  { name: 'Authentication', status: 'operational', uptime: '99.97%', responseTime: '67ms' },
  { name: 'Deployment Engine', status: 'operational', uptime: '99.92%', responseTime: '89ms' },
  { name: 'Monitoring', status: 'operational', uptime: '100.00%', responseTime: '23ms' },
];

export const initialRecentIncidents: SecurityIncident[] = [
  {
    id: 'INC-2025-001',
    title: 'Routine Security Patch Deployment',
    severity: 'info',
    status: 'resolved',
    timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
    description:
      'Applied latest security patches to infrastructure components. No service disruption.',
    impact: 'None',
    duration: '15 minutes',
  },
  {
    id: 'INC-2025-002',
    title: 'DDoS Mitigation Activated',
    severity: 'warning',
    status: 'resolved',
    timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // 1 day ago
    description: 'Automated DDoS protection system activated and mitigated a volumetric attack.',
    impact: 'Minimal latency increase (<5%)',
    duration: '8 minutes',
  },
];

export const initialSSLCertificates: SSLCertificate[] = [
  {
    domain: 'api.functionfly.com',
    issuer: "Let's Encrypt",
    expiryDate: new Date(Date.now() + 45 * 24 * 60 * 60 * 1000).toISOString(), // 45 days from now
    status: 'valid',
    autoRenewal: true,
  },
  {
    domain: 'app.functionfly.com',
    issuer: "Let's Encrypt",
    expiryDate: new Date(Date.now() + 52 * 24 * 60 * 60 * 1000).toISOString(), // 52 days from now
    status: 'valid',
    autoRenewal: true,
  },
  {
    domain: 'functionfly.com',
    issuer: 'DigiCert',
    expiryDate: new Date(Date.now() + 180 * 24 * 60 * 60 * 1000).toISOString(), // 180 days from now
    status: 'valid',
    autoRenewal: true,
  },
];

export const complianceFrameworks: ComplianceFramework[] = [
  {
    name: 'SOC 2 Type II',
    status: 'Certified',
    description: 'Security, Availability, and Confidentiality controls',
    auditor: 'Independent third-party audit',
    lastAudit: 'December 2025',
    nextAudit: 'December 2026',
  },
  {
    name: 'ISO 27001',
    status: 'Certified',
    description: 'Information Security Management Systems',
    auditor: 'ISO-accredited auditor',
    lastAudit: 'October 2025',
    nextAudit: 'October 2026',
  },
  {
    name: 'GDPR',
    status: 'Compliant',
    description: 'General Data Protection Regulation',
    auditor: 'Internal compliance team',
    lastAudit: 'Ongoing',
    nextAudit: 'Ongoing',
  },
  {
    name: 'CCPA',
    status: 'Compliant',
    description: 'California Consumer Privacy Act',
    auditor: 'Internal compliance team',
    lastAudit: 'Ongoing',
    nextAudit: 'Ongoing',
  },
];

export const securityMeasures: SecurityMeasure[] = [
  {
    category: 'Infrastructure Security',
    icon: 'Server',
    measures: [
      'Multi-cloud deployment with automatic failover',
      'End-to-end encryption (AES-256)',
      'Automated security patching and updates',
      'DDoS protection with global CDN',
      'Zero-trust network architecture',
      'Container security scanning',
    ],
  },
  {
    category: 'Application Security',
    icon: 'Code',
    measures: [
      'OWASP Top 10 compliance',
      'Automated vulnerability scanning',
      'Secure coding practices and reviews',
      'Runtime Application Self-Protection (RASP)',
      'API rate limiting and throttling',
      'Input validation and sanitization',
    ],
  },
  {
    category: 'Data Protection',
    icon: 'Database',
    measures: [
      'Data encryption at rest and in transit',
      'Database access controls and auditing',
      'Regular security assessments',
      'Backup encryption and integrity checks',
      'Data classification and handling procedures',
      'Secure deletion protocols',
    ],
  },
  {
    category: 'Access Control',
    icon: 'Key',
    measures: [
      'Multi-factor authentication (MFA)',
      'Role-based access control (RBAC)',
      'Single sign-on (SSO) integration',
      'Session management and timeout',
      'Audit logging for all access events',
      'Least privilege principle enforcement',
    ],
  },
];

export const incidentResponse: IncidentResponse = {
  detection: '24/7 automated monitoring and alerting',
  response: '< 15 minutes average response time',
  communication: 'Transparent incident communication',
  recovery: 'Automated failover and disaster recovery',
  learning: 'Post-incident analysis and improvement',
};

export const securityFAQs: SecurityFAQ[] = [
  {
    id: 'encryption',
    question: 'How is data encrypted?',
    answer:
      'All data is encrypted at rest using AES-256 encryption and in transit using TLS 1.3. Database connections, API communications, and file storage all use end-to-end encryption with perfect forward secrecy.',
  },
  {
    id: 'penetration-testing',
    question: 'Do you conduct penetration testing?',
    answer:
      'Yes, we conduct quarterly penetration testing by certified security researchers, annual red team exercises, and continuous automated security scanning. All findings are remediated within SLA timelines.',
  },
  {
    id: 'data-residency',
    question: 'Where is data stored?',
    answer:
      'Data can be stored in multiple regions (US East, US West, EU Central) based on your compliance requirements. Cross-region replication ensures high availability while maintaining data sovereignty.',
  },
  {
    id: 'third-party-risk',
    question: 'How do you manage third-party risks?',
    answer:
      'All third-party vendors undergo security assessments, contract reviews, and continuous monitoring. We maintain a vendor risk register and conduct annual reassessments of critical suppliers.',
  },
  {
    id: 'zero-trust',
    question: 'Do you use zero-trust architecture?',
    answer:
      'Yes, our platform implements zero-trust principles: every request is authenticated and authorized, network segmentation prevents lateral movement, and continuous monitoring detects anomalous behavior.',
  },
];

export const securityResources: SecurityResource[] = [
  {
    title: 'Security Overview',
    description: 'Technical security documentation',
    href: '#',
  },
  {
    title: 'Trust Center',
    description: 'Compliance certificates and audits',
    href: '#',
  },
  {
    title: 'Security Best Practices',
    description: 'Guidelines for secure deployments',
    href: '#',
  },
  {
    title: 'Contact Security Team',
    description: 'Report security vulnerabilities',
    href: '#',
  },
];

export const contactInfo: ContactInfo[] = [
  {
    type: 'security',
    title: 'Security Issues',
    email: 'security@functionfly.com',
    notes: 'PGP key available',
    icon: 'AlertTriangle',
  },
  {
    type: 'compliance',
    title: 'Compliance Questions',
    email: 'compliance@functionfly.com',
    notes: 'Business hours response',
    icon: 'Shield',
  },
];
