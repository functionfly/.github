import React from 'react';

interface StatCardProps {
  label: string;
  value: number;
  icon: React.ElementType;
  color: 'blue' | 'green' | 'purple' | 'red';
}

export function StatCard({ label, value, icon: Icon, color }: StatCardProps) {
  const colorMap = {
    blue: { bg: 'bg-blue-500/10', text: 'text-blue-500', border: 'border-blue-500/20' },
    green: { bg: 'bg-green-500/10', text: 'text-green-500', border: 'border-green-500/20' },
    purple: { bg: 'bg-purple-500/10', text: 'text-purple-500', border: 'border-purple-500/20' },
    red: { bg: 'bg-red-500/10', text: 'text-red-500', border: 'border-red-500/20' },
  };

  const styles = colorMap[color];

  return (
    <div className={`rounded-lg p-6 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 shadow-sm`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">{label}</p>
          <p className="mt-2 text-3xl font-bold text-gray-900 dark:text-white">{value.toLocaleString()}</p>
        </div>
        <div className={`p-3 rounded-lg ${styles.bg}`}>
          <Icon className={`w-6 h-6 ${styles.text}`} />
        </div>
      </div>
    </div>
  );
}