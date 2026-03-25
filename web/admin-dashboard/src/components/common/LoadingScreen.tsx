/**
 * Loading Screen Component
 * Displays a full-screen loading indicator
 */

import { Loader2 } from 'lucide-react';

export function LoadingScreen() {
  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="flex flex-col items-center gap-4">
        <Loader2 className="w-10 h-10 animate-spin text-blue-600" />
        <p className="text-gray-600 dark:text-gray-400">Loading...</p>
      </div>
    </div>
  );
}

/**
 * Loading Spinner Component
 * Displays a smaller inline loading indicator
 */
export function LoadingSpinner({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
  const sizeClasses = {
    sm: 'w-4 h-4',
    md: 'w-6 h-6',
    lg: 'w-8 h-8',
  };

  return (
    <Loader2 className={`${sizeClasses[size]} animate-spin text-blue-600`} />
  );
}

/**
 * Loading Overlay Component
 * Displays a loading overlay on top of content
 */
export function LoadingOverlay({ message }: { message?: string }) {
  return (
    <div className="absolute inset-0 bg-white/80 dark:bg-gray-900/80 flex items-center justify-center z-10">
      <div className="flex flex-col items-center gap-3">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
        {message && <p className="text-gray-600 dark:text-gray-400">{message}</p>}
      </div>
    </div>
  );
}
