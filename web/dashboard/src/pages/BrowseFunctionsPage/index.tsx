import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Search,
  Star,
  Download,
  Code,
  Filter,
  SortAsc,
  SortDesc,
  Grid,
  List,
  Play,
  LogIn,
  Zap,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Navbar } from "@/components/common/Navbar";
import { Footer } from "@/pages/LandingPage/components";
import { useAuthStore } from "@/stores/authStore";
import { registryApi, type RegistryFunction } from "@/api/registry";
import { useRealtime } from "@/hooks/useRealtime";
import { MetaTags } from "@/components/seo/MetaTags";

const CATEGORIES = [
  "All Categories",
  "API Tools",
  "Authentication",
  "Database",
  "Email",
  "File Processing",
  "Image Processing",
  "Machine Learning",
  "Payment",
  "Utility",
  "Web Scraping",
  "Workflow",
];

const SORT_OPTIONS = [
  { value: "popularity", label: "Popularity" },
  { value: "rating", label: "Rating" },
  { value: "reliability", label: "Reliability" },
  { value: "newest", label: "Newest" },
  { value: "name", label: "Name" },
];

export function BrowseFunctionsPage() {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const { subscribe, unsubscribe } = useRealtime();

  const [functions, setFunctions] = useState<RegistryFunction[]>([]);
  const [filteredFunctions, setFilteredFunctions] = useState<RegistryFunction[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All Categories");
  const [sortBy, setSortBy] = useState("popularity");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [showFilters, setShowFilters] = useState(false);

  useEffect(() => {
    const handleRegistryUpdate = (data: { event?: string; update_type?: string; function_id?: string; overall_score?: number; total_ratings?: number }) => {
      if (data.event === "registry_update" && data.update_type === "rating") {
        setFunctions((prev) =>
          prev.map((fn) =>
            fn.id === data.function_id
              ? {
                  ...fn,
                  overall_score: data.overall_score ?? fn.overall_score,
                  total_ratings: data.total_ratings ?? fn.total_ratings,
                }
              : fn
          )
        );
      }
    };
    subscribe("registry_updates", handleRegistryUpdate);
    return () => unsubscribe("registry_updates", handleRegistryUpdate);
  }, [subscribe, unsubscribe]);

  useEffect(() => {
    loadFunctions();
  }, []);

  useEffect(() => {
    let filtered = functions.filter((fn) => {
      const matchesSearch =
        searchQuery === "" ||
        fn.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        fn.author.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (fn.description?.toLowerCase().includes(searchQuery.toLowerCase()) ?? false);
      const matchesCategory =
        selectedCategory === "All Categories" || fn.category === selectedCategory;
      return matchesSearch && matchesCategory;
    });

    filtered.sort((a, b) => {
      let aVal: number | string | Date, bVal: number | string | Date;
      switch (sortBy) {
        case "popularity":
          aVal = a.popularity_score;
          bVal = b.popularity_score;
          break;
        case "rating":
          aVal = a.overall_score;
          bVal = b.overall_score;
          break;
        case "reliability":
          aVal = a.reliability_score;
          bVal = b.reliability_score;
          break;
        case "newest":
          aVal = new Date(a.created_at).getTime();
          bVal = new Date(b.created_at).getTime();
          break;
        case "name":
          aVal = a.name.toLowerCase();
          bVal = b.name.toLowerCase();
          break;
        default:
          return 0;
      }
      if (sortOrder === "asc") {
        return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
      }
      return aVal > bVal ? -1 : aVal < bVal ? 1 : 0;
    });
    setFilteredFunctions(filtered);
  }, [functions, searchQuery, selectedCategory, sortBy, sortOrder]);

  const loadFunctions = async () => {
    try {
      setLoading(true);
      const res = await registryApi.getFunctions({ limit: 500 });
      setFunctions(res.functions ?? []);
    } catch (e) {
      console.error("Failed to load registry functions:", e);
    } finally {
      setLoading(false);
    }
  };

  const getLoginRedirect = (path: string) => {
    return `/login?redirect=${encodeURIComponent(path)}`;
  };

  const handleDeploy = (fn: RegistryFunction) => {
    const path = `/functions/deploy?registry=${fn.author}/${fn.name}`;
    if (!isAuthenticated) {
      navigate(getLoginRedirect(path));
      return;
    }
    navigate(path);
  };

  const handleTryPlayground = (fn: RegistryFunction) => {
    const path = `/run/${fn.author}/${fn.name}`;
    if (!isAuthenticated) {
      navigate(getLoginRedirect(path));
      return;
    }
    navigate(path);
  };

  const handleViewDetails = (fn: RegistryFunction) => {
    navigate(`/fx/${fn.author}/${fn.name}`);
  };

  const formatScore = (score: number | undefined | null) =>
    typeof score === "number" && !Number.isNaN(score) ? score.toFixed(1) : "0.0";

  const renderCard = (fn: RegistryFunction) => {
    const actionSection = (
      <div className="flex flex-wrap gap-2 mt-auto">
        <Button
          variant="outline"
          size="sm"
          className="flex-1 min-w-[100px]"
          onClick={() => handleViewDetails(fn)}
        >
          View
        </Button>
        <Button
          size="sm"
          className="flex-1 min-w-[100px]"
          onClick={() => handleDeploy(fn)}
        >
          {isAuthenticated ? (
            <>Deploy</>
          ) : (
            <>
              <LogIn className="h-3.5 w-3.5 mr-1" />
              Deploy
            </>
          )}
        </Button>
        <Button
          variant="secondary"
          size="sm"
          className="flex-1 min-w-[100px]"
          onClick={() => handleTryPlayground(fn)}
        >
          {isAuthenticated ? (
            <>
              <Play className="h-3.5 w-3.5 mr-1" />
              Try
            </>
          ) : (
            <>
              <LogIn className="h-3.5 w-3.5 mr-1" />
              Try
            </>
          )}
        </Button>
      </div>
    );

    if (viewMode === "list") {
      return (
        <Card key={fn.id} className="border-border-subtle bg-bg-secondary/50 hover:bg-bg-tertiary/50 transition-colors">
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3">
                  <div className="shrink-0 w-10 h-10 rounded-lg bg-brand-500/10 border border-brand-500/20 flex items-center justify-center">
                    <Code className="h-5 w-5 text-brand-500" />
                  </div>
                  <div className="min-w-0">
                    <h3 className="font-semibold text-text-primary truncate">
                      {fn.author}/{fn.name}
                    </h3>
                    <p className="text-sm text-text-secondary line-clamp-1">
                      {fn.title || fn.description || "No description"}
                    </p>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-4 shrink-0">
                <div className="flex items-center gap-1 text-sm">
                  <Star className="h-4 w-4 fill-amber-400 text-amber-400" />
                  <span className="font-medium text-text-primary">{formatScore(fn.overall_score)}</span>
                  <span className="text-text-muted">({fn.total_ratings ?? 0})</span>
                </div>
                <div className="flex items-center gap-1 text-sm text-text-muted">
                  <Download className="h-4 w-4" />
                  <span>{Math.floor(Number(fn.popularity_score) || 0)}</span>
                </div>
                {actionSection}
              </div>
            </div>
          </CardContent>
        </Card>
      );
    }

    return (
      <Card
        key={fn.id}
        className="h-full flex flex-col border-border-subtle bg-bg-secondary/50 hover:bg-bg-tertiary/50 transition-colors overflow-hidden"
      >
        <CardHeader className="pb-2">
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className="shrink-0 w-10 h-10 rounded-lg bg-brand-500/10 border border-brand-500/20 flex items-center justify-center">
                <Code className="h-5 w-5 text-brand-500" />
              </div>
              <div className="min-w-0">
                <CardTitle className="text-lg truncate">{fn.name}</CardTitle>
                <p className="text-sm text-text-secondary">by {fn.author}</p>
              </div>
            </div>
            {fn.category && (
              <Badge variant="secondary" className="shrink-0">
                {fn.category}
              </Badge>
            )}
          </div>
        </CardHeader>
        <CardContent className="flex-1 flex flex-col pt-0">
          <p className="text-sm text-text-secondary mb-4 line-clamp-2">
            {fn.description || fn.title || "No description available"}
          </p>
          {fn.tags && fn.tags.length > 0 && (
            <div className="flex flex-wrap gap-1 mb-4">
              {fn.tags.slice(0, 3).map((tag, i) => (
                <Badge key={i} variant="outline" className="text-xs">
                  {tag}
                </Badge>
              ))}
              {fn.tags.length > 3 && (
                <Badge variant="outline" className="text-xs">
                  +{fn.tags.length - 3}
                </Badge>
              )}
            </div>
          )}
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-4 text-sm">
              <div className="flex items-center gap-1">
                <Star className="h-4 w-4 fill-amber-400 text-amber-400" />
                <span className="font-medium">{formatScore(fn.overall_score)}</span>
                <span className="text-text-muted">({fn.total_ratings ?? 0})</span>
              </div>
              <div className="flex items-center gap-1 text-text-muted">
                <Download className="h-4 w-4" />
                <span>{Math.floor(Number(fn.popularity_score) || 0)}</span>
              </div>
            </div>
            {Number(fn.price_per_call) > 0 && (
              <Badge variant="outline">${fn.price_per_call}/call</Badge>
            )}
          </div>
          {actionSection}
        </CardContent>
      </Card>
    );
  };

  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      <MetaTags
        title="Browse Functions | FunctionFly Registry"
        description="Browse and discover serverless functions from the FunctionFly registry. Deploy or try functions—sign in to use them."
        keywords={["function registry", "serverless", "browse functions", "deploy functions"]}
      />
      <Navbar variant="landing" />

      <main className="flex-1 pt-16">
        {/* Hero */}
        <section className="section-padding border-b border-border-subtle bg-gradient-radial">
          <div className="container-wide px-4 lg:px-6">
            <div className="max-w-3xl">
              <div className="flex items-center gap-2 text-brand-500 mb-3">
                <Zap className="h-5 w-5" />
                <span className="text-sm font-medium uppercase tracking-wider">Registry</span>
              </div>
              <h1 className="text-4xl md:text-5xl font-bold text-text-primary mb-4 text-balance">
                Browse functions
              </h1>
              <p className="text-lg text-text-secondary mb-6">
                Discover and explore serverless functions from the community. View docs and examples
                anytime; sign in to deploy or run functions.
              </p>
              {!isAuthenticated && (
                <div className="flex flex-wrap items-center gap-2 p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-700 dark:text-amber-400">
                  <LogIn className="h-4 w-4 shrink-0" />
                  <span className="text-sm">
                    You need an account to deploy or run functions.{" "}
                    <button
                      type="button"
                      onClick={() => navigate("/login")}
                      className="font-medium underline hover:no-underline"
                    >
                      Sign in
                    </button>{" "}
                    or{" "}
                    <button
                      type="button"
                      onClick={() => navigate("/signup")}
                      className="font-medium underline hover:no-underline"
                    >
                      sign up
                    </button>
                    .
                  </span>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Filters & list */}
        <section className="section-padding">
          <div className="container-wide px-4 lg:px-6">
            <div className="mb-6 space-y-4">
              <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
                  <Input
                    placeholder="Search by name, author, or description..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10 bg-bg-secondary border-border-subtle"
                  />
                </div>
                <Button
                  variant="outline"
                  onClick={() => setShowFilters(!showFilters)}
                  className="flex items-center gap-2 shrink-0"
                >
                  <Filter className="h-4 w-4" />
                  Filters
                </Button>
                <div className="flex items-center gap-2 shrink-0">
                  <Button
                    variant={viewMode === "grid" ? "default" : "outline"}
                    size="icon"
                    onClick={() => setViewMode("grid")}
                    aria-label="Grid view"
                  >
                    <Grid className="h-4 w-4" />
                  </Button>
                  <Button
                    variant={viewMode === "list" ? "default" : "outline"}
                    size="icon"
                    onClick={() => setViewMode("list")}
                    aria-label="List view"
                  >
                    <List className="h-4 w-4" />
                  </Button>
                </div>
              </div>

              {showFilters && (
                <Card className="border-border-subtle bg-bg-secondary/50">
                  <CardContent className="p-4">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                      <div>
                        <Label className="text-text-secondary">Category</Label>
                        <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                          <SelectTrigger className="mt-1 bg-bg-primary border-border-subtle">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {CATEGORIES.map((c) => (
                              <SelectItem key={c} value={c}>
                                {c}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <div>
                        <Label className="text-text-secondary">Sort by</Label>
                        <div className="flex gap-2 mt-1">
                          <Select value={sortBy} onValueChange={setSortBy}>
                            <SelectTrigger className="flex-1 bg-bg-primary border-border-subtle">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {SORT_OPTIONS.map((o) => (
                                <SelectItem key={o.value} value={o.value}>
                                  {o.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <Button
                            variant="outline"
                            size="icon"
                            onClick={() => setSortOrder(sortOrder === "asc" ? "desc" : "asc")}
                            aria-label={sortOrder === "asc" ? "Ascending" : "Descending"}
                          >
                            {sortOrder === "asc" ? (
                              <SortAsc className="h-4 w-4" />
                            ) : (
                              <SortDesc className="h-4 w-4" />
                            )}
                          </Button>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>

            <p className="text-sm text-text-muted mb-4">
              {loading ? "Loading…" : `${filteredFunctions.length} function${filteredFunctions.length !== 1 ? "s" : ""} found`}
            </p>

            {loading ? (
              <div
                className={
                  viewMode === "grid"
                    ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
                    : "space-y-2"
                }
              >
                {Array.from({ length: 6 }).map((_, i) => (
                  <Card key={i} className="border-border-subtle overflow-hidden">
                    <CardHeader>
                      <Skeleton className="h-6 w-3/4" />
                      <Skeleton className="h-4 w-1/2" />
                    </CardHeader>
                    <CardContent>
                      <Skeleton className="h-4 w-full mb-2" />
                      <Skeleton className="h-4 w-2/3 mb-4" />
                      <div className="flex gap-2">
                        <Skeleton className="h-8 flex-1" />
                        <Skeleton className="h-8 flex-1" />
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : (
              <div
                className={
                  viewMode === "grid"
                    ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
                    : "space-y-2"
                }
              >
                {filteredFunctions.map(renderCard)}
              </div>
            )}

            {!loading && filteredFunctions.length === 0 && (
              <div className="text-center py-16">
                <Code className="h-14 w-14 text-text-muted mx-auto mb-4 opacity-50" />
                <h3 className="text-xl font-semibold text-text-primary mb-2">No functions found</h3>
                <p className="text-text-secondary max-w-md mx-auto">
                  Try a different search or category, or check back later for new additions.
                </p>
              </div>
            )}
          </div>
        </section>
      </main>

      <Footer />
    </div>
  );
}
