/**
 * Security Store
 * Global state management for Security components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// Types
interface Permission { id: string; name: string; resource: string; action: string }
interface Role { id: string; name: string; permissions: string[]; userCount: number }
interface Threat { id: string; type: string; severity: string; source: string; target: string; timestamp: number; status: string }
interface ThreatSector { id: string; name: string; angle: number; radius: number; severity: string; threats: Threat[] }
interface SecurityEvent { id: string; type: string; severity: string; message: string; timestamp: number; userName?: string }
interface SandboxBoundary { id: string; name: string; type: string; boundaries: Array<{ id: string; type: string; status: string }>; status: string }
interface APIEndpoint { id: string; path: string; method: string; exposure: string; risk: string }
interface CredentialAccess { id: string; credentialName: string; accessedBy: string; accessedAt: number; accessType: string; success: boolean }
interface RuntimeIsolate { id: string; name: string; type: string; status: string; memoryUsage?: number; networkAllowed: boolean }
interface ComplianceFramework { id: string; name: string; version: string; status: string; score: number; requirements: Array<{ name: string; status: string }> }
interface ZeroTrustFlow { id: string; source: { type: string; identity: string; trustLevel: number }; destination: { type: string; identity: string }; status: string; timestamp: number }
interface MaliciousExecution { id: string; processName: string; detectedAt: number; threatType: string; confidence: number; severity: string; status: string }
interface AuditEntry { id: string; action: string; actor: { type: string; name: string }; timestamp: number }
interface EncryptionStatus { id: string; name: string; type: string; algorithm?: string; status: string }
interface SuspiciousBehavior { id: string; type: string; description: string; severity: string; detectedAt: number; score: number }
interface Vulnerability { id: string; cveId?: string; name: string; severity: string; status: string; cvssScore?: number }
interface SecurityPolicy { id: string; name: string; description: string; type: string; status: string; rules: Array<{ id: string; condition: string; action: string }>; hitCount?: number }

interface SecurityState {
  // Permission Matrix
  roles: Role[];
  permissions: Permission[];
  selectedRoleId: string | null;

  // Threat Detection
  threatSectors: ThreatSector[];
  selectedThreatId: string | null;

  // Security Timeline
  securityEvents: SecurityEvent[];
  selectedEventId: string | null;

  // Sandbox Boundaries
  boundaries: SandboxBoundary[];
  selectedBoundaryId: string | null;

  // API Exposure
  apiEndpoints: APIEndpoint[];
  selectedEndpointId: string | null;

  // Credential Access
  credentialAccesses: CredentialAccess[];
  selectedAccessId: string | null;

  // Runtime Isolation
  isolates: RuntimeIsolate[];
  selectedIsolateId: string | null;

  // Compliance
  frameworks: ComplianceFramework[];
  selectedFrameworkId: string | null;

  // Zero Trust
  zeroTrustFlows: ZeroTrustFlow[];
  selectedFlowId: string | null;

  // Malicious Execution
  maliciousExecutions: MaliciousExecution[];
  selectedExecutionId: string | null;

  // Audit Trail
  auditEntries: AuditEntry[];
  selectedEntryId: string | null;

  // Encryption Status
  encryptionStatuses: EncryptionStatus[];
  selectedEncryptionStatusId: string | null;

  // Suspicious Behavior
  suspiciousBehaviors: SuspiciousBehavior[];
  selectedBehaviorId: string | null;

  // Vulnerabilities
  vulnerabilities: Vulnerability[];
  selectedVulnerabilityId: string | null;

  // Policies
  policies: SecurityPolicy[];
  selectedPolicyId: string | null;

  // UI State
  activePanel: string;
  sidebarCollapsed: boolean;
}

export const useSecurityStore = create<SecurityState>()(
  immer((set) => ({
    // Permission Matrix
    roles: [],
    permissions: [],
    selectedRoleId: null,
    setRoles: (roles) => set((state) => { state.roles = roles }),
    setPermissions: (permissions) => set((state) => { state.permissions = permissions }),
    selectRole: (roleId) => set((state) => { state.selectedRoleId = roleId }),
    togglePermission: (roleId, permissionId) => set((state) => {
      const role = state.roles.find(r => r.id === roleId)
      if (role) {
        const idx = role.permissions.indexOf(permissionId)
        if (idx === -1) role.permissions.push(permissionId)
        else role.permissions.splice(idx, 1)
      }
    }),

    // Threat Detection
    threatSectors: [],
    selectedThreatId: null,
    setThreatSectors: (sectors) => set((state) => { state.threatSectors = sectors }),
    selectThreat: (threatId) => set((state) => { state.selectedThreatId = threatId }),

    // Security Timeline
    securityEvents: [],
    selectedEventId: null,
    setSecurityEvents: (events) => set((state) => { state.securityEvents = events }),
    selectEvent: (eventId) => set((state) => { state.selectedEventId = eventId }),

    // Sandbox Boundaries
    boundaries: [],
    selectedBoundaryId: null,
    setBoundaries: (boundaries) => set((state) => { state.boundaries = boundaries }),
    selectBoundary: (boundaryId) => set((state) => { state.selectedBoundaryId = boundaryId }),

    // API Exposure
    apiEndpoints: [],
    selectedEndpointId: null,
    setApiEndpoints: (endpoints) => set((state) => { state.apiEndpoints = endpoints }),
    selectEndpoint: (endpointId) => set((state) => { state.selectedEndpointId = endpointId }),
    updateEndpointExposure: (endpointId, exposure) => set((state) => {
      const endpoint = state.apiEndpoints.find(e => e.id === endpointId)
      if (endpoint) endpoint.exposure = exposure
    }),

    // Credential Access
    credentialAccesses: [],
    selectedAccessId: null,
    setCredentialAccesses: (accesses) => set((state) => { state.credentialAccesses = accesses }),
    selectAccess: (accessId) => set((state) => { state.selectedAccessId = accessId }),

    // Runtime Isolation
    isolates: [],
    selectedIsolateId: null,
    setIsolates: (isolates) => set((state) => { state.isolates = isolates }),
    selectIsolate: (isolateId) => set((state) => { state.selectedIsolateId = isolateId }),
    isolateAction: (isolateId, action) => set((state) => {
      const isolate = state.isolates.find(i => i.id === isolateId)
      if (isolate) {
        if (action === 'pause') isolate.status = 'paused'
        else if (action === 'resume') isolate.status = 'running'
        else if (action === 'terminate') isolate.status = 'terminated'
      }
    }),

    // Compliance
    frameworks: [],
    selectedFrameworkId: null,
    setFrameworks: (frameworks) => set((state) => { state.frameworks = frameworks }),
    selectFramework: (frameworkId) => set((state) => { state.selectedFrameworkId = frameworkId }),

    // Zero Trust
    zeroTrustFlows: [],
    selectedFlowId: null,
    setZeroTrustFlows: (flows) => set((state) => { state.zeroTrustFlows = flows }),
    selectFlow: (flowId) => set((state) => { state.selectedFlowId = flowId }),

    // Malicious Execution
    maliciousExecutions: [],
    selectedExecutionId: null,
    setMaliciousExecutions: (executions) => set((state) => { state.maliciousExecutions = executions }),
    selectExecution: (executionId) => set((state) => { state.selectedExecutionId = executionId }),
    blockExecution: (executionId) => set((state) => {
      const exec = state.maliciousExecutions.find(e => e.id === executionId)
      if (exec) exec.status = 'blocked'
    }),

    // Audit Trail
    auditEntries: [],
    selectedEntryId: null,
    setAuditEntries: (entries) => set((state) => { state.auditEntries = entries }),
    selectEntry: (entryId) => set((state) => { state.selectedEntryId = entryId }),

    // Encryption Status
    encryptionStatuses: [],
    selectedEncryptionStatusId: null,
    setEncryptionStatuses: (statuses) => set((state) => { state.encryptionStatuses = statuses }),
    selectEncryptionStatus: (statusId) => set((state) => { state.selectedEncryptionStatusId = statusId }),

    // Suspicious Behavior
    suspiciousBehaviors: [],
    selectedBehaviorId: null,
    setSuspiciousBehaviors: (behaviors) => set((state) => { state.suspiciousBehaviors = behaviors }),
    selectBehavior: (behaviorId) => set((state) => { state.selectedBehaviorId = behaviorId }),

    // Vulnerabilities
    vulnerabilities: [],
    selectedVulnerabilityId: null,
    setVulnerabilities: (vulnerabilities) => set((state) => { state.vulnerabilities = vulnerabilities }),
    selectVulnerability: (vulnerabilityId) => set((state) => { state.selectedVulnerabilityId = vulnerabilityId }),

    // Policies
    policies: [],
    selectedPolicyId: null,
    setPolicies: (policies) => set((state) => { state.policies = policies }),
    selectPolicy: (policyId) => set((state) => { state.selectedPolicyId = policyId }),
    togglePolicy: (policyId) => set((state) => {
      const policy = state.policies.find(p => p.id === policyId)
      if (policy) policy.status = policy.status === 'active' ? 'disabled' : 'active'
    }),

    // UI State
    activePanel: 'threats',
    sidebarCollapsed: false,
    setActivePanel: (panel) => set((state) => { state.activePanel = panel }),
    toggleSidebar: () => set((state) => { state.sidebarCollapsed = !state.sidebarCollapsed }),
  }))
)

export const usePermissionMatrix = () => useSecurityStore((s) => ({ roles: s.roles, permissions: s.permissions, selectedRoleId: s.selectedRoleId }))
export const useThreats = () => useSecurityStore((s) => ({ sectors: s.threatSectors, selectedId: s.selectedThreatId }))
export const useSecurityTimeline = () => useSecurityStore((s) => ({ events: s.securityEvents, selectedId: s.selectedEventId }))
export const useSandboxBoundaries = () => useSecurityStore((s) => ({ boundaries: s.boundaries, selectedId: s.selectedBoundaryId }))
export const useAPIExposure = () => useSecurityStore((s) => ({ endpoints: s.apiEndpoints, selectedId: s.selectedEndpointId }))
export const useCredentialAccess = () => useSecurityStore((s) => ({ accesses: s.credentialAccesses, selectedId: s.selectedAccessId }))
export const useRuntimeIsolation = () => useSecurityStore((s) => ({ isolates: s.isolates, selectedId: s.selectedIsolateId }))
export const useCompliance = () => useSecurityStore((s) => ({ frameworks: s.frameworks, selectedId: s.selectedFrameworkId }))
export const useZeroTrust = () => useSecurityStore((s) => ({ flows: s.zeroTrustFlows, selectedId: s.selectedFlowId }))
export const useMaliciousExecutions = () => useSecurityStore((s) => ({ executions: s.maliciousExecutions, selectedId: s.selectedExecutionId }))
export const useAuditTrail = () => useSecurityStore((s) => ({ entries: s.auditEntries, selectedId: s.selectedEntryId }))
export const useEncryptionStatus = () => useSecurityStore((s) => ({ statuses: s.encryptionStatuses, selectedId: s.selectedEncryptionStatusId }))
export const useSuspiciousBehavior = () => useSecurityStore((s) => ({ behaviors: s.suspiciousBehaviors, selectedId: s.selectedBehaviorId }))
export const useVulnerabilities = () => useSecurityStore((s) => ({ vulnerabilities: s.vulnerabilities, selectedId: s.selectedVulnerabilityId }))
export const useSecurityPolicies = () => useSecurityStore((s) => ({ policies: s.policies, selectedId: s.selectedPolicyId }))
export const useSecurityUI = () => useSecurityStore((s) => ({ activePanel: s.activePanel, collapsed: s.sidebarCollapsed }))
