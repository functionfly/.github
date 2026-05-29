import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/lib/utils';
import { ExternalLink } from 'lucide-react';
import { Link } from 'react-router-dom';
import { KIND_META } from '../constants';
import type { UnifiedPurchase } from '../utils';
import { formatDate, formatRelativeTime, formatUsd } from '../utils';
import { StatusLed } from './PurchaseCards';

type TFn = (key: string, opts?: Record<string, unknown>) => string;

function purchaseLink(entry: UnifiedPurchase): string | null {
  if (entry.functionRow) {
    const r = entry.functionRow;
    return `/functions/${encodeURIComponent(r.functionAuthor)}/${encodeURIComponent(r.functionName)}`;
  }
  if (entry.agentRow) {
    return `/agents/${encodeURIComponent(entry.agentRow.agentId)}`;
  }
  if (entry.licenseRow) {
    const r = entry.licenseRow;
    return `/functions/discovery?q=${encodeURIComponent(r.functionName || r.functionId)}`;
  }
  return null;
}

export function PurchaseTable({ entries, t }: { entries: UnifiedPurchase[]; t: TFn }) {
  if (entries.length === 0) {
    return <p className="text-sm text-aviation-text-secondary">{t('purchasesPage.noResults')}</p>;
  }

  return (
    <div className="overflow-hidden rounded-xl border border-aviation-border-instrument">
      <Table>
        <TableHeader>
          <TableRow className="border-aviation-border-instrument hover:bg-transparent">
            <TableHead className="text-aviation-text-secondary">{t('purchasesPage.colType')}</TableHead>
            <TableHead className="text-aviation-text-secondary">{t('purchasesPage.colName')}</TableHead>
            <TableHead className="text-aviation-text-secondary">{t('purchasesPage.colStatus')}</TableHead>
            <TableHead className="text-aviation-text-secondary">{t('purchasesPage.colAmount')}</TableHead>
            <TableHead className="text-aviation-text-secondary">{t('purchasesPage.colDate')}</TableHead>
            <TableHead className="w-[80px]" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {entries.map((entry) => {
            const meta = KIND_META[entry.kind];
            const Icon = meta.icon;
            const href = purchaseLink(entry);
            return (
              <TableRow key={entry.id} className="border-aviation-border-instrument/60">
                <TableCell>
                  <div className="flex items-center gap-2">
                    <div className={cn('rounded-md p-1.5', meta.iconBg)}>
                      <Icon className="h-3.5 w-3.5" />
                    </div>
                    <span className="text-xs capitalize text-aviation-text-secondary">{entry.kind}</span>
                  </div>
                </TableCell>
                <TableCell>
                  <div className="max-w-[240px]">
                    <p className="truncate font-medium text-aviation-text-primary">{entry.title}</p>
                    <p className="truncate text-xs text-aviation-text-dim">{entry.subtitle}</p>
                  </div>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <StatusLed status={entry.status} revoked={entry.licenseRow?.revoked} />
                    <Badge variant="outline" className="text-[10px] capitalize">
                      {entry.licenseRow?.revoked ? t('purchasesPage.revoked') : entry.status}
                    </Badge>
                  </div>
                </TableCell>
                <TableCell className="font-mono text-sm">
                  {entry.amount != null ? formatUsd(entry.amount) : '—'}
                </TableCell>
                <TableCell>
                  <span className="text-sm" title={formatDate(entry.dateMs)}>
                    {formatRelativeTime(entry.dateMs, t)}
                  </span>
                </TableCell>
                <TableCell>
                  {href && (
                    <Button asChild variant="ghost" size="sm" className="h-8 w-8 p-0">
                      <Link to={href} aria-label={t('purchasesPage.openItem')}>
                        <ExternalLink className="h-4 w-4" />
                      </Link>
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
