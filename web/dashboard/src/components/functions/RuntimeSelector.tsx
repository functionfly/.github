import React from 'react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { AlertCircle, CheckCircle2, Info, Shield } from 'lucide-react';
import { type SandboxTier, getTierColor, getTierLabel } from '@/api/sandbox';

export type RuntimeType = 'nodejs18' | 'nodejs20' | 'python3.11' | 'python3.12' | 'deno';

interface RuntimeOption {
  id: RuntimeType;
  name: string;
  description: string;
  status: 'stable' | 'beta' | 'deprecated';
  features: string[];
}

const RUNTIME_OPTIONS: RuntimeOption[] = [
  {
    id: 'nodejs20',
    name: 'Node.js 20',
    description: 'Latest LTS version with improved performance',
    status: 'stable',
    features: ['async/await', 'ES Modules', 'fetch API', 'Worker Threads'],
  },
  {
    id: 'nodejs18',
    name: 'Node.js 18',
    description: 'Previous LTS version',
    status: 'stable',
    features: ['async/await', 'ES Modules', 'fetch API'],
  },
  {
    id: 'deno',
    name: 'Deno',
    description: 'Secure runtime with TypeScript support',
    status: 'beta',
    features: ['TypeScript', 'Built-in formatting', 'Secure by default'],
  },
  {
    id: 'python3.12',
    name: 'Python 3.12',
    description: 'Latest Python with improved performance',
    status: 'stable',
    features: ['async/await', 'Type hints', 'Better error messages'],
  },
  {
    id: 'python3.11',
    name: 'Python 3.11',
    description: 'Previous stable Python version',
    status: 'stable',
    features: ['async/await', 'Type hints'],
  },
];

interface RuntimeSelectorProps {
  value: RuntimeType;
  onChange: (runtime: RuntimeType) => void;
  disabled?: boolean;
  showDescription?: boolean;
  size?: 'sm' | 'md' | 'lg';
  sandboxTier?: SandboxTier;
}

export function RuntimeSelector({
  value,
  onChange,
  disabled = false,
  showDescription = true,
  size = 'md',
  sandboxTier,
}: RuntimeSelectorProps) {
  const selectedRuntime = RUNTIME_OPTIONS.find((r) => r.id === value);

  const getStatusColor = (status: RuntimeOption['status']) => {
    switch (status) {
      case 'stable':
        return 'bg-green-500/10 text-green-500 border-green-500/20';
      case 'beta':
        return 'bg-blue-500/10 text-blue-500 border-blue-500/20';
      case 'deprecated':
        return 'bg-red-500/10 text-red-500 border-red-500/20';
    }
  };

  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return 'h-8 text-xs';
      case 'lg':
        return 'h-12 text-lg';
      default:
        return 'h-10 text-sm';
    }
  };

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium text-foreground">
        Runtime
      </label>
      
      <Select
        value={value}
        onValueChange={(val) => onChange(val as RuntimeType)}
        disabled={disabled}
      >
        <SelectTrigger className={getSizeClasses()}>
          <SelectValue placeholder="Select runtime">
            {selectedRuntime && (
              <div className="flex items-center gap-2">
                <RuntimeBadge 
                  runtime={selectedRuntime.id} 
                  size="sm" 
                  showLabel={true}
                />
                {sandboxTier && (
                  <Badge className={`${getTierColor(sandboxTier)} text-[10px] px-1.5 py-0 border`}>
                    <Shield className="w-2.5 h-2.5 mr-0.5" />
                    {getTierLabel(sandboxTier)}
                  </Badge>
                )}
              </div>
            )}
          </SelectValue>
        </SelectTrigger>
        
        <SelectContent>
          {RUNTIME_OPTIONS.map((runtime) => (
            <SelectItem key={runtime.id} value={runtime.id}>
              <div className="flex flex-col gap-1 py-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{runtime.name}</span>
                  <Badge variant="outline" className={getStatusColor(runtime.status)}>
                    {runtime.status}
                  </Badge>
                </div>
                {showDescription && (
                  <span className="text-xs text-muted-foreground">
                    {runtime.description}
                  </span>
                )}
                {showDescription && runtime.features.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {runtime.features.slice(0, 3).map((feature) => (
                      <Badge key={feature} variant="secondary" className="text-[10px] px-1 py-0">
                        {feature}
                      </Badge>
                    ))}
                  </div>
                )}
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {selectedRuntime && (
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger className="flex items-center gap-1">
                <Info className="w-3 h-3" />
                <span>{selectedRuntime.features.length} features available</span>
              </TooltipTrigger>
              <TooltipContent>
                <div className="space-y-1">
                  <p className="font-medium">Available features:</p>
                  <ul className="list-disc list-inside">
                    {selectedRuntime.features.map((f) => (
                      <li key={f}>{f}</li>
                    ))}
                  </ul>
                </div>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      )}
    </div>
  );
}

// Runtime Badge Component
interface RuntimeBadgeProps {
  runtime: RuntimeType;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
  showVersion?: boolean;
}

export function RuntimeBadge({ 
  runtime, 
  size = 'md', 
  showLabel = false,
  showVersion = true 
}: RuntimeBadgeProps) {
  const getRuntimeInfo = () => {
    switch (runtime) {
      case 'nodejs20':
        return {
          label: 'Node.js',
          version: '20.x',
          color: 'bg-green-500/10 text-green-500 border-green-500/20',
          icon: null,
        };
      case 'nodejs18':
        return {
          label: 'Node.js',
          version: '18.x',
          color: 'bg-green-500/10 text-green-500 border-green-500/20',
          icon: null,
        };
      case 'deno':
        return {
          label: 'Deno',
          version: 'latest',
          color: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
          icon: null,
        };
      case 'python3.12':
        return {
          label: 'Python',
          version: '3.12',
          color: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
          icon: null,
        };
      case 'python3.11':
        return {
          label: 'Python',
          version: '3.11',
          color: 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20',
          icon: null,
        };
      default:
        return {
          label: 'Unknown',
          version: '',
          color: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
          icon: null,
        };
    }
  };

  const info = getRuntimeInfo();

  const sizeClasses = {
    sm: 'text-[10px] px-1.5 py-0.5',
    md: 'text-xs px-2 py-1',
    lg: 'text-sm px-3 py-1.5',
  };

  return (
    <Badge className={`${info.color} ${sizeClasses[size]} border font-normal`}>
      {showLabel && <span className="mr-1">{info.label}</span>}
      {showVersion && <span className="opacity-75">{info.version}</span>}
    </Badge>
  );
}

export default RuntimeSelector;
