import { useState } from "react";
import {
  Trophy,
  Plus,
  Play,
  X,
  ChevronRight,
  Calendar,
  Clock,
  Users,
  Shield,
  AlertCircle,
  RefreshCw,
  CheckCircle,
  XCircle,
  Swords,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  useCityWars,
  useCityWar,
  useEligibleMetros,
  useCreateWar,
  useActivateWar,
  useCancelWar,
  useGenerateBracket,
  useSetQuarterfinals,
  useAdvanceWar,
  useOverrideMatch,
  useRecordMatch,
} from "@/hooks/useAdmin";
import type { War, WarMatch, EligibleMetro, MatchResultRequest } from "@/api/admin";
import { cn } from "@/lib/utils";

export function CityWarsAdminPage() {
  const [selectedWarId, setSelectedWarId] = useState<number | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showOverrideDialog, setShowOverrideDialog] = useState<{ match: WarMatch; warId: number } | null>(null);

  const { data: wars, isLoading: warsLoading, refetch: refetchWars } = useCityWars();
  const { data: eligibleMetros } = useEligibleMetros();

  const selectedWar = useCityWar(selectedWarId ?? 0);

  const createMutation = useCreateWar();
  const activateMutation = useActivateWar();
  const cancelMutation = useCancelWar();
  const generateBracketMutation = useGenerateBracket();
  const advanceMutation = useAdvanceWar();
  const overrideMutation = useOverrideMatch();
  const recordMutation = useRecordMatch();

  const getStatusColor = (status?: string) => {
    switch (status) {
      case "active":
        return "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20";
      case "scheduled":
        return "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20";
      case "complete":
        return "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20";
      case "cancelled":
        return "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20";
      default:
        return "bg-gray-500/10 text-gray-600 dark:text-gray-400 border-gray-500/20";
    }
  };

  const getRoundLabel = (round?: string) => {
    switch (round) {
      case "quarterfinal":
        return "Quarterfinals";
      case "semifinal":
        return "Semifinals";
      case "final":
        return "Final";
      case "complete":
        return "Champion";
      default:
        return round || "—";
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "—";
    return new Date(dateString).toLocaleDateString();
  };

  const formatDateTime = (dateString?: string) => {
    if (!dateString) return "—";
    return new Date(dateString).toLocaleString();
  };

  const handleCreateWar = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    await createMutation.mutateAsync({
      name: formData.get("name") as string,
      season: formData.get("season") as string,
      slug: formData.get("slug") as string,
      starts_at: new Date(formData.get("starts_at") as string).toISOString(),
      ends_at: new Date(formData.get("ends_at") as string).toISOString(),
    });
    setShowCreateDialog(false);
  };

  const handleActivate = async (id: number) => {
    await activateMutation.mutateAsync(id);
  };

  const handleCancel = async (id: number) => {
    await cancelMutation.mutateAsync(id);
  };

  const handleGenerateBracket = async (id: number) => {
    await generateBracketMutation.mutateAsync(id);
  };

  const handleAdvance = async (id: number, currentRound: string) => {
    const nextRound = currentRound === "quarterfinal" ? "semifinal" : "final";
    await advanceMutation.mutateAsync({ id, round: nextRound });
  };

  const handleOverride = async (data: MatchResultRequest) => {
    if (!showOverrideDialog) return;
    await overrideMutation.mutateAsync({ id: showOverrideDialog.warId, data });
    setShowOverrideDialog(null);
  };

  const isMutating =
    createMutation.isPending ||
    activateMutation.isPending ||
    cancelMutation.isPending ||
    generateBracketMutation.isPending ||
    advanceMutation.isPending ||
    overrideMutation.isPending ||
    recordMutation.isPending;

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground flex items-center gap-2">
            <Swords className="h-6 w-6" />
            City Wars Admin
          </h1>
          <p className="text-muted-foreground">
            Manage quarterly bracket tournaments between top AI cities
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetchWars()}
            disabled={warsLoading}
          >
            <RefreshCw className={cn("h-4 w-4 mr-2", warsLoading && "animate-spin")} />
            Refresh
          </Button>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Create War
          </Button>
        </div>
      </div>

      <Tabs defaultValue="list" className="space-y-6">
        <TabsList>
          <TabsTrigger value="list">All Wars</TabsTrigger>
          <TabsTrigger value="detail" disabled={!selectedWarId}>
            War Detail
          </TabsTrigger>
        </TabsList>

        <TabsContent value="list" className="space-y-6">
          {warsLoading ? (
            <div className="grid gap-4">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-24" />
              ))}
            </div>
          ) : wars?.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <Trophy className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-medium">No Wars Yet</h3>
                <p className="text-muted-foreground mt-2">
                  Create your first City War to start the bracket tournament.
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4">
              {wars?.map((war) => (
                <Card
                  key={war.id}
                  className={cn(
                    "cursor-pointer transition-colors hover:bg-accent",
                    selectedWarId === war.id && "border-primary"
                  )}
                  onClick={() => setSelectedWarId(war.id)}
                >
                  <CardContent className="py-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className="flex flex-col">
                          <span className="font-semibold">{war.name}</span>
                          <span className="text-sm text-muted-foreground">
                            {war.season} &middot; {war.slug}
                          </span>
                        </div>
                        <Badge className={getStatusColor(war.status)}>
                          {war.status}
                        </Badge>
                        <Badge variant="outline">
                          {getRoundLabel(war.round)}
                        </Badge>
                      </div>
                      <div className="flex items-center gap-6 text-sm text-muted-foreground">
                        <div className="flex items-center gap-1">
                          <Calendar className="h-4 w-4" />
                          {formatDate(war.starts_at)} - {formatDate(war.ends_at)}
                        </div>
                        <div className="flex items-center gap-1">
                          <Users className="h-4 w-4" />
                          {war.total_active_users.toLocaleString()} users
                        </div>
                        {war.champion_name && (
                          <div className="flex items-center gap-1 text-amber-600">
                            <Trophy className="h-4 w-4" />
                            {war.champion_name}
                          </div>
                        )}
                        <ChevronRight className="h-4 w-4" />
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="detail" className="space-y-6">
          {selectedWar.isLoading ? (
            <Skeleton className="h-96" />
          ) : selectedWar.data ? (
            <WarDetail
              war={selectedWar.data}
              eligibleMetros={eligibleMetros ?? []}
              onActivate={handleActivate}
              onCancel={handleCancel}
              onGenerateBracket={handleGenerateBracket}
              onAdvance={handleAdvance}
              onOverride={(match) => setShowOverrideDialog({ match, warId: selectedWar.data!.id })}
              isMutating={isMutating}
              getStatusColor={getStatusColor}
              getRoundLabel={getRoundLabel}
              formatDateTime={formatDateTime}
            />
          ) : (
            <Card>
              <CardContent className="py-12 text-center">
                <AlertCircle className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p>Select a war to view details</p>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>

      {/* Create Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New War</DialogTitle>
            <DialogDescription>
              Set up a new City War bracket tournament
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateWar} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">War Name</Label>
              <Input
                id="name"
                name="name"
                placeholder="Summer 2026 Championship"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="slug">Slug</Label>
              <Input
                id="slug"
                name="slug"
                placeholder="summer-2026"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="season">Season</Label>
              <Input
                id="season"
                name="season"
                placeholder="2026-Q3"
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="starts_at">Start Date</Label>
                <Input
                  id="starts_at"
                  name="starts_at"
                  type="datetime-local"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="ends_at">End Date</Label>
                <Input
                  id="ends_at"
                  name="ends_at"
                  type="datetime-local"
                  required
                />
              </div>
            </div>
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowCreateDialog(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? "Creating..." : "Create War"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Override Dialog */}
      <Dialog
        open={!!showOverrideDialog}
        onOpenChange={() => setShowOverrideDialog(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Override Match Result</DialogTitle>
            <DialogDescription>
              Manually set the winner and scores for this match
            </DialogDescription>
          </DialogHeader>
          {showOverrideDialog && (
            <OverrideMatchForm
              match={showOverrideDialog.match}
              onSubmit={handleOverride}
              onCancel={() => setShowOverrideDialog(null)}
              isPending={overrideMutation.isPending}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

interface WarDetailProps {
  war: War;
  eligibleMetros: EligibleMetro[];
  onActivate: (id: number) => void;
  onCancel: (id: number) => void;
  onGenerateBracket: (id: number) => void;
  onAdvance: (id: number, currentRound: string) => void;
  onOverride: (match: WarMatch) => void;
  isMutating: boolean;
  getStatusColor: (status?: string) => string;
  getRoundLabel: (round?: string) => string;
  formatDateTime: (date?: string) => string;
}

function WarDetail({
  war,
  eligibleMetros,
  onActivate,
  onCancel,
  onGenerateBracket,
  onAdvance,
  onOverride,
  isMutating,
  getStatusColor,
  getRoundLabel,
  formatDateTime,
}: WarDetailProps) {
  const [selectedMetros, setSelectedMetros] = useState<[number, number][]>([]);

  const canEdit = war.status === "scheduled";
  const canActivate = war.status === "scheduled" && war.quarterfinals && war.quarterfinals.length === 4;
  const canAdvance =
    war.status === "active" &&
    war.round !== "complete" &&
    war.quarterfinals?.every((m) => m.winner_metro_id) &&
    (war.round !== "semifinal" || war.semifinals?.every((m) => m.winner_metro_id));

  const handleMetroSelect = (pairIndex: number, side: "a" | "b", metroId: number) => {
    const newSelection = [...selectedMetros];
    if (!newSelection[pairIndex]) {
      newSelection[pairIndex] = [0, 0];
    }
    if (side === "a") {
      newSelection[pairIndex] = [metroId, newSelection[pairIndex][1]];
    } else {
      newSelection[pairIndex] = [newSelection[pairIndex][0], metroId];
    }
    setSelectedMetros(newSelection);
  };

  return (
    <div className="space-y-6">
      {/* Header Card */}
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle className="text-2xl">{war.name}</CardTitle>
              <CardDescription>
                {war.season} &middot; {war.slug}
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Badge className={getStatusColor(war.status)}>{war.status}</Badge>
              <Badge variant="outline">{getRoundLabel(war.round)}</Badge>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Starts</p>
              <p className="font-medium">{formatDateTime(war.starts_at)}</p>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Ends</p>
              <p className="font-medium">{formatDateTime(war.ends_at)}</p>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Active Users</p>
              <p className="font-medium">{war.total_active_users.toLocaleString()}</p>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Champion</p>
              <p className="font-medium flex items-center gap-1">
                {war.champion_name ? (
                  <>
                    <Trophy className="h-4 w-4 text-amber-500" />
                    {war.champion_name}
                  </>
                ) : (
                  "—"
                )}
              </p>
            </div>
          </div>

          {canEdit && (
            <div className="flex gap-2 mt-6 pt-6 border-t">
              {!war.quarterfinals || war.quarterfinals.length === 0 ? (
                <Button onClick={() => onGenerateBracket(war.id)} disabled={isMutating}>
                  <Shield className="h-4 w-4 mr-2" />
                  Auto-Generate Bracket
                </Button>
              ) : (
                <Button onClick={() => onGenerateBracket(war.id)} disabled={isMutating} variant="outline">
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Regenerate Bracket
                </Button>
              )}
              {canActivate && (
                <Button onClick={() => onActivate(war.id)} disabled={isMutating}>
                  <Play className="h-4 w-4 mr-2" />
                  Activate War
                </Button>
              )}
              <Button
                onClick={() => onCancel(war.id)}
                disabled={isMutating}
                variant="destructive"
              >
                <XCircle className="h-4 w-4 mr-2" />
                Cancel
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Bracket Selection (for scheduled wars) */}
      {canEdit && eligibleMetros.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Edit Bracket</CardTitle>
            <CardDescription>
              Select which metros compete in each quarterfinal match
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {[0, 1, 2, 3].map((pairIdx) => (
                <div key={pairIdx} className="space-y-2">
                  <p className="text-sm font-medium">Match {pairIdx + 1}</p>
                  <div className="space-y-2">
                    <select
                      className="w-full p-2 border rounded"
                      value={selectedMetros[pairIdx]?.[0] ?? ""}
                      onChange={(e) => handleMetroSelect(pairIdx, "a", Number(e.target.value))}
                    >
                      <option value="">Select Metro A</option>
                      {eligibleMetros.map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.name} (#{m.rank_position})
                        </option>
                      ))}
                    </select>
                    <p className="text-center text-muted-foreground">vs</p>
                    <select
                      className="w-full p-2 border rounded"
                      value={selectedMetros[pairIdx]?.[1] ?? ""}
                      onChange={(e) => handleMetroSelect(pairIdx, "b", Number(e.target.value))}
                    >
                      <option value="">Select Metro B</option>
                      {eligibleMetros.map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.name} (#{m.rank_position})
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Bracket Display */}
      <Card>
        <CardHeader>
          <CardTitle>Bracket</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-8">
            {/* Quarterfinals */}
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-3">Quarterfinals</h4>
              {war.quarterfinals && war.quarterfinals.length > 0 ? (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {war.quarterfinals.map((match) => (
                    <MatchCard
                      key={match.id}
                      match={match}
                      warId={war.id}
                      onOverride={() => onOverride(match)}
                      showOverride={war.status === "active"}
                      formatDateTime={formatDateTime}
                    />
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No quarterfinal matches yet</p>
              )}
            </div>

            {/* Semifinals */}
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-3">Semifinals</h4>
              {war.semifinals && war.semifinals.length > 0 ? (
                <div className="grid grid-cols-2 gap-4">
                  {war.semifinals.map((match) => (
                    <MatchCard
                      key={match.id}
                      match={match}
                      warId={war.id}
                      onOverride={() => onOverride(match)}
                      showOverride={war.status === "active"}
                      formatDateTime={formatDateTime}
                    />
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No semifinal matches yet</p>
              )}
            </div>

            {/* Final */}
            <div>
              <h4 className="text-sm font-medium text-muted-foreground mb-3">Final</h4>
              {war.final ? (
                <div className="max-w-md mx-auto">
                  <MatchCard
                    match={war.final}
                    warId={war.id}
                    onOverride={() => onOverride(war.final)}
                    showOverride={war.status === "active"}
                    formatDateTime={formatDateTime}
                    isFinal
                  />
                </div>
              ) : (
                <p className="text-muted-foreground">No final match yet</p>
              )}
            </div>
          </div>

          {/* Advance Button */}
          {canAdvance && (
            <div className="mt-6 pt-6 border-t flex justify-center">
              <Button
                onClick={() => onAdvance(war.id, war.round)}
                disabled={isMutating}
                size="lg"
              >
                <ChevronRight className="h-4 w-4 mr-2" />
                Advance to {war.round === "quarterfinal" ? "Semifinals" : "Champion"}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

interface MatchCardProps {
  match: WarMatch;
  warId: number;
  onOverride: () => void;
  showOverride: boolean;
  formatDateTime: (date?: string) => string;
  isFinal?: boolean;
}

function MatchCard({
  match,
  onOverride,
  showOverride,
  formatDateTime,
  isFinal,
}: MatchCardProps) {
  const hasWinner = !!match.winner_metro_id;
  const winnerIsA = match.winner_metro_id === match.metro_a_id;

  return (
    <div
      className={cn(
        "border rounded-lg p-4",
        hasWinner && "bg-emerald-500/5 border-emerald-500/20"
      )}
    >
      <div className="flex items-center justify-between mb-2">
        <Badge variant="outline" className="text-xs">
          {match.round} #{match.position}
        </Badge>
        {hasWinner && (
          <CheckCircle className="h-4 w-4 text-emerald-500" />
        )}
      </div>

      <div className="space-y-2">
        {/* Metro A */}
        <div
          className={cn(
            "p-2 rounded flex justify-between items-center",
            winnerIsA && "bg-emerald-500/10"
          )}
        >
          <div>
            <p className="font-medium text-sm">{match.metro_a_name}</p>
            <p className="text-xs text-muted-foreground">
              {match.metro_a_country}
            </p>
          </div>
          <div className="text-right">
            <p className="font-bold">{match.score_a.toFixed(4)}</p>
            <p className="text-xs text-muted-foreground">
              {match.active_users_a} users
            </p>
          </div>
        </div>

        <p className="text-center text-muted-foreground text-sm">vs</p>

        {/* Metro B */}
        <div
          className={cn(
            "p-2 rounded flex justify-between items-center",
            !winnerIsA && hasWinner && "bg-emerald-500/10"
          )}
        >
          <div>
            <p className="font-medium text-sm">{match.metro_b_name}</p>
            <p className="text-xs text-muted-foreground">
              {match.metro_b_country}
            </p>
          </div>
          <div className="text-right">
            <p className="font-bold">{match.score_b.toFixed(4)}</p>
            <p className="text-xs text-muted-foreground">
              {match.active_users_b} users
            </p>
          </div>
        </div>
      </div>

      {showOverride && (
        <Button
          variant="outline"
          size="sm"
          className="w-full mt-3"
          onClick={onOverride}
        >
          Override Result
        </Button>
      )}
    </div>
  );
}

interface OverrideMatchFormProps {
  match: WarMatch;
  onSubmit: (data: any) => void;
  onCancel: () => void;
  isPending: boolean;
}

function OverrideMatchForm({
  match,
  onSubmit,
  onCancel,
  isPending,
}: OverrideMatchFormProps) {
  const [winnerId, setWinnerId] = useState(match.metro_a_id);
  const [scoreA, setScoreA] = useState(match.score_a);
  const [scoreB, setScoreB] = useState(match.score_b);
  const [activeUsersA, setActiveUsersA] = useState(match.active_users_a);
  const [activeUsersB, setActiveUsersB] = useState(match.active_users_b);
  const [note, setNote] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      match_id: match.id,
      score_a: scoreA,
      score_b: scoreB,
      active_users_a: activeUsersA,
      active_users_b: activeUsersB,
      winner_id: winnerId,
      note,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{match.metro_a_name} Score</Label>
          <Input
            type="number"
            step="0.000001"
            value={scoreA}
            onChange={(e) => setScoreA(Number(e.target.value))}
          />
        </div>
        <div className="space-y-2">
          <Label>{match.metro_b_name} Score</Label>
          <Input
            type="number"
            step="0.000001"
            value={scoreB}
            onChange={(e) => setScoreB(Number(e.target.value))}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label>{match.metro_a_name} Users</Label>
          <Input
            type="number"
            value={activeUsersA}
            onChange={(e) => setActiveUsersA(Number(e.target.value))}
          />
        </div>
        <div className="space-y-2">
          <Label>{match.metro_b_name} Users</Label>
          <Input
            type="number"
            value={activeUsersB}
            onChange={(e) => setActiveUsersB(Number(e.target.value))}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Winner</Label>
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              name="winner"
              value={match.metro_a_id}
              checked={winnerId === match.metro_a_id}
              onChange={() => setWinnerId(match.metro_a_id)}
            />
            {match.metro_a_name}
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              name="winner"
              value={match.metro_b_id}
              checked={winnerId === match.metro_b_id}
              onChange={() => setWinnerId(match.metro_b_id)}
            />
            {match.metro_b_name}
          </label>
        </div>
      </div>

      <div className="space-y-2">
        <Label>Reason / Note</Label>
        <Input
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g., Data correction after ranking recalculation"
        />
      </div>

      <DialogFooter>
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={isPending}>
          {isPending ? "Saving..." : "Override Result"}
        </Button>
      </DialogFooter>
    </form>
  );
}

export default CityWarsAdminPage;
