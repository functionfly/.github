import { useState, useEffect, useCallback } from 'react';
import { CheckCircle, Circle, Loader2, AlertCircle, ArrowRight, Shield, CreditCard, Mail, BarChart3, Database, Store, MessageSquare, Bell, Brain, Cpu, Sparkles, MessageCircle } from 'lucide-react';
import { apiClient } from '@/api/client';

// ─── Types ───────────────────────────────────────────────────────────────────

interface ComponentState {
  status: 'pending' | 'provisioning' | 'active' | 'failed' | 'rolled_back';
  timestamp: string;
  error?: string;
  resource_id?: string;
}

interface ProvisioningResult {
  tenant_id: string;
  bundle_slug: string;
  status: 'pending' | 'provisioning' | 'active' | 'failed';
  components: Record<string, ComponentState>;
  started_at: string;
  finished_at: string;
  duration_ms: number;
  error_log?: string[];
}

interface ProvisioningStatusCardProps {
  bundleSlug: string;
  onComplete?: () => void;
}

// ─── Component definitions ───────────────────────────────────────────────────

const COMPONENT_META: Record<string, { icon: React.ComponentType<React.SVGAttributes<SVGElement>>; label: string; description: string }> = {
  user_db: {
    icon: Database,
    label: 'User Database',
    description: 'Dedicated PostgreSQL database with isolated schema',
  },
  auth: {
    icon: Shield,
    label: 'Authentication',
    description: 'JWT keys, OAuth configs, sessions, MFA',
  },
  payments: {
    icon: CreditCard,
    label: 'Payments',
    description: 'Stripe integration, products, invoices, webhooks',
  },
  email_workflows: {
    icon: Mail,
    label: 'Email Workflows',
    description: 'Templates, automated sequences, dunning flows',
  },
  analytics: {
    icon: BarChart3,
    label: 'Analytics',
    description: 'Event tracking, funnels, cohorts, dashboards',
  },
  // Marketplace-specific components
  marketplace: {
    icon: Store,
    label: 'Marketplace',
    description: 'Listings, categories, reviews, seller accounts',
  },
  messaging: {
    icon: MessageSquare,
    label: 'Messaging',
    description: 'Buyer-seller conversations, offers, attachments',
  },
  notifications: {
    icon: Bell,
    label: 'Notifications',
    description: 'In-app, email, push notification templates and delivery',
  },
  // AI App-specific components
  ai_app: {
    icon: Brain,
    label: 'AI Infrastructure',
    description: 'Vector DB, embeddings, RAG collections, memory system',
  },
  vector_db: {
    icon: Cpu,
    label: 'Vector Database',
    description: 'pgvector collections with HNSW indexing for fast similarity search',
  },
  chat_workflows: {
    icon: MessageCircle,
    label: 'Chat Workflows',
    description: 'AI assistants, conversations, RAG pipelines, tool calling',
  },
  memory_system: {
    icon: Sparkles,
    label: 'Memory System',
    description: 'Long-term memory, user profiles, semantic recall',
  },
};

// Bundle-specific component ordering
const BUNDLE_COMPONENTS: Record<string, string[]> = {
  'saas-starter': ['user_db', 'auth', 'payments', 'email_workflows', 'analytics'],
  'marketplace': ['user_db', 'auth', 'payments', 'email_workflows', 'analytics', 'marketplace'],
  'ai-app': ['user_db', 'auth', 'payments', 'email_workflows', 'analytics', 'ai_app'],
};

// ─── Status icons ────────────────────────────────────────────────────────────

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case 'active':
      return <CheckCircle className="w-5 h-5 text-emerald-500" />;
    case 'provisioning':
      return <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />;
    case 'failed':
      return <AlertCircle className="w-5 h-5 text-red-500" />;
    case 'pending':
      return <Circle className="w-5 h-5 text-zinc-400" />;
    default:
      return <Circle className="w-5 h-5 text-zinc-400" />;
  }
}

// ─── Main component ──────────────────────────────────────────────────────────

