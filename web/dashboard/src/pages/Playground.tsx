import { useState, useEffect, useCallback } from "react";
import { useParams, useSearchParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Play,
  Copy,
  Check,
  Share2,
  Clock,
  AlertCircle,
  CheckCircle2,
  XCircle,
  History,
  Trash2,
  ExternalLink,
  Server,
  Globe,
} from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  playgroundApi,
  type PlaygroundInfo,
  type ExecuteResponse,
  type ExecutionHistoryItem,
} from "@/api/playground";

// Registry function info type (from the existing registry API)
interface RegistryFunctionInfo {
  author: string;
  name: string;
  version: string;
  title?: string;
  description?: string;
  runtime: string;
  manifest?: {
    input?: Record<string, unknown>;
  };
}

// Status badge component
function StatusBadge({ status }: { status: string }) {
  const variants: Record<string, "success" | "warning" | "error" | "secondary"> = {
    deployed: "success",
    online: "success",
    running: "success",
    pending: "warning",
    degraded: "warning",
    offline: "error",
    error: "error",
    draft: "secondary",
  };

  return (
    <Badge variant={variants[status.toLowerCase()] || "secondary"}>
      {status}
    </Badge>
  );
}

// JSON Syntax highlighting component
function JsonViewer({ data }: { data: unknown }) {
  const [copied, setCopied] = useState(false);
  const json = JSON.stringify(data, null, 2);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(json);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div className="relative group">
      <Button
        variant="ghost"
        size="sm"
        onClick={handleCopy}
        className="absolute top-2 right-2 h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity z-10"
      >
        {copied ? (
          <Check className="h-4 w-4 text-green-500" />
        ) : (
          <Copy className="h-4 w-4" />
        )}
      </Button>
      <pre className="bg-muted p-4 rounded-md overflow-x-auto text-sm max-h-[400px] overflow-y-auto">
        <code className="language-json">{json}</code>
      </pre>
    </div>
  );
}

// History item component
function HistoryItem({
  item,
  onClick,
  onDelete,
}: {
  item: ExecutionHistoryItem;
  onClick: () => void;
  onDelete: () => void;
}) {
  const timestamp = new Date(item.timestamp).toLocaleString();

  return (
    <div className="flex items-center gap-3 p-3 rounded-lg hover:bg-secondary/50 transition-colors group">
      <button
        onClick={onClick}
        className="flex-1 text-left min-w-0"
      >
        <div className="flex items-center gap-2">
          {item.success ? (
            <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" />
          ) : (
            <XCircle className="h-4 w-4 text-red-500 shrink-0" />
          )}
          <span className="text-sm font-mono truncate">
            {typeof item.input === "object"
              ? JSON.stringify(item.input).slice(0, 50)
              : String(item.input).slice(0, 50)}
          </span>
        </div>
        <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
          <Clock className="h-3 w-3" />
          {timestamp}
          <span>•</span>
          <span>{item.latency_ms}ms</span>
        </div>
      </button>
      <Button
        variant="ghost"
        size="sm"
        onClick={(e) => {
          e.stopPropagation();
          onDelete();
        }}
        className="h-8 w-8 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
      >
        <Trash2 className="h-4 w-4 text-muted-foreground hover:text-red-500" />
      </Button>
    </div>
  );
}

