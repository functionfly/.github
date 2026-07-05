/**
 * PublishRegistryFunctionPage — author and publish a function to the
 * FunctionFly registry.
 *
 * Uses the Sealed Containment design system.
 * Uses presigned direct-upload when the source / WASM / README crosses 256 KiB
 * so the bytes stream straight to R2 instead of through the orchestrator.
 */

import { AlertCircle, ArrowLeft, CheckCircle2, FileCode, FileText, Info, Loader2, Rocket, Shield, Upload } from 'lucide-react';
import { editor } from 'monaco-editor';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Chamber } from '@/components/containment';
import { FrameButton, Input, SealedButton } from '@/components/containment';
import { usePageTitle } from '@/hooks';
import { useAuthStore } from '@/stores/authStore';
import { functionsApi } from '@/api';
import { formatBytes } from './formatters';
import { useRegistryPublish } from './useRegistryPublish';
import styles from './PublishRegistryFunctionPage.module.css';
import { LazyMonacoEditor, type OnMount } from '@/components/LazyMonacoEditor';

const RUNTIMES = [
  'python3.12',
  'python3.11',
  'python3.10',
  'node20',
  'node18',
  'deno',
  'wasm',
  'typescript',
];

const SIDE_EFFECTS = ['none', 'network', 'external_state'] as const;
const TRUST_LEVELS = ['standard', 'high', 'enterprise'] as const;
const CONFLICT_STRATEGIES = [
  { value: 'error', label: 'Error if exists', desc: 'Fail if version already exists' },
  { value: 'create_new', label: 'Create new version', desc: 'Always create a new version' },
  { value: 'overwrite', label: 'Overwrite', desc: 'Replace existing version (use with caution)' },
] as const;
const CHANGELOG_CATEGORIES = [
  { value: 'feature', label: 'Feature' },
  { value: 'improvement', label: 'Improvement' },
  { value: 'bug_fix', label: 'Bug Fix' },
  { value: 'breaking_change', label: 'Breaking Change' },
  { value: 'other', label: 'Other' },
] as const;

const DEFAULT_PYTHON_SOURCE = `def handler(event):
    """FunctionFly function: receives an event dict, returns a JSON-serialisable response."""
    name = event.get("name", "world")
    return {"ok": True, "message": f"Hello, {name}!"}
`;

function runtimeToMonacoLanguage(runtime: string): string {
  if (runtime.startsWith('python')) return 'python';
  if (runtime.startsWith('typescript')) return 'typescript';
  if (runtime.startsWith('node') || runtime === 'deno') return 'javascript';
  if (runtime === 'wasm') return 'plaintext';
  return 'plaintext';
}

function mapPlatformRuntimeToRegistry(platformRuntime: string): string {
  const map: Record<string, string> = {
    'typescript': 'typescript',
    'javascript': 'node20',
    'node20': 'node20',
    'node18': 'node18',
    'python': 'python3.12',
    'python-wasm': 'wasm',
    'rust-wasm': 'wasm',
    'browser-wasm': 'wasm',
    'go': 'node20',
    'deno': 'deno',
    'bun': 'node20',
  };
  return map[platformRuntime] || 'python3.12';
}

