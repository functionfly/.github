import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Search, MoreVertical, Rocket, Edit, Trash2, Eye, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { StatusBadge } from "@/components/common/StatusBadge";
import { ProviderIcon } from "@/components/common/ProviderIcon";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { functionsApi } from "@/api/functions";
import { toast } from "sonner";
import type { FunctionConfig } from "@/types";

export function FunctionsPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState("");

  const { data, isLoading, error } = useQuery({
    queryKey: ["functions"],
    queryFn: () => functionsApi.list(),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => functionsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["functions"] });
      toast.success("Function deleted successfully");
    },
    onError: () => {
      toast.error("Failed to delete function");
    },
  });

  const functions = data?.functions ?? [];

  const filteredFunctions = functions.filter((fn) =>
    fn.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleDelete = (fn: FunctionConfig) => {
    if (confirm(`Are you sure you want to delete "${fn.name}"?`)) {
      deleteMutation.mutate(fn.id);
    }
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

      {/* Functions List */}
      {!isLoading && (
        <div className="space-y-3">
          {filteredFunctions.map((fn) => (
            <Card key={fn.id} className="hover:border-[#6366f1]/30 transition-colors">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-[#6366f1]/10 flex items-center justify-center">
                      <Rocket className="w-5 h-5 text-[#6366f1]" />
                    </div>
                    <div>
                      <h3 className="font-medium text-white">{fn.name}</h3>
                      <p className="text-sm text-text-muted">
                        {fn.runtime || "unknown runtime"}
                        {fn.updated_at ? ` • Updated ${new Date(fn.updated_at).toLocaleDateString()}` : ""}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-6">
                    {/* Providers */}
                    {fn.providers && fn.providers.length > 0 && (
                      <div className="flex items-center gap-2">
                        <span className="text-sm text-text-muted">Providers:</span>
                        <div className="flex -space-x-1">
                          {fn.providers.map((provider) => (
                            <div
                              key={typeof provider === "string" ? provider : provider.id}
                              className="w-6 h-6 rounded-full bg-bg-tertiary border border-white/8 flex items-center justify-center"
                            >
                              <ProviderIcon
                                provider={typeof provider === "string" ? provider : provider.id}
                                size="sm"
                              />
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Status */}
                    <StatusBadge status={(fn.status as any) || "online"} />

                    {/* Actions */}
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="text-text-secondary">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="bg-bg-tertiary border-white/8">
                        <DropdownMenuItem className="gap-2" onClick={() => navigate(`/functions/${fn.id}`)}>
                          <Eye className="w-4 h-4" />
                          View Details
                        </DropdownMenuItem>
                        <DropdownMenuItem className="gap-2" onClick={() => navigate(`/functions/${fn.id}/edit`)}>
                          <Edit className="w-4 h-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem className="gap-2 text-red-400" onClick={() => handleDelete(fn)}>
                          <Trash2 className="w-4 h-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {!isLoading && filteredFunctions.length === 0 && (
        <Card className="p-12 text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
            <Rocket className="w-8 h-8 text-text-muted" />
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
    </div>
  );
}
