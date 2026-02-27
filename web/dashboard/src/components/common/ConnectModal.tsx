import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  Loader2,
  Key,
  Shield,
  ExternalLink,
  CheckCircle2,
  AlertCircle,
  Info,
  Globe,
  Zap,
  Settings
} from "lucide-react";
import { ProviderIcon } from "./ProviderIcon";

type ProviderType = "cloudflare" | "vercel" | "fly" | "deno" | "functionfly-edge";

interface ProviderConfig {
  name: string;
  displayName: string;
  description: string;
  icon: string;
  color: string;
  authMethods: ("oauth" | "api_key" | "credentials" | "none")[];
  features: string[];
  docsUrl: string;
  requiredScopes?: string[];
}

const providerConfigs: Record<ProviderType, ProviderConfig> = {
  cloudflare: {
    name: "cloudflare",
    displayName: "Cloudflare",
    description: "Deploy serverless functions to Cloudflare Workers with global edge network.",
    icon: "cloudflare",
    color: "#f48120",
    authMethods: ["oauth", "api_key"],
    features: ["Edge Runtime", "Global CDN", "Durable Objects", "KV Storage", "D1 Database"],
    docsUrl: "https://developers.cloudflare.com/workers/",
    requiredScopes: ["account:read", "user:read", "workers:write"],
  },
  vercel: {
    name: "vercel",
    displayName: "Vercel",
    description: "Deploy web applications and serverless functions with instant deployments.",
    icon: "vercel",
    color: "#ffffff",
    authMethods: ["oauth"],
    features: ["Instant Deployments", "Edge Network", "Preview Deployments", "Analytics"],
    docsUrl: "https://vercel.com/docs",
    requiredScopes: ["read", "write"],
  },
  fly: {
    name: "fly",
    displayName: "Fly.io",
    description: "Run your applications close to your users with global deployment.",
    icon: "fly",
    color: "#7b68ee",
    authMethods: ["oauth", "api_key"],
    features: ["Global Deployment", "Autoscaling", "Persistent Volumes", "Private Networking"],
    docsUrl: "https://fly.io/docs/",
    requiredScopes: ["read", "write", "admin"],
  },
  deno: {
    name: "deno",
    displayName: "Deno Deploy",
    description: "Deploy JavaScript and TypeScript functions with zero-configuration.",
    icon: "deno",
    color: "#ffffff",
    authMethods: ["oauth", "api_key"],
    features: ["TypeScript Support", "Zero Config", "Global Edge", "WebSocket Support"],
    docsUrl: "https://deno.com/deploy/docs",
    requiredScopes: ["read", "write"],
  },
  "functionfly-edge": {
    name: "functionfly-edge",
    displayName: "FunctionFly Edge",
    description: "Host your edge functions on FunctionFly's infrastructure - no deployment required. Just select your region and start deploying.",
    icon: "functionfly",
    color: "#6366f1",
    authMethods: ["none"],
    features: ["Zero Configuration", "Automatic Scaling", "Global Edge Network", "Managed Infrastructure"],
    docsUrl: "/docs/providers/functionfly-edge",
    requiredScopes: [],
  },
};

const apiKeySchema = z.object({
  apiKey: z.string().min(1, "API Key is required"),
  accountId: z.string().optional(),
  projectId: z.string().optional(),
});

const credentialsSchema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
  endpoint: z.string().url("Invalid endpoint URL").optional(),
});

type ApiKeyFormData = z.infer<typeof apiKeySchema>;
type CredentialsFormData = z.infer<typeof credentialsSchema>;

interface ConnectModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConnect: (provider: ProviderType, config: any) => Promise<void>;
  provider?: ProviderType;
  initialConfig?: any;
}

