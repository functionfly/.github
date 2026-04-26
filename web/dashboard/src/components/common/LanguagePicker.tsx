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
import { useDirection } from '@/hooks/useDirection';
import { useTheme } from '@/components/common/ThemeProvider';

import US from 'country-flag-icons/react/3x2/US';
import ES from 'country-flag-icons/react/3x2/ES';
import FR from 'country-flag-icons/react/3x2/FR';
import DE from 'country-flag-icons/react/3x2/DE';
import CN from 'country-flag-icons/react/3x2/CN';
import JP from 'country-flag-icons/react/3x2/JP';
import KR from 'country-flag-icons/react/3x2/KR';
import BR from 'country-flag-icons/react/3x2/BR';
import SA from 'country-flag-icons/react/3x2/SA';
import RU from 'country-flag-icons/react/3x2/RU';
import IN from 'country-flag-icons/react/3x2/IN';
import NL from 'country-flag-icons/react/3x2/NL';
import PL from 'country-flag-icons/react/3x2/PL';
import TR from 'country-flag-icons/react/3x2/TR';
import VN from 'country-flag-icons/react/3x2/VN';

const FLAG_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  en: US,
  es: ES,
  fr: FR,
  de: DE,
  zh: CN,
  ja: JP,
  ko: KR,
  pt: BR,
  ar: SA,
  ru: RU,
  hi: IN,
  nl: NL,
  pl: PL,
  tr: TR,
  vi: VN,
};

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
  const { isRtl, applyDir } = useDirection();
  const { theme } = useTheme();
  const isDark = theme === 'dark';

  const currentLang = getLanguage(i18n.language) ?? getLanguage('en')!;

  const handleSelect = useCallback(
    async (lang: Language) => {
      // Optimistically update i18n
      await i18n.changeLanguage(lang.code);
      persistLanguage(lang.code);
      applyDir(lang.code);
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
    [i18n, user, queryClient, applyDir]
  );

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
              {(() => {
                const FlagComponent = FLAG_ICONS[i18n.language.split('-')[0]];
                return FlagComponent ? (
                  <FlagComponent className="w-5 h-3.5 rounded-sm" />
                ) : null;
              })()}
              {showLabel && (
                <span className={cn('text-sm', isRtl && 'font-medium')} style={{ color: isDark ? '#e0e0e8' : 'inherit' }}>
                  {currentLang.name}
                </span>
              )}
              <ChevronDown
                className={cn('size-3.5 transition-transform duration-200', open && 'rotate-180')}
                style={{ color: isDark ? '#b0b0c0' : 'var(--text-muted)' }}
              />
            </>
          )}
        </Button>
      </PopoverTrigger>

      <PopoverContent
        align="end"
        sideOffset={8}
        className={cn(
          'w-80 p-0 overflow-hidden rounded-xl',
          'border shadow-xl outline-none',
          'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2'
        )}
        style={{
          backgroundColor: isDark ? '#151520' : '#ffffff',
          borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
          boxShadow: isDark
            ? '0 25px 50px -12px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(255, 107, 53, 0.1)'
            : '0 25px 50px -12px rgba(0, 0, 0, 0.15)',
        }}
      >
        {/* Header */}
        <div className="px-4 pt-4 pb-3" style={{ borderBottom: `1px solid ${isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)'}` }}>
          <div className="flex items-center gap-2.5">
            <div className="size-8 rounded-lg flex items-center justify-center" style={{ backgroundColor: isDark ? 'rgba(255, 107, 53, 0.15)' : 'rgba(255, 107, 53, 0.1)' }}>
              <Languages className="size-4" style={{ color: '#FF6B35' }} />
            </div>
            <div>
              <p className="text-sm font-semibold" style={{ color: isDark ? '#e0e0e8' : '#1a1a1f' }}>
                {t('language.title')}
              </p>
              <p className="text-xs mt-0.5" style={{ color: isDark ? '#a0a0b8' : '#5f6368' }}>
                {t('language.description')}
              </p>
            </div>
          </div>
        </div>

        {/* Auto-detected banner */}
        {user?.language === undefined && detectedLang !== i18n.language && (
          <div className="px-4 py-2 flex items-center gap-2" style={{ backgroundColor: isDark ? 'rgba(245, 158, 11, 0.1)' : 'rgba(245, 158, 11, 0.05)', borderBottom: `1px solid ${isDark ? 'rgba(245, 158, 11, 0.15)' : 'rgba(245, 158, 11, 0.1)'}` }}>
            <span className="text-[10px] uppercase tracking-wider font-semibold" style={{ color: '#f59e0b' }}>
              {t('language.detected')}
            </span>
            <span className="text-xs" style={{ color: isDark ? '#e0e0e8' : '#5f6368' }}>
              {getLanguage(detectedLang)?.name ?? 'English'}
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
                  isSelected && 'bg-brand-500/10'
                )}
                style={{
                  backgroundColor: isDark ? 'rgba(255, 255, 255, 0.02)' : 'transparent',
                }}
              >
                {/* Flag */}
                {(() => {
                  const FlagComponent = FLAG_ICONS[lang.code];
                  return FlagComponent ? (
                    <FlagComponent className="w-7 h-5 rounded-sm" />
                  ) : null;
                })()}

                {/* Names */}
                <div className="flex-1 text-left">
                  <p
                    className={cn(
                      'text-sm transition-colors duration-150',
                      isSelected
                        ? 'font-semibold'
                        : ''
                    )}
                    style={{
                      color: isSelected ? '#FF6B35' : (isDark ? '#e0e0e8' : 'var(--text-primary)')
                    }}
                    dir={lang.dir ?? 'ltr'}
                  >
                    {lang.name}
                  </p>
                  {lang.nameEn !== lang.name && (
                    <p className="text-xs mt-0.5" style={{ color: 'var(--text-muted)' }}>{lang.nameEn}</p>
                  )}
                </div>

                {/* Check mark */}
                {isSelected && (
                  <Check className="size-4 flex-shrink-0" style={{ color: '#FF6B35' }} />
                )}
              </button>
            );
          })}
        </div>

        {/* Footer hint */}
        <div className="px-4 py-2.5" style={{ borderTop: `1px solid ${isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)'}`, backgroundColor: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)' }}>
          <p className="text-[11px] text-center" style={{ color: 'var(--text-muted)' }}>
            {t('language.autoDetected')}
          </p>
        </div>
      </PopoverContent>
    </Popover>
  );
}
