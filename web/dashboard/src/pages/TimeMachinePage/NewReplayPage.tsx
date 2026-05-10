import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  ArrowRight,
  History,
  Clock,
  GitBranch,
  Settings,
  CheckCircle2,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useCreateReplay, useTimeMachineLimits } from '@/hooks/useTimeMachine';
import { useFunctions } from '@/hooks/useFunctions';
import { usePlan } from '@/hooks/usePlan';
import { cn } from '@/lib/utils';

interface FormData {
  functionId: string;
  windowStart: string;
  windowEnd: string;
  targetVersionId: string;
  reason: string;
  incidentUrl: string;
  reconciliationMode: string;
}

const INITIAL_FORM: FormData = {
  functionId: '',
  windowStart: '',
  windowEnd: '',
  targetVersionId: '',
  reason: '',
  incidentUrl: '',
  reconciliationMode: 'dry_run',
};

const STEPS = [
  { title: 'Function', icon: <GitBranch className="w-4 h-4" /> },
  { title: 'Time Window', icon: <Clock className="w-4 h-4" /> },
  { title: 'Target Version', icon: <GitBranch className="w-4 h-4" /> },
  { title: 'Configuration', icon: <Settings className="w-4 h-4" /> },
  { title: 'Review', icon: <CheckCircle2 className="w-4 h-4" /> },
];

