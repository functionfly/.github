import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Plus,
  Search,
  MoreVertical,
  User,
  Mail,
  Shield,
  ArrowLeft,
  ChevronUp,
  ChevronDown,
  Filter,
  Download,
  RefreshCw,
  UserCheck,
  UserX,
  Crown,
  Eye,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuSeparator } from "@/components/ui/dropdown-menu";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { adminUsersApi, type AdminUser, type AdminUserStats } from "@/api/admin";
import type { User as UserType } from "@/types";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

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

const ROLE_CONFIG: Record<string, { label: string; color: string; icon: React.ComponentType<{ className?: string }> }> = {
  super_admin: { label: "Super Admin", color: "bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20", icon: Crown },
  support: { label: "Support", color: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20", icon: UserCheck },
  billing_admin: { label: "Billing Admin", color: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20", icon: Shield },
  developer_admin: { label: "Dev Admin", color: "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20", icon: Shield },
  read_only: { label: "Read Only", color: "bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20", icon: Eye },
};

const PLAN_CONFIG: Record<string, { label: string; color: string }> = {
  free: { label: "Free", color: "bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/20" },
  starter: { label: "Starter", color: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20" },
  pro: { label: "Pro", color: "bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20" },
  enterprise: { label: "Enterprise", color: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20" },
};

type SortField = "name" | "email" | "role" | "plan" | "createdAt";
type SortDir = "asc" | "desc";

export function AdminUsersPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState("");
  const [roleFilter, setRoleFilter] = useState<string>("all");
  const [planFilter, setPlanFilter] = useState<string>("all");
  const [sortField, setSortField] = useState<SortField>("createdAt");
  const [sortDir, setSortDir] = useState<SortDir>("desc");
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
  const { data: usersData, isLoading: usersLoading, refetch } = useQuery({
    queryKey: ["admin-users", searchTerm, roleFilter],
    queryFn: () => adminUsersApi.listUsers({
      search: searchTerm || undefined,
      role: roleFilter === "all" ? undefined : roleFilter,
      limit: 100,
    }),
  });

  // Fetch user stats
  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ["admin-user-stats"],
    queryFn: () => adminUsersApi.getUserStats(),
  });

  const users = usersData?.users.map(convertAdminUserToUser) || [];

  // Sort and filter
  const filteredUsers = users
    .filter((user) => {
      if (planFilter !== "all" && user.plan !== planFilter) return false;
      return true;
    })
    .sort((a, b) => {
      let aVal = "";
      let bVal = "";
      switch (sortField) {
        case "name": aVal = a.name || a.email; bVal = b.name || b.email; break;
        case "email": aVal = a.email; bVal = b.email; break;
        case "role": aVal = a.role || ""; bVal = b.role || ""; break;
        case "plan": aVal = a.plan; bVal = b.plan; break;
        case "createdAt": aVal = a.createdAt || ""; bVal = b.createdAt || ""; break;
      }
      const cmp = aVal.localeCompare(bVal);
      return sortDir === "asc" ? cmp : -cmp;
    });

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDir(d => d === "asc" ? "desc" : "asc");
    } else {
      setSortField(field);
      setSortDir("asc");
    }
  };

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
      toast.success("User invited successfully");
    },
    onError: () => toast.error("Failed to invite user"),
  });

  const updateUserMutation = useMutation({
    mutationFn: ({ userId, updates }: { userId: string; updates: Partial<AdminUser> }) =>
      adminUsersApi.updateUser(userId, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      setIsEditDialogOpen(false);
      setSelectedUser(null);
      toast.success("User updated successfully");
    },
    onError: () => toast.error("Failed to update user"),
  });

  const handleInviteUser = () => {
    if (inviteEmail.trim()) {
      createUserMutation.mutate({
        email: inviteEmail.trim(),
        name: inviteEmail.split('@')[0],
        tenant_id: "559f8076-d1cf-484f-9fc0-77bb54e82e2b",
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

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return <ChevronUp className="w-3 h-3 opacity-30" />;
    return sortDir === "asc"
      ? <ChevronUp className="w-3 h-3 text-brand-500" />
      : <ChevronDown className="w-3 h-3 text-brand-500" />;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            onClick={() => navigate('/admin')}
            className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Users</h1>
            <p className="text-sm text-text-secondary">Manage all users and their permissions</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="border-border-default hover:bg-bg-hover text-text-secondary"
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
          <Dialog open={isInviteDialogOpen} onOpenChange={setIsInviteDialogOpen}>
            <DialogTrigger asChild>
              <Button className="bg-brand-500 hover:bg-brand-600 text-white">
                <Plus className="w-4 h-4 mr-2" />
                Invite User
              </Button>
            </DialogTrigger>
            <DialogContent className="bg-bg-tertiary border-border-default">
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
                    className="bg-bg-secondary border-border-default text-text-primary mt-1"
                  />
                </div>
                <div>
                  <Label htmlFor="invite-role" className="text-text-primary">Role</Label>
                  <Select value={inviteRole} onValueChange={setInviteRole}>
                    <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary mt-1">
                      <SelectValue placeholder="Select a role" />
                    </SelectTrigger>
                    <SelectContent className="bg-bg-tertiary border-border-default">
                      <SelectItem value="support">Support</SelectItem>
                      <SelectItem value="billing_admin">Billing Admin</SelectItem>
                      <SelectItem value="developer_admin">Developer Admin</SelectItem>
                      <SelectItem value="read_only">Read Only</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex justify-end gap-2 pt-2">
                  <Button variant="outline" onClick={() => setIsInviteDialogOpen(false)} className="border-border-default">
                    Cancel
                  </Button>
                  <Button
                    onClick={handleInviteUser}
                    disabled={!inviteEmail.trim() || !inviteRole || createUserMutation.isPending}
                    className="bg-brand-500 hover:bg-brand-600 text-white"
                  >
                    Send Invite
                  </Button>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {statsLoading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="glass-card">
              <CardContent className="p-5">
                <Skeleton className="h-4 w-24 mb-2" />
                <Skeleton className="h-8 w-16" />
              </CardContent>
            </Card>
          ))
        ) : statsData ? (
          [
            { label: "Total Users", value: statsData.total_users, icon: User, color: "text-blue-500", bg: "bg-blue-500/10" },
            { label: "Active Users", value: statsData.active_users, icon: UserCheck, color: "text-emerald-500", bg: "bg-emerald-500/10" },
            { label: "Admin Users", value: statsData.admin_users, icon: Crown, color: "text-red-500", bg: "bg-red-500/10" },
            { label: "Inactive Users", value: statsData.total_users - statsData.active_users, icon: UserX, color: "text-slate-500", bg: "bg-slate-500/10" },
          ].map((stat) => (
            <Card key={stat.label} className="glass-card hover-lift">
              <CardContent className="p-5">
                <div className="flex items-center justify-between mb-2">
                  <p className="text-xs font-medium text-text-secondary">{stat.label}</p>
                  <div className={cn("p-1.5 rounded-lg", stat.bg)}>
                    <stat.icon className={cn("w-4 h-4", stat.color)} />
                  </div>
                </div>
                <p className="text-2xl font-bold text-text-primary">{stat.value.toLocaleString()}</p>
              </CardContent>
            </Card>
          ))
        ) : null}
      </div>

      {/* Filters */}
      <Card className="glass-card">
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <Input
                placeholder="Search by name or email..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 bg-bg-secondary border-border-default text-text-primary"
              />
            </div>
            <Select value={roleFilter} onValueChange={setRoleFilter}>
              <SelectTrigger className="w-full sm:w-[160px] bg-bg-secondary border-border-default text-text-primary">
                <Filter className="w-4 h-4 mr-2 text-text-muted" />
                <SelectValue placeholder="Role" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-border-default">
                <SelectItem value="all">All Roles</SelectItem>
                <SelectItem value="super_admin">Super Admin</SelectItem>
                <SelectItem value="support">Support</SelectItem>
                <SelectItem value="billing_admin">Billing Admin</SelectItem>
                <SelectItem value="developer_admin">Developer Admin</SelectItem>
                <SelectItem value="read_only">Read Only</SelectItem>
                <SelectItem value="none">Regular Users</SelectItem>
              </SelectContent>
            </Select>
            <Select value={planFilter} onValueChange={setPlanFilter}>
              <SelectTrigger className="w-full sm:w-[140px] bg-bg-secondary border-border-default text-text-primary">
                <SelectValue placeholder="Plan" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-border-default">
                <SelectItem value="all">All Plans</SelectItem>
                <SelectItem value="free">Free</SelectItem>
                <SelectItem value="starter">Starter</SelectItem>
                <SelectItem value="pro">Pro</SelectItem>
                <SelectItem value="enterprise">Enterprise</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Users Table */}
      <Card className="glass-card">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-text-primary">
              Users
              <span className="ml-2 text-sm font-normal text-text-muted">
                ({usersLoading ? "..." : filteredUsers.length})
              </span>
            </CardTitle>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {/* Table Header */}
          <div className="hidden md:grid grid-cols-[2fr_2fr_1fr_1fr_auto] gap-4 px-6 py-3 border-b border-border-subtle text-xs font-semibold text-text-muted uppercase tracking-wider">
            <button
              className="flex items-center gap-1 hover:text-text-primary transition-colors text-left"
              onClick={() => handleSort("name")}
            >
              User <SortIcon field="name" />
            </button>
            <button
              className="flex items-center gap-1 hover:text-text-primary transition-colors text-left"
              onClick={() => handleSort("email")}
            >
              Email <SortIcon field="email" />
            </button>
            <button
              className="flex items-center gap-1 hover:text-text-primary transition-colors text-left"
              onClick={() => handleSort("role")}
            >
              Role <SortIcon field="role" />
            </button>
            <button
              className="flex items-center gap-1 hover:text-text-primary transition-colors text-left"
              onClick={() => handleSort("plan")}
            >
              Plan <SortIcon field="plan" />
            </button>
            <span>Actions</span>
          </div>

          <div className="divide-y divide-border-subtle">
            {usersLoading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-center gap-4 px-6 py-4">
                  <Skeleton className="w-9 h-9 rounded-full" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-48" />
                  </div>
                  <Skeleton className="h-6 w-20" />
                  <Skeleton className="h-6 w-16" />
                  <Skeleton className="w-8 h-8 rounded" />
                </div>
              ))
            ) : filteredUsers.length === 0 ? (
              <div className="text-center py-12 text-text-muted">
                <User className="w-10 h-10 mx-auto mb-3 opacity-30" />
                <p className="font-medium">No users found</p>
                <p className="text-sm mt-1">Try adjusting your search or filters</p>
              </div>
            ) : (
              filteredUsers.map((user) => {
                const roleConfig = user.role ? ROLE_CONFIG[user.role] : null;
                const planConfig = PLAN_CONFIG[user.plan] || PLAN_CONFIG.free;
                const RoleIcon = roleConfig?.icon || User;

                return (
                  <div
                    key={user.id}
                    className="flex flex-col md:grid md:grid-cols-[2fr_2fr_1fr_1fr_auto] gap-3 md:gap-4 px-6 py-4 hover:bg-bg-hover transition-colors"
                  >
                    {/* User */}
                    <div className="flex items-center gap-3">
                      <div className="w-9 h-9 rounded-full bg-brand-500/10 flex items-center justify-center flex-shrink-0">
                        <span className="text-sm font-semibold text-brand-500">
                          {(user.name || user.email).charAt(0).toUpperCase()}
                        </span>
                      </div>
                      <div className="min-w-0">
                        <p className="font-medium text-text-primary truncate">{user.name || "—"}</p>
                        <p className="text-xs text-text-muted truncate md:hidden">{user.email}</p>
                      </div>
                    </div>

                    {/* Email */}
                    <div className="hidden md:flex items-center">
                      <p className="text-sm text-text-secondary truncate">{user.email}</p>
                    </div>

                    {/* Role */}
                    <div className="flex items-center">
                      {roleConfig ? (
                        <Badge className={cn("text-xs border font-medium", roleConfig.color)}>
                          <RoleIcon className="w-3 h-3 mr-1" />
                          {roleConfig.label}
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="text-xs border-border-default text-text-muted">
                          User
                        </Badge>
                      )}
                    </div>

                    {/* Plan */}
                    <div className="flex items-center">
                      <Badge className={cn("text-xs border font-medium capitalize", planConfig.color)}>
                        {planConfig.label}
                      </Badge>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center justify-end">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm" className="text-text-muted hover:text-text-primary h-8 w-8 p-0">
                            <MoreVertical className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent className="bg-bg-tertiary border-border-default" align="end">
                          <DropdownMenuItem
                            className="text-text-primary hover:bg-bg-hover cursor-pointer"
                            onClick={() => handleViewDetails(user)}
                          >
                            <Eye className="w-4 h-4 mr-2" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-text-primary hover:bg-bg-hover cursor-pointer"
                            onClick={() => handleEditUser(user)}
                          >
                            <User className="w-4 h-4 mr-2" />
                            Edit User
                          </DropdownMenuItem>
                          <DropdownMenuItem className="text-text-primary hover:bg-bg-hover cursor-pointer">
                            <Mail className="w-4 h-4 mr-2" />
                            Reset Password
                          </DropdownMenuItem>
                          <DropdownMenuSeparator className="bg-border-subtle" />
                          <DropdownMenuItem className="text-red-500 hover:bg-red-500/10 cursor-pointer">
                            <UserX className="w-4 h-4 mr-2" />
                            Disable User
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </CardContent>
      </Card>

      {/* Edit User Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="bg-bg-tertiary border-border-default">
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
                className="bg-bg-secondary border-border-default text-text-primary mt-1"
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
                className="bg-bg-secondary border-border-default text-text-primary mt-1"
              />
            </div>
            <div>
              <Label htmlFor="edit-role" className="text-text-primary">Role</Label>
              <Select value={editRole} onValueChange={setEditRole}>
                <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary mt-1">
                  <SelectValue placeholder="Select a role" />
                </SelectTrigger>
                <SelectContent className="bg-bg-tertiary border-border-default">
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
                <SelectTrigger className="bg-bg-secondary border-border-default text-text-primary mt-1">
                  <SelectValue placeholder="Select a plan" />
                </SelectTrigger>
                <SelectContent className="bg-bg-tertiary border-border-default">
                  <SelectItem value="free">Free</SelectItem>
                  <SelectItem value="starter">Starter</SelectItem>
                  <SelectItem value="pro">Pro</SelectItem>
                  <SelectItem value="enterprise">Enterprise</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button variant="outline" onClick={() => setIsEditDialogOpen(false)} className="border-border-default">
                Cancel
              </Button>
              <Button
                onClick={handleSaveEdit}
                disabled={updateUserMutation.isPending}
                className="bg-brand-500 hover:bg-brand-600 text-white"
              >
                Save Changes
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* User Details Dialog */}
      <Dialog open={isDetailsOpen} onOpenChange={setIsDetailsOpen}>
        <DialogContent className="bg-bg-tertiary border-border-default max-w-2xl">
          <DialogHeader>
            <DialogTitle className="text-text-primary">
              User Details
            </DialogTitle>
          </DialogHeader>
          {selectedUser && (
            <div className="space-y-6">
              {/* User Avatar + Name */}
              <div className="flex items-center gap-4 p-4 rounded-xl bg-bg-secondary border border-border-subtle">
                <div className="w-14 h-14 rounded-full bg-brand-500/10 flex items-center justify-center">
                  <span className="text-xl font-bold text-brand-500">
                    {(selectedUser.name || selectedUser.email).charAt(0).toUpperCase()}
                  </span>
                </div>
                <div>
                  <p className="font-semibold text-text-primary text-lg">{selectedUser.name || "—"}</p>
                  <p className="text-text-secondary">{selectedUser.email}</p>
                  <div className="flex items-center gap-2 mt-1">
                    {selectedUser.role && ROLE_CONFIG[selectedUser.role] && (
                      <Badge className={cn("text-xs border", ROLE_CONFIG[selectedUser.role].color)}>
                        {ROLE_CONFIG[selectedUser.role].label}
                      </Badge>
                    )}
                    <Badge className={cn("text-xs border", (PLAN_CONFIG[selectedUser.plan] || PLAN_CONFIG.free).color)}>
                      {(PLAN_CONFIG[selectedUser.plan] || PLAN_CONFIG.free).label}
                    </Badge>
                  </div>
                </div>
              </div>

              {/* Details Grid */}
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <Label className="text-text-muted text-xs uppercase tracking-wider">User ID</Label>
                  <p className="text-text-primary font-mono text-sm bg-bg-secondary px-2 py-1 rounded border border-border-subtle truncate">
                    {selectedUser.id}
                  </p>
                </div>
                <div className="space-y-1">
                  <Label className="text-text-muted text-xs uppercase tracking-wider">Tenant ID</Label>
                  <p className="text-text-primary font-mono text-sm bg-bg-secondary px-2 py-1 rounded border border-border-subtle truncate">
                    {selectedUser.tenantId || "—"}
                  </p>
                </div>
                <div className="space-y-1">
                  <Label className="text-text-muted text-xs uppercase tracking-wider">Created</Label>
                  <p className="text-text-primary text-sm">
                    {selectedUser.createdAt ? new Date(selectedUser.createdAt).toLocaleDateString() : "—"}
                  </p>
                </div>
                <div className="space-y-1">
                  <Label className="text-text-muted text-xs uppercase tracking-wider">Last Updated</Label>
                  <p className="text-text-primary text-sm">
                    {selectedUser.updatedAt ? new Date(selectedUser.updatedAt).toLocaleDateString() : "—"}
                  </p>
                </div>
              </div>

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsDetailsOpen(false)} className="border-border-default">
                  Close
                </Button>
                <Button
                  onClick={() => {
                    setIsDetailsOpen(false);
                    handleEditUser(selectedUser);
                  }}
                  className="bg-brand-500 hover:bg-brand-600 text-white"
                >
                  Edit User
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
