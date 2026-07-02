import { Terminal, Copy, Check } from 'lucide-react';
import { useState } from 'react';

interface CodeOutputProps {
  stdout?: string;
  stderr?: string;
  exitCode?: number;
  language?: string;
}

export function CodeOutput({ stdout, stderr, exitCode, language }: CodeOutputProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    const text = [stdout, stderr].filter(Boolean).join('\n');
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div style={{
      background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
      border: '1px solid var(--panel-edge)', overflow: 'hidden', fontSize: '12px',
    }}>
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '6px 10px', background: 'var(--panel)',
        borderBottom: '1px solid var(--panel-edge)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <Terminal size={12} style={{ color: 'var(--accent)' }} />
          <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, color: 'var(--text-dim)' }}>
            {language || 'output'}
          </span>
          {exitCode !== undefined && (
            <span style={{
              fontFamily: 'var(--font-mono)', fontSize: '11px',
              color: exitCode === 0 ? 'var(--status-ok)' : 'var(--status-revoked)',
              padding: '1px 6px', borderRadius: 'var(--radius-sm)',
              background: exitCode === 0 ? 'rgba(34,197,94,0.1)' : 'rgba(239,68,68,0.1)',
            }}>
              exit {exitCode}
            </span>
          )}
        </div>
        <button
          onClick={handleCopy}
          style={{
            background: 'none', border: 'none', cursor: 'pointer',
            color: copied ? 'var(--status-ok)' : 'var(--text-faint)',
            padding: 2, display: 'flex',
          }}
        >
          {copied ? <Check size={12} /> : <Copy size={12} />}
        </button>
      </div>
      {stdout && (
        <pre style={{
          margin: 0, padding: '8px 10px', fontFamily: 'var(--font-mono)',
          fontSize: '12px', color: 'var(--text)', whiteSpace: 'pre-wrap',
          wordBreak: 'break-all', maxHeight: '300px', overflowY: 'auto',
        }}>
          {stdout}
        </pre>
      )}
      {stderr && (
        <pre style={{
          margin: 0, padding: '8px 10px', fontFamily: 'var(--font-mono)',
          fontSize: '12px', color: 'var(--status-revoked)', whiteSpace: 'pre-wrap',
          wordBreak: 'break-all', maxHeight: '200px', overflowY: 'auto',
          borderTop: stdout ? '1px solid var(--panel-edge)' : undefined,
        }}>
          {stderr}
        </pre>
      )}
      {!stdout && !stderr && (
        <div style={{ padding: '12px 10px', color: 'var(--text-faint)', fontStyle: 'italic' }}>
          No output
        </div>
      )}
    </div>
  );
}
