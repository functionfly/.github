import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ArrowLeft } from "lucide-react";
import { Navbar } from "@/components/common/Navbar";
import { Footer } from "@/pages/LandingPage/components";
import { CodeBlock } from "@/components/common/CodeBlock";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";

interface FunctionInfo {
  author: string;
  name: string;
  version: string;
  title?: string;
  description?: string;
  runtime: string;
  category?: string;
  tags: string[];
  price_per_call: number;
  reliability: number;
  deterministic: boolean;
  cache_ttl: number;
  input_type?: string;
  output_type?: string;
  input_example?: any;
  output_example?: any;
  manifest?: any;
}

export function FunctionPage() {
  const { author, name } = useParams<{ author: string; name: string }>();

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

  const generateCodeExamples = (fn: FunctionInfo) => {
    const baseUrl = window.location.origin;
    const executeUrl = `${baseUrl}/v1/fx/${fn.author}/${fn.name}`;

    const inputExample = fn.input_example || "{}";

    return {
      curl: `curl -X POST "${executeUrl}" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify(inputExample, null, 2)}'`,
      javascript: `const response = await fetch('${executeUrl}', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify(${JSON.stringify(inputExample, null, 2)})
});

const result = await response.json();
console.log(result);`,
      python: `import requests

response = requests.post('${executeUrl}', json=${JSON.stringify(inputExample, null, 2)})
result = response.json()
print(result)`,
    };
  };

  const codeExamples = generateCodeExamples(functionInfo);

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary">
      <Navbar variant="landing" />
      <main className="flex-1 pt-16">
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          {/* Back button */}
          <div className="mb-6">
            <Link to="/registry">
              <Button variant="ghost" size="sm" className="gap-2 text-muted-foreground hover:text-foreground">
                <ArrowLeft className="w-4 h-4" />
                Back to Registry
              </Button>
            </Link>
          </div>

          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-4">
              <span>{functionInfo.author}</span>
              <span>/</span>
              <span>{functionInfo.name}</span>
              <Badge variant="secondary">{functionInfo.version}</Badge>
            </div>

          <h1 className="text-4xl font-bold mb-4">
            {functionInfo.title || `${functionInfo.author}/${functionInfo.name}`}
          </h1>

          {functionInfo.description && (
            <p className="text-xl text-muted-foreground mb-6">
              {functionInfo.description}
            </p>
          )}

          {/* Tags and metadata */}
          <div className="flex flex-wrap gap-2 mb-6">
            {functionInfo.category && (
              <Badge variant="outline">{functionInfo.category}</Badge>
            )}
            {functionInfo.tags.map((tag) => (
              <Badge key={tag} variant="secondary">
                {tag}
              </Badge>
            ))}
            <Badge variant="outline">{functionInfo.runtime}</Badge>
            {functionInfo.deterministic && (
              <Badge variant="outline">Deterministic</Badge>
            )}
          </div>

          {/* Stats */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            <div className="text-center">
              <div className="text-2xl font-bold text-primary">
                ${functionInfo.price_per_call.toFixed(6)}
              </div>
              <div className="text-sm text-muted-foreground">per call</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {functionInfo.reliability.toFixed(1)}%
              </div>
              <div className="text-sm text-muted-foreground">reliability</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">
                {functionInfo.cache_ttl}s
              </div>
              <div className="text-sm text-muted-foreground">cache TTL</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold">
                {functionInfo.input_type || "any"}
              </div>
              <div className="text-sm text-muted-foreground">input type</div>
            </div>
          </div>

          {/* CTA */}
          <div className="flex gap-4">
            <Link to={`/run/${functionInfo.author}/${functionInfo.name}`}>
              <Button size="lg" className="px-8">
                Try it
              </Button>
            </Link>
            <Button variant="outline" size="lg">
              View on GitHub
            </Button>
          </div>
        </div>

        {/* Code Examples */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold mb-4">Code Examples</h2>

          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>cURL</CardTitle>
                <CardDescription>
                  Make a direct HTTP request to execute the function
                </CardDescription>
              </CardHeader>
              <CardContent>
                <CodeBlock code={codeExamples.curl} language="bash" />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>JavaScript</CardTitle>
                <CardDescription>
                  Execute the function from a web application
                </CardDescription>
              </CardHeader>
              <CardContent>
                <CodeBlock code={codeExamples.javascript} language="javascript" />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Python</CardTitle>
                <CardDescription>
                  Execute the function from a Python application
                </CardDescription>
              </CardHeader>
              <CardContent>
                <CodeBlock code={codeExamples.python} language="python" />
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Input/Output Schema */}
        {functionInfo.manifest && (
          <div className="mb-8">
            <h2 className="text-2xl font-bold mb-4">Schema</h2>

            <div className="grid md:grid-cols-2 gap-6">
              {functionInfo.manifest.input && (
                <Card>
                  <CardHeader>
                    <CardTitle>Input</CardTitle>
                    <CardDescription>
                      Type: {functionInfo.manifest.input.type || "any"}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <CodeBlock
                      code={JSON.stringify(functionInfo.manifest.input, null, 2)}
                      language="json"
                    />
                  </CardContent>
                </Card>
              )}

              {functionInfo.manifest.output && (
                <Card>
                  <CardHeader>
                    <CardTitle>Output</CardTitle>
                    <CardDescription>
                      Type: {functionInfo.manifest.output.type || "any"}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <CodeBlock
                      code={JSON.stringify(functionInfo.manifest.output, null, 2)}
                      language="json"
                    />
                  </CardContent>
                </Card>
              )}
            </div>
          </div>
        )}
        </div>
      </main>
      <Footer />
    </div>
  );
}
