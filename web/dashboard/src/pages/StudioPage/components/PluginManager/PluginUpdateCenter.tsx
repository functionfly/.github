import { useState, useMemo } from "react";
import { GlassCard, Badge, Button, Spinner, Input } from "@functionfly/ui-core";
import { Search, Download, Star, Shield, Clock, TrendingUp, Grid, List, Filter, ChevronLeft, ChevronRight, X } from "lucide-react";
import { type InstallPluginRequest, type PluginType } from "@/hooks/usePlugin";
import { useQuery } from "@tanstack/react-query";
import { marketplaceApi, type Extension, type MarketplaceFilters } from "@/api/marketplace";
import { PluginDetailsModal } from "./PluginDetailsModal";
import { ALL_FUNCTIONFLY_PLUGINS, FUNCTIONFLY_TEAM_PLUGINS, type DefaultPlugin } from "./defaultPlugins";
import { cn } from "@/lib/utils";

import { type Plugin } from "@/api/plugins";
interface PluginUpdateCenterProps {
  onInstall: (plugin: InstallPluginRequest) => void;
  installedPlugins?: Plugin[];
}

type SortOption = "trending" | "top_rated" | "newest" | "most_installed";
type CategoryFilter = "all" | "integration" | "infrastructure" | "ai_tool" | "runtime" | "ui";

const PAGE_SIZE = 12;

