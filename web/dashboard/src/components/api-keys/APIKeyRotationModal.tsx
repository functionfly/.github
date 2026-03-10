import { useState } from "react";
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
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  APIKey,
  RotationReason,
  ROTATION_REASON_LABELS,
} from "@/types/api-key";
import { apiKeysService } from "@/services/api-keys";

interface APIKeyRotationModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  apiKey: APIKey | null;
  onSuccess?: (newKey: string) => void;
}

export function APIKeyRotationModal({
  open,
  onOpenChange,
  apiKey,
  onSuccess,
}: APIKeyRotationModalProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [showKey, setShowKey] = useState<string | null>(null);
  const [reason, setReason] = useState<RotationReason>("manual");
  const [error, setError] = useState<string | null>(null);

  const handleRotate = async () => {
    if (!apiKey) return;

    setError(null);
    setIsLoading(true);
    try {
      const response = await apiKeysService.rotateKey(apiKey.id, { reason });
      setShowKey(response.plaintext);
      onSuccess?.(response.plaintext);
    } catch (err) {
      console.error("Failed to rotate API key:", err);
      setError(err instanceof Error ? err.message : "Failed to rotate API key");
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setShowKey(null);
    setReason("manual");
    setError(null);
    onOpenChange(false);
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
            <DialogTitle>API Key Rotated</DialogTitle>
            <DialogDescription>
              Your API key has been rotated. Copy the new key now as it will not
              be shown again.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="bg-muted p-4 rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <Label className="text-muted-foreground">New API Key</Label>
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
                <strong>Important:</strong> The old key is now invalid. Update any
                applications using the old key with the new one.
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
      <DialogContent className="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle>Rotate API Key</DialogTitle>
          <DialogDescription>
            Rotate the API key "{apiKey?.name}". This will invalidate the current
            key and create a new one.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          {error && (
            <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-3 text-sm text-red-600 dark:text-red-400">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="reason">Reason for rotation</Label>
            <Select value={reason} onValueChange={(v) => setReason(v as RotationReason)}>
              <SelectTrigger>
                <SelectValue placeholder="Select reason" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="manual">Manual</SelectItem>
                <SelectItem value="automatic">Automatic</SelectItem>
                <SelectItem value="compromised">Compromised</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {ROTATION_REASON_LABELS[reason]}
            </p>
          </div>

          <div className="bg-yellow-50 dark:bg-yellow-950 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              <strong>Warning:</strong> Any applications using the current key will
              stop working until updated with the new key.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleRotate}
            disabled={isLoading}
          >
            {isLoading ? "Rotating..." : "Rotate Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
