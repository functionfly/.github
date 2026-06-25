import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { format } from 'date-fns';
import {
  Check,
  X,
  AlertTriangle,
  Clock,
  DollarSign,
  RefreshCw,
  Eye,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { toast } from 'sonner';

interface PayoutApprovalRecord {
  id: string;
  payout_request_id: string;
  status: 'pending' | 'first_approved' | 'fully_approved' | 'rejected' | 'processing' | 'completed' | 'failed';
  amount_usd: number;
  submitted_by: string;
  submitted_at: string;
  first_approved_by?: string;
  first_approved_at?: string;
  second_approved_by?: string;
  second_approved_at?: string;
  rejected_by?: string;
  rejected_at?: string;
  rejection_reason?: string;
  approval_notes?: string;
}

interface PayoutApprovalSummary {
  total_pending: number;
  total_first_approved: number;
  total_fully_approved: number;
  total_rejected: number;
  total_amount_pending_usd: number;
  requires_attention: number;
}

interface PayoutApprovalRule {
  id: string;
  name: string;
  description: string;
  min_amount_usd: number;
  max_amount_usd?: number;
  requires_first_approval: boolean;
  requires_second_approval: boolean;
  first_approver_roles: string[];
  second_approver_roles: string[];
  is_active: boolean;
  priority: number;
}

interface ApprovalResponse {
  payouts: PayoutApprovalRecord[];
  summary: PayoutApprovalSummary;
  total: number;
  limit: number;
  offset: number;
}

interface ApproveResponse {
  success: boolean;
  payout_id: string;
  approved_by: string;
  approved_at: string;
  notes: string;
  fully_approved: boolean;
  processed: boolean;
}

interface RejectResponse {
  success: boolean;
  payout_id: string;
  rejected_by: string;
  rejected_at: string;
  reason: string;
}

export function AdminPayoutApprovalManager() {
  const [expandedPayoutId, setExpandedPayoutId] = useState<string | null>(null);
  const [rejectReason, setRejectReason] = useState<{ [payoutId: string]: string }>({});
  const [showRejectDialog, setShowRejectDialog] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery<ApprovalResponse>({
    queryKey: ['admin-payout-approvals'],
    queryFn: async () => {
      const response = await adminApiClient.get<ApprovalResponse>('/payouts/pending');
      return response.data;
    },
    refetchInterval: 30000,
  });

  const approveMutation = useMutation({
    mutationFn: async ({ payoutId, notes }: { payoutId: string; notes?: string }) => {
      const response = await adminApiClient.post<ApproveResponse>(`/payouts/${payoutId}/approve`, {
        notes: notes || '',
      });
      return response.data;
    },
    onSuccess: (data) => {
      toast.success(`Payout ${data.fully_approved ? 'fully approved and processed' : 'first-approved'}`);
      queryClient.invalidateQueries({ queryKey: ['admin-payout-approvals'] });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Failed to approve payout');
    },
  });

  const rejectMutation = useMutation({
    mutationFn: async ({ payoutId, reason }: { payoutId: string; reason: string }) => {
      const response = await adminApiClient.post<RejectResponse>(`/payouts/${payoutId}/reject`, {
        reason,
      });
      return response.data;
    },
    onSuccess: () => {
      toast.success('Payout rejected');
      setShowRejectDialog(null);
      setRejectReason({});
      queryClient.invalidateQueries({ queryKey: ['admin-payout-approvals'] });
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Failed to reject payout');
    },
  });

  const { data: rulesData } = useQuery<{ rules: PayoutApprovalRule[] }>({
    queryKey: ['admin-payout-approval-rules'],
    queryFn: async () => {
      const response = await adminApiClient.get<{ rules: PayoutApprovalRule[] }>('/payouts/approval-rules');
      return response.data;
    },
  });

  if (isLoading) return <LoadingScreen />;

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        <h3 className="font-semibold text-red-900">Error loading payout approvals</h3>
        <p className="text-red-700 mt-1">Failed to fetch approval data.</p>
      </div>
    );
  }

  const payouts = data?.payouts || [];
  const summary = data?.summary;

  const statusVariant = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300';
      case 'first_approved':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
      case 'fully_approved':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
      case 'rejected':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300';
    }
  };

  const formatStatus = (status: string) => {
    return status.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase());
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">Payout Approvals</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Review and approve high-value payout requests
          </p>
        </div>
        <button
          onClick={() => queryClient.invalidateQueries({ queryKey: ['admin-payout-approvals'] })}
          className="flex items-center gap-2 px-3 py-2 text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {summary && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Pending Approval</p>
                <p className="text-2xl font-bold text-yellow-600">{summary.total_pending}</p>
              </div>
              <Clock className="w-8 h-8 text-yellow-500" />
            </div>
          </div>
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Awaiting 2nd</p>
                <p className="text-2xl font-bold text-blue-600">{summary.total_first_approved}</p>
              </div>
              <AlertTriangle className="w-8 h-8 text-blue-500" />
            </div>
          </div>
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Pending Amount</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                  ${summary.total_amount_pending_usd.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </p>
              </div>
              <DollarSign className="w-8 h-8 text-green-500" />
            </div>
          </div>
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500 dark:text-gray-400">Requires Action</p>
                <p className="text-2xl font-bold text-red-600">{summary.requires_attention}</p>
              </div>
              <AlertTriangle className="w-8 h-8 text-red-500" />
            </div>
          </div>
        </div>
      )}

      {rulesData?.rules && rulesData.rules.length > 0 && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <h3 className="font-medium text-blue-900 dark:text-blue-300 mb-2">Approval Rules</h3>
          <div className="space-y-2">
            {rulesData.rules.map((rule) => (
              <div key={rule.id} className="text-sm">
                <span className="font-medium">
                  ${rule.min_amount_usd.toLocaleString()}
                  {rule.max_amount_usd && ` - $${rule.max_amount_usd.toLocaleString()}`}
                </span>
                <span className="text-blue-700 dark:text-blue-400">
                  {' '}
                  {rule.requires_second_approval ? '→ Requires 2 approvals' : '→ Requires 1 approval'}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="font-medium text-gray-900 dark:text-gray-100">
            Pending Payouts ({payouts.length})
          </h3>
        </div>

        {payouts.length === 0 ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">
            No payouts awaiting approval
          </div>
        ) : (
          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {payouts.map((payout) => (
              <div key={payout.id} className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <span className={`px-2 py-1 text-xs rounded font-medium ${statusVariant(payout.status)}`}>
                        {formatStatus(payout.status)}
                      </span>
                      <span className="font-mono text-sm text-gray-900 dark:text-gray-100">
                        ${payout.amount_usd.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                      </span>
                      <span className="text-sm text-gray-500">
                        by {payout.submitted_by.slice(0, 8)}...
                      </span>
                      <span className="text-xs text-gray-400">
                        {format(new Date(payout.submitted_at), 'MMM d, yyyy HH:mm')}
                      </span>
                    </div>
                    {payout.status === 'first_approved' && payout.first_approved_by && (
                      <p className="text-xs text-blue-600 mt-1">
                        First approved by {payout.first_approved_by.slice(0, 8)}... at{' '}
                        {payout.first_approved_at && format(new Date(payout.first_approved_at), 'HH:mm')}
                      </p>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      onClick={() =>
                        setExpandedPayoutId(expandedPayoutId === payout.id ? null : payout.id)
                      }
                      className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    >
                      {expandedPayoutId === payout.id ? (
                        <ChevronUp className="w-4 h-4" />
                      ) : (
                        <ChevronDown className="w-4 h-4" />
                      )}
                    </button>

                    {(payout.status === 'pending' || payout.status === 'first_approved') && (
                      <>
                        <button
                          onClick={() =>
                            approveMutation.mutate({
                              payoutId: payout.payout_request_id,
                              notes: '',
                            })
                          }
                          disabled={approveMutation.isPending}
                          className="flex items-center gap-1 px-3 py-1.5 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
                        >
                          <Check className="w-4 h-4" />
                          {payout.status === 'pending' ? 'Approve' : 'Final Approve'}
                        </button>
                        <button
                          onClick={() => setShowRejectDialog(payout.id)}
                          className="flex items-center gap-1 px-3 py-1.5 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                        >
                          <X className="w-4 h-4" />
                          Reject
                        </button>
                      </>
                    )}
                  </div>
                </div>

                {expandedPayoutId === payout.id && (
                  <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <p className="text-gray-500 dark:text-gray-400">Payout Request ID</p>
                        <p className="font-mono text-gray-900 dark:text-gray-100">{payout.payout_request_id}</p>
                      </div>
                      <div>
                        <p className="text-gray-500 dark:text-gray-400">Approval Record ID</p>
                        <p className="font-mono text-gray-900 dark:text-gray-100">{payout.id}</p>
                      </div>
                      <div>
                        <p className="text-gray-500 dark:text-gray-400">Submitted</p>
                        <p className="text-gray-900 dark:text-gray-100">
                          {format(new Date(payout.submitted_at), 'PPpp')}
                        </p>
                      </div>
                      {payout.first_approved_at && (
                        <div>
                          <p className="text-gray-500 dark:text-gray-400">First Approved</p>
                          <p className="text-gray-900 dark:text-gray-100">
                            {format(new Date(payout.first_approved_at), 'PPpp')}
                          </p>
                        </div>
                      )}
                      {payout.approval_notes && (
                        <div className="col-span-2">
                          <p className="text-gray-500 dark:text-gray-400">Notes</p>
                          <p className="text-gray-900 dark:text-gray-100">{payout.approval_notes}</p>
                        </div>
                      )}
                    </div>
                  </div>
                )}

                {showRejectDialog === payout.id && (
                  <div className="mt-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
                    <h4 className="font-medium text-red-900 dark:text-red-300 mb-2">Reject Payout</h4>
                    <textarea
                      value={rejectReason[payout.id] || ''}
                      onChange={(e) =>
                        setRejectReason({ ...rejectReason, [payout.id]: e.target.value })
                      }
                      placeholder="Reason for rejection (required)"
                      className="w-full px-3 py-2 border border-red-300 dark:border-red-700 rounded-lg bg-white dark:bg-red-950 text-gray-900 dark:text-gray-100 mb-2"
                      rows={2}
                    />
                    <div className="flex gap-2">
                      <button
                        onClick={() => {
                          if (!rejectReason[payout.id]?.trim()) {
                            toast.error('Please provide a rejection reason');
                            return;
                          }
                          rejectMutation.mutate({
                            payoutId: payout.payout_request_id,
                            reason: rejectReason[payout.id],
                          });
                        }}
                        disabled={rejectMutation.isPending}
                        className="px-3 py-1.5 text-sm bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50"
                      >
                        Confirm Reject
                      </button>
                      <button
                        onClick={() => {
                          setShowRejectDialog(null);
                          setRejectReason({ ...rejectReason, [payout.id]: '' });
                        }}
                        className="px-3 py-1.5 text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
