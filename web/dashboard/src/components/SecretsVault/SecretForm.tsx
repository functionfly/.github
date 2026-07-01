/**
 * SecretForm - Create/edit secret form with client-side encryption
 * Handles passphrase-based key derivation and AES-256-GCM encryption
 */

import { useCreateSecret, useUpdateSecret } from '@/hooks/useVault';
import { cn } from '@/lib/utils';
import type { CreateSecretRequest, SecretMetadata, SecretType } from '@/types/vault';
import { VaultCrypto } from '@/utils/vault-crypto';
import { Eye, EyeOff, FileKey, Key, Loader2, Lock, Shield, Wand2, Copy, Check } from 'lucide-react';
import { useCallback, useState } from 'react';
import { z } from 'zod';

import { PasswordStrengthIndicator } from '@/components/common/PasswordStrengthIndicator';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';

// Form validation schema
const secretFormSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .min(2, 'Name must be at least 2 characters')
    .max(100, 'Name must be less than 100 characters')
    .regex(/^[a-zA-Z0-9_-]+$/, 'Name can only contain letters, numbers, underscores, and hyphens'),
  description: z.string().max(500, 'Description must be less than 500 characters').optional(),
  secret_type: z.enum(['api_key', 'oauth_token', 'password', 'certificate'] as const),
  plaintext: z.string().min(1, 'Secret value is required'),
  passphrase: z
    .string()
    .min(12, 'Passphrase must be at least 12 characters')
    .min(1, 'Passphrase is required')
    .regex(/[A-Z]/, 'Passphrase must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Passphrase must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Passphrase must contain at least one digit'),
  scopes: z.array(z.string()).optional(),
});

type SecretFormData = z.infer<typeof secretFormSchema>;

// Available scopes for secrets
const AVAILABLE_SCOPES = [
  { value: 'read', label: 'Read', description: 'Can read secret value' },
  { value: 'write', label: 'Write', description: 'Can update secret metadata' },
  { value: 'delete', label: 'Delete', description: 'Can delete secret' },
  { value: 'admin', label: 'Admin', description: 'Full access' },
];

// Secret type options
const SECRET_TYPES: { value: SecretType; label: string; icon: typeof Key }[] = [
  { value: 'api_key', label: 'API Key', icon: Key },
  { value: 'oauth_token', label: 'OAuth Token', icon: Shield },
  { value: 'password', label: 'Password', icon: Lock },
  { value: 'certificate', label: 'Certificate', icon: FileKey },
];

// Characters for passphrase generation (no ambiguous chars like 0/O, 1/l/I)
const PASSPHRASE_CHARS = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789';
const PASSPHRASE_SYMBOLS = '!@#$%^&*+-=?';

function generatePassphrase(length: number = 20): string {
  const allChars = PASSPHRASE_CHARS + PASSPHRASE_SYMBOLS;
  const randomValues = crypto.getRandomValues(new Uint8Array(length));
  let result = '';
  for (let i = 0; i < length; i++) {
    result += allChars[randomValues[i] % allChars.length];
  }
  // Ensure at least one of each required class
  const classes = [
    'ABCDEFGHJKLMNPQRSTUVWXYZ',
    'abcdefghjkmnpqrstuvwxyz',
    '23456789',
    PASSPHRASE_SYMBOLS,
  ];
  const rng = crypto.getRandomValues(new Uint8Array(classes.length));
  classes.forEach((chars, i) => {
    const pos = (rng[i] & 0x7f) % result.length;
    const replacement = chars[rng[i] % chars.length];
    result = result.substring(0, pos) + replacement + result.substring(pos + 1);
  });
  return result;
}

export interface SecretFormProps {
  onSubmit: () => void;
  onCancel: () => void;
  initialData?: SecretMetadata;
}

