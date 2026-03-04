/**
 * SecretForm - Create/edit secret form with client-side encryption
 * Handles passphrase-based key derivation and AES-256-GCM encryption
 */

import { useState, useCallback } from "react";
import { z } from "zod";
import { Eye, EyeOff, Lock, Shield, Key, FileKey, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { VaultCrypto } from "@/utils/vault-crypto";
import { useCreateSecret, useUpdateSecret } from "@/hooks/useVault";
import { evaluatePasswordStrength } from "@/lib/validation";
import type { SecretMetadata, SecretType, CreateSecretRequest } from "@/types/vault";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { PasswordStrengthIndicator } from "@/components/common/PasswordStrengthIndicator";

// Form validation schema
const secretFormSchema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .min(2, "Name must be at least 2 characters")
    .max(100, "Name must be less than 100 characters")
    .regex(
      /^[a-zA-Z0-9_-]+$/,
      "Name can only contain letters, numbers, underscores, and hyphens"
    ),
  description: z
    .string()
    .max(500, "Description must be less than 500 characters")
    .optional(),
  secret_type: z.enum(["api_key", "oauth_token", "password", "certificate"] as const),
  plaintext: z.string().min(1, "Secret value is required"),
  passphrase: z
    .string()
    .min(8, "Passphrase must be at least 8 characters")
    .min(1, "Passphrase is required"),
  scopes: z.array(z.string()).optional(),
});

type SecretFormData = z.infer<typeof secretFormSchema>;

// Available scopes for secrets
const AVAILABLE_SCOPES = [
  { value: "read", label: "Read", description: "Can read secret value" },
  { value: "write", label: "Write", description: "Can update secret metadata" },
  { value: "delete", label: "Delete", description: "Can delete secret" },
  { value: "admin", label: "Admin", description: "Full access" },
];

// Secret type options
const SECRET_TYPES: { value: SecretType; label: string; icon: typeof Key }[] = [
  { value: "api_key", label: "API Key", icon: Key },
  { value: "oauth_token", label: "OAuth Token", icon: Shield },
  { value: "password", label: "Password", icon: Lock },
  { value: "certificate", label: "Certificate", icon: FileKey },
];

export interface SecretFormProps {
  onSubmit: () => void;
  onCancel: () => void;
  initialData?: SecretMetadata;
}

