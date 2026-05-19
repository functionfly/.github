/**
 * RegistryIntegration
 * Main container that wires all 13 Function Registry components together
 */

import * as React from "react"
import {
  FunctionCard,
  FunctionMarketplace,
  InstallFunctionModal,
  RevenueAnalytics,
  SubscriptionManager,
  UsageBillingPanel,
  CreatorProfile,
  MarketplaceLeaderboard,
  FunctionRoyaltiesPanel,
  AssetPricingEditor,
  LicenseManager,
  MonetizationOptimizer,
  RegistrySearch,
  RegistryCategoryExplorer,
  RuntimeCompatibilityMatrix,
  DependencyViewer,
  FunctionSandbox,
  VersionDiffViewer,
  FunctionUsageAnalytics,
  PublicWorkflowExplorer,
  WorkflowTemplateGallery,
  PackageIntegrityViewer,
  FunctionPermissionAudit,
  FunctionChangelog,
  FunctionBenchmarkPanel,
  type FunctionCardData,
  type RegistryFilters,
  type CategoryNode,
  type FunctionCompatibility,
  type WorkflowTemplate,
  type IntegrityCheck,
  type PermissionEntry,
  type ChangelogEntry,
  type BenchmarkResult,
  type UsageDataPoint,
  type SandboxInput,
} from "@functionfly/ui-marketplace"
import { useRegistryStore, type RegistryFunction } from "@/stores/registryStore"
import { cn } from "@functionfly/ui-core"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@functionfly/ui-core"

// ============================================================================
// Types
// ============================================================================

interface RegistryIntegrationProps {
  functions?: RegistryFunction[]
  isLoading?: boolean
  onFunctionSelect?: (id: string) => void
  onFunctionExecute?: (id: string) => void
  onFunctionFavorite?: (id: string, isFavorite: boolean) => void
  onFunctionShare?: (id: string) => void
  className?: string
}

// ============================================================================
// Sample Data Helpers
// ============================================================================

const sampleCategories: CategoryNode[] = [
  {
    id: "data-processing",
    name: "Data Processing",
    description: "Transform and process data at scale",
    icon: <span className="text-lg">📊</span>,
    count: 342,
    trending: true,
    subcategories: [
      { id: "etl", name: "ETL", description: "Extract, transform, load", icon: <span>🔄</span>, count: 89 },
      { id: "streaming", name: "Streaming", description: "Real-time data streams", icon: <span>⚡</span>, count: 67 },
    ],
  },
  {
    id: "ai-ml",
    name: "AI & ML",
    description: "Machine learning and AI functions",
    icon: <span className="text-lg">🤖</span>,
    count: 289,
    subcategories: [
      { id: "inference", name: "Inference", description: "Model inference", icon: <span>🎯</span>, count: 134 },
      { id: "training", name: "Training", description: "Model training", icon: <span>🏋️</span>, count: 45 },
    ],
  },
  {
    id: "automation",
    name: "Automation",
    description: "Automate repetitive tasks",
    icon: <span className="text-lg">⚙️</span>,
    count: 198,
  },
  {
    id: "integrations",
    name: "Integrations",
    description: "Connect with external services",
    icon: <span className="text-lg">🔗</span>,
    count: 156,
  },
]

const sampleUsageData: UsageDataPoint[] = Array.from({ length: 30 }, (_, i) => ({
  timestamp: new Date(Date.now() - (29 - i) * 24 * 60 * 60 * 1000).toISOString(),
  executions: Math.floor(Math.random() * 1000) + 100,
  errors: Math.floor(Math.random() * 20),
  latency: Math.floor(Math.random() * 100) + 20,
}))

const sampleIntegrityChecks: IntegrityCheck[] = [
  { id: "checksum", name: "Checksum Verification", status: "passed", description: "Package checksum matches", details: "sha256:abc123..." },
  { id: "signature", name: "Signature Validation", status: "passed", description: "Digital signature is valid" },
  { id: "dependencies", name: "Dependency Scan", status: "warning", description: "Some dependencies have known vulnerabilities", details: "package@1.2.3 has CVE-2024-1234" },
  { id: "sandbox", name: "Sandbox Execution", status: "passed", description: "Function executes in isolated environment" },
  { id: "timeout", name: "Timeout Protection", status: "passed", description: "Execution timeout is properly set" },
]

const samplePermissions: PermissionEntry[] = [
  { id: "p1", principal: "alice@acme.corp", principalType: "user", permissions: ["read", "execute", "admin"], grantedAt: "2024-01-15", isActive: true },
  { id: "p2", principal: "team-data-science", principalType: "team", permissions: ["read", "execute"], grantedAt: "2024-02-20", isActive: true },
  { id: "p3", principal: "api-prod-key", principalType: "api_key", permissions: ["read", "execute"], grantedAt: "2024-03-01", expiresAt: "2024-12-31", isActive: true },
  { id: "p4", principal: "Public", principalType: "public", permissions: ["read"], grantedAt: "2024-01-01", isActive: false },
]

