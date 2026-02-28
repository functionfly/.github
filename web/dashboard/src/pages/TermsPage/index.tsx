'use client';

import { Navbar } from '@/components/common/Navbar';
import { Footer } from '@/pages/LandingPage/components/Footer';
import {
  TermsHeader,
  TermsContent,
  BackToHome,
} from '@/components/terms';

export function TermsPage() {
  return (
    <div className="terms-page min-h-screen bg-bg-primary">
      <Navbar variant="landing" />

      <TermsHeader />

      <div className="container mx-auto px-4 py-8 pt-20 relative z-10">
        <div className="max-w-4xl mx-auto space-y-8">
          <TermsContent />
          <BackToHome />
        </div>
      </div>

      <Footer />
    </div>
  );
}
