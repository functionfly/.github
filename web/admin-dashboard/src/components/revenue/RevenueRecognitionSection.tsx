import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { format, subMonths } from 'date-fns';
import { DollarSign, TrendingUp, Calendar, ChevronDown, ChevronUp, RefreshCw, AlertCircle, Check } from 'lucide-react';
import { toast } from 'sonner';

interface DeferredRevenueResponse {
  tenant_id: string;
  period: string;
  opening_balance_cents: number;
  new_deferred_cents: number;
  recognized_cents: number;
  closing_balance_cents: number;
}

interface RecognizedRevenueResponse {
  tenant_id: string;
  period: string;
  subscription_revenue_cents: number;
  usage_revenue_cents: number;
  one_time_revenue_cents: number;
  total_cents: number;
}

interface RevenueReportResponse {
  report_id: string;
  period: string;
  total_revenue_cents: number;
  total_deferred_cents: number;
  total_recognized_cents: number;
  opening_deferred_cents: number;
  new_deferred_cents: number;
  recognized_from_deferred_cents: number;
  closing_deferred_cents: number;
  over_time_revenue_cents: number;
  point_in_time_revenue_cents: number;
}

interface UnbilledRevenueResponse {
  tenant_id: string;
  unbilled_revenue_cents: number;
}

function formatCents(cents: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(cents / 100);
}

function formatPeriod(period: string): string {
  const [year, month] = period.split('-');
  const date = new Date(parseInt(year), parseInt(month) - 1);
  return format(date, 'MMMM yyyy');
}

