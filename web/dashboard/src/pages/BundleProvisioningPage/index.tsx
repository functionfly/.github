import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { CheckCircle, Loader2, ArrowRight, ExternalLink, Settings, BarChart3, Shield, CreditCard } from 'lucide-react';
import { ProvisioningStatusCard } from '@/components/provisioning/ProvisioningStatusCard';
import { useProvisioningStatus, useProvisionBundle, useIsProvisioned } from '@/hooks/useProvisioning';

// ─── Bundle metadata ─────────────────────────────────────────────────────────

const BUNDLE_META: Record<string, { name: string; tagline: string; color: string }> = {
  'saas-starter': {
    name: 'SaaS Starter',
    tagline: 'Auth, Payments, Email Workflows, Analytics',
    color: 'indigo',
  },
  marketplace: {
    name: 'Marketplace',
    tagline: 'Listings, Stripe Connect, Messaging, Notifications',
    color: 'emerald',
  },
  'ai-app': {
    name: 'AI App',
    tagline: 'Vector DB, Embeddings, Chat Workflows, Memory',
    color: 'violet',
  },
};

// ─── Page Component ──────────────────────────────────────────────────────────

export default function BundleProvisioningPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const bundleSlug = searchParams.get('bundle') || 'saas-starter';
  const fromCheckout = searchParams.get('success') === 'true';

  const { data: statusData, isLoading } = useProvisioningStatus();
  const provisionMutation = useProvisionBundle();
  const isProvisioned = useIsProvisioned();

  const [autoProvisioned, setAutoProvisioned] = useState(false);

  const meta = BUNDLE_META[bundleSlug] || BUNDLE_META['saas-starter'];

  // Auto-trigger provisioning if coming from checkout
  useEffect(() => {
    if (fromCheckout && !autoProvisioned && !isLoading && !isProvisioned && !provisionMutation.isPending) {
      provisionMutation.mutate(bundleSlug);
      setAutoProvisioned(true);
    }
  }, [fromCheckout, autoProvisioned, isLoading, isProvisioned, bundleSlug]);

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-3 mb-2">
          <div className={`h-10 w-10 rounded-lg bg-${meta.color}-100 dark:bg-${meta.color}-500/10 flex items-center justify-center`}>
            <Shield className={`h-5 w-5 text-${meta.color}-600 dark:text-${meta.color}-400`} />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-zinc-900 dark:text-white">
              {meta.name}
            </h1>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              {meta.tagline}
            </p>
          </div>
        </div>

        {fromCheckout && (
          <div className="mt-4 rounded-lg bg-emerald-50 border border-emerald-200 px-4 py-3 dark:bg-emerald-500/10 dark:border-emerald-500/20">
            <div className="flex items-center gap-2">
              <CheckCircle className="h-5 w-5 text-emerald-600 dark:text-emerald-400" />
              <span className="text-sm font-medium text-emerald-800 dark:text-emerald-300">
                Payment confirmed! Your {meta.name} bundle is being provisioned.
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Provisioning Status Card */}
      <ProvisioningStatusCard
        bundleSlug={bundleSlug}
        onComplete={() => {
          // Auto-navigate to app after 3 seconds
          setTimeout(() => navigate('/dashboard'), 3000);
        }}
      />

      {/* Quick Links (shown after provisioning) */}
      {isProvisioned && (
        <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
          <QuickLink
            icon={Settings}
            title="Configure"
            description="Customize your bundle settings"
            href={`/settings?tab=billing`}
          />
          <QuickLink
            icon={BarChart3}
            title="Analytics"
            description="View usage and performance"
            href="/analytics"
          />
          <QuickLink
            icon={CreditCard}
            title="Billing"
            description="Manage subscription and invoices"
            href="/settings?tab=billing"
          />
        </div>
      )}

      {/* What's Included */}
      <div className="mt-12 rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900">
        <h2 className="text-lg font-semibold text-zinc-900 dark:text-white mb-4">
          What's Included
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {bundleSlug === 'saas-starter' && (
            <>
              <IncludedItem icon={Shield} title="Authentication" desc="JWT keys, OAuth (Google, GitHub), sessions, MFA, email verification" />
              <IncludedItem icon={CreditCard} title="Payments" desc="Stripe integration, 4 products, 7 prices, invoices, webhooks" />
              <IncludedItem icon={ExternalLink} title="Email Workflows" desc="20+ templates, welcome sequence, dunning, trial conversion" />
              <IncludedItem icon={BarChart3} title="Analytics" desc="Event tracking, funnels, cohorts, daily dashboards" />
            </>
          )}
          {bundleSlug === 'marketplace' && (
            <>
              <IncludedItem icon={ExternalLink} title="Listings" desc="Categories, variants, reviews, favorites, full-text search" />
              <IncludedItem icon={CreditCard} title="Stripe Connect" desc="Split payments, seller payouts, refunds, disputes" />
              <IncludedItem icon={Shield} title="Messaging" desc="Buyer-seller conversations, offers, file attachments" />
              <IncludedItem icon={BarChart3} title="Notifications" desc="22 templates, in-app + email, order lifecycle alerts" />
            </>
          )}
          {bundleSlug === 'ai-app' && (
            <>
              <IncludedItem icon={Shield} title="Vector DB" desc="pgvector with HNSW indexing, embedding collections, async queue" />
              <IncludedItem icon={CreditCard} title="Embeddings" desc="OpenAI text-embedding-3-small, document chunking, RAG pipeline" />
              <IncludedItem icon={ExternalLink} title="Chat Workflows" desc="AI assistants, tool calling, conversation memory, guardrails" />
              <IncludedItem icon={BarChart3} title="Memory System" desc="Long-term semantic memory, user profiles, memory graph" />
            </>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Sub-components ──────────────────────────────────────────────────────────

function QuickLink({ icon: Icon, title, description, href }: {
  icon: React.ComponentType<React.SVGAttributes<SVGElement>>;
  title: string;
  description: string;
  href: string;
}) {
  return (
    <a
      href={href}
      className="flex items-center gap-3 rounded-lg border border-zinc-200 bg-white p-4 transition-colors hover:border-indigo-300 hover:bg-indigo-50 dark:border-zinc-800 dark:bg-zinc-900 dark:hover:border-indigo-500/30 dark:hover:bg-indigo-500/5"
    >
      <Icon className="h-5 w-5 text-[#a3a3a3] dark:text-[#a1a1aa]" />
      <div>
        <div className="text-sm font-medium text-zinc-900 dark:text-white">{title}</div>
        <div className="text-xs text-zinc-500 dark:text-zinc-400">{description}</div>
      </div>
      <ArrowRight className="ml-auto h-4 w-4 text-zinc-400" />
    </a>
  );
}

function IncludedItem({ icon: Icon, title, desc }: {
  icon: React.ComponentType<React.SVGAttributes<SVGElement>>;
  title: string;
  desc: string;
}) {
  return (
    <div className="flex items-start gap-3">
      <div className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800">
        <Icon className="h-4 w-4 text-[#52525b] dark:text-[#71717a]" />
      </div>
      <div>
        <div className="text-sm font-medium text-zinc-900 dark:text-white">{title}</div>
        <div className="text-xs text-zinc-500 dark:text-zinc-400">{desc}</div>
      </div>
    </div>
  );
}
