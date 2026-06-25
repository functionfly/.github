import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { orgchartApi, type Goal } from '@/api/orgchart';
import { GitMerge, Plus, ChevronRight, ChevronDown, Target } from 'lucide-react';

const visibilityColors: Record<string, string> = {
  public: 'bg-green-500/20 text-green-400',
  department: 'bg-blue-500/20 text-blue-400',
  team: 'bg-purple-500/20 text-purple-400',
  private: 'bg-gray-500/20 text-gray-400',
};

const levelLabels: Record<string, string> = {
  company: 'Company',
  department: 'Department',
  team: 'Team',
  personal: 'Personal',
};

type GoalNode = Goal & { children?: GoalNode[] };

function GoalTree({ goals, depth = 0 }: { goals: GoalNode[]; depth?: number }) {
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  return (
    <div className={depth > 0 ? 'ml-6 border-l border-gray-800 pl-4' : ''}>
      {goals.map((goal) => {
        const hasChildren = goal.children && goal.children.length > 0;
        const isExpanded = expanded[goal.id] !== false;
        return (
          <div key={goal.id} className="py-2">
            <button
              onClick={() => setExpanded({ ...expanded, [goal.id]: !isExpanded })}
              className="flex w-full items-center gap-3 rounded-lg p-3 text-left transition-colors hover:bg-gray-800/50"
            >
              {hasChildren ? (
                isExpanded ? (
                  <ChevronDown className="h-4 w-4 flex-shrink-0 text-gray-500" />
                ) : (
                  <ChevronRight className="h-4 w-4 flex-shrink-0 text-gray-500" />
                )
              ) : (
                <div className="w-4" />
              )}
              <Target className={`h-4 w-4 flex-shrink-0 ${goal.level === 'company' ? 'text-blue-400' : goal.level === 'department' ? 'text-purple-400' : goal.level === 'team' ? 'text-cyan-400' : 'text-green-400'}`} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <h3 className="truncate font-medium text-gray-100">{goal.title}</h3>
                  <span className={`rounded-full px-2 py-0.5 text-xs ${visibilityColors[goal.visibility] || ''}`}>{goal.visibility}</span>
                  <span className="rounded bg-gray-700 px-1.5 py-0.5 text-xs text-gray-400">{levelLabels[goal.level] || goal.level}</span>
                </div>
                {goal.description && <p className="mt-0.5 truncate text-xs text-gray-500">{goal.description}</p>}
                <div className="mt-2 flex items-center gap-2">
                  <div className="h-1.5 w-32 overflow-hidden rounded-full bg-gray-700">
                    <div
                      className={`h-full rounded-full ${goal.progress >= 80 ? 'bg-green-500' : goal.progress >= 50 ? 'bg-blue-500' : goal.progress >= 25 ? 'bg-yellow-500' : 'bg-red-500'}`}
                      style={{ width: `${goal.progress}%` }}
                    />
                  </div>
                  <span className="text-xs text-gray-500">{goal.progress}%</span>
                </div>
              </div>
            </button>
            {hasChildren && isExpanded && <GoalTree goals={goal.children!} depth={depth + 1} />}
          </div>
        );
      })}
    </div>
  );
}

export function GoalCascadePage() {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({
    title: '',
    description: '',
    level: 'personal',
    visibility: 'team',
    parent_id: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['goals'],
    queryFn: () => orgchartApi.getGoals(),
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<Goal>) => orgchartApi.createGoal(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      setShowCreate(false);
      setForm({ title: '', description: '', level: 'personal', visibility: 'team', parent_id: '' });
    },
  });

  const allGoals = data?.data?.goals || [];

  function buildTree(goals: GoalNode[]): GoalNode[] {
    const map = new Map<string, GoalNode>();
    const roots: GoalNode[] = [];
    goals.forEach((g) => map.set(g.id, { ...g, children: [] }));
    goals.forEach((g) => {
      const node = map.get(g.id)!;
      if (g.parent_id && map.has(g.parent_id)) {
        map.get(g.parent_id)!.children!.push(node);
      } else {
        roots.push(node);
      }
    });
    return roots;
  }

  const tree = buildTree(allGoals);

  const stats = {
    total: allGoals.length,
    company: allGoals.filter((g) => g.level === 'company').length,
    department: allGoals.filter((g) => g.level === 'department').length,
    team: allGoals.filter((g) => g.level === 'team').length,
    personal: allGoals.filter((g) => g.level === 'personal').length,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <GitMerge className="h-6 w-6 text-purple-400" />
          <h1 className="text-2xl font-bold">Goal Cascade</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700"
        >
          <Plus className="h-4 w-4" />
          Create Goal
        </button>
      </div>

      <div className="grid grid-cols-5 gap-3">
        {[
          { label: 'Total', count: stats.total, color: 'text-gray-100' },
          { label: 'Company', count: stats.company, color: 'text-blue-400' },
          { label: 'Department', count: stats.department, color: 'text-purple-400' },
          { label: 'Team', count: stats.team, color: 'text-cyan-400' },
          { label: 'Personal', count: stats.personal, color: 'text-green-400' },
        ].map((s) => (
          <div key={s.label} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
            <span className="text-sm text-gray-400">{s.label}</span>
            <p className={`mt-1 text-2xl font-bold ${s.color}`}>{s.count}</p>
          </div>
        ))}
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : tree.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <GitMerge className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No goals defined yet</p>
        </div>
      ) : (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          <GoalTree goals={tree} />
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Create Goal</h2>
            <input
              type="text"
              placeholder="Goal title"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <textarea
              placeholder="Description (optional)"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={2}
            />
            <select
              value={form.level}
              onChange={(e) => setForm({ ...form, level: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="company">Company</option>
              <option value="department">Department</option>
              <option value="team">Team</option>
              <option value="personal">Personal</option>
            </select>
            <select
              value={form.visibility}
              onChange={(e) => setForm({ ...form, visibility: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="public">Public</option>
              <option value="department">Department</option>
              <option value="team">Team</option>
              <option value="private">Private</option>
            </select>
            <select
              value={form.parent_id}
              onChange={(e) => setForm({ ...form, parent_id: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="">No parent (root goal)</option>
              {allGoals.map((g) => (
                <option key={g.id} value={g.id}>{g.title}</option>
              ))}
            </select>
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowCreate(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => createMutation.mutate({
                  title: form.title,
                  description: form.description || undefined,
                  level: form.level,
                  visibility: form.visibility,
                  parent_id: form.parent_id || undefined,
                } as Partial<Goal>)}
                disabled={!form.title.trim()}
                className="rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700 disabled:opacity-50"
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
