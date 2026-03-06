/**
 * SecretPermissionMatrix - Grid/matrix view of secret permissions
 *
 * Displays a matrix view of secret permissions with users/roles as rows
 * and permissions as columns. Supports permission types: read, write,
 * rotate, delete, admin. Includes toggle/checkbox for each permission cell,
 * bulk edit capabilities, inheritance indicators, user avatars and role badges,
 * search/filter users, save/cancel actions, and loading state.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <SecretPermissionMatrix
 *   users={users}
 *   permissions={currentPermissions}
 *   onSave={handleSave}
 * />
 *
 * // With roles and inheritance
 * <SecretPermissionMatrix
 *   users={users}
 *   roles={roles}
 *   permissions={permissions}
 *   showInheritance
 *   inheritedPermissions={inheritedPerms}
 *   onSave={handleSave}
 *   onCancel={() => setIsEditing(false)}
 * />
 *
 * // Loading state
 * <SecretPermissionMatrix isLoading />
 * ```
 */

import { useState, useMemo, useCallback } from "react";
import {
  User,
  Users,
  Shield,
  Eye,
  FileEdit,
  RefreshCw,
  Trash2,
  Crown,
  Save,
  X,
  Search,
  Check,
  Loader2,
  AlertTriangle,
  ChevronRight,
  ChevronDown,
  Link,
  RotateCcw,
} from "lucide-react";
import { cn } from "@/lib/utils";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

/** Permission types available in the matrix */
export type PermissionType = "read" | "write" | "rotate" | "delete" | "admin";

/** User information for permission matrix */
export interface PermissionUser {
  id: string;
  name: string;
  email?: string;
  avatarUrl?: string;
  role?: string;
  type: "user" | "service_account" | "api_key";
}

/** Role information for group permissions */
export interface PermissionRole {
  id: string;
  name: string;
  description?: string;
  userCount?: number;
  isSystem?: boolean;
}

/** Permission entry for a user/role */
export interface PermissionEntry {
  userId: string;
  permissions: PermissionType[];
  inheritedFrom?: string;
  inheritedPermissions?: PermissionType[];
  expiresAt?: string;
}

export interface SecretPermissionMatrixProps {
  /** Array of users to display */
  users?: PermissionUser[];
  /** Array of roles to display */
  roles?: PermissionRole[];
  /** Current permission entries */
  permissions?: PermissionEntry[];
  /** Whether to show inheritance indicators */
  showInheritance?: boolean;
  /** Whether the component is loading */
  isLoading?: boolean;
  /** Whether the matrix is in edit mode */
  isEditing?: boolean;
  /** Whether save operation is in progress */
  isSaving?: boolean;
  /** Error message if save failed */
  saveError?: Error | null;
  /** Callback when permissions are saved */
  onSave?: (permissions: PermissionEntry[]) => Promise<void> | void;
  /** Callback when editing is cancelled */
  onCancel?: () => void;
  /** Callback when edit mode is toggled */
  onEditToggle?: (isEditing: boolean) => void;
  /** Additional CSS classes */
  className?: string;
}

// Permission metadata
const PERMISSIONS: { type: PermissionType; label: string; icon: typeof Eye; description: string }[] = [
  { type: "read", label: "Read", icon: Eye, description: "Can view secret value" },
  { type: "write", label: "Write", icon: FileEdit, description: "Can update secret metadata and value" },
  { type: "rotate", label: "Rotate", icon: RefreshCw, description: "Can rotate/renew the secret" },
  { type: "delete", label: "Delete", icon: Trash2, description: "Can delete the secret" },
  { type: "admin", label: "Admin", icon: Crown, description: "Full control including permissions" },
];

