import { useParams, useSearchParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useState, useEffect } from "react";
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
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
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
  Play,
} from "lucide-react";
import { Navbar } from "@/components/common/Navbar";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { dreApi, type Execution, type ExecutionDetail } from "@/api/dre";
import { toast } from "sonner";
import { formatDistanceToNow, format } from "date-fns";

// Import DRE Components
import {
  ExecutionHeader,
  ExecutionRootBadge,
  MerkleExecutionTree,
  HashDiffViewer,
  ReplayExecutionButton,
  ReplayModal,
  ReplayProgressTimeline,
  ReplayResultCard,
  FXCertViewer,
  TrustScoreBreakdown,
  CapsuleInspector,
  VerificationBadge,
  DeterminismBadge,
  CollapsibleSection,
  MetricCard,
  CertificateList,
  FunctionHistoryTab,
} from "@/components/dre";
import { mapCertificateDetailToFXCertData } from "@/lib/dre";
import { ReplayMode } from "@/components/dre/replay";

export default function ExecutionExplorerPage() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const [page, setPage] = useState(0);
  const [limit] = useState(20);
  const [selectedExecution, setSelectedExecution] = useState<Execution | null>(
    null
  );
  const [filters, setFilters] = useState({
    version: "__all",
    verifiedOnly: false,
  });
  // Replay modal state
  const [replayModalOpen, setReplayModalOpen] = useState(false);
  const [replayMode, setReplayMode] = useState<ReplayMode>("strict");
  const [selectedExecutionForReplay, setSelectedExecutionForReplay] = useState<ExecutionDetail | null>(null);

  // Certificates tab and full FXCERT modal
  const [searchParams, setSearchParams] = useSearchParams();
  const tabFromUrl = searchParams.get("tab");
  const certIdFromUrl = searchParams.get("certId");
  const [activeTab, setActiveTab] = useState<"executions" | "certificates" | "history">(
    tabFromUrl === "certificates" ? "certificates" : tabFromUrl === "history" ? "history" : "history"
  );
  const [selectedCertId, setSelectedCertId] = useState<string | null>(certIdFromUrl);

  useEffect(() => {
    if (tabFromUrl === "certificates") setActiveTab("certificates");
    else if (tabFromUrl === "history") setActiveTab("history");
    else if (tabFromUrl === "executions") setActiveTab("executions");
  }, [tabFromUrl]);
  useEffect(() => {
    if (certIdFromUrl) setSelectedCertId(certIdFromUrl);
  }, [certIdFromUrl]);

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
        version: filters.version && filters.version !== "__all" ? filters.version : undefined,
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
    <div className="execution-explorer-page min-h-screen flex flex-col bg-bg-primary">
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

          <Tabs
            value={activeTab}
            onValueChange={(v) => {
              setActiveTab(v as "executions" | "certificates" | "history");
              setSearchParams((p) => {
                const next = new URLSearchParams(p);
                if (v === "certificates") next.set("tab", "certificates");
                else if (v === "history") next.set("tab", "history");
                else if (v === "executions") next.delete("tab");
                return next;
              });
            }}
            className="w-full"
          >
            <TabsList className="mb-6">
              <TabsTrigger value="history">History</TabsTrigger>
              <TabsTrigger value="certificates">Certificates (FXCERT)</TabsTrigger>
              <TabsTrigger value="executions">Executions (MEG)</TabsTrigger>
            </TabsList>

            <TabsContent value="executions" className="mt-0">
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
                      <SelectItem value="__all">All versions</SelectItem>
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
                <CardContent className="p-12">
                  <div className="text-center mb-8">
                    <Hash className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-medium mb-2">
                      No executions found
                    </h3>
                    <p className="text-muted-foreground mb-6">
                      This function hasn't been executed yet or no execution
                      records are available.
                    </p>
                  </div>
                  <div className="border-t border-border-subtle pt-8">
                    <p className="text-sm font-medium text-foreground mb-3">
                      When executions exist, you'll see:
                    </p>
                    <ul className="text-sm text-muted-foreground space-y-2 list-disc list-inside">
                      <li>Execution cards with root hash, version, and verification status</li>
                      <li>Click any card to open the <strong className="text-foreground">DRE detail view</strong>: execution header, Merkle execution tree, trust score, FXCERT viewer, and replay</li>
                      <li>Filters by version and verified-only toggle</li>
                    </ul>
                  </div>
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
            </TabsContent>

            <TabsContent value="history" className="mt-0">
              <FunctionHistoryTab
                author={author!}
                name={name!}
                onViewCert={(certId) => {
                  setSelectedCertId(certId);
                  setSearchParams((p) => {
                    const next = new URLSearchParams(p);
                    next.set("certId", certId);
                    return next;
                  });
                }}
              />
            </TabsContent>

            <TabsContent value="certificates" className="mt-0">
              <CertificateList
                author={author!}
                name={name!}
                onViewCert={(certId) => {
                  setSelectedCertId(certId);
                  setSearchParams((p) => {
                    const next = new URLSearchParams(p);
                    next.set("certId", certId);
                    return next;
                  });
                }}
              />
            </TabsContent>
          </Tabs>
        </div>
      </main>

      {/* Execution Detail Modal */}
      <Dialog
        open={!!selectedExecution}
        onOpenChange={() => setSelectedExecution(null)}
      >
        <DialogContent className="execution-explorer-detail-dialog max-w-3xl max-h-[90vh] overflow-y-auto">
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
            <ExecutionDetailView
              execution={executionDetail.execution}
              onReplay={() => {
                setSelectedExecutionForReplay(executionDetail.execution);
                setReplayModalOpen(true);
              }}
              onViewFullCert={(certId) => {
                setSelectedExecution(null);
                setSelectedCertId(certId);
                setSearchParams((p) => {
                  const next = new URLSearchParams(p);
                  next.set("certId", certId);
                  return next;
                });
              }}
            />
          ) : (
            <div className="flex items-center justify-center py-12">
              <LoadingSpinner />
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Replay Modal */}
      <ReplayModal
        open={replayModalOpen}
        onOpenChange={setReplayModalOpen}
        capsule={{
          version: "1.0.0",
          runtime_version: "python-3.11",
          memory_limit: 512,
          instruction_limit: 1000000,
          float_mode: "deterministic",
          determinism_flags: ["no-random", "deterministic-float"],
        }}
        costEstimate={{ amount: 0.001, currency: "USD" }}
        onStartReplay={(mode, nodeId) => {
          console.log("Starting replay:", mode, nodeId);
          toast.success("Replay started!");
          setReplayModalOpen(false);
        }}
      />

      {/* Full FXCERT Modal */}
      <Dialog
        open={!!selectedCertId}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedCertId(null);
            setSearchParams((p) => {
              const next = new URLSearchParams(p);
              next.delete("certId");
              return next;
            });
          }
        }}
      >
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-brand-500" />
              FXCERT
            </DialogTitle>
            <DialogDescription>
              Execution certificate details
            </DialogDescription>
          </DialogHeader>
          {selectedCertId && author && name && (
            <FXCERTModalContent
              author={author}
              name={name}
              certId={selectedCertId}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

function FXCERTModalContent({
  author,
  name,
  certId,
}: {
  author: string;
  name: string;
  certId: string;
}) {
  const { data, isLoading, error } = useQuery({
    queryKey: ["certificate", author, name, certId],
    queryFn: () => dreApi.getCertificate(author, name, certId),
    enabled: !!author && !!name && !!certId,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <LoadingSpinner />
      </div>
    );
  }
  if (error || !data) {
    return (
      <ErrorMessage
        error={error instanceof Error ? error : new Error("Failed to load certificate")}
      />
    );
  }

  const fxcertData = mapCertificateDetailToFXCertData(data);
  const fullCertJson = JSON.stringify(data.cert, null, 2);

  return (
    <FXCertViewer
      certificate={fxcertData}
      showDetails
      fullCertJson={fullCertJson}
    />
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

function ExecutionDetailView({
  execution,
  onReplay,
  onViewFullCert,
}: {
  execution: ExecutionDetail;
  onReplay?: () => void;
  onViewFullCert?: (certId: string) => void;
}) {
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
      {/* Execution Header with new component */}
      <ExecutionHeader
        executionId={execution.execution_id}
        determinismTier={execution.determinism_tier as any}
        trustScore={85}
        verified={execution.replay_verified_at !== null}
        protocolVersion={execution.protocol_version}
        isLatestProtocol={true}
      />

      {/* Execution Root Badge */}
      <ExecutionRootBadge
        hash={execution.execution_root_hash}
        verified={execution.replay_verified_at !== null}
        anchored={execution.certificate?.anchored || false}
        chain={execution.certificate?.anchored ? "Ethereum" : undefined}
        blockNumber={execution.certificate?.anchored ? 12345678 : undefined}
      />

      {/* Replay Button */}
      {onReplay && (
        <div className="flex justify-end">
          <ReplayExecutionButton
            onClick={onReplay}
            capsuleVersion={execution.version}
            showWarning={execution.determinism_tier !== "full"}
            warningMessage={
              execution.determinism_tier !== "full"
                ? "This execution has non-deterministic components. Replay may differ."
                : undefined
            }
          />
        </div>
      )}

      {/* Merkle Execution Tree */}
      <MerkleExecutionTree
        hashes={{
          input: execution.component_hashes.input,
          output: execution.component_hashes.output,
          environment: execution.component_hashes.environment,
          dependency: execution.component_hashes.dependency,
          trace: execution.component_hashes.trace,
          resource: execution.component_hashes.resource,
          metadata: execution.component_hashes.metadata,
        }}
        onNodeClick={(type, hash) => {
          console.log("Node clicked:", type, hash);
        }}
      />

      {/* Trust Score Breakdown */}
      <TrustScoreBreakdown
        determinismScore={execution.determinism_tier === "full" ? 100 : execution.determinism_tier === "lite" ? 80 : 60}
        replayConsistency={execution.roots_match ? 100 : 50}
        resourceStability={95}
        driftIncidents={execution.roots_match ? 0 : 1}
        overallScore={execution.replay_verified_at ? 85 : 60}
      />

      {/* FX Certificate */}
      {execution.certificate && (
        <div className="space-y-2">
          <FXCertViewer
            certificate={{
              certificate_id: execution.certificate.certificate_id,
              level: execution.certificate.cert_level as "standard" | "extended" | "enterprise",
              certificate_hash: execution.certificate.certificate_hash,
              execution_root_hash: execution.execution_root_hash,
              issued_at: execution.certificate.created_at,
              expires_at: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString(),
              signatures: {
                node: { verified: true, key_id: "node-key-1" },
                platform: { verified: true, key_id: "platform-key-1" },
              },
              anchor: execution.certificate.anchored
                ? {
                    chain: "Ethereum",
                    block_number: 12345678,
                    tx_hash: "0x1234...",
                    timestamp: new Date().toISOString(),
                  }
                : undefined,
            }}
          />
          {onViewFullCert && (
            <Button
              variant="outline"
              size="sm"
              className="w-full gap-2"
              onClick={() => onViewFullCert(execution.certificate!.certificate_id)}
            >
              <Shield className="h-4 w-4" />
              View full FXCERT
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
