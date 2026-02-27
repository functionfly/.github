import { Button } from '@/components/ui/button';
import { Save } from 'lucide-react';

interface AutoSaveIndicatorProps {
  lastSaved: Date | null;
  showDraftIndicator: boolean;
  onClearDraft: () => void;
}

export function AutoSaveIndicator({ lastSaved, showDraftIndicator, onClearDraft }: AutoSaveIndicatorProps) {
  if (!lastSaved) return null;

  return (
    <div className="mb-4 p-3 bg-info/5 border border-info/20 rounded-md flex items-center gap-2">
      <Save className="h-4 w-4 text-info" />
      <span className="text-sm text-info">
        Draft saved automatically
        {showDraftIndicator && <span className="ml-2 text-info">•</span>}
      </span>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        onClick={onClearDraft}
        className="ml-auto text-info hover:text-info"
      >
        Clear draft
      </Button>
    </div>
  );
}