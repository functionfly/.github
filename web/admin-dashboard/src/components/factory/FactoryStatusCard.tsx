/**
 * Factory Status Card Component
 * Displays a single statistic with icon
 */

import clsx from 'clsx';
import type { LucideIcon } from 'lucide-react';

interface FactoryStatusCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  description?: string;
  status?: 'success' | 'warning' | 'error';
}

export function FactoryStatusCard({
  title,
  value,
  icon: Icon,
  description,
  status,
}: FactoryStatusCardProps) {
  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-5">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-sm text-gray-500 dark:text-gray-400">{title}</p>
          <p
            className={clsx(
              'text-2xl font-bold',
              status === 'success' && 'text-green-600 dark:text-green-400',
              status === 'warning' && 'text-yellow-600 dark:text-yellow-400',
              status === 'error' && 'text-red-600 dark:text-red-400',
              !status && 'text-gray-900 dark:text-gray-100'
            )}
          >
            {value}
          </p>
          {description && <p className="text-xs text-gray-500 dark:text-gray-400">{description}</p>}
        </div>
        <div
          className={clsx(
            'p-2 rounded-lg',
            status === 'success' && 'bg-green-100 dark:bg-green-900/50 text-green-600 dark:text-green-400',
            status === 'warning' && 'bg-yellow-100 dark:bg-yellow-900/50 text-yellow-600 dark:text-yellow-400',
            status === 'error' && 'bg-red-100 dark:bg-red-900/50 text-red-600 dark:text-red-400',
            !status && 'bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400'
          )}
        >
          <Icon className="h-5 w-5" />
        </div>
      </div>
    </div>
  );
}
