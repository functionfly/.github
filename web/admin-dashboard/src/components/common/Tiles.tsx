import React from 'react';

interface StatCardProps {
  label: string;
  value: number;
  icon: React.ElementType;
  color: 'blue' | 'green' | 'purple' | 'red';
}

export function StatCard({ label, value, icon: Icon, color }: StatCardProps) {
  const colorMap = {
    blue: { bg: 'rgba(59, 130, 246, 0.1)', text: '#3b82f6' },
    green: { bg: 'rgba(16, 185, 129, 0.1)', text: '#10b981' },
    purple: { bg: 'rgba(168, 85, 247, 0.1)', text: '#a855f7' },
    red: { bg: 'rgba(239, 68, 68, 0.1)', text: '#ef4444' },
  };

  return (
    <div className="rounded-lg p-6" style={{ backgroundColor: 'var(--bg-secondary)', boxShadow: 'var(--shadow-sm)', borderColor: 'var(--border-default)', borderWidth: '1px' }}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium" style={{ color: 'var(--text-secondary)' }}>{label}</p>
          <p className="mt-2 text-3xl font-bold" style={{ color: 'var(--text-primary)' }}>{value.toLocaleString()}</p>
        </div>
        <div className="p-3 rounded-lg" style={{ backgroundColor: colorMap[color].bg, color: colorMap[color].text }}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
    </div>
  );
}