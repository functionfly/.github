import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { toast } from 'sonner';
import {
  Plus,
  Pencil,
  Trash2,
  Save,
  X,
  Gavel,
  Shield,
  FileText,
  AlertTriangle,
  Scale,
  GripVertical,
  Eye,
  EyeOff,
} from 'lucide-react';

interface Rule {
  id: string;
  title: string;
  description: string;
  category: string;
  enforcement: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

const CATEGORIES = [
  { value: 'conduct', label: 'Conduct', icon: Shield, color: 'text-green-600' },
  { value: 'content', label: 'Content', icon: FileText, color: 'text-blue-600' },
  { value: 'safety', label: 'Safety', icon: AlertTriangle, color: 'text-red-600' },
  { value: 'legal', label: 'Legal', icon: Scale, color: 'text-purple-600' },
  { value: 'moderation', label: 'Moderation', icon: Gavel, color: 'text-orange-600' },
];

const ENFORCEMENTS = [
  { value: 'info', label: 'Info', color: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' },
  { value: 'warning', label: 'Warning', color: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' },
  { value: 'deletion', label: 'Deletion', color: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' },
  { value: 'suspension', label: 'Suspension', color: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' },
];

function getCategoryMeta(cat: string) {
  return CATEGORIES.find((c) => c.value === cat) ?? CATEGORIES[0];
}

function getEnforcementMeta(enc: string) {
  return ENFORCEMENTS.find((e) => e.value === enc) ?? ENFORCEMENTS[0];
}

export function AdminCommunityRulesPage() {
  const queryClient = useQueryClient();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({
    title: '',
    description: '',
    category: 'conduct',
    enforcement: 'warning',
    sort_order: 0,
    is_active: true,
  });

  const { data, isLoading } = useQuery({
    queryKey: ['admin-community-rules'],
    queryFn: async () => {
      const res = await adminApiClient.get<{ rules: Rule[] }>('/community/rules');
      return res.data.rules;
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: typeof form) => adminApiClient.post('/community/rules', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-community-rules'] });
      toast.success('Rule created');
      setShowCreate(false);
      resetForm();
    },
    onError: () => toast.error('Failed to create rule'),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, ...data }: { id: string } & Partial<typeof form>) =>
      adminApiClient.put(`/community/rules/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-community-rules'] });
      toast.success('Rule updated');
      setEditingId(null);
    },
    onError: () => toast.error('Failed to update rule'),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => adminApiClient.delete(`/community/rules/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-community-rules'] });
      toast.success('Rule deleted');
    },
    onError: () => toast.error('Failed to delete rule'),
  });

  const resetForm = () =>
    setForm({ title: '', description: '', category: 'conduct', enforcement: 'warning', sort_order: 0, is_active: true });

  const startEdit = (rule: Rule) => {
    setEditingId(rule.id);
    setForm({
      title: rule.title,
      description: rule.description,
      category: rule.category,
      enforcement: rule.enforcement,
      sort_order: rule.sort_order,
      is_active: rule.is_active,
    });
  };

  if (isLoading) return <LoadingScreen />;

  const rules = data ?? [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-3">
            <Gavel className="w-8 h-8" />
            Community Rules
          </h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Manage community guidelines displayed to all users. Rules are grouped by category and shown in the forum sidebar.
          </p>
        </div>
        <button
          onClick={() => { resetForm(); setShowCreate(true); }}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-5 h-5" />
          Add Rule
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        {CATEGORIES.map((cat) => {
          const count = rules.filter((r) => r.category === cat.value).length;
          const Icon = cat.icon;
          return (
            <div key={cat.value} className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-800 p-4">
              <div className="flex items-center gap-2 mb-1">
                <Icon className={`w-4 h-4 ${cat.color}`} />
                <span className="text-sm font-medium text-gray-600 dark:text-gray-400">{cat.label}</span>
              </div>
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{count}</p>
            </div>
          );
        })}
      </div>

      {/* Create Form */}
      {showCreate && (
        <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-800 p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">New Rule</h2>
          <RuleForm
            form={form}
            setForm={setForm}
            onSave={() => createMutation.mutate(form)}
            onCancel={() => { setShowCreate(false); resetForm(); }}
            saving={createMutation.isPending}
          />
        </div>
      )}

      {/* Rules List */}
      <div className="space-y-3">
        {rules.map((rule) => {
          const catMeta = getCategoryMeta(rule.category);
          const encMeta = getEnforcementMeta(rule.enforcement);
          const CatIcon = catMeta.icon;
          const isEditing = editingId === rule.id;

          return (
            <div
              key={rule.id}
              className={`bg-white dark:bg-gray-900 rounded-lg shadow-sm border transition-all ${
                isEditing ? 'border-blue-300 dark:border-blue-700' : 'border-gray-200 dark:border-gray-800'
              } ${!rule.is_active ? 'opacity-60' : ''}`}
            >
              {isEditing ? (
                <div className="p-6">
                  <RuleForm
                    form={form}
                    setForm={setForm}
                    onSave={() => updateMutation.mutate({ id: rule.id, ...form })}
                    onCancel={() => { setEditingId(null); resetForm(); }}
                    saving={updateMutation.isPending}
                  />
                </div>
              ) : (
                <div className="p-4 flex items-start gap-4">
                  <div className="flex items-center gap-2 text-gray-400">
                    <GripVertical className="w-4 h-4" />
                    <span className="text-xs font-mono text-gray-500 w-6 text-right">{rule.sort_order}</span>
                  </div>
                  <div className={`mt-0.5 ${catMeta.color}`}>
                    <CatIcon className="w-5 h-5" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="font-semibold text-gray-900 dark:text-gray-100">{rule.title}</h3>
                      <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${encMeta.color}`}>
                        {encMeta.label}
                      </span>
                      {!rule.is_active && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-500">
                          <EyeOff className="w-3 h-3" /> Hidden
                        </span>
                      )}
                    </div>
                    {rule.description && (
                      <p className="mt-1 text-sm text-gray-600 dark:text-gray-400 line-clamp-2">
                        {rule.description}
                      </p>
                    )}
                  </div>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => updateMutation.mutate({ id: rule.id, is_active: !rule.is_active })}
                      className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-500"
                      title={rule.is_active ? 'Hide rule' : 'Show rule'}
                    >
                      {rule.is_active ? <Eye className="w-4 h-4" /> : <EyeOff className="w-4 h-4" />}
                    </button>
                    <button
                      onClick={() => startEdit(rule)}
                      className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-500"
                    >
                      <Pencil className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => { if (confirm('Delete this rule?')) deleteMutation.mutate(rule.id); }}
                      className="p-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors text-red-500"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>

      {rules.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          <Gavel className="w-12 h-12 mx-auto mb-4 opacity-30" />
          <p className="text-lg font-medium">No rules defined</p>
          <p className="text-sm">Add your first community rule to get started.</p>
        </div>
      )}
    </div>
  );
}

function RuleForm({
  form,
  setForm,
  onSave,
  onCancel,
  saving,
}: {
  form: { title: string; description: string; category: string; enforcement: string; sort_order: number; is_active: boolean };
  setForm: (f: typeof form) => void;
  onSave: () => void;
  onCancel: () => void;
  saving: boolean;
}) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Title</label>
          <input
            type="text"
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            placeholder="Rule title"
          />
        </div>
        <div className="grid grid-cols-3 gap-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Category</label>
            <select
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            >
              {CATEGORIES.map((c) => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Enforcement</label>
            <select
              value={form.enforcement}
              onChange={(e) => setForm({ ...form, enforcement: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            >
              {ENFORCEMENTS.map((e) => (
                <option key={e.value} value={e.value}>{e.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Order</label>
            <input
              type="number"
              value={form.sort_order}
              onChange={(e) => setForm({ ...form, sort_order: parseInt(e.target.value) || 0 })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            />
          </div>
        </div>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
        <textarea
          value={form.description}
          onChange={(e) => setForm({ ...form, description: e.target.value })}
          rows={3}
          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          placeholder="Detailed explanation of this rule..."
        />
      </div>
      <div className="flex items-center gap-3">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={form.is_active}
            onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
            className="w-4 h-4 rounded border-gray-300"
          />
          <span className="text-sm text-gray-700 dark:text-gray-300">Active (visible to users)</span>
        </label>
      </div>
      <div className="flex items-center gap-2 justify-end">
        <button
          onClick={onCancel}
          className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
        >
          <X className="w-4 h-4" /> Cancel
        </button>
        <button
          onClick={onSave}
          disabled={saving || !form.title.trim()}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
        >
          <Save className="w-4 h-4" /> {saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  );
}
