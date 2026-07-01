import { useAgent, useUpdateAgent } from '@/hooks/useAgent';
import { useConnectorStatuses } from '@/hooks/useConnectors';
import { useMCPServers, useDeleteMCPServer } from '@/hooks/useMCPServers';
import { normalizeAgentIdentity } from '@/api/agent';
import { FrameButton } from '@/components/containment';
import { AddMCPServerModal } from '../components/AddMCPServerModal';
import {
  Bell,
  Code,
  Database,
  FileEdit,
  FileText,
  Globe,
  Image,
  Mail,
  Plus,
  Settings,
  Terminal,
  Trash2,
  Volume2,
  Wrench,
} from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

interface ToolsViewProps {
  agentId: string;
  setRightContext: (ctx: { type: string; id: string } | null) => void;
}

const TOOL_DEFINITIONS = [
  { key: 'web_search', name: 'Web Search', desc: 'Search the web for information', icon: Globe },
  { key: 'database_query', name: 'Database Query', desc: 'Query databases and data stores', icon: Database },
  { key: 'file_read', name: 'File Read', desc: 'Read files from the filesystem', icon: FileText },
  { key: 'file_write', name: 'File Write', desc: 'Write and create files', icon: FileEdit },
  { key: 'http_request', name: 'HTTP Request', desc: 'Make HTTP API requests', icon: Globe },
  { key: 'code_execution', name: 'Code Execution', desc: 'Execute code in sandboxed environments', icon: Terminal },
  { key: 'image_generation', name: 'Image Generation', desc: 'Generate images from prompts', icon: Image },
  { key: 'text_to_speech', name: 'Text to Speech', desc: 'Convert text to audio', icon: Volume2 },
  { key: 'email', name: 'Email', desc: 'Send emails and notifications', icon: Mail },
  { key: 'notification', name: 'Notification', desc: 'Push notifications and alerts', icon: Bell },
];

