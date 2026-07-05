import {
  Chamber,
  FrameButton,
  StatusPill,
} from '@/components/containment';
import type { Invoice } from '@/api/billing';
import { getInvoicesErrorMessage } from '@/api/billing';
import { formatCurrency, formatDate } from '@/pages/SettingsPage/settings-utils';
import { CreditCard, Download, FileText, AlertCircle } from 'lucide-react';
import styles from './InvoicesTab.module.css';

interface InvoicesTabProps {
  invoices: Invoice[];
  isLoading: boolean;
  error: Error | null;
}

export function InvoicesTab({ invoices, isLoading, error }: InvoicesTabProps) {
  return (
    <div className={styles.container}>
      <Chamber nested>
        <div className={styles.header}>
          <div className={styles.headerTitle}>
            <FileText className={styles.headerIcon} />
            Invoices
          </div>
          <div className={styles.headerDesc}>View and download your past invoices</div>
        </div>

        {isLoading ? (
          <div className={styles.skeletonList}>
            {[1, 2, 3].map((i) => (
              <div key={i} className={styles.skeletonRow} />
            ))}
          </div>
        ) : error ? (
          <div className={styles.warningBox}>
            <AlertCircle className={styles.warningIcon} />
            <div className={styles.warningContent}>
              <p className={styles.warningText}>{getInvoicesErrorMessage(error)}</p>
            </div>
          </div>
        ) : invoices.length === 0 ? (
          <div className={styles.emptyState}>
            <FileText className={styles.emptyIcon} />
            <p className={styles.emptyTitle}>No invoices yet</p>
            <p className={styles.emptyDesc}>Your invoices will appear here after your first payment</p>
          </div>
        ) : (
          <div className={styles.invoiceList}>
            {invoices.map((invoice) => (
              <div key={invoice.id} className={styles.invoiceRow}>
                <div className={styles.invoiceInfo}>
                  <div className={styles.invoiceIcon}>
                    <CreditCard className={styles.invoiceIconInner} />
                  </div>
                  <div className={styles.invoiceDetails}>
                    <p className={styles.invoiceAmount}>
                      {formatCurrency(invoice.amount, invoice.currency)}
                    </p>
                    <p className={styles.invoiceDate}>
                      {invoice.invoice_date ? formatDate(invoice.invoice_date) : '—'}
                    </p>
                  </div>
                </div>
                <div className={styles.invoiceActions}>
                  <StatusPill
                    status={invoice.status === 'paid' ? 'live' : invoice.status === 'open' ? 'pending' : 'revoked'}
                    label={invoice.status}
                  />
                  {invoice.invoice_pdf || invoice.hosted_invoice_url ? (
                    <FrameButton
                      size="sm"
                      onClick={() => window.open(invoice.invoice_pdf || invoice.hosted_invoice_url, '_blank')}
                      iconLeft={<Download style={{ width: 12, height: 12 }} />}
                    >
                      Download
                    </FrameButton>
                  ) : invoice.status === 'paid' ? (
                    <span className={styles.processingText}>Processing...</span>
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
