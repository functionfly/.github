import { teamsApi } from '@/api/teams';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { motion } from 'framer-motion';
import {
  AlertCircle,
  AlertTriangle,
  Check,
  CheckCircle2,
  Cloud,
  ExternalLink,
  Key,
  Loader2,
  Wrench,
} from 'lucide-react';
import React, { useEffect, useState } from 'react';
import Confetti from 'react-confetti';
import { toast } from 'sonner';

interface CostEstimate {
  monthlyCost: number;
  currency: string;
  breakdown: Record<string, number>;
  providerData?: Record<string, any>;
}

type ProviderStatus = 'available' | 'maintenance' | 'outage';

const providers = [
  {
    id: 'cloudflare',
    name: 'Cloudflare Workers',
    description: "Deploy to Cloudflare's edge network",
    tooltip:
      'Cloudflare Workers run your code at the edge, closer to your users for faster response times. They use JavaScript and support multiple runtimes.',
    color: '#f48120',
    docsUrl: 'https://developers.cloudflare.com/workers/',
    requiresApiToken: true,
    status: 'available' as ProviderStatus,
  },
  {
    id: 'vercel',
    name: 'Vercel',
    description: 'Deploy serverless functions on Vercel',
    tooltip:
      "Vercel's serverless functions automatically scale with your traffic. They support multiple languages and integrate seamlessly with their frontend hosting.",
    color: '#000000',
    docsUrl: 'https://vercel.com/docs',
    requiresApiToken: true,
    status: 'available' as ProviderStatus,
  },
  {
    id: 'fly',
    name: 'Fly.io',
    description: 'Run your functions close to users',
    tooltip:
      'Fly.io allows you to deploy applications and functions to servers worldwide. Your code runs in containers distributed across multiple regions.',
    color: '#7b68ee',
    docsUrl: 'https://fly.io/docs/',
    requiresApiToken: true,
    status: 'maintenance' as ProviderStatus,
  },
  {
    id: 'functionfly-edge',
    name: 'FunctionFly Edge',
    description: "Host on FunctionFly's infrastructure",
    tooltip:
      'FunctionFly Edge is our managed hosting solution. No deployment required - just select your region and start deploying. Perfect for getting started quickly.',
    color: '#6366f1',
    docsUrl: 'https://functionfly.com/docs/providers/functionfly-edge',
    requiresApiToken: false,
    status: 'available' as ProviderStatus,
    isManaged: true,
  },
  {
    id: 'deno',
    name: 'Deno Deploy',
    description: 'Deploy to Deno Deploy global edge network',
    tooltip:
      'Deno Deploy runs your JavaScript and TypeScript at the edge, closer to your users. Built on V8 and designed for secure, serverless deployments.',
    color: '#000000',
    docsUrl: 'https://docs.deno.com/deploy/',
    requiresApiToken: true,
    status: 'available' as ProviderStatus,
  },
  {
    id: 'aws-lambda',
    name: 'AWS Lambda',
    description: 'Deploy serverless functions to AWS Lambda',
    tooltip:
      'AWS Lambda lets you run code without provisioning or managing servers. Pay only for the compute time you consume, with automatic scaling.',
    color: '#FF9900',
    docsUrl: 'https://docs.aws.amazon.com/lambda/',
    requiresApiToken: true,
    status: 'available' as ProviderStatus,
  },
];

type ValidationState = 'idle' | 'validating' | 'valid' | 'invalid';

interface ValidationResult {
  isValid: boolean;
  message?: string;
  suggestion?: string;
}

