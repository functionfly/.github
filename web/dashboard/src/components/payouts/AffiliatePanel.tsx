import {
  applyAffiliateCode,
  getAffiliateEarningsSummary,
  getMyAffiliateCodes,
  getMyAffiliateCommissions,
  getMyAffiliateReferrals,
  type AffiliateCode,
  type AffiliateCommission,
  type AffiliateEarningsSummary,
  type AffiliateReferral,
} from '@/api/billing';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { formatDate } from '@/pages/SettingsPage/settings-utils';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { DollarSign, Gift, Plus, TrendingUp, Users, Loader2 } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';

export function AffiliatePanel() {
  const queryClient = useQueryClient();
  const [applyCodeModalOpen, setApplyCodeModalOpen] = useState(false);
  const [codeInput, setCodeInput] = useState('');
  const [applyLoading, setApplyLoading] = useState(false);

  const { data: summaryData, isLoading: summaryLoading } = useQuery({
    queryKey: ['affiliate', 'earnings-summary'],
    queryFn: getAffiliateEarningsSummary,
    retry: false,
  });

  const { data: codesData, isLoading: codesLoading } = useQuery({
    queryKey: ['affiliate', 'codes'],
    queryFn: getMyAffiliateCodes,
    retry: false,
  });

  const { data: referralsData, isLoading: referralsLoading } = useQuery({
    queryKey: ['affiliate', 'referrals'],
    queryFn: getMyAffiliateReferrals,
    retry: false,
  });

  const { data: commissionsData, isLoading: commissionsLoading } = useQuery({
    queryKey: ['affiliate', 'commissions'],
    queryFn: getMyAffiliateCommissions,
    retry: false,
  });

  const applyCodeMutation = useMutation({
    mutationFn: applyAffiliateCode,
    onSuccess: (data) => {
      toast.success(`Affiliate code "${data.code}" applied!`, {
        description: `You'll earn ${data.commission_type === 'percent' ? `${data.commission_value}%` : `$${data.commission_value}`} on qualifying subscriptions.`,
      });
      setApplyCodeModalOpen(false);
      setCodeInput('');
      queryClient.invalidateQueries({ queryKey: ['affiliate'] });
    },
    onError: (err) => {
      const msg = err instanceof Error ? err.message : 'Failed to apply code';
      toast.error(msg);
    },
  });

  const handleApplyCode = async () => {
    if (!codeInput.trim()) {
      toast.error('Please enter an affiliate code');
      return;
    }
    setApplyLoading(true);
    try {
      await applyCodeMutation.mutateAsync({ code: codeInput.trim().toUpperCase() });
    } finally {
      setApplyLoading(false);
    }
  };

  const codes: AffiliateCode[] = codesData ?? [];
  const referrals: AffiliateReferral[] = referralsData ?? [];
  const commissions: AffiliateCommission[] = commissionsData ?? [];
  const summary: AffiliateEarningsSummary | undefined = summaryData;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Gift className="w-6 h-6 text-purple-600" />
          <div>
            <h2 className="text-xl font-semibold">Affiliate & Referral Codes</h2>
            <p className="text-sm text-text-secondary">
              Earn commissions when others use your codes to subscribe.
            </p>
          </div>
        </div>
        <Dialog open={applyCodeModalOpen} onOpenChange={setApplyCodeModalOpen}>
          <DialogTrigger asChild>
            <Button variant="outline" className="gap-2">
              <Plus className="w-4 h-4" />
              Apply Code
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Apply Affiliate Code</DialogTitle>
              <DialogDescription>
                Enter a referral code to link your account and earn commissions on qualifying
                subscriptions.
              </DialogDescription>
            </DialogHeader>
            <div className="py-4">
              <Label htmlFor="affiliate-code">Affiliate Code</Label>
              <Input
                id="affiliate-code"
                value={codeInput}
                onChange={(e) => setCodeInput(e.target.value.toUpperCase())}
                placeholder="e.g. SAVE20"
                className="mt-2 font-mono text-lg uppercase"
              />
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setApplyCodeModalOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleApplyCode}
                disabled={applyLoading || !codeInput.trim()}
                className="bg-purple-600 hover:bg-purple-700"
              >
                {applyLoading ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : null}
                Apply Code
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Earnings Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
              <DollarSign className="w-5 h-5" />
              <span className="text-sm font-medium">Pending Earnings</span>
            </div>
            <p className="mt-1 text-2xl font-bold">
              ${((summary?.pending_earnings_cents ?? 0) / 100).toFixed(2)}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-blue-600 dark:text-blue-400">
              <TrendingUp className="w-5 h-5" />
              <span className="text-sm font-medium">Total Earnings</span>
            </div>
            <p className="mt-1 text-2xl font-bold">
              ${((summary?.total_earnings_cents ?? 0) / 100).toFixed(2)}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-purple-600 dark:text-purple-400">
              <Users className="w-5 h-5" />
              <span className="text-sm font-medium">Total Referrals</span>
            </div>
            <p className="mt-1 text-2xl font-bold">{summary?.total_referrals ?? 0}</p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-amber-600 dark:text-amber-400">
              <Gift className="w-5 h-5" />
              <span className="text-sm font-medium">Your Codes</span>
            </div>
            <p className="mt-1 text-2xl font-bold">{summary?.codes_count ?? 0}</p>
          </CardContent>
        </Card>
      </div>

      {/* My Codes */}
      <Card>
        <CardHeader>
          <CardTitle>Your Affiliate Codes</CardTitle>
          <CardDescription>Share these codes to earn commissions on new subscriptions.</CardDescription>
        </CardHeader>
        <CardContent>
          {codesLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : codes.length === 0 ? (
            <Alert>
              <AlertTitle>No affiliate codes yet</AlertTitle>
              <AlertDescription>
                Contact support to get your own affiliate code and start earning commissions.
              </AlertDescription>
            </Alert>
          ) : (
            <div className="space-y-3">
              {codes.map((code) => (
                <div
                  key={code.id}
                  className="flex items-center justify-between p-4 border rounded-lg"
                >
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-lg font-semibold text-purple-700 dark:text-purple-300">
                        {code.code}
                      </span>
                      <Badge variant={code.is_active ? 'default' : 'secondary'}>
                        {code.is_active ? 'active' : 'inactive'}
                      </Badge>
                      <Badge variant="outline">
                        {code.commission_type === 'percent'
                          ? `${code.commission_value}%`
                          : `$${code.commission_value} fixed`}
                      </Badge>
                    </div>
                    <p className="text-sm text-text-secondary mt-1">{code.name}</p>
                  </div>
                  <div className="text-right text-sm">
                    <p>
                      <span className="text-text-muted">Referrals: </span>
                      <span className="font-medium">{code.total_referrals}</span>
                    </p>
                    <p>
                      <span className="text-text-muted">Paid out: </span>
                      <span className="font-medium text-green-600">
                        ${(code.paid_out_earnings_cents / 100).toFixed(2)}
                      </span>
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Referrals */}
      <Card>
        <CardHeader>
          <CardTitle>Your Referrals</CardTitle>
          <CardDescription>People who signed up using your affiliate codes.</CardDescription>
        </CardHeader>
        <CardContent>
          {referralsLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : referrals.length === 0 ? (
            <p className="text-sm text-text-muted text-center py-4">No referrals yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Status</TableHead>
                    <TableHead>Code Used</TableHead>
                    <TableHead>Referred At</TableHead>
                    <TableHead>UTM Source</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {referrals.map((ref) => (
                    <TableRow key={ref.id}>
                      <TableCell>
                        <Badge
                          variant={
                            ref.status === 'qualified'
                              ? 'default'
                              : ref.status === 'converted'
                              ? 'secondary'
                              : 'outline'
                          }
                        >
                          {ref.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono">{ref.affiliate_code_id.slice(0, 8)}...</TableCell>
                      <TableCell>{formatDate(ref.referred_at)}</TableCell>
                      <TableCell>{ref.utm_source || '—'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Commissions */}
      <Card>
        <CardHeader>
          <CardTitle>Commission History</CardTitle>
          <CardDescription>Your earned commissions from referrals.</CardDescription>
        </CardHeader>
        <CardContent>
          {commissionsLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : commissions.length === 0 ? (
            <p className="text-sm text-text-muted text-center py-4">No commissions yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Amount</TableHead>
                    <TableHead>Rate</TableHead>
                    <TableHead>Base Amount</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {commissions.map((comm) => (
                    <TableRow key={comm.id}>
                      <TableCell className="font-semibold text-green-600">
                        ${comm.commission_usd.toFixed(2)}
                      </TableCell>
                      <TableCell>
                        {comm.commission_type === 'percent'
                          ? `${comm.commission_value}%`
                          : `$${comm.commission_value} fixed`}
                      </TableCell>
                      <TableCell>${comm.base_amount_usd.toFixed(2)}</TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            comm.status === 'paid'
                              ? 'default'
                              : comm.status === 'approved'
                              ? 'secondary'
                              : 'outline'
                          }
                        >
                          {comm.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{formatDate(comm.created_at)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}