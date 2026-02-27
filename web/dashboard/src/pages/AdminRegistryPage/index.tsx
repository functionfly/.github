import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Search,
  MoreVertical,
  Code,
  Trash2,
  RefreshCw,
  Eye,
  EyeOff,
  DollarSign,
  Flag,
  Star,
  Download,
  Shield,
  ArrowLeft,
  AlertTriangle,
  CheckCircle,
  Loader2,
} from "lucide-react";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Skeleton } from "@/components/ui/skeleton";
import { StatCard } from "@/components/common/StatCard";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import {
  adminRegistryApi,
  type AdminRegistryFunction,
  type AdminRegistryFunctionVersion,
} from "@/api/admin";

const visibilityColors = {
  public: "bg-emerald-500/10 text-emerald-400",
  private: "bg-blue-500/10 text-blue-400",
  unlisted: "bg-yellow-500/10 text-yellow-400",
};

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

export function AdminRegistryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState("");
  const [visibilityFilter, setVisibilityFilter] = useState<string>("all");
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [flaggedFilter, setFlaggedFilter] = useState<string>("all");
  const [selectedFunction, setSelectedFunction] = useState<AdminRegistryFunction | null>(null);
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [functionToDelete, setFunctionToDelete] = useState<AdminRegistryFunction | null>(null);
  const [isFlagDialogOpen, setIsFlagDialogOpen] = useState(false);
  const [flagReason, setFlagReason] = useState("");
  const [functionToFlag, setFunctionToFlag] = useState<AdminRegistryFunction | null>(null);
  const [isPricingDialogOpen, setIsPricingDialogOpen] = useState(false);
  const [newPrice, setNewPrice] = useState("");
  const [functionForPricing, setFunctionForPricing] = useState<AdminRegistryFunction | null>(null);

  // Fetch stats from admin registry API
  const { data: statsData } = useQuery({
    queryKey: ["admin-registry-stats"],
    queryFn: () => adminRegistryApi.getStats(),
  });

  // Fetch registry functions from admin API
  const { data: functionsData, isLoading: functionsLoading, refetch } = useQuery({
    queryKey: ["admin-registry-functions", visibilityFilter, categoryFilter, flaggedFilter, searchTerm],
    queryFn: () =>
      adminRegistryApi.listFunctions({
        visibility: visibilityFilter !== "all" ? visibilityFilter : undefined,
        category: categoryFilter !== "All Categories" && categoryFilter !== "all" ? categoryFilter : undefined,
        flagged: flaggedFilter === "flagged" ? true : flaggedFilter === "unflagged" ? false : undefined,
        search: searchTerm || undefined,
        limit: 100,
      }),
  });

  // Fetch function details by ID
  const { data: functionDetails, isLoading: functionDetailsLoading } = useQuery({
    queryKey: ["admin-registry-function-details", selectedFunction?.id],
    queryFn: () => adminRegistryApi.getFunction(selectedFunction!.id),
    enabled: !!selectedFunction && isDetailsOpen,
  });

  const functions = functionsData?.functions ?? [];
  const totalFunctions = functionsData?.total ?? 0;
  const stats = statsData ?? {
    total_functions: 0,
    public_functions: 0,
    private_functions: 0,
    unlisted_functions: 0,
    flagged_functions: 0,
    total_calls: 0,
    total_revenue: 0,
    avg_rating: 0,
  };

  // Update visibility mutation
  const updateVisibilityMutation = useMutation({
    mutationFn: ({ functionId, visibility }: { functionId: string; visibility: "public" | "private" | "unlisted" }) =>
      adminRegistryApi.updateVisibility(functionId, visibility),
    onSuccess: () => {
      toast.success("Visibility updated");
      queryClient.invalidateQueries({ queryKey: ["admin-registry-functions"] });
    },
    onError: () => {
      toast.error("Failed to update visibility");
    },
  });

  // Update pricing mutation
  const updatePricingMutation = useMutation({
    mutationFn: ({ functionId, price }: { functionId: string; price: number }) =>
      adminRegistryApi.updatePricing(functionId, price),
    onSuccess: () => {
      toast.success("Pricing updated");
      queryClient.invalidateQueries({ queryKey: ["admin-registry-functions"] });
      setIsPricingDialogOpen(false);
      setFunctionForPricing(null);
      setNewPrice("");
    },
    onError: () => {
      toast.error("Failed to update pricing");
    },
  });

  // Toggle flag mutation
  const toggleFlagMutation = useMutation({
    mutationFn: ({ functionId, flagged, reason }: { functionId: string; flagged: boolean; reason?: string }) =>
      adminRegistryApi.toggleFlag(functionId, flagged, reason),
    onSuccess: () => {
      toast.success(functionToFlag?.is_flagged ? "Function unflagged" : "Function flagged");
      queryClient.invalidateQueries({ queryKey: ["admin-registry-functions"] });
      setIsFlagDialogOpen(false);
      setFunctionToFlag(null);
      setFlagReason("");
    },
    onError: () => {
      toast.error("Failed to update flag status");
    },
  });

  // Delete function mutation
  const deleteFunctionMutation = useMutation({
    mutationFn: (functionId: string) => adminRegistryApi.deleteFunction(functionId),
    onSuccess: () => {
      toast.success("Function deleted successfully");
      queryClient.invalidateQueries({ queryKey: ["admin-registry-functions"] });
      setIsDeleteDialogOpen(false);
      setFunctionToDelete(null);
    },
    onError: () => {
      toast.error("Failed to delete function");
    },
  });

  const filteredFunctions = functions.filter((fn) => {
    const matchesSearch =
      !searchTerm ||
      fn.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      fn.author.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (fn.description && fn.description.toLowerCase().includes(searchTerm.toLowerCase()));
    return matchesSearch;
  });

  const handleViewDetails = (fn: AdminRegistryFunction) => {
    setSelectedFunction(fn);
    setIsDetailsOpen(true);
  };

  const handleVisibilityChange = (fn: AdminRegistryFunction, visibility: "public" | "private" | "unlisted") => {
    updateVisibilityMutation.mutate({ functionId: fn.id, visibility });
  };

  const handlePricingClick = (fn: AdminRegistryFunction) => {
    setFunctionForPricing(fn);
    setNewPrice(fn.price_per_call.toString());
    setIsPricingDialogOpen(true);
  };

  const handleUpdatePricing = () => {
    if (functionForPricing && newPrice) {
      updatePricingMutation.mutate({
        functionId: functionForPricing.id,
        price: parseFloat(newPrice),
      });
    }
  };

  const handleFlagClick = (fn: AdminRegistryFunction) => {
    setFunctionToFlag(fn);
    setFlagReason(fn.flag_reason || "");
    setIsFlagDialogOpen(true);
  };

  const handleToggleFlag = () => {
    if (functionToFlag) {
      toggleFlagMutation.mutate({
        functionId: functionToFlag.id,
        flagged: !functionToFlag.is_flagged,
        reason: flagReason,
      });
    }
  };

  const handleDeleteClick = (fn: AdminRegistryFunction) => {
    setFunctionToDelete(fn);
    setIsDeleteDialogOpen(true);
  };

  const confirmDelete = () => {
    if (functionToDelete) {
      deleteFunctionMutation.mutate(functionToDelete.id);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => navigate('/admin')}
          className="text-text-muted hover:text-text-primary hover:bg-bg-hover"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Dashboard
        </Button>
        <div className="flex-1 text-center">
          <h1 className="text-2xl font-bold text-text-primary">Registry</h1>
          <p className="text-text-secondary">
            Manage function registry - visibility, pricing, and content moderation
          </p>
        </div>
        <Button
          onClick={() => refetch()}
          variant="outline"
          className="border-border-subtle hover:bg-bg-hover"
        >
          <RefreshCw className="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatCard
          title="Total Functions"
          value={stats.total_functions}
          icon={<Code className="w-5 h-5 text-[#6366f1]" />}
          trend="neutral"
          change={{ value: 0, label: "in registry" }}
        />
        <StatCard
          title="Public Functions"
          value={stats.public_functions}
          icon={<Shield className="w-5 h-5 text-emerald-400" />}
          trend="up"
          change={{ value: 0, label: "visible" }}
        />
        <StatCard
          title="Total Calls"
          value={stats.total_calls.toLocaleString()}
          icon={<Download className="w-5 h-5 text-blue-400" />}
          trend="up"
          change={{ value: 0, label: "executions" }}
        />
        <StatCard
          title="Flagged"
          value={stats.flagged_functions}
          icon={<Flag className="w-5 h-5 text-red-400" />}
          trend={stats.flagged_functions > 0 ? "down" : "neutral"}
          change={{ value: 0, label: "needs review" }}
        />
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  placeholder="Search registry functions..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 bg-bg-secondary border-white/8 text-white"
                />
              </div>
            </div>
            <Select value={visibilityFilter} onValueChange={setVisibilityFilter}>
              <SelectTrigger className="w-full sm:w-[150px] bg-bg-secondary border-white/8 text-white">
                <SelectValue placeholder="Visibility" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-white/8">
                <SelectItem value="all">All Visibility</SelectItem>
                <SelectItem value="public">Public</SelectItem>
                <SelectItem value="private">Private</SelectItem>
                <SelectItem value="unlisted">Unlisted</SelectItem>
              </SelectContent>
            </Select>
            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger className="w-full sm:w-[180px] bg-bg-secondary border-white/8 text-white">
                <SelectValue placeholder="Category" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-white/8">
                {categories.map((cat) => (
                  <SelectItem key={cat} value={cat}>
                    {cat}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={flaggedFilter} onValueChange={setFlaggedFilter}>
              <SelectTrigger className="w-full sm:w-[150px] bg-bg-secondary border-white/8 text-white">
                <SelectValue placeholder="Flag Status" />
              </SelectTrigger>
              <SelectContent className="bg-bg-tertiary border-white/8">
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="flagged">Flagged</SelectItem>
                <SelectItem value="unflagged">Unflagged</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Functions List */}
      <Card>
        <CardHeader>
          <CardTitle className="text-white">
            Registry Functions ({filteredFunctions.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {functionsLoading ? (
            <div className="space-y-4">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-20 w-full" />
              ))}
            </div>
          ) : filteredFunctions.length === 0 ? (
            <div className="text-center py-12">
              <Code className="h-12 w-12 text-text-muted mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-white mb-2">
                No functions found
              </h3>
              <p className="text-text-secondary">
                {searchTerm
                  ? "Try adjusting your search or filters"
                  : "No functions in registry yet"}
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {filteredFunctions.map((fn) => (
                <div
                  key={fn.id}
                  className={`flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-white/8 hover:bg-white/5 transition-colors ${
                    fn.is_flagged ? "border-red-500/30" : ""
                  }`}
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-[#6366f1]/10 flex items-center justify-center">
                      <Code className="w-5 h-5 text-[#6366f1]" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <p className="font-medium text-white">{fn.name}</p>
                        {fn.is_flagged && (
                          <Flag className="w-4 h-4 text-red-400" />
                        )}
                      </div>
                      <p className="text-sm text-text-muted">by {fn.author}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-4">
                    <div className="text-right">
                      <Badge className={`${visibilityColors[fn.visibility as keyof typeof visibilityColors]}`}>
                        {fn.visibility}
                      </Badge>
                      <div className="flex items-center gap-2 mt-1">
                        <Star className="w-3 h-3 text-yellow-400 fill-yellow-400" />
                        <span className="text-xs text-text-muted">
                          {fn.overall_score.toFixed(1)} ({fn.total_ratings})
                        </span>
                      </div>
                    </div>

                    <div className="text-right min-w-[80px]">
                      <p className="text-white font-medium">
                        ${fn.price_per_call.toFixed(4)}
                      </p>
                      <p className="text-xs text-text-muted">per call</p>
                    </div>

                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-text-muted hover:text-white"
                        >
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="bg-bg-tertiary border-white/8">
                        <DropdownMenuItem
                          className="text-white hover:bg-white/5"
                          onClick={() => handleViewDetails(fn)}
                        >
                          <Eye className="w-4 h-4 mr-2" />
                          View Details
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-white hover:bg-white/5">
                          <Shield className="w-4 h-4 mr-2" />
                          Set Visibility
                          <div className="ml-auto flex flex-col">
                            {(["public", "private", "unlisted"] as const).map((v) => (
                              <button
                                key={v}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleVisibilityChange(fn, v);
                                }}
                                className={`px-2 py-1 text-xs hover:bg-white/10 ${
                                  fn.visibility === v ? "text-[#6366f1]" : "text-text-muted"
                                }`}
                              >
                                {v}
                              </button>
                            ))}
                          </div>
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-white hover:bg-white/5"
                          onClick={() => handlePricingClick(fn)}
                        >
                          <DollarSign className="w-4 h-4 mr-2" />
                          Set Pricing
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className={fn.is_flagged ? "text-emerald-400 hover:bg-emerald-500/10" : "text-red-400 hover:bg-red-500/10"}
                          onClick={() => handleFlagClick(fn)}
                        >
                          {fn.is_flagged ? (
                            <>
                              <Flag className="w-4 h-4 mr-2" />
                              Unflag
                            </>
                          ) : (
                            <>
                              <Flag className="w-4 h-4 mr-2" />
                              Flag
                            </>
                          )}
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-red-400 hover:bg-red-500/10"
                          onClick={() => handleDeleteClick(fn)}
                        >
                          <Trash2 className="w-4 h-4 mr-2" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Function Details Dialog */}
      <Dialog open={isDetailsOpen} onOpenChange={setIsDetailsOpen}>
        <DialogContent className="bg-bg-tertiary border-white/8 max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-white flex items-center gap-2">
              <Code className="w-5 h-5 text-[#6366f1]" />
              {selectedFunction?.name}
            </DialogTitle>
          </DialogHeader>

          {functionDetailsLoading ? (
            <LoadingSpinner />
          ) : functionDetails ? (
            <Tabs defaultValue="details" className="w-full">
              <TabsList className="bg-bg-secondary">
                <TabsTrigger value="details" className="text-white">
                  Details
                </TabsTrigger>
                <TabsTrigger value="versions" className="text-white">
                  Versions
                </TabsTrigger>
              </TabsList>

              <TabsContent value="details" className="space-y-4 mt-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label className="text-text-muted">Author</Label>
                    <p className="text-white">{functionDetails.function.author}</p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Name</Label>
                    <p className="text-white">{functionDetails.function.name}</p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Visibility</Label>
                    <p className="text-white">
                      <Badge className={`${visibilityColors[functionDetails.function.visibility]}`}>
                        {functionDetails.function.visibility}
                      </Badge>
                    </p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Price per Call</Label>
                    <p className="text-white">${functionDetails.function.price_per_call.toFixed(4)}</p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Category</Label>
                    <p className="text-white">{functionDetails.function.category || "Uncategorized"}</p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Latest Version</Label>
                    <p className="text-white">{functionDetails.function.latest_version || "N/A"}</p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Overall Score</Label>
                    <p className="text-white flex items-center gap-1">
                      <Star className="w-4 h-4 text-yellow-400 fill-yellow-400" />
                      {functionDetails.function.overall_score.toFixed(1)} ({functionDetails.function.total_ratings} ratings)
                    </p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Popularity</Label>
                    <p className="text-white">{Math.floor(functionDetails.function.popularity_score)}</p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Reliability Score</Label>
                    <p className="text-white">{functionDetails.function.reliability_score.toFixed(1)}</p>
                  </div>
                  <div>
                    <Label className="text-text-muted">Deterministic Score</Label>
                    <p className="text-white">{functionDetails.function.deterministic_score.toFixed(1)}</p>
                  </div>
                </div>

                <div>
                  <Label className="text-text-muted">Description</Label>
                  <p className="text-white mt-1">
                    {functionDetails.function.description || "No description"}
                  </p>
                </div>

                {functionDetails.function.tags && functionDetails.function.tags.length > 0 && (
                  <div>
                    <Label className="text-text-muted">Tags</Label>
                    <div className="flex flex-wrap gap-2 mt-2">
                      {functionDetails.function.tags.map((tag, i) => (
                        <Badge key={i} variant="outline" className="text-white">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                {functionDetails.function.is_flagged && (
                  <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
                    <div className="flex items-center gap-2 text-red-400">
                      <Flag className="w-4 h-4" />
                      <span className="font-medium">Flagged</span>
                    </div>
                    <p className="text-white mt-2">{functionDetails.function.flag_reason}</p>
                  </div>
                )}

                <div>
                  <Label className="text-text-muted">Created At</Label>
                  <p className="text-white">
                    {new Date(functionDetails.function.created_at).toLocaleString()}
                  </p>
                </div>
              </TabsContent>

              <TabsContent value="versions" className="mt-4">
                {functionDetails.versions && functionDetails.versions.length > 0 ? (
                  <div className="space-y-3">
                    {functionDetails.versions.map((version) => (
                      <div
                        key={version.id}
                        className="p-4 bg-bg-secondary rounded-lg border border-white/8"
                      >
                        <div className="flex justify-between items-start">
                          <div>
                            <p className="text-white font-medium">v{version.version}</p>
                            <p className="text-sm text-text-muted">
                              Runtime: {version.runtime}
                            </p>
                          </div>
                          <Badge
                            className={version.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-gray-500/10 text-gray-400"}
                          >
                            {version.is_active ? "Active" : "Inactive"}
                          </Badge>
                        </div>
                        <div className="grid grid-cols-3 gap-4 mt-3 text-sm">
                          <div>
                            <span className="text-text-muted">Timeout:</span>{" "}
                            <span className="text-white">{version.timeout_ms}ms</span>
                          </div>
                          <div>
                            <span className="text-text-muted">Memory:</span>{" "}
                            <span className="text-white">{version.memory_mb}MB</span>
                          </div>
                          <div>
                            <span className="text-text-muted">Cache TTL:</span>{" "}
                            <span className="text-white">{version.cache_ttl}s</span>
                          </div>
                        </div>
                        <div className="flex items-center gap-2 mt-3">
                          {version.deterministic && (
                            <Badge variant="outline" className="text-blue-400 border-blue-400/30">
                              Deterministic
                            </Badge>
                          )}
                        </div>
                        <p className="text-xs text-text-muted mt-2">
                          Published: {new Date(version.published_at).toLocaleString()}
                        </p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-text-muted text-center py-8">No versions found</p>
                )}
              </TabsContent>
            </Tabs>
          ) : null}
        </DialogContent>
      </Dialog>

      {/* Flag Dialog */}
      <Dialog open={isFlagDialogOpen} onOpenChange={setIsFlagDialogOpen}>
        <DialogContent className="bg-bg-tertiary border-white/8">
          <DialogHeader>
            <DialogTitle className="text-white flex items-center gap-2">
              <Flag className="w-5 h-5 text-red-400" />
              {functionToFlag?.is_flagged ? "Unflag Function" : "Flag Function"}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-text-secondary">
              {functionToFlag?.is_flagged
                ? `Unflag "${functionToFlag?.name}"? This will restore visibility.`
                : `Flag "${functionToFlag?.name}" for review.`}
            </p>
            <div>
              <Label className="text-white">Reason (optional)</Label>
              <Textarea
                value={flagReason}
                onChange={(e) => setFlagReason(e.target.value)}
                placeholder="Enter reason for flagging..."
                className="mt-2 bg-bg-secondary border-white/8 text-white"
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsFlagDialogOpen(false)}
              className="border-white/8 text-white hover:bg-white/5"
            >
              Cancel
            </Button>
            <Button
              onClick={handleToggleFlag}
              disabled={toggleFlagMutation.isPending}
              className={functionToFlag?.is_flagged ? "bg-emerald-600 hover:bg-emerald-700" : "bg-red-600 hover:bg-red-700"}
            >
              {toggleFlagMutation.isPending
                ? "Processing..."
                : functionToFlag?.is_flagged
                ? "Unflag"
                : "Flag"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Pricing Dialog */}
      <Dialog open={isPricingDialogOpen} onOpenChange={setIsPricingDialogOpen}>
        <DialogContent className="bg-bg-tertiary border-white/8">
          <DialogHeader>
            <DialogTitle className="text-white flex items-center gap-2">
              <DollarSign className="w-5 h-5 text-[#6366f1]" />
              Set Pricing
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-text-secondary">
              Set the price per call for "{functionForPricing?.name}"
            </p>
            <div>
              <Label className="text-white">Price per Call (USD)</Label>
              <Input
                type="number"
                step="0.0001"
                min="0"
                value={newPrice}
                onChange={(e) => setNewPrice(e.target.value)}
                placeholder="0.0000"
                className="mt-2 bg-bg-secondary border-white/8 text-white"
              />
              <p className="text-xs text-text-muted mt-1">
                Set to 0 for free functions
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsPricingDialogOpen(false)}
              className="border-white/8 text-white hover:bg-white/5"
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpdatePricing}
              disabled={updatePricingMutation.isPending || !newPrice}
              className="bg-[#6366f1] hover:bg-[#5855eb]"
            >
              {updatePricingMutation.isPending ? "Updating..." : "Update Pricing"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent className="bg-bg-tertiary border-white/8">
          <DialogHeader>
            <DialogTitle className="text-white flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-400" />
              Delete Function
            </DialogTitle>
          </DialogHeader>
          <p className="text-text-secondary">
            Are you sure you want to delete{" "}
            <span className="text-white font-medium">{functionToDelete?.name}</span>?
            This action cannot be undone and will remove it from the registry.
          </p>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(false)}
              className="border-white/8 text-white hover:bg-white/5"
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              disabled={deleteFunctionMutation.isPending}
              className="bg-red-600 hover:bg-red-700"
            >
              {deleteFunctionMutation.isPending ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
