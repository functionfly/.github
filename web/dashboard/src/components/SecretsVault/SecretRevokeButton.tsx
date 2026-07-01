/**
 * SecretRevokeButton - Button component for revoking secrets
 *
 * Provides a button with multiple revoke modes: immediate, scheduled,
 * and graceful (with grace period). Includes confirmation dialog with
 * safety checks, impact preview of what will be affected, danger styling
 * with red/destructive variants, loading state during revocation,
 * success/error feedback, option to provide revocation reason, and
 * cascading revoke options for dependent tokens.
 *
 * @example
 * ```tsx
 * // Basic usage - immediate revoke
 * <SecretRevokeButton
 *   secretId="secret-123"
 *   secretName="API Key"
 *   onRevoke={handleRevoke}
 * />
 *
 * // With all revoke modes enabled
 * <SecretRevokeButton
 *   secretId="secret-123"
 *   secretName="Production API Key"
 *   secretPreview="sk-...abc123"
 *   impactedServices={[
 *     { id: "svc1", name: "API Gateway", type: "service", criticality: "high" }
 *   ]}
 *   dependentTokens={3}
 *   revokeOptions={{
 *     allowImmediate: true,
 *     allowScheduled: true,
 *     allowGraceful: true,
 *     maxGracePeriodHours: 72,
 *   }}
 *   onRevoke={handleRevoke}
 *   onSchedule={handleScheduleRevoke}
 * />
 *
 * // Compact variant
 * <SecretRevokeButton
 *   secretId="secret-123"
 *   variant="compact"
 *   onRevoke={handleRevoke}
 * />
 * ```
 */

import { useState, useCallback, useMemo } from "react";
import {
  Ban,
  AlertTriangle,
  Clock,
  Calendar,
  Loader2,
  CheckCircle,
  XCircle,
  Server,
  Key,
  Zap,
  RefreshCw,
  X,
  ShieldAlert,
  FileKey,
  AlertOctagon,
  Trash2,
} from "lucide-react";
import { format, addHours } from "date-fns";
import { cn } from "@/lib/utils";

import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  RadioGroup,
  RadioGroupItem,
} from "@/components/ui/radio-group";
import { Slider } from "@/components/ui/slider";

/** Revoke mode options */
export type RevokeMode = "immediate" | "scheduled" | "graceful";

/** Service impacted by secret revocation */
export interface ImpactedService {
  id: string;
  name: string;
  type: "function" | "service" | "integration" | "workflow";
  criticality: "low" | "medium" | "high" | "critical";
  lastUsedAt?: string;
  estimatedDowntime?: string;
}

/** Revoke options configuration */
export interface RevokeOptions {
  /** Allow immediate revocation */
  allowImmediate?: boolean;
  /** Allow scheduled revocation */
  allowScheduled?: boolean;
  /** Allow graceful revocation with grace period */
  allowGraceful?: boolean;
  /** Maximum grace period in hours (default: 168 = 7 days) */
  maxGracePeriodHours?: number;
  /** Minimum grace period in hours (default: 0) */
  minGracePeriodHours?: number;
  /** Default grace period in hours (default: 24) */
  defaultGracePeriodHours?: number;
  /** Require confirmation reason */
  requireReason?: boolean;
}

/** Revoke request data */
export interface RevokeRequest {
  mode: RevokeMode;
  scheduledAt?: Date;
  gracePeriodHours: number;
  reason?: string;
  cascadeRevokeTokens: boolean;
  notifyStakeholders: boolean;
}

/** Revoke result */
export interface RevokeResult {
  success: boolean;
  revokeId?: string;
  scheduledAt?: string;
  gracePeriodEndsAt?: string;
  errorMessage?: string;
}

export interface SecretRevokeButtonProps {
  /** Secret ID to revoke */
  secretId: string;
  /** Secret name for display */
  secretName: string;
  /** Preview of secret value (masked) */
  secretPreview?: string;
  /** Number of tokens that will be affected */
  dependentTokens?: number;
  /** List of services impacted by revocation */
  impactedServices?: ImpactedService[];
  /** Revoke options configuration */
  revokeOptions?: RevokeOptions;
  /** Button variant */
  variant?: "default" | "compact" | "icon";
  /** Whether revocation is in progress */
  isRevoking?: boolean;
  /** Current revoke step (for controlled mode) */
  currentStep?: RevokeStep;
  /** Revoke result (for controlled mode) */
  revokeResult?: RevokeResult;
  /** Callback when revoke is requested */
  onRevoke?: (request: RevokeRequest) => Promise<RevokeResult> | void;
  /** Callback when dialog opens/closes */
  onOpenChange?: (open: boolean) => void;
  /** Additional CSS classes */
  className?: string;
}

