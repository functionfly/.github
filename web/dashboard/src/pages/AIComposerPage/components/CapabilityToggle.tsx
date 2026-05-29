import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Check, Settings2 } from 'lucide-react';
import { CAPABILITY_INFO } from '../constants';

export function CapabilityToggle({
  capability,
  enabled,
  onToggle,
}: {
  capability: string;
  enabled: boolean;
  onToggle: (capability: string) => void;
}) {
  const info = CAPABILITY_INFO[capability] || {
    description: `Enable ${capability} capability`,
    icon: Settings2,
  };
  const CapabilityIcon = info.icon;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={`
              flex items-center gap-2 px-2.5 py-1.5 rounded-md border cursor-pointer transition-all
              ${
                enabled
                  ? 'bg-violet-500/10 border-violet-500/30 text-violet-700 dark:text-violet-300'
                  : 'bg-muted/50 border-muted text-muted-foreground hover:bg-muted'
              }
            `}
            onClick={() => onToggle(capability)}
          >
            <CapabilityIcon className="w-3 h-3" />
            <span className="text-xs font-medium capitalize">{capability}</span>
            {enabled && <Check className="w-3 h-3 ml-auto" />}
          </div>
        </TooltipTrigger>
        <TooltipContent side="top">
          <p className="text-xs max-w-[200px]">{info.description}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