// Main Playground component
export function Playground() {
  const { author, name, appSlug, functionName } = useParams<{
    author?: string;
    name?: string;
    appSlug?: string;
    functionName?: string;
  }>();
  const [searchParams] = useSearchParams();

  // Determine if this is a registry function or user-deployed function
  const isRegistry = !!author && !!name;
  const slug = author || appSlug || "";
  const fnName = name || functionName || "";

  const [inputValue, setInputValue] = useState<string>("{}");
  const [isExecuting, setIsExecuting] = useState(false);
  const [executionResult, setExecutionResult] = useState<ExecuteResponse | null>(null);
  const [history, setHistory] = useState<ExecutionHistoryItem[]>([]);
  const [copied, setCopied] = useState(false);
  const [showHistory, setShowHistory] = useState(false);

  // Fetch user-deployed function info
  const { data: playgroundInfo, isLoading: isLoadingPlayground, error: playgroundError } = useQuery<PlaygroundInfo>({
    queryKey: ["playground-info", appSlug, functionName],
    queryFn: async () => {
      if (!appSlug || !functionName) throw new Error("Missing params");
      return playgroundApi.getInfo(appSlug, functionName);
    },
    enabled: !!appSlug && !!functionName,
    retry: false,
  });

  // Fetch registry function info
  const { data: registryInfo, isLoading: isLoadingRegistry, error: registryError } = useQuery({
    queryKey: ["registry-function", author, name],
    queryFn: async () => {
      if (!author || !name) throw new Error("Missing params");
      const response = await fetch(
        `/v1/registry/functions/${author}/${name}?expand=manifest`
      );
      if (response.status === 404) {
        throw new Error("Function not found");
      }
      if (!response.ok) {
        throw new Error("Failed to fetch function");
      }
      return response.json();
    },
    enabled: !!author && !!name,
    retry: false,
  });

  const isLoading = isLoadingPlayground || isLoadingRegistry;
  const error = playgroundError || registryError;
  const functionInfo = playgroundInfo || registryInfo;

  // Load history from localStorage
  useEffect(() => {
    if (slug && fnName) {
      const savedHistory = playgroundApi.getHistory(slug, fnName);
      setHistory(savedHistory);
    }
  }, [slug, fnName]);

  // Parse URL params for pre-filled input
  useEffect(() => {
    const urlInput = playgroundApi.parseUrlInput();
    if (urlInput !== null) {
      setInputValue(
        typeof urlInput === "string" ? urlInput : JSON.stringify(urlInput, null, 2)
      );
    }
  }, []);

  // Handle execution
  const handleExecute = useCallback(async () => {
    if (!slug || !fnName || !functionInfo) return;

    setIsExecuting(true);
    setExecutionResult(null);

    let parsedInput: unknown;
    try {
      parsedInput = JSON.parse(inputValue);
    } catch {
      parsedInput = inputValue;
    }

    try {
      let result: ExecuteResponse;

      if (isRegistry && registryInfo) {
        // Registry function - use registry API
        const response = await fetch(`/v1/fx/${slug}/${fnName}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-Playground': '1',
          },
          body: JSON.stringify(parsedInput),
        });
        const regResult = await response.json();
        result = {
          success: regResult.ok,
          output: regResult.data,
          error: regResult.error?.message,
          latency_ms: regResult.duration_ms,
          status_code: response.status,
        };

        // Save to history
        playgroundApi.saveToHistory(slug, fnName, {
          input: parsedInput,
          output: regResult.data,
          error: regResult.error?.message,
          latency_ms: regResult.duration_ms,
          status_code: response.status,
          success: regResult.ok,
        });
      } else if (playgroundInfo) {
        // User-deployed function - use playground API
        result = await playgroundApi.execute(slug, fnName, parsedInput);

        // Save to history
        playgroundApi.saveToHistory(slug, fnName, {
          input: parsedInput,
          output: result.output,
          error: result.error,
          latency_ms: result.latency_ms,
          status_code: result.status_code,
          success: result.success,
        });
      } else {
        throw new Error("Unknown function type");
      }

      setExecutionResult(result);

      // Refresh history
      setHistory(playgroundApi.getHistory(slug, fnName));
    } catch (err) {
      setExecutionResult({
        success: false,
        error: err instanceof Error ? err.message : "Failed to execute function",
        latency_ms: 0,
        status_code: 500,
      });

      // Save error to history
      playgroundApi.saveToHistory(slug, fnName, {
        input: parsedInput,
        error: err instanceof Error ? err.message : "Failed to execute function",
        latency_ms: 0,
        status_code: 500,
        success: false,
      });
      setHistory(playgroundApi.getHistory(slug, fnName));
    } finally {
      setIsExecuting(false);
    }
  }, [slug, fnName, functionInfo, isRegistry, registryInfo, playgroundInfo, inputValue]);

  // Copy shareable link
  const handleCopyShareableLink = async () => {
    if (!slug || !fnName) return;

    let parsedInput: unknown;
    try {
      parsedInput = JSON.parse(inputValue);
    } catch {
      parsedInput = inputValue;
    }

    const url = playgroundApi.createShareableUrl(slug, fnName, parsedInput);
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy link:", err);
    }
  };

  // Load history item
  const handleLoadFromHistory = (item: ExecutionHistoryItem) => {
    setInputValue(
      typeof item.input === "string"
        ? item.input
        : JSON.stringify(item.input, null, 2)
    );
    setExecutionResult({
      success: item.status_code >= 200 && item.status_code < 300,
      output: item.output,
      error: item.error,
      latency_ms: item.latency_ms,
      status_code: item.status_code,
    });
  };

  // Delete history item
  const handleDeleteHistory = (id: string) => {
    if (!slug || !fnName) return;
    const newHistory = history.filter((item) => item.id !== id);
    const key = `functionfly_playground_history_${slug}_${fnName}`;
    localStorage.setItem(key, JSON.stringify(newHistory));
    setHistory(newHistory);
  };

  // Clear all history
  const handleClearHistory = () => {
    if (!slug || !fnName) return;
    playgroundApi.clearHistory(slug, fnName);
    setHistory([]);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // Error state
  if (error || (!playgroundInfo && !registryInfo)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-red-500">
              <AlertCircle className="h-5 w-5" />
              Function Not Found
            </CardTitle>
            <CardDescription>
              The function {slug}/{fnName} could not be found or is not available for playground.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild>
              <Link to="/">Go to Home</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Check if playground is enabled (for user-deployed functions)
  if (playgroundInfo && !playgroundInfo.playground_enabled) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-amber-500">
              <AlertCircle className="h-5 w-5" />
              Playground Disabled
            </CardTitle>
            <CardDescription>
              The playground is not enabled for this function. Please contact the function owner.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild>
              <Link to="/">Go to Home</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Check if function is deployed (for user-deployed functions)
  const isDeployed = playgroundInfo
    ? playgroundInfo.status === "deployed" || playgroundInfo.status === "online"
    : true; // Registry functions are always considered deployed

  return (
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8 max-w-6xl">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-2 text-sm text-muted-foreground mb-4">
            {isRegistry ? (
              <Link
                to={`/fx/${slug}/${fnName}`}
                className="hover:underline flex items-center gap-1"
              >
                <Globe className="h-3 w-3" />
                {slug}/{fnName}
              </Link>
            ) : (
              <Link
                to={`/functions`}
                className="hover:underline flex items-center gap-1"
              >
                <Globe className="h-3 w-3" />
                {slug}
              </Link>
            )}
            <span>/</span>
            <span className="font-mono">{fnName}</span>
            <span>/</span>
            <span>Playground</span>
            <Badge variant="secondary" className="ml-2">
              v{isRegistry ? (registryInfo as any)?.version : playgroundInfo?.version}
            </Badge>
          </div>

          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-3xl font-bold mb-2">
                {String(
                  isRegistry
                    ? (registryInfo as any)?.title
                    : playgroundInfo?.playground_config?.title
                ) || fnName}
              </h1>
              {isRegistry ? (
                (registryInfo as any)?.description && (
                  <p className="text-muted-foreground">
                    {(registryInfo as any).description}
                  </p>
                )
              ) : (
                playgroundInfo?.playground_config?.description && (
                  <p className="text-muted-foreground">
                    {String(playgroundInfo.playground_config.description)}
                  </p>
                )
              )}
            </div>

            <div className="flex items-center gap-3">
              {isRegistry ? (
                <Badge variant="secondary">Registry</Badge>
              ) : (
                <>
                  <StatusBadge status={playgroundInfo?.status || "unknown"} />
                  {playgroundInfo?.provider && (
                    <Badge variant="outline" className="flex items-center gap-1">
                      <Server className="h-3 w-3" />
                      {playgroundInfo.provider}
                    </Badge>
                  )}
                  {playgroundInfo?.region && (
                    <span className="text-xs text-muted-foreground">
                      {playgroundInfo.region}
                    </span>
                  )}
                </>
              )}
            </div>
          </div>
        </div>

        {/* Not deployed warning */}
        {!isDeployed && (
          <div className="mb-6 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
            <div className="flex items-center gap-2 text-amber-500">
              <AlertCircle className="h-5 w-5" />
              <span className="font-medium">Function not deployed</span>
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              This function is not currently deployed. You can still test it, but execution may fail.
            </p>
          </div>
        )}

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Main content - Input and Output */}
          <div className="lg:col-span-2 space-y-6">
            <Tabs defaultValue="input" className="w-full">
              <TabsList className="w-full justify-start">
                <TabsTrigger value="input">Input</TabsTrigger>
                <TabsTrigger value="output">Output</TabsTrigger>
              </TabsList>

              <TabsContent value="input" className="mt-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">Request Body</CardTitle>
                    <CardDescription>
                      Enter your request as JSON. The function will receive this as input.
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Textarea
                      value={inputValue}
                      onChange={(e) => setInputValue(e.target.value)}
                      placeholder='{"text": "Hello, world!"}'
                      className="font-mono min-h-[200px]"
                    />

                    <div className="flex flex-wrap gap-2 mt-4">
                      <Button
                        onClick={handleExecute}
                        disabled={isExecuting || !isDeployed}
                        className="flex items-center gap-2"
                      >
                        {isExecuting ? (
                          <LoadingSpinner size="sm" />
                        ) : (
                          <Play className="h-4 w-4" />
                        )}
                        {isExecuting ? "Running..." : "Run Function"}
                      </Button>

                      <Button
                        variant="outline"
                        onClick={handleCopyShareableLink}
                        className="flex items-center gap-2"
                      >
                        {copied ? (
                          <Check className="h-4 w-4 text-green-500" />
                        ) : (
                          <Share2 className="h-4 w-4" />
                        )}
                        {copied ? "Copied!" : "Copy Link"}
                      </Button>

                      {playgroundInfo?.deployed_url && (
                        <Button
                          variant="ghost"
                          asChild
                          className="flex items-center gap-2"
                        >
                          <a
                            href={playgroundInfo.deployed_url}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            <ExternalLink className="h-4 w-4" />
                            View Deployment
                          </a>
                        </Button>
                      )}
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>

              <TabsContent value="output" className="mt-4">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">Response</CardTitle>
                    <CardDescription>
                      Function execution results will appear here.
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    {executionResult ? (
                      <div className="space-y-4">
                        {/* Status */}
                        <div className="flex items-center gap-3">
                          {executionResult.success ? (
                            <div className="flex items-center gap-2 text-green-500">
                              <CheckCircle2 className="h-5 w-5" />
                              <span className="font-medium">Success</span>
                            </div>
                          ) : (
                            <div className="flex items-center gap-2 text-red-500">
                              <XCircle className="h-5 w-5" />
                              <span className="font-medium">Error</span>
                            </div>
                          )}
                          <Badge variant="secondary">
                            {executionResult.status_code}
                          </Badge>
                          <span className="text-sm text-muted-foreground flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {executionResult.latency_ms}ms
                          </span>
                        </div>

                        {/* Output or Error */}
                        {executionResult.success ? (
                          <div>
                            <Label className="text-sm font-medium">Result</Label>
                            <div className="mt-2">
                              <JsonViewer data={executionResult.output} />
                            </div>
                          </div>
                        ) : (
                          <div>
                            <Label className="text-sm font-medium text-red-500">Error</Label>
                            <div className="mt-2 p-4 rounded-lg bg-red-500/10 border border-red-500/20">
                              <p className="text-red-500 font-mono text-sm">
                                {executionResult.error}
                              </p>
                            </div>
                          </div>
                        )}
                      </div>
                    ) : (
                      <div className="text-center py-12 text-muted-foreground">
                        <Play className="h-12 w-12 mx-auto mb-4 opacity-20" />
                        <p>Click "Run Function" to execute and see results</p>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          </div>

          {/* Sidebar - History */}
          <div className="space-y-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <div>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <History className="h-5 w-5" />
                    History
                  </CardTitle>
                  <CardDescription>Recent executions</CardDescription>
                </div>
                {history.length > 0 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleClearHistory}
                    className="text-muted-foreground hover:text-red-500"
                    aria-label="Clear execution history"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </CardHeader>
              <CardContent>
                {history.length > 0 ? (
                  <div className="space-y-1 max-h-[400px] overflow-y-auto">
                    {history.map((item) => (
                      <HistoryItem
                        key={item.id}
                        item={item}
                        onClick={() => handleLoadFromHistory(item)}
                        onDelete={() => handleDeleteHistory(item.id)}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    <Clock className="h-8 w-8 mx-auto mb-2 opacity-20" />
                    <p className="text-sm">No execution history</p>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Share Panel */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Share2 className="h-5 w-5" />
                  Share
                </CardTitle>
                <CardDescription>
                  Share this playground with pre-filled input
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div>
                    <Label className="text-xs text-muted-foreground">
                      Shareable URL
                    </Label>
                    <div className="flex gap-2 mt-1">
                      <Input
                        value={
                          (() => {
                            try {
                              const parsed = JSON.parse(inputValue);
                              return playgroundApi.createShareableUrl(
                                slug,
                                fnName,
                                parsed
                              );
                            } catch {
                              return playgroundApi.createShareableUrl(
                                slug,
                                fnName,
                                inputValue
                              );
                            }
                          })()
                        }
                        readOnly
                        className="font-mono text-xs"
                      />
                      <Button
                        variant="outline"
                        size="icon"
                        onClick={handleCopyShareableLink}
                        aria-label={copied ? 'Link copied' : 'Copy shareable link'}
                      >
                        {copied ? (
                          <Check className="h-4 w-4 text-green-500" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </div>

                  <p className="text-xs text-muted-foreground">
                    Anyone with this link can test the function with your
                    pre-filled input.
                  </p>
                </div>
              </CardContent>
            </Card>

            {/* Function Info */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Function Info</CardTitle>
              </CardHeader>
              <CardContent>
                <dl className="space-y-3 text-sm">
                  {isRegistry ? (
                    <>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">Type</dt>
                        <dd>Registry</dd>
                      </div>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">Runtime</dt>
                        <dd className="font-mono">{(registryInfo as Record<string, unknown>)?.runtime as string}</dd>
                      </div>
                    </>
                  ) : (
                    <>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">Status</dt>
                        <dd>
                          <StatusBadge status={playgroundInfo?.status || "unknown"} />
                        </dd>
                      </div>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">Version</dt>
                        <dd className="font-mono">{playgroundInfo?.version}</dd>
                      </div>
                      {playgroundInfo?.provider && (
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">Provider</dt>
                          <dd>{playgroundInfo.provider}</dd>
                        </div>
                      )}
                      {playgroundInfo?.region && (
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">Region</dt>
                          <dd>{playgroundInfo.region}</dd>
                        </div>
                      )}
                    </>
                  )}
                </dl>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Playground;
