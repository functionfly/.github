'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslation, Trans } from 'react-i18next';
import { useCookieConsentStore, CookieCategories } from '@/stores/cookieConsentStore';
import { CookieCategoryToggle } from './CookieCategoryToggle.tsx';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { getMarketingPageUrl } from '@/lib/constants';
import { Shield, Settings, Cookie } from 'lucide-react';

interface CookiePreferencesModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function CookiePreferencesModal({ isOpen, onClose }: CookiePreferencesModalProps) {
  const { t } = useTranslation();
  const { categories, setConsent } = useCookieConsentStore();
  const [tempCategories, setTempCategories] = useState<CookieCategories>(categories);

  const handleCategoryChange = (category: keyof CookieCategories, enabled: boolean) => {
    setTempCategories(prev => ({
      ...prev,
      [category]: category === 'necessary' ? true : enabled, // Necessary is always true
    }));
  };

  const handleAcceptAll = () => {
    const allCategories: CookieCategories = {
      necessary: true,
      analytics: true,
      marketing: true,
      functionality: true,
    };
    setConsent(allCategories);
    onClose();
  };

  const handleAcceptNecessary = () => {
    const necessaryOnly: CookieCategories = {
      necessary: true,
      analytics: false,
      marketing: false,
      functionality: false,
    };
    setConsent(necessaryOnly);
    onClose();
  };

  const handleSavePreferences = () => {
    setConsent(tempCategories);
    onClose();
  };

  const categoryDescriptions = {
    necessary: {
      title: t('cookieConsent.necessaryTitle'),
      description: t('cookieConsent.necessaryDescription'),
      readOnly: true,
    },
    analytics: {
      title: t('cookieConsent.analyticsTitle'),
      description: t('cookieConsent.analyticsDescription'),
      readOnly: false,
    },
    marketing: {
      title: t('cookieConsent.marketingTitle'),
      description: t('cookieConsent.marketingDescription'),
      readOnly: false,
    },
    functionality: {
      title: t('cookieConsent.functionalityTitle'),
      description: t('cookieConsent.functionalityDescription'),
      readOnly: false,
    },
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <Dialog open={isOpen} onOpenChange={onClose}>
          <DialogContent className="max-w-2xl max-h-[90vh] bg-[var(--bg-primary)] dark:bg-[#0f0f14] border border-[var(--border-subtle)] dark:border-[var(--border-default)] backdrop-blur-xl shadow-2xl dark:shadow-black/50">
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              transition={{ duration: 0.3, ease: [0.25, 0.46, 0.45, 0.94] }}
            >
              <DialogHeader className="relative pb-6">
                <div className="flex items-center gap-3 mb-2">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] dark:from-[#6366f1] dark:to-[#8b5cf6] flex items-center justify-center">
                    <Settings className="w-5 h-5 text-white" />
                  </div>
                  <DialogTitle className="text-2xl font-bold text-[var(--text-primary)]">
                    {t('cookieConsent.title')}
                  </DialogTitle>
                </div>
                <p className="text-[var(--text-secondary)] text-sm">
                  {t('cookieConsent.subtitle')}
                </p>
              </DialogHeader>

              <div className="max-h-[50vh] overflow-y-auto pr-2 custom-scrollbar">
                <div className="space-y-6">
                  {/* Introduction Section */}
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.1 }}
                    className="p-4 rounded-xl bg-[var(--bg-tertiary)] dark:bg-[#1a1a24] border border-[var(--border-subtle)] dark:border-[var(--border-default)]"
                  >
                    <div className="flex items-start gap-3">
                      <Cookie className="w-5 h-5 text-[#6366f1] mt-0.5 flex-shrink-0" />
                      <div>
                        <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-2">{t('cookieConsent.cookieUsage')}</h3>
                        <p className="text-[var(--text-secondary)] text-sm leading-relaxed">
                          {t('cookieConsent.cookieUsageDescription')}
                        </p>
                      </div>
                    </div>
                  </motion.div>

                  {/* Category Toggles */}
                  <div className="space-y-3">
                    {(Object.keys(categoryDescriptions) as Array<keyof typeof categoryDescriptions>).map((category, index) => (
                      <motion.div
                        key={category}
                        initial={{ opacity: 0, x: -20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: 0.2 + index * 0.1 }}
                      >
                        <CookieCategoryToggle
                          category={category}
                          title={categoryDescriptions[category].title}
                          description={categoryDescriptions[category].description}
                          enabled={tempCategories[category]}
                          readOnly={categoryDescriptions[category].readOnly}
                          onChange={(enabled: boolean) => handleCategoryChange(category, enabled)}
                        />
                      </motion.div>
                    ))}
                  </div>

                  {/* Footer Information */}
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.6 }}
                    className="pt-4 border-t border-[var(--border-subtle)] dark:border-[var(--border-default)]"
                  >
                    <div className="flex items-start gap-3">
                      <Shield className="w-5 h-5 text-green-500 mt-0.5 flex-shrink-0" />
                      <div>
                        <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-1">{t('cookieConsent.privacySecurity')}</h4>
                        <p className="text-[var(--text-secondary)] text-xs leading-relaxed">
                          <Trans i18nKey="cookieConsent.privacyContact" components={{
                            1: <a
                              href={getMarketingPageUrl('/privacy')}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-[var(--ff-flame)] hover:text-[var(--ff-afterburner)] transition-colors underline decoration-[var(--ff-flame)]/50 hover:decoration-[var(--ff-afterburner)]/50"
                            />
                          }} />
                        </p>
                      </div>
                    </div>
                  </motion.div>
                </div>
              </div>

              <DialogFooter className="flex-col sm:flex-row gap-3 pt-6 border-t border-[var(--border-subtle)]">
                <motion.div
                  className="flex gap-3 w-full sm:w-auto"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.7 }}
                >
                  <Button
                    variant="outline"
                    onClick={handleAcceptNecessary}
                    className="flex-1 sm:flex-none border-red-500/30 text-red-400 hover:bg-red-500/10 hover:border-red-500/50 dark:border-red-500/30 dark:text-red-400 dark:hover:bg-red-500/10 dark:hover:border-red-500/50"
                  >
                    {t('cookieConsent.rejectAll')}
                  </Button>
                  <Button
                    variant="outline"
                    onClick={onClose}
                    className="flex-1 sm:flex-none border-[var(--border-subtle)] text-[var(--text-primary)] hover:bg-[var(--bg-hover)] dark:border-[var(--border-subtle)] dark:text-[var(--text-primary)] dark:hover:bg-[var(--bg-hover)]"
                  >
                    {t('cookieConsent.cancel')}
                  </Button>
                </motion.div>

                <motion.div
                  className="flex gap-3 w-full sm:w-auto"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.8 }}
                >
                  <Button
                    onClick={handleSavePreferences}
                    className="flex-1 bg-[var(--ff-gradient-cta)] hover:opacity-90"
                  >
                    {t('cookieConsent.savePreferences')}
                  </Button>
                  <Button
                    onClick={handleAcceptAll}
                    className="flex-1 bg-green-600 hover:bg-green-700"
                  >
                    {t('cookieConsent.acceptAll')}
                  </Button>
                </motion.div>
              </DialogFooter>
            </motion.div>
          </DialogContent>
        </Dialog>
      )}
    </AnimatePresence>
  );
}