export function ToolsView({ agentId, setRightContext }: ToolsViewProps) {
  const { data: agentData } = useAgent(agentId);
  const updateAgent = useUpdateAgent(agentId);
  const { statuses: connectorStatuses, isLoading: connectorsLoading } = useConnectorStatuses();
  const { data: mcpServers, isLoading: mcpLoading } = useMCPServers(agentId);
  const deleteMCP = useDeleteMCPServer(agentId);
  const [showAddMCP, setShowAddMCP] = useState(false);

  const agent = agentData?.agent ? normalizeAgentIdentity(agentData.agent) : undefined;
  const capabilities = useMemo(() => (agent?.capabilities ?? {}) as Record<string, string>, [agent]);

  const isToolEnabled = useCallback((key: string) => {
    return capabilities[key] === 'true' || capabilities[key] === 'enabled';
  }, [capabilities]);

  const toggleTool = useCallback(async (key: string) => {
    if (!agent) return;
    const current = isToolEnabled(key);
    const newCaps = { ...capabilities, [key]: current ? 'false' : 'true' };
    await updateAgent.mutateAsync({ capabilities: newCaps } as any);
  }, [agent, capabilities, isToolEnabled, updateAgent]);

  const connectedConnectors = connectorStatuses.filter(c => c.connected);
  const disconnectedConnectors = connectorStatuses.filter(c => !c.connected);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Tools & Abilities</h2>
          <p className="aw-center__subtitle">Configure what this agent can do</p>
        </div>
      </div>

      {/* Tool Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 'var(--space-3)' }}>
        {TOOL_DEFINITIONS.map(tool => {
          const enabled = isToolEnabled(tool.key);
          const Icon = tool.icon;
          return (
            <div key={tool.key} className="aw-tool-card">
              <div className="aw-tool-card__icon">
                <Icon size={16} />
              </div>
              <div className="aw-tool-card__info">
                <p className="aw-tool-card__name">{tool.name}</p>
                <p className="aw-tool-card__desc">{tool.desc}</p>
              </div>
              <button
                className={`aw-switch ${enabled ? 'aw-switch--on' : ''}`}
                onClick={() => toggleTool(tool.key)}
                disabled={updateAgent.isPending}
              >
                <span className="aw-switch__thumb" />
              </button>
            </div>
          );
        })}
      </div>

      {/* MCP Servers */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <Wrench size={14} />
            MCP Servers
          </span>
          <FrameButton size="sm" iconLeft={<Plus size={12} />} onClick={() => setShowAddMCP(true)}>
            Add Server
          </FrameButton>
        </div>
        <div className="aw-card__body">
          {mcpLoading ? (
            <div style={{ display: 'flex', gap: 'var(--space-2)', flexDirection: 'column' }}>
              {Array.from({ length: 2 }).map((_, i) => (
                <div key={i} style={{ height: 44, background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', opacity: 0.4 }} />
              ))}
            </div>
          ) : !mcpServers || mcpServers.length === 0 ? (
            <div className="aw-empty" style={{ padding: 'var(--space-4)' }}>
              <Wrench size={32} className="aw-empty__icon" />
              <span className="aw-empty__title">No MCP servers configured</span>
              <span className="aw-empty__desc">Connect MCP servers to extend agent capabilities</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              {mcpServers.map(server => (
                <div key={server.id} style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 'var(--space-3)',
                  padding: 'var(--space-2) var(--space-3)',
                  background: 'var(--panel-raised)',
                  border: '1px solid var(--panel-edge)',
                  borderRadius: 'var(--radius)',
                  opacity: server.enabled ? 1 : 0.6,
                }}>
                  <div style={{
                    width: 32,
                    height: 32,
                    borderRadius: 'var(--radius-sm)',
                    background: 'var(--panel)',
                    border: '1px solid var(--panel-edge)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}>
                    <Globe size={14} style={{ color: 'var(--text-dim)' }} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                      <span style={{ fontFamily: 'var(--font-body)', fontSize: '13px', fontWeight: 600, color: 'var(--text)' }}>
                        {server.name}
                      </span>
                      <span style={{
                        fontFamily: 'var(--font-mono)',
                        fontSize: '9px',
                        fontWeight: 500,
                        letterSpacing: '0.04em',
                        padding: '1px var(--space-1)',
                        borderRadius: 'var(--radius-sm)',
                        color: 'var(--foil-a)',
                        background: 'rgba(159, 216, 255, 0.08)',
                        border: '1px solid rgba(159, 216, 255, 0.2)',
                      }}>
                        {server.transport}
                      </span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginTop: '2px' }}>
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-faint)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {server.url}
                      </span>
                      {server.tool_count > 0 && (
                        <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-faint)', whiteSpace: 'nowrap' }}>
                          {server.tool_count} tools
                        </span>
                      )}
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                    <span className={`aw-status-bar__dot ${server.last_error ? 'aw-status-bar__dot--error' : server.last_connected_at ? 'aw-status-bar__dot--live' : 'aw-status-bar__dot--off'}`} />
                    <button
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 24,
                        height: 24,
                        background: 'transparent',
                        border: 'none',
                        borderRadius: 'var(--radius-sm)',
                        color: 'var(--text-faint)',
                        cursor: 'pointer',
                      }}
                      onClick={() => deleteMCP.mutate(server.id)}
                      title="Remove server"
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Connectors */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title">Connector Status</span>
          <Link to="/settings#integrations">
            <FrameButton size="sm" iconLeft={<Settings size={12} />}>
              Manage
            </FrameButton>
          </Link>
        </div>
        <div className="aw-card__body">
          {connectorsLoading ? (
            <div style={{ display: 'flex', gap: 'var(--space-2)', flexDirection: 'column' }}>
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} style={{ height: 36, background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', opacity: 0.4 }} />
              ))}
            </div>
          ) : connectorStatuses.length === 0 ? (
            <div className="aw-empty" style={{ padding: 'var(--space-4)' }}>
              <Wrench size={32} className="aw-empty__icon" />
              <span className="aw-empty__title">No connectors available</span>
              <span className="aw-empty__desc">Connectors will appear here once configured</span>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              {connectedConnectors.length > 0 && connectedConnectors.map(conn => (
                <ConnectorRow key={conn.slug} name={conn.name} connected />
              ))}
              {connectedConnectors.length === 0 && (
                <div className="aw-empty" style={{ padding: 'var(--space-4)' }}>
                  <Wrench size={32} className="aw-empty__icon" />
                  <span className="aw-empty__title">No connectors enabled</span>
                  <span className="aw-empty__desc">
                    <Link to="/settings#integrations" style={{ color: 'var(--accent)', textDecoration: 'none' }}>
                      Enable connectors in Settings
                    </Link>
                    {' '}to extend agent capabilities
                  </span>
                </div>
              )}
              {disconnectedConnectors.length > 0 && (
                <>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', padding: 'var(--space-2) 0 var(--space-1)' }}>
                    <span style={{ flex: 1, height: 1, background: 'var(--panel-edge)' }} />
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>
                      Available
                    </span>
                    <span style={{ flex: 1, height: 1, background: 'var(--panel-edge)' }} />
                  </div>
                  {disconnectedConnectors.map(conn => (
                    <ConnectorRow key={conn.slug} name={conn.name} connected={false} />
                  ))}
                </>
              )}
            </div>
          )}
        </div>
      </div>

      <AddMCPServerModal open={showAddMCP} onClose={() => setShowAddMCP(false)} agentId={agentId} />
    </div>
  );
}

function ConnectorRow({ name, connected }: { name: string; connected: boolean }) {
  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: 'var(--space-2) var(--space-3)',
      background: 'var(--panel-raised)',
      border: '1px solid var(--panel-edge)',
      borderRadius: 'var(--radius)',
      opacity: connected ? 1 : 0.6,
    }}>
      <span style={{ fontFamily: 'var(--font-body)', fontSize: '13px', fontWeight: 500, color: 'var(--text)' }}>
        {name}
      </span>
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
        <span className={`aw-status-bar__dot ${connected ? 'aw-status-bar__dot--live' : 'aw-status-bar__dot--off'}`} />
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: connected ? 'var(--status-ok)' : 'var(--text-faint)' }}>
          {connected ? 'connected' : 'disconnected'}
        </span>
      </div>
    </div>
  );
}
