/**
 * Security Integration Component
 * Unified panel that wires all Security components together
 */

import React, { useState, useMemo } from 'react'
import { cn } from '@functionfly/ui-core'
import {
  Shield,
  ShieldCheck,
  ShieldAlert,
  AlertTriangle,
  Bug,
  Lock,
  Key,
  Eye,
  Radar,
  Globe,
  Container,
  Scan,
  Scale,
  Clipboard,
  History,
  ChevronRight,
  ChevronDown,
} from 'lucide-react'

import {
  PermissionMatrix,
  ThreatDetectionRadar,
  SecurityTimeline,
  SandboxBoundaryViewer,
  APIExposureAnalyzer,
  CredentialAccessMonitor,
  RuntimeIsolationMap,
  ComplianceDashboard,
  ZeroTrustFlowViewer,
  MaliciousExecutionDetector,
  AuditTrailExplorer,
  EncryptionStatusPanel,
  SuspiciousBehaviorTimeline,
  VulnerabilityScanner,
  PolicyEngineViewer,
} from '@functionfly/ui-security'

const NAV_ITEMS = [
  { id: 'threats', label: 'Threat Radar', icon: Radar },
  { id: 'timeline', label: 'Timeline', icon: History },
  { id: 'permissions', label: 'Permissions', icon: Shield },
  { id: 'sandbox', label: 'Sandbox', icon: Container },
  { id: 'api', label: 'API Exposure', icon: Globe },
  { id: 'credentials', label: 'Credentials', icon: Key },
  { id: 'isolation', label: 'Isolation', icon: Lock },
  { id: 'compliance', label: 'Compliance', icon: ShieldCheck },
  { id: 'zerotrust', label: 'Zero Trust', icon: Shield },
  { id: 'malicious', label: 'Malicious', icon: Bug },
  { id: 'audit', label: 'Audit Trail', icon: Clipboard },
  { id: 'encryption', label: 'Encryption', icon: Lock },
  { id: 'behavior', label: 'Behavior', icon: Eye },
  { id: 'vulns', label: 'Vulnerabilities', icon: Scan },
  { id: 'policies', label: 'Policies', icon: Scale },
] as const

type PanelId = typeof NAV_ITEMS[number]['id']

