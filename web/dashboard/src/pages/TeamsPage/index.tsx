import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { 
  Users, Plus, Settings, Trash2, Mail, MoreVertical, 
  Crown, Shield, User, Eye, Loader2, AlertCircle, CheckCircle 
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthStore } from "@/stores/authStore";
import { teamsApi, type Team, type TeamMember, type TeamInvite } from "@/api/teams";
import { toast } from "sonner";
import { format } from "date-fns";

const roleIcons = {
  owner: Crown,
  admin: Shield,
  member: User,
  viewer: Eye,
};

const roleColors = {
  owner: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
  admin: "bg-purple-500/20 text-purple-400 border-purple-500/30",
  member: "bg-blue-500/20 text-blue-400 border-blue-500/30",
  viewer: "bg-gray-500/20 text-gray-400 border-gray-500/30",
};

export function TeamsPage() {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [newTeamName, setNewTeamName] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<"admin" | "member" | "viewer">("member");

  // Fetch teams
  const { data: teamsData, isLoading } = useQuery({
    queryKey: ["teams"],
    queryFn: () => teamsApi.list(),
  });

  const teams = teamsData?.teams ?? [];

  // Fetch members for selected team
  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ["team-members", selectedTeam?.id],
    queryFn: () => teamsApi.listMembers(selectedTeam!.id),
    enabled: !!selectedTeam,
  });

  // Fetch invites for selected team
  const { data: invitesData } = useQuery({
    queryKey: ["team-invites", selectedTeam?.id],
    queryFn: () => teamsApi.listInvites(selectedTeam!.id),
    enabled: !!selectedTeam,
  });

  const members = membersData?.members ?? [];
  const invites = invitesData?.invites ?? [];

  // Create team mutation
  const createTeamMutation = useMutation({
    mutationFn: (name: string) => teamsApi.create({ name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["teams"] });
      setIsCreateDialogOpen(false);
      setNewTeamName("");
      toast.success("Team created successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to create team");
    },
  });

  // Delete team mutation
  const deleteTeamMutation = useMutation({
    mutationFn: (teamId: string) => teamsApi.delete(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["teams"] });
      if (selectedTeam) setSelectedTeam(null);
      toast.success("Team deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to delete team");
    },
  });

  // Invite member mutation
  const inviteMemberMutation = useMutation({
    mutationFn: ({ teamId, email, role }: { teamId: string; email: string; role: string }) =>
      teamsApi.addMember(teamId, { email, role: role as "admin" | "member" | "viewer" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["team-members", selectedTeam?.id] });
      queryClient.invalidateQueries({ queryKey: ["team-invites", selectedTeam?.id] });
      setIsInviteDialogOpen(false);
      setInviteEmail("");
      setInviteRole("member");
      toast.success("Invitation sent successfully");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to send invitation");
    },
  });

  // Update member role mutation
  const updateMemberMutation = useMutation({
    mutationFn: ({ teamId, memberId, role }: { teamId: string; memberId: string; role: string }) =>
      teamsApi.updateMember(teamId, memberId, { role: role as "admin" | "member" | "viewer" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["team-members", selectedTeam?.id] });
      toast.success("Member role updated");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to update role");
    },
  });

  // Remove member mutation
  const removeMemberMutation = useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) =>
      teamsApi.removeMember(teamId, memberId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["team-members", selectedTeam?.id] });
      toast.success("Member removed from team");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to remove member");
    },
  });

  // Cancel invite mutation
  const cancelInviteMutation = useMutation({
    mutationFn: ({ teamId, inviteId }: { teamId: string; inviteId: string }) =>
      teamsApi.cancelInvite(teamId, inviteId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["team-invites", selectedTeam?.id] });
      toast.success("Invitation cancelled");
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to cancel invitation");
    },
  });

  const handleCreateTeam = () => {
    if (!newTeamName.trim()) {
      toast.error("Team name is required");
      return;
    }
    createTeamMutation.mutate(newTeamName);
  };

  const handleInvite = () => {
    if (!inviteEmail.trim() || !selectedTeam) {
      toast.error("Email is required");
      return;
    }
    inviteMemberMutation.mutate({ teamId: selectedTeam.id, email: inviteEmail, role: inviteRole });
  };

  const getInitials = (name?: string, email?: string) => {
    if (name) return name.split(" ").map(n => n[0]).join("").toUpperCase().slice(0, 2);
    if (email) return email[0].toUpperCase();
    return "?";
  };

  const isCurrentUserOwner = (member: TeamMember) => 
    member.user_id === user?.id && member.role === "owner";

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Teams</h1>
          <p className="text-text-secondary">Manage your team members and permissions</p>
        </div>
        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button className="btn-primary">
              <Plus className="w-4 h-4 mr-2" />
              Create Team
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Team</DialogTitle>
              <DialogDescription>
                Create a new team to organize members and manage permissions.
              </DialogDescription>
            </DialogHeader>
            <div className="py-4">
              <Label htmlFor="team-name">Team Name</Label>
              <Input
                id="team-name"
                placeholder="e.g., Engineering, Marketing"
                value={newTeamName}
                onChange={(e) => setNewTeamName(e.target.value)}
                className="mt-2"
              />
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button 
                onClick={handleCreateTeam} 
                disabled={createTeamMutation.isPending || !newTeamName.trim()}
              >
                {createTeamMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                Create Team
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Teams Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-text-muted" />
        </div>
      ) : teams.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Users className="w-12 h-12 mx-auto text-text-muted mb-4" />
            <h3 className="text-lg font-medium text-text-primary mb-2">No teams yet</h3>
            <p className="text-text-secondary mb-4">Create your first team to start collaborating.</p>
            <Button onClick={() => setIsCreateDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create Team
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {teams.map((team) => (
            <Card 
              key={team.id} 
              className={`cursor-pointer transition-all hover:shadow-md ${
                selectedTeam?.id === team.id ? "ring-2 ring-brand-500" : ""
              }`}
              onClick={() => setSelectedTeam(team)}
            >
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-lg bg-brand-500/20 flex items-center justify-center">
                      <Users className="w-5 h-5 text-brand-400" />
                    </div>
                    <div>
                      <CardTitle className="text-base">{team.name}</CardTitle>
                      <CardDescription>
                        {team.members?.length ?? 0} members
                      </CardDescription>
                    </div>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={(e) => {
                        e.stopPropagation();
                        // Edit team settings
                      }}>
                        <Settings className="w-4 h-4 mr-2" />
                        Settings
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem 
                        className="text-error"
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
              </CardHeader>
              <CardContent>
                <p className="text-xs text-text-muted">
                  Created {format(new Date(team.created_at), "MMM d, yyyy")}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Team Detail / Members */}
      {selectedTeam && (
        <div className="mt-8">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>{selectedTeam.name}</CardTitle>
                  <CardDescription>Manage team members and their roles</CardDescription>
                </div>
                <Dialog open={isInviteDialogOpen} onOpenChange={setIsInviteDialogOpen}>
                  <DialogTrigger asChild>
                    <Button>
                      <Mail className="w-4 h-4 mr-2" />
                      Invite Member
                    </Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Invite Team Member</DialogTitle>
                      <DialogDescription>
                        Send an invitation to join this team.
                      </DialogDescription>
                    </DialogHeader>
                    <div className="py-4 space-y-4">
                      <div>
                        <Label htmlFor="invite-email">Email Address</Label>
                        <Input
                          id="invite-email"
                          type="email"
                          placeholder="colleague@company.com"
                          value={inviteEmail}
                          onChange={(e) => setInviteEmail(e.target.value)}
                          className="mt-2"
                        />
                      </div>
                      <div>
                        <Label htmlFor="invite-role">Role</Label>
                        <Select value={inviteRole} onValueChange={(v) => setInviteRole(v as "admin" | "member" | "viewer")}>
                          <SelectTrigger className="mt-2">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="admin">Admin - Full access</SelectItem>
                            <SelectItem value="member">Member - Can manage resources</SelectItem>
                            <SelectItem value="viewer">Viewer - Read-only access</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    <DialogFooter>
                      <Button variant="outline" onClick={() => setIsInviteDialogOpen(false)}>
                        Cancel
                      </Button>
                      <Button 
                        onClick={handleInvite} 
                        disabled={inviteMemberMutation.isPending || !inviteEmail.trim()}
                      >
                        {inviteMemberMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                        Send Invitation
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </div>
            </CardHeader>
            <CardContent>
              {membersLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
                </div>
              ) : (
                <div className="space-y-4">
                  {/* Pending Invites */}
                  {invites.length > 0 && (
                    <div className="mb-6">
                      <h4 className="text-sm font-medium text-text-secondary mb-3">Pending Invitations</h4>
                      <div className="space-y-2">
                        {invites.map((invite) => (
                          <div 
                            key={invite.id}
                            className="flex items-center justify-between p-3 bg-card rounded-lg border"
                          >
                            <div className="flex items-center gap-3">
                              <div className="w-8 h-8 rounded-full bg-amber-500/20 flex items-center justify-center">
                                <Mail className="w-4 h-4 text-amber-400" />
                              </div>
                              <div>
                                <p className="text-sm font-medium text-text-primary">{invite.email}</p>
                                <p className="text-xs text-text-muted">
                                  Invited as {invite.role} • Expires {format(new Date(invite.expires_at), "MMM d")}
                                </p>
                              </div>
                            </div>
                            <div className="flex items-center gap-2">
                              <Badge variant="outline" className="bg-amber-500/10 text-amber-400">
                                Pending
                              </Badge>
                              <Button 
                                variant="ghost" 
                                size="sm"
                                onClick={() => cancelInviteMutation.mutate({ 
                                  teamId: selectedTeam.id, 
                                  inviteId: invite.id 
                                })}
                              >
                                Cancel
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Team Members */}
                  <h4 className="text-sm font-medium text-text-secondary mb-3">Team Members</h4>
                  {members.length === 0 ? (
                    <div className="text-center py-8">
                      <Users className="w-8 h-8 mx-auto text-text-muted mb-2" />
                      <p className="text-text-secondary">No members yet</p>
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {members.map((member) => {
                        const RoleIcon = roleIcons[member.role];
                        return (
                          <div 
                            key={member.id}
                            className="flex items-center justify-between p-3 bg-card rounded-lg border"
                          >
                            <div className="flex items-center gap-3">
                              <Avatar>
                                <AvatarFallback className="bg-brand-500/20 text-brand-400">
                                  {getInitials(member.user?.name, member.user?.email)}
                                </AvatarFallback>
                              </Avatar>
                              <div>
                                <p className="text-sm font-medium text-text-primary">
                                  {member.user?.name || member.user?.username || member.user?.email}
                                  {isCurrentUserOwner(member) && (
                                    <span className="ml-2 text-xs text-text-muted">(You)</span>
                                  )}
                                </p>
                                <p className="text-xs text-text-muted">{member.user?.email}</p>
                              </div>
                            </div>
                            <div className="flex items-center gap-2">
                              <Badge className={roleColors[member.role]}>
                                <RoleIcon className="w-3 h-3 mr-1" />
                                {member.role}
                              </Badge>
                              {!isCurrentUserOwner(member) && (
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <Button variant="ghost" size="sm">
                                      <MoreVertical className="w-4 h-4" />
                                    </Button>
                                  </DropdownMenuTrigger>
                                  <DropdownMenuContent align="end">
                                    <DropdownMenuItem 
                                      onClick={() => updateMemberMutation.mutate({
                                        teamId: selectedTeam.id,
                                        memberId: member.id,
                                        role: "admin"
                                      })}
                                    >
                                      <Shield className="w-4 h-4 mr-2" />
                                      Make Admin
                                    </DropdownMenuItem>
                                    <DropdownMenuItem 
                                      onClick={() => updateMemberMutation.mutate({
                                        teamId: selectedTeam.id,
                                        memberId: member.id,
                                        role: "member"
                                      })}
                                    >
                                      <User className="w-4 h-4 mr-2" />
                                      Make Member
                                    </DropdownMenuItem>
                                    <DropdownMenuItem 
                                      onClick={() => updateMemberMutation.mutate({
                                        teamId: selectedTeam.id,
                                        memberId: member.id,
                                        role: "viewer"
                                      })}
                                    >
                                      <Eye className="w-4 h-4 mr-2" />
                                      Make Viewer
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem 
                                      className="text-error"
                                      onClick={() => {
                                        if (confirm(`Remove ${member.user?.email} from this team?`)) {
                                          removeMemberMutation.mutate({
                                            teamId: selectedTeam.id,
                                            memberId: member.id
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
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

export default TeamsPage;
