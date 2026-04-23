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
import { toast } from "sonner";

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

  const resetForm = () => {
    setName("");
    setDescription("");
    setKeyType("platform");
    setRotationDays(DEFAULT_ROTATION_DAYS);
    setRpm(DEFAULT_RATE_LIMIT.rpm);
    setRph(DEFAULT_RATE_LIMIT.rph);
    setRpd(DEFAULT_RATE_LIMIT.rpd);
    setError(null);
  };

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
      const data: CreateAPIKeyRequest = {
        name: name.trim(),
        description: description.trim() || undefined,
        key_type: keyType,
        rotation_frequency_days: rotationDays,
        rate_limit_rpm: rpm,
        rate_limit_rph: rph,
        rate_limit_rpd: rpd,
      };

      const response = await apiKeysService.createKey(data);
      storeNewApiKey(response);
      const plaintext = response.plaintext ?? "";
      setCopied(false);
      setShowKey(plaintext || null);
      onSuccess?.({ ...data, plaintext });
      // If no plaintext in response, close modal so user isn't stuck (key still created, list will update)
      if (!plaintext) {
        resetForm();
        onOpenChange(false);
      }
    } catch (err: unknown) {
      console.error("Failed to create API key:", err);
      const apiMessage =
        err &&
        typeof err === "object" &&
        "response" in err &&
        (err as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message;
      setError(
        typeof apiMessage === "string" && apiMessage
          ? apiMessage
          : err instanceof Error
            ? err.message
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
                {(Object.keys(API_KEY_TYPE_LABELS) as APIKeyType[]).map((type) => (
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
              onChange={(e) => setRotationDays(parseInt(e.target.value, 10) || DEFAULT_ROTATION_DAYS)}
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
                onChange={(e) => setRpm(parseInt(e.target.value, 10) || DEFAULT_RATE_LIMIT.rpm)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="rph">RPH</Label>
              <Input
                id="rph"
                type="number"
                min={1}
                value={rph}
                onChange={(e) => setRph(parseInt(e.target.value, 10) || DEFAULT_RATE_LIMIT.rph)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="rpd">RPD</Label>
              <Input
                id="rpd"
                type="number"
                min={1}
                value={rpd}
                onChange={(e) => setRpd(parseInt(e.target.value, 10) || DEFAULT_RATE_LIMIT.rpd)}
              />
            </div>
          </div>

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
