/**
 * PublishCardForm — Form to publish/update an A2A agent card.
 */

import { useState } from 'react';
import { Send, Loader2, CheckCircle2 } from 'lucide-react';
import type { PublishCardRequest, AgentSkill } from '../types';

export function PublishCardForm() {
  const [form, setForm] = useState<PublishCardRequest>({
    id: '',
    name: '',
    description: '',
    url: '',
    version: '1.0',
    protocol_version: '0.3.0',
    capabilities: [],
    skills: [],
    auth_schemes: ['bearer'],
    input_modes: ['application/json'],
    output_modes: ['application/json'],
  });
  const [skillInput, setSkillInput] = useState('');
  const [capInput, setCapInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      const res = await fetch('/api/v1/a2a/agents/cards', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data?.error?.message || 'Failed to publish agent card');
      }
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }

  function addCapability() {
    if (capInput.trim() && !form.capabilities?.includes(capInput.trim())) {
      setForm({ ...form, capabilities: [...(form.capabilities || []), capInput.trim()] });
      setCapInput('');
    }
  }

  function addSkill() {
    if (skillInput.trim()) {
      const skills = [...(form.skills || [])];
      skills.push({ id: skillInput.trim(), description: '' });
      setForm({ ...form, skills });
      setSkillInput('');
    }
  }

  return (
    <form onSubmit={handleSubmit} className="max-w-2xl space-y-6">
      {success && (
        <div className="flex items-center gap-2 p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-lg text-emerald-400 text-sm">
          <CheckCircle2 className="w-4 h-4" />
          Agent card published successfully
        </div>
      )}

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1.5">Agent ID</label>
          <input
            type="text"
            required
            value={form.id}
            onChange={(e) => setForm({ ...form, id: e.target.value })}
            placeholder="my-org/my-agent"
            className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1.5">Name</label>
          <input
            type="text"
            required
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            placeholder="My Agent"
            className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-text-primary mb-1.5">Description</label>
        <textarea
          value={form.description}
          onChange={(e) => setForm({ ...form, description: e.target.value })}
          placeholder="What does this agent do?"
          rows={3}
          className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
        />
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1.5">URL</label>
          <input
            type="url"
            value={form.url}
            onChange={(e) => setForm({ ...form, url: e.target.value })}
            placeholder="https://..."
            className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1.5">Version</label>
          <input
            type="text"
            value={form.version}
            onChange={(e) => setForm({ ...form, version: e.target.value })}
            className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-text-primary mb-1.5">JWKS URL (peer auth)</label>
          <input
            type="url"
            value={form.peer_jwks_url || ''}
            onChange={(e) => setForm({ ...form, peer_jwks_url: e.target.value })}
            placeholder="https://.../.well-known/jwks.json"
            className="w-full px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
          />
        </div>
      </div>

      {/* Capabilities */}
      <div>
        <label className="block text-sm font-medium text-text-primary mb-1.5">Capabilities</label>
        <div className="flex gap-2 mb-2">
          <input
            type="text"
            value={capInput}
            onChange={(e) => setCapInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addCapability())}
            placeholder="e.g. streaming"
            className="flex-1 px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
          />
          <button type="button" onClick={addCapability} className="px-3 py-2 bg-zinc-700 text-white rounded-lg hover:bg-zinc-600">
            Add
          </button>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {form.capabilities?.map((cap) => (
            <span key={cap} className="px-2 py-0.5 text-xs bg-emerald-500/10 text-emerald-300 rounded-full border border-emerald-500/20 flex items-center gap-1">
              {cap}
              <button type="button" onClick={() => setForm({ ...form, capabilities: form.capabilities?.filter((c) => c !== cap) })} className="hover:text-red-400">×</button>
            </span>
          ))}
        </div>
      </div>

      {/* Skills */}
      <div>
        <label className="block text-sm font-medium text-text-primary mb-1.5">Skills</label>
        <div className="flex gap-2 mb-2">
          <input
            type="text"
            value={skillInput}
            onChange={(e) => setSkillInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addSkill())}
            placeholder="e.g. summarize"
            className="flex-1 px-3 py-2 bg-bg-secondary border border-border-subtle rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:ring-2 focus:ring-brand-500/30"
          />
          <button type="button" onClick={addSkill} className="px-3 py-2 bg-zinc-700 text-white rounded-lg hover:bg-zinc-600">
            Add
          </button>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {form.skills?.map((skill, idx) => (
            <span key={idx} className="px-2 py-0.5 text-xs bg-blue-500/10 text-blue-300 rounded-full border border-blue-500/20 flex items-center gap-1">
              {skill.id}
              <button type="button" onClick={() => setForm({ ...form, skills: form.skills?.filter((_, i) => i !== idx) })} className="hover:text-red-400">×</button>
            </span>
          ))}
        </div>
      </div>

      <button
        type="submit"
        disabled={loading}
        className="flex items-center gap-2 px-6 py-2.5 bg-brand-500 text-white rounded-lg font-medium hover:bg-brand-600 disabled:opacity-50 transition-colors"
      >
        {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
        Publish Agent Card
      </button>
    </form>
  );
}
