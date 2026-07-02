import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Shield, CreditCard, Mail, BarChart3, CheckCircle, XCircle,
  Loader2, ExternalLink, ChevronRight, Settings, ToggleLeft, ToggleRight,
} from 'lucide-react';
import { apiClient } from '@/api/client';
import { usePageTitle } from '@/hooks';
import {
  Chamber, CornerBrace, PageGrid, SealedButton, FrameButton, StatusPill,
} from '@/components/containment';

// ─── Types ───────────────────────────────────────────────────────────────────

interface OAuthProvider {
  provider: string;
  enabled: boolean;
  client_id: string;
  has_secret: boolean;
}

interface BundleConfig {
  bundle_slug: string;
  tenant_id: string;
  auth: {
    oauth_providers: OAuthProvider[];
    has_jwt_key: boolean;
    jwt_key_id: string;
    email_templates: number;
  };
  payments: {
    mode: string;
    default_currency: string;
    tax_calculation_mode: string;
    webhook_url: string;
    products: { name: string; active: boolean; prices: { amount_cents: number; currency: string; interval: string; trial_days: number }[] }[];
    email_templates: number;
  };
  email: {
    transaction_templates: number;
    workflow_templates: number;
    workflows: { slug: string; name: string; trigger_type: string; is_active: boolean; steps: number }[];
  };
  analytics: {
    dashboards: number;
    funnels: number;
    events: number;
    retention_days: number;
  };
}

// ─── API ─────────────────────────────────────────────────────────────────────

function useBundleConfig(appId: string) {
  return useQuery<BundleConfig>({
    queryKey: ['bundle-config', appId],
    queryFn: () => apiClient.get<BundleConfig>(`/v1/apps/${appId}/bundle/config`),
    enabled: !!appId,
  });
}

function useToggleOAuth(appId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ provider, enabled }: { provider: string; enabled: boolean }) =>
      apiClient.put(`/v1/apps/${appId}/bundle/config/auth/oauth/${provider}`, { enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bundle-config', appId] }),
  });
}

function useToggleWorkflow(appId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ slug, active }: { slug: string; active: boolean }) =>
      apiClient.put(`/v1/apps/${appId}/bundle/config/email/workflows/${slug}`, { is_active: active }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bundle-config', appId] }),
  });
}

// ─── Page ────────────────────────────────────────────────────────────────────

