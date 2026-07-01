import { useState } from 'react';
import { Modal } from '@/components/containment';
import { useAddMCPServer } from '@/hooks/useMCPServers';
import { Globe, Terminal } from 'lucide-react';

interface AddMCPServerModalProps {
  open: boolean;
  onClose: () => void;
  agentId: string;
}

export function AddMCPServerModal({ open, onClose, agentId }: AddMCPServerModalProps) {
  const addServer = useAddMCPServer(agentId);
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [transport, setTransport] = useState<'streamable-http' | 'stdio' | 'sse'>('streamable-http');
  const [description, setDescription] = useState('');
  const [headersJson, setHeadersJson] = useState('');
  const [headerError, setHeaderError] = useState('');

  const resetForm = () => {
    setName('');
    setUrl('');
    setTransport('streamable-http');
    setDescription('');
    setHeadersJson('');
    setHeaderError('');
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !url.trim()) return;

    let headers: Record<string, string> | undefined;
    if (headersJson.trim()) {
      try {
        headers = JSON.parse(headersJson);
        setHeaderError('');
      } catch {
        setHeaderError('Invalid JSON');
        return;
      }
    }

    await addServer.mutateAsync({
      name: name.trim(),
      url: url.trim(),
      transport,
      description: description.trim() || undefined,
      headers,
    });
    resetForm();
    onClose();
  };

  return (
    <Modal open={open} onClose={handleClose} title="Add MCP Server">
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
        <p style={{ fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text-dim)', margin: 0 }}>
          Connect an external MCP server to give this agent additional tools and capabilities.
        </p>

        <FormField label="Server Name" required>
          <input
            className="input"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. My MCP Server"
            required
          />
        </FormField>

        <FormField label="Endpoint URL" required>
          <input
            className="input"
            type="text"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder={transport === 'stdio' ? 'e.g. npx @modelcontextprotocol/server-filesystem' : 'e.g. https://mcp.example.com/sse'}
            required
          />
        </FormField>

        <FormField label="Transport">
          <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
            {(['streamable-http', 'sse', 'stdio'] as const).map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => setTransport(t)}
                style={{
                  flex: 1,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 'var(--space-1)',
                  padding: 'var(--space-2) var(--space-3)',
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  fontWeight: transport === t ? 600 : 400,
                  color: transport === t ? 'var(--text)' : 'var(--text-dim)',
                  background: transport === t ? 'var(--panel-raised)' : 'transparent',
                  border: `1px solid ${transport === t ? 'var(--steel-light)' : 'var(--panel-edge)'}`,
                  borderRadius: 'var(--radius)',
                  cursor: 'pointer',
                  transition: 'all 120ms ease-out',
                }}
              >
                {t === 'stdio' ? <Terminal size={12} /> : <Globe size={12} />}
                {t}
              </button>
            ))}
          </div>
        </FormField>

        <FormField label="Description">
          <input
            className="input"
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Optional description"
          />
        </FormField>

        <FormField label="Headers (JSON)" error={headerError}>
          <input
            className="input"
            type="text"
            value={headersJson}
            onChange={(e) => { setHeadersJson(e.target.value); setHeaderError(''); }}
            placeholder='{"Authorization": "Bearer ..."}'
            style={{ fontFamily: 'var(--font-mono)', fontSize: '13px' }}
          />
        </FormField>

        <div style={{ display: 'flex', gap: 'var(--space-2)', justifyContent: 'flex-end', paddingTop: 'var(--space-2)' }}>
          <button
            type="button"
            className="frame-button frame-button--md"
            onClick={handleClose}
          >
            Cancel
          </button>
          <button
            type="submit"
            className="sealed-button sealed-button--md"
            disabled={!name.trim() || !url.trim() || addServer.isPending}
          >
            {addServer.isPending ? 'Adding…' : 'Add Server'}
          </button>
        </div>
      </form>
    </Modal>
  );
}

function FormField({ label, required, error, children }: { label: string; required?: boolean; error?: string; children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
      <label style={{ fontFamily: 'var(--font-body)', fontSize: '13px', fontWeight: 500, color: 'var(--text)' }}>
        {label}
        {required && <span style={{ color: 'var(--accent)', marginLeft: '2px' }}>*</span>}
      </label>
      {children}
      {error && (
        <span style={{ fontFamily: 'var(--font-body)', fontSize: '12px', color: 'var(--status-revoked)' }}>{error}</span>
      )}
    </div>
  );
}
