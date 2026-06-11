import { useStateFabricEntitlements } from '@/hooks/useBilling';
import type { StateFabricAddOnId } from './StateFabricAddonGate';

export function useHasStateFabricAddon(addonId: StateFabricAddOnId): boolean {
  const { data } = useStateFabricEntitlements();
  return (data?.addon_ids ?? []).includes(addonId);
}
