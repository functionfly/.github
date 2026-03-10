import { useState } from "react";
import { useParams } from "react-router-dom";
import { toast } from "sonner";
import { AlertTriangle } from "lucide-react";
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
import {
  APIKeyDetails,
  APIKeyRotationModal,
} from "@/components/api-keys";
import { APIKey } from "@/types/api-key";
import { apiKeysService } from "@/services/api-keys";

export function APIKeyDetailPage() {
  const { keyId } = useParams<{ keyId: string }>();
  const [showRotationModal, setShowRotationModal] = useState(false);
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null);
  const [deleteKey, setDeleteKey] = useState<APIKey | null>(null);

  const handleRotate = (key: APIKey) => {
    setSelectedKey(key);
    setShowRotationModal(true);
  };

  const handleDelete = async (key: APIKey) => {
    setDeleteKey(key);
  };

  const confirmDelete = async () => {
    if (!deleteKey) return;

    try {
      await apiKeysService.deleteKey(deleteKey.id);
      toast.success("API key deleted");
      window.location.href = "/dashboard/api-keys";
    } catch (error) {
      toast.error("Failed to delete API key", {
        description: error instanceof Error ? error.message : "Unknown error",
      });
    }
  };

  if (!keyId) {
    return (
      <div className="container mx-auto py-8">
        <p>API key ID is required</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8">
      <APIKeyDetails
        onRotate={handleRotate}
        onDelete={handleDelete}
      />

      {/* Rotation Modal */}
      <APIKeyRotationModal
        open={showRotationModal}
        onOpenChange={setShowRotationModal}
        apiKey={selectedKey}
        onSuccess={() => {
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
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