/** Revoke step in the flow */
export type RevokeStep = "confirm" | "processing" | "success" | "error";

/** Default revoke options */
const DEFAULT_REVOKE_OPTIONS: RevokeOptions = {
  allowImmediate: true,
  allowScheduled: true,
  allowGraceful: true,
  maxGracePeriodHours: 168,
  minGracePeriodHours: 0,
  defaultGracePeriodHours: 24,
  requireReason: false,
};

/** Grace period preset options */
const GRACE_PERIOD_PRESETS = [
  { value: 0, label: "Immediate", description: "Revoke instantly" },
  { value: 1, label: "1 hour", description: "Brief grace period" },
  { value: 24, label: "24 hours", description: "Standard grace period" },
  { value: 72, label: "3 days", description: "Extended grace period" },
  { value: 168, label: "7 days", description: "Maximum grace period" },
];

// Service icon by type
function ServiceIcon({ type }: { type: ImpactedService["type"] }) {
  switch (type) {
    case "function":
      return <Zap className="h-4 w-4" />;
    case "service":
      return <Server className="h-4 w-4" />;
    case "integration":
      return <FileKey className="h-4 w-4" />;
    case "workflow":
      return <RefreshCw className="h-4 w-4" />;
    default:
      return <Server className="h-4 w-4" />;
  }
}

// Get criticality color
function getCriticalityColor(criticality: ImpactedService["criticality"]): string {
  switch (criticality) {
    case "critical":
      return "text-error bg-error/10 border-error/30";
    case "high":
      return "text-orange-500 bg-orange-500/10 border-orange-500/30";
    case "medium":
      return "text-yellow-500 bg-yellow-500/10 border-yellow-500/30";
    case "low":
      return "text-green-500 bg-green-500/10 border-green-500/30";
    default:
      return "text-gray-500 bg-gray-500/10 border-gray-500/30";
  }
}

