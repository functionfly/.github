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

export interface SecretListProps {
  className?: string;
}

export function SecretList({ className }: SecretListProps) {
  const { data: secrets, isLoading, error } = useVaultSecrets();
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
  const filteredSecrets = secrets?.filter((secret) => {
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
      <div className={cn("space-y-4", className)}>
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="rounded-lg border border-border-subtle">
          <div className="space-y-2 p-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={cn("rounded-lg border border-error/20 bg-error-glow p-8 text-center", className)}>
        <Shield className="mx-auto h-12 w-12 text-error mb-4" />
        <h3 className="text-lg font-semibold text-error mb-2">Failed to load secrets</h3>
        <p className="text-text-secondary mb-4">
          {error instanceof Error ? error.message : "An unexpected error occurred"}
        </p>
        <Button onClick={() => window.location.reload()} variant="outline">
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className={cn("space-y-4", className)}>
      {/* Header with search and create button */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
          <Input
            placeholder="Search secrets..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex items-center gap-3">
          {/* Secrets limit indicator */}
          {secretsLimit > 0 && (
            <div className={cn(
              "flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm",
              hasReachedLimit
                ? "bg-warning/10 text-warning"
                : "bg-muted text-muted-foreground"
            )}>
              <AlertCircle className="h-4 w-4" />
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
            className="gap-2"
            disabled={hasReachedLimit}
          >
            <Plus className="h-4 w-4" />
            Create Secret
          </Button>
        </div>
      </div>

      {/* Empty state */}
      {filteredSecrets?.length === 0 && (
        <div className="rounded-lg border border-border-subtle bg-card p-12 text-center">
          <div className="mx-auto h-16 w-16 rounded-full bg-gradient-to-br from-brand-500/20 to-purple-500/20 flex items-center justify-center mb-4">
            <Key className="h-8 w-8 text-brand-500" />
          </div>
          <h3 className="text-lg font-semibold text-card-foreground mb-2">
            {searchQuery ? "No secrets found" : canCreateSecrets ? "No secrets yet" : "Secrets not available"}
          </h3>
          <p className="text-text-secondary max-w-sm mx-auto mb-6">
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
            >
              <Plus className="h-4 w-4 mr-2" />
              Create Secret
            </Button>
          )}
          {!canCreateSecrets && (
            <Button onClick={() => window.location.href = '/pricing'}>
              Upgrade Plan
            </Button>
          )}
        </div>
      )}

      {/* Secrets table */}
      {filteredSecrets && filteredSecrets.length > 0 && (
        <div className="rounded-lg border border-border-subtle overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Last Accessed</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="w-16"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredSecrets.map((secret) => {
                const Icon = secretTypeIcons[secret.secret_type];
                return (
                  <TableRow
                    key={secret.id}
                    className="cursor-pointer"
                    onClick={() => setSelectedSecretId(secret.id)}
                  >
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-brand-500/10 to-purple-500/10 flex items-center justify-center">
                          <Icon className="h-4 w-4 text-brand-500" />
                        </div>
                        <span className="font-medium text-card-foreground">
                          {secret.name}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={secretTypeVariants[secret.secret_type]}>
                        {secretTypeLabels[secret.secret_type]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className="text-text-secondary line-clamp-1 max-w-[200px]">
                        {secret.description || "—"}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1.5 text-text-secondary">
                        <Clock className="h-3.5 w-3.5" />
                        <span className="text-sm">
                          {secret.last_accessed_at
                            ? formatDistanceToNow(new Date(secret.last_accessed_at), {
                                addSuffix: true,
                              })
                            : "Never"}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-text-secondary">
                        {formatDistanceToNow(new Date(secret.created_at), {
                          addSuffix: true,
                        })}
                      </span>
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={(e) => e.stopPropagation()}
                            aria-label="More actions"
                          >
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedSecretId(secret.id);
                            }}
                          >
                            <Eye className="h-4 w-4 mr-2" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-error focus:text-error"
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
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Create New Secret</DialogTitle>
            <DialogDescription>
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
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Secret</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete <strong>{secretToDelete?.name}</strong>?
              This action cannot be undone. All access tokens for this secret will also be revoked.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
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
