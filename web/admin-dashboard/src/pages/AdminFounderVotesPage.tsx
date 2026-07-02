import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import { LoadingSpinner } from '@/components/ui/LoadingSpinner';
import { Dialog, DialogContent, DialogHeader, DialogFooter } from '@/components/ui/Dialog';
import {
  Vote,
  Plus,
  Trash2,
  X,
  CheckCircle2,
  Clock,
  XCircle,
  BarChart3,
} from 'lucide-react';
import { useToastHelpers } from '@/components/ui/Toast';
import type { AdminAPIResponse } from '@/types';

interface VoteOption {
  id: string;
  label: string;
}

interface ChangeDiff {
  summary?: string;
  changes?: { field: string; label?: string; before?: string; after?: string; type?: string }[];
  impact?: string;
  rationale?: string;
  category?: string;
}

interface FounderVote {
  id: string;
  title: string;
  description: string;
  options: VoteOption[];
  change_diff?: ChangeDiff;
  status: string;
  quorum: number;
  total_votes?: number;
  results?: Record<string, number>;
  created_at: string;
  updated_at: string;
}

interface VotesResponse {
  votes: FounderVote[];
  total: number;
}

const STATUS_COLORS: Record<string, string> = {
  active: 'bg-green-500/10 text-green-400 border-green-500/20',
  closed: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20',
  passed: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  rejected: 'bg-red-500/10 text-red-400 border-red-500/20',
};

function StatusIcon({ status }: { status: string }) {
  switch (status) {
    case 'active':
      return <Clock size={14} className="text-green-400" />;
    case 'passed':
      return <CheckCircle2 size={14} className="text-blue-400" />;
    case 'rejected':
      return <XCircle size={14} className="text-red-400" />;
    default:
      return <XCircle size={14} className="text-zinc-400" />;
  }
}

