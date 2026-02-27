import { Shield } from 'lucide-react';

export function PrivacyPageHeader() {
  return (
    <div className="privacy-header">
      <div className="container mx-auto px-4">
        <div className="icon-container">
          <Shield className="h-8 w-8" />
        </div>
        <h1>Privacy & Cookie Policy</h1>
        <p>
          Your privacy is important to us. This page explains how we collect, use, and protect your data.
        </p>
      </div>
    </div>
  );
}