export function ConnectProviderStep() {
  const { updateStepData } = useOnboardingStore();
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
  const [apiToken, setApiToken] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [validationState, setValidationState] = useState<ValidationState>('idle');
  const [validationMessage, setValidationMessage] = useState<string>('');
  const [validationSuggestion, setValidationSuggestion] = useState<string>('');
  const [showConfetti, setShowConfetti] = useState(false);
  const [costEstimate, setCostEstimate] = useState<CostEstimate | null>(null);
  const [isEstimatingCost, setIsEstimatingCost] = useState(false);
  const [shareWithTeam, setShareWithTeam] = useState(false);
  const [isSharing, setIsSharing] = useState(false);

  const handleConnect = async () => {
    if (!selectedProvider || !apiToken || validationState !== 'valid') return;

    setIsConnecting(true);

    try {
      const cleanToken = apiToken.startsWith('Bearer ') ? apiToken.slice(7) : apiToken;

      const response = await fetch('/v1/providers/validate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          provider: selectedProvider,
          token: cleanToken,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        setIsConnecting(false);
        toast.error(
          errorData.message ||
            'Failed to connect provider. Please check your API token and try again.'
        );
        return;
      }

      const validationResult = await response.json();

      if (!validationResult.is_valid) {
        setIsConnecting(false);
        toast.error(
          validationResult.message ||
            'Failed to connect provider. Please check your API token and try again.'
        );
        return;
      }

      setIsConnecting(false);
      setIsConnected(true);
      setShowConfetti(true);

      const providerConfig: {
        id: string;
        provider: string;
        providerName: string;
        connectedAt: string;
        isShared?: boolean;
        teamId?: string;
      } = {
        id: `${selectedProvider}-${Date.now()}`,
        provider: selectedProvider,
        providerName: selectedProviderData?.name || selectedProvider,
        connectedAt: new Date().toISOString(),
      };

      updateStepData('connect-provider', {
        selectedProvider,
        providerName: selectedProviderData?.name,
        connectedAt: new Date().toISOString(),
        userId: validationResult.user_id,
        email: validationResult.email,
        providerConfig,
      });

      if (shareWithTeam) {
        setIsSharing(true);
        try {
          const { teams } = await teamsApi.list();
          const teamId = teams[0]?.id;

          if (!teamId) {
            toast.warning('Provider connected but no team found to share with yet.');
            return;
          }

          const shareResponse = await fetch(`/v1/providers/${providerConfig.id}/share`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              team_id: teamId,
            }),
          });

          if (shareResponse.ok) {
            providerConfig.isShared = true;
            providerConfig.teamId = teamId;
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

      setTimeout(() => setShowConfetti(false), 3000);

      toast.success(`${selectedProviderData?.name} connected successfully!`);
    } catch (error) {
      setIsConnecting(false);
      console.error('Connection error:', error);
      toast.error('Failed to connect provider. Please check your API token and try again.');
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
          function_name: 'sample-function',
          runtime: 'cloudflare',
          memory_mb: 128,
          requests_per_day: 1000,
          compute_duration_ms: 100,
          regions: ['us-east-1'],
        }),
      });

      if (response.ok) {
        const estimate: CostEstimate = await response.json();
        setCostEstimate(estimate);
      }
    } catch (error) {
      console.error('Cost estimation error:', error);
    } finally {
      setIsEstimatingCost(false);
    }
  };

  const selectedProviderData = providers.find((p) => p.id === selectedProvider);

  const validateProviderToken = async (
    token: string,
    provider: string
  ): Promise<ValidationResult> => {
    switch (provider) {
      case 'cloudflare':
        return await validateCloudflareToken(token);
      case 'vercel':
        return await validateVercelToken(token);
      case 'fly':
        return await validateFlyToken(token);
      case 'deno':
        return await validateDenoToken(token);
      case 'aws-lambda':
        return await validateAWSLambdaToken(token);
      default:
        return { isValid: false, message: 'Unsupported provider' };
    }
  };

  const validateCloudflareToken = async (token: string): Promise<ValidationResult> => {
    try {
      const cleanToken = token.startsWith('Bearer ') ? token.slice(7) : token;

      const response = await fetch('https://api.cloudflare.com/client/v4/user/tokens/verify', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${cleanToken}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        return {
          isValid: false,
          message: 'Invalid Cloudflare API token',
          suggestion:
            'Please check your API token in Cloudflare dashboard and ensure it has the correct permissions',
        };
      }

      const data = await response.json();
      if (data.success && data.result.status === 'active') {
        return { isValid: true };
      }

      return {
        isValid: false,
        message: 'Cloudflare API token is not active',
        suggestion: 'Please regenerate your API token in Cloudflare dashboard',
      };
    } catch (error) {
      return {
        isValid: false,
        message: 'Unable to connect to Cloudflare API',
        suggestion: 'Check your internet connection and try again',
      };
    }
  };

  const validateVercelToken = async (token: string): Promise<ValidationResult> => {
    try {
      const cleanToken = token.startsWith('Bearer ') ? token.slice(7) : token;

      const response = await fetch('https://api.vercel.com/v2/user', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${cleanToken}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.status === 401) {
        return {
          isValid: false,
          message: 'Invalid Vercel API token',
          suggestion: 'Please check your API token in Vercel dashboard',
        };
      }

      if (!response.ok) {
        return {
          isValid: false,
          message: 'Unable to validate Vercel token',
          suggestion: 'Please try again or check your token permissions',
        };
      }

      const data = await response.json();
      if (data.user) {
        return { isValid: true };
      }

      return {
        isValid: false,
        message: 'Vercel API token validation failed',
        suggestion: 'Please regenerate your API token in Vercel dashboard',
      };
    } catch (error) {
      return {
        isValid: false,
        message: 'Unable to connect to Vercel API',
        suggestion: 'Check your internet connection and try again',
      };
    }
  };

  const validateFlyToken = async (token: string): Promise<ValidationResult> => {
    try {
      const cleanToken = token.startsWith('Bearer ') ? token.slice(7) : token;

      const response = await fetch('https://api.fly.io/v1/apps', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${cleanToken}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.status === 401) {
        return {
          isValid: false,
          message: 'Invalid Fly.io API token',
          suggestion: 'Please check your API token in Fly.io dashboard',
        };
      }

      if (!response.ok) {
        return {
          isValid: false,
          message: 'Unable to validate Fly.io token',
          suggestion: 'Please try again or check your token permissions',
        };
      }

      return { isValid: true };
    } catch (error) {
      return {
        isValid: false,
        message: 'Unable to connect to Fly.io API',
        suggestion: 'Check your internet connection and try again',
      };
    }
  };

  const validateDenoToken = async (token: string): Promise<ValidationResult> => {
    try {
      const cleanToken = token.startsWith('Bearer ') ? token.slice(7) : token;

      const response = await fetch('https://api.deno.com/v1/user', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${cleanToken}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.status === 401) {
        return {
          isValid: false,
          message: 'Invalid Deno Deploy API token',
          suggestion: 'Please check your API token in the Deno Deploy dashboard',
        };
      }

      if (!response.ok) {
        return {
          isValid: false,
          message: 'Unable to validate Deno Deploy token',
          suggestion: 'Please try again or check your token permissions',
        };
      }

      return { isValid: true };
    } catch (error) {
      return {
        isValid: false,
        message: 'Unable to connect to Deno Deploy API',
        suggestion: 'Check your internet connection and try again',
      };
    }
  };

  const validateAWSLambdaToken = async (token: string): Promise<ValidationResult> => {
    try {
      const cleanToken = token.startsWith('Bearer ') ? token.slice(7) : token;

      const response = await fetch('https://lambda.amazonaws.com/2018-10-17/functions/', {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${cleanToken}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.status === 401 || response.status === 403) {
        return {
          isValid: false,
          message: 'Invalid AWS credentials',
          suggestion: 'Please check your AWS access key and secret in the AWS dashboard',
        };
      }

      if (!response.ok) {
        return {
          isValid: false,
          message: 'Unable to validate AWS credentials',
          suggestion: 'Please try again or check your IAM permissions',
        };
      }

      return { isValid: true };
    } catch (error) {
      return {
        isValid: false,
        message: 'Unable to connect to AWS Lambda API',
        suggestion: 'Check your internet connection and try again',
      };
    }
  };

  const validateApiToken = async (token: string, provider: string): Promise<ValidationResult> => {
    if (!token.trim()) {
      return { isValid: false };
    }

    const cleanToken = token.startsWith('Bearer ') ? token.slice(7) : token;

    const basicValidation = {
      cloudflare: cleanToken.length > 40 && /^[a-zA-Z0-9_-]+$/.test(cleanToken),
      vercel:
        cleanToken.length > 20 &&
        (cleanToken.startsWith('vercel_') || /^[a-zA-Z0-9_-]+$/.test(cleanToken)),
      fly:
        cleanToken.length > 20 && !cleanToken.includes(' ') && /^[A-Za-z0-9_-]+$/.test(cleanToken),
      deno: cleanToken.length > 20 && /^[A-Za-z0-9_-]+$/.test(cleanToken),
      'aws-lambda':
        cleanToken.length > 20 && !cleanToken.includes(' ') && /^[A-Za-z0-9+/=]+$/.test(cleanToken),
    };

    if (!basicValidation[provider as keyof typeof basicValidation]) {
      return {
        isValid: false,
        message: 'Invalid token format',
        suggestion:
          provider === 'cloudflare'
            ? 'Cloudflare API tokens are long alphanumeric strings (40+ characters). Get yours from the Cloudflare dashboard under My Profile > API Tokens.'
            : provider === 'vercel'
              ? 'Vercel API tokens start with "vercel_" or are long alphanumeric strings. Create one in your Vercel dashboard under Account Settings > Tokens.'
              : provider === 'fly'
                ? 'Fly.io API tokens are alphanumeric strings without spaces. Create one using "flyctl tokens create" in your terminal or in the Fly.io dashboard.'
                : provider === 'deno'
                  ? 'Deno Deploy API tokens are alphanumeric strings. Create one in your Deno Deploy dashboard under Account Settings > API Tokens.'
                  : provider === 'aws-lambda'
                    ? 'AWS credentials consist of an Access Key ID and Secret Access Key. Create IAM credentials in the AWS Console with Lambda permissions.'
                    : 'Please check your API token format and try again.',
      };
    }

    setValidationState('validating');

    try {
      const validationResult = await validateProviderToken(token, provider);

      if (!validationResult.isValid) {
        return {
          isValid: false,
          message: validationResult.message || 'Invalid or expired API token',
          suggestion:
            validationResult.suggestion ||
            "Please check your API token in the provider's dashboard and ensure it has the correct permissions",
        };
      }

      return { isValid: true };
    } catch (error) {
      console.error(`Token validation error for ${provider}:`, error);
      return {
        isValid: false,
        message: 'Unable to validate token',
        suggestion: 'Check your internet connection or try again later',
      };
    }
  };

  useEffect(() => {
    if (!apiToken || !selectedProvider) {
      setValidationState('idle');
      setValidationMessage('');
      setValidationSuggestion('');
      return;
    }

    const timeoutId = setTimeout(async () => {
      const result = await validateApiToken(apiToken, selectedProvider);
      setValidationState(result.isValid ? 'valid' : 'invalid');
      setValidationMessage(result.message || '');
      setValidationSuggestion(result.suggestion || '');
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [apiToken, selectedProvider]);

  useEffect(() => {
    if (selectedProvider && validationState === 'valid') {
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
        <div className="w-16 h-16 bg-aviation-green-dim rounded-full flex items-center justify-center mx-auto mb-4">
          <Check className="w-8 h-8 text-aviation-green" />
        </div>
        <h3 className="text-xl font-mono font-semibold text-aviation-text-primary mb-2">
          {selectedProviderData?.name} Connected!
        </h3>
        <p className="text-aviation-text-secondary font-mono">Your provider is now ready to deploy functions.</p>
      </motion.div>
    );
  }

  const unavailableProviders = providers.filter((p) => p.status !== 'available');
  const availableProviders = providers.filter((p) => p.status === 'available');

  return (
    <>
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
      <div className="space-y-6">
        {unavailableProviders.length > 0 && (
          <div className="bg-aviation-cyan-dim border border-aviation-cyan/30 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-aviation-cyan flex-shrink-0 mt-0.5" />
              <div>
                <h4 className="font-mono font-medium text-aviation-cyan mb-1">
                  Some providers are currently unavailable
                </h4>
                <p className="text-sm font-mono text-aviation-text-secondary">
                  {unavailableProviders.length === 1
                    ? `${unavailableProviders[0].name} is undergoing maintenance.`
                    : `${unavailableProviders.length} providers are currently unavailable.`}{' '}
                  You can still get started with the {availableProviders.length} available provider
                  {availableProviders.length > 1 ? 's' : ''} below.
                </p>
              </div>
            </div>
          </div>
        )}

        {!selectedProvider ? (
          <div className="grid gap-4">
            {providers.map((provider) => {
              const isAvailable = provider.status === 'available';
              const statusIcon =
                provider.status === 'maintenance'
                  ? Wrench
                  : provider.status === 'outage'
                    ? AlertTriangle
                    : CheckCircle2;
              const statusColor =
                provider.status === 'maintenance'
                  ? 'text-aviation-cyan'
                  : provider.status === 'outage'
                    ? 'text-aviation-red'
                    : 'text-aviation-green';

              return (
                <Card
                  key={provider.id}
                  className={`aviation-instrument p-4 transition-all ${
                    isAvailable
                      ? 'cursor-pointer hover:border-aviation-amber-dim'
                      : 'opacity-60 cursor-not-allowed'
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
                        {React.createElement(statusIcon, {
                          className: `w-4 h-4 ${statusColor} bg-aviation-bg-primary rounded-full p-0.5`,
                        })}
                      </div>
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <h3
                          className={`font-mono font-semibold ${isAvailable ? 'text-aviation-text-primary' : 'text-aviation-text-muted'}`}
                        >
                          {provider.name}
                        </h3>
                        <HelpTooltip content={provider.tooltip} />
                        {!isAvailable && (
                          <span
                            className={`text-xs px-2 py-0.5 rounded-full font-mono font-medium ${
                              provider.status === 'maintenance'
                                ? 'bg-aviation-cyan/20 text-aviation-cyan'
                                : 'bg-aviation-red/20 text-aviation-red'
                            }`}
                          >
                            {provider.status === 'maintenance' ? 'Maintenance' : 'Outage'}
                          </span>
                        )}
                      </div>
                      <p
                        className={`text-sm font-mono ${isAvailable ? 'text-aviation-text-secondary' : 'text-aviation-text-muted'}`}
                      >
                        {provider.description}
                      </p>
                      {!isAvailable && (
                        <p className="text-xs font-mono text-aviation-text-muted mt-1">
                          {provider.status === 'maintenance'
                            ? 'Temporarily unavailable for scheduled maintenance'
                            : 'Service is currently experiencing issues'}
                        </p>
                      )}
                    </div>
                    <Button variant="ghost" size="sm" disabled={!isAvailable} className="text-aviation-text-muted">
                      {isAvailable ? 'Connect' : 'Unavailable'}
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
            <div className="flex items-center gap-4 p-4 bg-aviation-bg-tertiary rounded-lg">
              <div
                className="w-12 h-12 rounded-lg flex items-center justify-center"
                style={{ backgroundColor: `${selectedProviderData?.color}20` }}
              >
                <Cloud className="w-6 h-6" style={{ color: selectedProviderData?.color }} />
              </div>
              <div className="flex-1">
                <h3 className="font-mono font-semibold text-aviation-text-primary">{selectedProviderData?.name}</h3>
                <a
                  href={selectedProviderData?.docsUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm font-mono text-aviation-cyan hover:underline inline-flex items-center gap-1"
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
                  setApiToken('');
                }}
                className="text-aviation-text-muted hover:text-aviation-amber"
              >
                Change
              </Button>
            </div>

            {selectedProvider && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                className="bg-aviation-bg-instrument rounded-lg p-4 border border-aviation-border-panel"
              >
                <div className="flex items-center gap-2 mb-3">
                  <span className="text-sm font-mono font-medium text-aviation-text-primary">Cost Preview</span>
                  <HelpTooltip content="Estimated monthly costs for a typical function deployment. Actual costs may vary based on usage." />
                </div>

                {isEstimatingCost ? (
                  <div className="flex items-center gap-2 text-aviation-text-muted">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span className="text-sm font-mono">Calculating costs...</span>
                  </div>
                ) : costEstimate ? (
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-mono text-aviation-text-secondary">Monthly Estimate</span>
                      <span className="text-lg font-mono font-semibold text-aviation-text-primary">
                        ${costEstimate.monthlyCost.toFixed(2)} {costEstimate.currency}
                      </span>
                    </div>

                    <div className="space-y-2">
                      <div className="text-xs font-mono text-aviation-text-muted">Breakdown:</div>
                      {Object.entries(costEstimate.breakdown).map(([key, value]) => (
                        <div key={key} className="flex justify-between text-sm">
                          <span className="text-aviation-text-secondary font-mono capitalize">
                            {key.replace('_', ' ')}
                          </span>
                          <span className="text-aviation-text-primary font-mono">${value.toFixed(3)}</span>
                        </div>
                      ))}
                    </div>

                    {costEstimate.providerData && (
                      <div className="text-xs font-mono text-aviation-text-muted pt-2 border-t border-aviation-border-panel">
                        Based on 1,000 requests/day, 100ms compute time, 128MB memory
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="text-sm font-mono text-aviation-text-muted">
                    Connect your provider to see cost estimates
                  </div>
                )}
              </motion.div>
            )}

            <div className="space-y-2">
              <Label htmlFor="apiToken" className="flex items-center gap-2 font-mono text-aviation-text-secondary">
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
                  className={`aviation-input pr-10 ${
                    validationState === 'invalid'
                      ? 'border-aviation-red focus:border-aviation-red'
                      : validationState === 'valid'
                        ? 'border-aviation-green focus:border-aviation-green'
                        : ''
                  }`}
                />
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                  {validationState === 'validating' && (
                    <Loader2 className="w-4 h-4 animate-spin text-aviation-text-muted" />
                  )}
                  {validationState === 'valid' && (
                    <CheckCircle2 className="w-4 h-4 text-aviation-green" />
                  )}
                  {validationState === 'invalid' && (
                    <AlertCircle className="w-4 h-4 text-aviation-red" />
                  )}
                </div>
              </div>
              <div className="text-xs font-mono space-y-1">
                {validationState === 'invalid' && validationMessage && (
                  <div className="flex items-start gap-2 text-aviation-red">
                    <AlertCircle className="w-3 h-3 flex-shrink-0 mt-0.5" />
                    <div>
                      <p className="font-medium">{validationMessage}</p>
                      {validationSuggestion && (
                        <p className="text-aviation-red/80 mt-1">{validationSuggestion}</p>
                      )}
                    </div>
                  </div>
                )}
                {validationState === 'valid' && (
                  <p className="text-aviation-green flex items-center gap-1">
                    <CheckCircle2 className="w-3 h-3" />
                    API token validated successfully
                  </p>
                )}
                {validationState === 'idle' && (
                  <>
                    <p className="text-aviation-text-muted">
                      Find your API token in {selectedProviderData?.name}'s dashboard under API
                      settings.
                    </p>
                    <p className="text-aviation-text-muted">
                      Your API token is securely encrypted and stored. We never share your
                      credentials.
                    </p>
                  </>
                )}
              </div>
            </div>

            <div className="flex items-center gap-3 p-3 bg-aviation-bg-tertiary rounded-lg">
              <input
                type="checkbox"
                id="shareWithTeam"
                checked={shareWithTeam}
                onChange={(e) => setShareWithTeam(e.target.checked)}
                className="w-4 h-4 text-aviation-amber border-aviation-border-instrument rounded focus:ring-aviation-amber focus:ring-1"
              />
              <div className="flex-1">
                <Label
                  htmlFor="shareWithTeam"
                  className="text-sm font-mono font-medium text-aviation-text-primary cursor-pointer"
                >
                  Share with Team
                </Label>
                <p className="text-xs font-mono text-aviation-text-muted">
                  Allow team members to use this provider for deployments
                </p>
              </div>
            </div>

            <Button
              onClick={handleConnect}
              disabled={
                !apiToken ||
                validationState === 'invalid' ||
                validationState === 'validating' ||
                isConnecting ||
                isSharing
              }
              className="aviation-button-primary w-full font-mono"
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
              ) : validationState === 'validating' ? (
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