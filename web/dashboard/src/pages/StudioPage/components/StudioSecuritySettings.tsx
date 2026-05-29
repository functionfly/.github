import {
  PermissionMatrix,
  SecurityTimeline,
  ThreatDetectionRadar,
} from '@functionfly/ui-security';

const sampleRoles = [
  {
    id: 'admin',
    name: 'Admin',
    permissions: ['p1', 'p2', 'p3', 'p4'],
    userCount: 2,
    createdAt: Date.now() - 86400000 * 30,
    isSystem: true,
  },
  {
    id: 'developer',
    name: 'Developer',
    permissions: ['p2', 'p4'],
    userCount: 8,
    createdAt: Date.now() - 86400000 * 14,
  },
  {
    id: 'viewer',
    name: 'Viewer',
    permissions: ['p3'],
    userCount: 12,
    createdAt: Date.now() - 86400000 * 7,
  },
];

const samplePermissions = [
  { id: 'p1', name: 'Manage Agents', resource: 'agents', action: 'admin' as const },
  { id: 'p2', name: 'Edit Workflows', resource: 'workflows', action: 'write' as const },
  { id: 'p3', name: 'Read Vault', resource: 'vault', action: 'read' as const },
  { id: 'p4', name: 'Deploy Functions', resource: 'functions', action: 'write' as const },
];

const sampleSectors = [
  {
    id: 'runtime',
    name: 'Runtime',
    angle: 0,
    radius: 0.85,
    severity: 'high' as const,
    threats: [
      {
        id: 't1',
        type: 'intrusion' as const,
        severity: 'high' as const,
        source: 'agent-runtime',
        target: 'workflow-parser',
        timestamp: Date.now() - 3600000,
        status: 'detected' as const,
        description: 'Unsanitized input in workflow node parser',
      },
    ],
  },
  {
    id: 'plugins',
    name: 'Plugins',
    angle: 120,
    radius: 0.65,
    severity: 'medium' as const,
    threats: [
      {
        id: 't2',
        type: 'data-breach' as const,
        severity: 'medium' as const,
        source: 'plugin-sandbox',
        target: 'network-egress',
        timestamp: Date.now() - 7200000,
        status: 'investigating' as const,
        description: 'Plugin requested elevated network scope',
      },
    ],
  },
  {
    id: 'auth',
    name: 'Auth',
    angle: 240,
    radius: 0.45,
    severity: 'low' as const,
    threats: [],
  },
];

const sampleEvents = [
  {
    id: 'e1',
    type: 'login' as const,
    severity: 'info' as const,
    message: 'User signed in from new device',
    timestamp: Date.now() - 1800000,
    userName: 'admin@functionfly.local',
    source: 'auth',
  },
  {
    id: 'e2',
    type: 'alert' as const,
    severity: 'warning' as const,
    message: 'Agent attempted outbound connection to unknown host',
    timestamp: Date.now() - 5400000,
    userName: 'agent-7',
    source: 'runtime-guard',
  },
];

export function StudioSecuritySettings() {
  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-text-primary mb-2">Permission Matrix</h3>
        <p className="text-sm text-text-muted mb-4">
          Review role-based access for Studio agents, workflows, and vault operations.
        </p>
        <PermissionMatrix
          roles={sampleRoles}
          permissions={samplePermissions}
          className="rounded-lg border border-border-subtle overflow-hidden min-h-[240px]"
        />
      </div>

      <div>
        <h3 className="text-lg font-semibold text-text-primary mb-2">Threat Detection</h3>
        <ThreatDetectionRadar
          sectors={sampleSectors}
          className="rounded-lg border border-border-subtle p-4 bg-bg-secondary min-h-[280px]"
        />
      </div>

      <div>
        <h3 className="text-lg font-semibold text-text-primary mb-2">Security Timeline</h3>
        <SecurityTimeline
          events={sampleEvents}
          className="rounded-lg border border-border-subtle p-4 bg-bg-secondary"
        />
      </div>
    </div>
  );
}
