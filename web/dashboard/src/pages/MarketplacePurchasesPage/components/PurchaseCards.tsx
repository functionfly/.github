import { createBillingPortalSession, getBillingPortalErrorMessage } from '@/api/billing';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import { Bot, Copy, CreditCard, ExternalLink, Key } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import { KIND_META } from '../constants';
import type { UnifiedPurchase } from '../utils';
import {
  formatDate,
  formatRelativeTime,
  formatUsd,
  isActiveStatus,
  isRevokedStatus,
} from '../utils';

type TFn = (key: string, opts?: Record<string, unknown>) => string;

export function StatusLed({ status, revoked }: { status: string; revoked?: boolean }) {
  const active = isActiveStatus(status) && !revoked;
  const bad = isRevokedStatus(status, revoked);
  return (
    <span
      className={cn(
        'inline-block h-2 w-2 shrink-0 rounded-full',
        active && 'animate-pulse bg-aviation-green shadow-[0_0_6px_rgba(34,197,94,0.6)]',
        bad && 'bg-destructive',
        !active && !bad && 'bg-aviation-amber/80'
      )}
      aria-hidden
    />
  );
}

function PurchaseCardShell({
  entry,
  header,
  body,
  index = 0,
}: {
  entry: UnifiedPurchase;
  header: React.ReactNode;
  body?: React.ReactNode;
  index?: number;
}) {
  const meta = KIND_META[entry.kind];
  const Icon = meta.icon;

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: Math.min(index * 0.04, 0.24), duration: 0.25 }}
    >
      <Card
        className={cn(
          'border-l-4 transition-all duration-200 hover:border-aviation-border-panel hover:shadow-[0_0_20px_rgba(6,182,212,0.08)]',
          meta.accent
        )}
      >
        <CardHeader className="pb-2">
          <div className="flex min-w-0 items-start gap-3">
            <div className={cn('mt-0.5 shrink-0 rounded-lg p-2', meta.iconBg)}>
              <Icon className="h-4 w-4" />
            </div>
            <div className="min-w-0 flex-1">{header}</div>
          </div>
        </CardHeader>
        {body ? <CardContent className="pt-0">{body}</CardContent> : null}
      </Card>
    </motion.div>
  );
}

export function UnifiedPurchaseCard({
  entry,
  t,
  index = 0,
}: {
  entry: UnifiedPurchase;
  t: TFn;
  index?: number;
}) {
  switch (entry.kind) {
    case 'function':
      return entry.functionRow ? (
        <FunctionPurchaseCard row={entry.functionRow} t={t} index={index} />
      ) : null;
    case 'agent':
      return entry.agentRow ? <AgentHiringCard row={entry.agentRow} t={t} index={index} /> : null;
    case 'license':
      return entry.licenseRow ? <LicenseCard row={entry.licenseRow} t={t} index={index} /> : null;
    case 'subscription':
      return entry.subscriptionRow ? (
        <SubscriptionCard row={entry.subscriptionRow} t={t} index={index} />
      ) : null;
    default:
      return null;
  }
}

function FunctionPurchaseCard({
  row,
  t,
  index,
}: {
  row: NonNullable<UnifiedPurchase['functionRow']>;
  t: TFn;
  index?: number;
}) {
  const fnPath = `/functions/${encodeURIComponent(row.functionAuthor)}/${encodeURIComponent(row.functionName)}`;
  const entry: UnifiedPurchase = {
    id: row.id,
    kind: 'function',
    title: `${row.functionAuthor}/${row.functionName}`,
    subtitle: row.agentId,
    dateMs: row.createdAt,
    amount: row.pricePaidUsd,
    status: row.status,
    functionRow: row,
  };

  return (
    <PurchaseCardShell
      entry={entry}
      index={index}
      header={
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <CardTitle className="text-base font-mono">{entry.title}</CardTitle>
            <CardDescription className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1">
              <StatusLed status={row.status} />
              <span>
                {t('purchasesPage.agent')}: {row.agentId}
              </span>
              <span>·</span>
              <span title={formatDate(row.createdAt)}>{formatRelativeTime(row.createdAt, t)}</span>
            </CardDescription>
          </div>
          <div className="flex flex-col items-end gap-1">
            <Badge variant="secondary">{formatUsd(row.pricePaidUsd)}</Badge>
            <Badge variant={isActiveStatus(row.status) ? 'default' : 'outline'} className="text-[10px]">
              {row.status}
            </Badge>
          </div>
        </div>
      }
      body={
        <Button asChild variant="outline" size="sm">
          <Link to={fnPath}>
            <ExternalLink className="mr-2 h-3.5 w-3.5" />
            {t('purchasesPage.openFunction')}
          </Link>
        </Button>
      }
    />
  );
}

