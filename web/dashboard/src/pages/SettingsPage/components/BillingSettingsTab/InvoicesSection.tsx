import type { Invoice } from '@/api/billing';
import { getInvoicesErrorMessage } from '@/api/billing';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Chamber, CornerBrace, FrameButton, StatusPill } from '@/components/sc';
import { CreditCard, Download } from 'lucide-react';

interface InvoicesSectionProps {
  invoices: Invoice[];
  invoicesLoading: boolean;
  invoicesError: Error | null;
  formatCurrency: (amount: number, currency: string) => string;
  formatDate: (date: string) => string;
}

export function InvoicesSection({
  invoices,
  invoicesLoading,
  invoicesError,
  formatCurrency,
  formatDate,
}: InvoicesSectionProps) {
  return (
    <Chamber>
      <CornerBrace position="tl" />
      <CornerBrace position="br" />
      <div className="mb-4">
        <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
          Invoices
        </h3>
        <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
          View and download your past invoices
        </p>
      </div>
      <div>
        {invoicesLoading ? (
          <div className="flex items-center justify-center p-4">
            <div
              className="animate-spin rounded-full h-6 w-6 border-b-2"
              style={{ borderColor: 'var(--accent)' }}
            />
          </div>
        ) : invoicesError ? (
          <div
            className="flex items-center gap-2 p-4 rounded-lg"
            style={{
              background: 'rgba(232, 196, 104, 0.06)',
              border: '1px solid rgba(232, 196, 104, 0.3)',
            }}
          >
            <StatusPill status="pending" label="Error" />
            <p className="text-sm" style={{ color: 'var(--status-pending)' }}>
              {getInvoicesErrorMessage(invoicesError)}
            </p>
          </div>
        ) : invoices.length === 0 ? (
          <div className="text-center p-6">
            <CreditCard className="w-12 h-12 mx-auto mb-3" style={{ color: 'var(--text-faint)' }} />
            <p style={{ color: 'var(--text-dim)' }}>No invoices yet</p>
            <p className="text-sm" style={{ color: 'var(--text-faint)' }}>
              Your invoices will appear here after your first payment
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {invoices.map((invoice) => (
              <div
                key={invoice.id}
                className="flex items-center justify-between p-4 rounded-lg hover:border-border-strong transition-colors"
                style={{ background: 'var(--panel-raised)', border: '1px solid var(--panel-edge)' }}
              >
                <div className="flex items-center gap-4">
                  <div
                    className="w-10 h-10 rounded-lg flex items-center justify-center"
                    style={{ background: 'rgba(59, 130, 246, 0.1)' }}
                  >
                    <CreditCard className="w-5 h-5" style={{ color: 'var(--accent)' }} />
                  </div>
                  <div>
                    <p
                      className="font-medium font-mono tabular-nums"
                      style={{ color: 'var(--text)' }}
                    >
                      {formatCurrency(invoice.amount, invoice.currency)}
                    </p>
                    <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                      {formatDate(invoice.invoice_date)}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <StatusPill
                    status={invoice.status === 'paid' ? 'live' : 'pending'}
                    label={invoice.status}
                  />
                  {invoice.invoice_pdf || invoice.hosted_invoice_url ? (
                    <FrameButton size="sm" iconLeft={<Download className="w-4 h-4" />}>
                      <a
                        href={invoice.invoice_pdf || invoice.hosted_invoice_url}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        Download
                      </a>
                    </FrameButton>
                  ) : invoice.status === 'paid' ? (
                    <span
                      className="text-xs"
                      style={{ color: 'var(--text-faint)' }}
                      title="Invoice PDF will be available shortly"
                    >
                      Processing...
                    </span>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </Chamber>
  );
}
