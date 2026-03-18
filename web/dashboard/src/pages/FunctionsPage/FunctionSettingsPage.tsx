import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { apiClient } from "@/api/client";
import { registryApi, type FunctionSettingsResponse } from "@/api/registry";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  ArrowLeft,
  Save,
  Key,
  Globe,
  Shield,
  Bell,
  Trash2,
  Copy,
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  Plus,
  X,
} from "lucide-react";
import { usePlan } from "@/hooks/usePlan";
import { EnterpriseFeature } from "@/components/enterprise/EnterpriseFeature";
import {
  getCustomDomainsLimit,
  canAddCustomDomain,
  formatCustomDomainsRemaining,
} from "@/lib/plan-utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import "@/styles/components.css";
import { EmbedTab } from "@/components/embed";

type FunctionSettings = FunctionSettingsResponse & { webhookUrl?: string };

export function FunctionSettingsPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState("general");
  const [isSaving, setIsSaving] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showApiKeyDialog, setShowApiKeyDialog] = useState(false);
  const [apiKey, setApiKey] = useState("");
  const [isRegenerating, setIsRegenerating] = useState(false);

  // Form state
  const [functionSettings, setFunctionSettings] = useState<FunctionSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Form fields
  const [description, setDescription] = useState("");
  const [isPublic, setIsPublic] = useState(false);
  const [allowAnonymousInvoke, setAllowAnonymousInvoke] = useState(false);
  const [corsEnabled, setCorsEnabled] = useState(false);
  const [corsOrigins, setCorsOrigins] = useState("");
  const [timeout, setTimeout] = useState(30);
  const [memory, setMemory] = useState(128);
  const [runtime, setRuntime] = useState("python3.11");
  const [selectedProviders, setSelectedProviders] = useState<string[]>([]);

  // Environment variables
  const [envVars, setEnvVars] = useState<{ key: string; value: string }[]>([]);
  const [newEnvKey, setNewEnvKey] = useState("");
  const [newEnvValue, setNewEnvValue] = useState("");

  // Custom domains
  const [showAddDomainDialog, setShowAddDomainDialog] = useState(false);
  const [newDomainValue, setNewDomainValue] = useState("");
  const [addDomainError, setAddDomainError] = useState<string | null>(null);

  const { plan, hasFeature } = usePlan();

  // Fetch function settings
  useEffect(() => {
    const fetchFunctionSettings = async () => {
      if (!author || !name) return;

      try {
        setLoading(true);
        setError(null);

        const response = await registryApi.getFunctionSettings(author, name);
        setFunctionSettings(response);

        // Populate form fields
        setDescription(response.description || "");
        setIsPublic(response.isPublic || false);
        setAllowAnonymousInvoke(response.allowAnonymousInvoke || false);
        setCorsEnabled(response.corsEnabled || false);
        setCorsOrigins(response.corsOrigins?.join(", ") || "");
        setTimeout(response.timeout || 30);
        setMemory(response.memory || 128);
        setRuntime(response.runtime || "python3.11");
        setSelectedProviders(response.providers || []);

        // Parse environment variables
        if (response.environmentVariables) {
          setEnvVars(
            Object.entries(response.environmentVariables).map(([key, value]) => ({
              key,
              value,
            }))
          );
        }
      } catch (err) {
        console.error("Failed to load function settings:", err);
        setError("Failed to load function settings");
        toast.error("Failed to load function settings");
      } finally {
        setLoading(false);
      }
    };

    fetchFunctionSettings();
  }, [author, name]);

  const handleSave = async () => {
    if (!author || !name) return;

    setIsSaving(true);
    try {
      const settings = {
        description,
        isPublic,
        allowAnonymousInvoke,
        corsEnabled,
        corsOrigins: corsOrigins.split(",").map((o) => o.trim()).filter(Boolean),
        timeout,
        memory,
        runtime,
        providers: selectedProviders,
        environmentVariables: envVars.reduce((acc, { key, value }) => {
          if (key) acc[key] = value;
          return acc;
        }, {} as Record<string, string>),
        customDomains: functionSettings?.customDomains ?? [],
      };

      await apiClient.patch(`/v1/functions/${author}/${name}/settings`, settings);
      toast.success("Settings saved successfully");
    } catch (error) {
      console.error("Failed to save settings:", error);
      toast.error("Failed to save settings");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = () => {
    setShowDeleteDialog(true);
  };

  const customDomainsList = functionSettings?.customDomains ?? [];
  const customDomainsLimit = getCustomDomainsLimit(plan);
  const canAddMore = canAddCustomDomain(plan, customDomainsList.length);

  const validateDomain = (value: string): string | null => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) return "Enter a domain";
    if (/^https?:\/\//i.test(trimmed)) return "Do not include http:// or https://";
    if (/\s/.test(trimmed)) return "Domain cannot contain spaces";
    if (customDomainsList.includes(trimmed)) return "This domain is already added";
    return null;
  };

  const handleAddDomain = () => {
    setAddDomainError(null);
    const err = validateDomain(newDomainValue);
    if (err) {
      setAddDomainError(err);
      return;
    }
    const domain = newDomainValue.trim().toLowerCase();
    setFunctionSettings((prev) =>
      prev
        ? { ...prev, customDomains: [...(prev.customDomains ?? []), domain] }
        : prev
    );
    setNewDomainValue("");
    setShowAddDomainDialog(false);
  };

  const handleRemoveDomain = (domain: string) => {
    setFunctionSettings((prev) =>
      prev
        ? {
            ...prev,
            customDomains: (prev.customDomains ?? []).filter((d) => d !== domain),
          }
        : prev
    );
  };

  const confirmDelete = async () => {
    if (!author || !name || !functionSettings) return;

    try {
      setIsDeleting(true);
      await apiClient.delete(`/v1/functions/${author}/${name}`);
      toast.success(`Function "${functionSettings.name}" has been deleted`);
      navigate("/functions");
    } catch (error) {
      console.error("Failed to delete function:", error);
      toast.error("Failed to delete function");
    } finally {
      setIsDeleting(false);
      setShowDeleteDialog(false);
    }
  };

  const handleRegenerateApiKey = async () => {
    if (!author || !name) return;

    setIsRegenerating(true);
    try {
      const response = await apiClient.post<{ apiKey: string }>(
        `/v1/functions/${author}/${name}/regenerate-api-key`
      );
      setApiKey(response.apiKey);
      setShowApiKeyDialog(true);
    } catch (error) {
      console.error("Failed to regenerate API key:", error);
      toast.error("Failed to regenerate API key");
    } finally {
      setIsRegenerating(false);
    }
  };

  const addEnvironmentVariable = () => {
    if (newEnvKey && newEnvValue) {
      setEnvVars([...envVars, { key: newEnvKey, value: newEnvValue }]);
      setNewEnvKey("");
      setNewEnvValue("");
    }
  };

  const removeEnvironmentVariable = (index: number) => {
    setEnvVars(envVars.filter((_, i) => i !== index));
  };

  const copyApiKey = () => {
    navigator.clipboard.writeText(apiKey);
    toast.success("API key copied to clipboard");
  };

  if (loading || !functionSettings) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 bg-muted rounded animate-pulse" />
          <div className="space-y-2">
            <div className="w-48 h-6 bg-muted rounded animate-pulse" />
            <div className="w-32 h-4 bg-muted rounded animate-pulse" />
          </div>
        </div>
        <div className="p-6 border rounded-lg">
          <div className="w-full h-64 bg-muted rounded animate-pulse" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate(`/functions/${functionSettings.id}`)}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-white">
                {functionSettings.name}
              </h1>
              <Badge variant="secondary">Settings</Badge>
            </div>
            <p className="text-text-secondary">
              Configure function settings and preferences
            </p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline" onClick={() => navigate(`/functions/${functionSettings.id}`)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            <Save className="w-4 h-4 mr-2" />
            {isSaving ? "Saving..." : "Save Changes"}
          </Button>
        </div>
      </div>

      {/* Settings Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-6">
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
          <TabsTrigger value="providers">Providers</TabsTrigger>
          <TabsTrigger value="environment">Environment</TabsTrigger>
          <TabsTrigger value="embed">Embed</TabsTrigger>
          <TabsTrigger value="danger">Danger Zone</TabsTrigger>
        </TabsList>

        {/* General Settings */}
        <TabsContent value="general" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Basic Information</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Function Name</Label>
                  <Input value={functionSettings.name} disabled />
                  <p className="text-xs text-text-muted">
                    Function name cannot be changed
                  </p>
                </div>
                <div className="space-y-2">
                  <Label>Author</Label>
                  <Input value={functionSettings.author} disabled />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Describe what your function does"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="runtime">Runtime</Label>
                <Select value={runtime} onValueChange={setRuntime}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="python3.11">Python 3.11</SelectItem>
                    <SelectItem value="python3.12">Python 3.12</SelectItem>
                    <SelectItem value="nodejs18">Node.js 18</SelectItem>
                    <SelectItem value="nodejs20">Node.js 20</SelectItem>
                    <SelectItem value="deno">Deno</SelectItem>
                    <SelectItem value="rust">Rust</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <Separator />

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Public Function</Label>
                    <p className="text-sm text-text-muted">
                      Make this function visible in the public registry
                    </p>
                  </div>
                  <Switch
                    checked={isPublic}
                    onCheckedChange={setIsPublic}
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label>Allow Anonymous Invocation</Label>
                    <p className="text-sm text-text-muted">
                      Allow anyone to invoke this function without authentication
                    </p>
                  </div>
                  <Switch
                    checked={allowAnonymousInvoke}
                    onCheckedChange={setAllowAnonymousInvoke}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Execution Settings</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="timeout">Timeout (seconds)</Label>
                  <Input
                    id="timeout"
                    type="number"
                    value={timeout}
                    onChange={(e) => setTimeout(parseInt(e.target.value))}
                    min={1}
                    max={300}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="memory">Memory (MB)</Label>
                  <Input
                    id="memory"
                    type="number"
                    value={memory}
                    onChange={(e) => setMemory(parseInt(e.target.value))}
                    min={64}
                    max={1024}
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Security Settings */}
        <TabsContent value="security" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">API Authentication</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg bg-bg-tertiary">
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-primary/10 rounded-lg">
                    <Key className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <h4 className="font-medium text-text-primary">API Key</h4>
                    <p className="text-sm text-text-muted">
                      Use this key to authenticate API requests
                    </p>
                  </div>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRegenerateApiKey}
                  disabled={isRegenerating}
                >
                  <RefreshCw className={`w-4 h-4 mr-2 ${isRegenerating ? "animate-spin" : ""}`} />
                  Regenerate
                </Button>
              </div>

              <p className="text-xs text-text-muted">
                Regenerating your API key will invalidate the previous key. Make sure to update
                your applications accordingly.
              </p>
            </CardContent>
          </Card>

          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">CORS Configuration</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label>Enable CORS</Label>
                  <p className="text-sm text-text-muted">
                    Allow cross-origin requests to your function
                  </p>
                </div>
                <Switch checked={corsEnabled} onCheckedChange={setCorsEnabled} />
              </div>

              {corsEnabled && (
                <div className="space-y-2">
                  <Label htmlFor="cors-origins">Allowed Origins</Label>
                  <Input
                    id="cors-origins"
                    value={corsOrigins}
                    onChange={(e) => setCorsOrigins(e.target.value)}
                    placeholder="https://example.com, https://app.example.com"
                  />
                  <p className="text-xs text-text-muted">
                    Comma-separated list of allowed origins. Use * to allow all origins.
                  </p>
                </div>
              )}
            </CardContent>
          </Card>

          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Custom Domains</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <EnterpriseFeature
                feature="CUSTOM_DOMAINS"
                fallback="upgrade"
                upgradeMessage="Custom domains are available on Starter, Professional, and Enterprise plans. Upgrade to connect your own domains."
              >
                {hasFeature("CUSTOM_DOMAINS") && (
                  <>
                    <p className="text-text-muted text-sm">
                      Connect custom domains to this function. Plan limit:{" "}
                      <span className="text-text-primary font-medium">
                        {formatCustomDomainsRemaining(customDomainsList.length, plan)}
                      </span>
                    </p>
                    {customDomainsList.length > 0 ? (
                      <div className="space-y-2">
                        {customDomainsList.map((domain) => (
                          <div
                            key={domain}
                            className="flex items-center justify-between p-3 rounded-lg bg-bg-tertiary"
                          >
                            <div className="flex items-center gap-2">
                              <Globe className="w-4 h-4 text-text-muted shrink-0" />
                              <span className="text-text-primary">{domain}</span>
                            </div>
                            <div className="flex items-center gap-2">
                              <Badge variant="secondary">Active</Badge>
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-text-muted hover:text-destructive"
                                onClick={() => handleRemoveDomain(domain)}
                                aria-label={`Remove ${domain}`}
                              >
                                <X className="w-4 h-4" />
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="text-text-muted text-sm">
                        No custom domains configured. Add a domain to serve this function from your own hostname.
                      </p>
                    )}

                    <Button
                      variant="outline"
                      className="w-full"
                      disabled={!canAddMore}
                      onClick={() => {
                        setNewDomainValue("");
                        setAddDomainError(null);
                        setShowAddDomainDialog(true);
                      }}
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      Add Custom Domain
                      {!canAddMore && customDomainsLimit > 0 && (
                        <span className="ml-2 text-xs">(limit reached)</span>
                      )}
                    </Button>
                  </>
                )}
              </EnterpriseFeature>
            </CardContent>
          </Card>

          <Dialog open={showAddDomainDialog} onOpenChange={setShowAddDomainDialog}>
            <DialogContent className="sm:max-w-md">
              <DialogHeader>
                <DialogTitle>Add Custom Domain</DialogTitle>
                <DialogDescription>
                  Enter the hostname (e.g. api.example.com). Do not include https://.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-2">
                <Input
                  placeholder="api.example.com"
                  value={newDomainValue}
                  onChange={(e) => {
                    setNewDomainValue(e.target.value);
                    setAddDomainError(null);
                  }}
                  onKeyDown={(e) => e.key === "Enter" && handleAddDomain()}
                />
                {addDomainError && (
                  <p className="text-sm text-destructive">{addDomainError}</p>
                )}
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setShowAddDomainDialog(false)}>
                  Cancel
                </Button>
                <Button onClick={handleAddDomain}>Add domain</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </TabsContent>

        {/* Providers Settings */}
        <TabsContent value="providers" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Deployment Providers</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <p className="text-sm text-text-muted">
                Select which providers to deploy your function to. You can configure
                provider-specific settings for each selected provider.
              </p>

              <div className="space-y-3">
                {[
                  { id: "workers", name: "Cloudflare Workers", color: "#f48120" },
                  { id: "vercel", name: "Vercel", color: "#000000" },
                  { id: "fly", name: "Fly.io", color: "#7b68ee" },
                  { id: "deno", name: "Deno Deploy", color: "#000000" },
                ].map((provider) => (
                  <div
                    key={provider.id}
                    className={`flex items-center justify-between p-4 rounded-lg border transition-colors ${
                      selectedProviders.includes(provider.id)
                        ? "border-primary bg-primary/5"
                        : "border-border bg-bg-secondary"
                    }`}
                  >
                    <div className="flex items-center gap-3">
                      <input
                        type="checkbox"
                        checked={selectedProviders.includes(provider.id)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setSelectedProviders([...selectedProviders, provider.id]);
                          } else {
                            setSelectedProviders(
                              selectedProviders.filter((p) => p !== provider.id)
                            );
                          }
                        }}
                        className="w-4 h-4 rounded border-border"
                      />
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: provider.color }}
                      />
                      <span className="text-text-primary">{provider.name}</span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Environment Variables */}
        <TabsContent value="environment" className="space-y-6">
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Environment Variables</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <p className="text-sm text-text-muted">
                Environment variables are key-value pairs that are available to your
                function at runtime. Secrets are encrypted and masked in logs.
              </p>

              <div className="space-y-3">
                {envVars.map((envVar, index) => (
                  <div key={index} className="flex items-center gap-3">
                    <Input
                      value={envVar.key}
                      onChange={(e) => {
                        const newEnvVars = [...envVars];
                        newEnvVars[index].key = e.target.value;
                        setEnvVars(newEnvVars);
                      }}
                      placeholder="KEY"
                      className="flex-1"
                    />
                    <Input
                      value={envVar.value}
                      onChange={(e) => {
                        const newEnvVars = [...envVars];
                        newEnvVars[index].value = e.target.value;
                        setEnvVars(newEnvVars);
                      }}
                      placeholder="Value"
                      type="password"
                      className="flex-1"
                    />
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => removeEnvironmentVariable(index)}
                    >
                      <Trash2 className="w-4 h-4 text-red-400" />
                    </Button>
                  </div>
                ))}

                <div className="flex items-center gap-3">
                  <Input
                    value={newEnvKey}
                    onChange={(e) => setNewEnvKey(e.target.value)}
                    placeholder="New Key"
                    className="flex-1"
                  />
                  <Input
                    value={newEnvValue}
                    onChange={(e) => setNewEnvValue(e.target.value)}
                    placeholder="Value"
                    type="password"
                    className="flex-1"
                  />
                  <Button variant="outline" onClick={addEnvironmentVariable}>
                    Add
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Embed Settings */}
        <TabsContent value="embed" className="space-y-6">
          <EmbedTab author={author || ""} name={name || ""} />
        </TabsContent>

        {/* Danger Zone */}
        <TabsContent value="danger" className="space-y-6">
          <Card className="card border-red-500/20">
            <CardHeader className="card-header">
              <CardTitle className="card-title text-red-400">Danger Zone</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div className="flex items-center justify-between p-4 rounded-lg bg-red-500/5 border border-red-500/20">
                <div>
                  <h4 className="font-medium text-text-primary">Delete Function</h4>
                  <p className="text-sm text-text-muted">
                    Permanently delete this function and all its data
                  </p>
                </div>
                <Button variant="destructive" onClick={handleDelete}>
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Function</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{functionSettings.name}"? This action
              cannot be undone and all deployments will be lost.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteDialog(false)}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={isDeleting}>
              {isDeleting ? "Deleting..." : "Delete Function"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* API Key Dialog */}
      <Dialog open={showApiKeyDialog} onOpenChange={setShowApiKeyDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New API Key</DialogTitle>
            <DialogDescription>
              Your new API key has been generated. Copy it now as you won't be able
              to see it again.
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2">
            <Input value={apiKey} readOnly type="password" />
            <Button variant="outline" onClick={copyApiKey}>
              <Copy className="w-4 h-4" />
            </Button>
          </div>
          <DialogFooter>
            <Button onClick={() => setShowApiKeyDialog(false)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default FunctionSettingsPage;