function AgentHiringCard({
  row,
  t,
  index,
}: {
  row: NonNullable<UnifiedPurchase['agentRow']>;
  t: TFn;
  index?: number;
}) {
  const entry: UnifiedPurchase = {
    id: row.id,
    kind: 'agent',
    title: row.taskType,
    subtitle: row.agentId,
    dateMs: row.createdAt,
    amount: row.budgetUsd,
    status: row.status,
    agentRow: row,
  };

  return (
    <PurchaseCardShell
      entry={entry}
      index={index}
      header={
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <CardTitle className="text-base">{row.taskType}</CardTitle>
            <CardDescription className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1">
              <StatusLed status={row.status} />
              <span>
                {t('purchasesPage.agent')}: {row.agentId}
              </span>
              <span>·</span>
              <span title={formatDate(row.createdAt)}>{formatRelativeTime(row.createdAt, t)}</span>
            </CardDescription>
          </div>
          <Badge variant={isActiveStatus(row.status) ? 'default' : 'secondary'}>{row.status}</Badge>
        </div>
      }
      body={
        <div className="flex items-center justify-between gap-2">
          <span className="text-sm text-aviation-text-secondary">
            {t('purchasesPage.budget')}: {formatUsd(row.budgetUsd)}
          </span>
          <Button asChild variant="outline" size="sm">
            <Link to={`/agents/${encodeURIComponent(row.agentId)}`}>
              <Bot className="mr-2 h-3.5 w-3.5" />
              {t('purchasesPage.openAgent')}
            </Link>
          </Button>
        </div>
      }
    />
  );
}

function LicenseCard({
  row,
  t,
  index,
}: {
  row: NonNullable<UnifiedPurchase['licenseRow']>;
  t: TFn;
  index?: number;
}) {
  const [copying, setCopying] = useState(false);
  const fnPath = row.functionId
    ? `/functions/discovery?q=${encodeURIComponent(row.functionName || row.functionId)}`
    : '/functions/discovery';
  const maxActivations = row.maxActivations ?? null;
  const activationPct =
    maxActivations && maxActivations > 0
      ? Math.min(100, (row.activationCount / maxActivations) * 100)
      : null;

  const copyPrefix = async () => {
    if (!row.keyPrefix) return;
    setCopying(true);
    try {
      await navigator.clipboard.writeText(row.keyPrefix);
      toast.success(t('purchasesPage.keyCopied'));
    } catch {
      toast.error(t('purchasesPage.keyCopyFailed'));
    } finally {
      setCopying(false);
    }
  };

  const entry: UnifiedPurchase = {
    id: row.id,
    kind: 'license',
    title: row.functionName || row.functionId,
    subtitle: row.purchaserName,
    dateMs: row.issuedAt,
    status: row.revoked ? 'revoked' : row.type,
    licenseRow: row,
  };

  return (
    <PurchaseCardShell
      entry={entry}
      index={index}
      header={
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <CardTitle className="text-base">{row.functionName || row.functionId}</CardTitle>
            <CardDescription className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1">
              <StatusLed status={row.type} revoked={row.revoked} />
              <Key className="inline h-3.5 w-3.5" />
              {row.keyPrefix ? `${row.keyPrefix}****` : t('purchasesPage.licenseMasked')}
              {row.expiresAt ? ` · ${t('purchasesPage.expires')} ${formatDate(row.expiresAt)}` : ''}
            </CardDescription>
          </div>
          <Badge variant={row.revoked ? 'destructive' : 'secondary'}>
            {row.revoked ? t('purchasesPage.revoked') : row.type}
          </Badge>
        </div>
      }
      body={
        <div className="space-y-3">
          {activationPct != null && (
            <div className="space-y-1">
              <div className="flex justify-between text-xs text-aviation-text-dim">
                <span>
                  {t('purchasesPage.activations', {
                    count: row.activationCount,
                    max: maxActivations ?? '∞',
                  })}
                </span>
                <span>{Math.round(activationPct)}%</span>
              </div>
              <Progress value={activationPct} indicatorClassName="bg-emerald-500" />
            </div>
          )}
          {!activationPct && (
            <p className="text-xs text-aviation-text-dim">
              {t('purchasesPage.activations', {
                count: row.activationCount,
                max: row.maxActivations ?? '∞',
              })}
            </p>
          )}
          <div className="flex flex-wrap gap-2">
            {row.keyPrefix && (
              <Button variant="outline" size="sm" onClick={copyPrefix} disabled={copying}>
                <Copy className="mr-2 h-3.5 w-3.5" />
                {t('purchasesPage.copyKeyPrefix')}
              </Button>
            )}
            <Button asChild variant="outline" size="sm">
              <Link to={fnPath}>
                <ExternalLink className="mr-2 h-3.5 w-3.5" />
                {t('purchasesPage.openFunction')}
              </Link>
            </Button>
          </div>
        </div>
      }
    />
  );
}

