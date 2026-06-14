import { useState, useEffect, useRef } from "react";
import { motion } from "framer-motion";
import { useOnboardingStore } from "@/stores/onboardingStore";
import { useTranslation } from "react-i18next";
import { Rocket, Code, Check, Loader2, Terminal, Copy, ExternalLink, AlertTriangle, RefreshCw, Sparkles, Globe, CheckCircle2, XCircle } from "lucide-react";
import Confetti from "react-confetti";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { HelpTooltip } from "@/components/ui/help-tooltip";
import { Skeleton } from "@/components/ui/skeleton";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { trackEvent } from "@/lib/analytics";

const sampleFunctions = [
  {
    id: "hello-world",
    name: "Hello World",
    nameKey: "onboarding.deploy.sampleFunctions.helloWorld.name",
    description: "onboarding.deploy.sampleFunctions.helloWorld.description",
    code: `export default {
  async fetch(request) {
    return new Response("Hello from FunctionFly!", {
      headers: { "content-type": "text/plain" },
    });
  },
};`,
    runtime: "cloudflare",
  },
  {
    id: "api-echo",
    name: "API Echo",
    nameKey: "onboarding.deploy.sampleFunctions.apiEcho.name",
    description: "onboarding.deploy.sampleFunctions.apiEcho.description",
    code: `export default {
  async fetch(request) {
    const data = {
      message: "Echo from FunctionFly",
      method: request.method,
      url: request.url,
      headers: Object.fromEntries(request.headers),
      timestamp: new Date().toISOString(),
    };

    return new Response(JSON.stringify(data, null, 2), {
      headers: { "content-type": "application/json" },
    });
  },
};`,
    runtime: "cloudflare",
  },
];

interface FunctionMetrics {
  requests: number;
  latency: number;
  errors: number;
  uptime: number;
}

interface DeploymentResult {
  function_id: string;
  deployment_id: string;
  url: string;
  region: string;
  providers: string[];
}

const DEPLOY_STEPS = {
  VALIDATING: 'onboarding.deploy.steps.validating',
  CONNECTING: 'onboarding.deploy.steps.connecting',
  UPLOADING: 'onboarding.deploy.steps.uploading',
  DEPLOYING: 'onboarding.deploy.steps.deploying',
  CONFIGURING: 'onboarding.deploy.steps.configuring',
  COMPLETE: 'onboarding.deploy.steps.complete',
} as const;

