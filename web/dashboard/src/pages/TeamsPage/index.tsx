import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AnimatePresence, motion } from 'framer-motion';
import {
  AlertCircle,
  ChevronRight,
  Clock,
  Crown,
  Edit3,
  Eye,
  LayoutGrid,
  List,
  Loader2,
  Mail,
  MoreVertical,
  Plus,
  Search,
  Settings,
  Shield,
  Trash2,
  User,
  UserPlus,
  Users,
  X,
} from 'lucide-react';
import { useState } from 'react';
// import { ToggleButtonGroup } from "@/components/ui";
import { teamsApi, type Team, type TeamMember } from '@/api/teams';
import { useAuthStore } from '@/stores/authStore';
import { format, formatDistanceToNow } from 'date-fns';
import { toast } from 'sonner';

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: User,
  viewer: Eye,
};

const roleColors = {
  owner: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  admin: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  member: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  viewer: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
};

const roleLabels = {
  owner: 'Owner',
  admin: 'Admin',
  member: 'Member',
  viewer: 'Viewer',
};

export function TeamsPage() {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [newTeamName, setNewTeamName] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<'admin' | 'member' | 'viewer'>('member');
  const [searchQuery, setSearchQuery] = useState('');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [activeTab, setActiveTab] = useState('members');

  // Fetch teams
  const { data: teamsData, isLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: () => teamsApi.list(),
  });

  const teams = teamsData?.teams ?? [];

  // Filter teams by search
  const filteredTeams = teams.filter((team) =>
    team.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Fetch members for selected team
  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ['team-members', selectedTeam?.id],
    queryFn: () => teamsApi.listMembers(selectedTeam!.id),
    enabled: !!selectedTeam,
  });

  // Fetch invites for selected team
  const { data: invitesData, isLoading: invitesLoading } = useQuery({
    queryKey: ['team-invites', selectedTeam?.id],
    queryFn: () => teamsApi.listInvites(selectedTeam!.id),
    enabled: !!selectedTeam,
  });

  const members = membersData?.members ?? [];
  const invites = invitesData?.invites ?? [];

  // Create team mutation
  const createTeamMutation = useMutation({
    mutationFn: (name: string) => teamsApi.create({ name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      setIsCreateDialogOpen(false);
      setNewTeamName('');
      toast.success('Team created successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to create team');
    },
  });

  // Delete team mutation
  const deleteTeamMutation = useMutation({
    mutationFn: (teamId: string) => teamsApi.delete(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      if (selectedTeam) setSelectedTeam(null);
      toast.success('Team deleted successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to delete team');
    },
  });

  // Invite member mutation
  const inviteMemberMutation = useMutation({
    mutationFn: ({ teamId, email, role }: { teamId: string; email: string; role: string }) =>
      teamsApi.addMember(teamId, { email, role: role as 'admin' | 'member' | 'viewer' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', selectedTeam?.id] });
      queryClient.invalidateQueries({ queryKey: ['team-invites', selectedTeam?.id] });
      setIsInviteDialogOpen(false);
      setInviteEmail('');
      setInviteRole('member');
      toast.success('Invitation sent successfully');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to send invitation');
    },
  });

  // Update member role mutation
  const updateMemberMutation = useMutation({
    mutationFn: ({ teamId, memberId, role }: { teamId: string; memberId: string; role: string }) =>
      teamsApi.updateMember(teamId, memberId, { role: role as 'admin' | 'member' | 'viewer' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', selectedTeam?.id] });
      toast.success('Member role updated');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to update role');
    },
  });

  // Remove member mutation
  const removeMemberMutation = useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) =>
      teamsApi.removeMember(teamId, memberId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members', selectedTeam?.id] });
      toast.success('Member removed from team');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to remove member');
    },
  });

  // Cancel invite mutation
  const cancelInviteMutation = useMutation({
    mutationFn: ({ teamId, inviteId }: { teamId: string; inviteId: string }) =>
      teamsApi.cancelInvite(teamId, inviteId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-invites', selectedTeam?.id] });
      toast.success('Invitation cancelled');
    },
    onError: (error: Error) => {
      toast.error(error.message || 'Failed to cancel invitation');
    },
  });

  const handleCreateTeam = () => {
    if (!newTeamName.trim()) {
      toast.error('Team name is required');
      return;
    }
    createTeamMutation.mutate(newTeamName);
  };

  const handleInvite = () => {
    if (!inviteEmail.trim() || !selectedTeam) {
      toast.error('Email is required');
      return;
    }
    inviteMemberMutation.mutate({ teamId: selectedTeam.id, email: inviteEmail, role: inviteRole });
  };

  const getInitials = (name?: string, email?: string) => {
    if (name)
      return name
        .split(' ')
        .map((n) => n[0])
        .join('')
        .toUpperCase()
        .slice(0, 2);
    if (email) return email[0].toUpperCase();
    return '?';
  };

  const isCurrentUserOwner = (member: TeamMember) =>
    member.user_id === user?.id && member.role === 'owner';

  const getUserRoleInTeam = (team: Team) => {
    if (team.owner_id === user?.id) return 'owner';
    const member = team.members?.find((m) => m.user_id === user?.id);
    return member?.role || 'member';
  };

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: 'var(--text-primary)' }}>Teams</h1>
          <p className="text-gray-400 mt-1" style={{ color: 'var(--text-secondary)' }}>Manage your teams, members, and permissions</p>
        </div>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button className="bg-[#6366f1] hover:bg-[#5558e0]">
              <Plus className="w-4 h-4 mr-2" />
              Create Team
            </Button>
          </DialogTrigger>
          <DialogContent className="[&>button]:text-gray-400" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)', color: 'var(--team-card-text)' }}>
            <DialogHeader>
              <DialogTitle style={{ color: 'var(--team-card-text)' }}>Create New Team</DialogTitle>
              <DialogDescription style={{ color: 'var(--team-card-text-secondary)' }}>
                Create a new team to organize members and manage permissions.
              </DialogDescription>
            </DialogHeader>
            <div className="py-4">
              <Label htmlFor="team-name" style={{ color: 'var(--team-card-text-secondary)' }}>
                Team Name
              </Label>
              <Input
                id="team-name"
                placeholder="e.g., Engineering, Marketing"
                value={newTeamName}
                onChange={(e) => setNewTeamName(e.target.value)}
                style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-input-border)', color: 'var(--team-input-text)' }}
                className="mt-2 placeholder:text-gray-500"
              />
            </div>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setIsCreateDialogOpen(false)}
                style={{ borderColor: 'var(--team-card-border)', color: 'var(--team-card-text-secondary)' }}
                className="hover:bg-[#21262d]"
              >
                Cancel
              </Button>
              <Button
                onClick={handleCreateTeam}
                disabled={createTeamMutation.isPending || !newTeamName.trim()}
                className="bg-[#6366f1] hover:bg-[#5558e0]"
              >
                {createTeamMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Create Team
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Search and Filters */}
      {teams.length > 0 && (
        <div className="flex items-center gap-4">
          <div className="relative flex-1 max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" style={{ color: 'var(--team-input-placeholder)' }} />
            <Input
              placeholder="Search teams..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-card-border)', color: 'var(--team-input-text)' }}
              className="pl-10"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 hover:opacity-70"
                style={{ color: 'var(--team-input-placeholder)' }}
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
          <div className="flex items-center gap-2 ml-auto rounded-lg p-1" style={{ backgroundColor: 'var(--team-card-bg)', border: '1px solid var(--team-card-border)' }}>
            <Button
              variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
              size="icon"
              className={`h-8 w-8 ${viewMode === 'grid' ? '' : ''}`}
              onClick={() => setViewMode('grid')}
              aria-label="Grid view"
              style={viewMode === 'grid' ? { backgroundColor: 'var(--team-input-bg)', color: 'var(--team-card-text)' } : { color: 'var(--team-card-text-secondary)' }}
            >
              <LayoutGrid className="w-4 h-4" />
            </Button>
            <Button
              variant={viewMode === 'list' ? 'secondary' : 'ghost'}
              size="icon"
              className={`h-8 w-8 ${viewMode === 'list' ? '' : ''}`}
              onClick={() => setViewMode('list')}
              aria-label="List view"
              style={viewMode === 'list' ? { backgroundColor: 'var(--team-input-bg)', color: 'var(--team-card-text)' } : { color: 'var(--team-card-text-secondary)' }}
            >
              <List className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Teams Grid/List */}
      {isLoading ? (
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="w-8 h-8 animate-spin text-[#6366f1]" />
        </div>
      ) : teams.length === 0 ? (
        <Card style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
          <CardContent className="py-12 text-center">
            <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 flex items-center justify-center mx-auto mb-4">
              <Users className="w-8 h-8 text-[#6366f1]" />
            </div>
            <h3 className="text-lg font-semibold mb-2" style={{ color: 'var(--team-card-text)' }}>No teams yet</h3>
            <p className="mb-6 max-w-sm mx-auto" style={{ color: 'var(--team-card-text-secondary)' }}>
              Create your first team to start collaborating with your colleagues.
            </p>
            <Button
              onClick={() => setIsCreateDialogOpen(true)}
              className="bg-[#6366f1] hover:bg-[#5558e0]"
            >
              <Plus className="w-4 h-4 mr-2" />
              Create Team
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div
          className={
            viewMode === 'grid'
              ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
              : 'space-y-3'
          }
        >
          <AnimatePresence mode="popLayout">
            {filteredTeams.map((team, index) => {
              const userRole = getUserRoleInTeam(team);
              const RoleIcon = roleIcons[userRole as keyof typeof roleIcons] || User;
              const isSelected = selectedTeam?.id === team.id;

              return (
                <motion.div
                  key={team.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.05 }}
                >
                  <Card
                    style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}
                    className={`transition-all duration-200 cursor-pointer group ${
                      viewMode === 'list'
                        ? 'hover:border-[#6366f1]/30'
                        : 'hover:border-[#6366f1]/50 hover:-translate-y-0.5'
                    } ${isSelected ? 'ring-2 ring-[#6366f1] border-[#6366f1]' : ''}`}
                    onClick={() => setSelectedTeam(team)}
                  >
                    <CardHeader className={viewMode === 'list' ? 'pb-3' : 'pb-3'}>
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-3">
                          <div
                            className={`rounded-xl bg-gradient-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 border border-[#6366f1]/20 flex items-center justify-center ${
                              viewMode === 'list' ? 'w-10 h-10' : 'w-12 h-12'
                            }`}
                          >
                            <Users
                              className={`text-[#6366f1] ${viewMode === 'list' ? 'w-5 h-5' : 'w-6 h-6'}`}
                            />
                          </div>
                          <div>
                            <CardTitle className="text-base" style={{ color: 'var(--team-card-text)' }}>{team.name}</CardTitle>
                            <CardDescription style={{ color: 'var(--team-card-text-secondary)' }}>
                              {team.members?.length ?? 0} members
                            </CardDescription>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <Badge
                            variant="outline"
                            className={`${roleColors[userRole as keyof typeof roleColors]} text-xs`}
                          >
                            <RoleIcon className="w-3 h-3 mr-1" />
                            {roleLabels[userRole as keyof typeof roleLabels]}
                          </Badge>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                style={{ color: 'var(--team-card-text-secondary)' }}
                                aria-label="Team options"
                              >
                                <MoreVertical className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                              align="end"
                              style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}
                            >
                              <DropdownMenuItem
                                onClick={(e) => e.stopPropagation()}
                                style={{ color: 'var(--team-card-text-secondary)' }}
                                className="focus:bg-[var(--team-input-bg)] focus:text-[var(--team-card-text)]"
                              >
                                <Settings className="w-4 h-4 mr-2" />
                                Settings
                              </DropdownMenuItem>
                              <DropdownMenuSeparator style={{ backgroundColor: 'var(--team-card-border)' }} />
                              <DropdownMenuItem
                                className="text-red-400 focus:bg-red-500/10 focus:text-red-400"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  if (confirm(`Are you sure you want to delete "${team.name}"?`)) {
                                    deleteTeamMutation.mutate(team.id);
                                  }
                                }}
                              >
                                <Trash2 className="w-4 h-4 mr-2" />
                                Delete Team
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      </div>
                    </CardHeader>
                    {viewMode === 'grid' && (
                      <CardContent className="pt-0">
                        <div className="flex items-center justify-between text-xs text-gray-500">
                          <span>
                            Created{' '}
                            {formatDistanceToNow(new Date(team.created_at), { addSuffix: true })}
                          </span>
                          <ChevronRight
                            className={`w-4 h-4 transition-transform ${isSelected ? 'text-[#6366f1]' : ''}`}
                          />
                        </div>
                      </CardContent>
                    )}
                  </Card>
                </motion.div>
              );
            })}
          </AnimatePresence>
        </div>
      )}

      {/* Team Detail Panel */}
      <AnimatePresence mode="wait">
        {selectedTeam && (
          <motion.div
            key={selectedTeam.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.2 }}
            className="mt-8"
          >
            <Card style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-4">
                    <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 border border-[#6366f1]/20 flex items-center justify-center">
                      <Users className="w-7 h-7 text-[#6366f1]" />
                    </div>
                    <div>
                      <CardTitle className="text-xl" style={{ color: 'var(--team-card-text)' }}>{selectedTeam.name}</CardTitle>
                      <CardDescription style={{ color: 'var(--team-card-text-secondary)' }}>
                        Manage team members, invites, and settings
                      </CardDescription>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setSelectedTeam(null)}
                      className="border-[#30363d] text-gray-300 hover:bg-[#21262d]"
                    >
                      <X className="w-4 h-4 mr-2" />
                      Close
                    </Button>
                    <Dialog open={isInviteDialogOpen} onOpenChange={setIsInviteDialogOpen}>
                      <DialogTrigger asChild>
                        <Button size="sm" className="bg-[#6366f1] hover:bg-[#5558e0]">
                          <UserPlus className="w-4 h-4 mr-2" />
                          Invite
                        </Button>
                      </DialogTrigger>
                      <DialogContent className="[&>button]:text-gray-400" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)', color: 'var(--team-card-text)' }}>
                        <DialogHeader>
                          <DialogTitle style={{ color: 'var(--team-card-text)' }}>Invite Team Member</DialogTitle>
                          <DialogDescription style={{ color: 'var(--team-card-text-secondary)' }}>
                            Send an invitation to join {selectedTeam.name}.
                          </DialogDescription>
                        </DialogHeader>
                        <div className="py-4 space-y-4">
                          <div>
                            <Label htmlFor="invite-email" style={{ color: 'var(--team-card-text-secondary)' }}>
                              Email Address
                            </Label>
                            <Input
                              id="invite-email"
                              type="email"
                              placeholder="colleague@company.com"
                              value={inviteEmail}
                              onChange={(e) => setInviteEmail(e.target.value)}
                              style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-input-border)', color: 'var(--team-input-text)' }}
                              className="mt-2 placeholder:text-gray-500"
                            />
                          </div>
                          <div>
                            <Label htmlFor="invite-role" style={{ color: 'var(--team-card-text-secondary)' }}>
                              Role
                            </Label>
                            <Select
                              value={inviteRole}
                              onValueChange={(v) =>
                                setInviteRole(v as 'admin' | 'member' | 'viewer')
                              }
                            >
                              <SelectTrigger className="mt-2" style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-input-border)', color: 'var(--team-input-text)' }}>
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
                                <SelectItem value="admin" style={{ color: 'var(--team-card-text)' }} className="focus:bg-[var(--team-input-bg)]">
                                  <div className="flex items-center gap-2">
                                    <Shield className="w-4 h-4 text-purple-400" />
                                    Admin - Full access
                                  </div>
                                </SelectItem>
                                <SelectItem
                                  value="member"
                                  style={{ color: 'var(--team-card-text)' }}
                                  className="focus:bg-[var(--team-input-bg)]"
                                >
                                  <div className="flex items-center gap-2">
                                    <User className="w-4 h-4 text-blue-400" />
                                    Member - Can manage resources
                                  </div>
                                </SelectItem>
                                <SelectItem
                                  value="viewer"
                                  style={{ color: 'var(--team-card-text)' }}
                                  className="focus:bg-[var(--team-input-bg)]"
                                >
                                  <div className="flex items-center gap-2">
                                    <Eye className="w-4 h-4 text-gray-400" />
                                    Viewer - Read-only access
                                  </div>
                                </SelectItem>
                              </SelectContent>
                            </Select>
                          </div>
                        </div>
                        <DialogFooter>
                          <Button
                            variant="outline"
                            onClick={() => setIsInviteDialogOpen(false)}
                            style={{ borderColor: 'var(--team-card-border)', color: 'var(--team-card-text-secondary)' }}
                            className="hover:bg-[#21262d]"
                          >
                            Cancel
                          </Button>
                          <Button
                            onClick={handleInvite}
                            disabled={inviteMemberMutation.isPending || !inviteEmail.trim()}
                            className="bg-[#6366f1] hover:bg-[#5558e0]"
                          >
                            {inviteMemberMutation.isPending && (
                              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                            )}
                            Send Invitation
                          </Button>
                        </DialogFooter>
                      </DialogContent>
                    </Dialog>
                  </div>
                </div>
              </CardHeader>

              <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                <TabsList
                  style={
                    {
                      backgroundColor: 'var(--bg-secondary)',
                      borderColor: 'var(--border-default)',
                    } as React.CSSProperties
                  }
                  className="mx-6"
                >
                  <TabsTrigger
                    value="members"
                    style={
                      {
                        color: 'var(--text-secondary)',
                      } as React.CSSProperties
                    }
                    className="data-[state=active]:shadow-sm"
                  >
                    <Users className="w-4 h-4 mr-2" />
                    Members ({members.length})
                  </TabsTrigger>
                  <TabsTrigger
                    value="invites"
                    style={
                      {
                        color: 'var(--text-secondary)',
                      } as React.CSSProperties
                    }
                    className="data-[state=active]:shadow-sm"
                  >
                    <Mail className="w-4 h-4 mr-2" />
                    Invites ({invites.length})
                  </TabsTrigger>
                  <TabsTrigger
                    value="settings"
                    style={
                      {
                        color: 'var(--text-secondary)',
                      } as React.CSSProperties
                    }
                    className="data-[state=active]:shadow-sm"
                  >
                    <Settings className="w-4 h-4 mr-2" />
                    Settings
                  </TabsTrigger>
                </TabsList>

                <CardContent className="pt-6">
                  {/* Members Tab */}
                  <TabsContent value="members" className="mt-0">
                    {membersLoading ? (
                      <div className="flex items-center justify-center py-12">
                        <Loader2 className="w-6 h-6 animate-spin text-[#6366f1]" />
                      </div>
                    ) : members.length === 0 ? (
                      <div className="text-center py-12">
                        <Users className="w-12 h-12 text-gray-600 mx-auto mb-4" />
                        <h3 className="text-lg font-medium mb-2" style={{ color: 'var(--team-card-text)' }}>No members yet</h3>
                        <p className="mb-4" style={{ color: 'var(--team-card-text-secondary)' }}>
                          Invite your first team member to start collaborating.
                        </p>
                        <Button
                          onClick={() => setIsInviteDialogOpen(true)}
                          className="bg-[#6366f1] hover:bg-[#5558e0]"
                        >
                          <UserPlus className="w-4 h-4 mr-2" />
                          Invite Member
                        </Button>
                      </div>
                    ) : (
                      <div className="space-y-2">
                        {members.map((member) => {
                          const RoleIcon = roleIcons[member.role];
                          return (
                            <div
                              key={member.id}
                              style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-card-border)' }}
                              className="flex items-center justify-between p-4 rounded-lg border hover:border-[#6366f1]/30 transition-colors"
                            >
                              <div className="flex items-center gap-4">
                                <Avatar className="w-10 h-10 border" style={{ borderColor: 'var(--team-card-border)' }}>
                                  <AvatarFallback className="text-sm font-medium" style={{ backgroundColor: 'var(--team-card-bg)', color: 'var(--team-card-text)' }}>
                                    {getInitials(member.user?.name, member.user?.email)}
                                  </AvatarFallback>
                                </Avatar>
                                <div>
                                  <div className="font-medium flex items-center gap-2" style={{ color: 'var(--team-card-text)' }}>
                                    {member.user?.name ||
                                      member.user?.username ||
                                      member.user?.email}
                                    {isCurrentUserOwner(member) && (
                                      <Badge
                                        variant="outline"
                                        className="bg-[#6366f1]/10 text-[#6366f1] text-xs"
                                      >
                                        You
                                      </Badge>
                                    )}
                                  </div>
                                  <div className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>{member.user?.email}</div>
                                </div>
                              </div>
                              <div className="flex items-center gap-3">
                                <Badge
                                  variant="outline"
                                  className={`${roleColors[member.role as keyof typeof roleColors]} text-xs`}
                                >
                                  <RoleIcon className="w-3 h-3 mr-1" />
                                  {roleLabels[member.role as keyof typeof roleLabels]}
                                </Badge>
                                {!isCurrentUserOwner(member) && (
                                  <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                      <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8"
                                        style={{ color: 'var(--team-card-text-secondary)' }}
                                      >
                                        <MoreVertical className="w-4 h-4" />
                                      </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent
                                      align="end"
                                      style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}
                                    >
                                      <DropdownMenuItem
                                        onClick={() =>
                                          updateMemberMutation.mutate({
                                            teamId: selectedTeam.id,
                                            memberId: member.id,
                                            role: 'admin',
                                          })
                                        }
                                        style={{ color: 'var(--team-card-text-secondary)' }}
                                        className="focus:bg-[var(--team-input-bg)] focus:text-[var(--team-card-text)]"
                                      >
                                        <Shield className="w-4 h-4 mr-2 text-purple-400" />
                                        Make Admin
                                      </DropdownMenuItem>
                                      <DropdownMenuItem
                                        onClick={() =>
                                          updateMemberMutation.mutate({
                                            teamId: selectedTeam.id,
                                            memberId: member.id,
                                            role: 'member',
                                          })
                                        }
                                        style={{ color: 'var(--team-card-text-secondary)' }}
                                        className="focus:bg-[var(--team-input-bg)] focus:text-[var(--team-card-text)]"
                                      >
                                        <User className="w-4 h-4 mr-2 text-blue-400" />
                                        Make Member
                                      </DropdownMenuItem>
                                      <DropdownMenuItem
                                        onClick={() =>
                                          updateMemberMutation.mutate({
                                            teamId: selectedTeam.id,
                                            memberId: member.id,
                                            role: 'viewer',
                                          })
                                        }
                                        style={{ color: 'var(--team-card-text-secondary)' }}
                                        className="focus:bg-[var(--team-input-bg)] focus:text-[var(--team-card-text)]"
                                      >
                                        <Eye className="w-4 h-4 mr-2 text-gray-400" />
                                        Make Viewer
                                      </DropdownMenuItem>
                                      <DropdownMenuSeparator style={{ backgroundColor: 'var(--team-card-border)' }} />
                                      <DropdownMenuItem
                                        className="text-red-400 focus:bg-red-500/10 focus:text-red-400"
                                        onClick={() => {
                                          if (
                                            confirm(`Remove ${member.user?.email} from this team?`)
                                          ) {
                                            removeMemberMutation.mutate({
                                              teamId: selectedTeam.id,
                                              memberId: member.id,
                                            });
                                          }
                                        }}
                                      >
                                        <Trash2 className="w-4 h-4 mr-2" />
                                        Remove
                                      </DropdownMenuItem>
                                    </DropdownMenuContent>
                                  </DropdownMenu>
                                )}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </TabsContent>

                  {/* Invites Tab */}
                  <TabsContent value="invites" className="mt-0">
                    {invitesLoading ? (
                      <div className="flex items-center justify-center py-12">
                        <Loader2 className="w-6 h-6 animate-spin text-[#6366f1]" />
                      </div>
                    ) : invites.length === 0 ? (
                      <div className="text-center py-12">
                        <Mail className="w-12 h-12 text-gray-600 mx-auto mb-4" />
                        <h3 className="text-lg font-medium mb-2" style={{ color: 'var(--team-card-text)' }}>No pending invites</h3>
                        <p className="mb-4" style={{ color: 'var(--team-card-text-secondary)' }}>
                          Invite team members to see pending invitations here.
                        </p>
                        <Button
                          onClick={() => setIsInviteDialogOpen(true)}
                          className="bg-[#6366f1] hover:bg-[#5558e0]"
                        >
                          <UserPlus className="w-4 h-4 mr-2" />
                          Invite Member
                        </Button>
                      </div>
                    ) : (
                      <div className="space-y-2">
                        {invites.map((invite) => (
                          <div
                            key={invite.id}
                            style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-card-border)' }}
                            className="flex items-center justify-between p-4 rounded-lg border"
                          >
                            <div className="flex items-center gap-4">
                              <div className="w-10 h-10 rounded-lg bg-amber-500/10 flex items-center justify-center">
                                <Mail className="w-5 h-5 text-amber-400" />
                              </div>
                              <div>
                                <div className="font-medium" style={{ color: 'var(--team-card-text)' }}>{invite.email}</div>
                                <div className="text-sm flex items-center gap-2" style={{ color: 'var(--team-card-text-secondary)' }}>
                                  <Clock className="w-3 h-3" />
                                  Invited{' '}
                                  {formatDistanceToNow(new Date(invite.created_at), {
                                    addSuffix: true,
                                  })}
                                  • Expires {format(new Date(invite.expires_at), 'MMM d')}
                                </div>
                              </div>
                            </div>
                            <div className="flex items-center gap-3">
                              <Badge
                                variant="outline"
                                className="bg-amber-500/10 text-amber-400 border-amber-500/30"
                              >
                                <Clock className="w-3 h-3 mr-1" />
                                Pending
                              </Badge>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() =>
                                  cancelInviteMutation.mutate({
                                    teamId: selectedTeam.id,
                                    inviteId: invite.id,
                                  })
                                }
                                disabled={cancelInviteMutation.isPending}
                                className="hover:bg-red-500/10"
                                style={{ color: 'var(--team-card-text-secondary)' }}
                              >
                                {cancelInviteMutation.isPending ? (
                                  <Loader2 className="w-4 h-4 animate-spin" />
                                ) : (
                                  <X className="w-4 h-4" />
                                )}
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </TabsContent>

                  {/* Settings Tab */}
                  <TabsContent value="settings" className="mt-0">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
<Card style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-card-border)' }}>
                        <CardHeader>
                          <CardTitle className="flex items-center gap-2 text-base" style={{ color: 'var(--team-card-text)' }}>
                            <Shield className="w-5 h-5 text-[#6366f1]" />
                            Team Permissions
                          </CardTitle>
                          <CardDescription style={{ color: 'var(--team-card-text-secondary)' }}>
                            Understanding roles and access levels
                          </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                          {[
                            {
                              role: 'owner',
                              desc: 'Full control - manage members, settings, billing, and delete team',
                            },
                            {
                              role: 'admin',
                              desc: 'Can manage members and resources, but cannot delete team',
                            },
                            {
                              role: 'member',
                              desc: 'Can create and manage functions and resources',
                            },
                            {
                              role: 'viewer',
                              desc: 'Read-only access to view team resources and activity',
                            },
                          ].map(({ role, desc }) => {
                            const RoleIcon = roleIcons[role as keyof typeof roleIcons];
                            return (
                              <div key={role} className="flex items-start gap-3">
                                <div
                                  className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 ${
                                    role === 'owner'
                                      ? 'bg-yellow-500/10'
                                      : role === 'admin'
                                        ? 'bg-purple-500/10'
                                        : role === 'member'
                                          ? 'bg-blue-500/10'
                                          : 'bg-gray-500/10'
                                  }`}
                                >
                                  <RoleIcon
                                    className={`w-4 h-4 ${
                                      role === 'owner'
                                        ? 'text-yellow-400'
                                        : role === 'admin'
                                          ? 'text-purple-400'
                                          : role === 'member'
                                            ? 'text-blue-400'
                                            : 'text-gray-400'
                                    }`}
                                  />
                                </div>
                                <div>
                                  <div className="font-medium text-sm capitalize" style={{ color: 'var(--team-card-text)' }}>
                                    {role}
                                  </div>
                                  <div className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>{desc}</div>
                                </div>
                              </div>
                            );
                          })}
                        </CardContent>
                      </Card>

                      <Card style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-card-border)' }}>
                        <CardHeader>
                          <CardTitle className="flex items-center gap-2 text-base" style={{ color: 'var(--team-card-text)' }}>
                            <Settings className="w-5 h-5 text-[#6366f1]" />
                            Team Settings
                          </CardTitle>
                          <CardDescription style={{ color: 'var(--team-card-text-secondary)' }}>
                            Manage your team configuration
                          </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                          <div className="p-4 rounded-lg" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium" style={{ color: 'var(--team-card-text)' }}>Team Name</span>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-8 text-[#6366f1] hover:text-[#5558e0] hover:bg-[#6366f1]/10"
                              >
                                <Edit3 className="w-4 h-4 mr-2" />
                                Edit
                              </Button>
                            </div>
                            <p className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>{selectedTeam.name}</p>
                          </div>
                          <div className="p-4 rounded-lg" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium" style={{ color: 'var(--team-card-text)' }}>Team ID</span>
                              <Badge
                                variant="outline"
                                style={{ color: 'var(--team-card-text-secondary)', borderColor: 'var(--team-card-border)' }}
                                className="font-mono text-xs"
                              >
                                {selectedTeam.id.slice(0, 8)}...
                              </Badge>
                            </div>
                            <p className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>
                              Unique identifier for API access
                            </p>
                          </div>
                          <div className="p-4 rounded-lg" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium" style={{ color: 'var(--team-card-text)' }}>Created</span>
                            </div>
                            <p className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>
                              {format(new Date(selectedTeam.created_at), 'MMMM d, yyyy')}
                            </p>
                          </div>
                          <Separator style={{ backgroundColor: 'var(--team-card-border)' }} />
                          <div className="p-4 rounded-lg bg-red-500/5 border border-red-500/20">
                            <div className="flex items-center gap-2 mb-2">
                              <AlertCircle className="w-5 h-5 text-red-400" />
                              <span className="font-medium text-red-400">Danger Zone</span>
                            </div>
                            <p className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>
                              Deleting a team will remove all members and data. This cannot be undone.
                            </p>
                            <Button
                              variant="destructive"
                              size="sm"
                              onClick={() => {
                                if (
                                  confirm(
                                    `Are you sure you want to delete "${selectedTeam.name}"? This cannot be undone.`
                                  )
                                ) {
                                  deleteTeamMutation.mutate(selectedTeam.id);
                                }
                              }}
                              disabled={deleteTeamMutation.isPending}
                              className="bg-red-500/10 text-red-400 border border-red-500/30 hover:bg-red-500/20 hover:text-red-400"
                            >
                              {deleteTeamMutation.isPending ? (
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                              ) : (
                                <Trash2 className="w-4 h-4 mr-2" />
                              )}
                              Delete Team
                            </Button>
                          </div>
                        </CardContent>
                      </Card>

<Card style={{ backgroundColor: 'var(--team-input-bg)', borderColor: 'var(--team-card-border)' }}>
                        <CardHeader>
                          <CardTitle className="flex items-center gap-2 text-base" style={{ color: 'var(--team-card-text)' }}>
                            <Settings className="w-5 h-5 text-[#6366f1]" />
                            Team Settings
                          </CardTitle>
                          <CardDescription style={{ color: 'var(--team-card-text-secondary)' }}>
                            Manage your team configuration
                          </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                          <div className="p-4 rounded-lg" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium" style={{ color: 'var(--team-card-text)' }}>Team Name</span>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-8 text-[#6366f1] hover:text-[#5558e0] hover:bg-[#6366f1]/10"
                              >
                                <Edit3 className="w-4 h-4 mr-2" />
                                Edit
                              </Button>
                            </div>
                            <p className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>{selectedTeam.name}</p>
                          </div>
                          <div className="p-4 rounded-lg" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium" style={{ color: 'var(--team-card-text)' }}>Team ID</span>
                              <Badge
                                variant="outline"
                                style={{ color: 'var(--team-card-text-secondary)', borderColor: 'var(--team-card-border)' }}
                                className="font-mono text-xs"
                              >
                                {selectedTeam.id.slice(0, 8)}...
                              </Badge>
                            </div>
                            <p className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>
                              Unique identifier for API access
                            </p>
                          </div>
                          <div className="p-4 rounded-lg" style={{ backgroundColor: 'var(--team-card-bg)', borderColor: 'var(--team-card-border)' }}>
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium" style={{ color: 'var(--team-card-text)' }}>Created</span>
                            </div>
                            <p className="text-sm" style={{ color: 'var(--team-card-text-secondary)' }}>
                              {format(new Date(selectedTeam.created_at), 'MMMM d, yyyy')}
                            </p>
                          </div>
                          <Separator style={{ backgroundColor: 'var(--team-card-border)' }} />
                          <div className="p-4 rounded-lg bg-red-500/5 border border-red-500/20">
                            <div className="flex items-center gap-2 mb-2">
                              <AlertCircle className="w-5 h-5 text-red-400" />
                              <span className="font-medium text-red-400">Danger Zone</span>
                            </div>
                            <p className="text-sm mb-3" style={{ color: 'var(--team-card-text-secondary)' }}>
                              Deleting a team will remove all members and data. This cannot be undone.
                            </p>
                          </div>
                          <div className="p-4 rounded-lg bg-[#161b22] border border-[#30363d]">
                            <div className="flex items-center justify-between mb-2">
                              <span className="font-medium text-white">Created</span>
                            </div>
                            <p className="text-sm text-gray-400">
                              {format(new Date(selectedTeam.created_at), 'MMMM d, yyyy')}
                            </p>
                          </div>
                          <Separator className="bg-[#30363d]" />
                          <div className="p-4 rounded-lg bg-red-500/5 border border-red-500/20">
                            <div className="flex items-center gap-2 mb-2">
                              <AlertCircle className="w-5 h-5 text-red-400" />
                              <span className="font-medium text-red-400">Danger Zone</span>
                            </div>
                            <p className="text-sm text-gray-400 mb-3">
                              Deleting a team will remove all members and data. This cannot be
                              undone.
                            </p>
                            <Button
                              variant="destructive"
                              size="sm"
                              onClick={() => {
                                if (
                                  confirm(
                                    `Are you sure you want to delete "${selectedTeam.name}"? This cannot be undone.`
                                  )
                                ) {
                                  deleteTeamMutation.mutate(selectedTeam.id);
                                }
                              }}
                              disabled={deleteTeamMutation.isPending}
                              className="bg-red-500/10 text-red-400 border border-red-500/30 hover:bg-red-500/20 hover:text-red-400"
                            >
                              {deleteTeamMutation.isPending ? (
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                              ) : (
                                <Trash2 className="w-4 h-4 mr-2" />
                              )}
                              Delete Team
                            </Button>
                          </div>
                        </CardContent>
                      </Card>
                    </div>
                  </TabsContent>
                </CardContent>
              </Tabs>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export default TeamsPage;
