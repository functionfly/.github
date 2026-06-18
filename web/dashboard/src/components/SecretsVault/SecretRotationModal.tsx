/**
 * SecretRotationModal - Modal for rotating secrets with safety checks
 *
 * Provides a comprehensive interface for rotating secrets with options for
 * immediate, scheduled, or automatic rotation. Includes impact analysis,
 * grace period configuration, and rollback capabilities.
 *
 * @example
 * ```tsx
 * // Basic usage
 * <SecretRotationModal
 *   isOpen={isModalOpen}
 *   secretId="secret-123"
 *   secretName="API Key"
 *   onClose={() => setIsModalOpen(false)}
 *   onRotate={handleRotation}
 * />
 *
 * // With custom rotation options
 * <SecretRotationModal
 *   isOpen={isOpen}
 *   secretId="secret-123"
 *   secretName="Production API Key"
 *   secretPreview="sk-...abc123"
 *   impactedServices={["api-gateway", "webhook-handler"]}
 *   rotationOptions={{
 *     allowScheduled: true,
 *     allowAutomatic: true,
 *     maxGracePeriodHours: 168,
 *   }}
 *   onClose={onClose}
 *   onRotate={handleRotate}
 *   onRollback={handleRollback}
 * />
 *
 * // Loading state (for fetching secret data)
 * <SecretRotationModal isOpen={isOpen} isLoading />
 * ```
 */

import { useState, useCallback, useEffect } from "react";
import {
  X,
  RefreshCw,
  AlertTriangle,
  Check,
  Clock,
  Calendar,
  Server,
  Shield,
  AlertCircle,
  Loader2,
  ChevronRight,
  ChevronLeft,
  Eye,
  EyeOff,
  RotateCcw,
  FileKey,
  Zap,
  Timer,
} from "lucide-react";
import { format, addHours } from "date-fns";
import { cn } from "@/lib/utils";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
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
  Alert,
  AlertDescription,
  AlertTitle,
} from "@/components/ui/alert";
import {
  RadioGroup,
  RadioGroupItem,
} from "@/components/ui/radio-group";
import { Slider } from "@/components/ui/slider";
import { vaultApi } from "@/api/vault";

/** Rotation type options */
export type RotationType = "immediate" | "scheduled" | "automatic";

/** Rotation step in the wizard */
export type RotationStep = "options" | "confirm" | "processing" | "success" | "error";

/** Impact analysis for services using the secret */
export interface ImpactedService {
  id: string;
  name: string;
  type: "function" | "service" | "integration" | "workflow";
  criticality: "low" | "medium" | "high" | "critical";
  lastUsedAt?: string;
  estimatedDowntime?: string;
}

/** Rotation options configuration */
export interface RotationOptions {
  /** Allow scheduled rotation */
  allowScheduled?: boolean;
  /** Allow automatic rotation */
  allowAutomatic?: boolean;
  /** Maximum grace period in hours (default: 168 = 7 days) */
  maxGracePeriodHours?: number;
  /** Minimum grace period in hours (default: 0) */
  minGracePeriodHours?: number;
  /** Default grace period in hours (default: 24) */
  defaultGracePeriodHours?: number;
}

/** Rotation request data */
export interface RotationRequest {
  rotationType: RotationType;
  scheduledAt?: Date;
  gracePeriodHours: number;
  autoRotateIntervalDays?: number;
  notifyStakeholders: boolean;
  requireApproval: boolean;
}

/** Rotation result */
export interface RotationResult {
  success: boolean;
  newSecretValue?: string;
  rotationId?: string;
  errorMessage?: string;
  rollbackAvailable?: boolean;
}

export interface SecretRotationModalProps {
  /** Whether the modal is open */
  isOpen: boolean;
  /** Secret ID being rotated */
  secretId: string;
  /** Secret name for display */
  secretName: string;
  /** Preview of current secret value (masked) */
  secretPreview?: string;
  /** Whether the component is loading */
  isLoading?: boolean;
  /** List of services impacted by rotation */
  impactedServices?: ImpactedService[];
  /** Rotation options configuration */
  rotationOptions?: RotationOptions;
  /** Current rotation step (for controlled mode) */
  currentStep?: RotationStep;
  /** Rotation result (for controlled mode) */
  rotationResult?: RotationResult;
  /** Processing progress (0-100) */
  processingProgress?: number;
  /** Processing status message */
  processingMessage?: string;
  /** Callback when modal closes */
  onClose: () => void;
  /** Callback when rotation is requested */
  onRotate?: (request: RotationRequest) => Promise<RotationResult> | void;
  /** Callback when rollback is requested */
  onRollback?: (rotationId: string) => Promise<void> | void;
  /** Callback when step changes */
  onStepChange?: (step: RotationStep) => void;
  /** Additional CSS classes */
  className?: string;
}

