'use client';

import { useEffect, useRef, useState } from 'react';
import { Loader2, AlertCircle } from 'lucide-react';

/**
 * ConnectorsCallbackPage handles the initial popup redirect in the OAuth flow.
 *
 * Flow:
 * 1. Parent opens popup to this page with ?slug=xxx&oauth_url=<encoded_url>
 * 2. This page redirects the popup to the OAuth provider (Google, GitHub, etc.)
 * 3. User authorizes on the provider
 * 4. Provider redirects the popup to the backend /v1/connectors/callback
 * 5. Backend exchanges code for tokens, creates the connector, and renders an
 *    HTML page that uses postMessage to notify this window's opener, then auto-closes
 *
 * If the OAuth provider returns an error (e.g., user denied), this page shows the error.
 */
export function ConnectorsCallbackPage() {
  const redirected = useRef(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (redirected.current) return;
    redirected.current = true;

    // Use native URL parsing for reliability
    const params = new URLSearchParams(window.location.search);

    // Check if OAuth provider returned an error (e.g., user denied access)
    const oauthError = params.get('error');
    const oauthErrorDesc = params.get('error_description');
    if (oauthError) {
      const msg = oauthErrorDesc || oauthError;
      setError(msg);

      // Notify parent window of the error
      try {
        if (window.opener && !window.opener.closed) {
          window.opener.postMessage(
            {
              type: 'oauth_callback',
              status: 'error',
              message: `Authorization denied: ${msg}`,
            },
            window.location.origin
          );
        }
      } catch (e) {
        console.error('postMessage failed:', e);
      }
      return;
    }

    const slug = params.get('slug');
    const oauthUrl = params.get('oauth_url');

    if (!slug || !oauthUrl) {
      setError('Missing required parameters. Please try connecting again.');
      return;
    }

    // URLSearchParams.get() already URL-decodes, so oauthUrl is the direct redirect URL
    window.location.href = oauthUrl;
  }, []);

  // Re-read params for rendering (after state updates)
  const params = new URLSearchParams(window.location.search);
  const slug = params.get('slug');
  const oauthUrl = params.get('oauth_url');

  // Error state
  if (error) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4">
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-8 max-w-md w-full text-center">
          <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-white mb-2">Authorization Failed</h2>
          <p className="text-gray-400 mb-4">{error}</p>
          <button
            onClick={() => window.close()}
            className="px-4 py-2 bg-gray-800 text-white rounded-md hover:bg-gray-700 transition-colors"
          >
            Close Window
          </button>
        </div>
      </div>
    );
  }

  // Missing params state
  if (!slug || !oauthUrl) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4">
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-8 max-w-md w-full text-center">
          <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-white mb-2">Error</h2>
          <p className="text-gray-400 mb-4">Missing required parameters. Please try connecting again.</p>
          <button
            onClick={() => window.close()}
            className="px-4 py-2 bg-gray-800 text-white rounded-md hover:bg-gray-700 transition-colors"
          >
            Close Window
          </button>
        </div>
      </div>
    );
  }

  // Loading/redirecting state
  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4">
      <div className="bg-gray-900 border border-gray-800 rounded-lg p-8 max-w-md w-full text-center">
        <Loader2 className="h-12 w-12 text-blue-500 mx-auto mb-4 animate-spin" />
        <h2 className="text-xl font-semibold text-white mb-2">Redirecting...</h2>
        <p className="text-gray-400">Please wait while we redirect you to the OAuth provider.</p>
        <p className="text-gray-500 text-sm mt-4">
          If you are not redirected,{' '}
          <a href={oauthUrl} className="text-blue-400 hover:underline">
            click here
          </a>
          .
        </p>
      </div>
    </div>
  );
}
