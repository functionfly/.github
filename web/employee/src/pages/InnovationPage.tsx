import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { innovationApi, type InnovationGrant } from '@/api/innovation';
import { Lightbulb, Plus, ThumbsUp, ThumbsDown, Send, Check, X, DollarSign, Filter } from 'lucide-react';

const statusColors: Record<string, string> = {
  draft: 'bg-gray-500/20 text-gray-400',
  submitted: 'bg-blue-500/20 text-blue-400',
  under_review: 'bg-yellow-500/20 text-yellow-400',
  approved: 'bg-green-500/20 text-green-400',
  rejected: 'bg-red-500/20 text-red-400',
  funded: 'bg-purple-500/20 text-purple-400',
};

const categories = ['process_improvement', 'technology', 'sustainability', 'cost_saving', 'employee_wellness', 'other'];

export function InnovationPage() {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showVote, setShowVote] = useState<{ id: string; vote: boolean } | null>(null);
  const [voteComment, setVoteComment] = useState('');
  const [form, setForm] = useState({ title: '', description: '', category: 'process_improvement', requested_amount_cents: '' });

  const { data, isLoading } = useQuery({
    queryKey: ['innovation', 'grants', statusFilter],
    queryFn: () => innovationApi.list(statusFilter ? { status: statusFilter } : undefined),
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<InnovationGrant>) => innovationApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['innovation'] });
      setShowCreate(false);
      setForm({ title: '', description: '', category: 'process_improvement', requested_amount_cents: '' });
    },
  });

  const submitMutation = useMutation({
    mutationFn: (id: string) => innovationApi.submit(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['innovation'] }),
  });

  const voteMutation = useMutation({
    mutationFn: ({ id, vote, comment }: { id: string; vote: boolean; comment?: string }) => innovationApi.vote(id, vote, comment),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['innovation'] });
      setShowVote(null);
      setVoteComment('');
    },
  });

  const grants = data?.data?.grants || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Innovation Grants</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          New Proposal
        </button>
      </div>

      <div className="flex items-center gap-3">
        <Filter className="h-4 w-4 text-gray-400" />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Statuses</option>
          <option value="draft">Draft</option>
          <option value="submitted">Submitted</option>
          <option value="under_review">Under Review</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
          <option value="funded">Funded</option>
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : grants.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Lightbulb className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No innovation grants yet</p>
        </div>
      ) : (
        <div className="space-y-4">
          {grants.map((grant) => (
            <div key={grant.id} className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-3 flex items-start justify-between">
                <div>
                  <h3 className="font-semibold text-gray-100">{grant.title}</h3>
                  <p className="mt-1 text-sm text-gray-400">{grant.description}</p>
                </div>
                <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${statusColors[grant.status] || ''}`}>
                  {grant.status.replace('_', ' ')}
                </span>
              </div>

              <div className="mb-4 flex flex-wrap items-center gap-4 text-xs text-gray-500">
                <span className="rounded bg-gray-800 px-2 py-1 capitalize">{grant.category.replace('_', ' ')}</span>
                {grant.requested_amount_cents != null && (
                  <span className="flex items-center gap-1">
                    <DollarSign className="h-3 w-3" />
                    {(grant.requested_amount_cents / 100).toLocaleString('en-US', { style: 'currency', currency: 'USD' })}
                  </span>
                )}
                {grant.feasibility_score != null && (
                  <span>Feasibility: {grant.feasibility_score}/100</span>
                )}
                <span className="flex items-center gap-2">
                  <span className="flex items-center gap-1 text-green-400">
                    <ThumbsUp className="h-3 w-3" /> {grant.votes_for}
                  </span>
                  <span className="flex items-center gap-1 text-red-400">
                    <ThumbsDown className="h-3 w-3" /> {grant.votes_against}
                  </span>
                </span>
              </div>

              <div className="flex gap-2">
                {grant.status === 'draft' && (
                  <button
                    onClick={() => submitMutation.mutate(grant.id)}
                    className="flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
                  >
                    <Send className="h-3 w-3" />
                    Submit for Review
                  </button>
                )}
                {(grant.status === 'submitted' || grant.status === 'under_review') && (
                  <>
                    <button
                      onClick={() => setShowVote({ id: grant.id, vote: true })}
                      className="flex items-center gap-1.5 rounded-lg border border-green-600/30 bg-green-600/10 px-3 py-1.5 text-xs font-medium text-green-400 hover:bg-green-600/20"
                    >
                      <ThumbsUp className="h-3 w-3" />
                      Vote For
                    </button>
                    <button
                      onClick={() => setShowVote({ id: grant.id, vote: false })}
                      className="flex items-center gap-1.5 rounded-lg border border-red-600/30 bg-red-600/10 px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-600/20"
                    >
                      <ThumbsDown className="h-3 w-3" />
                      Vote Against
                    </button>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">New Innovation Proposal</h2>
            <input
              type="text"
              placeholder="Title"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <textarea
              placeholder="Description"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={4}
            />
            <select
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              {categories.map((c) => (
                <option key={c} value={c}>{c.replace('_', ' ')}</option>
              ))}
            </select>
            <input
              type="number"
              placeholder="Requested amount (cents, optional)"
              value={form.requested_amount_cents}
              onChange={(e) => setForm({ ...form, requested_amount_cents: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowCreate(false)}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => createMutation.mutate({
                  title: form.title,
                  description: form.description,
                  category: form.category,
                  requested_amount_cents: form.requested_amount_cents ? parseInt(form.requested_amount_cents) : undefined,
                })}
                disabled={!form.title.trim() || !form.description.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {showVote && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">
              {showVote.vote ? 'Vote For' : 'Vote Against'}
            </h2>
            <textarea
              placeholder="Comment (optional)"
              value={voteComment}
              onChange={(e) => setVoteComment(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => { setShowVote(null); setVoteComment(''); }}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => voteMutation.mutate({ id: showVote.id, vote: showVote.vote, comment: voteComment || undefined })}
                className={`rounded-lg px-4 py-2 text-sm font-medium text-white ${showVote.vote ? 'bg-green-600 hover:bg-green-700' : 'bg-red-600 hover:bg-red-700'}`}
              >
                Submit Vote
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