/** Default rotation options */
const DEFAULT_ROTATION_OPTIONS: RotationOptions = {
  allowScheduled: true,
  allowAutomatic: true,
  maxGracePeriodHours: 168,
  minGracePeriodHours: 0,
  defaultGracePeriodHours: 24,
};

/** Grace period preset options */
const GRACE_PERIOD_PRESETS = [
  { value: 0, label: "No grace period", description: "Old secret invalidates immediately" },
  { value: 1, label: "1 hour", description: "Brief overlap for in-flight requests" },
  { value: 24, label: "24 hours", description: "Standard grace period" },
  { value: 72, label: "3 days", description: "Extended for complex deployments" },
  { value: 168, label: "7 days", description: "Maximum safety window" },
];

/** Automatic rotation interval options */
const AUTO_ROTATE_INTERVALS = [
  { value: 30, label: "30 days" },
  { value: 90, label: "90 days" },
  { value: 180, label: "6 months" },
  { value: 365, label: "1 year" },
];

/**
 * Mask a secret value showing only last 4 characters
 */
function maskSecret(value: string): string {
  if (!value || value.length <= 8) return "•".repeat(value?.length || 8);
  return value.slice(0, 4) + "•".repeat(value.length - 8) + value.slice(-4);
}

/**
 * Get criticality color
 */
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

/**
 * Service icon by type
 */
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

/**
 * Impact analysis component
 */
