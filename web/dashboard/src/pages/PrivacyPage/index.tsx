'use client';

import { Navbar } from '@/components/common/Navbar';
import { Footer } from '@/pages/LandingPage/components';
import {
  PrivacyPageHeader,
  CookieConsentStatus,
  PrivacyPolicy,
  CookieDetails,
  UserRights,
  ContactInformation,
  BackToHome,
} from '@/components/privacy';

export function PrivacyPage() {

  return (
    <div className="privacy-page">
      <Navbar variant="landing" />

      <PrivacyPageHeader />

      <div className="container mx-auto px-4 py-8 pt-20 relative z-10">
        <div className="max-w-4xl mx-auto space-y-8">
          <CookieConsentStatus />
          <PrivacyPolicy />
          <CookieDetails />
          <UserRights />
          <ContactInformation />
          <BackToHome />
        </div>
      </div>

      <Footer />
    </div>
  );
}