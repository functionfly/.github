import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { ArrowLeft, Rocket, Settings, Play, Code, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { useRealtime } from "@/hooks/useRealtime";
import { registryApi, RegistryFunction, RegistryFunctionVersion } from "@/api/registry";
import { apiClient } from "@/api/client";

interface App {
  id: string;
  name: string;
  slug: string;
}

interface Backend {
  id: string;
  provider: string;
  region: string;
  url: string;
}

export default function RegistryDeployPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { subscribe, unsubscribe } = useRealtime();

  const registryPath = searchParams.get("registry");
  const [author, name] = registryPath ? registryPath.split("/") : ["", ""];

  const [registryFunction, setRegistryFunction] = useState<RegistryFunction | null>(null);
  const [versions, setVersions] = useState<RegistryFunctionVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<string>("latest");
  const [currentVersion, setCurrentVersion] = useState<RegistryFunctionVersion | null>(null);

  const [apps, setApps] = useState<App[]>([]);
  const [appsLoading, setAppsLoading] = useState(false);
  const [selectedApp, setSelectedApp] = useState<string>("");

  const [loading, setLoading] = useState(true);
  const [deploying, setDeploying] = useState(false);
  const [testing, setTesting] = useState(false);

  // Deployment configuration
  const [functionName, setFunctionName] = useState("");
  const [route, setRoute] = useState("");
  const [envVars, setEnvVars] = useState<Record<string, string>>({});
  const [secrets, setSecrets] = useState<Record<string, string>>({});

  // Test execution
  const [testInput, setTestInput] = useState("{}");
  const [testOutput, setTestOutput] = useState<string>("");
  const [testError, setTestError] = useState<string>("");

  // Real-time deployment updates
  useEffect(() => {
    const handleDeploymentUpdate = (data: any) => {
      if (data.event === "deployment_update" && data.app_id === selectedApp) {
        // Handle deployment status updates
        console.log("Deployment update:", data);
      }
    };

    subscribe("deployments", handleDeploymentUpdate);

    return () => {
      unsubscribe("deployments", handleDeploymentUpdate);
    };
  }, [subscribe, unsubscribe, selectedApp]);

  // Load function details
  useEffect(() => {
    if (author && name) {
      loadFunctionDetails();
      loadUserApps();
    }
  }, [author, name]);

  // Update current version when selection changes
  useEffect(() => {
    const list = Array.isArray(versions) ? versions : [];
    if (selectedVersion === "latest" && list.length > 0) {
      setCurrentVersion(list[0]);
    } else {
      const version = list.find(v => v.version === selectedVersion);
      setCurrentVersion(version || null);
    }
  }, [selectedVersion, versions]);

  const loadFunctionDetails = async () => {
    try {
      setLoading(true);
      const [functionData, versionsData] = await Promise.all([
        registryApi.getFunction(author, name),
        registryApi.getFunctionVersions(author, name),
      ]);

      // Backend returns the function object directly (no .function wrapper)
      const raw = functionData as unknown as { function?: RegistryFunction } | RegistryFunction;
      const fn = "function" in raw && raw.function ? raw.function : (functionData as unknown as RegistryFunction);
      setRegistryFunction(fn);

      // Backend may return { versions: [...] } or the raw versions array
      const vers = Array.isArray(versionsData)
        ? versionsData
        : (versionsData as { versions?: RegistryFunctionVersion[] })?.versions ?? [];
      setVersions(vers);
      setFunctionName(name); // Default function name
    } catch (error) {
      console.error("Failed to load function details:", error);
    } finally {
      setLoading(false);
    }
  };

  const loadUserApps = async () => {
    try {
      setAppsLoading(true);
      const response = await apiClient.get<{ apps?: App[] }>("/v1/apps");
      const list = Array.isArray(response?.apps) ? response.apps : [];
      setApps(list);
    } catch (error) {
      console.error("Failed to load apps:", error);
      setApps([]);
    } finally {
      setAppsLoading(false);
    }
  };

  const handleTestFunction = async () => {
    if (!currentVersion) return;

    try {
      setTesting(true);
      setTestError("");
      setTestOutput("");

      let input;
      try {
        input = JSON.parse(testInput);
      } catch (e) {
        setTestError("Invalid JSON input");
        return;
      }

      const result = await registryApi.testFunction(author, name, input);

      if (result.ok) {
        setTestOutput(JSON.stringify(result.output, null, 2));
      } else {
        setTestError("Function execution failed");
      }
    } catch (error: any) {
      setTestError(error.response?.data?.message || "Test failed");
    } finally {
      setTesting(false);
    }
  };

  const handleDeploy = async () => {
    if (!selectedApp || !functionName || !currentVersion) return;

    try {
      setDeploying(true);

      // Prepare deployment spec
      const deploySpec = {
        app_id: selectedApp,
        function_name: functionName,
        registry_function: `${author}/${name}`,
        version: selectedVersion,
        route: route,
        env_vars: envVars,
        secrets: secrets,
      };

      const response = await apiClient.post<{ function_id: string }>(`/v1/functions/deploy`, deploySpec);

      // Navigate to the function details page
      navigate(`/functions/${response.function_id}`);
    } catch (error: any) {
      console.error("Deployment failed:", error);
      setTestError(error.response?.data?.message || "Deployment failed");
    } finally {
      setDeploying(false);
    }
  };

  const addEnvVar = () => {
    setEnvVars({ ...envVars, "": "" });
  };

  const updateEnvVar = (oldKey: string, newKey: string, value: string) => {
    const updated = { ...envVars };
    delete updated[oldKey];
    updated[newKey] = value;
    setEnvVars(updated);
  };

  const removeEnvVar = (key: string) => {
    const updated = { ...envVars };
    delete updated[key];
    setEnvVars(updated);
  };

  const addSecret = () => {
    setSecrets({ ...secrets, "": "" });
  };

  const updateSecret = (oldKey: string, newKey: string, value: string) => {
    const updated = { ...secrets };
    delete updated[oldKey];
    updated[newKey] = value;
    setSecrets(updated);
  };

  const removeSecret = (key: string) => {
    const updated = { ...secrets };
    delete updated[key];
    setSecrets(updated);
  };

  if (loading) {
    return (
      <div className="container mx-auto px-6 py-8">
        <div className="mb-8">
          <Skeleton className="h-8 w-64 mb-2" />
          <Skeleton className="h-4 w-96" />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <Card>
              <CardHeader>
                <Skeleton className="h-6 w-48" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-32 w-full" />
              </CardContent>
            </Card>
          </div>

          <div className="space-y-6">
            <Card>
              <CardHeader>
                <Skeleton className="h-6 w-32" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-48 w-full" />
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    );
  }

  if (!registryFunction) {
    return (
      <div className="container mx-auto px-6 py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Function not found or not available for deployment. Check that{" "}
            <strong>{author}/{name}</strong> exists in the registry and try again.
          </AlertDescription>
        </Alert>
        <Button variant="outline" className="mt-4" onClick={() => navigate("/registry")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Registry
        </Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-6 py-8">
      {/* Header */}
      <div className="mb-8">
        <Button
          variant="ghost"
          onClick={() => navigate("/registry")}
          className="mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Registry
        </Button>

        <div className="flex items-center gap-3 mb-2">
          <Code className="h-8 w-8 text-muted-foreground" />
          <div>
            <h1 className="text-3xl font-bold">{registryFunction.author}/{registryFunction.name}</h1>
            <p className="text-muted-foreground">{registryFunction.title || registryFunction.description}</p>
          </div>
        </div>

        <div className="flex items-center gap-4">
          {registryFunction.category && (
            <Badge variant="secondary">{registryFunction.category}</Badge>
          )}
          {(registryFunction.overall_score != null || (registryFunction as { total_ratings?: number }).total_ratings != null) && (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <span>⭐ {(registryFunction.overall_score ?? 0).toFixed(1)}</span>
              <span>({(registryFunction as { total_ratings?: number }).total_ratings ?? 0} ratings)</span>
            </div>
          )}
          {(registryFunction as { popularity_score?: number }).popularity_score != null && (
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <span>📥 {Math.floor((registryFunction as { popularity_score?: number }).popularity_score ?? 0)} downloads</span>
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          <Tabs defaultValue="deploy" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="deploy">Deploy</TabsTrigger>
              <TabsTrigger value="test">Test Function</TabsTrigger>
            </TabsList>

            <TabsContent value="deploy" className="space-y-6">
              {/* Deployment Configuration */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Settings className="h-5 w-5" />
                    Deployment Configuration
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="app">Target App</Label>
                      {appsLoading ? (
                        <div className="flex h-10 items-center rounded-md border border-input bg-muted/50 px-3 py-2 text-sm text-muted-foreground">
                          Loading apps…
                        </div>
                      ) : apps.length === 0 ? (
                        <div className="rounded-md border border-dashed border-input bg-muted/30 p-4 text-sm">
                          <p className="text-muted-foreground mb-2">No apps yet.</p>
                          <p className="text-muted-foreground text-xs mb-3">
                            Create an app in your dashboard to deploy this function to.
                          </p>
                          <Button type="button" variant="outline" size="sm" onClick={() => navigate("/dashboard")}>
                            Go to Dashboard
                          </Button>
                        </div>
                      ) : (
                        <Select value={selectedApp} onValueChange={setSelectedApp}>
                          <SelectTrigger>
                            <SelectValue placeholder="Select an app" />
                          </SelectTrigger>
                          <SelectContent>
                            {apps.map((app) => (
                              <SelectItem key={app.id} value={app.id}>
                                {app.name} ({app.slug})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      )}
                    </div>

                    <div>
                      <Label htmlFor="version">Version</Label>
                      <Select value={selectedVersion} onValueChange={setSelectedVersion}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="latest">Latest ({registryFunction.latest_version})</SelectItem>
                          {versions.map((version) => (
                            <SelectItem key={version.id} value={version.version}>
                              {version.version}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div>
                    <Label htmlFor="functionName">Function Name</Label>
                    <Input
                      id="functionName"
                      value={functionName}
                      onChange={(e) => setFunctionName(e.target.value)}
                      placeholder="my-function"
                    />
                  </div>

                  <div>
                    <Label htmlFor="route">Route (optional)</Label>
                    <Input
                      id="route"
                      value={route}
                      onChange={(e) => setRoute(e.target.value)}
                      placeholder="/api/my-function"
                    />
                  </div>
                </CardContent>
              </Card>

              {/* Environment Variables */}
              <Card>
                <CardHeader>
                  <CardTitle>Environment Variables</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {Object.entries(envVars).map(([key, value], index) => (
                    <div key={index} className="flex gap-2">
                      <Input
                        placeholder="KEY"
                        value={key}
                        onChange={(e) => updateEnvVar(key, e.target.value, value)}
                      />
                      <Input
                        placeholder="VALUE"
                        value={value}
                        onChange={(e) => updateEnvVar(key, key, e.target.value)}
                      />
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => removeEnvVar(key)}
                      >
                        Remove
                      </Button>
                    </div>
                  ))}
                  <Button variant="outline" onClick={addEnvVar}>
                    Add Environment Variable
                  </Button>
                </CardContent>
              </Card>

              {/* Secrets */}
              <Card>
                <CardHeader>
                  <CardTitle>Secrets</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {Object.entries(secrets).map(([key, value], index) => (
                    <div key={index} className="flex gap-2">
                      <Input
                        placeholder="KEY"
                        value={key}
                        onChange={(e) => updateSecret(key, e.target.value, value)}
                      />
                      <Input
                        type="password"
                        placeholder="VALUE"
                        value={value}
                        onChange={(e) => updateSecret(key, key, e.target.value)}
                      />
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => removeSecret(key)}
                      >
                        Remove
                      </Button>
                    </div>
                  ))}
                  <Button variant="outline" onClick={addSecret}>
                    Add Secret
                  </Button>
                </CardContent>
              </Card>

              {testError && (
                <Alert>
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{testError}</AlertDescription>
                </Alert>
              )}

              <Button
                onClick={handleDeploy}
                disabled={deploying || !selectedApp || !functionName}
                className="w-full"
                size="lg"
              >
                <Rocket className="h-4 w-4 mr-2" />
                {deploying ? "Deploying..." : "Deploy Function"}
              </Button>
            </TabsContent>

            <TabsContent value="test" className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Play className="h-5 w-5" />
                    Test Function
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="testInput">Input (JSON)</Label>
                    <Textarea
                      id="testInput"
                      value={testInput}
                      onChange={(e) => setTestInput(e.target.value)}
                      placeholder='{"key": "value"}'
                      rows={4}
                    />
                  </div>

                  <Button
                    onClick={handleTestFunction}
                    disabled={testing}
                    className="w-full"
                  >
                    <Play className="h-4 w-4 mr-2" />
                    {testing ? "Testing..." : "Test Function"}
                  </Button>

                  {testOutput && (
                    <div>
                      <Label>Output</Label>
                      <pre className="bg-muted p-4 rounded-md text-sm overflow-x-auto">
                        {testOutput}
                      </pre>
                    </div>
                  )}

                  {testError && (
                    <Alert>
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription>{testError}</AlertDescription>
                    </Alert>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Function Info */}
          <Card>
            <CardHeader>
              <CardTitle>Function Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {currentVersion && (
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm font-medium">Runtime</span>
                    <Badge variant="outline">{currentVersion.runtime}</Badge>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm font-medium">Timeout</span>
                    <span className="text-sm text-muted-foreground">{currentVersion.timeout_ms}ms</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm font-medium">Memory</span>
                    <span className="text-sm text-muted-foreground">{currentVersion.memory_mb}MB</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm font-medium">Deterministic</span>
                    <Badge variant={currentVersion.deterministic ? "default" : "secondary"}>
                      {currentVersion.deterministic ? "Yes" : "No"}
                    </Badge>
                  </div>
                </div>
              )}

              {registryFunction.tags && registryFunction.tags.length > 0 && (
                <div>
                  <span className="text-sm font-medium">Tags</span>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {registryFunction.tags.map((tag, index) => (
                      <Badge key={index} variant="outline" className="text-xs">
                        {tag}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Stats */}
          <Card>
            <CardHeader>
              <CardTitle>Statistics</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between">
                <span className="text-sm">Rating</span>
                <div className="flex items-center gap-1">
                  <span className="font-medium">{(registryFunction.overall_score ?? 0).toFixed(1)}</span>
                  <span className="text-muted-foreground">({(registryFunction as { total_ratings?: number }).total_ratings ?? 0})</span>
                </div>
              </div>
              <div className="flex justify-between">
                <span className="text-sm">Downloads</span>
                <span className="font-medium">{Math.floor((registryFunction as { popularity_score?: number }).popularity_score ?? 0)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm">Reliability</span>
                <span className="font-medium">{((registryFunction as { reliability?: number; reliability_score?: number }).reliability ?? (registryFunction as { reliability_score?: number }).reliability_score ?? 0).toFixed(1)}%</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
