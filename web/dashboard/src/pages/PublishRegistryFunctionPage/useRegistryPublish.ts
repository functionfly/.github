import { useCallback, useMemo, useState } from 'react';
import { toast } from 'sonner';
import { registryApi } from '@/api/registry';

export type PublishStage = 'idle' | 'wasm' | 'source' | 'readme' | 'publish' | 'done' | 'error';

export type ConflictStrategy = 'error' | 'overwrite' | 'create_new';

export type ChangelogCategory = 'feature' | 'bug_fix' | 'breaking_change' | 'improvement' | 'other';

export interface PublishFormState {
  author: string;
  name: string;
  version: string;
  runtime: string;
  title: string;
  description: string;
  category: string;
  tags: string; // comma-separated; the dashboard parses on submit
  timeoutMs: number;
  memoryMb: number;
  deterministic: boolean;
  idempotent: boolean;
  sideEffects: 'none' | 'network' | 'external_state';
  trustLevel: 'standard' | 'high' | 'enterprise';
  conflictStrategy: ConflictStrategy;
  changelogCategory: ChangelogCategory;
  changelogTitle: string;
  changelogDescription: string;
}

export const DEFAULT_FORM_STATE: PublishFormState = {
  author: '',
  name: '',
  version: '1.0.0',
  runtime: 'python3.12',
  title: '',
  description: '',
  category: '',
  tags: '',
  timeoutMs: 5000,
  memoryMb: 128,
  deterministic: false,
  idempotent: false,
  sideEffects: 'none',
  trustLevel: 'standard',
  conflictStrategy: 'error',
  changelogCategory: 'feature',
  changelogTitle: 'Initial release',
  changelogDescription: '',
};

export interface PublishProgress {
  stage: PublishStage;
  /** 0–100 within the current stage, or 100 once complete. */
  percent: number;
  message: string;
}

const INITIAL_PROGRESS: PublishProgress = {
  stage: 'idle',
  percent: 0,
  message: '',
};

export interface UseRegistryPublishReturn {
  state: PublishFormState;
  setState: React.Dispatch<React.SetStateAction<PublishFormState>>;
  progress: PublishProgress;
  isPublishing: boolean;
  result: { ok: boolean; function_id: string; version: string; function: string } | null;
  error: string | null;
  publish: (args: {
    sourceCode: string;
    sourceBlob?: Blob;
    wasmBase64?: string;
    wasmBlob?: Blob;
    readme?: string;
  }) => Promise<boolean>;
  reset: () => void;
}

function buildManifest(form: PublishFormState) {
  const tags = form.tags
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean);
  return {
    name: form.name,
    version: form.version,
    runtime: form.runtime,
    title: form.title || undefined,
    description: form.description || undefined,
    category: form.category || undefined,
    tags: tags.length > 0 ? tags : undefined,
    timeout_ms: form.timeoutMs,
    memory_mb: form.memoryMb,
    deterministic: form.deterministic,
    side_effects: form.sideEffects,
    idempotent: form.idempotent,
  };
}

/**
 * UseRegistryPublish — orchestrates the publish flow that prefers
 * presigned-direct-upload when the source/WASM is large, and falls back to
 * the regular JSON publish endpoint otherwise. Surfaces per-stage progress
 * for the dashboard's progress UI.
 */
