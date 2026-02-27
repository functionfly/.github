import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { CookiePreferencesModal } from '@/components/cookie-consent';
import { useCookieConsent } from '@/hooks/useCookieConsent';
import { Cookie, Settings } from 'lucide-react';

export function CookieConsentStatus() {
  const [showPreferencesModal, setShowPreferencesModal] = useState(false);
  const { categories, consentTimestamp } = useCookieConsent();

  const formatDate = (date: Date | string | null) => {
    if (!date) return 'Not given';

    // Handle both Date objects and date strings (from localStorage persistence)
    const dateObj = typeof date === 'string' ? new Date(date) : date;

    // Check if the date is valid
    if (isNaN(dateObj.getTime())) return 'Not given';

    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(dateObj);
  };

  return (
    <>
      <div className="privacy-card">
        <div className="space-y-6">
          <div>
            <h3 className="text-xl font-bold text-white flex items-center gap-2">
              <Cookie className="h-5 w-5" />
              Your Cookie Preferences
            </h3>
          </div>
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-muted-foreground">Consent Status</p>
                <p className="font-medium">
                  {consentTimestamp ? 'Consent Given' : 'No Consent Given'}
                </p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Last Updated</p>
                <p className="font-medium">{formatDate(consentTimestamp)}</p>
              </div>
            </div>
            <Button
              variant="outline"
              onClick={() => setShowPreferencesModal(true)}
              className="w-full sm:w-auto"
            >
              <Settings className="h-4 w-4 mr-2" />
              Manage Cookie Preferences
            </Button>
          </div>
        </div>
      </div>

      <CookiePreferencesModal
        isOpen={showPreferencesModal}
        onClose={() => setShowPreferencesModal(false)}
      />
    </>
  );
}