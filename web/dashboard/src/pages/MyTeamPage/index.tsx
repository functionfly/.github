import "./styles.css";
import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import {
  Users,
  Settings,
  Mail,
  Crown,
  Shield,
  User,
  Eye,
  Loader2,
  ArrowRight,
  Brain,
  Activity,
  Clock,
  MoreVertical,
  UserPlus,
  LogOut,
  CreditCard,
  Bell,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useAuthStore } from "@/stores/authStore";
import { teamsApi, type Team, type TeamMember } from "@/api/teams";
import { useToast } from "@/components/ui/use-toast";
import { formatDistanceToNow } from "date-fns";

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

const roleLabels = {
  owner: "Owner",
  admin: "Admin",
  member: "Member",
  viewer: "Viewer",
};

// Mock activity data - replace with real API calls when available
const mockActivities = [
  {
    id: "1",
    type: "member_joined",
    user: "Sarah Chen",
    action: "joined the team",
    timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
    icon: UserPlus,
  },
  {
    id: "2",
    type: "memory_created",
    user: "Alex Rivera",
    action: "created a new team memory",
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
    icon: Brain,
  },
  {
    id: "3",
    type: "deployment",
    user: "System",
    action: "deployed 3 functions to production",
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 4).toISOString(),
    icon: Activity,
  },
  {
    id: "4",
    type: "settings_changed",
    user: "You",
    action: "updated team notification settings",
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
    icon: Settings,
  },
];

