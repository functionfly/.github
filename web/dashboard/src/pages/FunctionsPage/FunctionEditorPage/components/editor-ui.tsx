import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { AlertCircle, HelpCircle } from 'lucide-react';
import type { ReactNode } from 'react';

export function SectionCard({
  icon,
  title,
  description,
  step,
  children,
}: {
  icon: ReactNode;
  title: string;
  description?: string;
  step?: number;
  children: ReactNode;
}) {
  return (
    <Card
      className="card overflow-hidden border-border-subtle/50"
      style={{
        background: 'var(--bg-secondary, #12121a)',
      }}
    >
      <CardHeader className="pb-3 pt-4 px-5">
        <div className="flex items-center gap-3">
          {step !== undefined && (
            <div
              className="flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold shadow-sm"
              style={{
                background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)',
                color: '#fff',
              }}
            >
              {step}
            </div>
          )}
          <span className="text-indigo-400 flex-shrink-0">{icon}</span>
          <div className="min-w-0">
            <CardTitle className="text-sm font-semibold text-text-primary">{title}</CardTitle>
            {description && (
              <p className="text-xs text-text-muted mt-0.5 leading-relaxed">{description}</p>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-5 pb-5 space-y-4">{children}</CardContent>
    </Card>
  );
}

export function FieldError({ message }: { message?: string }) {
  if (!message) return null;
  return (
    <p className="flex items-center gap-1.5 text-xs text-red-400 mt-1.5">
      <AlertCircle className="w-3 h-3 flex-shrink-0" />
      {message}
    </p>
  );
}

export function InfoTip({ content }: { content: string }) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <HelpCircle className="w-3.5 h-3.5 text-text-muted cursor-help inline-block ml-1 align-middle" />
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-xs text-xs">
          {content}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
