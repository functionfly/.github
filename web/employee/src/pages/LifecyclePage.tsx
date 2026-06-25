import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { lifecycleApi, type LifecycleWorkflow, type LifecycleWorkflowInstance } from '@/api/lifecycle';
import { GitPullRequest, Plus, Clock, CheckCircle2, ChevronRight } from 'lucide-react';

const eventTypeColors: Record<string, string> = {
  hired: 'bg-green-500/20 text-green-400',
  promoted: 'bg-blue-500/20 text-blue-400',
  transferred: 'bg-purple-500/20 text-purple-400',
  on_leave: 'bg-yellow-500/20 text-yellow-400',
  returned: 'bg-cyan-500/20 text-cyan-400',
  resigned: 'bg-red-500/20 text-red-400',
  terminated: 'bg-red-500/20 text-red-400',
};

export function LifecyclePage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'events' | 'workflows'>('events');
  const [selectedWorkflow, setSelectedWorkflow] = useState<string | null>(null);
  const [selectedInstance, setSelectedInstance] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [stepNotes, setStepNotes] = useState('');
  const [form, setForm] = useState({ name: '', description: '', trigger_event: 'hired' });

  const { data: eventsData, isLoading: eventsLoading } = useQuery({
    queryKey: ['lifecycle-events'],
    queryFn: () => lifecycleApi.listEvents(),
  });

  const { data: workflowsData, isLoading: workflowsLoading } = useQuery({
    queryKey: ['lifecycle-workflows'],
    queryFn: () => lifecycleApi.listWorkflows(),
  });

  const { data: instanceData } = useQuery({
    queryKey: ['lifecycle-instance', selectedInstance],
    queryFn: () => lifecycleApi.getInstance(selectedInstance!),
    enabled: !!selectedInstance,
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<LifecycleWorkflow>) => lifecycleApi.createWorkflow(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lifecycle-workflows'] });
      setShowCreate(false);
      setForm({ name: '', description: '', trigger_event: 'hired' });
    },
  });

  const completeStepMutation = useMutation({
    mutationFn: ({ instanceId, stepIdx }: { instanceId: string; stepIdx: number }) =>
      lifecycleApi.completeStep(instanceId, stepIdx, stepNotes || undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lifecycle-instance'] });
      setStepNotes('');
    },
  });

  const events = eventsData?.data?.events || [];
  const workflows = workflowsData?.data?.workflows || [];
  const instance = instanceData?.data?.instance;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <GitPullRequest className="h-6 w-6 text-purple-400" />
          <h1 className="text-2xl font-bold">Employee Lifecycle</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          New Workflow
        </button>
      </div>

      <div className="flex gap-1 rounded-lg bg-gray-900 p-1">
        <button
          onClick={() => { setTab('events'); setSelectedInstance(null); }}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${tab === 'events' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'}`}
        >
          Events
        </button>
        <button
          onClick={() => { setTab('workflows'); setSelectedInstance(null); }}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${tab === 'workflows' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'}`}
        >
          Workflows
        </button>
      </div>

      {tab === 'events' && (
        <>
          {eventsLoading ? (
            <div className="flex justify-center py-12">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
            </div>
          ) : events.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <GitPullRequest className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No lifecycle events</p>
            </div>
          ) : (
            <div className="space-y-3">
              {events.map((ev) => (
                <div key={ev.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className={`rounded-full px-2 py-0.5 text-xs ${eventTypeColors[ev.event_type] || 'bg-gray-500/20 text-gray-400'}`}>
                        {ev.event_type}
                      </span>
                      <span className="text-sm text-gray-300">Employee: {ev.employee_id}</span>
                    </div>
                    <span className="text-xs text-gray-500">{new Date(ev.created_at).toLocaleString()}</span>
                  </div>
                  {ev.notes && <p className="mt-2 text-sm text-gray-400">{ev.notes}</p>}
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {tab === 'workflows' && (
        <>
          {workflowsLoading ? (
            <div className="flex justify-center py-12">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
            </div>
          ) : workflows.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <GitPullRequest className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No workflow templates</p>
            </div>
          ) : (
            <div className="space-y-3">
              {workflows.map((wf) => (
                <button
                  key={wf.id}
                  onClick={() => setSelectedWorkflow(wf.id === selectedWorkflow ? null : wf.id)}
                  className={`w-full rounded-xl border p-4 text-left transition-colors ${
                    wf.id === selectedWorkflow
                      ? 'border-blue-600 bg-gray-800'
                      : 'border-gray-800 bg-gray-900 hover:bg-gray-800'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-medium text-gray-100">{wf.name}</h3>
                      <div className="mt-1 flex items-center gap-2 text-xs text-gray-500">
                        <span>Trigger: {wf.trigger_event}</span>
                        <span>·</span>
                        <span>{wf.steps.length} steps</span>
                        <span>·</span>
                        <span className={wf.is_active ? 'text-green-400' : 'text-gray-500'}>{wf.is_active ? 'Active' : 'Inactive'}</span>
                      </div>
                      {wf.description && <p className="mt-1 text-sm text-gray-400">{wf.description}</p>}
                    </div>
                    <ChevronRight className={`h-4 w-4 text-gray-500 transition-transform ${wf.id === selectedWorkflow ? 'rotate-90' : ''}`} />
                  </div>
                  {wf.id === selectedWorkflow && (
                    <div className="mt-4 space-y-2 border-t border-gray-700 pt-4">
                      <h4 className="text-xs font-medium text-gray-400 uppercase">Steps</h4>
                      {wf.steps.map((step, i) => (
                        <div key={i} className="flex items-center gap-3 rounded-lg bg-gray-800 p-3">
                          <span className="flex h-6 w-6 items-center justify-center rounded-full bg-gray-700 text-xs font-medium text-gray-300">{i + 1}</span>
                          <div>
                            <p className="text-sm text-gray-200">{step.title}</p>
                            <p className="text-xs text-gray-500">{step.assignee_role} · Due in {step.due_days}d</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </button>
              ))}
            </div>
          )}

          {selectedInstance && instance && (
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-6 space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-100">Workflow Instance</h2>
                <span className={`rounded-full px-2 py-0.5 text-xs ${instance.status === 'completed' ? 'bg-green-500/20 text-green-400' : 'bg-blue-500/20 text-blue-400'}`}>{instance.status}</span>
              </div>
              <div className="grid grid-cols-3 gap-4 text-sm">
                <div><span className="text-gray-500">Employee:</span> <span className="text-gray-200">{instance.employee_id}</span></div>
                <div><span className="text-gray-500">Current Step:</span> <span className="text-gray-200">{instance.current_step + 1}</span></div>
                <div><span className="text-gray-500">Started:</span> <span className="text-gray-200">{new Date(instance.started_at).toLocaleDateString()}</span></div>
              </div>
              <div className="space-y-2">
                {instance.steps_status.map((step) => (
                  <div key={step.step_idx} className="flex items-center justify-between rounded-lg bg-gray-800 p-3">
                    <div className="flex items-center gap-3">
                      {step.status === 'completed' ? (
                        <CheckCircle2 className="h-5 w-5 text-green-400" />
                      ) : (
                        <Clock className="h-5 w-5 text-yellow-400" />
                      )}
                      <div>
                        <p className="text-sm text-gray-200">Step {step.step_idx + 1}</p>
                        {step.completed_at && <p className="text-xs text-gray-500">Completed: {new Date(step.completed_at).toLocaleString()}</p>}
                        {step.notes && <p className="text-xs text-gray-400">{step.notes}</p>}
                      </div>
                    </div>
                    {step.status !== 'completed' && step.step_idx === instance.current_step && (
                      <button
                        onClick={() => completeStepMutation.mutate({ instanceId: instance.id, stepIdx: step.step_idx })}
                        className="rounded-lg bg-green-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-700"
                      >
                        Complete
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">New Workflow Template</h2>
            <input
              type="text"
              placeholder="Workflow name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <select
              value={form.trigger_event}
              onChange={(e) => setForm({ ...form, trigger_event: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="hired">Hired</option>
              <option value="promoted">Promoted</option>
              <option value="transferred">Transferred</option>
              <option value="resigned">Resigned</option>
            </select>
            <textarea
              placeholder="Description (optional)"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={3}
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowCreate(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => createMutation.mutate({ name: form.name, description: form.description || undefined, trigger_event: form.trigger_event, steps: [] })}
                disabled={!form.name.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
