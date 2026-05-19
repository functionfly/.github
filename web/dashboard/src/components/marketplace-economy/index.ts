/**
 * Marketplace Economy Components Index
 * Re-exports from @functionfly/ui-marketplace-economy and local components
 */

// Re-export all components from the marketplace-economy package
export * from '@functionfly/ui-marketplace-economy'

// Export local integration component
export { MarketplaceEconomyIntegration } from './MarketplaceEconomyIntegration'

// Export store
export { useMarketplaceEconomyStore } from '@/stores/marketplaceEconomyStore'
export {
  selectTopCreators,
  selectActiveSubscriptions,
  selectTotalRevenue,
  selectAlertCount,
} from '@/stores/marketplaceEconomyStore'
