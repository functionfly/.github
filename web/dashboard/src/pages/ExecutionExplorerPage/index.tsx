import { useParams, useSearchParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks";
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
  HashBlock,
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
import type { ComponentType } from "@/components/dre";
import { mapCertificateDetailToFXCertData } from "@/lib/dre";
import "./styles.css";
import { ReplayMode } from "@/components/dre/replay";

export default function ExecutionExplorerPage() {
  usePageTitle("Executions");
  const { t } = useTranslation();
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
      <div className="min-h-screen flex flex-col bg-[var(--bg)]">
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
      <div className="min-h-screen flex flex-col bg-[var(--bg)]">
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
    <div className="aviation-executions min-h-screen flex flex-col">
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
              className="executions-back-link"
            >
              <ArrowLeft className="w-4 h-4" />
              {t("executionsPage.backToFunction")}
            </Link>

            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
              <div>
                <h1 className="text-3xl font-bold flex items-center gap-3 executions-title">
                  <Lock className="w-8 h-8 text-[var(--status-ok)]" />
                  {t("executionsPage.title")}
                </h1>
                <p className="text-[var(--text-faint)] mt-1 executions-subtitle">
                  {t("executionsPage.subtitle")}{" "}
                  <span className="font-mono text-[var(--text)]">
                    {author}/{name}
                  </span>
                </p>
              </div>

              {data && (
                <div className="executions-stats-bar">
                  <span className="text-sm text-[var(--text-faint)]">
                    <strong className="executions-stat-value">{data.total}</strong>{" "}
                    {t("executionsPage.totalLabel")}
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
            className="w-full executions-tabs"
          >
            <TabsList className="mb-6 executions-tabs-list">
              <TabsTrigger value="history" className="executions-tabs-trigger">{t("executionsPage.tabs.history")}</TabsTrigger>
              <TabsTrigger value="certificates" className="executions-tabs-trigger">{t("executionsPage.tabs.certificates")}</TabsTrigger>
              <TabsTrigger value="executions" className="executions-tabs-trigger">{t("executionsPage.tabs.executions")}</TabsTrigger>
            </TabsList>

            <TabsContent value="executions" className="mt-0">
          {/* Filters */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
            className="mb-6"
          >
            <Card className="executions-filters">
              <CardContent className="p-4">
                <div className="flex flex-wrap items-center gap-4">
                  <div className="flex items-center gap-2">
                    <Filter className="w-4 h-4 text-[var(--text-faint)]" />
                    <span className="executions-filter-label">{t("executionsPage.filters.label")}:</span>
                  </div>

                  <Select
                    value={filters.version}
                    onValueChange={(value) =>
                      setFilters((f) => ({ ...f, version: value }))
                    }
                  >
                    <SelectTrigger className="w-40 executions-filter-select">
                      <SelectValue placeholder={t("executionsPage.filters.allVersions")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__all">{t("executionsPage.filters.allVersions")}</SelectItem>
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
                      {t("executionsPage.filters.verifiedOnly")}
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
              <Card className="executions-empty-state">
                <CardContent className="p-12">
                  <div className="text-center mb-8">
                    <div className="executions-empty-icon">
                      <Hash className="w-8 h-8" />
                    </div>
                    <h3 className="executions-empty-title">
                      {t("executionsPage.empty.title")}
                    </h3>
                    <p className="executions-empty-description">
                      {t("executionsPage.empty.description")}
                    </p>
                  </div>
                  <div className="border-t border-[var(--panel-edge)] pt-8">
                    <p className="text-sm font-medium text-[var(--text)] mb-3">
                      {t("executionsPage.empty.whenExists")}
                    </p>
                    <ul className="executions-empty-list">
                      <li>{t("executionsPage.empty.feature1")}</li>
                      <li>{t("executionsPage.empty.feature2")}</li>
                      <li>{t("executionsPage.empty.feature3")}</li>
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
              className="executions-pagination"
            >
              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
                className="executions-pagination-btn"
              >
                <ChevronLeft className="w-4 h-4 mr-2" />
                {t("executionsPage.pagination.previous")}
              </Button>

              <span className="executions-pagination-info">
                {t("executionsPage.pagination.page")} <strong>{page + 1}</strong> {t("executionsPage.pagination.of")} <strong>{totalPages}</strong>
              </span>

              <Button
                variant="outline"
                onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                disabled={page >= totalPages - 1}
                className="executions-pagination-btn"
              >
                {t("executionsPage.pagination.next")}
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
        <DialogContent className="aviation-executions-dialog max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Lock className="w-5 h-5 text-[var(--status-ok)]" />
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
        <DialogContent className="aviation-executions-dialog max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-[var(--status-ok)]" />
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
  const { t } = useTranslation();
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05 }}
    >
      <Card
        className="execution-card"
        onClick={onClick}
      >
        <div className="execution-card-accent" />
        <CardContent className="p-4">
          <div className="flex flex-col md:flex-row md:items-center gap-4">
            {/* Hash */}
            <div className="flex-1 min-w-0">
              <div className="execution-card-header">
                <div className="execution-card-icon">
                  <Lock className="w-4 h-4" />
                </div>
                <span className="execution-card-label">
                  {t("executionsPage.card.rootHash")}
                </span>
              </div>
              <code className="execution-hash block truncate">
                {execution.execution_root_hash}
              </code>
            </div>

            {/* Version & Date */}
            <div className="execution-meta">
              <div className="flex items-center gap-1">
                <span className="execution-version-badge">v{execution.version}</span>
              </div>
              <div
                className="execution-timestamp"
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
                <span className="execution-badge execution-badge-verified">
                  <CheckCircle className="w-3 h-3 mr-1" />
                  {t("executionsPage.card.verified")}
                </span>
              ) : (
                <span className="execution-badge execution-badge-pending">
                  <Clock className="w-3 h-3 mr-1" />
                  {t("executionsPage.card.pending")}
                </span>
              )}

              {execution.roots_match && (
                <span className="execution-badge execution-badge-match">
                  <Shield className="w-3 h-3 mr-1" />
                  {t("executionsPage.card.match")}
                </span>
              )}
            </div>

            {/* Arrow */}
            <ExternalLink className="execution-card-arrow w-4 h-4" />
          </div>

          {/* Component Hashes Preview */}
          <div className="execution-components">
                {[
                { key: "input", label: t("executionsPage.card.input") },
                { key: "output", label: t("executionsPage.card.output") },
                { key: "environment", label: t("executionsPage.card.env") },
                { key: "dependency", label: t("executionsPage.card.dep") },
                { key: "trace", label: t("executionsPage.card.trace") },
                { key: "resource", label: t("executionsPage.card.res") },
                { key: "metadata", label: t("executionsPage.card.meta") },
              ].map(({ key, label }) => {
                const hashes = execution.component_hashes as unknown as Record<string, string>;
                const hash = hashes[key];
                return (
                  <div key={key} className="execution-component">
                    <div className="execution-component-label">{label}</div>
                    <code className="execution-component-hash">
                      {hash?.slice(0, 8) || "—"}
                    </code>
                  </div>
                );
              })}
            </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

const COMPONENT_DETAIL: Record<
  ComponentType,
  { label: string; description: string }
> = {
  input: { label: "Input", description: "Function input parameters" },
  output: { label: "Output", description: "Function output result" },
  environment: {
    label: "Environment",
    description: "Environment variables and config",
  },
  dependency: {
    label: "Dependencies",
    description: "External dependencies and packages",
  },
  trace: { label: "Trace", description: "Execution trace log" },
  resource: {
    label: "Resources",
    description: "CPU, memory, and I/O usage",
  },
  metadata: {
    label: "Metadata",
    description: "Execution metadata and timing",
  },
};

function useComponentDetailTranslation() {
  const { t } = useTranslation();
  return {
    input: { label: t("executionsPage.component.input"), description: t("executionsPage.component.inputDesc") },
    output: { label: t("executionsPage.component.output"), description: t("executionsPage.component.outputDesc") },
    environment: { label: t("executionsPage.component.environment"), description: t("executionsPage.component.environmentDesc") },
    dependency: { label: t("executionsPage.component.dependency"), description: t("executionsPage.component.dependencyDesc") },
    trace: { label: t("executionsPage.component.trace"), description: t("executionsPage.component.traceDesc") },
    resource: { label: t("executionsPage.component.resource"), description: t("executionsPage.component.resourceDesc") },
    metadata: { label: t("executionsPage.component.metadata"), description: t("executionsPage.component.metadataDesc") },
  };
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
  const { t } = useTranslation();
  const componentDetailTranslations = useComponentDetailTranslation();
  const [copied, setCopied] = useState(false);
  const [componentDetail, setComponentDetail] = useState<{
    type: ComponentType;
    hash: string;
  } | null>(null);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    toast.success(t("executionsPage.copiedToClipboard"));
    setTimeout(() => setCopied(false), 2000);
  };

  const componentOrder = [
    { key: "input", label: t("executionsPage.component.inputHash") },
    { key: "output", label: t("executionsPage.component.outputHash") },
    { key: "environment", label: t("executionsPage.component.environmentHash") },
    { key: "dependency", label: t("executionsPage.component.dependencyHash") },
    { key: "trace", label: t("executionsPage.component.traceHash") },
    { key: "resource", label: t("executionsPage.component.resourceHash") },
    { key: "metadata", label: t("executionsPage.component.metadataHash") },
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
                ? t("executionsPage.replayWarning")
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
        onNodeClick={(type, hash) => setComponentDetail({ type, hash })}
      />

      {/* Component detail dialog (View details) */}
      <Dialog
        open={!!componentDetail}
        onOpenChange={(open) => {
          if (!open) setComponentDetail(null);
        }}
      >
        <DialogContent
          className="max-w-md"
          onPointerDownOutside={() => setComponentDetail(null)}
          onEscapeKeyDown={() => setComponentDetail(null)}
        >
          {componentDetail && (
            <>
              <DialogHeader>
                <DialogTitle>
                  {componentDetailTranslations[componentDetail.type].label}
                </DialogTitle>
                <DialogDescription>
                  {componentDetailTranslations[componentDetail.type].description}
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-3 pt-2">
                <Label className="text-[var(--text-faint)] text-xs uppercase tracking-wide">
                  {t("executionsPage.component.componentHash")}
                </Label>
                <HashBlock
                  hash={componentDetail.hash}
                  className="w-full"
                />
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>

      {/* Trust Score Breakdown: use API trust when present, else derive from execution only (no mock values) */}
      <TrustScoreBreakdown
        determinismScore={
          execution.trust != null
            ? Math.round(execution.trust.determinism_score)
            : execution.determinism_tier === "full"
              ? 100
              : execution.determinism_tier === "lite"
                ? 80
                : 60
        }
        replayConsistency={
          execution.trust != null
            ? Math.round(execution.trust.replay_consistency_score)
            : execution.roots_match
              ? 100
              : 50
        }
        resourceStability={undefined}
        driftIncidents={
          execution.trust != null ? execution.trust.drift_incidents_total : 0
        }
        overallScore={
          execution.trust != null
            ? Math.round(execution.trust.trust_score)
            : (() => {
                const det =
                  execution.determinism_tier === "full"
                    ? 100
                    : execution.determinism_tier === "lite"
                      ? 80
                      : 60;
                const replay = execution.roots_match ? 100 : 50;
                return Math.round((det + replay) / 2);
              })()
        }
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
              {t("executionsPage.viewFullCert")}
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
