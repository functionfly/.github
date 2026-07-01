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
  History,
  AlertCircle,
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
import { SecretVersionHistory } from "./SecretVersionHistory";

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
  const [isVersionHistoryOpen, setIsVersionHistoryOpen] = useState(false);
  const [showPassphraseWarning, setShowPassphraseWarning] = useState(true);

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
        <DialogContent className="secrets-dialog-content">
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-[var(--status-ok)]" />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  // Error state
  if (error || !secret) {
    return (
      <Dialog open onOpenChange={onClose}>
        <DialogContent className="secrets-dialog-content">
          <div className="secrets-error-state">
            <AlertTriangle className="secrets-error-icon" />
            <h3 className="secrets-error-title">
              Failed to load secret
            </h3>
            <p className="secrets-error-description">
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
        <DialogContent className="secrets-dialog-content">
          <DialogHeader className="secrets-dialog-header">
            <div className="secrets-detail-header">
              <div className="secrets-detail-icon">
                <Icon className="h-6 w-6" />
              </div>
              <div className="flex-1 min-w-0">
                <DialogTitle className="secrets-detail-title">
                  {secret.name}
                </DialogTitle>
                <div className="secrets-detail-meta">
                  <span className={cn(
                    "secret-type-badge",
                    secret.secret_type === 'api_key' && "secret-type-badge-api",
                    secret.secret_type === 'oauth_token' && "secret-type-badge-oauth",
                    secret.secret_type === 'password' && "secret-type-badge-password",
                    secret.secret_type === 'certificate' && "secret-type-badge-certificate"
                  )}>
                    {secretTypeLabels[secret.secret_type]}
                  </span>
                  {secret.scopes?.map((scope) => (
                    <span key={scope} className="rotation-badge rotation-badge-active" style={{ fontSize: '0.65rem' }}>
                      {scope}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </DialogHeader>

          <div className="space-y-6">
            {/* Description */}
            {secret.description && (
              <div className="secrets-detail-section">
                <Label className="secrets-detail-label">Description</Label>
                <p className="secrets-detail-value">{secret.description}</p>
              </div>
            )}

            {/* Passphrase Recovery Warning */}
            {showPassphraseWarning && !decryptedValue && (
              <div className="rounded-lg rgba(232,196,104,0.04)  border border-[rgba(232,196,104,0.15)]  p-4">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-[var(--status-pending)] mt-0.5 shrink-0" />
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <p className="text-sm font-medium text-[var(--status-pending)] ">
                        Zero-Knowledge Encryption
                      </p>
                      <button
                        onClick={() => setShowPassphraseWarning(false)}
                        className="text-[var(--status-pending)] hover:text-[var(--status-pending)] "
                        aria-label="Dismiss warning"
                      >
                        <X className="h-4 w-4" />
                      </button>
                    </div>
                    <p className="text-sm text-[var(--status-pending)]  mt-1">
                      Your passphrase is never sent to the server. We cannot recover your passphrase if you lose it.
                      Store it securely in a password manager.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Decryption Section */}
            <div className="secrets-decrypt-section">
              <div className="secrets-decrypt-header">
                <ShieldCheck className="h-5 w-5" />
                <h4 className="secrets-decrypt-title">
                  Decrypt Secret
                </h4>
              </div>

              {!decryptedValue ? (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="passphrase" className="secrets-detail-label">Passphrase</Label>
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
                          "secrets-decrypt-input pr-10",
                          decryptError && "border-error"
                        )}
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassphrase(!showPassphrase)}
                        className="secrets-toggle-btn absolute right-3 top-1/2 -translate-y-1/2"
                      >
                        {showPassphrase ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                    {decryptError && (
                      <p className="secrets-form-error flex items-center gap-1">
                        <AlertTriangle className="h-3.5 w-3.5" />
                        {decryptError}
                      </p>
                    )}
                  </div>
                  <Button
                    onClick={handleDecrypt}
                    disabled={!passphrase || isDecrypting}
                    className="btn-secrets-create w-full"
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
                    <Label className="secrets-detail-label">Decrypted Value</Label>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setShowDecrypted(!showDecrypted)}
                        className="secrets-toggle-btn"
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
                        className="secrets-toggle-btn"
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
                      className="secrets-decrypted-value"
                    />
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setDecryptedValue(null);
                      setPassphrase("");
                      setDecryptError(null);
                    }}
                    className="btn-secrets-cancel w-full"
                  >
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Lock & Clear
                  </Button>
                </div>
              )}
            </div>

            {/* Metadata */}
            <div className="secrets-detail-grid text-sm">
              <div>
                <Label className="secrets-detail-label">Created</Label>
                <p className="secrets-time-cell mt-1">
                  <Clock className="secrets-time-icon" />
                  {format(new Date(secret.created_at), "MMM d, yyyy HH:mm")}
                </p>
              </div>
              <div>
                <Label className="secrets-detail-label">Last Accessed</Label>
                <p className="secrets-time-cell mt-1">
                  <Clock className="secrets-time-icon" />
                  {secret.last_accessed_at
                    ? formatDistanceToNow(new Date(secret.last_accessed_at), {
                        addSuffix: true,
                      })
                    : "Never"}
                </p>
              </div>
              <div>
                <Label className="secrets-detail-label">Access Count</Label>
                <p className="secrets-detail-value mt-1">{secret.access_count}</p>
              </div>
              <div>
                <Label className="secrets-detail-label">Key Version</Label>
                <p className="secrets-detail-value mt-1">
                  v{secret.encrypted_data.key_version}
                </p>
              </div>
            </div>

            {/* Actions */}
            <div className="secrets-detail-actions">
              <Button
                variant="outline"
                onClick={() => setIsTokenModalOpen(true)}
                className="secrets-action-btn-detail"
              >
                <Key className="h-4 w-4 mr-2" />
                Generate Token
              </Button>
              <Button
                variant="outline"
                onClick={() => setIsVersionHistoryOpen(true)}
                className="secrets-action-btn-detail"
              >
                <History className="h-4 w-4 mr-2" />
                Version History
              </Button>
              <Button
                variant="outline"
                onClick={() => setIsEditModalOpen(true)}
                className="secrets-action-btn-detail"
              >
                <Edit3 className="h-4 w-4 mr-2" />
                Edit
              </Button>
              <Button
                variant="destructive"
                onClick={() => setIsDeleteDialogOpen(true)}
                className="secrets-action-btn-detail"
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
          <DialogContent className="secrets-dialog-content">
            <DialogHeader className="secrets-dialog-header">
              <DialogTitle className="secrets-dialog-title">Edit Secret</DialogTitle>
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
        <AlertDialogContent className="secrets-delete-content">
          <AlertDialogHeader className="secrets-delete-header">
            <AlertDialogTitle className="secrets-delete-title">Delete Secret</AlertDialogTitle>
            <AlertDialogDescription className="secrets-delete-description">
              Are you sure you want to delete <strong>{secret.name}</strong>?
              This action cannot be undone. All access tokens for this secret
              will also be revoked.
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

      {/* Version History Drawer */}
      <SecretVersionHistory
        secretId={secretId}
        isOpen={isVersionHistoryOpen}
        onClose={() => setIsVersionHistoryOpen(false)}
        onRollbackSuccess={() => {
          // Could refetch secret data here if needed
        }}
      />
    </>
  );
}
