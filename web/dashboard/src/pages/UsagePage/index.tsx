/**
 * Usage Page — Sealed Containment
 */

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { usePageTitle } from '@/hooks';
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
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  StatusPill,
  GaugeStrip,
  Gauge,
  AnnotationTag,
} from '@/components/containment';
import { createBillingPortalSession, getBillingPortalErrorMessage } from '@/api/billing';
import { toast } from 'sonner';
import { ROUTES } from '@/lib/constants';
import { usePlan } from '@/hooks/usePlan';
import { useAuthStore } from '@/stores/authStore';
import { cn } from '@/lib/utils';

import { DATE_RANGES, type DateRangeValue, getDateRangeDates } from './constants';
import { useUsagePageData } from './hooks/useUsageData';
import { useChartData } from './hooks/useChartData';
import { useInsights } from './hooks/useInsights';
import { DateRangePicker, type DateRangeSelection } from '@/components/ui/date-picker';
import './styles.css';

import { InsightsSection } from './components/InsightsSection';
import { OverviewTab } from './components/OverviewTab';
import { ResourcesTab } from './components/ResourcesTab';
import { CostsTab } from './components/CostsTab';
import { ForecastTab } from './components/ForecastTab';

function DateRangeSelect({ value, onChange }: { value: DateRangeValue; onChange: (v: DateRangeValue) => void }) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value as DateRangeValue)}
      className="usage-select"
    >
      {DATE_RANGES.map((range) => (
        <option key={range.value} value={range.value}>{range.label}</option>
      ))}
    </select>
  );
}

function TabButton({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: React.ReactNode; label: string }) {
  return (
    <button onClick={onClick} className={cn('usage-tab', active && 'usage-tab--active')}>
      {icon}
      {label}
    </button>
  );
}