function ImpactAnalysis({ services }: { services: ImpactedService[] }) {
  const criticalCount = services.filter((s) => s.criticality === "critical").length;
  const highCount = services.filter((s) => s.criticality === "high").length;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-medium text-(--color-text-primary)">
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
          <Badge variant="secondary" className="text-xs">
            {services.length} Total
          </Badge>
        </div>
      </div>

      <div className="space-y-2 max-h-40 overflow-y-auto">
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
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <span className="text-(--color-text-muted)">
                        ~{service.estimatedDowntime}
                      </span>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Estimated downtime during rotation</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/**
 * Secret rotation modal component
 */
export function SecretRotationModal({
  isOpen,
  secretId,
  secretName,
  secretPreview,
  isLoading = false,
  impactedServices = [],
  rotationOptions = DEFAULT_ROTATION_OPTIONS,
  currentStep,
  rotationResult,
  processingProgress = 0,
  processingMessage = "Processing...",
  onClose,
  onRotate,
  onRollback,
  onStepChange,
  className,
}: SecretRotationModalProps) {
  // Local state for uncontrolled mode
  const [localStep, setLocalStep] = useState<RotationStep>("options");
  const [localResult, setLocalResult] = useState<RotationResult | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);

  // Dependency state
  const [fetchedDependencies, setFetchedDependencies] = useState<ImpactedService[]>([]);
  const [isLoadingDependencies, setIsLoadingDependencies] = useState(false);

  // Form state
  const [rotationType, setRotationType] = useState<RotationType>("immediate");
  const [scheduledAt, setScheduledAt] = useState<Date | undefined>();
  const [gracePeriodHours, setGracePeriodHours] = useState(
    rotationOptions.defaultGracePeriodHours || 24
  );
  const [autoRotateInterval, setAutoRotateInterval] = useState(90);
  const [notifyStakeholders, setNotifyStakeholders] = useState(true);
  const [requireApproval, setRequireApproval] = useState(false);
  const [showSecret, setShowSecret] = useState(false);

  // Confirmation state
  const [confirmationText, setConfirmationText] = useState("");
  const requiredConfirmation = `rotate ${secretName}`;

  // Use controlled or uncontrolled step
  const step = currentStep ?? localStep;
  const result = rotationResult ?? localResult;

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen) {
      setLocalStep("options");
      setLocalResult(null);
      setConfirmationText("");
      setRotationType("immediate");
      setGracePeriodHours(rotationOptions.defaultGracePeriodHours || 24);
    }
  }, [isOpen, rotationOptions.defaultGracePeriodHours]);

  // Fetch dependencies when modal opens if not provided via props
  useEffect(() => {
    if (!isOpen) return;

    // If impactedServices are already provided via props, don't fetch
    if (impactedServices && impactedServices.length > 0) {
      setFetchedDependencies([]);
      return;
    }

    // Fetch dependencies from API
    const fetchDependencies = async () => {
      setIsLoadingDependencies(true);
      try {
        const response = await vaultApi.getSecretDependencies(secretId);
        const mapped: ImpactedService[] = response.dependencies.map((dep) => ({
          id: dep.dependent_id,
          name: dep.name,
          type: dep.dependent_type as ImpactedService["type"],
          criticality: dep.criticality as ImpactedService["criticality"],
          lastUsedAt: undefined,
          estimatedDowntime: undefined,
        }));
        setFetchedDependencies(mapped);
      } catch (error) {
        console.error("Failed to fetch secret dependencies:", error);
        setFetchedDependencies([]);
      } finally {
        setIsLoadingDependencies(false);
      }
    };

    fetchDependencies();
  }, [isOpen, secretId, impactedServices]);

  // Use fetched dependencies if no prop-provided dependencies
  const effectiveImpactedServices = impactedServices && impactedServices.length > 0
    ? impactedServices
    : fetchedDependencies;

  // Handle step changes
  const goToStep = useCallback(
    (newStep: RotationStep) => {
      setLocalStep(newStep);
      onStepChange?.(newStep);
    },
    [onStepChange]
  );

  // Handle rotation initiation
  const handleRotate = useCallback(async () => {
    if (!onRotate) return;

    const request: RotationRequest = {
      rotationType,
      scheduledAt,
      gracePeriodHours,
      autoRotateIntervalDays: rotationType === "automatic" ? autoRotateInterval : undefined,
      notifyStakeholders,
      requireApproval,
    };

    goToStep("processing");
    setIsProcessing(true);

    try {
      const result = await onRotate(request);
      if (result) {
        setLocalResult(result);
        goToStep(result.success ? "success" : "error");
      } else {
        // If onRotate doesn't return a result, assume success
        goToStep("success");
      }
    } catch (error) {
      setLocalResult({
        success: false,
        errorMessage: error instanceof Error ? error.message : "Rotation failed",
        rollbackAvailable: false,
      });
      goToStep("error");
    } finally {
      setIsProcessing(false);
    }
  }, [
    onRotate,
    rotationType,
    scheduledAt,
    gracePeriodHours,
    autoRotateInterval,
    notifyStakeholders,
    requireApproval,
    goToStep,
  ]);

  // Handle rollback
  const handleRollback = useCallback(async () => {
    if (!onRollback || !result?.rotationId) return;

    setIsProcessing(true);
    try {
      await onRollback(result.rotationId);
      goToStep("success");
      setLocalResult((prev) =>
        prev
          ? {
              ...prev,
              success: true,
              errorMessage: undefined,
            }
          : null
      );
    } catch (error) {
      // Keep error state but update message
      setLocalResult((prev) =>
        prev
          ? {
              ...prev,
              errorMessage: `Rollback failed: ${error instanceof Error ? error.message : "Unknown error"}`,
            }
          : null
      );
    } finally {
      setIsProcessing(false);
    }
  }, [onRollback, result?.rotationId, goToStep]);

  // Check if can proceed to confirmation
  const canProceedToConfirm = useCallback(() => {
    if (rotationType === "scheduled" && !scheduledAt) return false;
    return true;
  }, [rotationType, scheduledAt]);

  // Check if can confirm rotation
  const canConfirm = useCallback(() => {
    return confirmationText.toLowerCase() === requiredConfirmation.toLowerCase();
  }, [confirmationText, requiredConfirmation]);

  // Loading skeleton
  if (isLoading) {
    return (
      <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
        <DialogContent className={cn("sm:max-w-lg", className)}>
          <DialogHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-64" />
          </DialogHeader>
          <div className="space-y-4 py-4">
            <Skeleton className="h-32 w-full" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className={cn("sm:max-w-2xl max-h-[90vh] overflow-y-auto", className)}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5 text-brand-500" />
            Rotate Secret
          </DialogTitle>
          <DialogDescription>
            {secretName}
            {secretPreview && (
              <span className="ml-2 text-(--color-text-muted)">
                ({maskSecret(secretPreview)})
              </span>
            )}
          </DialogDescription>
        </DialogHeader>

        {/* Step indicator */}
        <div className="flex items-center gap-2 mb-4">
          {["options", "confirm", "processing", "success"].map((s, i) => (
            <div key={s} className="flex items-center gap-2 flex-1">
              <div
                className={cn(
                  "h-2 flex-1 rounded-full transition-colors",
                  step === s || (i < ["options", "confirm", "processing", "success"].indexOf(step))
                    ? "bg-brand-500"
                    : "bg-(--color-bg-tertiary)"
                )}
              />
            </div>
          ))}
        </div>

        {/* Step: Options */}
        {step === "options" && (
          <div className="space-y-6 py-2">
            {/* Current secret preview */}
            {secretPreview && (
              <div className="p-4 rounded-lg bg-(--color-bg-secondary) border border-(--border-subtle)">
                <Label className="text-xs text-(--color-text-muted)">Current Secret Preview</Label>
                <div className="flex items-center gap-2 mt-1">
                  <code className="flex-1 font-mono text-sm text-(--color-text-primary)">
                    {showSecret ? secretPreview : maskSecret(secretPreview)}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => setShowSecret(!showSecret)}
                    aria-label={showSecret ? 'Hide secret' : 'Show secret'}
                  >
                    {showSecret ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>
            )}

            {/* Rotation type selection */}
            <div className="space-y-3">
              <Label className="text-sm font-medium">Rotation Type</Label>
              <RadioGroup
                value={rotationType}
                onValueChange={(v) => setRotationType(v as RotationType)}
                className="grid grid-cols-1 gap-3"
              >
                <label
                  className={cn(
                    "flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors",
                    rotationType === "immediate"
                      ? "border-brand-500 bg-brand-500/5"
                      : "border-(--border-subtle) hover:border-(--border-default)"
                  )}
                >
                  <RadioGroupItem value="immediate" className="mt-1" />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Zap className="h-4 w-4 text-brand-500" />
                      <span className="font-medium">Immediate Rotation</span>
                    </div>
                    <p className="text-xs text-(--color-text-muted) mt-1">
                      Rotate the secret now. The old value will be invalidated after the grace period.
                    </p>
                  </div>
                </label>

                {rotationOptions.allowScheduled && (
                  <label
                    className={cn(
                      "flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors",
                      rotationType === "scheduled"
                        ? "border-brand-500 bg-brand-500/5"
                        : "border-(--border-subtle) hover:border-(--border-default)"
                    )}
                  >
                    <RadioGroupItem value="scheduled" className="mt-1" />
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <Calendar className="h-4 w-4 text-purple-500" />
                        <span className="font-medium">Scheduled Rotation</span>
                      </div>
                      <p className="text-xs text-(--color-text-muted) mt-1">
                        Schedule the rotation for a future date and time.
                      </p>
                      {rotationType === "scheduled" && (
                        <div className="mt-3">
                          <Input
                            type="datetime-local"
                            min={format(new Date(), "yyyy-MM-dd'T'HH:mm")}
                            onChange={(e) => setScheduledAt(new Date(e.target.value))}
                            className="w-full"
                          />
                        </div>
                      )}
                    </div>
                  </label>
                )}

                {rotationOptions.allowAutomatic && (
                  <label
                    className={cn(
                      "flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors",
                      rotationType === "automatic"
                        ? "border-brand-500 bg-brand-500/5"
                        : "border-(--border-subtle) hover:border-(--border-default)"
                    )}
                  >
                    <RadioGroupItem value="automatic" className="mt-1" />
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <Timer className="h-4 w-4 text-green-500" />
                        <span className="font-medium">Automatic Rotation</span>
                      </div>
                      <p className="text-xs text-(--color-text-muted) mt-1">
                        Enable automatic rotation at regular intervals.
                      </p>
                      {rotationType === "automatic" && (
                        <div className="mt-3">
                          <Label className="text-xs mb-2 block">Rotation Interval</Label>
                          <Select
                            value={autoRotateInterval.toString()}
                            onValueChange={(v) => setAutoRotateInterval(parseInt(v))}
                          >
                            <SelectTrigger className="w-full">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {AUTO_ROTATE_INTERVALS.map((interval) => (
                                <SelectItem key={interval.value} value={interval.value.toString()}>
                                  {interval.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      )}
                    </div>
                  </label>
                )}
              </RadioGroup>
            </div>

            <Separator className="bg-(--border-subtle)" />

            {/* Grace period */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label className="text-sm font-medium">Grace Period</Label>
                <Badge variant="outline" className="text-xs">
                  {gracePeriodHours} hours
                </Badge>
              </div>
              <p className="text-xs text-(--color-text-muted)">
                Old secret remains valid during this period to allow services to update.
              </p>
              <Slider
                value={[gracePeriodHours]}
                onValueChange={([v]) => setGracePeriodHours(v)}
                min={rotationOptions.minGracePeriodHours || 0}
                max={rotationOptions.maxGracePeriodHours || 168}
                step={1}
              />
              <div className="flex flex-wrap gap-2">
                {GRACE_PERIOD_PRESETS.filter(
                  (p) =>
                    p.value >= (rotationOptions.minGracePeriodHours || 0) &&
                    p.value <= (rotationOptions.maxGracePeriodHours || 168)
                ).map((preset) => (
                  <button
                    key={preset.value}
                    onClick={() => setGracePeriodHours(preset.value)}
                    className={cn(
                      "px-2 py-1 rounded text-xs transition-colors",
                      gracePeriodHours === preset.value
                        ? "bg-brand-500 text-white"
                        : "bg-(--color-bg-tertiary) text-(--color-text-muted) hover:bg-(--color-bg-secondary)"
                    )}
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Impact analysis */}
            {(effectiveImpactedServices.length > 0 || isLoadingDependencies) && (
              <>
                <Separator className="bg-(--border-subtle)" />
                {isLoadingDependencies ? (
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <h4 className="text-sm font-medium text-(--color-text-primary)">
                        Impact Analysis
                      </h4>
                      <Loader2 className="h-4 w-4 animate-spin text-(--color-text-muted)" />
                    </div>
                    <div className="space-y-2">
                      <Skeleton className="h-8 w-full" />
                      <Skeleton className="h-8 w-full" />
                    </div>
                  </div>
                ) : (
                  <ImpactAnalysis services={effectiveImpactedServices} />
                )}
              </>
            )}

            {/* Options */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label className="text-sm">Notify Stakeholders</Label>
                  <p className="text-xs text-(--color-text-muted)">
                    Send notification to affected service owners
                  </p>
                </div>
                <Switch checked={notifyStakeholders} onCheckedChange={setNotifyStakeholders} />
              </div>
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label className="text-sm">Require Approval</Label>
                  <p className="text-xs text-(--color-text-muted)">
                    Require admin approval before rotation
                  </p>
                </div>
                <Switch checked={requireApproval} onCheckedChange={setRequireApproval} />
              </div>
            </div>
          </div>
        )}

        {/* Step: Confirm */}
        {step === "confirm" && (
          <div className="space-y-6 py-2">
            <Alert variant="destructive" className="border-error/50 bg-error/10">
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Confirm Rotation</AlertTitle>
              <AlertDescription>
                This action will generate a new secret value. Services using the old value may fail
                after the grace period expires.
              </AlertDescription>
            </Alert>

            <div className="p-4 rounded-lg bg-(--color-bg-secondary) border border-(--border-subtle) space-y-3">
              <h4 className="font-medium text-sm">Rotation Summary</h4>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-(--color-text-muted)">Type</span>
                  <span className="font-medium capitalize">{rotationType}</span>
                </div>
                {rotationType === "scheduled" && scheduledAt && (
                  <div className="flex justify-between">
                    <span className="text-(--color-text-muted)">Scheduled For</span>
                    <span className="font-medium">{format(scheduledAt, "PPP p")}</span>
                  </div>
                )}
                {rotationType === "automatic" && (
                  <div className="flex justify-between">
                    <span className="text-(--color-text-muted)">Auto-rotate Interval</span>
                    <span className="font-medium">{autoRotateInterval} days</span>
                  </div>
                )}
                <div className="flex justify-between">
                  <span className="text-(--color-text-muted)">Grace Period</span>
                  <span className="font-medium">{gracePeriodHours} hours</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-(--color-text-muted)">Impacted Services</span>
                  <span className="font-medium">{effectiveImpactedServices.length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-(--color-text-muted)">Stakeholders Notified</span>
                  <span className="font-medium">{notifyStakeholders ? "Yes" : "No"}</span>
                </div>
              </div>
            </div>

            <div className="space-y-3">
              <Label htmlFor="confirmation" className="text-sm font-medium">
                Type "rotate {secretName}" to confirm
              </Label>
              <Input
                id="confirmation"
                value={confirmationText}
                onChange={(e) => setConfirmationText(e.target.value)}
                placeholder={`rotate ${secretName}`}
                className={cn(
                  confirmationText &&
                    confirmationText.toLowerCase() !== requiredConfirmation.toLowerCase() &&
                    "border-error"
                )}
              />
            </div>
          </div>
        )}

        {/* Step: Processing */}
        {step === "processing" && (
          <div className="py-12 flex flex-col items-center text-center space-y-4">
            <div className="relative">
              <div className="h-16 w-16 rounded-full border-4 border-(--border-subtle) border-t-brand-500 animate-spin" />
              <RefreshCw className="h-6 w-6 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-brand-500" />
            </div>
            <div className="space-y-1">
              <h4 className="font-medium text-(--color-text-primary)">Rotating Secret...</h4>
              <p className="text-sm text-(--color-text-muted)">{processingMessage}</p>
            </div>
            <div className="w-full max-w-xs space-y-2">
              <Progress value={processingProgress} className="h-2" />
              <p className="text-xs text-(--color-text-muted)">{processingProgress}% complete</p>
            </div>
          </div>
        )}

        {/* Step: Success */}
        {step === "success" && result?.success && (
          <div className="py-8 flex flex-col items-center text-center space-y-4">
            <div className="h-16 w-16 rounded-full bg-success/10 flex items-center justify-center">
              <Check className="h-8 w-8 text-success" />
            </div>
            <div className="space-y-1">
              <h4 className="font-medium text-(--color-text-primary)">Rotation Successful</h4>
              <p className="text-sm text-(--color-text-muted)">
                The secret has been rotated successfully. Make sure to update your services with the new value.
              </p>
            </div>
            {result.newSecretValue && (
              <div className="w-full max-w-md p-4 rounded-lg bg-(--color-bg-secondary) border border-(--border-subtle)">
                <Label className="text-xs text-(--color-text-muted)">New Secret Value</Label>
                <div className="flex items-center gap-2 mt-1">
                  <code className="flex-1 font-mono text-sm text-(--color-text-primary) truncate">
                    {showSecret ? result.newSecretValue : maskSecret(result.newSecretValue)}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => setShowSecret(!showSecret)}
                    aria-label={showSecret ? 'Hide secret' : 'Show secret'}
                  >
                    {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </Button>
                </div>
                <p className="text-xs text-error mt-2 flex items-center gap-1">
                  <AlertCircle className="h-3 w-3" />
                  Copy this now. It won't be shown again.
                </p>
              </div>
            )}
          </div>
        )}

        {/* Step: Error */}
        {step === "error" && !result?.success && (
          <div className="py-8 flex flex-col items-center text-center space-y-4">
            <div className="h-16 w-16 rounded-full bg-error/10 flex items-center justify-center">
              <X className="h-8 w-8 text-error" />
            </div>
            <div className="space-y-1">
              <h4 className="font-medium text-(--color-text-primary)">Rotation Failed</h4>
              <p className="text-sm text-(--color-text-muted)">
                {result?.errorMessage || "An unexpected error occurred during rotation."}
              </p>
            </div>
            {result?.rollbackAvailable && (
              <Alert className="max-w-md border-warning/50 bg-warning/10">
                <RotateCcw className="h-4 w-4" />
                <AlertTitle>Rollback Available</AlertTitle>
                <AlertDescription className="space-y-3">
                  <p>You can rollback to the previous secret value.</p>
                  <Button
                    variant="outline"
                    onClick={handleRollback}
                    disabled={isProcessing}
                    className="w-full"
                  >
                    {isProcessing ? (
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : (
                      <RotateCcw className="h-4 w-4 mr-2" />
                    )}
                    Rollback to Previous Version
                  </Button>
                </AlertDescription>
              </Alert>
            )}
          </div>
        )}

        {/* Footer */}
        <DialogFooter className="gap-2">
          {step === "options" && (
            <>
              <Button variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button
                onClick={() => goToStep("confirm")}
                disabled={!canProceedToConfirm()}
                className="gap-2"
              >
                Continue
                <ChevronRight className="h-4 w-4" />
              </Button>
            </>
          )}
          {step === "confirm" && (
            <>
              <Button variant="outline" onClick={() => goToStep("options")}>
                <ChevronLeft className="h-4 w-4 mr-2" />
                Back
              </Button>
              <Button
                onClick={handleRotate}
                disabled={!canConfirm() || isProcessing}
                variant="destructive"
                className="gap-2"
              >
                {isProcessing ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4" />
                )}
                Rotate Secret
              </Button>
            </>
          )}
          {(step === "success" || step === "error") && (
            <Button onClick={onClose} className="w-full sm:w-auto">
              {step === "success" ? "Done" : "Close"}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default SecretRotationModal;
