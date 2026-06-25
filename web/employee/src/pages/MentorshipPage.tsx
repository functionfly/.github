import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { mentorshipApi, type MentorshipMatch } from '@/api/mentorship';
import { HeartHandshake, Plus, Calendar, MessageSquare, Clock } from 'lucide-react';

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  pending: 'bg-yellow-500/20 text-yellow-400',
  completed: 'bg-blue-500/20 text-blue-400',
  paused: 'bg-gray-500/20 text-gray-400',
};

export function MentorshipPage() {
  const queryClient = useQueryClient();
  const [showRequest, setShowRequest] = useState(false);
  const [mentorId, setMentorId] = useState('');
  const [focusArea, setFocusArea] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['mentorship'],
    queryFn: () => mentorshipApi.list(),
  });

  const requestMutation = useMutation({
    mutationFn: () => mentorshipApi.request(mentorId, focusArea || undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mentorship'] });
      setShowRequest(false);
      setMentorId('');
      setFocusArea('');
    },
  });

  const matches = data?.data?.matches || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Mentorship</h1>
        <button
          onClick={() => setShowRequest(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Request Mentorship
        </button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : matches.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <HeartHandshake className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No mentorship matches yet</p>
          <p className="mt-1 text-sm text-gray-500">Request mentorship to get started</p>
        </div>
      ) : (
        <div className="space-y-4">
          {matches.map((match) => (
            <div key={match.id} className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-3 flex items-start justify-between">
                <div>
                  <h3 className="font-semibold text-gray-100">
                    Mentor: {match.mentor_id}
                  </h3>
                  {match.focus_area && (
                    <p className="mt-1 text-sm text-gray-400">Focus: {match.focus_area}</p>
                  )}
                </div>
                <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${statusColors[match.status] || ''}`}>
                  {match.status}
                </span>
              </div>

              <div className="flex flex-wrap items-center gap-4 text-xs text-gray-500">
                <span className="flex items-center gap-1">
                  <Calendar className="h-3 w-3" />
                  Started {new Date(match.started_at).toLocaleDateString()}
                </span>
                {match.meeting_frequency && (
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {match.meeting_frequency}
                  </span>
                )}
                {match.notes && (
                  <span className="flex items-center gap-1">
                    <MessageSquare className="h-3 w-3" />
                    Has notes
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showRequest && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Request Mentorship</h2>
            <input
              type="text"
              placeholder="Mentor ID"
              value={mentorId}
              onChange={(e) => setMentorId(e.target.value)}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <input
              type="text"
              placeholder="Focus area (optional)"
              value={focusArea}
              onChange={(e) => setFocusArea(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => { setShowRequest(false); setMentorId(''); setFocusArea(''); }}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => requestMutation.mutate()}
                disabled={!mentorId.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Send Request
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
