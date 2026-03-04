/**
 * SecretDetail - View and decrypt secret values
 * Handles passphrase-based key derivation and AES-256-GCM decryption
 */

import { useState, useCallback } from "react";
import {
  X,
  Eye,
  EyeOff,
  Copy,
  Key,
  Shield,
  Lock,
  FileKey,
  Clock,
  RefreshCw,
  Trash2,
  Edit3,
  AlertTriangle,
  Check,
  Loader2,
  ShieldCheck,
} from "lucide-react";
import { formatDistanceToNow, format } from "date-fns";
import { cn } from "@/lib/utils";
import { VaultCrypto } from "@/utils/vault-crypto";
import {
  useVaultSecret,
  useDeleteSecret,
  useUpdateSecret,
} from "@/hooks/useVault";
import type { SecretType, Secret } from "@/types/vault";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
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
import { TokenGenerator } from "./TokenGenerator";
import { SecretForm } from "./SecretForm";

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

export interface SecretDetailProps {
  secretId: string;
  onClose: () => void;
}

export function SecretDetail({ secretId, onClose }: SecretDetailProps) {
  const { data: secret, isLoading, error } = useVaultSecret(secretId);
  const deleteSecret = useDeleteSecret();
  const updateSecret = useUpdateSecret(secretId);

  const [passphrase, setPassphrase] = useState("");
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [decryptedValue, setDecryptedValue] = useState<string | null>(null);
  const [isDecrypting, setIsDecrypting] = useState(false);
  const [decryptError, setDecryptError] = useState<string | null>(null);
  const [showDecrypted, setShowDecrypted] = useState(false);
  const [copied, setCopied] = useState(false);

  const [isTokenModalOpen, setIsTokenModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

  // Handle decryption
  const handleDecrypt = useCallback(async () => {
    if (!secret || !passphrase) return;

    setIsDecrypting(true);
    setDecryptError(null);

    try {
      const encryptedData = VaultCrypto.fromPayload(secret.encrypted_data);
      const decrypted = await VaultCrypto.decryptWithPassphrase(
        encryptedData,
        passphrase
      );
      setDecryptedValue(decrypted);
      setShowDecrypted(true);
    } catch (err) {
      console.error("Decryption failed:", err);
      setDecryptError(
        "Failed to decrypt. Please check your passphrase and try again."
      );
    } finally {
      setIsDecrypting(false);
    }
  }, [secret, passphrase]);

  // Copy decrypted value to clipboard
  const handleCopy = useCallback(async () => {
    if (!decryptedValue) return;

    try {
      await navigator.clipboard.writeText(decryptedValue);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  }, [decryptedValue]);

  // Handle deletion
  const handleDelete = async () => {
    await deleteSecret.mutateAsync(secretId);
    onClose();
  };

  // Loading state
  if (isLoading) {
    return (
      <Dialog open onOpenChange={onClose}>
        <DialogContent className="max-w-2xl">
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  // Error state
  if (error || !secret) {
    return (
      <Dialog open onOpenChange={onClose}>
        <DialogContent className="max-w-2xl">
          <div className="text-center py-8">
            <AlertTriangle className="mx-auto h-12 w-12 text-error mb-4" />
            <h3 className="text-lg font-semibold text-error mb-2">
              Failed to load secret
            </h3>
            <p className="text-text-secondary">
              {error instanceof Error ? error.message : "Secret not found"}
            </p>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  const Icon = secretTypeIcons[secret.secret_type];

  return (
    <>
      <Dialog open onOpenChange={onClose}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <div className="flex items-start gap-4">
              <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-brand-500/10 to-purple-500/10 flex items-center justify-center shrink-0">
                <Icon className="h-6 w-6 text-brand-500" />
              </div>
              <div className="flex-1 min-w-0">
                <DialogTitle className="text-xl break-words">
                  {secret.name}
                </DialogTitle>
                <div className="flex items-center gap-2 mt-1 flex-wrap">
                  <Badge variant="secondary">
                    {secretTypeLabels[secret.secret_type]}
                  </Badge>
                  {secret.scopes?.map((scope) => (
                    <Badge key={scope} variant="outline" className="text-xs">
                      {scope}
                    </Badge>
                  ))}
                </div>
              </div>
            </div>
          </DialogHeader>

          <div className="space-y-6">
            {/* Description */}
            {secret.description && (
              <div>
                <Label className="text-text-muted">Description</Label>
                <p className="text-card-foreground mt-1">{secret.description}</p>
              </div>
            )}

            {/* Decryption Section */}
            <div className="rounded-lg border border-border-subtle p-4 space-y-4">
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-5 w-5 text-brand-500" />
                <h4 className="font-semibold text-card-foreground">
                  Decrypt Secret
                </h4>
              </div>

              {!decryptedValue ? (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="passphrase">Passphrase</Label>
                    <div className="relative">
                      <Input
                        id="passphrase"
                        type={showPassphrase ? "text" : "password"}
                        placeholder="Enter your encryption passphrase"
                        value={passphrase}
                        onChange={(e) => setPassphrase(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") handleDecrypt();
                        }}
                        className={cn(
                          "pr-10",
                          decryptError && "border-error"
                        )}
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassphrase(!showPassphrase)}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary"
                      >
                        {showPassphrase ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                    {decryptError && (
                      <p className="text-sm text-error flex items-center gap-1">
                        <AlertTriangle className="h-3.5 w-3.5" />
                        {decryptError}
                      </p>
                    )}
                  </div>
                  <Button
                    onClick={handleDecrypt}
                    disabled={!passphrase || isDecrypting}
                    className="w-full"
                  >
                    {isDecrypting ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Decrypting...
                      </>
                    ) : (
                      <>
                        <Lock className="h-4 w-4 mr-2" />
                        Decrypt Secret
                      </>
                    )}
                  </Button>
                </>
              ) : (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <Label>Decrypted Value</Label>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setShowDecrypted(!showDecrypted)}
                        className="text-text-muted hover:text-text-primary p-1"
                        title={showDecrypted ? "Hide" : "Show"}
                      >
                        {showDecrypted ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                      <button
                        onClick={handleCopy}
                        className="text-text-muted hover:text-text-primary p-1"
                        title="Copy to clipboard"
                      >
                        {copied ? (
                          <Check className="h-4 w-4 text-success" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </div>
                  <div className="relative">
                    <textarea
                      readOnly
                      value={showDecrypted ? decryptedValue : "•".repeat(20)}
                      className="w-full min-h-[100px] p-3 rounded-lg border border-border-subtle bg-bg-secondary font-mono text-sm resize-none"
                    />
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setDecryptedValue(null);
                      setPassphrase("");
                      setDecryptError(null);
                    }}
                    className="w-full"
                  >
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Lock & Clear
                  </Button>
                </div>
              )}
            </div>

            {/* Metadata */}
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <Label className="text-text-muted text-xs">Created</Label>
                <p className="flex items-center gap-1.5 text-card-foreground mt-1">
                  <Clock className="h-3.5 w-3.5" />
                  {format(new Date(secret.created_at), "MMM d, yyyy HH:mm")}
                </p>
              </div>
              <div>
                <Label className="text-text-muted text-xs">Last Accessed</Label>
                <p className="flex items-center gap-1.5 text-card-foreground mt-1">
                  <Clock className="h-3.5 w-3.5" />
                  {secret.last_accessed_at
                    ? formatDistanceToNow(new Date(secret.last_accessed_at), {
                        addSuffix: true,
                      })
                    : "Never"}
                </p>
              </div>
              <div>
                <Label className="text-text-muted text-xs">Access Count</Label>
                <p className="text-card-foreground mt-1">{secret.access_count}</p>
              </div>
              <div>
                <Label className="text-text-muted text-xs">Key Version</Label>
                <p className="text-card-foreground mt-1">
                  v{secret.encrypted_data.key_version}
                </p>
              </div>
            </div>

            {/* Actions */}
            <div className="flex flex-wrap gap-2 pt-4 border-t border-border-subtle">
              <Button
                variant="outline"
                onClick={() => setIsTokenModalOpen(true)}
                className="flex-1"
              >
                <Key className="h-4 w-4 mr-2" />
                Generate Token
              </Button>
              <Button
                variant="outline"
                onClick={() => setIsEditModalOpen(true)}
                className="flex-1"
              >
                <Edit3 className="h-4 w-4 mr-2" />
                Edit
              </Button>
              <Button
                variant="destructive"
                onClick={() => setIsDeleteDialogOpen(true)}
                className="flex-1"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Token Generator Modal */}
      {isTokenModalOpen && (
        <TokenGenerator
          secretId={secretId}
          onClose={() => setIsTokenModalOpen(false)}
          onGenerated={() => {
            // Token was generated
          }}
        />
      )}

      {/* Edit Modal */}
      {isEditModalOpen && (
        <Dialog open onOpenChange={setIsEditModalOpen}>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Edit Secret</DialogTitle>
            </DialogHeader>
            <SecretForm
              onSubmit={() => setIsEditModalOpen(false)}
              onCancel={() => setIsEditModalOpen(false)}
              initialData={secret}
            />
          </DialogContent>
        </Dialog>
      )}

      {/* Delete Confirmation */}
      <AlertDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Secret</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete <strong>{secret.name}</strong>?
              This action cannot be undone. All access tokens for this secret
              will also be revoked.
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
    </>
  );
}