const sampleChangelog: ChangelogEntry[] = [
  { version: "2.1.0", date: "2024-05-01", author: "jane@author.com", type: "minor", changes: ["Added streaming support", "Improved error handling", "Bug fixes"] },
  { version: "2.0.0", date: "2024-03-15", author: "jane@author.com", type: "major", changes: ["Breaking: New API structure", "Added ML inference", "Performance improvements 2x"] },
  { version: "1.5.0", date: "2024-01-10", author: "john@author.com", type: "minor", changes: ["Added batch processing", "New configuration options"] },
  { version: "1.4.2", date: "2023-12-01", author: "john@author.com", type: "patch", changes: ["Bug fix: Memory leak in loops", "Dependency updates"] },
  { version: "1.4.0", date: "2023-11-15", author: "jane@author.com", type: "minor", changes: ["Added webhooks", "Security hardening"], },
  { version: "1.0.0", date: "2023-06-01", author: "admin@functionfly.io", type: "major", changes: ["Initial release"], },
]

const sampleBenchmarks: BenchmarkResult[] = [
  { name: "Cold Start", iterations: 1000, avgMs: 45.2, p50Ms: 42.1, p95Ms: 68.3, p99Ms: 89.2, minMs: 28.5, maxMs: 145.6, stdDev: 12.3, isBaseline: true },
  { name: "Warm Execution", iterations: 5000, avgMs: 12.8, p50Ms: 11.9, p95Ms: 18.2, p99Ms: 24.6, minMs: 8.2, maxMs: 35.1, stdDev: 4.2 },
  { name: "Batch Processing", iterations: 500, avgMs: 234.5, p50Ms: 220.0, p95Ms: 312.4, p99Ms: 398.1, minMs: 180.2, maxMs: 520.0, stdDev: 45.6 },
]

const sampleCompatibility: FunctionCompatibility[] = [
  {
    functionId: "fn-1",
    functionName: "DataTransformer",
    runtimes: [
      { runtime: "nodejs", version: "18", isCompatible: true },
      { runtime: "python", version: "3.11", isCompatible: true },
      { runtime: "go", version: "1.21", isCompatible: true, notes: "Requires CGO" },
      { runtime: "rust", version: "1.70", isCompatible: false, notes: "No WASM support yet" },
      { runtime: "deno", version: "1.37", isCompatible: true },
    ],
  },
  {
    functionId: "fn-2",
    functionName: "ML inference",
    runtimes: [
      { runtime: "nodejs", version: "18", isCompatible: false, notes: "Memory constraints" },
      { runtime: "python", version: "3.11", isCompatible: true },
      { runtime: "go", version: "1.21", isCompatible: false, notes: "No ML libraries" },
      { runtime: "rust", version: "1.70", isCompatible: true },
      { runtime: "deno", version: "1.37", isCompatible: false, notes: "No WASM support" },
    ],
  },
]

const sampleWorkflows: WorkflowTemplate[] = [
  { id: "wf-1", name: "ETL Pipeline", description: "Extract, transform, and load data from multiple sources", author: "data-team", category: "data-processing", steps: 5, executions: 12500, rating: 4.8, tags: ["etl", "streaming"], isFeatured: true },
  { id: "wf-2", name: "ML Training Pipeline", description: "End-to-end machine learning workflow with model serving", author: "ml-team", category: "ai-ml", steps: 8, executions: 3200, rating: 4.6, tags: ["ml", "training"] },
  { id: "wf-3", name: "Slack Notifications", description: "Send alerts to Slack based on function events", author: "devops", category: "integrations", steps: 2, executions: 45000, rating: 4.9, tags: ["slack", "notifications"] },
]

const sampleSandboxInputs: SandboxInput[] = [
  { name: "inputData", type: "string", value: '{"records": [{"id": 1, "value": "sample"}]}', required: true },
  { name: "options", type: "object", value: '{"strict": true, "timeout": 30}', required: false },
]

// ============================================================================
// Component
// ============================================================================

