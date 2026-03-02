import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { motion, AnimatePresence } from "framer-motion";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  ArrowLeft,
  Play,
  Github,
  DollarSign,
  Shield,
  Clock,
  Type,
  Terminal,
  FileJson,
  Zap,
  Users,
  Star,
  Package,
  ChevronRight,
  Share2,
  TrendingUp,
  Activity,
  Calendar,
  Hash,
  Sparkles,
  Loader2,
  Code2,
  BookOpen,
  BarChart3,
  Lock,
  User,
  Layers,
  ExternalLink,
} from "lucide-react";
import { Navbar } from "@/components/common/Navbar";
import { Footer } from "@/pages/LandingPage/components";
import { CodeBlock } from "@/components/common/CodeBlock";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { TrustScoreBadge, TrustLevel } from "@/components/common/TrustScoreBadge";
import { DeterministicReliabilityBadge } from "@/components/dre";
import { FollowFunctionButton } from "@/components/follow";
import { useState } from "react";
import { toast } from "sonner";
import { formatDistanceToNow } from "date-fns";
import ReactMarkdown from "react-markdown";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";

interface FunctionInfo {
  id: string;
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
  stars?: number;
  executions?: number;
  created_at?: string;
  updated_at?: string;
  popularity_score?: number;
  readme?: string;
  documentation_url?: string;
  /** Trust score 0–100 from ratings (included when rating exists) */
  trust_score?: number;
  trust_level?: string;
  /** Declared capabilities for sandbox / integrity */
  capabilities?: string[];
  /** Version integrity hash (displayed as ExecutionRootHash badge) */
  source_hash?: string;
}

