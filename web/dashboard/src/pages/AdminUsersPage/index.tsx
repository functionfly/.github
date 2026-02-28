import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Search, MoreVertical, User, Mail, Shield, ArrowLeft } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { StatCard } from "@/components/common/StatCard";
import { adminUsersApi, type AdminUser, type AdminUserStats } from "@/api/admin";
import type { User as UserType } from "@/types";

// Convert AdminUser to User type for compatibility
const convertAdminUserToUser = (adminUser: AdminUser): UserType => ({
  id: adminUser.id,
  email: adminUser.email,
  name: adminUser.name,
  tenantId: adminUser.tenant_id,
  plan: adminUser.plan as any,
  role: adminUser.role,
  createdAt: adminUser.created_at,
  updatedAt: adminUser.updated_at,
});

// Convert AdminUserStats to stats format
const convertUserStatsToStats = (stats: AdminUserStats) => [
  {
    title: "Total Users",
    value: stats.total_users,
    change: { value: 0, label: "from last month" }, // TODO: Add real change calculation
    icon: <User className="w-5 h-5 text-[#6366f1]" />,
    trend: "neutral" as const,
  },
  {
    title: "Admin Users",
    value: stats.admin_users,
    change: { value: 0, label: "from last month" }, // TODO: Add real change calculation
    icon: <Shield className="w-5 h-5 text-[#6366f1]" />,
    trend: "neutral" as const,
  },
  {
    title: "Active Users",
    value: stats.active_users,
    change: { value: 0, label: "from last month" }, // TODO: Add real change calculation
    icon: <Mail className="w-5 h-5 text-[#6366f1]" />,
    trend: "neutral" as const,
  },
];