function SubscriptionCard({
  row,
  t,
  index,
}: {
  row: NonNullable<UnifiedPurchase['subscriptionRow']>;
  t: TFn;
  index?: number;
}) {
  const [billingLoading, setBillingLoading] = useState(false);

  const openBillingPortal = async () => {
    setBillingLoading(true);
    try {
      const { url } = await createBillingPortalSession(
        `${window.location.origin}/marketplace/purchases?tab=subscriptions`
      );
      if (url) window.location.href = url;
    } catch (e) {
      toast.error(getBillingPortalErrorMessage(e));
    } finally {
      setBillingLoading(false);
    }
  };

  const entry: UnifiedPurchase = {
    id: row.id,
    kind: 'subscription',
    title: row.planName,
    subtitle: row.billingCycle,
    dateMs: row.currentPeriodStart,
    amount: row.amount,
    status: row.status,
    subscriptionRow: row,
  };

  return (
    <PurchaseCardShell
      entry={entry}
      index={index}
      header={
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <CardTitle className="text-base">{row.planName}</CardTitle>
            <CardDescription className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1">
              <StatusLed status={row.status} />
              <CreditCard className="inline h-3.5 w-3.5" />
              {row.billingCycle}
              <span>·</span>
              <span>
                {formatDate(row.currentPeriodStart)} – {formatDate(row.currentPeriodEnd)}
              </span>
            </CardDescription>
          </div>
          <Badge variant={isActiveStatus(row.status) ? 'default' : 'secondary'}>{row.status}</Badge>
        </div>
      }
      body={
        <div className="space-y-2">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="text-sm font-medium">{formatUsd(row.amount)}</span>
            <span className="text-xs text-aviation-text-secondary">
              {t('purchasesPage.nextRenewal')}: {formatDate(row.currentPeriodEnd)}
            </span>
          </div>
          {row.cancelAtPeriodEnd && (
            <p className="text-xs text-amber-500">{t('purchasesPage.cancelAtPeriodEnd')}</p>
          )}
          <Button variant="outline" size="sm" onClick={openBillingPortal} disabled={billingLoading}>
            {t('purchasesPage.manageBilling')}
          </Button>
        </div>
      }
    />
  );
}

export function PurchaseSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h3 className="mb-3 text-sm font-medium uppercase tracking-wider text-aviation-text-secondary">
        {title}
      </h3>
      <div className="space-y-3">{children}</div>
    </div>
  );
}
