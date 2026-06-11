import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import {
  ArrowLeft,
  Brain,
  Trash2,
  Save,
  Clock,
  Star,
  FileText,
  Calendar,
  Edit3,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useAgentMemory,
  useCreateAgentMemory,
  useUpdateAgentMemory,
  useDeleteAgentMemory,
} from "@/hooks/useAgentMemory";
import type { AgentMemoryType } from "@/types";

const memoryTypeColors: Record<AgentMemoryType, string> = {
  working: "bg-blue-500/10 border-blue-500/20 text-blue-700",
  longterm: "bg-green-500/10 border-green-500/20 text-green-700",
  context: "bg-purple-500/10 border-purple-500/20 text-purple-700",
  episodic: "bg-orange-500/10 border-orange-500/20 text-orange-700",
};

const memoryTypeLabels: Record<AgentMemoryType, string> = {
  working: "Working Memory",
  longterm: "Long-term Memory",
  context: "Context",
  episodic: "Episodic Memory",
};

export function AgentMemoryDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = id === "new";

  const { data: memory, isLoading, error } = useAgentMemory(id || "");
  const createMutation = useCreateAgentMemory();
  const updateMutation = useUpdateAgentMemory(id || "");
  const deleteMutation = useDeleteAgentMemory();

  const [isEditing, setIsEditing] = useState(isNew);
  const [formData, setFormData] = useState({
    agent_id: "",
    memory_type: "working" as AgentMemoryType,
    content: "",
    importance_score: 0.5,
    ttl_days: 30,
  });

  // Initialize form data when memory is loaded
  if (memory && !isEditing && formData.content === "" && !isNew) {
    setFormData({
      agent_id: memory.agent_id,
      memory_type: memory.memory_type,
      content: memory.content || "",
      importance_score: memory.importance_score,
      ttl_days: memory.ttl_days,
    });
  }

  const handleSave = async () => {
    if (isNew) {
      try {
        await createMutation.mutateAsync({
          agent_id: formData.agent_id,
          memory_type: formData.memory_type,
          content: formData.content,
          importance_score: formData.importance_score,
          ttl_days: formData.ttl_days,
        });
        navigate("/agent-memories");
      } catch (err) {
        // Error handled by mutation
      }
      return;
    }

    if (!id) {
      return;
    }

    try {
      await updateMutation.mutateAsync({
        content: formData.content,
        importance_score: formData.importance_score,
      });
      setIsEditing(false);
    } catch (err) {
      // Error handled by mutation
    }
  };

  const handleDelete = () => {
    if (!id) return;

    if (confirm("Are you sure you want to delete this memory?")) {
      deleteMutation.mutate(id, {
        onSuccess: () => {
          navigate("/agent-memories");
        },
      });
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      </div>
    );
  }

  if (error && !isNew) {
    return (
      <div className="container mx-auto py-8">
        <div className="text-center py-12 text-destructive">
          <p>Failed to load memory</p>
          <p className="text-sm text-muted-foreground">{error.message}</p>
          <Button variant="outline" className="mt-4" onClick={() => navigate("/agent-memories")}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Memories
          </Button>
        </div>
      </div>
    );
  }

  if (isNew) {
    return (
      <div className="container mx-auto py-8">
        <Button variant="ghost" className="mb-4" onClick={() => navigate("/agent-memories")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Memories
        </Button>

        <Card>
          <CardHeader>
            <CardTitle>Create New Memory</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">
              Creating new memories via the dashboard is not yet implemented.
              Memories are currently created by agents through the API.
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const createdAt = memory ? new Date(memory.created_at) : null;
  const updatedAt = memory ? new Date(memory.updated_at) : null;
  const expiresAt = memory?.expires_at ? new Date(memory.expires_at) : null;

  return (
    <div className="container mx-auto py-8">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={() => navigate("/agent-memories")}>
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div className="flex items-center gap-3">
            <Brain className="h-8 w-8" />
            <h1 className="text-2xl font-bold">Memory Details</h1>
          </div>
        </div>
        <div className="flex gap-2">
          {isEditing ? (
            <>
              <Button variant="outline" onClick={() => setIsEditing(false)}>
                Cancel
              </Button>
              <Button onClick={handleSave} disabled={updateMutation.isPending}>
                <Save className="h-4 w-4 mr-2" />
                {updateMutation.isPending ? "Saving..." : "Save"}
              </Button>
            </>
          ) : (
            <>
              <Button variant="outline" onClick={() => setIsEditing(true)}>
                <Edit3 className="h-4 w-4 mr-2" />
                Edit
              </Button>
              <Button variant="destructive" onClick={handleDelete}>
                <Trash2 className="h-4 w-4 mr-2" />
                Delete
              </Button>
            </>
          )}
        </div>
      </div>

      {memory && (
        <div className="grid gap-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Memory Information</CardTitle>
                <Badge className={memoryTypeColors[memory.memory_type]}>
                  {memoryTypeLabels[memory.memory_type]}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Agent ID</label>
                  {isEditing ? (
                    <Input
                      value={formData.agent_id}
                      onChange={(e) => setFormData({ ...formData, agent_id: e.target.value })}
                      className="mt-1"
                    />
                  ) : (
                    <p className="mt-1 font-mono">{memory.agent_id}</p>
                  )}
                </div>
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Memory Type</label>
                  {isEditing ? (
                    <Select
                      value={formData.memory_type}
                      onValueChange={(value) =>
                        setFormData({ ...formData, memory_type: value as AgentMemoryType })
                      }
                    >
                      <SelectTrigger className="mt-1">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="working">Working Memory</SelectItem>
                        <SelectItem value="longterm">Long-term Memory</SelectItem>
                        <SelectItem value="context">Context</SelectItem>
                        <SelectItem value="episodic">Episodic Memory</SelectItem>
                      </SelectContent>
                    </Select>
                  ) : (
                    <p className="mt-1">{memoryTypeLabels[memory.memory_type]}</p>
                  )}
                </div>
              </div>

              <div>
                <label className="text-sm font-medium text-muted-foreground">Content</label>
                {isEditing ? (
                  <Textarea
                    value={formData.content}
                    onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                    className="mt-1 min-h-[200px]"
                  />
                ) : (
                  <p className="mt-1 whitespace-pre-wrap">{memory.content || "No content"}</p>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Metadata</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="flex items-center gap-2">
                  <Star className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <p className="text-sm text-muted-foreground">Importance</p>
                    <p className="font-medium">{memory.importance_score.toFixed(2)}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <FileText className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <p className="text-sm text-muted-foreground">Access Count</p>
                    <p className="font-medium">{memory.access_count}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <p className="text-sm text-muted-foreground">TTL (days)</p>
                    <p className="font-medium">{memory.ttl_days}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <p className="text-sm text-muted-foreground">Expires</p>
                    <p className="font-medium">{expiresAt ? expiresAt.toLocaleDateString() : "Never"}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Timestamps</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-muted-foreground">Created</p>
                  <p className="font-medium">
                    {createdAt ? createdAt.toLocaleString() : "Unknown"}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Last Updated</p>
                  <p className="font-medium">
                    {updatedAt ? updatedAt.toLocaleString() : "Unknown"}
                  </p>
                </div>
                {memory.last_accessed_at && (
                  <div>
                    <p className="text-sm text-muted-foreground">Last Accessed</p>
                    <p className="font-medium">
                      {new Date(memory.last_accessed_at).toLocaleString()}
                    </p>
                  </div>
                )}
                {memory.source_event_id && (
                  <div>
                    <p className="text-sm text-muted-foreground">Source Event ID</p>
                    <p className="font-medium font-mono text-xs">{memory.source_event_id}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {memory.structured_data && Object.keys(memory.structured_data).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Structured Data</CardTitle>
              </CardHeader>
              <CardContent>
                <pre className="bg-muted p-4 rounded-lg overflow-auto text-sm">
                  {JSON.stringify(memory.structured_data, null, 2)}
                </pre>
              </CardContent>
            </Card>
          )}

          {memory.embedding && memory.embedding.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>Embedding</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Vector with {memory.embedding.length} dimensions
                </p>
                <p className="font-mono text-xs mt-2">
                  [{memory.embedding.slice(0, 5).join(", ")}, ...]
                </p>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}