export function PublishRegistryFunctionPage() {
  usePageTitle('Publish Function');

  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user } = useAuthStore();
  const authorDefault = useMemo(() => user?.username ?? '', [user?.username]);
  const { state, setState, progress, isPublishing, result, error, publish, reset } =
    useRegistryPublish(authorDefault);

  // Pre-fill from FunctionEditorPage when navigated with a functionId
  const [isPrefilling, setIsPrefilling] = useState(false);
  useEffect(() => {
    const functionId = searchParams.get('functionId');
    if (!functionId) return;
    let cancelled = false;
    setIsPrefilling(true);
    functionsApi
      .get(functionId)
      .then((fn) => {
        if (cancelled) return;
        const fnAny = fn as any;
        const runtimeRegistry = mapPlatformRuntimeToRegistry(fnAny.runtime || 'python3.12');
        setState((prev) => ({
          ...prev,
          name: fn.name || prev.name,
          runtime: runtimeRegistry,
          description: fnAny.description || prev.description,
          tags: ((fnAny.tags as string[]) || []).join(','),
        }));
        setSourceCode(fn.code || DEFAULT_PYTHON_SOURCE);
        setIsPrefilling(false);
      })
      .catch(() => {
        if (!cancelled) setIsPrefilling(false);
      });
    return () => {
      cancelled = true;
    };
  }, [searchParams, setState]);

  const [sourceCode, setSourceCode] = useState<string>(DEFAULT_PYTHON_SOURCE);
  const [sourceBlob, setSourceBlob] = useState<Blob | undefined>(undefined);
  const [wasmFile, setWasmFile] = useState<File | null>(null);
  const [wasmBase64, setWasmBase64] = useState<string | undefined>(undefined);
  const [readme, setReadme] = useState<string>('');
  const [activeTab, setActiveTab] = useState<'manifest' | 'source' | 'extras'>('manifest');
  const sourceEditorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const wasSourceLoaded = useRef(false);
  const MonacoEditorRef = useRef<typeof import('@/components/LazyMonacoEditor').LazyMonacoEditor | null>(null);

  useEffect(() => {
    setSourceBlob(new Blob([sourceCode], { type: textContentTypeFor(state.runtime) }));
  }, [sourceCode, state.runtime]);

  const handleMonacoLoad = (ed: unknown) => {
    sourceEditorRef.current = ed as unknown as editor.IStandaloneCodeEditor;
  };

  function pickWasmFile(file: File) {
    setWasmFile(file);
    const reader = new FileReader();
    reader.onload = () => {
      const result = reader.result;
      if (typeof result === 'string') {
        const comma = result.indexOf(',');
        setWasmBase64(comma >= 0 ? result.slice(comma + 1) : result);
      }
    };
    reader.readAsDataURL(file);
  }

  async function onPublish() {
    if (isPublishing) return;
    const ok = await publish({
      sourceCode,
      sourceBlob,
      wasmBase64,
      wasmBlob: wasmFile ?? undefined,
      readme: readme.trim() || undefined,
    });
    if (ok && !wasSourceLoaded.current) {
      wasSourceLoaded.current = true;
    }
  }

  return (
    <div className={styles.page}>
      <div className="page-grid" />

      {/* Top action bar */}
      <div className={styles.topBar}>
        <div className={styles.topBarInner}>
          <div className={styles.topBarLeft}>
            {searchParams.has('functionId') ? (
              <FrameButton size="sm" onClick={() => navigate(-1)} className={styles.backButton}>
                <ArrowLeft className={styles.backIcon} />
                Back to Editor
              </FrameButton>
            ) : (
              <FrameButton size="sm" onClick={() => navigate('/registry')} className={styles.backButton}>
                <ArrowLeft className={styles.backIcon} />
                Registry
              </FrameButton>
            )}
            <h1 className={styles.pageTitle}>
              Publish function
              {isPrefilling && <Loader2 className={styles.pageTitleSpinner} />}
            </h1>
          </div>
          <div className={styles.topBarRight}>
            <FrameButton size="sm" onClick={reset} disabled={isPublishing}>
              Reset
            </FrameButton>
            <SealedButton
              size="sm"
              onClick={onPublish}
              loading={isPublishing}
              iconLeft={isPublishing ? undefined : <Rocket className={styles.rocketIcon} />}
            >
              {isPublishing ? 'Publishing…' : 'Publish'}
            </SealedButton>
          </div>
        </div>
      </div>

      <div className={styles.content}>
        <div className={styles.mainColumn}>
          {/* Tabs */}
          <div className={styles.tabs}>
            <button
              type="button"
              className={`${styles.tab} ${activeTab === 'manifest' ? styles.tabActive : ''}`}
              onClick={() => setActiveTab('manifest')}
            >
              Manifest
            </button>
            <button
              type="button"
              className={`${styles.tab} ${activeTab === 'source' ? styles.tabActive : ''}`}
              onClick={() => setActiveTab('source')}
            >
              Source
            </button>
            <button
              type="button"
              className={`${styles.tab} ${activeTab === 'extras' ? styles.tabActive : ''}`}
              onClick={() => setActiveTab('extras')}
            >
              WASM &amp; README
            </button>
          </div>

          {/* Tab panels */}
          <div className={styles.tabContent}>
            {activeTab === 'manifest' && (
              <ManifestCard state={state} setState={setState} />
            )}

            {activeTab === 'source' && (
              <SourceCard
                runtime={state.runtime}
                sourceCode={sourceCode}
                setSourceCode={setSourceCode}
                onMount={handleMonacoLoad}
                sourceBytes={sourceBlob?.size ?? 0}
              />
            )}

            {activeTab === 'extras' && (
              <>
                <WasmCard wasmFile={wasmFile} setWasmFile={pickWasmFile} />
                <ReadmeCard readme={readme} setReadme={setReadme} />
                <ChangelogCard
                  category={state.changelogCategory}
                  title={state.changelogTitle}
                  description={state.changelogDescription}
                  onCategoryChange={(v) => setState((s) => ({ ...s, changelogCategory: v as typeof state.changelogCategory }))}
                  onTitleChange={(v) => setState((s) => ({ ...s, changelogTitle: v }))}
                  onDescriptionChange={(v) => setState((s) => ({ ...s, changelogDescription: v }))}
                />
              </>
            )}
          </div>

          {/* Progress + result banner */}
          <PublishProgress progress={progress} error={error} result={result} />
        </div>

        {/* Right sidebar */}
        <div className={styles.sidebar}>
          <Chamber className={styles.infoCard}>
            <h3 className={styles.infoCardTitle}>What happens on publish</h3>
            <p className={styles.infoCardDesc}>
              The dashboard may stream large artifacts straight to R2 via a
              short-lived presigned URL. Small payloads use JSON publish as
              usual. Either way, only the content hash + size land in
              Postgres; bytes live in object storage.
            </p>
            <ul className={styles.infoCardList}>
              <li>Author, name, version follow semver.</li>
              <li>Tags are comma-separated.</li>
              <li>Trust level affects moderation gating.</li>
              <li>Changelog is auto-generated if left empty.</li>
            </ul>
          </Chamber>

          <PublishFeeEstimate isNewFunction={!state.name} />

          {state.trustLevel !== 'standard' && (
            <Chamber className={styles.trustGatingCard}>
              <div className={styles.trustGatingHeader}>
                <Shield className={styles.trustGatingIcon} />
                <h3 className={styles.trustGatingTitle}>
                  {state.trustLevel === 'enterprise' ? 'Enterprise Trust' : 'High Trust'}
                </h3>
              </div>
              <p className={styles.trustGatingDesc}>
                {state.trustLevel === 'enterprise'
                  ? 'Enterprise functions require manual review and approval before they can be executed. This may take up to 48 hours.'
                  : 'High trust functions require manual review before execution. You will be notified when approval is granted.'}
              </p>
            </Chamber>
          )}

          {result?.ok && (
            <Chamber className={styles.resultCard}>
              <div className={styles.resultHeader}>
                <CheckCircle2 className={styles.resultIcon} />
                <h3 className={styles.resultTitle}>Published</h3>
              </div>
              <div className={styles.resultBody}>
                <code className={styles.resultFunction}>{result.function}</code>
                <div className={styles.resultVersion}>
                  Version <span className={styles.versionBadge}>{result.version}</span>
                </div>
                <div className={styles.resultActions}>
                  <SealedButton size="sm" onClick={() => navigate(`/fx/${result.function}`)}>
                    View
                  </SealedButton>
                  <FrameButton size="sm" onClick={reset}>
                    Publish another
                  </FrameButton>
                </div>
              </div>
            </Chamber>
          )}
        </div>
      </div>
    </div>
  );
}

