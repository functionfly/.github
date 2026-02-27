import React, { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { Cloud, Check, Key, ExternalLink, Loader2, AlertCircle, CheckCircle2, Wrench, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { HelpTooltip } from "@/components/ui/help-tooltip";
import { toast } from "sonner";
import { useOnboardingStore } from "@/stores/onboardingStore";
import Confetti from "react-confetti";

interface CostEstimate {
  monthlyCost: number;
  currency: string;
  breakdown: Record<string, number>;
  providerData?: Record<string, any>;
}

type ProviderStatus = "available" | "maintenance" | "outage";

const providers = [
  {
    id: "cloudflare",
    name: "Cloudflare Workers",
    description: "Deploy to Cloudflare's edge network",
    tooltip: "Cloudflare Workers run your code at the edge, closer to your users for faster response times. They use JavaScript and support multiple runtimes.",
    color: "#f48120",
    docsUrl: "https://developers.cloudflare.com/workers/",
    requiresApiToken: true,
    status: "available" as ProviderStatus,
  },
  {
    id: "vercel",
    name: "Vercel",
    description: "Deploy serverless functions on Vercel",
    tooltip: "Vercel's serverless functions automatically scale with your traffic. They support multiple languages and integrate seamlessly with their frontend hosting.",
    color: "#000000",
    docsUrl: "https://vercel.com/docs",
    requiresApiToken: true,
    status: "available" as ProviderStatus,
  },
  {
    id: "fly",
    name: "Fly.io",
    description: "Run your functions close to users",
    tooltip: "Fly.io allows you to deploy applications and functions to servers worldwide. Your code runs in containers distributed across multiple regions.",
    color: "#7b68ee",
    docsUrl: "https://fly.io/docs/",
    requiresApiToken: true,
    status: "maintenance" as ProviderStatus,
  },
  {
    id: "functionfly-edge",
    name: "FunctionFly Edge",
    description: "Host on FunctionFly's infrastructure",
    tooltip: "FunctionFly Edge is our managed hosting solution. No deployment required - just select your region and start deploying. Perfect for getting started quickly.",
    color: "#6366f1",
    docsUrl: "https://functionfly.com/docs/providers/functionfly-edge",
    requiresApiToken: false,
    status: "available" as ProviderStatus,
    isManaged: true,
  },
];

type ValidationState = "idle" | "validating" | "valid" | "invalid";

interface ValidationResult {
  isValid: boolean;
  message?: string;
  suggestion?: string;
}

export function ConnectProviderStep() {
  const { updateStepData } = useOnboardingStore();
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
  const [apiToken, setApiToken] = useState("");
  const [isConnecting, setIsConnecting] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [validationState, setValidationState] = useState<ValidationState>("idle");
  const [validationMessage, setValidationMessage] = useState<string>("");
  const [validationSuggestion, setValidationSuggestion] = useState<string>("");
  const [showConfetti, setShowConfetti] = useState(false);
  const [costEstimate, setCostEstimate] = useState<CostEstimate | null>(null);
  const [isEstimatingCost, setIsEstimatingCost] = useState(false);
  const [shareWithTeam, setShareWithTeam] = useState(false);
  const [isSharing, setIsSharing] = useState(false);

  const handleConnect = async () => {
    if (!selectedProvider || !apiToken || validationState !== "valid") return;

    setIsConnecting(true);

    try {
      // Use backend API for provider validation
      const response = await fetch('/v1/providers/validate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          provider: selectedProvider,
          token: apiToken,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        setIsConnecting(false);
        toast.error(errorData.message || "Failed to connect provider. Please check your API token and try again.");
        return;
      }

      const validationResult = await response.json();

      if (!validationResult.is_valid) {
        setIsConnecting(false);
        toast.error(validationResult.message || "Failed to connect provider. Please check your API token and try again.");
        return;
      }

      setIsConnecting(false);
      setIsConnected(true);
      setShowConfetti(true);

      // Create provider config
      const providerConfig = {
        id: `${selectedProvider}-${Date.now()}`, // Generate temporary ID
        provider: selectedProvider,
        providerName: selectedProviderData?.name || selectedProvider,
        connectedAt: new Date().toISOString(),
      };

      // Save step data to onboarding store
      updateStepData("connect-provider", {
        selectedProvider,
        providerName: selectedProviderData?.name,
        connectedAt: new Date().toISOString(),
        userId: validationResult.user_id,
        email: validationResult.email,
        providerConfig,
      });

      // Share with team if requested
      if (shareWithTeam) {
        setIsSharing(true);
        try {
          // Note: In a real implementation, you'd get the team ID from user context
          // For now, we'll assume the user has a team or will create one
          const shareResponse = await fetch(`/v1/providers/${providerConfig.id}/share`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              team_id: 'current-user-team', // This would come from user context
            }),
          });

          if (shareResponse.ok) {
            providerConfig.isShared = true;
            providerConfig.teamId = 'current-user-team';
            toast.success(`Provider shared with team successfully!`);
          } else {
            toast.warning('Provider connected but could not be shared with team');
          }
        } catch (error) {
          console.error('Provider sharing error:', error);
          toast.warning('Provider connected but could not be shared with team');
        } finally {
          setIsSharing(false);
        }
      }

      // Hide confetti after 3 seconds
      setTimeout(() => setShowConfetti(false), 3000);

      toast.success(`${selectedProviderData?.name} connected successfully!`);
    } catch (error) {
      setIsConnecting(false);
      console.error("Connection error:", error);
      toast.error("Failed to connect provider. Please check your API token and try again.");
    }
  };

  const estimateCost = async (provider: string) => {
    setIsEstimatingCost(true);
    try {
      const response = await fetch('/v1/providers/cost-estimate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          provider,
          function_name: "sample-function",
          runtime: "cloudflare",
          memory_mb: 128,
          requests_per_day: 1000,
          compute_duration_ms: 100,
          regions: ["us-east-1"],
        }),
      });

      if (response.ok) {
        const estimate: CostEstimate = await response.json();
        setCostEstimate(estimate);
      }
    } catch (error) {
      console.error("Cost estimation error:", error);
    } finally {
      setIsEstimatingCost(false);
    }
  };

  const selectedProviderData = providers.find((p) => p.id === selectedProvider);

  // Real API validation functions
  const validateProviderToken = async (token: string, provider: string): Promise<ValidationResult> => {
    switch (provider) {
      case 'cloudflare':
        return await validateCloudflareToken(token);
      case 'vercel':
        return await validateVercelToken(token);
      case 'fly':
        return await validateFlyToken(token);
      default:
        return { isValid: false, message: "Unsupported provider" };
    }
  };

  const validateCloudflareToken = async (token: string): Promise<ValidationResult> => {
    try {
      // Cloudflare API token validation
      const response = await fetch('https://api.cloudflare.com/client/v4/user/tokens/verify', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        return {
          isValid: false,
          message: "Invalid Cloudflare API token",
          suggestion: "Please check your API token in Cloudflare dashboard and ensure it has the correct permissions"
        };
      }

      const data = await response.json();
      if (data.success && data.result.status === 'active') {
        return { isValid: true };
      }

      return {
        isValid: false,
        message: "Cloudflare API token is not active",
        suggestion: "Please regenerate your API token in Cloudflare dashboard"
      };
    } catch (error) {
      return {
        isValid: false,
        message: "Unable to connect to Cloudflare API",
        suggestion: "Check your internet connection and try again"
      };
    }
  };

  const validateVercelToken = async (token: string): Promise<ValidationResult> => {
    try {
      // Vercel API token validation
      const response = await fetch('https://api.vercel.com/v2/user', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.status === 401) {
        return {
          isValid: false,
          message: "Invalid Vercel API token",
          suggestion: "Please check your API token in Vercel dashboard"
        };
      }

      if (!response.ok) {
        return {
          isValid: false,
          message: "Unable to validate Vercel token",
          suggestion: "Please try again or check your token permissions"
        };
      }

      const data = await response.json();
      if (data.user) {
        return { isValid: true };
      }

      return {
        isValid: false,
        message: "Vercel API token validation failed",
        suggestion: "Please regenerate your API token in Vercel dashboard"
      };
    } catch (error) {
      return {
        isValid: false,
        message: "Unable to connect to Vercel API",
        suggestion: "Check your internet connection and try again"
      };
    }
  };

  const validateFlyToken = async (token: string): Promise<ValidationResult> => {
    try {
      // Fly.io API token validation
      const response = await fetch('https://api.fly.io/v1/apps', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.status === 401) {
        return {
          isValid: false,
          message: "Invalid Fly.io API token",
          suggestion: "Please check your API token in Fly.io dashboard"
        };
      }

      if (!response.ok) {
        return {
          isValid: false,
          message: "Unable to validate Fly.io token",
          suggestion: "Please try again or check your token permissions"
        };
      }

      // If we can access apps, the token is valid
      return { isValid: true };
    } catch (error) {
      return {
        isValid: false,
        message: "Unable to connect to Fly.io API",
        suggestion: "Check your internet connection and try again"
      };
    }
  };

  const validateApiToken = async (token: string, provider: string): Promise<ValidationResult> => {
    if (!token.trim()) {
      return { isValid: false };
    }

    // Basic format validation
    const basicValidation = {
      cloudflare: token.length > 40 && token.startsWith('Bearer '),
      vercel: token.length > 20 && (token.startsWith('vercel_') || token.startsWith('Bearer ')),
      fly: token.length > 20 && !token.includes(' ') && /^[A-Za-z0-9_-]+$/.test(token),
    };

    if (!basicValidation[provider as keyof typeof basicValidation]) {
      return {
        isValid: false,
        message: "Invalid token format",
        suggestion: provider === 'cloudflare'
          ? 'Cloudflare API tokens should be Bearer tokens obtained from your Cloudflare dashboard API section'
          : provider === 'vercel'
          ? 'Vercel API tokens should start with "vercel_" and can be created in your Vercel dashboard'
          : 'Fly.io API tokens should be alphanumeric strings without spaces, created in your Fly.io dashboard'
      };
    }

    // Simulate API validation
    setValidationState("validating");

    try {
      // Real API validation for each provider
      const validationResult = await validateProviderToken(token, provider);

      if (!validationResult.isValid) {
        return {
          isValid: false,
          message: validationResult.message || "Invalid or expired API token",
          suggestion: validationResult.suggestion || "Please check your API token in the provider's dashboard and ensure it has the correct permissions"
        };
      }

      return { isValid: true };
    } catch (error) {
      console.error(`Token validation error for ${provider}:`, error);
      return {
        isValid: false,
        message: "Unable to validate token",
        suggestion: "Check your internet connection or try again later"
      };
    }
  };

  // Debounced validation effect
  useEffect(() => {
    if (!apiToken || !selectedProvider) {
      setValidationState("idle");
      setValidationMessage("");
      setValidationSuggestion("");
      return;
    }

    const timeoutId = setTimeout(async () => {
      const result = await validateApiToken(apiToken, selectedProvider);
      setValidationState(result.isValid ? "valid" : "invalid");
      setValidationMessage(result.message || "");
      setValidationSuggestion(result.suggestion || "");
    }, 500); // 500ms debounce

    return () => clearTimeout(timeoutId);
  }, [apiToken, selectedProvider]);

  // Effect to estimate costs when provider is selected
  useEffect(() => {
    if (selectedProvider && validationState === "valid") {
      estimateCost(selectedProvider);
    } else {
      setCostEstimate(null);
    }
  }, [selectedProvider, validationState]);

  if (isConnected) {
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        className="text-center py-8"
      >
        <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-4">
          <Check className="w-8 h-8 text-green-500" />
        </div>
        <h3 className="text-xl font-semibold text-text-primary mb-2">
          {selectedProviderData?.name} Connected!
        </h3>
        <p className="text-text-secondary">
          Your provider is now ready to deploy functions.
        </p>
      </motion.div>
    );
  }

  const unavailableProviders = providers.filter(p => p.status !== "available");
  const availableProviders = providers.filter(p => p.status === "available");

  return (
    <>
      {showConfetti && (
        <Confetti
          width={window.innerWidth}
          height={window.innerHeight}
          recycle={false}
          numberOfPieces={80}
          gravity={0.3}
          colors={['#6366f1', '#8b5cf6', '#10b981', '#f59e0b', '#ef4444']}
        />
      )}
      <div className="space-y-6">
      {unavailableProviders.length > 0 && (
        <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" />
            <div>
              <h4 className="font-medium text-yellow-400 mb-1">
                Some providers are currently unavailable
              </h4>
              <p className="text-sm text-yellow-300">
                {unavailableProviders.length === 1
                  ? `${unavailableProviders[0].name} is undergoing maintenance.`
                  : `${unavailableProviders.length} providers are currently unavailable.`
                } You can still get started with the {availableProviders.length} available provider{availableProviders.length > 1 ? 's' : ''} below.
              </p>
            </div>
          </div>
        </div>
      )}

      {!selectedProvider ? (
        <div className="grid gap-4">
          {providers.map((provider) => {
            const isAvailable = provider.status === "available";
            const statusIcon = provider.status === "maintenance" ? Wrench :
                              provider.status === "outage" ? AlertTriangle : CheckCircle2;
            const statusColor = provider.status === "maintenance" ? "text-yellow-500" :
                               provider.status === "outage" ? "text-red-500" : "text-green-500";

            return (
              <Card
                key={provider.id}
                className={`card p-4 transition-all ${
                  isAvailable
                    ? "cursor-pointer hover:border-[#6366f1]/50"
                    : "opacity-60 cursor-not-allowed"
                }`}
                onClick={() => isAvailable && setSelectedProvider(provider.id)}
              >
                <div className="flex items-center gap-4">
                  <div
                    className="w-12 h-12 rounded-lg flex items-center justify-center relative"
                    style={{ backgroundColor: `${provider.color}20` }}
                  >
                    <Cloud className="w-6 h-6" style={{ color: provider.color }} />
                    <div className="absolute -top-1 -right-1">
                      {React.createElement(statusIcon, {className: `w-4 h-4 ${statusColor} bg-bg-primary rounded-full p-0.5`})}
                    </div>
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h3 className={`font-medium ${isAvailable ? "text-text-primary" : "text-text-muted"}`}>
                        {provider.name}
                      </h3>
                      <HelpTooltip content={provider.tooltip} />
                      {!isAvailable && (
                        <span className={`text-xs px-2 py-0.5 rounded-full ${
                          provider.status === "maintenance"
                            ? "bg-yellow-500/20 text-yellow-400"
                            : "bg-red-500/20 text-red-400"
                        }`}>
                          {provider.status === "maintenance" ? "Maintenance" : "Outage"}
                        </span>
                      )}
                    </div>
                    <p className={`text-sm ${isAvailable ? "text-text-secondary" : "text-text-muted"}`}>
                      {provider.description}
                    </p>
                    {!isAvailable && (
                      <p className="text-xs text-text-muted mt-1">
                        {provider.status === "maintenance"
                          ? "Temporarily unavailable for scheduled maintenance"
                          : "Service is currently experiencing issues"
                        }
                      </p>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={!isAvailable}
                  >
                    {isAvailable ? "Connect" : "Unavailable"}
                  </Button>
                </div>
              </Card>
            );
          })}
        </div>
      ) : (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="space-y-6"
        >
          <div className="flex items-center gap-4 p-4 bg-bg-tertiary rounded-lg">
            <div
              className="w-12 h-12 rounded-lg flex items-center justify-center"
              style={{ backgroundColor: `${selectedProviderData?.color}20` }}
            >
              <Cloud className="w-6 h-6" style={{ color: selectedProviderData?.color }} />
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-text-primary">{selectedProviderData?.name}</h3>
              <a
                href={selectedProviderData?.docsUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-text-accent hover:underline inline-flex items-center gap-1"
              >
                View Documentation
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setSelectedProvider(null);
                setApiToken("");
              }}
            >
              Change
            </Button>
          </div>

          {/* Cost Preview */}
          {selectedProvider && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              className="bg-bg-tertiary rounded-lg p-4 border border-border-subtle"
            >
              <div className="flex items-center gap-2 mb-3">
                <span className="text-sm font-medium text-text-primary">Cost Preview</span>
                <HelpTooltip content="Estimated monthly costs for a typical function deployment. Actual costs may vary based on usage." />
              </div>

              {isEstimatingCost ? (
                <div className="flex items-center gap-2 text-text-muted">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span className="text-sm">Calculating costs...</span>
                </div>
              ) : costEstimate ? (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-text-secondary">Monthly Estimate</span>
                    <span className="text-lg font-semibold text-text-primary">
                      ${costEstimate.monthlyCost.toFixed(2)} {costEstimate.currency}
                    </span>
                  </div>

                  <div className="space-y-2">
                    <div className="text-xs text-text-muted">Breakdown:</div>
                    {Object.entries(costEstimate.breakdown).map(([key, value]) => (
                      <div key={key} className="flex justify-between text-sm">
                        <span className="text-text-secondary capitalize">
                          {key.replace('_', ' ')}
                        </span>
                        <span className="text-text-primary">${value.toFixed(3)}</span>
                      </div>
                    ))}
                  </div>

                  {costEstimate.providerData && (
                    <div className="text-xs text-text-muted pt-2 border-t border-border-subtle">
                      Based on 1,000 requests/day, 100ms compute time, 128MB memory
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-sm text-text-muted">
                  Connect your provider to see cost estimates
                </div>
              )}
            </motion.div>
          )}

          <div className="space-y-2">
            <Label htmlFor="apiToken" className="flex items-center gap-2">
              <Key className="w-4 h-4" />
              API Token
              <HelpTooltip content="An API token is a secure key that allows FunctionFly to deploy functions to your provider account. You can generate this in your provider's dashboard under API settings." />
            </Label>
            <div className="relative">
              <Input
                id="apiToken"
                type="password"
                placeholder={`Enter your ${selectedProviderData?.name} API token`}
                value={apiToken}
                onChange={(e) => setApiToken(e.target.value)}
                className={`input pr-10 ${
                  validationState === "invalid"
                    ? "border-red-500 focus:border-red-500"
                    : validationState === "valid"
                    ? "border-green-500 focus:border-green-500"
                    : ""
                }`}
              />
              <div className="absolute right-3 top-1/2 -translate-y-1/2">
                {validationState === "validating" && (
                  <Loader2 className="w-4 h-4 animate-spin text-text-muted" />
                )}
                {validationState === "valid" && (
                  <CheckCircle2 className="w-4 h-4 text-green-500" />
                )}
                {validationState === "invalid" && (
                  <AlertCircle className="w-4 h-4 text-red-500" />
                )}
              </div>
            </div>
            <div className="text-xs space-y-1">
              {validationState === "invalid" && validationMessage && (
                <div className="flex items-start gap-2 text-red-400">
                  <AlertCircle className="w-3 h-3 flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="font-medium">{validationMessage}</p>
                    {validationSuggestion && (
                      <p className="text-red-300 mt-1">{validationSuggestion}</p>
                    )}
                  </div>
                </div>
              )}
              {validationState === "valid" && (
                <p className="text-green-400 flex items-center gap-1">
                  <CheckCircle2 className="w-3 h-3" />
                  API token validated successfully
                </p>
              )}
              {validationState === "idle" && (
                <>
                  <p className="text-text-muted">
                    Find your API token in {selectedProviderData?.name}'s dashboard under API settings.
                  </p>
                  <p className="text-text-muted">
                    Your API token is securely encrypted and stored. We never share your credentials.
                  </p>
                </>
              )}
            </div>
          </div>

          {/* Team Sharing Option */}
          <div className="flex items-center gap-3 p-3 bg-bg-tertiary rounded-lg">
            <input
              type="checkbox"
              id="shareWithTeam"
              checked={shareWithTeam}
              onChange={(e) => setShareWithTeam(e.target.checked)}
              className="w-4 h-4 text-[#6366f1] border-border-subtle rounded focus:ring-[#6366f1] focus:ring-1"
            />
            <div className="flex-1">
              <Label htmlFor="shareWithTeam" className="text-sm font-medium text-text-primary cursor-pointer">
                Share with Team
              </Label>
              <p className="text-xs text-text-muted">
                Allow team members to use this provider for deployments
              </p>
            </div>
          </div>

          <Button
            onClick={handleConnect}
            disabled={!apiToken || validationState === "invalid" || validationState === "validating" || isConnecting || isSharing}
            className="btn-primary w-full"
          >
            {isConnecting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Connecting to {selectedProviderData?.name}...
              </>
            ) : isSharing ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Sharing with team...
              </>
            ) : validationState === "validating" ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Validating token...
              </>
            ) : (
              `Connect ${selectedProviderData?.name}`
            )}
          </Button>
        </motion.div>
      )}
    </div>
    </>
  );
}