export function MyTeamPage() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [isLeaveDialogOpen, setIsLeaveDialogOpen] = useState(false);

  // Fetch user's teams
  const { data: teamsData, isLoading: teamsLoading } = useQuery({
    queryKey: ["teams"],
    queryFn: () => teamsApi.list(),
  });

  const teams = teamsData?.teams ?? [];
  const primaryTeam = teams[0];

  // Fetch members for primary team
  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ["team-members", primaryTeam?.id],
    queryFn: () => teamsApi.listMembers(primaryTeam!.id),
    enabled: !!primaryTeam,
  });

  const members = membersData?.members ?? [];

  // Get current user's role in the team
  const currentMember = members.find((m) => m.user_id === user?.id);
  const userRole = currentMember?.role || "viewer";
  const isOwner = userRole === "owner";
  const isAdmin = userRole === "admin" || isOwner;

  // Leave team mutation
  const leaveTeamMutation = useMutation({
    mutationFn: () => teamsApi.removeMember(primaryTeam!.id, user!.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["teams"] });
      toast({
        title: "Left team",
        description: "You have successfully left the team.",
      });
      navigate("/dashboard/teams");
    },
    onError: (error: Error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to leave team",
        variant: "destructive",
      });
    },
  });

  if (teamsLoading) {
    return (
      <div className="container mx-auto py-8 flex items-center justify-center min-h-[400px]">
        <Loader2 className="w-8 h-8 animate-spin text-[#6366f1]" />
      </div>
    );
  }

  if (!primaryTeam) {
    return (
      <div className="container mx-auto py-8">
        <Card className="bg-[#161b22] border-[#30363d]">
          <CardContent className="py-12 text-center">
            <Users className="w-12 h-12 text-gray-500 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-white mb-2">No Team Found</h3>
            <p className="text-gray-400 mb-6">
              You are not a member of any team yet. Create a team or accept an invitation to get started.
            </p>
            <Button
              onClick={() => navigate("/dashboard/teams")}
              className="bg-[#6366f1] hover:bg-[#5558e0]"
            >
              <Users className="w-4 h-4 mr-2" />
              Manage Teams
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">My Team</h1>
          <p className="text-gray-400 mt-1">Manage your team, members, and activity</p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => navigate("/dashboard/teams")}
            className="border-[#30363d] text-gray-300 hover:bg-[#21262d]"
          >
            <Settings className="w-4 h-4 mr-2" />
            All Teams
          </Button>
          {isAdmin && (
            <Button
              onClick={() => setIsInviteDialogOpen(true)}
              className="bg-[#6366f1] hover:bg-[#5558e0]"
            >
              <UserPlus className="w-4 h-4 mr-2" />
              Invite Member
            </Button>
          )}
        </div>
      </div>

      {/* Team Overview Card */}
      <Card className="bg-[#161b22] border-[#30363d]">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-4">
              <div className="w-16 h-16 rounded-xl bg-gradient-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 border border-[#6366f1]/20 flex items-center justify-center">
                <Users className="w-8 h-8 text-[#6366f1]" />
              </div>
              <div>
                <CardTitle className="text-xl text-white">{primaryTeam.name}</CardTitle>
                <CardDescription className="text-gray-400">
                  Created {formatDistanceToNow(new Date(primaryTeam.created_at), { addSuffix: true })}
                </CardDescription>
              </div>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="text-gray-400">
                  <MoreVertical className="w-5 h-5" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="bg-[#21262d] border-[#30363d]">
                <DropdownMenuItem
                  onClick={() => navigate(`/dashboard/teams/${primaryTeam.id}/memory`)}
                  className="text-gray-300 focus:bg-[#30363d] focus:text-white"
                >
                  <Brain className="w-4 h-4 mr-2" />
                  Team Memory
                </DropdownMenuItem>
                <DropdownMenuItem
                  onClick={() => navigate("/dashboard/teams")}
                  className="text-gray-300 focus:bg-[#30363d] focus:text-white"
                >
                  <Settings className="w-4 h-4 mr-2" />
                  Team Settings
                </DropdownMenuItem>
                <DropdownMenuSeparator className="bg-[#30363d]" />
                {!isOwner && (
                  <DropdownMenuItem
                    onClick={() => setIsLeaveDialogOpen(true)}
                    className="text-red-400 focus:bg-red-500/10 focus:text-red-400"
                  >
                    <LogOut className="w-4 h-4 mr-2" />
                    Leave Team
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="p-4 rounded-lg bg-[#0d1117] border border-[#30363d]">
              <div className="text-2xl font-bold text-white">{members.length}</div>
              <div className="text-sm text-gray-400">Team Members</div>
            </div>
            <div className="p-4 rounded-lg bg-[#0d1117] border border-[#30363d]">
              <div className="text-2xl font-bold text-[#6366f1]">{mockActivities.length}</div>
              <div className="text-sm text-gray-400">Activities Today</div>
            </div>
            <div className="p-4 rounded-lg bg-[#0d1117] border border-[#30363d]">
              <div className="text-2xl font-bold text-green-400">Active</div>
              <div className="text-sm text-gray-400">Team Status</div>
            </div>
            <div className="p-4 rounded-lg bg-[#0d1117] border border-[#30363d]">
              <div className="text-2xl font-bold text-yellow-400">{roleLabels[userRole as keyof typeof roleLabels]}</div>
              <div className="text-sm text-gray-400">Your Role</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Main Content Tabs */}
      <Tabs defaultValue="members" className="w-full">
        <TabsList className="bg-[#161b22] border border-[#30363d]">
          <TabsTrigger value="members" className="data-[state=active]:bg-[#21262d]">
            <Users className="w-4 h-4 mr-2" />
            Members
          </TabsTrigger>
          <TabsTrigger value="activity" className="data-[state=active]:bg-[#21262d]">
            <Activity className="w-4 h-4 mr-2" />
            Activity
          </TabsTrigger>
          <TabsTrigger value="settings" className="data-[state=active]:bg-[#21262d]">
            <Settings className="w-4 h-4 mr-2" />
            Settings
          </TabsTrigger>
        </TabsList>

        {/* Members Tab */}
        <TabsContent value="members" className="mt-6">
          <Card className="bg-[#161b22] border-[#30363d]">
            <CardHeader>
              <CardTitle className="text-white">Team Members</CardTitle>
              <CardDescription className="text-gray-400">
                {members.length} member{members.length !== 1 ? "s" : ""} in your team
              </CardDescription>
            </CardHeader>
            <CardContent>
              {membersLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin text-[#6366f1]" />
                </div>
              ) : (
                <div className="space-y-3">
                  {members.map((member) => {
                    const RoleIcon = roleIcons[member.role as keyof typeof roleIcons] || User;
                    return (
                      <div
                        key={member.id}
                        className="flex items-center justify-between p-4 rounded-lg bg-[#0d1117] border border-[#30363d] hover:border-[#6366f1]/30 transition-colors"
                      >
                        <div className="flex items-center gap-4">
                          <Avatar className="w-10 h-10 border border-[#30363d]">
                            <AvatarImage src={member.user?.avatar} />
                            <AvatarFallback className="bg-[#21262d] text-white">
                              {member.user?.name?.charAt(0) ||
                                member.user?.email?.charAt(0) ||
                                "?"}
                            </AvatarFallback>
                          </Avatar>
                          <div>
                            <div className="font-medium text-white">
                              {member.user?.name || member.user?.email}
                            </div>
                            <div className="text-sm text-gray-400">{member.user?.email}</div>
                          </div>
                        </div>
                        <div className="flex items-center gap-3">
                          <Badge
                            variant="outline"
                            className={roleColors[member.role as keyof typeof roleColors]}
                          >
                            <RoleIcon className="w-3 h-3 mr-1" />
                            {roleLabels[member.role as keyof typeof roleLabels]}
                          </Badge>
                          {member.user_id === user?.id && (
                            <Badge variant="outline" className="bg-[#6366f1]/10 text-[#6366f1]">
                              You
                            </Badge>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Activity Tab */}
        <TabsContent value="activity" className="mt-6">
          <Card className="bg-[#161b22] border-[#30363d]">
            <CardHeader>
              <CardTitle className="text-white">Recent Activity</CardTitle>
              <CardDescription className="text-gray-400">
                Latest actions and updates from your team
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {mockActivities.map((activity) => {
                  const ActivityIcon = activity.icon;
                  return (
                    <div
                      key={activity.id}
                      className="flex items-start gap-4 p-4 rounded-lg bg-[#0d1117] border border-[#30363d]"
                    >
                      <div className="w-10 h-10 rounded-lg bg-[#21262d] flex items-center justify-center shrink-0">
                        <ActivityIcon className="w-5 h-5 text-[#6366f1]" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-white">{activity.user}</span>
                          <span className="text-gray-400">{activity.action}</span>
                        </div>
                        <div className="flex items-center gap-2 mt-1 text-sm text-gray-500">
                          <Clock className="w-3 h-3" />
                          {formatDistanceToNow(new Date(activity.timestamp), { addSuffix: true })}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Settings Tab */}
        <TabsContent value="settings" className="mt-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card className="bg-[#161b22] border-[#30363d]">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Bell className="w-5 h-5 text-[#6366f1]" />
                  Notifications
                </CardTitle>
                <CardDescription className="text-gray-400">
                  Configure how you receive team updates
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between p-3 rounded-lg bg-[#0d1117] border border-[#30363d]">
                  <div>
                    <div className="font-medium text-white">Member Activity</div>
                    <div className="text-sm text-gray-400">When members join or leave</div>
                  </div>
                  <Badge className="bg-green-500/20 text-green-400">Enabled</Badge>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[#0d1117] border border-[#30363d]">
                  <div>
                    <div className="font-medium text-white">Team Memory Updates</div>
                    <div className="text-sm text-gray-400">When new memories are created</div>
                  </div>
                  <Badge className="bg-green-500/20 text-green-400">Enabled</Badge>
                </div>
                <div className="flex items-center justify-between p-3 rounded-lg bg-[#0d1117] border border-[#30363d]">
                  <div>
                    <div className="font-medium text-white">Deployment Alerts</div>
                    <div className="text-sm text-gray-400">Function deployment notifications</div>
                  </div>
                  <Badge className="bg-gray-500/20 text-gray-400">Disabled</Badge>
                </div>
              </CardContent>
            </Card>

            <Card className="bg-[#161b22] border-[#30363d]">
              <CardHeader>
                <CardTitle className="text-white flex items-center gap-2">
                  <Shield className="w-5 h-5 text-[#6366f1]" />
                  Permissions
                </CardTitle>
                <CardDescription className="text-gray-400">
                  Your current role and capabilities
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="p-4 rounded-lg bg-[#0d1117] border border-[#30363d]">
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-10 h-10 rounded-lg bg-[#6366f1]/20 flex items-center justify-center">
                      {(() => {
                        const RoleIcon = roleIcons[userRole as keyof typeof roleIcons] || User;
                        return <RoleIcon className="w-5 h-5 text-[#6366f1]" />;
                      })()}
                    </div>
                    <div>
                      <div className="font-medium text-white">
                        {roleLabels[userRole as keyof typeof roleLabels]}
                      </div>
                      <div className="text-sm text-gray-400">Current Role</div>
                    </div>
                  </div>
                  <Separator className="bg-[#30363d] my-3" />
                  <div className="space-y-2 text-sm">
                    <div className="flex items-center gap-2 text-gray-300">
                      <div className="w-1.5 h-1.5 rounded-full bg-green-400" />
                      View team members and activity
                    </div>
                    <div className="flex items-center gap-2 text-gray-300">
                      <div className="w-1.5 h-1.5 rounded-full bg-green-400" />
                      Access team memory
                    </div>
                    {(isAdmin || isOwner) && (
                      <>
                        <div className="flex items-center gap-2 text-gray-300">
                          <div className="w-1.5 h-1.5 rounded-full bg-green-400" />
                          Invite and manage members
                        </div>
                        <div className="flex items-center gap-2 text-gray-300">
                          <div className="w-1.5 h-1.5 rounded-full bg-green-400" />
                          Edit team settings
                        </div>
                      </>
                    )}
                    {isOwner && (
                      <div className="flex items-center gap-2 text-gray-300">
                        <div className="w-1.5 h-1.5 rounded-full bg-green-400" />
                        Transfer ownership or delete team
                      </div>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card
          className="bg-[#161b22] border-[#30363d] hover:border-[#6366f1]/50 transition-colors cursor-pointer group"
          onClick={() => navigate(`/dashboard/teams/${primaryTeam.id}/memory`)}
        >
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 flex items-center justify-center">
                  <Brain className="w-6 h-6 text-[#6366f1]" />
                </div>
                <div>
                  <h3 className="font-semibold text-white group-hover:text-[#6366f1] transition-colors">
                    Team Memory
                  </h3>
                  <p className="text-sm text-gray-400">Shared knowledge base</p>
                </div>
              </div>
              <ArrowRight className="w-5 h-5 text-gray-500 group-hover:text-[#6366f1] transition-colors" />
            </div>
          </CardContent>
        </Card>

        <Card
          className="bg-[#161b22] border-[#30363d] hover:border-[#6366f1]/50 transition-colors cursor-pointer group"
          onClick={() => navigate("/dashboard/teams")}
        >
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center">
                  <Settings className="w-6 h-6 text-green-400" />
                </div>
                <div>
                  <h3 className="font-semibold text-white group-hover:text-green-400 transition-colors">
                    Team Settings
                  </h3>
                  <p className="text-sm text-gray-400">Configure your team</p>
                </div>
              </div>
              <ArrowRight className="w-5 h-5 text-gray-500 group-hover:text-green-400 transition-colors" />
            </div>
          </CardContent>
        </Card>

        <Card
          className="bg-[#161b22] border-[#30363d] hover:border-[#6366f1]/50 transition-colors cursor-pointer group"
          onClick={() => navigate("/dashboard/billing")}
        >
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-yellow-500/20 to-orange-500/20 flex items-center justify-center">
                  <CreditCard className="w-6 h-6 text-yellow-400" />
                </div>
                <div>
                  <h3 className="font-semibold text-white group-hover:text-yellow-400 transition-colors">
                    Billing
                  </h3>
                  <p className="text-sm text-gray-400">Manage subscription</p>
                </div>
              </div>
              <ArrowRight className="w-5 h-5 text-gray-500 group-hover:text-yellow-400 transition-colors" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Leave Team Dialog */}
      <Dialog open={isLeaveDialogOpen} onOpenChange={setIsLeaveDialogOpen}>
        <DialogContent className="bg-[#161b22] border-[#30363d]">
          <DialogHeader>
            <DialogTitle className="text-white">Leave Team?</DialogTitle>
            <DialogDescription className="text-gray-400">
              Are you sure you want to leave <strong>{primaryTeam.name}</strong>? You will lose
              access to all team resources and memories.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsLeaveDialogOpen(false)}
              className="border-[#30363d] text-gray-300 hover:bg-[#21262d]"
            >
              Cancel
            </Button>
            <Button
              onClick={() => leaveTeamMutation.mutate()}
              disabled={leaveTeamMutation.isPending}
              className="bg-red-500 hover:bg-red-600"
            >
              {leaveTeamMutation.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <LogOut className="w-4 h-4 mr-2" />
              )}
              Leave Team
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}


