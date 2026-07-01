import { agentApi } from '@/api/agent';
import { SealedButton, Table } from '@/components/containment';
import { useQuery } from '@tanstack/react-query';
import { DollarSign, TrendingDown, TrendingUp, Wallet } from 'lucide-react';
import { useState } from 'react';

interface CostsViewProps {
  agentId: string;
}

export function CostsView({ agentId }: CostsViewProps) {
  const [spendCap, setSpendCap] = useState('');

  const { data: billingData, isLoading: billingLoading } = useQuery({
    queryKey: ['agent-billing', agentId],
    queryFn: () => agentApi.getBillingSummary(agentId),
    enabled: !!agentId,
  });

  const { data: costData } = useQuery({
    queryKey: ['agent-cost-breakdown', agentId],
    queryFn: () => agentApi.getCostBreakdown(agentId),
    enabled: !!agentId,
  });

  const { data: walletData } = useQuery({
    queryKey: ['agent-wallet', agentId],
    queryFn: () => agentApi.getWallet(agentId),
    enabled: !!agentId,
  });

  const { data: txData } = useQuery({
    queryKey: ['agent-transactions', agentId],
    queryFn: () => agentApi.listWalletTransactions(agentId, { limit: 50 }),
    enabled: !!agentId,
  });

  const billing = billingData?.summary as any;
  const breakdown = costData?.breakdown as any;
  const wallet = walletData?.wallet as any;
  const transactions = txData?.transactions ?? [];

  if (billingLoading) {
    return <div className="aw-loading"><div className="aw-loading__spinner" /></div>;
  }

  const costByFunction = breakdown?.by_function ?? [];
  const maxCost = costByFunction.length > 0 ? Math.max(...costByFunction.map((f: any) => f.cost_usd ?? 0)) : 1;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <div className="aw-center__header">
        <div>
          <h2 className="aw-center__title">Costs</h2>
          <p className="aw-center__subtitle">Budget intelligence and spending controls</p>
        </div>
      </div>

      <div className="aw-stats">
        <div className="aw-stat">
          <p className="aw-stat__label">Total Spend</p>
          <p className="aw-stat__value">${(billing?.total_spend_usd ?? 0).toFixed(2)}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Balance</p>
          <p className="aw-stat__value">${(wallet?.balance_usd ?? 0).toFixed(2)}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Avg / Execution</p>
          <p className="aw-stat__value">${(billing?.avg_cost_per_execution ?? 0).toFixed(4)}</p>
        </div>
        <div className="aw-stat">
          <p className="aw-stat__label">Today</p>
          <p className="aw-stat__value">${(billing?.spend_today_usd ?? 0).toFixed(2)}</p>
        </div>
      </div>

      {/* Cost by Function */}
      {costByFunction.length > 0 && (
        <div className="aw-card">
          <div className="aw-card__header">
            <span className="aw-card__title">Cost by Function</span>
          </div>
          <div className="aw-card__body">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
              {costByFunction.slice(0, 10).map((fn: any, i: number) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text)', width: '140px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {fn.function_name ?? fn.name ?? 'unknown'}
                  </span>
                  <div className="aw-progress" style={{ flex: 1 }}>
                    <div
                      className="aw-progress__fill"
                      style={{ width: `${((fn.cost_usd ?? 0) / maxCost) * 100}%` }}
                    />
                  </div>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-dim)', width: '60px', textAlign: 'right' }}>
                    ${(fn.cost_usd ?? 0).toFixed(4)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Budget Controls */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <Wallet size={14} />
            Budget Controls
          </span>
        </div>
        <div className="aw-card__body">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 'var(--space-3)' }}>
              {['Daily', 'Weekly', 'Monthly'].map(period => (
                <div key={period} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-1)' }}>
                  <label style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 500, letterSpacing: '0.06em', textTransform: 'uppercase', color: 'var(--text-faint)' }}>
                    {period} Cap
                  </label>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)' }}>
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text-dim)' }}>$</span>
                    <input
                      style={{ flex: 1, padding: 'var(--space-1) var(--space-2)', fontFamily: 'var(--font-mono)', fontSize: '13px', color: 'var(--text)', background: 'var(--panel-raised)', border: '1px solid var(--steel)', borderRadius: 'var(--radius)', textAlign: 'right' }}
                      placeholder="0.00"
                      value={period === 'Daily' ? spendCap : ''}
                      onChange={period === 'Daily' ? e => setSpendCap(e.target.value) : undefined}
                    />
                  </div>
                </div>
              ))}
            </div>
            <SealedButton size="sm" iconLeft={<DollarSign size={12} />}>
              Save Budget Limits
            </SealedButton>
          </div>
        </div>
      </div>

      {/* Recent Transactions */}
      <div className="aw-card">
        <div className="aw-card__header">
          <span className="aw-card__title">Recent Transactions</span>
        </div>
        <div className="aw-card__body aw-card__body--flush">
          {transactions.length === 0 ? (
            <div className="aw-empty" style={{ padding: 'var(--space-5)' }}>
              <DollarSign size={32} className="aw-empty__icon" />
              <span className="aw-empty__title">No transactions</span>
            </div>
          ) : (
            <Table
              columns={[
                { key: 'type', header: 'Type', render: (row: any) => (
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', textTransform: 'uppercase', color: row.type === 'credit' ? 'var(--status-ok)' : 'var(--text-dim)' }}>
                    {row.type ?? 'debit'}
                  </span>
                )},
                { key: 'amount', header: 'Amount', align: 'right', render: (row: any) => (
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '13px', display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: '4px' }}>
                    {(row.type === 'credit') ? <TrendingUp size={12} style={{ color: 'var(--status-ok)' }} /> : <TrendingDown size={12} style={{ color: 'var(--text-faint)' }} /> }
                    ${(row.amount_usd ?? row.amount ?? 0).toFixed(4)}
                  </span>
                )},
                { key: 'description', header: 'Description' },
                { key: 'date', header: 'Date', render: (row: any) => (
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-faint)' }}>
                    {row.created_at ? new Date(row.created_at).toLocaleDateString() : '—'}
                  </span>
                )},
              ]}
              data={transactions}
              emptyMessage="No transactions"
            />
          )}
        </div>
      </div>
    </div>
  );
}
