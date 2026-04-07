/**
 * Stub for @marsidev/react-turnstile
 * This is a placeholder to fix TypeScript build errors.
 * Install the actual package for production use:
 *   bun add @marsidev/react-turnstile
 */

import React from 'react';

interface TurnstileProps {
  siteKey?: string;
  action?: string;
  onSuccess?: (token: string) => void;
  onError?: () => void;
  onExpire?: () => void;
  onLoad?: () => void;
  options?: {
    theme?: 'light' | 'dark' | 'auto';
    size?: 'normal' | 'compact' | 'invisible';
    language?: string;
    appearance?: 'always' | 'execute' | 'interaction-only';
  };
  className?: string;
  style?: React.CSSProperties;
  scriptOptions?: {
    async?: boolean;
    defer?: boolean;
    appendTo?: 'head' | 'body';
  };
}

export function Turnstile({ 
  siteKey, 
  onSuccess,
  className,
  style 
}: TurnstileProps): React.ReactElement {
  // Stub implementation - renders a placeholder div
  return (
    <div 
      className={className} 
      style={{
        ...style,
        width: '100%',
        height: '65px',
        background: 'var(--bg-secondary, #f3f4f6)',
        border: '1px dashed var(--border-default, #d1d5db)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: '12px',
        color: 'var(--text-secondary, #6b7280)',
      }}
    >
      Turnstile Captcha (stub - install @marsidev/react-turnstile)
    </div>
  );
}

export default Turnstile;
