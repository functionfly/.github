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
      className="overflow-hidden"
      style={{
        background: 'var(--panel)',
        backgroundImage: 'radial-gradient(140% 100% at 15% 0%, var(--glass-tint), transparent 55%)',
        borderColor: 'var(--panel-edge)',
        borderRadius: 'var(--radius-lg)',
        boxShadow: 'var(--shadow-chamber)',
      }}
    >
      <CardHeader className="pb-3 pt-4 px-5">
        <div className="flex items-center gap-3">
          {step !== undefined && (
            <div
              className="flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold"
              style={{
                background: 'var(--status-ok)',
                color: 'var(--bg)',
              }}
            >
              {step}
            </div>
          )}
          <span className="text-[var(--status-ok)] flex-shrink-0">{icon}</span>
          <div className="min-w-0">
            <CardTitle className="text-sm font-semibold text-[var(--text)]" style={{ fontFamily: 'var(--font-display)' }}>{title}</CardTitle>
            {description && (
              <p className="text-xs text-[var(--text-faint)] mt-0.5 leading-relaxed">{description}</p>
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
    <p className="flex items-center gap-1.5 text-xs text-[var(--status-revoked)] mt-1.5">
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