// Impact analysis component
function ImpactAnalysis({
  services,
  tokenCount,
}: {
  services: ImpactedService[];
  tokenCount: number;
}) {
  const criticalCount = services.filter((s) => s.criticality === "critical").length;
  const highCount = services.filter((s) => s.criticality === "high").length;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-[var(--text)] flex items-center gap-2">
          <AlertTriangle className="h-4 w-4 text-warning" />
          Impact Analysis
        </h4>
        <div className="flex items-center gap-2">
          {criticalCount > 0 && (
            <Badge variant="destructive" className="text-xs">
              {criticalCount} Critical
            </Badge>
          )}
          {highCount > 0 && (
            <Badge variant="warning" className="text-xs">
              {highCount} High
            </Badge>
          )}
        </div>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-2 gap-2 text-sm">
        <div className="flex items-center gap-2 text-[var(--text-dim)]">
          <Server className="h-4 w-4" />
          <span>{services.length} services affected</span>
        </div>
        {tokenCount > 0 && (
          <div className="flex items-center gap-2 text-[var(--text-dim)]">
            <Key className="h-4 w-4" />
            <span>{tokenCount} tokens revoked</span>
          </div>
        )}
      </div>

      {/* Service list */}
      {services.length > 0 && (
        <div className="space-y-2 max-h-40 overflow-y-auto rounded-lg border border-[var(--panel-edge)]-subtle p-2">
          {services.map((service) => (
            <div
              key={service.id}
              className={cn(
                "flex items-center justify-between p-2 rounded-lg border text-xs",
                getCriticalityColor(service.criticality)
              )}
            >
              <div className="flex items-center gap-2">
                <ServiceIcon type={service.type} />
                <span className="font-medium">{service.name}</span>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="text-[10px] uppercase">
                  {service.criticality}
                </Badge>
                {service.estimatedDowntime && (
                  <span className="text-[var(--text-faint)]">~{service.estimatedDowntime}</span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * SecretRevokeButton component
 */
export function SecretRevokeButton({
  secretId,
  secretName,
  secretPreview,
  dependentTokens = 0,
  impactedServices = [],
  revokeOptions = DEFAULT_REVOKE_OPTIONS,
  variant = "default",
  isRevoking: controlledRevoking,
  currentStep,
  revokeResult,
  onRevoke,
  onOpenChange,
  className,
}: SecretRevokeButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [localStep, setLocalStep] = useState<RevokeStep>("confirm");
  const [localRevoking, setLocalRevoking] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  // Form state
  const [revokeMode, setRevokeMode] = useState<RevokeMode>("immediate");
  const [scheduledDate, setScheduledDate] = useState<Date>(addHours(new Date(), 24));
  const [gracePeriodHours, setGracePeriodHours] = useState(
    revokeOptions.defaultGracePeriodHours ?? 24
  );
  const [reason, setReason] = useState("");
  const [cascadeRevoke, setCascadeRevoke] = useState(true);
  const [notifyStakeholders, setNotifyStakeholders] = useState(true);

  // Controlled vs uncontrolled
  const step = currentStep ?? localStep;
  const isRevoking = controlledRevoking ?? localRevoking;

  // Merge with defaults
  const options = { ...DEFAULT_REVOKE_OPTIONS, ...revokeOptions };

  // Handle dialog open
  const handleOpenChange = useCallback(
    (open: boolean) => {
      setIsOpen(open);
      if (!open) {
        // Reset form when closing
        setLocalStep("confirm");
        setLocalError(null);
        setReason("");
        setRevokeMode("immediate");
      }
      onOpenChange?.(open);
    },
    [onOpenChange]
  );

  // Handle revoke submission
  const handleRevoke = useCallback(async () => {
    if (!onRevoke) return;

    setLocalRevoking(true);
    setLocalStep("processing");
    setLocalError(null);

    const request: RevokeRequest = {
      mode: revokeMode,
      scheduledAt: revokeMode === "scheduled" ? scheduledDate : undefined,
      gracePeriodHours: revokeMode === "graceful" ? gracePeriodHours : 0,
      reason: reason || undefined,
      cascadeRevokeTokens: cascadeRevoke,
      notifyStakeholders,
    };

    try {
      const result = await onRevoke(request);

      if (!result || result.success) {
        setLocalStep("success");
      } else {
        setLocalStep("error");
        setLocalError(result.errorMessage ?? "Revocation failed");
      }
    } catch (error) {
      setLocalStep("error");
      setLocalError(error instanceof Error ? error.message : "An error occurred");
    } finally {
      setLocalRevoking(false);
    }
  }, [
    onRevoke,
    revokeMode,
    scheduledDate,
    gracePeriodHours,
    reason,
    cascadeRevoke,
    notifyStakeholders,
  ]);

  // Handle close after success/error
  const handleClose = useCallback(() => {
    handleOpenChange(false);
  }, [handleOpenChange]);

  // Render button based on variant
  const renderButton = () => {
    const baseProps = {
      onClick: () => handleOpenChange(true),
      className: cn(
        variant === "compact" && "gap-2",
        variant === "icon" && "h-8 w-8 p-0",
        className
      ),
    };

    switch (variant) {
      case "compact":
        return (
          <Button variant="outline" {...baseProps}>
            <Ban className="h-4 w-4 text-error" />
            Revoke
          </Button>
        );
      case "icon":
        return (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" {...baseProps}>
                  <Ban className="h-4 w-4 text-error" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Revoke secret</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        );
      default:
        return (
          <Button variant="destructive" {...baseProps}>
            <ShieldAlert className="h-4 w-4 mr-2" />
            Revoke Secret
          </Button>
        );
    }
  };

  return (
    <>
      {renderButton()}

      <Dialog open={isOpen} onOpenChange={handleOpenChange}>
        <DialogContent className="sm:max-w-lg">
          {step === "confirm" && (
            <>
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2 text-error">
                  <AlertOctagon className="h-5 w-5" />
                  Revoke Secret
                </DialogTitle>
                <DialogDescription>
                  This action will revoke access to{" "}
                  <strong>{secretName}</strong>. This cannot be undone.
                </DialogDescription>
              </DialogHeader>

              {/* Secret preview */}
              {secretPreview && (
                <div className="rounded-lg bg-bg-secondary p-3">
                  <Label className="text-xs text-[var(--text-faint)]">Secret Preview</Label>
                  <code className="block mt-1 text-sm font-mono text-[var(--text)]">
                    {secretPreview}
                  </code>
                </div>
              )}

              {/* Impact analysis */}
              {(impactedServices.length > 0 || dependentTokens > 0) && (
                <ImpactAnalysis
                  services={impactedServices}
                  tokenCount={dependentTokens}
                />
              )}

              {/* Revoke mode selection */}
              <div className="space-y-3">
                <Label>Revoke Mode</Label>
                <RadioGroup
                  value={revokeMode}
                  onValueChange={(value) => setRevokeMode(value as RevokeMode)}
                  className="grid grid-cols-3 gap-2"
                >
                  {options.allowImmediate && (
                    <div>
                      <RadioGroupItem
                        value="immediate"
                        id="immediate"
                        className="sr-only"
                      />
                      <Label
                        htmlFor="immediate"
                        className={cn(
                          "flex flex-col items-center justify-center rounded-lg border-2 border-[var(--panel-edge)]-subtle p-3 cursor-pointer transition-colors hover:border-[var(--panel-edge)]-hover",
                          revokeMode === "immediate" && "border-error bg-error/5"
                        )}
                      >
                        <Ban className="h-5 w-5 mb-1 text-error" />
                        <span className="text-xs font-medium">Immediate</span>
                      </Label>
                    </div>
                  )}
                  {options.allowScheduled && (
                    <div>
                      <RadioGroupItem
                        value="scheduled"
                        id="scheduled"
                        className="sr-only"
                      />
                      <Label
                        htmlFor="scheduled"
                        className={cn(
                          "flex flex-col items-center justify-center rounded-lg border-2 border-[var(--panel-edge)]-subtle p-3 cursor-pointer transition-colors hover:border-[var(--panel-edge)]-hover",
                          revokeMode === "scheduled" && "border-[rgba(143,255,208,0.3)] rgba(143,255,208,0.15)/5"
                        )}
                      >
                        <Calendar className="h-5 w-5 mb-1 text-[var(--status-ok)]" />
                        <span className="text-xs font-medium">Scheduled</span>
                      </Label>
                    </div>
                  )}
                  {options.allowGraceful && (
                    <div>
                      <RadioGroupItem
                        value="graceful"
                        id="graceful"
                        className="sr-only"
                      />
                      <Label
                        htmlFor="graceful"
                        className={cn(
                          "flex flex-col items-center justify-center rounded-lg border-2 border-[var(--panel-edge)]-subtle p-3 cursor-pointer transition-colors hover:border-[var(--panel-edge)]-hover",
                          revokeMode === "graceful" && "border-warning bg-warning/5"
                        )}
                      >
                        <Clock className="h-5 w-5 mb-1 text-warning" />
                        <span className="text-xs font-medium">Graceful</span>
                      </Label>
                    </div>
                  )}
                </RadioGroup>
              </div>

              {/* Scheduled date picker */}
              {revokeMode === "scheduled" && (
                <div className="space-y-2">
                  <Label htmlFor="scheduled-date">Schedule For</Label>
                  <Input
                    id="scheduled-date"
                    type="datetime-local"
                    value={format(scheduledDate, "yyyy-MM-dd'T'HH:mm")}
                    onChange={(e) => setScheduledDate(new Date(e.target.value))}
                    min={format(new Date(), "yyyy-MM-dd'T'HH:mm")}
                  />
                </div>
              )}

              {/* Grace period slider */}
              {revokeMode === "graceful" && (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <Label>Grace Period</Label>
                    <Badge variant="outline">{gracePeriodHours} hours</Badge>
                  </div>
                  <Slider
                    value={[gracePeriodHours]}
                    onValueChange={([value]) => setGracePeriodHours(value)}
                    min={options.minGracePeriodHours ?? 0}
                    max={options.maxGracePeriodHours ?? 168}
                    step={1}
                  />
                  <div className="flex justify-between text-xs text-[var(--text-faint)]">
                    <span>Immediate</span>
                    <span>{options.maxGracePeriodHours} hours max</span>
                  </div>
                </div>
              )}

              {/* Revocation reason */}
              <div className="space-y-2">
                <Label htmlFor="reason">
                  Reason for Revocation
                  {options.requireReason && (
                    <span className="text-error ml-1">*</span>
                  )}
                </Label>
                <Textarea
                  id="reason"
                  placeholder="Why is this secret being revoked?"
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  rows={2}
                />
              </div>

              {/* Options */}
              <div className="space-y-3">
                {dependentTokens > 0 && (
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label className="text-sm">Revoke Dependent Tokens</Label>
                      <p className="text-xs text-[var(--text-faint)]">
                        Also revoke {dependentTokens} access token
                        {dependentTokens !== 1 ? "s" : ""}
                      </p>
                    </div>
                    <Switch checked={cascadeRevoke} onCheckedChange={setCascadeRevoke} />
                  </div>
                )}

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label className="text-sm">Notify Stakeholders</Label>
                    <p className="text-xs text-[var(--text-faint)]">
                      Send notifications to affected users
                    </p>
                  </div>
                  <Switch
                    checked={notifyStakeholders}
                    onCheckedChange={setNotifyStakeholders}
                  />
                </div>
              </div>

              {/* Warning */}
              <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertTitle>This action cannot be undone</AlertTitle>
                <AlertDescription>
                  Once revoked, this secret will no longer be accessible. Any services
                  using it will fail.
                </AlertDescription>
              </Alert>

              <DialogFooter>
                <Button variant="outline" onClick={() => handleOpenChange(false)}>
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  onClick={handleRevoke}
                  disabled={
                    isRevoking ||
                    (options.requireReason && !reason.trim()) ||
                    (revokeMode === "scheduled" && scheduledDate <= new Date())
                  }
                >
                  {isRevoking ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Revoking...
                    </>
                  ) : (
                    <>
                      <Ban className="h-4 w-4 mr-2" />
                      {revokeMode === "immediate"
                        ? "Revoke Now"
                        : revokeMode === "scheduled"
                        ? "Schedule Revoke"
                        : "Start Grace Period"}
                    </>
                  )}
                </Button>
              </DialogFooter>
            </>
          )}

          {step === "processing" && (
            <div className="py-12 text-center">
              <Loader2 className="h-12 w-12 animate-spin text-[var(--status-ok)] mx-auto mb-4" />
              <h3 className="text-lg font-semibold text-[var(--text)] mb-2">
                Processing Revocation...
              </h3>
              <p className="text-[var(--text-dim)]">
                {revokeMode === "immediate"
                  ? "Revoking secret access..."
                  : revokeMode === "scheduled"
                  ? "Scheduling revocation..."
                  : "Setting up grace period..."}
              </p>
            </div>
          )}

          {step === "success" && (
            <div className="py-12 text-center">
              <div className="mx-auto h-12 w-12 rounded-full bg-success-glow flex items-center justify-center mb-4">
                <CheckCircle className="h-6 w-6 text-success" />
              </div>
              <h3 className="text-lg font-semibold text-[var(--text)] mb-2">
                Secret Revoked Successfully
              </h3>
              <p className="text-[var(--text-dim)] mb-4">
                {revokeMode === "immediate"
                  ? "The secret has been immediately revoked."
                  : revokeMode === "scheduled"
                  ? `Revocation scheduled for ${format(scheduledDate, "MMM d, yyyy HH:mm")}.`
                  : `Grace period started. Secret will be revoked in ${gracePeriodHours} hours.`}
              </p>
              {cascadeRevoke && dependentTokens > 0 && (
                <p className="text-sm text-[var(--text-faint)]">
                  {dependentTokens} access token{dependentTokens !== 1 ? "s" : ""} also
                  revoked.
                </p>
              )}
              <Button onClick={handleClose} className="mt-6">
                Close
              </Button>
            </div>
          )}

          {step === "error" && (
            <div className="py-12 text-center">
              <div className="mx-auto h-12 w-12 rounded-full bg-error-glow flex items-center justify-center mb-4">
                <XCircle className="h-6 w-6 text-error" />
              </div>
              <h3 className="text-lg font-semibold text-error mb-2">
                Revocation Failed
              </h3>
              <p className="text-[var(--text-dim)] mb-4">
                {localError || revokeResult?.errorMessage || "An unexpected error occurred."}
              </p>
              <div className="flex justify-center gap-2">
                <Button variant="outline" onClick={() => setLocalStep("confirm")}>
                  Try Again
                </Button>
                <Button onClick={handleClose}>Close</Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}
