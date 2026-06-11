import { useState } from "react";
import { motion } from "framer-motion";
import { Shield, Play, Check, Loader2, AlertTriangle, RefreshCw, Globe, Activity } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { HelpTooltip } from "@/components/ui/help-tooltip";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";
import { useOnboardingStore } from "@/stores/onboardingStore";
import Confetti from "react-confetti";
import { apiClient } from "@/api/client";

type TestStatus = "idle" | "running" | "success" | "failed";

interface TestResult {
  region: string;
  provider: string;
  status: "success" | "failed";
  latency: number;
}

interface FailoverTestResponse {
  success: boolean;
  message?: string;
  results: Array<{
    provider: string;
    region: string;
    status: string;
    latency_ms: number;
  }>;
  failover_occurred: boolean;
  test_duration_ms: number;
}

export function TestFailoverStep() {
  const { updateStepData, stepData } = useOnboardingStore();
  const [testStatus, setTestStatus] = useState<TestStatus>("idle");
  const [progress, setProgress] = useState(0);
  const [currentStep, setCurrentStep] = useState(0);
  const [results, setResults] = useState<TestResult[]>([]);
  const [testError, setTestError] = useState<string | null>(null);
  const [showSkeleton, setShowSkeleton] = useState(false);
  const [showConfetti, setShowConfetti] = useState(false);

  const testSteps = [
    { name: "Checking primary endpoint", description: "Testing connection to primary provider" },
    { name: "Simulating failure", description: "Temporarily disabling primary endpoint" },
    { name: "Detecting failover", description: "Waiting for automatic failover to trigger" },
    { name: "Verifying backup", description: "Confirming request handled by backup provider" },
    { name: "Restoring primary", description: "Bringing primary endpoint back online" },
  ];

  const runFailoverTest = async () => {
    setTestStatus("running");
    setProgress(0);
    setCurrentStep(0);
    setResults([]);
    setTestError(null);
    setShowSkeleton(true);

    try {
      const connectedProvider = stepData["connect-provider"]?.providerId || "cloudflare";
      const backupProvider = stepData["connect-provider"]?.backupProviderId;

      const response = await apiClient.post<FailoverTestResponse>(
        "/v1/providers/failover-test",
        {
          primary_provider_id: connectedProvider,
          backup_provider_id: backupProvider,
        }
      );

      for (let i = 0; i < testSteps.length; i++) {
        setCurrentStep(i);
        await new Promise((resolve) => setTimeout(resolve, 800));
        setProgress(((i + 1) / testSteps.length) * 100);
      }

      const mappedResults: TestResult[] = response.results.map(r => ({
        region: r.region,
        provider: r.provider,
        status: r.status as "success" | "failed",
        latency: r.latency_ms,
      }));
      setResults(mappedResults);

      setShowSkeleton(false);
      setTestStatus(response.success ? "success" : "failed");
      if (response.success) {
        setShowConfetti(true);
      }

      updateStepData("test-failover", {
        testResults: mappedResults,
        testCompletedAt: new Date().toISOString(),
        testStatus: response.success ? "success" : "failed",
        failoverOccurred: response.failover_occurred,
        testDurationMs: response.test_duration_ms,
      });

      setTimeout(() => setShowConfetti(false), 3000);
      toast.success(response.success ? "Failover test completed successfully!" : "Failover test failed");
    } catch (error: any) {
      setShowSkeleton(false);
      setTestStatus("failed");
      const errorMessage = error?.response?.data?.message || "Failover test failed due to an unexpected error.";
      const suggestion = "Please check your provider connections and try again.";
      setTestError(`${errorMessage} ${suggestion}`);
      toast.error("Failover test failed");
    }
  };

  if (testStatus === "success" || testStatus === "failed") {
    return (
      <>
        {showConfetti && testStatus === "success" && (
          <Confetti
            width={window.innerWidth}
            height={window.innerHeight}
            recycle={false}
            numberOfPieces={50}
            gravity={0.3}
            colors={['#f59e0b', '#ffb800', '#06b6d4', '#5b7cf5', '#10b981']}
          />
        )}
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="space-y-6"
        >
          <div className="text-center py-4 relative">
            {testStatus === "success" && !showSkeleton && (
              <div className="absolute inset-0 pointer-events-none">
                {[...Array(12)].map((_, i) => (
                  <motion.div
                    key={i}
                    className="absolute w-1 h-1 bg-gradient-to-r from-aviation-green to-aviation-cyan rounded-full"
                    initial={{
                      x: "50%",
                      y: "50%",
                      scale: 0,
                      opacity: 1
                    }}
                    animate={{
                      x: `${50 + (Math.random() - 0.5) * 250}%`,
                      y: `${50 + (Math.random() - 0.5) * 250}%`,
                      scale: [0, 1, 0],
                      opacity: [1, 1, 0]
                    }}
                    transition={{
                      duration: 3,
                      delay: Math.random() * 0.8,
                      ease: "easeOut"
                    }}
                    style={{
                      left: `${Math.random() * 100}%`,
                      top: `${Math.random() * 100}%`,
                    }}
                  />
                ))}
              </div>
            )}

            <motion.div
              className={`w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4 relative z-10 ${
                testStatus === "success" ? "bg-aviation-green-dim" : "bg-aviation-red-dim"
              }`}
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ type: "spring", bounce: 0.5, delay: 0.2 }}
            >
              <motion.div
                initial={{ scale: 0, rotate: testStatus === "success" ? -180 : 0 }}
                animate={{ scale: 1, rotate: 0 }}
                transition={{ type: "spring", bounce: 0.6, delay: 0.4 }}
              >
                {testStatus === "success" ? (
                  <Shield className="w-8 h-8 text-aviation-green" />
                ) : (
                  <AlertTriangle className="w-8 h-8 text-aviation-red" />
                )}
              </motion.div>
            </motion.div>

            <motion.h3
              className="text-xl font-mono font-bold text-aviation-text-primary mb-2"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.6 }}
            >
              {testStatus === "success" ? "Failover Test Passed!" : "Failover Test Failed"}
            </motion.h3>

            <motion.p
              className="text-aviation-text-secondary font-mono"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.8 }}
            >
              {testStatus === "success"
                ? "Your setup is working correctly. Automatic failover is ready."
                : "There was an issue with your failover configuration."
              }
            </motion.p>
          </div>

          <div className="space-y-3">
            <h4 className="text-sm font-mono font-medium text-aviation-text-primary">Test Results</h4>
            {showSkeleton ? (
              <div className="space-y-3">
                {[1, 2].map((index) => (
                  <Card key={index} className="aviation-instrument p-3 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Skeleton className="w-8 h-8 rounded-full" />
                      <div>
                        <Skeleton className="h-4 w-24 mb-1" />
                        <Skeleton className="h-3 w-16" />
                      </div>
                    </div>
                    <div className="text-right">
                      <Skeleton className="h-4 w-12 mb-1" />
                      <Skeleton className="h-3 w-16" />
                    </div>
                  </Card>
                ))}
              </div>
            ) : (
              results.map((result, index) => (
                <Card
                  key={index}
                  className="aviation-instrument p-3 flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={`w-8 h-8 rounded-full flex items-center justify-center ${
                        result.status === "success" ? "bg-aviation-green-dim" : "bg-aviation-red-dim"
                      }`}
                    >
                      {result.status === "success" ? (
                        <Check className="w-4 h-4 text-aviation-green" />
                      ) : (
                        <AlertTriangle className="w-4 h-4 text-aviation-red" />
                      )}
                    </div>
                    <div>
                      <p className="text-sm font-mono font-medium text-aviation-text-primary">{result.provider}</p>
                      <p className="text-xs font-mono text-aviation-text-muted">{result.region}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="flex items-center gap-1">
                      <p className="text-sm font-mono text-aviation-text-primary">{result.latency}ms</p>
                      <HelpTooltip content="Latency measures response time. Lower values mean faster performance for your users." />
                    </div>
                    <p className="text-xs font-mono text-aviation-text-muted">Latency</p>
                  </div>
                </Card>
              ))
            )}
          </div>

          <div className="bg-aviation-green-dim border border-aviation-green/30 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <Activity className="w-5 h-5 text-aviation-green flex-shrink-0 mt-0.5" />
              <div>
                <h4 className="font-mono font-medium text-aviation-green mb-1">What happened?</h4>
                <p className="text-sm font-mono text-aviation-text-secondary">
                  When the primary provider (Cloudflare) was temporarily unavailable,
                  FunctionFly automatically routed traffic to your backup provider
                  (Vercel) with zero downtime.
                </p>
              </div>
            </div>
          </div>

          {testError && (
            <div className="bg-aviation-red-dim border border-aviation-red/30 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <AlertTriangle className="w-5 h-5 text-aviation-red flex-shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-mono font-medium text-aviation-red mb-1">Test Error</h4>
                  <p className="text-sm font-mono text-aviation-red/80">{testError}</p>
                </div>
              </div>
            </div>
          )}

          <Button
            variant="outline"
            onClick={runFailoverTest}
            className="w-full font-mono border-aviation-border-instrument text-aviation-text-primary hover:border-aviation-amber hover:text-aviation-amber"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            {testStatus === "failed" ? "Retry Test" : "Run Test Again"}
          </Button>
        </motion.div>
      </>
    );
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <p className="text-aviation-text-secondary font-mono">
          Test your failover setup by simulating a provider failure.
          We'll verify that traffic automatically routes to your backup.
        </p>
        <div className="mt-2">
          <HelpTooltip
            content="Failover ensures your functions stay online even if one provider experiences issues. FunctionFly automatically routes traffic to healthy backup providers."
            className="inline-block"
          />
        </div>
      </div>

      {testStatus === "running" && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="space-y-4"
        >
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-aviation-text-primary font-mono">{testSteps[currentStep].name}</span>
              <span className="text-aviation-text-muted font-mono">{Math.round(progress)}%</span>
            </div>
            <Progress value={progress} className="h-2 bg-aviation-bg-tertiary [&>div]:bg-gradient-to-r [&>div]:from-aviation-amber [&>div]:to-aviation-cyan" />
            <p className="text-xs font-mono text-aviation-text-muted">
              {testSteps[currentStep].description}
            </p>
          </div>

          {results.length > 0 && (
            <div className="space-y-2">
              {results.map((result, index) => (
                <Card
                  key={index}
                  className="aviation-instrument p-3 flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <Globe className="w-4 h-4 text-aviation-cyan" />
                    <span className="text-sm font-mono text-aviation-text-primary">{result.provider}</span>
                  </div>
                  <Check className="w-4 h-4 text-aviation-green" />
                </Card>
              ))}
            </div>
          )}
        </motion.div>
      )}

      <Button
        onClick={runFailoverTest}
        disabled={testStatus === "running"}
        className="aviation-button-primary w-full font-mono"
      >
        {testStatus === "running" ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Testing Failover...
          </>
        ) : (
          <>
            <Play className="w-4 h-4 mr-2" />
            Run Failover Test
          </>
        )}
      </Button>

      <div className="grid grid-cols-2 gap-3 text-center">
        <Card className="aviation-instrument p-3">
          <div className="flex items-center justify-center gap-1 mb-2">
            <Globe className="w-5 h-5 text-aviation-cyan" />
            <HelpTooltip content="Your primary provider handles all traffic under normal conditions. It's the first choice for serving your functions." />
          </div>
          <p className="text-xs font-mono text-aviation-text-muted">Primary</p>
          <p className="text-sm font-mono font-semibold text-aviation-text-primary">Cloudflare</p>
        </Card>
        <Card className="aviation-instrument p-3">
          <div className="flex items-center justify-center gap-1 mb-2">
            <Shield className="w-5 h-5 text-aviation-green" />
            <HelpTooltip content="Your backup provider automatically takes over if the primary fails. This ensures your functions stay available 24/7." />
          </div>
          <p className="text-xs font-mono text-aviation-text-muted">Backup</p>
          <p className="text-sm font-mono font-semibold text-aviation-text-primary">Vercel</p>
        </Card>
      </div>
    </div>
  );
}