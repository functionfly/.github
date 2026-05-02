import { Globe, Lock, EyeOff } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

type Visibility = 'public' | 'private' | 'unlisted';

interface VisibilityOption {
  value: Visibility;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  description: string;
}

const VISIBILITY_OPTIONS: VisibilityOption[] = [
  {
    value: 'public',
    icon: Globe,
    label: 'Public',
    description: 'Visible to everyone and listed in the marketplace',
  },
  {
    value: 'private',
    icon: Lock,
    label: 'Private',
    description: 'Only visible to your organization members',
  },
  {
    value: 'unlisted',
    icon: EyeOff,
    label: 'Unlisted',
    description: 'Accessible by link but not listed publicly',
  },
];

interface VisibilitySelectorProps {
  value: Visibility;
  onChange: (value: Visibility) => void;
  size?: 'sm' | 'md';
  className?: string;
}

export function VisibilitySelector({ value, onChange, size = 'md', className }: VisibilitySelectorProps) {
  return (
    <div
      className={cn(
        'inline-flex rounded-lg border border-border-subtle bg-bg-secondary p-0.5',
        className
      )}
      role="radiogroup"
      aria-label="Function visibility"
    >
      <TooltipProvider>
        {VISIBILITY_OPTIONS.map((option) => {
          const Icon = option.icon;
          const isActive = value === option.value;
          return (
            <Tooltip key={option.value}>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  role="radio"
                  aria-checked={isActive}
                  aria-label={option.label}
                  onClick={() => onChange(option.value)}
                  className={cn(
                    'inline-flex items-center justify-center rounded-md transition-all duration-200',
                    size === 'sm' ? 'h-6 w-6' : 'h-8 w-8',
                    isActive
                      ? 'bg-brand-500 text-white shadow-sm'
                      : 'text-text-muted hover:text-text-primary hover:bg-muted'
                  )}
                >
                  <Icon className={cn(size === 'sm' ? 'h-3 w-3' : 'h-4 w-4')} />
                </button>
              </TooltipTrigger>
              <TooltipContent side="top">
                <p className="font-medium">{option.label}</p>
                <p className="text-text-muted text-[10px] mt-0.5">{option.description}</p>
              </TooltipContent>
            </Tooltip>
          );
        })}
      </TooltipProvider>
    </div>
  );
}
