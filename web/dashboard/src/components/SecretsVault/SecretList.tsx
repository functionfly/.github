/**
 * SecretList - List view of all vault secrets
 * Displays secrets in a table with metadata and action buttons
 */

import { useState } from "react";
import {
  Plus,
  Key,
  Trash2,
  Eye,
  Shield,
  Lock,
  FileKey,
  Clock,
  MoreHorizontal,
  Search,
  AlertCircle,
  AlertTriangle,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";
import { useVaultSecrets, useDeleteSecret } from "@/hooks/useVault";
import { useAuthStore } from "@/stores/authStore";
import { getSecretsLimit, formatSecretsRemaining } from "@/lib/plan-utils";
import type { SecretMetadata, SecretType } from "@/types/vault";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
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
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { SecretForm } from "./SecretForm";
import { SecretDetail } from "./SecretDetail";

// Secret type icon mapping
const secretTypeIcons: Record<SecretType, typeof Key> = {
  api_key: Key,
  oauth_token: Shield,
  password: Lock,
  certificate: FileKey,
};

// Secret type label mapping
const secretTypeLabels: Record<SecretType, string> = {
  api_key: "API Key",
  oauth_token: "OAuth Token",
  password: "Password",
  certificate: "Certificate",
};

// Secret type badge variant mapping
const secretTypeVariants: Record<SecretType, "default" | "secondary" | "outline" | "success" | "warning"> = {
  api_key: "default",
  oauth_token: "success",
  password: "warning",
  certificate: "outline",
};

// Stale threshold in days
const STALE_THRESHOLD_DAYS = 90;

function getRotationStatus(secret: SecretMetadata): { label: string; variant: 'default' | 'secondary' | 'outline' | 'success' | 'warning'; icon: typeof Clock } {
  const updatedAt = secret.updated_at ? new Date(secret.updated_at) : new Date(secret.created_at);
  const daysSinceUpdate = Math.floor((Date.now() - updatedAt.getTime()) / (1000 * 60 * 60 * 24));

  if (daysSinceUpdate > STALE_THRESHOLD_DAYS) {
    return { label: 'Stale', variant: 'warning', icon: AlertTriangle };
  }
  if (secret.current_version && secret.current_version > 1) {
    return { label: 'Rotated', variant: 'success', icon: Shield };
  }
  return { label: 'Active', variant: 'secondary', icon: Clock };
}

export interface SecretListProps {
  className?: string;
}

export function SecretList({ className }: SecretListProps) {
  const { data: secretsResponse, isLoading, error } = useVaultSecrets();
  const secrets = secretsResponse?.secrets ?? [];
  const deleteSecret = useDeleteSecret();
  const user = useAuthStore((state) => state.user);
  const userPlan = user?.plan;
  const secretsLimit = getSecretsLimit(userPlan);
  const currentSecretCount = secrets?.length ?? 0;

  const [searchQuery, setSearchQuery] = useState("");
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [selectedSecretId, setSelectedSecretId] = useState<string | null>(null);
  const [secretToDelete, setSecretToDelete] = useState<SecretMetadata | null>(null);

  // Check if user has reached their secrets limit
  const hasReachedLimit = secretsLimit > 0 && currentSecretCount >= secretsLimit;
  const canCreateSecrets = secretsLimit > 0;

  // Filter secrets based on search query
  const filteredSecrets = secrets.filter((secret) => {
    const query = searchQuery.toLowerCase();
    return (
      secret.name.toLowerCase().includes(query) ||
      secret.description?.toLowerCase().includes(query) ||
      secret.secret_type.toLowerCase().includes(query)
    );
  });

  // Handle secret deletion
  const handleDelete = async () => {
    if (secretToDelete) {
      await deleteSecret.mutateAsync(secretToDelete.id);
      setSecretToDelete(null);
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className={cn("secrets-loading-container", className)}>
        <div className="secrets-loading-header">
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="secrets-table-container">
          <div className="secrets-loading-container p-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="secrets-skeleton-row secrets-skeleton" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={cn("secrets-error-state", className)}>
        <Shield className="secrets-error-icon mx-auto mb-4" />
        <h3 className="secrets-error-title">Failed to load secrets</h3>
        <p className="secrets-error-description">
          {error instanceof Error ? error.message : "An unexpected error occurred"}
        </p>
        <Button onClick={() => window.location.reload()} className="btn-secrets-outline">
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header with search and create button */}
      <div className="secrets-toolbar">
        <div className="secrets-search-container">
          <Search className="secrets-search-icon" />
          <Input
            placeholder="Search secrets..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="secrets-search-input pl-10"
          />
        </div>
        <div className="secrets-toolbar-actions">
          {/* Secrets limit indicator */}
          {secretsLimit > 0 && (
            <div className={cn(
              "secrets-limit-indicator",
              hasReachedLimit
                ? "secrets-limit-indicator-warning"
                : "secrets-limit-indicator-normal"
            )}>
              <AlertCircle className="secrets-limit-icon" />
              <span className="hidden sm:inline">
                {formatSecretsRemaining(currentSecretCount, userPlan)}
              </span>
              <span className="sm:hidden">
                {currentSecretCount}/{secretsLimit}
              </span>
            </div>
          )}
          <Button
            onClick={() => setIsCreateModalOpen(true)}
            className="btn-secrets-create"
            disabled={hasReachedLimit}
          >
            <Plus className="h-4 w-4" />
            Create Secret
          </Button>
        </div>
      </div>

      {/* Empty state */}
      {filteredSecrets?.length === 0 && (
        <div className="secrets-empty-state">
          <div className="secrets-empty-icon">
            <Key className="h-8 w-8" />
          </div>
          <h3 className="secrets-empty-title">
            {searchQuery ? "No secrets found" : canCreateSecrets ? "No secrets yet" : "Secrets not available"}
          </h3>
          <p className="secrets-empty-description">
            {searchQuery
              ? "No secrets match your search query. Try a different search term."
              : canCreateSecrets
                ? "Create your first secret to securely store API keys, passwords, and other sensitive data."
                : "Secrets are not available on your current plan. Upgrade to Starter or higher to use secrets."}
          </p>
          {!searchQuery && canCreateSecrets && (
            <Button
              onClick={() => setIsCreateModalOpen(true)}
              disabled={hasReachedLimit}
              className="btn-secrets-create"
            >
              <Plus className="h-4 w-4 mr-2" />
              Create Secret
            </Button>
          )}
          {!canCreateSecrets && (
            <Button onClick={() => window.location.href = '/pricing'} className="btn-bundle-primary">
              Upgrade Plan
            </Button>
          )}
        </div>
      )}

      {/* Secrets table */}
      {filteredSecrets && filteredSecrets.length > 0 && (
        <div className="secrets-table-container">
          <Table>
            <TableHeader>
              <TableRow className="secrets-table-header-row">
                <TableHead className="secrets-table-head">Name</TableHead>
                <TableHead className="secrets-table-head">Type</TableHead>
                <TableHead className="secrets-table-head">Description</TableHead>
                <TableHead className="secrets-table-head">Rotation</TableHead>
                <TableHead className="secrets-table-head">Last Accessed</TableHead>
                <TableHead className="secrets-table-head">Created</TableHead>
                <TableHead className="w-16"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredSecrets.map((secret) => {
                const Icon = secretTypeIcons[secret.secret_type];
                const rotationStatus = getRotationStatus(secret);
                return (
                  <TableRow
                    key={secret.id}
                    className="secrets-table-body-row"
                    onClick={() => setSelectedSecretId(secret.id)}
                  >
                    <TableCell className="secrets-table-cell">
                      <div className="secrets-name-cell">
                        <div className="secrets-name-icon">
                          <Icon className="h-4 w-4" />
                        </div>
                        <span className="secrets-name-text">
                          {secret.name}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="secrets-table-cell">
                      <span className={`secret-type-badge secret-type-badge-${secret.secret_type === 'api_key' ? 'api' : secret.secret_type === 'oauth_token' ? 'oauth' : secret.secret_type === 'password' ? 'password' : 'certificate'}`}>
                        {secretTypeLabels[secret.secret_type]}
                      </span>
                    </TableCell>
                    <TableCell className="secrets-table-cell">
                      <span className="secrets-description-cell">
                        {secret.description || "—"}
                      </span>
                    </TableCell>
                    <TableCell className="secrets-table-cell">
                      <span className={`rotation-badge rotation-badge-${rotationStatus.label.toLowerCase()}`}>
                        <rotationStatus.icon className="rotation-badge-icon" />
                        {rotationStatus.label}
                      </span>
                    </TableCell>
                    <TableCell className="secrets-table-cell">
                      <div className="secrets-time-cell">
                        <Clock className="secrets-time-icon" />
                        <span className="secrets-time-text">
                          {secret.last_accessed_at
                            ? formatDistanceToNow(new Date(secret.last_accessed_at), {
                                addSuffix: true,
                              })
                            : "Never"}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="secrets-table-cell">
                      <div className="secrets-time-cell">
                        <Clock className="secrets-time-icon" />
                        <span className="secrets-time-text">
                          {formatDistanceToNow(new Date(secret.created_at), {
                            addSuffix: true,
                          })}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="secrets-actions-cell">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="secrets-action-btn"
                            onClick={(e) => e.stopPropagation()}
                            aria-label="More actions"
                          >
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="secrets-dialog-content">
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedSecretId(secret.id);
                            }}
                            className="focus:bg-[var(--secrets-input-bg)]"
                          >
                            <Eye className="h-4 w-4 mr-2" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="btn-secrets-delete focus:bg-[var(--secrets-input-bg)]"
                            onClick={(e) => {
                              e.stopPropagation();
                              setSecretToDelete(secret);
                            }}
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Create Secret Modal */}
      <Dialog open={isCreateModalOpen} onOpenChange={setIsCreateModalOpen}>
        <DialogContent className="secrets-dialog-content max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader className="secrets-dialog-header">
            <DialogTitle className="secrets-dialog-title">Create New Secret</DialogTitle>
            <DialogDescription className="secrets-dialog-description">
              Store sensitive data securely with client-side encryption.
            </DialogDescription>
          </DialogHeader>
          <SecretForm
            onSubmit={() => setIsCreateModalOpen(false)}
            onCancel={() => setIsCreateModalOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Secret Detail Modal */}
      {selectedSecretId && (
        <SecretDetail
          secretId={selectedSecretId}
          onClose={() => setSelectedSecretId(null)}
        />
      )}

      {/* Delete Confirmation */}
      <AlertDialog
        open={!!secretToDelete}
        onOpenChange={(open) => !open && setSecretToDelete(null)}
      >
        <AlertDialogContent className="secrets-delete-content">
          <AlertDialogHeader className="secrets-delete-header">
            <AlertDialogTitle className="secrets-delete-title">Delete Secret</AlertDialogTitle>
            <AlertDialogDescription className="secrets-delete-description">
              Are you sure you want to delete <strong>{secretToDelete?.name}</strong>?
              This action cannot be undone. All access tokens for this secret will also be revoked.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter className="secrets-delete-footer">
            <AlertDialogCancel className="btn-secrets-cancel">Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="btn-secrets-delete"
              disabled={deleteSecret.isPending}
            >
              {deleteSecret.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}