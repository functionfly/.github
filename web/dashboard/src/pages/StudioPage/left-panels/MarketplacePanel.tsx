import React from "react";
import {
  FunctionMarketplace,
  CreatorProfile,
  LicenseManager,
  MonetizationOptimizer,
  AssetPricingEditor,
} from "@functionfly/ui-marketplace";
import {
  UsageBillingPanel,
  SubscriptionManager,
  FunctionRoyaltiesPanel,
  MarketplaceLeaderboard,
  RevenueAnalytics,
} from "@functionfly/ui-marketplace-economy/components";
import { GlassCard, Badge, cn } from "@functionfly/ui-core";
import {
  useExecuteFunction,
  useFavoriteFunction,
  useUpdateLicense,
  useUpdatePricing,
} from "@/hooks/useStudioMarketplace";
import { TrendingUp, Users, DollarSign, Star } from "lucide-react";

interface MarketplacePanelProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  categoryFilter: string;
  onCategoryChange: (category: string) => void;
  currentUserName: string;
}

export function MarketplacePanel({
  searchQuery,
  onSearchChange,
  categoryFilter,
  onCategoryChange,
  currentUserName,
}: MarketplacePanelProps) {
  const executeFunction = useExecuteFunction();
  const favoriteFunction = useFavoriteFunction();
  const updateLicense = useUpdateLicense();
  const updatePricing = useUpdatePricing();

  const creatorProfile = {
    id: "creator-1",
    username: currentUserName.toLowerCase().replace(/\s+/g, "-"),
    name: currentUserName,
    avatar: undefined,
    profileUrl: `/profile/${currentUserName.toLowerCase().replace(/\s+/g, "-")}`,
  };

  const usageMetrics = [
    { name: "API Calls", used: 4521, limit: 5000, unit: "calls", cost: 0 },
    { name: "Storage", used: 1.2, limit: 5, unit: "GB", cost: 0 },
    { name: "Bandwidth", used: 2.1, limit: 10, unit: "GB", cost: 0 },
  ];

  const subscriptions = [
    {
      id: "sub-1",
      customerName: currentUserName,
      plan: "Pro",
      amount: 29,
      billingCycle: "month",
      status: "active" as const,
      currentPeriodEnd: new Date("2026-06-18").toISOString(),
    },
  ];

  const royalties = [
    {
      id: "royalty-1",
      functionId: "fn-1",
      functionName: "Data Parser",
      licensee: "Acme Corp",
      licenseType: "commercial",
      royaltyAmount: 234.5,
      royaltyPercentage: 20,
      saleAmount: 1172.5,
      saleDate: new Date("2026-05-15").toISOString(),
      paidOut: false,
    },
    {
      id: "royalty-2",
      functionId: "fn-2",
      functionName: "Text Analyzer",
      licensee: "Beta Inc",
      licenseType: "restricted",
      royaltyAmount: 156.78,
      royaltyPercentage: 15,
      saleAmount: 1045.2,
      saleDate: new Date("2026-05-10").toISOString(),
      paidOut: true,
    },
  ];

  const leaderboardEntries = [
    { rank: 1, creatorId: "c1", creatorName: "Data Parser", functionId: "fn-1", functionName: "Data Parser", creatorAvatar: "", sales: 1234, revenue: 456.78, trend: "up" as const, rating: 4.8 },
    { rank: 2, creatorId: "c2", creatorName: "Text Analyzer", functionId: "fn-2", functionName: "Text Analyzer", creatorAvatar: "", sales: 890, revenue: 234.56, trend: "up" as const, rating: 4.5 },
    { rank: 3, creatorId: "c3", creatorName: "JSON Transformer", functionId: "fn-3", functionName: "JSON Transformer", creatorAvatar: "", sales: 567, revenue: 123.45, trend: "down" as const, rating: 4.2 },
  ];

  const revenueData = [
    { month: "Jan", revenue: 120 },
    { month: "Feb", revenue: 156 },
    { month: "Mar", revenue: 189 },
    { month: "Apr", revenue: 210 },
    { month: "May", revenue: 234 },
  ];

  return (
    <div className="p-3 space-y-3">
      <FunctionMarketplace
        functions={[]}
        onSelect={(id) => executeFunction.mutate({ functionId: id })}
        onExecute={(id) => executeFunction.mutate({ functionId: id })}
        onFavorite={(id, fav) => favoriteFunction.mutate({ functionId: id, favorite: fav })}
        searchQuery={searchQuery}
        onSearchChange={onSearchChange}
        categoryFilter={categoryFilter}
        onCategoryChange={onCategoryChange}
      />

      <div className="border-t border-border-subtle pt-3 mt-3">
        <div className="text-[10px] text-text-muted uppercase tracking-wider mb-2">Creator Tools</div>

        <CreatorProfile
          creator={creatorProfile}
          stats={{
            totalFunctions: 12,
            totalDownloads: 4521,
            totalRevenue: 892.34,
            averageRating: 4.8,
          }}
          className="bg-bg-primary rounded-lg border border-border-subtle p-3"
        />

        <div className="grid grid-cols-3 gap-2 mt-3">
          <div className="bg-bg-primary rounded-lg border border-border-subtle p-2 text-center">
            <TrendingUp className="size-4 text-success mx-auto mb-1" />
            <div className="text-lg font-semibold">4.5K</div>
            <div className="text-[10px] text-text-muted">Downloads</div>
          </div>
          <div className="bg-bg-primary rounded-lg border border-border-subtle p-2 text-center">
            <Users className="size-4 text-brand-400 mx-auto mb-1" />
            <div className="text-lg font-semibold">892</div>
            <div className="text-[10px] text-text-muted">Users</div>
          </div>
          <div className="bg-bg-primary rounded-lg border border-border-subtle p-2 text-center">
            <DollarSign className="size-4 text-warning mx-auto mb-1" />
            <div className="text-lg font-semibold">$892</div>
            <div className="text-[10px] text-text-muted">Revenue</div>
          </div>
        </div>

        <UsageBillingPanel
          metrics={usageMetrics}
          totalCost={0}
          billingCycle="monthly"
          currency="USD"
          onUpgradeClick={() => console.log("Upgrade clicked")}
          className="bg-bg-primary rounded-lg border border-border-subtle p-3 mt-3"
        />

        <SubscriptionManager
          subscriptions={subscriptions}
          activeCount={1}
          cancelledCount={0}
          pastDueCount={0}
          onSubscriptionSelect={(sub) => console.log("Selected subscription:", sub)}
          className="bg-bg-primary rounded-lg border border-border-subtle p-3 mt-3"
        />

        <div className="grid grid-cols-2 gap-2 mt-3">
          <FunctionRoyaltiesPanel
            royalties={royalties}
            totalEarned={391.28}
            totalPending={234.5}
            currency="USD"
            onRoyaltySelect={(r) => console.log("Selected royalty:", r)}
            onClaimPending={() => console.log("Claim pending")}
            className="bg-bg-primary rounded-lg border border-border-subtle p-3"
          />
          <MarketplaceLeaderboard
            entries={leaderboardEntries}
            category="functions"
            timeRange="30d"
            onEntrySelect={(e) => console.log("Selected entry:", e)}
            className="bg-bg-primary rounded-lg border border-border-subtle p-3"
          />
        </div>

        <RevenueAnalytics
          data={revenueData}
          totalRevenue={909}
          revenueByCategory={{}}
          revenueByPeriod={{}}
          className="bg-bg-primary rounded-lg border border-border-subtle p-3 mt-3"
        />

        <LicenseManager
          functionId="fn-1"
          currentLicense="mit"
          onLicenseChange={(lic) => updateLicense.mutate({ functionId: "fn-1", license: lic })}
          className="bg-bg-primary rounded-lg border border-border-subtle p-3 mt-3"
        />

        <MonetizationOptimizer
          functionId="fn-1"
          currentMetrics={{
            executionCount: 4521,
            averageLatency: 42,
            errorRate: 0.02,
            userCount: 892,
          }}
          onApplyRecommendation={(model, price) =>
            updatePricing.mutate({ functionId: "fn-1", price, model: model as "per_call" })
          }
          className="bg-bg-primary rounded-lg border border-border-subtle p-3 mt-3"
        />

        <AssetPricingEditor
          functionId="fn-1"
          currentPrice={0.01}
          pricingModel="per_call"
          onPriceChange={(p) => updatePricing.mutate({ functionId: "fn-1", price: p, model: "per_call" })}
          onModelChange={(m) =>
            updatePricing.mutate({ functionId: "fn-1", price: 0.01, model: m as "per_call" })
          }
          className="bg-bg-primary rounded-lg border border-border-subtle p-3 mt-3"
        />
      </div>
    </div>
  );
}