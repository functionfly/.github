/**
 * Marketplace Economy Page
 * Main page component for the Marketplace Economy section
 */

import { useEffect } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { Calendar, TrendingUp, DollarSign, Users, BarChart3, Zap, Settings } from 'lucide-react'
import {
  CreatorEconomy,
  RevenueAnalytics,
  SubscriptionManager,
  UsageBillingPanel,
  LicenseManager,
  CreatorProfile,
  MarketplaceLeaderboard,
  FunctionRoyaltiesPanel,
  AssetPricingEditor,
  SalesConversionAnalytics,
  MonetizationOptimizer,
  MarketplaceTrendRadar,
} from '@functionfly/ui-marketplace-economy/components'
import { useMarketplaceEconomyStore } from '@/stores/marketplaceEconomyStore'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'

const VIEW_LABELS: Record<string, { label: string; icon: React.ComponentType<{ className?: string }> }> = {
  economy: { label: 'Economy', icon: DollarSign },
  revenue: { label: 'Revenue', icon: TrendingUp },
  subscriptions: { label: 'Subscriptions', icon: Users },
  billing: { label: 'Billing', icon: Calendar },
  licenses: { label: 'Licenses', icon: BarChart3 },
  profile: { label: 'Profile', icon: Users },
  leaderboard: { label: 'Leaderboard', icon: TrendingUp },
  royalties: { label: 'Royalties', icon: DollarSign },
  pricing: { label: 'Pricing', icon: Zap },
  conversion: { label: 'Conversion', icon: BarChart3 },
  optimizer: { label: 'Optimizer', icon: Settings },
  trends: { label: 'Trends', icon: TrendingUp },
}

const PERIOD_OPTIONS = [
  { value: '7d', label: 'Last 7 days' },
  { value: '30d', label: 'Last 30 days' },
  { value: '90d', label: 'Last 90 days' },
  { value: '1y', label: 'Last year' },
]

const DATE_RANGE_OPTIONS = [
  { value: '7d', label: '7 Days' },
  { value: '30d', label: '30 Days' },
  { value: '90d', label: '90 Days' },
]