export function UsagePage() {
  usePageTitle('Usage');
  const user = useAuthStore((s) => s.user);
  const { limits, displayName, isEnterprise } = usePlan();
  const [billingLoading, setBillingLoading] = useState(false);
  const [dateRange, setDateRange] = useState<DateRangeValue>('30d');
  const [customDateRange, setCustomDateRange] = useState<DateRangeSelection>({ from: null, to: null });
  const [activeTab, setActiveTab] = useState('overview');

  const { from: periodStart, to: periodEnd } = getDateRangeDates(dateRange, customDateRange);

  const {
    meData, displayPlan, username,
    usageData, executionRateDataRes, functionsData, providersData, appsData,
    stateFabricsList, agentsListRes, agentsUsageAndBalance,
    costSummary, functionCosts, periodData, regionData, forecast, spendCap,
    functionsCount, providersCount, appsCount, fabricIds, agentIds,
    functionsLimit, providersLimit, stateFabricsLimit, agentsLimit, stateFabricTotals,
    isLoading, usageLoading, executionRateLoading, costSummaryLoading,
    functionCostsLoading, periodLoading, regionLoading, forecastLoading, spendCapLoading,
  } = useUsagePageData(dateRange);

  const {
    usageGraphData, executionRateData, totalUsage, periodComparison,
    dailyChartData, costBreakdownData, regionChartData, functionChartData,
  } = useChartData({ usageData, executionRateDataRes, periodData, costSummary, regionData, functionCosts });

  const insights = useInsights({ costSummary, forecast, spendCap, functionCosts, periodComparison });

  const requestLimit = limits?.requests ?? 0;
  const isUnlimited = requestLimit === Infinity || isEnterprise;
  const remaining = isUnlimited ? null : Math.max(0, requestLimit - totalUsage);
  const usagePercent = isUnlimited ? 0 : requestLimit > 0 ? Math.min(100, (totalUsage / requestLimit) * 100) : 0;
  const isOverLimit = !isUnlimited && totalUsage > requestLimit;

  const openBillingPortal = async () => {
    setBillingLoading(true);
    try {
      const { url } = await createBillingPortalSession(`${window.location.origin}${ROUTES.USAGE}`);
      if (url) window.location.href = url;
    } catch (e) {
      toast.error(getBillingPortalErrorMessage(e));
    } finally {
      setBillingLoading(false);
    }
  };

  const customDomainsLimit = (limits as { customDomains?: number })?.customDomains ?? 0;

  return (
    <div className="usage-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="usage-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE US-01" secondary="Usage & Analytics" position="top-right" />

        <div className="usage-hero__header">
          <div className="usage-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="usage-hero__title">Usage & Analytics</h1>
          </div>
          <p className="usage-hero__subtitle">
            Track platform usage, costs, and resource limits across all services.
          </p>
          <div className="usage-hero__controls">
            <div className="usage-hero__control-group">
              <Calendar className="usage-hero__control-icon" />
              <DateRangeSelect value={dateRange} onChange={setDateRange} />
            </div>
            {dateRange === 'custom' && (
              <DateRangePicker value={customDateRange} onChange={setCustomDateRange} showPresets />
            )}
            <FrameButton
              size="sm"
              onClick={openBillingPortal}
              disabled={billingLoading}
              iconLeft={billingLoading ? <Loader2 className="usage-spinner" /> : <CreditCard className="usage-icon-sm" />}
            >
              Manage billing
            </FrameButton>
          </div>
        </div>

        <GaugeStrip>
          <Gauge isFirst data={{ value: isLoading ? '...' : totalUsage.toLocaleString(), label: 'Total Executions' }} />
          <Gauge data={{ value: isLoading ? '...' : `${periodComparison.change > 0 ? '+' : ''}${periodComparison.change.toFixed(1)}%`, label: 'Period Trend' }} />
          <Gauge data={{ value: isLoading ? '...' : appsCount, label: 'Apps' }} />
          {!isUnlimited && (
            <Gauge data={{ value: isLoading ? '...' : (remaining !== null ? remaining.toLocaleString() : '—'), label: 'Remaining' }} />
          )}
        </GaugeStrip>
      </Chamber>

      {/* Insights */}
      <InsightsSection insights={insights} />

      {/* Tabs */}
      <div className="usage-tabs">
        <TabButton active={activeTab === 'overview'} onClick={() => setActiveTab('overview')} icon={<BarChart3 className="usage-icon-sm" />} label="Overview" />
        <TabButton active={activeTab === 'resources'} onClick={() => setActiveTab('resources')} icon={<Zap className="usage-icon-sm" />} label="Resources" />
        <TabButton active={activeTab === 'costs'} onClick={() => setActiveTab('costs')} icon={<DollarSign className="usage-icon-sm" />} label="Cost Details" />
        <TabButton active={activeTab === 'forecast'} onClick={() => setActiveTab('forecast')} icon={<TrendingUp className="usage-icon-sm" />} label="Forecast" />
      </div>

      {/* Tab Content */}
      <div className="usage-tab-content">
        {activeTab === 'overview' && (
          <OverviewTab
            usageLoading={usageLoading}
            executionRateLoading={executionRateLoading}
            periodLoading={periodLoading}
            totalUsage={totalUsage}
            usageGraphData={usageGraphData}
            executionRateData={executionRateData}
            dailyChartData={dailyChartData}
            limits={{ isUnlimited, requestLimit, usagePercent, isOverLimit, remaining }}
          />
        )}
        {activeTab === 'resources' && (
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
        )}
        {activeTab === 'costs' && (
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
        )}
        {activeTab === 'forecast' && (
          <ForecastTab
            forecastLoading={forecastLoading}
            spendCapLoading={spendCapLoading}
            billingLoading={billingLoading}
            forecast={forecast}
            spendCap={spendCap}
            username={username}
            openBillingPortal={openBillingPortal}
          />
        )}
      </div>

      {/* Footer */}
      <Chamber className="usage-footer-chamber">
        <CornerBrace position="tr" />
        <CornerBrace position="bl" />
        <div className="usage-footer__header">
          <h2 className="usage-footer__title">Billing & Documentation</h2>
          <p className="usage-footer__desc">Manage your subscription, payment methods, and view invoices.</p>
        </div>
        <div className="usage-footer__actions">
          <SealedButton
            onClick={openBillingPortal}
            disabled={billingLoading}
            loading={billingLoading}
            iconLeft={<CreditCard className="usage-icon-sm" />}
          >
            Open billing portal
          </SealedButton>
          <Link to={username ? `/u/${username}/settings/billing` : ROUTES.SETTINGS}>
            <FrameButton size="sm" iconRight={<ExternalLink className="usage-icon-sm" />}>
              Settings
            </FrameButton>
          </Link>
          <Link to={ROUTES.PRICING}>
            <FrameButton size="sm" iconRight={<ExternalLink className="usage-icon-sm" />}>
              View plans
            </FrameButton>
          </Link>
        </div>
      </Chamber>
    </div>
  );
}
