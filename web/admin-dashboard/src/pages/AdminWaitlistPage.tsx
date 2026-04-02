import { CheckCircle, Clock, Filter, Mail, Search, Send, Trash2, Users, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingSpinner } from '../components/ui/LoadingSpinner';
import { adminApiClient as adminClient } from '../lib/api/adminClient';

interface WaitlistEntry {
  id: string;
  email: string;
  name?: string;
  company?: string;
  useCase?: string;
  source: string;
  status: 'pending' | 'approved' | 'rejected' | 'invited';
  inviteCodeId?: string;
  invitedAt?: string;
  notes?: string;
  ip?: string;
  createdAt: string;
  updatedAt: string;
}

interface WaitlistStats {
  total: number;
  pending: number;
  approved: number;
  invited: number;
  rejected: number;
}

export default function AdminWaitlistPage() {
  const [entries, setEntries] = useState<WaitlistEntry[]>([]);
  const [stats, setStats] = useState<WaitlistStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [issuingInvite, setIssuingInvite] = useState<string | null>(null);
  const [issuedCode, setIssuedCode] = useState<string | null>(null);
  const [selectedEntry, setSelectedEntry] = useState<WaitlistEntry | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [entriesRes, statsRes] = await Promise.all([
        adminClient.get(`/waitlist?status=${statusFilter}&limit=100`),
        adminClient.get('/waitlist/stats'),
      ]);

      const entriesPayload = entriesRes.data as
        | WaitlistEntry[]
        | { data?: WaitlistEntry[] }
        | undefined;
      if (Array.isArray(entriesPayload)) {
        setEntries(entriesPayload);
      } else {
        setEntries(entriesPayload?.data || []);
      }

      const statsPayload = statsRes.data as WaitlistStats | { data?: WaitlistStats } | undefined;
      if (statsPayload && 'total' in statsPayload) {
        setStats(statsPayload);
      } else {
        setStats(statsPayload?.data || null);
      }
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch waitlist data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [statusFilter]);

  const handleIssueInvite = async (entryId: string) => {
    try {
      setIssuingInvite(entryId);
      const response = await adminClient.post(`/waitlist/${entryId}/invite`, {});
      const invitePayload =
        (response.data as { inviteCode?: string; data?: { inviteCode?: string } }) || {};
      const inviteCode = invitePayload.inviteCode || invitePayload.data?.inviteCode;

      if (inviteCode) {
        setIssuedCode(inviteCode);
        setSelectedEntry(entries.find((e) => e.id === entryId) || null);
        fetchData();
      }
    } catch (err: any) {
      setError(err.message || 'Failed to issue invite');
    } finally {
      setIssuingInvite(null);
    }
  };

  const handleUpdateStatus = async (entryId: string, status: string) => {
    try {
      await adminClient.patch(`/waitlist/${entryId}/status`, { status });
      fetchData();
    } catch (err: any) {
      setError(err.message || 'Failed to update status');
    }
  };

  const handleDelete = async (entryId: string) => {
    if (!confirm('Are you sure you want to delete this entry?')) return;

    try {
      await adminClient.delete(`/waitlist/${entryId}`);
      fetchData();
    } catch (err: any) {
      setError(err.message || 'Failed to delete entry');
    }
  };

  const filteredEntries = entries.filter(
    (entry) =>
      entry.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      entry.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      entry.company?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return (
          <Badge variant="warning" icon={<Clock className="w-3 h-3" />}>
            Pending
          </Badge>
        );
      case 'approved':
        return (
          <Badge variant="info" icon={<CheckCircle className="w-3 h-3" />}>
            Approved
          </Badge>
        );
      case 'invited':
        return (
          <Badge variant="success" icon={<Mail className="w-3 h-3" />}>
            Invited
          </Badge>
        );
      case 'rejected':
        return (
          <Badge variant="error" icon={<X className="w-3 h-3" />}>
            Rejected
          </Badge>
        );
      default:
        return <Badge>{status}</Badge>;
    }
  };

  if (loading && !entries.length) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner text="Loading waitlist..." />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-gray-900">Waitlist Management</h1>
        <Button variant="outline" onClick={fetchData}>
          Refresh
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">{error}</div>
      )}

      {/* Stats */}
      {stats && (
        <div className="grid grid-cols-5 gap-4">
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Users className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Total</p>
                <p className="text-xl font-semibold text-gray-900">{stats.total}</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-yellow-100 rounded-lg">
                <Clock className="w-5 h-5 text-yellow-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Pending</p>
                <p className="text-xl font-semibold text-gray-900">{stats.pending}</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-cyan-100 rounded-lg">
                <CheckCircle className="w-5 h-5 text-cyan-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Approved</p>
                <p className="text-xl font-semibold text-gray-900">{stats.approved}</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-100 rounded-lg">
                <Mail className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Invited</p>
                <p className="text-xl font-semibold text-gray-900">{stats.invited}</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-red-100 rounded-lg">
                <X className="w-5 h-5 text-red-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Rejected</p>
                <p className="text-xl font-semibold text-gray-900">{stats.rejected}</p>
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search by email, name, or company..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-white border border-gray-300 rounded-lg text-gray-900 placeholder-gray-400 focus:outline-none focus:border-[#6366f1] focus:ring-1 focus:ring-[#6366f1]"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-gray-500" />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 bg-white border border-gray-300 rounded-lg text-gray-900 focus:outline-none focus:border-[#6366f1] focus:ring-1 focus:ring-[#6366f1]"
          >
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="approved">Approved</option>
            <option value="invited">Invited</option>
            <option value="rejected">Rejected</option>
          </select>
        </div>
      </div>

      {/* Entries Table */}
      <Card>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Email</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Name</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Company</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Status</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-700">Date</th>
                <th className="px-4 py-3 text-right text-sm font-medium text-gray-700">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {filteredEntries.map((entry) => (
                <tr key={entry.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm text-gray-900">{entry.email}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{entry.name || '-'}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{entry.company || '-'}</td>
                  <td className="px-4 py-3">{getStatusBadge(entry.status)}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {new Date(entry.createdAt).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      {entry.status === 'pending' && (
                        <>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleUpdateStatus(entry.id, 'approved')}
                          >
                            <CheckCircle className="w-4 h-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleUpdateStatus(entry.id, 'rejected')}
                          >
                            <X className="w-4 h-4" />
                          </Button>
                        </>
                      )}
                      {(entry.status === 'pending' || entry.status === 'approved') && (
                        <Button
                          size="sm"
                          onClick={() => handleIssueInvite(entry.id)}
                          disabled={issuingInvite === entry.id}
                        >
                          {issuingInvite === entry.id ? (
                            <LoadingSpinner size="sm" />
                          ) : (
                            <Send className="w-4 h-4" />
                          )}
                        </Button>
                      )}
                      <Button size="sm" variant="ghost" onClick={() => handleDelete(entry.id)}>
                        <Trash2 className="w-4 h-4 text-red-500" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredEntries.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <Users className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>No waitlist entries found</p>
          </div>
        )}
      </Card>

      {/* Invite Code Modal */}
      {issuedCode && selectedEntry && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md p-6 m-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Invite Code Generated</h3>
            <p className="text-gray-600 mb-4">
              Send this invite code to{' '}
              <strong className="text-gray-900">{selectedEntry.email}</strong>:
            </p>
            <div className="bg-gray-50 border border-[#6366f1]/30 rounded-lg p-4 mb-4">
              <code className="text-2xl font-mono text-[#6366f1] tracking-wider">{issuedCode}</code>
            </div>
            <p className="text-sm text-gray-500 mb-6">
              This code can only be viewed once. Make sure to copy and send it to the user.
            </p>
            <div className="flex justify-end gap-3">
              <Button
                variant="outline"
                onClick={() => {
                  setIssuedCode(null);
                  setSelectedEntry(null);
                }}
              >
                Close
              </Button>
              <Button
                onClick={() => {
                  navigator.clipboard.writeText(issuedCode);
                }}
              >
                Copy to Clipboard
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}