export function MarketplaceEconomyPage() {
  const { t } = useTranslation()
  const { panel } = useParams<{ panel?: string }>()
  const [searchParams, setSearchParams] = useSearchParams()

  const {
    activeView,
    setActiveView,
    dateRange,
    setDateRange,
    revenue,
    creators,
    subscriptions,
    alerts,
    dismissAlert,
    selectedCreatorId,
    selectCreator,
  } = useMarketplaceEconomyStore()

  useEffect(() => {
    if (panel && Object.keys(VIEW_LABELS).includes(panel)) {
      setActiveView(panel as keyof typeof VIEW_LABELS)
    }
  }, [panel, setActiveView])

  const handleViewChange = (view: string) => {
    setActiveView(view as keyof typeof VIEW_LABELS)
    setSearchParams({ panel: view })
  }

  const handlePeriodChange = (period: string) => {
    const now = new Date()
    let start: Date
    switch (period) {
      case '7d':
        start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case '30d':
        start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
      case '90d':
        start = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000)
        break
      case '1y':
        start = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000)
        break
      default:
        start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    }
    setDateRange({ start, end: now })
  }

  const activeAlerts = alerts.filter(a => !a.dismissed)

  const selectedCreator = creators.find(c => c.id === selectedCreatorId)

  return (
    <div className="aviation-dashboard min-h-screen bg-aviation-bg-primary">
      {/* Header */}
      <div className="aviation-header border-b border-aviation-border-panel bg-aviation-bg-panel/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="aviation-container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="aviation-title text-2xl font-bold text-aviation-text-primary">
                Marketplace Economy
              </h1>
              <p className="aviation-subtitle text-sm text-aviation-text-secondary mt-1">
                Monitor your creator earnings, royalties, and marketplace performance
              </p>
            </div>

            {/* Date Range Selector */}
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <Calendar className="w-4 h-4 text-aviation-text-muted" />
                <Select
                  value={revenue.period}
                  onValueChange={handlePeriodChange}
                >
                  <SelectTrigger className="aviation-select w-[160px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {PERIOD_OPTIONS.map(opt => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>

          {/* Summary Stats */}
          <div className="aviation-stats-grid grid grid-cols-4 gap-4 mt-6">
            <div className="aviation-stat-card bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg p-4">
              <div className="flex items-center gap-3">
                <div className="aviation-stat-icon bg-aviation-amber/10 text-aviation-amber rounded-lg p-2">
                  <DollarSign className="w-5 h-5" />
                </div>
                <div>
                  <p className="text-xs text-aviation-text-muted uppercase tracking-wider">Total Revenue</p>
                  <p className="text-2xl font-bold text-aviation-text-primary">${revenue.total.toLocaleString()}</p>
                </div>
              </div>
            </div>

            <div className="aviation-stat-card bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg p-4">
              <div className="flex items-center gap-3">
                <div className="aviation-stat-icon bg-aviation-cyan/10 text-aviation-cyan rounded-lg p-2">
                  <Users className="w-5 h-5" />
                </div>
                <div>
                  <p className="text-xs text-aviation-text-muted uppercase tracking-wider">Active Creators</p>
                  <p className="text-2xl font-bold text-aviation-text-primary">{creators.length}</p>
                </div>
              </div>
            </div>

            <div className="aviation-stat-card bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg p-4">
              <div className="flex items-center gap-3">
                <div className="aviation-stat-icon bg-aviation-emerald/10 text-aviation-emerald rounded-lg p-2">
                  <Users className="w-5 h-5" />
                </div>
                <div>
                  <p className="text-xs text-aviation-text-muted uppercase tracking-wider">Active Subscriptions</p>
                  <p className="text-2xl font-bold text-aviation-text-primary">
                    {subscriptions.filter(s => s.status === 'active').length}
                  </p>
                </div>
              </div>
            </div>

            <div className="aviation-stat-card bg-aviation-bg-instrument/50 border border-aviation-border-instrument rounded-lg p-4">
              <div className="flex items-center gap-3">
                <div className="aviation-stat-icon bg-aviation-rose/10 text-aviation-rose rounded-lg p-2">
                  <TrendingUp className="w-5 h-5" />
                </div>
                <div>
                  <p className="text-xs text-aviation-text-muted uppercase tracking-wider">Alerts</p>
                  <p className="text-2xl font-bold text-aviation-text-primary">{activeAlerts.length}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Alerts Banner */}
      {activeAlerts.length > 0 && (
        <div className="aviation-alerts-banner mx-6 mt-6 bg-aviation-bg-panel border border-aviation-amber/30 rounded-lg p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="aviation-alert-icon bg-aviation-amber/10 text-aviation-amber rounded-lg p-2">
                <Zap className="w-4 h-4" />
              </div>
              <div>
                <p className="text-sm font-medium text-aviation-text-primary">You have {activeAlerts.length} pending alert{activeAlerts.length !== 1 ? 's' : ''}</p>
                <p className="text-xs text-aviation-text-secondary">{activeAlerts[0].message}</p>
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => dismissAlert(activeAlerts[0].id)}
              className="text-aviation-text-muted hover:text-aviation-text-primary"
            >
              Dismiss
            </Button>
          </div>
        </div>
      )}

      {/* Navigation Tabs */}
      <div className="aviation-tabs-container mx-6 mt-6">
        <Tabs
          value={activeView}
          onValueChange={handleViewChange}
          className="w-full"
        >
          <TabsList className="aviation-tabs-list bg-aviation-bg-panel border border-aviation-border-panel rounded-lg p-1 inline-flex">
            {Object.entries(VIEW_LABELS).map(([key, { label, icon: Icon }]) => (
              <TabsTrigger
                key={key}
                value={key}
                className={cn(
                  'aviation-tab',
                  activeView === key && 'aviation-tab-active'
                )}
              >
                <Icon className="w-4 h-4 mr-2" />
                {label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>

      {/* Main Content Area */}
      <div className="aviation-content mx-6 mt-6 mb-6">
        <div className="grid grid-cols-12 gap-6">
          {/* Sidebar */}
          <div className="col-span-3">
            <div className="aviation-sidebar-panel bg-aviation-bg-panel border border-aviation-border-panel rounded-lg p-4 sticky top-40">
              <h3 className="text-sm font-semibold text-aviation-text-primary mb-4">Quick Actions</h3>
              <div className="space-y-2">
                <Button
                  variant="ghost"
                  className="aviation-btn-ghost w-full justify-start text-left"
                  onClick={() => handleViewChange('economy')}
                >
                  <DollarSign className="w-4 h-4 mr-3 text-aviation-amber" />
                  View Economy Overview
                </Button>
                <Button
                  variant="ghost"
                  className="aviation-btn-ghost w-full justify-start text-left"
                  onClick={() => handleViewChange('revenue')}
                >
                  <TrendingUp className="w-4 h-4 mr-3 text-aviation-cyan" />
                  Revenue Analytics
                </Button>
                <Button
                  variant="ghost"
                  className="aviation-btn-ghost w-full justify-start text-left"
                  onClick={() => handleViewChange('leaderboard')}
                >
                  <Users className="w-4 h-4 mr-3 text-aviation-emerald" />
                  Top Creators
                </Button>
                <Button
                  variant="ghost"
                  className="aviation-btn-ghost w-full justify-start text-left"
                  onClick={() => handleViewChange('royalties')}
                >
                  <DollarSign className="w-4 h-4 mr-3 text-aviation-rose" />
                  Royalty Reports
                </Button>
                <Button
                  variant="ghost"
                  className="aviation-btn-ghost w-full justify-start text-left"
                  onClick={() => handleViewChange('optimizer')}
                >
                  <Zap className="w-4 h-4 mr-3 text-aviation-amber" />
                  Optimization Tips
                </Button>
              </div>

              {/* Date Range Quick Select */}
              <div className="mt-6 pt-4 border-t border-aviation-border-panel">
                <h3 className="text-sm font-semibold text-aviation-text-primary mb-3">Date Range</h3>
                <div className="space-y-1">
                  {DATE_RANGE_OPTIONS.map(opt => (
                    <Button
                      key={opt.value}
                      variant="ghost"
                      size="sm"
                      className={cn(
                        'w-full justify-start text-left text-xs',
                        revenue.period === opt.value
                          ? 'bg-aviation-bg-instrument text-aviation-text-primary'
                          : 'text-aviation-text-muted hover:text-aviation-text-primary'
                      )}
                      onClick={() => handlePeriodChange(opt.value)}
                    >
                      {opt.label}
                    </Button>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Main Panel */}
          <div className="col-span-9">
            <div className="aviation-main-panel bg-aviation-bg-panel border border-aviation-border-panel rounded-lg p-6">
              {activeView === 'economy' && (
                <CreatorEconomy
                  earnings={{
                    totalRevenue: revenue.total,
                    subscriptionRevenue: revenue.breakdown.subscription,
                    oneTimeRevenue: revenue.breakdown.oneTime,
                    royaltyRevenue: revenue.breakdown.royalty,
                    pendingPayout: 0,
                    currency: 'USD',
                  }}
                  monthlyTrend={12.5}
                  activeSubscribers={subscriptions.filter(s => s.status === 'active').length}
                  className="aviation-component"
                />
              )}

              {activeView === 'revenue' && (
                <RevenueAnalytics
                  data={revenue.dataPoints}
                  totalRevenue={revenue.total}
                  period={revenue.period}
                  onPeriodChange={handlePeriodChange}
                  className="aviation-component"
                />
              )}

              {activeView === 'subscriptions' && (
                <SubscriptionManager
                  subscriptions={subscriptions}
                  activeCount={subscriptions.filter(s => s.status === 'active').length}
                  cancelledCount={subscriptions.filter(s => s.status === 'cancelled').length}
                  pastDueCount={subscriptions.filter(s => s.status === 'past_due').length}
                  onSubscriptionSelect={(sub) => console.log('Selected subscription:', sub.id)}
                  className="aviation-component"
                />
              )}

              {activeView === 'billing' && (
                <UsageBillingPanel
                  metrics={[]}
                  totalCost={0}
                  billingCycle="monthly"
                  currency="USD"
                  className="aviation-component"
                  onUpgradeClick={undefined}
                  onMetricClick={undefined}
                />
              )}

              {activeView === 'licenses' && (
                <LicenseManager
                  licenses={[]}
                  totalActive={0}
                  totalRevoked={0}
                  className="aviation-component"
                />
              )}

              {activeView === 'profile' && selectedCreator && (
                <CreatorProfile
                  creatorId={selectedCreator.id}
                  creatorName={selectedCreator.name}
                  creatorAvatar={selectedCreator.avatar}
                  stats={selectedCreator.stats}
                  payoutInfo={selectedCreator.payoutInfo}
                  className="aviation-component"
                  onEditProfile={undefined}
                  onViewPayoutHistory={undefined}
                />
              )}

              {activeView === 'profile' && !selectedCreator && (
                <div className="aviation-empty-state flex items-center justify-center h-64">
                  <p className="text-aviation-text-muted">Select a creator to view their profile</p>
                </div>
              )}

              {activeView === 'leaderboard' && (
                <MarketplaceLeaderboard
                  entries={creators.map((c, i) => ({
                    rank: i + 1,
                    creatorId: c.id,
                    creatorName: c.name,
                    creatorAvatar: c.avatar,
                    functionId: '',
                    functionName: '',
                    sales: c.stats.totalSales,
                    revenue: c.stats.totalRevenue,
                    rating: c.stats.averageRating,
                    trend: 'up' as const,
                  }))}
                  category="creators"
                  timeRange="30d"
                  onEntrySelect={(entry) => selectCreator(entry.creatorId)}
                  className="aviation-component"
                />
              )}

              {activeView === 'royalties' && (
                <FunctionRoyaltiesPanel
                  royalties={[]}
                  totalEarned={revenue.breakdown.royalty}
                  totalPending={0}
                  currency="USD"
                  className="aviation-component"
                />
              )}

              {activeView === 'pricing' && (
                <AssetPricingEditor
                  functionId=""
                  functionName=""
                  currentPricing={[]}
                  className="aviation-component"
                  onPriceUpdate={undefined}
                  onSave={undefined}
                />
              )}

              {activeView === 'conversion' && (
                <SalesConversionAnalytics
                  funnelSteps={[]}
                  totalVisitors={0}
                  totalPurchases={0}
                  overallConversionRate={0}
                  averageOrderValue={0}
                  className="aviation-component"
                  onStageClick={undefined}
                />
              )}

              {activeView === 'optimizer' && (
                <MonetizationOptimizer
                  suggestions={[]}
                  currentRevenue={revenue.total}
                  projectedRevenue={revenue.total * 1.12}
                  className="aviation-component"
                />
              )}

              {activeView === 'trends' && (
                <MarketplaceTrendRadar
                  trends={[]}
                  timeRange="30d"
                  className="aviation-component"
                  onTrendSelect={undefined}
                  onCategoryChange={undefined}
                />
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
