import {
  UserPlus,
  ArrowUpRight,
  ArrowRightLeft,
  FolderCheck,
  GraduationCap,
  Trophy,
  Calendar,
} from 'lucide-react';
import { formatDate } from '@/lib/utils';
import type { TimelineEvent } from '@/api/identity';

const eventIcons: Record<string, typeof UserPlus> = {
  joined: UserPlus,
  promoted: ArrowUpRight,
  transferred: ArrowRightLeft,
  project_completed: FolderCheck,
  certification_earned: GraduationCap,
  achievement_unlocked: Trophy,
};

const eventColors: Record<string, string> = {
  joined: 'bg-blue-500',
  promoted: 'bg-green-500',
  transferred: 'bg-purple-500',
  project_completed: 'bg-yellow-500',
  certification_earned: 'bg-cyan-500',
  achievement_unlocked: 'bg-orange-500',
};

interface CareerTimelineProps {
  events: TimelineEvent[];
  className?: string;
}

export function CareerTimeline({ events, className = '' }: CareerTimelineProps) {
  if (events.length === 0) {
    return (
      <div className={`rounded-xl border border-gray-800 bg-gray-900 p-6 ${className}`}>
        <p className="py-4 text-center text-sm text-gray-500">No career events yet</p>
      </div>
    );
  }

  return (
    <div className={`rounded-xl border border-gray-800 bg-gray-900 p-6 ${className}`}>
      <div className="relative">
        <div className="absolute left-4 top-0 bottom-0 w-px bg-gray-800" />

        <div className="space-y-6">
          {events.map((event) => {
            const Icon = eventIcons[event.event_type] || Trophy;
            const dotColor = eventColors[event.event_type] || 'bg-gray-500';

            return (
              <div key={event.id} className="relative flex gap-4 pl-10">
                <div className={`absolute left-2.5 top-1 h-3 w-3 rounded-full border-2 border-gray-900 ${dotColor}`} />
                <div className="flex-1 rounded-lg bg-gray-800/50 p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                      <Icon className="h-4 w-4 text-gray-400" />
                      <h4 className="text-sm font-semibold text-gray-100">{event.title}</h4>
                    </div>
                    <span className="flex items-center gap-1 text-xs text-gray-500">
                      <Calendar className="h-3 w-3" />
                      {formatDate(event.event_date)}
                    </span>
                  </div>
                  {event.description && (
                    <p className="mt-1 text-sm text-gray-400">{event.description}</p>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
