/**
 * FunctionHeader Component
 *
 * A comprehensive page header component for function detail pages in the FunctionFly dashboard.
 * Displays execution hash, trust tier, economic score, runtime badge, resource signature,
 * and certificate verification status.
 *
 * @example
 * <FunctionHeader
 *   data={{
 *     name: "hash-sha256",
 *     id: "func_123",
 *     executionRootHash: "0x7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa5d0...",
 *     trustTier: "high",
 *     economicScore: 87,
 *     runtime: "workers",
 *     resourceSignature: "res_sig_a1b2c3d4...",
 *     fxcert: { verified: true, issuedAt: "2024-01-15T10:30:00Z" },
 *     status: "online",
 *     version: "v2.1.0"
 *   }}
 *   onBack={() => navigate(-1)}
 *   onEdit={() => openEditor()}
 * />
 */

import * as React from 'react';
import { Link } from 'react-router-dom';
import {
  ArrowLeft,
  Shield,
  ShieldCheck,
  Hash,
  Database,
  TrendingUp,
  Award,
  FileCode2,
  Copy,
  Check,
  MoreVertical,
  Play,
  Pause,
  Trash2,
  AlertCircle,
} from 'lucide-react';
import { cn, truncate } from '@/lib/utils';
import type { FunctionHeaderData, FunctionHeaderProps, TrustTier } from '@/types';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusBadge } from '@/components/common/StatusBadge';
import { ProviderIcon } from '@/components/common/ProviderIcon';

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Get trust tier configuration for styling
 */
function getTrustTierConfig(tier: TrustTier) {
  const configs = {
    critical: {
      color: 'text-purple-400',
      bgColor: 'bg-purple-500/10',
      borderColor: 'border-purple-500/30',
      iconColor: '#a855f7',
      label: 'Critical Trust',
      description: 'Maximum security and verification',
    },
    high: {
      color: 'text-emerald-400',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/30',
      iconColor: '#10b981',
      label: 'High Trust',
      description: 'Verified and highly trusted',
    },
    medium: {
      color: 'text-brand-400',
      bgColor: 'bg-brand-500/10',
      borderColor: 'border-brand-500/30',
      iconColor: '#6366f1',
      label: 'Medium Trust',
      description: 'Standard verification level',
    },
    low: {
      color: 'text-amber-400',
      bgColor: 'bg-amber-500/10',
      borderColor: 'border-amber-500/30',
      iconColor: '#f59e0b',
      label: 'Low Trust',
      description: 'Basic verification only',
    },
    untrusted: {
      color: 'text-red-400',
      bgColor: 'bg-red-500/10',
      borderColor: 'border-red-500/30',
      iconColor: '#ef4444',
      label: 'Untrusted',
      description: 'Not verified - use with caution',
    },
  };

  return configs[tier] || configs.medium;
}

/**
 * Get economic score color based on value
 */
function getEconomicScoreColor(score: number): string {
  if (score >= 80) return 'text-emerald-400';
  if (score >= 60) return 'text-brand-400';
  if (score >= 40) return 'text-amber-400';
  return 'text-red-400';
}

/**
 * Get economic score background color
 */
function getEconomicScoreBgColor(score: number): string {
  if (score >= 80) return 'bg-emerald-500/10';
  if (score >= 60) return 'bg-brand-500/10';
  if (score >= 40) return 'bg-amber-500/10';
  return 'bg-red-500/10';
}

// ============================================================================
// Sub-components
// ============================================================================

/**
 * Copy to clipboard button with tooltip
 */
