import { useState, useRef, useEffect } from "react";
import { useNavigate, useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import Editor from "@monaco-editor/react";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Play,
  Rocket,
  Save,
  ArrowLeft,
  Plus,
  X,
  Eye,
  EyeOff,
  Copy,
  CheckCircle2,
  AlertCircle,
  Terminal,
  Link2,
  Loader2,
  Shield,
  Key,
} from "lucide-react";
import { ProviderIcon } from "@/components/common/ProviderIcon";
import { StatusBadge } from "@/components/common/StatusBadge";
import { functionsApi } from "@/api";
import { apiClient } from "@/api/client";
import { providersApi } from "@/api/providers";
import { vaultApi } from "@/api/vault";
import { useVaultSecrets } from "@/hooks/useVault";
import { SecretRevealGate } from "@/components/SecretsVault";
import { FunctionConfig, TestFunctionRequest } from "@/types";
import type { SecretMetadata } from "@/types/vault";
import "@/styles/components.css";

interface EnvironmentVariable {
  id: string;
  key: string;
  value: string;
  isSecret: boolean;
}

interface DeploymentLog {
  id: string;
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'success';
  message: string;
}

const defaultCode = `export default {
  async fetch(request, env, ctx) {
    // Your function code here
    return new Response('Hello World!', {
      headers: { 'content-type': 'text/plain' },
    });
  }
};`;

