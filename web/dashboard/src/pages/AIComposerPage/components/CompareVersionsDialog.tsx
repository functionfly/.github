import type { FunctionGenerationResponse, FunctionGenerationResult } from '@/api/composer';
import { RUNTIME_MONACO_LANG } from '@/api/composer';
import { LazyMonacoEditor } from '@/components/LazyMonacoEditor';
import {
  Button,
} from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import type { GenerationHistoryItem } from '../useGenerationHistory';

interface CompareVersionsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  compareItems: [GenerationHistoryItem | null, GenerationHistoryItem | null];
  monacoTheme: string;
}

export function CompareVersionsDialog({
  open,
  onOpenChange,
  compareItems,
  monacoTheme,
}: CompareVersionsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh]">
        <DialogHeader>
          <DialogTitle>Compare Versions</DialogTitle>
          <DialogDescription>Comparing code changes between generations</DialogDescription>
        </DialogHeader>
        <div className="grid grid-cols-2 gap-4 mt-4">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Badge variant="default">Current</Badge>
              <span className="text-xs text-muted-foreground">
                {compareItems[0] && new Date(compareItems[0].timestamp).toLocaleString()}
              </span>
            </div>
            <div className="rounded-md border overflow-hidden">
              <LazyMonacoEditor
                height="400px"
                language={
                  compareItems[0]
                    ? RUNTIME_MONACO_LANG[compareItems[0].runtime] || 'plaintext'
                    : 'plaintext'
                }
                value={compareItems[0]?.result.code || ''}
                theme={monacoTheme}
                options={{
                  readOnly: true,
                  minimap: { enabled: false },
                  fontSize: 12,
                  lineNumbers: 'on',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  wordWrap: 'on',
                }}
              />
            </div>
          </div>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <Badge variant="secondary">Previous</Badge>
              <span className="text-xs text-muted-foreground">
                {compareItems[1]
                  ? new Date(compareItems[1].timestamp).toLocaleString()
                  : 'None selected'}
              </span>
            </div>
            <div className="rounded-md border overflow-hidden">
              <LazyMonacoEditor
                height="400px"
                language={
                  compareItems[1]
                    ? RUNTIME_MONACO_LANG[compareItems[1].runtime] || 'plaintext'
                    : 'plaintext'
                }
                value={compareItems[1]?.result.code || '// No previous version to compare'}
                theme={monacoTheme}
                options={{
                  readOnly: true,
                  minimap: { enabled: false },
                  fontSize: 12,
                  lineNumbers: 'on',
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                  wordWrap: 'on',
                }}
              />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function buildGenerationResponseFromHistory(
  item: GenerationHistoryItem
): FunctionGenerationResponse {
  return {
    success: true,
    result: item.result,
    generation_id: item.id,
    latency_ms: 0,
    tokens_used: { prompt: 0, completion: 0, total: 0 },
  };
}

export function buildStreamingResultFromHistory(
  result: FunctionGenerationResult
): Partial<FunctionGenerationResult> & { code: string } {
  return { ...result, code: result.code };
}
