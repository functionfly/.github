import { useState as useReactState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  ArrowLeft,
  Database,
  Clock,
  Save,
  Trash2,
  History,
  Camera,
  RefreshCw,
  Key,
  Edit,
  Plus,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  useState as useSimpleStateQuery,
  useStateValue,
  useSetStateValue,
  useDeleteStateValue,
  useStateHistory,
  useStateSnapshots,
  useCreateSnapshot,
  useRestoreSnapshot,
  useStatePermissions,
  useGrantPermission,
} from "@/hooks/useState";
import type { StateValue } from "@/types";

export function StateDetailPage() {
  const { path } = useParams<{ path: string }>();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [isCreating, setIsCreating] = useReactState(path === "new");
  const [editPath, setEditPath] = useReactState("");
  const [editKey, setEditKey] = useReactState("");
  const [editValue, setEditValue] = useReactState("");
  const [editTtl, setEditTtl] = useReactState("");
  const [jsonError, setJsonError] = useReactState<string | null>(null);
  const [newPermissionPrincipal, setNewPermissionPrincipal] = useReactState("");
  const [newPermissionType, setNewPermissionType] = useReactState<"user" | "team" | "service">("user");
  const [newPermissionPerms, setNewPermissionPerms] = useReactState<string[]>(["read"]);

  const decodedPath = path ? decodeURIComponent(path) : "";

  const { data: stateData, isLoading, error, refetch } = useSimpleStateQuery(decodedPath);
  const { data: valueData } = useStateValue(decodedPath);
  const setValue = useSetStateValue(decodedPath);
  const deleteValue = useDeleteStateValue(decodedPath);
  const { data: historyData } = useStateHistory(decodedPath);
  const { data: snapshotsData } = useStateSnapshots(decodedPath);
  const createSnapshot = useCreateSnapshot(decodedPath);
  const restoreSnapshot = useRestoreSnapshot(decodedPath);
  const { data: permissionsData } = useStatePermissions(decodedPath);
  const grantPermission = useGrantPermission(decodedPath);

  // Initialize form for create mode
  useEffect(() => {
    if (isCreating) {
      setEditPath("");
      setEditKey("");
      setEditValue("");
      setEditTtl("");
    }
  }, [isCreating]);

  // Initialize form for edit mode
  useEffect(() => {
    if (stateData) {
      setEditPath(stateData.path);
      setEditKey(stateData.key);
      setEditValue(
        typeof stateData.value === "object"
          ? JSON.stringify(stateData.value, null, 2)
          : String(stateData.value)
      );
      setEditTtl(stateData.ttl?.toString() || "");
    }
  }, [stateData]);

  const validateJson = (value: string): boolean => {
    try {
      JSON.parse(value);
      setJsonError(null);
      return true;
    } catch (e) {
      setJsonError((e as Error).message);
      return false;
    }
  };

  const handleValueChange = (value: string) => {
    setEditValue(value);
    validateJson(value);
  };

  const handleSaveValue = async () => {
    if (!validateJson(editValue)) return;

    const parsedValue = JSON.parse(editValue);
    await setValue.mutateAsync({
      value: parsedValue,
      ttl: editTtl ? parseInt(editTtl) : undefined,
    });
    refetch();
  };

  const handleDeleteValue = async () => {
    if (window.confirm("Are you sure you want to delete this value?")) {
      await deleteValue.mutateAsync();
      navigate("/state");
    }
  };

  const handleCreateState = async () => {
    // This would use createState mutation
    // For now, we'll just navigate back
    navigate("/state");
  };

  const handleCreateSnapshot = async () => {
    const name = prompt("Enter snapshot name:");
    if (name) {
      await createSnapshot.mutateAsync({
        path: decodedPath,
        name,
      });
    }
  };

  const handleRestoreSnapshot = async (snapshotId: string) => {
    if (window.confirm("Are you sure you want to restore this snapshot?")) {
      await restoreSnapshot.mutateAsync({ snapshotId });
      refetch();
    }
  };

  const handleGrantPermission = async () => {
    if (!newPermissionPrincipal) return;
    await grantPermission.mutateAsync({
      path: decodedPath,
      principal: newPermissionPrincipal,
      principalType: newPermissionType,
      permissions: newPermissionPerms as ("read" | "write" | "admin")[],
    });
    setNewPermissionPrincipal("");
  };

  const formatJson = (value: unknown): string => {
    if (value === undefined) return "";
    return JSON.stringify(value, null, 2);
  };

  if (path === "new") {
    return (
      <div className="container mx-auto py-6 space-y-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => navigate("/state")}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h1 className="text-2xl font-bold">Create New State</h1>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>State Details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Path</label>
              <Input
                value={editPath}
                onChange={(e) => setEditPath(e.target.value)}
                placeholder="e.g., users/settings"
              />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Key</label>
              <Input
                value={editKey}
                onChange={(e) => setEditKey(e.target.value)}
                placeholder="e.g., theme"
              />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Value (JSON)</label>
              <Textarea
                value={editValue}
                onChange={(e) => handleValueChange(e.target.value)}
                placeholder='{"color": "dark"}'
                rows={6}
                className={jsonError ? "border-error" : ""}
              />
              {jsonError && (
                <p className="text-sm text-error">{jsonError}</p>
              )}
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">TTL (optional, seconds)</label>
              <Input
                value={editTtl}
                onChange={(e) => setEditTtl(e.target.value)}
                placeholder="e.g., 3600"
              />
            </div>
            <Button onClick={handleCreateState} className="gap-2">
              <Save className="h-4 w-4" />
              Create State
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="container mx-auto py-6 space-y-6">
        <Skeleton className="h-8 w-64" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center gap-4 mb-6">
          <Button variant="ghost" size="sm" onClick={() => navigate("/state")}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h1 className="text-2xl font-bold">Error</h1>
        </div>
        <Card>
          <CardContent className="py-8 text-center text-error">
            <p>Failed to load state: {(error as Error).message}</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => navigate("/state")}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="flex items-center gap-3">
            <div className="p-2 bg-brand-100 rounded-lg">
              <Database className="h-5 w-5 text-brand-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold font-mono">{stateData?.path}</h1>
              <p className="text-sm text-text-secondary">
                Key: {stateData?.key} • Version {stateData?.version}
              </p>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button variant="outline" size="sm" onClick={handleDeleteValue} className="text-error">
            <Trash2 className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="value" className="space-y-4">
        <TabsList>
          <TabsTrigger value="value" className="gap-2">
            <Key className="h-4 w-4" />
            Value
          </TabsTrigger>
          <TabsTrigger value="history" className="gap-2">
            <History className="h-4 w-4" />
            History
          </TabsTrigger>
          <TabsTrigger value="snapshots" className="gap-2">
            <Camera className="h-4 w-4" />
            Snapshots
          </TabsTrigger>
          <TabsTrigger value="permissions" className="gap-2">
            <Edit className="h-4 w-4" />
            Permissions
          </TabsTrigger>
        </TabsList>

        {/* Value Tab */}
        <TabsContent value="value">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>Current Value</span>
                <Badge variant="outline">v{stateData?.version}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Value (JSON)</label>
                <Textarea
                  value={editValue}
                  onChange={(e) => handleValueChange(e.target.value)}
                  rows={12}
                  className={`font-mono text-sm ${jsonError ? "border-error" : ""}`}
                />
                {jsonError && (
                  <p className="text-sm text-error">{jsonError}</p>
                )}
              </div>
              <div className="flex items-center gap-4">
                <div className="grid gap-2 flex-1">
                  <label className="text-sm font-medium">TTL (optional, seconds)</label>
                  <Input
                    value={editTtl}
                    onChange={(e) => setEditTtl(e.target.value)}
                    placeholder="Leave empty for no expiration"
                  />
                </div>
                <div className="flex items-end">
                  <Button
                    onClick={handleSaveValue}
                    disabled={!!jsonError || setValue.isPending}
                    className="gap-2"
                  >
                    <Save className="h-4 w-4" />
                    {setValue.isPending ? "Saving..." : "Save Changes"}
                  </Button>
                </div>
              </div>
              {stateData?.metadata && Object.keys(stateData.metadata).length > 0 && (
                <div className="pt-4 border-t">
                  <h4 className="text-sm font-medium mb-2">Metadata</h4>
                  <div className="flex flex-wrap gap-2">
                    {Object.entries(stateData.metadata).map(([key, value]) => (
                      <Badge key={key} variant="outline">
                        {key}: {value}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* History Tab */}
        <TabsContent value="history">
          <Card>
            <CardHeader>
              <CardTitle>Change History</CardTitle>
            </CardHeader>
            <CardContent>
              {historyData?.history && historyData.history.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Operation</TableHead>
                      <TableHead>Previous Value</TableHead>
                      <TableHead>New Value</TableHead>
                      <TableHead>Actor</TableHead>
                      <TableHead>Timestamp</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {historyData.history.map((entry) => (
                      <TableRow key={entry.id}>
                        <TableCell>
                          <Badge
                            variant={
                              entry.operation === "create"
                                ? "default"
                                : entry.operation === "delete"
                                ? "destructive"
                                : "outline"
                            }
                          >
                            {entry.operation}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs max-w-[200px] truncate">
                          {entry.previousValue !== undefined
                            ? formatJson(entry.previousValue)
                            : "-"}
                        </TableCell>
                        <TableCell className="font-mono text-xs max-w-[200px] truncate">
                          {entry.newValue !== undefined ? formatJson(entry.newValue) : "-"}
                        </TableCell>
                        <TableCell>{entry.actor || "-"}</TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1 text-sm">
                            <Clock className="h-3 w-3" />
                            {new Date(entry.timestamp).toLocaleString()}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="text-center py-8 text-text-secondary">
                  <History className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>No history available</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Snapshots Tab */}
        <TabsContent value="snapshots">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Snapshots</CardTitle>
              <Button size="sm" onClick={handleCreateSnapshot} className="gap-2">
                <Plus className="h-4 w-4" />
                Create Snapshot
              </Button>
            </CardHeader>
            <CardContent>
              {snapshotsData && snapshotsData.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Description</TableHead>
                      <TableHead>Version</TableHead>
                      <TableHead>Created</TableHead>
                      <TableHead>Expires</TableHead>
                      <TableHead className="w-[100px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {snapshotsData.map((snapshot) => (
                      <TableRow key={snapshot.id}>
                        <TableCell className="font-medium">{snapshot.name}</TableCell>
                        <TableCell>{snapshot.description || "-"}</TableCell>
                        <TableCell>
                          <Badge variant="outline">v{snapshot.version}</Badge>
                        </TableCell>
                        <TableCell>
                          {new Date(snapshot.createdAt).toLocaleDateString()}
                        </TableCell>
                        <TableCell>
                          {snapshot.expiresAt
                            ? new Date(snapshot.expiresAt).toLocaleDateString()
                            : "-"}
                        </TableCell>
                        <TableCell>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleRestoreSnapshot(snapshot.id)}
                          >
                            <RefreshCw className="h-3 w-3 mr-1" />
                            Restore
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="text-center py-8 text-text-secondary">
                  <Camera className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>No snapshots available</p>
                  <Button
                    size="sm"
                    variant="outline"
                    className="mt-2"
                    onClick={handleCreateSnapshot}
                  >
                    Create your first snapshot
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Permissions Tab */}
        <TabsContent value="permissions">
          <Card>
            <CardHeader>
              <CardTitle>Access Permissions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Grant Permission Form */}
              <div className="flex gap-2 items-end">
                <div className="flex-1 grid gap-2">
                  <label className="text-sm font-medium">Principal (email/ID)</label>
                  <Input
                    value={newPermissionPrincipal}
                    onChange={(e) => setNewPermissionPrincipal(e.target.value)}
                    placeholder="user@example.com"
                  />
                </div>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">Type</label>
                  <select
                    value={newPermissionType}
                    onChange={(e) => setNewPermissionType(e.target.value as typeof newPermissionType)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="user">User</option>
                    <option value="team">Team</option>
                    <option value="service">Service</option>
                  </select>
                </div>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">Permissions</label>
                  <select
                    multiple
                    value={newPermissionPerms}
                    onChange={(e) =>
                      setNewPermissionPerms(Array.from(e.target.selectedOptions, (o) => o.value))
                    }
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="read">Read</option>
                    <option value="write">Write</option>
                    <option value="admin">Admin</option>
                  </select>
                </div>
                <Button onClick={handleGrantPermission} disabled={!newPermissionPrincipal}>
                  Grant
                </Button>
              </div>

              {/* Permissions List */}
              {permissionsData && permissionsData.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Principal</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Permissions</TableHead>
                      <TableHead>Granted By</TableHead>
                      <TableHead>Granted At</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {permissionsData.map((perm) => (
                      <TableRow key={perm.id}>
                        <TableCell className="font-mono">{perm.principal}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{perm.principalType}</Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            {perm.permissions.map((p) => (
                              <Badge key={p} variant="secondary">
                                {p}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell>{perm.grantedBy || "-"}</TableCell>
                        <TableCell>
                          {new Date(perm.grantedAt).toLocaleDateString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="text-center py-8 text-text-secondary">
                  <Edit className="h-8 w-8 mx-auto mb-2 opacity-50" />
                  <p>No permissions set</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default StateDetailPage;
