import { useQuery } from '@tanstack/react-query';
import { compensationApi, type CompensationRecord, type EquityGrant } from '@/api/compensation';
import { useAuthStore } from '@/stores/authStore';
import { DollarSign, TrendingUp, Calendar, Shield } from 'lucide-react';
import { formatCurrency, formatDate } from '@/lib/utils';

export function CompensationPage() {
  const { user } = useAuthStore();

  const { data: compData } = useQuery({
    queryKey: ['compensation', user?.id],
    queryFn: () => compensationApi.get(user!.id),
    enabled: !!user?.id,
  });

  const { data: equityData } = useQuery({
    queryKey: ['equity', user?.id],
    queryFn: () => compensationApi.listEquity(user!.id),
    enabled: !!user?.id,
  });

  const compensation = compData?.data?.compensation;
  const grants = equityData?.data?.grants || [];

  const now = new Date();

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Compensation</h1>

      <div className="rounded-lg border border-yellow-600/30 bg-yellow-600/10 p-3 text-sm text-yellow-300">
        <Shield className="mr-2 inline h-4 w-4" />
        Compensation data is confidential. Access is logged and audited.
      </div>

      {compensation ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
            <div className="mb-2 flex items-center gap-2 text-gray-400">
              <DollarSign className="h-4 w-4" />
              <span className="text-sm">Base Salary</span>
            </div>
            <p className="text-2xl font-bold text-gray-100">
              {formatCurrency(compensation.base_salary_cents, compensation.currency)}
            </p>
            <p className="mt-1 text-xs text-gray-500">
              Paid {compensation.pay_frequency}
            </p>
          </div>

          <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
            <div className="mb-2 flex items-center gap-2 text-gray-400">
              <Calendar className="h-4 w-4" />
              <span className="text-sm">Effective Date</span>
            </div>
            <p className="text-lg font-semibold text-gray-100">
              {formatDate(compensation.effective_date)}
            </p>
            {compensation.review_date && (
              <p className="mt-1 text-xs text-gray-500">
                Next review: {formatDate(compensation.review_date)}
              </p>
            )}
          </div>

          <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
            <div className="mb-2 flex items-center gap-2 text-gray-400">
              <TrendingUp className="h-4 w-4" />
              <span className="text-sm">Currency</span>
            </div>
            <p className="text-lg font-semibold text-gray-100">{compensation.currency}</p>
          </div>
        </div>
      ) : (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-8 text-center">
          <DollarSign className="mx-auto mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">Compensation data not available</p>
        </div>
      )}

      <h2 className="mt-8 text-xl font-semibold">Equity Grants</h2>

      {grants.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-8 text-center">
          <TrendingUp className="mx-auto mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No equity grants</p>
        </div>
      ) : (
        <div className="space-y-4">
          {grants.map((grant) => {
            const vestingStart = new Date(grant.vesting_start);
            const vestingEnd = new Date(grant.vesting_end);
            const totalDuration = vestingEnd.getTime() - vestingStart.getTime();
            const elapsed = now.getTime() - vestingStart.getTime();
            const vestingPct = Math.min(100, Math.max(0, (elapsed / totalDuration) * 100));

            return (
              <div key={grant.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
                <div className="mb-3 flex items-start justify-between">
                  <div>
                    <p className="font-semibold text-gray-100">{grant.grant_type.toUpperCase()}</p>
                    <p className="text-sm text-gray-400">{grant.shares.toLocaleString()} shares</p>
                  </div>
                  <span className={`rounded-full px-2 py-0.5 text-xs ${
                    grant.status === 'active' ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'
                  }`}>
                    {grant.status}
                  </span>
                </div>

                {grant.strike_price_cents && (
                  <p className="mb-2 text-sm text-gray-400">
                    Strike price: {formatCurrency(grant.strike_price_cents)}
                  </p>
                )}

                <div className="mb-2 flex justify-between text-xs text-gray-500">
                  <span>{formatDate(grant.vesting_start)}</span>
                  <span>{formatDate(grant.vesting_end)}</span>
                </div>

                <div className="h-3 rounded-full bg-gray-800">
                  <div
                    className="h-full rounded-full bg-blue-600 transition-all"
                    style={{ width: `${vestingPct}%` }}
                  />
                </div>

                <div className="mt-2 flex justify-between text-sm">
                  <span className="text-gray-400">
                    {grant.vested_shares.toLocaleString()} / {grant.shares.toLocaleString()} vested
                  </span>
                  <span className="font-medium text-blue-400">{Math.round(vestingPct)}%</span>
                </div>

                {grant.cliff_date && (
                  <p className="mt-2 text-xs text-gray-500">
                    Cliff: {formatDate(grant.cliff_date)}
                  </p>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
