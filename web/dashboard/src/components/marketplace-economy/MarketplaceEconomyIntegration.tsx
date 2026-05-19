/**
 * Marketplace Economy Integration Component
 * Wires together all marketplace-economy sub-components with the store
 */

import React from 'react'
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
} from '@functionfly/ui-marketplace-economy'
import { useMarketplaceEconomyStore } from '@/stores/marketplaceEconomyStore'
import type {
  Subscription,
  License,
  RoyaltyRecord,
  LeaderboardEntry,
  PricingTier,
  ConversionFunnelStep,
  OptimizationSuggestion,
  TrendItem,
} from '@functionfly/ui-marketplace-economy'

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

  const handleSubscriptionSelect = (subscription: Subscription) => {
    console.log('Subscription selected:', subscription)
  }

  const handleSubscriptionCancel = (subscriptionId: string) => {
    console.log('Cancel subscription:', subscriptionId)
  }

  const handleLicenseSelect = (license: License) => {
    console.log('License selected:', license)
  }

  const handleLicenseRevoke = (licenseId: string) => {
    console.log('Revoke license:', licenseId)
  }

  const handleRoyaltySelect = (royalty: RoyaltyRecord) => {
    console.log('Royalty selected:', royalty)
  }

  const handleLeaderboardSelect = (entry: LeaderboardEntry) => {
    selectCreator(entry.creatorId)
  }

  const handlePricingUpdate = (tiers: PricingTier[]) => {
    console.log('Pricing updated:', tiers)
  }

  const handleConversionStageClick = (step: ConversionFunnelStep) => {
    console.log('Conversion stage clicked:', step)
  }

  const handleSuggestionApply = (suggestion: OptimizationSuggestion) => {
    console.log('Apply suggestion:', suggestion)
  }

  const handleSuggestionDismiss = (suggestionId: string) => {
    console.log('Dismiss suggestion:', suggestionId)
  }

  const handleTrendSelect = (trend: TrendItem) => {
    console.log('Trend selected:', trend)
  }

  return (
    <div className={className}>
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
          className="aviation-component"
        />
      )}

      {activeView === 'subscriptions' && (
        <SubscriptionManager
          subscriptions={subscriptions}
          activeCount={subscriptions.filter(s => s.status === 'active').length}
          cancelledCount={subscriptions.filter(s => s.status === 'cancelled').length}
          pastDueCount={subscriptions.filter(s => s.status === 'past_due').length}
          onSubscriptionSelect={handleSubscriptionSelect}
          onSubscriptionCancel={handleSubscriptionCancel}
          className="aviation-component"
        />
      )}

      {activeView === 'licenses' && (
        <LicenseManager
          licenses={[]}
          totalActive={0}
          totalRevoked={0}
          onLicenseSelect={handleLicenseSelect}
          onLicenseRevoke={handleLicenseRevoke}
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
        />
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
            sales: c.stats?.totalSales || 0,
            revenue: c.stats?.totalRevenue || 0,
            rating: c.stats?.averageRating || 0,
            trend: 'up' as const,
          }))}
          category="creators"
          timeRange="30d"
          onEntrySelect={handleLeaderboardSelect}
          className="aviation-component"
        />
      )}

      {activeView === 'royalties' && (
        <FunctionRoyaltiesPanel
          royalties={[]}
          totalEarned={revenue.breakdown.royalty}
          totalPending={0}
          currency="USD"
          onRoyaltySelect={handleRoyaltySelect}
          className="aviation-component"
        />
      )}

      {activeView === 'pricing' && (
        <AssetPricingEditor
          functionId=""
          functionName=""
          currentPricing={[]}
          onPriceUpdate={handlePricingUpdate}
          className="aviation-component"
        />
      )}

      {activeView === 'conversion' && (
        <SalesConversionAnalytics
          funnelSteps={[]}
          totalVisitors={0}
          totalPurchases={0}
          overallConversionRate={0}
          averageOrderValue={0}
          onStageClick={handleConversionStageClick}
          className="aviation-component"
        />
      )}

      {activeView === 'optimizer' && (
        <MonetizationOptimizer
          suggestions={[]}
          currentRevenue={revenue.total}
          onSuggestionApply={handleSuggestionApply}
          onSuggestionDismiss={handleSuggestionDismiss}
          className="aviation-component"
        />
      )}

      {activeView === 'trends' && (
        <MarketplaceTrendRadar
          trends={[]}
          timeRange="30d"
          onTrendSelect={handleTrendSelect}
          className="aviation-component"
        />
      )}
    </div>
  )
}
