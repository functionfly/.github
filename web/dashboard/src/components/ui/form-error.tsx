import { AlertCircle, AlertTriangle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';

interface FormErrorProps {
  error?: string | null;
  variant?: 'error' | 'warning' | 'info';
  className?: string;
  showIcon?: boolean;
}

export function FormError({
  error,
  variant = 'error',
  className,
  showIcon = true
}: FormErrorProps) {
  if (!error) return null;

  const variants = {
    error: {
      bg: 'bg-red-500/10',
      border: 'border-red-500/20',
      text: 'text-red-400',
      icon: AlertCircle,
      iconColor: 'text-red-500',
    },
    warning: {
      bg: 'bg-yellow-500/10',
      border: 'border-yellow-500/20',
      text: 'text-yellow-400',
      icon: AlertTriangle,
      iconColor: 'text-yellow-500',
    },
    info: {
      bg: 'bg-blue-500/10',
      border: 'border-blue-500/20',
      text: 'text-blue-400',
      icon: Info,
      iconColor: 'text-blue-500',
    },
  };

  const config = variants[variant];
  const Icon = config.icon;

  return (
    <div className={cn(
      'p-3 rounded-lg border flex items-start gap-3',
      config.bg,
      config.border,
      config.text,
      className
    )}>
      {showIcon && <Icon className={cn('w-4 h-4 mt-0.5 flex-shrink-0', config.iconColor)} />}
      <div className="text-sm flex-1 break-words whitespace-pre-wrap min-w-0">
        {error}
      </div>
    </div>
  );
}
