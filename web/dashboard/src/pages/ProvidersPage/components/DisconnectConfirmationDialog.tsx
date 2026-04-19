import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { AlertCircle, AlertTriangle, Loader2, X } from 'lucide-react';

interface DisconnectConfirmationDialogProps {
  providerName: string;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isDisconnecting: boolean;
}

export function DisconnectConfirmationDialog({
  providerName,
  isOpen,
  onClose,
  onConfirm,
  isDisconnecting,
}: DisconnectConfirmationDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="bg-bg-tertiary border-border-subtle sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center bg-amber-100 dark:bg-amber-900/30">
              <AlertTriangle className="w-5 h-5 text-amber-600 dark:text-amber-400" />
            </div>
            <div>
              <DialogTitle className="text-text-primary text-lg">Disconnect Provider</DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-text-secondary">
            Are you sure you want to disconnect <strong>{providerName}</strong>? This action cannot
            be undone.
          </DialogDescription>
        </DialogHeader>

        <Alert className="bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-800/50">
          <AlertCircle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
          <AlertTitle className="text-amber-800 dark:text-amber-300">Warning</AlertTitle>
          <AlertDescription className="text-amber-700 dark:text-amber-400 text-sm">
            Disconnecting this provider will:
            <ul className="list-disc list-inside mt-1 space-y-0.5">
              <li>Prevent new deployments to {providerName}</li>
              <li>Remove stored API credentials (they cannot be recovered)</li>
              <li>Keep existing deployments running but disable updates</li>
            </ul>
          </AlertDescription>
        </Alert>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isDisconnecting}
            className="border-border-default"
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isDisconnecting}
            className="gap-2"
          >
            {isDisconnecting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Disconnecting...
              </>
            ) : (
              <>
                <X className="w-4 h-4" />
                Disconnect
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