export function DeployFunctionStep() {
  const { t } = useTranslation();
  const { updateStepData, stepData } = useOnboardingStore();
  const [selectedFunction, setSelectedFunction] = useState<string | null>(null);
  const [functionName, setFunctionName] = useState("");
  const [isDeploying, setIsDeploying] = useState(false);
  const [isDeployed, setIsDeployed] = useState(false);
  const [deployedUrl, setDeployedUrl] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState("preset");
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [deployError, setDeployError] = useState<string | null>(null);
  const [showSkeleton, setShowSkeleton] = useState(false);
  const [deployProgress, setDeployProgress] = useState(0);
  const [deployStep, setDeployStep] = useState<string>("");
  const [showConfetti, setShowConfetti] = useState(false);
  const [functionMetrics, setFunctionMetrics] = useState<FunctionMetrics | null>(null);
  const [isLoadingMetrics, setIsLoadingMetrics] = useState(false);
  const [deploymentResult, setDeploymentResult] = useState<DeploymentResult | null>(null);

  const [urlSuggestionsOpen, setUrlSuggestionsOpen] = useState(false);
  const [urlSuggestions, setUrlSuggestions] = useState<Array<{ url: string; available: boolean; reason?: string }>>([]);
  const urlInputRef = useRef<HTMLDivElement>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const selectedFunctionData = sampleFunctions.find((f) => f.id === selectedFunction);

  const validateFunctionName = (name: string): { valid: boolean; message?: string } => {
    const cleanName = name.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
    if (cleanName.length < 2) {
      return { valid: false, message: t('onboarding.deploy.errors.nameTooShort') };
    }
    if (cleanName.length > 52) {
      return { valid: false, message: t('onboarding.deploy.errors.nameTooLong') };
    }
    if (!/^[a-z0-9-]+$/.test(cleanName)) {
      return { valid: false, message: t('onboarding.deploy.errors.nameInvalidChars') };
    }
    return { valid: true };
  };

  const generateUrlSuggestions = (name: string): Array<{ url: string; available: boolean; reason?: string }> => {
    if (!name || name.length < 2) return [];

    const cleanName = name.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
    const baseUrl = `${cleanName}.functionfly.app`;

    const suggestions = [
      { url: baseUrl, available: true, reason: '' },
    ];

    if (cleanName.length > 3) {
      suggestions.push({ url: `api-${baseUrl}`, available: true, reason: '' });
      suggestions.push({ url: `${cleanName}-api.functionfly.app`, available: true, reason: '' });
    }

    if (cleanName.length > 5) {
      suggestions.push({ url: `${cleanName}-v1.functionfly.app`, available: true, reason: '' });
      suggestions.push({ url: `app-${baseUrl}`, available: true, reason: '' });
    }

    if (cleanName.includes('-')) {
      const singleWord = cleanName.split('-')[0];
      suggestions.push({ url: `${singleWord}.functionfly.app`, available: true, reason: '' });
    }

    return suggestions.slice(0, 5);
  };

  useEffect(() => {
    const trimmed = functionName.trim();
    if (trimmed.length < 2) {
      setUrlSuggestions([]);
      setUrlSuggestionsOpen(false);
      return;
    }

    const timer = setTimeout(() => {
      const suggestions = generateUrlSuggestions(trimmed);
      setUrlSuggestions(suggestions);
      setUrlSuggestionsOpen(suggestions.length > 0);
    }, 150);

    return () => clearTimeout(timer);
  }, [functionName, t]);

  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  const handleDeploy = async () => {
    if (!selectedFunction && activeTab === "preset") return;
    const nameValidation = validateFunctionName(functionName);
    if (!nameValidation.valid) {
      setDeployError(nameValidation.message || t('onboarding.deploy.errors.invalidName'));
      return;
    }

    setIsDeploying(true);
    setDeployError(null);
    setDeployProgress(0);

    abortControllerRef.current = new AbortController();

    try {
      setDeployStep(t(DEPLOY_STEPS.VALIDATING));
      setDeployProgress(10);
      await new Promise(resolve => setTimeout(resolve, 300));

      const connectedProvider = stepData["connect-provider"];
      const providerId = connectedProvider?.providerConfig?.id;

      if (!providerId) {
        throw new Error(t('onboarding.deploy.errors.noProvider'));
      }

      setDeployStep(t(DEPLOY_STEPS.CONNECTING));
      setDeployProgress(25);
      await new Promise(resolve => setTimeout(resolve, 300));

      const functionCode = selectedFunctionData?.code || '';
      const cleanName = functionName.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');

      setDeployStep(t(DEPLOY_STEPS.UPLOADING));
      setDeployProgress(45);

      const createResponse = await apiClient.post<{ function_id: string }>('/v1/functions', {
        name: cleanName,
        code: functionCode,
        runtime: selectedFunctionData?.runtime || 'cloudflare',
        providers: [providerId],
        region: 'us-east-1',
      }, { signal: abortControllerRef.current.signal });

      const functionId = createResponse.function_id;

      setDeployStep(t(DEPLOY_STEPS.DEPLOYING));
      setDeployProgress(70);
      await new Promise(resolve => setTimeout(resolve, 300));

      const deployResponse = await apiClient.post<DeploymentResult>('/v1/functions/deploy', {
        function_id: functionId,
        backend_id: providerId,
      }, { signal: abortControllerRef.current.signal });

      setDeploymentResult(deployResponse);

      setDeployStep(t(DEPLOY_STEPS.CONFIGURING));
      setDeployProgress(90);
      await new Promise(resolve => setTimeout(resolve, 300));

      setDeployProgress(100);
      setDeployStep(t(DEPLOY_STEPS.COMPLETE));

      setShowSkeleton(true);
      await new Promise(resolve => setTimeout(resolve, 500));

      setShowSkeleton(false);
      setDeployedUrl(deployResponse.url);
      setIsDeployed(true);
      setShowConfetti(true);

      trackEvent('onboarding_function_deployed', {
        function_id: functionId,
        function_name: cleanName,
        provider: providerId,
      });

      loadFunctionMetrics(functionId);

      updateStepData("deploy-function", {
        functionName: cleanName,
        functionId,
        deployedUrl: deployResponse.url,
        selectedFunction,
        deployedAt: new Date().toISOString(),
        deploymentId: deployResponse.deployment_id,
      });

      setTimeout(() => setShowConfetti(false), 3000);

      toast.success(t('onboarding.deploy.toast.success', { name: cleanName }));
    } catch (error: any) {
      setShowSkeleton(false);

      if (error.name === 'AbortError') {
        setDeployError(t('onboarding.deploy.errors.cancelled'));
        return;
      }

      trackEvent('onboarding_function_deploy_failed', {
        error: error?.message || 'Unknown error',
        function_name: functionName,
      });

      const errorMessage = error?.response?.data?.message || error?.message || t('onboarding.deploy.errors.generic');
      setDeployError(errorMessage);
      toast.error(t('onboarding.deploy.toast.failed'));
    } finally {
      setIsDeploying(false);
      setDeployProgress(0);
      setDeployStep("");
      abortControllerRef.current = null;
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success(t('onboarding.deploy.toast.copied'));
  };

  const loadFunctionMetrics = async (functionId?: string) => {
    setIsLoadingMetrics(true);
    try {
      const id = functionId || deploymentResult?.function_id;
      if (!id) {
        throw new Error('No function ID');
      }

      await new Promise(resolve => setTimeout(resolve, 500));

      const metricsResponse = await apiClient.get<{ requests: number; latency_ms: number; error_rate: number; uptime_percent: number }>(
        `/v1/functions/${id}/metrics`,
        { params: { period: '24h' } }
      );

      setFunctionMetrics({
        requests: metricsResponse.requests || 0,
        latency: metricsResponse.latency_ms || 0,
        errors: Math.round((metricsResponse.error_rate || 0) * (metricsResponse.requests || 0)),
        uptime: metricsResponse.uptime_percent || 99.9,
      });
    } catch (error) {
      console.error("Failed to load metrics:", error);
      setFunctionMetrics({
        requests: 0,
        latency: 0,
        errors: 0,
        uptime: 99.9,
      });
    } finally {
      setIsLoadingMetrics(false);
    }
  };

  if (isDeployed && deployedUrl) {
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="space-y-6"
      >
        {showConfetti && (
          <Confetti
            width={window.innerWidth}
            height={window.innerHeight}
            recycle={false}
            numberOfPieces={50}
            gravity={0.3}
            colors={['#f59e0b', '#ffb800', '#06b6d4', '#5b7cf5', '#10b981']}
          />
        )}

        {showSkeleton ? (
          <div className="space-y-6">
            <div className="text-center py-4">
              <Skeleton className="w-16 h-16 rounded-full mx-auto mb-4" />
              <Skeleton className="h-6 w-48 mx-auto mb-2" />
              <Skeleton className="h-4 w-64 mx-auto" />
            </div>

            <Card className="aviation-instrument p-4">
              <div className="flex items-center justify-between mb-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-8 w-8 rounded" />
              </div>
              <Skeleton className="h-12 w-full" />
            </Card>

            <div className="flex gap-3">
              <Skeleton className="h-10 flex-1" />
            </div>
          </div>
        ) : (
          <div>
            <div className="text-center py-4 relative">
              <div className="absolute inset-0 pointer-events-none" aria-hidden="true">
                {[...Array(10)].map((_, i) => (
                  <motion.div
                    key={i}
                    className="absolute w-2 h-2 bg-gradient-to-r from-aviation-amber to-aviation-cyan rounded-full"
                    initial={{
                      x: "50%",
                      y: "50%",
                      scale: 0,
                      opacity: 1
                    }}
                    animate={{
                      x: `${50 + (Math.random() - 0.5) * 200}%`,
                      y: `${50 + (Math.random() - 0.5) * 200}%`,
                      scale: [0, 1, 0],
                      opacity: [1, 1, 0]
                    }}
                    transition={{
                      duration: 2,
                      delay: Math.random() * 0.5,
                      ease: "easeOut"
                    }}
                    style={{
                      left: `${Math.random() * 100}%`,
                      top: `${Math.random() * 100}%`,
                    }}
                  />
                ))}
              </div>

              <motion.div
                className="w-16 h-16 bg-aviation-green-dim rounded-full flex items-center justify-center mx-auto mb-4 relative z-10"
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ type: "spring", bounce: 0.5, delay: 0.2 }}
              >
                <motion.div
                  initial={{ scale: 0, rotate: -180 }}
                  animate={{ scale: 1, rotate: 0 }}
                  transition={{ type: "spring", bounce: 0.6, delay: 0.4 }}
                >
                  <Check className="w-8 h-8 text-aviation-green" aria-hidden="true" />
                </motion.div>
              </motion.div>

              <motion.h3
                className="text-xl font-mono font-bold text-aviation-text-primary mb-2"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.6 }}
              >
                {t('onboarding.deploy.deployed.title')}
              </motion.h3>

              <motion.p
                className="text-aviation-text-secondary font-mono mb-4"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.8 }}
              >
                {t('onboarding.deploy.deployed.subtitle')}
              </motion.p>
            </div>

            <Card className="aviation-instrument p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-mono text-aviation-text-secondary">{t('onboarding.deploy.deployed.functionUrl')}</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(deployedUrl)}
                  className="text-aviation-text-muted hover:text-aviation-amber"
                  aria-label={t('onboarding.deploy.actions.copyUrl')}
                >
                  <Copy className="w-4 h-4" aria-hidden="true" />
                </Button>
              </div>
              <code className="block p-3 bg-aviation-bg-tertiary rounded text-sm font-mono text-aviation-amber break-all">
                {deployedUrl}
              </code>
            </Card>

            <Card className="aviation-instrument p-4">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-mono font-semibold text-aviation-text-primary">{t('onboarding.deploy.deployed.liveMetrics')}</h4>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => loadFunctionMetrics()}
                  disabled={isLoadingMetrics}
                  className="text-aviation-text-muted hover:text-aviation-amber"
                  aria-label={t('onboarding.deploy.actions.refreshMetrics')}
                >
                  {isLoadingMetrics ? (
                    <Loader2 className="w-4 h-4 animate-spin" aria-hidden="true" />
                  ) : (
                    <RefreshCw className="w-4 h-4" aria-hidden="true" />
                  )}
                </Button>
              </div>

              {isLoadingMetrics ? (
                <div className="grid grid-cols-2 gap-4">
                  {[1, 2, 3, 4].map((i) => (
                    <div key={i} className="text-center">
                      <Skeleton className="h-8 w-16 mx-auto mb-1" />
                      <Skeleton className="h-3 w-12 mx-auto" />
                    </div>
                  ))}
                </div>
              ) : functionMetrics ? (
                <div className="grid grid-cols-2 gap-4">
                  <div className="text-center">
                    <div className="text-2xl font-mono font-bold text-aviation-text-primary">
                      {functionMetrics.requests.toLocaleString()}
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">{t('onboarding.deploy.metrics.requests')}</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-mono font-bold text-aviation-text-primary">
                      {functionMetrics.latency}ms
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">{t('onboarding.deploy.metrics.latency')}</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-mono font-bold text-aviation-red">
                      {functionMetrics.errors}
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">{t('onboarding.deploy.metrics.errors')}</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-mono font-bold text-aviation-green">
                      {functionMetrics.uptime.toFixed(1)}%
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">{t('onboarding.deploy.metrics.uptime')}</div>
                  </div>
                </div>
              ) : (
                <div className="text-center text-aviation-text-muted py-4">
                  <div className="text-sm font-mono">{t('onboarding.deploy.metrics.loading')}</div>
                </div>
              )}
            </Card>

            <div className="flex gap-3">
              <Button
                variant="outline"
                className="flex-1 font-mono border-aviation-border-instrument text-aviation-text-primary hover:border-aviation-amber hover:text-aviation-amber"
                onClick={() => window.open(deployedUrl, "_blank")}
              >
                <ExternalLink className="w-4 h-4 mr-2" aria-hidden="true" />
                {t('onboarding.deploy.actions.openFunction')}
              </Button>
            </div>

            <div className="mt-6 pt-4 border-t border-aviation-border-panel">
              <Button
                variant="ghost"
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="w-full justify-between font-mono text-aviation-text-secondary hover:text-aviation-amber"
                aria-expanded={showAdvanced}
                aria-controls="advanced-config"
              >
                <span className="text-sm font-mono font-medium">{t('onboarding.deploy.advanced.title')}</span>
                <HelpTooltip content={t('onboarding.deploy.advanced.tooltip')} />
              </Button>

              {showAdvanced && (
                <motion.div
                  id="advanced-config"
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  className="mt-4 space-y-4"
                >
                  <Card className="aviation-panel p-4">
                    <h4 className="font-mono font-medium text-aviation-text-primary mb-3 flex items-center gap-2">
                      {t('onboarding.deploy.advanced.envVars')}
                      <HelpTooltip content={t('onboarding.deploy.advanced.envVarsTooltip')} />
                    </h4>
                    <div className="space-y-2">
                      <div className="flex gap-2">
                        <Input placeholder="KEY" className="aviation-input flex-1" aria-label={t('onboarding.deploy.advanced.envKeyPlaceholder')} />
                        <Input placeholder="VALUE" className="aviation-input flex-1" type="password" aria-label={t('onboarding.deploy.advanced.envValuePlaceholder')} />
                        <Button variant="outline" size="sm" className="border-aviation-border-instrument">{t('onboarding.deploy.advanced.add')}</Button>
                      </div>
                      <p className="text-xs font-mono text-aviation-text-muted">
                        {t('onboarding.deploy.advanced.envVarsHint')}
                      </p>
                    </div>
                  </Card>

                  <Card className="aviation-panel p-4">
                    <h4 className="font-mono font-medium text-aviation-text-primary mb-3 flex items-center gap-2">
                      {t('onboarding.deploy.advanced.scaling')}
                      <HelpTooltip content={t('onboarding.deploy.advanced.scalingTooltip')} />
                    </h4>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label className="text-sm font-mono text-aviation-text-secondary">{t('onboarding.deploy.advanced.maxConcurrent')}</Label>
                        <Input type="number" defaultValue="100" className="aviation-input mt-1" aria-label={t('onboarding.deploy.advanced.maxConcurrentLabel')} />
                      </div>
                      <div>
                        <Label className="text-sm font-mono text-aviation-text-secondary">{t('onboarding.deploy.advanced.timeout')}</Label>
                        <Input type="number" defaultValue="30" className="aviation-input mt-1" aria-label={t('onboarding.deploy.advanced.timeoutLabel')} />
                      </div>
                    </div>
                  </Card>

                  <div className="bg-aviation-cyan-dim border border-aviation-cyan/30 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <Code className="w-5 h-5 text-aviation-cyan flex-shrink-0 mt-0.5" aria-hidden="true" />
                      <div>
                        <h4 className="font-mono font-medium text-aviation-cyan mb-1">{t('onboarding.deploy.advanced.proTipTitle')}</h4>
                        <p className="text-sm font-mono text-aviation-text-secondary">
                          {t('onboarding.deploy.advanced.proTipDesc')}
                        </p>
                      </div>
                    </div>
                  </div>
                </motion.div>
              )}
            </div>
          </div>
        )}
      </motion.div>
    );
  }

  return (
    <div className="space-y-6">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2 bg-aviation-bg-tertiary">
          <TabsTrigger value="preset" className="flex items-center gap-2 font-mono text-aviation-text-secondary data-[state=active]:text-aviation-amber">
            {t('onboarding.deploy.tabs.sampleFunctions')}
            <HelpTooltip content={t('onboarding.deploy.tabs.sampleFunctionsTooltip')} />
          </TabsTrigger>
          <TabsTrigger value="custom" className="flex items-center gap-2 font-mono text-aviation-text-secondary data-[state=active]:text-aviation-amber">
            {t('onboarding.deploy.tabs.customCode')}
            <HelpTooltip content={t('onboarding.deploy.tabs.customCodeTooltip')} />
          </TabsTrigger>
        </TabsList>

        <TabsContent value="preset" className="space-y-4">
          <div className="grid gap-3" role="listbox" aria-label={t('onboarding.deploy.functionListLabel')}>
            {sampleFunctions.map((func) => (
              <Card
                key={func.id}
                className={`aviation-instrument p-4 cursor-pointer transition-all ${
                  selectedFunction === func.id
                    ? "border-aviation-amber ring-1 ring-aviation-amber"
                    : "hover:border-aviation-border-glow"
                }`}
                onClick={() => setSelectedFunction(func.id)}
                role="option"
                aria-selected={selectedFunction === func.id}
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    setSelectedFunction(func.id);
                  }
                }}
              >
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-lg bg-aviation-amber-subtle flex items-center justify-center flex-shrink-0">
                    <Code className="w-5 h-5 text-aviation-amber" aria-hidden="true" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-mono font-semibold text-aviation-text-primary">{t(func.nameKey)}</h3>
                      <HelpTooltip content={t('onboarding.deploy.functionEdgeHelp')} />
                      {selectedFunction === func.id && (
                        <Check className="w-4 h-4 text-aviation-green" aria-hidden="true" />
                      )}
                    </div>
                    <p className="text-sm font-mono text-aviation-text-secondary">{t(func.description)}</p>
                  </div>
                </div>
              </Card>
            ))}
          </div>

          {selectedFunctionData && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              className="space-y-4"
            >
              <Card className="aviation-panel overflow-hidden">
                <div className="flex items-center justify-between px-4 py-2 bg-aviation-bg-tertiary border-b border-aviation-border-panel">
                  <div className="flex items-center gap-2">
                    <Terminal className="w-4 h-4 text-aviation-text-muted" aria-hidden="true" />
                    <span className="text-xs font-mono text-aviation-text-muted">{t('onboarding.deploy.codePreview')}</span>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(selectedFunctionData.code)}
                    className="text-aviation-text-muted hover:text-aviation-amber"
                    aria-label={t('onboarding.deploy.actions.copyCode')}
                  >
                    <Copy className="w-3 h-3" aria-hidden="true" />
                  </Button>
                </div>
                <pre className="p-4 text-sm font-mono text-aviation-text-secondary overflow-x-auto" role="region" aria-label={t('onboarding.deploy.codePreviewRegion')}>
                  <code>{selectedFunctionData.code}</code>
                </pre>
              </Card>
            </motion.div>
          )}
        </TabsContent>

        <TabsContent value="custom" className="space-y-4">
          <Card className="aviation-panel p-4">
            <p className="text-sm font-mono text-aviation-text-secondary mb-4">
              {t('onboarding.deploy.customCodeNotice')}
            </p>
            <Button
              variant="outline"
              onClick={() => setActiveTab("preset")}
              className="w-full font-mono border-aviation-border-instrument text-aviation-text-primary hover:border-aviation-amber hover:text-aviation-amber"
            >
              {t('onboarding.deploy.actions.chooseSample')}
            </Button>
          </Card>
        </TabsContent>
      </Tabs>

      <div className="space-y-2" ref={urlInputRef}>
        <Label htmlFor="functionName" className="flex items-center gap-2 font-mono text-aviation-text-secondary">
          {t('onboarding.deploy.functionNameLabel')}
          <HelpTooltip content={t('onboarding.deploy.functionNameTooltip')} />
        </Label>
        <Popover open={urlSuggestionsOpen} onOpenChange={setUrlSuggestionsOpen}>
          <PopoverTrigger asChild>
            <Input
              id="functionName"
              placeholder="my-first-function"
              value={functionName}
              onChange={(e) => setFunctionName(e.target.value)}
              onFocus={() => {
                if (urlSuggestions.length > 0) setUrlSuggestionsOpen(true);
              }}
              className="aviation-input"
              aria-describedby="functionName-hint"
              aria-invalid={functionName.length >= 2 && !/^[a-z0-9-]+$/.test(functionName.toLowerCase())}
            />
          </PopoverTrigger>
          <PopoverContent
            className="w-[400px] p-0"
            align="start"
            sideOffset={4}
            style={{
              backgroundColor: 'var(--ff-bg-secondary)',
              borderColor: 'var(--ff-border-default)',
            }}
          >
            <div className="px-3 py-2 border-b border-aviation-border-panel" style={{ borderColor: 'var(--ff-border-subtle)' }}>
              <div className="flex items-center gap-2 text-xs font-mono text-aviation-text-muted">
                <Sparkles className="w-3 h-3 text-aviation-amber" aria-hidden="true" />
                <span>{t('onboarding.deploy.urlSuggestions')}</span>
              </div>
            </div>
            <div className="py-1 max-h-[240px] overflow-y-auto" role="listbox">
              {urlSuggestions.map((suggestion, index) => (
                <button
                  key={index}
                  onClick={() => {
                    const cleanName = suggestion.url.replace('.functionfly.app', '');
                    setFunctionName(cleanName);
                    setUrlSuggestionsOpen(false);
                  }}
                  className="w-full px-3 py-2 flex items-center justify-between hover:bg-aviation-bg-tertiary transition-colors text-left"
                  style={{ borderColor: 'var(--ff-border-subtle)' }}
                  role="option"
                  aria-selected={false}
                >
                  <div className="flex items-center gap-2 min-w-0 flex-1">
                    <Globe className="w-4 h-4 text-aviation-text-muted flex-shrink-0" aria-hidden="true" />
                    <span className="font-mono text-sm text-aviation-text-primary truncate">
                      {suggestion.url}
                    </span>
                  </div>
                  <div className="flex items-center gap-1.5 flex-shrink-0 ml-2">
                    {suggestion.reason && (
                      <span className="text-xs font-mono text-aviation-text-muted">{suggestion.reason}</span>
                    )}
                    {suggestion.available ? (
                      <span className="flex items-center gap-1 text-xs font-mono text-aviation-green" role="status">
                        <CheckCircle2 className="w-3 h-3" aria-hidden="true" />
                        {t('onboarding.deploy.suggestions.available')}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-xs font-mono text-aviation-red" role="status">
                        <XCircle className="w-3 h-3" aria-hidden="true" />
                        {t('onboarding.deploy.suggestions.taken')}
                      </span>
                    )}
                  </div>
                </button>
              ))}
            </div>
            <div className="px-3 py-2 border-t border-aviation-border-panel text-xs font-mono text-aviation-text-muted" style={{ borderColor: 'var(--ff-border-subtle)' }}>
              {t('onboarding.deploy.urlSuggestionsHint')}
            </div>
          </PopoverContent>
        </Popover>
        <div id="functionName-hint" className="text-xs font-mono text-aviation-text-muted space-y-1">
          <p>
            {t('onboarding.deploy.functionNameHint')}
          </p>
          <p className="text-aviation-stratosphere">
            {t('onboarding.deploy.functionNamePreview', { name: functionName ? functionName.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-') : 'your-function' })}
          </p>
        </div>
      </div>

      <Button
        onClick={handleDeploy}
        disabled={(!selectedFunction && activeTab === "preset") || !functionName.trim() || isDeploying}
        className="aviation-button-primary w-full font-mono"
        aria-describedby={deployError ? 'deploy-error' : undefined}
      >
        {isDeploying ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" aria-hidden="true" />
            {t('onboarding.deploy.actions.deploying')}
          </>
        ) : (
          <>
            <Rocket className="w-4 h-4 mr-2" aria-hidden="true" />
            {t('onboarding.deploy.actions.deploy')}
          </>
        )}
      </Button>

      {isDeploying && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: "auto" }}
          className="space-y-2"
          role="status"
          aria-live="polite"
        >
          <div className="flex items-center justify-between text-sm">
            <span className="text-aviation-text-primary font-mono">{deployStep}</span>
            <span className="text-aviation-text-muted font-mono">{Math.round(deployProgress)}%</span>
          </div>
          <Progress value={deployProgress} className="h-2 bg-aviation-bg-tertiary [&>div]:bg-gradient-to-r [&>div]:from-aviation-amber [&>div]:to-aviation-cyan" aria-label={t('onboarding.deploy.deployProgress', { percent: Math.round(deployProgress) })} />
        </motion.div>
      )}

      {deployError && (
        <motion.div
          id="deploy-error"
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-aviation-red-dim border border-aviation-red/30 rounded-lg p-4"
          role="alert"
        >
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-aviation-red flex-shrink-0 mt-0.5" aria-hidden="true" />
            <div>
              <h4 className="font-mono font-medium text-aviation-red mb-1">{t('onboarding.deploy.errors.deploymentFailed')}</h4>
              <p className="text-sm font-mono text-aviation-red/80">{deployError}</p>
              <div className="flex gap-2 mt-3">
                <Button
                  variant="outline"
                  size="sm"
                  className="border-aviation-red/30 text-aviation-red hover:bg-aviation-red-dim font-mono"
                  onClick={() => {
                    setDeployError(null);
                    handleDeploy();
                  }}
                >
                  {t('onboarding.deploy.actions.tryAgain')}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="border-aviation-cyan/30 text-aviation-cyan hover:bg-aviation-cyan-dim font-mono"
                  onClick={() => {
                    setActiveTab("custom");
                    setDeployError(null);
                  }}
                >
                  {t('onboarding.deploy.actions.tryCustomCode')}
                </Button>
              </div>
              <div className="mt-3 p-3 bg-aviation-bg-tertiary rounded text-xs font-mono text-aviation-text-muted">
                <strong>{t('onboarding.deploy.errors.alternative')}</strong> {t('onboarding.deploy.errors.alternativeDesc')}
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
}