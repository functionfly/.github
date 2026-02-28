import { useState } from "react";
import { Camera, Plus, RotateCcw, Trash2, Clock, Database, HardDrive } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreateSnapshot, useDeleteSnapshot, useCreateReplay } from "@/hooks/useStateFabric";
import type { Snapshot } from "@/types";

interface SnapshotManagerProps {
  fabricId: string;
  snapshots: Snapshot[];
}

export function SnapshotManager({ fabricId, snapshots }: SnapshotManagerProps) {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newSnapshotName, setNewSnapshotName] = useState("");
  const [newSnapshotDescription, setNewSnapshotDescription] = useState("");
  const [selectedSnapshot, setSelectedSnapshot] = useState<Snapshot | null>(null);
  const [isReplayOpen, setIsReplayOpen] = useState(false);

  const createSnapshot = useCreateSnapshot(fabricId);
  const deleteSnapshot = useDeleteSnapshot(fabricId);
  const createReplay = useCreateReplay(fabricId);

  const handleCreate = async () => {
    if (!newSnapshotName.trim()) return;
    await createSnapshot.mutateAsync({
      name: newSnapshotName,
      description: newSnapshotDescription,
    });
    setIsCreateOpen(false);
    setNewSnapshotName("");
    setNewSnapshotDescription("");
  };

  const handleDelete = async (snapshotId: string) => {
    if (confirm("Are you sure you want to delete this snapshot?")) {
      await deleteSnapshot.mutateAsync(snapshotId);
    }
  };

  const handleReplay = async () => {
    if (!selectedSnapshot) return;
    await createReplay.mutateAsync({
      snapshotId: selectedSnapshot.id,
    });
    setIsReplayOpen(false);
    setSelectedSnapshot(null);
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getExpiryStatus = (snapshot: Snapshot) => {
    if (!snapshot.expiresAt) return null;
    const expiry = new Date(snapshot.expiresAt);
    const now = new Date();
    const daysUntilExpiry = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

    if (daysUntilExpiry < 0) return { label: "Expired", color: "bg-red-500/10 text-red-400" };
    if (daysUntilExpiry <= 7) return { label: `Expires in ${daysUntilExpiry} days`, color: "bg-yellow-500/10 text-yellow-400" };
    return { label: `Expires in ${daysUntilExpiry} days`, color: "bg-green-500/10 text-green-400" };
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-text-primary">Snapshots</h3>
          <p className="text-sm text-text-muted">
            Point-in-time backups of your state fabric data
          </p>
        </div>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="w-4 h-4 mr-2" />
              Create Snapshot
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Snapshot</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={newSnapshotName}
                  onChange={(e) => setNewSnapshotName(e.target.value)}
                  placeholder="Enter snapshot name"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description (optional)</Label>
                <Input
                  id="description"
                  value={newSnapshotDescription}
                  onChange={(e) => setNewSnapshotDescription(e.target.value)}
                  placeholder="Enter description"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={!newSnapshotName.trim()}>
                Create Snapshot
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Snapshot List */}
      {snapshots.length === 0 ? (
        <Card className="p-8 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
            <Camera className="w-8 h-8 text-text-muted" />
          </div>
          <p className="text-text-muted mb-4">No snapshots yet</p>
          <Button onClick={() => setIsCreateOpen(true)}>
            <Plus className="w-4 h-4 mr-2" />
            Create Your First Snapshot
          </Button>
        </Card>
      ) : (
        <div className="grid gap-4">
          {snapshots.map((snapshot) => {
            const expiryStatus = getExpiryStatus(snapshot);
            return (
              <Card key={snapshot.id}>
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-lg bg-bg-secondary flex items-center justify-center">
                        <Camera className="w-5 h-5 text-blue-400" />
                      </div>
                      <div>
                        <CardTitle className="text-lg">{snapshot.name}</CardTitle>
                        <p className="text-sm text-text-muted">{snapshot.description}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Dialog open={isReplayOpen && selectedSnapshot?.id === snapshot.id} onOpenChange={(open) => {
                        setIsReplayOpen(open);
                        if (!open) setSelectedSnapshot(null);
                      }}>
                        <DialogTrigger asChild>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setSelectedSnapshot(snapshot)}
                          >
                            <RotateCcw className="w-4 h-4 mr-2" />
                            Replay
                          </Button>
                        </DialogTrigger>
                        <DialogContent>
                          <DialogHeader>
                            <DialogTitle>Replay from Snapshot</DialogTitle>
                          </DialogHeader>
                          <div className="py-4">
                            <p className="text-text-muted mb-4">
                              This will create a new replay session starting from snapshot "{snapshot.name}".
                              All events after this snapshot will be replayed.
                            </p>
                            <div className="bg-bg-secondary p-4 rounded-lg">
                              <div className="grid grid-cols-2 gap-4 text-sm">
                                <div>
                                  <span className="text-text-muted">Snapshot:</span>
                                  <p className="font-medium">{snapshot.name}</p>
                                </div>
                                <div>
                                  <span className="text-text-muted">Events:</span>
                                  <p className="font-medium">{snapshot.eventCount.toLocaleString()}</p>
                                </div>
                                <div>
                                  <span className="text-text-muted">Created:</span>
                                  <p className="font-medium">{formatDate(snapshot.createdAt)}</p>
                                </div>
                                <div>
                                  <span className="text-text-muted">Size:</span>
                                  <p className="font-medium">{formatBytes(snapshot.sizeBytes)}</p>
                                </div>
                              </div>
                            </div>
                          </div>
                          <DialogFooter>
                            <Button variant="outline" onClick={() => setIsReplayOpen(false)}>
                              Cancel
                            </Button>
                            <Button onClick={handleReplay}>
                              <RotateCcw className="w-4 h-4 mr-2" />
                              Start Replay
                            </Button>
                          </DialogFooter>
                        </DialogContent>
                      </Dialog>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(snapshot.id)}
                      >
                        <Trash2 className="w-4 h-4 text-red-400" />
                      </Button>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div className="flex items-center gap-2">
                      <Database className="w-4 h-4 text-text-muted" />
                      <div>
                        <p className="text-xs text-text-muted">Events</p>
                        <p className="font-medium">{snapshot.eventCount.toLocaleString()}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <HardDrive className="w-4 h-4 text-text-muted" />
                      <div>
                        <p className="text-xs text-text-muted">Size</p>
                        <p className="font-medium">{formatBytes(snapshot.sizeBytes)}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Clock className="w-4 h-4 text-text-muted" />
                      <div>
                        <p className="text-xs text-text-muted">Created</p>
                        <p className="font-medium">{formatDate(snapshot.createdAt)}</p>
                      </div>
                    </div>
                    {expiryStatus && (
                      <div className="flex items-center gap-2">
                        <Badge className={expiryStatus.color}>
                          {expiryStatus.label}
                        </Badge>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