export function ProvisioningStatusCard({ bundleSlug, onComplete }: ProvisioningStatusCardProps) {
  const [result, setResult] = useState<ProvisioningResult | null>(null);
  const [isProvisioning, setIsProvisioning] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = useCallback(async () => {
    try {
      const res = await apiClient.get<ProvisioningResult | { status: 'not_provisioned' }>('/provisioning/status');
      if (res && 'status' in res && res.status === 'not_provisioned') {
        setResult(null);
      } else {
        setResult(res as ProvisioningResult);
      }
    } catch {
      // Silently fail — status endpoint is optional
    }
  }, []);

  // Poll for status updates while provisioning
  useEffect(() => {
    if (!isProvisioning) return;
    const interval = setInterval(fetchStatus, 2000);
    return () => clearInterval(interval);
  }, [isProvisioning, fetchStatus]);

  // Check for existing provisioning on mount
  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  const handleProvision = async () => {
    setIsProvisioning(true);
    setError(null);

    try {
      const res = await apiClient.post<ProvisioningResult>('/provisioning/bundle', {
        bundle_slug: bundleSlug,
      });

      setResult(res);

      if (res.status === 'active') {
        onComplete?.();
      }
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Provisioning failed');
    } finally {
      setIsProvisioning(false);
    }
  };

  const handleRetry = async () => {
    setIsProvisioning(true);
    setError(null);

    try {
      await apiClient.post('/provisioning/retry');
      // Poll will pick up the new state
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Retry failed');
      setIsProvisioning(false);
    }
  };

  const allActive = result?.status === 'active';
  const hasFailures = result?.status === 'failed' || Object.values(result?.components || {}).some(c => c.status === 'failed');
  const componentOrder = BUNDLE_COMPONENTS[bundleSlug] || BUNDLE_COMPONENTS['saas-starter'];

  const bundleNames: Record<string, string> = {
    'saas-starter': 'SaaS Starter',
    'marketplace': 'Marketplace',
    'ai-app': 'AI App',
  };

  return (
    <div className="rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
      {/* Header */}
      <div className="border-b border-zinc-200 px-6 py-4 dark:border-zinc-800">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-zinc-900 dark:text-white">
              {bundleNames[bundleSlug] || bundleSlug} — One-Click Provisioning
            </h3>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              {bundleSlug === 'ai-app'
                ? 'AI infrastructure — Vector DB, embeddings, chat, and memory'
                : 'Fully isolated production stack — zero platform leakage'}
            </p>
          </div>
          {allActive && (
            <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400">
              <CheckCircle className="h-4 w-4" />
              All Systems Active
            </span>
          )}
        </div>
      </div>

      {/* Component list */}
      <div className="divide-y divide-zinc-100 px-6 dark:divide-zinc-800">
        {componentOrder.map((key) => {
          const meta = COMPONENT_META[key];
          const state = result?.components?.[key];
          const Icon = meta.icon;

          return (
            <div key={key} className="flex items-center gap-4 py-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800">
                <Icon className="h-5 w-5 text-[#71717a] dark:text-[#71717a]" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-zinc-900 dark:text-white">
                    {meta.label}
                  </span>
                  {state && <StatusIcon status={state.status} />}
                </div>
                <p className="text-xs text-zinc-500 dark:text-zinc-400">{meta.description}</p>
                {state?.error && (
                  <p className="mt-1 text-xs text-red-500">{state.error}</p>
                )}
              </div>
              {state?.resource_id && (
                <span className="text-xs text-zinc-400 font-mono">{state.resource_id}</span>
              )}
            </div>
          );
        })}
      </div>

      {/* Error message */}
      {error && (
        <div className="mx-6 mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-700 dark:bg-red-500/10 dark:text-red-400">
          {error}
        </div>
      )}

      {/* Actions */}
      <div className="border-t border-zinc-200 px-6 py-4 dark:border-zinc-800">
        {!result && !isProvisioning && (
          <button
            onClick={handleProvision}
            className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
          >
            Provision {bundleNames[bundleSlug] || 'Everything'}
            <ArrowRight className="h-4 w-4" />
          </button>
        )}

        {isProvisioning && (
          <div className="flex items-center justify-center gap-2 text-sm text-blue-600 dark:text-blue-400">
            <Loader2 className="h-4 w-4 animate-spin" />
            Provisioning isolated infrastructure...
          </div>
        )}

        {allActive && (
          <div className="text-center text-sm text-zinc-500 dark:text-zinc-400">
            Provisioned in {((result?.duration_ms || 0) / 1000).toFixed(1)}s
          </div>
        )}

        {hasFailures && !isProvisioning && (
          <button
            onClick={handleRetry}
            className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-amber-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-amber-500"
          >
            Retry Failed Components
          </button>
        )}
      </div>
    </div>
  );
}
