import { Folder, File, ArrowLeft, RefreshCw, HardDrive, Clock, FileText, Database, Image, Volume2, Code, Search } from 'lucide-react';
import { FrameButton } from '@/components/containment';
import { useWorkspace, type WorkspaceEntry } from '../hooks/useWorkspace';
import { formatDistanceToNow } from 'date-fns';

interface WorkspaceViewProps {
  agentId: string;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function fileIcon(name: string, isDir: boolean) {
  if (isDir) return <Folder size={16} />;
  const ext = name.split('.').pop()?.toLowerCase();
  switch (ext) {
    case 'py': case 'js': case 'ts': case 'tsx': case 'go': case 'rs': case 'sh':
      return <Code size={16} />;
    case 'png': case 'jpg': case 'jpeg': case 'gif': case 'svg': case 'webp':
      return <Image size={16} />;
    case 'mp3': case 'wav': case 'ogg': case 'flac': case 'aac':
      return <Volume2 size={16} />;
    case 'csv': case 'json': case 'parquet': case 'sql':
      return <Database size={16} />;
    case 'md': case 'txt': case 'log':
      return <FileText size={16} />;
    default:
      return <File size={16} />;
  }
}

function toolIcon(tool: string) {
  switch (tool) {
    case 'file_write': return <File size={12} />;
    case 'file_read': return <FileText size={12} />;
    case 'code_execution': return <Code size={12} />;
    case 'image_generation': return <Image size={12} />;
    case 'text_to_speech': return <Volume2 size={12} />;
    case 'database_query': return <Database size={12} />;
    case 'web_search': return <Search size={12} />;
    default: return <HardDrive size={12} />;
  }
}

export function WorkspaceView({ agentId }: WorkspaceViewProps) {
  const { entries, currentPath, manifest, history, loading, error, navigate, refresh } = useWorkspace(agentId);

  const pathParts = currentPath.split('/').filter(Boolean);
  const canGoUp = currentPath !== '/';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)', flex: 1, minHeight: 0 }}>
      {/* Header */}
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Workspace</h2>
          <p className="aw-center__subtitle">Agent filesystem, outputs, and execution history</p>
        </div>
        <FrameButton size="sm" onClick={refresh} iconLeft={<RefreshCw size={12} />}>
          Refresh
        </FrameButton>
      </div>

      {/* Stats */}
      {manifest && (
        <div className="aw-stats">
          <div className="aw-stat">
            <span className="aw-stat__label">Files</span>
            <span className="aw-stat__value">{manifest.file_count}</span>
          </div>
          <div className="aw-stat">
            <span className="aw-stat__label">Total Size</span>
            <span className="aw-stat__value">{formatBytes(manifest.total_bytes)}</span>
          </div>
          <div className="aw-stat">
            <span className="aw-stat__label">Created</span>
            <span className="aw-stat__value">{manifest.created_at ? formatDistanceToNow(new Date(manifest.created_at), { addSuffix: true }) : '—'}</span>
          </div>
          <div className="aw-stat">
            <span className="aw-stat__label">Last Updated</span>
            <span className="aw-stat__value">{manifest.updated_at ? formatDistanceToNow(new Date(manifest.updated_at), { addSuffix: true }) : '—'}</span>
          </div>
        </div>
      )}