export function ConnectModal({
  isOpen,
  onClose,
  onConnect,
  provider,
  initialConfig,
}: ConnectModalProps) {
  const [selectedProvider, setSelectedProvider] = useState<ProviderType | null>(provider || null);
  const [authMethod, setAuthMethod] = useState<"oauth" | "api_key" | "credentials">("oauth");
  const [isConnecting, setIsConnecting] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<"idle" | "connecting" | "success" | "error">("idle");
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [oauthUrl, setOauthUrl] = useState<string | null>(null);

  const apiKeyForm = useForm<ApiKeyFormData>({
    resolver: zodResolver(apiKeySchema),
    defaultValues: {
      apiKey: initialConfig?.apiKey || "",
      accountId: initialConfig?.accountId || "",
      projectId: initialConfig?.projectId || "",
    },
  });

  const credentialsForm = useForm<CredentialsFormData>({
    resolver: zodResolver(credentialsSchema),
    defaultValues: {
      username: initialConfig?.username || "",
      password: initialConfig?.password || "",
      endpoint: initialConfig?.endpoint || "",
    },
  });

  useEffect(() => {
    if (provider) {
      setSelectedProvider(provider);
    }
  }, [provider]);

  const handleProviderSelect = (providerType: ProviderType) => {
    setSelectedProvider(providerType);
    setAuthMethod("oauth"); // Default to OAuth
    setConnectionStatus("idle");
    setConnectionError(null);
  };

  const handleOAuthConnect = async (providerType: ProviderType) => {
    setIsConnecting(true);
    setConnectionStatus("connecting");

    try {
      // Generate OAuth URL
      const response = await fetch(`/api/auth/${providerType}/oauth-url`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      });

      if (!response.ok) {
        throw new Error("Failed to generate OAuth URL");
      }

      const { url } = await response.json();
      setOauthUrl(url);

      // Open OAuth window
      const oauthWindow = window.open(url, `${providerType}-oauth`, "width=600,height=700");

      // Poll for completion
      const checkCompletion = setInterval(async () => {
        try {
          const statusResponse = await fetch(`/api/auth/${providerType}/status`);
          const { connected, error } = await statusResponse.json();

          if (connected) {
            clearInterval(checkCompletion);
            setConnectionStatus("success");
            setIsConnecting(false);

            if (oauthWindow) {
              oauthWindow.close();
            }

            // Auto-close after success
            setTimeout(() => {
              onClose();
              setOauthUrl(null);
            }, 2000);
          } else if (error) {
            clearInterval(checkCompletion);
            setConnectionStatus("error");
            setConnectionError(error);
            setIsConnecting(false);

            if (oauthWindow) {
              oauthWindow.close();
            }
          }
        } catch (err) {
          // Continue polling
        }
      }, 2000);

      // Cleanup after 5 minutes
      setTimeout(() => {
        clearInterval(checkCompletion);
        if (oauthWindow && !oauthWindow.closed) {
          oauthWindow.close();
        }
      }, 300000);

    } catch (error) {
      setConnectionStatus("error");
      setConnectionError(error instanceof Error ? error.message : "OAuth connection failed");
      setIsConnecting(false);
    }
  };

  const handleApiKeyConnect = async (data: ApiKeyFormData) => {
    if (!selectedProvider) return;

    setIsConnecting(true);
    setConnectionStatus("connecting");

    try {
      await onConnect(selectedProvider, {
        method: "api_key",
        ...data,
      });

      setConnectionStatus("success");
      setTimeout(() => onClose(), 2000);
    } catch (error) {
      setConnectionStatus("error");
      setConnectionError(error instanceof Error ? error.message : "API Key connection failed");
    } finally {
      setIsConnecting(false);
    }
  };

  const handleCredentialsConnect = async (data: CredentialsFormData) => {
    if (!selectedProvider) return;

    setIsConnecting(true);
    setConnectionStatus("connecting");

    try {
      await onConnect(selectedProvider, {
        method: "credentials",
        ...data,
      });

      setConnectionStatus("success");
      setTimeout(() => onClose(), 2000);
    } catch (error) {
      setConnectionStatus("error");
      setConnectionError(error instanceof Error ? error.message : "Credentials connection failed");
    } finally {
      setIsConnecting(false);
    }
  };

  const handleClose = () => {
    if (!isConnecting) {
      onClose();
      setSelectedProvider(null);
      setAuthMethod("oauth");
      setConnectionStatus("idle");
      setConnectionError(null);
      setOauthUrl(null);
      apiKeyForm.reset();
      credentialsForm.reset();
    }
  };

  const selectedConfig = selectedProvider ? providerConfigs[selectedProvider] : null;

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Shield className="w-5 h-5" />
            Connect Provider
          </DialogTitle>
          <DialogDescription>
            Connect your account to deploy functions and manage resources.
          </DialogDescription>
        </DialogHeader>

        {!selectedProvider ? (
          // Provider Selection
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              {Object.entries(providerConfigs).map(([key, config]) => (
                <Card
                  key={key}
                  className="cursor-pointer transition-all hover:shadow-md hover:border-indigo-500/50"
                  onClick={() => handleProviderSelect(key as ProviderType)}
                >
                  <CardContent className="p-4">
                    <div className="flex items-center gap-3 mb-3">
                      <ProviderIcon provider={key as ProviderType} size="md" />
                      <div>
                        <h3 className="font-medium">{config.displayName}</h3>
                        <p className="text-xs text-text-secondary">{config.description}</p>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {config.features.slice(0, 2).map((feature) => (
                        <Badge key={feature} variant="secondary" className="text-xs">
                          {feature}
                        </Badge>
                      ))}
                      {config.features.length > 2 && (
                        <Badge variant="secondary" className="text-xs">
                          +{config.features.length - 2} more
                        </Badge>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        ) : (
          // Connection Configuration
          <div className="space-y-6">
            {/* Selected Provider Header */}
            <Card className="bg-indigo-500/5 border-indigo-500/20">
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <ProviderIcon provider={selectedProvider} size="md" />
                    <div>
                      <h3 className="font-medium">{selectedConfig?.displayName}</h3>
                      <p className="text-sm text-text-secondary">{selectedConfig?.description}</p>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedProvider(null)}
                    className="text-text-secondary hover:text-white"
                  >
                    Change
                  </Button>
                </div>
              </CardContent>
            </Card>

            {/* Auth Method Selection */}
            <div className="space-y-4">
              <Label className="text-sm font-medium">Authentication Method</Label>
              <Tabs value={authMethod} onValueChange={(value) => setAuthMethod(value as any)}>
                <TabsList className="grid w-full grid-cols-3">
                  {selectedConfig?.authMethods.map((method) => (
                    <TabsTrigger key={method} value={method} className="text-xs">
                      {method === "oauth" && "OAuth"}
                      {method === "api_key" && "API Key"}
                      {method === "credentials" && "Credentials"}
                    </TabsTrigger>
                  ))}
                </TabsList>

                {/* OAuth Tab */}
                {selectedConfig?.authMethods.includes("oauth") && (
                  <TabsContent value="oauth" className="space-y-4">
                    <Card>
                      <CardContent className="pt-6">
                        <div className="text-center space-y-4">
                          <div className="mx-auto w-16 h-16 bg-indigo-500/10 rounded-full flex items-center justify-center">
                            <ExternalLink className="w-8 h-8 text-indigo-400" />
                          </div>
                          <div>
                            <h3 className="font-medium mb-2">Connect with OAuth</h3>
                            <p className="text-sm text-text-secondary mb-4">
                              Securely connect your {selectedConfig.displayName} account using OAuth 2.0
                            </p>
                            {selectedConfig.requiredScopes && (
                              <div className="text-left">
                                <p className="text-xs text-text-secondary mb-2">Required permissions:</p>
                                <div className="flex flex-wrap gap-1">
                                  {selectedConfig.requiredScopes.map((scope) => (
                                    <Badge key={scope} variant="outline" className="text-xs">
                                      {scope}
                                    </Badge>
                                  ))}
                                </div>
                              </div>
                            )}
                          </div>
                          <Button
                            onClick={() => handleOAuthConnect(selectedProvider)}
                            disabled={isConnecting}
                            className="w-full"
                          >
                            {isConnecting ? (
                              <>
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                Connecting...
                              </>
                            ) : (
                              <>
                                <ExternalLink className="w-4 h-4 mr-2" />
                                Connect with {selectedConfig.displayName}
                              </>
                            )}
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  </TabsContent>
                )}

                {/* API Key Tab */}
                {selectedConfig?.authMethods.includes("api_key") && (
                  <TabsContent value="api_key" className="space-y-4">
                    <Card>
                      <CardContent className="pt-6">
                        <form onSubmit={apiKeyForm.handleSubmit(handleApiKeyConnect)} className="space-y-4">
                          <div className="space-y-2">
                            <Label htmlFor="apiKey">
                              API Key <span className="text-red-500">*</span>
                            </Label>
                            <Input
                              id="apiKey"
                              type="password"
                              {...apiKeyForm.register("apiKey")}
                              placeholder="Enter your API key"
                              className={cn(apiKeyForm.formState.errors.apiKey && "border-red-500")}
                            />
                            {apiKeyForm.formState.errors.apiKey && (
                              <p className="text-xs text-red-500">{apiKeyForm.formState.errors.apiKey.message}</p>
                            )}
                          </div>

                          {(selectedProvider === "cloudflare" || selectedProvider === "fly") && (
                            <div className="space-y-2">
                              <Label htmlFor="accountId">Account ID</Label>
                              <Input
                                id="accountId"
                                {...apiKeyForm.register("accountId")}
                                placeholder="Enter account ID"
                              />
                            </div>
                          )}

                          {selectedProvider === "vercel" && (
                            <div className="space-y-2">
                              <Label htmlFor="projectId">Project ID</Label>
                              <Input
                                id="projectId"
                                {...apiKeyForm.register("projectId")}
                                placeholder="Enter project ID"
                              />
                            </div>
                          )}

                          <Button
                            type="submit"
                            disabled={!apiKeyForm.formState.isValid || isConnecting}
                            className="w-full"
                          >
                            {isConnecting ? (
                              <>
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                Connecting...
                              </>
                            ) : (
                              <>
                                <Key className="w-4 h-4 mr-2" />
                                Connect with API Key
                              </>
                            )}
                          </Button>
                        </form>
                      </CardContent>
                    </Card>
                  </TabsContent>
                )}

                {/* Credentials Tab */}
                {selectedConfig?.authMethods.includes("credentials") && (
                  <TabsContent value="credentials" className="space-y-4">
                    <Card>
                      <CardContent className="pt-6">
                        <form onSubmit={credentialsForm.handleSubmit(handleCredentialsConnect)} className="space-y-4">
                          <div className="space-y-2">
                            <Label htmlFor="username">
                              Username <span className="text-red-500">*</span>
                            </Label>
                            <Input
                              id="username"
                              {...credentialsForm.register("username")}
                              placeholder="Enter username"
                              className={cn(credentialsForm.formState.errors.username && "border-red-500")}
                            />
                            {credentialsForm.formState.errors.username && (
                              <p className="text-xs text-red-500">{credentialsForm.formState.errors.username.message}</p>
                            )}
                          </div>

                          <div className="space-y-2">
                            <Label htmlFor="password">
                              Password <span className="text-red-500">*</span>
                            </Label>
                            <Input
                              id="password"
                              type="password"
                              {...credentialsForm.register("password")}
                              placeholder="Enter password"
                              className={cn(credentialsForm.formState.errors.password && "border-red-500")}
                            />
                            {credentialsForm.formState.errors.password && (
                              <p className="text-xs text-red-500">{credentialsForm.formState.errors.password.message}</p>
                            )}
                          </div>

                          <div className="space-y-2">
                            <Label htmlFor="endpoint">Endpoint URL</Label>
                            <Input
                              id="endpoint"
                              {...credentialsForm.register("endpoint")}
                              placeholder="https://api.example.com"
                              className={cn(credentialsForm.formState.errors.endpoint && "border-red-500")}
                            />
                            {credentialsForm.formState.errors.endpoint && (
                              <p className="text-xs text-red-500">{credentialsForm.formState.errors.endpoint.message}</p>
                            )}
                          </div>

                          <Button
                            type="submit"
                            disabled={!credentialsForm.formState.isValid || isConnecting}
                            className="w-full"
                          >
                            {isConnecting ? (
                              <>
                                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                                Connecting...
                              </>
                            ) : (
                              <>
                                <Settings className="w-4 h-4 mr-2" />
                                Connect with Credentials
                              </>
                            )}
                          </Button>
                        </form>
                      </CardContent>
                    </Card>
                  </TabsContent>
                )}
              </Tabs>
            </div>

            {/* Status Messages */}
            {connectionStatus === "error" && connectionError && (
              <Alert className="border-red-500/20 bg-red-500/10">
                <AlertCircle className="h-4 w-4 text-red-500" />
                <AlertDescription className="text-red-600 dark:text-red-400">
                  {connectionError}
                </AlertDescription>
              </Alert>
            )}

            {connectionStatus === "success" && (
              <Alert className="border-green-500/20 bg-green-500/10">
                <CheckCircle2 className="h-4 w-4 text-green-500" />
                <AlertDescription className="text-green-600 dark:text-green-400">
                  Successfully connected to {selectedConfig?.displayName}!
                </AlertDescription>
              </Alert>
            )}

            {/* Documentation Link */}
            {selectedConfig && (
              <div className="text-center">
                <a
                  href={selectedConfig.docsUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 text-sm text-indigo-400 hover:text-indigo-300"
                >
                  <Info className="w-4 h-4" />
                  View {selectedConfig.displayName} documentation
                  <ExternalLink className="w-3 h-3" />
                </a>
              </div>
            )}
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={isConnecting}>
            {selectedProvider ? "Back" : "Cancel"}
          </Button>
          {!selectedProvider && (
            <div className="text-xs text-text-secondary">
              Need help? Check our{" "}
              <a href="/docs/providers" className="text-indigo-400 hover:text-indigo-300">
                provider setup guide
              </a>
            </div>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
