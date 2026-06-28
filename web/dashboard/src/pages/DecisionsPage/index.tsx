import type { CreateDecisionRequest, DecisionStatus, TeamDecision, UpdateDecisionRequest } from '@/api/decisions';
import { DecisionCard, DecisionDetail, DecisionForm } from '@/components/decisions';
import {
  Chamber,
  PageGrid,
  SealedButton,
  FrameButton,
  Input,
} from '@/components/containment';
import { toast } from 'sonner';
import {
  useTeams,
  useDecisions,
  useCreateDecision,
  useUpdateDecision,
  useDeleteDecision,
  useApproveDecision,
} from '@/hooks';
import {
  AlertCircle,
  CheckCircle,
  Clock,
  Grid,
  List,
  Plus,
  Search,
  XCircle,
} from 'lucide-react';
import { useMemo, useState } from 'react';

import './decisions-page.css';

const statusToPill = (s: string): 'live' | 'pending' | 'revoked' => {
  if (s === 'approved') return 'live';
  if (s === 'pending') return 'pending';
  return 'revoked';
};

export default function DecisionsPage() {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<DecisionStatus | 'all'>('all');
  const [teamFilter, setTeamFilter] = useState<string>('all');
  const [tagFilter, setTagFilter] = useState('');

  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedDecision, setSelectedDecision] = useState<TeamDecision | null>(null);
  const [selectedTeamForCreate, setSelectedTeamForCreate] = useState('');

  const { data: teamsData } = useTeams();
  const {
    data: decisionsData,
    isLoading: isDecisionsLoading,
    error: decisionsError,
  } = useDecisions(
    teamFilter !== 'all' ? teamFilter : '',
    { status: statusFilter !== 'all' ? statusFilter : undefined, tag: tagFilter || undefined }
  );

  const createMutation = useCreateDecision();
  const updateMutation = useUpdateDecision();
  const deleteMutation = useDeleteDecision();
  const approveMutation = useApproveDecision();

  const teams = teamsData?.teams ?? [];
  const decisions = decisionsData?.decisions || [];

  const filteredDecisions = useMemo(() => {
    if (!searchQuery.trim()) return decisions;
    const q = searchQuery.toLowerCase();
    return decisions.filter(
      (d) =>
        d.title.toLowerCase().includes(q) ||
        d.description?.toLowerCase().includes(q) ||
        d.rationale?.toLowerCase().includes(q) ||
        d.outcome?.toLowerCase().includes(q) ||
        d.tags?.some((t) => t.toLowerCase().includes(q))
    );
  }, [decisions, searchQuery]);

  const statusCounts = useMemo(() => {
    const counts = { pending: 0, approved: 0, superseded: 0, deprecated: 0 };
    decisions.forEach((d) => { if (counts[d.status] !== undefined) counts[d.status]++; });
    return counts;
  }, [decisions]);

  const handleCreateDecision = () => {
    if (teams.length === 0) { toast.error('You need to create or join a team first'); return; }
    if (teams.length === 1) setSelectedTeamForCreate(teams[0].id);
    setCreateDialogOpen(true);
  };

  const handleDeleteDecision = (decisionId: string) => {
    const decision = decisions.find((d) => d.id === decisionId);
    if (decision) {
      deleteMutation.mutate(
        { teamId: decision.team_id, decisionId: decision.id },
        { onSuccess: () => { setDetailDialogOpen(false); setSelectedDecision(null); } }
      );
    }
  };

  const handleApproveDecision = (decision: TeamDecision) => {
    const action = window.confirm('Do you want to approve this decision?\n\nClick "Cancel" to mark as superseded or "OK" to approve.');
    const status: 'approved' | 'superseded' | 'deprecated' = action ? 'approved' : 'superseded';
    approveMutation.mutate(
      { teamId: decision.team_id, decisionId: decision.id, status },
      { onSuccess: () => setDetailDialogOpen(false) }
    );
  };

  const handleUpdateDecision = (data: UpdateDecisionRequest) => {
    if (!selectedDecision) return;
    updateMutation.mutate(
      { teamId: selectedDecision.team_id, decisionId: selectedDecision.id, data },
      { onSuccess: () => { setEditDialogOpen(false); setSelectedDecision(null); } }
    );
  };

  return (
    <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <PageGrid />

      {/* Header */}
      <div className="decisions-header">
        <div>
          <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, letterSpacing: '-0.005em', color: 'var(--text)', margin: 0 }}>Decision Recorder</h1>
          <p style={{ fontSize: 14, color: 'var(--text-dim)', marginTop: 'var(--space-2)', maxWidth: 600, lineHeight: 1.5 }}>
            Track team decisions, rationale, and approvals. Never wonder "why did we choose X?" again.
          </p>
        </div>
        <SealedButton onClick={handleCreateDecision} iconLeft={<Plus style={{ width: 14, height: 14 }} />}>
          Record Decision
        </SealedButton>
      </div>

      {/* Status Summary */}
      <div className="decisions-status-grid">
        {([
          { key: 'pending' as const, icon: Clock, label: 'Pending' },
          { key: 'approved' as const, icon: CheckCircle, label: 'Approved' },
          { key: 'superseded' as const, icon: AlertCircle, label: 'Superseded' },
          { key: 'deprecated' as const, icon: XCircle, label: 'Deprecated' },
        ]).map(({ key, icon: Icon, label }) => (
          <Chamber
            nested
            key={key}
            className={`decisions-status-card ${key} ${statusFilter === key ? 'active' : ''}`}
            onClick={() => setStatusFilter(statusFilter === key ? 'all' : key)}
            style={{ cursor: 'pointer' }}
          >
            <div className="decisions-status-icon">
              <Icon style={{ width: 22, height: 22 }} />
            </div>
            <p style={{ fontFamily: 'var(--font-mono)', fontSize: 26, fontWeight: 700, color: 'var(--text)', marginBottom: 'var(--space-1)' }}>{statusCounts[key]}</p>
            <p style={{ fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)' }}>{label}</p>
          </Chamber>
        ))}
      </div>

      {/* Filters */}
      <div className="decisions-filters">
        <div style={{ position: 'relative', flex: 1, maxWidth: 320 }}>
          <Search style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', width: 14, height: 14, color: 'var(--text-faint)', pointerEvents: 'none' }} />
          <input
            type="text"
            placeholder="Search decisions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input"
            style={{ paddingLeft: 36 }}
          />
        </div>

        <select value={teamFilter} onChange={(e) => setTeamFilter(e.target.value)} className="input" style={{ minWidth: 160, cursor: 'pointer', appearance: 'none' }}>
          <option value="all">All Teams</option>
          {teams.map((team) => <option key={team.id} value={team.id}>{team.name}</option>)}
        </select>

        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value as DecisionStatus | 'all')} className="input" style={{ minWidth: 160, cursor: 'pointer', appearance: 'none' }}>
          <option value="all">All Statuses</option>
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="superseded">Superseded</option>
          <option value="deprecated">Deprecated</option>
        </select>

        <div style={{ display: 'flex', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', overflow: 'hidden' }}>
          {([['grid', Grid], ['list', List]] as const).map(([mode, Icon]) => (
            <button
              key={mode}
              onClick={() => setViewMode(mode)}
              style={{
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                width: 36, height: 36, background: viewMode === mode ? 'var(--panel-raised)' : 'transparent',
                border: 'none', cursor: 'pointer',
                color: viewMode === mode ? 'var(--status-ok)' : 'var(--text-faint)',
                transition: 'all var(--duration-fast) var(--ease-out)',
              }}
            >
              <Icon style={{ width: 14, height: 14 }} />
            </button>
          ))}
        </div>
      </div>

      {/* Content */}
      {teamFilter === 'all' ? (
        <Chamber>
          <div style={{ textAlign: 'center', padding: 'var(--space-8) 0' }}>
            <div style={{ width: 64, height: 64, margin: '0 auto var(--space-4)', borderRadius: 'var(--radius-lg)', background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <AlertCircle style={{ width: 32, height: 32, color: 'var(--text-faint)' }} />
            </div>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Select a Team</h3>
            <p style={{ fontSize: 14, color: 'var(--text-dim)', maxWidth: 500, margin: '0 auto', lineHeight: 1.5 }}>
              Choose a team above to view their decisions. Each team maintains its own decision history with full context and approval tracking.
            </p>
          </div>
        </Chamber>
      ) : isDecisionsLoading ? (
        <Chamber>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 'var(--space-8) 0' }}>
            <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)' }}>Loading decisions...</div>
          </div>
        </Chamber>
      ) : decisionsError ? (
        <Chamber>
          <div style={{ textAlign: 'center', padding: 'var(--space-8) 0' }}>
            <div style={{ width: 64, height: 64, margin: '0 auto var(--space-4)', borderRadius: 'var(--radius-lg)', background: 'rgba(255, 107, 107, 0.1)', border: '1px solid rgba(255, 107, 107, 0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <AlertCircle style={{ width: 32, height: 32, color: 'var(--status-revoked)' }} />
            </div>
            <p style={{ fontSize: 14, color: 'var(--status-revoked)' }}>Failed to load decisions. Please try again.</p>
          </div>
        </Chamber>
      ) : filteredDecisions.length === 0 ? (
        <Chamber>
          <div style={{ textAlign: 'center', padding: 'var(--space-8) 0' }}>
            <div style={{ width: 64, height: 64, margin: '0 auto var(--space-4)', borderRadius: 'var(--radius-lg)', background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <CheckCircle style={{ width: 32, height: 32, color: 'var(--status-ok)' }} />
            </div>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>No decisions yet</h3>
            <p style={{ fontSize: 14, color: 'var(--text-dim)', maxWidth: 400, margin: '0 auto var(--space-5)', lineHeight: 1.5 }}>
              {searchQuery || statusFilter !== 'all'
                ? 'No decisions match your filters.'
                : 'Start recording team decisions to build your decision history.'}
            </p>
            <SealedButton onClick={handleCreateDecision} iconLeft={<Plus style={{ width: 14, height: 14 }} />}>
              Record Your First Decision
            </SealedButton>
          </div>
        </Chamber>
      ) : (
        <div className={viewMode === 'grid' ? 'decisions-grid' : 'decisions-list'}>
          {filteredDecisions.map((decision) => (
            <DecisionCard
              key={decision.id}
              decision={decision}
              teamId={decision.team_id}
              onView={(d) => { setSelectedDecision(d); setDetailDialogOpen(true); }}
              onEdit={(d) => { setSelectedDecision(d); setEditDialogOpen(true); }}
              onDelete={handleDeleteDecision}
              onApprove={handleApproveDecision}
            />
          ))}
        </div>
      )}

      {/* Create Dialog */}
      {teams.length > 0 && (
        <DecisionForm
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          onSubmit={(data) =>
            createMutation.mutate(
              { teamId: selectedTeamForCreate || teams[0].id, data: data as CreateDecisionRequest },
              { onSuccess: () => { setCreateDialogOpen(false); setSelectedTeamForCreate(''); } }
            )
          }
          isLoading={createMutation.isPending}
        />
      )}

      {/* Edit Dialog */}
      <DecisionForm
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        onSubmit={handleUpdateDecision}
        decision={selectedDecision}
        isLoading={updateMutation.isPending}
      />

      {/* Detail Dialog */}
      <DecisionDetail
        open={detailDialogOpen}
        onOpenChange={setDetailDialogOpen}
        decision={selectedDecision}
        onEdit={(d) => { setSelectedDecision(d); setEditDialogOpen(true); }}
        onDelete={handleDeleteDecision}
        onApprove={handleApproveDecision}
      />
    </div>
  );
}
