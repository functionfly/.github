import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
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
import { Switch } from "@/components/ui/switch";
import { AlertCircle } from "lucide-react";
import {
  APIKey,
  API_KEY_TYPE_LABELS,
  API_KEY_TYPE_DESCRIPTIONS,
  DEFAULT_RATE_LIMIT,
  DEFAULT_ROTATION_DAYS,
} from "@/types/api-key";
import { apiKeysService } from "@/services/api-keys";

interface EditAPIKeyModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  apiKey: APIKey | null;
}

export function EditAPIKeyModal({ open, onOpenChange, apiKey }: EditAPIKeyModalProps) {
  const queryClient = useQueryClient();
  const [name, setName] = useState(apiKey?.name ?? "");
  const [description, setDescription] = useState(apiKey?.description ?? "");
  const [keyType, setKeyType] = useState<APIKey["key_type"]>(apiKey?.key_type ?? "platform");
  const [rotationDays, setRotationDays] = useState(apiKey?.rotation_frequency_days ?? DEFAULT_ROTATION_DAYS);
  const [rpm, setRpm] = useState(apiKey?.rate_limit_rpm ?? DEFAULT_RATE_LIMIT.rpm);
  const [rph, setRph] = useState(apiKey?.rate_limit_rph ?? DEFAULT_RATE_LIMIT.rph);
  const [rpd, setRpd] = useState(apiKey?.rate_limit_rpd ?? DEFAULT_RATE_LIMIT.rpd);
  const [isActive, setIsActive] = useState(apiKey?.is_active ?? true);
  const [error, setError] = useState<string | null>(null);

  // Re-initialise form when the underlying key changes.
  const initFromKey = (k: APIKey | null) => {
    if (!k) return;
    setName(k.name);
    setDescription(k.description ?? "");
    setKeyType(k.key_type);
    setRotationDays(k.rotation_frequency_days);
    setRpm(k.rate_limit_rpm ?? DEFAULT_RATE_LIMIT.rpm);
    setRph(k.rate_limit_rph ?? DEFAULT_RATE_LIMIT.rph);
    setRpd(k.rate_limit_rpd ?? DEFAULT_RATE_LIMIT.rpd);
    setIsActive(k.is_active);
    setError(null);
  };

  // Refresh when a different key is opened.
  if (open && apiKey) initFromKey(apiKey);

  const updateMutation = useMutation({
    mutationFn: () =>
      apiKeysService.updateKey(apiKey!.id, {
        name: name.trim() || undefined,
        description: description.trim() || undefined,
        is_active: isActive,
        rotation_frequency_days: Math.max(0, rotationDays),
        rate_limit_rpm: Math.max(0, rpm),
        rate_limit_rph: Math.max(0, rph),
        rate_limit_rpd: Math.max(0, rpd),
      }),
    onSuccess: (updated) => {
      queryClient.invalidateQueries({ queryKey: ["api-key", apiKey!.id] });
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      toast.success("API key updated");
      onOpenChange(false);
    },
    onError: (err: unknown) => {
      const msg =
        err instanceof Error ? err.message : "Failed to update API key";
      setError(msg);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!apiKey) return;
    if (!name.trim()) {
      setError("Name is required");
      return;
    }
    updateMutation.mutate();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[540px]">
        <DialogHeader>
          <DialogTitle>Edit API Key</DialogTitle>
          <DialogDescription>
            Update the key name, type, rate limits, and expiry settings.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="edit-name">Name</Label>
            <Input
              id="edit-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              maxLength={255}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-desc">Description</Label>
            <Textarea
              id="edit-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={2000}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-type">Type</Label>
            <Select value={keyType} onValueChange={(v) => setKeyType(v as APIKey["key_type"])}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {(Object.keys(API_KEY_TYPE_LABELS) as APIKey["key_type"][]).map((t) => (
                  <SelectItem key={t} value={t}>
                    <div className="flex flex-col">
                      <span>{API_KEY_TYPE_LABELS[t]}</span>
                      <span className="text-xs text-muted-foreground">
                        {API_KEY_TYPE_DESCRIPTIONS[t]}
                      </span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-rotation">Rotation Frequency (days)</Label>
            <Input
              id="edit-rotation"
              type="number"
              min={1}
              max={3650}
              value={rotationDays}
              onChange={(e) => {
                const n = parseInt(e.target.value, 10);
                setRotationDays(Number.isFinite(n) && n >= 1 ? n : DEFAULT_ROTATION_DAYS);
              }}
            />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-rpm">RPM</Label>
              <Input
                id="edit-rpm"
                type="number"
                min={1}
                value={rpm}
                onChange={(e) => {
                  const n = parseInt(e.target.value, 10);
                  setRpm(Number.isFinite(n) && n >= 1 ? n : DEFAULT_RATE_LIMIT.rpm);
                }}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-rph">RPH</Label>
              <Input
                id="edit-rph"
                type="number"
                min={1}
                value={rph}
                onChange={(e) => {
                  const n = parseInt(e.target.value, 10);
                  setRph(Number.isFinite(n) && n >= 1 ? n : DEFAULT_RATE_LIMIT.rph);
                }}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-rpd">RPD</Label>
              <Input
                id="edit-rpd"
                type="number"
                min={1}
                value={rpd}
                onChange={(e) => {
                  const n = parseInt(e.target.value, 10);
                  setRpd(Number.isFinite(n) && n >= 1 ? n : DEFAULT_RATE_LIMIT.rpd);
                }}
              />
            </div>
          </div>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label>Active</Label>
              <p className="text-xs text-muted-foreground">
                Inactive keys cannot be used for authentication.
              </p>
            </div>
            <Switch checked={isActive} onCheckedChange={setIsActive} />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
