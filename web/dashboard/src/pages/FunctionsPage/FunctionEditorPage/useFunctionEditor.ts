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
  const [logs, setLogs] = useState<DeploymentLog[]>([
    {
      id: '1',
      timestamp: new Date().toLocaleString(),
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
    logs?: { level: string; message: string }[];
  } | null>(null);
  const [testTab, setTestTab] = useState<'input' | 'output'>('input');

  const slugManuallyEdited = useRef(false);
  const autoSaveTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const tagInputRef = useRef<HTMLInputElement>(null);

  const addLog = useCallback((level: DeploymentLog['level'], message: string) => {
    setLogs((prev) => [
      ...prev,
      {
        id: Date.now().toString(),
        timestamp: new Date().toLocaleString(),
        level,
        message,
      },
    ]);
  }, []);

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
        setSlug(d.slug);
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
    if (isEditing || !isDirty) return;
    if (autoSaveTimer.current) clearTimeout(autoSaveTimer.current);
    autoSaveTimer.current = setTimeout(() => {
      try {
        localStorage.setItem(
          DRAFT_KEY,
          JSON.stringify({
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
          })
        );
        setLastSaved(new Date());
      } catch {
        /* ignore */
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
    const load = async () => {
      try {
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
    if (!currentDeploymentId) return;
    const poll = async () => {
      try {
        const data = await apiClient.get<{ deployment: { status: string } }>(
          `/v1/functions/deployments/${currentDeploymentId}`
        );
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
      } catch {
        /* ignore */
      }
    };
    poll();
    const interval = setInterval(poll, 5000);
    return () => clearInterval(interval);
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
      setRuntime(r);
      setRuntimeVersion(RUNTIME_VERSIONS[r][0]);
      setCode(CODE_TEMPLATES[r]);
      // Apply smart defaults for new functions only
      if (!isEditing) {
        const defaults = getRuntimeDefaults(r);
        setResources((prev) => ({
          ...prev,
          ...defaults,
        }));
      }
      markDirty();
    },
    [markDirty, isEditing, getRuntimeDefaults]
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
    if (!validate()) {
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
        localStorage.removeItem(DRAFT_KEY);
        addLog('success', `Function created: ${created.name}`);
        toast.success('Function created');
        navigate(`/functions/${functionId}/edit`, { replace: true, state: { justCreated: true } });
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
    setIsTesting(true);
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
      setTestResult(result);
      setTestTab('output');
      if (result.success) {
        addLog('success', `Test passed in ${result.executionTimeMs}ms`);
        if (result.output) addLog('info', `Output: ${JSON.stringify(result.output)}`);
      } else {
        addLog('error', `Test failed: ${result.error}`);
      }
      result.logs.forEach((l) => addLog(l.level as DeploymentLog['level'], l.message));
    } catch (err) {
      addLog('error', `Test error: ${err}`);
      setTestResult({ success: false, error: String(err) });
    } finally {
      setIsTesting(false);
    }
  }, [isEditing, id, code, envVars, testInput, addLog]);

  // Keyboard shortcuts
  const [keyboardShortcutsOpen, setKeyboardShortcutsOpen] = useState(false);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      // Save shortcut
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        void handleSaveDraft();
      }

      // Test shortcut: Ctrl/Cmd + Enter
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        void handleTest();
      }

      // Deploy shortcut: Ctrl/Cmd + Shift + Enter
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'Enter') {
        e.preventDefault();
        void handleDeploy();
      }

      // Help shortcut: ? (when not in an input)
      if (e.key === '?' && !['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement).tagName)) {
        e.preventDefault();
        setKeyboardShortcutsOpen(true);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [handleSaveDraft, handleTest, handleDeploy]);

  // Code validation effect
  useEffect(() => {
    const timer = setTimeout(() => {
      const result = validateCode(code, runtime);
      setValidationIssues(result.issues);
    }, 1000); // Debounce validation

    return () => clearTimeout(timer);
  }, [code, runtime]);

  const isLoading = isSaving || isDeploying || isTesting;

  return {
    navigate,
    id,
    isEditing,
    providers,
    vaultSecrets,
    functionName,
    setFunctionName,
    slug,
    description,
    setDescription,
    runtime,
    runtimeVersion,
    setRuntimeVersion,
    code,
    setCode,
    visibility,
    setVisibility,
    tags,
    newTag,
    setNewTag,
    selectedProviders,
    setSelectedProviders,
    selectedRegion,
    setSelectedRegion,
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
    activeTab,
    setActiveTab,
    isDeploying,
    isSaving,
    isTesting,
    isLoading,
    isDirty,
    lastSaved,
    draftTimestamp,
    logs,
    vaultPickerOpen,
    setVaultPickerOpen,
    pickingSecretId,
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
    showValidationPanel,
    setShowValidationPanel,
    keyboardShortcutsOpen,
    setKeyboardShortcutsOpen,
    tagInputRef,
    showDraftRestorePrompt,
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
    testInput,
    setTestInput,
    testResult,
    testTab,
    setTestTab,
    linkedAppId,
    deployBackendsLoading,
    filteredDeployBackends,
    selectedDeployBackendId,
    setDeployBackendId,
  };
}

export type FunctionEditorModel = ReturnType<typeof useFunctionEditor>;