export default function BundleConfigPage() {
  usePageTitle('Bundle Configuration');
  const { slug } = useParams<{ slug: string }>();
  const appId = slug || '';
  const { data, isLoading, error } = useBundleConfig(appId);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-32">
        <Loader2 className="w-6 h-6 animate-spin" style={{ color: 'var(--accent)' }} />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="flex flex-col items-center justify-center py-32 gap-4">
        <XCircle className="w-8 h-8" style={{ color: 'var(--status-revoked)' }} />
        <p style={{ color: 'var(--text-dim)' }}>Failed to load bundle configuration</p>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '72rem', margin: '0 auto', padding: 'var(--space-6)', display: 'flex', flexDirection: 'column', gap: 'var(--space-5)' }}>
      <PageGrid />

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingBottom: 'var(--space-5)', borderBottom: '1px solid var(--chamber-edge)' }}>
        <div>
          <h1 style={{ fontFamily: 'var(--font-display)', fontSize: '1.75rem', fontWeight: 700, color: 'var(--text)', display: 'flex', alignItems: 'center', gap: '0.625rem', margin: 0 }}>
            <span style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: '2.25rem', height: '2.25rem', borderRadius: 'var(--radius)', background: 'linear-gradient(135deg, var(--foil-a), var(--foil-b), var(--foil-c))' }}>
              <Settings className="h-5 w-5" style={{ color: 'var(--text-on-light)' }} />
            </span>
            Bundle Configuration
          </h1>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-dim)', marginTop: 'var(--space-1)' }}>
            Manage your {data.bundle_slug} infrastructure components
          </p>
        </div>
        <Link to={`/apps/${appId}`}>
          <FrameButton size="sm" iconRight={<ChevronRight className="w-3.5 h-3.5" />}>
            Back to App
          </FrameButton>
        </Link>
      </div>

      {/* Component Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(22rem, 1fr))', gap: 'var(--space-4)' }}>
        <AuthCard config={data.auth} appId={appId || ''} />
        <PaymentsCard config={data.payments} />
        <EmailCard config={data.email} appId={appId || ''} />
        <AnalyticsCard config={data.analytics} />
      </div>
    </div>
  );
}

// ─── Auth Card ───────────────────────────────────────────────────────────────

function AuthCard({ config, appId }: { config: BundleConfig['auth']; appId: string }) {
  const toggleOAuth = useToggleOAuth(appId);

  return (
    <Chamber style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)', padding: 'var(--space-5)' }}>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
        <div style={{ width: '2.5rem', height: '2.5rem', borderRadius: 'var(--radius)', background: 'rgba(143,255,208,0.08)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Shield className="h-5 w-5" style={{ color: 'var(--status-ok)' }} />
        </div>
        <div style={{ flex: 1 }}>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '1rem', fontWeight: 700, color: 'var(--text)', margin: 0 }}>Authentication</h3>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-dim)' }}>JWT, OAuth, MFA</span>
        </div>
        <StatusPill status="live" label="Active" />
      </div>

      <div style={{ padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel-raised)', borderLeft: '2px solid var(--steel)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', fontSize: '0.8125rem', color: 'var(--text-dim)' }}>
          <CheckCircle className="h-4 w-4" style={{ color: config.has_jwt_key ? 'var(--status-ok)' : 'var(--status-revoked)' }} />
          <span>JWT Key: <strong style={{ color: 'var(--text)' }}>{config.jwt_key_id || 'Not configured'}</strong></span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', fontSize: '0.8125rem', color: 'var(--text-dim)', marginTop: '4px' }}>
          <Mail className="h-4 w-4" style={{ color: 'var(--text-faint)' }} />
          <span>{config.email_templates} auth email templates</span>
        </div>
      </div>

      <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-faint)', textTransform: 'uppercase', letterSpacing: '0.06em', marginTop: 'var(--space-2)' }}>
        OAuth Providers
      </div>
      {config.oauth_providers.map((p) => (
        <div key={p.provider} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel-raised)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <span style={{ fontSize: '0.8125rem', fontWeight: 600, color: 'var(--text)', textTransform: 'capitalize' }}>{p.provider}</span>
            {p.enabled ? (
              <span style={{ fontSize: '0.6875rem', color: 'var(--status-ok)' }}>Enabled</span>
            ) : (
              <span style={{ fontSize: '0.6875rem', color: 'var(--text-faint)' }}>Disabled</span>
            )}
          </div>
          <button
            onClick={() => toggleOAuth.mutate({ provider: p.provider, enabled: !p.enabled })}
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: p.enabled ? 'var(--status-ok)' : 'var(--text-faint)', padding: 0 }}
            aria-label={`Toggle ${p.provider}`}
          >
            {p.enabled ? <ToggleRight className="w-6 h-6" /> : <ToggleLeft className="w-6 h-6" />}
          </button>
        </div>
      ))}
    </Chamber>
  );
}

// ─── Payments Card ───────────────────────────────────────────────────────────

function PaymentsCard({ config }: { config: BundleConfig['payments'] }) {
  return (
    <Chamber style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)', padding: 'var(--space-5)' }}>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
        <div style={{ width: '2.5rem', height: '2.5rem', borderRadius: 'var(--radius)', background: 'rgba(143,255,208,0.08)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <CreditCard className="h-5 w-5" style={{ color: 'var(--status-ok)' }} />
        </div>
        <div style={{ flex: 1 }}>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '1rem', fontWeight: 700, color: 'var(--text)', margin: 0 }}>Payments</h3>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-dim)' }}>Stripe · {config.default_currency.toUpperCase()}</span>
        </div>
        <StatusPill status="live" label={config.mode} />
      </div>

      <div style={{ padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel-raised)', borderLeft: '2px solid var(--steel)' }}>
        <div style={{ fontSize: '0.8125rem', color: 'var(--text-dim)' }}>
          Tax: <strong style={{ color: 'var(--text)' }}>{config.tax_calculation_mode}</strong>
        </div>
        <div style={{ fontSize: '0.8125rem', color: 'var(--text-dim)', marginTop: '4px' }}>
          {config.products.length} products · {config.email_templates} billing templates
        </div>
      </div>

      <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-faint)', textTransform: 'uppercase', letterSpacing: '0.06em', marginTop: 'var(--space-2)' }}>
        Products
      </div>
      {config.products.map((p) => (
        <div key={p.name} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel-raised)' }}>
          <span style={{ fontSize: '0.8125rem', fontWeight: 600, color: 'var(--text)' }}>{p.name}</span>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-dim)' }}>
            {p.prices.length > 0 ? `$${(p.prices[0].amount_cents / 100).toFixed(0)}/mo` : 'Free'}
          </span>
        </div>
      ))}

      {config.webhook_url && (
        <div style={{ fontSize: '0.6875rem', color: 'var(--text-faint)', wordBreak: 'break-all', marginTop: 'var(--space-2)' }}>
          Webhook: {config.webhook_url}
        </div>
      )}
    </Chamber>
  );
}

// ─── Email Card ──────────────────────────────────────────────────────────────

function EmailCard({ config, appId }: { config: BundleConfig['email']; appId: string }) {
  const toggleWorkflow = useToggleWorkflow(appId);

  return (
    <Chamber style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)', padding: 'var(--space-5)' }}>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
        <div style={{ width: '2.5rem', height: '2.5rem', borderRadius: 'var(--radius)', background: 'rgba(143,255,208,0.08)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Mail className="h-5 w-5" style={{ color: 'var(--status-ok)' }} />
        </div>
        <div style={{ flex: 1 }}>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '1rem', fontWeight: 700, color: 'var(--text)', margin: 0 }}>Email Workflows</h3>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-dim)' }}>{config.transaction_templates + config.workflow_templates} templates</span>
        </div>
        <StatusPill status="live" label="Active" />
      </div>

      <div style={{ padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel-raised)', borderLeft: '2px solid var(--steel)' }}>
        <div style={{ fontSize: '0.8125rem', color: 'var(--text-dim)' }}>
          {config.transaction_templates} transactional · {config.workflow_templates} workflow templates
        </div>
      </div>

      <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-faint)', textTransform: 'uppercase', letterSpacing: '0.06em', marginTop: 'var(--space-2)' }}>
        Workflows
      </div>
      {config.workflows.map((wf) => (
        <div key={wf.slug} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel-raised)' }}>
          <div>
            <div style={{ fontSize: '0.8125rem', fontWeight: 600, color: 'var(--text)' }}>{wf.name}</div>
            <div style={{ fontSize: '0.6875rem', color: 'var(--text-faint)' }}>{wf.steps} steps · {wf.trigger_type}</div>
          </div>
          <button
            onClick={() => toggleWorkflow.mutate({ slug: wf.slug, active: !wf.is_active })}
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: wf.is_active ? 'var(--status-ok)' : 'var(--text-faint)', padding: 0 }}
            aria-label={`Toggle ${wf.name}`}
          >
            {wf.is_active ? <ToggleRight className="w-6 h-6" /> : <ToggleLeft className="w-6 h-6" />}
          </button>
        </div>
      ))}
    </Chamber>
  );
}

// ─── Analytics Card ──────────────────────────────────────────────────────────

function AnalyticsCard({ config }: { config: BundleConfig['analytics'] }) {
  return (
    <Chamber style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)', padding: 'var(--space-5)' }}>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
        <div style={{ width: '2.5rem', height: '2.5rem', borderRadius: 'var(--radius)', background: 'rgba(143,255,208,0.08)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <BarChart3 className="h-5 w-5" style={{ color: 'var(--status-ok)' }} />
        </div>
        <div style={{ flex: 1 }}>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '1rem', fontWeight: 700, color: 'var(--text)', margin: 0 }}>Analytics</h3>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-dim)' }}>Dashboards, funnels, events</span>
        </div>
        <StatusPill status="live" label="Active" />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--space-2)' }}>
        <StatBlock label="Dashboards" value={config.dashboards} />
        <StatBlock label="Funnels" value={config.funnels} />
        <StatBlock label="Events" value={config.events} />
        <StatBlock label="Retention" value={`${config.retention_days}d`} />
      </div>
    </Chamber>
  );
}

function StatBlock({ label, value }: { label: string; value: string | number }) {
  return (
    <div style={{ padding: 'var(--space-2) var(--space-3)', borderRadius: 'var(--radius)', background: 'var(--panel-raised)', textAlign: 'center' }}>
      <div style={{ fontFamily: 'var(--font-mono)', fontSize: '1.25rem', fontWeight: 600, color: 'var(--text)' }}>{value}</div>
      <div style={{ fontSize: '0.6875rem', color: 'var(--text-faint)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</div>
    </div>
  );
}
