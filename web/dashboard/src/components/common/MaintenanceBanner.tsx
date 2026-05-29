import { AlertTriangle, X } from 'lucide-react';
import { useState } from 'react';

interface MaintenanceBannerProps {
  message?: string;
}

export function MaintenanceBanner({ message }: MaintenanceBannerProps) {
  const [dismissed, setDismissed] = useState(false);

  if (dismissed) {
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
          <AlertTriangle className="h-4 w-4 flex-shrink-0" />
          <p className="text-sm font-medium">
            {message || 'System maintenance in progress. Some features may be temporarily unavailable.'}
          </p>
        </div>
        <button
          onClick={() => setDismissed(true)}
          className="p-1 hover:bg-amber-600/20 dark:hover:bg-amber-500/20 rounded transition-colors"
          aria-label="Dismiss maintenance banner"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}

export function useMaintenanceBanner() {
  const enabled = import.meta.env.VITE_MAINTENANCE_MODE === 'true';
  const message = import.meta.env.VITE_MAINTENANCE_MESSAGE as string | undefined;
  return { enabled, message };
}