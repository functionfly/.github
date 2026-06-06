import { useQuery } from "@tanstack/react-query";
import { marketplaceApi, type ExtensionUpdate, type InstalledPluginInfo } from "@/api/marketplace";

export function useUpdateCheck(installed: InstalledPluginInfo[]) {
  return useQuery({
    queryKey: ["marketplace-updates", installed.map(p => `${p.id}:${p.version}`).join(",")],
    queryFn: () => marketplaceApi.checkUpdates(installed),
    enabled: installed.length > 0,
    staleTime: 1000 * 60 * 15,
  });
}

export function useExtensionUpdates(installed: InstalledPluginInfo[]) {
  const { data, ...rest } = useUpdateCheck(installed);
  return { updates: data?.updates || ([] as ExtensionUpdate[]), ...rest };
}