      <div style={{ display: 'flex', gap: 'var(--space-4)', flex: 1, minHeight: 0 }}>
        {/* File Browser */}
        <div style={{ flex: 2, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <div className="aw-card" style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
            {/* Breadcrumb */}
            <div className="aw-card__header" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
              <button
                onClick={() => canGoUp ? navigate(pathParts.slice(0, -1).join('/') || '') : undefined}
                style={{ background: 'none', border: 'none', color: canGoUp ? 'var(--accent)' : 'var(--text-faint)', cursor: canGoUp ? 'pointer' : 'default', padding: 2, display: 'flex' }}
                disabled={!canGoUp}
              >
                <ArrowLeft size={14} />
              </button>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-dim)' }}>
                /{pathParts.join(' / ') || 'root'}
              </span>
            </div>

            {/* File List */}
            <div className="aw-card__body aw-card__body--flush" style={{ flex: 1, overflowY: 'auto' }}>
              {loading ? (
                <div className="aw-loading"><div className="aw-loading__spinner" /></div>
              ) : error ? (
                <div className="aw-empty">
                  <HardDrive size={32} className="aw-empty__icon" />
                  <span className="aw-empty__title">Error</span>
                  <span className="aw-empty__desc">{error}</span>
                </div>
              ) : entries.length === 0 ? (
                <div className="aw-empty">
                  <Folder size={32} className="aw-empty__icon" />
                  <span className="aw-empty__title">Empty directory</span>
                  <span className="aw-empty__desc">No files in this location yet</span>
                </div>
              ) : (
                <div>
                  {entries
                    .sort((a, b) => {
                      if (a.is_dir !== b.is_dir) return a.is_dir ? -1 : 1;
                      return a.name.localeCompare(b.name);
                    })
                    .map((entry) => (
                      <EntryRow key={entry.path} entry={entry} onNavigate={navigate} />
                    ))}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* History Panel */}
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <div className="aw-card" style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
            <div className="aw-card__header">
              <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                <Clock size={14} />
                History
                <span style={{ fontSize: '11px', color: 'var(--text-faint)' }}>{history.length} actions</span>
              </span>
            </div>
            <div className="aw-card__body aw-card__body--flush" style={{ flex: 1, overflowY: 'auto' }}>
              {history.length === 0 ? (
                <div className="aw-empty">
                  <Clock size={32} className="aw-empty__icon" />
                  <span className="aw-empty__title">No activity</span>
                  <span className="aw-empty__desc">Tool actions will appear here</span>
                </div>
              ) : (
                <div>
                  {[...history].reverse().slice(0, 50).map((entry, i) => (
                    <div key={i} style={{
                      display: 'flex', alignItems: 'flex-start', gap: 'var(--space-2)',
                      padding: 'var(--space-2) var(--space-3)',
                      borderBottom: '1px solid var(--panel-edge)',
                      fontSize: '12px',
                    }}>
                      <span style={{ color: 'var(--accent)', flexShrink: 0, marginTop: 2 }}>
                        {toolIcon(entry.tool)}
                      </span>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontWeight: 600, color: 'var(--text)' }}>{entry.tool}</div>
                        <div style={{ color: 'var(--text-dim)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {entry.description}
                        </div>
                        {entry.path && (
                          <div style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-faint)' }}>
                            {entry.path}
                          </div>
                        )}
                      </div>
                      <span style={{ color: 'var(--text-faint)', fontSize: '11px', flexShrink: 0, whiteSpace: 'nowrap' }}>
                        {formatDistanceToNow(new Date(entry.timestamp), { addSuffix: true })}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Structure */}
      {manifest && manifest.structure && manifest.structure.length > 0 && (
        <div className="aw-card">
          <div className="aw-card__header">
            <span className="aw-card__title">Directory Structure</span>
          </div>
          <div className="aw-card__body">
            <pre style={{
              fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-dim)',
              margin: 0, whiteSpace: 'pre-wrap', lineHeight: 1.6,
            }}>
              {manifest.structure.join('\n')}
            </pre>
          </div>
        </div>
      )}
    </div>
  );
}

function EntryRow({ entry, onNavigate }: { entry: WorkspaceEntry; onNavigate: (path: string) => void }) {
  return (
    <div
      onClick={() => entry.is_dir ? onNavigate(entry.path) : undefined}
      style={{
        display: 'flex', alignItems: 'center', gap: 'var(--space-2)',
        padding: 'var(--space-2) var(--space-3)',
        borderBottom: '1px solid var(--panel-edge)',
        cursor: entry.is_dir ? 'pointer' : 'default',
        transition: 'background var(--duration-fast) var(--ease-out)',
      }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--panel-raised)'; }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = ''; }}
    >
      <span style={{ color: entry.is_dir ? 'var(--accent)' : 'var(--text-faint)', flexShrink: 0 }}>
        {fileIcon(entry.name, entry.is_dir)}
      </span>
      <span style={{
        flex: 1, fontFamily: 'var(--font-mono)', fontSize: '13px',
        color: entry.is_dir ? 'var(--accent)' : 'var(--text)',
        overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
      }}>
        {entry.name}{entry.is_dir ? '/' : ''}
      </span>
      {!entry.is_dir && (
        <span style={{ fontSize: '11px', color: 'var(--text-faint)', flexShrink: 0 }}>
          {formatBytes(entry.size)}
        </span>
      )}
      <span style={{ fontSize: '11px', color: 'var(--text-faint)', flexShrink: 0 }}>
        {formatDistanceToNow(new Date(entry.mod_time), { addSuffix: true })}
      </span>
    </div>
  );
}