export function FunctionEditorPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEditing = !!id;

  // Fetch real providers from API
  const { data: connectedProviders } = useQuery({
    queryKey: ["providers"],
    queryFn: () => providersApi.getConnectedProviders(),
  });

  // Map connected providers to the format expected by the editor
  const providers = (connectedProviders ?? []).map((p) => ({
    id: (p as { provider_type?: string; id: string }).provider_type || p.id,
    name: p.name,
    regions: (p as { region?: string }).region ? [(p as { region: string }).region] : ["global"],
  }));

  const { data: vaultSecrets } = useVaultSecrets();

  const [functionName, setFunctionName] = useState("");
  const [selectedProviders, setSelectedProviders] = useState<string[]>([]);
  const [selectedRegion, setSelectedRegion] = useState("");
  const [code, setCode] = useState(defaultCode);
  const [envVars, setEnvVars] = useState<EnvironmentVariable[]>([]);
  const [newEnvKey, setNewEnvKey] = useState("");
  const [newEnvValue, setNewEnvValue] = useState("");
  const [isNewEnvSecret, setIsNewEnvSecret] = useState(false);
  const [isDeploying, setIsDeploying] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [activeTab, setActiveTab] = useState("editor");
  const [currentDeploymentId, setCurrentDeploymentId] = useState<string | null>(null);
  const [deploymentStatus, setDeploymentStatus] = useState<string | null>(null);
  const [vaultPickerOpen, setVaultPickerOpen] = useState(false);
  const [pickingSecretId, setPickingSecretId] = useState<string | null>(null);
  const [revealEnvVarId, setRevealEnvVarId] = useState<string | null>(null);
  const [revealGateOpen, setRevealGateOpen] = useState(false);
  const [logs, setLogs] = useState<DeploymentLog[]>([
    {
      id: "1",
      timestamp: "2024-01-15 10:30:15",
      level: "info",
      message: isEditing ? "Editor initialized for editing" : "Editor initialized"
    }
  ]);

  // Document title
  useEffect(() => {
    const base = "FunctionFly";
    if (isEditing && functionName) {
      document.title = `Edit ${functionName} | ${base}`;
    } else {
      document.title = functionName ? `New: ${functionName} | ${base}` : `New Function | ${base}`;
    }
    return () => {
      document.title = base;
    };
  }, [isEditing, functionName]);

  // Auto-select region when only one provider/region is available
  useEffect(() => {
    if (selectedProviders.length === 1 && !selectedRegion) {
      const provider = providers.find((p) => p.id === selectedProviders[0]);
      if (provider?.regions?.length === 1) {
        setSelectedRegion(provider.regions[0]);
      }
    }
  }, [selectedProviders, providers, selectedRegion]);

  // Load function data if editing
  useEffect(() => {
    if (isEditing && id) {
      // Fetch function data from API
      const fetchFunctionData = async () => {
        try {
          addLog("info", `Loading function data for ${id}...`);

          const functionData = await functionsApi.get(id);

          setFunctionName(functionData.name);
          setSelectedProviders(functionData.providers);
          setSelectedRegion(functionData.region);
          setCode(functionData.code);
          setEnvVars(functionData.envVars.map((env, index) => ({
            id: `env-${index + 1}`,
            ...env
          })));

          addLog("success", `Function data loaded successfully`);
        } catch (error) {
          addLog("error", `Failed to load function data: ${error}`);
        }
      };

      fetchFunctionData();
    }
  }, [isEditing, id]);

  // Poll for deployment status updates
  useEffect(() => {
    if (!currentDeploymentId) return;

    const pollDeploymentStatus = async () => {
      try {
        const data = await apiClient.get<{ deployment: { status: string } }>(`/v1/functions/deployments/${currentDeploymentId}`);
        const status = data.deployment?.status;

        if (status !== deploymentStatus) {
          setDeploymentStatus(status);
          addLog("info", `Deployment status: ${status}`);

          if (status === "success") {
            addLog("success", "Deployment completed successfully!");
            setCurrentDeploymentId(null);
          } else if (status === "failed") {
            addLog("error", "Deployment failed");
            setCurrentDeploymentId(null);
          }
        }
      } catch (error) {
        console.error("Failed to poll deployment status:", error);
      }
    };

    // Poll immediately and then every 5 seconds
    pollDeploymentStatus();
    const interval = setInterval(pollDeploymentStatus, 5000);

    return () => clearInterval(interval);
  }, [currentDeploymentId, deploymentStatus]);

  const addLog = (level: DeploymentLog['level'], message: string) => {
    const newLog: DeploymentLog = {
      id: Date.now().toString(),
      timestamp: new Date().toLocaleString(),
      level,
      message
    };
    setLogs(prev => [...prev, newLog]);
  };

  const handleProviderToggle = (providerId: string) => {
    setSelectedProviders(prev =>
      prev.includes(providerId)
        ? prev.filter(id => id !== providerId)
        : [...prev, providerId]
    );
  };

  const addEnvironmentVariable = () => {
    if (!newEnvKey.trim() || !newEnvValue.trim()) return;

    const newVar: EnvironmentVariable = {
      id: Date.now().toString(),
      key: newEnvKey.trim(),
      value: newEnvValue.trim(),
      isSecret: isNewEnvSecret
    };

    setEnvVars(prev => [...prev, newVar]);
    setNewEnvKey("");
    setNewEnvValue("");
    setIsNewEnvSecret(false);
    addLog("info", `Added environment variable: ${newVar.key}`);
  };

  const removeEnvironmentVariable = (id: string) => {
    setEnvVars(prev => {
      const removed = prev.find(v => v.id === id);
      if (removed) {
        addLog("info", `Removed environment variable: ${removed.key}`);
      }
      return prev.filter(v => v.id !== id);
    });
  };

  const handleSelectVaultSecret = async (secret: SecretMetadata) => {
    setPickingSecretId(secret.id);
    try {
      const { value } = await vaultApi.decryptSecret(secret.id);
      setNewEnvKey(secret.name);
      setNewEnvValue(value);
      setIsNewEnvSecret(true);
      setVaultPickerOpen(false);
      addLog("info", `Using secret from Vault: ${secret.name}`);
      toast.success(`Added "${secret.name}" from Vault`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Decrypt failed";
      addLog("error", `Vault decrypt failed: ${msg}`);
      toast.error("Could not use this secret here. Try adding it manually or reveal in Vault first.");
    } finally {
      setPickingSecretId(null);
    }
  };

  const handleRevealVerified = (envVar: EnvironmentVariable) => {
    try {
      navigator.clipboard.writeText(envVar.value);
      toast.success("Value copied to clipboard");
    } catch {
      toast.error("Could not copy to clipboard");
    }
    setRevealGateOpen(false);
    setRevealEnvVarId(null);
  };

  const handleOpenReveal = (envVarId: string) => {
    setRevealEnvVarId(envVarId);
    setRevealGateOpen(true);
  };

  const handleDeploy = async () => {
    if (!functionName.trim() || selectedProviders.length === 0) {
      addLog("error", "Function name and at least one provider are required");
      toast.error("Function name and at least one provider are required");
      return;
    }

    if (!selectedRegion?.trim()) {
      addLog("error", "Please select a region");
      toast.error("Please select a region");
      return;
    }

    setIsDeploying(true);
    addLog("info", "Starting deployment...");

    try {
      let functionId = id ?? null;

      // New function: create first so we have a functionId
      if (!isEditing || !id) {
        addLog("info", "Creating function...");
        const created = await functionsApi.create({
          name: functionName.trim(),
          providers: selectedProviders,
          region: selectedRegion,
          code,
          envVars: envVars.map(({ key, value, isSecret }) => ({ key, value, isSecret })),
        });
        functionId = created.id;
        addLog("success", `Function created: ${created.name}`);
        toast.success("Function created");
        // Redirect to edit page so URL has the new id; user can deploy from there
        navigate(`/functions/${functionId}/edit`, { replace: true, state: { justCreated: true } });
        setIsDeploying(false);
        return;
      }

      // Existing function: deploy (backend expects functionId + backend_id)
      const result = await functionsApi.deploy({
        functionId,
        providers: selectedProviders,
        region: selectedRegion,
      } as any);
      setCurrentDeploymentId(result.deploymentId);
      setDeploymentStatus("pending");
      addLog("success", `Deployment started with ID: ${result.deploymentId}`);
      toast.success("Deployment started");
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      addLog("error", `Deployment failed: ${message}`);
      toast.error(`Deployment failed: ${message}`);
    } finally {
      setIsDeploying(false);
    }
  };

  const handleTest = async () => {
    setIsTesting(true);
    addLog("info", "Running function test...");

    try {
      const testData: TestFunctionRequest = {
        functionId: isEditing ? id : undefined,
        code: isEditing ? undefined : code,
        envVars,
        testInput: {} // Default test input
      };

      const result = await functionsApi.test(testData);

      if (result.success) {
        addLog("success", `Test completed in ${result.executionTimeMs}ms`);
        if (result.output) {
          addLog("info", `Output: ${JSON.stringify(result.output)}`);
        }
      } else {
        addLog("error", `Test failed: ${result.error}`);
      }

      // Log any additional logs from the test
      result.logs.forEach(log => {
        addLog(log.level as any, log.message);
      });

    } catch (error) {
      addLog("error", `Test failed: ${error}`);
    } finally {
      setIsTesting(false);
    }
  };

  const handleSaveDraft = async () => {
    if (!functionName.trim()) {
      addLog("error", "Function name is required to save a draft");
      toast.error("Function name is required");
      return;
    }
    if (selectedProviders.length === 0) {
      toast.error("Select at least one provider");
      return;
    }
    if (!selectedRegion?.trim()) {
      toast.error("Select a region");
      return;
    }
    try {
      addLog("info", "Saving draft...");
      if (isEditing && id) {
        await functionsApi.update(id, {
          name: functionName,
          code,
          providers: selectedProviders,
          region: selectedRegion,
          envVars: envVars.map(({ key, value, isSecret }) => ({ key, value, isSecret })),
        });
        toast.success("Draft saved");
      } else {
        const created = await functionsApi.create({
          name: functionName,
          code,
          providers: selectedProviders,
          region: selectedRegion,
          envVars: envVars.map(({ key, value, isSecret }) => ({ key, value, isSecret })),
        });
        toast.success("Function created");
        navigate(`/functions/${created.id}/edit`, { replace: true });
      }
      addLog("success", "Draft saved successfully");
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      addLog("error", `Failed to save draft: ${message}`);
      toast.error(`Failed to save draft: ${message}`);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate("/functions")}
            className="text-text-secondary hover:text-text-primary"
          >
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <div>
            <h1 className="text-2xl page-title">
              {isEditing ? `Edit ${functionName || "Function"}` : (functionName || "New Function")}
            </h1>
            <p className="text-sm page-title-subtle mt-0.5">
              {isEditing ? "Edit and redeploy your edge function" : "Create and deploy your edge function"}
            </p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleSaveDraft}
            disabled={isDeploying || isTesting}
          >
            <Save className="w-4 h-4 mr-2" />
            Save Draft
          </Button>
          <Button
            variant="outline"
            onClick={handleTest}
            disabled={isDeploying || isTesting}
          >
            <Play className="w-4 h-4 mr-2" />
            Test
          </Button>
          <Button
            onClick={handleDeploy}
            disabled={isDeploying || isTesting}
            className="gap-2"
          >
            <Rocket className="w-4 h-4" />
            {isDeploying ? "Deploying..." : "Deploy"}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left Column - Form */}
        <div className="space-y-6">
          {/* Basic Configuration */}
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Configuration</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              <div>
                <Label htmlFor="function-name">Function Name</Label>
                <Input
                  id="function-name"
                  placeholder="my-function"
                  value={functionName}
                  onChange={(e) => setFunctionName(e.target.value)}
                  className="input mt-1"
                />
              </div>

              <div>
                <Label>Providers</Label>
                {providers.length === 0 ? (
                  <div className="mt-2 p-4 rounded-lg border border-border-subtle bg-bg-tertiary/50">
                    <p className="text-sm text-text-secondary mb-3">
                      Connect a provider to deploy your function. You need at least one provider (e.g. Cloudflare Workers, Vercel) to deploy.
                    </p>
                    <Button variant="outline" size="sm" asChild>
                      <Link to="/providers" className="gap-2">
                        <Link2 className="w-4 h-4" />
                        Connect a provider
                      </Link>
                    </Button>
                  </div>
                ) : (
                  <div className="flex flex-wrap gap-2 mt-2">
                    {providers.map((provider) => (
                      <button
                        key={provider.id}
                        onClick={() => handleProviderToggle(provider.id)}
                        className={`flex items-center gap-2 px-3 py-2 rounded-lg border transition-colors ${
                          selectedProviders.includes(provider.id)
                            ? "bg-indigo-500/10 border-indigo-500/30 text-indigo-400"
                            : "bg-bg-tertiary border-border-subtle text-text-secondary hover:border-border-default"
                        }`}
                      >
                        <ProviderIcon provider={provider.id} size="sm" />
                        <span className="text-sm">{provider.name}</span>
                      </button>
                    ))}
                  </div>
                )}
              </div>

              {selectedProviders.length > 0 && (
                <div>
                  <Label htmlFor="region">Region</Label>
                  <Select value={selectedRegion} onValueChange={setSelectedRegion}>
                    <SelectTrigger className="select mt-1">
                      <SelectValue placeholder="Select a region" />
                    </SelectTrigger>
                    <SelectContent>
                      {providers
                        .filter(p => selectedProviders.includes(p.id))
                        .flatMap(p => p.regions)
                        .filter((region, index, arr) => arr.indexOf(region) === index)
                        .map((region) => (
                          <SelectItem key={region} value={region}>
                            {region.toUpperCase()}
                          </SelectItem>
                        ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Environment Variables */}
          <Card className="card">
            <CardHeader className="card-header">
              <CardTitle className="card-title">Environment Variables</CardTitle>
            </CardHeader>
            <CardContent className="card-content space-y-4">
              {envVars.length > 0 && (
                <div className="space-y-2">
                  {envVars.map((envVar) => (
                    <div key={envVar.id} className="flex items-center gap-3 p-3 rounded-lg bg-bg-tertiary">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <code className="text-sm font-medium text-text-primary">{envVar.key}</code>
                          {envVar.isSecret && <EyeOff className="w-3 h-3 text-text-muted shrink-0" />}
                        </div>
                        <div className="text-sm text-text-secondary mt-1 truncate">
                          {envVar.isSecret ? "••••••••" : envVar.value}
                        </div>
                      </div>
                      <div className="flex items-center gap-1 shrink-0">
                        {envVar.isSecret && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="text-text-muted hover:text-text-primary"
                            onClick={() => handleOpenReveal(envVar.id)}
                          >
                            <Eye className="w-4 h-4" />
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => removeEnvironmentVariable(envVar.id)}
                          className="text-text-muted hover:text-red-400"
                        >
                          <X className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              <Separator />

              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <Label htmlFor="env-key">Key</Label>
                    <Input
                      id="env-key"
                      placeholder="API_KEY"
                      value={newEnvKey}
                      onChange={(e) => setNewEnvKey(e.target.value)}
                      className="input mt-1"
                    />
                  </div>
                  <div>
                    <Label htmlFor="env-value">Value</Label>
                    <Input
                      id="env-value"
                      type={isNewEnvSecret ? "password" : "text"}
                      placeholder="your-api-key"
                      value={newEnvValue}
                      onChange={(e) => setNewEnvValue(e.target.value)}
                      className="input mt-1"
                    />
                  </div>
                </div>

                <div className="flex flex-wrap items-center gap-3">
                  <div className="flex items-center gap-2">
                    <Checkbox
                      id="isSecret"
                      checked={isNewEnvSecret}
                      onCheckedChange={(checked) => setIsNewEnvSecret(checked === true)}
                    />
                    <Label htmlFor="isSecret" className="text-sm cursor-pointer">
                      Mark as secret
                    </Label>
                  </div>
                  <Button
                    size="sm"
                    onClick={addEnvironmentVariable}
                    disabled={!newEnvKey.trim() || !newEnvValue.trim()}
                  >
                    <Plus className="w-3 h-3 mr-1" />
                    Add
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setVaultPickerOpen(true)}
                    className="gap-1.5"
                  >
                    <Shield className="w-3.5 h-3.5" />
                    Use from Vault
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Vault secret picker dialog */}
          <Dialog open={vaultPickerOpen} onOpenChange={setVaultPickerOpen}>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                  <Key className="w-5 h-5" />
                  Use secret from Vault
                </DialogTitle>
                <DialogDescription>
                  Choose a secret to add as an environment variable. Its value will be filled in and marked as secret.
                </DialogDescription>
              </DialogHeader>
              <ScrollArea className="max-h-[280px] rounded-md border border-border-subtle">
                <div className="p-2 space-y-1">
                  {!vaultSecrets?.length && (
                    <p className="text-sm text-text-secondary py-4 text-center">No secrets in Vault yet.</p>
                  )}
                  {vaultSecrets?.map((secret) => (
                    <button
                      key={secret.id}
                      type="button"
                      onClick={() => handleSelectVaultSecret(secret)}
                      disabled={pickingSecretId !== null}
                      className="w-full flex items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-bg-hover disabled:opacity-50"
                    >
                      <Shield className="w-4 h-4 text-text-muted shrink-0" />
                      <div className="flex-1 min-w-0">
                        <span className="text-sm font-medium text-text-primary block truncate">{secret.name}</span>
                        {secret.description && (
                          <span className="text-xs text-text-secondary block truncate">{secret.description}</span>
                        )}
                      </div>
                      {pickingSecretId === secret.id ? (
                        <Loader2 className="w-4 h-4 animate-spin text-text-muted shrink-0" />
                      ) : null}
                    </button>
                  ))}
                </div>
              </ScrollArea>
            </DialogContent>
          </Dialog>

          {/* Reveal secret env var: SecretRevealGate (controlled, no visible trigger) */}
          {revealEnvVarId && (() => {
            const envVar = envVars.find((e) => e.id === revealEnvVarId);
            if (!envVar || !envVar.isSecret) return null;
            return (
              <SecretRevealGate
                trigger={<span className="hidden" aria-hidden />}
                isOpen={revealGateOpen}
                onOpenChange={(open) => {
                  setRevealGateOpen(open);
                  if (!open) setRevealEnvVarId(null);
                }}
                onVerified={() => handleRevealVerified(envVar)}
                onCancelled={() => {
                  setRevealGateOpen(false);
                  setRevealEnvVarId(null);
                }}
                title="Reveal secret value"
                description={`Verify your identity to reveal the value of ${envVar.key}. It will be copied to your clipboard.`}
                requiredLevel="medium"
              />
            );
          })()}
        </div>

        {/* Right Column - Editor & Preview */}
        <div className="space-y-6">
          <Card className="card h-[600px] flex flex-col overflow-hidden">
            <CardHeader className="card-header shrink-0">
              <div className="flex items-center justify-between">
                <CardTitle className="card-title">Code Editor</CardTitle>
                <div className="flex gap-2">
                  <Button variant="ghost" size="sm">
                    <Copy className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent className="card-content p-0 flex-1 flex flex-col min-h-0">
              <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
                <TabsList className="grid w-full grid-cols-2 rounded-none border-b border-border-subtle shrink-0">
                  <TabsTrigger value="editor" className="rounded-none">Editor</TabsTrigger>
                  <TabsTrigger value="logs" className="rounded-none">Logs</TabsTrigger>
                </TabsList>

                <TabsContent value="editor" className="mt-0 flex-1 min-h-0 data-[state=inactive]:hidden">
                  <div className="w-full min-h-[480px]" style={{ height: "100%" }}>
                    {activeTab === "editor" && (
                      <Editor
                        height="100%"
                        defaultLanguage="javascript"
                        value={code}
                        onChange={(value) => setCode(value || "")}
                        theme="vs-dark"
                        loading={
                          <div className="flex items-center justify-center w-full min-h-[400px] bg-[#1e1e1e] text-text-secondary rounded-b-lg">
                            <Loader2 className="w-8 h-8 animate-spin" />
                          </div>
                        }
                        options={{
                          minimap: { enabled: false },
                          fontSize: 14,
                          lineNumbers: "on",
                          roundedSelection: false,
                          scrollBeyondLastLine: false,
                          automaticLayout: true,
                          tabSize: 2,
                          wordWrap: "on",
                        }}
                      />
                    )}
                  </div>
                </TabsContent>

                <TabsContent value="logs" className="mt-0 flex-1 min-h-0 overflow-auto">
                  <ScrollArea className="h-full p-4">
                    <div className="space-y-2">
                      {logs.map((log) => (
                        <div key={log.id} className="flex items-start gap-3 text-sm">
                          <div className="text-text-muted font-mono text-xs w-20">
                            {log.timestamp.split(' ')[1]}
                          </div>
                          <div className="flex items-center gap-2 flex-1">
                            {log.level === 'success' && <CheckCircle2 className="w-4 h-4 text-green-400" />}
                            {log.level === 'error' && <AlertCircle className="w-4 h-4 text-red-400" />}
                            {log.level === 'warn' && <AlertCircle className="w-4 h-4 text-yellow-400" />}
                            {log.level === 'info' && <Terminal className="w-4 h-4 text-blue-400" />}
                            <span className="font-mono">{log.message}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </ScrollArea>
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
