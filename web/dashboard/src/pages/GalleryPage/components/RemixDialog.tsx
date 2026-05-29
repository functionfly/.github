import { GitFork, Loader2 } from 'lucide-react';
import type { GalleryFunction } from '@/api/composer';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

interface RemixDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  fn: GalleryFunction | null;
  customization: string;
  onCustomizationChange: (v: string) => void;
  remixCost: number;
  walletBalance: number;
  canRemix: boolean;
  isOwnFunction: boolean;
  isPending: boolean;
  onConfirm: () => void;
  onAddFunds: () => void;
}

export function RemixDialog({
  open,
  onOpenChange,
  fn,
  customization,
  onCustomizationChange,
  remixCost,
  walletBalance,
  canRemix,
  isOwnFunction,
  isPending,
  onConfirm,
  onAddFunds,
}: RemixDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Remix Function</DialogTitle>
          <DialogDescription>
            Create your own copy of <strong>{fn?.title || fn?.name}</strong> by @{fn?.author}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="bg-muted rounded-lg p-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Remix Cost</span>
              <span className="font-semibold">${remixCost.toFixed(2)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Your Balance</span>
              <span className={walletBalance < remixCost ? 'text-destructive font-semibold' : 'text-emerald-500 font-semibold'}>
                ${walletBalance.toFixed(2)}
              </span>
            </div>
          </div>

          <div>
            <label className="text-sm font-medium mb-2 block">Customizations (optional)</label>
            <textarea
              className="w-full min-h-[90px] rounded-md border bg-muted/50 p-3 text-sm resize-none"
              placeholder="Describe changes you'd like..."
              value={customization}
              onChange={(e) => onCustomizationChange(e.target.value)}
              disabled={!canRemix && !isOwnFunction}
            />
          </div>

          {!canRemix && !isOwnFunction && (
            <p className="text-sm text-destructive bg-destructive/10 p-3 rounded-md">
              You need ${(remixCost - walletBalance).toFixed(2)} more to remix.
            </p>
          )}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          {!canRemix && !isOwnFunction ? (
            <Button onClick={onAddFunds}>Add Funds</Button>
          ) : (
            <Button onClick={onConfirm} disabled={isPending}>
              {isPending ? (
                <><Loader2 className="mr-2 h-4 w-4 animate-spin" />Remixing...</>
              ) : (
                <><GitFork className="mr-2 h-4 w-4" />Remix for ${remixCost.toFixed(2)}</>
              )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
