import { agentApi, type AgentIdentity } from '@/api/agent';
import { notificationsApi } from '@/api/notifications';
import { WalletDashboard } from '@/components/swarm/WalletDashboard';
import { Button } from '@/components/ui/button';
import { ROUTES } from '@/lib/constants';
import { useNotificationStore } from '@/stores/notificationStore';
import { Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

const LAST_WALLET_AGENT_KEY = 'ff-last-wallet-agent-id';

function persistLastWalletAgent(id: string) {
  try {
    localStorage.setItem(LAST_WALLET_AGENT_KEY, id);
  } catch {
    /* ignore */
  }
}

/** Route param was missing `agentId` before API normalization; URL became literally `/wallet/undefined`. */
function sanitizeWalletAgentIdParam(raw: string | null | undefined): string | null {
  const t = raw?.trim();
  if (!t || t === 'undefined' || t === 'null') return null;
  return t;
}

/**
 * Wallet is scoped per agent. Routes: /wallet (resolve agent) and /wallet/:agentId.
 * The sidebar uses /wallet; we resolve a real agent id instead of the invalid placeholder "default-agent".
 */
export function WalletPage() {
  const { slug: pathAgentId } = useParams<{ slug: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const queryAgent = searchParams.get('agent')?.trim() || undefined;
  const updateUnreadCounts = useNotificationStore((s) => s.updateUnreadCounts);

  const explicitId =
    sanitizeWalletAgentIdParam(pathAgentId) ?? sanitizeWalletAgentIdParam(queryAgent) ?? null;

  const [phase, setPhase] = useState<'idle' | 'resolving' | 'pick' | 'error'>('idle');
  const [agents, setAgents] = useState<AgentIdentity[]>([]);
  const [resolveError, setResolveError] = useState<string | null>(null);

  // Handle return from Stripe checkout
  useEffect(() => {
    const credits = searchParams.get('credits');
    if (!credits) return;

    // Remove the param from the URL immediately so a refresh doesn't re-trigger
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        next.delete('credits');
        return next;
      },
      { replace: true }
    );

    if (credits === 'success') {
      toast.success('Funds added successfully', {
        description: 'Your wallet has been topped up. Check your notifications for details.',
      });
      // Refresh the bell badge so the new notification is reflected right away
      notificationsApi
        .fetchUnreadCounts()
        .then((counts) => {
          const byCategory = counts?.byCategory || {};
          updateUnreadCounts({
            all: counts?.total || 0,
            trust: byCategory.trust || 0,
            revenue: byCategory.revenue || 0,
            issues: byCategory.issues || 0,
            messages: byCategory.messages || 0,
            security: byCategory.security || 0,
          });
        })
        .catch(() => {
          /* silent – badge will catch up on next poll */
        });
    } else if (credits === 'cancel') {
      toast.info('Payment cancelled', {
        description: 'No funds were added to your wallet.',
      });
    }
  }, [searchParams, setSearchParams, updateUnreadCounts]);

  useEffect(() => {
    if (explicitId) {
      persistLastWalletAgent(explicitId);
      setPhase('idle');
      return;
    }

    let cancelled = false;
    setPhase('resolving');
    setResolveError(null);

    (async () => {
      try {
        const res = await agentApi.listAgents({ limit: 100 });
        if (cancelled) return;
        const list = res.agents ?? [];
        if (list.length === 0) {
          setPhase('error');
          setResolveError('No agents yet. Create an agent to use the wallet.');
          return;
        }
        if (list.length === 1) {
          const id = list[0].agentId;
          persistLastWalletAgent(id);
          navigate(`/wallet/agents/${encodeURIComponent(id)}`, { replace: true });
          return;
        }
        let last: string | null = null;
        try {
          last = localStorage.getItem(LAST_WALLET_AGENT_KEY);
        } catch {
          /* ignore */
        }
        if (last && list.some((a) => a.agentId === last)) {
          navigate(`/wallet/agents/${encodeURIComponent(last)}`, { replace: true });
          return;
        }
        setAgents(list);
        setPhase('pick');
      } catch {
        if (!cancelled) {
          setPhase('error');
          setResolveError('Could not load your agents. Try again.');
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [explicitId, navigate]);

  if (explicitId) {
    return <WalletDashboard agentId={explicitId} />;
  }

  if (phase === 'resolving' || phase === 'idle') {
    return (
      <div className="flex flex-col items-center justify-center gap-3 p-12 text-muted-foreground">
        <Loader2 className="h-8 w-8 animate-spin" />
        <p className="text-sm">Loading wallet…</p>
      </div>
    );
  }

  if (phase === 'error') {
    return (
      <div className="mx-auto max-w-md space-y-4 rounded-lg border border-border p-6">
        <h1 className="text-xl font-semibold">Agent wallet</h1>
        <p className="text-sm text-muted-foreground">{resolveError}</p>
        <Button asChild>
          <Link to={ROUTES.AGENTS}>Go to agents</Link>
        </Button>
      </div>
    );
  }

  if (phase === 'pick') {
    return (
      <div className="mx-auto max-w-lg space-y-4 p-6">
        <h1 className="text-2xl font-bold">Choose an agent</h1>
        <p className="text-sm text-muted-foreground">
          The wallet is tied to a specific agent. Select one to continue.
        </p>
        <ul className="space-y-2">
          {agents.map((a) => (
            <li key={a.agentId}>
              <Button variant="outline" className="h-auto w-full justify-start py-3" asChild>
                <Link
                  to={`/wallet/agents/${encodeURIComponent(a.agentId)}`}
                  onClick={() => persistLastWalletAgent(a.agentId)}
                >
                  <span className="font-medium">{a.name || a.agentId}</span>
                  <span className="ml-2 text-xs text-muted-foreground">{a.agentId}</span>
                </Link>
              </Button>
            </li>
          ))}
        </ul>
      </div>
    );
  }

  return null;
}

export default WalletPage;
