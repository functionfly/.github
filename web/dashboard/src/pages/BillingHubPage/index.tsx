import { useBillingHub } from './useBillingHub';
import { OverviewTab } from './components/OverviewTab';
import { PlansTab } from './components/PlansTab';
import { AddOnsTab } from './components/AddOnsTab';
import { InvoicesTab } from './components/InvoicesTab';
import { PaymentMethodsTab } from './components/PaymentMethodsTab';
import { WalletTab } from './components/WalletTab';
import { UsageTab } from './components/UsageTab';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  CreditCard,
  FileText,
  LayoutDashboard,
  Wallet,
  Zap,
  BarChart3,
  Package,
} from 'lucide-react';

type BillingTab = 'overview' | 'plans' | 'addons' | 'invoices' | 'payment' | 'wallet' | 'usage';

const TABS: { id: BillingTab; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'overview', label: 'Overview', icon: LayoutDashboard },
  { id: 'plans', label: 'Plans', icon: Zap },
  { id: 'addons', label: 'Add-ons', icon: Package },
  { id: 'invoices', label: 'Invoices', icon: FileText },
  { id: 'payment', label: 'Payment', icon: CreditCard },
  { id: 'wallet', label: 'Credits', icon: Wallet },
  { id: 'usage', label: 'Usage', icon: BarChart3 },
];

function BillingHubSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-64" />
        </div>
      </div>
      <Skeleton className="h-10 w-full max-w-2xl" />
      <div className="grid gap-6">
        <Skeleton className="h-48 w-full" />
        <Skeleton className="h-48 w-full" />
      </div>
    </div>
  );
}

export function BillingHubPage() {
  const { state, actions, isLoading, errors, projectedBilling, usageMetrics, planLimits } = useBillingHub();

  const isInitialLoading = isLoading.subscription && !state.subscription;

  if (isInitialLoading) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-8">
        <BillingHubSkeleton />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      {/* Page Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="font-display text-3xl font-bold text-text-primary">Billing Hub</h1>
          <p className="text-text-secondary mt-1">
            Manage your subscription, invoices, and payment methods
          </p>
        </div>
        <Button
          variant="outline"
          onClick={actions.openBillingPortal}
          className="border-border-strong"
        >
          <CreditCard className="mr-2 h-4 w-4" />
          Open Billing Portal
        </Button>
      </div>

      {/* Tab Navigation */}
      <Tabs defaultValue="overview" className="w-full">
        <div className="overflow-x-auto pb-2">
          <TabsList className="mb-6 grid w-full grid-cols-7 gap-1">
            {TABS.map((tab) => {
              const Icon = tab.icon;
              return (
                <TabsTrigger
                  key={tab.id}
                  value={tab.id}
                  className="flex items-center gap-2 text-xs sm:text-sm"
                >
                  <Icon className="h-4 w-4" />
                  <span className="hidden sm:inline">{tab.label}</span>
                </TabsTrigger>
              );
            })}
          </TabsList>
        </div>

        <TabsContent value="overview">
          <OverviewTab
            subscription={state.subscription}
            walletInfo={state.walletInfo}
            paymentMethods={state.paymentMethods}
            projectedBilling={projectedBilling}
            usageMetrics={usageMetrics}
            planLimits={planLimits}
            isLoading={isLoading}
            errors={errors}
            onOpenPortal={actions.openBillingPortal}
          />
        </TabsContent>

        <TabsContent value="plans">
          <PlansTab
            subscription={state.subscription}
            onOpenPortal={actions.openBillingPortal}
          />
        </TabsContent>

        <TabsContent value="addons">
          <AddOnsTab
            addOnCatalog={state.addOnCatalog}
            entitledAddOnIds={state.entitledAddOnIds}
            onPurchase={actions.handleAddOnPurchase}
            isLoading={isLoading.addOns}
          />
        </TabsContent>

        <TabsContent value="invoices">
          <InvoicesTab
            invoices={state.invoices}
            isLoading={isLoading.invoices}
            error={errors.invoices}
          />
        </TabsContent>

        <TabsContent value="payment">
          <PaymentMethodsTab
            paymentMethods={state.paymentMethods}
            isLoading={isLoading.paymentMethods}
            error={errors.paymentMethods}
            onOpenPortal={actions.openBillingPortal}
          />
        </TabsContent>

        <TabsContent value="wallet">
          <WalletTab
            walletInfo={state.walletInfo}
            walletTransactions={state.walletTransactions}
            isLoading={isLoading.wallet}
            error={errors.wallet}
            onTopUp={actions.handleTopUp}
          />
        </TabsContent>

        <TabsContent value="usage">
          <UsageTab
            usageData={state.usageData}
            projectedBilling={projectedBilling}
            usageMetrics={usageMetrics}
            costData={state.costData}
            isLoading={isLoading.usage}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}