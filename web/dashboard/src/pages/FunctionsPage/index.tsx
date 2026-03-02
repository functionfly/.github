import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Search, Loader2, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FunctionCardCompact } from "@/components/functions";
import { functionsApi } from "@/api/functions";
import { toast } from "sonner";
import type { FunctionConfig, FunctionCardData } from "@/types";

/**
 * Map FunctionConfig (API data) to FunctionCardData format
 */
function mapToFunctionCardData(fn: FunctionConfig & { runtime?: string; updated_at?: string }): FunctionCardData {
  return {
    id: fn.id,
    name: fn.name,
    description: `Function deployed on ${fn.providers.join(", ")} in ${fn.region}`,
    author: {
      id: fn.tenantId,
      username: "current-user",
      name: "Current User",
    },
    trustScore: 75, // Default trust score for user functions
    metrics: {
      executionCount: 0, // Default since not tracked in FunctionConfig
      averageLatency: 0,
      errorRate: 0,
    },
    pricing: {
      model: "free",
      pricePerCall: 0,
      currency: "USD",
    },
    isVerified: false,
    isDeterministic: true,
    rating: {
      average: 0,
      count: 0,
    },
    tags: fn.providers,
    category: "Edge Function",
    language: fn.runtime || "python",
    lastUpdated: fn.updated_at || fn.updatedAt,
    version: fn.version,
    isFavorite: false,
    isFeatured: false,
  };
}

export function FunctionsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<FunctionConfig | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ["functions"],
    queryFn: () => functionsApi.list(),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => functionsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functions"] });
      toast.success("Function deleted successfully");
      setDeleteDialogOpen(false);
      setFunctionToDelete(null);
    },
    onError: () => {
      toast.error("Failed to delete function");
      setDeleteDialogOpen(false);
    },
  });

  const functions = data?.functions ?? [];

  const filteredFunctions = functions.filter((fn) =>
    fn.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleDeleteClick = (fn: FunctionConfig) => {
    setFunctionToDelete(fn);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = () => {
    if (functionToDelete) {
      deleteMutation.mutate(functionToDelete.id);
    }
  };

  const handleCancelDelete = () => {
    setDeleteDialogOpen(false);
    setFunctionToDelete(null);
  };

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-white">Functions</h1>
            <p className="text-text-secondary">Manage and deploy your edge functions</p>
          </div>
        </div>
        <Card className="p-12 text-center">
          <p className="text-text-secondary">Failed to load functions. Please try again.</p>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Functions</h1>
          <p className="text-text-secondary">Manage and deploy your edge functions</p>
        </div>
        <Button className="gap-2" onClick={() => navigate("/functions/new")}>
          <Plus className="w-4 h-4" />
          Deploy New
        </Button>
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            placeholder="Search functions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <Button variant="outline">Filter</Button>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-text-muted" />
        </div>
      )}

      {/* Functions Grid */}
      {!isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredFunctions.map((fn) => (
            <FunctionCardCompact
              key={fn.id}
              data={mapToFunctionCardData(fn)}
              onView={(id) => navigate(`/functions/${id}`)}
              onEdit={(id) => navigate(`/functions/${id}/edit`)}
              onDelete={() => handleDeleteClick(fn)}
            />
          ))}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && filteredFunctions.length === 0 && (
        <Card className="p-12 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
            <Plus className="w-8 h-8 text-text-muted" />
          </div>
          <h3 className="text-lg font-medium text-white mb-2">
            {searchQuery ? "No functions match your search" : "No functions yet"}
          </h3>
          <p className="text-text-secondary mb-6">
            {searchQuery ? "Try a different search term" : "Deploy your first function to get started"}
          </p>
          {!searchQuery && (
            <Button onClick={() => navigate("/functions/new")}>
              <Plus className="w-4 h-4 mr-2" />
              Deploy Function
            </Button>
          )}
        </Card>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              Delete Function
            </DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{functionToDelete?.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={handleCancelDelete}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                "Delete"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