export default function FunctionPage() {
  const { author, name } = useParams<{ author: string; name: string }>();

  const { data: functionInfo, isLoading, error } = useQuery<FunctionInfo>({
    queryKey: ["function", author, name],
    queryFn: async () => {
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
  });

  // Enhanced Loading Skeleton
  if (isLoading) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16">
          <div className="container mx-auto px-4 py-8 max-w-5xl">
            <div className="animate-pulse">
              <div className="h-4 w-32 bg-bg-tertiary rounded mb-6" />
              <div className="rounded-2xl bg-bg-tertiary/50 p-8 mb-8">
                <div className="h-4 w-48 bg-bg-secondary rounded mb-4" />
                <div className="h-12 w-3/4 bg-bg-secondary rounded mb-4" />
                <div className="h-6 w-1/2 bg-bg-secondary rounded mb-6" />
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                  {[...Array(4)].map((_, i) => (
                    <div key={i} className="h-24 bg-bg-secondary rounded-lg" />
                  ))}
                </div>
                <div className="flex gap-4">
                  <div className="h-12 w-32 bg-bg-secondary rounded" />
                  <div className="h-12 w-32 bg-bg-secondary rounded" />
                </div>
              </div>
              <div className="h-96 bg-bg-tertiary rounded-lg" />
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  if (error) {
    const isNotFound = (error as Error).message === "Function not found";
    if (isNotFound) {
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

    const inputExample = fn.input_example || fn.manifest?.input?.example || {};

    return {
      curl: `curl -X POST "${executeUrl}" \\
  -H "Content-Type: application/json" \\
  -d '${JSON.stringify(inputExample)}'`,
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

response = requests.post('${executeUrl}', json=${JSON.stringify(inputExample)})
result = response.json()
print(result)`,
    };
  };

  function ShareButton({ functionInfo }: { functionInfo: FunctionInfo }) {
    const [isSharing, setIsSharing] = useState(false);

    const handleShare = async () => {
      setIsSharing(true);
      const shareUrl = window.location.href;
      const shareData = {
        title: `${functionInfo.author}/${functionInfo.name}`,
        text: functionInfo.description || `Check out ${functionInfo.name} on FunctionFly`,
        url: shareUrl,
      };

      try {
        if (navigator.share) {
          await navigator.share(shareData);
          toast.success("Shared successfully");
        } else {
          await navigator.clipboard.writeText(shareUrl);
          toast.success("Link copied to clipboard");
        }
      } catch (err) {
        if ((err as Error).name !== "AbortError") {
          toast.error("Failed to share");
        }
      } finally {
        setIsSharing(false);
      }
    };

    return (
      <Button
        variant="outline"
        size="lg"
        className="gap-2"
        onClick={handleShare}
        disabled={isSharing}
      >
        {isSharing ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : (
          <Share2 className="w-4 h-4" />
        )}
        Share
      </Button>
    );
  }

  const codeExamples = generateCodeExamples(functionInfo);

  // Markdown renderer component (theme-aware: light mode = dark text, dark mode = inverted)
  const MarkdownRenderer = ({ content }: { content: string }) => (
    <div className="function-page-prose prose prose-sm max-w-none prose-invert">
      <ReactMarkdown
        components={{
          code({ node, inline, className, children, ...props }: any) {
            const match = /language-(\w+)/.exec(className || "");
            return !inline && match ? (
              <SyntaxHighlighter
                style={vscDarkPlus}
                language={match[1]}
                PreTag="div"
                {...props}
              >
                {String(children).replace(/\n$/, "")}
              </SyntaxHighlighter>
            ) : (
              <code className={className} {...props}>
                {children}
              </code>
            );
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary">
      <Navbar variant="landing" />
      <main className="flex-1 pt-16">
        <div className="container mx-auto px-4 py-8 max-w-5xl">
          {/* Back button */}
          <div className="mb-6">
            <Link to="/registry">
              <Button variant="ghost" size="sm" className="gap-2 text-muted-foreground hover:text-foreground">
                <ArrowLeft className="w-4 h-4" />
                Back to Registry
              </Button>
            </Link>
          </div>

          {/* Hero Section with Animation */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: "easeOut" }}
            className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-brand-500/10 via-brand-500/5 to-transparent border border-border-subtle p-8 mb-8"
          >
            <div className="relative z-10">
              {/* Breadcrumb */}
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-4">
                <Link to={`/registry?author=${functionInfo.author}`} className="hover:text-foreground transition-colors">
                  {functionInfo.author}
                </Link>
                <ChevronRight className="w-4 h-4" />
                <span className="text-foreground font-medium">{functionInfo.name}</span>
                <Badge variant="secondary" className="ml-2">v{functionInfo.version}</Badge>
                {functionInfo.deterministic && (
                  <Badge variant="outline" className="gap-1">
                    <Shield className="w-3 h-3" />
                    Deterministic
                  </Badge>
                )}
              </div>

              {/* Title */}
              <h1 className="text-4xl md:text-5xl font-bold mb-4">
                {functionInfo.title || functionInfo.name}
              </h1>

              {/* Description */}
              {functionInfo.description ? (
                <p className="text-xl text-muted-foreground mb-6 max-w-2xl">
                  {functionInfo.description}
                </p>
              ) : (
                <p className="text-xl text-muted-foreground/60 mb-6 italic">
                  No description available
                </p>
              )}

              {/* Tags */}
              <div className="flex flex-wrap gap-2 mb-6">
                {functionInfo.category && (
                  <Badge variant="outline" className="gap-1">
                    <Package className="w-3 h-3" />
                    {functionInfo.category}
                  </Badge>
                )}
                {functionInfo.tags.map((tag) => (
                  <Badge key={tag} variant="secondary">
                    {tag}
                  </Badge>
                ))}
                <Badge variant="outline" className="gap-1">
                  <Terminal className="w-3 h-3" />
                  {functionInfo.runtime}
                </Badge>
              </div>

              {/* Function profile: Owner, Trust Score, ExecutionRootHash, Capabilities (description & version live in hero/breadcrumb only) */}
              <Card className="function-profile-card mb-8 bg-bg-primary/60 border-border-subtle">
                <CardHeader className="pb-3">
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Layers className="w-5 h-5 text-brand-500" />
                    Function profile
                  </CardTitle>
                  <CardDescription>Identity, integrity, and capabilities for this function</CardDescription>
                </CardHeader>
                <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                  <div>
                    <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">Owner</p>
                    <Link to={`/registry?author=${functionInfo.author}`} className="text-sm font-medium text-brand-500 hover:underline flex items-center gap-1">
                      <User className="w-4 h-4" />
                      {functionInfo.author}
                    </Link>
                  </div>
                  <div>
                    <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">Trust score</p>
                    {functionInfo.trust_score != null ? (
                      <TrustScoreBadge
                        trustScore={functionInfo.trust_score}
                        trustLevel={functionInfo.trust_level as TrustLevel}
                        showScore
                        size="sm"
                      />
                    ) : (
                      <span className="text-sm text-muted-foreground">—</span>
                    )}
                  </div>
                  <div className="min-w-0">
                    <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">Execution root hash</p>
                    {functionInfo.source_hash ? (
                      <Link
                        to={`/registry/${functionInfo.author}/${functionInfo.name}/executions`}
                        className="inline-flex items-center min-w-0 max-w-full"
                      >
                        <Badge
                          variant="outline"
                          className="font-mono text-xs gap-1 min-w-0 max-w-full hover:bg-brand-500/10 hover:border-brand-500/30 transition-colors cursor-pointer"
                          title={`${functionInfo.source_hash} - Click to explore all executions`}
                        >
                          <Lock className="w-3 h-3 shrink-0" />
                          <span className="truncate block min-w-0">{functionInfo.source_hash}</span>
                          <ExternalLink className="w-3 h-3 shrink-0 ml-1" />
                        </Badge>
                      </Link>
                    ) : (
                      <Link
                        to={`/registry/${functionInfo.author}/${functionInfo.name}/executions`}
                        className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1 transition-colors"
                      >
                        <Lock className="w-3 h-3" />
                        View executions
                      </Link>
                    )}
                  </div>
                  <div>
                    <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">Capabilities</p>
                    {functionInfo.capabilities && functionInfo.capabilities.length > 0 ? (
                      <div className="flex flex-wrap gap-1">
                        {functionInfo.capabilities.map((cap) => (
                          <Badge key={cap} variant="outline" className="text-xs">
                            {cap}
                          </Badge>
                        ))}
                      </div>
                    ) : (
                      <span className="text-sm text-muted-foreground">None declared</span>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* Stats Cards with Animation */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8 function-page-hero-stats">
                {[
                  {
                    icon: <DollarSign className="w-5 h-5 text-green-500" />,
                    value: `$${functionInfo.price_per_call.toFixed(6)}`,
                    label: "per call",
                    color: "green",
                    delay: 0,
                  },
                  {
                    icon: <Shield className="w-5 h-5 text-blue-500" />,
                    value: (
                      <DeterministicReliabilityBadge
                        percentage={functionInfo.reliability * 100}
                        showTrend
                        trend="stable"
                      />
                    ),
                    label: "reliability",
                    color: "blue",
                    delay: 0.1,
                  },
                  {
                    icon: <Clock className="w-5 h-5 text-yellow-500" />,
                    value: `${functionInfo.cache_ttl}s`,
                    label: "cache TTL",
                    color: "yellow",
                    delay: 0.2,
                  },
                  {
                    icon: <Type className="w-5 h-5 text-purple-500" />,
                    value: functionInfo.input_type || "any",
                    label: "input type",
                    color: "purple",
                    delay: 0.3,
                  },
                ].map((stat, index) => (
                  <motion.div
                    key={stat.label}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.4, delay: 0.3 + stat.delay }}
                    whileHover={{ scale: 1.02, transition: { duration: 0.2 } }}
                  >
                    <Card className="function-stat-card bg-bg-primary/50 backdrop-blur border-border-subtle hover:border-brand-500/30 transition-colors cursor-default">
                      <CardContent className="p-4">
                        <div className="flex items-center gap-3">
                          <div className={`w-10 h-10 rounded-lg bg-${stat.color}-500/10 flex items-center justify-center`}>
                            {stat.icon}
                          </div>
                          <div>
                            <div className="text-2xl font-bold">{stat.value}</div>
                            <div className="text-xs text-muted-foreground">{stat.label}</div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </motion.div>
                ))}
              </div>

              {/* CTA Buttons */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4, delay: 0.6 }}
                className="flex flex-wrap gap-4"
              >
                <Link to={`/run/${functionInfo.author}/${functionInfo.name}`}>
                  <Button size="lg" className="gap-2 px-8">
                    <Play className="w-4 h-4" />
                    Try it Now
                  </Button>
                </Link>
                <Link to={`/registry/${functionInfo.author}/${functionInfo.name}/executions`}>
                  <Button variant="outline" size="lg" className="gap-2">
                    <Activity className="w-4 h-4" />
                    Executions
                  </Button>
                </Link>
                <Button variant="outline" size="lg" className="gap-2">
                  <Github className="w-4 h-4" />
                  View on GitHub
                </Button>
                <ShareButton functionInfo={functionInfo} />
                <FollowFunctionButton
                  functionId={functionInfo.id}
                  functionName={functionInfo.name}
                  size="lg"
                />
              </motion.div>

              {/* Auto-generated Documentation — directly under Try it Now */}
              <motion.div
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.35 }}
                className="function-page-api-reference mb-8 mt-8 overflow-hidden rounded-2xl border border-border-subtle bg-gradient-to-b from-bg-primary via-bg-primary/95 to-bg-secondary/30 shadow-sm"
              >
                <div className="flex flex-wrap items-center justify-between gap-4 border-b border-border-subtle bg-bg-tertiary/50 px-6 py-4">
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-500/10 text-brand-500">
                      <BookOpen className="h-5 w-5" />
                    </div>
                    <div>
                      <h2 className="text-lg font-semibold tracking-tight text-foreground">
                        API Reference
                      </h2>
                      <p className="text-xs text-muted-foreground">
                        Auto-generated from the function manifest
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" className="gap-2" asChild>
                      <a
                        href={`/v1/docs/${functionInfo.author}/${functionInfo.name}`}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <ExternalLink className="h-3.5 w-3.5" />
                        Full docs
                      </a>
                    </Button>
                  </div>
                </div>

                <div className="grid gap-0 lg:grid-cols-[minmax(0,200px)_1fr]">
                  <div className="border-b border-border-subtle bg-bg-primary/50 px-4 py-4 lg:border-b-0 lg:border-r lg:py-6">
                    <nav className="flex flex-wrap gap-1 lg:flex-col" aria-label="Documentation sections">
                      <a href="#doc-overview" className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-bg-tertiary hover:text-foreground">
                        Overview
                      </a>
                      {(functionInfo.manifest?.input || functionInfo.input_example != null) && (
                        <a href="#doc-input" className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-bg-tertiary hover:text-foreground">
                          Input
                        </a>
                      )}
                      {(functionInfo.manifest?.output || functionInfo.output_example != null) && (
                        <a href="#doc-output" className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-bg-tertiary hover:text-foreground">
                          Output
                        </a>
                      )}
                    </nav>
                  </div>

                  <ScrollArea className="max-h-[70vh] lg:max-h-none">
                    <div className="space-y-0 p-6">
                      <section id="doc-overview" className="scroll-mt-6">
                        <div className="rounded-xl border border-border-subtle bg-bg-primary/80 p-5 shadow-sm">
                          <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wider text-foreground">
                            <span className="flex h-6 w-6 items-center justify-center rounded bg-brand-500/15 text-brand-500 text-xs">1</span>
                            Overview
                          </h3>
                          <p className="mb-4 text-sm leading-relaxed text-muted-foreground">
                            {functionInfo.description || "No description provided."}
                          </p>
                          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                            <div className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-3 py-2.5">
                              <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">Runtime</p>
                              <p className="mt-0.5 font-mono text-sm text-foreground">{functionInfo.runtime}</p>
                            </div>
                            {functionInfo.manifest?.deterministic != null && (
                              <div className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-3 py-2.5">
                                <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">Determinism</p>
                                <p className="mt-0.5 text-sm text-foreground">
                                  {functionInfo.manifest.deterministic ? "Deterministic" : "Non-deterministic"}
                                </p>
                              </div>
                            )}
                            {functionInfo.cache_ttl != null && functionInfo.cache_ttl > 0 && (
                              <div className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-3 py-2.5">
                                <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">Cache TTL</p>
                                <p className="mt-0.5 text-sm text-foreground">{functionInfo.cache_ttl}s</p>
                              </div>
                            )}
                          </div>
                        </div>
                      </section>

                      {(functionInfo.manifest?.input || functionInfo.input_example != null) && (
                        <section id="doc-input" className="scroll-mt-6 pt-6">
                          <div className="rounded-xl border border-border-subtle bg-bg-primary/80 p-5 shadow-sm">
                            <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wider text-foreground">
                              <span className="flex h-6 w-6 items-center justify-center rounded bg-emerald-500/15 text-emerald-600 text-xs dark:text-emerald-400">2</span>
                              Input
                            </h3>
                            {functionInfo.manifest?.input?.properties && typeof functionInfo.manifest.input.properties === "object" && (
                              <div className="mb-4 overflow-hidden rounded-lg border border-border-subtle">
                                <table className="w-full text-left text-sm">
                                  <thead>
                                    <tr className="border-b border-border-subtle bg-bg-tertiary/70">
                                      <th className="px-4 py-2.5 font-medium text-foreground">Property</th>
                                      <th className="px-4 py-2.5 font-medium text-foreground">Type</th>
                                      <th className="px-4 py-2.5 font-medium text-foreground">Description</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {Object.entries(functionInfo.manifest.input.properties).map(([key, val]: [string, unknown]) => {
                                      const v = val as { type?: string; description?: string; default?: unknown };
                                      return (
                                        <tr key={key} className="border-b border-border-subtle/50 last:border-0">
                                          <td className="px-4 py-2 font-mono text-xs text-foreground">{key}</td>
                                          <td className="px-4 py-2 text-muted-foreground">{v?.type ?? "—"}</td>
                                          <td className="px-4 py-2 text-muted-foreground">{v?.description ?? "—"}</td>
                                        </tr>
                                      );
                                    })}
                                  </tbody>
                                </table>
                              </div>
                            )}
                            <Tabs defaultValue="schema" className="w-full">
                              <TabsList className="mb-2 h-9">
                                <TabsTrigger value="schema" className="text-xs">Schema</TabsTrigger>
                                {functionInfo.input_example != null && (
                                  <TabsTrigger value="example" className="text-xs">Example</TabsTrigger>
                                )}
                              </TabsList>
                              <TabsContent value="schema" className="mt-0">
                                {functionInfo.manifest?.input ? (
                                  <CodeBlock code={JSON.stringify(functionInfo.manifest.input, null, 2)} language="json" />
                                ) : (
                                  <p className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-4 py-3 text-sm text-muted-foreground">No schema defined.</p>
                                )}
                              </TabsContent>
                              {functionInfo.input_example != null && (
                                <TabsContent value="example" className="mt-0">
                                  <CodeBlock
                                    code={typeof functionInfo.input_example === "string"
                                      ? functionInfo.input_example
                                      : JSON.stringify(functionInfo.input_example, null, 2)}
                                    language="json"
                                  />
                                </TabsContent>
                              )}
                            </Tabs>
                          </div>
                        </section>
                      )}

                      {(functionInfo.manifest?.output || functionInfo.output_example != null) && (
                        <section id="doc-output" className="scroll-mt-6 pt-6">
                          <div className="rounded-xl border border-border-subtle bg-bg-primary/80 p-5 shadow-sm">
                            <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wider text-foreground">
                              <span className="flex h-6 w-6 items-center justify-center rounded bg-violet-500/15 text-violet-600 text-xs dark:text-violet-400">3</span>
                              Output
                            </h3>
                            {functionInfo.manifest?.output?.properties && typeof functionInfo.manifest.output.properties === "object" && (
                              <div className="mb-4 overflow-hidden rounded-lg border border-border-subtle">
                                <table className="w-full text-left text-sm">
                                  <thead>
                                    <tr className="border-b border-border-subtle bg-bg-tertiary/70">
                                      <th className="px-4 py-2.5 font-medium text-foreground">Property</th>
                                      <th className="px-4 py-2.5 font-medium text-foreground">Type</th>
                                      <th className="px-4 py-2.5 font-medium text-foreground">Description</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {Object.entries(functionInfo.manifest.output.properties).map(([key, val]: [string, unknown]) => {
                                      const v = val as { type?: string; description?: string };
                                      return (
                                        <tr key={key} className="border-b border-border-subtle/50 last:border-0">
                                          <td className="px-4 py-2 font-mono text-xs text-foreground">{key}</td>
                                          <td className="px-4 py-2 text-muted-foreground">{v?.type ?? "—"}</td>
                                          <td className="px-4 py-2 text-muted-foreground">{v?.description ?? "—"}</td>
                                        </tr>
                                      );
                                    })}
                                  </tbody>
                                </table>
                              </div>
                            )}
                            <Tabs defaultValue="schema" className="w-full">
                              <TabsList className="mb-2 h-9">
                                <TabsTrigger value="schema" className="text-xs">Schema</TabsTrigger>
                                {functionInfo.output_example != null && (
                                  <TabsTrigger value="example" className="text-xs">Example</TabsTrigger>
                                )}
                              </TabsList>
                              <TabsContent value="schema" className="mt-0">
                                {functionInfo.manifest?.output ? (
                                  <CodeBlock code={JSON.stringify(functionInfo.manifest.output, null, 2)} language="json" />
                                ) : (
                                  <p className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-4 py-3 text-sm text-muted-foreground">No schema defined.</p>
                                )}
                              </TabsContent>
                              {functionInfo.output_example != null && (
                                <TabsContent value="example" className="mt-0">
                                  <CodeBlock
                                    code={typeof functionInfo.output_example === "string"
                                      ? functionInfo.output_example
                                      : JSON.stringify(functionInfo.output_example, null, 2)}
                                    language="json"
                                  />
                                </TabsContent>
                              )}
                            </Tabs>
                          </div>
                        </section>
                      )}
                    </div>
                  </ScrollArea>
                </div>
              </motion.div>
            </div>
          </motion.div>

          {/* Code Examples with Tabs */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.7 }}
          >
          <Card className="mb-8 code-examples-card">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Terminal className="w-5 h-5 text-brand-500" />
                Code Examples
              </CardTitle>
              <CardDescription className="text-text-secondary code-examples-description">
                Use these examples to integrate this function into your application
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Tabs defaultValue="curl" className="w-full code-examples-tabs">
                <TabsList className="grid w-full grid-cols-3 mb-4 code-examples-tabs-list">
                  <TabsTrigger value="curl" className="gap-2">
                    <Terminal className="w-4 h-4" />
                    cURL
                  </TabsTrigger>
                  <TabsTrigger value="javascript" className="gap-2">
                    <FileJson className="w-4 h-4" />
                    JavaScript
                  </TabsTrigger>
                  <TabsTrigger value="python" className="gap-2">
                    <Zap className="w-4 h-4" />
                    Python
                  </TabsTrigger>
                </TabsList>
                <TabsContent value="curl">
                  <CodeBlock code={codeExamples.curl} language="bash" />
                </TabsContent>
                <TabsContent value="javascript">
                  <CodeBlock code={codeExamples.javascript} language="javascript" />
                </TabsContent>
                <TabsContent value="python">
                  <CodeBlock code={codeExamples.python} language="python" />
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>
          </motion.div>

          {/* Schema Section */}
          {(functionInfo.manifest?.input || functionInfo.manifest?.output) && (
            <div className="mb-8">
              <h2 className="text-2xl font-bold mb-4 flex items-center gap-2">
                <FileJson className="w-6 h-6 text-brand-500" />
                Input / Output Schema
              </h2>

              <div className="grid md:grid-cols-2 gap-6">
                {functionInfo.manifest.input && (
                  <Card>
                    <CardHeader className="pb-3">
                      <CardTitle className="text-lg">Input</CardTitle>
                      <CardDescription>
                        Expected input structure
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
                    <CardHeader className="pb-3">
                      <CardTitle className="text-lg">Output</CardTitle>
                      <CardDescription>
                        Expected output structure
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

          {/* Related Functions Placeholder */}
          <Card className="bg-gradient-to-br from-brand-500/5 to-transparent border-dashed">
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-xl bg-brand-500/10 flex items-center justify-center">
                  <Package className="w-6 h-6 text-brand-500" />
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold text-lg">Explore More Functions</h3>
                  <p className="text-muted-foreground text-sm">
                    Discover related functions in the registry to build powerful workflows
                  </p>
                </div>
                <Link to="/registry">
                  <Button variant="outline">
                    Browse Registry
                    <ChevronRight className="w-4 h-4 ml-1" />
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        </div>
      </main>
      <Footer />
    </div>
  );
}

