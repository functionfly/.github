import { Brain, ExternalLink, Key, X } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useLocalStorage } from '@/hooks/useInfiniteScroll';
import { Button } from '@/components/ui/button';

interface BYOKPromptBannerProps {
  className?: string;
}

export function BYOKPromptBanner({ className }: BYOKPromptBannerProps) {
  const navigate = useNavigate();
  const [dismissed, setDismissed] = useLocalStorage('ff-byok-banner-dismissed', false);

  if (dismissed) return null;

  return (
    <div
      className={className}
      style={{
        background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.08) 0%, rgba(59, 130, 246, 0.08) 100%)',
        border: '1px solid rgba(139, 92, 246, 0.2)',
        borderRadius: 'var(--radius-lg)',
        padding: '12px 16px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: '16px',
      }}
    >
      <div className="flex items-center gap-3">
        <div
          style={{
            width: 36,
            height: 36,
            borderRadius: 'var(--radius)',
            background: 'rgba(139, 92, 246, 0.15)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Brain className="h-5 w-5" style={{ color: 'rgba(139, 92, 246, 0.9)' }} />
        </div>
        <div>
          <p className="text-sm font-medium" style={{ color: 'var(--text)' }}>
            Bring Your Own AI Keys
          </p>
          <p className="text-xs" style={{ color: 'var(--text-dim)' }}>
            Connect your OpenAI, Anthropic, or OpenRouter key to enable AI features. Pay providers directly — no platform markup.
          </p>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Button
          size="sm"
          onClick={() => navigate('/settings#ai-keys')}
          className="gap-1.5"
        >
          <Key className="h-3.5 w-3.5" />
          Connect AI Keys
        </Button>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => window.open('/docs/ai-models', '_blank')}
          className="gap-1 text-muted-foreground"
        >
          <ExternalLink className="h-3.5 w-3.5" />
          Learn More
        </Button>
        <button
          onClick={() => setDismissed(true)}
          className="p-1 rounded transition-colors"
          style={{ color: 'var(--text-faint)' }}
          aria-label="Dismiss"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
