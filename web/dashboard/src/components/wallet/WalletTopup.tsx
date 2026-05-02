import { useState, useCallback } from 'react';
import { agentApi } from '@/api/agent';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { toast } from 'sonner';
import {
  Plus,
  CreditCard,
  DollarSign,
  Loader2,
  Wallet,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface WalletTopupProps {
  agentId: string;
  currentBalance?: number;
  className?: string;
  onSuccess?: (newBalance: number) => void;
}

const AMOUNT_PRESETS = [10, 25, 50, 100];
const MINIMUM_AMOUNT = 0.5;

export function WalletTopup({
  agentId,
  currentBalance,
  className,
  onSuccess,
}: WalletTopupProps) {
  const { t } = useTranslation();
  const [showDialog, setShowDialog] = useState(false);
  const [amount, setAmount] = useState('25');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value);
  };

  const validateAmount = useCallback((value: string): string | null => {
    const num = parseFloat(value);
    if (isNaN(num) || num <= 0) {
      return t('walletTopup:invalidAmount');
    }
    if (num < MINIMUM_AMOUNT) {
      return t('walletTopup:minimumAmount', { amount: formatCurrency(MINIMUM_AMOUNT) });
    }
    return null;
  }, [t]);

  const handlePresetClick = (preset: number) => {
    setAmount(preset.toString());
    setError(null);
  };

  const handleAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAmount(e.target.value);
    setError(null);
  };

  const handleSubmit = async () => {
    const validationError = validateAmount(amount);
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsLoading(true);
    setError(null);

    const amountNum = parseFloat(amount);

    try {
      const successUrl = `${window.location.origin}/wallet/agents/${encodeURIComponent(agentId)}?credits=success`;
      const cancelUrl = `${window.location.origin}/wallet/agents/${encodeURIComponent(agentId)}?credits=cancel`;
      
      const res = await agentApi.createCreditsCheckout(agentId, amountNum, successUrl, cancelUrl);
      
      if (res?.url) {
        window.location.href = res.url;
        return;
      }

      setError(t('walletTopup:checkoutError'));
    } catch (e: unknown) {
      const err = e as {
        response?: { data?: { error?: { code?: string; message?: string } } };
        message?: string;
      };
      const errObj = err.response?.data?.error;

      if (errObj?.code === 'PAYMENTS_NOT_CONFIGURED') {
        try {
          const purchaseRes = await agentApi.purchaseCredits(agentId, amountNum);
          const newBalance = purchaseRes?.new_balance_usd ?? currentBalance ?? 0;
          setShowDialog(false);
          toast.success(t('walletTopup:purchaseSuccess'), {
            description: t('walletTopup:fundsAdded', { amount: formatCurrency(amountNum) }),
          });
          onSuccess?.(newBalance);
          return;
        } catch (innerErr: unknown) {
          const msg = innerErr instanceof Error ? innerErr.message : t('walletTopup:purchaseError');
          setError(msg);
          return;
        }
      }

      const apiMsg = errObj?.message && (errObj.code ? `${errObj.code}: ${errObj.message}` : errObj.message);
      setError(apiMsg ?? (e instanceof Error ? e.message : t('walletTopup:checkoutError')));
    } finally {
      setIsLoading(false);
    }
  };

  const handleDialogClose = (open: boolean) => {
    if (!open) {
      setShowDialog(false);
      setAmount('25');
      setError(null);
    }
  };

  return (
    <div className={className}>
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Wallet className="h-5 w-5" />
            {t('walletTopup:walletBalance')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">{t('walletTopup:currentBalance')}</span>
            <span className="text-2xl font-bold">
              {formatCurrency(currentBalance ?? 0)}
            </span>
          </div>
          <Button
            variant="default"
            className="w-full"
            onClick={() => setShowDialog(true)}
          >
            <Plus className="h-4 w-4 mr-2" />
            {t('walletTopup:addFunds')}
          </Button>
        </CardContent>
      </Card>

      <Dialog open={showDialog} onOpenChange={handleDialogClose}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <CreditCard className="h-5 w-5" />
              {t('walletTopup:addFundsDialog')}
            </DialogTitle>
            <DialogDescription>
              {t('walletTopup:addFundsDescription')}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6 py-4">
            <div className="space-y-2">
              <Label htmlFor="amount">{t('walletTopup:selectAmount')}</Label>
              <div className="grid grid-cols-4 gap-2">
                {AMOUNT_PRESETS.map((preset) => (
                  <Button
                    key={preset}
                    type="button"
                    variant={amount === preset.toString() ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => handlePresetClick(preset)}
                  >
                    ${preset}
                  </Button>
                ))}
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="custom-amount">{t('walletTopup:customAmount')}</Label>
              <div className="relative">
                <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="custom-amount"
                  type="number"
                  step="0.01"
                  min={MINIMUM_AMOUNT}
                  placeholder="0.00"
                  value={amount}
                  onChange={handleAmountChange}
                  className="pl-9"
                />
              </div>
              <p className="text-xs text-muted-foreground">
                {t('walletTopup:minimumHint', { amount: formatCurrency(MINIMUM_AMOUNT) })}
              </p>
            </div>

            {error && (
              <p className="text-sm text-destructive">{error}</p>
            )}
          </div>

          <DialogFooter className="gap-2 sm:gap-0">
            <Button
              variant="outline"
              onClick={() => setShowDialog(false)}
              disabled={isLoading}
            >
              {t('walletTopup:cancel')}
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={isLoading}
              isLoading={isLoading}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {t('walletTopup:processing')}
                </>
              ) : (
                <>
                  <CreditCard className="mr-2 h-4 w-4" />
                  {t('walletTopup:continueToCheckout')}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
