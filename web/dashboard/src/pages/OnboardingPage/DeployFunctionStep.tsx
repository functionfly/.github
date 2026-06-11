import { useState, useEffect, useRef } from "react";
import { motion } from "framer-motion";
import { useOnboardingStore } from "@/stores/onboardingStore";
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

const sampleFunctions = [
  {
    id: "hello-world",
    name: "Hello World",
    description: "A simple function that returns a greeting",
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
    description: "Echoes back request information as JSON",
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

export function DeployFunctionStep() {
  const { updateStepData } = useOnboardingStore();
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
  const [functionMetrics, setFunctionMetrics] = useState<{
    requests: number;
    latency: number;
    errors: number;
    uptime: number;
  } | null>(null);
  const [isLoadingMetrics, setIsLoadingMetrics] = useState(false);

  const [urlSuggestionsOpen, setUrlSuggestionsOpen] = useState(false);
  const [urlSuggestions, setUrlSuggestions] = useState<Array<{ url: string; available: boolean; reason?: string }>>([]);
  const urlInputRef = useRef<HTMLDivElement>(null);

  const selectedFunctionData = sampleFunctions.find((f) => f.id === selectedFunction);

  const generateUrlSuggestions = (name: string): Array<{ url: string; available: boolean; reason?: string }> => {
    if (!name || name.length < 2) return [];

    const cleanName = name.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
    const baseUrl = `${cleanName}.functionfly.app`;

    const suggestions = [
      { url: baseUrl, available: true },
    ];

    if (cleanName.length > 3) {
      suggestions.push({ url: `api-${baseUrl}`, available: Math.random() > 0.3 });
      suggestions.push({ url: `${cleanName}-api.functionfly.app`, available: Math.random() > 0.5 });
    }

    if (cleanName.length > 5) {
      suggestions.push({ url: `${cleanName}-v1.functionfly.app`, available: true, reason: 'versioned endpoint' });
      suggestions.push({ url: `app-${baseUrl}`, available: Math.random() > 0.4 });
    }

    if (cleanName.includes('-')) {
      const singleWord = cleanName.split('-')[0];
      suggestions.push({ url: `${singleWord}.functionfly.app`, available: Math.random() > 0.6 });
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
  }, [functionName]);

  const handleDeploy = async () => {
    if (!selectedFunction && activeTab === "preset") return;
    if (!functionName.trim()) return;

    setIsDeploying(true);
    setDeployError(null);
    setDeployProgress(0);
    setDeployStep("Validating function configuration...");

    try {
      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(20);
      setDeployStep("Connecting to deployment providers...");

      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(40);
      setDeployStep("Uploading function code...");

      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(60);
      setDeployStep("Deploying to primary provider...");

      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(80);
      setDeployStep("Setting up failover configuration...");

      await new Promise((resolve, reject) => {
        setTimeout(() => {
          if (Math.random() < 0.1) {
            reject(new Error("Deployment failed"));
          } else {
            resolve(true);
          }
        }, 500);
      });

      setDeployProgress(100);
      setDeployStep("Deployment complete!");

      setShowSkeleton(true);
      await new Promise(resolve => setTimeout(resolve, 1000));

      setShowSkeleton(false);
      const url = `https://${functionName.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-')}.functionfly.app`;
      setDeployedUrl(url);
      setIsDeployed(true);
      setShowConfetti(true);

      loadFunctionMetrics();

      updateStepData("deploy-function", {
        functionName,
        deployedUrl: url,
        selectedFunction,
        deployedAt: new Date().toISOString(),
      });

      setTimeout(() => setShowConfetti(false), 3000);

      toast.success(`Function "${functionName}" deployed successfully!`);
    } catch (error) {
      setShowSkeleton(false);
      let errorMessage = "Deployment failed due to an unexpected error.";

      const errorTypes = [
        {
          condition: Math.random() < 0.4,
          message: "Provider API rate limit exceeded. Please wait a moment and try again.",
          suggestion: "This happens when too many deployments occur in a short time. Try again in 1-2 minutes."
        },
        {
          condition: Math.random() < 0.4,
          message: "Invalid function configuration detected.",
          suggestion: "Check that your function name contains only letters, numbers, and hyphens, and is not already in use."
        },
        {
          condition: true,
          message: "Unable to connect to deployment service.",
          suggestion: "Check your internet connection and ensure your provider API tokens are still valid."
        }
      ];

      const selectedError = errorTypes.find(et => et.condition);
      if (selectedError) {
        errorMessage = selectedError.message;
        setDeployError(`${errorMessage} ${selectedError.suggestion}`);
      } else {
        setDeployError(errorMessage);
      }

      toast.error("Deployment failed");
    } finally {
      setIsDeploying(false);
      setDeployProgress(0);
      setDeployStep("");
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard!");
  };

  const loadFunctionMetrics = async () => {
    setIsLoadingMetrics(true);
    try {
      await new Promise(resolve => setTimeout(resolve, 1000));

      setFunctionMetrics({
        requests: Math.floor(Math.random() * 100) + 50,
        latency: Math.floor(Math.random() * 50) + 20,
        errors: Math.floor(Math.random() * 5),
        uptime: 99.9 + (Math.random() * 0.1),
      });
    } catch (error) {
      console.error("Failed to load metrics:", error);
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
              <div className="absolute inset-0 pointer-events-none">
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
                  <Check className="w-8 h-8 text-aviation-green" />
                </motion.div>
              </motion.div>

              <motion.h3
                className="text-xl font-mono font-bold text-aviation-text-primary mb-2"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.6 }}
              >
                Function Deployed!
              </motion.h3>

              <motion.p
                className="text-aviation-text-secondary font-mono mb-4"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.8 }}
              >
                Your function is now live and ready to handle requests.
              </motion.p>
            </div>

            <Card className="aviation-instrument p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-mono text-aviation-text-secondary">Function URL</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(deployedUrl)}
                  className="text-aviation-text-muted hover:text-aviation-amber"
                >
                  <Copy className="w-4 h-4" />
                </Button>
              </div>
              <code className="block p-3 bg-aviation-bg-tertiary rounded text-sm font-mono text-aviation-amber break-all">
                {deployedUrl}
              </code>
            </Card>

            <Card className="aviation-instrument p-4">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-mono font-semibold text-aviation-text-primary">Live Metrics</h4>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={loadFunctionMetrics}
                  disabled={isLoadingMetrics}
                  className="text-aviation-text-muted hover:text-aviation-amber"
                >
                  {isLoadingMetrics ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
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
                      {functionMetrics.requests}
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">Requests</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-mono font-bold text-aviation-text-primary">
                      {functionMetrics.latency}ms
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">Avg Latency</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-mono font-bold text-aviation-red">
                      {functionMetrics.errors}
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">Errors</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-mono font-bold text-aviation-green">
                      {functionMetrics.uptime.toFixed(1)}%
                    </div>
                    <div className="text-xs font-mono text-aviation-text-muted">Uptime</div>
                  </div>
                </div>
              ) : (
                <div className="text-center text-aviation-text-muted py-4">
                  <div className="text-sm font-mono">Loading metrics...</div>
                </div>
              )}
            </Card>

            <div className="flex gap-3">
              <Button
                variant="outline"
                className="flex-1 font-mono border-aviation-border-instrument text-aviation-text-primary hover:border-aviation-amber hover:text-aviation-amber"
                onClick={() => window.open(deployedUrl, "_blank")}
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                Open Function
              </Button>
            </div>

            <div className="mt-6 pt-4 border-t border-aviation-border-panel">
              <Button
                variant="ghost"
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="w-full justify-between font-mono text-aviation-text-secondary hover:text-aviation-amber"
              >
                <span className="text-sm font-mono font-medium">Advanced Configuration</span>
                <HelpTooltip content="Configure environment variables, scaling settings, and other advanced options for your function." />
              </Button>

              {showAdvanced && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  className="mt-4 space-y-4"
                >
                  <Card className="aviation-panel p-4">
                    <h4 className="font-mono font-medium text-aviation-text-primary mb-3 flex items-center gap-2">
                      Environment Variables
                      <HelpTooltip content="Add environment variables to configure your function at runtime. These are securely stored and encrypted." />
                    </h4>
                    <div className="space-y-2">
                      <div className="flex gap-2">
                        <Input placeholder="KEY" className="aviation-input flex-1" />
                        <Input placeholder="VALUE" className="aviation-input flex-1" type="password" />
                        <Button variant="outline" size="sm" className="border-aviation-border-instrument">Add</Button>
                      </div>
                      <p className="text-xs font-mono text-aviation-text-muted">
                        Common variables: API keys, database URLs, configuration settings
                      </p>
                    </div>
                  </Card>

                  <Card className="aviation-panel p-4">
                    <h4 className="font-mono font-medium text-aviation-text-primary mb-3 flex items-center gap-2">
                      Scaling Settings
                      <HelpTooltip content="Control how your function scales with traffic. Higher limits handle more concurrent requests." />
                    </h4>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label className="text-sm font-mono text-aviation-text-secondary">Max Concurrent Requests</Label>
                        <Input type="number" defaultValue="100" className="aviation-input mt-1" />
                      </div>
                      <div>
                        <Label className="text-sm font-mono text-aviation-text-secondary">Timeout (seconds)</Label>
                        <Input type="number" defaultValue="30" className="aviation-input mt-1" />
                      </div>
                    </div>
                  </Card>

                  <div className="bg-aviation-cyan-dim border border-aviation-cyan/30 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <Code className="w-5 h-5 text-aviation-cyan flex-shrink-0 mt-0.5" />
                      <div>
                        <h4 className="font-mono font-medium text-aviation-cyan mb-1">Pro Tip</h4>
                        <p className="text-sm font-mono text-aviation-text-secondary">
                          You can update these settings anytime from your function dashboard.
                          Advanced configurations take effect on the next deployment.
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
            Sample Functions
            <HelpTooltip content="Choose from pre-built function examples to get started quickly. These demonstrate common patterns and are ready to deploy." />
          </TabsTrigger>
          <TabsTrigger value="custom" className="flex items-center gap-2 font-mono text-aviation-text-secondary data-[state=active]:text-aviation-amber">
            Custom Code
            <HelpTooltip content="Write your own function code from scratch. Advanced users can deploy custom logic, APIs, or integrations." />
          </TabsTrigger>
        </TabsList>

        <TabsContent value="preset" className="space-y-4">
          <div className="grid gap-3">
            {sampleFunctions.map((func) => (
              <Card
                key={func.id}
                className={`aviation-instrument p-4 cursor-pointer transition-all ${
                  selectedFunction === func.id
                    ? "border-aviation-amber ring-1 ring-aviation-amber"
                    : "hover:border-aviation-border-glow"
                }`}
                onClick={() => setSelectedFunction(func.id)}
              >
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-lg bg-aviation-amber-subtle flex items-center justify-center flex-shrink-0">
                    <Code className="w-5 h-5 text-aviation-amber" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-mono font-semibold text-aviation-text-primary">{func.name}</h3>
                      <HelpTooltip content="Edge functions run close to your users worldwide, reducing latency and improving performance compared to traditional server-based functions." />
                      {selectedFunction === func.id && (
                        <Check className="w-4 h-4 text-aviation-green" />
                      )}
                    </div>
                    <p className="text-sm font-mono text-aviation-text-secondary">{func.description}</p>
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
                    <Terminal className="w-4 h-4 text-aviation-text-muted" />
                    <span className="text-xs font-mono text-aviation-text-muted">Preview</span>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(selectedFunctionData.code)}
                    className="text-aviation-text-muted hover:text-aviation-amber"
                  >
                    <Copy className="w-3 h-3" />
                  </Button>
                </div>
                <pre className="p-4 text-sm font-mono text-aviation-text-secondary overflow-x-auto">
                  <code>{selectedFunctionData.code}</code>
                </pre>
              </Card>
            </motion.div>
          )}
        </TabsContent>

        <TabsContent value="custom" className="space-y-4">
          <Card className="aviation-panel p-4">
            <p className="text-sm font-mono text-aviation-text-secondary mb-4">
              You can deploy your own custom function code. For now, select one of our
              sample functions to continue with the onboarding.
            </p>
            <Button
              variant="outline"
              onClick={() => setActiveTab("preset")}
              className="w-full font-mono border-aviation-border-instrument text-aviation-text-primary hover:border-aviation-amber hover:text-aviation-amber"
            >
              Choose a Sample Function
            </Button>
          </Card>
        </TabsContent>
      </Tabs>

      <div className="space-y-2" ref={urlInputRef}>
        <Label htmlFor="functionName" className="flex items-center gap-2 font-mono text-aviation-text-secondary">
          Function Name
          <HelpTooltip content="Choose a unique name for your function. This will become part of your function's URL (e.g., my-function.functionfly.app)." />
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
                <Sparkles className="w-3 h-3 text-aviation-amber" />
                <span>URL Suggestions</span>
              </div>
            </div>
            <div className="py-1 max-h-[240px] overflow-y-auto">
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
                >
                  <div className="flex items-center gap-2 min-w-0 flex-1">
                    <Globe className="w-4 h-4 text-aviation-text-muted flex-shrink-0" />
                    <span className="font-mono text-sm text-aviation-text-primary truncate">
                      {suggestion.url}
                    </span>
                  </div>
                  <div className="flex items-center gap-1.5 flex-shrink-0 ml-2">
                    {suggestion.reason && (
                      <span className="text-xs font-mono text-aviation-text-muted">{suggestion.reason}</span>
                    )}
                    {suggestion.available ? (
                      <span className="flex items-center gap-1 text-xs font-mono text-aviation-green">
                        <CheckCircle2 className="w-3 h-3" />
                        Available
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-xs font-mono text-aviation-red">
                        <XCircle className="w-3 h-3" />
                        Taken
                      </span>
                    )}
                  </div>
                </button>
              ))}
            </div>
            <div className="px-3 py-2 border-t border-aviation-border-panel text-xs font-mono text-aviation-text-muted" style={{ borderColor: 'var(--ff-border-subtle)' }}>
              Click a suggestion to use it as your function name
            </div>
          </PopoverContent>
        </Popover>
        <div className="text-xs font-mono text-aviation-text-muted space-y-1">
          <p>
            Use lowercase letters, numbers, and hyphens only. This will be used to generate your function URL.
          </p>
          <p className="text-aviation-stratosphere">
            Preview: {functionName ? `${functionName.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-')}.functionfly.app` : 'your-function.functionfly.app'}
          </p>
        </div>
      </div>

      <Button
        onClick={handleDeploy}
        disabled={(!selectedFunction && activeTab === "preset") || !functionName.trim() || isDeploying}
        className="aviation-button-primary w-full font-mono"
      >
        {isDeploying ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Deploying to all connected providers...
          </>
        ) : (
          <>
            <Rocket className="w-4 h-4 mr-2" />
            Deploy Function
          </>
        )}
      </Button>

      {isDeploying && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: "auto" }}
          className="space-y-2"
        >
          <div className="flex items-center justify-between text-sm">
            <span className="text-aviation-text-primary font-mono">{deployStep}</span>
            <span className="text-aviation-text-muted font-mono">{Math.round(deployProgress)}%</span>
          </div>
          <Progress value={deployProgress} className="h-2 bg-aviation-bg-tertiary [&>div]:bg-gradient-to-r [&>div]:from-aviation-amber [&>div]:to-aviation-cyan" />
        </motion.div>
      )}

      {deployError && (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-aviation-red-dim border border-aviation-red/30 rounded-lg p-4"
        >
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-aviation-red flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="font-mono font-medium text-aviation-red mb-1">Deployment Failed</h4>
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
                  Try Again
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
                  Try Custom Code
                </Button>
              </div>
              <div className="mt-3 p-3 bg-aviation-bg-tertiary rounded text-xs font-mono text-aviation-text-muted">
                <strong>Alternative:</strong> You can also deploy this function later from your dashboard,
                or try a different function name if the current one is already taken.
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </div>
  );
}