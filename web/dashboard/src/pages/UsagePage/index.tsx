/**
 * Usage Page - Modular Version
 * Refactored into smaller, manageable components
 */

import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  CreditCard,
  Loader2,
  TrendingUp,
  ExternalLink,
  BarChart3,
  Cloud,
  DollarSign,
  Calendar,
  Zap,
} from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { createBillingPortalSession, getBillingPortalErrorMessage } from '@/api/billing';
import { toast } from 'sonner';
import { ROUTES } from '@/lib/constants';
import { usePlan } from '@/hooks/usePlan';
import { useAuthStore } from '@/stores/authStore';

// Local imports
import { DATE_RANGES, type DateRangeValue, getDateRangeDates } from './constants';
import { useUsagePageData } from './hooks/useUsageData';
import { useChartData } from './hooks/useChartData';
import { useInsights } from './hooks/useInsights';
import { DateRangePicker, type DateRangeSelection } from '@/components/ui/date-picker';

// Components
import { InsightsSection } from './components/InsightsSection';
import { OverviewTab } from './components/OverviewTab';
import { ResourcesTab } from './components/ResourcesTab';
import { CostsTab } from './components/CostsTab';
import { ForecastTab } from './components/ForecastTab';

export function UsagePage() {
  const user = useAuthStore((s) => s.user);
  const { limits, displayName, isEnterprise } = usePlan();
  const [billingLoading, setBillingLoading] = useState(false);
  const [dateRange, setDateRange] = useState<DateRangeValue>('30d');
  const [customDateRange, setCustomDateRange] = useState<DateRangeSelection>({ from: null, to: null });
  const [activeTab, setActiveTab] = useState('overview');

  // Calculate actual date range based on selection
  const { from: periodStart, to: periodEnd } = getDateRangeDates(dateRange, customDateRange);

  // Fetch all data via custom hook
  const {
    meData,
    displayPlan,
    username,

    // Data
    usageData,
    executionRateDataRes,
    functionsData,
    providersData,
    appsData,
    stateFabricsList,
    agentsListRes,
    agentsUsageAndBalance,
    costSummary,
    functionCosts,
    periodData,
    regionData,
    forecast,
    spendCap,

    // Counts
    functionsCount,
    providersCount,
    appsCount,
    fabricIds,
    agentIds,

    // Limits
    functionsLimit,
    providersLimit,
    stateFabricsLimit,
    agentsLimit,
    stateFabricTotals,

    // Loading states
    isLoading,
    usageLoading,
    executionRateLoading,
    costSummaryLoading,
    functionCostsLoading,
    periodLoading,
    regionLoading,
    forecastLoading,
    spendCapLoading,
  } = useUsagePageData(dateRange);

  // Chart data transformations
  const {
    usageGraphData,
    executionRateData,
    totalUsage,
    periodComparison,
    dailyChartData,
    costBreakdownData,
    regionChartData,
    functionChartData,
  } = useChartData({
    usageData,
    executionRateDataRes,
    periodData,
    costSummary,
    regionData,
    functionCosts,
  });

  // Generate insights
  const insights = useInsights({
    costSummary,
    forecast,
    spendCap,
    functionCosts,
  });

  // Plan limit calculations
  const requestLimit = limits?.requests ?? 0;
  const isUnlimited = requestLimit === Infinity || isEnterprise;
  const remaining = isUnlimited ? null : Math.max(0, requestLimit - totalUsage);
  const usagePercent = isUnlimited
    ? 0
    : requestLimit > 0
      ? Math.min(100, (totalUsage / requestLimit) * 100)
      : 0;
  const isOverLimit = !isUnlimited && totalUsage > requestLimit;

  // Billing portal handler
  const openBillingPortal = async () => {
    setBillingLoading(true);
    try {
      const { url } = await createBillingPortalSession(
        `${window.location.origin}${ROUTES.USAGE}`
      );
      if (url) window.location.href = url;
    } catch (e) {
      toast.error(getBillingPortalErrorMessage(e));
    } finally {
      setBillingLoading(false);
    }
  };

  // Custom domains limit (may not exist on all limit types)
  const customDomainsLimit = (limits as { customDomains?: number })?.customDomains ?? 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-primary tracking-tight">
            Usage & Analytics
          </h1>
          <p className="text-text-secondary mt-1">
            Track platform usage, costs, and resource limits across all services.
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <Select
            value={dateRange}
            onValueChange={(v) => setDateRange(v as DateRangeValue)}
          >
            <SelectTrigger className="w-[160px]">
              <Calendar className="h-4 w-4 mr-2" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {DATE_RANGES.map((range) => (
                <SelectItem key={range.value} value={range.value}>
                  {range.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {dateRange === 'custom' && (
            <DateRangePicker
              value={customDateRange}
              onChange={setCustomDateRange}
              showPresets
            />
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={openBillingPortal}
            disabled={billingLoading}
          >
            {billingLoading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <CreditCard className="h-4 w-4 mr-2" />
            )}
            Manage billing
          </Button>
        </div>
      </div>

      {/* Insights */}
      <InsightsSection insights={insights} />

      {/* Quick Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">
              Total Executions
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : (
              <p className="text-2xl font-semibold text-text-primary">
                {totalUsage.toLocaleString()}
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              Period Trend
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : (
              <p
                className={`text-2xl font-semibold ${
                  periodComparison.change > 0
                    ? 'text-amber-500'
                    : periodComparison.change < 0
                      ? 'text-emerald-500'
                      : 'text-text-primary'
                }`}
              >
                {periodComparison.change > 0 ? '+' : ''}
                {periodComparison.change.toFixed(1)}%
              </p>
            )}
          </CardContent>
        </Card>

        <Card className="border-theme bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
              <Cloud className="h-4 w-4" />
              Apps
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
            ) : (
              <p className="text-2xl font-semibold text-text-primary">{appsCount}</p>
            )}
          </CardContent>
        </Card>

        {!isUnlimited && (
          <Card className="border-theme bg-card">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
                <TrendingUp className="h-4 w-4" />
                Remaining Requests
              </CardTitle>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
              ) : (
                <p
                  className={`text-xl font-semibold ${
                    isOverLimit ? 'text-destructive' : 'text-text-primary'
                  }`}
                >
                  {remaining !== null ? remaining.toLocaleString() : '—'}
                </p>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      {/* Tabs - Modern Segmented Control Style */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <div className="flex justify-center">
          <TabsList className="h-11 p-1.5 rounded-xl bg-bg-secondary/80 border border-border-subtle backdrop-blur-sm">
            <TabsTrigger
              value="overview"
              className="relative px-5 py-2 text-sm font-medium rounded-lg data-[state=active]:bg-card data-[state=active]:text-text-primary data-[state=active]:shadow-sm data-[state=active]:border-border-subtle transition-all duration-200 flex items-center gap-2"
            >
              <BarChart3 className="w-4 h-4" />
              Overview
            </TabsTrigger>
            <TabsTrigger
              value="resources"
              className="relative px-5 py-2 text-sm font-medium rounded-lg data-[state=active]:bg-card data-[state=active]:text-text-primary data-[state=active]:shadow-sm data-[state=active]:border-border-subtle transition-all duration-200 flex items-center gap-2"
            >
              <Zap className="w-4 h-4" />
              Resources
            </TabsTrigger>
            <TabsTrigger
              value="costs"
              className="relative px-5 py-2 text-sm font-medium rounded-lg data-[state=active]:bg-card data-[state=active]:text-text-primary data-[state=active]:shadow-sm data-[state=active]:border-border-subtle transition-all duration-200 flex items-center gap-2"
            >
              <DollarSign className="w-4 h-4" />
              Cost Details
            </TabsTrigger>
            <TabsTrigger
              value="forecast"
              className="relative px-5 py-2 text-sm font-medium rounded-lg data-[state=active]:bg-card data-[state=active]:text-text-primary data-[state=active]:shadow-sm data-[state=active]:border-border-subtle transition-all duration-200 flex items-center gap-2"
            >
              <TrendingUp className="w-4 h-4" />
              Forecast
            </TabsTrigger>
          </TabsList>
        </div>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4 mt-4">
          <OverviewTab
            usageLoading={usageLoading}
            executionRateLoading={executionRateLoading}
            periodLoading={periodLoading}
            totalUsage={totalUsage}
            usageGraphData={usageGraphData}
            executionRateData={executionRateData}
            dailyChartData={dailyChartData}
            limits={{
              isUnlimited,
              requestLimit,
              usagePercent,
              isOverLimit,
              remaining,
            }}
          />
        </TabsContent>

        {/* Resources Tab */}
        <TabsContent value="resources" className="space-y-4 mt-4">
          <ResourcesTab
            displayPlan={displayPlan}
            displayName={displayName}
            requestLimit={requestLimit}
            functionsLimit={functionsLimit}
            providersLimit={providersLimit}
            stateFabricsLimit={stateFabricsLimit}
            agentsLimit={agentsLimit}
            customDomainsLimit={customDomainsLimit}
            functionsCount={functionsCount}
            providersCount={providersCount}
            stateFabricsCount={(stateFabricsList ?? []).length}
            agentIds={agentIds}
            agentsUsageAndBalance={agentsUsageAndBalance}
            stateFabricTotals={stateFabricTotals}
          />
        </TabsContent>

        {/* Costs Tab */}
        <TabsContent value="costs" className="space-y-4 mt-4">
          <CostsTab
            costSummaryLoading={costSummaryLoading}
            functionCostsLoading={functionCostsLoading}
            periodLoading={periodLoading}
            regionLoading={regionLoading}
            costSummary={costSummary}
            costBreakdownData={costBreakdownData}
            regionChartData={regionChartData}
            functionChartData={functionChartData}
            dailyChartData={dailyChartData}
          />
        </TabsContent>

        {/* Forecast Tab */}
        <TabsContent value="forecast" className="space-y-4 mt-4">
          <ForecastTab
            forecastLoading={forecastLoading}
            spendCapLoading={spendCapLoading}
            billingLoading={billingLoading}
            forecast={forecast}
            spendCap={spendCap}
            username={username}
            openBillingPortal={openBillingPortal}
          />
        </TabsContent>
      </Tabs>

      {/* Footer */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base">Billing & Documentation</CardTitle>
          <CardDescription>
            Manage your subscription, payment methods, and view invoices.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={openBillingPortal}
            disabled={billingLoading}
          >
            {billingLoading ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <CreditCard className="h-4 w-4 mr-2" />
            )}
            Open billing portal
          </Button>
          <Button variant="ghost" size="sm" asChild>
            <Link
              to={username ? `/u/${username}/settings/billing` : ROUTES.SETTINGS}
            >
              Settings <ExternalLink className="h-4 w-4 ml-2" />
            </Link>
          </Button>
          <Button variant="ghost" size="sm" asChild>
            <Link to={ROUTES.PRICING}>
              View plans <ExternalLink className="h-4 w-4 ml-2" />
            </Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
