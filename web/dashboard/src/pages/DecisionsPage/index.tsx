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
    <div className="container mx-auto py-8 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Decision Recorder</h1>
          <p className="text-muted-foreground mt-1">
            Track team decisions, rationale, and approvals. Never wonder "why did we choose X?"
            again.
          </p>
        </div>
        <Button onClick={handleCreateDecision}>
          <Plus className="mr-2 h-4 w-4" />
          Record Decision
        </Button>
      </div>

      {/* Status Summary Cards */}
      <div className="grid grid-cols-4 gap-4">
        <Card
          className={`cursor-pointer transition-all ${
            statusFilter === 'pending' ? 'border-yellow-500 bg-yellow-500/5' : ''
          }`}
          onClick={() => setStatusFilter(statusFilter === 'pending' ? 'all' : 'pending')}
        >
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-yellow-500/10">
                <Clock className="h-6 w-6 text-yellow-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">{statusCounts.pending}</p>
                <p className="text-sm text-muted-foreground">Pending</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card
          className={`cursor-pointer transition-all ${
            statusFilter === 'approved' ? 'border-green-500 bg-green-500/5' : ''
          }`}
          onClick={() => setStatusFilter(statusFilter === 'approved' ? 'all' : 'approved')}
        >
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-green-500/10">
                <CheckCircle className="h-6 w-6 text-green-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">{statusCounts.approved}</p>
                <p className="text-sm text-muted-foreground">Approved</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card
          className={`cursor-pointer transition-all ${
            statusFilter === 'superseded' ? 'border-orange-500 bg-orange-500/5' : ''
          }`}
          onClick={() => setStatusFilter(statusFilter === 'superseded' ? 'all' : 'superseded')}
        >
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-orange-500/10">
                <AlertCircle className="h-6 w-6 text-orange-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">{statusCounts.superseded}</p>
                <p className="text-sm text-muted-foreground">Superseded</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card
          className={`cursor-pointer transition-all ${
            statusFilter === 'deprecated' ? 'border-gray-500 bg-gray-500/5' : ''
          }`}
          onClick={() => setStatusFilter(statusFilter === 'deprecated' ? 'all' : 'deprecated')}
        >
          <CardContent className="pt-6">
            <div className="flex items-center gap-4">
              <div className="p-3 rounded-lg bg-gray-500/10">
                <XCircle className="h-6 w-6 text-gray-500" />
              </div>
              <div>
                <p className="text-2xl font-bold">{statusCounts.deprecated}</p>
                <p className="text-sm text-muted-foreground">Deprecated</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search decisions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <Select value={teamFilter} onValueChange={setTeamFilter}>
          <SelectTrigger className="w-[200px]">
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
          <SelectTrigger className="w-[150px]">
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

        <div className="flex items-center gap-1 border rounded-md">
          <Button
            variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
            size="icon"
            className="h-8 w-8"
            onClick={() => setViewMode('grid')}
          >
            <Grid className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'secondary' : 'ghost'}
            size="icon"
            className="h-8 w-8"
            onClick={() => setViewMode('list')}
          >
            <List className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Content */}
      {teamFilter === 'all' ? (
        <Card>
          <CardContent className="py-12 text-center">
            <AlertCircle className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">Select a Team</h3>
            <p className="text-muted-foreground max-w-md mx-auto">
              Choose a team above to view their decisions. Each team maintains its own decision
              history with full context and approval tracking.
            </p>
          </CardContent>
        </Card>
      ) : isDecisionsLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : decisionsError ? (
        <Card className="border-destructive">
          <CardContent className="py-12 text-center text-destructive">
            <AlertCircle className="h-12 w-12 mx-auto mb-4" />
            <p>Failed to load decisions. Please try again.</p>
          </CardContent>
        </Card>
      ) : filteredDecisions.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <CheckCircle className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">No decisions yet</h3>
            <p className="text-muted-foreground max-w-md mx-auto mb-4">
              {searchQuery || statusFilter !== 'all'
                ? 'No decisions match your filters.'
                : 'Start recording team decisions to build your decision history.'}
            </p>
            <Button onClick={handleCreateDecision}>
              <Plus className="mr-2 h-4 w-4" />
              Record Your First Decision
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div
          className={viewMode === 'grid' ? 'grid gap-4 md:grid-cols-2 lg:grid-cols-3' : 'space-y-4'}
        >
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
