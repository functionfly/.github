import type { CreateDecisionRequest, DecisionStatus, TeamDecision, UpdateDecisionRequest } from '@/api/decisions';
import { DecisionCard, DecisionDetail, DecisionForm } from '@/components/decisions';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  useTeams,
  useDecisions,
  useCreateDecision,
  useUpdateDecision,
  useDeleteDecision,
  useApproveDecision,
} from '@/hooks';
import type { Team } from '@/api/teams';
import {
  AlertCircle,
  CheckCircle,
  Clock,
  Grid,
  List,
  Loader2,
  Plus,
  Search,
  XCircle,
} from 'lucide-react';
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import './decisions-page.css';

export default function DecisionsPage() {
  const navigate = useNavigate();

  // UI State
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<DecisionStatus | 'all'>('all');
  const [teamFilter, setTeamFilter] = useState<string>('all');
  const [tagFilter, setTagFilter] = useState<string>('');

  // Dialog State
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedDecision, setSelectedDecision] = useState<TeamDecision | null>(null);
  const [selectedTeamForCreate, setSelectedTeamForCreate] = useState<string>('');

  // Fetch teams for filter and creation
  const { data: teamsData, isLoading: isTeamsLoading } = useTeams();

  // Fetch decisions for selected team
  const {
    data: decisionsData,
    isLoading: isDecisionsLoading,
    error: decisionsError,
  } = useDecisions(
    teamFilter !== 'all' ? teamFilter : '',
    {
      status: statusFilter !== 'all' ? statusFilter : undefined,
      tag: tagFilter || undefined,
    }
  );

  // Mutations
  const createMutation = useCreateDecision();
  const updateMutation = useUpdateDecision();
  const deleteMutation = useDeleteDecision();
  const approveMutation = useApproveDecision();

  const teams = teamsData?.teams ?? [];
  const decisions = decisionsData?.decisions || [];

  // Filter decisions by search query (client-side)
  const filteredDecisions = useMemo(() => {
    if (!searchQuery.trim()) return decisions;
    const query = searchQuery.toLowerCase();
    return decisions.filter(
      (d) =>
        d.title.toLowerCase().includes(query) ||
        d.description?.toLowerCase().includes(query) ||
        d.rationale?.toLowerCase().includes(query) ||
        d.outcome?.toLowerCase().includes(query) ||
        d.tags?.some((t) => t.toLowerCase().includes(query))
    );
  }, [decisions, searchQuery]);

  // Status counts
  const statusCounts = useMemo(() => {
    const counts = { pending: 0, approved: 0, superseded: 0, deprecated: 0 };
    decisions.forEach((d) => {
      if (counts[d.status] !== undefined) counts[d.status]++;
    });
    return counts;
  }, [decisions]);

  const handleCreateDecision = () => {
    if (teams.length === 0) {
      // Team is required to record decisions - let user know they need a team first
      toast.error('You need to create or join a team first');
      return;
    }
    // If only one team, select it automatically
    if (teams.length === 1) {
      setSelectedTeamForCreate(teams[0].id);
    }
    setCreateDialogOpen(true);
  };

  const handleViewDecision = (decision: TeamDecision) => {
    setSelectedDecision(decision);
    setDetailDialogOpen(true);
  };

  const handleEditDecision = (decision: TeamDecision) => {
    setSelectedDecision(decision);
    setEditDialogOpen(true);
  };

  const handleDeleteDecision = (decisionId: string) => {
    const decision = decisions.find((d) => d.id === decisionId);
    if (decision) {
      deleteMutation.mutate(
        { teamId: decision.team_id, decisionId: decision.id },
        {
          onSuccess: () => {
            setDetailDialogOpen(false);
            setSelectedDecision(null);
          },
        }
      );
    }
  };

  const handleApproveDecision = (decision: TeamDecision) => {
    // Show a simple confirmation dialog
    const action = window.confirm(
      `Do you want to approve this decision?\n\nClick "Cancel" to mark as superseded or "OK" to approve.`
    );
    const status: 'approved' | 'superseded' | 'deprecated' = action ? 'approved' : 'superseded';
    approveMutation.mutate(
      {
        teamId: decision.team_id,
        decisionId: decision.id,
        status,
      },
      {
        onSuccess: () => {
          setDetailDialogOpen(false);
        },
      }
    );
  };

  const handleUpdateDecision = (data: UpdateDecisionRequest) => {
    if (!selectedDecision) return;
    updateMutation.mutate(
      {
        teamId: selectedDecision.team_id,
        decisionId: selectedDecision.id,
        data,
      },
      {
        onSuccess: () => {
          setEditDialogOpen(false);
          setSelectedDecision(null);
        },
      }
    );
  };

  return (
    <div className="decisions-page container mx-auto">
      {/* Header */}
      <div className="decisions-header">
        <div className="decisions-header-content">
          <h1 className="decisions-header-title">Decision Recorder</h1>
          <p className="decisions-header-subtitle">
            Track team decisions, rationale, and approvals. Never wonder "why did we choose X?" again.
          </p>
        </div>
        <button onClick={handleCreateDecision} className="decisions-record-btn">
          <Plus className="h-4 w-4" />
          Record Decision
        </button>
      </div>

      {/* Status Summary Cards */}
      <div className="decisions-status-grid">
        <div
          className={`decisions-status-card pending ${statusFilter === 'pending' ? 'active' : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'pending' ? 'all' : 'pending')}
        >
          <div className="decisions-status-icon">
            <Clock className="h-6 w-6" />
          </div>
          <p className="decisions-status-count">{statusCounts.pending}</p>
          <p className="decisions-status-label">Pending</p>
        </div>

        <div
          className={`decisions-status-card approved ${statusFilter === 'approved' ? 'active' : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'approved' ? 'all' : 'approved')}
        >
          <div className="decisions-status-icon">
            <CheckCircle className="h-6 w-6" />
          </div>
          <p className="decisions-status-count">{statusCounts.approved}</p>
          <p className="decisions-status-label">Approved</p>
        </div>

        <div
          className={`decisions-status-card superseded ${statusFilter === 'superseded' ? 'active' : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'superseded' ? 'all' : 'superseded')}
        >
          <div className="decisions-status-icon">
            <AlertCircle className="h-6 w-6" />
          </div>
          <p className="decisions-status-count">{statusCounts.superseded}</p>
          <p className="decisions-status-label">Superseded</p>
        </div>

        <div
          className={`decisions-status-card deprecated ${statusFilter === 'deprecated' ? 'active' : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'deprecated' ? 'all' : 'deprecated')}
        >
          <div className="decisions-status-icon">
            <XCircle className="h-6 w-6" />
          </div>
          <p className="decisions-status-count">{statusCounts.deprecated}</p>
          <p className="decisions-status-label">Deprecated</p>
        </div>
      </div>

      {/* Filters */}
      <div className="decisions-filters">
        <div className="decisions-search">
          <Search className="decisions-search-icon" />
          <input
            type="text"
            placeholder="Search decisions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="decisions-search-input"
          />
        </div>

        <Select value={teamFilter} onValueChange={setTeamFilter}>
          <SelectTrigger className="decisions-select-trigger">
            <SelectValue placeholder="Select team" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Teams</SelectItem>
            {teams.map((team) => (
              <SelectItem key={team.id} value={team.id}>
                {team.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select
          value={statusFilter}
          onValueChange={(v) => setStatusFilter(v as DecisionStatus | 'all')}
        >
          <SelectTrigger className="decisions-select-trigger">
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Statuses</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="approved">Approved</SelectItem>
            <SelectItem value="superseded">Superseded</SelectItem>
            <SelectItem value="deprecated">Deprecated</SelectItem>
          </SelectContent>
        </Select>

        <div className="decisions-view-toggle">
          <button
            className={`decisions-view-btn ${viewMode === 'grid' ? 'active' : ''}`}
            onClick={() => setViewMode('grid')}
          >
            <Grid className="h-4 w-4" />
          </button>
          <button
            className={`decisions-view-btn ${viewMode === 'list' ? 'active' : ''}`}
            onClick={() => setViewMode('list')}
          >
            <List className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Content */}
      {teamFilter === 'all' ? (
        <div className="decisions-select-team-state">
          <div className="decisions-select-team-state-icon">
            <AlertCircle className="h-12 w-12" />
          </div>
          <h3 className="decisions-select-team-state-title">Select a Team</h3>
          <p className="decisions-select-team-state-description">
            Choose a team above to view their decisions. Each team maintains its own decision history
            with full context and approval tracking.
          </p>
        </div>
      ) : isDecisionsLoading ? (
        <div className="decisions-loading">
          <div className="decisions-loading-spinner" />
        </div>
      ) : decisionsError ? (
        <div className="decisions-error-state">
          <div className="decisions-error-state-icon">
            <AlertCircle className="h-12 w-12" />
          </div>
          <p className="decisions-error-state-text">Failed to load decisions. Please try again.</p>
        </div>
      ) : filteredDecisions.length === 0 ? (
        <div className="decisions-empty-state">
          <div className="decisions-empty-state-icon success">
            <CheckCircle className="h-12 w-12" />
          </div>
          <h3 className="decisions-empty-state-title">No decisions yet</h3>
          <p className="decisions-empty-state-description">
            {searchQuery || statusFilter !== 'all'
              ? 'No decisions match your filters.'
              : 'Start recording team decisions to build your decision history.'}
          </p>
          <button onClick={handleCreateDecision} className="decisions-empty-state-btn">
            <Plus className="h-4 w-4" />
            Record Your First Decision
          </button>
        </div>
      ) : (
        <div className={viewMode === 'grid' ? 'decisions-grid' : 'decisions-list'}>
          {filteredDecisions.map((decision) => (
            <DecisionCard
              key={decision.id}
              decision={decision}
              teamId={decision.team_id}
              onView={handleViewDecision}
              onEdit={handleEditDecision}
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
              {
                onSuccess: () => {
                  setCreateDialogOpen(false);
                  setSelectedTeamForCreate('');
                },
              }
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
        onEdit={handleEditDecision}
        onDelete={handleDeleteDecision}
        onApprove={handleApproveDecision}
      />
    </div>
  );
}
