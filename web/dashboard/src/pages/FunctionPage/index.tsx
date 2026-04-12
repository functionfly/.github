import { registryApi, type RegistryFunctionReview } from '@/api/registry';
import { CodeBlock } from '@/components/common/CodeBlock';
import { ErrorMessage } from '@/components/common/ErrorMessage';
import { Navbar } from '@/components/common/Navbar';
import { FollowFunctionButton } from '@/components/follow';
import { FunctionCard, FunctionHeader, TrustScoreBadge } from '@/components/functions';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { Footer } from '@/pages/LandingPage/components';
import { useAuthStore } from '@/stores/authStore';
import type {
  FraudRiskLevel,
  FunctionCardData,
  FunctionHeaderData,
  TrustMetrics,
  TrustTier,
} from '@/types';
import { Icon } from '@iconify/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { formatDistanceToNow } from 'date-fns';
import { motion } from 'framer-motion';
import {
  Activity,
  ArrowLeft,
  BookOpen,
  ChevronRight,
  ExternalLink,
  FileJson,
  Loader2,
  Package,
  Play,
  Share2,
  Shield,
  Star,
  Terminal,
  Zap,
} from 'lucide-react';
import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { toast } from 'sonner';

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
  /** Whether this version is verified (e.g. functionfly official) */
  verified?: boolean;
  /** Declared capabilities for sandbox / integrity */
  capabilities?: string[];
  /** Version integrity hash (displayed as ExecutionRootHash badge) */
  source_hash?: string;
}

/**
 * Map FunctionInfo to FunctionHeaderData format
 */
function mapToFunctionHeaderData(
  data: FunctionInfo,
  trustTier: TrustTier = 'medium'
): FunctionHeaderData {
  // Use source_hash if available, otherwise generate a placeholder
  const executionRootHash = data.source_hash
    ? data.source_hash.startsWith('0x')
      ? data.source_hash
      : `0x${data.source_hash}`
    : `0x${data.author}${data.name}`.slice(0, 66).padEnd(66, '0');

  // Generate resource signature from author/name
  const resourceSignature = `res_sig_${data.author.slice(0, 8)}${data.name.slice(0, 8)}`;

  // Map trust_level to TrustTier (backend sends high|medium|low|untrusted)
  const mappedTrustTier: TrustTier =
    (data.trust_level?.toLowerCase() as TrustTier) ||
    (data.trust_score != null && data.trust_score >= 80
      ? 'high'
      : data.trust_score != null && data.trust_score >= 50
        ? 'medium'
        : data.trust_score != null && data.trust_score >= 20
          ? 'low'
          : trustTier);

  // Calculate economic score from reliability and other factors
  const economicScore = Math.round((data.reliability || 0.5) * 100);

  return {
    name: data.title || data.name,
    id: `${data.author}/${data.name}`,
    executionRootHash,
    trustTier: mappedTrustTier,
    economicScore,
    runtime: data.runtime,
    resourceSignature,
    fxcert: {
      verified: data.verified === true || (data.verified !== false && data.deterministic) || false,
      issuedAt: data.created_at,
      issuer: 'FunctionFly Registry',
    },
    status: 'online',
    version: `v${data.version}`,
    description: data.description,
  };
}

/**
 * Map FunctionInfo to FunctionCardData format
 */
function mapToFunctionCardData(data: FunctionInfo): FunctionCardData {
  return {
    id: `${data.author}/${data.name}`,
    name: data.title || data.name,
    description: data.description || 'No description available',
    author: {
      id: data.author,
      username: data.author,
      name: data.author,
    },
    trustScore: data.trust_score ?? Math.round((data.reliability || 0.5) * 100),
    metrics: {
      executionCount: data.executions || 0,
      executionTrend: data.popularity_score ? [data.popularity_score] : undefined,
    },
    pricing: {
      model: data.price_per_call === 0 ? 'free' : 'per_call',
      pricePerCall: data.price_per_call,
      currency: 'USD',
    },
    isVerified: data.deterministic || false,
    isDeterministic: data.deterministic || false,
    rating: {
      average: data.stars ? Math.min(data.stars / 20, 5) : 0,
      count: data.stars || 0,
    },
    tags: data.tags,
    category: data.category,
    language: data.runtime,
    lastUpdated: data.updated_at,
    version: data.version,
    isFavorite: false,
    isFeatured: data.popularity_score ? data.popularity_score > 80 : false,
  };
}

