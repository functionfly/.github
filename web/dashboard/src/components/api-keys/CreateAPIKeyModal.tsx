import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  APIKeyType,
  CreateAPIKeyRequest,
  API_KEY_TYPE_LABELS,
  API_KEY_TYPE_DESCRIPTIONS,
  DEFAULT_RATE_LIMIT,
  DEFAULT_ROTATION_DAYS,
} from "@/types/api-key";
import { apiKeysService, storeNewApiKey } from "@/services/api-keys";
import {
  storeApiKeyInVault,
  isVaultPassphraseSet,
  getVaultPassphrase,
  markVaultApiKeyStored,
} from "@/services/vault-api-key-storage";
import { toast } from "sonner";
import { usePlan } from "@/hooks";

interface CreateAPIKeyModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (apiKey: CreateAPIKeyRequest & { plaintext: string }) => void;
}

export function CreateAPIKeyModal({
  open,
  onOpenChange,
  onSuccess,
}: CreateAPIKeyModalProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isEnterprise } = usePlan();
  const [isLoading, setIsLoading] = useState(false);
  const [showKey, setShowKey] = useState<string | null>(null);
  const [countdownSeconds, setCountdownSeconds] = useState(20);
  const [copied, setCopied] = useState(false);

  // Form state
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [keyType, setKeyType] = useState<APIKeyType>("platform");
  const [rotationDays, setRotationDays] = useState(DEFAULT_ROTATION_DAYS);
  const [rpm, setRpm] = useState(DEFAULT_RATE_LIMIT.rpm);
  const [rph, setRph] = useState(DEFAULT_RATE_LIMIT.rph);
  const [rpd, setRpd] = useState(DEFAULT_RATE_LIMIT.rpd);
  const [error, setError] = useState<string | null>(null);
  const [saveToVault, setSaveToVault] = useState(false);
  const [vaultSaveStatus, setVaultSaveStatus] = useState<'idle' | 'saving' | 'saved' | 'failed'>('idle');
  const [hasVaultPassphrase, setHasVaultPassphrase] = useState(false);

  const resetForm = () => {
    setName("");
    setDescription("");
    setKeyType("platform");
    setRotationDays(DEFAULT_ROTATION_DAYS);
    setRpm(DEFAULT_RATE_LIMIT.rpm);
    setRph(DEFAULT_RATE_LIMIT.rph);
    setRpd(DEFAULT_RATE_LIMIT.rpd);
    setError(null);
    setVaultSaveStatus('idle');
  };

  useEffect(() => {
    const hasVault = isVaultPassphraseSet();
    setHasVaultPassphrase(hasVault);
    setSaveToVault(hasVault);
  }, [open]);

  // Live countdown and auto-close when showing the new key
  useEffect(() => {
    if (!showKey || !open) return;
    setCountdownSeconds(20);
    const interval = setInterval(() => {
      setCountdownSeconds((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          setShowKey(null);
          resetForm();
          onOpenChange(false);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [showKey, open, onOpenChange]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError(t("createApiKey.nameRequired"));
      return;
    }

    setIsLoading(true);
    try {
      // Sanitize numeric inputs: parseInt("-5", 10) === -5 (truthy), so we
      // explicitly clamp to >= 0 and fall back to the default.
      const safeInt = (val: number, fallback: number): number => {
        if (!Number.isFinite(val) || val < 0) return fallback;
        return Math.floor(val);
      };

      const data: CreateAPIKeyRequest = {
        name: name.trim(),
        description: description.trim() || undefined,
        key_type: keyType,
        rotation_frequency_days: safeInt(rotationDays, DEFAULT_ROTATION_DAYS),
        // Send rate_limit as a nested object — this matches the backend
        // apikey.CreateAPIKeyRequest.RateLimit field. Sending flat fields
        // would be silently dropped and the default rate limit would apply.
        rate_limit: {
          rpm: safeInt(rpm, DEFAULT_RATE_LIMIT.rpm),
          rph: safeInt(rph, DEFAULT_RATE_LIMIT.rph),
          rpd: safeInt(rpd, DEFAULT_RATE_LIMIT.rpd),
        },
      };

      const response = await apiKeysService.createKey(data);
      storeNewApiKey(response);
      const plaintext = response.plaintext ?? "";
      setCopied(false);
      setShowKey(plaintext || null);

      if (saveToVault && plaintext && hasVaultPassphrase) {
        setVaultSaveStatus('saving');
        try {
          const passphrase = await getVaultPassphrase();
          if (passphrase) {
            await storeApiKeyInVault(response, passphrase);
            markVaultApiKeyStored(response.id);
            setVaultSaveStatus('saved');
            toast.success('API key saved to vault', {
              description: 'Your key is now encrypted and stored securely',
            });
          } else {
            setVaultSaveStatus('failed');
          }
        } catch (vaultErr) {
          // Sanitized log: do not print the full error object as it may
          // contain request bodies or partial key fragments.
          // eslint-disable-next-line no-console
          console.error('Failed to save API key to vault');
          setVaultSaveStatus('failed');
          toast.error('Failed to save to vault', {
            description: 'Your API key was created but could not be saved to vault',
          });
        }
      }

      onSuccess?.({ ...data, plaintext });
      if (!plaintext) {
        resetForm();
        onOpenChange(false);
      }
    } catch (err: unknown) {
      // Sanitized log: extract message only, do not log the full error.
      const safeMessage =
        err && typeof err === "object" && "response" in err
          ? (err as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message
          : err instanceof Error
            ? err.message
            : String(err);
      // eslint-disable-next-line no-console
      console.error("Failed to create API key:", safeMessage);
      setError(
        typeof safeMessage === "string" && safeMessage
          ? safeMessage
          : t("createApiKey.failedToCreate")
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    resetForm();
    setShowKey(null);
    onOpenChange(false);
    if (showKey) {
      // Navigate to the keys list after showing the key
      navigate("/dashboard/api-keys");
    }
  };

  const copyToClipboard = async () => {
    if (!showKey) return;
    try {
      await navigator.clipboard.writeText(showKey);
      setCopied(true);
      toast.success(t("createApiKey.apiKeyCopied"));
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error(t("createApiKey.failedToCopy"));
    }
  };

  if (showKey) {
    return (
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>{t("createApiKey.keyCreatedTitle")}</DialogTitle>
            <DialogDescription>
              {t("createApiKey.keyCreatedDescription")}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="bg-muted p-4 rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <Label className="text-muted-foreground">{t("createApiKey.apiKey")}</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={copyToClipboard}
                  className="text-xs"
                  disabled={copied}
                >
                  {copied ? t("createApiKey.copied") : t("createApiKey.copy")}
                </Button>
              </div>
              <code className="text-sm break-all font-mono">{showKey}</code>
            </div>
            <div className="bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
              <p className="text-sm text-amber-800 dark:text-amber-200">
                <strong>{t("createApiKey.important")}</strong> {t("createApiKey.importantDescription")}
              </p>
            </div>
          </div>
          <DialogFooter>
              <p className="text-xs text-muted-foreground mr-auto">
                {t("createApiKey.closesAutomatically", { countdownSeconds })}
              </p>
            <Button onClick={handleClose}>{t("createApiKey.done")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{t("createApiKey.createTitle")}</DialogTitle>
          <DialogDescription>
            {t("createApiKey.createDescription")}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-3 text-sm text-red-600 dark:text-red-400">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="name">{t("createApiKey.name")}</Label>
            <Input
              id="name"
              placeholder={t("createApiKey.namePlaceholder")}
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">{t("createApiKey.description")}</Label>
            <Textarea
              id="description"
              placeholder={t("createApiKey.descriptionPlaceholder")}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="keyType">{t("createApiKey.keyType")}</Label>
            <Select value={keyType} onValueChange={(v) => setKeyType(v as APIKeyType)}>
              <SelectTrigger className="create-api-key-key-type-trigger">
                <SelectValue placeholder={t("createApiKey.selectKeyType")} />
              </SelectTrigger>
              <SelectContent className="create-api-key-key-type-dropdown">
                {(Object.keys(API_KEY_TYPE_LABELS) as APIKeyType[]).filter((type) => {
                  // MicroPython runtime keys are Enterprise-only
                  if (type === 'micropython' && !isEnterprise) return false;
                  return true;
                }).map((type) => (
                  <SelectItem key={type} value={type}>
                    <div className="flex flex-col">
                      <span>{API_KEY_TYPE_LABELS[type]}</span>
                      <span className="text-xs text-muted-foreground">
                        {API_KEY_TYPE_DESCRIPTIONS[type]}
                      </span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {t("createApiKey.keyTypeHelp")}
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="rotationDays">{t("createApiKey.rotationFrequency")}</Label>
            <Input
              id="rotationDays"
              type="number"
              min={1}
              max={365}
              value={rotationDays}
              onChange={(e) => {
                const n = parseInt(e.target.value, 10);
                // Guard against NaN, negative numbers, and oversized values.
                setRotationDays(
                  Number.isFinite(n) && n >= 1 && n <= 365 ? n : DEFAULT_ROTATION_DAYS
                );
              }}
            />
            <p className="text-xs text-muted-foreground">
              {t("createApiKey.rotationHelp")}
            </p>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="rpm">RPM</Label>
              <Input
                id="rpm"
                type="number"
                min={1}
                value={rpm}
                onChange={(e) => {
                  const n = parseInt(e.target.value, 10);
                  setRpm(
                    Number.isFinite(n) && n >= 1 ? n : DEFAULT_RATE_LIMIT.rpm
                  );
                }}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="rph">RPH</Label>
              <Input
                id="rph"
                type="number"
                min={1}
                value={rph}
                onChange={(e) => {
                  const n = parseInt(e.target.value, 10);
                  setRph(
                    Number.isFinite(n) && n >= 1 ? n : DEFAULT_RATE_LIMIT.rph
                  );
                }}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="rpd">RPD</Label>
              <Input
                id="rpd"
                type="number"
                min={1}
                value={rpd}
                onChange={(e) => {
                  const n = parseInt(e.target.value, 10);
                  setRpd(
                    Number.isFinite(n) && n >= 1 ? n : DEFAULT_RATE_LIMIT.rpd
                  );
                }}
              />
            </div>
          </div>

          {hasVaultPassphrase ? (
            <div className="flex items-start space-x-3 p-4 rounded-lg bg-muted/50 border border-border-subtle">
              <input
                type="checkbox"
                id="saveToVault"
                checked={saveToVault}
                onChange={(e) => setSaveToVault(e.target.checked)}
                className="mt-1 h-4 w-4 rounded border-border text-brand-500 focus:ring-brand-500"
              />
              <div className="flex-1">
                <Label htmlFor="saveToVault" className="text-sm font-medium cursor-pointer">
                  Encrypt and store in Vault
                </Label>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Your API key will be encrypted client-side and stored securely in your vault.
                  {vaultSaveStatus === 'saved' && (
                    <span className="text-green-600 ml-1">Saved</span>
                  )}
                  {vaultSaveStatus === 'saving' && (
                    <span className="text-amber-600 ml-1">Saving...</span>
                  )}
                  {vaultSaveStatus === 'failed' && (
                    <span className="text-red-600 ml-1">Failed - key created but not stored in vault</span>
                  )}
                </p>
              </div>
            </div>
          ) : (
            <div className="p-4 rounded-lg bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800">
              <p className="text-sm text-amber-800 dark:text-amber-200">
                <strong>Want to secure your API keys?</strong>{" "}
                <button
                  type="button"
                  onClick={() => {
                    onOpenChange(false);
                    window.dispatchEvent(new CustomEvent('openVaultSetup'));
                  }}
                  className="underline hover:no-underline"
                >
                  Set up Vault
                </button>{" "}
                to encrypt and store your keys securely.
              </p>
            </div>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t("createApiKey.cancel")}
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? t("createApiKey.creating") : t("createApiKey.create")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
