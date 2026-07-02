import { Image, ExternalLink } from 'lucide-react';

interface ImageViewerProps {
  url?: string;
  prompt?: string;
  size?: string;
  path?: string;
}

export function ImageViewer({ url, prompt, size, path }: ImageViewerProps) {
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
        <Image size={12} style={{ color: 'var(--accent)' }} />
        <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '12px', color: 'var(--text-dim)' }}>
          Image Generation
        </span>
        {size && (
          <span style={{ fontSize: '11px', color: 'var(--text-faint)' }}>{size}</span>
        )}
      </div>

      {url ? (
        <div style={{ padding: 'var(--space-2)', textAlign: 'center' }}>
          <img
            src={url}
            alt={prompt || 'Generated image'}
            style={{
              maxWidth: '100%', maxHeight: '400px', borderRadius: 'var(--radius-sm)',
              objectFit: 'contain',
            }}
          />
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            style={{
              display: 'inline-flex', alignItems: 'center', gap: '4px',
              marginTop: 'var(--space-2)', fontSize: '11px', color: 'var(--accent)',
              textDecoration: 'none',
            }}
          >
            <ExternalLink size={10} /> Open full size
          </a>
        </div>
      ) : (
        <div style={{ padding: 'var(--space-3)', textAlign: 'center' }}>
          {path && (
            <div style={{
              fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-dim)',
              padding: 'var(--space-2)', background: 'var(--panel)',
              borderRadius: 'var(--radius-sm)', marginBottom: 'var(--space-2)',
            }}>
              Saved to: {path}
            </div>
          )}
          <div style={{ color: 'var(--text-faint)', fontSize: '12px' }}>
            {prompt ? `Prompt: "${prompt.slice(0, 100)}${prompt.length > 100 ? '...' : ''}"` : 'Image generation queued'}
          </div>
        </div>
      )}
    </div>
  );
}
