import { AffiliatePanel } from '@/components/payouts/AffiliatePanel';
import { PublisherPayoutsPanel } from '@/components/payouts/PublisherPayoutsPanel';
import { PayoutScheduleSettings } from '@/components/payouts/PayoutScheduleSettings';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ROUTES } from '@/lib/constants';
import { useAuthStore } from '@/stores/authStore';
import { ArrowLeft } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';

export function PayoutsSettingsPage() {
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);

  const settingsHref = user?.username ? `/u/${user.username}/settings` : ROUTES.SETTINGS;

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-start gap-3">
          <Button
            variant="ghost"
            size="icon"
            className="mt-0.5 shrink-0 text-text-secondary hover:text-text-primary"
            aria-label="Back to settings"
            onClick={() => navigate(settingsHref)}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">Publisher payouts</h1>
            <p className="text-text-secondary">
              Connect a bank account via Stripe, manage payout schedule, and withdraw earnings.
            </p>
            <p className="mt-2 text-sm text-text-muted">
              Looking for subscription billing?{' '}
              <Link
                to={user?.username ? `/u/${user.username}/settings/billing` : ROUTES.SETTINGS}
                className="text-primary underline-offset-4 hover:underline"
              >
                Open Billing
              </Link>
            </p>
          </div>
        </div>
      </div>

      <Tabs defaultValue="payouts" className="space-y-6">
        <TabsList>
          <TabsTrigger value="payouts">Payouts</TabsTrigger>
          <TabsTrigger value="schedule">Schedule</TabsTrigger>
          <TabsTrigger value="affiliate">Affiliate</TabsTrigger>
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
    </div>
  );
}

export default PayoutsSettingsPage;
