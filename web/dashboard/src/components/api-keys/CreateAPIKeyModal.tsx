import { useState } from "react";
import { useNavigate } from "react-router-dom";
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
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [showKey, setShowKey] = useState<string | null>(null);

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError("Name is required");
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
      // Store the new key in localStorage for one-time display
      storeNewApiKey(response);
      setShowKey(response.plaintext);
      onSuccess?.({ ...data, plaintext: response.plaintext });
    } catch (err) {
      console.error("Failed to create API key:", err);
      setError(err instanceof Error ? err.message : "Failed to create API key");
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
    if (showKey) {
      await navigator.clipboard.writeText(showKey);
    }
  };

  if (showKey) {
    return (
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>API Key Created</DialogTitle>
            <DialogDescription>
              Your new API key has been created. Copy it now as it will not be
              shown again.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="bg-muted p-4 rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <Label className="text-muted-foreground">API Key</Label>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={copyToClipboard}
                  className="text-xs"
                >
                  Copy
                </Button>
              </div>
              <code className="text-sm break-all font-mono">{showKey}</code>
            </div>
            <div className="bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800 rounded-lg p-4">
              <p className="text-sm text-amber-800 dark:text-amber-200">
                <strong>Important:</strong> This is the only time you will see
                this key. Store it securely. If you lose it, you will need to
                create a new API key.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={handleClose}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Create API Key</DialogTitle>
          <DialogDescription>
            Create a new API key for programmatic access to the platform.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-3 text-sm text-red-600 dark:text-red-400">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              placeholder="My API Key"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description (optional)</Label>
            <Textarea
              id="description"
              placeholder="Describe what this key is for..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="keyType">Key Type *</Label>
            <Select value={keyType} onValueChange={(v) => setKeyType(v as APIKeyType)}>
              <SelectTrigger>
                <SelectValue placeholder="Select key type" />
              </SelectTrigger>
              <SelectContent>
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
              The type determines what resources this key can access.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="rotationDays">Rotation Frequency (days)</Label>
            <Input
              id="rotationDays"
              type="number"
              min={1}
              max={365}
              value={rotationDays}
              onChange={(e) => setRotationDays(parseInt(e.target.value, 10) || DEFAULT_ROTATION_DAYS)}
            />
            <p className="text-xs text-muted-foreground">
              How often the key should be rotated (1-365 days)
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
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Creating..." : "Create API Key"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