// Get initials from name
function getInitials(name: string): string {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

// Get user type icon
function UserTypeIcon({ type }: { type: PermissionUser["type"] }) {
  switch (type) {
    case "service_account":
      return <RefreshCw className="h-3.5 w-3.5" />;
    case "api_key":
      return <Shield className="h-3.5 w-3.5" />;
    default:
      return <User className="h-3.5 w-3.5" />;
  }
}

// Get user type label
function getUserTypeLabel(type: PermissionUser["type"]): string {
  switch (type) {
    case "service_account":
      return "Service Account";
    case "api_key":
      return "API Key";
    default:
      return "User";
  }
}

/**
 * SecretPermissionMatrix component
 */
export function SecretPermissionMatrix({
  users = [],
  roles = [],
  permissions = [],
  showInheritance = false,
  isLoading = false,
  isEditing: controlledEditing,
  isSaving = false,
  saveError = null,
  onSave,
  onCancel,
  onEditToggle,
  className,
}: SecretPermissionMatrixProps) {
  const [localEditing, setLocalEditing] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterType, setFilterType] = useState<"all" | "user" | "role">("all");
  const [expandedRoles, setExpandedRoles] = useState<Set<string>>(new Set());
  const [pendingChanges, setPendingChanges] = useState<Map<string, PermissionType[]>>(new Map());

  // Determine edit mode
  const isEditing = controlledEditing ?? localEditing;
  const setIsEditing = (value: boolean) => {
    if (controlledEditing === undefined) {
      setLocalEditing(value);
    }
    onEditToggle?.(value);
  };

  // Filter users/roles
  const filteredUsers = useMemo(() => {
    if (filterType === "role") return [];
    if (!searchQuery) return users;
    const query = searchQuery.toLowerCase();
    return users.filter(
      (u) =>
        u.name.toLowerCase().includes(query) ||
        u.email?.toLowerCase().includes(query) ||
        u.role?.toLowerCase().includes(query)
    );
  }, [users, searchQuery, filterType]);

  const filteredRoles = useMemo(() => {
    if (filterType === "user") return [];
    if (!searchQuery) return roles;
    const query = searchQuery.toLowerCase();
    return roles.filter(
      (r) =>
        r.name.toLowerCase().includes(query) ||
        r.description?.toLowerCase().includes(query)
    );
  }, [roles, searchQuery, filterType]);

  // Get permissions for a user (considering pending changes)
  const getUserPermissions = useCallback(
    (userId: string): PermissionType[] => {
      if (pendingChanges.has(userId)) {
        return pendingChanges.get(userId)!;
      }
      const entry = permissions.find((p) => p.userId === userId);
      return entry?.permissions ?? [];
    },
    [permissions, pendingChanges]
  );

  // Check if user has permission
  const hasPermission = useCallback(
    (userId: string, permission: PermissionType): boolean => {
      const userPerms = getUserPermissions(userId);
      // Admin implies all permissions
      if (userPerms.includes("admin")) return true;
      return userPerms.includes(permission);
    },
    [getUserPermissions]
  );

  // Toggle permission for a user
  const togglePermission = useCallback(
    (userId: string, permission: PermissionType) => {
      if (!isEditing) return;

      setPendingChanges((prev) => {
        const currentPerms = getUserPermissions(userId);
        let newPerms: PermissionType[];

        if (currentPerms.includes(permission)) {
          newPerms = currentPerms.filter((p) => p !== permission);
        } else {
          newPerms = [...currentPerms, permission];
        }

        const next = new Map(prev);
        next.set(userId, newPerms);
        return next;
      });
    },
    [isEditing, getUserPermissions]
  );

  // Check if has changes
  const hasChanges = pendingChanges.size > 0;

  // Handle save
  const handleSave = useCallback(async () => {
    if (!hasChanges || !onSave) return;

    // Build updated permissions array
    const updatedPermissions: PermissionEntry[] = users.map((user) => {
      const existing = permissions.find((p) => p.userId === user.id);
      const newPerms = pendingChanges.get(user.id) ?? existing?.permissions ?? [];
      return {
        userId: user.id,
        permissions: newPerms,
        inheritedFrom: existing?.inheritedFrom,
        inheritedPermissions: existing?.inheritedPermissions,
        expiresAt: existing?.expiresAt,
      };
    });

    await onSave(updatedPermissions);
    setPendingChanges(new Map());
    setIsEditing(false);
  }, [hasChanges, onSave, users, permissions, pendingChanges, setIsEditing]);

  // Handle cancel
  const handleCancel = useCallback(() => {
    setPendingChanges(new Map());
    setIsEditing(false);
    onCancel?.();
  }, [onCancel, setIsEditing]);

  // Toggle role expansion
  const toggleRole = useCallback((roleId: string) => {
    setExpandedRoles((prev) => {
      const next = new Set(prev);
      if (next.has(roleId)) {
        next.delete(roleId);
      } else {
        next.add(roleId);
      }
      return next;
    });
  }, []);

  // Bulk set permissions
  const bulkSetPermissions = useCallback(
    (permission: PermissionType, grant: boolean) => {
      if (!isEditing) return;

      setPendingChanges((prev) => {
        const next = new Map(prev);
        filteredUsers.forEach((user) => {
          const currentPerms = getUserPermissions(user.id);
          let newPerms: PermissionType[];

          if (grant) {
            newPerms = currentPerms.includes(permission)
              ? currentPerms
              : [...currentPerms, permission];
          } else {
            newPerms = currentPerms.filter((p) => p !== permission);
          }

          next.set(user.id, newPerms);
        });
        return next;
      });
    },
    [isEditing, filteredUsers, getUserPermissions]
  );

  // Loading state
  if (isLoading) {
    return (
      <div className={cn("space-y-4", className)}>
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-10 w-24" />
        </div>
        <div className="border border-border-subtle rounded-lg overflow-hidden">
          <div className="space-y-2 p-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4">
                <Skeleton className="h-10 w-10 rounded-full" />
                <Skeleton className="h-4 w-32" />
                <div className="flex-1 flex justify-end gap-2">
                  {Array.from({ length: 5 }).map((_, j) => (
                    <Skeleton key={j} className="h-8 w-8" />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <TooltipProvider>
      <div className={cn("space-y-4", className)}>
        {/* Header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3 flex-1">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
              <Input
                placeholder="Search users or roles..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
              {searchQuery && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6"
                  onClick={() => setSearchQuery("")}
                  aria-label="Clear search"
                >
                  <X className="h-3 w-3" />
                </Button>
              )}
            </div>
            <Select
              value={filterType}
              onValueChange={(value) => setFilterType(value as typeof filterType)}
            >
              <SelectTrigger className="w-[130px]">
                <Users className="h-4 w-4 mr-2" />
                <SelectValue placeholder="Filter" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="user">Users Only</SelectItem>
                <SelectItem value="role">Roles Only</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            {isEditing ? (
              <>
                <Button
                  variant="outline"
                  onClick={handleCancel}
                  disabled={isSaving}
                >
                  <X className="h-4 w-4 mr-2" />
                  Cancel
                </Button>
                <Button
                  onClick={handleSave}
                  disabled={!hasChanges || isSaving}
                  className="gap-2"
                >
                  {isSaving ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Save className="h-4 w-4" />
                  )}
                  Save Changes
                </Button>
              </>
            ) : (
              <Button onClick={() => setIsEditing(true)} variant="outline">
                <FileEdit className="h-4 w-4 mr-2" />
                Edit Permissions
              </Button>
            )}
          </div>
        </div>

        {/* Error alert */}
        {saveError && (
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Failed to save permissions</AlertTitle>
            <AlertDescription>{saveError.message}</AlertDescription>
          </Alert>
        )}

        {/* Pending changes indicator */}
        {hasChanges && (
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              You have unsaved changes. Don't forget to save your modifications.
            </AlertDescription>
          </Alert>
        )}

        {/* Permission matrix */}
        <div className="border border-border-subtle rounded-lg overflow-hidden">
          {/* Header row */}
          <div className="bg-bg-secondary border-b border-border-subtle">
            <div className="grid grid-cols-[1fr,repeat(5,auto)] gap-2 items-center p-3 text-sm">
              <div className="font-medium text-text-primary">
                User / Role
                <span className="text-text-muted ml-2">
                  ({filteredUsers.length + filteredRoles.length})
                </span>
              </div>
              {PERMISSIONS.map((perm) => (
                <Tooltip key={perm.type}>
                  <TooltipTrigger asChild>
                    <div className="flex flex-col items-center gap-1 w-16">
                      <perm.icon className="h-4 w-4 text-text-secondary" />
                      <span className="text-xs text-text-muted">{perm.label}</span>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{perm.description}</p>
                  </TooltipContent>
                </Tooltip>
              ))}
            </div>

            {/* Bulk actions row (only when editing) */}
            {isEditing && (
              <div className="grid grid-cols-[1fr,repeat(5,auto)] gap-2 items-center px-3 pb-3 text-sm">
                <div className="text-xs text-text-muted">Bulk actions:</div>
                {PERMISSIONS.map((perm) => (
                  <div key={perm.type} className="flex items-center justify-center w-16 gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => bulkSetPermissions(perm.type, true)}
                      title={`Grant ${perm.label} to all visible`}
                    >
                      <Check className="h-3 w-3" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      onClick={() => bulkSetPermissions(perm.type, false)}
                      title={`Revoke ${perm.label} from all visible`}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Empty state */}
          {filteredUsers.length === 0 && filteredRoles.length === 0 && (
            <div className="p-12 text-center">
              <div className="mx-auto h-16 w-16 rounded-full bg-gradient-to-br from-brand-500/20 to-purple-500/20 flex items-center justify-center mb-4">
                <Users className="h-8 w-8 text-brand-500" />
              </div>
              <h3 className="text-lg font-semibold text-card-foreground mb-2">
                No users or roles found
              </h3>
              <p className="text-text-secondary max-w-sm mx-auto">
                {searchQuery
                  ? "No results match your search. Try different keywords."
                  : "Users and roles with secret permissions will appear here."}
              </p>
            </div>
          )}

          {/* Users list */}
          <div className="divide-y divide-border-subtle">
            {filteredUsers.map((user) => {
              const userPerms = getUserPermissions(user.id);
              const isAdmin = userPerms.includes("admin");
              const hasChangesForUser = pendingChanges.has(user.id);

              return (
                <div
                  key={user.id}
                  className={cn(
                    "grid grid-cols-[1fr,repeat(5,auto)] gap-2 items-center p-3",
                    "hover:bg-bg-secondary/50 transition-colors",
                    hasChangesForUser && "bg-warning-glow/10"
                  )}
                >
                  {/* User info */}
                  <div className="flex items-center gap-3 min-w-0">
                    <Avatar className="h-9 w-9 shrink-0">
                      <AvatarImage src={user.avatarUrl} alt={user.name} />
                      <AvatarFallback>{getInitials(user.name)}</AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm text-text-primary truncate">
                          {user.name}
                        </span>
                        {isAdmin && (
                          <Tooltip>
                            <TooltipTrigger>
                              <Crown className="h-3.5 w-3.5 text-warning" />
                            </TooltipTrigger>
                            <TooltipContent>
                              <p>Admin access</p>
                            </TooltipContent>
                          </Tooltip>
                        )}
                      </div>
                      <div className="flex items-center gap-2 text-xs text-text-muted">
                        <UserTypeIcon type={user.type} />
                        <span className="truncate">
                          {user.email || getUserTypeLabel(user.type)}
                        </span>
                        {user.role && (
                          <Badge variant="outline" className="text-[10px] px-1 py-0">
                            {user.role}
                          </Badge>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Permission checkboxes */}
                  {PERMISSIONS.map((perm) => {
                    const isGranted = hasPermission(user.id, perm.type);
                    const isInherited =
                      showInheritance &&
                      permissions
                        .find((p) => p.userId === user.id)
                        ?.inheritedPermissions?.includes(perm.type);

                    return (
                      <div key={perm.type} className="flex justify-center w-16">
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div>
                              <Checkbox
                                checked={isGranted}
                                disabled={!isEditing || isAdmin}
                                onCheckedChange={() =>
                                  togglePermission(user.id, perm.type)
                                }
                                className={cn(
                                  isGranted && "data-[state=checked]:bg-success data-[state=checked]:border-success",
                                  isInherited && "data-[state=checked]:bg-brand-500 data-[state=checked]:border-brand-500"
                                )}
                              />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>
                              {isInherited
                                ? `${perm.label} (inherited)`
                                : isGranted
                                ? `Has ${perm.label.toLowerCase()} permission`
                                : `No ${perm.label.toLowerCase()} permission`}
                            </p>
                          </TooltipContent>
                        </Tooltip>
                        {isInherited && (
                          <Link className="h-3 w-3 text-brand-500 ml-1" />
                        )}
                      </div>
                    );
                  })}
                </div>
              );
            })}

            {/* Roles list */}
            {filteredRoles.map((role) => {
              const isExpanded = expandedRoles.has(role.id);
              const rolePerms = getUserPermissions(role.id);
              const isAdmin = rolePerms.includes("admin");

              return (
                <div key={role.id} className="divide-y divide-border-subtle">
                  <div
                    className={cn(
                      "grid grid-cols-[1fr,repeat(5,auto)] gap-2 items-center p-3",
                      "hover:bg-bg-secondary/50 transition-colors cursor-pointer"
                    )}
                    onClick={() => toggleRole(role.id)}
                  >
                    {/* Role info */}
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="h-9 w-9 rounded-full bg-brand-500/10 flex items-center justify-center shrink-0">
                        <Users className="h-4 w-4 text-brand-500" />
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-sm text-text-primary">
                            {role.name}
                          </span>
                          {role.isSystem && (
                            <Badge variant="outline" className="text-[10px]">
                              System
                            </Badge>
                          )}
                          {isAdmin && (
                            <Crown className="h-3.5 w-3.5 text-warning" />
                          )}
                        </div>
                        <div className="flex items-center gap-2 text-xs text-text-muted">
                          <span>{role.userCount ?? 0} members</span>
                          {role.description && (
                            <span className="truncate">• {role.description}</span>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Permission indicators (read-only for roles) */}
                    {PERMISSIONS.map((perm) => {
                      const isGranted = hasPermission(role.id, perm.type);

                      return (
                        <div key={perm.type} className="flex justify-center w-16">
                          {isGranted ? (
                            <div className="h-5 w-5 rounded bg-success/20 flex items-center justify-center">
                              <Check className="h-3 w-3 text-success" />
                            </div>
                          ) : (
                            <div className="h-5 w-5 rounded border border-border-subtle" />
                          )}
                        </div>
                      );
                    })}
                  </div>

                  {/* Expanded role members (placeholder) */}
                  {isExpanded && (
                    <div className="bg-bg-secondary/30 px-12 py-3 text-sm text-text-muted">
                      <div className="flex items-center gap-2">
                        <ChevronRight className="h-4 w-4" />
                        <span>Role members would be listed here</span>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Legend */}
        <div className="flex flex-wrap items-center gap-4 text-xs text-text-muted">
          <div className="flex items-center gap-1.5">
            <div className="h-4 w-4 rounded bg-success/20 flex items-center justify-center">
              <Check className="h-2.5 w-2.5 text-success" />
            </div>
            <span>Granted</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="h-4 w-4 rounded bg-brand-500/20 flex items-center justify-center">
              <Link className="h-2.5 w-2.5 text-brand-500" />
            </div>
            <span>Inherited</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="h-4 w-4 rounded border border-border-subtle" />
            <span>Not granted</span>
          </div>
          <div className="flex items-center gap-1.5">
            <Crown className="h-3.5 w-3.5 text-warning" />
            <span>Admin (all permissions)</span>
          </div>
          {showInheritance && (
            <div className="flex items-center gap-1.5">
              <RotateCcw className="h-3.5 w-3.5 text-text-muted" />
              <span>Inheritance enabled</span>
            </div>
          )}
        </div>
      </div>
    </TooltipProvider>
  );
}
