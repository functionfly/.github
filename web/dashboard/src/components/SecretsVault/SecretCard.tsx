/**
 * SecretCard - Card display for a single secret
 *
 * Displays secret metadata in a card format with masked value preview,
 * copy functionality, edit/delete actions, and scope badges.
 * Supports different secret types with appropriate icons and styling.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <SecretCard
 *   secret={secretMetadata}
 *   onEdit={(id) => openEditModal(id)}
 *   onDelete={(id) => confirmDelete(id)}
 * />
 *
 * // With decrypted value display
 * <SecretCard
 *   secret={secretMetadata}
 *   decryptedValue="sk-abc123..."
 *   onCopy={(value) => copyToClipboard(value)}
 * />
 *
 * // Loading state
 * <SecretCard isLoading />
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
  ShieldCheck,
  ShieldAlert,
  ShieldQuestion,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { cn } from "@/lib/utils";
import type { SecretMetadata, SecretType } from "@/types/vault";

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Secret status for visual indicators */
export type SecretStatus = "active" | "expired" | "rotating" | "compromised";

/** Extended secret metadata with optional status */
export interface SecretWithStatus extends SecretMetadata {
  status?: SecretStatus;
  expires_at?: string;
}

export interface SecretCardProps {
  /** Secret metadata to display */
  secret?: SecretWithStatus;
  /** Whether the component is in loading state */
  isLoading?: boolean;
  /** Decrypted secret value (shown only when explicitly provided) */
  decryptedValue?: string | null;
  /** Callback when edit is clicked */
  onEdit?: (secretId: string) => void;
  /** Callback when delete is clicked */
  onDelete?: (secretId: string) => void;
  /** Callback when copy is clicked */
  onCopy?: (value: string) => void;
  /** Callback when reveal is requested (triggers security gate) */
  onReveal?: (secretId: string) => void;
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

// Status icon mapping
const statusIcons: Record<SecretStatus, typeof ShieldCheck> = {
  active: ShieldCheck,
  expired: ShieldAlert,
  rotating: ShieldQuestion,
  compromised: ShieldAlert,
};

// Status color mapping
const statusColors: Record<SecretStatus, string> = {
  active: "text-success",
  expired: "text-error",
  rotating: "text-warning",
  compromised: "text-error",
};

// Status label mapping
const statusLabels: Record<SecretStatus, string> = {
  active: "Active",
  expired: "Expired",
  rotating: "Rotating",
  compromised: "Compromised",
};

/**
 * Masks a secret value, showing only the last 4 characters
 * @param value - The secret value to mask
 * @returns Masked string with only last 4 chars visible
 */
function maskSecretValue(value: string): string {
  if (value.length <= 4) return "•".repeat(value.length);
  return "•".repeat(Math.min(value.length - 4, 20)) + value.slice(-4);
}

/**
 * Skeleton loader for the secret card
 */
function SecretCardSkeleton({ className }: { className?: string }) {
  return (
    <Card className={cn("overflow-hidden", className)}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <Skeleton className="h-10 w-10 rounded-lg" />
            <div className="space-y-2">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-3 w-20" />
            </div>
          </div>
          <Skeleton className="h-6 w-16 rounded-full" />
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-10 w-full" />
        <div className="flex gap-2">
          <Skeleton className="h-5 w-16 rounded-full" />
          <Skeleton className="h-5 w-20 rounded-full" />
        </div>
        <div className="flex justify-between">
          <Skeleton className="h-3 w-24" />
          <div className="flex gap-2">
            <Skeleton className="h-8 w-8 rounded-md" />
            <Skeleton className="h-8 w-8 rounded-md" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * SecretCard component
 *
 * Renders a card displaying secret information with actions,
 * masked value preview, and metadata.
 */
export function SecretCard({
  secret,
  isLoading = false,
  decryptedValue,
  onEdit,
  onDelete,
  onCopy,
  onReveal,
  className,
}: SecretCardProps) {
  const [copied, setCopied] = useState(false);
  const [showDecrypted, setShowDecrypted] = useState(false);

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

  if (isLoading) {
    return <SecretCardSkeleton className={className} />;
  }

  if (!secret) {
    return null;
  }

  const TypeIcon = secretTypeIcons[secret.secret_type];
  const badgeVariant = secretTypeVariants[secret.secret_type];
  const status: SecretStatus = secret.status || "active";
  const StatusIcon = statusIcons[status];

  const displayValue = showDecrypted && decryptedValue
    ? decryptedValue
    : maskSecretValue(decryptedValue || "••••••••••••••••••••••••");

  return (
    <TooltipProvider>
      <Card
        className={cn(
          "overflow-hidden transition-all duration-200",
          "border border-(--border-subtle)",
          "bg-(--color-bg-primary)",
          "hover:border-(--border-default)",
          "hover:shadow-md",
          className
        )}
      >
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-3 min-w-0">
              <div
                className={cn(
                  "flex h-10 w-10 shrink-0 items-center justify-center rounded-lg",
                  "bg-gradient-to-br from-(--color-brand-500) to-purple-500"
                )}
              >
                <TypeIcon className="h-5 w-5 text-white" />
              </div>
              <div className="min-w-0">
                <CardTitle className="text-base font-semibold text-(--color-text-primary) truncate">
                  {secret.name}
                </CardTitle>
                <CardDescription className="text-xs text-(--color-text-muted)">
                  {secretTypeLabels[secret.secret_type]}
                </CardDescription>
              </div>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <StatusIcon className={cn("h-4 w-4", statusColors[status])} />
              <Badge variant={badgeVariant} className="text-xs">
                {statusLabels[status]}
              </Badge>
            </div>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Description */}
          {secret.description && (
            <p className="text-sm text-(--color-text-secondary) line-clamp-2">
              {secret.description}
            </p>
          )}

          {/* Masked Value Display */}
          <div
            className={cn(
              "flex items-center gap-2 p-3 rounded-lg",
              "bg-(--color-bg-secondary)",
              "border border-(--border-subtle)"
            )}
          >
            <code className="flex-1 font-mono text-sm text-(--color-text-primary) truncate">
              {displayValue}
            </code>
            <div className="flex items-center gap-1">
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
                  <p>{showDecrypted ? "Hide value" : "Reveal value"}</p>
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
                  <p>{copied ? "Copied!" : "Copy to clipboard"}</p>
                </TooltipContent>
              </Tooltip>
            </div>
          </div>

          {/* Scope Badges */}
          {secret.scopes && secret.scopes.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {secret.scopes.slice(0, 3).map((scope) => (
                <Badge key={scope} variant="secondary" className="text-xs">
                  {scope}
                </Badge>
              ))}
              {secret.scopes.length > 3 && (
                <Badge variant="secondary" className="text-xs">
                  +{secret.scopes.length - 3}
                </Badge>
              )}
            </div>
          )}

          {/* Footer with metadata and actions */}
          <div className="flex items-center justify-between pt-2 border-t border-(--border-subtle)">
            <div className="flex items-center gap-2 text-xs text-(--color-text-muted)">
              <Clock className="h-3 w-3" />
              {secret.last_accessed_at ? (
                <span>
                  Accessed {formatDistanceToNow(new Date(secret.last_accessed_at), { addSuffix: true })}
                </span>
              ) : (
                <span>Never accessed</span>
              )}
            </div>
            <div className="flex items-center gap-1">
              {onEdit && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => onEdit(secret.id)}
                    >
                      <Edit3 className="h-4 w-4 text-(--color-text-muted)" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Edit secret</p>
                  </TooltipContent>
                </Tooltip>
              )}
              {onDelete && (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 hover:text-destructive"
                      onClick={() => onDelete(secret.id)}
                    >
                      <Trash2 className="h-4 w-4 text-(--color-text-muted)" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Delete secret</p>
                  </TooltipContent>
                </Tooltip>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </TooltipProvider>
  );
}