export function AdminUsersPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState("");
  const [roleFilter, setRoleFilter] = useState<string>("all");
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("");
  const [selectedUser, setSelectedUser] = useState<UserType | null>(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editName, setEditName] = useState("");
  const [editEmail, setEditEmail] = useState("");
  const [editRole, setEditRole] = useState("");
  const [editPlan, setEditPlan] = useState("");

  // Fetch users data
  const { data: usersData, isLoading: usersLoading } = useQuery({
    queryKey: ["admin-users", searchTerm, roleFilter],
    queryFn: () => adminUsersApi.listUsers({
      search: searchTerm || undefined,
      role: roleFilter === "all" ? undefined : roleFilter,
      limit: 100, // TODO: Add pagination
    }),
  });

  // Fetch user stats
  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ["admin-user-stats"],
    queryFn: () => adminUsersApi.getUserStats(),
  });

  // Convert data to expected formats
  const users = usersData?.users.map(convertAdminUserToUser) || [];
  const stats = statsData ? convertUserStatsToStats(statsData) : [];

  // Filter users (additional client-side filtering beyond API)
  const filteredUsers = users.filter((user) => {
    // API already handles search and role filtering, but we can add additional filtering here if needed
    return true;
  });

  // Mutations
  const createUserMutation = useMutation({
    mutationFn: (userData: { email: string; name?: string; tenant_id: string; plan: string; role?: string }) =>
      adminUsersApi.createUser(userData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      queryClient.invalidateQueries({ queryKey: ["admin-user-stats"] });
      setIsInviteDialogOpen(false);
      setInviteEmail("");
      setInviteRole("");
    },
  });

  const updateUserMutation = useMutation({
    mutationFn: ({ userId, updates }: { userId: string; updates: Partial<AdminUser> }) =>
      adminUsersApi.updateUser(userId, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      setIsEditDialogOpen(false);
      setSelectedUser(null);
    },
  });

  const handleInviteUser = () => {
    if (inviteEmail.trim()) {
      createUserMutation.mutate({
        email: inviteEmail.trim(),
        name: inviteEmail.split('@')[0],
        tenant_id: "559f8076-d1cf-484f-9fc0-77bb54e82e2b", // Default tenant - TODO: Make this configurable
        plan: "starter",
        role: inviteRole || undefined,
      });
    }
  };

  const handleViewDetails = (user: UserType) => {
    setSelectedUser(user);
    setIsDetailsOpen(true);
  };

  const handleEditUser = (user: UserType) => {
    setSelectedUser(user);
    setEditName(user.name || "");
    setEditEmail(user.email);
    setEditRole(user.role || "none");
    setEditPlan(user.plan);
    setIsEditDialogOpen(true);
  };

  const handleSaveEdit = () => {
    if (selectedUser) {
      updateUserMutation.mutate({
        userId: selectedUser.id,
        updates: {
          name: editName.trim() || undefined,
          email: editEmail.trim(),
          role: editRole === "none" ? undefined : editRole,
          plan: editPlan,
        },
      });
    }
  };

  const getRoleBadgeColor = (role?: string) => {
    switch (role) {
      case "super_admin":
        return "bg-red-500/10 text-red-400";
      case "support":
        return "bg-blue-500/10 text-blue-400";
      case "billing_admin":
        return "bg-green-500/10 text-green-400";
      case "developer_admin":
        return "bg-purple-500/10 text-purple-400";
      default:
        return "bg-gray-500/10 text-text-secondary";
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => navigate('/admin')}
          className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Dashboard
        </Button>
        <div className="flex-1 text-center">
          <h1 className="text-2xl font-bold text-text-primary">Users</h1>
          <p className="text-text-secondary">Manage all users and their permissions</p>
        </div>
        <Dialog open={isInviteDialogOpen} onOpenChange={setIsInviteDialogOpen}>
          <DialogTrigger asChild>
            <Button className="bg-[#6366f1] hover:bg-[#5855eb]">
              <Plus className="w-4 h-4 mr-2" />
              Invite User
            </Button>
          </DialogTrigger>
          <DialogContent className="bg-bg-tertiary border-white/8">
            <DialogHeader>
              <DialogTitle className="text-text-primary">Invite New User</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <Label htmlFor="invite-email" className="text-text-primary">Email Address</Label>
                <Input
                  id="invite-email"
                  type="email"
                  placeholder="user@example.com"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  className="bg-bg-secondary border-border-default text-text-primary"
                />
              </div>
              <div>
                <Label htmlFor="invite-role" className="text-text-primary">Role</Label>
                <Select value={inviteRole} onValueChange={setInviteRole}>
                  <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary">
                    <SelectValue placeholder="Select a role" />
                  </SelectTrigger>
                  <SelectContent className="bg-bg-tertiary border-white/8">
                    <SelectItem value="support">Support</SelectItem>
                    <SelectItem value="billing_admin">Billing Admin</SelectItem>
                    <SelectItem value="developer_admin">Developer Admin</SelectItem>
                    <SelectItem value="read_only">Read Only</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsInviteDialogOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleInviteUser}
                  disabled={!inviteEmail.trim() || !inviteRole}
                  className="bg-[#6366f1] hover:bg-[#5855eb]"
                >
                  Send Invite
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>

        {/* Edit User Dialog */}
        <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
          <DialogContent className="bg-bg-tertiary border-white/8">
            <DialogHeader>
              <DialogTitle className="text-text-primary">Edit User</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <Label htmlFor="edit-name" className="text-text-primary">Name</Label>
                <Input
                  id="edit-name"
                  type="text"
                  placeholder="User name"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="bg-bg-secondary border-border-default text-text-primary"
                />
              </div>
              <div>
                <Label htmlFor="edit-email" className="text-text-primary">Email Address</Label>
                <Input
                  id="edit-email"
                  type="email"
                  placeholder="user@example.com"
                  value={editEmail}
                  onChange={(e) => setEditEmail(e.target.value)}
                  className="bg-bg-secondary border-border-default text-text-primary"
                />
              </div>
              <div>
                <Label htmlFor="edit-role" className="text-text-primary">Role</Label>
                <Select value={editRole} onValueChange={setEditRole}>
                  <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary">
                    <SelectValue placeholder="Select a role" />
                  </SelectTrigger>
                  <SelectContent className="bg-bg-tertiary border-white/8">
                    <SelectItem value="none">No Role</SelectItem>
                    <SelectItem value="support">Support</SelectItem>
                    <SelectItem value="billing_admin">Billing Admin</SelectItem>
                    <SelectItem value="developer_admin">Developer Admin</SelectItem>
                    <SelectItem value="super_admin">Super Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="edit-plan" className="text-text-primary">Plan</Label>
                <Select value={editPlan} onValueChange={setEditPlan}>
                  <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary">
                    <SelectValue placeholder="Select a plan" />
                  </SelectTrigger>
                  <SelectContent className="bg-bg-tertiary border-white/8">
                    <SelectItem value="free">Free</SelectItem>
                    <SelectItem value="starter">Starter</SelectItem>
                    <SelectItem value="pro">Pro</SelectItem>
                    <SelectItem value="enterprise">Enterprise</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsEditDialogOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleSaveEdit}
                  className="bg-[#6366f1] hover:bg-[#5855eb]"
                >
                  Save Changes
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {statsLoading ? (
          // Loading skeletons for stats
          Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="bg-bg-secondary border-white/8">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div className="space-y-2">
                    <Skeleton className="h-4 w-24 bg-white/10" />
                    <Skeleton className="h-8 w-16 bg-white/10" />
                  </div>
                  <Skeleton className="h-10 w-10 rounded-lg bg-white/10" />
                </div>
              </CardContent>
            </Card>
          ))
        ) : (
          stats.map((stat) => (
            <StatCard key={stat.title} {...stat} />
          ))
        )}
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  placeholder="Search users..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-bg-secondary border-border-default text-text-primary"
                />
              </div>
            </div>
            <Select value={roleFilter} onValueChange={setRoleFilter}>
              <SelectTrigger className="w-full sm:w-[180px] bg-bg-secondary border-border-default text-text-primary">
                <SelectValue placeholder="Filter by role" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-white/8">
                <SelectItem value="all">All Roles</SelectItem>
                <SelectItem value="super_admin">Super Admin</SelectItem>
                <SelectItem value="support">Support</SelectItem>
                <SelectItem value="billing_admin">Billing Admin</SelectItem>
                <SelectItem value="developer_admin">Developer Admin</SelectItem>
                <SelectItem value="read_only">Read Only</SelectItem>
                <SelectItem value="none">Regular Users</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Users List */}
      <Card>
        <CardHeader>
          <CardTitle>All Users ({usersLoading ? "..." : filteredUsers.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {usersLoading ? (
              // Loading skeletons for users
              Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-white/8">
                  <div className="flex items-center gap-4">
                    <Skeleton className="w-10 h-10 rounded-lg bg-white/10" />
                    <div className="space-y-2">
                      <Skeleton className="h-4 w-32 bg-white/10" />
                      <Skeleton className="h-3 w-48 bg-white/10" />
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right space-y-2">
                      <Skeleton className="h-5 w-20 bg-white/10" />
                      <Skeleton className="h-3 w-16 bg-white/10" />
                    </div>
                    <Skeleton className="w-8 h-8 rounded bg-white/10" />
                  </div>
                </div>
              ))
            ) : filteredUsers.length === 0 ? (
              <div className="text-center py-8 text-text-muted">
                No users found matching your criteria.
              </div>
            ) : (
              filteredUsers.map((user) => (
              <div
                key={user.id}
                className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-white/8 hover:bg-white/5 transition-colors"
              >
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-lg bg-[#6366f1]/10 flex items-center justify-center">
                    <User className="w-5 h-5 text-[#6366f1]" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{user.name || user.email}</p>
                    <p className="text-sm text-text-secondary">{user.email}</p>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  <div className="text-right">
                    {user.role ? (
                      <Badge className={getRoleBadgeColor(user.role)}>
                        {user.role.replace('_', ' ')}
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="border-border-default text-text-secondary">
                        Regular User
                      </Badge>
                    )}
                    <p className="text-xs text-text-secondary mt-1">
                      Plan: {user.plan}
                    </p>
                  </div>

                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm" className="text-text-muted hover:text-text-primary">
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="bg-bg-tertiary border-white/8">
                      <DropdownMenuItem
                        className="text-text-primary hover:bg-bg-hover"
                        onClick={() => handleViewDetails(user)}
                      >
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="text-text-primary hover:bg-bg-hover"
                        onClick={() => handleEditUser(user)}
                      >
                        Edit User
                      </DropdownMenuItem>
                      <DropdownMenuItem className="text-text-primary hover:bg-bg-hover">
                        Reset Password
                      </DropdownMenuItem>
                      <DropdownMenuItem className="text-red-400 hover:bg-red-500/10">
                        Disable User
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            ))
            )}
          </div>
        </CardContent>
      </Card>

      {/* User Details Dialog */}
      <Dialog open={isDetailsOpen} onOpenChange={setIsDetailsOpen}>
        <DialogContent className="bg-bg-tertiary border-white/8 max-w-2xl">
          <DialogHeader>
            <DialogTitle className="text-text-primary">
              User Details - {selectedUser?.name || selectedUser?.email}
            </DialogTitle>
          </DialogHeader>
          {selectedUser && (
            <div className="space-y-6">
              {/* User Info */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-secondary">User ID</Label>
                  <p className="text-text-primary font-mono text-sm">{selectedUser.id}</p>
                </div>
                <div>
                  <Label className="text-text-secondary">Role</Label>
                  <div className="mt-1">
                    {selectedUser.role ? (
                      <Badge className={getRoleBadgeColor(selectedUser.role)}>
                        {selectedUser.role.replace('_', ' ')}
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="border-border-default text-text-secondary">
                        Regular User
                      </Badge>
                    )}
                  </div>
                </div>
                <div>
                  <Label className="text-text-secondary">Email</Label>
                  <p className="text-text-primary">{selectedUser.email}</p>
                </div>
                <div>
                  <Label className="text-text-secondary">Plan</Label>
                  <p className="text-text-primary capitalize">{selectedUser.plan}</p>
                </div>
                <div>
                  <Label className="text-text-secondary">Tenant ID</Label>
                  <p className="text-text-primary font-mono text-sm">{selectedUser.tenantId}</p>
                </div>
                <div>
                  <Label className="text-text-secondary">Name</Label>
                  <p className="text-text-primary">{selectedUser.name || 'Not set'}</p>
                </div>
              </div>

              {/* Dates */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-text-secondary">Created At</Label>
                  <p className="text-text-primary">
                    {new Date(selectedUser.createdAt).toLocaleString()}
                  </p>
                </div>
                <div>
                  <Label className="text-text-secondary">Last Updated</Label>
                  <p className="text-text-primary">
                    {selectedUser.updatedAt ? new Date(selectedUser.updatedAt).toLocaleString() : 'Never'}
                  </p>
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
