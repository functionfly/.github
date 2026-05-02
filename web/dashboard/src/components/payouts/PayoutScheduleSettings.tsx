import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { formatPayoutUsd } from '@/lib/payout-utils';
import { usePayoutSchedule, useUpdatePayoutSchedule } from '@/hooks/usePayouts';
import { CalendarClock, Loader2 } from 'lucide-react';
import { useState, useEffect } from 'react';

const FREQUENCIES = [
  { value: 'weekly', label: 'Weekly' },
  { value: 'biweekly', label: 'Every two weeks' },
  { value: 'monthly', label: 'Monthly' },
] as const;

const DAYS_OF_WEEK = [
  { value: '0', label: 'Sunday' },
  { value: '1', label: 'Monday' },
  { value: '2', label: 'Tuesday' },
  { value: '3', label: 'Wednesday' },
  { value: '4', label: 'Thursday' },
  { value: '5', label: 'Friday' },
  { value: '6', label: 'Saturday' },
] as const;

export function PayoutScheduleSettings() {
  const { data: schedule, isLoading } = usePayoutSchedule();
  const updateMutation = useUpdatePayoutSchedule();

  const [enabled, setEnabled] = useState(false);
  const [frequency, setFrequency] = useState<'weekly' | 'biweekly' | 'monthly'>('weekly');
  const [minAmount, setMinAmount] = useState('50');
  const [dayOfWeek, setDayOfWeek] = useState('1');
  const [dayOfMonth, setDayOfMonth] = useState('1');

  useEffect(() => {
    if (schedule) {
      setEnabled(schedule.schedule_enabled);
      setFrequency(schedule.frequency);
      setMinAmount((schedule.minimum_amount_cents / 100).toString());
      if (schedule.day_of_week != null) setDayOfWeek(schedule.day_of_week.toString());
      if (schedule.day_of_month != null) setDayOfMonth(schedule.day_of_month.toString());
    }
  }, [schedule]);

  const handleSave = () => {
    const minCents = Math.round(parseFloat(minAmount || '0') * 100);
    updateMutation.mutate({
      schedule_enabled: enabled,
      frequency,
      minimum_amount_cents: minCents,
      day_of_week: frequency !== 'monthly' ? parseInt(dayOfWeek, 10) : undefined,
      day_of_month: frequency === 'monthly' ? parseInt(dayOfMonth, 10) : undefined,
    });
  };

  const nextPayout = schedule?.next_scheduled_at
    ? new Date(schedule.next_scheduled_at).toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : null;

  return (
    <Card className="border-border-default">
      <CardHeader>
        <div className="flex items-center gap-2">
          <CalendarClock className="h-5 w-5 text-text-muted" />
          <div>
            <CardTitle className="text-lg">Auto-payout schedule</CardTitle>
            <CardDescription className="text-text-secondary">
              Automatically withdraw available earnings on a schedule.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-5">
        {isLoading ? (
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        ) : (
          <>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="auto-payout-toggle">Enable auto-payouts</Label>
                <p className="text-xs text-text-muted">
                  Automatically withdraw available balance when it meets the minimum.
                </p>
              </div>
              <Switch
                id="auto-payout-toggle"
                checked={enabled}
                onCheckedChange={setEnabled}
              />
            </div>

            {enabled && (
              <div className="space-y-4 border-t border-border-subtle pt-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label>Frequency</Label>
                    <Select value={frequency} onValueChange={(v) => setFrequency(v as typeof frequency)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {FREQUENCIES.map((f) => (
                          <SelectItem key={f.value} value={f.value}>
                            {f.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {frequency !== 'monthly' && (
                    <div className="space-y-2">
                      <Label>Day of week</Label>
                      <Select value={dayOfWeek} onValueChange={setDayOfWeek}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {DAYS_OF_WEEK.map((d) => (
                            <SelectItem key={d.value} value={d.value}>
                              {d.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}

                  {frequency === 'monthly' && (
                    <div className="space-y-2">
                      <Label>Day of month</Label>
                      <Select value={dayOfMonth} onValueChange={setDayOfMonth}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {Array.from({ length: 28 }, (_, i) => i + 1).map((d) => (
                            <SelectItem key={d} value={d.toString()}>
                              {d}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="min-payout-amount">Minimum balance to trigger payout (USD)</Label>
                  <Input
                    id="min-payout-amount"
                    type="text"
                    inputMode="decimal"
                    placeholder="50.00"
                    value={minAmount}
                    onChange={(e) => setMinAmount(e.target.value)}
                    className="max-w-[200px] font-mono"
                  />
                  <p className="text-xs text-text-muted">
                    Auto-payouts only trigger when your available balance is at least this amount.
                  </p>
                </div>
              </div>
            )}

            {nextPayout && enabled && (
              <div className="flex items-center gap-2 text-sm text-text-secondary">
                <span>Next scheduled payout:</span>
                <Badge variant="secondary">{nextPayout}</Badge>
              </div>
            )}

            {schedule?.last_auto_payout_at && (
              <p className="text-xs text-text-muted">
                Last auto-payout:{' '}
                {new Date(schedule.last_auto_payout_at).toLocaleDateString()}
              </p>
            )}

            <Button
              onClick={handleSave}
              disabled={updateMutation.isPending}
            >
              {updateMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save schedule
            </Button>
          </>
        )}
      </CardContent>
    </Card>
  );
}
