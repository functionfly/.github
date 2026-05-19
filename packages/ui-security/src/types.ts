/**
 * @functionfly/ui-security
 * Security Components - Types and Interfaces
 */

// ============================================================================
// Permission Matrix
// ============================================================================

export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: 'read' | 'write' | 'delete' | 'admin';
  description?: string;
}

export interface Role {
  id: string;
  name: string;
  permissions: string[];
  userCount: number;
  createdAt: number;
  isSystem?: boolean;
}

export interface PermissionMatrixProps {
  roles: Role[];
  permissions: Permission[];
  selectedRoleId?: string | null;
  onRoleSelect?: (role: Role) => void;
  onPermissionToggle?: (roleId: string, permissionId: string) => void;
  showUserCount?: boolean;
  className?: string;
}

// ============================================================================
// Threat Detection Radar
// ============================================================================

export interface ThreatSector {
  id: string;
  name: string;
  angle: number;
  radius: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  threats: Threat[];
  color?: string;
}

export interface Threat {
  id: string;
  type: 'malware' | 'intrusion' | 'dos' | 'data-breach' | 'insider' | 'phishing';
  severity: 'low' | 'medium' | 'high' | 'critical';
  source: string;
  target: string;
  timestamp: number;
  status: 'detected' | 'investigating' | 'mitigated' | 'false-positive';
  description?: string;
}

export interface ThreatDetectionRadarProps {
  sectors: ThreatSector[];
  selectedThreatId?: string | null;
  onThreatSelect?: (threat: Threat) => void;
  showLegend?: boolean;
  className?: string;
}

// ============================================================================
// Security Timeline
// ============================================================================

export interface SecurityEvent {
  id: string;
  type: 'login' | 'logout' | 'permission-change' | 'resource-access' | 'config-change' | 'alert';
  severity: 'info' | 'warning' | 'error' | 'critical';
  message: string;
  timestamp: number;
  userId?: string;
  userName?: string;
  ipAddress?: string;
  metadata?: Record<string, unknown>;
  source?: string;
}

export interface SecurityTimelineProps {
  events: SecurityEvent[];
  selectedEventId?: string | null;
  onEventSelect?: (event: SecurityEvent) => void;
  filters?: {
    severity?: SecurityEvent['severity'][];
    type?: SecurityEvent['type'][];
    dateRange?: { start: number; end: number };
  };
  className?: string;
}

// ============================================================================
// Sandbox Boundary Viewer
// ============================================================================

export interface SandboxBoundary {
  id: string;
  name: string;
  type: 'function' | 'container' | 'vm' | 'process';
  boundaries: Array<{
    id: string;
    type: 'network' | 'filesystem' | 'memory' | 'cpu';
    status: 'enforced' | 'relaxed' | 'bypassed';
    rules?: string[];
  }>;
  status: 'active' | 'inactive' | 'breached';
  breachedAt?: number;
}

export interface SandboxBoundaryViewerProps {
  boundaries: SandboxBoundary[];
  selectedBoundaryId?: string | null;
  onBoundarySelect?: (boundary: SandboxBoundary) => void;
  className?: string;
}

// ============================================================================
// API Exposure Analyzer
// ============================================================================

export interface APIEndpoint {
  id: string;
  path: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  exposure: 'public' | 'authenticated' | 'internal' | 'private';
  risk: 'low' | 'medium' | 'high' | 'critical';
  callsLast7Days?: number;
  authRequired: boolean;
  rateLimited: boolean;
  lastAccessed?: number;
}

export interface APIExposureAnalyzerProps {
  endpoints: APIEndpoint[];
  selectedEndpointId?: string | null;
  onEndpointSelect?: (endpoint: APIEndpoint) => void;
  onExposureChange?: (endpointId: string, exposure: APIEndpoint['exposure']) => void;
  className?: string;
}

// ============================================================================
// Credential Access Monitor
// ============================================================================

export interface CredentialAccess {
  id: string;
  credentialId: string;
  credentialName: string;
  accessedBy: string;
  accessedByUserId?: string;
  accessedAt: number;
  accessType: 'read' | 'write' | 'delete' | 'use';
  success: boolean;
  ipAddress?: string;
  userAgent?: string;
  location?: string;
}

export interface CredentialAccessMonitorProps {
  accesses: CredentialAccess[];
  selectedAccessId?: string | null;
  onAccessSelect?: (access: CredentialAccess) => void;
  onExport?: () => void;
  className?: string;
}

// ============================================================================
// Runtime Isolation Map
// ============================================================================

export interface RuntimeIsolate {
  id: string;
  name: string;
  type: 'v8-isolate' | 'wasm-instance' | 'container' | 'process';
  status: 'running' | 'paused' | 'terminated' | 'crashed';
  memoryUsage?: number;
  cpuUsage?: number;
  networkAllowed: boolean;
  filesystemAllowed: boolean;
  environment: Record<string, string>;
  parent?: string;
}

export interface RuntimeIsolationMapProps {
  isolates: RuntimeIsolate[];
  selectedIsolateId?: string | null;
  onIsolateSelect?: (isolate: RuntimeIsolate) => void;
  onIsolateAction?: (isolateId: string, action: 'pause' | 'resume' | 'terminate') => void;
  className?: string;
}

// ============================================================================
// Compliance Dashboard
// ============================================================================

export interface ComplianceFramework {
  id: string;
  name: string;
  version: string;
  status: 'compliant' | 'partial' | 'non-compliant' | 'audit-required';
  score: number;
  requirements: Array<{
    id: string;
    name: string;
    status: 'met' | 'partial' | ' unmet' | 'not-applicable';
    evidence?: string[];
    lastChecked?: number;
  }>;
  lastAuditDate?: number;
  nextAuditDate?: number;
}

