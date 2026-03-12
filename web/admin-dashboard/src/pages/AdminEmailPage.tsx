/**
 * Admin Email Page
 * Tabbed interface: Newsletter (subscribers & campaigns) and Email Settings
 */

import { useState } from 'react';
import { Mail, Settings, Send } from 'lucide-react';

type TabId = 'newsletter' | 'settings';

export function AdminEmailPage() {
  const [activeTab, setActiveTab] = useState<TabId>('newsletter');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Email</h1>
        <p className="mt-2 text-gray-600">
          Manage newsletter subscribers, campaigns, and platform email settings.
        </p>
      </div>

      <div className="border-b border-gray-200">
        <nav className="flex gap-6">
          <button
            type="button"
            onClick={() => setActiveTab('newsletter')}
            className={`flex items-center gap-2 pb-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'newsletter'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Send className="w-4 h-4" />
            Newsletter
          </button>
          <button
            type="button"
            onClick={() => setActiveTab('settings')}
            className={`flex items-center gap-2 pb-3 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'settings'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Settings className="w-4 h-4" />
            Email Settings
          </button>
        </nav>
      </div>

      {activeTab === 'newsletter' && <NewsletterTab />}
      {activeTab === 'settings' && <EmailSettingsTab />}
    </div>
  );
}

function NewsletterTab() {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6">
      <div className="flex items-center gap-2 mb-4">
        <Mail className="w-5 h-5 text-indigo-600" />
        <h2 className="text-lg font-semibold text-gray-900">Newsletter</h2>
      </div>
      <p className="text-gray-600 mb-4">
        Subscribers and campaigns. This area is wired in the new dashboard and ready for full migration.
      </p>
      <div className="rounded-lg bg-gray-50 border border-gray-200 p-4 text-sm text-gray-500">
        Newsletter features (subscriber list, campaigns, templates) can be implemented here.
      </div>
    </div>
  );
}

function EmailSettingsTab() {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6 max-w-2xl">
      <div className="flex items-center gap-2 mb-6">
        <Settings className="w-5 h-5 text-indigo-600" />
        <h2 className="text-lg font-semibold text-gray-900">Email Settings</h2>
      </div>
      <p className="text-gray-600 mb-6">
        Configure how the platform sends transactional and marketing email. Some options are set via server environment variables.
      </p>

      <div className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">From address</label>
          <input
            type="email"
            placeholder="noreply@functionfly.dev"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
            readOnly
            disabled
          />
          <p className="mt-1 text-xs text-gray-500">Set via FROM_EMAIL on the server.</p>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">SMTP host</label>
          <input
            type="text"
            placeholder="localhost or smtp.example.com"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 bg-gray-50"
            readOnly
            disabled
          />
          <p className="mt-1 text-xs text-gray-500">Set via SMTP_HOST (and SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD) on the server.</p>
        </div>
        <div className="rounded-lg bg-amber-50 border border-amber-200 p-4 text-sm text-amber-800">
          <strong>Note:</strong> To change SMTP or from-address, update environment variables on the orchestrator API server and restart. Admin UI controls for these may be added in a future release.
        </div>
      </div>
    </div>
  );
}