export const SecurityIntegration: React.FC<{ className?: string }> = ({ className }) => {
  const [activePanel, setActivePanel] = useState<PanelId>('threats')
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  const mockRoles = useMemo(() => [
    { id: 'r1', name: 'Admin', permissions: ['p1', 'p2', 'p3'], userCount: 5, isSystem: true },
    { id: 'r2', name: 'Developer', permissions: ['p1', 'p2'], userCount: 23 },
    { id: 'r3', name: 'Viewer', permissions: ['p1'], userCount: 45 },
  ], [])

  const mockPermissions = useMemo(() => [
    { id: 'p1', name: 'Read', resource: 'functions', action: 'read' as const },
    { id: 'p2', name: 'Write', resource: 'functions', action: 'write' as const },
    { id: 'p3', name: 'Delete', resource: 'functions', action: 'delete' as const },
    { id: 'p4', name: 'Admin', resource: 'users', action: 'admin' as const },
    { id: 'p5', name: 'Read', resource: 'logs', action: 'read' as const },
    { id: 'p6', name: 'Write', resource: 'config', action: 'write' as const },
  ], [])

  const mockThreatSectors = useMemo(() => [
    { id: 's1', name: 'Network', angle: 45, radius: 0.6, severity: 'high' as const, threats: [
      { id: 't1', type: 'intrusion' as const, severity: 'high' as const, source: '185.220.101.x', target: 'api-server', timestamp: Date.now(), status: 'detected' as const },
    ]},
    { id: 's2', name: 'Application', angle: 135, radius: 0.4, severity: 'medium' as const, threats: [] },
    { id: 's3', name: 'Data', angle: 225, radius: 0.8, severity: 'critical' as const, threats: [
      { id: 't2', type: 'data-breach' as const, severity: 'critical' as const, source: 'internal', target: 'user-db', timestamp: Date.now() - 3600000, status: 'investigating' as const },
    ]},
  ], [])

  const mockEvents = useMemo(() => [
    { id: 'e1', type: 'login' as const, severity: 'info' as const, message: 'User sarah@company.com logged in', timestamp: Date.now() - 60000, userName: 'Sarah Chen', ipAddress: '192.168.1.1' },
    { id: 'e2', type: 'permission-change' as const, severity: 'warning' as const, message: 'Admin role granted to mike@company.com', timestamp: Date.now() - 300000, userName: 'Mike Johnson' },
    { id: 'e3', type: 'resource-access' as const, severity: 'info' as const, message: 'API key created for production', timestamp: Date.now() - 600000, userName: 'Sarah Chen' },
  ], [])

  const mockBoundaries = useMemo(() => [
    { id: 'b1', name: 'API Gateway', type: 'container' as const, boundaries: [
      { id: 'bn1', type: 'network' as const, status: 'enforced' as const },
      { id: 'bn2', type: 'filesystem' as const, status: 'enforced' as const },
    ], status: 'active' as const },
    { id: 'b2', name: 'Worker Process', type: 'process' as const, boundaries: [
      { id: 'bn3', type: 'memory' as const, status: 'relaxed' as const },
      { id: 'bn4', type: 'cpu' as const, status: 'enforced' as const },
    ], status: 'active' as const },
  ], [])

  const mockEndpoints = useMemo(() => [
    { id: 'ep1', path: '/api/v1/users', method: 'GET' as const, exposure: 'public' as const, risk: 'high' as const },
    { id: 'ep2', path: '/api/v1/admin', method: 'POST' as const, exposure: 'internal' as const, risk: 'critical' as const },
    { id: 'ep3', path: '/api/v1/functions', method: 'GET' as const, exposure: 'authenticated' as const, risk: 'low' as const },
  ], [])

  const mockAccesses = useMemo(() => [
    { id: 'ca1', credentialName: 'DATABASE_PASSWORD', accessedBy: 'api-server', accessedAt: Date.now() - 120000, accessType: 'use' as const, success: true, ipAddress: '10.0.0.1' },
    { id: 'ca2', credentialName: 'API_SECRET_KEY', accessedBy: 'worker-1', accessedAt: Date.now() - 300000, accessType: 'read' as const, success: true },
  ], [])

  const mockIsolates = useMemo(() => [
    { id: 'i1', name: 'User Function 1', type: 'v8-isolate' as const, status: 'running' as const, memoryUsage: 45, networkAllowed: false },
    { id: 'i2', name: 'Data Processor', type: 'wasm-instance' as const, status: 'running' as const, memoryUsage: 72, networkAllowed: true },
  ], [])

  const mockFrameworks = useMemo(() => [
    { id: 'f1', name: 'SOC 2 Type II', version: '2019', status: 'compliant' as const, score: 96, requirements: [{ name: 'Access Control', status: 'met' }, { name: 'Encryption', status: 'met' }] },
    { id: 'f2', name: 'GDPR', version: '2018', status: 'partial' as const, score: 78, requirements: [{ name: 'Data Protection', status: 'partial' }] },
  ], [])

  const mockFlows = useMemo(() => [
    { id: 'f1', source: { type: 'user' as const, identity: 'sarah@company.com', trustLevel: 0.95 }, destination: { type: 'resource' as const, identity: '/api/users', sensitivity: 'high' as const }, status: 'allowed' as const, timestamp: Date.now() },
    { id: 'f2', source: { type: 'service' as const, identity: 'worker-1', trustLevel: 0.85 }, destination: { type: 'data' as const, identity: 'user-db', sensitivity: 'critical' as const }, status: 'pending' as const, timestamp: Date.now() },
  ], [])

  const mockMalicious = useMemo(() => [
    { id: 'm1', processName: 'suspicious_download.sh', detectedAt: Date.now() - 1800000, threatType: 'Trojan', confidence: 0.89, severity: 'high' as const, status: 'blocked' as const },
  ], [])

  const mockAudit = useMemo(() => [
    { id: 'a1', action: 'User permissions updated', actor: { type: 'user' as const, name: 'Admin' }, timestamp: Date.now() - 300000 },
    { id: 'a2', action: 'API key created', actor: { type: 'user' as const, name: 'Sarah Chen' }, timestamp: Date.now() - 600000 },
  ], [])

  const mockEncryption = useMemo(() => [
    { id: 'e1', name: 'User Data at Rest', type: 'at-rest' as const, algorithm: 'AES-256-GCM', status: 'encrypted' as const },
    { id: 'e2', name: 'API Traffic', type: 'in-transit' as const, algorithm: 'TLS 1.3', status: 'encrypted' as const },
  ], [])

  const mockBehavior = useMemo(() => [
    { id: 'b1', type: 'anomaly' as const, description: 'Unusual login time detected for user sarah@company.com', severity: 'medium' as const, detectedAt: Date.now() - 3600000, score: 72 },
  ], [])

  const mockVulns = useMemo(() => [
    { id: 'v1', cveId: 'CVE-2024-1234', name: 'Critical API Vulnerability', severity: 'critical' as const, status: 'open' as const, cvssScore: 9.8 },
    { id: 'v2', name: 'Information Disclosure', severity: 'medium' as const, status: 'in-remediation' as const, cvssScore: 5.3 },
  ], [])

  const mockPolicies = useMemo(() => [
    { id: 'p1', name: 'Deny External API Access', description: 'Block all external traffic to internal APIs', type: 'access' as const, status: 'active' as const, rules: [{ id: 'r1', condition: 'source.ip not in internal', action: 'deny' }], hitCount: 1250 },
    { id: 'p2', name: 'Require MFA for Admins', description: 'All admin users must use multi-factor authentication', type: 'access' as const, status: 'active' as const, rules: [], hitCount: 8920 },
  ], [])

  return (
    <div className={cn('flex h-full bg-aviation-bg-panel rounded-lg border border-aviation-border-panel overflow-hidden', className)}>
      <div className={cn('flex flex-col border-r border-aviation-border-panel transition-all duration-300', sidebarCollapsed ? 'w-12' : 'w-56')}>
        <div className="flex items-center justify-end px-2 py-2 border-b border-aviation-border-panel">
          <button onClick={() => setSidebarCollapsed(!sidebarCollapsed)} className="p-1.5 hover:bg-aviation-bg-instrument rounded">
            {sidebarCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
        </div>
        <nav className="flex-1 overflow-auto py-2">
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon
            const isActive = activePanel === item.id
            return (
              <button
                key={item.id}
                onClick={() => setActivePanel(item.id)}
                className={cn(
                  'flex items-center gap-3 w-full px-3 py-2 text-left transition-colors',
                  isActive ? 'bg-aviation-cyan/20 text-aviation-cyan border-l-2 border-aviation-cyan' : 'text-aviation-text-muted hover:text-aviation-text-primary hover:bg-aviation-bg-secondary',
                  sidebarCollapsed && 'justify-center px-0'
                )}
                title={sidebarCollapsed ? item.label : undefined}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                {!sidebarCollapsed && <span className="text-sm truncate">{item.label}</span>}
              </button>
            )
          })}
        </nav>
        {!sidebarCollapsed && (
          <div className="px-3 py-2 border-t border-aviation-border-panel">
            <div className="flex items-center gap-2 text-xs text-aviation-text-muted">
              <div className="w-2 h-2 rounded-full bg-green-400" />
              <span>Security Active</span>
            </div>
          </div>
        )}
      </div>

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-aviation-border-panel bg-aviation-bg-secondary">
          <div className="flex items-center gap-2">
            <Shield className="w-5 h-5 text-aviation-cyan" />
            <span className="text-sm font-medium">Security Center</span>
          </div>
          <span className="text-xs text-aviation-text-muted">{NAV_ITEMS.find(i => i.id === activePanel)?.label}</span>
        </div>

        <div className="flex-1 overflow-hidden">
          {activePanel === 'threats' && <ThreatDetectionRadar sectors={mockThreatSectors} className="h-full" />}
          {activePanel === 'timeline' && <SecurityTimeline events={mockEvents} className="h-full" />}
          {activePanel === 'permissions' && <PermissionMatrix roles={mockRoles} permissions={mockPermissions} className="h-full" />}
          {activePanel === 'sandbox' && <SandboxBoundaryViewer boundaries={mockBoundaries} className="h-full" />}
          {activePanel === 'api' && <APIExposureAnalyzer endpoints={mockEndpoints} className="h-full" />}
          {activePanel === 'credentials' && <CredentialAccessMonitor accesses={mockAccesses} className="h-full" />}
          {activePanel === 'isolation' && <RuntimeIsolationMap isolates={mockIsolates} className="h-full" />}
          {activePanel === 'compliance' && <ComplianceDashboard frameworks={mockFrameworks} className="h-full" />}
          {activePanel === 'zerotrust' && <ZeroTrustFlowViewer flows={mockFlows} className="h-full" />}
          {activePanel === 'malicious' && <MaliciousExecutionDetector executions={mockMalicious} className="h-full" />}
          {activePanel === 'audit' && <AuditTrailExplorer entries={mockAudit} className="h-full" />}
          {activePanel === 'encryption' && <EncryptionStatusPanel statuses={mockEncryption} className="h-full" />}
          {activePanel === 'behavior' && <SuspiciousBehaviorTimeline behaviors={mockBehavior} className="h-full" />}
          {activePanel === 'vulns' && <VulnerabilityScanner vulnerabilities={mockVulns} className="h-full" />}
          {activePanel === 'policies' && <PolicyEngineViewer policies={mockPolicies} className="h-full" />}
        </div>
      </div>
    </div>
  )
}

export default SecurityIntegration