export function SecretForm({ onSubmit, onCancel, initialData }: SecretFormProps) {
  const isEditMode = !!initialData;
  const createSecret = useCreateSecret();
  const updateSecret = useUpdateSecret(initialData?.id || '');

  const [formData, setFormData] = useState<Partial<SecretFormData>>({
    name: initialData?.name || '',
    description: initialData?.description || '',
    secret_type: initialData?.secret_type || 'api_key',
    plaintext: '',
    passphrase: '',
    scopes: initialData?.scopes || ['read'],
  });

  const [errors, setErrors] = useState<Partial<Record<keyof SecretFormData, string>>>({});
  const [showPlaintext, setShowPlaintext] = useState(false);
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [isEncrypting, setIsEncrypting] = useState(false);
  const [passphraseCopied, setPassphraseCopied] = useState(false);

  const handleGeneratePassphrase = useCallback(() => {
    const generated = generatePassphrase(20);
    setFormData((prev) => ({ ...prev, passphrase: generated }));
    setShowPassphrase(true);
    setPassphraseCopied(false);
    setErrors((prev) => ({ ...prev, passphrase: undefined }));
  }, []);

  const handleCopyPassphrase = useCallback(async () => {
    if (!formData.passphrase) return;
    try {
      await navigator.clipboard.writeText(formData.passphrase);
      setPassphraseCopied(true);
      setTimeout(() => setPassphraseCopied(false), 2000);
    } catch {
      // clipboard API may fail in some environments
    }
  }, [formData.passphrase]);

  // Validate form field
  const validateField = useCallback(
    (field: keyof SecretFormData, value: unknown) => {
      const result = secretFormSchema.safeParse({ ...formData, [field]: value });
      if (!result.success) {
        const fieldError = result.error.issues.find((e) => e.path[0] === field);
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
      result.error.issues.forEach((err) => {
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

        const encrypted = await VaultCrypto.encryptWithPassphrase(data.plaintext, data.passphrase);

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
      console.error('Failed to save secret:', error);
    } finally {
      setIsEncrypting(false);
    }
  };

  const isSubmitting = createSecret.isPending || updateSecret.isPending || isEncrypting;

  return (
    <form onSubmit={handleSubmit} className="secrets-form space-y-6">
      {/* Name field */}
      <div className="space-y-2">
        <Label htmlFor="name" className="secrets-form-label">
          Name <span className="text-[var(--status-revoked)]">*</span>
        </Label>
        <Input
          id="name"
          placeholder="my-api-key"
          value={formData.name}
          onChange={(e) => handleChange('name', e.target.value)}
          disabled={isEditMode} // Name cannot be changed in edit mode
          className={cn("secrets-search-input", errors.name && 'border-[var(--status-revoked)]')}
        />
        {errors.name && <p className="secrets-form-error">{errors.name}</p>}
        <p className="secrets-form-hint">
          Unique identifier for this secret. Cannot be changed after creation.
        </p>
      </div>

      {/* Secret Type field */}
      <div className="space-y-2">
        <Label htmlFor="secret_type" className="secrets-form-label">
          Secret Type <span className="text-[var(--status-revoked)]">*</span>
        </Label>
        <Select
          value={formData.secret_type}
          onValueChange={(value) => handleChange('secret_type', value)}
          disabled={isEditMode} // Type cannot be changed in edit mode
        >
          <SelectTrigger className={cn("secrets-search-input", errors.secret_type && 'border-[var(--status-revoked)]')}>
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
        <Label htmlFor="description" className="secrets-form-label">Description</Label>
        <Textarea
          id="description"
          placeholder="What is this secret used for?"
          value={formData.description}
          onChange={(e) => handleChange('description', e.target.value)}
          className={cn("secrets-search-input", errors.description && 'border-[var(--status-revoked)]')}
          rows={3}
        />
        {errors.description && <p className="secrets-form-error">{errors.description}</p>}
      </div>

      {/* Scopes field */}
      <div className="space-y-2">
        <Label className="secrets-form-label">Scopes</Label>
        <div className="flex flex-wrap gap-2">
          {AVAILABLE_SCOPES.map((scope) => (
            <button
              key={scope.value}
              type="button"
              onClick={() => toggleScope(scope.value)}
              className={cn(
                'secrets-scope-btn',
                formData.scopes?.includes(scope.value)
                  ? 'secrets-scope-btn-active'
                  : ''
              )}
            >
              {scope.label}
              {formData.scopes?.includes(scope.value) && <span className="text-xs">✓</span>}
            </button>
          ))}
        </div>
        <p className="secrets-form-hint">Select the access levels for this secret</p>
      </div>

      {/* Plaintext secret field (only for create mode) */}
      {!isEditMode && (
        <>
          <div className="space-y-2">
            <Label htmlFor="plaintext" className="secrets-form-label">
              Secret Value <span className="text-[var(--status-revoked)]">*</span>
            </Label>
            <div className="relative">
              {showPlaintext ? (
                <Textarea
                  id="plaintext"
                  placeholder="Enter the secret value to encrypt..."
                  value={formData.plaintext}
                  onChange={(e) => handleChange('plaintext', e.target.value)}
                  className={cn("secrets-search-input font-mono text-sm", errors.plaintext && 'border-[var(--status-revoked)]')}
                  rows={4}
                />
              ) : (
                <Input
                  id="plaintext-hidden"
                  type="password"
                  placeholder="Enter the secret value to encrypt..."
                  value={formData.plaintext}
                  onChange={(e) => handleChange('plaintext', e.target.value)}
                  className={cn("secrets-search-input font-mono text-sm h-24", errors.plaintext && 'border-[var(--status-revoked)]')}
                />
              )}
              <button
                type="button"
                onClick={() => setShowPlaintext(!showPlaintext)}
                className="absolute right-3 top-3 text-[var(--text-faint)] hover:text-[var(--text)] secrets-toggle-btn"
              >
                {showPlaintext ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
            {errors.plaintext && <p className="secrets-form-error">{errors.plaintext}</p>}
          </div>

          {/* Passphrase field */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="passphrase" className="secrets-form-label">
                Encryption Passphrase <span className="text-[var(--status-revoked)]">*</span>
              </Label>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={handleGeneratePassphrase}
                  className="secrets-generate-btn"
                >
                  <Wand2 className="h-3 w-3" />
                  Generate
                </button>
                {formData.passphrase && (
                  <button
                    type="button"
                    onClick={handleCopyPassphrase}
                    className="secrets-copy-btn ml-2"
                  >
                    {passphraseCopied ? (
                      <>
                        <Check className="h-3 w-3 text-[var(--status-ok)]" />
                        Copied
                      </>
                    ) : (
                      <>
                        <Copy className="h-3 w-3" />
                        Copy
                      </>
                    )}
                  </button>
                )}
              </div>
            </div>
            <div className="relative">
              <Input
                id="passphrase"
                type={showPassphrase ? 'text' : 'password'}
                placeholder="Enter a strong passphrase to encrypt this secret"
                value={formData.passphrase}
                onChange={(e) => handleChange('passphrase', e.target.value)}
                className={cn("secrets-search-input pr-10", errors.passphrase && 'border-[var(--status-revoked)]')}
              />
              <button
                type="button"
                onClick={() => setShowPassphrase(!showPassphrase)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--text-faint)] hover:text-[var(--text)] secrets-toggle-btn"
              >
                {showPassphrase ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
            {formData.passphrase && <PasswordStrengthIndicator password={formData.passphrase} />}
            {errors.passphrase && <p className="secrets-form-error">{errors.passphrase}</p>}
            <div className="secrets-warning-box">
              <p className="text-sm text-[var(--status-pending)]">
                <strong>Important:</strong> This passphrase will be used to encrypt your secret. You
                will need this same passphrase to decrypt the secret later. Store it securely - we
                cannot recover it if you lose it!
              </p>
            </div>
          </div>
        </>
      )}

      {/* Actions */}
      <div className="secrets-form-actions">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting} className="btn-secrets-cancel">
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting} className="btn-secrets-create">
          {isSubmitting ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              {isEncrypting ? 'Encrypting...' : isEditMode ? 'Updating...' : 'Creating...'}
            </>
          ) : isEditMode ? (
            'Update Secret'
          ) : (
            'Create Secret'
          )}
        </Button>
      </div>
    </form>
  );
}
