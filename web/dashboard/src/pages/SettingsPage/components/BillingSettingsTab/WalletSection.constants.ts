export const WALLET_TOP_UP_PRESETS = [10, 25, 50, 100] as const;
export const MIN_WALLET_TOP_UP_USD = 1;
export const MAX_WALLET_TOP_UP_USD = 10_000;

export function formatUsd(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}
