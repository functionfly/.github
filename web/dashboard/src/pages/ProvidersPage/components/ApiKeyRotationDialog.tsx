import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { AlertCircle, Loader2, RefreshCw, Shield, Check, Key } from 'lucide-react';
import { useState } from 'react';
import type { ProviderConfig } from '../constants/providerMeta';

interface ApiKeyRotationDialogProps {
  provider: ProviderConfig;
  accent: { border: string; text: string };
  isOpen: boolean;
  onClose: () => void;
  onRotate: (providerId: string, newApiKey: string) => Promise<void>;
  isRotating?: boolean;
}

export function ApiKeyRotationDialog({
  provider,
  accent,
  isOpen,
  onClose,
  onRotate,
  isRotating = false,
}: ApiKeyRotationDialogProps) {
  const [newApiKey, setNewApiKey] = useState('');
  const [confirmKey, setConfirmKey] = useState('');
  const [validationError, setValidationError] = useState<string | null>(null);
  const [showSuccess, setShowSuccess] = useState(false);

  const handleClose = () => {
    setNewApiKey('');
    setConfirmKey('');
    setValidationError(null);
    setShowSuccess(false);
    onClose();
  };

  const handleSubmit = async () => {
    setValidationError(null);

    // Validation
    if (!newApiKey.trim()) {
      setValidationError('New API key is required');
      return;
    }
    if (newApiKey.length < 10) {
      setValidationError('API key must be at least 10 characters');
      return;
    }
    if (newApiKey !== confirmKey) {
      setValidationError('API keys do not match');
      return;
    }

    try {
      await onRotate(provider.id, newApiKey);
      setShowSuccess(true);
      setTimeout(() => {
        handleClose();
      }, 1500);
    } catch (error) {
      setValidationError(error instanceof Error ? error.message : 'Failed to rotate API key');
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogContent className="bg-bg-tertiary border-border-subtle sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div
              className="w-10 h-10 rounded-xl flex items-center justify-center"
              style={{ backgroundColor: `${accent.border}15` }}
            >
              <Key className="w-5 h-5" style={{ color: accent.text }} />
            </div>
            <div>
              <DialogTitle className="text-text-primary text-lg">Rotate API Key</DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-text-secondary">
            Update your {provider.name} API credentials. This will immediately replace the stored
            credentials with the new key.
          </DialogDescription>
        </DialogHeader>

        {showSuccess ? (
          <div className="p-4 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/50 animate-in slide-in-from-top-2">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center">
                <Check className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
              </div>
              <div>
                <h4 className="font-medium text-emerald-800 dark:text-emerald-400">API Key Rotated</h4>
                <p className="text-sm text-emerald-700 dark:text-emerald-400">
                  Your credentials have been updated successfully.
                </p>
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-4 pt-2">
            {validationError && (
              <div className="p-3 rounded-lg bg-error/10 border border-error/20 animate-in slide-in-from-top-1">
                <div className="flex items-start gap-2">
                  <AlertCircle className="w-4 h-4 text-error mt-0.5 shrink-0" />
                  <p className="text-sm text-error">{validationError}</p>
                </div>
              </div>
            )}

            <Alert className="bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-800/50">
              <Shield className="h-4 w-4 text-amber-600 dark:text-amber-400" />
              <AlertTitle className="text-amber-800 dark:text-amber-300">Security Notice</AlertTitle>
              <AlertDescription className="text-amber-700 dark:text-amber-400 text-sm">
                The old API key will be immediately invalidated. Ensure your new key is active and
                working before rotating.
              </AlertDescription>
            </Alert>

            <div className="space-y-2">
              <Label htmlFor="new-api-key" className="text-text-primary">
                New API Key
              </Label>
              <Input
                id="new-api-key"
                type="password"
                placeholder={`Enter your new ${provider.name} API key`}
                value={newApiKey}
                onChange={(e) => {
                  setNewApiKey(e.target.value);
                  if (validationError) setValidationError(null);
                }}
                className="bg-bg-secondary border-border-subtle focus:border-border-default"
                disabled={isRotating}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="confirm-api-key" className="text-text-primary">
                Confirm API Key
              </Label>
              <Input
                id="confirm-api-key"
                type="password"
                placeholder="Re-enter the API key to confirm"
                value={confirmKey}
                onChange={(e) => {
                  setConfirmKey(e.target.value);
                  if (validationError) setValidationError(null);
                }}
                className="bg-bg-secondary border-border-subtle focus:border-border-default"
                disabled={isRotating}
              />
            </div>

            <div className="flex items-start gap-2 p-3 rounded-lg bg-bg-secondary border border-border-subtle">
              <RefreshCw className="w-4 h-4 text-text-muted mt-0.5 shrink-0" />
              <p className="text-xs text-text-secondary">
                After rotation, all deployments will use the new credentials. Existing functions
                will continue running but new deployments require the new key.
              </p>
            </div>

            <DialogFooter className="gap-2 sm:gap-0 pt-2">
              <Button
                variant="outline"
                onClick={handleClose}
                disabled={isRotating}
                className="border-border-default"
              >
                Cancel
              </Button>
              <Button
                onClick={handleSubmit}
                disabled={!newApiKey.trim() || newApiKey !== confirmKey || isRotating}
                className="gap-2"
                style={{
                  backgroundColor: accent.border,
                  color: 'white',
                }}
              >
                {isRotating ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Rotating...
                  </>
                ) : (
                  <>
                    <RefreshCw className="w-4 h-4" />
                    Rotate Key
                  </>
                )}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
