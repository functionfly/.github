import {
  Chamber,
  FrameButton,
  StatusPill,
} from '@/components/containment';
import type { Invoice } from '@/api/billing';
import { getInvoicesErrorMessage } from '@/api/billing';
import { formatCurrency, formatDate } from '@/pages/SettingsPage/settings-utils';
import { CreditCard, Download, FileText, AlertCircle } from 'lucide-react';

interface InvoicesTabProps {
  invoices: Invoice[];
  isLoading: boolean;
  error: Error | null;
}

export function InvoicesTab({ invoices, isLoading, error }: InvoicesTabProps) {
  return (
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <Chamber nested>
        <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
          <div className="sc-billing-card-title">
            <FileText style={{ width: 14, height: 14 }} />
            Invoices
          </div>
          <div className="sc-billing-card-description">View and download your past invoices</div>
        </div>

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            {[1, 2, 3].map((i) => (
              <div key={i} style={{ height: 64, background: 'var(--panel)', borderRadius: 'var(--radius)' }} />
            ))}
          </div>
        ) : error ? (
          <div className="sc-billing-info sc-billing-info-warning">
            <AlertCircle style={{ width: 18, height: 18 }} />
            <div className="sc-billing-info-content">
              <div className="sc-billing-info-text">{getInvoicesErrorMessage(error)}</div>
            </div>
          </div>
        ) : invoices.length === 0 ? (
          <div className="empty-state" style={{ minHeight: 160, flexDirection: 'column', gap: 'var(--space-3)' }}>
            <FileText style={{ width: 48, height: 48, color: 'var(--text-faint)' }} />
            <p style={{ color: 'var(--text-faint)', fontFamily: 'var(--font-mono)', fontSize: 11, fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.06em' }}>No invoices yet</p>
            <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>Your invoices will appear here after your first payment</p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
            {invoices.map((invoice) => (
              <div
                key={invoice.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: 'var(--space-4)',
                  borderRadius: 'var(--radius)',
                  background: 'var(--panel)',
                  border: '1px solid var(--panel-edge)',
                  transition: 'border-color var(--duration-fast) var(--ease-out)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
                  <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: 'rgba(143, 255, 208, 0.1)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    <CreditCard style={{ width: 18, height: 18, color: 'var(--status-ok)' }} />
                  </div>
                  <div>
                    <p style={{ fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 600, fontVariantNumeric: 'tabular-nums', color: 'var(--text)' }}>
                      {formatCurrency(invoice.amount, invoice.currency)}
                    </p>
                    <p style={{ fontSize: 12, color: 'var(--text-dim)' }}>
                      {invoice.invoice_date ? formatDate(invoice.invoice_date) : '—'}
                    </p>
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                  <StatusPill status={invoice.status === 'paid' ? 'live' : invoice.status === 'open' ? 'pending' : 'revoked'} label={invoice.status} />
                  {invoice.invoice_pdf || invoice.hosted_invoice_url ? (
                    <FrameButton
                      size="sm"
                      onClick={() => window.open(invoice.invoice_pdf || invoice.hosted_invoice_url, '_blank')}
                      iconLeft={<Download style={{ width: 12, height: 12 }} />}
                    >
                      Download
                    </FrameButton>
                  ) : invoice.status === 'paid' ? (
                    <span style={{ fontSize: 11, color: 'var(--text-faint)' }}>Processing...</span>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        )}
      </Chamber>
    </div>
  );
}
