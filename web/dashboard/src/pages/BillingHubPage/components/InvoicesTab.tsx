import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import type { Invoice } from '@/api/billing';
import { getInvoicesErrorMessage } from '@/api/billing';
import { formatCurrency, formatDate } from '@/pages/SettingsPage/settings-utils';
import { CreditCard, Download, FileText } from 'lucide-react';

interface InvoicesTabProps {
  invoices: Invoice[];
  isLoading: boolean;
  error: Error | null;
}

export function InvoicesTab({ invoices, isLoading, error }: InvoicesTabProps) {
  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <FileText className="h-5 w-5 text-brand-500" />
            Invoices
          </CardTitle>
          <CardDescription>View and download your past invoices</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : error ? (
            <div className="flex items-center gap-2 p-4 rounded-lg bg-amber-500/10 border border-amber-500/20">
              <span className="w-5 h-5 text-amber-500 shrink-0">⚠️</span>
              <p className="text-amber-500 text-sm">{getInvoicesErrorMessage(error)}</p>
            </div>
          ) : invoices.length === 0 ? (
            <div className="text-center py-12">
              <FileText className="h-12 w-12 text-text-muted mx-auto mb-3" />
              <p className="text-text-muted">No invoices yet</p>
              <p className="text-sm text-text-muted">
                Your invoices will appear here after your first payment
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {invoices.map((invoice) => (
                <div
                  key={invoice.id}
                  className="flex items-center justify-between p-4 rounded-lg bg-bg-secondary border border-border-default hover:border-border-strong transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-brand-500/20 flex items-center justify-center">
                      <CreditCard className="w-5 h-5 text-brand-500" />
                    </div>
                    <div>
                      <p className="font-medium font-mono tabular-nums text-text-primary">
                        {formatCurrency(invoice.amount, invoice.currency)}
                      </p>
                      <p className="text-sm text-text-muted">
                        {invoice.invoice_date ? formatDate(invoice.invoice_date) : '—'}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge
                      variant={invoice.status === 'paid' ? 'default' : 'secondary'}
                      className={invoice.status === 'paid' ? 'ff-badge-success' : ''}
                    >
                      {invoice.status}
                    </Badge>
                    {invoice.invoice_pdf || invoice.hosted_invoice_url ? (
                      <Button variant="ghost" size="sm" asChild>
                        <a
                          href={invoice.invoice_pdf || invoice.hosted_invoice_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1"
                        >
                          <Download className="w-4 h-4" />
                          Download
                        </a>
                      </Button>
                    ) : invoice.status === 'paid' ? (
                      <span className="text-xs text-text-muted">Processing...</span>
                    ) : null}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}