export function useRegistryPublish(authorDefault = ''): UseRegistryPublishReturn {
  const [state, setState] = useState<PublishFormState>({
    ...DEFAULT_FORM_STATE,
    author: authorDefault,
  });
  const [progress, setProgress] = useState<PublishProgress>(INITIAL_PROGRESS);
  const [result, setResult] = useState<UseRegistryPublishReturn['result']>(null);
  const [error, setError] = useState<string | null>(null);

  const isPublishing = useMemo(
    () => progress.stage !== 'idle' && progress.stage !== 'done' && progress.stage !== 'error',
    [progress.stage]
  );

  const reset = useCallback(() => {
    setState({ ...DEFAULT_FORM_STATE, author: authorDefault });
    setProgress(INITIAL_PROGRESS);
    setResult(null);
    setError(null);
  }, [authorDefault]);

  const publish = useCallback(
    async (args: {
      sourceCode: string;
      sourceBlob?: Blob;
      wasmBase64?: string;
      wasmBlob?: Blob;
      readme?: string;
    }): Promise<boolean> => {
      // Validate locally before hitting the server.
      if (!state.author.trim() || !state.name.trim() || !state.version.trim()) {
        setError('author, name, and version are required');
        setProgress({ stage: 'error', percent: 100, message: 'missing required fields' });
        return false;
      }

      // Validate name pattern: lowercase alphanumeric and hyphens only
      if (!/^[a-z0-9-]+$/.test(state.name.trim())) {
        setError('Name must be lowercase alphanumeric with hyphens only (e.g. my-function)');
        setProgress({ stage: 'error', percent: 100, message: 'invalid name format' });
        return false;
      }

      // Validate semver format
      if (!/^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$/.test(state.version.trim())) {
        setError('Version must be a valid semver (e.g. 1.0.0)');
        setProgress({ stage: 'error', percent: 100, message: 'invalid version format' });
        return false;
      }

      if (!args.sourceCode.trim() && !args.sourceBlob) {
        setError('source code is required');
        setProgress({ stage: 'error', percent: 100, message: 'missing source code' });
        return false;
      }

      setError(null);
      setResult(null);

      const manifest = buildManifest(state);
      const totalStages =
        (args.wasmBlob || args.wasmBase64 ? 1 : 0) +
        (args.sourceBlob && args.sourceCode.length > 256 * 1024 ? 1 : 0) +
        (args.readme ? 1 : 0) +
        1;
      let stageIdx = 0;
      const advance = (stage: PublishStage, message: string) => {
        stageIdx++;
        setProgress({
          stage,
          percent: Math.round((stageIdx / Math.max(totalStages, 1)) * 100),
          message,
        });
      };

      // Build changelog if user provided one
      const hasChangelog = state.changelogTitle.trim() || state.changelogDescription.trim();
      const changelog = hasChangelog ? {
        category: state.changelogCategory,
        title: state.changelogTitle.trim() || 'Updated',
        description: state.changelogDescription.trim(),
        changes: [],
      } : undefined;

      try {
        // Decide whether to use presigned-direct-upload. Threshold is 256 KiB;
        // matches the server's PresignSmallThreshold.
        const PRESIGN_THRESHOLD = 256 * 1024;
        const useDirectUpload =
          (args.wasmBlob && args.wasmBlob.size > PRESIGN_THRESHOLD) ||
          (args.sourceBlob && args.sourceBlob.size > PRESIGN_THRESHOLD) ||
          (args.readme && args.readme.length > PRESIGN_THRESHOLD);

        if (useDirectUpload) {
          // Per-stage progress callback bridges the upload's per-file events
          // into our overall progress bar.
          const publishResult = await registryApi.publishFunctionViaPresigned({
            author: state.author,
            name: state.name,
            version: state.version,
            manifest,
            source: {
              code: args.sourceCode,
              codeBlob: args.sourceBlob,
            },
            readme: args.readme,
            wasm: args.wasmBlob
              ? { blob: args.wasmBlob }
              : args.wasmBase64
                ? { binaryBase64: args.wasmBase64 }
                : undefined,
            conflictStrategy: state.conflictStrategy,
            changelog,
            onProgress: (stage, loaded, total) => {
              const localPct = total > 0 ? Math.round((loaded / total) * 100) : 0;
              setProgress({
                stage,
                percent: Math.round((stageIdx / totalStages) * 100) + Math.round(localPct / totalStages),
                message: `${stage} uploading ${formatBytes(loaded)}/${formatBytes(total)}`,
              });
            },
          });

          advance('publish', 'finalising publish');
          setResult({
            ok: publishResult.ok,
            function: `${state.author}/${state.name}`,
            function_id: publishResult.function_id,
            version: state.version,
          });
          setProgress({ stage: 'done', percent: 100, message: 'published' });
          toast.success('Function published');
          return true;
        }

        // Small-payload path: single JSON publish.
        if (args.wasmBase64 || args.wasmBlob) {
          advance('wasm', 'preparing WASM');
        }
        if (args.readme) {
          advance('readme', 'preparing README');
        }
        advance('source', 'preparing source');

        const publishResult = await registryApi.publishFunction({
          author: state.author,
          name: state.name,
          version: state.version,
          manifest,
          source: {
            code: args.sourceCode,
            language: state.runtime,
            wasmBinary: args.wasmBase64,
          },
          readme: args.readme,
          conflictStrategy: state.conflictStrategy,
          changelog,
        });

        advance('publish', 'finalising publish');
        setResult({
          ok: publishResult.ok,
          function: `${state.author}/${state.name}`,
          function_id: publishResult.function_id,
          version: state.version,
        });
        setProgress({ stage: 'done', percent: 100, message: 'published' });
        toast.success('Function published');
        return true;
      } catch (e: unknown) {
        const friendlyMessage = mapApiErrorToUserMessage(e);
        setError(friendlyMessage);
        setProgress({ stage: 'error', percent: 100, message: friendlyMessage });
        toast.error(`Publish failed: ${friendlyMessage}`);
        return false;
      }
    },
    [state]
  );

  return {
    state,
    setState,
    progress,
    isPublishing,
    result,
    error,
    publish,
    reset,
  };
}

