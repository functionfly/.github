import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { identityApi } from '@/api/identity';
import { employeesApi } from '@/api/employees';
import { CareerTimeline } from '@/components/identity/CareerTimeline';
import { Clock, Filter } from 'lucide-react';
import type { TimelineEvent } from '@/api/identity';

const eventTypes = [
  { value: '', label: 'All Events' },
  { value: 'joined', label: 'Joined' },
  { value: 'promoted', label: 'Promoted' },
  { value: 'transferred', label: 'Transferred' },
  { value: 'project_completed', label: 'Project Completed' },
  { value: 'certification_earned', label: 'Certification Earned' },
  { value: 'achievement_unlocked', label: 'Achievement Unlocked' },
];

export function CareerTimelinePage() {
  const [filterType, setFilterType] = useState('');

  const { data: employeeData } = useQuery({
    queryKey: ['employee', 'me'],
    queryFn: () => employeesApi.list({ limit: 1 }).then((r) => ({ data: { employee: r.data.employees[0] } })),
  });

  const employee = employeeData?.data?.employee;
  const employeeId = employee?.id || '';

  const { data: timelineData, isLoading } = useQuery({
    queryKey: ['timeline', employeeId],
    queryFn: () => identityApi.getTimeline(employeeId),
    enabled: !!employeeId,
  });

  const events = timelineData?.data?.events || [];
  const filtered = filterType
    ? events.filter((e) => e.event_type === filterType)
    : events;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Clock className="h-6 w-6 text-cyan-400" />
          <h1 className="text-2xl font-bold">Career Timeline</h1>
        </div>
        <span className="text-sm text-gray-500">{events.length} events</span>
      </div>

      <div className="flex items-center gap-3">
        <Filter className="h-4 w-4 text-gray-500" />
        <div className="flex flex-wrap gap-2">
          {eventTypes.map((type) => (
            <button
              key={type.value}
              onClick={() => setFilterType(type.value)}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
                filterType === type.value
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {type.label}
            </button>
          ))}
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : (
        <CareerTimeline events={filtered} />
      )}
    </div>
  );
}