function CopyButton({ value, label }: { value: string; label: string }) {
  const [copied, setCopied] = React.useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6 text-text-muted hover:text-text-primary"
            onClick={handleCopy}
            aria-label={`Copy ${label}`}
          >
            {copied ? (
              <Check className="h-3.5 w-3.5 text-emerald-400" />
            ) : (
              <Copy className="h-3.5 w-3.5" />
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>{copied ? 'Copied!' : `Copy ${label}`}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Trust Tier Badge Component
 */
function TrustTierBadge({ tier, className }: { tier: TrustTier; className?: string }) {
  const config = getTrustTierConfig(tier);

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={cn(
              'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full border',
              config.bgColor,
              config.borderColor,
              className
            )}
            aria-label={`Trust tier: ${config.label}`}
          >
            <Shield className={cn('h-3.5 w-3.5', config.color)} />
            <span className={cn('text-xs font-medium', config.color)}>{config.label}</span>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>{config.description}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Economic Score Badge Component
 */
function EconomicScoreBadge({ score, className }: { score: number; className?: string }) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={cn(
              'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-border-subtle',
              getEconomicScoreBgColor(score),
              className
            )}
            aria-label={`Economic score: ${score}%`}
          >
            <TrendingUp className={cn('h-3.5 w-3.5', getEconomicScoreColor(score))} />
            <span className={cn('text-xs font-medium', getEconomicScoreColor(score))}>
              {score}%
            </span>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>Economic Efficiency Score</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Verified Certificate Badge (fxcert) – clickable to open Certificates tab when link is provided.
 */
function FxcertBadge({
  fxcert,
  to,
  className,
}: {
  fxcert: FunctionHeaderData['fxcert'];
  /** Optional route to certificates (e.g. /registry/author/name/executions?tab=certificates) */
  to?: string;
  className?: string;
}) {
  if (!fxcert.verified) {
    return (
      <Badge
        variant="outline"
        className={cn('gap-1.5 text-text-muted', className)}
        aria-label="Certificate not verified"
      >
        <Shield className="h-3.5 w-3.5" />
        <span>Unverified</span>
      </Badge>
    );
  }

  const badge = (
    <Badge
      variant="success"
      className={cn('gap-1.5', to && 'cursor-pointer hover:opacity-90', className)}
      aria-label="Verified certificate"
    >
      <ShieldCheck className="h-3.5 w-3.5" />
      <span>fxcert</span>
    </Badge>
  );

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          {to ? (
            <Link to={to} className="inline-flex">
              {badge}
            </Link>
          ) : (
            badge
          )}
        </TooltipTrigger>
        <TooltipContent className="max-w-xs">
          <div className="space-y-1">
            <p className="font-medium">Verified Certificate</p>
            {to && <p className="text-xs text-text-muted">Click to view certificates</p>}
            {fxcert.issuer && <p className="text-xs text-text-muted">Issuer: {fxcert.issuer}</p>}
            {fxcert.issuedAt && (
              <p className="text-xs text-text-muted">
                Issued: {new Date(fxcert.issuedAt).toLocaleDateString()}
              </p>
            )}
            {fxcert.expiresAt && (
              <p className="text-xs text-text-muted">
                Expires: {new Date(fxcert.expiresAt).toLocaleDateString()}
              </p>
            )}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Runtime Badge Component
 */
function RuntimeBadge({ runtime, className }: { runtime: string; className?: string }) {
  const runtimeLabels: Record<string, string> = {
    workers: 'Cloudflare Workers',
    vercel: 'Vercel',
    fly: 'Fly.io',
    deno: 'Deno Deploy',
    'functionfly-edge': 'FunctionFly Edge',
  };

  const label = runtimeLabels[runtime] || runtime;

  return (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-border-subtle bg-bg-tertiary',
        className
      )}
      aria-label={`Runtime: ${label}`}
    >
      <ProviderIcon provider={runtime} size="sm" />
      <span className="text-xs font-medium text-text-primary">{label}</span>
    </div>
  );
}

/**
 * Report Issue Badge Component
 * A sleek warning badge that lets users report if a function is down or doesn't work.
 */
function ReportIssueBadge({
  functionName,
  author,
  onClick,
}: {
  functionName: string;
  author: string;
  onClick: () => void;
}) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            onClick={onClick}
            className={cn(
              'p-1.5 rounded-md border border-transparent',
              'text-text-muted hover:text-amber-500 hover:border-amber-500/20 hover:bg-amber-500/5',
              'transition-all cursor-pointer'
            )}
            aria-label="Report an issue with this function"
          >
            <AlertCircle className="h-4 w-4" />
          </button>
        </TooltipTrigger>
        <TooltipContent className="max-w-xs">
          <div className="space-y-1">
            <p className="font-medium">Report a Function Issue</p>
            <p className="text-xs text-text-muted">
              Notify @{author} that {functionName} is down or not working.
            </p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Hash Display Component
 */
function HashDisplay({
  value,
  label,
  truncatedLength = 16,
  className,
}: {
  value: string;
  label: string;
  truncatedLength?: number;
  className?: string;
}) {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Hash className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
      <div className="flex items-center gap-1.5 min-w-0">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <code className="text-xs font-mono text-text-secondary truncate">
                {truncate(value, truncatedLength)}
              </code>
            </TooltipTrigger>
            <TooltipContent>
              <code className="text-xs font-mono">{value}</code>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <CopyButton value={value} label={label} />
      </div>
    </div>
  );
}

/**
 * Resource Signature Display Component
 */
function ResourceSignatureDisplay({
  signature,
  className,
}: {
  signature: string;
  className?: string;
}) {
  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Database className="h-3.5 w-3.5 text-text-muted flex-shrink-0" />
      <div className="flex items-center gap-1.5 min-w-0">
        <span className="text-xs text-text-muted">Resource:</span>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <code className="text-xs font-mono text-text-secondary truncate">
                {truncate(signature, 12)}
              </code>
            </TooltipTrigger>
            <TooltipContent>
              <code className="text-xs font-mono">{signature}</code>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <CopyButton value={signature} label="resource signature" />
      </div>
    </div>
  );
}

// ============================================================================
// Main Component
// ============================================================================

/**
 * FunctionHeader Component
 *
 * A comprehensive page header for function detail pages.
 * Displays all critical function metadata in a two-row layout:
 * - Row 1: Title section with name, status, version, and actions
 * - Row 2: Badges and metrics row (trust tier, economic score, runtime, hashes)
 */
export function FunctionHeader({
  data,
  className,
  onBack,
  onEdit,
  onDeploy,
  onTest,
  onShare,
  onReportIssue,
}: FunctionHeaderProps) {
  return (
    <div
      className={cn(
        'w-full space-y-4 p-6 rounded-xl border border-border-subtle bg-bg-tertiary/50',
        className
      )}
      role="banner"
      aria-label={`Function header: ${data.name}`}
    >
      {/* Row 1: Title and Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-start gap-3">
          {onBack && (
            <Button
              variant="ghost"
              size="icon"
              onClick={onBack}
              className="text-text-secondary hover:text-text-primary mt-0.5"
              aria-label="Go back"
            >
              <ArrowLeft className="h-4 w-4" />
            </Button>
          )}
          <div className="space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-2xl font-bold text-text-primary">{data.name}</h1>
              {data.status && <StatusBadge status={data.status} />}
              {data.version && (
                <Badge variant="secondary" className="text-xs">
                  <FileCode2 className="h-3 w-3 mr-1" />
                  {data.version}
                </Badge>
              )}
            </div>
            {data.description && (
              <p className="text-sm text-text-secondary max-w-2xl">{data.description}</p>
            )}
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2 flex-wrap">
          {onTest && (
            <Button variant="outline" size="sm" onClick={onTest}>
              <Play className="h-4 w-4 mr-1.5" />
              Test
            </Button>
          )}
          {onEdit && (
            <Button variant="outline" size="sm" onClick={onEdit}>
              Edit
            </Button>
          )}
          {onDeploy && (
            <Button size="sm" onClick={onDeploy}>
              Deploy
            </Button>
          )}

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="h-8 w-8"
                aria-label="Function options"
              >
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="bg-bg-tertiary border-white/8">
              {onShare && <DropdownMenuItem onClick={onShare}>Share Function</DropdownMenuItem>}
              <DropdownMenuSeparator />
              <DropdownMenuItem className="text-red-400">
                <Trash2 className="h-4 w-4 mr-2" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Divider */}
      <div className="h-px bg-border-subtle" />

      {/* Row 2: Badges and Metrics */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        {/* Left: Trust, Economic Score, Runtime */}
        <div className="flex flex-wrap items-center gap-2">
          <TrustTierBadge tier={data.trustTier} />
          <EconomicScoreBadge score={data.economicScore} />
          <RuntimeBadge runtime={data.runtime} />
          <FxcertBadge
            fxcert={data.fxcert}
            to={
              data.id && data.id.includes('/')
                ? `/registry/${data.id.split('/').slice(0, 2).join('/')}/executions?tab=certificates`
                : undefined
            }
          />
        </div>

        {/* Right: Hashes and Signatures */}
        <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-4 text-left sm:text-right">
          <HashDisplay
            value={data.executionRootHash}
            label="execution root hash"
            truncatedLength={20}
          />
          <ResourceSignatureDisplay signature={data.resourceSignature} />
          {onReportIssue && (
            <ReportIssueBadge
              functionName={data.name}
              author={data.id.split('/')[0] ?? ''}
              onClick={onReportIssue}
            />
          )}
        </div>
      </div>
    </div>
  );
}

export default FunctionHeader;
