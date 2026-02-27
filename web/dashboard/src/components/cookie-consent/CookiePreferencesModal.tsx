'use client';

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useCookieConsentStore, CookieCategories } from '@/stores/cookieConsentStore';
import { CookieCategoryToggle } from './CookieCategoryToggle.tsx';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Shield, Settings, Cookie } from 'lucide-react';

interface CookiePreferencesModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function CookiePreferencesModal({ isOpen, onClose }: CookiePreferencesModalProps) {
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
      title: 'Strictly Necessary Cookies',
      description: 'These cookies are essential for the proper functioning of our website. Without these cookies, the website would not work properly.',
      readOnly: true,
    },
    analytics: {
      title: 'Analytics Cookies',
      description: 'These cookies help us understand how visitors interact with our website by collecting and reporting information anonymously.',
      readOnly: false,
    },
    marketing: {
      title: 'Marketing Cookies',
      description: 'These cookies are used to track visitors across websites to display ads that are relevant and engaging for individual users.',
      readOnly: false,
    },
    functionality: {
      title: 'Functionality Cookies',
      description: 'These cookies enable the website to remember choices you make and provide enhanced, more personal features.',
      readOnly: false,
    },
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <Dialog open={isOpen} onOpenChange={onClose}>
          <DialogContent className="max-w-2xl max-h-[90vh] bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 border border-white/10 backdrop-blur-xl">
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              transition={{ duration: 0.3, ease: [0.25, 0.46, 0.45, 0.94] }}
            >
              <DialogHeader className="relative pb-6">
                <div className="flex items-center gap-3 mb-2">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                    <Settings className="w-5 h-5 text-white" />
                  </div>
                  <DialogTitle className="text-2xl font-bold bg-gradient-to-r from-white to-text-secondary bg-clip-text text-transparent">
                    Cookie Preferences
                  </DialogTitle>
                </div>
                <p className="text-text-secondary text-sm">
                  Manage your cookie preferences and privacy settings
                </p>
              </DialogHeader>

              <div className="max-h-[50vh] overflow-y-auto pr-2 custom-scrollbar">
                <div className="space-y-6">
                  {/* Introduction Section */}
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.1 }}
                    className="p-4 rounded-xl bg-gradient-to-r from-white/5 to-white/10 border border-white/10"
                  >
                    <div className="flex items-start gap-3">
                      <Cookie className="w-5 h-5 text-[#6366f1] mt-0.5 flex-shrink-0" />
                      <div>
                        <h3 className="text-lg font-semibold text-white mb-2">Cookie Usage</h3>
                        <p className="text-text-secondary text-sm leading-relaxed">
                          We use cookies to ensure the basic functionalities of the website and to enhance your online experience.
                          Choose which categories you'd like to enable.
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
                    className="pt-4 border-t border-white/10"
                  >
                    <div className="flex items-start gap-3">
                      <Shield className="w-5 h-5 text-green-400 mt-0.5 flex-shrink-0" />
                      <div>
                        <h4 className="text-sm font-semibold text-white mb-1">Privacy & Security</h4>
                        <p className="text-text-secondary text-xs leading-relaxed">
                          For any queries about our cookie policy or privacy practices, please{' '}
                          <a
                            href="/privacy"
                            className="text-[#6366f1] hover:text-[#8b5cf6] transition-colors underline decoration-[#6366f1]/50 hover:decoration-[#8b5cf6]/50"
                          >
                            contact us
                          </a>
                          .
                        </p>
                      </div>
                    </div>
                  </motion.div>
                </div>
              </div>

              <DialogFooter className="flex-col sm:flex-row gap-3 pt-6 border-t border-white/10">
                <motion.div
                  className="flex gap-3 w-full sm:w-auto"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.7 }}
                >
                  <Button
                    variant="outline"
                    onClick={handleAcceptNecessary}
                    className="flex-1 sm:flex-none border-red-500/30 text-red-400 hover:bg-red-500/10 hover:border-red-500/50"
                  >
                    Reject All
                  </Button>
                  <Button
                    variant="outline"
                    onClick={onClose}
                    className="flex-1 sm:flex-none"
                  >
                    Cancel
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
                    className="flex-1 bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] hover:from-[#6366f1]/90 hover:to-[#8b5cf6]/90"
                  >
                    Save Preferences
                  </Button>
                  <Button
                    onClick={handleAcceptAll}
                    className="flex-1 bg-gradient-to-r from-green-600 to-green-500 hover:from-green-700 hover:to-green-600"
                  >
                    Accept All
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