export function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  if (n < 1024 * 1024 * 1024) return `${(n / (1024 * 1024)).toFixed(1)} MB`;
  return `${(n / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

export interface ApiErrorResponse {
  code?: string;
  message?: string;
}

export function mapApiErrorToUserMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const axiosErr = e as { response?: { data?: ApiErrorResponse; status?: number } };
    const data = axiosErr.response?.data;
    const status = axiosErr.response?.status;

    if (status === 409) {
      return 'A version with this name and version already exists. Choose "Create new version" or "Overwrite existing" from the conflict strategy below.';
    }
    if (status === 402) {
      return data?.message || 'Payment required. Please add funds to your wallet.';
    }
    if (status === 429) {
      return 'Too many requests. Please wait a moment and try again.';
    }
    if (status === 401) {
      return 'Session expired. Please log in again.';
    }
    if (status === 403) {
      return 'You do not have permission to publish. Contact support if you believe this is an error.';
    }

    if (data?.code === 'QUOTA_EXCEEDED') {
      return 'Storage quota exceeded. Please upgrade your plan or delete unused functions.';
    }
    if (data?.code === 'RESOURCE_EXHAUSTED') {
      return 'Resource limit reached. Please upgrade your plan.';
    }
    if (data?.code === 'BILLING_ERROR') {
      return data?.message || 'Billing error. Please check your wallet balance.';
    }
    if (data?.code === 'VALIDATION_ERROR') {
      return data?.message || 'Validation failed. Please check your inputs.';
    }
    if (data?.message) {
      if (
        data.message.toLowerCase().includes('insufficient wallet') ||
        data.message.toLowerCase().includes('balance')
      ) {
        return `Insufficient wallet balance. Please add funds to your wallet.`;
      }
      if (data.message.toLowerCase().includes('semver') || data.message.toLowerCase().includes('version')) {
        return `Invalid version format: ${data.message}`;
      }
      if (data.message.toLowerCase().includes('name')) {
        return `Invalid name: ${data.message}`;
      }
      return data.message;
    }
  }
  if (e instanceof Error) {
    if (e.message.toLowerCase().includes('network') || e.message.toLowerCase().includes('fetch')) {
      return 'Network error. Please check your connection and try again.';
    }
    return e.message;
  }
  return 'Publish failed. Please try again.';
}