export function RegistryIntegration({
  functions = [],
  isLoading = false,
  onFunctionSelect,
  onFunctionExecute,
  onFunctionFavorite,
  onFunctionShare,
  className,
}: RegistryIntegrationProps) {
  const {
    searchQuery,
    selectedCategory,
    sortBy,
    filters,
    selectedFunction,
    selectedVersion,
    setSearchQuery,
    setSelectedCategory,
    setSortBy,
    setFilters,
    setSelectedFunction,
    setSelectedVersion,
    sandboxInputs,
    setSandboxInputs,
  } = useRegistryStore()

  const [activeTab, setActiveTab] = React.useState("marketplace")
  const [installModalOpen, setInstallModalOpen] = React.useState(false)
  const [functionToInstall, setFunctionToInstall] = React.useState<FunctionCardData | undefined>()

  // Handle function selection
  const handleFunctionSelect = (id: string) => {
    const fn = functions.find((f) => f.id === id)
    setSelectedFunction(fn ?? null)
    onFunctionSelect?.(id)
  }

  // Handle function install
  const handleInstall = (fn: FunctionCardData) => {
    setFunctionToInstall(fn)
    setInstallModalOpen(true)
  }

  // Initialize sandbox inputs when function is selected
  React.useEffect(() => {
    if (selectedFunction) {
      setSandboxInputs(sampleSandboxInputs)
    }
  }, [selectedFunction, setSandboxInputs])

  return (
    <div className={cn("space-y-6", className)}>
      {/* Search and Filter Bar */}
      <RegistrySearch
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        selectedCategory={selectedCategory}
        onCategorySelect={setSelectedCategory}
        sortBy={sortBy}
        onSortChange={setSortBy}
        filters={filters}
        onFilterChange={setFilters}
      />

      {/* Main Content Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="marketplace">Marketplace</TabsTrigger>
          <TabsTrigger value="explorer">Category Explorer</TabsTrigger>
          <TabsTrigger value="workflows">Workflows</TabsTrigger>
          {selectedFunction && (
            <>
              <TabsTrigger value="detail">Function Detail</TabsTrigger>
              <TabsTrigger value="sandbox">Sandbox</TabsTrigger>
              <TabsTrigger value="analytics">Analytics</TabsTrigger>
            </>
          )}
        </TabsList>

        {/* Marketplace Tab */}
        <TabsContent value="marketplace">
          <FunctionMarketplace
            functions={functions}
            onSelect={handleFunctionSelect}
            onExecute={onFunctionExecute}
            onFavorite={onFunctionFavorite}
            onShare={onFunctionShare}
            searchQuery={searchQuery}
            onSearchChange={setSearchQuery}
            categoryFilter={selectedCategory}
            onCategoryChange={setSelectedCategory}
            isLoading={isLoading}
          />
        </TabsContent>

        {/* Category Explorer Tab */}
        <TabsContent value="explorer">
          <RegistryCategoryExplorer
            categories={sampleCategories}
            onCategorySelect={(id) => setSelectedCategory(id)}
            selectedCategory={selectedCategory}
          />
        </TabsContent>

        {/* Workflows Tab */}
        <TabsContent value="workflows">
          <div className="space-y-6">
            <PublicWorkflowExplorer
              workflows={sampleWorkflows}
              onSelect={(id) => console.log("Selected workflow:", id)}
              onUse={(id) => console.log("Using workflow:", id)}
            />
            <WorkflowTemplateGallery
              categories={[
                { id: "featured", name: "Featured Templates", templates: sampleWorkflows.filter((w) => w.isFeatured) },
                { id: "popular", name: "Popular Templates", templates: sampleWorkflows },
              ]}
              onSelect={(id) => console.log("Selected template:", id)}
              onUse={(id) => console.log("Using template:", id)}
            />
          </div>
        </TabsContent>

        {/* Function Detail Tab */}
        {selectedFunction && (
          <>
            <TabsContent value="detail">
              <div className="space-y-6">
                {/* Function Card */}
                <FunctionCard
                  {...selectedFunction}
                  variant="expanded"
                  onView={() => { }}
                  onExecute={() => onFunctionExecute?.(selectedFunction.id)}
                  onFavorite={() => onFunctionFavorite?.(selectedFunction.id, !selectedFunction.isFavorite)}
                  onShare={() => onFunctionShare?.(selectedFunction.id)}
                />

                {/* Runtime Compatibility */}
                <RuntimeCompatibilityMatrix
                  functions={sampleCompatibility}
                  selectedFunctionId={selectedFunction.id}
                  onFunctionSelect={(id) => {
                    const fn = functions.find((f) => f.id === id)
                    if (fn) setSelectedFunction(fn)
                  }}
                />

                {/* Version Diff */}
                <VersionDiffViewer
                  oldVersion={{ version: "1.0.0", timestamp: "2024-01-01", author: "author" }}
                  newVersion={{ version: selectedFunction.version ?? "2.0.0", timestamp: new Date().toISOString(), author: "author" }}
                  oldCode={`// Version 1.0.0
export function transform(data) {
  return data.map(item => ({
    ...item,
    processed: true
  }));
}`}
                  newCode={`// Version 2.0.0
export function transform(data, options = {}) {
  return data.map(item => ({
    ...item,
    processed: true,
    timestamp: Date.now(),
    ...options
  }));
}`}
                  onRestoreVersion={(v) => console.log("Restoring to version:", v)}
                />

                {/* Changelog */}
                <FunctionChangelog
                  functionId={selectedFunction.id}
                  functionName={selectedFunction.name}
                  entries={sampleChangelog}
                  onVersionSelect={setSelectedVersion}
                />

                {/* Dependency Viewer */}
                <DependencyViewer
                  nodes={[
                    { id: "fn", name: selectedFunction.name, version: selectedFunction.version ?? "1.0.0", isExternal: false },
                    { id: "lodash", name: "lodash", version: "4.17.21", isExternal: true },
                    { id: "axios", name: "axios", version: "1.6.0", isExternal: true },
                    { id: "pg", name: "pg", version: "8.11.0", isExternal: true },
                  ]}
                  edges={[
                    { from: "fn", to: "lodash", type: "npm" },
                    { from: "fn", to: "axios", type: "npm" },
                    { from: "fn", to: "pg", type: "npm" },
                  ]}
                />

                {/* Package Integrity */}
                <PackageIntegrityViewer
                  functionId={selectedFunction.id}
                  functionName={selectedFunction.name}
                  version={selectedFunction.version ?? "1.0.0"}
                  checks={sampleIntegrityChecks}
                  onReverify={() => console.log("Reverifying...")}
                />

                {/* Permission Audit */}
                <FunctionPermissionAudit
                  functionId={selectedFunction.id}
                  functionName={selectedFunction.name}
                  permissions={samplePermissions}
                  onRevoke={(id) => console.log("Revoking permission:", id)}
                  onAdd={() => console.log("Adding permission...")}
                />
              </div>
            </TabsContent>

            {/* Sandbox Tab */}
            <TabsContent value="sandbox">
              <FunctionSandbox
                functionId={selectedFunction.id}
                functionName={selectedFunction.name}
                inputs={sandboxInputs.length > 0 ? sandboxInputs : sampleSandboxInputs}
                onRun={(inputs) => console.log("Running with inputs:", inputs)}
              />
            </TabsContent>

            {/* Analytics Tab */}
            <TabsContent value="analytics">
              <FunctionUsageAnalytics
                functionId={selectedFunction.id}
                functionName={selectedFunction.name}
                usageData={sampleUsageData}
                totalExecutions={selectedFunction.metrics.executionCount}
                avgLatency={selectedFunction.metrics.averageLatency ?? 45}
                errorRate={selectedFunction.metrics.errorRate ?? 0.5}
                topUsers={[
                  { userId: "u1", name: "Acme Corp", calls: 12500 },
                  { userId: "u2", name: "Beta Inc", calls: 8300 },
                  { userId: "u3", name: "Gamma LLC", calls: 4100 },
                ]}
              />

              {/* Benchmarks */}
              <FunctionBenchmarkPanel
                functionId={selectedFunction.id}
                functionName={selectedFunction.name}
                benchmarks={sampleBenchmarks}
                onRunBenchmark={() => console.log("Running benchmarks...")}
                onSetBaseline={(name) => console.log("Setting baseline:", name)}
              />
            </TabsContent>
          </>
        )}
      </Tabs>

      {/* Install Modal */}
      <InstallFunctionModal
        isOpen={installModalOpen}
        onClose={() => setInstallModalOpen(false)}
        onInstall={() => console.log("Installing function:", functionToInstall?.name)}
        functionData={functionToInstall}
      />
    </div>
  )
}

// ============================================================================
// Export all registry components for direct access
// ============================================================================

export {
  FunctionCard,
  FunctionMarketplace,
  InstallFunctionModal,
  RevenueAnalytics,
  SubscriptionManager,
  UsageBillingPanel,
  CreatorProfile,
  MarketplaceLeaderboard,
  FunctionRoyaltiesPanel,
  AssetPricingEditor,
  LicenseManager,
  MonetizationOptimizer,
  RegistrySearch,
  RegistryCategoryExplorer,
  RuntimeCompatibilityMatrix,
  DependencyViewer,
  FunctionSandbox,
  VersionDiffViewer,
  FunctionUsageAnalytics,
  PublicWorkflowExplorer,
  WorkflowTemplateGallery,
  PackageIntegrityViewer,
  FunctionPermissionAudit,
  FunctionChangelog,
  FunctionBenchmarkPanel,
} from "@functionfly/ui-marketplace"

export type {
  FunctionCardData,
  RegistryFilters,
  CategoryNode,
  FunctionCompatibility,
  WorkflowTemplate,
  IntegrityCheck,
  PermissionEntry,
  ChangelogEntry,
  BenchmarkResult,
  UsageDataPoint,
  SandboxInput,
} from "@functionfly/ui-marketplace"
