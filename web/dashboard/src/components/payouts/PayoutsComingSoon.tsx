import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { CalendarClock, Rocket } from 'lucide-react';

interface PayoutsComingSoonProps {
  compact?: boolean;
}

export function PayoutsComingSoon({ compact = false }: PayoutsComingSoonProps) {
  return (
    <Card className="border-border-default">
      <CardHeader className={compact ? 'p-4 pb-2' : undefined}>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="space-y-1">
            <CardTitle className={compact ? 'text-base' : 'text-lg'}>Publisher payouts</CardTitle>
            <CardDescription>
              Withdrawals and payout history are temporarily paused while we finalize rollout.
            </CardDescription>
          </div>
          <Badge variant="outline" className="gap-1">
            <CalendarClock className="h-3.5 w-3.5" />
            Coming soon
          </Badge>
        </div>
      </CardHeader>
      <CardContent className={compact ? 'p-4 pt-0' : undefined}>
        <Alert className="border-border-default bg-muted/40">
          <Rocket className="h-4 w-4" />
          <AlertTitle>Payouts are not available yet</AlertTitle>
          <AlertDescription>
            We are finishing the payouts experience and turning it back on shortly. You can keep using
            your wallet and the rest of your account normally in the meantime.
          </AlertDescription>
        </Alert>
      </CardContent>
    </Card>
  );
}
