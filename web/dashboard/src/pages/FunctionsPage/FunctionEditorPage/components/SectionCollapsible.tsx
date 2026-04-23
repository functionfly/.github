import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

interface SectionCollapsibleProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
  badge?: string;
}

export function SectionCollapsible({
  title,
  children,
  defaultOpen = false,
  badge,
}: SectionCollapsibleProps) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <Card
      className="overflow-hidden border-border-subtle/50 transition-all duration-200"
      style={{ background: 'var(--bg-secondary)' }}
    >
      <CardHeader className="pb-0 pt-3 px-4">
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center justify-between w-full py-2 text-left"
          aria-expanded={isOpen}
        >
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-text-primary">{title}</span>
            {badge && (
              <span className="text-xs px-2 py-0.5 rounded-full bg-[#FF6B35]/10 text-[#FF6B35]">
                {badge}
              </span>
            )}
          </div>
          <div className="flex items-center gap-2 text-text-muted">
            <span className="text-xs hidden sm:inline">{isOpen ? t('common.hide', { defaultValue: 'Hide' }) : t('common.show', { defaultValue: 'Show' })}</span>
            {isOpen ? (
              <ChevronUp className="w-4 h-4" />
            ) : (
              <ChevronDown className="w-4 h-4" />
            )}
          </div>
        </button>
      </CardHeader>
      <CardContent
        className={`px-4 transition-all duration-200 ${
          isOpen ? 'pb-4 pt-2 opacity-100' : 'pb-0 pt-0 opacity-0 h-0 overflow-hidden'
        }`}
      >
        {children}
      </CardContent>
    </Card>
  );
}