export function AdminFounderVotesPage() {
  const queryClient = useQueryClient();
  const { success, error: showError } = useToastHelpers();
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const { data: votesResponse, isLoading } = useQuery<AdminAPIResponse<VotesResponse>>({
    queryKey: ['admin-founder-votes'],
    queryFn: () => adminApiClient.get<VotesResponse>('/founders/votes'),
  });

  const createMutation = useMutation({
    mutationFn: (body: {
      title: string;
      description: string;
      options: VoteOption[];
      change_diff?: ChangeDiff;
      quorum: number;
    }) => adminApiClient.post('/founders/votes', body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-founder-votes'] });
      success('Proposal created');
      setCreateOpen(false);
    },
    onError: (err: Error) => showError(err.message || 'Failed to create proposal'),
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      adminApiClient.patch(`/founders/votes/${id}`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-founder-votes'] });
      success('Proposal updated');
    },
    onError: (err: Error) => showError(err.message || 'Failed to update proposal'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.delete(`/founders/votes/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-founder-votes'] });
      success('Proposal deleted');
      setDeleteId(null);
    },
    onError: (err: Error) => showError(err.message || 'Failed to delete proposal'),
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <LoadingSpinner size="lg" text="Loading proposals..." />
      </div>
    );
  }

  const votes = votesResponse?.data?.votes ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Founder Governance</h1>
          <p className="text-sm text-zinc-400 mt-1">
            Create and manage voting proposals for founders
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)} className="gap-2">
          <Plus size={16} />
          New Proposal
        </Button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
        <Card className="p-4">
          <div className="text-sm text-zinc-400">Total</div>
          <div className="text-2xl font-bold text-white">{votes.length}</div>
        </Card>
        <Card className="p-4">
          <div className="text-sm text-zinc-400">Active</div>
          <div className="text-2xl font-bold text-green-400">
            {votes.filter((v) => v.status === 'active').length}
          </div>
        </Card>
        <Card className="p-4">
          <div className="text-sm text-zinc-400">Passed</div>
          <div className="text-2xl font-bold text-blue-400">
            {votes.filter((v) => v.status === 'passed').length}
          </div>
        </Card>
        <Card className="p-4">
          <div className="text-sm text-zinc-400">Rejected</div>
          <div className="text-2xl font-bold text-red-400">
            {votes.filter((v) => v.status === 'rejected').length}
          </div>
        </Card>
      </div>

      {votes.length === 0 ? (
        <Card className="p-12 text-center">
          <Vote size={32} className="mx-auto text-zinc-500 mb-3" />
          <p className="text-zinc-400">No proposals yet</p>
          <p className="text-sm text-zinc-500 mt-1">
            Create a proposal to start gathering founder votes
          </p>
        </Card>
      ) : (
        <div className="space-y-3">
          {votes.map((vote) => (
            <Card key={vote.id} className="p-4">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <StatusIcon status={vote.status} />
                    <h3 className="font-semibold text-white truncate">{vote.title}</h3>
                    <span
                      className={`text-xs px-2 py-0.5 rounded-full border ${STATUS_COLORS[vote.status] ?? 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20'}`}
                    >
                      {vote.status}
                    </span>
                  </div>
                  {vote.description && (
                    <p className="text-sm text-zinc-400 line-clamp-2 mb-2">
                      {vote.description}
                    </p>
                  )}
                  <div className="flex items-center gap-4 text-xs text-zinc-500">
                    <span className="flex items-center gap-1">
                      <BarChart3 size={12} />
                      {vote.options.length} options
                    </span>
                    {vote.quorum > 0 && <span>Quorum: {vote.quorum}</span>}
                    <span>
                      Created {new Date(vote.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  {vote.status === 'active' && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        updateStatusMutation.mutate({ id: vote.id, status: 'closed' })
                      }
                    >
                      Close
                    </Button>
                  )}
                  {(vote.status === 'closed' || vote.status === 'active') && (
                    <>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          updateStatusMutation.mutate({ id: vote.id, status: 'passed' })
                        }
                      >
                        Pass
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() =>
                          updateStatusMutation.mutate({ id: vote.id, status: 'rejected' })
                        }
                      >
                        Reject
                      </Button>
                    </>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setDeleteId(vote.id)}
                    className="text-red-400 hover:text-red-300"
                  >
                    <Trash2 size={14} />
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      <CreateProposalDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSubmit={(data) => createMutation.mutate(data)}
        isSubmitting={createMutation.isPending}
      />

      <Dialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader title="Delete Proposal" onClose={() => setDeleteId(null)} />
          <p className="text-zinc-400 text-sm">
            Are you sure you want to delete this proposal? This action cannot be undone. All votes
            will be lost.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>
              Cancel
            </Button>
            <Button
              onClick={() => deleteId && deleteMutation.mutate(deleteId)}
              disabled={deleteMutation.isPending}
              className="bg-red-600 hover:bg-red-500"
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreateProposalDialog({
  open,
  onClose,
  onSubmit,
  isSubmitting,
}: {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: {
    title: string;
    description: string;
    options: VoteOption[];
    change_diff?: ChangeDiff;
    quorum: number;
  }) => void;
  isSubmitting: boolean;
}) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [options, setOptions] = useState<VoteOption[]>([
    { id: 'yes', label: 'Yes' },
    { id: 'no', label: 'No' },
  ]);
  const [quorum, setQuorum] = useState(0);
  const [diffSummary, setDiffSummary] = useState('');
  const [diffImpact, setDiffImpact] = useState('');
  const [diffRationale, setDiffRationale] = useState('');
  const [diffCategory, setDiffCategory] = useState('');

  const handleSubmit = () => {
    if (!title.trim() || options.length < 2) return;
    const changeDiff: ChangeDiff = {};
    if (diffSummary) changeDiff.summary = diffSummary;
    if (diffImpact) changeDiff.impact = diffImpact;
    if (diffRationale) changeDiff.rationale = diffRationale;
    if (diffCategory) changeDiff.category = diffCategory;
    onSubmit({
      title: title.trim(),
      description: description.trim(),
      options,
      change_diff: Object.keys(changeDiff).length > 0 ? changeDiff : undefined,
      quorum,
    });
  };

  const addOption = () => {
    setOptions([...options, { id: `opt_${Date.now()}`, label: '' }]);
  };

  const removeOption = (index: number) => {
    if (options.length <= 2) return;
    setOptions(options.filter((_, i) => i !== index));
  };

  const updateOption = (index: number, label: string) => {
    const updated = [...options];
    updated[index] = { ...updated[index], label };
    setOptions(updated);
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader title="Create Proposal" onClose={onClose} />

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">Title *</label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g., Increase free tier function limit"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">Description</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Explain what this proposal is about..."
              className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[80px]"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">
              Voting Options * (min 2)
            </label>
            <div className="space-y-2">
              {options.map((opt, i) => (
                <div key={opt.id} className="flex items-center gap-2">
                  <Input
                    value={opt.label}
                    onChange={(e) => updateOption(i, e.target.value)}
                    placeholder={`Option ${i + 1}`}
                    className="flex-1"
                  />
                  {options.length > 2 && (
                    <button
                      type="button"
                      onClick={() => removeOption(i)}
                      className="text-zinc-500 hover:text-red-400 p-1"
                    >
                      <X size={14} />
                    </button>
                  )}
                </div>
              ))}
            </div>
            <Button variant="ghost" size="sm" onClick={addOption} className="mt-2 gap-1">
              <Plus size={12} /> Add option
            </Button>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">
              Quorum (0 = no minimum)
            </label>
            <Input
              type="number"
              value={quorum}
              onChange={(e) => setQuorum(Math.max(0, parseInt(e.target.value) || 0))}
              min={0}
            />
          </div>

          <div className="border-t border-zinc-700 pt-4">
            <h3 className="text-sm font-medium text-zinc-300 mb-3">
              Proposed Changes (optional)
            </h3>
            <div className="space-y-3">
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Category</label>
                <Input
                  value={diffCategory}
                  onChange={(e) => setDiffCategory(e.target.value)}
                  placeholder="e.g., Pricing, Limits, Features"
                />
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Summary</label>
                <textarea
                  value={diffSummary}
                  onChange={(e) => setDiffSummary(e.target.value)}
                  placeholder="Brief summary of the changes..."
                  className="w-full bg-zinc-800 border border-zinc-700 rounded-md px-3 py-2 text-sm text-white placeholder:text-zinc-500 focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[60px]"
                />
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Expected Impact</label>
                <Input
                  value={diffImpact}
                  onChange={(e) => setDiffImpact(e.target.value)}
                  placeholder="What effect will this have?"
                />
              </div>
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Rationale</label>
                <Input
                  value={diffRationale}
                  onChange={(e) => setDiffRationale(e.target.value)}
                  placeholder="Why is this change needed?"
                />
              </div>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={
              !title.trim() || options.filter((o) => o.label.trim()).length < 2 || isSubmitting
            }
          >
            {isSubmitting ? 'Creating...' : 'Create Proposal'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
