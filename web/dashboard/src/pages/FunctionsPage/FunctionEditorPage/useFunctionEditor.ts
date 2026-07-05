import { functionsApi } from '@/api';
import { fetchDeployBackendOptions } from '@/api/apps';
import { apiClient } from '@/api/client';
import { providersApi } from '@/api/providers';
import { vaultApi } from '@/api/vault';
import { useVaultSecrets } from '@/hooks/useVault';
import type { DeployFunctionRequest, TestFunctionRequest } from '@/types';
import type { SecretMetadata } from '@/types/vault';
import { useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { CODE_TEMPLATES, DRAFT_KEY, RUNTIME_VERSIONS } from './constants';
import type {
  DeploymentLog,
  EnvironmentVariable,
  HttpTrigger,
  ResourceLimits,
  RetryPolicy,
  Runtime,
  ScheduleTrigger,
  Visibility,
} from './types';
import { slugify } from './utils';
import { validateCode, type ValidationIssue } from './utils/codeValidation';

export function useFunctionEditor() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEditing = !!id;

  const { data: connectedProviders } = useQuery({
    queryKey: ['providers'],
    queryFn: () => providersApi.getConnectedProviders(),
  });
  const providers = (connectedProviders ?? []).map((p) => {
    const prov = p as typeof p & { provider_type?: string; region?: string };
    const slug = prov.provider_type ?? prov.name ?? prov.id;
    return { id: slug, name: prov.name, regions: prov.region ? [prov.region] : ['global'] };
  });

  const { data: vaultSecrets } = useVaultSecrets();

  const { data: deployBackendOptions = [], isLoading: deployBackendsLoading } = useQuery({
    queryKey: ['deploy-backend-options'],
    queryFn: fetchDeployBackendOptions,
    staleTime: 60_000,
  });

  const [functionName, setFunctionName] = useState('');
  const [slug, setSlug] = useState('');
  const [description, setDescription] = useState('');
  const [runtime, setRuntime] = useState<Runtime>('typescript');
  const [runtimeVersion, setRuntimeVersion] = useState('ES2022');
  const [code, setCode] = useState(CODE_TEMPLATES.typescript);
  const [visibility, setVisibility] = useState<Visibility>('private');
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');

  const [selectedProviders, setSelectedProviders] = useState<string[]>([]);
  const [selectedRegion, setSelectedRegion] = useState('');

  /** When set, deploy backends are limited to this app (matches API function.app_id) */
  const [linkedAppId, setLinkedAppId] = useState<string | null>(null);
  const [selectedDeployBackendId, setSelectedDeployBackendId] = useState('');

  const filteredDeployBackends = useMemo(() => {
    if (!linkedAppId) return deployBackendOptions;
    return deployBackendOptions.filter((b) => b.appId === linkedAppId);
  }, [deployBackendOptions, linkedAppId]);

  const [envVars, setEnvVars] = useState<EnvironmentVariable[]>([]);
  const [newEnvKey, setNewEnvKey] = useState('');
  const [newEnvValue, setNewEnvValue] = useState('');
  const [isNewEnvSecret, setIsNewEnvSecret] = useState(false);
  const [showEnvValues, setShowEnvValues] = useState<Record<string, boolean>>({});

  const [resources, setResources] = useState<ResourceLimits>({
    memoryMb: 128,
    timeoutMs: 30000,
    maxConcurrency: 10,
  });

  const [httpTrigger, setHttpTrigger] = useState<HttpTrigger>({
    enabled: true,
    method: 'ANY',
    path: '/',
  });
  const [scheduleTrigger, setScheduleTrigger] = useState<ScheduleTrigger>({
    enabled: false,
    cron: '0 * * * *',
    timezone: 'UTC',
  });

  const [retryPolicy, setRetryPolicy] = useState<RetryPolicy>({
    maxRetries: 3,
    backoffMs: 1000,
    backoffStrategy: 'exponential',
  });
  const [warmInstances, setWarmInstances] = useState(0);
  const [showDraftRestorePrompt, setShowDraftRestorePrompt] = useState(false);

  const [activeTab, setActiveTab] = useState('editor');
  const [isDeploying, setIsDeploying] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [isDirty, setIsDirty] = useState(false);
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [draftTimestamp, setDraftTimestamp] = useState<Date | null>(null);
  const [currentDeploymentId, setCurrentDeploymentId] = useState<string | null>(null);
  const [deploymentStatus, setDeploymentStatus] = useState<string | null>(null);
  const [postCreateFunctionId, setPostCreateFunctionId] = useState<string | null>(null);
  const [isPublishedToRegistry, setIsPublishedToRegistry] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    try {
      const saved = sessionStorage.getItem('functionfly:editor-layout');
      if (saved) {
        const layout = JSON.parse(saved);
        if (layout.activeTab) setActiveTab(layout.activeTab);
      }
    } catch {
      /* ignore */
    }
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    try {
      sessionStorage.setItem(
        'functionfly:editor-layout',
        JSON.stringify({ activeTab })
      );
    } catch {
      /* ignore */
    }
  }, [activeTab]);
  const [logs, setLogs] = useState<DeploymentLog[]>([
    {
      id: '1',
      timestamp: new Date().toISOString(),
      level: 'info',
      message: 'Editor initialized',
    },
  ]);

  const [vaultPickerOpen, setVaultPickerOpen] = useState(false);
  const [pickingSecretId, setPickingSecretId] = useState<string | null>(null);
  const [pendingSecretForDecrypt, setPendingSecretForDecrypt] = useState<SecretMetadata | null>(
    null
  );
  const [vaultDecryptPassphrase, setVaultDecryptPassphrase] = useState('');
  const [revealEnvVarId, setRevealEnvVarId] = useState<string | null>(null);
  const [revealGateOpen, setRevealGateOpen] = useState(false);

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Code validation state
  const [validationIssues, setValidationIssues] = useState<ValidationIssue[]>([]);
  const [showValidationPanel, setShowValidationPanel] = useState(false);

  // Test panel states
  const [testInput, setTestInput] = useState<string>('{}');
  const [testResult, setTestResult] = useState<{
    success: boolean;
    output?: unknown;
    error?: string;
    executionTimeMs?: number;
    coldStartMs?: number;
    logs?: { level: DeploymentLog['level']; message: string }[];
  } | null>(null);
  const [testTab, setTestTab] = useState<'input' | 'output'>('input');

  const slugManuallyEdited = useRef(false);
  const autoSaveTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const tagInputRef = useRef<HTMLInputElement>(null);
  const isMountedRef = useRef(true);
  const deployRef = useRef(false);
  const testAbortRef = useRef<AbortController | null>(null);
  const testTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const addLog = useCallback((level: DeploymentLog['level'], message: string) => {
    setLogs((prev) => [
      {
        id: Date.now().toString(),
        timestamp: new Date().toISOString(),
        level,
        message,
      },
      ...prev,
    ]);
  }, []);

  const cancelTest = useCallback(() => {
    if (testAbortRef.current) {
      testAbortRef.current.abort();
      testAbortRef.current = null;
    }
    if (testTimeoutRef.current) {
      clearTimeout(testTimeoutRef.current);
      testTimeoutRef.current = undefined;
    }
    setIsTesting(false);
    addLog('warn', 'Test cancelled');
  }, [addLog]);

  const markDirty = useCallback(() => setIsDirty(true), []);

  useEffect(() => {
    const base = 'FunctionFly';
    if (isEditing && functionName) document.title = `Edit ${functionName} | ${base}`;
    else
      document.title = functionName ? `New: ${functionName} | ${base}` : `New Function | ${base}`;
    return () => {
      document.title = base;
    };
  }, [isEditing, functionName]);

  useEffect(() => {
    if (isEditing) return;
    try {
      const saved = localStorage.getItem(DRAFT_KEY);
      if (saved) {
        const d = JSON.parse(saved);
        // Show restore prompt if there's a meaningful draft
        if (d.functionName || d.code) {
          setShowDraftRestorePrompt(true);
          if (d.savedAt) {
            setDraftTimestamp(new Date(d.savedAt));
          }
        }
      }
    } catch {
      /* ignore */
    }
  }, [isEditing]);

  const handleRestoreDraft = useCallback(() => {
    try {
      const saved = localStorage.getItem(DRAFT_KEY);
      if (!saved) return;
      const d = JSON.parse(saved);
      if (d.functionName) setFunctionName(d.functionName);
      if (d.slug) {
        setSlug(slugify(d.slug));
        slugManuallyEdited.current = true;
      }
      if (d.description) setDescription(d.description);
      if (d.runtime) {
        setRuntime(d.runtime);
        setCode(d.code ?? CODE_TEMPLATES[d.runtime as Runtime]);
      }
      if (d.runtimeVersion) setRuntimeVersion(d.runtimeVersion);
      if (d.code) setCode(d.code);
      if (d.visibility) setVisibility(d.visibility);
      if (d.tags) setTags(d.tags);
      if (d.envVars) setEnvVars(d.envVars);
      if (d.resources) setResources(d.resources);
      if (d.httpTrigger) setHttpTrigger(d.httpTrigger);
      if (d.scheduleTrigger) setScheduleTrigger(d.scheduleTrigger);
      if (d.retryPolicy) setRetryPolicy({ backoffStrategy: 'exponential', ...d.retryPolicy });
      if (typeof d.warmInstances === 'number') setWarmInstances(d.warmInstances);
      setShowDraftRestorePrompt(false);
      toast.success('Draft restored');
    } catch {
      toast.error('Failed to restore draft');
    }
  }, []);

  const handleDiscardDraft = useCallback(() => {
    localStorage.removeItem(DRAFT_KEY);
    setShowDraftRestorePrompt(false);
  }, []);

  useEffect(() => {
    if (isEditing || !isDirty || isDeploying || isTesting) return;
    if (autoSaveTimer.current) clearTimeout(autoSaveTimer.current);
    autoSaveTimer.current = setTimeout(() => {
      let payload: string;
      try {
        payload = JSON.stringify({
          functionName,
          slug,
          description,
          runtime,
          runtimeVersion,
          code,
          visibility,
          tags,
          envVars,
          resources,
          httpTrigger,
          scheduleTrigger,
          retryPolicy,
          warmInstances,
          savedAt: new Date().toISOString(),
        });
      } catch {
        toast.error('Draft autosave failed: unable to serialize state.');
        return;
      }
      if (payload.length > 4 * 1024 * 1024) {
        toast.error('Draft is too large to autosave (~4MB localStorage limit). Consider saving or trimming code.');
        return;
      }
      try {
        localStorage.setItem(DRAFT_KEY, payload);
        setLastSaved(new Date());
      } catch (err) {
        if (err instanceof DOMException && (err.code === 22 || err.name === 'QuotaExceededError')) {
          toast.error('Draft autosave failed: localStorage quota exceeded. Consider saving or trimming code.');
        } else if (err instanceof Error) {
          toast.error('Draft autosave failed: ' + err.message);
        } else {
          toast.error('Draft autosave failed: ' + String(err));
        }
      }
    }, 1500);
    return () => {
      if (autoSaveTimer.current) clearTimeout(autoSaveTimer.current);
    };
  }, [
    isEditing,
    isDirty,
    functionName,
    slug,
    description,
    runtime,
    runtimeVersion,
    code,
    visibility,
    tags,
    envVars,
    resources,
    httpTrigger,
    scheduleTrigger,
    retryPolicy,
    warmInstances,
  ]);

  useEffect(() => {
    if (selectedProviders.length === 1 && !selectedRegion) {
      const p = providers.find((p) => p.id === selectedProviders[0]);
      if (p?.regions?.length === 1) setSelectedRegion(p.regions[0]);
    }
  }, [selectedProviders, providers, selectedRegion]);

  useEffect(() => {
    if (filteredDeployBackends.length === 0) {
      setSelectedDeployBackendId('');
      return;
    }
    const ok = filteredDeployBackends.some((b) => b.id === selectedDeployBackendId);
    if (!selectedDeployBackendId || !ok) {
      setSelectedDeployBackendId(filteredDeployBackends[0].id);
    }
  }, [filteredDeployBackends, selectedDeployBackendId]);

  useEffect(() => {
    if (!isEditing || !id) return;
    setCurrentDeploymentId(null);
    setDeploymentStatus(null);
    const load = async () => {
      try {
        setPostCreateFunctionId(null);
        addLog('info', `Loading function ${id}…`);
        const fn = await functionsApi.get(id);
        setFunctionName(fn.name);
        setSlug(slugify(fn.name));
        setLinkedAppId(fn.appId ?? null);
        setSelectedProviders(fn.providers);
        setSelectedRegion(fn.region);
        setCode(fn.code);
        setEnvVars(
          fn.envVars.map((e, i) => ({
            id: `env-${i}`,
            key: e.key ?? '',
            value: e.value ?? '',
            isSecret: e.isSecret ?? false,
          }))
        );
        addLog('success', 'Function loaded');
      } catch (err) {
        addLog('error', `Failed to load: ${err}`);
      }
    };
    load();
  }, [isEditing, id, addLog]);

  useEffect(() => {
    isMountedRef.current = true;
    if (!currentDeploymentId) return;
    const controller = new AbortController();
    const poll = async () => {
      try {
        if (!isMountedRef.current) return;
        const data = await apiClient.get<{ deployment: { status: string } }>(
          `/v1/functions/deployments/${currentDeploymentId}`,
          { signal: controller.signal }
        );
        if (!isMountedRef.current) return;
        const status = data.deployment?.status;
        if (status !== deploymentStatus) {
          setDeploymentStatus(status);
          addLog('info', `Deployment status: ${status}`);
          if (status === 'success') {
            addLog('success', 'Deployment completed!');
            setCurrentDeploymentId(null);
          } else if (status === 'failed') {
            addLog('error', 'Deployment failed');
            setCurrentDeploymentId(null);
          }
        }
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          /* ignore poll errors */
        }
      }
    };
    poll();
    const interval = setInterval(poll, 5000);
    return () => {
      isMountedRef.current = false;
      controller.abort();
      clearInterval(interval);
    };
  }, [currentDeploymentId, deploymentStatus, addLog]);

  const handleNameChange = useCallback(
    (name: string) => {
      setFunctionName(name);
      if (!slugManuallyEdited.current) setSlug(slugify(name));
      markDirty();
      setErrors((e) => {
        const next = { ...e };
        delete next.name;
        return next;
      });
    },
    [markDirty]
  );

  const handleSlugChange = useCallback(
    (s: string) => {
      slugManuallyEdited.current = true;
      setSlug(s);
      markDirty();
    },
    [markDirty]
  );

  // Smart defaults based on runtime characteristics
  const getRuntimeDefaults = useCallback((r: Runtime) => {
    switch (r) {
      case 'python':
        return { memoryMb: 256, timeoutMs: 30000 }; // Python needs more memory for imports
      case 'python-wasm':
        return { memoryMb: 128, timeoutMs: 15000 }; // MicroPython is lighter
      case 'rust-wasm':
      case 'browser-wasm':
        return { memoryMb: 64, timeoutMs: 5000 }; // WASM is efficient
      case 'go':
        return { memoryMb: 128, timeoutMs: 10000 }; // Go is fast
      case 'deno':
      case 'bun':
        return { memoryMb: 128, timeoutMs: 10000 }; // Modern JS runtimes
      case 'typescript':
      case 'javascript':
      default:
        return { memoryMb: 128, timeoutMs: 30000 }; // Default
    }
  }, []);

  const handleRuntimeChange = useCallback(
    (r: Runtime) => {
      if (isDirty && r !== runtime) {
        if (!window.confirm('You have unsaved changes. Switching runtime will discard them. Continue?')) {
          return;
        }
      }
      setRuntime(r);
      setRuntimeVersion(RUNTIME_VERSIONS[r][0]);
      setCode(CODE_TEMPLATES[r]);
      if (!isEditing) {
        const defaults = getRuntimeDefaults(r);
        setResources((prev) => ({
          ...prev,
          ...defaults,
        }));
      }
      markDirty();
    },
    [markDirty, isEditing, getRuntimeDefaults, isDirty, runtime]
  );

  const handleProviderToggle = useCallback(
    (pid: string) => {
      setSelectedProviders((prev) =>
        prev.includes(pid) ? prev.filter((x) => x !== pid) : [...prev, pid]
      );
      markDirty();
    },
    [markDirty]
  );

  const addEnvironmentVariable = useCallback(() => {
    if (!newEnvKey.trim()) return;
    setEnvVars((prev) => [
      ...prev,
      {
        id: Date.now().toString(),
        key: newEnvKey.trim(),
        value: newEnvValue.trim(),
        isSecret: isNewEnvSecret,
      },
    ]);
    setNewEnvKey('');
    setNewEnvValue('');
    setIsNewEnvSecret(false);
    markDirty();
  }, [newEnvKey, newEnvValue, isNewEnvSecret, markDirty]);

  const removeEnvironmentVariable = useCallback(
    (envId: string) => {
      setEnvVars((prev) => prev.filter((v) => v.id !== envId));
      markDirty();
    },
    [markDirty]
  );

  const addTag = useCallback(() => {
    const t = newTag.trim().toLowerCase();
    if (!t || tags.includes(t)) return;
    setTags((prev) => [...prev, t]);
    setNewTag('');
    markDirty();
    tagInputRef.current?.focus();
  }, [newTag, tags, markDirty]);

  const removeTag = useCallback(
    (t: string) => {
      setTags((prev) => prev.filter((x) => x !== t));
      markDirty();
    },
    [markDirty]
  );

  const handleSelectVaultSecret = useCallback((secret: SecretMetadata) => {
    setPendingSecretForDecrypt(secret);
    setVaultDecryptPassphrase('');
    setVaultPickerOpen(false);
  }, []);

  const handleConfirmVaultPassphrase = useCallback(async () => {
    if (!pendingSecretForDecrypt || !vaultDecryptPassphrase.trim()) return;
    setPickingSecretId(pendingSecretForDecrypt.id);
    try {
      const { value } = await vaultApi.decryptSecret(
        pendingSecretForDecrypt.id,
        vaultDecryptPassphrase.trim()
      );
      setNewEnvKey(pendingSecretForDecrypt.name);
      setNewEnvValue(value);
      setIsNewEnvSecret(true);
      toast.success(`Added "${pendingSecretForDecrypt.name}" from Vault`);
      setPendingSecretForDecrypt(null);
      setVaultDecryptPassphrase('');
    } catch {
      toast.error('Wrong passphrase or decrypt failed.');
    } finally {
      setPickingSecretId(null);
    }
  }, [pendingSecretForDecrypt, vaultDecryptPassphrase]);

  const handleRevealVerified = useCallback((envVar: EnvironmentVariable) => {
    try {
      navigator.clipboard.writeText(envVar.value);
      toast.success('Value copied to clipboard');
    } catch {
      toast.error('Could not copy to clipboard');
    }
    setRevealGateOpen(false);
    setRevealEnvVarId(null);
  }, []);

  const setDeployBackendId = useCallback((v: string) => {
    setSelectedDeployBackendId(v);
    setErrors((e) => {
      const next = { ...e };
      delete next.deployBackend;
      return next;
    });
  }, []);

  const validate = useCallback((): boolean => {
    const errs: Record<string, string> = {};
    if (!functionName.trim()) errs.name = 'Function name is required';
    if (!slug.trim()) errs.slug = 'Slug is required';
    else if (!/^[a-z0-9][a-z0-9-]*$/.test(slug))
      errs.slug = 'Slug must be lowercase letters, numbers, and hyphens';
    if (!code.trim()) errs.code = 'Code is required';
    if (httpTrigger.enabled && !httpTrigger.path.startsWith('/'))
      errs.httpPath = 'Path must start with /';
    if (isEditing) {
      if (!deployBackendsLoading && filteredDeployBackends.length === 0) {
        errs.deployBackend = 'Add a backend under Apps before deploying.';
      } else if (filteredDeployBackends.length > 0 && !selectedDeployBackendId) {
        errs.deployBackend = 'Select a deploy backend.';
      }
    }

    // Add code validation errors
    const validation = validateCode(code, runtime);
    const codeErrors = validation.issues.filter((i) => i.type === 'error');
    if (codeErrors.length > 0) {
      errs.code = `Code validation: ${codeErrors[0].message}`;
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  }, [
    functionName,
    slug,
    code,
    runtime,
    httpTrigger,
    isEditing,
    deployBackendsLoading,
    filteredDeployBackends,
    selectedDeployBackendId,
  ]);

  const handleSaveDraft = useCallback(async () => {
    if (!functionName.trim()) {
      toast.error('Function name is required');
      return;
    }
    setIsSaving(true);
    addLog('info', 'Saving draft…');
    try {
      const payload = {
        name: functionName.trim(),
        providers: selectedProviders.length > 0 ? selectedProviders : ['functionfly-edge'],
        region: selectedRegion || 'global',
        code,
        envVars: envVars.map(({ key, value, isSecret }) => ({ key, value, isSecret })),
      };
      if (isEditing && id) {
        await functionsApi.update(id, payload);
        toast.success('Saved');
      } else {
        const created = await functionsApi.create(payload);
        localStorage.removeItem(DRAFT_KEY);
        toast.success('Draft saved');
        navigate(`/functions/${created.id}/edit`, { replace: true });
      }
      setIsDirty(false);
      setLastSaved(new Date());
      addLog('success', 'Draft saved');
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      addLog('error', `Save failed: ${msg}`);
      toast.error(`Save failed: ${msg}`);
    } finally {
      setIsSaving(false);
    }
  }, [
    functionName,
    selectedProviders,
    selectedRegion,
    code,
    envVars,
    isEditing,
    id,
    navigate,
    addLog,
  ]);

  const handleDeploy = useCallback(async () => {
    if (deployRef.current) return;
    deployRef.current = true;
    if (!validate()) {
      deployRef.current = false;
      toast.error('Please fix the errors before deploying');
      return;
    }
    setIsDeploying(true);
    addLog('info', 'Starting deployment…');
    try {
      let functionId = id ?? null;
      if (!isEditing || !id) {
        const created = await functionsApi.create({
          name: functionName.trim(),
          providers: selectedProviders.length > 0 ? selectedProviders : ['functionfly-edge'],
          region: selectedRegion || 'global',
          code,
          envVars: envVars.map(({ key, value, isSecret }) => ({ key, value, isSecret })),
        });
        functionId = created.id;
        setPostCreateFunctionId(created.id);
        localStorage.removeItem(DRAFT_KEY);
        addLog('success', `Function created: ${created.name}`);
        toast.success('Function created');
        navigate(`/functions/${functionId}/edit`, { replace: true, state: { justCreated: true } });
        deployRef.current = false;
        setIsDeploying(false);
        return;
      }
      const deployPayload: DeployFunctionRequest = {
        functionId: functionId!,
        backendId: selectedDeployBackendId,
      };
      const result = await functionsApi.deploy(deployPayload);
      setCurrentDeploymentId(result.deploymentId);
      setDeploymentStatus('pending');
      addLog('success', `Deployment started: ${result.deploymentId}`);
      toast.success('Deployment started');
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      addLog('error', `Deployment failed: ${msg}`);
      toast.error(`Deployment failed: ${msg}`);
    } finally {
      deployRef.current = false;
      setIsDeploying(false);
    }
  }, [
    validate,
    isEditing,
    id,
    functionName,
    selectedProviders,
    selectedRegion,
    code,
    envVars,
    selectedDeployBackendId,
    navigate,
    addLog,
  ]);

  const handleTest = useCallback(async () => {
    if (isTesting || testAbortRef.current) return;
    setIsTesting(true);
    const controller = new AbortController();
    testAbortRef.current = controller;
    testTimeoutRef.current = setTimeout(() => controller.abort(), 120_000);
    addLog('info', 'Running test…');
    try {
      let parsedInput: Record<string, unknown> = {};
      try {
        parsedInput = JSON.parse(testInput);
      } catch {
        addLog('warn', 'Invalid JSON input, using empty object');
      }

      const testData: TestFunctionRequest = {
        functionId: isEditing ? id : undefined,
        code: isEditing ? undefined : code,
        envVars: envVars.map(({ key, value, isSecret }) => ({ key, value, isSecret })),
        testInput: parsedInput,
      };
      const result = await functionsApi.test(testData);
      setTestResult({
        success: result.success,
        output: result.output,
        error: result.error,
        executionTimeMs: result.executionTimeMs,
        logs: result.logs.map((l) => ({ level: l.level as DeploymentLog['level'], message: l.message })),
      });
      setTestTab('output');
      if (result.success) {
        addLog('success', `Test passed in ${result.executionTimeMs}ms`);
        if (result.output) addLog('info', `Output: ${JSON.stringify(result.output)}`);
      } else {
        addLog('error', `Test failed: ${result.error}`);
      }
      result.logs.forEach((l) => addLog(l.level as DeploymentLog['level'], l.message));
    } catch (err) {
      if (testAbortRef.current === controller) {
        addLog('error', `Test error: ${err}`);
        setTestResult({ success: false, error: String(err) });
      }
    } finally {
      if (testAbortRef.current === controller) {
        setIsTesting(false);
        testAbortRef.current = null;
      }
      if (testTimeoutRef.current) {
        clearTimeout(testTimeoutRef.current);
        testTimeoutRef.current = undefined;
      }
    }
  }, [isEditing, id, code, envVars, testInput, addLog]);

  // Keyboard shortcuts
  const [keyboardShortcutsOpen, setKeyboardShortcutsOpen] = useState(false);

  const isLoading = isSaving || isDeploying || isTesting;

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      const inInput = ['INPUT', 'TEXTAREA', 'SELECT'].includes(tag);

      // Save shortcut: Ctrl/Cmd + S
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        if (inInput) return;
        e.preventDefault();
        void handleSaveDraft();
        return;
      }

      if (inInput) return;

      // Test shortcut: Ctrl/Cmd + Enter
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        if (!isLoading) void handleTest();
        return;
      }

      // Deploy shortcut: Ctrl/Cmd + Shift + Enter
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'Enter') {
        e.preventDefault();
        if (!isLoading) void handleDeploy();
        return;
      }

      // Help shortcut: ?
      if (e.key === '?' && !inInput) {
        e.preventDefault();
        setKeyboardShortcutsOpen(true);
        return;
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [handleSaveDraft, handleTest, handleDeploy, isLoading]);

  // Code validation effect

  const editor = useMemo(() => ({
    navigate,
    id,
    isEditing,
    providers,
    vaultSecrets,
    functionName,
    setFunctionName,
    slug,
    setSlug,
    description,
    setDescription,
    runtime,
    setRuntime,
    runtimeVersion,
    setRuntimeVersion,
    code,
    setCode,
    visibility,
    setVisibility,
    tags,
    setTags,
    newTag,
    setNewTag,
    selectedProviders,
    setSelectedProviders,
    selectedRegion,
    setSelectedRegion,
    linkedAppId,
    setLinkedAppId,
    selectedDeployBackendId,
    setSelectedDeployBackendId,
    filteredDeployBackends,
    envVars,
    setEnvVars,
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
    warmInstances,
    setWarmInstances,
    showDraftRestorePrompt,
    activeTab,
    setActiveTab,
    isDeploying,
    isSaving,
    isTesting,
    isDirty,
    lastSaved,
    draftTimestamp,
    currentDeploymentId,
    setCurrentDeploymentId,
    deploymentStatus,
    postCreateFunctionId,
    setPostCreateFunctionId,
    isPublishedToRegistry,
    setIsPublishedToRegistry,
    logs,
    vaultPickerOpen,
    setVaultPickerOpen,
    pickingSecretId,
    setPickingSecretId,
    pendingSecretForDecrypt,
    setPendingSecretForDecrypt,
    vaultDecryptPassphrase,
    setVaultDecryptPassphrase,
    revealEnvVarId,
    setRevealEnvVarId,
    revealGateOpen,
    setRevealGateOpen,
    errors,
    validationIssues,
    setValidationIssues,
    showValidationPanel,
    setShowValidationPanel,
    keyboardShortcutsOpen,
    setKeyboardShortcutsOpen,
    tagInputRef,
    handleRestoreDraft,
    handleDiscardDraft,
    handleNameChange,
    handleSlugChange,
    handleRuntimeChange,
    handleProviderToggle,
    addEnvironmentVariable,
    removeEnvironmentVariable,
    addTag,
    removeTag,
    handleSelectVaultSecret,
    handleConfirmVaultPassphrase,
    handleRevealVerified,
    markDirty,
    handleSaveDraft,
    handleDeploy,
    handleTest,
    cancelTest,
    testInput,
    setTestInput,
    testResult,
    testTab,
    setTestTab,
    deployBackendsLoading,
    setDeployBackendId,
    isLoading,
  }), [
    isEditing,
    providers,
    vaultSecrets,
    functionName,
    slug,
    description,
    runtime,
    runtimeVersion,
    code,
    visibility,
    tags,
    newTag,
    selectedProviders,
    selectedRegion,
    linkedAppId,
    selectedDeployBackendId,
    filteredDeployBackends,
    envVars,
    newEnvKey,
    newEnvValue,
    isNewEnvSecret,
    showEnvValues,
    resources,
    httpTrigger,
    scheduleTrigger,
    retryPolicy,
    warmInstances,
    showDraftRestorePrompt,
    activeTab,
    isDeploying,
    isSaving,
    isTesting,
    isDirty,
    lastSaved,
    draftTimestamp,
    currentDeploymentId,
    deploymentStatus,
    postCreateFunctionId,
    isPublishedToRegistry,
    logs,
    vaultPickerOpen,
    pickingSecretId,
    pendingSecretForDecrypt,
    vaultDecryptPassphrase,
    revealEnvVarId,
    revealGateOpen,
    errors,
    validationIssues,
    showValidationPanel,
    keyboardShortcutsOpen,
    testInput,
    testResult,
    testTab,
    deployBackendsLoading,
    linkedAppId,
    filteredDeployBackends,
    selectedDeployBackendId,
    deploymentStatus,
    isPublishedToRegistry,
    postCreateFunctionId,
    isLoading,
    handleRestoreDraft,
    handleDiscardDraft,
    handleNameChange,
    handleSlugChange,
    handleRuntimeChange,
    handleProviderToggle,
    addEnvironmentVariable,
    removeEnvironmentVariable,
    addTag,
    removeTag,
    handleSelectVaultSecret,
    handleConfirmVaultPassphrase,
    handleRevealVerified,
    markDirty,
    handleSaveDraft,
    handleDeploy,
    handleTest,
    cancelTest,
    setDeployBackendId,
    setIsPublishedToRegistry,
    setPostCreateFunctionId,
  ]);

  return editor;
}

export type FunctionEditorModel = ReturnType<typeof useFunctionEditor>;
