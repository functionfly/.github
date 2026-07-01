'use client';

import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Activity,
  Clock,
  BarChart3,
  Cpu,
  Hash,
} from 'lucide-react';
import { agentApi, type AgentAnalytics as AgentAnalyticsType, type ExecutionRecord, type CostBreakdown, type AgentUsage, type ModelBreakdownItem } from '@/api/agent';

type TimeRange = '24h' | '7d' | '30d';

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function StatCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="aw-stat">
      <p className="aw-stat__label">{label}</p>
      <p className="aw-stat__value">{value}</p>
      {sub && <p className="aw-stat__sub">{sub}</p>}
    </div>
  );
}

function ModelBreakdownTable({ models, isLoading }: { models: ModelBreakdownItem[]; isLoading: boolean }) {
  if (isLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  if (!models || models.length === 0) {
    return (
      <div className="aw-empty">
        <Cpu size={40} className="aw-empty__icon" />
        <span className="aw-empty__title">No model data yet</span>
        <span className="aw-empty__desc">Model usage will appear once agents make LLM calls</span>
      </div>
    );
  }

  const maxCost = Math.max(...models.map(m => m.total_cost_usd));

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
      {models.map((model, i) => (
        <div key={i} className="aw-card">
          <div className="aw-card__body" style={{ padding: 'var(--space-3) var(--space-4)' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-2)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                <Cpu size={14} style={{ color: 'var(--accent)' }} />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '13px', fontWeight: 600, color: 'var(--text)' }}>
                  {model.model_name}
                </span>
                <span style={{
                  fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase',
                  padding: '2px 6px', borderRadius: 'var(--radius-sm)',
                  color: 'var(--text-faint)', background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)',
                }}>
                  {model.provider}
                </span>
              </div>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--text)' }}>
                ${model.total_cost_usd.toFixed(4)}
              </span>
            </div>

            <div className="aw-progress" style={{ marginBottom: 'var(--space-2)' }}>
              <div className="aw-progress__fill" style={{ width: `${maxCost > 0 ? (model.total_cost_usd / maxCost) * 100 : 0}%` }} />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 'var(--space-3)' }}>
              <div>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '9px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Calls</span>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--text)', margin: '2px 0 0' }}>{model.total_calls}</p>
              </div>
              <div>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '9px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Prompt Tokens</span>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--foil-a)', margin: '2px 0 0' }}>{formatTokens(model.total_prompt_tokens)}</p>
              </div>
              <div>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '9px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Completion</span>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--foil-b)', margin: '2px 0 0' }}>{formatTokens(model.total_completion_tokens)}</p>
              </div>
              <div>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '9px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Reasoning</span>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--status-pending)', margin: '2px 0 0' }}>{formatTokens(model.total_reasoning_tokens)}</p>
              </div>
              <div>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '9px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>Avg/Call</span>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', fontWeight: 500, color: 'var(--text-dim)', margin: '2px 0 0' }}>{formatTokens(Math.round(model.avg_tokens_per_call))}</p>
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function RecentExecutions({ executions, isLoading }: { executions: ExecutionRecord[]; isLoading: boolean }) {
  if (isLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  if (!executions || executions.length === 0) {
    return (
      <div className="aw-empty">
        <Activity size={40} className="aw-empty__icon" />
        <span className="aw-empty__title">No recent executions</span>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
      {executions.slice(0, 15).map((exec) => {
        const uri = exec.functionUri ?? exec.functionName ?? 'unknown';
        const parts = uri.replace('fx://', '').replace('chat://', 'chat: ').split('/');
        const displayName = parts.length > 1 ? parts[1] : uri;

        return (
          <div key={exec.id} className="aw-feed-item">
            <span className={`aw-feed-item__dot ${exec.outcome === 'success' ? 'aw-feed-item__dot--result' : 'aw-feed-item__dot--error'}`} />
            <div className="aw-feed-item__content">
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-1)' }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text)' }}>
                  {displayName}
                </span>
                {exec.model_name && (
                  <span style={{
                    fontFamily: 'var(--font-mono)', fontSize: '9px', fontWeight: 500, letterSpacing: '0.06em',
                    padding: '1px 4px', borderRadius: 'var(--radius-sm)',
                    color: 'var(--foil-a)', background: 'rgba(159,216,255,0.08)', border: '1px solid rgba(159,216,255,0.2)',
                  }}>
                    {exec.model_name}
                  </span>
                )}
                <span style={{
                  fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 600, letterSpacing: '0.06em', textTransform: 'uppercase',
                  color: exec.outcome === 'success' ? 'var(--status-ok)' : 'var(--status-revoked)',
                }}>
                  {exec.outcome}
                </span>
              </div>
              <div className="aw-feed-item__meta">
                <span><DollarSign size={10} /> ${(exec.costUsd ?? 0).toFixed(4)}</span>
                <span><Clock size={10} /> {((exec.latencyMs ?? 0) / 1000).toFixed(2)}s</span>
                {(exec.total_tokens ?? 0) > 0 && (
                  <span><Hash size={10} /> {formatTokens(exec.total_tokens!)} tokens</span>
                )}
                {(exec.prompt_tokens ?? 0) > 0 && (
                  <span style={{ color: 'var(--foil-a)' }}>in:{formatTokens(exec.prompt_tokens!)}</span>
                )}
                {(exec.completion_tokens ?? 0) > 0 && (
                  <span style={{ color: 'var(--foil-b)' }}>out:{formatTokens(exec.completion_tokens!)}</span>
                )}
                <span>{new Date(exec.timestamp).toLocaleString()}</span>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

interface AgentAnalyticsProps {
  agentId: string;
  className?: string;
}

export function AgentAnalyticsComponent({ agentId, className }: AgentAnalyticsProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>('7d');

  const sinceMap: Record<TimeRange, string> = {
    '24h': new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    '7d': new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    '30d': new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
  };

  const { data: analyticsData, isLoading: analyticsLoading } = useQuery({
    queryKey: ['agent-analytics', agentId, timeRange],
    queryFn: () => agentApi.getAnalytics(agentId, { since: sinceMap[timeRange] }),
    enabled: !!agentId,
  });

  const { data: executionsData, isLoading: executionsLoading } = useQuery({
    queryKey: ['agent-executions', agentId, timeRange],
    queryFn: () => agentApi.listExecutions(agentId, { limit: 100 }),
    enabled: !!agentId,
  });

  const { data: modelData, isLoading: modelLoading } = useQuery({
    queryKey: ['agent-model-breakdown', agentId],
    queryFn: () => agentApi.getModelBreakdown(agentId),
    enabled: !!agentId,
  });

  const analytics = analyticsData?.analytics;
  const executions = executionsData?.executions ?? [];
  const models = modelData?.models ?? [];

  const [activeTab, setActiveTab] = useState<'overview' | 'models' | 'executions'>('overview');
  const tabs = [
    { key: 'overview' as const, label: 'Overview', icon: BarChart3 },
    { key: 'models' as const, label: 'Models', icon: Cpu },
    { key: 'executions' as const, label: 'Executions', icon: Activity },
  ];

  return (
    <div className={className}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-5)' }}>
        <div>
          <h2 style={{ fontFamily: 'var(--font-display)', fontSize: '20px', fontWeight: 600, color: 'var(--text)', margin: 0 }}>
            Agent Analytics
          </h2>
          <p style={{ fontFamily: 'var(--font-body)', fontSize: '13px', color: 'var(--text-dim)', margin: 'var(--space-1) 0 0' }}>
            Execution statistics, model usage, and cost breakdown
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--space-1)', padding: 'var(--space-1)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)' }}>
          {(['24h', '7d', '30d'] as TimeRange[]).map(range => (
            <button
              key={range}
              className={`aw-nav-item ${timeRange === range ? 'aw-nav-item--active' : ''}`}
              style={{ padding: 'var(--space-1) var(--space-2)', fontSize: '12px' }}
              onClick={() => setTimeRange(range)}
            >
              {range}
            </button>
          ))}
        </div>
      </div>

      {/* Stats Grid */}
      {analyticsLoading ? (
        <div className="aw-loading"><div className="aw-loading__spinner" /></div>
      ) : analytics ? (
        <div className="aw-stats">
          <StatCard label="Total Executions" value={(analytics.total_calls ?? 0).toLocaleString()} />
          <StatCard label="Success Rate" value={`${((analytics.success_rate ?? 0) * 100).toFixed(1)}%`} />
          <StatCard label="Avg Latency" value={`${((analytics.avg_latency_ms ?? 0) / 1000).toFixed(2)}s`} />
          <StatCard label="Total Cost" value={`$${(analytics.total_cost_usd ?? 0).toFixed(4)}`} />
          <StatCard label="Total Tokens" value={formatTokens(analytics.total_all_tokens ?? 0)} sub={`${formatTokens(analytics.total_prompt_tokens ?? 0)} in / ${formatTokens(analytics.total_completion_tokens ?? 0)} out`} />
          <StatCard label="Reasoning Tokens" value={formatTokens(analytics.total_reasoning_tokens ?? 0)} />
          <StatCard label="P50 Latency" value={`${((analytics.p50_latency_ms ?? 0) / 1000).toFixed(2)}s`} />
          <StatCard label="P95 Latency" value={`${((analytics.p95_latency_ms ?? 0) / 1000).toFixed(2)}s`} />
        </div>
      ) : null}

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 'var(--space-1)', padding: 'var(--space-1)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', marginTop: 'var(--space-5)', marginBottom: 'var(--space-4)' }}>
        {tabs.map(tab => (
          <button
            key={tab.key}
            className={`aw-nav-item ${activeTab === tab.key ? 'aw-nav-item--active' : ''}`}
            onClick={() => setActiveTab(tab.key)}
          >
            <tab.icon size={14} />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          {/* Token Distribution */}
          {analytics && (analytics.total_all_tokens ?? 0) > 0 && (
            <div className="aw-card">
              <div className="aw-card__header">
                <span className="aw-card__title">Token Distribution</span>
              </div>
              <div className="aw-card__body">
                <div style={{ display: 'flex', gap: 'var(--space-4)', alignItems: 'center' }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', height: '24px', borderRadius: 'var(--radius-sm)', overflow: 'hidden', background: 'var(--panel)' }}>
                      {(() => {
                        const total = analytics.total_all_tokens || 1;
                        const promptPct = (analytics.total_prompt_tokens / total) * 100;
                        const completionPct = (analytics.total_completion_tokens / total) * 100;
                        const reasoningPct = (analytics.total_reasoning_tokens / total) * 100;
                        const otherCompletionPct = Math.max(0, completionPct - reasoningPct);
                        return (
                          <>
                            <div style={{ width: `${promptPct}%`, background: 'var(--foil-a)', transition: 'width var(--duration-base)' }} title={`Prompt: ${formatTokens(analytics.total_prompt_tokens)}`} />
                            <div style={{ width: `${otherCompletionPct}%`, background: 'var(--foil-b)', transition: 'width var(--duration-base)' }} title={`Completion: ${formatTokens(analytics.total_completion_tokens - analytics.total_reasoning_tokens)}`} />
                            <div style={{ width: `${reasoningPct}%`, background: 'var(--status-pending)', transition: 'width var(--duration-base)' }} title={`Reasoning: ${formatTokens(analytics.total_reasoning_tokens)}`} />
                          </>
                        );
                      })()}
                    </div>
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)', flexShrink: 0 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                      <span style={{ width: '8px', height: '8px', borderRadius: '2px', background: 'var(--foil-a)', flexShrink: 0 }} />
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-dim)' }}>Prompt ({formatTokens(analytics.total_prompt_tokens)})</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                      <span style={{ width: '8px', height: '8px', borderRadius: '2px', background: 'var(--foil-b)', flexShrink: 0 }} />
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-dim)' }}>Completion ({formatTokens(analytics.total_completion_tokens - analytics.total_reasoning_tokens)})</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                      <span style={{ width: '8px', height: '8px', borderRadius: '2px', background: 'var(--status-pending)', flexShrink: 0 }} />
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-dim)' }}>Reasoning ({formatTokens(analytics.total_reasoning_tokens)})</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Recent Executions */}
          <RecentExecutions executions={executions} isLoading={executionsLoading} />
        </div>
      )}

      {activeTab === 'models' && (
        <ModelBreakdownTable models={models} isLoading={modelLoading} />
      )}

      {activeTab === 'executions' && (
        <RecentExecutions executions={executions} isLoading={executionsLoading} />
      )}
    </div>
  );
}

export default AgentAnalyticsComponent;
