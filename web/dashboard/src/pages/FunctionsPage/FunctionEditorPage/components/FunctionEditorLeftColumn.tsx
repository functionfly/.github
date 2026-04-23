import { ProviderIcon } from '@/components/common/ProviderIcon';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import {
  Check,
  Clock,
  Code2,
  Eye,
  EyeOff,
  Globe,
  Hash,
  Key,
  Link2,
  Lock,
  Plus,
  Settings2,
  Shield,
  Timer,
  Webhook,
  X,
  Zap,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import {
  HTTP_METHODS,
  MEMORY_OPTIONS,
  RUNTIME_META,
  RUNTIME_VERSIONS,
  TIMEOUT_OPTIONS,
} from '../constants';
import type { HttpMethod, Runtime } from '../types';
import type { FunctionEditorModel } from '../useFunctionEditor';
import { formatTimeout } from '../utils';
import { FieldError, InfoTip, SectionCard } from './editor-ui';

type Props = { editor: FunctionEditorModel };

export function FunctionEditorLeftColumn({ editor }: Props) {
  const {
    providers,
    functionName,
    slug,
    description,
    setDescription,
    runtime,
    runtimeVersion,
    setRuntimeVersion,
    visibility,
    setVisibility,
    tags,
    newTag,
    setNewTag,
    selectedProviders,
    selectedRegion,
    setSelectedRegion,
    envVars,
    newEnvKey,
    setNewEnvKey,
    newEnvValue,
    setNewEnvValue,
    isNewEnvSecret,
    setIsNewEnvSecret,
    showEnvValues,
    setShowEnvValues,
    resources,
    setResources,
    httpTrigger,
    setHttpTrigger,
    scheduleTrigger,
    setScheduleTrigger,
    retryPolicy,
    setRetryPolicy,
    errors,
    tagInputRef,
    handleNameChange,
    handleSlugChange,
    handleRuntimeChange,
    handleProviderToggle,
    addEnvironmentVariable,
    removeEnvironmentVariable,
    addTag,
    removeTag,
    markDirty,
    setVaultPickerOpen,
  } = editor;

  return (
    <div className="space-y-5">
      <SectionCard icon={<Zap className="w-4 h-4" />} title="Function Basics" step={1}>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <Label
              htmlFor="fn-name"
              className="text-xs font-medium text-text-secondary mb-1.5 block"
            >
              Name <span className="text-red-400">*</span>
            </Label>
            <Input
              id="fn-name"
              placeholder="my-awesome-function"
              value={functionName}
              onChange={(e) => handleNameChange(e.target.value)}
              className="input"
              aria-describedby={errors.name ? 'fn-name-error' : undefined}
            />
            <FieldError message={errors.name} />
          </div>
          <div>
            <Label
              htmlFor="fn-slug"
              className="text-xs font-medium text-text-secondary mb-1.5 block"
            >
              Slug / Identifier
              <InfoTip content="URL-safe identifier used in API calls. Auto-generated from name." />
            </Label>
            <div className="input-with-icon">
              <Hash className="icon h-3.5 w-3.5 shrink-0" aria-hidden />
              <Input
                id="fn-slug"
                placeholder="my-awesome-function"
                value={slug}
                onChange={(e) => handleSlugChange(e.target.value)}
                className="input w-full font-mono text-sm"
              />
            </div>
            <FieldError message={errors.slug} />
          </div>
        </div>
        <div>
          <Label htmlFor="fn-desc" className="text-xs font-medium text-text-secondary mb-1.5 block">
            Description
          </Label>
          <Textarea
            id="fn-desc"
            placeholder="What does this function do?"
            value={description}
            onChange={(e) => {
              setDescription(e.target.value);
              markDirty();
            }}
            className="input resize-none"
            rows={2}
          />
        </div>
      </SectionCard>

      <SectionCard icon={<Code2 className="w-4 h-4" />} title="Runtime Configuration" step={2}>
        <div>
          <Label className="text-xs font-medium text-text-secondary mb-2 block">Runtime</Label>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
            {(Object.entries(RUNTIME_META) as [Runtime, (typeof RUNTIME_META)[Runtime]][]).map(
              ([key, meta]) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => handleRuntimeChange(key)}
                  className={`relative flex flex-col gap-1.5 rounded-lg border-2 p-3 text-left transition-all duration-200 ${
                    runtime === key
                      ? 'border-[#FF6B35] bg-[#FFF1EB] shadow-sm dark:border-[#FF8C42]/90 dark:bg-[#FF6B35]/25 dark:shadow-[0_0_0_1px_rgba(255,140,66,0.35)]'
                      : 'border-transparent bg-bg-tertiary hover:border-border-default hover:bg-bg-hover'
                  }`}
                  aria-pressed={runtime === key}
                >
                  {runtime === key ? (
                    <span
                      className="absolute right-2 top-2 flex h-6 w-6 items-center justify-center rounded-full bg-[#FF6B35] text-white shadow-sm dark:bg-[#FF8C42]"
                      aria-hidden
                    >
                      <Check className="h-3.5 w-3.5" strokeWidth={2.5} />
                    </span>
                  ) : null}
                  <div className="flex items-center gap-2 pr-7">
                    <span
                      className="h-2 w-2 shrink-0 rounded-full"
                      style={{ background: meta.color }}
                    />
                    <span className="text-sm font-semibold text-text-primary">{meta.label}</span>
                  </div>
                  <span className="text-xs text-text-muted leading-relaxed">
                    {meta.description}
                  </span>
                </button>
              )
            )}
          </div>
        </div>
        <div className="w-40">
          <Label
            htmlFor="runtime-version"
            className="text-xs font-medium text-text-secondary mb-1.5 block"
          >
            Version
          </Label>
          <Select
            value={runtimeVersion}
            onValueChange={(v) => {
              setRuntimeVersion(v);
              markDirty();
            }}
          >
            <SelectTrigger id="runtime-version">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {RUNTIME_VERSIONS[runtime].map((v) => (
                <SelectItem key={v} value={v}>
                  {v}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </SectionCard>

      <SectionCard
        icon={<Settings2 className="w-4 h-4" />}
        title="Deployment Target"
        step={3}
        description="Choose where your function runs"
      >
        {providers.length === 0 ? (
          <div className="p-4 rounded-lg border border-border-subtle bg-bg-tertiary/50">
            <p className="text-sm text-text-secondary mb-3">
              Connect a provider to deploy. You need at least one (e.g. Cloudflare Workers, Vercel).
            </p>
            <Button variant="outline" size="sm" asChild>
              <Link to="/providers" className="gap-2">
                <Link2 className="w-4 h-4" />
                Connect a provider
              </Link>
            </Button>
          </div>
        ) : (
          <div className="flex flex-wrap gap-2">
            {providers.map((provider) => (
              <button
                key={provider.id}
                type="button"
                onClick={() => handleProviderToggle(provider.id)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg border transition-colors ${
                  selectedProviders.includes(provider.id)
                    ? 'bg-[#FF6B35]/10 border-[#FF6B35]/30 text-[#FF6B35]'
                    : 'bg-bg-tertiary border-border-subtle text-text-secondary hover:border-border-default'
                }`}
              >
                <ProviderIcon provider={provider.id} size="sm" />
                <span className="text-sm">{provider.name}</span>
              </button>
            ))}
          </div>
        )}
        {selectedProviders.length > 0 && (
          <div className="w-48">
            <Label
              htmlFor="region"
              className="text-xs font-medium text-text-secondary mb-1.5 block"
            >
              Region
            </Label>
            <Select
              value={selectedRegion}
              onValueChange={(v) => {
                setSelectedRegion(v);
                markDirty();
              }}
            >
              <SelectTrigger id="region">
                <SelectValue placeholder="Select region" />
              </SelectTrigger>
              <SelectContent>
                {providers
                  .filter((p) => selectedProviders.includes(p.id))
                  .flatMap((p) => p.regions)
                  .filter((r, i, arr) => arr.indexOf(r) === i)
                  .map((r) => (
                    <SelectItem key={r} value={r}>
                      {r.toUpperCase()}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>
        )}
      </SectionCard>

      <SectionCard
        icon={<Key className="w-4 h-4" />}
        title="Environment Variables"
        step={4}
        description="Key-value pairs injected at runtime"
      >
        {envVars.length > 0 && (
          <div className="space-y-2">
            {envVars.map((envVar) => (
              <div
                key={envVar.id}
                className="flex items-center gap-3 p-3 rounded-lg"
                style={{ background: 'var(--bg-tertiary)' }}
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <code className="text-sm font-medium text-text-primary">{envVar.key}</code>
                    {envVar.isSecret && <EyeOff className="w-3 h-3 text-text-muted shrink-0" />}
                  </div>
                  <div className="text-sm text-text-secondary mt-0.5 truncate font-mono">
                    {envVar.isSecret
                      ? showEnvValues[envVar.id]
                        ? envVar.value
                        : '••••••••'
                      : envVar.value}
                  </div>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  {envVar.isSecret && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-text-muted hover:text-text-primary h-7 w-7 p-0"
                      onClick={() =>
                        setShowEnvValues((p) => ({ ...p, [envVar.id]: !p[envVar.id] }))
                      }
                      aria-label={showEnvValues[envVar.id] ? 'Hide value' : 'Show value'}
                    >
                      {showEnvValues[envVar.id] ? (
                        <EyeOff className="w-3.5 h-3.5" />
                      ) : (
                        <Eye className="w-3.5 h-3.5" />
                      )}
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => removeEnvironmentVariable(envVar.id)}
                    className="text-text-muted hover:text-red-400 h-7 w-7"
                    aria-label={`Remove ${envVar.key}`}
                  >
                    <X className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}

        <Separator className="opacity-30" />

        <div className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <Label htmlFor="env-key" className="text-xs text-text-secondary mb-1 block">
                Key
              </Label>
              <Input
                id="env-key"
                placeholder="API_KEY"
                value={newEnvKey}
                onChange={(e) => setNewEnvKey(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && addEnvironmentVariable()}
                className="input font-mono text-sm"
              />
            </div>
            <div>
              <Label htmlFor="env-value" className="text-xs text-text-secondary mb-1 block">
                Value
              </Label>
              <Input
                id="env-value"
                type={isNewEnvSecret ? 'password' : 'text'}
                placeholder="your-value"
                value={newEnvValue}
                onChange={(e) => setNewEnvValue(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && addEnvironmentVariable()}
                className="input font-mono text-sm"
              />
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2">
              <Checkbox
                id="isSecret"
                checked={isNewEnvSecret}
                onCheckedChange={(c) => setIsNewEnvSecret(c === true)}
              />
              <Label htmlFor="isSecret" className="text-xs cursor-pointer text-text-secondary">
                Mark as secret
              </Label>
            </div>
            <Button
              size="sm"
              onClick={addEnvironmentVariable}
              disabled={!newEnvKey.trim()}
              className="gap-1"
            >
              <Plus className="w-3 h-3" />
              Add
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setVaultPickerOpen(true)}
              className="gap-1.5"
            >
              <Shield className="w-3.5 h-3.5" />
              Use from Vault
            </Button>
          </div>
        </div>
      </SectionCard>

      <SectionCard
        icon={<Timer className="w-4 h-4" />}
        title="Resource Limits"
        step={5}
        description="Control memory, timeout, and concurrency"
      >
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div>
            <Label className="text-xs text-text-secondary mb-1.5 block">
              Memory
              <InfoTip content="Maximum RAM allocated to your function per invocation." />
            </Label>
            <Select
              value={String(resources.memoryMb)}
              onValueChange={(v) => {
                setResources((r) => ({ ...r, memoryMb: Number(v) }));
                markDirty();
              }}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {MEMORY_OPTIONS.map((m) => (
                  <SelectItem key={m} value={String(m)}>
                    {m} MB
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label className="text-xs text-text-secondary mb-1.5 block">
              Timeout
              <InfoTip content="Maximum execution time before the function is terminated." />
            </Label>
            <Select
              value={String(resources.timeoutMs)}
              onValueChange={(v) => {
                setResources((r) => ({ ...r, timeoutMs: Number(v) }));
                markDirty();
              }}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {TIMEOUT_OPTIONS.map((t) => (
                  <SelectItem key={t} value={String(t)}>
                    {formatTimeout(t)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div>
            <Label htmlFor="concurrency" className="text-xs text-text-secondary mb-1.5 block">
              Max Concurrency
              <InfoTip content="Maximum simultaneous executions of this function." />
            </Label>
            <Input
              id="concurrency"
              type="number"
              min={1}
              max={1000}
              value={resources.maxConcurrency}
              onChange={(e) => {
                setResources((r) => ({ ...r, maxConcurrency: Number(e.target.value) }));
                markDirty();
              }}
              className="input"
            />
          </div>
        </div>
      </SectionCard>

      <SectionCard
        icon={<Webhook className="w-4 h-4" />}
        title="Triggers"
        step={6}
        description="Define how your function is invoked"
      >
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Globe className="w-4 h-4 text-text-muted" />
              <span className="text-sm font-medium text-text-primary">HTTP Trigger</span>
            </div>
            <Switch
              checked={httpTrigger.enabled}
              onCheckedChange={(c) => {
                setHttpTrigger((t) => ({ ...t, enabled: c }));
                markDirty();
              }}
              aria-label="Enable HTTP trigger"
            />
          </div>
          {httpTrigger.enabled && (
            <div className="grid grid-cols-[120px_1fr] gap-3 pl-6">
              <div>
                <Label className="text-xs text-text-secondary mb-1 block">Method</Label>
                <Select
                  value={httpTrigger.method}
                  onValueChange={(v) => {
                    setHttpTrigger((t) => ({ ...t, method: v as HttpMethod }));
                    markDirty();
                  }}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {HTTP_METHODS.map((m) => (
                      <SelectItem key={m} value={m}>
                        {m}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="http-path" className="text-xs text-text-secondary mb-1 block">
                  Path
                </Label>
                <Input
                  id="http-path"
                  placeholder="/api/my-function"
                  value={httpTrigger.path}
                  onChange={(e) => {
                    setHttpTrigger((t) => ({ ...t, path: e.target.value }));
                    markDirty();
                  }}
                  className="input font-mono text-sm"
                />
                <FieldError message={errors.httpPath} />
              </div>
            </div>
          )}
        </div>

        <Separator className="opacity-30" />

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4 text-text-muted" />
              <span className="text-sm font-medium text-text-primary">Schedule (Cron)</span>
            </div>
            <Switch
              checked={scheduleTrigger.enabled}
              onCheckedChange={(c) => {
                setScheduleTrigger((t) => ({ ...t, enabled: c }));
                markDirty();
              }}
              aria-label="Enable schedule trigger"
            />
          </div>
          {scheduleTrigger.enabled && (
            <div className="grid grid-cols-2 gap-3 pl-6">
              <div>
                <Label htmlFor="cron" className="text-xs text-text-secondary mb-1 block">
                  Cron Expression
                  <InfoTip content="Standard 5-field cron: minute hour day month weekday" />
                </Label>
                <Input
                  id="cron"
                  placeholder="0 * * * *"
                  value={scheduleTrigger.cron}
                  onChange={(e) => {
                    setScheduleTrigger((t) => ({ ...t, cron: e.target.value }));
                    markDirty();
                  }}
                  className="input font-mono text-sm"
                />
              </div>
              <div>
                <Label htmlFor="tz" className="text-xs text-text-secondary mb-1 block">
                  Timezone
                </Label>
                <Input
                  id="tz"
                  placeholder="UTC"
                  value={scheduleTrigger.timezone}
                  onChange={(e) => {
                    setScheduleTrigger((t) => ({ ...t, timezone: e.target.value }));
                    markDirty();
                  }}
                  className="input text-sm"
                />
              </div>
            </div>
          )}
        </div>
      </SectionCard>

      <SectionCard icon={<Globe className="w-4 h-4" />} title="Visibility & Tags" step={7}>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-text-primary flex items-center gap-2">
              {visibility === 'public' ? (
                <>
                  <Globe className="w-4 h-4 text-emerald-400" /> Public
                </>
              ) : (
                <>
                  <Lock className="w-4 h-4 text-text-muted" /> Private
                </>
              )}
            </p>
            <p className="text-xs text-text-muted mt-0.5">
              {visibility === 'public'
                ? 'Anyone can discover and call this function'
                : 'Only you and your team can access this function'}
            </p>
          </div>
          <Switch
            checked={visibility === 'public'}
            onCheckedChange={(c) => {
              setVisibility(c ? 'public' : 'private');
              markDirty();
            }}
            aria-label="Toggle visibility"
          />
        </div>

        <Separator className="opacity-30" />

        <div>
          <Label className="text-xs text-text-secondary mb-2 block">
            Tags
            <InfoTip content="Labels for organizing and discovering functions." />
          </Label>
          <div className="flex flex-wrap gap-1.5 mb-2">
            {tags.map((t) => (
              <Badge
                key={t}
                variant="secondary"
                className="gap-1 text-xs cursor-pointer hover:bg-red-500/20 hover:text-red-400 transition-colors"
                onClick={() => removeTag(t)}
              >
                {t}
                <X className="w-2.5 h-2.5" />
              </Badge>
            ))}
          </div>
          <div className="flex gap-2">
            <Input
              ref={tagInputRef}
              placeholder="Add tag…"
              value={newTag}
              onChange={(e) => setNewTag(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  addTag();
                }
              }}
              className="input text-sm"
            />
            <Button size="sm" variant="outline" onClick={addTag} disabled={!newTag.trim()}>
              <Plus className="w-3.5 h-3.5" />
            </Button>
          </div>
        </div>
      </SectionCard>

      <SectionCard
        icon={<Settings2 className="w-4 h-4" />}
        title="Advanced Settings"
        step={8}
        description="Retry policy and concurrency controls"
      >
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Label htmlFor="max-retries" className="text-xs text-text-secondary mb-1.5 block">
              Max Retries
              <InfoTip content="Number of automatic retries on failure before giving up." />
            </Label>
            <Input
              id="max-retries"
              type="number"
              min={0}
              max={10}
              value={retryPolicy.maxRetries}
              onChange={(e) => {
                setRetryPolicy((r) => ({ ...r, maxRetries: Number(e.target.value) }));
                markDirty();
              }}
              className="input"
            />
          </div>
          <div>
            <Label htmlFor="backoff" className="text-xs text-text-secondary mb-1.5 block">
              Retry Backoff
              <InfoTip content="Delay in milliseconds between retry attempts." />
            </Label>
            <Input
              id="backoff"
              type="number"
              min={100}
              max={60000}
              step={100}
              value={retryPolicy.backoffMs}
              onChange={(e) => {
                setRetryPolicy((r) => ({ ...r, backoffMs: Number(e.target.value) }));
                markDirty();
              }}
              className="input"
            />
          </div>
        </div>
      </SectionCard>
    </div>
  );
}
