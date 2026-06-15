/**
 * Marketplace Economy Store
 * Global state management for Marketplace Economy components
 */

import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import type {
  CreatorEarnings,
  RevenueDataPoint,
  Subscription,
  SubscriptionStatus,
  BillingCycle,
  LicenseType,
  License,
  CreatorStats,
  PayoutInfo,
  RoyaltyRecord,
  LeaderboardEntry,
  PricingTier,
  ConversionFunnelStep,
  OptimizationSuggestion,
  TrendItem,
  TrendDirection,
} from '@functionfly/ui-marketplace-economy'

// ============================================================================
// Types
// ============================================================================

export interface Creator {
  id: string
  name: string
  avatar?: string | null
  stats: CreatorStats
  payoutInfo: PayoutInfo
}

export interface RevenueData {
  total: number
  breakdown: {
    subscription: number
    oneTime: number
    royalty: number
    tier: number
  }
  dataPoints: RevenueDataPoint[]
  period: '7d' | '30d' | '90d' | '1y'
}

export interface MarketplaceAlert {
  id: string
  type: 'warning' | 'error' | 'info' | 'success'
  title: string
  message: string
  dismissed: boolean
  timestamp: number
}

export type MarketplaceView = 'economy' | 'revenue' | 'subscriptions' | 'billing' | 'licenses' | 'profile' | 'leaderboard' | 'royalties' | 'pricing' | 'conversion' | 'optimizer' | 'trends'

export interface MarketplaceEconomyState {
  creators: Creator[]
  subscriptions: Subscription[]
  revenue: RevenueData
  selectedCreatorId: string | null
  activeView: MarketplaceView
  dateRange: { start: Date; end: Date }
  alerts: MarketplaceAlert[]
}

// ============================================================================
// Store
// ============================================================================

interface MarketplaceEconomyActions {
  selectCreator: (creatorId: string | null) => void
  setActiveView: (view: MarketplaceView) => void
  setDateRange: (range: { start: Date; end: Date }) => void
  dismissAlert: (alertId: string) => void
  updateRevenue: (revenue: Partial<RevenueData>) => void
  setCreators: (creators: Creator[]) => void
  setSubscriptions: (subscriptions: Subscription[]) => void
  addSubscription: (subscription: Subscription) => void
  updateSubscription: (subscriptionId: string, updates: Partial<Subscription>) => void
  addAlert: (alert: MarketplaceAlert) => void
  clearAlerts: () => void
}

export const useMarketplaceEconomyStore = create<MarketplaceEconomyState & MarketplaceEconomyActions>()(
  immer((set) => ({
    creators: [],
    subscriptions: [],
    revenue: {
      total: 0,
      breakdown: { subscription: 0, oneTime: 0, royalty: 0, tier: 0 },
      dataPoints: [],
      period: '30d',
    },
    selectedCreatorId: null,
    activeView: 'economy',
    dateRange: {
      start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
      end: new Date(),
    },
    alerts: [],

    selectCreator: (creatorId) => set((state) => {
      state.selectedCreatorId = creatorId
    }),

    setActiveView: (view) => set((state) => {
      state.activeView = view
    }),

    setDateRange: (range) => set((state) => {
      state.dateRange = range
    }),

    dismissAlert: (alertId) => set((state) => {
      state.alerts = state.alerts.filter(a => a.id !== alertId)
    }),

    updateRevenue: (revenue) => set((state) => {
      Object.assign(state.revenue, revenue)
    }),

    setCreators: (creators) => set((state) => {
      state.creators = creators
    }),

    setSubscriptions: (subscriptions) => set((state) => {
      state.subscriptions = subscriptions
    }),

    addSubscription: (subscription) => set((state) => {
      state.subscriptions.push(subscription)
    }),

    updateSubscription: (subscriptionId, updates) => set((state) => {
      const sub = state.subscriptions.find(s => s.id === subscriptionId)
      if (sub) Object.assign(sub, updates)
    }),

    addAlert: (alert) => set((state) => {
      state.alerts.push(alert)
    }),

    clearAlerts: () => set((state) => {
      state.alerts = []
    }),
  }))
)

// ============================================================================
// Selectors
// ============================================================================

export const selectTopCreators = (state: MarketplaceEconomyState) =>
  [...state.creators].sort((a, b) => (b.stats?.totalRevenue || 0) - (a.stats?.totalRevenue || 0)).slice(0, 10)

export const selectActiveSubscriptions = (state: MarketplaceEconomyState) =>
  state.subscriptions.filter(s => s.status === 'active')

export const selectTotalRevenue = (state: MarketplaceEconomyState) =>
  state.revenue.total

export const selectAlertCount = (state: MarketplaceEconomyState) =>
  state.alerts.filter(a => !a.dismissed).length

// ============================================================================
// Custom Hooks
// ============================================================================

export const useMarketplaceEconomy = () => {
  const store = useMarketplaceEconomyStore()
  return {
    ...store,
    topCreators: selectTopCreators(store),
    activeSubscriptions: selectActiveSubscriptions(store),
    totalRevenue: selectTotalRevenue(store),
    alertCount: selectAlertCount(store),
  }
}

export const useRevenueAnalytics = () => useMarketplaceEconomyStore((state) => ({
  revenue: state.revenue,
  dateRange: state.dateRange,
  updateRevenue: state.updateRevenue,
  setDateRange: state.setDateRange,
}))

export const useSubscriptions = () => useMarketplaceEconomyStore((state) => ({
  subscriptions: state.subscriptions,
  activeCount: state.subscriptions.filter(s => s.status === 'active').length,
  cancelledCount: state.subscriptions.filter(s => s.status === 'cancelled').length,
  pastDueCount: state.subscriptions.filter(s => s.status === 'past_due').length,
  setSubscriptions: state.setSubscriptions,
  addSubscription: state.addSubscription,
  updateSubscription: state.updateSubscription,
}))

export const useUsageBilling = () => useMarketplaceEconomyStore((state) => ({
  dateRange: state.dateRange,
  activeView: state.activeView,
}))

export const useLicenses = () => useMarketplaceEconomyStore((state) => ({
  selectedCreatorId: state.selectedCreatorId,
  selectCreator: state.selectCreator,
}))

export const useCreatorProfile = () => useMarketplaceEconomyStore((state) => ({
  creators: state.creators,
  selectedCreatorId: state.selectedCreatorId,
  selectCreator: state.selectCreator,
}))

export const useLeaderboard = () => useMarketplaceEconomyStore((state) => ({
  creators: state.creators,
  activeView: state.activeView,
}))

export const useRoyalties = () => useMarketplaceEconomyStore((state) => ({
  selectedCreatorId: state.selectedCreatorId,
  selectCreator: state.selectCreator,
}))

export const useAssetPricing = () => useMarketplaceEconomyStore((state) => ({
  selectedCreatorId: state.selectedCreatorId,
  selectCreator: state.selectCreator,
}))

export const useConversionAnalytics = () => useMarketplaceEconomyStore((state) => ({
  activeView: state.activeView,
  dateRange: state.dateRange,
}))

export const useOptimizer = () => useMarketplaceEconomyStore((state) => ({
  activeView: state.activeView,
  revenue: state.revenue,
}))

export const useTrendRadar = () => useMarketplaceEconomyStore((state) => ({
  activeView: state.activeView,
  dateRange: state.dateRange,
}))
