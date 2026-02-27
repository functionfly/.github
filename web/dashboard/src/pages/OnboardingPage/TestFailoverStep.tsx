import { useState } from "react";
import { motion } from "framer-motion";
import { Shield, Play, Check, Loader2, AlertTriangle, RefreshCw, Globe, Activity, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { HelpTooltip } from "@/components/ui/help-tooltip";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";
import { useOnboardingStore } from "@/stores/onboardingStore";
import Confetti from "react-confetti";

type TestStatus = "idle" | "running" | "success" | "failed";

interface TestResult {
  region: string;
  provider: string;
  status: "success" | "failed";
  latency: number;
}

export function TestFailoverStep() {
  const { updateStepData } = useOnboardingStore();
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
      // Simulate test steps with potential failure
      for (let i = 0; i < testSteps.length; i++) {
        setCurrentStep(i);

        // Simulate potential failure at different steps
        if (i === 2 && Math.random() < 0.2) { // 20% chance of failure during failover detection
          throw new Error("failover_detection");
        }

        await new Promise((resolve) => setTimeout(resolve, 1500));
        setProgress(((i + 1) / testSteps.length) * 100);

        // Add mock results for certain steps
        if (i === 0) {
          setResults((prev) => [
            ...prev,
            { region: "US-East", provider: "Cloudflare", status: "success", latency: 45 },
          ]);
        } else if (i === 3) {
          setResults((prev) => [
            ...prev,
            { region: "US-West", provider: "Vercel", status: "success", latency: 62 },
          ]);
        }
      }

      setShowSkeleton(false);
      setTestStatus("success");
      setShowConfetti(true);

      // Save step data to onboarding store
      updateStepData("test-failover", {
        testResults: results,
        testCompletedAt: new Date().toISOString(),
        testStatus: "success",
      });

      // Hide confetti after 3 seconds
      setTimeout(() => setShowConfetti(false), 3000);

      toast.success("Failover test completed successfully!");
    } catch (error) {
      setShowSkeleton(false);
      setTestStatus("failed");

      let errorMessage = "Failover test failed due to an unexpected error.";
      let suggestion = "Please check your provider connections and try again.";

      if (error instanceof Error && error.message === "failover_detection") {
        errorMessage = "Automatic failover detection failed.";
        suggestion = "This could be due to network issues or provider API problems. Verify your API tokens are still valid and try again.";
      }

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
            numberOfPieces={100}
            gravity={0.3}
            colors={['#6366f1', '#8b5cf6', '#10b981', '#f59e0b', '#ef4444']}
          />
        )}
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="space-y-6"
        >
        <div className="text-center py-4 relative">
          {testStatus === "success" && !showSkeleton && (
            /* Celebration particles for success */
            <div className="absolute inset-0 pointer-events-none">
              {[...Array(25)].map((_, i) => (
                <motion.div
                  key={i}
                  className="absolute w-1 h-1 bg-gradient-to-r from-green-400 to-blue-500 rounded-full"
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
              testStatus === "success" ? "bg-green-500/20" : "bg-red-500/20"
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
                <Shield className="w-8 h-8 text-green-500" />
              ) : (
                <AlertTriangle className="w-8 h-8 text-red-500" />
              )}
            </motion.div>
          </motion.div>

          <motion.h3
            className="text-xl font-semibold text-text-primary mb-2"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
          >
            {testStatus === "success" ? "Failover Test Passed! 🛡️" : "Failover Test Failed"}
          </motion.h3>

          <motion.p
            className="text-text-secondary"
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
          <h4 className="text-sm font-medium text-text-primary">Test Results</h4>
          {showSkeleton ? (
            // Skeleton loading state
            <div className="space-y-3">
              {[1, 2].map((index) => (
                <Card key={index} className="card p-3 flex items-center justify-between">
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
                className="card p-3 flex items-center justify-between"
              >
                <div className="flex items-center gap-3">
                  <div
                    className={`w-8 h-8 rounded-full flex items-center justify-center ${
                      result.status === "success" ? "bg-green-500/20" : "bg-red-500/20"
                    }`}
                  >
                    {result.status === "success" ? (
                      <Check className="w-4 h-4 text-green-500" />
                    ) : (
                      <AlertTriangle className="w-4 h-4 text-red-500" />
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-text-primary">{result.provider}</p>
                    <p className="text-xs text-text-muted">{result.region}</p>
                  </div>
                </div>
                <div className="text-right">
                  <div className="flex items-center gap-1">
                    <p className="text-sm text-text-primary">{result.latency}ms</p>
                    <HelpTooltip content="Latency measures response time. Lower values mean faster performance for your users." />
                  </div>
                  <p className="text-xs text-text-muted">Latency</p>
                </div>
              </Card>
            ))
          )}
        </div>

        <div className="bg-green-500/10 border border-green-500/20 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <Activity className="w-5 h-5 text-green-500 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="font-medium text-green-400 mb-1">What happened?</h4>
              <p className="text-sm text-text-secondary">
                When the primary provider (Cloudflare) was temporarily unavailable,
                FunctionFly automatically routed traffic to your backup provider
                (Vercel) with zero downtime.
              </p>
            </div>
          </div>
        </div>

        {testError && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-red-500 flex-shrink-0 mt-0.5" />
              <div>
                <h4 className="font-medium text-red-400 mb-1">Test Error</h4>
                <p className="text-sm text-red-300">{testError}</p>
              </div>
            </div>
          </div>
        )}

        <Button
          variant="outline"
          onClick={runFailoverTest}
          className="w-full"
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
        <p className="text-text-secondary">
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
              <span className="text-text-primary">{testSteps[currentStep].name}</span>
              <span className="text-text-muted">{Math.round(progress)}%</span>
            </div>
            <Progress value={progress} className="h-2" />
            <p className="text-xs text-text-muted">
              {testSteps[currentStep].description}
            </p>
          </div>

          {results.length > 0 && (
            <div className="space-y-2">
              {results.map((result, index) => (
                <Card
                  key={index}
                  className="card p-3 flex items-center justify-between"
                >
                  <div className="flex items-center gap-3">
                    <Globe className="w-4 h-4 text-text-accent" />
                    <span className="text-sm text-text-primary">{result.provider}</span>
                  </div>
                  <Check className="w-4 h-4 text-green-500" />
                </Card>
              ))}
            </div>
          )}
        </motion.div>
      )}

      <Button
        onClick={runFailoverTest}
        disabled={testStatus === "running"}
        className="btn-primary w-full"
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
        <Card className="card p-3">
          <div className="flex items-center justify-center gap-1 mb-2">
            <Globe className="w-5 h-5 text-text-accent" />
            <HelpTooltip content="Your primary provider handles all traffic under normal conditions. It's the first choice for serving your functions." />
          </div>
          <p className="text-xs text-text-muted">Primary</p>
          <p className="text-sm font-medium text-text-primary">Cloudflare</p>
        </Card>
        <Card className="card p-3">
          <div className="flex items-center justify-center gap-1 mb-2">
            <Shield className="w-5 h-5 text-green-500" />
            <HelpTooltip content="Your backup provider automatically takes over if the primary fails. This ensures your functions stay available 24/7." />
          </div>
          <p className="text-xs text-text-muted">Backup</p>
          <p className="text-sm font-medium text-text-primary">Vercel</p>
        </Card>
      </div>
    </div>
  );
}
