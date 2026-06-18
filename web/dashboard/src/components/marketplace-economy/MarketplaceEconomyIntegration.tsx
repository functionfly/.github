/**
 * Marketplace Economy Integration Component
 * Wires together all marketplace-economy sub-components with the store
 */

import React from 'react'
import {
  CreatorEconomy,
  RevenueAnalytics,
  SubscriptionManager,
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

interface MarketplaceEconomyIntegrationProps {
  className?: string
}

export function MarketplaceEconomyIntegration({ className }: MarketplaceEconomyIntegrationProps) {
  const {
    creators,
    subscriptions,
    revenue,
    selectedCreatorId,
    selectCreator,
    activeView,
  } = useMarketplaceEconomyStore()

  const selectedCreator = creators.find(c => c.id === selectedCreatorId)

  return (
    <div className={className}>
      {activeView === 'economy' && (
        <CreatorEconomy
          earnings={{
            totalRevenue: revenue.total,
            pendingPayout: 0,
            currency: 'USD',
            subscriptionRevenue: revenue.breakdown.subscription,
            oneTimeRevenue: revenue.breakdown.oneTime,
            royaltyRevenue: revenue.breakdown.royalty,
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
          period={revenue.period as '7d' | '30d' | '90d' | '1y'}
          className="aviation-component"
        />
      )}

      {activeView === 'subscriptions' && (
        <SubscriptionManager
          subscriptions={subscriptions}
          activeCount={subscriptions.filter(s => s.status === 'active').length}
          cancelledCount={subscriptions.filter(s => s.status === 'cancelled').length}
          pastDueCount={subscriptions.filter(s => s.status === 'past_due').length}
          className="aviation-component"
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

      {activeView === 'leaderboard' && (
        <MarketplaceLeaderboard
          entries={creators.map((c, i) => ({
            rank: i + 1,
            creatorId: c.id,
            creatorName: c.name,
            functionId: '',
            functionName: '',
            sales: c.stats?.totalSales || 0,
            revenue: c.stats?.totalRevenue || 0,
            rating: c.stats?.averageRating || 0,
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
  )
}