import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Shield,
  Plus,
  Trash2,
  Loader2,
  AlertCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  APIKeyPermission,
  Permission,
  ResourceType,
  PERMISSION_LABELS,
  RESOURCE_TYPE_LABELS,
} from "@/types/api-key";
import { apiKeysService } from "@/services/api-keys";

interface APIKeyPermissionsProps {
  keyId: string;
}

export function APIKeyPermissions({ keyId }: APIKeyPermissionsProps) {
  const queryClient = useQueryClient();
  const [showAddForm, setShowAddForm] = useState(false);
  const [newPermission, setNewPermission] = useState<{
    permission: Permission;
    resource_type: ResourceType;
    resource_id: string;
  }>({
    permission: "read",
    resource_type: "function",
    resource_id: "",
  });

  const { data: permissions, isLoading, error } = useQuery({
    queryKey: ["api-key-permissions", keyId],
    queryFn: () => apiKeysService.getPermissions(keyId),
  });

  const addPermissionMutation = useMutation({
    mutationFn: (data: typeof newPermission) =>
      apiKeysService.addPermission(keyId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-key-permissions", keyId] });
      setShowAddForm(false);
      setNewPermission({
        permission: "read",
        resource_type: "function",
        resource_id: "",
      });
    },
  });

  const removePermissionMutation = useMutation({
    mutationFn: (permissionId: string) =>
      apiKeysService.removePermission(keyId, permissionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-key-permissions", keyId] });
    },
  });

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const handleAddPermission = () => {
    if (!newPermission.resource_id.trim()) return;
    addPermissionMutation.mutate(newPermission);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8">
        <AlertCircle className="w-6 h-6 text-red-500 mx-auto mb-2" />
        <p className="text-sm text-muted-foreground">
          Failed to load permissions
        </p>
      </div>
    );
  }

  const perms = permissions as APIKeyPermission[] | undefined;

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <Shield className="w-5 h-5" />
          Permissions
        </CardTitle>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setShowAddForm(!showAddForm)}
        >
          <Plus className="w-4 h-4 mr-2" />
          Add Permission
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        {showAddForm && (
          <div className="p-4 border rounded-lg space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Permission</Label>
                <Select
                  value={newPermission.permission}
                  onValueChange={(v) =>
                    setNewPermission({ ...newPermission, permission: v as Permission })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="read">Read</SelectItem>
                    <SelectItem value="write">Write</SelectItem>
                    <SelectItem value="execute">Execute</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Resource Type</Label>
                <Select
                  value={newPermission.resource_type}
                  onValueChange={(v) =>
                    setNewPermission({
                      ...newPermission,
                      resource_type: v as ResourceType,
                    })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="function">Function</SelectItem>
                    <SelectItem value="app">App</SelectItem>
                    <SelectItem value="tenant">Tenant</SelectItem>
                    <SelectItem value="registry">Registry</SelectItem>
                    <SelectItem value="deployment">Deployment</SelectItem>
                    <SelectItem value="secret">Secret</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>Resource ID</Label>
              <Input
                placeholder="Enter resource ID (UUID)"
                value={newPermission.resource_id}
                onChange={(e) =>
                  setNewPermission({ ...newPermission, resource_id: e.target.value })
                }
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowAddForm(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleAddPermission}
                disabled={
                  !newPermission.resource_id.trim() ||
                  addPermissionMutation.isPending
                }
              >
                {addPermissionMutation.isPending ? "Adding..." : "Add"}
              </Button>
            </div>
          </div>
        )}

        {perms && perms.length > 0 ? (
          <div className="space-y-2">
            {perms.map((permission) => (
              <div
                key={permission.id}
                className="flex items-center justify-between p-3 border rounded-lg"
              >
                <div className="flex items-center gap-3">
                  <Badge variant="outline">
                    {RESOURCE_TYPE_LABELS[permission.resource_type]}
                  </Badge>
                  <span className="font-medium">
                    {PERMISSION_LABELS[permission.permission]}
                  </span>
                  <span className="text-xs text-muted-foreground font-mono">
                    {permission.resource_id.slice(0, 8)}...
                  </span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-muted-foreground">
                    {formatDate(permission.created_at)}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-red-600 hover:text-red-600"
                    onClick={() => removePermissionMutation.mutate(permission.id)}
                    disabled={removePermissionMutation.isPending}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-center py-8 text-muted-foreground">
            No permissions configured
          </p>
        )}
      </CardContent>
    </Card>
  );
}
