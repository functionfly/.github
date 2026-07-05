import { Brain, Key, Loader2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAIKeys } from '@/api/ai-keys';
import { Button } from '@/components/ui/button';
import { useTranslation } from 'react-i18next';

interface AIFeatureGateProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  showPlatformAIIndicator?: boolean;
}

export function AIFeatureGate({ children, fallback, showPlatformAIIndicator = false }: AIFeatureGateProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: keys = [], isLoading } = useAIKeys();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-6 w-6 animate-spin" style={{ color: 'var(--text-faint)' }} />
      </div>
    );
  }

  const hasKeys = keys.length > 0;

  if (hasKeys || fallback) {
    return (
      <>
        {children}
        {showPlatformAIIndicator && !hasKeys && (
          <div
            className="inline-flex items-center gap-1.5 px-2 py-1 rounded text-xs"
            style={{
              background: 'rgba(59, 130, 246, 0.1)',
              color: 'rgba(59, 130, 246, 0.9)',
              border: '1px solid rgba(59, 130, 246, 0.2)',
            }}
          >
            <Brain className="h-3 w-3" />
            Using platform AI (no key configured)
          </div>
        )}
      </>
    );
  }

  return (
    fallback ?? (
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '48px 24px',
          textAlign: 'center',
          background: 'var(--panel)',
          border: '1px solid var(--panel-edge)',
          borderRadius: 'var(--radius-lg)',
        }}
      >
        <div
          style={{
            width: 64,
            height: 64,
            borderRadius: 'var(--radius-lg)',
            background: 'rgba(139, 92, 246, 0.1)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginBottom: 16,
          }}
        >
          <Brain className="h-8 w-8" style={{ color: 'rgba(139, 92, 246, 0.8)' }} />
        </div>
        <h3 className="text-lg font-semibold mb-2" style={{ fontFamily: 'var(--font-display)', color: 'var(--text)' }}>
          AI Keys Required
        </h3>
        <p className="text-sm max-w-sm mb-6" style={{ color: 'var(--text-dim)' }}>
          Connect at least one AI provider key to enable AI-powered agents and workflows.
          Your keys are encrypted and never stored in plaintext.
        </p>
        <Button onClick={() => navigate('/settings#ai-keys')} className="gap-2">
          <Key className="h-4 w-4" />
          Connect AI Keys
        </Button>
        <p className="text-xs mt-4" style={{ color: 'var(--text-faint)' }}>
          Don't have an AI key?{' '}
          <a
            href="/docs/ai-models"
            target="_blank"
            rel="noopener noreferrer"
            className="underline underline-offset-2"
            style={{ color: 'var(--status-ok)' }}
          >
            Learn about supported providers
          </a>
        </p>
      </div>
    )
  );
}