/**
 * Map FunctionInfo to TrustMetrics format for TrustScoreBadge
 */
function mapToTrustMetrics(data: FunctionInfo): TrustMetrics {
  // Calculate overall score from trust_score or derive from reliability
  const overallScore = data.trust_score ?? Math.round((data.reliability || 0.5) * 100);

  // Reliability is already a 0-1 value, convert to 0-100
  const reliability = Math.round((data.reliability || 0.5) * 100);

  // Latency score - assume 100 if no data (higher is better for this metric)
  // In a real scenario, this would be calculated based on avg_response_time
  // For now, default to good score
  const latency = 85;

  // Determinism score - based on deterministic flag
  const determinism = data.deterministic ? 95 : 50;

  // Community reputation - based on stars/ratings
  // Normalize stars (assuming 0-100 scale) to 0-100 score
  const communityReputation = data.stars ? Math.min(data.stars, 100) : 50;

  // Fraud risk - default to low
  const fraudRisk: FraudRiskLevel = 'low';

  return {
    overallScore,
    reliability,
    latency,
    determinism,
    communityReputation,
    fraudRisk,
    details: {
      totalExecutions: data.executions,
      lastUpdated: data.updated_at,
    },
  };
}

export default function FunctionPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const {
    data: functionInfo,
    isLoading,
    error,
  } = useQuery<FunctionInfo>({
    queryKey: ['function', author, name],
    queryFn: async () => {
      const response = await fetch(`/v1/registry/functions/${author}/${name}?expand=manifest`);
      if (response.status === 404) {
        throw new Error('Function not found');
      }
      if (!response.ok) {
        throw new Error('Failed to fetch function');
      }
      return response.json();
    },
    enabled: !!author && !!name,
  });

  // Reviews (public list + authenticated submit)
  // NOTE: hooks must stay above early returns to avoid hook-order errors during loading.
  const [reviewDialogOpen, setReviewDialogOpen] = useState(false);
  const [reviewStars, setReviewStars] = useState(5);
  const [reviewTitle, setReviewTitle] = useState('');
  const [reviewBody, setReviewBody] = useState('');

  const { data: reviewsResp, isLoading: isLoadingReviews } = useQuery({
    queryKey: ['function-reviews', author, name],
    queryFn: async () => {
      if (!author || !name) return null;
      return registryApi.listReviews(author, name, { limit: 20, offset: 0 });
    },
    enabled: !!author && !!name,
  });

  const reviews: RegistryFunctionReview[] = reviewsResp?.reviews ?? [];
  const reviewsTotal = reviewsResp?.total ?? reviews.length;

  const submitReviewMutation = useMutation({
    mutationFn: async () => {
      if (!author || !name) throw new Error('Missing function');
      return registryApi.submitReview(author, name, {
        stars: reviewStars,
        title: reviewTitle,
        body: reviewBody,
      });
    },
    onSuccess: () => {
      toast.success('Review submitted');
      setReviewDialogOpen(false);
      queryClient.invalidateQueries({ queryKey: ['function-reviews', author, name] });
      // Optional: also submit aggregate rating signal so overall_score updates quickly
      registryApi
        .submitRating(author!, name!, {
          overall_score: reviewStars,
          reliability_score: 0,
          latency_score: 0,
          documentation_score: 0,
        })
        .catch(() => {
          /* ignore */
        });
    },
    onError: (e) => {
      const msg = (e as Error)?.message || 'Failed to submit review';
      if (msg.toLowerCase().includes('execute this function')) {
        toast.info('Run it once to unlock reviews', {
          action: {
            label: 'Try it now',
            onClick: () => navigate(`/run/${author}/${name}`),
          },
        });
        return;
      }
      toast.error(msg);
    },
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
    const isNotFound = (error as Error).message === 'Function not found';
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
          toast.success('Shared successfully');
        } else {
          await navigator.clipboard.writeText(shareUrl);
          toast.success('Link copied to clipboard');
        }
      } catch (err) {
        if ((err as Error).name !== 'AbortError') {
          toast.error('Failed to share');
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
        {isSharing ? <Loader2 className="w-4 h-4 animate-spin" /> : <Share2 className="w-4 h-4" />}
        Share
      </Button>
    );
  }

  const codeExamples = generateCodeExamples(functionInfo);

  // Markdown renderer component (theme-aware: light mode = dark text, dark mode = inverted)
  const MarkdownRenderer = ({ content }: { content: string }) => (
    <div className="function-page-prose prose prose-sm max-w-none text-foreground dark:prose-invert">
      <ReactMarkdown
        components={{
          code({ node, inline, className, children, ...props }: any) {
            const match = /language-(\w+)/.exec(className || '');
            return !inline && match ? (
              <SyntaxHighlighter style={vscDarkPlus} language={match[1]} PreTag="div" {...props}>
                {String(children).replace(/\n$/, '')}
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
          {/* Function Header */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: 'easeOut' }}
            className="mb-8"
          >
            <FunctionHeader
              data={mapToFunctionHeaderData(functionInfo)}
              onBack={() => navigate('/registry')}
              onShare={() => {
                const shareUrl = window.location.href;
                navigator.clipboard.writeText(shareUrl).then(() => {
                  toast.success('Link copied to clipboard');
                });
              }}
            />

            {/* CTA Buttons */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: 0.3 }}
              className="flex flex-wrap gap-4 mt-6"
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
              <Link
                to={`/registry/${functionInfo.author}/${functionInfo.name}/executions?tab=certificates`}
              >
                <Button variant="outline" size="lg" className="gap-2">
                  <Shield className="w-4 h-4" />
                  Certificates
                </Button>
              </Link>
              <Button variant="outline" size="lg" className="gap-2">
                <Icon icon="simple-icons:github" className="w-4 h-4" />
                View on GitHub
              </Button>
              <ShareButton functionInfo={functionInfo} />
              <FollowFunctionButton
                functionId={functionInfo.id}
                functionName={functionInfo.name}
                size="lg"
              />
            </motion.div>
          </motion.div>

          {/* Function Card - Function Overview */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="mb-8"
          >
            <FunctionCard
              data={mapToFunctionCardData(functionInfo)}
              variant="expanded"
              onView={() => {
                document.getElementById('function-api-reference')?.scrollIntoView({
                  behavior: 'smooth',
                  block: 'start',
                });
              }}
              onExecute={() => navigate(`/run/${functionInfo.author}/${functionInfo.name}`)}
              onShare={() => {
                const shareUrl = window.location.href;
                navigator.clipboard.writeText(shareUrl).then(() => {
                  toast.success('Link copied to clipboard');
                });
              }}
              onFavorite={(id, isFav) => {
                toast.info(isFav ? 'Added to favorites' : 'Removed from favorites');
              }}
            />
          </motion.div>

          {/* Trust Score Badge - Expanded View */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="mb-8"
          >
            <div className="grid md:grid-cols-2 gap-6">
              <TrustScoreBadge
                metrics={mapToTrustMetrics(functionInfo)}
                variant="expanded"
                showDetails={true}
              />
              <div className="flex items-center justify-center rounded-xl border border-border-subtle bg-bg-secondary/50 p-4">
                <div className="text-center">
                  <div className="flex items-center justify-center gap-2 mb-2">
                    <Shield className="h-5 w-5 text-brand-500" />
                    <h3 className="font-semibold text-foreground">Trust & Verification</h3>
                  </div>
                  <p className="text-sm text-muted-foreground max-w-sm">
                    This function has been verified by the FunctionFly registry. Trust scores are
                    calculated based on execution reliability, community ratings, and deterministic
                    behavior.
                  </p>
                </div>
              </div>
            </div>
          </motion.div>

          {/* Reviews */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.35 }}
            className="mb-10"
          >
            <div className="flex items-center justify-between gap-4 mb-4">
              <div>
                <h2 className="text-lg font-semibold text-foreground">Community reviews</h2>
                <p className="text-sm text-muted-foreground">
                  {reviewsTotal} {reviewsTotal === 1 ? 'review' : 'reviews'}
                </p>
              </div>
              <Button
                variant="outline"
                onClick={() => {
                  if (!isAuthenticated) {
                    toast.info('Sign in to leave a review');
                    navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
                    return;
                  }
                  setReviewDialogOpen(true);
                }}
                className="gap-2"
              >
                <Star className="h-4 w-4" />
                Write a review
              </Button>
            </div>

            <div className="grid gap-3">
              {isLoadingReviews ? (
                <div className="rounded-xl border border-border-subtle bg-bg-secondary/30 p-4 text-sm text-muted-foreground">
                  Loading reviews…
                </div>
              ) : reviews.length === 0 ? (
                <div className="rounded-xl border border-border-subtle bg-bg-secondary/30 p-6 text-center">
                  <p className="text-sm text-muted-foreground">No reviews yet. Be the first.</p>
                </div>
              ) : (
                reviews.map((r) => (
                  <div
                    key={r.id}
                    className="rounded-xl border border-border-subtle bg-bg-secondary/30 p-4"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <div className="flex items-center gap-0.5">
                            {Array.from({ length: 5 }).map((_, i) => (
                              <Star
                                key={i}
                                className={`h-4 w-4 ${
                                  i < (r.stars ?? 0)
                                    ? 'fill-amber-400 text-amber-400'
                                    : 'text-muted-foreground/40'
                                }`}
                              />
                            ))}
                          </div>
                          <span className="text-xs text-muted-foreground">
                            {formatDistanceToNow(new Date(r.created_at), { addSuffix: true })}
                          </span>
                        </div>
                        {r.title && (
                          <div className="mt-2 font-medium text-foreground truncate">{r.title}</div>
                        )}
                        {r.body && (
                          <div className="mt-1 text-sm text-muted-foreground whitespace-pre-wrap">
                            {r.body}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </motion.div>

          <Dialog open={reviewDialogOpen} onOpenChange={setReviewDialogOpen}>
            <DialogContent className="max-w-lg">
              <DialogHeader>
                <DialogTitle>Write a review</DialogTitle>
                <DialogDescription>
                  Share your experience. Your star rating will also contribute to the function’s
                  community score.
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-4">
                <div>
                  <div className="text-sm font-medium text-foreground mb-2">Rating</div>
                  <div className="flex items-center gap-1">
                    {Array.from({ length: 5 }).map((_, i) => {
                      const val = i + 1;
                      const active = val <= reviewStars;
                      return (
                        <button
                          key={val}
                          type="button"
                          onClick={() => setReviewStars(val)}
                          className="p-1 rounded-md hover:bg-bg-tertiary transition-colors"
                          aria-label={`${val} star${val === 1 ? '' : 's'}`}
                        >
                          <Star
                            className={`h-5 w-5 ${
                              active ? 'fill-amber-400 text-amber-400' : 'text-muted-foreground/50'
                            }`}
                          />
                        </button>
                      );
                    })}
                    <span className="ml-2 text-sm text-muted-foreground">{reviewStars}/5</span>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="text-sm font-medium text-foreground">Title (optional)</div>
                  <Input
                    value={reviewTitle}
                    onChange={(e) => setReviewTitle(e.target.value)}
                    placeholder="Short summary"
                    maxLength={120}
                  />
                </div>

                <div className="space-y-2">
                  <div className="text-sm font-medium text-foreground">Review (optional)</div>
                  <Textarea
                    value={reviewBody}
                    onChange={(e) => setReviewBody(e.target.value)}
                    placeholder="What worked well? Anything to watch out for?"
                    maxLength={5000}
                  />
                </div>
              </div>

              <DialogFooter>
                <Button variant="outline" onClick={() => setReviewDialogOpen(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={() => submitReviewMutation.mutate()}
                  disabled={submitReviewMutation.isPending}
                >
                  {submitReviewMutation.isPending ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Submitting…
                    </>
                  ) : (
                    'Submit'
                  )}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Auto-generated Documentation — directly under Try it Now */}
          <motion.div
            id="function-api-reference"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.35 }}
            className="function-page-api-reference scroll-mt-24 mb-8 mt-8 overflow-hidden rounded-2xl border border-border-subtle bg-linear-to-b from-bg-primary via-bg-primary/95 to-bg-secondary/30 shadow-sm"
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
              <div className="border-b border-border-subtle bg-bg-primary/20 px-4 py-4 lg:border-b-0 lg:border-r lg:py-6">
                <nav
                  className="flex flex-wrap gap-1 lg:flex-col"
                  aria-label="Documentation sections"
                >
                  <a
                    href="#doc-overview"
                    className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-bg-tertiary hover:text-foreground"
                  >
                    Overview
                  </a>
                  {(functionInfo.manifest?.input || functionInfo.input_example != null) && (
                    <a
                      href="#doc-input"
                      className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-bg-tertiary hover:text-foreground"
                    >
                      Input
                    </a>
                  )}
                  {(functionInfo.manifest?.output || functionInfo.output_example != null) && (
                    <a
                      href="#doc-output"
                      className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-bg-tertiary hover:text-foreground"
                    >
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
                        <span className="flex h-6 w-6 items-center justify-center rounded bg-brand-500/15 text-brand-500 text-xs">
                          1
                        </span>
                        Overview
                      </h3>
                      <p className="mb-4 text-sm leading-relaxed text-muted-foreground">
                        {functionInfo.description || 'No description provided.'}
                      </p>
                      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                        <div className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-3 py-2.5">
                          <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                            Runtime
                          </p>
                          <p className="mt-0.5 font-mono text-sm text-foreground">
                            {functionInfo.runtime}
                          </p>
                        </div>
                        {functionInfo.manifest?.deterministic != null && (
                          <div className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-3 py-2.5">
                            <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                              Determinism
                            </p>
                            <p className="mt-0.5 text-sm text-foreground">
                              {functionInfo.manifest.deterministic
                                ? 'Deterministic'
                                : 'Non-deterministic'}
                            </p>
                          </div>
                        )}
                        {functionInfo.cache_ttl != null && functionInfo.cache_ttl > 0 && (
                          <div className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-3 py-2.5">
                            <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                              Cache TTL
                            </p>
                            <p className="mt-0.5 text-sm text-foreground">
                              {functionInfo.cache_ttl}s
                            </p>
                          </div>
                        )}
                      </div>
                    </div>
                  </section>

                  {(functionInfo.manifest?.input || functionInfo.input_example != null) && (
                    <section id="doc-input" className="scroll-mt-6 pt-6">
                      <div className="rounded-xl border border-border-subtle bg-bg-primary/80 p-5 shadow-sm">
                        <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wider text-foreground">
                          <span className="flex h-6 w-6 items-center justify-center rounded bg-emerald-500/15 text-emerald-600 text-xs dark:text-emerald-400">
                            2
                          </span>
                          Input
                        </h3>
                        {functionInfo.manifest?.input?.properties &&
                          typeof functionInfo.manifest.input.properties === 'object' && (
                            <div className="mb-4 overflow-hidden rounded-lg border border-border-subtle">
                              <table className="w-full text-left text-sm">
                                <thead>
                                  <tr className="border-b border-border-subtle bg-bg-tertiary/70">
                                    <th className="px-4 py-2.5 font-medium text-foreground">
                                      Property
                                    </th>
                                    <th className="px-4 py-2.5 font-medium text-foreground">
                                      Type
                                    </th>
                                    <th className="px-4 py-2.5 font-medium text-foreground">
                                      Description
                                    </th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {Object.entries(functionInfo.manifest.input.properties).map(
                                    ([key, val]: [string, unknown]) => {
                                      const v = val as {
                                        type?: string;
                                        description?: string;
                                        default?: unknown;
                                      };
                                      return (
                                        <tr
                                          key={key}
                                          className="border-b border-border-subtle/50 last:border-0"
                                        >
                                          <td className="px-4 py-2 font-mono text-xs text-foreground">
                                            {key}
                                          </td>
                                          <td className="px-4 py-2 text-muted-foreground">
                                            {v?.type ?? '—'}
                                          </td>
                                          <td className="px-4 py-2 text-muted-foreground">
                                            {v?.description ?? '—'}
                                          </td>
                                        </tr>
                                      );
                                    }
                                  )}
                                </tbody>
                              </table>
                            </div>
                          )}
                        <Tabs defaultValue="schema" className="w-full">
                          <TabsList className="mb-2 h-9">
                            <TabsTrigger value="schema" className="text-xs">
                              Schema
                            </TabsTrigger>
                            {functionInfo.input_example != null && (
                              <TabsTrigger value="example" className="text-xs">
                                Example
                              </TabsTrigger>
                            )}
                          </TabsList>
                          <TabsContent value="schema" className="mt-0">
                            {functionInfo.manifest?.input ? (
                              <CodeBlock
                                code={JSON.stringify(functionInfo.manifest.input, null, 2)}
                                language="json"
                              />
                            ) : (
                              <p className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-4 py-3 text-sm text-muted-foreground">
                                No schema defined.
                              </p>
                            )}
                          </TabsContent>
                          {functionInfo.input_example != null && (
                            <TabsContent value="example" className="mt-0">
                              <CodeBlock
                                code={
                                  typeof functionInfo.input_example === 'string'
                                    ? functionInfo.input_example
                                    : JSON.stringify(functionInfo.input_example, null, 2)
                                }
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
                          <span className="flex h-6 w-6 items-center justify-center rounded bg-violet-500/15 text-violet-600 text-xs dark:text-violet-400">
                            3
                          </span>
                          Output
                        </h3>
                        {functionInfo.manifest?.output?.properties &&
                          typeof functionInfo.manifest.output.properties === 'object' && (
                            <div className="mb-4 overflow-hidden rounded-lg border border-border-subtle">
                              <table className="w-full text-left text-sm">
                                <thead>
                                  <tr className="border-b border-border-subtle bg-bg-tertiary/70">
                                    <th className="px-4 py-2.5 font-medium text-foreground">
                                      Property
                                    </th>
                                    <th className="px-4 py-2.5 font-medium text-foreground">
                                      Type
                                    </th>
                                    <th className="px-4 py-2.5 font-medium text-foreground">
                                      Description
                                    </th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {Object.entries(functionInfo.manifest.output.properties).map(
                                    ([key, val]: [string, unknown]) => {
                                      const v = val as { type?: string; description?: string };
                                      return (
                                        <tr
                                          key={key}
                                          className="border-b border-border-subtle/50 last:border-0"
                                        >
                                          <td className="px-4 py-2 font-mono text-xs text-foreground">
                                            {key}
                                          </td>
                                          <td className="px-4 py-2 text-muted-foreground">
                                            {v?.type ?? '—'}
                                          </td>
                                          <td className="px-4 py-2 text-muted-foreground">
                                            {v?.description ?? '—'}
                                          </td>
                                        </tr>
                                      );
                                    }
                                  )}
                                </tbody>
                              </table>
                            </div>
                          )}
                        <Tabs defaultValue="schema" className="w-full">
                          <TabsList className="mb-2 h-9">
                            <TabsTrigger value="schema" className="text-xs">
                              Schema
                            </TabsTrigger>
                            {functionInfo.output_example != null && (
                              <TabsTrigger value="example" className="text-xs">
                                Example
                              </TabsTrigger>
                            )}
                          </TabsList>
                          <TabsContent value="schema" className="mt-0">
                            {functionInfo.manifest?.output ? (
                              <CodeBlock
                                code={JSON.stringify(functionInfo.manifest.output, null, 2)}
                                language="json"
                              />
                            ) : (
                              <p className="rounded-lg border border-border-subtle bg-bg-tertiary/50 px-4 py-3 text-sm text-muted-foreground">
                                No schema defined.
                              </p>
                            )}
                          </TabsContent>
                          {functionInfo.output_example != null && (
                            <TabsContent value="example" className="mt-0">
                              <CodeBlock
                                code={
                                  typeof functionInfo.output_example === 'string'
                                    ? functionInfo.output_example
                                    : JSON.stringify(functionInfo.output_example, null, 2)
                                }
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
                      <CardDescription>Expected input structure</CardDescription>
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
                      <CardDescription>Expected output structure</CardDescription>
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
          <Card className="bg-linear-to-br from-brand-500/5 to-transparent border-dashed">
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
