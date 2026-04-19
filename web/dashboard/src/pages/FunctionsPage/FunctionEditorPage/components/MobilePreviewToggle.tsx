import { Button } from '@/components/ui/button';
import { Eye, EyeOff, Monitor } from 'lucide-react';
import type { FunctionEditorModel } from '../useFunctionEditor';

type Props = { editor: FunctionEditorModel };

export function MobilePreviewToggle({ editor }: Props) {
  const { activeTab, setActiveTab } = editor;

  // Only show on small screens
  return (
    <div className="lg:hidden sticky top-14 z-[25] border-b border-border-subtle bg-bg-secondary/95 backdrop-blur-xl">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 h-12 flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <Monitor className="w-4 h-4" />
          <span className="hidden sm:inline">Preview Mode</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-text-muted">Tap sections below to edit</span>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setActiveTab(activeTab === 'editor' ? 'logs' : 'editor')}
            className="h-8 gap-1.5 text-xs"
          >
            {activeTab === 'editor' ? (
              <>
                <Eye className="w-3.5 h-3.5" />
                <span className="hidden sm:inline">View Logs</span>
              </>
            ) : (
              <>
                <EyeOff className="w-3.5 h-3.5" />
                <span className="hidden sm:inline">View Editor</span>
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
