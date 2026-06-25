import { Award, BadgeCheck, Medal, Banknote, GraduationCap } from 'lucide-react';

interface WalletItem {
  id: string;
  name: string;
  description?: string;
  icon?: string;
  earned_at?: string;
  category?: string;
}

interface WalletSectionProps {
  title: string;
  type: 'credentials' | 'badges' | 'awards' | 'grants' | 'training';
  items: WalletItem[];
  className?: string;
}

const typeIcons: Record<string, typeof Award> = {
  credentials: BadgeCheck,
  badges: Award,
  awards: Medal,
  grants: Banknote,
  training: GraduationCap,
};

const typeColors: Record<string, string> = {
  credentials: 'text-blue-400',
  badges: 'text-yellow-400',
  awards: 'text-purple-400',
  grants: 'text-green-400',
  training: 'text-cyan-400',
};

export function WalletSection({ title, type, items, className = '' }: WalletSectionProps) {
  const Icon = typeIcons[type] || Award;

  return (
    <div className={`rounded-xl border border-gray-800 bg-gray-900 p-5 ${className}`}>
      <h3 className="mb-4 flex items-center gap-2 text-sm font-medium text-gray-300">
        <Icon className={`h-4 w-4 ${typeColors[type] || 'text-gray-400'}`} />
        {title} ({items.length})
      </h3>

      {items.length === 0 ? (
        <p className="py-4 text-center text-sm text-gray-500">No {type} earned yet</p>
      ) : (
        <div className="space-y-2">
          {items.map((item) => (
            <div
              key={item.id}
              className="flex items-center gap-3 rounded-lg bg-gray-800/50 px-3 py-2.5"
            >
              <span className="text-lg">{item.icon || '🏅'}</span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-gray-100">{item.name}</p>
                {item.description && (
                  <p className="truncate text-xs text-gray-500">{item.description}</p>
                )}
              </div>
              {item.earned_at && (
                <span className="shrink-0 text-xs text-gray-500">
                  {new Date(item.earned_at).toLocaleDateString()}
                </span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
