import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Search, Star, Download, Code, Filter, SortAsc, SortDesc, Grid, List } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { useRealtime } from "@/hooks/useRealtime";
import { apiClient } from "@/api/client";

interface RegistryFunction {
  id: string;
  author: string;
  name: string;
  title?: string;
  description?: string;
  category?: string;
  tags: string[];
  visibility: string;
  price_per_call: number;
  popularity_score: number;
  reliability_score: number;
  deterministic_score: number;
  latest_version?: string;
  total_ratings: number;
  overall_score: number;
  created_at: string;
}

interface RegistryFunctionVersion {
  id: string;
  version: string;
  manifest: any;
  runtime: string;
  timeout_ms: number;
  memory_mb: number;
  deterministic: boolean;
  cache_ttl: number;
  published_at: string;
}

const categories = [
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

const runtimes = [
  "All Runtimes",
  "node18",
  "node20",
  "python3.9",
  "python3.10",
  "python3.11",
  "deno",
  "go",
  "rust",
];

export default function RegistryPage() {
  const navigate = useNavigate();
  const { subscribe, unsubscribe } = useRealtime();

  const [functions, setFunctions] = useState<RegistryFunction[]>([]);
  const [filteredFunctions, setFilteredFunctions] = useState<RegistryFunction[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All Categories");
  const [selectedRuntime, setSelectedRuntime] = useState("All Runtimes");
  const [sortBy, setSortBy] = useState("popularity");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [showFilters, setShowFilters] = useState(false);

  // Real-time updates for registry changes
  useEffect(() => {
    const handleRegistryUpdate = (data: any) => {
      if (data.event === "registry_update" && data.update_type === "rating") {
        // Update the function's rating in the list
        setFunctions(prev => prev.map(fn =>
          fn.id === data.function_id
            ? { ...fn, overall_score: data.overall_score, total_ratings: data.total_ratings }
            : fn
        ));
      }
    };

    subscribe("registry_updates", handleRegistryUpdate);

    return () => {
      unsubscribe("registry_updates", handleRegistryUpdate);
    };
  }, [subscribe, unsubscribe]);

  // Load functions from registry
  useEffect(() => {
    loadRegistryFunctions();
  }, []);

  // Filter and sort functions
  useEffect(() => {
    let filtered = functions.filter(fn => {
      const matchesSearch = searchQuery === "" ||
        fn.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        fn.author.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (fn.description && fn.description.toLowerCase().includes(searchQuery.toLowerCase()));

      const matchesCategory = selectedCategory === "All Categories" || fn.category === selectedCategory;

      return matchesSearch && matchesCategory;
    });

    // Sort functions
    filtered.sort((a, b) => {
      let aValue: any, bValue: any;

      switch (sortBy) {
        case "popularity":
          aValue = a.popularity_score;
          bValue = b.popularity_score;
          break;
        case "rating":
          aValue = a.overall_score;
          bValue = b.overall_score;
          break;
        case "reliability":
          aValue = a.reliability_score;
          bValue = b.reliability_score;
          break;
        case "newest":
          aValue = new Date(a.created_at);
          bValue = new Date(b.created_at);
          break;
        case "name":
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
          break;
        default:
          return 0;
      }

      if (sortOrder === "asc") {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
      }
    });

    setFilteredFunctions(filtered);
  }, [functions, searchQuery, selectedCategory, sortBy, sortOrder]);

  const loadRegistryFunctions = async () => {
    try {
      setLoading(true);
      const response = await apiClient.get<{ functions: RegistryFunction[] }>("/v1/registry/functions");
      setFunctions(response.functions || []);
    } catch (error) {
      console.error("Failed to load registry functions:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleDeployFunction = async (fn: RegistryFunction) => {
    // Navigate to deploy page with function details
    navigate(`/functions/deploy?registry=${fn.author}/${fn.name}`);
  };

  const handleViewFunction = (fn: RegistryFunction) => {
    navigate(`/registry/${fn.author}/${fn.name}`);
  };

  const toggleSortOrder = () => {
    setSortOrder(sortOrder === "asc" ? "desc" : "asc");
  };

  const formatScore = (score: number) => {
    return score.toFixed(1);
  };

  const renderFunctionCard = (fn: RegistryFunction) => {
    if (viewMode === "list") {
      return (
        <Card key={fn.id} className="mb-2">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-3">
                  <Code className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <h3 className="font-semibold">{fn.author}/{fn.name}</h3>
                    <p className="text-sm text-muted-foreground">{fn.title || fn.description}</p>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-4">
                <div className="text-center">
                  <div className="flex items-center gap-1">
                    <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                    <span className="font-medium">{formatScore(fn.overall_score)}</span>
                  </div>
                  <p className="text-xs text-muted-foreground">({fn.total_ratings})</p>
                </div>

                <div className="text-center">
                  <Download className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
                  <p className="text-xs text-muted-foreground">{Math.floor(fn.popularity_score)}</p>
                </div>

                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleViewFunction(fn)}
                  >
                    View
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => handleDeployFunction(fn)}
                  >
                    Deploy
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      );
    }

    return (
      <Card key={fn.id} className="h-full">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-2">
              <Code className="h-5 w-5 text-muted-foreground" />
              <div>
                <CardTitle className="text-lg">{fn.name}</CardTitle>
                <p className="text-sm text-muted-foreground">by {fn.author}</p>
              </div>
            </div>
            {fn.category && (
              <Badge variant="secondary">{fn.category}</Badge>
            )}
          </div>
        </CardHeader>

        <CardContent className="flex-1 flex flex-col">
          <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
            {fn.description || "No description available"}
          </p>

          {fn.tags && fn.tags.length > 0 && (
            <div className="flex flex-wrap gap-1 mb-4">
              {fn.tags.slice(0, 3).map((tag, index) => (
                <Badge key={index} variant="outline" className="text-xs">
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
                <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                <span className="font-medium">{formatScore(fn.overall_score)}</span>
                <span className="text-muted-foreground">({fn.total_ratings})</span>
              </div>

              <div className="flex items-center gap-1">
                <Download className="h-4 w-4 text-muted-foreground" />
                <span className="text-muted-foreground">{Math.floor(fn.popularity_score)}</span>
              </div>
            </div>

            {fn.price_per_call > 0 && (
              <Badge variant="outline">
                ${fn.price_per_call}/call
              </Badge>
            )}
          </div>

          <div className="flex gap-2 mt-auto">
            <Button
              variant="outline"
              className="flex-1"
              onClick={() => handleViewFunction(fn)}
            >
              View Details
            </Button>
            <Button
              className="flex-1"
              onClick={() => handleDeployFunction(fn)}
            >
              Deploy
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  };

  return (
    <div className="container mx-auto px-6 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Function Registry</h1>
        <p className="text-muted-foreground">
          Browse and deploy functions from the community registry
        </p>
      </div>

      {/* Search and Filters */}
      <div className="mb-6 space-y-4">
        <div className="flex gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search functions..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          <Button
            variant="outline"
            onClick={() => setShowFilters(!showFilters)}
            className="flex items-center gap-2"
          >
            <Filter className="h-4 w-4" />
            Filters
          </Button>

          <div className="flex items-center gap-2">
            <Button
              variant={viewMode === "grid" ? "default" : "outline"}
              size="sm"
              onClick={() => setViewMode("grid")}
            >
              <Grid className="h-4 w-4" />
            </Button>
            <Button
              variant={viewMode === "list" ? "default" : "outline"}
              size="sm"
              onClick={() => setViewMode("list")}
            >
              <List className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {showFilters && (
          <Card>
            <CardContent className="p-4">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <Label htmlFor="category">Category</Label>
                  <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {categories.map((category) => (
                        <SelectItem key={category} value={category}>
                          {category}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div>
                  <Label htmlFor="runtime">Runtime</Label>
                  <Select value={selectedRuntime} onValueChange={setSelectedRuntime}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {runtimes.map((runtime) => (
                        <SelectItem key={runtime} value={runtime}>
                          {runtime}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div>
                  <Label htmlFor="sort">Sort By</Label>
                  <div className="flex gap-2">
                    <Select value={sortBy} onValueChange={setSortBy}>
                      <SelectTrigger className="flex-1">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="popularity">Popularity</SelectItem>
                        <SelectItem value="rating">Rating</SelectItem>
                        <SelectItem value="reliability">Reliability</SelectItem>
                        <SelectItem value="newest">Newest</SelectItem>
                        <SelectItem value="name">Name</SelectItem>
                      </SelectContent>
                    </Select>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={toggleSortOrder}
                    >
                      {sortOrder === "asc" ? <SortAsc className="h-4 w-4" /> : <SortDesc className="h-4 w-4" />}
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Results */}
      <div className="mb-4">
        <p className="text-sm text-muted-foreground">
          {loading ? "Loading..." : `${filteredFunctions.length} functions found`}
        </p>
      </div>

      {loading ? (
        <div className={viewMode === "grid" ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" : "space-y-2"}>
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i}>
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
        <div className={
          viewMode === "grid"
            ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
            : "space-y-2"
        }>
          {filteredFunctions.map(renderFunctionCard)}
        </div>
      )}

      {filteredFunctions.length === 0 && !loading && (
        <div className="text-center py-12">
          <Code className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">No functions found</h3>
          <p className="text-muted-foreground">
            Try adjusting your search or filters
          </p>
        </div>
      )}
    </div>
  );
}
