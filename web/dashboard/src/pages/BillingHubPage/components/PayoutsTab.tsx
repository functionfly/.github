import { useState } from 'react';
import { AffiliatePanel } from '@/components/payouts/AffiliatePanel';
import { PublisherPayoutsPanel } from '@/components/payouts/PublisherPayoutsPanel';
import { PayoutScheduleSettings } from '@/components/payouts/PayoutScheduleSettings';
import { Chamber } from '@/components/containment';
import { Banknote, Calendar, Gift, Tabs, TabsContent, TabsList, TabsTrigger } from 'lucide-react';

type PayoutTab = 'payouts' | 'schedule' | 'affiliate';

const TABS: { id: PayoutTab; label: string; icon: typeof Banknote }[] = [
  { id: 'payouts', label: 'Payouts', icon: Banknote },
  { id: 'schedule', label: 'Schedule', icon: Calendar },
  { id: 'affiliate', label: 'Affiliate', icon: Gift },
];

export function PayoutsTab() {
  const [activeTab, setActiveTab] = useState<PayoutTab>('payouts');

  return (
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <Chamber nested>
        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as PayoutTab)} className="space-y-6">
          <TabsList>
            {TABS.map((tab) => {
              const Icon = tab.icon;
              return (
                <TabsTrigger key={tab.id} value={tab.id} className="gap-2">
                  <Icon style={{ width: 14, height: 14 }} />
                  {tab.label}
                </TabsTrigger>
              );
            })}
          </TabsList>

          <TabsContent value="payouts" className="space-y-6">
            <PublisherPayoutsPanel />
          </TabsContent>

          <TabsContent value="schedule" className="space-y-6">
            <PayoutScheduleSettings />
          </TabsContent>

          <TabsContent value="affiliate" className="space-y-6">
            <AffiliatePanel />
          </TabsContent>
        </Tabs>
      </Chamber>
    </div>
  );
}
