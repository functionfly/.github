import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { useGitHubRepo, useScanGitHubRepo, useGitHubBranches } from '@/hooks/useGitHubRepos';
import { usePreviewImport, useStartImport, useImportProgress } from '@/hooks/useGitHubImport';
import { useGitHubStore } from '@/stores/githubStore';
import type { DetectedFunction, ScanResult, ImportPreview, StartImportRequest, Branch, GitHubRepo, ImportProgressEvent, ImportCompleteEvent, ImportErrorEvent } from '@/types/github';
import { AnimatePresence, motion } from 'framer-motion';
import {
  AlertTriangle,
  ArrowLeft,
  ArrowRight,
  Check,
  CheckCircle,
  Clock,
  Eye,
  FileCode,
  GitBranch,
  Globe,
  Hash,
  Loader2,
  Lock,
  Play,
  RefreshCw,
  Rocket,
  Scan,
  Shield,
  Star,
  X,
  Zap,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

const STEPS = [
  { id: 1, label: 'Scan', icon: Scan },
  { id: 2, label: 'Configure', icon: FileCode },
  { id: 3, label: 'Confirm', icon: Eye },
  { id: 4, label: 'Progress', icon: Rocket },
] as const;

type StepId = 1 | 2 | 3 | 4;

export default function GitHubRepoImportPage() {
  const { repoId } = useParams<{ repoId: string }>();
  const navigate = useNavigate();

  const {
    scanResult,
    setScanResult,
    importConfig,
    setSelectedFunctions,
    toggleFunction,
    setGlobalVisibility,
    setAutoSync,
    setSyncBranches,
    setEnvironmentMappings,
    activeImportId,
    setActiveImportId,
    resetImportConfig,
  } = useGitHubStore();

  const [currentStep, setCurrentStep] = useState<StepId>(
    activeImportId ? 4 : 1
  );
  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [preview, setPreview] = useState<ImportPreview | null>(null);

  const { data: repo, isLoading: repoLoading } = useGitHubRepo(repoId || '');
  const { data: branches } = useGitHubBranches(repoId || '');
  const scanMutation = useScanGitHubRepo(repoId || '');
  const previewMutation = usePreviewImport();
  const startImportMutation = useStartImport();
  const { progress, complete, error: progressError, status: progressStatus } = useImportProgress(activeImportId);

  useEffect(() => {
    if (!repoId) {
      navigate('/github', { replace: true });
    }
  }, [repoId, navigate]);

  const handleScan = useCallback(() => {
    if (!repoId) return;
    scanMutation.mutate(undefined, {
      onSuccess: (data) => {
        setScanResult(data);
      },
    });
  }, [repoId, scanMutation, setScanResult]);

  const handleAutoSelect = useCallback(
    (result: ScanResult) => {
      const highConfidence = result.functions
        .filter((f) => f.confidence >= 0.7)
        .map((f) => f.name);
      setSelectedFunctions(highConfidence);
    },
    [setSelectedFunctions]
  );

  useEffect(() => {
    if (scanResult && importConfig.selectedFunctions.length === 0) {
      handleAutoSelect(scanResult);
    }
  }, [scanResult, handleAutoSelect, importConfig.selectedFunctions.length]);

  const handlePreview = useCallback(() => {
    if (!repoId || !scanResult) return;
    const primaryFunction = scanResult.functions.find((f) =>
      importConfig.selectedFunctions.includes(f.name)
    );
    if (!primaryFunction) return;

    const request: StartImportRequest = {
      connection_id: '',
      repo_id: repoId,
      source_branch: repo?.default_branch || 'main',
      source_path: primaryFunction.sub_directory || '/',
      function_name: primaryFunction.name,
      visibility: importConfig.globalVisibility,
      auto_sync_enabled: importConfig.autoSync,
      sync_branches: importConfig.syncBranches.length > 0 ? importConfig.syncBranches : undefined,
      environment_mappings:
        Object.keys(importConfig.environmentMappings).length > 0
          ? importConfig.environmentMappings
          : undefined,
    };

    previewMutation.mutate(request, {
      onSuccess: (data) => {
        setPreview(data);
        setCurrentStep(3);
      },
    });
  }, [repoId, scanResult, importConfig, repo, previewMutation]);

  const handleStartImport = useCallback(() => {
    if (!repoId || !scanResult) return;
    const primaryFunction = scanResult.functions.find((f) =>
      importConfig.selectedFunctions.includes(f.name)
    );
    if (!primaryFunction) return;

    const request: StartImportRequest = {
      connection_id: '',
      repo_id: repoId,
      source_branch: repo?.default_branch || 'main',
      source_path: primaryFunction.sub_directory || '/',
      function_name: primaryFunction.name,
      visibility: importConfig.globalVisibility,
      auto_sync_enabled: importConfig.autoSync,
      sync_branches: importConfig.syncBranches.length > 0 ? importConfig.syncBranches : undefined,
      environment_mappings:
        Object.keys(importConfig.environmentMappings).length > 0
          ? importConfig.environmentMappings
          : undefined,
    };

    startImportMutation.mutate(request, {
      onSuccess: (data) => {
        setActiveImportId(data.id);
        setCurrentStep(4);
      },
    });
  }, [repoId, scanResult, importConfig, repo, startImportMutation, setActiveImportId]);

  const handleCancel = useCallback(() => {
    resetImportConfig();
    setScanResult(null);
    setActiveImportId(null);
    navigate('/github');
  }, [resetImportConfig, setScanResult, setActiveImportId, navigate]);

  const progressPercent = progress?.progress ?? 0;

  if (!repoId) return null;

  return (
    <div className="min-h-[calc(100vh-4rem)] p-6 md:p-8">
      <div className="max-w-3xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => {
              if (currentStep > 1 && currentStep < 4) {
                setCurrentStep((s) => (s - 1) as StepId);
              } else {
                handleCancel();
              }
            }}
            disabled={currentStep === 4}
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h1 className="text-xl font-bold text-text-primary">Import from GitHub</h1>
            <p className="text-sm text-text-secondary">
              {repo?.full_name || 'Loading repository...'}
            </p>
          </div>
        </div>

        {/* Step Progress */}
        <div className="flex items-center gap-2">
          {STEPS.map((step, index) => {
            const isActive = step.id === currentStep;
            const isCompleted = step.id < currentStep;
            const Icon = step.icon;
            return (
              <div key={step.id} className="flex items-center gap-2 flex-1">
                <div
                  className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all flex-1 justify-center ${
                    isActive
                      ? 'bg-brand-500/20 text-brand-500 border border-brand-500/30'
                      : isCompleted
                        ? 'bg-green-500/10 text-green-500 border border-green-500/20'
                        : 'bg-bg-secondary text-text-muted border border-border-default'
                  }`}
                >
                  {isCompleted ? (
                    <CheckCircle className="w-4 h-4" />
                  ) : (
                    <Icon className="w-4 h-4" />
                  )}
                  <span className="hidden sm:inline">{step.label}</span>
                </div>
                {index < STEPS.length - 1 && (
                  <div
                    className={`h-px flex-1 max-w-8 ${
                      step.id < currentStep ? 'bg-green-500/50' : 'bg-border-default'
                    }`}
                  />
                )}
              </div>
            );
          })}
        </div>

        {/* Step Content */}
        <AnimatePresence mode="wait">
          <motion.div
            key={currentStep}
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
            transition={{ duration: 0.2 }}
          >
            {currentStep === 1 && (
              <StepScan
                repo={repo}
                repoLoading={repoLoading}
                scanResult={scanResult}
                isScanning={scanMutation.isPending}
                onScan={handleScan}
                onNext={() => setCurrentStep(2)}
              />
            )}
            {currentStep === 2 && (
              <StepConfigure
                scanResult={scanResult}
                importConfig={importConfig}
                branches={branches}
                toggleFunction={toggleFunction}
                setGlobalVisibility={setGlobalVisibility}
                setAutoSync={setAutoSync}
                setSyncBranches={setSyncBranches}
                onPreview={handlePreview}
                isPreviewing={previewMutation.isPending}
              />
            )}
            {currentStep === 3 && (
              <StepConfirm
                preview={preview}
                selectedFunctions={importConfig.selectedFunctions}
                onStartImport={handleStartImport}
                isStarting={startImportMutation.isPending}
                onBack={() => setCurrentStep(2)}
              />
            )}
            {currentStep === 4 && (
              <StepProgress
                progress={progress}
                complete={complete}
                progressError={progressError}
                progressStatus={progressStatus}
                progressPercent={progressPercent}
                onViewFunction={() => {
                  if (complete?.function_id) {
                    navigate(`/functions/${complete.function_id}`);
                  }
                }}
                onImportAnother={() => {
                  resetImportConfig();
                  setScanResult(null);
                  setActiveImportId(null);
                  setCurrentStep(1);
                }}
                onRetry={() => {
                  setActiveImportId(null);
                  setCurrentStep(1);
                }}
              />
            )}
          </motion.div>
        </AnimatePresence>

        {/* Cancel Dialog */}
        <AlertDialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Discard import?</AlertDialogTitle>
              <AlertDialogDescription>
                This will discard your current import progress. You can always start a new import
                later.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Continue Import</AlertDialogCancel>
              <AlertDialogAction onClick={handleCancel}>Discard</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  );
}

// ============================================================================
// Step 1: Scan
// ============================================================================
function StepScan({
  repo,
  repoLoading,
  scanResult,
  isScanning,
  onScan,
  onNext,
}: {
  repo: GitHubRepo | undefined;
  repoLoading: boolean;
  scanResult: ScanResult | null;
  isScanning: boolean;
  onScan: () => void;
  onNext: () => void;
}) {
  return (
    <div className="space-y-6">
      {/* Repo Info */}
      {repoLoading ? (
        <Skeleton className="h-32 rounded-lg" />
      ) : repo ? (
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 rounded-lg bg-bg-secondary border border-border-default flex items-center justify-center shrink-0">
                <svg role="img" viewBox="0 0 24 24" className="w-6 h-6 text-text-primary" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <h3 className="font-semibold text-text-primary">{repo.full_name}</h3>
                  {repo.is_private ? (
                    <Badge variant="outline" className="text-xs">
                      <Lock className="w-3 h-3 mr-1" />
                      Private
                    </Badge>
                  ) : (
                    <Badge variant="outline" className="text-xs">
                      <Globe className="w-3 h-3 mr-1" />
                      Public
                    </Badge>
                  )}
                </div>
                {repo.description && (
                  <p className="text-sm text-text-secondary mb-2">{repo.description}</p>
                )}
                <div className="flex items-center gap-4 text-xs text-text-muted">
                  {repo.language && (
                    <span className="flex items-center gap-1">
                      <span className="w-2 h-2 rounded-full bg-brand-500" />
                      {repo.language}
                    </span>
                  )}
                  <span className="flex items-center gap-1">
                    <Star className="w-3 h-3" />
                    {repo.stars_count} stars
                  </span>
                  <span className="flex items-center gap-1">
                    <GitBranch className="w-3 h-3" />
                    {repo.default_branch}
                  </span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      ) : null}

      {/* Scan Button / Results */}
      {!scanResult ? (
        <div className="text-center py-8">
          <Scan className="w-12 h-12 text-text-muted mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-text-primary mb-2">Scan for Functions</h3>
          <p className="text-text-secondary mb-6 max-w-md mx-auto">
            We'll analyze the repository to detect deployable functions, their runtimes, and
            configurations.
          </p>
          <Button onClick={onScan} disabled={isScanning} size="lg" className="gap-2">
            {isScanning ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Scanning...
              </>
            ) : (
              <>
                <Scan className="w-4 h-4" />
                Scan for Functions
              </>
            )}
          </Button>
        </div>
      ) : (
        <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">Detection Results</CardTitle>
              <CardDescription>
                Found {scanResult.functions.length} function{scanResult.functions.length !== 1 ? 's' : ''} using{' '}
                {scanResult.strategy_used} strategy
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {scanResult.functions.map((fn) => (
                <div
                  key={fn.name}
                  className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-default"
                >
                  <div className="flex items-center gap-3">
                    <FileCode className="w-4 h-4 text-brand-500" />
                    <div>
                      <span className="font-medium text-text-primary">{fn.name}</span>
                      <span className="text-xs text-text-muted ml-2">{fn.entry_point}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className="text-xs">
                      {fn.runtime}
                    </Badge>
                    <span
                      className={`text-xs font-mono ${
                        fn.confidence >= 0.8
                          ? 'text-green-500'
                          : fn.confidence >= 0.5
                            ? 'text-amber-500'
                            : 'text-red-500'
                      }`}
                    >
                      {Math.round(fn.confidence * 100)}%
                    </span>
                  </div>
                </div>
              ))}

              {scanResult.warnings.length > 0 && (
                <div className="mt-3 p-3 rounded-lg bg-amber-500/10 border border-amber-500/20">
                  {scanResult.warnings.map((warning, i) => (
                    <div key={i} className="flex items-start gap-2 text-sm text-amber-500">
                      <AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" />
                      <span>{warning}</span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <div className="flex justify-end">
            <Button onClick={onNext} className="gap-2">
              Next
              <ArrowRight className="w-4 h-4" />
            </Button>
          </div>
        </motion.div>
      )}
    </div>
  );
}

// ============================================================================
// Step 2: Configure
// ============================================================================
function StepConfigure({
  scanResult,
  importConfig,
  branches,
  toggleFunction,
  setGlobalVisibility,
  setAutoSync,
  setSyncBranches,
  onPreview,
  isPreviewing,
}: {
  scanResult: ScanResult | null;
  importConfig: any;
  branches: Branch[] | undefined;
  toggleFunction: (name: string) => void;
  setGlobalVisibility: (v: 'public' | 'private' | 'unlisted') => void;
  setAutoSync: (enabled: boolean) => void;
  setSyncBranches: (branches: string[]) => void;
  onPreview: () => void;
  isPreviewing: boolean;
}) {
  const [branchInput, setBranchInput] = useState('');

  if (!scanResult) {
    return (
      <div className="text-center py-8">
        <p className="text-text-secondary">No scan results. Go back and scan first.</p>
      </div>
    );
  }

  const handleAddBranch = () => {
    if (branchInput.trim() && !importConfig.syncBranches.includes(branchInput.trim())) {
      setSyncBranches([...importConfig.syncBranches, branchInput.trim()]);
      setBranchInput('');
    }
  };

  return (
    <div className="space-y-6">
      {/* Function Selection */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Select Functions</CardTitle>
          <CardDescription>
            {importConfig.selectedFunctions.length} of {scanResult.functions.length} selected
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          {scanResult.functions.map((fn) => {
            const isSelected = importConfig.selectedFunctions.includes(fn.name);
            return (
              <div
                key={fn.name}
                onClick={() => toggleFunction(fn.name)}
                className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                  isSelected
                    ? 'border-brand-500/50 bg-brand-500/5'
                    : 'border-border-default bg-bg-secondary hover:border-border-default/80'
                }`}
              >
                <div
                  className={`w-5 h-5 rounded border flex items-center justify-center shrink-0 ${
                    isSelected
                      ? 'bg-brand-500 border-brand-500'
                      : 'border-border-default'
                  }`}
                >
                  {isSelected && <Check className="w-3 h-3 text-white" />}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-text-primary">{fn.name}</span>
                    <Badge variant="outline" className="text-xs">
                      {fn.runtime}
                    </Badge>
                  </div>
                  <span className="text-xs text-text-muted">{fn.entry_point}</span>
                </div>
                <span
                  className={`text-xs font-mono ${
                    fn.confidence >= 0.8
                      ? 'text-green-500'
                      : fn.confidence >= 0.5
                        ? 'text-amber-500'
                        : 'text-red-500'
                  }`}
                >
                  {Math.round(fn.confidence * 100)}%
                </span>
              </div>
            );
          })}
        </CardContent>
      </Card>

      {/* Global Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Import Settings</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Visibility */}
          <div className="space-y-2">
            <Label>Default Visibility</Label>
            <Select
              value={importConfig.globalVisibility}
              onValueChange={(v) => setGlobalVisibility(v as any)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="public">
                  <div className="flex items-center gap-2">
                    <Globe className="w-4 h-4" />
                    Public
                  </div>
                </SelectItem>
                <SelectItem value="private">
                  <div className="flex items-center gap-2">
                    <Lock className="w-4 h-4" />
                    Private
                  </div>
                </SelectItem>
                <SelectItem value="unlisted">
                  <div className="flex items-center gap-2">
                    <Eye className="w-4 h-4" />
                    Unlisted
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Auto-sync */}
          <div className="flex items-center justify-between">
            <div>
              <Label>Auto-sync</Label>
              <p className="text-xs text-text-muted">
                Automatically re-import when code changes are pushed
              </p>
            </div>
            <Switch checked={importConfig.autoSync} onCheckedChange={setAutoSync} />
          </div>

          {/* Branch Selector */}
          {importConfig.autoSync && (
            <div className="space-y-2">
              <Label>Sync Branches</Label>
              <div className="flex gap-2">
                <Input
                  placeholder="Branch name (e.g., main)"
                  value={branchInput}
                  onChange={(e) => setBranchInput(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAddBranch()}
                />
                <Button variant="outline" onClick={handleAddBranch} type="button">
                  Add
                </Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {importConfig.syncBranches.map((branch: string) => (
                  <Badge key={branch} variant="secondary" className="gap-1">
                    <GitBranch className="w-3 h-3" />
                    {branch}
                    <button
                      onClick={() =>
                        setSyncBranches(
                          importConfig.syncBranches.filter((b: string) => b !== branch)
                        )
                      }
                      className="ml-1 hover:text-red-500"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </Badge>
                ))}
                {importConfig.syncBranches.length === 0 && branches && (
                  <p className="text-xs text-text-muted">
                    Available: {branches.slice(0, 5).map((b) => b.name).join(', ')}
                  </p>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Next Button */}
      <div className="flex justify-end">
        <Button
          onClick={onPreview}
          disabled={importConfig.selectedFunctions.length === 0 || isPreviewing}
          className="gap-2"
        >
          {isPreviewing ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              Generating Preview...
            </>
          ) : (
            <>
              Preview Import
              <ArrowRight className="w-4 h-4" />
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

// ============================================================================
// Step 3: Confirm
// ============================================================================
function StepConfirm({
  preview,
  selectedFunctions,
  onStartImport,
  isStarting,
  onBack,
}: {
  preview: ImportPreview | null;
  selectedFunctions: string[];
  onStartImport: () => void;
  isStarting: boolean;
  onBack: () => void;
}) {
  return (
    <div className="space-y-6">
      {preview ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Import Preview</CardTitle>
            <CardDescription>
              Review what will be created before starting the import
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Functions to import */}
            <div className="space-y-2">
              <Label>Functions ({preview.functions.length})</Label>
              {preview.functions.map((fn) => (
                <div
                  key={fn.name}
                  className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-default"
                >
                  <div className="flex items-center gap-2">
                    <FileCode className="w-4 h-4 text-brand-500" />
                    <span className="text-sm font-medium">{fn.name}</span>
                  </div>
                  <Badge variant="outline" className="text-xs">
                    {fn.runtime}
                  </Badge>
                </div>
              ))}
            </div>

            {/* Cost */}
            {preview.total_estimated_cost_usd > 0 && (
              <div className="flex items-center justify-between p-3 rounded-lg bg-bg-secondary border border-border-default">
                <span className="text-sm text-text-secondary">Estimated Cost</span>
                <span className="font-mono font-semibold text-text-primary">
                  ${preview.total_estimated_cost_usd.toFixed(4)}
                </span>
              </div>
            )}

            {/* Warnings */}
            {preview.warnings.length > 0 && (
              <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 space-y-2">
                {preview.warnings.map((warning, i) => (
                  <div key={i} className="flex items-start gap-2 text-sm text-amber-500">
                    <AlertTriangle className="w-4 h-4 shrink-0 mt-0.5" />
                    <span>{warning}</span>
                  </div>
                ))}
              </div>
            )}

            {/* Conflicts */}
            {preview.conflicts.length > 0 && (
              <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/20 space-y-2">
                <Label className="text-red-500">Conflicts Detected</Label>
                {preview.conflicts.map((conflict, i) => (
                  <div key={i} className="flex items-center justify-between text-sm">
                    <span className="text-text-primary">{conflict.function_name}</span>
                    <Badge variant="destructive" className="text-xs">
                      {conflict.existing_version}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="text-center py-8">
          <p className="text-text-secondary">No preview available</p>
        </div>
      )}

      <div className="flex items-center justify-between">
        <Button variant="outline" onClick={onBack} className="gap-2">
          <ArrowLeft className="w-4 h-4" />
          Back
        </Button>
        <Button
          onClick={onStartImport}
          disabled={isStarting}
          size="lg"
          className="gap-2"
        >
          {isStarting ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              Starting Import...
            </>
          ) : (
            <>
              <Rocket className="w-4 h-4" />
              Confirm Import
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

// ============================================================================
// Step 4: Progress
// ============================================================================
function StepProgress({
  progress,
  complete,
  progressError,
  progressStatus,
  progressPercent,
  onViewFunction,
  onImportAnother,
  onRetry,
}: {
  progress: ImportProgressEvent | null;
  complete: ImportCompleteEvent | null;
  progressError: ImportErrorEvent | null;
  progressStatus: string;
  progressPercent: number;
  onViewFunction: () => void;
  onImportAnother: () => void;
  onRetry: () => void;
}) {
  if (progressStatus === 'completed' || complete) {
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="text-center py-12"
      >
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: 'spring', bounce: 0.5, delay: 0.1 }}
          className="w-20 h-20 rounded-full bg-green-500/20 border border-green-500/30 flex items-center justify-center mx-auto mb-6"
        >
          <CheckCircle className="w-10 h-10 text-green-500" />
        </motion.div>
        <h2 className="text-2xl font-bold text-text-primary mb-2">Import Complete!</h2>
        <p className="text-text-secondary mb-2">
          Successfully imported {complete?.files_imported ?? 0} files
        </p>
        {complete?.function_name && (
          <p className="text-sm text-text-muted mb-6">Function: {complete.function_name}</p>
        )}
        <div className="flex items-center justify-center gap-3">
          <Button onClick={onViewFunction} className="gap-2">
            <Eye className="w-4 h-4" />
            View Function
          </Button>
          <Button variant="outline" onClick={onImportAnother} className="gap-2">
            <RefreshCw className="w-4 h-4" />
            Import Another
          </Button>
        </div>
      </motion.div>
    );
  }

  if (progressStatus === 'failed' || progressError) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center py-12"
      >
        <div className="w-20 h-20 rounded-full bg-red-500/20 border border-red-500/30 flex items-center justify-center mx-auto mb-6">
          <X className="w-10 h-10 text-red-500" />
        </div>
        <h2 className="text-2xl font-bold text-text-primary mb-2">Import Failed</h2>
        <p className="text-text-secondary mb-6 max-w-md mx-auto">
          {progressError?.message || 'An unexpected error occurred during import'}
        </p>
        <Button onClick={onRetry} className="gap-2">
          <RefreshCw className="w-4 h-4" />
          Retry
        </Button>
      </motion.div>
    );
  }

  return (
    <div className="space-y-6 py-8">
      <div className="text-center">
        <Loader2 className="w-12 h-12 text-brand-500 animate-spin mx-auto mb-4" />
        <h3 className="text-lg font-semibold text-text-primary mb-1">
          {progress?.message || 'Importing...'}
        </h3>
        <p className="text-sm text-text-muted">
          {progress?.stage ? `Stage: ${progress.stage}` : 'Connecting to import stream...'}
        </p>
      </div>

      {/* Progress Bar */}
      <div className="w-full h-2 bg-bg-secondary rounded-full overflow-hidden">
        <motion.div
          className="h-full bg-gradient-to-r from-brand-500 to-brand-600 rounded-full"
          initial={{ width: 0 }}
          animate={{ width: `${progressPercent}%` }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
        />
      </div>

      <div className="text-center text-sm text-text-muted">
        {progressPercent}% complete
      </div>
    </div>
  );
}