function ManifestCard({
  state,
  setState,
}: {
  state: ReturnType<typeof useRegistryPublish>['state'];
  setState: ReturnType<typeof useRegistryPublish>['setState'];
}) {
  return (
    <Chamber>
      <h2 className={styles.cardTitle}>Identity &amp; metadata</h2>
      <p className={styles.cardDesc}>Names, runtime, limits. The function ID is auto-generated.</p>

      <div className={styles.formGrid}>
        <Field label="Author" required>
          <Input
            value={state.author}
            onChange={(e) => setState((s) => ({ ...s, author: e.target.value }))}
            placeholder="my-org"
          />
        </Field>
        <Field label="Name" required>
          <Input
            value={state.name}
            onChange={(e) => setState((s) => ({ ...s, name: e.target.value }))}
            placeholder="my-function"
          />
        </Field>
        <Field label="Version" required>
          <Input
            value={state.version}
            onChange={(e) => setState((s) => ({ ...s, version: e.target.value }))}
            placeholder="1.0.0"
          />
        </Field>
      </div>

      <div className={styles.formGrid2}>
        <Field label="Title">
          <Input
            value={state.title}
            onChange={(e) => setState((s) => ({ ...s, title: e.target.value }))}
            placeholder="My Function"
          />
        </Field>
        <Field label="Category">
          <Input
            value={state.category}
            onChange={(e) => setState((s) => ({ ...s, category: e.target.value }))}
            placeholder="utility"
          />
        </Field>
      </div>

      <Field label="Description">
        <textarea
          className={styles.textarea}
          value={state.description}
          onChange={(e) => setState((s) => ({ ...s, description: e.target.value }))}
          placeholder="What this function does, when to use it…"
          rows={3}
        />
      </Field>

      <Field label="Tags" hint="comma-separated">
        <Input
          value={state.tags}
          onChange={(e) => setState((s) => ({ ...s, tags: e.target.value }))}
          placeholder="encoding, utility, text"
        />
      </Field>

      <div className={styles.formGrid3}>
        <Field label="Runtime">
          <div className={styles.selectWrapper}>
            <select
              className={styles.select}
              value={state.runtime}
              onChange={(e) => setState((s) => ({ ...s, runtime: e.target.value }))}
            >
              {RUNTIMES.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
            <div className={styles.selectChevron}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="m6 9 6 6 6-6" />
              </svg>
            </div>
          </div>
        </Field>
        <Field label="Side effects">
          <div className={styles.selectWrapper}>
            <select
              className={styles.select}
              value={state.sideEffects}
              onChange={(e) =>
                setState((s) => ({
                  ...s,
                  sideEffects: e.target.value as (typeof SIDE_EFFECTS)[number],
                }))
              }
            >
              {SIDE_EFFECTS.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
            <div className={styles.selectChevron}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="m6 9 6 6 6-6" />
              </svg>
            </div>
          </div>
        </Field>
        <Field label="Trust level">
          <div className={styles.selectWrapper}>
            <select
              className={styles.select}
              value={state.trustLevel}
              onChange={(e) =>
                setState((s) => ({ ...s, trustLevel: e.target.value as (typeof TRUST_LEVELS)[number] }))
              }
            >
              {TRUST_LEVELS.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
            <div className={styles.selectChevron}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="m6 9 6 6 6-6" />
              </svg>
            </div>
          </div>
        </Field>
      </div>

      <div className={styles.formGrid2}>
        <Field label="Timeout (ms)" hint="100–30000">
          <Input
            type="number"
            value={state.timeoutMs}
            min={100}
            max={30000}
            onChange={(e) =>
              setState((s) => ({ ...s, timeoutMs: Number(e.target.value) || 0 }))
            }
          />
        </Field>
        <Field label="Memory (MB)" hint="32–1024">
          <Input
            type="number"
            value={state.memoryMb}
            min={32}
            max={1024}
            onChange={(e) =>
              setState((s) => ({ ...s, memoryMb: Number(e.target.value) || 0 }))
            }
          />
        </Field>
      </div>

      <div className={styles.checkboxGroup}>
        <label className={styles.checkbox}>
          <input
            type="checkbox"
            checked={state.deterministic}
            onChange={(e) => setState((s) => ({ ...s, deterministic: e.target.checked }))}
          />
          <span>Deterministic</span>
        </label>
        <label className={styles.checkbox}>
          <input
            type="checkbox"
            checked={state.idempotent}
            onChange={(e) => setState((s) => ({ ...s, idempotent: e.target.checked }))}
          />
          <span>Idempotent</span>
        </label>
      </div>

      <div className={styles.conflictSection}>
        <div className={styles.conflictHeader}>
          <h3 className={styles.conflictTitle}>Version conflict strategy</h3>
        </div>
        <div className={styles.conflictOptions}>
          {CONFLICT_STRATEGIES.map((strategy) => (
            <label
              key={strategy.value}
              className={`${styles.conflictOption} ${state.conflictStrategy === strategy.value ? styles.conflictOptionActive : ''}`}
            >
              <input
                type="radio"
                name="conflictStrategy"
                value={strategy.value}
                checked={state.conflictStrategy === strategy.value}
                onChange={(e) =>
                  setState((s) => ({ ...s, conflictStrategy: e.target.value as typeof s.conflictStrategy }))
                }
                className={styles.conflictRadio}
              />
              <div className={styles.conflictOptionContent}>
                <span className={styles.conflictOptionLabel}>{strategy.label}</span>
                <span className={styles.conflictOptionDesc}>{strategy.desc}</span>
              </div>
            </label>
          ))}
        </div>
      </div>
    </Chamber>
  );
}

function SourceCard({
  runtime,
  sourceCode,
  setSourceCode,
  onMount,
  sourceBytes,
}: {
  runtime: string;
  sourceCode: string;
  setSourceCode: (v: string) => void;
  onMount: OnMount;
  sourceBytes: number;
}) {
  const language = runtimeToMonacoLanguage(runtime);
  const willUseDirectUpload = sourceBytes > 256 * 1024;

  return (
    <Chamber>
      <div className={styles.sourceHeader}>
        <h2 className={styles.cardTitle}>Source</h2>
        <div className={styles.sourceMeta}>
          <FileCode className={styles.sourceIcon} />
          <span className={styles.sourceSize}>{formatBytes(sourceBytes)} source.</span>
          <span className={willUseDirectUpload ? styles.badgeDirect : styles.badgeJson}>
            {willUseDirectUpload ? 'direct-to-R2' : 'JSON publish'}
          </span>
        </div>
      </div>
      <div className={styles.editorWrapper}>
        <LazyMonacoEditor
          height="60vh"
          language={language}
          value={sourceCode}
          onChange={(v) => setSourceCode(v ?? '')}
          onMount={onMount}
          theme="vs-dark"
          options={{
            minimap: { enabled: false },
            fontSize: 13,
            wordWrap: 'on',
            scrollBeyondLastLine: false,
          }}
        />
      </div>
    </Chamber>
  );
}

function WasmCard({
  wasmFile,
  setWasmFile,
}: {
  wasmFile: File | null;
  setWasmFile: (f: File) => void;
}) {
  return (
    <Chamber>
      <h2 className={styles.cardTitle}>WASM binary (optional)</h2>
      <p className={styles.cardDesc}>
        <span className={styles.wasmDesc}>
          <Upload className={styles.wasmIcon} />
          <span>
            Required only when the runtime is <code>wasm</code> and you want
            to pre-compile rather than bundle lazily.
          </span>
        </span>
      </p>
      <div className={styles.fileInputWrapper}>
        <input
          type="file"
          accept=".wasm,application/wasm"
          className={styles.fileInput}
          onChange={(e) => {
            const f = e.target.files?.[0];
            if (f) setWasmFile(f);
          }}
        />
      </div>
      {wasmFile && (
        <p className={styles.fileInfo}>
          {wasmFile.name} ({formatBytes(wasmFile.size)})
        </p>
      )}
    </Chamber>
  );
}

function ReadmeCard({
  readme,
  setReadme,
}: {
  readme: string;
  setReadme: (v: string) => void;
}) {
  const willUseDirectUpload = readme.length > 256 * 1024;

  return (
    <Chamber>
      <div className={styles.readmeHeader}>
        <h2 className={styles.cardTitle}>README (optional)</h2>
        <div className={styles.sourceMeta}>
          <FileText className={styles.sourceIcon} />
          <span className={styles.sourceSize}>{formatBytes(readme.length)} markdown.</span>
          {willUseDirectUpload && (
            <span className={styles.badgeDirect}>direct-to-R2</span>
          )}
        </div>
      </div>
      <textarea
        className={styles.textarea}
        value={readme}
        onChange={(e) => setReadme(e.target.value)}
        placeholder="# My function\n\nDescribe inputs, outputs, and example payloads."
        rows={8}
      />
    </Chamber>
  );
}

function ChangelogCard({
  category,
  title,
  description,
  onCategoryChange,
  onTitleChange,
  onDescriptionChange,
}: {
  category: string;
  title: string;
  description: string;
  onCategoryChange: (v: string) => void;
  onTitleChange: (v: string) => void;
  onDescriptionChange: (v: string) => void;
}) {
  return (
    <Chamber>
      <div className={styles.changelogHeader}>
        <h2 className={styles.cardTitle}>Changelog (optional)</h2>
        <p className={styles.cardDesc}>
          Describe what changed in this version. If left empty, a changelog will be auto-generated.
        </p>
      </div>

      <div className={styles.formGrid2}>
        <Field label="Category">
          <div className={styles.selectWrapper}>
            <select
              className={styles.select}
              value={category}
              onChange={(e) => onCategoryChange(e.target.value)}
            >
              {CHANGELOG_CATEGORIES.map((c) => (
                <option key={c.value} value={c.value}>
                  {c.label}
                </option>
              ))}
            </select>
            <div className={styles.selectChevron}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="m6 9 6 6 6-6" />
              </svg>
            </div>
          </div>
        </Field>
        <Field label="Title">
          <Input
            value={title}
            onChange={(e) => onTitleChange(e.target.value)}
            placeholder="What changed"
          />
        </Field>
      </div>

      <Field label="Description">
        <textarea
          className={styles.textarea}
          value={description}
          onChange={(e) => onDescriptionChange(e.target.value)}
          placeholder="Details about this change…"
          rows={3}
        />
      </Field>
    </Chamber>
  );
}

function PublishFeeEstimate({ isNewFunction }: { isNewFunction: boolean }) {
  const publishFee = 2.99;
  const updateFee = 0.99;
  const fee = isNewFunction ? publishFee : updateFee;
  const feeLabel = isNewFunction ? 'New function publish' : 'Version update';

  return (
    <Chamber className={styles.feeCard}>
      <div className={styles.feeHeader}>
        <Info className={styles.feeIcon} />
        <h3 className={styles.feeTitle}>Publish fee</h3>
      </div>
      <div className={styles.feeBody}>
        <div className={styles.feeRow}>
          <span className={styles.feeLabel}>{feeLabel}</span>
          <span className={styles.feeAmount}>${fee.toFixed(2)}</span>
        </div>
        <p className={styles.feeNote}>
          This fee will be charged to your wallet. Ensure you have sufficient balance to publish.
        </p>
      </div>
    </Chamber>
  );
}

function PublishProgress({
  progress,
  error,
  result,
}: {
  progress: ReturnType<typeof useRegistryPublish>['progress'];
  error: string | null;
  result: ReturnType<typeof useRegistryPublish>['result'];
}) {
  if (progress.stage === 'idle') return null;
  const isError = progress.stage === 'error';

  return (
    <div className={`${styles.progress} ${isError ? styles.progressError : ''}`}>
      <div className={styles.progressHeader}>
        <span className={styles.progressLabel}>
          {isError ? <AlertCircle className={styles.progressIcon} /> : <Loader2 className={`${styles.progressIcon} ${styles.spinning}`} />}
          <span className={styles.progressStage}>{stageLabel(progress.stage)}</span>
          {progress.message && (
            <span className={styles.progressMessage}>{progress.message}</span>
          )}
        </span>
        {result?.ok && <span className={styles.progressBadge}>done</span>}
      </div>
      <div className={styles.progressBar}>
        <div className={styles.progressFill} style={{ width: `${progress.percent}%` }} />
      </div>
      {error && <p className={styles.progressErrorText}>{error}</p>}
    </div>
  );
}

function Field({
  label,
  required,
  hint,
  children,
}: {
  label: string;
  required?: boolean;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div className={styles.field}>
      <label className={styles.fieldLabel}>
        {label}
        {required && <span className={styles.required}>*</span>}
      </label>
      {children}
      {hint && <p className={styles.fieldHint}>{hint}</p>}
    </div>
  );
}

function stageLabel(stage: ReturnType<typeof useRegistryPublish>['progress']['stage']): string {
  switch (stage) {
    case 'idle':
      return 'Ready';
    case 'wasm':
      return 'Uploading WASM';
    case 'source':
      return 'Uploading source';
    case 'readme':
      return 'Uploading README';
    case 'publish':
      return 'Publishing';
    case 'done':
      return 'Published';
    case 'error':
      return 'Failed';
  }
}

function textContentTypeFor(runtime: string): string {
  if (runtime.startsWith('python')) return 'text/x-python; charset=utf-8';
  if (runtime.startsWith('typescript')) return 'text/typescript; charset=utf-8';
  if (runtime.startsWith('node') || runtime === 'deno') return 'text/javascript; charset=utf-8';
  return 'text/plain; charset=utf-8';
}

export default PublishRegistryFunctionPage;
