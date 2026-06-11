import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import type { PaymentMethod } from '@/api/billing';
import { CreditCard, Plus, Trash2 } from 'lucide-react';

interface PaymentMethodsTabProps {
  paymentMethods: PaymentMethod[];
  isLoading: boolean;
  error: Error | null;
  onOpenPortal: () => void;
}

export function PaymentMethodsTab({
  paymentMethods,
  isLoading,
  error,
  onOpenPortal,
}: PaymentMethodsTabProps) {
  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <CreditCard className="h-5 w-5 text-brand-500" />
            Payment Methods
          </CardTitle>
          <CardDescription>Manage your payment methods for subscriptions and top-ups</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[1, 2].map((i) => (
                <Skeleton key={i} className="h-20 w-full" />
              ))}
            </div>
          ) : error ? (
            <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <span className="w-5 h-5 text-amber-500 shrink-0">⚠️</span>
              <p className="text-amber-500 text-sm">{error.message}</p>
            </div>
          ) : paymentMethods.length === 0 ? (
            <div className="text-center py-12">
              <CreditCard className="h-12 w-12 text-text-muted mx-auto mb-3" />
              <p className="text-text-muted">No payment methods</p>
              <p className="text-sm text-text-muted mb-4">
                Add a payment method to manage your subscription
              </p>
              <Button variant="outline" className="border-border-strong" onClick={onOpenPortal}>
                <Plus className="mr-2 h-4 w-4" />
                Add Payment Method
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {paymentMethods.map((pm) => (
                <div
                  key={pm.stripe_payment_method_id}
                  className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-default"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-brand-500/20 flex items-center justify-center">
                      <CreditCard className="w-5 h-5 text-brand-500" />
                    </div>
                    <div>
                      <p className="font-medium text-text-primary">
                        {pm.brand} •••• {pm.last4}
                      </p>
                      <p className="text-sm text-text-muted">
                        Expires {pm.exp_month.toString().padStart(2, '0')}/{pm.exp_year}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    {pm.is_default && (
                      <Badge variant="success" className="ff-badge-success">
                        Default
                      </Badge>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-red-400 hover:text-red-300 hover:bg-red-500/10"
                      onClick={onOpenPortal}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}

              <div className="pt-4 border-t border-border-default">
                <Button variant="outline" className="border-border-strong" onClick={onOpenPortal}>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Payment Method
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display">Billing Portal</CardTitle>
          <CardDescription>
            Manage your payment methods and subscription through our secure billing portal
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-text-muted mb-4">
            For advanced payment method management, invoice download, and subscription changes,
            use our secure billing portal.
          </p>
          <Button variant="outline" className="border-border-strong" onClick={onOpenPortal}>
            <CreditCard className="mr-2 h-4 w-4" />
            Open Billing Portal
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}