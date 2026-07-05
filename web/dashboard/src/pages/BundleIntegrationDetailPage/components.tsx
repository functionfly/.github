import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Shield, CreditCard, Mail, BarChart3, CheckCircle,
  ToggleLeft, ToggleRight,
} from 'lucide-react';
import { apiClient } from '@/api/client';
import {
  Chamber, CornerBrace, StatusPill,
} from '@/components/containment';

// ─── Types ─────────────────────────────────────────────────────────────────

export interface OAuthProvider {
  provider: string;
  enabled: boolean;
  client_id: string;
  has_secret: boolean;
}

export interface BundleConfig {
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

// ─── Hooks ──────────────────────────────────────────────────────────────────

export function useBundleConfig(appId: string) {
  return useQuery<BundleConfig>({
    queryKey: ['bundle-config', appId],
    queryFn: () => apiClient.get<BundleConfig>(`/v1/apps/${appId}/bundle/config`),
    enabled: !!appId,
  });
}

export function useToggleOAuth(appId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ provider, enabled }: { provider: string; enabled: boolean }) =>
      apiClient.put(`/v1/apps/${appId}/bundle/config/auth/oauth/${provider}`, { enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bundle-config', appId] }),
  });
}

export function useToggleWorkflow(appId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ slug, active }: { slug: string; active: boolean }) =>
      apiClient.put(`/v1/apps/${appId}/bundle/config/email/workflows/${slug}`, { is_active: active }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bundle-config', appId] }),
  });
}

// ─── Cards ───────────────────────────────────────────────────────────────────

export function AuthCard({ config, appId }: { config: BundleConfig['auth']; appId: string }) {
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

export function PaymentsCard({ config }: { config: BundleConfig['payments'] }) {
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

export function EmailCard({ config, appId }: { config: BundleConfig['email']; appId: string }) {
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

export function AnalyticsCard({ config }: { config: BundleConfig['analytics'] }) {
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
