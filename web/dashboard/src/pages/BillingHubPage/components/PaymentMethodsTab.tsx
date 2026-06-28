import {
  Chamber,
  FrameButton,
  StatusPill,
} from '@/components/containment';
import type { PaymentMethod } from '@/api/billing';
import { CreditCard, Plus, Trash2, AlertCircle } from 'lucide-react';

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
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <Chamber nested>
        <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
          <div className="sc-billing-card-title">
            <CreditCard style={{ width: 14, height: 14 }} />
            Payment Methods
          </div>
          <div className="sc-billing-card-description">Manage your payment methods for subscriptions and top-ups</div>
        </div>

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            {[1, 2].map((i) => (
              <div key={i} style={{ height: 80, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
            ))}
          </div>
        ) : error ? (
          <div className="sc-billing-info sc-billing-info-warning">
            <AlertCircle style={{ width: 18, height: 18 }} />
            <div className="sc-billing-info-content">
              <div className="sc-billing-info-text">{error.message}</div>
            </div>
          </div>
        ) : paymentMethods.length === 0 ? (
          <div className="empty-state" style={{ minHeight: 160, flexDirection: 'column', gap: 'var(--space-3)' }}>
            <CreditCard style={{ width: 48, height: 48, color: 'var(--text-faint)' }} />
            <p style={{ color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.06em' }}>No payment methods</p>
            <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-4)' }}>Add a payment method to manage your subscription</p>
            <FrameButton onClick={onOpenPortal} iconLeft={<Plus style={{ width: 14, height: 14 }} />}>
              Add Payment Method
            </FrameButton>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            {paymentMethods.map((pm) => (
              <div
                key={pm.stripe_payment_method_id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: 'var(--space-4)',
                  borderRadius: 'var(--radius)',
                  background: 'var(--panel)',
                  border: '1px solid var(--panel-edge)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
                  <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: 'rgba(143, 255, 208, 0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    <CreditCard style={{ width: 18, height: 18, color: 'var(--status-ok)' }} />
                  </div>
                  <div>
                    <p style={{ fontWeight: 500, color: 'var(--text)' }}>
                      {pm.brand} •••• {pm.last4}
                    </p>
                    <p style={{ fontSize: 12, color: 'var(--text-dim)' }}>
                      Expires {pm.exp_month.toString().padStart(2, '0')}/{pm.exp_year}
                    </p>
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                  {pm.is_default && <StatusPill status="live" label="Default" />}
                  <button
                    onClick={onOpenPortal}
                    style={{
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      width: 28, height: 28, borderRadius: 'var(--radius)',
                      background: 'transparent', border: 'none', cursor: 'pointer',
                      color: 'var(--status-revoked)',
                      transition: 'background var(--duration-fast) var(--ease-out)',
                    }}
                    title="Remove payment method"
                  >
                    <Trash2 style={{ width: 14, height: 14 }} />
                  </button>
                </div>
              </div>
            ))}

            <div style={{ paddingTop: 'var(--space-4)', borderTop: '1px solid var(--panel-edge)' }}>
              <FrameButton onClick={onOpenPortal} iconLeft={<Plus style={{ width: 14, height: 14 }} />}>
                Add Payment Method
              </FrameButton>
            </div>
          </div>
        )}
      </Chamber>

      {/* Billing Portal */}
      <Chamber nested>
        <div className="sc-billing-card-title" style={{ marginBottom: 'var(--space-2)' }}>Billing Portal</div>
        <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-4)' }}>
          For advanced payment method management, invoice download, and subscription changes, use our secure billing portal.
        </p>
        <FrameButton onClick={onOpenPortal} iconLeft={<CreditCard style={{ width: 14, height: 14 }} />}>
          Open Billing Portal
        </FrameButton>
      </Chamber>
    </div>
  );
}
