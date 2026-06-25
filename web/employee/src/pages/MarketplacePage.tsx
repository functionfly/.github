import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { marketplaceApi, type MarketplaceOpportunity } from '@/api/marketplace';
import { Store, Plus, Clock, Globe, MapPin, Briefcase, Filter, Send } from 'lucide-react';

const typeColors: Record<string, string> = {
  project: 'bg-blue-500/20 text-blue-400',
  gig: 'bg-purple-500/20 text-purple-400',
  mentorship: 'bg-green-500/20 text-green-400',
  hackathon: 'bg-orange-500/20 text-orange-400',
};

export function MarketplacePage() {
  const queryClient = useQueryClient();
  const [typeFilter, setTypeFilter] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showApply, setShowApply] = useState<string | null>(null);
  const [applyMessage, setApplyMessage] = useState('');
  const [form, setForm] = useState({
    title: '',
    description: '',
    opportunity_type: 'project',
    skills_required: '',
    hours_per_week: '',
    duration_weeks: '',
    is_remote: true,
    max_applicants: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['marketplace', 'opportunities', typeFilter],
    queryFn: () => marketplaceApi.list(typeFilter ? { type: typeFilter } : undefined),
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<MarketplaceOpportunity>) => marketplaceApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace'] });
      setShowCreate(false);
      setForm({ title: '', description: '', opportunity_type: 'project', skills_required: '', hours_per_week: '', duration_weeks: '', is_remote: true, max_applicants: '' });
    },
  });

  const applyMutation = useMutation({
    mutationFn: ({ id, message }: { id: string; message?: string }) => marketplaceApi.apply(id, message),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['marketplace'] });
      setShowApply(null);
      setApplyMessage('');
    },
  });

  const opportunities = data?.data?.opportunities || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Talent Marketplace</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Post Opportunity
        </button>
      </div>

      <div className="flex items-center gap-3">
        <Filter className="h-4 w-4 text-gray-400" />
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Types</option>
          <option value="project">Project</option>
          <option value="gig">Gig</option>
          <option value="mentorship">Mentorship</option>
          <option value="hackathon">Hackathon</option>
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : opportunities.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Store className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No opportunities available</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {opportunities.map((opp) => (
            <div key={opp.id} className="rounded-xl border border-gray-800 bg-gray-900 p-5">
              <div className="mb-3 flex items-start justify-between">
                <h3 className="font-semibold text-gray-100">{opp.title}</h3>
                <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${typeColors[opp.opportunity_type] || ''}`}>
                  {opp.opportunity_type}
                </span>
              </div>

              <p className="mb-3 line-clamp-2 text-sm text-gray-400">{opp.description}</p>

              {opp.skills_required.length > 0 && (
                <div className="mb-3 flex flex-wrap gap-1.5">
                  {opp.skills_required.map((skill) => (
                    <span key={skill} className="rounded bg-gray-800 px-2 py-0.5 text-xs text-gray-400">
                      {skill}
                    </span>
                  ))}
                </div>
              )}

              <div className="mb-4 flex flex-wrap items-center gap-3 text-xs text-gray-500">
                {opp.hours_per_week && (
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {opp.hours_per_week}h/wk
                  </span>
                )}
                {opp.duration_weeks && (
                  <span className="flex items-center gap-1">
                    <Briefcase className="h-3 w-3" />
                    {opp.duration_weeks} weeks
                  </span>
                )}
                {opp.is_remote && (
                  <span className="flex items-center gap-1 text-green-400">
                    <Globe className="h-3 w-3" />
                    Remote
                  </span>
                )}
                {!opp.is_remote && (
                  <span className="flex items-center gap-1">
                    <MapPin className="h-3 w-3" />
                    On-site
                  </span>
                )}
              </div>

              <button
                onClick={() => setShowApply(opp.id)}
                className="flex w-full items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
              >
                <Send className="h-4 w-4" />
                Apply
              </button>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Post Opportunity</h2>
            <input
              type="text"
              placeholder="Title"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <textarea
              placeholder="Description"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
            />
            <select
              value={form.opportunity_type}
              onChange={(e) => setForm({ ...form, opportunity_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="project">Project</option>
              <option value="gig">Gig</option>
              <option value="mentorship">Mentorship</option>
              <option value="hackathon">Hackathon</option>
            </select>
            <input
              type="text"
              placeholder="Skills required (comma separated)"
              value={form.skills_required}
              onChange={(e) => setForm({ ...form, skills_required: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <div className="mb-3 grid grid-cols-2 gap-3">
              <input
                type="number"
                placeholder="Hours/week"
                value={form.hours_per_week}
                onChange={(e) => setForm({ ...form, hours_per_week: e.target.value })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              />
              <input
                type="number"
                placeholder="Duration (weeks)"
                value={form.duration_weeks}
                onChange={(e) => setForm({ ...form, duration_weeks: e.target.value })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              />
            </div>
            <label className="mb-4 flex items-center gap-2 text-sm text-gray-300">
              <input
                type="checkbox"
                checked={form.is_remote}
                onChange={(e) => setForm({ ...form, is_remote: e.target.checked })}
                className="rounded border-gray-600 bg-gray-800"
              />
              Remote friendly
            </label>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowCreate(false)}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => createMutation.mutate({
                  title: form.title,
                  description: form.description,
                  opportunity_type: form.opportunity_type,
                  skills_required: form.skills_required ? form.skills_required.split(',').map((s) => s.trim()).filter(Boolean) : [],
                  hours_per_week: form.hours_per_week ? parseInt(form.hours_per_week) : undefined,
                  duration_weeks: form.duration_weeks ? parseInt(form.duration_weeks) : undefined,
                  is_remote: form.is_remote,
                  max_applicants: form.max_applicants ? parseInt(form.max_applicants) : undefined,
                })}
                disabled={!form.title.trim() || !form.description.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Post
              </button>
            </div>
          </div>
        </div>
      )}

      {showApply && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-sm rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Apply</h2>
            <textarea
              placeholder="Why are you a good fit? (optional)"
              value={applyMessage}
              onChange={(e) => setApplyMessage(e.target.value)}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={4}
            />
            <div className="flex justify-end gap-3">
              <button
                onClick={() => { setShowApply(null); setApplyMessage(''); }}
                className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={() => applyMutation.mutate({ id: showApply, message: applyMessage || undefined })}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
              >
                Submit Application
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
