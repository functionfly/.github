import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  LANGUAGES_BY_NATIVE_NAME,
  getLanguage,
  detectLanguage,
  persistLanguage,
  type Language,
} from '@/lib/i18n/languages';
import { usersApi } from '@/api/users';
import { useAuthStore } from '@/stores/authStore';
import { useQueryClient } from '@tanstack/react-query';
import { Check, ChevronDown, Globe, Languages } from 'lucide-react';
import { cn } from '@/lib/utils';

interface LanguagePickerProps {
  className?: string;
  variant?: 'ghost' | 'default' | 'icon';
  showLabel?: boolean;
}

export function LanguagePicker({
  className,
  variant = 'ghost',
  showLabel = true,
}: LanguagePickerProps) {
  const { i18n, t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [detectedLang] = useState(detectLanguage);
  const { user } = useAuthStore();
  const queryClient = useQueryClient();

  const currentLang = getLanguage(i18n.language) ?? getLanguage('en')!;

  const handleSelect = useCallback(
    async (lang: Language) => {
      // Optimistically update i18n
      await i18n.changeLanguage(lang.code);
      persistLanguage(lang.code);
      setOpen(false);

      // Persist to backend if user is logged in
      if (user) {
        try {
          await usersApi.updateMe({ language: lang.code });
          queryClient.invalidateQueries({ queryKey: ['user-me'] });
        } catch {
          // Non-critical: language preference is stored locally even if API fails
        }
      }
    },
    [i18n, user, queryClient]
  );

  const isRtl = currentLang.dir === 'rtl';

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant={variant === 'icon' ? 'ghost' : variant}
          className={cn(
            'gap-2 transition-all duration-200',
            variant === 'icon' && 'size-9 p-0 rounded-lg',
            className
          )}
          aria-label={t('language.changeLanguage')}
        >
          {variant === 'icon' ? (
            <Globe className="size-4" />
          ) : (
            <>
              <span className="text-base leading-none">{currentLang.flag}</span>
              {showLabel && (
                <span className={cn('text-sm', isRtl && 'font-medium')}>
                  {currentLang.name}
                </span>
              )}
              <ChevronDown
                className={cn(
                  'size-3.5 text-text-muted transition-transform duration-200',
                  open && 'rotate-180'
                )}
              />
            </>
          )}
        </Button>
      </PopoverTrigger>

      <PopoverContent
        align="end"
        sideOffset={8}
        className={cn(
          'w-80 p-0 overflow-hidden',
          'bg-bg-secondary border border-border-subtle shadow-xl',
          'rounded-xl'
        )}
      >
        {/* Header */}
        <div className="px-4 pt-4 pb-3 border-b border-border-subtle">
          <div className="flex items-center gap-2.5">
            <div className="size-8 rounded-lg bg-brand-500/10 flex items-center justify-center">
              <Languages className="size-4 text-brand-400" />
            </div>
            <div>
              <p className="text-sm font-semibold text-text-primary">
                {t('language.title')}
              </p>
              <p className="text-xs text-text-muted mt-0.5">
                {t('language.description')}
              </p>
            </div>
          </div>
        </div>

        {/* Auto-detected banner */}
        {user?.language === undefined && detectedLang !== i18n.language && (
          <div className="px-4 py-2 bg-amber-500/5 border-b border-amber-500/10 flex items-center gap-2">
            <span className="text-[10px] uppercase tracking-wider text-amber-500 font-semibold">
              {t('language.detected')}
            </span>
            <span className="text-xs text-text-secondary">
              {getLanguage(detectLanguage)?.name ?? 'English'}
            </span>
          </div>
        )}

        {/* Language list */}
        <div className="max-h-80 overflow-y-auto py-1 overflow-x-hidden">
          {LANGUAGES_BY_NATIVE_NAME.map((lang) => {
            const isSelected = lang.code === currentLang.code;
            return (
              <button
                key={lang.code}
                onClick={() => handleSelect(lang)}
                className={cn(
                  'w-full flex items-center gap-3 px-4 py-2.5 transition-colors duration-150',
                  'hover:bg-bg-hover',
                  isSelected && 'bg-brand-500/5'
                )}
              >
                {/* Flag */}
                <span
                  className={cn(
                    'text-xl leading-none transition-transform duration-200',
                    isSelected && 'scale-110'
                  )}
                  style={{ fontFamily: 'sans-serif' }}
                >
                  {lang.flag}
                </span>

                {/* Names */}
                <div className="flex-1 text-left">
                  <p
                    className={cn(
                      'text-sm transition-colors duration-150',
                      isSelected
                        ? 'text-brand-400 font-semibold'
                        : 'text-text-primary'
                    )}
                    dir={lang.dir ?? 'ltr'}
                  >
                    {lang.name}
                  </p>
                  {lang.nameEn !== lang.name && (
                    <p className="text-xs text-text-muted mt-0.5">{lang.nameEn}</p>
                  )}
                </div>

                {/* Check mark */}
                {isSelected && (
                  <Check className="size-4 text-brand-400 flex-shrink-0" />
                )}
              </button>
            );
          })}
        </div>

        {/* Footer hint */}
        <div className="px-4 py-2.5 border-t border-border-subtle bg-bg-tertiary/30">
          <p className="text-[11px] text-text-muted text-center">
            {t('language.autoDetected')}
          </p>
        </div>
      </PopoverContent>
    </Popover>
  );
}
