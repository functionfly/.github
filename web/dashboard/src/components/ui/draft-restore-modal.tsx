import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { FileText, Clock, Trash2 } from 'lucide-react';

interface DraftRestoreModalProps {
  isOpen: boolean;
  onClose: () => void;
  onRestore: () => void;
  onDiscard: () => void;
  lastSavedAt: Date | null;
  draftAge: number | null;
  formName?: string;
}

export function DraftRestoreModal({
  isOpen,
  onClose,
  onRestore,
  onDiscard,
  lastSavedAt,
  draftAge,
  formName = 'this form',
}: DraftRestoreModalProps) {
  const formatAge = (ageMs: number | null) => {
    if (ageMs === null) return 'unknown time ago';
    
    const minutes = Math.floor(ageMs / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
    if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    if (minutes > 0) return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
    return 'just now';
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-full bg-primary/10">
              <FileText className="h-5 w-5 text-primary" />
            </div>
            <div>
              <DialogTitle>Unsaved Draft Found</DialogTitle>
              <DialogDescription className="mt-1.5">
                You have an unsaved draft for {formName}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>
        
        <div className="py-4">
          <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/50 border border-border-subtle">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <div>
              <p className="text-sm font-medium text-text-primary">
                Last saved {formatAge(draftAge)}
              </p>
              {lastSavedAt && (
                <p className="text-xs text-text-secondary">
                  {lastSavedAt.toLocaleString()}
                </p>
              )}
            </div>
          </div>
          
          <p className="text-sm text-text-secondary mt-4">
            Would you like to restore your previous progress or start fresh?
          </p>
        </div>
        
        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="ghost" onClick={onDiscard} className="gap-2">
            <Trash2 className="h-4 w-4" />
            Discard
          </Button>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={onRestore} className="gap-2">
            <FileText className="h-4 w-4" />
            Restore Draft
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
