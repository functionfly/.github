/**
 * Admin Newsletter Page
 * Manage marketing newsletters, campaigns, and subscriber lists.
 */

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Plus, Send, Users, Eye, Archive, Trash2, Search, Filter, Download, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Dialog, DialogContent, DialogHeader, DialogFooter } from '@/components/ui/Dialog';
import { DataTable, TableBadge } from '@/components/ui/DataTable';
import { Card } from '@/components/ui/Card';
import { useToastHelpers } from '@/components/ui/Toast';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface Newsletter {
  id: string;
  subject: string;
  status: 'draft' | 'scheduled' | 'sending' | 'sent' | 'archived';
  recipient_count: number;
  open_rate?: number;
  click_rate?: number;
  sent_at?: string;
  scheduled_at?: string;
  created_at: string;
  author: string;
}

interface NewsletterStats {
  total_sent: number;
  total_subscribers: number;
  avg_open_rate: number;
  avg_click_rate: number;
}

export function AdminNewsletterPage() {
  const queryClient = useQueryClient();
  const toast = useToastHelpers();
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isPreviewDialogOpen, setIsPreviewDialogOpen] = useState(false);
  const [selectedNewsletter, setSelectedNewsletter] = useState<Newsletter | null>(null);
  const [sortColumn, setSortColumn] = useState<string>('created_at');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

  // Form state
  const [newNewsletter, setNewNewsletter] = useState({
    subject: '',
    content: '',
    recipient_segment: 'all',
    scheduled_at: '',
  });

  // Fetch newsletters
  const { data: newslettersResponse, isLoading, isError, refetch } = useQuery({
    queryKey: ['admin-newsletters', sortColumn, sortDirection],
    queryFn: async () => {
      try {
        const res = await adminApiClient.get<Newsletter[]>('/content/newsletters', {
          params: { sort: sortColumn, order: sortDirection },
        });
        return res;
      } catch {
        toast.error('Failed to load newsletters');
        return { data: [] as Newsletter[] };
      }
    },
  });

  // Fetch stats
  const { data: statsResponse } = useQuery({
    queryKey: ['admin-newsletter-stats'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<NewsletterStats>('/content/newsletters/stats');
      } catch {
        return { data: null };
      }
    },
  });

  const newsletters = (newslettersResponse as { data?: Newsletter[] })?.data ?? [];
  const stats = (statsResponse as { data?: NewsletterStats })?.data;

  // Create newsletter mutation
  const createMutation = useMutation({
    mutationFn: (data: typeof newNewsletter) =>
      adminApiClient.post('/content/newsletters', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-newsletters'] });
      setNewNewsletter({ subject: '', content: '', recipient_segment: 'all', scheduled_at: '' });
      setIsCreateDialogOpen(false);
      toast.success('Newsletter created successfully');
    },
    onError: () => {
      toast.error('Failed to create newsletter');
    },
  });

  // Send newsletter mutation
  const sendMutation = useMutation({
    mutationFn: (id: string) =>
      adminApiClient.post(`/content/newsletters/${id}/send`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-newsletters'] });
      toast.success('Newsletter sent successfully');
    },
    onError: () => {
      toast.error('Failed to send newsletter');
    },
  });

  // Archive newsletter mutation
  const archiveMutation = useMutation({
    mutationFn: (id: string) =>
      adminApiClient.patch(`/content/newsletters/${id}`, { status: 'archived' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-newsletters'] });
      toast.success('Newsletter archived');
    },
  });

  // Delete newsletter mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      adminApiClient.delete(`/content/newsletters/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-newsletters'] });
      toast.success('Newsletter deleted');
    },
  });

  const handleSort = (column: string) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  const handleCreate = () => {
    if (!newNewsletter.subject.trim() || !newNewsletter.content.trim()) {
      toast.error('Subject and content are required');
      return;
    }
    createMutation.mutate(newNewsletter);
  };

  const handlePreview = (newsletter: Newsletter) => {
    setSelectedNewsletter(newsletter);
    setIsPreviewDialogOpen(true);
  };

  const filteredNewsletters = newsletters.filter((newsletter) => {
    const matchesSearch =
      newsletter.subject.toLowerCase().includes(searchTerm.toLowerCase()) ||
      newsletter.author.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === 'all' || newsletter.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const statusOptions = [
    { value: 'all', label: 'All Statuses' },
    { value: 'draft', label: 'Draft' },
    { value: 'scheduled', label: 'Scheduled' },
    { value: 'sending', label: 'Sending' },
    { value: 'sent', label: 'Sent' },
    { value: 'archived', label: 'Archived' },
  ];

  const segmentOptions = [
    { value: 'all', label: 'All Subscribers' },
    { value: 'active', label: 'Active Users' },
    { value: 'admins', label: 'Admins Only' },
    { value: 'beta', label: 'Beta Users' },
  ];

  const getStatusVariant = (status: string): Parameters<typeof TableBadge>[0]['variant'] => {
    switch (status) {
      case 'sent':
        return 'success';
      case 'scheduled':
        return 'info';
      case 'sending':
        return 'warning';
      case 'archived':
        return 'neutral';
      default:
        return 'default';
    }
  };

  const columns = [
    {
      key: 'subject',
      header: 'Subject',
      accessor: (row: Newsletter) => (
        <div>
          <div className="font-medium text-gray-900">{row.subject}</div>
          <div className="text-xs text-gray-500">by {row.author}</div>
        </div>
      ),
      sortable: true,
    },
    {
      key: 'status',
      header: 'Status',
      accessor: (row: Newsletter) => (
        <TableBadge variant={getStatusVariant(row.status)}>
          {row.status}
        </TableBadge>
      ),
      sortable: true,
      width: '120px',
    },
    {
      key: 'recipients',
      header: 'Recipients',
      accessor: (row: Newsletter) => (
        <div className="text-gray-600">
          {row.recipient_count.toLocaleString()}
        </div>
      ),
      sortable: true,
      width: '120px',
    },
    {
      key: 'rates',
      header: 'Performance',
      accessor: (row: Newsletter) => (
        <div className="text-sm">
          {row.open_rate !== undefined ? (
            <div className="flex gap-3">
              <span className="text-green-600">
                <Eye className="w-3 h-3 inline mr-1" />
                {row.open_rate.toFixed(1)}%
              </span>
              <span className="text-blue-600">
                Click: {row.click_rate?.toFixed(1) ?? 0}%
              </span>
            </div>
          ) : (
            <span className="text-gray-400">-</span>
          )}
        </div>
      ),
      width: '180px',
    },
    {
      key: 'date',
      header: 'Date',
      accessor: (row: Newsletter) => (
        <div className="text-sm text-gray-600">
          {row.sent_at
            ? new Date(row.sent_at).toLocaleDateString()
            : row.scheduled_at
            ? `Scheduled: ${new Date(row.scheduled_at).toLocaleDateString()}`
            : new Date(row.created_at).toLocaleDateString()}
        </div>
      ),
      sortable: true,
      width: '150px',
    },
  ];

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <p className="text-red-600 mb-4">Failed to load newsletters</p>
          <Button onClick={() => refetch()} variant="outline">
            <RefreshCw className="w-4 h-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">Newsletters</h1>
          <p className="text-gray-600 dark:text-gray-400">Manage email campaigns and subscriber communications</p>
        </div>
        <Button onClick={() => setIsCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create Newsletter
        </Button>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-4 gap-4">
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <Send className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Total Sent</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.total_sent.toLocaleString()}</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-100 dark:bg-green-900/30 rounded-lg">
                <Users className="w-5 h-5 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Subscribers</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.total_subscribers.toLocaleString()}</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-100 dark:bg-indigo-900/30 rounded-lg">
                <Eye className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Avg Open Rate</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.avg_open_rate.toFixed(1)}%</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                <Filter className="w-5 h-5 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Avg Click Rate</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.avg_click_rate.toFixed(1)}%</p>
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1">
          <Input
            placeholder="Search newsletters..."
            leftIcon={<Search className="w-4 h-4" />}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div className="w-48">
          <Select
            options={statusOptions}
            value={statusFilter}
            onChange={(value) => setStatusFilter(value)}
          />
        </div>
        <Button variant="outline">
          <Download className="w-4 h-4 mr-2" />
          Export
        </Button>
      </div>

      {/* Table */}
      <DataTable
        data={filteredNewsletters}
        columns={columns}
        keyExtractor={(row) => row.id}
        sortColumn={sortColumn}
        sortDirection={sortDirection}
        onSort={handleSort}
        emptyMessage="No newsletters found. Create one to get started."
        rowActions={(row) => (
          <div className="flex gap-2">
            <button
              onClick={() => handlePreview(row)}
              className="p-1 text-gray-500 hover:text-indigo-600 transition-colors"
              title="Preview"
            >
              <Eye className="w-4 h-4" />
            </button>
            {row.status === 'draft' && (
              <button
                onClick={() => sendMutation.mutate(row.id)}
                disabled={sendMutation.isPending}
                className="p-1 text-gray-500 hover:text-green-600 transition-colors"
                title="Send now"
              >
                <Send className="w-4 h-4" />
              </button>
            )}
            {row.status !== 'archived' && (
              <button
                onClick={() => archiveMutation.mutate(row.id)}
                className="p-1 text-gray-500 hover:text-yellow-600 transition-colors"
                title="Archive"
              >
                <Archive className="w-4 h-4" />
              </button>
            )}
            <button
              onClick={() => deleteMutation.mutate(row.id)}
              disabled={deleteMutation.isPending}
              className="p-1 text-gray-500 hover:text-red-600 transition-colors"
              title="Delete"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        )}
      />

      {/* Create Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogContent>
          <DialogHeader
            title="Create Newsletter"
            description="Compose a new email campaign for your subscribers."
            onClose={() => setIsCreateDialogOpen(false)}
          />
          <div className="space-y-4 mt-4">
            <Input
              label="Subject"
              placeholder="Enter newsletter subject..."
              value={newNewsletter.subject}
              onChange={(e) => setNewNewsletter({ ...newNewsletter, subject: e.target.value })}
              required
            />
            <Select
              label="Recipient Segment"
              options={segmentOptions}
              value={newNewsletter.recipient_segment}
              onChange={(value) => setNewNewsletter({ ...newNewsletter, recipient_segment: value })}
            />
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Content</label>
              <textarea
                value={newNewsletter.content}
                onChange={(e) => setNewNewsletter({ ...newNewsletter, content: e.target.value })}
                placeholder="Write your newsletter content..."
                rows={6}
                className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 resize-y bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                required
              />
            </div>
            <Input
              label="Schedule Send (optional)"
              type="datetime-local"
              value={newNewsletter.scheduled_at}
              onChange={(e) => setNewNewsletter({ ...newNewsletter, scheduled_at: e.target.value })}
              helperText="Leave blank to save as draft"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={createMutation.isPending || !newNewsletter.subject.trim()}
            >
              {createMutation.isPending ? 'Creating...' : 'Create Newsletter'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Preview Dialog */}
      <Dialog open={isPreviewDialogOpen} onOpenChange={setIsPreviewDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader
            title={selectedNewsletter?.subject ?? 'Newsletter Preview'}
            onClose={() => setIsPreviewDialogOpen(false)}
          />
          <div className="mt-4">
            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-6 border border-gray-200 dark:border-gray-700">
              <div className="text-sm text-gray-500 dark:text-gray-400 mb-4">
                From: FunctionFly Team &lt;team@functionfly.com&gt;
                <br />
                To: {selectedNewsletter?.recipient_count.toLocaleString() ?? 0} subscribers
              </div>
              <div className="prose max-w-none">
                <p className="text-gray-400 dark:text-gray-500 italic">Newsletter content would be rendered here...</p>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsPreviewDialogOpen(false)}>
              Close
            </Button>
            {selectedNewsletter?.status === 'draft' && (
              <Button onClick={() => selectedNewsletter && sendMutation.mutate(selectedNewsletter.id)}>
                <Send className="w-4 h-4 mr-2" />
                Send Now
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
