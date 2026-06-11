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
import "./StateDetailPage/styles.css";

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
      <div className="state-detail-container state-detail-content">
        <div className="state-detail-header">
          <div className="state-detail-header-left">
            <Button variant="ghost" size="sm" onClick={() => navigate("/state")} className="state-detail-back-btn">
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <h1 className="state-detail-title">Create New State</h1>
          </div>
        </div>

        <Card className="state-detail-card">
          <CardHeader className="state-detail-card-header">
            <CardTitle className="state-detail-card-title">State Details</CardTitle>
          </CardHeader>
          <CardContent className="state-detail-card-content space-y-4">
            <div className="state-detail-form-group">
              <label className="state-detail-label">Path</label>
              <Input
                value={editPath}
                onChange={(e) => setEditPath(e.target.value)}
                placeholder="e.g., users/settings"
                className="state-detail-input"
              />
            </div>
            <div className="state-detail-form-group">
              <label className="state-detail-label">Key</label>
              <Input
                value={editKey}
                onChange={(e) => setEditKey(e.target.value)}
                placeholder="e.g., theme"
                className="state-detail-input"
              />
            </div>
            <div className="state-detail-form-group">
              <label className="state-detail-label">Value (JSON)</label>
              <Textarea
                value={editValue}
                onChange={(e) => handleValueChange(e.target.value)}
                placeholder='{"color": "dark"}'
                rows={6}
                className={`state-detail-textarea ${jsonError ? "state-detail-textarea-error" : ""}`}
              />
              {jsonError && (
                <p className="state-detail-error-text">{jsonError}</p>
              )}
            </div>
            <div className="state-detail-form-group">
              <label className="state-detail-label">TTL (optional, seconds)</label>
              <Input
                value={editTtl}
                onChange={(e) => setEditTtl(e.target.value)}
                placeholder="e.g., 3600"
                className="state-detail-input"
              />
            </div>
            <Button onClick={handleCreateState} className="btn-state-detail-primary">
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
      <div className="state-detail-container state-detail-content">
        <div className="state-detail-loading-container">
          <Skeleton className="state-detail-skeleton state-detail-skeleton-header" />
          <Skeleton className="state-detail-skeleton state-detail-skeleton-content" />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="state-detail-container">
        <div className="state-detail-header">
          <div className="state-detail-header-left">
            <Button variant="ghost" size="sm" onClick={() => navigate("/state")} className="state-detail-back-btn">
              <ArrowLeft className="h-4 w-4" />
            </Button>
            <h1 className="state-detail-title">Error</h1>
          </div>
        </div>
        <Card className="state-detail-card">
          <CardContent className="py-8 text-center state-detail-error">
            <p>Failed to load state: {(error as Error).message}</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="state-detail-container state-detail-content">
      {/* Header */}
      <div className="state-detail-header">
        <div className="state-detail-header-left">
          <Button variant="ghost" size="sm" onClick={() => navigate("/state")} className="state-detail-back-btn">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="flex items-center gap-3">
            <div className="state-detail-icon-container">
              <Database className="state-detail-icon" />
            </div>
            <div>
              <h1 className="state-detail-title font-mono">{stateData?.path}</h1>
              <p className="state-detail-subtitle">
                Key: {stateData?.key} • Version {stateData?.version}
              </p>
            </div>
          </div>
        </div>
        <div className="state-detail-header-right">
          <Button variant="outline" size="sm" onClick={() => refetch()} className="btn-state-detail-outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button variant="outline" size="sm" onClick={handleDeleteValue} className="btn-state-detail-danger">
            <Trash2 className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="value" className="state-detail-tabs space-y-4">
        <TabsList className="state-detail-tabs-list">
          <TabsTrigger value="value" className="state-detail-tabs-trigger">
            <Key className="state-detail-tabs-trigger-icon" />
            Value
          </TabsTrigger>
          <TabsTrigger value="history" className="state-detail-tabs-trigger">
            <History className="state-detail-tabs-trigger-icon" />
            History
          </TabsTrigger>
          <TabsTrigger value="snapshots" className="state-detail-tabs-trigger">
            <Camera className="state-detail-tabs-trigger-icon" />
            Snapshots
          </TabsTrigger>
          <TabsTrigger value="permissions" className="state-detail-tabs-trigger">
            <Edit className="state-detail-tabs-trigger-icon" />
            Permissions
          </TabsTrigger>
        </TabsList>

        {/* Value Tab */}
        <TabsContent value="value" className="state-detail-tabs-content">
          <Card className="state-detail-card">
            <CardHeader className="state-detail-card-header">
              <CardTitle className="state-detail-card-title">
                <span>Current Value</span>
                <Badge variant="outline">v{stateData?.version}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="state-detail-card-content space-y-4">
              <div className="state-detail-form-group">
                <label className="state-detail-label">Value (JSON)</label>
                <Textarea
                  value={editValue}
                  onChange={(e) => handleValueChange(e.target.value)}
                  rows={12}
                  className={`state-detail-textarea font-mono text-sm ${jsonError ? "state-detail-textarea-error" : ""}`}
                />
                {jsonError && (
                  <p className="state-detail-error-text">{jsonError}</p>
                )}
              </div>
              <div className="state-detail-form-row">
                <div className="state-detail-form-group">
                  <label className="state-detail-label">TTL (optional, seconds)</label>
                  <Input
                    value={editTtl}
                    onChange={(e) => setEditTtl(e.target.value)}
                    placeholder="Leave empty for no expiration"
                    className="state-detail-input"
                  />
                </div>
                <div className="state-detail-form-actions">
                  <Button
                    onClick={handleSaveValue}
                    disabled={!!jsonError || setValue.isPending}
                    className="btn-state-detail-primary"
                  >
                    <Save className="h-4 w-4" />
                    {setValue.isPending ? "Saving..." : "Save Changes"}
                  </Button>
                </div>
              </div>
              {stateData?.metadata && Object.keys(stateData.metadata).length > 0 && (
                <div className="state-detail-metadata">
                  <h4 className="state-detail-metadata-title">Metadata</h4>
                  <div className="state-detail-metadata-tags">
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
        <TabsContent value="history" className="state-detail-tabs-content">
          <Card className="state-detail-card">
            <CardHeader className="state-detail-card-header">
              <CardTitle className="state-detail-card-title">Change History</CardTitle>
            </CardHeader>
            <CardContent className="state-detail-card-content">
              {historyData?.history && historyData.history.length > 0 ? (
                <Table className="state-detail-table">
                  <TableHeader>
                    <TableRow className="state-detail-table-header-row">
                      <TableHead className="state-detail-table-header-cell">Operation</TableHead>
                      <TableHead className="state-detail-table-header-cell">Previous Value</TableHead>
                      <TableHead className="state-detail-table-header-cell">New Value</TableHead>
                      <TableHead className="state-detail-table-header-cell">Actor</TableHead>
                      <TableHead className="state-detail-table-header-cell">Timestamp</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {historyData.history.map((entry) => (
                      <TableRow key={entry.id} className="state-detail-table-body-row">
                        <TableCell className="state-detail-table-cell">
                          <Badge
                            className={
                              entry.operation === "create"
                                ? "state-detail-op-badge state-detail-op-badge-create"
                                : entry.operation === "delete"
                                ? "state-detail-op-badge state-detail-op-badge-delete"
                                : "state-detail-op-badge state-detail-op-badge-update"
                            }
                          >
                            {entry.operation}
                          </Badge>
                        </TableCell>
                        <TableCell className="state-detail-table-cell state-detail-table-cell-mono">
                          {entry.previousValue !== undefined
                            ? formatJson(entry.previousValue)
                            : "-"}
                        </TableCell>
                        <TableCell className="state-detail-table-cell state-detail-table-cell-mono">
                          {entry.newValue !== undefined ? formatJson(entry.newValue) : "-"}
                        </TableCell>
                        <TableCell className="state-detail-table-cell">{entry.actor || "-"}</TableCell>
                        <TableCell className="state-detail-table-cell">
                          <div className="state-detail-table-cell-timestamp">
                            <Clock className="h-3 w-3" />
                            {new Date(entry.timestamp).toLocaleString()}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="state-detail-empty-state">
                  <History className="state-detail-empty-icon" />
                  <p className="state-detail-empty-title">No history available</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Snapshots Tab */}
        <TabsContent value="snapshots" className="state-detail-tabs-content">
          <Card className="state-detail-card">
            <CardHeader className="state-detail-card-header flex flex-row items-center justify-between">
              <CardTitle className="state-detail-card-title">Snapshots</CardTitle>
              <Button size="sm" onClick={handleCreateSnapshot} className="btn-state-detail-primary gap-2">
                <Plus className="h-4 w-4" />
                Create Snapshot
              </Button>
            </CardHeader>
            <CardContent className="state-detail-card-content">
              {snapshotsData && snapshotsData.length > 0 ? (
                <Table className="state-detail-table">
                  <TableHeader>
                    <TableRow className="state-detail-table-header-row">
                      <TableHead className="state-detail-table-header-cell">Name</TableHead>
                      <TableHead className="state-detail-table-header-cell">Description</TableHead>
                      <TableHead className="state-detail-table-header-cell">Version</TableHead>
                      <TableHead className="state-detail-table-header-cell">Created</TableHead>
                      <TableHead className="state-detail-table-header-cell">Expires</TableHead>
                      <TableHead className="state-detail-table-header-cell w-[100px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {snapshotsData.map((snapshot) => (
                      <TableRow key={snapshot.id} className="state-detail-table-body-row">
                        <TableCell className="state-detail-table-cell font-medium">{snapshot.name}</TableCell>
                        <TableCell className="state-detail-table-cell">{snapshot.description || "-"}</TableCell>
                        <TableCell className="state-detail-table-cell">
                          <Badge variant="outline">v{snapshot.version}</Badge>
                        </TableCell>
                        <TableCell className="state-detail-table-cell">
                          {new Date(snapshot.createdAt).toLocaleDateString()}
                        </TableCell>
                        <TableCell className="state-detail-table-cell">
                          {snapshot.expiresAt
                            ? new Date(snapshot.expiresAt).toLocaleDateString()
                            : "-"}
                        </TableCell>
                        <TableCell className="state-detail-table-cell">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleRestoreSnapshot(snapshot.id)}
                            className="btn-state-detail-outline"
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
                <div className="state-detail-empty-state">
                  <Camera className="state-detail-empty-icon" />
                  <p className="state-detail-empty-title">No snapshots available</p>
                  <Button
                    size="sm"
                    variant="outline"
                    className="mt-2 btn-state-detail-outline"
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
        <TabsContent value="permissions" className="state-detail-tabs-content">
          <Card className="state-detail-card">
            <CardHeader className="state-detail-card-header">
              <CardTitle className="state-detail-card-title">Access Permissions</CardTitle>
            </CardHeader>
            <CardContent className="state-detail-card-content space-y-4">
              {/* Grant Permission Form */}
              <div className="state-detail-form-row">
                <div className="state-detail-form-group">
                  <label className="state-detail-label">Principal (email/ID)</label>
                  <Input
                    value={newPermissionPrincipal}
                    onChange={(e) => setNewPermissionPrincipal(e.target.value)}
                    placeholder="user@example.com"
                    className="state-detail-input"
                  />
                </div>
                <div className="state-detail-form-group">
                  <label className="state-detail-label">Type</label>
                  <select
                    value={newPermissionType}
                    onChange={(e) => setNewPermissionType(e.target.value as typeof newPermissionType)}
                    className="state-detail-input"
                  >
                    <option value="user">User</option>
                    <option value="team">Team</option>
                    <option value="service">Service</option>
                  </select>
                </div>
                <div className="state-detail-form-group">
                  <label className="state-detail-label">Permissions</label>
                  <select
                    multiple
                    value={newPermissionPerms}
                    onChange={(e) =>
                      setNewPermissionPerms(Array.from(e.target.selectedOptions, (o) => o.value))
                    }
                    className="state-detail-input"
                  >
                    <option value="read">Read</option>
                    <option value="write">Write</option>
                    <option value="admin">Admin</option>
                  </select>
                </div>
                <Button onClick={handleGrantPermission} disabled={!newPermissionPrincipal} className="btn-state-detail-primary">
                  Grant
                </Button>
              </div>

              {/* Permissions List */}
              {permissionsData && permissionsData.length > 0 ? (
                <Table className="state-detail-table">
                  <TableHeader>
                    <TableRow className="state-detail-table-header-row">
                      <TableHead className="state-detail-table-header-cell">Principal</TableHead>
                      <TableHead className="state-detail-table-header-cell">Type</TableHead>
                      <TableHead className="state-detail-table-header-cell">Permissions</TableHead>
                      <TableHead className="state-detail-table-header-cell">Granted By</TableHead>
                      <TableHead className="state-detail-table-header-cell">Granted At</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {permissionsData.map((perm) => (
                      <TableRow key={perm.id} className="state-detail-table-body-row">
                        <TableCell className="state-detail-table-cell font-mono">{perm.principal}</TableCell>
                        <TableCell className="state-detail-table-cell">
                          <Badge variant="outline">{perm.principalType}</Badge>
                        </TableCell>
                        <TableCell className="state-detail-table-cell">
                          <div className="flex gap-1">
                            {perm.permissions.map((p) => (
                              <Badge key={p} variant="secondary">
                                {p}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell className="state-detail-table-cell">{perm.grantedBy || "-"}</TableCell>
                        <TableCell className="state-detail-table-cell">
                          {new Date(perm.grantedAt).toLocaleDateString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="state-detail-empty-state">
                  <Edit className="state-detail-empty-icon" />
                  <p className="state-detail-empty-title">No permissions set</p>
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
