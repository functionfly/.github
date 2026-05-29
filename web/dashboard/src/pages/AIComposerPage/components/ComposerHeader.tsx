import { History, Save, Sparkles, XCircle } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

interface ComposerHeaderProps {
  hasDraft: boolean;
  lastSaved: Date | null;
  historyCount: number;
  historySidebarOpen: boolean;
  onClearDraft: () => void;
  onToggleHistory: () => void;
}

export function ComposerHeader({
  hasDraft,
  lastSaved,
  historyCount,
  historySidebarOpen,
  onClearDraft,
  onToggleHistory,
}: ComposerHeaderProps) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-4">
        <div className="p-3 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 shadow-lg">
          <Sparkles className="h-8 w-8 text-white" />
        </div>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">AI Composer</h1>
          <p className="text-muted-foreground">
            Describe what you need. FlyMind AI generates production-ready functions.
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3">
        {hasDraft && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-muted/50 text-xs text-muted-foreground">
                  <Save className="w-3.5 h-3.5" />
                  <span>Draft saved</span>
                  <button onClick={onClearDraft} className="hover:text-destructive">
                    <XCircle className="w-3.5 h-3.5" />
                  </button>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p className="text-xs">
                  Last saved: {lastSaved?.toLocaleTimeString()}
                  <br />
                  Click × to clear
                </p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}

        <Button variant="outline" size="sm" onClick={onToggleHistory} className="gap-2">
          <History className="h-4 w-4" />
          {historySidebarOpen ? 'Hide History' : 'Show History'}
          <Badge variant="secondary" className="ml-1 text-xs">
            {historyCount}
          </Badge>
        </Button>
      </div>
    </div>
  );
}