export function PluginUpdateCenter({ onInstall, installedPlugins = [] }: PluginUpdateCenterProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState<SortOption>("trending");
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>("all");
  const [featuredFilter, setFeaturedFilter] = useState<CategoryFilter | null>(null);
  const [page, setPage] = useState(1);
  const [selectedExtension, setSelectedExtension] = useState<Extension | null>(null);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

  const { data: marketplaceData, isLoading } = useQuery({
    queryKey: ["marketplace-extensions"],
    queryFn: () => marketplaceApi.list({ status: "published", limit: 100 }),
    staleTime: 1000 * 60 * 5,
  });

  const marketplaceExtensions = marketplaceData?.extensions || [];
  const defaultPlugins = ALL_FUNCTIONFLY_PLUGINS;

  const { data: installCountsData } = useQuery({
    queryKey: ["marketplace-install-counts"],
    queryFn: async () => {
      const allIds = [...defaultPlugins.map(p => p.id), ...marketplaceExtensions.map(e => e.id)];
      if (allIds.length === 0) return { install_counts: {} as Record<string, number> };
      const result = await marketplaceApi.getInstallCounts(allIds);
      return result;
    },
    staleTime: 1000 * 60 * 5,
  });

  const installedPluginNames = useMemo(() => {
    return new Set(installedPlugins.map((p) => p.name.toLowerCase()));
  }, [installedPlugins]);

  const combinedExtensions = useMemo(() => {
    const marketplaceIds = new Set(marketplaceExtensions.map((e) => e?.id).filter(Boolean));
    const defaultIds = new Set(defaultPlugins.map((e) => e?.id).filter(Boolean));

    const pluginsById = new Map<string, Extension>();
    const installCounts = installCountsData?.install_counts || {};

    for (const plugin of defaultPlugins) {
      if (!plugin?.id || !plugin?.name) continue;
      const realCount = installCounts[plugin.id];
      const updatedPlugin = realCount !== undefined
        ? { ...plugin, install_count: realCount }
        : plugin;
      pluginsById.set(plugin.id, updatedPlugin);
    }

    for (const plugin of marketplaceExtensions) {
      if (!plugin?.id || !plugin?.name) continue;
      if (pluginsById.has(plugin.id)) {
        const existing = pluginsById.get(plugin.id);
        if (plugin.install_count > (existing?.install_count || 0)) {
          pluginsById.set(plugin.id, plugin);
        }
      } else {
        pluginsById.set(plugin.id, plugin);
      }
    }

    return Array.from(pluginsById.values()).filter(ext => ext?.name);
  }, [defaultPlugins, marketplaceExtensions, installCountsData]);

  const filteredExtensions = useMemo(() => {
    let results = combinedExtensions;

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      results = results.filter(
        (p) =>
          p.name.toLowerCase().includes(query) ||
          p.description.toLowerCase().includes(query) ||
          p.tags?.some((t) => t.toLowerCase().includes(query))
      );
    }

    if (categoryFilter !== "all") {
      results = results.filter((p) => p.category === categoryFilter);
    }

    switch (sortBy) {
      case "top_rated":
        results = [...results].sort((a, b) => b.rating_average - a.rating_average);
        break;
      case "newest":
        results = [...results].sort((a, b) => {
          const dateA = new Date(a.changelog ? 0 : 0).getTime();
          const dateB = new Date(b.changelog ? 0 : 0).getTime();
          return dateB - dateA;
        });
        break;
      case "most_installed":
        results = [...results].sort((a, b) => b.install_count - a.install_count);
        break;
      case "trending":
      default:
        results = [...results].sort((a, b) => {
          const scoreA = a.install_count * 0.3 + a.rating_average * 100 * 0.7;
          const scoreB = b.install_count * 0.3 + b.rating_average * 100 * 0.7;
          return scoreB - scoreA;
        });
        break;
    }

    return results;
  }, [combinedExtensions, searchQuery, categoryFilter, sortBy]);

  const totalPages = Math.ceil(filteredExtensions.length / PAGE_SIZE);
  const paginatedExtensions = filteredExtensions.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  const formatNumber = (n: number) => {
    if (n >= 1000) return (n / 1000).toFixed(1) + "k";
    return n.toString();
  };

  const handleInstall = async (extension: Extension) => {
    await onInstall({
      manifest: extension.manifest || { name: extension.name, version: extension.version },
      plugin_type: (extension.category as PluginType) || "ui",
      name: extension.name,
      version: extension.version,
      description: extension.description,
      author_name: extension.creator_id,
      category: extension.category,
      size_bytes: 0,
    });
  };

  const categoryBadges: { value: CategoryFilter; label: string; icon: React.ReactNode }[] = [
    { value: "all", label: "All", icon: <Grid className="w-3 h-3" /> },
    { value: "integration", label: "Integrations", icon: <TrendingUp className="w-3 h-3" /> },
    { value: "infrastructure", label: "Infrastructure", icon: <Shield className="w-3 h-3" /> },
    { value: "ai_tool", label: "AI Tools", icon: <Star className="w-3 h-3" /> },
    { value: "runtime", label: "Runtime", icon: <Clock className="w-3 h-3" /> },
    { value: "ui", label: "UI", icon: <List className="w-3 h-3" /> },
  ];

  const sortOptions: { value: SortOption; label: string }[] = [
    { value: "trending", label: "Trending" },
    { value: "top_rated", label: "Top Rated" },
    { value: "newest", label: "Newest" },
    { value: "most_installed", label: "Most Installed" },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4 flex-wrap">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
          <Input
            placeholder="Search plugins..."
            value={searchQuery}
            onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
            className="pl-9 bg-white/5 border-white/10"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery("")}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-white/40 hover:text-white/60"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        <select
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value as SortOption)}
          className="bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm"
        >
          {sortOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>

        <div className="flex items-center gap-0 p-0.5 rounded-lg bg-white/5 border border-white/10">
          <button
            onClick={() => setViewMode("grid")}
            className={cn("p-2 rounded-md transition-all", viewMode === "grid" ? "bg-white/10 text-white" : "text-white/40 hover:text-white/70")}
          >
            <Grid className="w-4 h-4" />
          </button>
          <button
            onClick={() => setViewMode("list")}
            className={cn("p-2 rounded-md transition-all", viewMode === "list" ? "bg-white/10 text-white" : "text-white/40 hover:text-white/70")}
          >
            <List className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        {categoryBadges.map((badge) => (
          <button
            key={badge.value}
            onClick={() => { setCategoryFilter(badge.value); setPage(1); }}
            className={cn(
              "flex items-center gap-2 px-3 py-1.5 rounded-full border text-sm font-medium transition-all",
              categoryFilter === badge.value
                ? "bg-white/10 text-white border-white/20"
                : "bg-white/5 text-white/60 border-white/10 hover:text-white/80 hover:bg-white/10"
            )}
          >
            {badge.icon}
            {badge.label}
          </button>
        ))}
      </div>

      <div className="flex items-center justify-between">
        <p className="text-sm text-white/60">
          {filteredExtensions.length} plugins found
          {searchQuery && ` for "${searchQuery}"`}
        </p>
        {page > 1 && (
          <Badge variant="outline" className="text-white/60 border-white/20">
            Page {page} of {totalPages}
          </Badge>
        )}
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Spinner className="w-8 h-8" />
        </div>
      ) : paginatedExtensions.length === 0 ? (
        <GlassCard className="flex flex-col items-center justify-center h-64">
          <Search className="w-12 h-12 text-white/20 mb-3" />
          <p className="text-white/60 mb-2">No plugins found</p>
          <p className="text-sm text-white/40">Try adjusting your search or filters</p>
        </GlassCard>
      ) : viewMode === "grid" ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {paginatedExtensions.map((extension) => (
            <GlassCard
              key={extension.id}
              className="p-4 hover:bg-white/10 transition-all cursor-pointer group"
              onClick={() => setSelectedExtension(extension)}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-purple-500/30 to-blue-500/30 flex items-center justify-center text-xl font-bold text-white/80">
                  {extension && extension.name ? extension.name.charAt(0) : '?'}
                </div>
                {extension.verified && (
                  <Shield className="w-5 h-5 text-green-400" />
                )}
              </div>

              <h4 className="font-medium text-white mb-1">{extension.name}</h4>
              <p className="text-xs text-white/60 mb-2">by {extension.creator_id}</p>

              <p className="text-sm text-white/60 line-clamp-2 mb-3 min-h-[2.5rem]">
                {extension.description}
              </p>

              <div className="flex flex-wrap gap-1 mb-3">
                {extension.tags?.slice(0, 3).map((tag) => (
                  <Badge key={tag} className="text-xs bg-white/10 text-white/70 border-white/20">
                    {tag}
                  </Badge>
                ))}
              </div>

              <div className="flex items-center gap-3 text-xs text-white/60 mb-4">
                <span className="flex items-center gap-1">
                  <Download className="w-3 h-3" />
                  {formatNumber(extension.install_count)}
                </span>
                <span className="flex items-center gap-1 text-yellow-400">
                  <Star className="w-3 h-3 fill-current" />
                  {extension.rating_average.toFixed(1)}
                </span>
              </div>

              <Button
                size="sm"
                className="w-full opacity-0 group-hover:opacity-100 transition-opacity"
                variant="default"
                disabled={installedPluginNames.has(extension.name.toLowerCase())}
                onClick={(e) => { e.stopPropagation(); handleInstall(extension); }}
              >
                <Download className="w-4 h-4 mr-1" />
                {installedPluginNames.has(extension.name.toLowerCase()) ? "Installed" : "Install"}
              </Button>
            </GlassCard>
          ))}
        </div>
      ) : (
        <div className="space-y-3">
          {paginatedExtensions.map((extension) => (
            <GlassCard
              key={extension.id}
              className="p-4 hover:bg-white/10 transition-all cursor-pointer"
              onClick={() => setSelectedExtension(extension)}
            >
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-purple-500/30 to-blue-500/30 flex items-center justify-center text-xl font-bold text-white/80 shrink-0">
                  {extension && extension.name ? extension.name.charAt(0) : '?'}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h4 className="font-medium text-white">{extension.name}</h4>
                    {extension.verified && <Shield className="w-4 h-4 text-green-400" />}
                    <Badge className="text-xs bg-white/10 text-white/80 border-white/20 shrink-0">
                      v{extension.version}
                    </Badge>
                  </div>
                  <p className="text-sm text-white/60 mt-1">{extension.description}</p>
                  <div className="flex items-center gap-4 mt-2">
                    <span className="text-xs text-white/40">by {extension.creator_id}</span>
                    <span className="text-xs text-white/40 flex items-center gap-1">
                      <Download className="w-3 h-3" />
                      {formatNumber(extension.install_count)}
                    </span>
                    <span className="text-xs text-yellow-400 flex items-center gap-1">
                      <Star className="w-3 h-3 fill-current" />
                      {extension.rating_average.toFixed(1)}
                    </span>
                  </div>
                </div>
                <Button
                  size="sm"
                  variant="default"
                  disabled={installedPluginNames.has(extension.name.toLowerCase())}
                  onClick={(e) => { e.stopPropagation(); handleInstall(extension); }}
                >
                  <Download className="w-4 h-4 mr-1" />
                  {installedPluginNames.has(extension.name.toLowerCase()) ? "Installed" : "Install"}
                </Button>
              </div>
            </GlassCard>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="w-4 h-4" />
          </Button>
          {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
            const pageNum = i + 1;
            return (
              <Button
                key={pageNum}
                variant={page === pageNum ? "default" : "outline"}
                size="sm"
                onClick={() => setPage(pageNum)}
              >
                {pageNum}
              </Button>
            );
          })}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page === totalPages}
          >
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>
      )}

      {selectedExtension && (
        <PluginDetailsModal
          extension={selectedExtension}
          onClose={() => setSelectedExtension(null)}
          onInstall={onInstall}
        />
      )}
    </div>
  );
}