import { useState, useEffect } from "react";
import { useParams, useSearchParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Copy, Check, Play, ExternalLink, ArrowLeft } from "lucide-react";
import { Navbar } from "@/components/common/Navbar";
import { Footer } from "@/pages/LandingPage/components";
import { ManifestInputForm } from "@/components/common/ManifestInputForm";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";

interface FunctionInfo {
  author: string;
  name: string;
  version: string;
  title?: string;
  description?: string;
  runtime: string;
  manifest?: any;
}

interface ExecutionResponse {
  ok: boolean;
  data?: any;
  cached: boolean;
  duration_ms: number;
  version: string;
  execution_id?: string;
  error?: {
    code: string;
    message: string;
  };
}

export function PlaygroundPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [inputValue, setInputValue] = useState<any>({});
  const [isExecuting, setIsExecuting] = useState(false);
  const [executionResult, setExecutionResult] = useState<ExecutionResponse | null>(null);
  const [shareableLink, setShareableLink] = useState<string | null>(null);
  const [copiedLink, setCopiedLink] = useState(false);

  // Get input from URL params
  useEffect(() => {
    const inputParam = searchParams.get('input');
    if (inputParam) {
      try {
        // Try to decode as base64url first, then as regular JSON
        let decodedInput;
        try {
          decodedInput = JSON.parse(atob(inputParam.replace(/-/g, '+').replace(/_/g, '/')));
        } catch {
          decodedInput = JSON.parse(decodeURIComponent(inputParam));
        }
        setInputValue(decodedInput);
      } catch (error) {
        console.warn('Failed to parse input from URL:', error);
      }
    }
  }, [searchParams]);

  const { data: functionInfo, isLoading, error } = useQuery<FunctionInfo>({
    queryKey: ["function", author, name],
    queryFn: async () => {
      const response = await fetch(
        `/v1/registry/functions/${author}/${name}?expand=manifest`
      );
      if (!response.ok) {
        throw new Error("Failed to fetch function");
      }
      return response.json();
    },
    enabled: !!author && !!name,
  });

  const handleExecute = async () => {
    if (!functionInfo) return;

    setIsExecuting(true);
    setExecutionResult(null);
    setShareableLink(null);

    try {
      const response = await fetch(`/v1/fx/${functionInfo.author}/${functionInfo.name}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Playground': '1', // Playground marker
        },
        body: JSON.stringify(inputValue),
      });

      const result: ExecutionResponse = await response.json();
      setExecutionResult(result);

      // If successful and has execution_id, create shareable link
      if (result.ok && result.execution_id) {
        const shareLink = `${window.location.origin}/replay/${result.execution_id}`;
        setShareableLink(shareLink);
      }
    } catch (error) {
      setExecutionResult({
        ok: false,
        cached: false,
        duration_ms: 0,
        version: functionInfo.version,
        error: {
          code: 'NETWORK_ERROR',
          message: 'Failed to execute function',
        },
      });
    } finally {
      setIsExecuting(false);
    }
  };

  const handleCopyShareableLink = async () => {
    if (!shareableLink) return;

    try {
      await navigator.clipboard.writeText(shareableLink);
      setCopiedLink(true);
      setTimeout(() => setCopiedLink(false), 2000);
    } catch (error) {
      console.error('Failed to copy link:', error);
    }
  };

  const createShareableInputLink = () => {
    if (!functionInfo) return '';

    // Encode input as base64url for URL safety
    const inputJson = JSON.stringify(inputValue);
    const encodedInput = btoa(inputJson).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    const baseUrl = `${window.location.origin}/run/${functionInfo.author}/${functionInfo.name}`;

    return `${baseUrl}?input=${encodedInput}`;
  };

  const handleCopyInputLink = async () => {
    const link = createShareableInputLink();
    try {
      await navigator.clipboard.writeText(link);
      setCopiedLink(true);
      setTimeout(() => setCopiedLink(false), 2000);
    } catch (error) {
      console.error('Failed to copy link:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <LoadingSpinner />
        </main>
        <Footer />
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <ErrorMessage error={error as Error} />
        </main>
        <Footer />
      </div>
    );
  }

  if (!functionInfo) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <div className="text-center">
            <h1 className="text-2xl font-bold mb-2">Function not found</h1>
            <p className="text-muted-foreground">
              The function {author}/{name} could not be found.
            </p>
            <Link to="/registry">
              <Button variant="outline" className="mt-4">
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Registry
              </Button>
            </Link>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary">
      <Navbar variant="landing" />
      <main className="flex-1 pt-16">
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          {/* Back button */}
          <div className="mb-6">
            <Link to={`/fx/${functionInfo.author}/${functionInfo.name}`}>
              <Button variant="ghost" size="sm" className="gap-2 text-muted-foreground hover:text-foreground">
                <ArrowLeft className="w-4 h-4" />
                Back to {functionInfo.title || functionInfo.name}
              </Button>
            </Link>
          </div>

          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-4">
              <Link to={`/fx/${functionInfo.author}/${functionInfo.name}`} className="hover:underline">
                {functionInfo.author}/{functionInfo.name}
              </Link>
              <span>/</span>
              <span>Playground</span>
              <Badge variant="secondary">{functionInfo.version}</Badge>
            </div>

          <h1 className="text-4xl font-bold mb-2">
            {functionInfo.title || `${functionInfo.author}/${functionInfo.name}`} Playground
          </h1>

          {functionInfo.description && (
            <p className="text-xl text-muted-foreground">
              {functionInfo.description}
            </p>
          )}
        </div>

        <div className="grid lg:grid-cols-2 gap-8">
          {/* Input Section */}
          <div>
            <Card>
              <CardHeader>
                <CardTitle>Input</CardTitle>
                <CardDescription>
                  Configure the function parameters
                </CardDescription>
              </CardHeader>
              <CardContent>
                {functionInfo.manifest?.input ? (
                  <ManifestInputForm
                    inputSpec={functionInfo.manifest.input}
                    value={inputValue}
                    onChange={setInputValue}
                  />
                ) : (
                  <Textarea
                    value={typeof inputValue === 'string' ? inputValue : JSON.stringify(inputValue, null, 2)}
                    onChange={(e) => {
                      try {
                        setInputValue(JSON.parse(e.target.value));
                      } catch {
                        setInputValue(e.target.value);
                      }
                    }}
                    placeholder="Enter input as JSON..."
                    className="font-mono"
                    rows={8}
                  />
                )}

                <div className="flex gap-2 mt-4">
                  <Button
                    onClick={handleExecute}
                    disabled={isExecuting}
                    className="flex items-center gap-2"
                  >
                    {isExecuting ? (
                      <LoadingSpinner size="sm" />
                    ) : (
                      <Play className="h-4 w-4" />
                    )}
                    {isExecuting ? 'Running...' : 'Run'}
                  </Button>

                  <Button
                    variant="outline"
                    onClick={handleCopyInputLink}
                    className="flex items-center gap-2"
                  >
                    {copiedLink ? (
                      <Check className="h-4 w-4 text-green-500" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                    Copy Link
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Output Section */}
          <div>
            <Card>
              <CardHeader>
                <CardTitle>Output</CardTitle>
                <CardDescription>
                  Function execution results
                </CardDescription>
              </CardHeader>
              <CardContent>
                {executionResult ? (
                  <div className="space-y-4">
                    {/* Status */}
                    <div className="flex items-center gap-2">
                      <Badge variant={executionResult.ok ? "default" : "destructive"}>
                        {executionResult.ok ? "Success" : "Error"}
                      </Badge>
                      <span className="text-sm text-muted-foreground">
                        {executionResult.duration_ms}ms
                      </span>
                      {executionResult.cached && (
                        <Badge variant="secondary">Cached</Badge>
                      )}
                    </div>

                    {/* Result */}
                    {executionResult.ok ? (
                      <div>
                        <Label className="text-sm font-medium">Result</Label>
                        <Textarea
                          value={JSON.stringify(executionResult.data, null, 2)}
                          readOnly
                          className="font-mono mt-1"
                          rows={8}
                        />
                      </div>
                    ) : (
                      <div>
                        <Label className="text-sm font-medium text-red-600">Error</Label>
                        <div className="mt-1 p-3 bg-red-50 border border-red-200 rounded-md">
                          <div className="font-medium text-red-800">
                            {executionResult.error?.code || 'execution_failed'}
                          </div>
                          <div className="text-red-700 mt-1">
                            {executionResult.error?.message || executionResult.error?.code || 'Execution failed (no details)'}
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Share Link */}
                    {shareableLink && (
                      <div>
                        <Label className="text-sm font-medium">Share Result</Label>
                        <div className="flex gap-2 mt-1">
                          <Input
                            value={shareableLink}
                            readOnly
                            className="font-mono text-sm"
                          />
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={handleCopyShareableLink}
                            className="flex items-center gap-2"
                          >
                            {copiedLink ? (
                              <Check className="h-4 w-4 text-green-500" />
                            ) : (
                              <Copy className="h-4 w-4" />
                            )}
                            Copy
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            asChild
                          >
                            <a
                              href={shareableLink}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="flex items-center gap-2"
                            >
                              <ExternalLink className="h-4 w-4" />
                              Open
                            </a>
                          </Button>
                        </div>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="text-center text-muted-foreground py-8">
                    Run the function to see results here
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
