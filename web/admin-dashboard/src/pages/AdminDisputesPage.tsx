/**
 * Admin Disputes Page
 * Manage chargebacks and disputes with automated response workflow
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import {
  Shield,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  DollarSign,
  Search,
  ChevronDown,
  ChevronUp,
  Eye,
  Send,
  RotateCcw,
  Ban,
  FileText,
  Activity,
  Filter,
} from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { toast } from 'sonner';
import { format } from 'date-fns';

interface PaymentDispute {
  id: string;
  stripe_dispute_id: string;
  stripe_charge_id: string;
  amount_cents: number;
  currency: string;
  reason: string;
  status: string;
  evidence_due_by?: string;
  evidence_submitted: boolean;
  outcome?: string;
  tenant_id?: string;
  tenant_name?: string;
  user_id?: string;
  user_email?: string;
  created_at: string;
  updated_at: string;
}

interface DisputeStats {
  total_disputes: number;
  open_disputes: number;
  won_disputes: number;
  lost_disputes: number;
  total_disputed_cents: number;
  by_status: Record<string, number>;
}

interface EvidenceDetails {
  product_description?: string;
  customer_email?: string;
  customer_name?: string;
  customer_purchase_ip?: string;
  billing_address?: string;
  receipt_url?: string;
  service_date?: string;
  service_document?: string;
  refund_policy_url?: string;
  access_activity_log?: string;
  customer_communication?: string;
}

interface AutomationLog {
  action: string;
  outcome: string;
  details?: Record<string, unknown>;
  created_at: string;
}

const STATUS_COLORS: Record<string, string> = {
  needs_response: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  warning_needs_response: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300',
  needs_review: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  under_review: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  won: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-300',
  lost: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  closed: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300',
  charge_refunded: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300',
};

const REASON_LABELS: Record<string, string> = {
  duplicate: 'Duplicate Charge',
  fraudulent: 'Fraudulent',
  product_not_received: 'Product Not Received',
  product_unacceptable: 'Product Unacceptable',
  subscription_canceled: 'Subscription Canceled',
  credit_not_processed: 'Credit Not Processed',
  incorrect_details: 'Incorrect Details',
  unrecognized: 'Unrecognized',
  bank_cannot_process: 'Bank Cannot Process',
  debit_not_authorized: 'Debit Not Authorized',
};

export function AdminDisputesPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [selectedDispute, setSelectedDispute] = useState<PaymentDispute | null>(null);
  const [showEvidenceModal, setShowEvidenceModal] = useState(false);
  const queryClient = useQueryClient();

  const { data: statsResponse, isLoading: statsLoading } = useQuery({
    queryKey: ['admin-disputes-stats'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<DisputeStats>('/billing/disputes/stats');
      } catch {
        return { data: null };
      }
    },
    staleTime: 1000 * 60,
  });

  const { data: disputesResponse, isLoading: disputesLoading, refetch: refetchDisputes } = useQuery({
    queryKey: ['admin-disputes', statusFilter],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (statusFilter !== 'all') {
        params.set('status', statusFilter);
      }
      const queryString = params.toString();
      const url = `/billing/disputes${queryString ? `?${queryString}` : ''}`;
      return adminApiClient.get<{ disputes: PaymentDispute[]; total: number }>(url);
    },
    staleTime: 1000 * 30,
  });

  const { data: evidenceData, isLoading: evidenceLoading } = useQuery({
    queryKey: ['admin-dispute-evidence', selectedDispute?.id],
    queryFn: async () => {
      if (!selectedDispute) return null;
      try {
        return await adminApiClient.get<EvidenceDetails>(`/billing/disputes/${selectedDispute.id}/evidence`);
      } catch {
        return null;
      }
    },
    enabled: !!selectedDispute && showEvidenceModal,
  });

  const { data: automationLogs } = useQuery({
    queryKey: ['admin-dispute-logs', selectedDispute?.id],
    queryFn: async () => {
      if (!selectedDispute) return null;
      try {
        const response = await adminApiClient.get<{ logs: AutomationLog[] }>(`/billing/disputes/${selectedDispute.id}/automation-log`);
        return response.data?.logs || [];
      } catch {
        return [];
      }
    },
    enabled: !!selectedDispute,
  });

  const submitEvidenceMutation = useMutation({
    mutationFn: async (disputeId: string) => {
      return adminApiClient.post(`/billing/disputes/${disputeId}/submit`, {});
    },
    onSuccess: () => {
      toast.success('Evidence submitted successfully');
      queryClient.invalidateQueries({ queryKey: ['admin-disputes'] });
      queryClient.invalidateQueries({ queryKey: ['admin-disputes-stats'] });
      setShowEvidenceModal(false);
    },
    onError: () => {
      toast.error('Failed to submit evidence');
    },
  });

  const issueRefundMutation = useMutation({
    mutationFn: async ({ disputeId, reason }: { disputeId: string; reason: string }) => {
      return adminApiClient.post(`/billing/disputes/${disputeId}/refund`, { reason });
    },
    onSuccess: () => {
      toast.success('Refund issued successfully');
      queryClient.invalidateQueries({ queryKey: ['admin-disputes'] });
      queryClient.invalidateQueries({ queryKey: ['admin-disputes-stats'] });
    },
    onError: () => {
      toast.error('Failed to issue refund');
    },
  });

  const skipAutoMutation = useMutation({
    mutationFn: async (disputeId: string) => {
      return adminApiClient.post(`/billing/disputes/${disputeId}/skip`, {});
    },
    onSuccess: () => {
      toast.success('Auto-response skipped');
      queryClient.invalidateQueries({ queryKey: ['admin-disputes'] });
    },
    onError: () => {
      toast.error('Failed to skip auto-response');
    },
  });

  const stats = statsResponse?.data;
  const disputes = disputesResponse?.data?.disputes || [];

  const filteredDisputes = disputes.filter((d) => {
    if (!searchTerm) return true;
    const term = searchTerm.toLowerCase();
    return (
      d.stripe_dispute_id.toLowerCase().includes(term) ||
      d.tenant_name?.toLowerCase().includes(term) ||
      d.user_email?.toLowerCase().includes(term) ||
      d.reason.toLowerCase().includes(term)
    );
  });

  const formatAmount = (cents: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency.toUpperCase(),
    }).format(cents / 100);
  };

  const formatDate = (dateStr: string) => {
    return format(new Date(dateStr), 'MMM d, yyyy HH:mm');
  };

  if (statsLoading && !stats) {
    return <LoadingScreen />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Shield className="w-8 h-8" />
            Chargeback Disputes
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Automated dispute response workflow management
          </p>
        </div>
      </div>

      {/* Stats Overview */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <StatCard
          label="Total Disputes"
          value={stats?.total_disputes || 0}
          icon={FileText}
          color="gray"
        />
        <StatCard
          label="Open"
          value={stats?.open_disputes || 0}
          icon={Clock}
          color="yellow"
        />
        <StatCard
          label="Won"
          value={stats?.won_disputes || 0}
          icon={CheckCircle}
          color="emerald"
        />
        <StatCard
          label="Lost"
          value={stats?.lost_disputes || 0}
          icon={XCircle}
          color="red"
        />
        <StatCard
          label="Total Disputed"
          value={formatAmount(stats?.total_disputed_cents || 0)}
          icon={DollarSign}
          color="blue"
        />
      </div>

      {/* Filters and Search */}
      <div className="flex flex-col sm:flex-row gap-4 items-center justify-between bg-white dark:bg-gray-800 rounded-lg p-4 shadow">
        <div className="relative flex-1 w-full max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search by dispute ID, tenant, email, or reason..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-gray-500" />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          >
            <option value="all">All Status</option>
            <option value="needs_response">Needs Response</option>
            <option value="warning_needs_response">Warning - Needs Response</option>
            <option value="needs_review">Needs Review</option>
            <option value="won">Won</option>
            <option value="lost">Lost</option>
            <option value="closed">Closed</option>
          </select>
        </div>
      </div>

      {/* Disputes Table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-300">Dispute</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-300">Amount</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-300">Reason</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-300">Status</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-300">Evidence Due</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-300">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              {disputesLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                    Loading disputes...
                  </td>
                </tr>
              ) : filteredDisputes.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-gray-500">
                    No disputes found
                  </td>
                </tr>
              ) : (
                filteredDisputes.map((dispute) => (
                  <tr key={dispute.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-4 py-3">
                      <div className="text-sm">
                        <div className="font-medium text-gray-900 dark:text-white">
                          {dispute.tenant_name || 'Unknown Tenant'}
                        </div>
                        <div className="text-xs text-gray-500 font-mono">
                          {dispute.stripe_dispute_id.slice(0, 20)}...
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">
                      {formatAmount(dispute.amount_cents, dispute.currency)}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                      {REASON_LABELS[dispute.reason] || dispute.reason}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                          STATUS_COLORS[dispute.status] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
                        }`}
                      >
                        {dispute.status.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                      {dispute.evidence_due_by
                        ? formatDate(dispute.evidence_due_by)
                        : '-'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => {
                            setSelectedDispute(dispute);
                            setShowEvidenceModal(true);
                          }}
                          className="p-1.5 text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/30 rounded"
                          title="View Evidence"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        {(dispute.status === 'needs_response' ||
                          dispute.status === 'warning_needs_response' ||
                          dispute.status === 'needs_review') && (
                          <>
                            <button
                              onClick={() => submitEvidenceMutation.mutate(dispute.id)}
                              className="p-1.5 text-emerald-600 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-900/30 rounded"
                              title="Submit Evidence"
                              disabled={submitEvidenceMutation.isPending}
                            >
                              <Send className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() =>
                                issueRefundMutation.mutate({ disputeId: dispute.id, reason: 'fraudulent' })
                              }
                              className="p-1.5 text-orange-600 hover:bg-orange-50 dark:text-orange-400 dark:hover:bg-orange-900/30 rounded"
                              title="Issue Refund"
                              disabled={issueRefundMutation.isPending}
                            >
                              <RotateCcw className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() => skipAutoMutation.mutate(dispute.id)}
                              className="p-1.5 text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-700 rounded"
                              title="Skip Auto-Response"
                              disabled={skipAutoMutation.isPending}
                            >
                              <Ban className="w-4 h-4" />
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Evidence Modal */}
      {showEvidenceModal && selectedDispute && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden">
            <div className="p-6 border-b dark:border-gray-700">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold">Dispute Evidence Preview</h2>
                <button
                  onClick={() => setShowEvidenceModal(false)}
                  className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                >
                  <XCircle className="w-5 h-5" />
                </button>
              </div>
              <p className="text-sm text-gray-500 mt-1">
                {selectedDispute.stripe_dispute_id}
              </p>
            </div>

            <div className="p-6 overflow-y-auto max-h-[calc(90vh-180px)]">
              {evidenceLoading ? (
                <div className="text-center py-8 text-gray-500">Loading evidence...</div>
              ) : evidenceData?.data ? (
                <div className="space-y-6">
                  <EvidenceSection title="Customer Information">
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="text-gray-500">Name:</span>
                        <span className="ml-2 font-medium">{evidenceData.data.customer_name || 'N/A'}</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Email:</span>
                        <span className="ml-2 font-medium">{evidenceData.data.customer_email || 'N/A'}</span>
                      </div>
                      <div className="col-span-2">
                        <span className="text-gray-500">Billing Address:</span>
                        <span className="ml-2 font-medium">{evidenceData.data.billing_address || 'N/A'}</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Purchase IP:</span>
                        <span className="ml-2 font-medium font-mono">{evidenceData.data.customer_purchase_ip || 'N/A'}</span>
                      </div>
                    </div>
                  </EvidenceSection>

                  <EvidenceSection title="Product/Service">
                    <p className="text-sm">{evidenceData.data.product_description || 'N/A'}</p>
                    <div className="mt-2 text-sm">
                      <span className="text-gray-500">Service Date:</span>
                      <span className="ml-2">{evidenceData.data.service_date || 'N/A'}</span>
                    </div>
                  </EvidenceSection>

                  <EvidenceSection title="Access Activity Log">
                    <pre className="text-xs bg-gray-50 dark:bg-gray-900 p-3 rounded overflow-x-auto whitespace-pre-wrap dark:text-gray-300">
                      {evidenceData.data.access_activity_log || 'No activity log available'}
                    </pre>
                  </EvidenceSection>

                  <EvidenceSection title="Customer Communication">
                    <p className="text-sm">{evidenceData.data.customer_communication || 'No communication on file'}</p>
                  </EvidenceSection>

                  <EvidenceSection title="Refund Policy">
                    <p className="text-sm">{evidenceData.data.refund_policy_url || 'N/A'}</p>
                  </EvidenceSection>

                  {evidenceData.data.receipt_url && (
                    <EvidenceSection title="Receipt">
                      <a
                        href={evidenceData.data.receipt_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-blue-600 hover:underline"
                      >
                        View Receipt →
                      </a>
                    </EvidenceSection>
                  )}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  Evidence not available
                </div>
              )}

              {/* Automation Log */}
              {automationLogs && automationLogs.length > 0 && (
                <EvidenceSection title="Automation History">
                  <div className="space-y-2">
                    {automationLogs.map((log, i) => (
                      <div
                        key={i}
                        className="flex items-start gap-3 text-sm border-l-2 border-gray-200 dark:border-gray-600 pl-3"
                      >
                        <Activity className="w-4 h-4 text-gray-400 mt-0.5" />
                        <div>
                          <div className="font-medium">
                            {log.action.replace(/_/g, ' ')}
                            <span
                              className={`ml-2 px-1.5 py-0.5 rounded text-xs ${
                                log.outcome === 'success'
                                  ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300'
                                  : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
                              }`}
                            >
                              {log.outcome}
                            </span>
                          </div>
                          <div className="text-gray-500 text-xs">{formatDate(log.created_at)}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </EvidenceSection>
              )}
            </div>

            <div className="p-6 border-t dark:border-gray-700 flex justify-end gap-3">
              <button
                onClick={() => setShowEvidenceModal(false)}
                className="px-4 py-2 border rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                Close
              </button>
              {(selectedDispute.status === 'needs_response' ||
                selectedDispute.status === 'warning_needs_response' ||
                selectedDispute.status === 'needs_review') && (
                <button
                  onClick={() => {
                    submitEvidenceMutation.mutate(selectedDispute.id);
                    setShowEvidenceModal(false);
                  }}
                  disabled={submitEvidenceMutation.isPending}
                  className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 disabled:opacity-50"
                >
                  Submit Evidence to Stripe
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function StatCard({
  label,
  value,
  icon: Icon,
  color,
}: {
  label: string;
  value: string | number;
  icon: React.ComponentType<{ className?: string }>;
  color: string;
}) {
  const colorClasses: Record<string, string> = {
    gray: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300',
    yellow: 'bg-yellow-100 text-yellow-600 dark:bg-yellow-900 dark:text-yellow-300',
    emerald: 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900 dark:text-emerald-300',
    red: 'bg-red-100 text-red-600 dark:bg-red-900 dark:text-red-300',
    blue: 'bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-300',
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <div className="flex items-center gap-3">
        <div className={`p-2 rounded-lg ${colorClasses[color]}`}>
          <Icon className="w-5 h-5" />
        </div>
        <div>
          <p className="text-sm text-gray-500 dark:text-gray-400">{label}</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{value}</p>
        </div>
      </div>
    </div>
  );
}

function EvidenceSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{title}</h3>
      {children}
    </div>
  );
}
