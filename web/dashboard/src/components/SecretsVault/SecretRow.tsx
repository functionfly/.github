/**
 * SecretRow - Compact row display for secrets in tables/lists
 *
 * Displays secret metadata in a compact row format suitable for tables
 * and lists. Includes checkbox for bulk selection, quick actions, and
 * expandable details section with status indicators.
 *
 * @example
 * ```tsx
 * // Basic usage in a table
 * <TableRow>
 *   <SecretRow
 *     secret={secretMetadata}
 *     onEdit={(id) => openEditModal(id)}
 *     onDelete={(id) => confirmDelete(id)}
 *   />
 * </TableRow>
 *
 * // With selection enabled
 * <SecretRow
 *   secret={secretMetadata}
 *   isSelected={selectedIds.has(secret.id)}
 *   onSelect={(id, checked) => toggleSelection(id, checked)}
 * />
 *
 * // With expanded details
 * <SecretRow
 *   secret={secretMetadata}
 *   defaultExpanded={true}
 * />
 *
 * // Loading state
 * <SecretRow isLoading />
 * ```
 */

import { useState, useCallback } from "react";
import {
  Key,
  Shield,
  Lock,
  FileKey,
  Copy,
  Check,
  Eye,
  EyeOff,
  Edit3,
  Trash2,
  Clock,
  ChevronDown,
  ChevronRight,
  ShieldCheck,
  ShieldAlert,
  ShieldQuestion,
  RefreshCw,
  MoreHorizontal,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";
import type { SecretMetadata, SecretType } from "@/types/vault";

import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

/** Secret status for visual indicators */
export type SecretRowStatus = "active" | "expired" | "rotating" | "compromised";

/** Extended secret metadata with optional status for row display */
export interface SecretRowData extends SecretMetadata {
  status?: SecretRowStatus;
  expires_at?: string;
}

export interface SecretRowProps {
  /** Secret metadata to display */
  secret?: SecretRowData;
  /** Whether the component is in loading state */
  isLoading?: boolean;
  /** Whether the row is selected (for bulk operations) */
  isSelected?: boolean;
  /** Whether the row is disabled */
  isDisabled?: boolean;
  /** Whether to show the checkbox for selection */
  showCheckbox?: boolean;
  /** Whether details are expanded by default */
  defaultExpanded?: boolean;
  /** Whether the row is expandable */
  isExpandable?: boolean;
  /** Decrypted secret value (shown only when explicitly provided) */
  decryptedValue?: string | null;
  /** Callback when selection changes */
  onSelect?: (secretId: string, selected: boolean) => void;
  /** Callback when edit is clicked */
  onEdit?: (secretId: string) => void;
  /** Callback when delete is clicked */
  onDelete?: (secretId: string) => void;
  /** Callback when copy is clicked */
  onCopy?: (value: string) => void;
  /** Callback when reveal is requested */
  onReveal?: (secretId: string) => void;
  /** Callback when rotate is clicked */
  onRotate?: (secretId: string) => void;
  /** Additional CSS classes */
  className?: string;
}

// Secret type icon mapping
const secretTypeIcons: Record<SecretType, typeof Key> = {
  api_key: Key,
  oauth_token: Shield,
  password: Lock,
  certificate: FileKey,
};

// Secret type label mapping
const secretTypeLabels: Record<SecretType, string> = {
  api_key: "API Key",
  oauth_token: "OAuth Token",
  password: "Password",
  certificate: "Certificate",
};

// Secret type badge variant mapping
const secretTypeVariants: Record<SecretType, "default" | "secondary" | "outline" | "success" | "warning"> = {
  api_key: "default",
  oauth_token: "success",
  password: "warning",
  certificate: "outline",
};

// Status configuration
const statusConfig: Record<SecretRowStatus, { icon: typeof ShieldCheck; color: string; label: string }> = {
  active: { icon: ShieldCheck, color: "text-success", label: "Active" },
  expired: { icon: ShieldAlert, color: "text-error", label: "Expired" },
  rotating: { icon: RefreshCw, color: "text-warning", label: "Rotating" },
  compromised: { icon: ShieldAlert, color: "text-error", label: "Compromised" },
};

/**
 * Masks a secret value, showing only the last 4 characters
 * @param value - The secret value to mask
 * @returns Masked string with only last 4 chars visible
 */
function maskSecretValue(value: string): string {
  if (value.length <= 4) return "•".repeat(value.length);
  return "•".repeat(Math.min(value.length - 4, 12)) + value.slice(-4);
}

/**
 * Skeleton loader for the secret row
 */
function SecretRowSkeleton({ showCheckbox = true }: { showCheckbox?: boolean }) {
  return (
    <div className="flex items-center gap-4 p-4 border-b border-(--border-subtle)">
      {showCheckbox && <Skeleton className="h-4 w-4 rounded-sm" />}
      <Skeleton className="h-8 w-8 rounded-lg" />
      <div className="flex-1 min-w-0 space-y-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-3 w-48" />
      </div>
      <Skeleton className="h-5 w-16 rounded-full" />
      <Skeleton className="h-5 w-20 rounded-full" />
      <Skeleton className="h-3 w-24" />
      <div className="flex gap-1">
        <Skeleton className="h-8 w-8 rounded-md" />
        <Skeleton className="h-8 w-8 rounded-md" />
      </div>
    </div>
  );
}

/**
 * SecretRow component
 *
 * Renders a compact row for table/list display with selection,
 * quick actions, and expandable details.
 */
export function SecretRow({
  secret,
  isLoading = false,
  isSelected = false,
  isDisabled = false,
  showCheckbox = true,
  defaultExpanded = false,
  isExpandable = true,
  decryptedValue,
  onSelect,
  onEdit,
  onDelete,
  onCopy,
  onReveal,
  onRotate,
  className,
}: SecretRowProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const [copied, setCopied] = useState(false);
  const [showDecrypted, setShowDecrypted] = useState(false);

  const handleSelect = useCallback(
    (checked: boolean) => {
      if (secret?.id && onSelect) {
        onSelect(secret.id, checked);
      }
    },
    [secret?.id, onSelect]
  );

  const handleCopy = useCallback(() => {
    const valueToCopy = decryptedValue || secret?.id || "";
    if (onCopy) {
      onCopy(valueToCopy);
    } else {
      navigator.clipboard.writeText(valueToCopy);
    }
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [decryptedValue, secret?.id, onCopy]);

  const handleReveal = useCallback(() => {
    if (decryptedValue) {
      setShowDecrypted(!showDecrypted);
    } else if (onReveal && secret?.id) {
      onReveal(secret.id);
    }
  }, [decryptedValue, onReveal, secret?.id, showDecrypted]);

  const toggleExpand = useCallback(() => {
    if (isExpandable) {
      setIsExpanded(!isExpanded);
    }
  }, [isExpandable, isExpanded]);

  if (isLoading) {
    return <SecretRowSkeleton showCheckbox={showCheckbox} />;
  }

  if (!secret) {
    return null;
  }

  const TypeIcon = secretTypeIcons[secret.secret_type];
  const badgeVariant = secretTypeVariants[secret.secret_type];
  const status: SecretRowStatus = secret.status || "active";
  const statusInfo = statusConfig[status];
  const StatusIcon = statusInfo.icon;

  const displayValue = showDecrypted && decryptedValue
    ? decryptedValue
    : maskSecretValue(decryptedValue || "••••••••••••••••");

  return (
    <TooltipProvider>
      <div
        className={cn(
          "group border-b border-(--border-subtle)",
          "transition-colors duration-200",
          isSelected && "bg-(--color-brand-500)/5",
          !isDisabled && "hover:bg-(--color-bg-secondary)",
          isDisabled && "opacity-50 cursor-not-allowed",
          className
        )}
      >
        {/* Main Row */}
        <div className="flex items-center gap-3 p-3">
          {/* Checkbox */}
          {showCheckbox && (
            <Checkbox
              checked={isSelected}
              onCheckedChange={handleSelect}
              disabled={isDisabled}
              className="shrink-0"
            />
          )}

          {/* Expand Toggle */}
          {isExpandable && (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 shrink-0 p-0"
              onClick={toggleExpand}
              aria-label={isExpanded ? 'Collapse secret details' : 'Expand secret details'}
              aria-expanded={isExpanded}
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4 text-(--color-text-muted)" />
              ) : (
                <ChevronRight className="h-4 w-4 text-(--color-text-muted)" />
              )}
            </Button>
          )}

          {/* Type Icon */}
          <div
            className={cn(
              "flex h-8 w-8 shrink-0 items-center justify-center rounded-lg",
              "bg-gradient-to-br from-(--color-brand-500) to-purple-500"
            )}
          >
            <TypeIcon className="h-4 w-4 text-white" />
          </div>

          {/* Secret Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-medium text-(--color-text-primary) truncate">
                {secret.name}
              </span>
              <Badge variant={badgeVariant} className="text-xs shrink-0">
                {secretTypeLabels[secret.secret_type]}
              </Badge>
            </div>
            {secret.description && (
              <p className="text-xs text-(--color-text-muted) truncate">
                {secret.description}
              </p>
            )}
          </div>

          {/* Status */}
          <div className="flex items-center gap-1.5 shrink-0">
            <StatusIcon className={cn("h-4 w-4", statusInfo.color)} />
            <span className={cn("text-xs font-medium", statusInfo.color)}>
              {statusInfo.label}
            </span>
          </div>

          {/* Scope Badges */}
          <div className="hidden md:flex items-center gap-1 shrink-0">
            {secret.scopes?.slice(0, 2).map((scope) => (
              <Badge key={scope} variant="secondary" className="text-xs">
                {scope}
              </Badge>
            ))}
            {secret.scopes && secret.scopes.length > 2 && (
              <Badge variant="secondary" className="text-xs">
                +{secret.scopes.length - 2}
              </Badge>
            )}
          </div>

          {/* Last Accessed */}
          <div className="hidden lg:flex items-center gap-1.5 text-xs text-(--color-text-muted) shrink-0 w-32">
            <Clock className="h-3 w-3" />
            {secret.last_accessed_at ? (
              <span>
                {formatDistanceToNow(new Date(secret.last_accessed_at), { addSuffix: true })}
              </span>
            ) : (
              <span>Never</span>
            )}
          </div>

          {/* Quick Actions */}
          <div className="flex items-center gap-1 shrink-0">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={handleReveal}
                >
                  {showDecrypted ? (
                    <EyeOff className="h-4 w-4 text-(--color-text-muted)" />
                  ) : (
                    <Eye className="h-4 w-4 text-(--color-text-muted)" />
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{showDecrypted ? "Hide" : "Reveal"}</p>
              </TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={handleCopy}
                >
                  {copied ? (
                    <Check className="h-4 w-4 text-success" />
                  ) : (
                    <Copy className="h-4 w-4 text-(--color-text-muted)" />
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{copied ? "Copied!" : "Copy"}</p>
              </TooltipContent>
            </Tooltip>

            {/* More Actions Dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8" aria-label="More actions">
                  <MoreHorizontal className="h-4 w-4 text-(--color-text-muted)" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {onEdit && (
                  <DropdownMenuItem onClick={() => onEdit(secret.id)}>
                    <Edit3 className="h-4 w-4 mr-2" />
                    Edit
                  </DropdownMenuItem>
                )}
                {onRotate && (
                  <DropdownMenuItem onClick={() => onRotate(secret.id)}>
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Rotate
                  </DropdownMenuItem>
                )}
                {onDelete && (
                  <DropdownMenuItem
                    onClick={() => onDelete(secret.id)}
                    className="text-error focus:text-error"
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Expanded Details */}
        {isExpanded && (
          <div className="px-3 pb-3 pt-0">
            <div
              className={cn(
                "ml-11 p-3 rounded-lg",
                "bg-(--color-bg-secondary)",
                "border border-(--border-subtle)"
              )}
            >
              {/* Value Display */}
              <div className="flex items-center gap-2 mb-3">
                <code className="flex-1 font-mono text-sm text-(--color-text-primary) truncate bg-(--color-bg-primary) px-2 py-1 rounded">
                  {displayValue}
                </code>
              </div>

              {/* Metadata Grid */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs">
                <div>
                  <span className="text-(--color-text-muted)">ID:</span>
                  <span className="ml-1 text-(--color-text-secondary) font-mono">
                    {secret.id.slice(0, 8)}...
                  </span>
                </div>
                <div>
                  <span className="text-(--color-text-muted)">Access Count:</span>
                  <span className="ml-1 text-(--color-text-secondary)">
                    {secret.access_count}
                  </span>
                </div>
                <div>
                  <span className="text-(--color-text-muted)">Created:</span>
                  <span className="ml-1 text-(--color-text-secondary)">
                    {formatDistanceToNow(new Date(secret.created_at), { addSuffix: true })}
                  </span>
                </div>
                <div>
                  <span className="text-(--color-text-muted)">Updated:</span>
                  <span className="ml-1 text-(--color-text-secondary)">
                    {formatDistanceToNow(new Date(secret.updated_at), { addSuffix: true })}
                  </span>
                </div>
              </div>

              {/* All Scopes */}
              {secret.scopes && secret.scopes.length > 0 && (
                <div className="mt-3 pt-3 border-t border-(--border-subtle)">
                  <span className="text-xs text-(--color-text-muted)">Scopes:</span>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {secret.scopes.map((scope) => (
                      <Badge key={scope} variant="secondary" className="text-xs">
                        {scope}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </TooltipProvider>
  );
}
