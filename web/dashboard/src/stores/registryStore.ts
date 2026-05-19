/**
 * Registry Store
 * Global state management for Function Registry & Marketplace components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'

// ============================================================================
// Types
// ============================================================================

export interface FunctionAuthor {
  id: string
  username: string
  name: string
  avatar?: string
  profileUrl?: string
}

export interface FunctionMetrics {
  executionCount: number
  executionTrend?: number[]
  averageLatency?: number
  errorRate?: number
}

export interface FunctionRating {
  average: number
  count: number
  distribution?: Record<number, number>
}

export type PricingModel = "free" | "per_call" | "subscription" | "revenue_share"

export interface RegistryFunction {
  id: string
  name: string
  description: string
  author: FunctionAuthor
  trustScore: number
  metrics: FunctionMetrics
  pricing: {
    model: PricingModel
    pricePerCall?: number
    currency?: string
  }
  isVerified: boolean
  isDeterministic: boolean
  rating: FunctionRating
  tags?: string[]
  category?: string
  language?: string
  lastUpdated?: string
  version?: string
  isFavorite?: boolean
  isFeatured?: boolean
}

export interface DependencyNode {
  id: string
  name: string
  version: string
  isExternal: boolean
}

export interface DependencyEdge {
  from: string
  to: string
  type: "runtime" | "npm" | "go" | "pip" | "cargo"
  isExternal?: boolean
}

export interface RuntimeVersion {
  runtime: string
  version: string
  isCompatible: boolean
  minVersion?: string
  notes?: string
}

export interface IntegrityCheck {
  id: string
  name: string
  status: "passed" | "warning" | "failed"
  description: string
  details?: string
}

export interface PermissionEntry {
  id: string
  principal: string
  principalType: "user" | "team" | "api_key" | "public"
  permissions: string[]
  grantedAt: string
  expiresAt?: string
  isActive: boolean
}

export interface ChangelogEntry {
  version: string
  date: string
  author: string
  type: "major" | "minor" | "patch" | "security"
  changes: string[]
}

export interface BenchmarkResult {
  name: string
  iterations: number
  avgMs: number
  p50Ms: number
  p95Ms: number
  p99Ms: number
  minMs: number
  maxMs: number
  stdDev: number
  isBaseline?: boolean
}

export interface UsageDataPoint {
  timestamp: string
  executions: number
  errors: number
  latency: number
}

export interface SandboxInput {
  name: string
  type: string
  value: string
  required: boolean
}

export interface SandboxResult {
  success: boolean
  output: string
  executionTime: number
  memoryUsed?: number
  error?: string
}

export interface CategoryNode {
  id: string
  name: string
  description: string
  icon: React.ReactNode
  count: number
  subcategories?: CategoryNode[]
  trending?: boolean
}

// ============================================================================
// Store Interface
// ============================================================================

interface RegistryState {
  // Search & Filter
  searchQuery: string
  selectedCategory: string
  sortBy: "popular" | "recent" | "rating" | "price"
  filters: {
    isVerified?: boolean
    isFree?: boolean
    minTrustScore?: number
    maxPrice?: number
    language?: string
    runtime?: string
  }

  // Selected function
  selectedFunctionId: string | null
  selectedFunction: RegistryFunction | null

  // Function versions
  selectedVersion: string | null
  oldVersionCode: string
  newVersionCode: string

  // Sandbox
  sandboxInputs: SandboxInput[]
  sandboxResult: SandboxResult | null
  isRunningSandbox: boolean

  // Dependencies
  dependencyNodes: DependencyNode[]
  dependencyEdges: DependencyEdge[]

  // Runtime compatibility
  runtimeVersions: RuntimeVersion[]

  // Integrity
  integrityChecks: IntegrityCheck[]

  // Permissions
  permissions: PermissionEntry[]

  // Changelog
  changelogEntries: ChangelogEntry[]

  // Benchmarks
  benchmarkResults: BenchmarkResult[]
  selectedBenchmark: string | null

  // Usage analytics
  usageData: UsageDataPoint[]

  // Categories (for category explorer)
  categories: CategoryNode[]

  // Actions
  setSearchQuery: (query: string) => void
  setSelectedCategory: (category: string) => void
  setSortBy: (sort: "popular" | "recent" | "rating" | "price") => void
  setFilters: (filters: Partial<RegistryState['filters']>) => void
  clearFilters: () => void

  setSelectedFunction: (fn: RegistryFunction | null) => void
  setSelectedVersion: (version: string | null) => void
  setVersionCode: (which: "old" | "new", code: string) => void

  setSandboxInputs: (inputs: SandboxInput[]) => void
  setSandboxResult: (result: SandboxResult | null) => void
  setIsRunningSandbox: (running: boolean) => void
  runSandbox: () => Promise<void>

  setDependencyData: (nodes: DependencyNode[], edges: DependencyEdge[]) => void
  setRuntimeVersions: (versions: RuntimeVersion[]) => void
  setIntegrityChecks: (checks: IntegrityCheck[]) => void
  setPermissions: (permissions: PermissionEntry[]) => void
  setChangelogEntries: (entries: ChangelogEntry[]) => void
  setBenchmarkResults: (results: BenchmarkResult[]) => void
  setSelectedBenchmark: (name: string | null) => void
  setUsageData: (data: UsageDataPoint[]) => void
  setCategories: (categories: CategoryNode[]) => void

  // Permission actions
  revokePermission: (permissionId: string) => void

  // Benchmark actions
  runBenchmark: () => Promise<void>
  setBaseline: (benchmarkName: string) => void

  // Integrity
  verifyIntegrity: () => Promise<void>

  // Reset
  reset: () => void
}

// ============================================================================
// Initial State
// ============================================================================

const initialState = {
  searchQuery: "",
  selectedCategory: "all",
  sortBy: "popular" as const,
  filters: {} as RegistryState['filters'],
  selectedFunctionId: null,
  selectedFunction: null,
  selectedVersion: null,
  oldVersionCode: "",
  newVersionCode: "",
  sandboxInputs: [],
  sandboxResult: null,
  isRunningSandbox: false,
  dependencyNodes: [],
  dependencyEdges: [],
  runtimeVersions: [],
  integrityChecks: [],
  permissions: [],
  changelogEntries: [],
  benchmarkResults: [],
  selectedBenchmark: null,
  usageData: [],
  categories: [],
}

// ============================================================================
// Store
// ============================================================================

export const useRegistryStore = create<RegistryState>()(
  immer((set, get) => ({
    ...initialState,

    // Search & Filter
    setSearchQuery: (query) => set((state) => {
      state.searchQuery = query
    }),
    setSelectedCategory: (category) => set((state) => {
      state.selectedCategory = category
    }),
    setSortBy: (sort) => set((state) => {
      state.sortBy = sort
    }),
    setFilters: (filters) => set((state) => {
      Object.assign(state.filters, filters)
    }),
    clearFilters: () => set((state) => {
      state.filters = {}
    }),

    setSelectedFunction: (fn) => set((state) => {
      state.selectedFunction = fn
      state.selectedFunctionId = fn?.id ?? null
    }),
    setSelectedVersion: (version) => set((state) => {
      state.selectedVersion = version
    }),
    setVersionCode: (which, code) => set((state) => {
      if (which === "old") state.oldVersionCode = code
      else state.newVersionCode = code
    }),

    setSandboxInputs: (inputs) => set((state) => {
      state.sandboxInputs = inputs
    }),
    setSandboxResult: (result) => set((state) => {
      state.sandboxResult = result
    }),
    setIsRunningSandbox: (running) => set((state) => {
      state.isRunningSandbox = running
    }),
    runSandbox: async () => {
      const { sandboxInputs, selectedFunction } = get()
      if (!selectedFunction) return

      set((state) => { state.isRunningSandbox = true })

      // Simulate execution
      await new Promise((resolve) => setTimeout(resolve, 1500))

      const result: SandboxResult = {
        success: Math.random() > 0.1,
        output: JSON.stringify({ message: "Function executed successfully", data: sandboxInputs.reduce((acc, inp) => ({ ...acc, [inp.name]: inp.value }), {}) }, null, 2),
        executionTime: Math.random() * 200 + 50,
        memoryUsed: Math.random() * 50 + 10,
      }

      set((state) => {
        state.sandboxResult = result
        state.isRunningSandbox = false
      })
    },

    setDependencyData: (nodes, edges) => set((state) => {
      state.dependencyNodes = nodes
      state.dependencyEdges = edges
    }),
    setRuntimeVersions: (versions) => set((state) => {
      state.runtimeVersions = versions
    }),
    setIntegrityChecks: (checks) => set((state) => {
      state.integrityChecks = checks
    }),
    setPermissions: (permissions) => set((state) => {
      state.permissions = permissions
    }),
    setChangelogEntries: (entries) => set((state) => {
      state.changelogEntries = entries
    }),
    setBenchmarkResults: (results) => set((state) => {
      state.benchmarkResults = results
    }),
    setSelectedBenchmark: (name) => set((state) => {
      state.selectedBenchmark = name
    }),
    setUsageData: (data) => set((state) => {
      state.usageData = data
    }),
    setCategories: (categories) => set((state) => {
      state.categories = categories
    }),

    revokePermission: (permissionId) => set((state) => {
      const perm = state.permissions.find(p => p.id === permissionId)
      if (perm) perm.isActive = false
    }),

    runBenchmark: async () => {
      await new Promise((resolve) => setTimeout(resolve, 3000))
      // Benchmark results would be updated via API in real implementation
    },

    setBaseline: (benchmarkName) => set((state) => {
      state.benchmarkResults.forEach(b => {
        b.isBaseline = b.name === benchmarkName
      })
    }),

    verifyIntegrity: async () => {
      await new Promise((resolve) => setTimeout(resolve, 2000))
      // Would update via API in real implementation
    },

    reset: () => set(() => ({ ...initialState })),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const selectFilteredFunctions = (functions: RegistryFunction[], state: RegistryState) => {
  return functions.filter((fn) => {
    const matchesSearch = !state.searchQuery ||
      fn.name.toLowerCase().includes(state.searchQuery.toLowerCase()) ||
      fn.description.toLowerCase().includes(state.searchQuery.toLowerCase())
    const matchesCategory = state.selectedCategory === "all" || fn.category === state.selectedCategory
    const matchesVerified = !state.filters.isVerified || fn.isVerified
    const matchesFree = !state.filters.isFree || fn.pricing.model === "free"
    const matchesTrust = !state.filters.minTrustScore || fn.trustScore >= state.filters.minTrustScore
    const matchesPrice = !state.filters.maxPrice || !fn.pricing.pricePerCall || fn.pricing.pricePerCall <= state.filters.maxPrice
    const matchesLanguage = !state.filters.language || fn.language === state.filters.language
    return matchesSearch && matchesCategory && matchesVerified && matchesFree && matchesTrust && matchesPrice && matchesLanguage
  })
}

export const selectSortedFunctions = (functions: RegistryFunction[], sortBy: RegistryState['sortBy']) => {
  return [...functions].sort((a, b) => {
    switch (sortBy) {
      case "popular":
        return b.metrics.executionCount - a.metrics.executionCount
      case "recent":
        return new Date(b.lastUpdated ?? 0).getTime() - new Date(a.lastUpdated ?? 0).getTime()
      case "rating":
        return b.rating.average - a.rating.average
      case "price":
        return (a.pricing.pricePerCall ?? 0) - (b.pricing.pricePerCall ?? 0)
      default:
        return 0
    }
  })
}