export function NewReplayPage() {
  const navigate = useNavigate();
  const [step, setStep] = useState(0);
  const [form, setForm] = useState<FormData>(INITIAL_FORM);
  const [functionSearch, setFunctionSearch] = useState('');
  const [functionDropdownOpen, setFunctionDropdownOpen] = useState(false);
  const createReplay = useCreateReplay();
  const { data: limits } = useTimeMachineLimits();
  const { displayName } = usePlan();
  const { data: functions } = useFunctions();

  const filteredFunctions = functions?.functions?.filter((fn) =>
    fn.name.toLowerCase().includes(functionSearch.toLowerCase())
  ) ?? [];

  const updateField = <K extends keyof FormData>(key: K, value: FormData[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const windowHours = (() => {
    if (!form.windowStart || !form.windowEnd) return 0;
    const ms = new Date(form.windowEnd).getTime() - new Date(form.windowStart).getTime();
    return Math.round((ms / (1000 * 60 * 60)) * 10) / 10;
  })();

  const exceedsWindowLimit = limits && !limits.unlimited && windowHours > limits.replay_window_hours;

  const canProceed = () => {
    switch (step) {
      case 0:
        return form.functionId.trim().length > 0;
      case 1:
        return form.windowStart.length > 0 && form.windowEnd.length > 0 && windowHours > 0 && !exceedsWindowLimit;
      case 2:
        return form.targetVersionId.trim().length > 0;
      case 3:
        return form.reason.trim().length > 0;
      default:
        return true;
    }
  };

  const handleSubmit = () => {
    createReplay.mutate(
      {
        function_id: form.functionId,
        window_start: new Date(form.windowStart).toISOString(),
        window_end: new Date(form.windowEnd).toISOString(),
        target_version_id: form.targetVersionId,
        reason: form.reason,
        incident_url: form.incidentUrl || undefined,
        reconciliation_mode: form.reconciliationMode,
      },
      {
        onSuccess: (job) => {
          navigate(`/time-machine/${job.id}`);
        },
      }
    );
  };

  return (
    <div className="space-y-6 max-w-3xl mx-auto">
      <div className="flex items-center gap-2 text-sm text-text-secondary">
        <Link to="/time-machine" className="hover:text-text-primary transition-colors">
          Time Machine
        </Link>
        <span>/</span>
        <span className="text-text-primary">New Replay</span>
      </div>

      <div>
        <h1 className="text-2xl font-bold text-text-primary tracking-tight">
          Create New Replay
        </h1>
        <p className="text-text-secondary mt-1">
          Configure a replay job to re-execute past function calls against a new version
        </p>
      </div>

      <div className="flex items-center gap-1">
        {STEPS.map((s, i) => (
          <div key={s.title} className="flex items-center gap-1">
            <button
              onClick={() => i < step && setStep(i)}
              disabled={i > step}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-colors',
                i === step
                  ? 'bg-brand-500/15 text-brand-500 border border-brand-500/30'
                  : i < step
                    ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 cursor-pointer'
                    : 'bg-bg-secondary text-text-muted border border-white/10 cursor-not-allowed'
              )}
            >
              {i < step ? <CheckCircle2 className="w-3.5 h-3.5" /> : s.icon}
              {s.title}
            </button>
            {i < STEPS.length - 1 && (
              <div className={cn('w-6 h-px', i < step ? 'bg-emerald-500/40' : 'bg-white/10')} />
            )}
          </div>
        ))}
      </div>

      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            {STEPS[step].icon}
            {STEPS[step].title}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {step === 0 && (
            <>
              <p className="text-sm text-text-secondary">
                Select the function you want to replay executions for.
              </p>
              <div className="space-y-2">
                <label className="text-sm font-medium text-text-primary">Function</label>
                <div className="relative">
                  <Input
                    placeholder="Search functions..."
                    value={functionSearch}
                    onChange={(e) => setFunctionSearch(e.target.value)}
                    onFocus={() => setFunctionDropdownOpen(true)}
                  />
                  {functionDropdownOpen && (
                    <div
                      className="absolute z-50 mt-1 w-full rounded-md border bg-card shadow-lg max-h-[300px] overflow-y-auto"
                      style={{
                        borderColor: 'var(--ff-border-default)',
                        backgroundColor: 'var(--ff-bg-secondary)',
                      }}
                    >
                      {filteredFunctions.length === 0 ? (
                        <div className="px-3 py-2 text-sm text-text-muted">No functions found.</div>
                      ) : (
                        filteredFunctions.map((fn) => (
                          <button
                            key={fn.id}
                            type="button"
                            className="w-full px-3 py-2 text-left text-sm hover:bg-accent transition-colors"
                            style={{ color: 'var(--ff-text-primary)' }}
                            onClick={() => {
                              updateField('functionId', fn.id);
                              setFunctionDropdownOpen(false);
                              setFunctionSearch(fn.name);
                            }}
                          >
                            <div className="font-medium">{fn.name}</div>
                            <div className="text-xs opacity-50">{fn.id}</div>
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
              </div>
            </>
          )}

          {step === 1 && (
            <>
              <p className="text-sm text-text-secondary">
                Select the time window of executions to replay.
              </p>
              {limits && !limits.unlimited && (
                <div className="flex items-center gap-2 p-3 rounded-lg bg-blue-500/10 border border-blue-500/20 text-sm text-blue-400">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  Your plan ({displayName}) allows up to {limits.replay_window_hours}h replay window
                </div>
              )}
              {exceedsWindowLimit && (
                <div className="flex items-center gap-2 p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-sm text-red-400">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  Selected window ({windowHours}h) exceeds your plan limit of {limits!.replay_window_hours}h.
                  Reduce the time range or upgrade your plan.
                </div>
              )}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium text-text-primary">Start</label>
                  <Input
                    type="datetime-local"
                    value={form.windowStart}
                    onChange={(e) => updateField('windowStart', e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium text-text-primary">End</label>
                  <Input
                    type="datetime-local"
                    value={form.windowEnd}
                    onChange={(e) => updateField('windowEnd', e.target.value)}
                  />
                </div>
              </div>
              {windowHours > 0 && !exceedsWindowLimit && (
                <p className="text-xs text-text-muted">
                  Selected window: {windowHours} hours
                </p>
              )}
            </>
          )}

          {step === 2 && (
            <>
              <p className="text-sm text-text-secondary">
                Specify the target function version to replay executions against.
              </p>
              <div className="space-y-2">
                <label className="text-sm font-medium text-text-primary">Target Version</label>
                <Select
                  value={form.targetVersionId}
                  onValueChange={(v) => updateField('targetVersionId', v)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select a version..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="latest">Latest</SelectItem>
                    <SelectItem value="stable">Stable</SelectItem>
                    <SelectItem value="previous">Previous</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-text-muted">
                  Choose "Latest" to use the most recent version, or select a specific version.
                </p>
              </div>
            </>
          )}

          {step === 3 && (
            <>
              <p className="text-sm text-text-secondary">
                Provide context for this replay and choose the reconciliation mode.
              </p>
              <div className="space-y-2">
                <label className="text-sm font-medium text-text-primary">Reason *</label>
                <Textarea
                  placeholder="Why are you creating this replay? e.g. 'Fix incorrect pricing calculation introduced in v1.3.2'"
                  value={form.reason}
                  onChange={(e) => updateField('reason', e.target.value)}
                  rows={3}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-text-primary">Incident URL (optional)</label>
                <Input
                  placeholder="https://status.example.com/incidents/..."
                  value={form.incidentUrl}
                  onChange={(e) => updateField('incidentUrl', e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-text-primary">Reconciliation Mode</label>
                <Select
                  value={form.reconciliationMode}
                  onValueChange={(v) => updateField('reconciliationMode', v)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="dry_run">Dry Run — preview changes only</SelectItem>
                    <SelectItem value="preview_only">Preview Only — generate diff report</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </>
          )}

          {step === 4 && (
            <div className="space-y-4">
              <p className="text-sm text-text-secondary">
                Review your replay configuration before submitting.
              </p>
              <div className="space-y-3 rounded-lg border border-white/10 divide-y divide-white/10">
                {[
                  { label: 'Function', value: form.functionId },
                  { label: 'Time Window', value: `${form.windowStart} → ${form.windowEnd} (${windowHours}h)` },
                  { label: 'Target Version', value: form.targetVersionId },
                  { label: 'Reason', value: form.reason },
                  { label: 'Incident URL', value: form.incidentUrl || '—' },
                  { label: 'Reconciliation Mode', value: form.reconciliationMode === 'dry_run' ? 'Dry Run' : 'Preview Only' },
                ].map((row) => (
                  <div key={row.label} className="flex items-start justify-between gap-4 py-3 px-4 first:pt-0 last:pb-0">
                    <span className="text-sm text-text-secondary shrink-0">{row.label}</span>
                    <span className="text-sm text-text-primary text-right break-all">{row.value}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {createReplay.isError && (
        <div className="flex items-center gap-2 p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-sm text-red-400">
          <AlertCircle className="w-4 h-4 flex-shrink-0" />
          {createReplay.error?.message ?? 'Failed to create replay'}
        </div>
      )}

      <div className="flex items-center justify-between">
        <Button
          variant="outline"
          onClick={() => (step === 0 ? navigate('/time-machine') : setStep((s) => s - 1))}
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          {step === 0 ? 'Cancel' : 'Back'}
        </Button>
        {step < STEPS.length - 1 ? (
          <Button onClick={() => setStep((s) => s + 1)} disabled={!canProceed()}>
            Next
            <ArrowRight className="w-4 h-4 ml-2" />
          </Button>
        ) : (
          <Button onClick={handleSubmit} disabled={createReplay.isPending}>
            {createReplay.isPending ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <History className="w-4 h-4 mr-2" />
            )}
            Create Replay
          </Button>
        )}
      </div>
    </div>
  );
}
