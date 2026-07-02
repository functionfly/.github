import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import {
  CheckCircle,
  Rocket,
  Shield,
  CreditCard,
  Mail,
  BarChart3,
  Store,
  MessageSquare,
  Bell,
  Brain,
  Cpu,
  Database,
  Sparkles,
  Settings,
  ExternalLink,
  ArrowRight,
  Zap,
  Clock,
  TrendingUp,
  Server,
  Code,
  Loader2,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { appsApi } from '@/api/apps';
import { usePageTitle } from '@/hooks';

// ─── Bundle metadata ─────────────────────────────────────────────────────────

const BUNDLES: Record<string, {
  name: string;
  tagline: string;
  price: string;
  gradient: string;
  icon: typeof Rocket;
  features: { icon: typeof Shield; title: string; desc: string }[];
  quickActions: { label: string; href: string; icon: typeof Settings }[];
}> = {
  'saas-starter': {
    name: 'SaaS Starter',
    tagline: 'Full SaaS backend ready to customize',
    price: '$29/mo',
    gradient: 'from-indigo-500 to-blue-500',
    icon: Rocket,
    features: [
      { icon: Shield, title: 'Authentication', desc: 'JWT, OAuth (Google, GitHub), sessions, MFA' },
      { icon: CreditCard, title: 'Payments', desc: 'Stripe integration, products, invoices, webhooks' },
      { icon: Mail, title: 'Email Workflows', desc: '20+ templates, welcome sequence, dunning' },
      { icon: BarChart3, title: 'Analytics', desc: 'Event tracking, funnels, cohorts, dashboards' },
    ],
    quickActions: [
      { label: 'View Functions', href: '/functions', icon: Code },
      { label: 'App Settings', href: '/settings', icon: Settings },
      { label: 'Billing', href: '/billing', icon: CreditCard },
    ],
  },
  marketplace: {
    name: 'Marketplace',
    tagline: 'Multi-vendor marketplace backend',
    price: '$49/mo',
    gradient: 'from-emerald-500 to-teal-500',
    icon: Store,
    features: [
      { icon: Store, title: 'Listings', desc: 'Categories, variants, reviews, full-text search' },
      { icon: CreditCard, title: 'Stripe Connect', desc: 'Split payments, seller payouts, refunds' },
      { icon: MessageSquare, title: 'Messaging', desc: 'Buyer-seller conversations, offers, files' },
      { icon: Bell, title: 'Notifications', desc: '22 templates, in-app + email alerts' },
    ],
    quickActions: [
      { label: 'View Functions', href: '/functions', icon: Code },
      { label: 'App Settings', href: '/settings', icon: Settings },
      { label: 'Billing', href: '/billing', icon: CreditCard },
    ],
  },
  'ai-app': {
    name: 'AI App',
    tagline: 'AI-powered backend with vector search',
    price: '$39/mo',
    gradient: 'from-violet-500 to-purple-500',
    icon: Brain,
    features: [
      { icon: Cpu, title: 'Vector DB', desc: 'pgvector with HNSW indexing, collections' },
      { icon: Database, title: 'Embeddings', desc: 'OpenAI embeddings, document chunking, RAG' },
      { icon: MessageSquare, title: 'Chat Workflows', desc: 'AI assistants, tool calling, guardrails' },
      { icon: Sparkles, title: 'Memory System', desc: 'Long-term semantic memory, user profiles' },
    ],
    quickActions: [
      { label: 'View Functions', href: '/functions', icon: Code },
      { label: 'App Settings', href: '/settings', icon: Settings },
      { label: 'Billing', href: '/billing', icon: CreditCard },
    ],
  },
};

// ─── Page Component ──────────────────────────────────────────────────────────

export default function BundleOverviewPage() {
  usePageTitle('Bundle Overview');
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const bundleSlug = searchParams.get('bundle') || 'saas-starter';
  const justDeployed = searchParams.get('deployed') === 'true';
  const [showConfetti, setShowConfetti] = useState(justDeployed);

  const bundle = BUNDLES[bundleSlug] || BUNDLES['saas-starter'];
  const Icon = bundle.icon;

  const { data: appsData, isLoading: appsLoading } = useQuery({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await appsApi.list();
      return res.apps;
    },
  });

  const apps = appsData ?? [];
  const latestApp = apps[0];

  useEffect(() => {
    if (justDeployed) {
      const timer = setTimeout(() => setShowConfetti(false), 5000);
      return () => clearTimeout(timer);
    }
  }, [justDeployed]);

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Success Banner */}
      {justDeployed && (
        <div className={`mb-8 overflow-hidden rounded-2xl bg-gradient-to-r ${bundle.gradient} p-[1px]`}>
          <div className="rounded-2xl bg-white p-6 dark:bg-zinc-900">
            <div className="flex items-start gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-500/20">
                <CheckCircle className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
              </div>
              <div className="flex-1">
                <h2 className="text-xl font-bold text-zinc-900 dark:text-white">
                  {bundle.name} deployed successfully!
                </h2>
                <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                  Your backend is live and ready. All functions, auth, and integrations have been configured.
                </p>
                <div className="mt-4 flex flex-wrap gap-3">
                  {latestApp && (
                    <Link
                      to={`/apps/${latestApp.id}`}
                      className="inline-flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-700 dark:bg-white dark:text-zinc-900 dark:hover:bg-zinc-200"
                    >
                      <ExternalLink className="h-4 w-4" />
                      Open App
                    </Link>
                  )}
                  <Link
                    to="/functions"
                    className="inline-flex items-center gap-2 rounded-lg border border-zinc-300 px-4 py-2 text-sm font-medium text-zinc-700 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-800"
                  >
                    <Code className="h-4 w-4" />
                    View Functions
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-4">
          <div className={`flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-r ${bundle.gradient} text-white shadow-lg`}>
            <Icon className="h-7 w-7" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-zinc-900 dark:text-white">{bundle.name}</h1>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">{bundle.tagline}</p>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-3 py-1.5 text-sm font-medium text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400">
              <Zap className="h-3.5 w-3.5" />
              Active
            </span>
            <span className="text-sm font-medium text-zinc-500 dark:text-zinc-400">{bundle.price}</span>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="mb-8 grid grid-cols-1 gap-3 sm:grid-cols-3">
        {bundle.quickActions.map((action) => (
          <Link
            key={action.label}
            to={action.href}
            className="flex items-center gap-3 rounded-xl border border-zinc-200 bg-white p-4 transition-all hover:border-zinc-300 hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900 dark:hover:border-zinc-700"
          >
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800">
              <action.icon className="h-5 w-5 text-zinc-600 dark:text-zinc-400" />
            </div>
            <div className="flex-1">
              <div className="text-sm font-medium text-zinc-900 dark:text-white">{action.label}</div>
            </div>
            <ArrowRight className="h-4 w-4 text-zinc-400" />
          </Link>
        ))}
      </div>

      {/* Features Grid */}
      <div className="mb-8 rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900">
        <h2 className="mb-4 text-lg font-semibold text-zinc-900 dark:text-white">What's Included</h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {bundle.features.map((f) => (
            <div key={f.title} className="flex items-start gap-3 rounded-lg bg-zinc-50 p-4 dark:bg-zinc-800/50">
              <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-white shadow-sm dark:bg-zinc-800">
                <f.icon className="h-4.5 w-4.5 text-zinc-600 dark:text-zinc-400" />
              </div>
              <div>
                <div className="text-sm font-medium text-zinc-900 dark:text-white">{f.title}</div>
                <div className="text-xs text-zinc-500 dark:text-zinc-400">{f.desc}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Apps created by this bundle */}
      {appsLoading ? (
        <div className="flex items-center justify-center gap-2 py-8 text-sm text-zinc-500">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading your apps...
        </div>
      ) : apps.length > 0 ? (
        <div className="rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900">
          <h2 className="mb-4 text-lg font-semibold text-zinc-900 dark:text-white">Your Apps</h2>
          <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
            {apps.map((app) => (
              <Link
                key={app.id}
                to={`/apps/${app.id}`}
                className="flex items-center gap-4 py-3 transition-colors hover:bg-zinc-50 -mx-2 px-2 rounded-lg dark:hover:bg-zinc-800/50"
              >
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800">
                  <Server className="h-5 w-5 text-zinc-500" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium text-zinc-900 dark:text-white truncate">{app.name}</div>
                  <div className="text-xs text-zinc-500 dark:text-zinc-400">{app.id}</div>
                </div>
                <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
                  app.status === 'active'
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400'
                    : 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'
                }`}>
                  {app.status}
                </span>
                <ArrowRight className="h-4 w-4 text-zinc-400" />
              </Link>
            ))}
          </div>
        </div>
      ) : null}

      {/* Getting Started */}
      <div className="mt-8 rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-900">
        <h2 className="mb-4 text-lg font-semibold text-zinc-900 dark:text-white">Next Steps</h2>
        <div className="space-y-3">
          <NextStep
            num={1}
            title="Explore your functions"
            desc="Your bundle came with pre-configured function templates. Customize them for your use case."
            href="/functions"
          />
          <NextStep
            num={2}
            title="Configure integrations"
            desc="Set up your Stripe keys, OAuth providers, and email templates."
            href="/settings"
          />
          <NextStep
            num={3}
            title="Deploy to production"
            desc="When you're ready, deploy your functions to the edge."
            href="/functions"
          />
        </div>
      </div>
    </div>
  );
}

// ─── Sub-components ──────────────────────────────────────────────────────────

function NextStep({ num, title, desc, href }: { num: number; title: string; desc: string; href: string }) {
  return (
    <Link
      to={href}
      className="flex items-start gap-4 rounded-lg p-3 transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
    >
      <div className="flex h-8 w-8 items-center justify-center rounded-full bg-zinc-100 text-sm font-bold text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400">
        {num}
      </div>
      <div className="flex-1">
        <div className="text-sm font-medium text-zinc-900 dark:text-white">{title}</div>
        <div className="text-xs text-zinc-500 dark:text-zinc-400">{desc}</div>
      </div>
      <ArrowRight className="mt-1 h-4 w-4 text-zinc-400" />
    </Link>
  );
}
