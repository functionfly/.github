import './styles.css';

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
    <div className="tm-container">
      <div className="tm-form-container">
        <div className="tm-form-header">
          <div className="tm-form-breadcrumb">
            <Link to="/time-machine">Time Machine</Link>
            <span>/</span>
            <span>New Replay</span>
          </div>
          <h1 className="tm-form-title">Create New Replay</h1>
          <p className="tm-form-subtitle">
            Configure a replay job to re-execute past function calls against a new version
          </p>
        </div>

        <div className="tm-stepper">
          {STEPS.map((s, i) => (
            <div key={s.title} className="flex items-center">
              <button
                onClick={() => i < step && setStep(i)}
                disabled={i > step}
                className={cn(
                  'tm-step',
                  i === step ? 'tm-step-active' :
                  i < step ? 'tm-step-completed' : 'tm-step-pending'
                )}
              >
                {i < step ? <CheckCircle2 className="w-4 h-4" /> : s.icon}
                {s.title}
              </button>
              {i < STEPS.length - 1 && (
                <div className={cn('tm-step-separator', i < step && 'completed')} />
              )}
            </div>
          ))}
        </div>

        <div className="tm-form-card">
          <div className="tm-form-card-header">
            {STEPS[step].icon}
            {STEPS[step].title}
          </div>
          <div className="tm-form-card-content">
          {step === 0 && (
            <>
              <p className="tm-form-description">
                Select the function you want to replay executions for.
              </p>
              <div className="tm-form-field">
                <label className="tm-form-label">Function</label>
                <div className="relative">
                  <Input
                    className="tm-form-input"
                    placeholder="Search functions..."
                    value={functionSearch}
                    onChange={(e) => setFunctionSearch(e.target.value)}
                    onFocus={() => setFunctionDropdownOpen(true)}
                  />
                  {functionDropdownOpen && (
                    <div
                      className="absolute z-50 mt-1 w-full rounded-md border shadow-lg max-h-[300px] overflow-y-auto"
                      style={{
                        borderColor: 'var(--tm-panel-border)',
                        backgroundColor: 'var(--tm-panel-bg)',
                      }}
                    >
                      {filteredFunctions.length === 0 ? (
                        <div className="px-3 py-2 text-sm" style={{ color: 'var(--tm-panel-text-muted)' }}>No functions found.</div>
                      ) : (
                        filteredFunctions.map((fn) => (
                          <button
                            key={fn.id}
                            type="button"
                            className="w-full px-3 py-2 text-left text-sm transition-colors"
                            style={{ color: 'var(--tm-panel-text)' }}
                            onClick={() => {
                              updateField('functionId', fn.id);
                              setFunctionDropdownOpen(false);
                              setFunctionSearch(fn.name);
                            }}
                          >
                            <div className="font-medium">{fn.name}</div>
                            <div className="text-xs" style={{ opacity: 0.5 }}>{fn.id}</div>
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
              <p className="tm-form-description">
                Select the time window of executions to replay.
              </p>
              {limits && !limits.unlimited && (
                <div className="tm-form-alert tm-form-alert-info">
                  <AlertCircle />
                  Your plan ({displayName}) allows up to {limits.replay_window_hours}h replay window
                </div>
              )}
              {exceedsWindowLimit && (
                <div className="tm-form-alert tm-form-alert-error">
                  <AlertCircle />
                  Selected window ({windowHours}h) exceeds your plan limit of {limits!.replay_window_hours}h.
                  Reduce the time range or upgrade your plan.
                </div>
              )}
              <div className="tm-form-row">
                <div className="tm-form-field">
                  <label className="tm-form-label">Start</label>
                  <Input
                    className="tm-form-input"
                    type="datetime-local"
                    value={form.windowStart}
                    onChange={(e) => updateField('windowStart', e.target.value)}
                  />
                </div>
                <div className="tm-form-field">
                  <label className="tm-form-label">End</label>
                  <Input
                    className="tm-form-input"
                    type="datetime-local"
                    value={form.windowEnd}
                    onChange={(e) => updateField('windowEnd', e.target.value)}
                  />
                </div>
              </div>
              {windowHours > 0 && !exceedsWindowLimit && (
                <p className="tm-form-hint">
                  Selected window: {windowHours} hours
                </p>
              )}
            </>
          )}

          {step === 2 && (
            <>
              <p className="tm-form-description">
                Specify the target function version to replay executions against.
              </p>
              <div className="tm-form-field">
                <label className="tm-form-label">Target Version</label>
                <Select
                  value={form.targetVersionId}
                  onValueChange={(v) => updateField('targetVersionId', v)}
                >
                  <SelectTrigger className="tm-form-input">
                    <SelectValue placeholder="Select a version..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="latest">Latest</SelectItem>
                    <SelectItem value="stable">Stable</SelectItem>
                    <SelectItem value="previous">Previous</SelectItem>
                  </SelectContent>
                </Select>
                <p className="tm-form-hint">
                  Choose "Latest" to use the most recent version, or select a specific version.
                </p>
              </div>
            </>
          )}

          {step === 3 && (
            <>
              <p className="tm-form-description">
                Provide context for this replay and choose the reconciliation mode.
              </p>
              <div className="tm-form-field">
                <label className="tm-form-label">Reason *</label>
                <Textarea
                  className="tm-form-textarea"
                  placeholder="Why are you creating this replay? e.g. 'Fix incorrect pricing calculation introduced in v1.3.2'"
                  value={form.reason}
                  onChange={(e) => updateField('reason', e.target.value)}
                  rows={3}
                />
              </div>
              <div className="tm-form-field">
                <label className="tm-form-label">Incident URL (optional)</label>
                <Input
                  className="tm-form-input"
                  placeholder="https://status.example.com/incidents/..."
                  value={form.incidentUrl}
                  onChange={(e) => updateField('incidentUrl', e.target.value)}
                />
              </div>
              <div className="tm-form-field">
                <label className="tm-form-label">Reconciliation Mode</label>
                <Select
                  value={form.reconciliationMode}
                  onValueChange={(v) => updateField('reconciliationMode', v)}
                >
                  <SelectTrigger className="tm-form-input">
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
            <div>
              <p className="tm-form-description">
                Review your replay configuration before submitting.
              </p>
              <div className="tm-form-review">
                {[
                  { label: 'Function', value: form.functionId },
                  { label: 'Time Window', value: `${form.windowStart} → ${form.windowEnd} (${windowHours}h)` },
                  { label: 'Target Version', value: form.targetVersionId },
                  { label: 'Reason', value: form.reason },
                  { label: 'Incident URL', value: form.incidentUrl || '—' },
                  { label: 'Reconciliation Mode', value: form.reconciliationMode === 'dry_run' ? 'Dry Run' : 'Preview Only' },
                ].map((row) => (
                  <div key={row.label} className="tm-form-review-row">
                    <span className="tm-form-review-label">{row.label}</span>
                    <span className="tm-form-review-value">{row.value}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
          </div>
        </div>
      </div>

      {createReplay.isError && (
        <div className="tm-form-alert tm-form-alert-error">
          <AlertCircle />
          {createReplay.error?.message ?? 'Failed to create replay'}
        </div>
      )}

      <div className="tm-form-actions">
        <div className="tm-form-actions-left">
          <Button
            variant="outline"
            className="tm-button"
            onClick={() => (step === 0 ? navigate('/time-machine') : setStep((s) => s - 1))}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            {step === 0 ? 'Cancel' : 'Back'}
          </Button>
        </div>
        <div className="tm-form-actions-right">
          {step < STEPS.length - 1 ? (
            <Button className="tm-button" onClick={() => setStep((s) => s + 1)} disabled={!canProceed()}>
              Next
              <ArrowRight className="w-4 h-4 ml-2" />
            </Button>
          ) : (
            <Button className="tm-button" onClick={handleSubmit} disabled={createReplay.isPending}>
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
    </div>
  );
}
