import { useApiReachableStore } from '@/stores/apiReachableStore';
import { AlertCircle, X } from 'lucide-react';
import { useEffect, useState } from 'react';

/**
 * API Health Banner - Shows a persistent banner when the API is unreachable
 * This provides user feedback during API outages and allows users to dismiss the warning
 */
export function ApiHealthBanner() {
  const apiReachable = useApiReachableStore((state) => state.apiReachable);
  const [dismissed, setDismissed] = useState(false);
  const [showBanner, setShowBanner] = useState(false);

  useEffect(() => {
    // Only show banner if API is explicitly unreachable (not undefined/unknown)
    if (apiReachable === false && !dismissed) {
      setShowBanner(true);
    } else if (apiReachable === true) {
      // Reset dismissed state when API comes back online
      setDismissed(false);
      setShowBanner(false);
    }
  }, [apiReachable, dismissed]);

  const handleDismiss = () => {
    setDismissed(true);
    setShowBanner(false);
  };

  if (!showBanner) {
    return null;
  }

  return (
    <div
      className="fixed top-0 left-0 right-0 z-50 bg-amber-500/90 dark:bg-amber-600/90 text-amber-950 dark:text-amber-50 backdrop-blur-sm"
      role="alert"
      aria-live="polite"
    >
      <div className="container mx-auto px-4 py-2 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <AlertCircle className="h-4 w-4 flex-shrink-0" />
          <p className="text-sm font-medium">
            Connection issues detected. Some features may be unavailable.
          </p>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-xs opacity-75 hidden sm:inline">
            Retrying automatically...
          </span>
          <button
            onClick={handleDismiss}
            className="p-1 hover:bg-amber-600/20 dark:hover:bg-amber-500/20 rounded transition-colors"
            aria-label="Dismiss warning"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
