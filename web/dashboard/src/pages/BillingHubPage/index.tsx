import { useState } from 'react';
import { useBillingHub } from './useBillingHub';
import { usePageTitle } from '@/hooks';
import { OverviewTab } from './components/OverviewTab';
import { PlansTab } from './components/PlansTab';
import { AddOnsTab } from './components/AddOnsTab';
import { InvoicesTab } from './components/InvoicesTab';
import { PaymentMethodsTab } from './components/PaymentMethodsTab';
import { WalletTab } from './components/WalletTab';
import { UsageTab } from './components/UsageTab';
import {
  Chamber,
  CornerBrace,
  PageGrid,
  SealedButton,
  FrameButton,
  AnnotationTag,
} from '@/components/containment';
import {
  CreditCard,
  FileText,
  LayoutDashboard,
  Wallet,
  Zap,
  BarChart3,
  Package,
} from 'lucide-react';
import '@/styles/sc-billing.css';

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
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
          <div style={{ height: 32, width: 192, background: 'var(--panel-raised)', borderRadius: 'var(--radius)' }} />
          <div style={{ height: 16, width: 256, background: 'var(--panel-raised)', borderRadius: 'var(--radius)' }} />
        </div>
      </div>
      <div style={{ height: 40, width: '100%', maxWidth: 640, background: 'var(--panel-raised)', borderRadius: 'var(--radius)' }} />
      <Chamber nested>
        <div style={{ height: 192, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
      </Chamber>
      <Chamber nested>
        <div style={{ height: 192, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
      </Chamber>
    </div>
  );
}

export function BillingHubPage() {
  usePageTitle('Billing');
  const { state, actions, isLoading, errors, projectedBilling, usageMetrics, planLimits } = useBillingHub();
  const [activeTab, setActiveTab] = useState<BillingTab>('overview');

  const isInitialLoading = isLoading.subscription && !state.subscription;

  if (isInitialLoading) {
    return (
      <div className="sc-billing-page">
        <BillingHubSkeleton />
      </div>
    );
  }

  return (
    <div className="sc-billing-page">
      <PageGrid />

      {/* Page Header */}
      <div className="sc-billing-header">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 'var(--space-4)' }}>
          <div>
            <h1 className="sc-billing-title">
              <span className="icon-wrapper">
                <CreditCard style={{ width: 18, height: 18, color: 'var(--bg)' }} />
              </span>
              Billing Hub
            </h1>
            <p className="sc-billing-subtitle">
              Manage your subscription, invoices, and payment methods
            </p>
          </div>
          <FrameButton
            onClick={actions.openBillingPortal}
            iconLeft={<CreditCard style={{ width: 14, height: 14 }} />}
          >
            Open Billing Portal
          </FrameButton>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="sc-billing-tabs" role="tablist">
        {TABS.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              role="tab"
              aria-selected={isActive}
              className={`sc-billing-tab ${isActive ? '' : ''}`}
              data-state={isActive ? 'active' : 'inactive'}
              onClick={() => setActiveTab(tab.id)}
            >
              <Icon style={{ width: 14, height: 14 }} />
              <span>{tab.label}</span>
            </button>
          );
        })}
      </div>

      {/* Tab Content */}
      <div style={{ paddingTop: 'var(--space-6)' }}>
        {activeTab === 'overview' && (
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
        )}

        {activeTab === 'plans' && (
          <PlansTab
            subscription={state.subscription}
            onOpenPortal={actions.openBillingPortal}
          />
        )}

        {activeTab === 'addons' && (
          <AddOnsTab
            addOnCatalog={state.addOnCatalog}
            entitledAddOnIds={state.entitledAddOnIds}
            onPurchase={actions.handleAddOnPurchase}
            isLoading={isLoading.addOns}
          />
        )}

        {activeTab === 'invoices' && (
          <InvoicesTab
            invoices={state.invoices}
            isLoading={isLoading.invoices}
            error={errors.invoices}
          />
        )}

        {activeTab === 'payment' && (
          <PaymentMethodsTab
            paymentMethods={state.paymentMethods}
            isLoading={isLoading.paymentMethods}
            error={errors.paymentMethods}
            onOpenPortal={actions.openBillingPortal}
          />
        )}

        {activeTab === 'wallet' && (
          <WalletTab
            walletInfo={state.walletInfo}
            walletTransactions={state.walletTransactions}
            isLoading={isLoading.wallet}
            error={errors.wallet}
            onTopUp={actions.handleTopUp}
          />
        )}

        {activeTab === 'usage' && (
          <UsageTab
            usageData={state.usageData}
            projectedBilling={projectedBilling}
            usageMetrics={usageMetrics}
            costData={state.costData}
            isLoading={isLoading.usage}
          />
        )}
      </div>
    </div>
  );
}
