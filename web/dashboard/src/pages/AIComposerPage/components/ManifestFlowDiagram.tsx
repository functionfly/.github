import { Zap } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import type { ManifestFlow } from '../types';

export function ManifestFlowDiagram({ manifest }: { manifest?: ManifestFlow }) {
  if (!manifest) return null;

  return (
    <div className="relative py-4">
      <div className="flex items-center justify-center gap-4">
        <div className="flex flex-col gap-2">
          <span className="text-xs font-medium text-muted-foreground text-center">Inputs</span>
          <div className="space-y-1.5">
            {manifest.inputs.length > 0 ? (
              manifest.inputs.slice(0, 3).map((input, i) => (
                <div
                  key={i}
                  className="px-2 py-1 rounded bg-blue-500/10 border border-blue-500/20 text-xs flex items-center gap-1.5 min-w-[100px]"
                >
                  <div className="w-1.5 h-1.5 rounded-full bg-blue-500" />
                  <span className="truncate">{input.name}</span>
                </div>
              ))
            ) : (
              <div className="px-2 py-1 rounded bg-muted text-xs text-muted-foreground italic">
                No inputs
              </div>
            )}
            {manifest.inputs.length > 3 && (
              <div className="text-xs text-muted-foreground text-center">
                +{manifest.inputs.length - 3} more
              </div>
            )}
          </div>
        </div>

        <div className="flex flex-col items-center gap-1">
          <div className="w-12 h-0.5 bg-gradient-to-r from-blue-500 via-violet-500 to-green-500 rounded-full" />
          <div className="flex items-center gap-1">
            <div className="w-0 h-0 border-t-[4px] border-t-transparent border-l-[6px] border-l-violet-500 border-b-[4px] border-b-transparent" />
          </div>
          <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4">
            <Zap className="w-3 h-3 mr-1" />
            Function
          </Badge>
        </div>

        <div className="flex flex-col gap-2">
          <span className="text-xs font-medium text-muted-foreground text-center">Outputs</span>
          <div className="space-y-1.5">
            {manifest.outputs.length > 0 ? (
              manifest.outputs.slice(0, 3).map((output, i) => (
                <div
                  key={i}
                  className="px-2 py-1 rounded bg-green-500/10 border border-green-500/20 text-xs flex items-center gap-1.5 min-w-[100px]"
                >
                  <div className="w-1.5 h-1.5 rounded-full bg-green-500" />
                  <span className="truncate">{output.name}</span>
                </div>
              ))
            ) : (
              <div className="px-2 py-1 rounded bg-muted text-xs text-muted-foreground italic">
                No outputs
              </div>
            )}
            {manifest.outputs.length > 3 && (
              <div className="text-xs text-muted-foreground text-center">
                +{manifest.outputs.length - 3} more
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
