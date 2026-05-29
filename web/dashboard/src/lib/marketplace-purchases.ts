/** When false, hides My Purchases nav and shows rollout placeholder. */
export function isMarketplacePurchasesEnabled(): boolean {
  const raw = (import.meta.env.VITE_MARKETPLACE_PURCHASES_ENABLED ?? 'true')
    .toString()
    .trim()
    .toLowerCase();
  return !['0', 'false', 'no', 'off'].includes(raw);
}