export interface ComplianceDashboardProps {
  frameworks: ComplianceFramework[];
  selectedFrameworkId?: string | null;
  onFrameworkSelect?: (framework: ComplianceFramework) => void;
  className?: string;
}

// ============================================================================
// Zero Trust Flow Viewer
// ============================================================================

export interface ZeroTrustFlow {
  id: string;
  source: {
    type: 'user' | 'service' | 'device' | 'external';
    identity: string;
    trustLevel: number;
  };
  destination: {
    type: 'resource' | 'service' | 'data';
    identity: string;
    sensitivity: 'low' | 'medium' | 'high' | 'critical';
  };
  path: string[];
  verificationSteps: Array<{
    type: 'identity' | 'device' | 'context' | 'policy';
    status: 'passed' | 'failed' | 'pending';
    timestamp?: number;
  }>;
  status: 'allowed' | 'denied' | 'pending' | 'timeout';
  timestamp: number;
}

export interface ZeroTrustFlowViewerProps {
  flows: ZeroTrustFlow[];
  selectedFlowId?: string | null;
  onFlowSelect?: (flow: ZeroTrustFlow) => void;
  onFlowRefresh?: () => void;
  className?: string;
}

// ============================================================================
// Malicious Execution Detector
// ============================================================================

export interface MaliciousExecution {
  id: string;
  processName: string;
  pid?: number;
  parentProcess?: string;
  detectedAt: number;
  threatType: string;
  confidence: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  status: 'detected' | 'blocked' | 'terminated' | 'escaped';
  command?: string;
  userId?: string;
  machineName?: string;
}

export interface MaliciousExecutionDetectorProps {
  executions: MaliciousExecution[];
  selectedExecutionId?: string | null;
  onExecutionSelect?: (execution: MaliciousExecution) => void;
  onBlock?: (executionId: string) => void;
  onTerminate?: (executionId: string) => void;
  className?: string;
}

// ============================================================================
// Audit Trail Explorer
// ============================================================================

export interface AuditEntry {
  id: string;
  action: string;
  actor: {
    type: 'user' | 'system' | 'service';
    id: string;
    name: string;
  };
  target?: {
    type: string;
    id: string;
    name: string;
  };
  timestamp: number;
  ipAddress?: string;
  userAgent?: string;
  changes?: Array<{
    field: string;
    oldValue: unknown;
    newValue: unknown;
  }>;
  metadata?: Record<string, unknown>;
}

export interface AuditTrailExplorerProps {
  entries: AuditEntry[];
  selectedEntryId?: string | null;
  onEntrySelect?: (entry: AuditEntry) => void;
  onExport?: (format: 'csv' | 'json') => void;
  className?: string;
}

// ============================================================================
// Encryption Status Panel
// ============================================================================

export interface EncryptionStatus {
  id: string;
  name: string;
  type: 'at-rest' | 'in-transit' | 'in-use';
  algorithm?: string;
  keySize?: number;
  status: 'encrypted' | 'decrypted' | 'key-missing' | 'error';
  lastVerified?: number;
  nextRotation?: number;
  keyId?: string;
}

export interface EncryptionStatusPanelProps {
  statuses: EncryptionStatus[];
  selectedStatusId?: string | null;
  onStatusSelect?: (status: EncryptionStatus) => void;
  className?: string;
}

// ============================================================================
// Suspicious Behavior Timeline
// ============================================================================

export interface SuspiciousBehavior {
  id: string;
  type: 'anomaly' | 'signature-match' | 'heuristic' | 'behavioral';
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  detectedAt: number;
  score: number;
  indicators: string[];
  affectedEntities?: string[];
  status: 'new' | 'investigating' | 'resolved' | 'dismissed';
}

export interface SuspiciousBehaviorTimelineProps {
  behaviors: SuspiciousBehavior[];
  selectedBehaviorId?: string | null;
  onBehaviorSelect?: (behavior: SuspiciousBehavior) => void;
  className?: string;
}

// ============================================================================
// Vulnerability Scanner
// ============================================================================

export interface Vulnerability {
  id: string;
  cveId?: string;
  name: string;
  description: string;
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  status: 'open' | 'in-remediation' | 'resolved' | 'false-positive' | 'accepted';
  cvssScore?: number;
  affectedComponents: string[];
  discoveredAt: number;
  discoveredBy?: string;
  remediation?: string;
}

export interface VulnerabilityScannerProps {
  vulnerabilities: Vulnerability[];
  selectedVulnerabilityId?: string | null;
  onVulnerabilitySelect?: (vulnerability: Vulnerability) => void;
  onRemediate?: (vulnerabilityId: string) => void;
  className?: string;
}

// ============================================================================
// Policy Engine Viewer
// ============================================================================

export interface SecurityPolicy {
  id: string;
  name: string;
  description: string;
  type: 'access' | 'network' | 'data' | 'audit';
  status: 'active' | 'draft' | 'disabled';
  rules: Array<{
    id: string;
    condition: string;
    action: 'allow' | 'deny' | 'alert' | 'log';
    priority: number;
  }>;
  lastEvaluated?: number;
  hitCount?: number;
}

export interface PolicyEngineViewerProps {
  policies: SecurityPolicy[];
  selectedPolicyId?: string | null;
  onPolicySelect?: (policy: SecurityPolicy) => void;
  onPolicyToggle?: (policyId: string) => void;
  className?: string;
}