export function SecretForm({ onSubmit, onCancel, initialData }: SecretFormProps) {
  const isEditMode = !!initialData;
  const createSecret = useCreateSecret();
  const updateSecret = useUpdateSecret(initialData?.id || "");

  const [formData, setFormData] = useState<Partial<SecretFormData>>({
    name: initialData?.name || "",
    description: initialData?.description || "",
    secret_type: initialData?.secret_type || "api_key",
    plaintext: "",
    passphrase: "",
    scopes: initialData?.scopes || ["read"],
  });

  const [errors, setErrors] = useState<Partial<Record<keyof SecretFormData, string>>>({});
  const [showPlaintext, setShowPlaintext] = useState(false);
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [isEncrypting, setIsEncrypting] = useState(false);

  // Validate form field
  const validateField = useCallback(
    (field: keyof SecretFormData, value: unknown) => {
      const result = secretFormSchema.safeParse({ ...formData, [field]: value });
      if (!result.success) {
        const fieldError = result.error.errors.find((e) => e.path[0] === field);
        setErrors((prev) => ({
          ...prev,
          [field]: fieldError?.message,
        }));
        return false;
      }
      setErrors((prev) => ({ ...prev, [field]: undefined }));
      return true;
    },
    [formData]
  );

  // Handle input change
  const handleChange = useCallback(
    (field: keyof SecretFormData, value: unknown) => {
      setFormData((prev) => ({ ...prev, [field]: value }));
      validateField(field, value);
    },
    [validateField]
  );

  // Toggle scope selection
  const toggleScope = useCallback((scope: string) => {
    setFormData((prev) => {
      const currentScopes = prev.scopes || [];
      const newScopes = currentScopes.includes(scope)
        ? currentScopes.filter((s) => s !== scope)
        : [...currentScopes, scope];
      return { ...prev, scopes: newScopes };
    });
  }, []);

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate all fields
    const result = secretFormSchema.safeParse(formData);
    if (!result.success) {
      const newErrors: Partial<Record<keyof SecretFormData, string>> = {};
      result.error.errors.forEach((err) => {
        const field = err.path[0] as keyof SecretFormData;
        newErrors[field] = err.message;
      });
      setErrors(newErrors);
      return;
    }

    const data = result.data;

    try {
      if (isEditMode) {
        // For edit mode, only update metadata (not the encrypted value)
        await updateSecret.mutateAsync({
          name: data.name,
          description: data.description,
          scopes: data.scopes,
        });
      } else {
        // For create mode, encrypt and submit
        setIsEncrypting(true);

        const encrypted = await VaultCrypto.encryptWithPassphrase(
          data.plaintext,
          data.passphrase
        );

        const request: CreateSecretRequest = {
          name: data.name,
          description: data.description,
          secret_type: data.secret_type,
          encrypted_data: VaultCrypto.toPayload(encrypted),
          scopes: data.scopes,
        };

        await createSecret.mutateAsync(request);
      }

      onSubmit();
    } catch (error) {
      console.error("Failed to save secret:", error);
    } finally {
      setIsEncrypting(false);
    }
  };

  const isSubmitting = createSecret.isPending || updateSecret.isPending || isEncrypting;

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Name field */}
      <div className="space-y-2">
        <Label htmlFor="name">
          Name <span className="text-error">*</span>
        </Label>
        <Input
          id="name"
          placeholder="my-api-key"
          value={formData.name}
          onChange={(e) => handleChange("name", e.target.value)}
          disabled={isEditMode} // Name cannot be changed in edit mode
          className={cn(errors.name && "border-error")}
        />
        {errors.name && (
          <p className="text-sm text-error">{errors.name}</p>
        )}
        <p className="text-xs text-text-muted">
          Unique identifier for this secret. Cannot be changed after creation.
        </p>
      </div>

      {/* Secret Type field */}
      <div className="space-y-2">
        <Label htmlFor="secret_type">
          Secret Type <span className="text-error">*</span>
        </Label>
        <Select
          value={formData.secret_type}
          onValueChange={(value) => handleChange("secret_type", value)}
          disabled={isEditMode} // Type cannot be changed in edit mode
        >
          <SelectTrigger className={cn(errors.secret_type && "border-error")}>
            <SelectValue placeholder="Select type" />
          </SelectTrigger>
          <SelectContent>
            {SECRET_TYPES.map((type) => {
              const Icon = type.icon;
              return (
                <SelectItem key={type.value} value={type.value}>
                  <div className="flex items-center gap-2">
                    <Icon className="h-4 w-4" />
                    {type.label}
                  </div>
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
      </div>

      {/* Description field */}
      <div className="space-y-2">
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          placeholder="What is this secret used for?"
          value={formData.description}
          onChange={(e) => handleChange("description", e.target.value)}
          className={cn(errors.description && "border-error")}
          rows={3}
        />
        {errors.description && (
          <p className="text-sm text-error">{errors.description}</p>
        )}
      </div>

      {/* Scopes field */}
      <div className="space-y-2">
        <Label>Scopes</Label>
        <div className="flex flex-wrap gap-2">
          {AVAILABLE_SCOPES.map((scope) => (
            <button
              key={scope.value}
              type="button"
              onClick={() => toggleScope(scope.value)}
              className={cn(
                "inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-all duration-200",
                formData.scopes?.includes(scope.value)
                  ? "bg-brand-500/20 text-brand-600 border border-brand-500/30"
                  : "bg-bg-tertiary text-text-muted border border-border-subtle hover:border-border-default"
              )}
            >
              {scope.label}
              {formData.scopes?.includes(scope.value) && (
                <span className="text-xs">✓</span>
              )}
            </button>
          ))}
        </div>
        <p className="text-xs text-text-muted">
          Select the access levels for this secret
        </p>
      </div>

      {/* Plaintext secret field (only for create mode) */}
      {!isEditMode && (
        <>
          <div className="space-y-2">
            <Label htmlFor="plaintext">
              Secret Value <span className="text-error">*</span>
            </Label>
            <div className="relative">
              {showPlaintext ? (
                <Textarea
                  id="plaintext"
                  placeholder="Enter the secret value to encrypt..."
                  value={formData.plaintext}
                  onChange={(e) => handleChange("plaintext", e.target.value)}
                  className={cn(
                    errors.plaintext && "border-error",
                    "pr-10 font-mono text-sm"
                  )}
                  rows={4}
                />
              ) : (
                <Input
                  id="plaintext-hidden"
                  type="password"
                  placeholder="Enter the secret value to encrypt..."
                  value={formData.plaintext}
                  onChange={(e) => handleChange("plaintext", e.target.value)}
                  className={cn(
                    errors.plaintext && "border-error",
                    "pr-10 font-mono text-sm h-24"
                  )}
                />
              )}
              <button
                type="button"
                onClick={() => setShowPlaintext(!showPlaintext)}
                className="absolute right-3 top-3 text-text-muted hover:text-text-primary"
              >
                {showPlaintext ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </button>
            </div>
            {errors.plaintext && (
              <p className="text-sm text-error">{errors.plaintext}</p>
            )}
          </div>

          {/* Passphrase field */}
          <div className="space-y-2">
            <Label htmlFor="passphrase">
              Encryption Passphrase <span className="text-error">*</span>
            </Label>
            <div className="relative">
              <Input
                id="passphrase"
                type={showPassphrase ? "text" : "password"}
                placeholder="Enter a strong passphrase to encrypt this secret"
                value={formData.passphrase}
                onChange={(e) => handleChange("passphrase", e.target.value)}
                className={cn(errors.passphrase && "border-error", "pr-10")}
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
            {formData.passphrase && (
              <PasswordStrengthIndicator password={formData.passphrase} />
            )}
            {errors.passphrase && (
              <p className="text-sm text-error">{errors.passphrase}</p>
            )}
            <div className="rounded-lg bg-warning-glow border border-warning/20 p-3">
              <p className="text-sm text-warning">
                <strong>Important:</strong> This passphrase will be used to encrypt your secret.
                You will need this same passphrase to decrypt the secret later.
                Store it securely - we cannot recover it if you lose it!
              </p>
            </div>
          </div>
        </>
      )}

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-4 border-t border-border-subtle">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isSubmitting}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              {isEncrypting ? "Encrypting..." : isEditMode ? "Updating..." : "Creating..."}
            </>
          ) : isEditMode ? (
            "Update Secret"
          ) : (
            "Create Secret"
          )}
        </Button>
      </div>
    </form>
  );
}
