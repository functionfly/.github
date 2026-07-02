import { CodeOutput } from './CodeOutput';
import { ImageViewer } from './ImageViewer';
import { AudioPlayer } from './AudioPlayer';
import { Database, Globe, FileText, Bell, Mail, HardDrive } from 'lucide-react';

interface ToolResultRendererProps {
  tool: string;
  data: Record<string, unknown>;
}

function formatJSON(data: unknown): string {
  if (typeof data === 'string') return data;
  try {
    return JSON.stringify(data, null, 2);
  } catch {
    return String(data);
  }
}

export function ToolResultRenderer({ tool, data }: ToolResultRendererProps) {
  switch (tool) {
    case 'code_execution':
      return (
        <CodeOutput
          stdout={data.stdout as string}
          stderr={data.stderr as string}
          exitCode={data.exit_code as number}
          language={data.language as string}
        />
      );

    case 'image_generation':
      return (
        <ImageViewer
          url={data.url as string}
          prompt={data.prompt as string}
          size={data.size as string}
          path={data.path as string}
        />
      );

    case 'text_to_speech':
      return (
        <AudioPlayer
          url={data.url as string}
          path={data.path as string}
          format={data.format as string}
          text={data.text as string}
          voice={data.voice as string}
        />
      );

    case 'database_query': {
      const rows = data.rows as Array<Record<string, unknown>> | undefined;
      const columns = data.columns as string[] | undefined;
      const rowCount = data.row_count as number | undefined;
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', overflow: 'hidden',
        }}>
          <div style={{
            display: 'flex', alignItems: 'center', gap: '6px',
            padding: '6px 10px', background: 'var(--panel)',
            borderBottom: '1px solid var(--panel-edge)',
          }}>
            <Database size={12} style={{ color: 'var(--accent)' }} />
            <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '12px', color: 'var(--text-dim)' }}>
              Query Result
            </span>
            {rowCount !== undefined && (
              <span style={{ fontSize: '11px', color: 'var(--text-faint)' }}>{rowCount} rows</span>
            )}
          </div>
          {rows && rows.length > 0 ? (
            <div style={{ overflowX: 'auto', maxHeight: '300px', overflowY: 'auto' }}>
              <table style={{
                width: '100%', borderCollapse: 'collapse',
                fontFamily: 'var(--font-mono)', fontSize: '11px',
              }}>
                <thead>
                  <tr>
                    {(columns || Object.keys(rows[0])).map((col) => (
                      <th key={col} style={{
                        textAlign: 'left', padding: '4px 8px',
                        background: 'var(--panel)', color: 'var(--text-dim)',
                        borderBottom: '1px solid var(--panel-edge)',
                        fontWeight: 600, position: 'sticky', top: 0,
                      }}>
                        {col}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row, i) => (
                    <tr key={i}>
                      {(columns || Object.keys(row)).map((col) => (
                        <td key={col} style={{
                          padding: '4px 8px', color: 'var(--text)',
                          borderBottom: '1px solid var(--panel-edge)',
                          maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                        }}>
                          {row[col] === null ? <span style={{ color: 'var(--text-faint)' }}>NULL</span> : String(row[col])}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div style={{ padding: 'var(--space-3)', color: 'var(--text-faint)', fontSize: '12px', textAlign: 'center' }}>
              No results
            </div>
          )}
        </div>
      );
    }

    case 'http_request':
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', overflow: 'hidden',
        }}>
          <div style={{
            display: 'flex', alignItems: 'center', gap: '6px',
            padding: '6px 10px', background: 'var(--panel)',
            borderBottom: '1px solid var(--panel-edge)',
          }}>
            <Globe size={12} style={{ color: 'var(--accent)' }} />
            <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '12px', color: 'var(--text-dim)' }}>
              HTTP Response
            </span>
            {data.status !== undefined && (
              <span style={{
                fontFamily: 'var(--font-mono)', fontSize: '11px',
                color: (data.status as number) < 400 ? 'var(--status-ok)' : 'var(--status-revoked)',
                padding: '1px 6px', borderRadius: 'var(--radius-sm)',
                background: (data.status as number) < 400 ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
              }}>
                {data.status as number}
              </span>
            )}
          </div>
          <pre style={{
            margin: 0, padding: '8px 10px', fontFamily: 'var(--font-mono)',
            fontSize: '11px', color: 'var(--text)', whiteSpace: 'pre-wrap',
            wordBreak: 'break-all', maxHeight: '300px', overflowY: 'auto',
          }}>
            {formatJSON(data.body || data)}
          </pre>
        </div>
      );

    case 'web_search':
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', overflow: 'hidden',
        }}>
          <div style={{
            display: 'flex', alignItems: 'center', gap: '6px',
            padding: '6px 10px', background: 'var(--panel)',
            borderBottom: '1px solid var(--panel-edge)',
          }}>
            <Globe size={12} style={{ color: 'var(--accent)' }} />
            <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '12px', color: 'var(--text-dim)' }}>
              Search Results
            </span>
          </div>
          <pre style={{
            margin: 0, padding: '8px 10px', fontFamily: 'var(--font-mono)',
            fontSize: '11px', color: 'var(--text)', whiteSpace: 'pre-wrap',
            wordBreak: 'break-all', maxHeight: '300px', overflowY: 'auto',
          }}>
            {formatJSON(data)}
          </pre>
        </div>
      );

    case 'email':
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', padding: 'var(--space-3)',
          display: 'flex', alignItems: 'center', gap: 'var(--space-2)',
        }}>
          <Mail size={14} style={{ color: 'var(--status-ok)' }} />
          <span style={{ fontSize: '12px', color: 'var(--text)' }}>
            Email {data.status === 'sent' ? 'sent' : 'logged'}: {data.to as string} — {data.subject as string}
          </span>
        </div>
      );

    case 'notification':
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', padding: 'var(--space-3)',
          display: 'flex', alignItems: 'center', gap: 'var(--space-2)',
        }}>
          <Bell size={14} style={{ color: 'var(--status-ok)' }} />
          <span style={{ fontSize: '12px', color: 'var(--text)' }}>
            Notification {data.status === 'sent' ? 'sent' : 'logged'}: {data.title as string}
          </span>
        </div>
      );

    case 'file_read':
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', overflow: 'hidden',
        }}>
          <div style={{
            display: 'flex', alignItems: 'center', gap: '6px',
            padding: '6px 10px', background: 'var(--panel)',
            borderBottom: '1px solid var(--panel-edge)',
          }}>
            <FileText size={12} style={{ color: 'var(--accent)' }} />
            <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '12px', color: 'var(--text-dim)' }}>
              {data.path as string}
            </span>
            {data.size !== undefined && (
              <span style={{ fontSize: '11px', color: 'var(--text-faint)' }}>{data.size as number} bytes</span>
            )}
          </div>
          <pre style={{
            margin: 0, padding: '8px 10px', fontFamily: 'var(--font-mono)',
            fontSize: '11px', color: 'var(--text)', whiteSpace: 'pre-wrap',
            wordBreak: 'break-all', maxHeight: '300px', overflowY: 'auto',
          }}>
            {data.content as string}
          </pre>
        </div>
      );

    case 'file_write':
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', padding: 'var(--space-3)',
          display: 'flex', alignItems: 'center', gap: 'var(--space-2)',
        }}>
          <FileText size={14} style={{ color: 'var(--status-ok)' }} />
          <span style={{ fontSize: '12px', color: 'var(--text)' }}>
            Written: {data.path as string} ({data.bytes as number} bytes)
          </span>
        </div>
      );

    default:
      return (
        <div style={{
          background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--panel-edge)', overflow: 'hidden',
        }}>
          <div style={{
            display: 'flex', alignItems: 'center', gap: '6px',
            padding: '6px 10px', background: 'var(--panel)',
            borderBottom: '1px solid var(--panel-edge)',
          }}>
            <HardDrive size={12} style={{ color: 'var(--accent)' }} />
            <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '12px', color: 'var(--text-dim)' }}>
              {tool}
            </span>
          </div>
          <pre style={{
            margin: 0, padding: '8px 10px', fontFamily: 'var(--font-mono)',
            fontSize: '11px', color: 'var(--text)', whiteSpace: 'pre-wrap',
            wordBreak: 'break-all', maxHeight: '300px', overflowY: 'auto',
          }}>
            {formatJSON(data)}
          </pre>
        </div>
      );
  }
}
