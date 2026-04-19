import { Clock, Link2, Unlink, RefreshCw, AlertCircle, Check, Shield } from 'lucide-react';
import { ScrollArea } from '@/components/ui/scroll-area';

export interface AuditLogEntry {
  id: string;
  timestamp: string;
  action: 'connected' | 'disconnected' | 'tested' | 'rotated' | 'failed';
  providerName: string;
  providerId: string;
  details?: string;
  actor?: string;
}

interface ConnectionAuditLogProps {
  entries: AuditLogEntry[];
  maxHeight?: number;
}

const ACTION_ICONS: Record<string, React.ReactNode> = {
  connected: <Link2 className="w-4 h-4 text-emerald-500" />,
  disconnected: <Unlink className="w-4 h-4 text-amber-500" />,
  tested: <Check className="w-4 h-4 text-blue-500" />,
  rotated: <RefreshCw className="w-4 h-4 text-purple-500" />,
  failed: <AlertCircle className="w-4 h-4 text-red-500" />,
};

const ACTION_COLORS: Record<string, string> = {
  connected: 'text-emerald-600 dark:text-emerald-400',
  disconnected: 'text-amber-600 dark:text-amber-400',
  tested: 'text-blue-600 dark:text-blue-400',
  rotated: 'text-purple-600 dark:text-purple-400',
  failed: 'text-red-600 dark:text-red-400',
};

const ACTION_LABELS: Record<string, string> = {
  connected: 'Connected',
  disconnected: 'Disconnected',
  tested: 'Connection Tested',
  rotated: 'API Key Rotated',
  failed: 'Connection Failed',
};

function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffInMs = now.getTime() - date.getTime();
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60));
  const diffInMinutes = Math.floor(diffInMs / (1000 * 60));

  if (diffInMinutes < 1) return 'Just now';
  if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
  if (diffInHours < 24) return `${diffInHours}h ago`;
  if (diffInDays < 7) return `${diffInDays}d ago`;
  return date.toLocaleDateString();
}

export function ConnectionAuditLog({ entries, maxHeight = 200 }: ConnectionAuditLogProps) {
  if (entries.length === 0) {
    return (
      <div className="p-4 text-center text-text-tertiary text-sm">
        <Shield className="w-8 h-8 mx-auto mb-2 opacity-50" />
        <p>No recent activity</p>
        <p className="text-xs opacity-70">Connection events will appear here</p>
      </div>
    );
  }

  return (
    <ScrollArea className={`h-${maxHeight}`} style={{ maxHeight }}>
      <div className="space-y-2 p-1">
        {entries.map((entry) => (
          <div
            key={entry.id}
            className="flex items-start gap-3 p-2 rounded-lg bg-bg-secondary/30 hover:bg-bg-secondary/50 transition-colors"
          >
            <div className="mt-0.5">{ACTION_ICONS[entry.action]}</div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className={`text-sm font-medium ${ACTION_COLORS[entry.action]}`}>
                  {ACTION_LABELS[entry.action]}
                </span>
                <span className="text-xs text-text-tertiary">{formatRelativeTime(entry.timestamp)}</span>
              </div>
              <p className="text-sm text-text-secondary truncate">{entry.providerName}</p>
              {entry.details && (
                <p className="text-xs text-text-tertiary mt-0.5">{entry.details}</p>
              )}
            </div>
          </div>
        ))}
      </div>
    </ScrollArea>
  );
}

// Mock data generator
export function generateMockAuditLog(): AuditLogEntry[] {
  const actions: Array<AuditLogEntry['action']> = ['connected', 'disconnected', 'tested', 'rotated', 'failed'];
  const providers = [
    { name: 'Cloudflare Workers', id: 'workers' },
    { name: 'Vercel', id: 'vercel' },
    { name: 'Fly.io', id: 'fly' },
    { name: 'Deno Deploy', id: 'deno' },
    { name: 'FunctionFly Edge', id: 'functionfly-edge' },
  ];

  const entries: AuditLogEntry[] = [];
  const now = Date.now();

  for (let i = 0; i < 10; i++) {
    const provider = providers[Math.floor(Math.random() * providers.length)];
    const action = actions[Math.floor(Math.random() * actions.length)];
    const timestamp = new Date(now - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString();

    entries.push({
      id: `audit-${i}`,
      timestamp,
      action,
      providerName: provider.name,
      providerId: provider.id,
      details: action === 'failed' ? 'Authentication failed - invalid API key' : undefined,
    });
  }

  return entries.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
}
