import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Key,
  Plus,
  RotateCcw,
  Trash2,
  AlertTriangle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { toast } from "sonner";
import {
  APIKeyList,
  CreateAPIKeyModal,
  APIKeyRotationModal,
} from "@/components/api-keys";
import {
  APIKey,
  APIKeyFilters,
} from "@/types/api-key";
import { apiKeysService, getStoredApiKey } from "@/services/api-keys";

const PAGE_SIZE = 10;

export function APIKeysPage() {
  const queryClient = useQueryClient();
  const [filters, setFilters] = useState<APIKeyFilters>({});
  const [page, setPage] = useState(1);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showRotationModal, setShowRotationModal] = useState(false);
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null);
  const [deleteKey, setDeleteKey] = useState<APIKey | null>(null);

  // Check for stored newly created API key
  useEffect(() => {
    const stored = getStoredApiKey();
    if (stored) {
      toast.success("API key created successfully", {
        description: "Copy your new API key now. It won't be shown again.",
        duration: 10000,
      });
    }
  }, []);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ["api-keys", filters, page],
    queryFn: () => apiKeysService.listKeys(filters, page, PAGE_SIZE),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiKeysService.deleteKey(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      toast.success("API key deleted");
      setDeleteKey(null);
    },
    onError: (error) => {
      toast.error("Failed to delete API key", {
        description: error instanceof Error ? error.message : "Unknown error",
      });
    },
  });

  const handleFiltersChange = (newFilters: APIKeyFilters) => {
    setFilters(newFilters);
    setPage(1);
  };

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  const handleRotate = (key: APIKey) => {
    setSelectedKey(key);
    setShowRotationModal(true);
  };

  const handleDelete = (key: APIKey) => {
    setDeleteKey(key);
  };

  const confirmDelete = () => {
    if (deleteKey) {
      deleteMutation.mutate(deleteKey.id);
    }
  };

  const apiKeys = data?.data || [];
  const total = data?.total || 0;

  return (
    <div className="container mx-auto py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold">API Keys</h1>
          <p className="text-muted-foreground mt-1">
            Manage your API keys for programmatic access
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="w-5 h-5" />
            Your API Keys
          </CardTitle>
        </CardHeader>
        <CardContent>
          <APIKeyList
            apiKeys={apiKeys}
            isLoading={isLoading}
            total={total}
            page={page}
            pageSize={PAGE_SIZE}
            filters={filters}
            onFiltersChange={handleFiltersChange}
            onPageChange={handlePageChange}
            onCreateNew={() => setShowCreateModal(true)}
            onRotate={handleRotate}
            onDelete={handleDelete}
          />
        </CardContent>
      </Card>

      {/* Create Modal */}
      <CreateAPIKeyModal
        open={showCreateModal}
        onOpenChange={setShowCreateModal}
      />

      {/* Rotation Modal */}
      <APIKeyRotationModal
        open={showRotationModal}
        onOpenChange={setShowRotationModal}
        apiKey={selectedKey}
        onSuccess={() => {
          refetch();
          toast.success("API key rotated successfully");
        }}
      />

      {/* Delete Confirmation */}
      <AlertDialog open={!!deleteKey} onOpenChange={() => setDeleteKey(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              Delete API Key
            </AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the API key "{deleteKey?.name}"?
              This action cannot be undone and any applications using this key
              will stop working.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-red-600 hover:bg-red-700"
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
