import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { ArrowLeft } from "lucide-react";
import { Navbar } from "@/components/common/Navbar";
import { Footer } from "@/pages/LandingPage/components";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { Clock, Zap, Calendar } from "lucide-react";

interface ReplayData {
  function_id: string;
  author: string;
  name: string;
  version: string;
  input_json: any;
  output_json: any;
  duration_ms: number;
  cached: boolean;
  execution_id: string;
  created_at: string;
}

export function ReplayPage() {
  const { execId } = useParams<{ execId: string }>();

  const { data: replayData, isLoading, error } = useQuery<ReplayData>({
    queryKey: ["replay", execId],
    queryFn: async () => {
      const response = await fetch(`/v1/registry/replay/${execId}`);
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error("Execution not found or not shareable");
        }
        throw new Error("Failed to fetch replay data");
      }
      return response.json();
    },
    enabled: !!execId,
  });

  const createRunAgainLink = (input: any) => {
    if (!replayData) return '';

    // Encode input as base64url for URL safety
    const inputJson = JSON.stringify(input);
    const encodedInput = btoa(inputJson).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    const baseUrl = `${window.location.origin}/run/${replayData.author}/${replayData.name}`;

    return `${baseUrl}?input=${encodedInput}`;
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

  if (!replayData) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16 flex items-center justify-center">
          <div className="text-center">
            <h1 className="text-2xl font-bold mb-2">Execution not found</h1>
            <p className="text-muted-foreground">
              The execution {execId} could not be found or is not shareable.
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

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary">
      <Navbar variant="landing" />
      <main className="flex-1 pt-16">
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          {/* Back button */}
          <div className="mb-6">
            <Link to={`/run/${replayData.author}/${replayData.name}`}>
              <Button variant="ghost" size="sm" className="gap-2 text-muted-foreground hover:text-foreground">
                <ArrowLeft className="w-4 h-4" />
                Back to Playground
              </Button>
            </Link>
          </div>

          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-4">
              <Link to={`/fx/${replayData.author}/${replayData.name}`} className="hover:underline">
                {replayData.author}/{replayData.name}
              </Link>
              <span>/</span>
              <span>Replay</span>
              <Badge variant="secondary">{replayData.execution_id}</Badge>
            </div>

          <h1 className="text-4xl font-bold mb-2">
            Function Execution Replay
          </h1>

          <p className="text-xl text-muted-foreground">
            Recorded execution of {replayData.author}/{replayData.name} v{replayData.version}
          </p>
        </div>

        {/* Metadata */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center space-x-2">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-2xl font-bold">{replayData.duration_ms}ms</p>
                  <p className="text-xs text-muted-foreground">Execution time</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center space-x-2">
                <Zap className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-2xl font-bold">
                    {replayData.cached ? "Cached" : "Live"}
                  </p>
                  <p className="text-xs text-muted-foreground">Execution type</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center space-x-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm font-bold">{formatTimestamp(replayData.created_at)}</p>
                  <p className="text-xs text-muted-foreground">Executed at</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="grid lg:grid-cols-2 gap-8">
          {/* Input */}
          <Card>
            <CardHeader>
              <CardTitle>Input</CardTitle>
              <CardDescription>
                Parameters passed to the function
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Textarea
                value={JSON.stringify(replayData.input_json, null, 2)}
                readOnly
                className="font-mono"
                rows={12}
              />
            </CardContent>
          </Card>

          {/* Output */}
          <Card>
            <CardHeader>
              <CardTitle>Output</CardTitle>
              <CardDescription>
                Result returned by the function
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Textarea
                value={JSON.stringify(replayData.output_json, null, 2)}
                readOnly
                className="font-mono"
                rows={12}
              />
            </CardContent>
          </Card>
        </div>

        {/* Actions */}
        <div className="mt-8 flex gap-4">
          <Link to={`/fx/${replayData.author}/${replayData.name}`}>
            <Button variant="outline">
              View Function Docs
            </Button>
          </Link>

          <Link to={createRunAgainLink(replayData.input_json)}>
            <Button>
              Run Again
            </Button>
          </Link>
        </div>

        {/* Footer note */}
        <div className="mt-12 pt-8 border-t border-border">
          <div className="text-center text-sm text-muted-foreground">
            <p>
              This is a recorded execution from the FunctionFly playground.
              Results may vary based on the current function version and external dependencies.
            </p>
          </div>
        </div>
        </div>
      </main>
      <Footer />
    </div>
  );
}