export function RevenueRecognitionSection() {
  const queryClient = useQueryClient();
  const [selectedPeriod, setSelectedPeriod] = useState(format(new Date(), 'yyyy-MM'));
  const [expandedSection, setExpandedSection] = useState<string | null>('summary');

  const periods = Array.from({ length: 12 }, (_, i) => {
    const date = subMonths(new Date(), i);
    return format(date, 'yyyy-MM');
  });

  const { data: deferredData, isLoading: deferredLoading } = useQuery({
    queryKey: ['admin-revenue', 'deferred', selectedPeriod],
    queryFn: () => adminApiClient.get<DeferredRevenueResponse>(`/billing/revenue/deferred?period=${selectedPeriod}`),
  });

  const { data: recognizedData, isLoading: recognizedLoading } = useQuery({
    queryKey: ['admin-revenue', 'recognized', selectedPeriod],
    queryFn: () => adminApiClient.get<RecognizedRevenueResponse>(`/billing/revenue/recognized?period=${selectedPeriod}`),
  });

  const { data: reportData, isLoading: reportLoading } = useQuery({
    queryKey: ['admin-revenue', 'report', selectedPeriod],
    queryFn: () => adminApiClient.get<RevenueReportResponse>(`/billing/revenue/report?period=${selectedPeriod}`),
  });

  const { data: unbilledData, isLoading: unbilledLoading } = useQuery({
    queryKey: ['admin-revenue', 'unbilled'],
    queryFn: () => adminApiClient.get<UnbilledRevenueResponse>('/billing/revenue/unbilled'),
  });

  const processRecognitionMutation = useMutation({
    mutationFn: async () => {
      return adminApiClient.post('/billing/revenue/process', {});
    },
    onSuccess: () => {
      toast.success('Revenue recognition processed successfully');
      queryClient.invalidateQueries({ queryKey: ['admin-revenue'] });
    },
    onError: () => {
      toast.error('Failed to process revenue recognition');
    },
  });

  const recognized = recognizedData?.data;
  const deferred = deferredData?.data;
  const report = reportData?.data;
  const unbilled = unbilledData?.data;

  return (
    <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
      <div className="p-6 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Revenue Recognition (ASC 606/IFRS 15)
          </h2>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Deferred revenue tracking and recognition scheduling for accrual accounting
          </p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={selectedPeriod}
            onChange={(e) => setSelectedPeriod(e.target.value)}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 text-sm"
          >
            {periods.map((period) => (
              <option key={period} value={period}>
                {formatPeriod(period)}
              </option>
            ))}
          </select>
          <button
            onClick={() => processRecognitionMutation.mutate()}
            disabled={processRecognitionMutation.isPending}
            className="flex items-center gap-2 px-3 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-60 text-sm"
          >
            <RefreshCw className={`w-4 h-4 ${processRecognitionMutation.isPending ? 'animate-spin' : ''}`} />
            Process Recognition
          </button>
        </div>
      </div>

      <div className="p-6 space-y-6">
        {/* Key Metrics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-slate-50 dark:bg-slate-800/50 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-600 dark:text-gray-400">Recognized Revenue</p>
              <TrendingUp className="w-5 h-5 text-green-600" />
            </div>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {recognizedLoading ? '...' : formatCents(recognized?.total_cents ?? 0)}
            </p>
          </div>

          <div className="bg-slate-50 dark:bg-slate-800/50 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-600 dark:text-gray-400">Closing Deferred</p>
              <Calendar className="w-5 h-5 text-amber-600" />
            </div>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {deferredLoading ? '...' : formatCents(deferred?.closing_balance_cents ?? 0)}
            </p>
          </div>

          <div className="bg-slate-50 dark:bg-slate-800/50 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-600 dark:text-gray-400">Unbilled Revenue</p>
              <DollarSign className="w-5 h-5 text-blue-600" />
            </div>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {unbilledLoading ? '...' : formatCents(unbilled?.unbilled_revenue_cents ?? 0)}
            </p>
          </div>

          <div className="bg-slate-50 dark:bg-slate-800/50 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-600 dark:text-gray-400">Point-in-Time</p>
              <Check className="w-5 h-5 text-purple-600" />
            </div>
            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
              {reportLoading ? '...' : formatCents(report?.point_in_time_revenue_cents ?? 0)}
            </p>
          </div>
        </div>

        {/* Revenue Breakdown */}
        <div>
          <button
            onClick={() => setExpandedSection(expandedSection === 'breakdown' ? null : 'breakdown')}
            className="flex items-center justify-between w-full text-left"
          >
            <h3 className="font-medium text-gray-900 dark:text-gray-100">Revenue Breakdown</h3>
            {expandedSection === 'breakdown' ? (
              <ChevronUp className="w-5 h-5 text-gray-400" />
            ) : (
              <ChevronDown className="w-5 h-5 text-gray-400" />
            )}
          </button>

          {expandedSection === 'breakdown' && (
            <div className="mt-4 space-y-4">
              {/* Deferred Revenue Movement */}
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Deferred Revenue Movement</h4>
                <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Opening Balance</p>
                    <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                      {deferredLoading ? '...' : formatCents(deferred?.opening_balance_cents ?? 0)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">+ New Deferred</p>
                    <p className="text-lg font-semibold text-green-600">
                      {deferredLoading ? '...' : `+${formatCents(deferred?.new_deferred_cents ?? 0)}`}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">- Recognized</p>
                    <p className="text-lg font-semibold text-red-600">
                      {deferredLoading ? '...' : `-${formatCents(deferred?.recognized_cents ?? 0)}`}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Closing Balance</p>
                    <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                      {deferredLoading ? '...' : formatCents(deferred?.closing_balance_cents ?? 0)}
                    </p>
                  </div>
                </div>
              </div>

              {/* Revenue by Type */}
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Recognized by Type</h4>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Subscription</p>
                    <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                      {recognizedLoading ? '...' : formatCents(recognized?.subscription_revenue_cents ?? 0)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Usage</p>
                    <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                      {recognizedLoading ? '...' : formatCents(recognized?.usage_revenue_cents ?? 0)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">One-time</p>
                    <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                      {recognizedLoading ? '...' : formatCents(recognized?.one_time_revenue_cents ?? 0)}
                    </p>
                  </div>
                </div>
              </div>

              {/* Recognition Method Split */}
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Recognition Method</h4>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Over Time (Linear)</p>
                    <p className="text-lg font-semibold text-blue-600">
                      {reportLoading ? '...' : formatCents(report?.over_time_revenue_cents ?? 0)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">Point in Time</p>
                    <p className="text-lg font-semibold text-purple-600">
                      {reportLoading ? '...' : formatCents(report?.point_in_time_revenue_cents ?? 0)}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Full Report Details */}
        <div>
          <button
            onClick={() => setExpandedSection(expandedSection === 'report' ? null : 'report')}
            className="flex items-center justify-between w-full text-left"
          >
            <h3 className="font-medium text-gray-900 dark:text-gray-100">Recognition Report Details</h3>
            {expandedSection === 'report' ? (
              <ChevronUp className="w-5 h-5 text-gray-400" />
            ) : (
              <ChevronDown className="w-5 h-5 text-gray-400" />
            )}
          </button>

          {expandedSection === 'report' && (
            <div className="mt-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4 space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">Report ID</span>
                <span className="font-mono text-sm text-gray-900 dark:text-gray-100">{report?.report_id ?? '—'}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">Period</span>
                <span className="text-sm text-gray-900 dark:text-gray-100">{formatPeriod(selectedPeriod)}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">Total Revenue</span>
                <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {reportLoading ? '...' : formatCents(report?.total_revenue_cents ?? 0)}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">Total Recognized</span>
                <span className="text-sm font-medium text-green-600">
                  {reportLoading ? '...' : formatCents(report?.total_recognized_cents ?? 0)}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600 dark:text-gray-400">Total Deferred</span>
                <span className="text-sm font-medium text-amber-600">
                  {reportLoading ? '...' : formatCents(report?.total_deferred_cents ?? 0)}
                </span>
              </div>
            </div>
          )}
        </div>

        {/* Accounting Notes */}
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
          <div className="flex gap-3">
            <AlertCircle className="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
            <div className="space-y-1">
              <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
                ASC 606 / IFRS 15 Compliance
              </p>
              <p className="text-xs text-amber-700 dark:text-amber-300">
                Revenue is recognized when control of goods/services is transferred to customers, 
                either at a point in time or over time. Deferred revenue represents performance 
                obligations that are not yet satisfied.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}