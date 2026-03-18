import { useState } from "react";
import { motion } from "framer-motion";
import { useOnboardingStore } from "@/stores/onboardingStore";
import { Rocket, Code, Check, Loader2, Terminal, Copy, ExternalLink, AlertTriangle, Sparkles, RefreshCw } from "lucide-react";
import Confetti from "react-confetti";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Progress } from "@/components/ui/progress";
import { HelpTooltip } from "@/components/ui/help-tooltip";
import { Skeleton } from "@/components/ui/skeleton";
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

  const selectedFunctionData = sampleFunctions.find((f) => f.id === selectedFunction);

  const handleDeploy = async () => {
    if (!selectedFunction && activeTab === "preset") return;
    if (!functionName.trim()) return;

    setIsDeploying(true);
    setDeployError(null);
    setDeployProgress(0);
    setDeployStep("Validating function configuration...");

    try {
      // Step 1: Validation (20%)
      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(20);
      setDeployStep("Connecting to deployment providers...");

      // Step 2: Connection (40%)
      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(40);
      setDeployStep("Uploading function code...");

      // Step 3: Upload (60%)
      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(60);
      setDeployStep("Deploying to primary provider...");

      // Step 4: Primary deployment (80%)
      await new Promise(resolve => setTimeout(resolve, 500));
      setDeployProgress(80);
      setDeployStep("Setting up failover configuration...");

      // Step 5: Finalize (90%)
      await new Promise((resolve, reject) => {
        setTimeout(() => {
          // Simulate occasional deployment failure (10% chance)
          if (Math.random() < 0.1) {
            reject(new Error("Deployment failed"));
          } else {
            resolve(true);
          }
        }, 500);
      });

      setDeployProgress(100);
      setDeployStep("Deployment complete!");

      // Show skeleton briefly before showing results
      setShowSkeleton(true);
      await new Promise(resolve => setTimeout(resolve, 1000));

      setShowSkeleton(false);
      const url = `https://${functionName.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-')}.functionfly.app`;
      setDeployedUrl(url);
      setIsDeployed(true);
      setShowConfetti(true);

      // Load initial metrics
      loadFunctionMetrics();

      // Save step data to onboarding store
      updateStepData("deploy-function", {
        functionName,
        deployedUrl: url,
        selectedFunction,
        deployedAt: new Date().toISOString(),
      });

      // Hide confetti after 3 seconds
      setTimeout(() => setShowConfetti(false), 3000);

      toast.success(`Function "${functionName}" deployed successfully!`);
    } catch (error) {
      setShowSkeleton(false);
      let errorMessage = "Deployment failed due to an unexpected error.";

      // Simulate different types of deployment errors
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

  // Simulate real-time metrics for deployed function
  const loadFunctionMetrics = async () => {
    setIsLoadingMetrics(true);
    try {
      // Simulate API call to get metrics
      await new Promise(resolve => setTimeout(resolve, 1000));

      // Mock metrics data
      setFunctionMetrics({
        requests: Math.floor(Math.random() * 100) + 50, // 50-150 requests
        latency: Math.floor(Math.random() * 50) + 20,   // 20-70ms latency
        errors: Math.floor(Math.random() * 5),          // 0-5 errors
        uptime: 99.9 + (Math.random() * 0.1),           // 99.9-100% uptime
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
            numberOfPieces={100}
            gravity={0.3}
            colors={['#6366f1', '#8b5cf6', '#10b981', '#f59e0b', '#ef4444']}
          />
        )}

        {showSkeleton ? (
          // Skeleton loading state
          <div className="space-y-6">
            <div className="text-center py-4">
              <Skeleton className="w-16 h-16 rounded-full mx-auto mb-4" />
              <Skeleton className="h-6 w-48 mx-auto mb-2" />
              <Skeleton className="h-4 w-64 mx-auto" />
            </div>

            <Card className="card p-4">
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
              {/* Celebration particles */}
              <div className="absolute inset-0 pointer-events-none">
                {[...Array(20)].map((_, i) => (
                  <motion.div
                    key={i}
                    className="absolute w-2 h-2 bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] rounded-full"
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
                className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4 relative z-10"
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ type: "spring", bounce: 0.5, delay: 0.2 }}
              >
                <motion.div
                  initial={{ scale: 0, rotate: -180 }}
                  animate={{ scale: 1, rotate: 0 }}
                  transition={{ type: "spring", bounce: 0.6, delay: 0.4 }}
                >
                  <Check className="w-8 h-8 text-green-500" />
                </motion.div>
              </motion.div>

              <motion.h3
                className="text-xl font-semibold text-text-primary mb-2"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.6 }}
              >
                Function Deployed! 🎉
              </motion.h3>

              <motion.p
                className="text-text-secondary mb-4"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.8 }}
              >
                Your function is now live and ready to handle requests.
              </motion.p>
            </div>

            <Card className="card p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-text-secondary">Function URL</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(deployedUrl)}
                >
                  <Copy className="w-4 h-4" />
                </Button>
              </div>
              <code className="block p-3 bg-bg-primary rounded text-sm text-[#6366f1] break-all">
                {deployedUrl}
              </code>
            </Card>

            {/* Real-time Metrics */}
            <Card className="card p-4">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-medium text-text-primary">Live Metrics</h4>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={loadFunctionMetrics}
                  disabled={isLoadingMetrics}
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
                    <div className="text-2xl font-bold text-text-primary">
                      {functionMetrics.requests}
                    </div>
                    <div className="text-xs text-text-muted">Requests</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-text-primary">
                      {functionMetrics.latency}ms
                    </div>
                    <div className="text-xs text-text-muted">Avg Latency</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-red-500">
                      {functionMetrics.errors}
                    </div>
                    <div className="text-xs text-text-muted">Errors</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-green-500">
                      {functionMetrics.uptime.toFixed(1)}%
                    </div>
                    <div className="text-xs text-text-muted">Uptime</div>
                  </div>
                </div>
              ) : (
                <div className="text-center text-text-muted py-4">
                  <div className="text-sm">Loading metrics...</div>
                </div>
              )}
            </Card>

            <div className="flex gap-3">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => window.open(deployedUrl, "_blank")}
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                Open Function
              </Button>
            </div>

            {/* Progressive Disclosure - Advanced Options */}
            <div className="mt-6 pt-4 border-t border-border-subtle">
              <Button
                variant="ghost"
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="w-full justify-between"
              >
                <span className="text-sm font-medium">Advanced Configuration</span>
                <HelpTooltip content="Configure environment variables, scaling settings, and other advanced options for your function." />
              </Button>

              {showAdvanced && (
                <motion.div
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: "auto" }}
                  className="mt-4 space-y-4"
                >
                  <Card className="card p-4">
                    <h4 className="font-medium text-text-primary mb-3 flex items-center gap-2">
                      Environment Variables
                      <HelpTooltip content="Add environment variables to configure your function at runtime. These are securely stored and encrypted." />
                    </h4>
                    <div className="space-y-2">
                      <div className="flex gap-2">
                        <Input placeholder="KEY" className="flex-1" />
                        <Input placeholder="VALUE" className="flex-1" type="password" />
                        <Button variant="outline" size="sm">Add</Button>
                      </div>
                      <p className="text-xs text-text-muted">
                        Common variables: API keys, database URLs, configuration settings
                      </p>
                    </div>
                  </Card>

                  <Card className="card p-4">
                    <h4 className="font-medium text-text-primary mb-3 flex items-center gap-2">
                      Scaling Settings
                      <HelpTooltip content="Control how your function scales with traffic. Higher limits handle more concurrent requests." />
                    </h4>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label className="text-sm">Max Concurrent Requests</Label>
                        <Input type="number" defaultValue="100" className="mt-1" />
                      </div>
                      <div>
                        <Label className="text-sm">Timeout (seconds)</Label>
                        <Input type="number" defaultValue="30" className="mt-1" />
                      </div>
                    </div>
                  </Card>

                  <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
                    <div className="flex items-start gap-3">
                      <Code className="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" />
                      <div>
                        <h4 className="font-medium text-blue-400 mb-1">Pro Tip</h4>
                        <p className="text-sm text-text-secondary">
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
        <TabsList className="grid w-full grid-cols-2 bg-bg-tertiary">
          <TabsTrigger value="preset" className="flex items-center gap-2">
            Sample Functions
            <HelpTooltip content="Choose from pre-built function examples to get started quickly. These demonstrate common patterns and are ready to deploy." />
          </TabsTrigger>
          <TabsTrigger value="custom" className="flex items-center gap-2">
            Custom Code
            <HelpTooltip content="Write your own function code from scratch. Advanced users can deploy custom logic, APIs, or integrations." />
          </TabsTrigger>
        </TabsList>

        <TabsContent value="preset" className="space-y-4">
          <div className="grid gap-3">
            {sampleFunctions.map((func) => (
              <Card
                key={func.id}
                className={`card p-4 cursor-pointer transition-all ${
                  selectedFunction === func.id
                    ? "border-[#6366f1] ring-1 ring-[#6366f1]"
                    : "hover:border-border-default"
                }`}
                onClick={() => setSelectedFunction(func.id)}
              >
                <div className="flex items-start gap-3">
                  <div className="w-10 h-10 rounded-lg bg-[#6366f1]/20 flex items-center justify-center flex-shrink-0">
                    <Code className="w-5 h-5 text-[#6366f1]" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium text-text-primary">{func.name}</h3>
                      <HelpTooltip content="Edge functions run close to your users worldwide, reducing latency and improving performance compared to traditional server-based functions." />
                      {selectedFunction === func.id && (
                        <Check className="w-4 h-4 text-green-500" />
                      )}
                    </div>
                    <p className="text-sm text-text-secondary">{func.description}</p>
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
              <Card className="card overflow-hidden">
                <div className="flex items-center justify-between px-4 py-2 bg-bg-tertiary border-b border-border-subtle">
                  <div className="flex items-center gap-2">
                    <Terminal className="w-4 h-4 text-text-muted" />
                    <span className="text-xs text-text-muted">Preview</span>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(selectedFunctionData.code)}
                  >
                    <Copy className="w-3 h-3" />
                  </Button>
                </div>
                <pre className="p-4 text-sm text-text-secondary overflow-x-auto">
                  <code>{selectedFunctionData.code}</code>
                </pre>
              </Card>
            </motion.div>
          )}
        </TabsContent>

        <TabsContent value="custom" className="space-y-4">
          <Card className="card p-4">
            <p className="text-sm text-text-secondary mb-4">
              You can deploy your own custom function code. For now, select one of our
              sample functions to continue with the onboarding.
            </p>
            <Button
              variant="outline"
              onClick={() => setActiveTab("preset")}
              className="w-full"
            >
              Choose a Sample Function
            </Button>
          </Card>
        </TabsContent>
      </Tabs>

      <div className="space-y-2">
        <Label htmlFor="functionName" className="flex items-center gap-2">
          Function Name
          <HelpTooltip content="Choose a unique name for your function. This will become part of your function's URL (e.g., my-function.functionfly.app)." />
        </Label>
        <Input
          id="functionName"
          placeholder="my-first-function"
          value={functionName}
          onChange={(e) => setFunctionName(e.target.value)}
          className="input"
        />
        <div className="text-xs text-text-muted space-y-1">
          <p>
            Use lowercase letters, numbers, and hyphens only. This will be used to generate your function URL.
          </p>
          <p className="text-[#6366f1]">
            Preview: {functionName ? `${functionName.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-')}.functionfly.app` : 'your-function.functionfly.app'}
          </p>
        </div>
      </div>

      <Button
        onClick={handleDeploy}
        disabled={(!selectedFunction && activeTab === "preset") || !functionName.trim() || isDeploying}
        className="btn-primary w-full"
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

      {/* Progress Indicator */}
      {isDeploying && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: "auto" }}
          className="space-y-2"
        >
          <div className="flex items-center justify-between text-sm">
            <span className="text-text-primary">{deployStep}</span>
            <span className="text-text-muted">{Math.round(deployProgress)}%</span>
          </div>
          <Progress value={deployProgress} className="h-2" />
        </motion.div>
      )}

      {deployError && (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-red-500/10 border border-red-500/20 rounded-lg p-4"
        >
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="font-medium text-red-400 mb-1">Deployment Failed</h4>
              <p className="text-sm text-red-300">{deployError}</p>
              <div className="flex gap-2 mt-3">
                <Button
                  variant="outline"
                  size="sm"
                  className="border-red-500/30 text-red-400 hover:bg-red-500/10"
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
                  className="border-blue-500/30 text-blue-400 hover:bg-blue-500/10"
                  onClick={() => {
                    setActiveTab("custom");
                    setDeployError(null);
                  }}
                >
                  Try Custom Code
                </Button>
              </div>
              <div className="mt-3 p-3 bg-bg-tertiary rounded text-xs text-text-muted">
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
