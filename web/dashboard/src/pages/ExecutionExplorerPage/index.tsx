import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { motion } from "framer-motion";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  ArrowLeft,
  Lock,
  CheckCircle,
  Shield,
  Clock,
  Hash,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Copy,
  Check,
  AlertCircle,
  Filter,
  Calendar,
} from "lucide-react";
import { Navbar } from "@/components/common/Navbar";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { dreApi, type Execution, type ExecutionDetail } from "@/api/dre";
import { toast } from "sonner";
import { formatDistanceToNow, format } from "date-fns";

export default function ExecutionExplorerPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const [page, setPage] = useState(0);
  const [limit] = useState(20);
  const [selectedExecution, setSelectedExecution] = useState<Execution | null>(
    null
  );
  const [filters, setFilters] = useState({
    version: "",
    verifiedOnly: false,
  });

  const { data, isLoading, error } = useQuery({
    queryKey: [
      "executions",
      author,
      name,
      page,
      limit,
      filters.version,
      filters.verifiedOnly,
    ],
    queryFn: () =>
      dreApi.listExecutions(author!, name!, {
        offset: page * limit,
        limit,
        version: filters.version || undefined,
        verified_only: filters.verifiedOnly || undefined,
      }),
    enabled: !!author && !!name,
  });

  const { data: executionDetail } = useQuery({
    queryKey: ["execution-detail", selectedExecution?.execution_id],
    queryFn: () =>
      dreApi.getExecution(author!, name!, selectedExecution!.execution_id),
    enabled: !!selectedExecution && !!author && !!name,
  });

  const totalPages = data ? Math.ceil(data.total / limit) : 0;

  if (isLoading) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16">
          <div className="container mx-auto px-4 py-8">
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner size="lg" />
            </div>
          </div>
        </main>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16">
          <div className="container mx-auto px-4 py-8">
            <ErrorMessage error={error as Error} />
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-bg-primary">
      <Navbar variant="landing" />
      <main className="flex-1 pt-16">
        <div className="container mx-auto px-4 py-8 max-w-6xl">
          {/* Header */}
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-8"
          >
            <Link
              to={`/fx/${author}/${name}`}
              className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors mb-4"
            >
              <ArrowLeft className="w-4 h-4" />
              Back to Function
            </Link>

            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
              <div>
                <h1 className="text-3xl font-bold flex items-center gap-3">
                  <Lock className="w-8 h-8 text-brand-500" />
                  Execution Explorer
                </h1>
                <p className="text-muted-foreground mt-1">
                  Browse all execution root hashes for{" "}
                  <span className="font-mono text-foreground">
                    {author}/{name}
                  </span>
                </p>
              </div>

              {data && (
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>
                    <strong className="text-foreground">{data.total}</strong>{" "}
                    total executions
                  </span>
                </div>
              )}
            </div>
          </motion.div>

          {/* Filters */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="mb-6"
          >
            <Card className="bg-bg-primary/60 border-border-subtle">
              <CardContent className="p-4">
                <div className="flex flex-wrap items-center gap-4">
                  <div className="flex items-center gap-2">
                    <Filter className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Filters:</span>
                  </div>

                  <Select
                    value={filters.version}
                    onValueChange={(value) =>
                      setFilters((f) => ({ ...f, version: value }))
                    }
                  >
                    <SelectTrigger className="w-40">
                      <SelectValue placeholder="All versions" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">All versions</SelectItem>
                      {/* Could dynamically load versions */}
                    </SelectContent>
                  </Select>

                  <div className="flex items-center gap-2">
                    <Switch
                      id="verified-only"
                      checked={filters.verifiedOnly}
                      onCheckedChange={(checked) =>
                        setFilters((f) => ({ ...f, verifiedOnly: checked }))
                      }
                    />
                    <Label
                      htmlFor="verified-only"
                      className="text-sm cursor-pointer"
                    >
                      Verified only
                    </Label>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Executions Grid */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
            className="grid gap-4"
          >
            {data?.executions.map((execution, index) => (
              <ExecutionCard
                key={execution.execution_id}
                execution={execution}
                index={index}
                onClick={() => setSelectedExecution(execution)}
              />
            ))}

            {data?.executions.length === 0 && (
              <Card className="bg-bg-primary/60 border-border-subtle">
                <CardContent className="p-12 text-center">
                  <Hash className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                  <h3 className="text-lg font-medium mb-2">
                    No executions found
                  </h3>
                  <p className="text-muted-foreground">
                    This function hasn't been executed yet or no execution
                    records are available.
                  </p>
                </CardContent>
              </Card>
            )}
          </motion.div>

          {/* Pagination */}
          {totalPages > 1 && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.3 }}
              className="flex items-center justify-between mt-8"
            >
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
              >
                <ChevronLeft className="w-4 h-4 mr-2" />
                Previous
              </Button>

              <span className="text-sm text-muted-foreground">
                Page {page + 1} of {totalPages}
              </span>

              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                disabled={page >= totalPages - 1}
              >
                Next
                <ChevronRight className="w-4 h-4 ml-2" />
              </Button>
            </motion.div>
          )}
        </div>
      </main>

      {/* Execution Detail Modal */}
      <Dialog
        open={!!selectedExecution}
        onOpenChange={() => setSelectedExecution(null)}
      >
        <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Lock className="w-5 h-5 text-brand-500" />
              Execution Details
            </DialogTitle>
            <DialogDescription>
              Complete information about this execution
            </DialogDescription>
          </DialogHeader>

          {executionDetail ? (
            <ExecutionDetailView execution={executionDetail.execution} />
          ) : (
            <div className="flex items-center justify-center py-12">
              <LoadingSpinner />
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

function ExecutionCard({
  execution,
  index,
  onClick,
}: {
  execution: Execution;
  index: number;
  onClick: () => void;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05 }}
    >
      <Card
        className="bg-bg-primary/60 border-border-subtle hover:border-brand-500/30 transition-all cursor-pointer group"
        onClick={onClick}
      >
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row md:items-center gap-4">
            {/* Hash */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-2">
                <Lock className="w-4 h-4 text-brand-500" />
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  Execution Root Hash
                </span>
              </div>
              <code className="text-sm font-mono bg-bg-secondary px-2 py-1 rounded block truncate">
                {execution.execution_root_hash}
              </code>
            </div>

            {/* Version & Date */}
            <div className="flex items-center gap-4 text-sm">
              <div className="flex items-center gap-1">
                <Badge variant="outline">v{execution.version}</Badge>
              </div>
              <div
                className="flex items-center gap-1 text-muted-foreground"
                title={format(new Date(execution.created_at), "PPpp")}
              >
                <Calendar className="w-4 h-4" />
                {formatDistanceToNow(new Date(execution.created_at), {
                  addSuffix: true,
                })}
              </div>
            </div>

            {/* Status Badges */}
            <div className="flex items-center gap-2">
              {execution.replay_verified ? (
                <Badge
                  variant="outline"
                  className="bg-green-500/10 text-green-500 border-green-500/20"
                >
                  <CheckCircle className="w-3 h-3 mr-1" />
                  Verified
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="text-muted-foreground"
                >
                  <Clock className="w-3 h-3 mr-1" />
                  Pending
                </Badge>
              )}

              {execution.roots_match && (
                <Badge
                  variant="outline"
                  className="bg-blue-500/10 text-blue-500 border-blue-500/20"
                >
                  <Shield className="w-3 h-3 mr-1" />
                  Match
                </Badge>
              )}
            </div>

            {/* Arrow */}
            <ExternalLink className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
          </div>

          {/* Component Hashes Preview */}
          <div className="mt-4 pt-4 border-t border-border-subtle">
            <div className="grid grid-cols-7 gap-2 text-xs">
                {[
                { key: "input", label: "Input" },
                { key: "output", label: "Output" },
                { key: "environment", label: "Env" },
                { key: "dependency", label: "Dep" },
                { key: "trace", label: "Trace" },
                { key: "resource", label: "Res" },
                { key: "metadata", label: "Meta" },
              ].map(({ key, label }) => {
                const hashes = execution.component_hashes as unknown as Record<string, string>;
                const hash = hashes[key];
                return (
                  <div key={key} className="text-center">
                    <div className="text-muted-foreground mb-1">{label}</div>
                    <code className="font-mono text-[10px] bg-bg-secondary px-1 py-0.5 rounded block truncate">
                      {hash?.slice(0, 8) || "—"}
                    </code>
                  </div>
                );
              })}
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function ExecutionDetailView({ execution }: { execution: ExecutionDetail }) {
  const [copied, setCopied] = useState(false);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    toast.success("Copied to clipboard");
    setTimeout(() => setCopied(false), 2000);
  };

  const componentOrder = [
    { key: "input", label: "Input Hash" },
    { key: "output", label: "Output Hash" },
    { key: "environment", label: "Environment Hash" },
    { key: "dependency", label: "Dependency Hash" },
    { key: "trace", label: "Trace Hash" },
    { key: "resource", label: "Resource Hash" },
    { key: "metadata", label: "Metadata Hash" },
  ] as const;

  return (
    <div className="space-y-6">
      {/* Execution Root Hash */}
      <div className="space-y-2">
        <Label className="text-xs uppercase tracking-wide text-muted-foreground">
          Execution Root Hash
        </Label>
        <div className="flex gap-2">
          <code className="flex-1 font-mono text-sm bg-bg-secondary p-3 rounded break-all">
            {execution.execution_root_hash}
          </code>
          <Button
            variant="outline"
            size="icon"
            onClick={() => copyToClipboard(execution.execution_root_hash)}
          >
            {copied ? (
              <Check className="w-4 h-4" />
            ) : (
              <Copy className="w-4 h-4" />
            )}
          </Button>
        </div>
      </div>

      {/* Status Grid */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1">
          <Label className="text-xs uppercase tracking-wide text-muted-foreground">
            Version
          </Label>
          <div className="text-sm">{execution.version}</div>
        </div>
        <div className="space-y-1">
          <Label className="text-xs uppercase tracking-wide text-muted-foreground">
            Created
          </Label>
          <div className="text-sm">
            {format(new Date(execution.created_at), "PPpp")}
          </div>
        </div>
        <div className="space-y-1">
          <Label className="text-xs uppercase tracking-wide text-muted-foreground">
            Determinism Tier
          </Label>
          <Badge variant="outline" className="capitalize">
            {execution.determinism_tier}
          </Badge>
        </div>
        <div className="space-y-1">
          <Label className="text-xs uppercase tracking-wide text-muted-foreground">
            Protocol Version
          </Label>
          <div className="text-sm font-mono">
            {execution.protocol_version}
          </div>
        </div>
      </div>

      {/* Verification Status */}
      <div className="space-y-2">
        <Label className="text-xs uppercase tracking-wide text-muted-foreground">
          Replay Verification
        </Label>
        <div className="bg-bg-secondary rounded-lg p-4 space-y-3">
          {execution.replay_verified_at ? (
            <>
              <div className="flex items-center gap-2 text-green-500">
                <CheckCircle className="w-5 h-5" />
                <span className="font-medium">Replay Verified</span>
              </div>
              <div className="text-sm text-muted-foreground space-y-1">
                <p>
                  Verified at:{" "}
                  {format(new Date(execution.replay_verified_at), "PPpp")}
                </p>
                {execution.replay_node_id && (
                  <p>Node: {execution.replay_node_id}</p>
                )}
              </div>
              {execution.roots_match ? (
                <Badge className="bg-green-500/10 text-green-500 border-green-500/20">
                  <Shield className="w-3 h-3 mr-1" />
                  Root hashes match
                </Badge>
              ) : (
                <Badge
                  variant="destructive"
                  className="bg-red-500/10 text-red-500 border-red-500/20"
                >
                  <AlertCircle className="w-3 h-3 mr-1" />
                  Root hash mismatch
                </Badge>
              )}
            </>
          ) : (
            <div className="flex items-center gap-2 text-muted-foreground">
              <Clock className="w-5 h-5" />
              <span>Replay verification pending</span>
            </div>
          )}
        </div>
      </div>

      {/* Component Hashes */}
      <div className="space-y-2">
        <Label className="text-xs uppercase tracking-wide text-muted-foreground">
          Component Hashes (MEG)
        </Label>
        <div className="space-y-2">
          {componentOrder.map(({ key, label }) => {
            const hash = execution.component_hashes[key];
            return (
              <div
                key={key}
                className="flex items-center gap-3 bg-bg-secondary rounded p-2"
              >
                <span className="text-xs text-muted-foreground w-32 shrink-0">
                  {label}
                </span>
                <code className="flex-1 font-mono text-xs truncate">
                  {hash || "—"}
                </code>
                {hash && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 shrink-0"
                    onClick={() => copyToClipboard(hash)}
                  >
                    <Copy className="w-3 h-3" />
                  </Button>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Certificate Info */}
      {execution.certificate && (
        <div className="space-y-2">
          <Label className="text-xs uppercase tracking-wide text-muted-foreground">
            Certificate
          </Label>
          <div className="bg-bg-secondary rounded-lg p-4 space-y-2">
            <div className="flex items-center gap-2">
              <Badge variant="outline" className="capitalize">
                {execution.certificate.cert_level}
              </Badge>
              {execution.certificate.anchored && (
                <Badge className="bg-blue-500/10 text-blue-500 border-blue-500/20">
                  Anchored
                </Badge>
              )}
            </div>
            <div className="text-xs text-muted-foreground">
              ID: {execution.certificate.certificate_id}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
