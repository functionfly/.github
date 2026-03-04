/**
 * SecretRevealGate - Security gate requiring passkey/WebAuthn reauthentication
 *
 * A security component that requires users to authenticate via WebAuthn/passkey
 * before revealing sensitive secret values. Includes configurable session duration,
 * visual security prompts, and graceful fallbacks for unsupported browsers.
 *
 * @example
 * ```tsx
 * // Basic usage with callback
 * <SecretRevealGate
 *   onVerified={() => setRevealSecret(true)}
 *   requiredLevel="high"
 * >
 *   <SecretCard secret={secret} decryptedValue={value} />
 * </SecretRevealGate>
 *
 * // With custom session duration
 * <SecretRevealGate
 *   onVerified={handleVerified}
 *   sessionDurationMinutes={15}
 *   allowRememberDevice={true}
 * >
 *   <SecretValueDisplay />
 * </SecretRevealGate>
 *
 * // As a modal trigger
 * <SecretRevealGate
 *   onVerified={handleVerified}
 *   trigger={<Button>Reveal Secret</Button>}
 * />
 *
 * // Controlled open state
 * <SecretRevealGate
 *   isOpen={isGateOpen}
 *   onOpenChange={setIsGateOpen}
 *   onVerified={handleVerified}
 * />
 * ```
 */

import { useState, useCallback, useEffect, useRef } from "react";
import {
  Shield,
  ShieldCheck,
  ShieldAlert,
  Fingerprint,
  Key,
  Clock,
  AlertTriangle,
  Check,
  X,
  Loader2,
  Smartphone,
  Laptop,
  Lock,
} from "lucide-react";
import { cn } from "@/lib/utils";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Security level requirements */
export type SecurityLevel = "low" | "medium" | "high" | "critical";

/** Authentication method type */
export type AuthMethod = "webauthn" | "passkey" | "password" | "totp";

/** Verification result from authentication */
export interface VerificationResult {
  success: boolean;
  method: AuthMethod;
  timestamp: string;
  deviceId?: string;
  expiresAt?: string;
}

export interface SecretRevealGateProps {
  /** Children to render after successful verification */
  children?: React.ReactNode;
  /** Custom trigger element to open the gate */
  trigger?: React.ReactNode;
  /** Callback when verification is successful */
  onVerified: (result: VerificationResult) => void;
  /** Callback when verification fails or is cancelled */
  onCancelled?: () => void;
  /** Required security level for this operation */
  requiredLevel?: SecurityLevel;
  /** Session duration in minutes (default: 5) */
  sessionDurationMinutes?: number;
  /** Whether to allow "Remember this device" option */
  allowRememberDevice?: boolean;
  /** Whether the dialog is open (controlled) */
  isOpen?: boolean;
  /** Callback when open state changes (controlled) */
  onOpenChange?: (open: boolean) => void;
  /** Additional CSS classes for the trigger wrapper */
  className?: string;
  /** Description text for the security prompt */
  description?: string;
  /** Title for the security dialog */
  title?: string;
}

// Security level configuration
const securityLevelConfig: Record<
  SecurityLevel,
  { label: string; color: string; icon: typeof Shield }
> = {
  low: { label: "Low Security", color: "text-warning", icon: Shield },
  medium: { label: "Medium Security", color: "text-brand-500", icon: ShieldCheck },
  high: { label: "High Security", color: "text-success", icon: ShieldCheck },
  critical: { label: "Critical Security", color: "text-error", icon: ShieldAlert },
};

// WebAuthn credential options for authentication
interface WebAuthnOptions {
  publicKey: PublicKeyCredentialRequestOptions;
}

/**
 * Check if WebAuthn is supported in the current browser
 */
function isWebAuthnSupported(): boolean {
  return (
    typeof window !== "undefined" &&
    window.PublicKeyCredential !== undefined &&
    typeof window.PublicKeyCredential === "function"
  );
}

/**
 * Check if passkeys are supported (platform authenticator)
 */
async function isPasskeySupported(): Promise<boolean> {
  if (!isWebAuthnSupported()) return false;
  try {
    const available = await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
    return available;
  } catch {
    return false;
  }
}

/**
 * Perform WebAuthn authentication
 */
async function performWebAuthnAuth(
  challenge: string,
  timeout?: number
): Promise<VerificationResult> {
  const publicKeyCredentialRequestOptions: PublicKeyCredentialRequestOptions = {
    challenge: Uint8Array.from(atob(challenge), (c) => c.charCodeAt(0)),
    timeout: timeout || 60000,
    userVerification: "required",
    rpId: window.location.hostname,
  };

  try {
    const credential = await navigator.credentials.get({
      publicKey: publicKeyCredentialRequestOptions,
    });

    if (!credential) {
      throw new Error("No credential returned");
    }

    return {
      success: true,
      method: "webauthn",
      timestamp: new Date().toISOString(),
    };
  } catch (error) {
    throw new Error(
      error instanceof Error ? error.message : "Authentication failed"
    );
  }
}

/**
 * Format remaining time for display
 */
function formatRemainingTime(minutes: number): string {
  if (minutes >= 60) {
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `${hours}h ${mins > 0 ? `${mins}m` : ""}`;
  }
  return `${minutes}m`;
}

/**
 * SecretRevealGate component
 *
 * Renders a security gate that requires WebAuthn/passkey authentication
 * before allowing access to protected content.
 */
export function SecretRevealGate({
  children,
  trigger,
  onVerified,
  onCancelled,
  requiredLevel = "medium",
  sessionDurationMinutes = 5,
  allowRememberDevice = false,
  isOpen: controlledIsOpen,
  onOpenChange,
  className,
  description,
  title,
}: SecretRevealGateProps) {
  // State
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isSupported, setIsSupported] = useState<boolean | null>(null);
  const [isPasskeyAvailable, setIsPasskeyAvailable] = useState(false);
  const [rememberDevice, setRememberDevice] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isVerified, setIsVerified] = useState(false);
  const [verificationResult, setVerificationResult] = useState<VerificationResult | null>(null);
  const [sessionExpiry, setSessionExpiry] = useState<Date | null>(null);

  const sessionTimerRef = useRef<NodeJS.Timeout | null>(null);

  // Determine controlled vs uncontrolled
  const isControlled = controlledIsOpen !== undefined;
  const dialogOpen = isControlled ? controlledIsOpen : isOpen;

  const handleOpenChange = useCallback(
    (open: boolean) => {
      if (!isControlled) {
        setIsOpen(open);
      }
      onOpenChange?.(open);

      if (!open && !isVerified) {
        onCancelled?.();
      }
    },
    [isControlled, onOpenChange, isVerified, onCancelled]
  );

  // Check WebAuthn support on mount
  useEffect(() => {
    const checkSupport = async () => {
      const supported = isWebAuthnSupported();
      setIsSupported(supported);

      if (supported) {
        const passkeyAvailable = await isPasskeySupported();
        setIsPasskeyAvailable(passkeyAvailable);
      }
    };
    checkSupport();
  }, []);

  // Session timeout management
  useEffect(() => {
    if (isVerified && sessionExpiry) {
      const checkExpiry = () => {
        if (new Date() >= sessionExpiry) {
          setIsVerified(false);
          setVerificationResult(null);
          setSessionExpiry(null);
        }
      };

      sessionTimerRef.current = setInterval(checkExpiry, 1000);

      return () => {
        if (sessionTimerRef.current) {
          clearInterval(sessionTimerRef.current);
        }
      };
    }
  }, [isVerified, sessionExpiry]);

  const handleAuthenticate = useCallback(async () => {
    if (!isSupported) {
      setError("WebAuthn is not supported in this browser");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Generate a challenge (in production, this would come from the server)
      const challenge = btoa(String.fromCharCode(...crypto.getRandomValues(new Uint8Array(32))));

      const result = await performWebAuthnAuth(challenge);

      if (result.success) {
        // Calculate session expiry
        const expiry = new Date();
        expiry.setMinutes(expiry.getMinutes() + sessionDurationMinutes);

        const finalResult: VerificationResult = {
          ...result,
          deviceId: rememberDevice ? `device_${Date.now()}` : undefined,
          expiresAt: expiry.toISOString(),
        };

        setIsVerified(true);
        setVerificationResult(finalResult);
        setSessionExpiry(expiry);
        setIsOpen(false);
        onVerified(finalResult);
      }
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Authentication failed. Please try again."
      );
    } finally {
      setIsLoading(false);
    }
  }, [isSupported, sessionDurationMinutes, rememberDevice, onVerified]);

  const handleFallbackAuth = useCallback(() => {
    // Fallback to password authentication
    // In production, this would open a password dialog
    setError("Password authentication not implemented in this demo");
  }, []);

  const handleTriggerClick = useCallback(() => {
    if (isVerified) {
      // Already verified, just call onVerified again with existing result
      if (verificationResult) {
        onVerified(verificationResult);
      }
    } else {
      handleOpenChange(true);
    }
  }, [isVerified, verificationResult, onVerified, handleOpenChange]);

  const securityConfig = securityLevelConfig[requiredLevel];
  const SecurityIcon = securityConfig.icon;

  // Render trigger or children based on verification state
  if (isVerified && children) {
    return <>{children}</>;
  }

  return (
    <TooltipProvider>
      <>
        {/* Trigger Element */}
        {trigger ? (
          <div onClick={handleTriggerClick} className={cn("cursor-pointer", className)}>
            {trigger}
          </div>
        ) : (
          <Button
            onClick={handleTriggerClick}
            className={cn("gap-2", className)}
            variant={isVerified ? "default" : "outline"}
          >
            {isVerified ? (
              <>
                <ShieldCheck className="h-4 w-4" />
                Verified
              </>
            ) : (
              <>
                <Lock className="h-4 w-4" />
                Reveal Secret
              </>
            )}
          </Button>
        )}

        {/* Authentication Dialog */}
        <Dialog open={dialogOpen} onOpenChange={handleOpenChange}>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <div className="flex items-center gap-3">
                <div
                  className={cn(
                    "flex h-12 w-12 items-center justify-center rounded-xl",
                    "bg-gradient-to-br from-(--color-brand-500) to-purple-500"
                  )}
                >
                  <Shield className="h-6 w-6 text-white" />
                </div>
                <div>
                  <DialogTitle className="text-lg font-semibold">
                    {title || "Security Verification Required"}
                  </DialogTitle>
                  <DialogDescription className="text-sm text-(--color-text-muted)">
                    {description ||
                      "Authenticate to access sensitive secret information"}
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>

            {/* Security Level Badge */}
            <div className="flex items-center justify-between p-3 rounded-lg bg-(--color-bg-secondary)">
              <span className="text-sm text-(--color-text-secondary)">
                Required Security Level
              </span>
              <Badge
                variant="outline"
                className={cn("gap-1.5", securityConfig.color)}
              >
                <SecurityIcon className="h-3 w-3" />
                {securityConfig.label}
              </Badge>
            </div>

            {/* Loading State */}
            {isSupported === null && (
              <div className="space-y-4">
                <Skeleton className="h-16 w-full rounded-lg" />
                <Skeleton className="h-10 w-full" />
              </div>
            )}

            {/* WebAuthn Not Supported */}
            {isSupported === false && (
              <div className="p-4 rounded-lg bg-warning/10 border border-warning/20">
                <div className="flex items-start gap-3">
                  <AlertTriangle className="h-5 w-5 text-warning shrink-0 mt-0.5" />
                  <div>
                    <h4 className="font-medium text-warning">
                      Browser Not Supported
                    </h4>
                    <p className="text-sm text-(--color-text-secondary) mt-1">
                      Your browser does not support WebAuthn. Please use a
                      modern browser like Chrome, Firefox, Safari, or Edge.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Authentication Options */}
            {isSupported === true && (
              <div className="space-y-4">
                {/* Passkey Option */}
                <Button
                  onClick={handleAuthenticate}
                  disabled={isLoading}
                  className="w-full h-14 justify-start gap-4 text-left"
                  variant="outline"
                >
                  {isLoading ? (
                    <Loader2 className="h-5 w-5 animate-spin" />
                  ) : (
                    <div
                      className={cn(
                        "flex h-10 w-10 items-center justify-center rounded-lg",
                        "bg-(--color-brand-500)/10"
                      )}
                    >
                      <Fingerprint className="h-5 w-5 text-(--color-brand-500)" />
                    </div>
                  )}
                  <div className="flex-1">
                    <div className="font-medium">Use Passkey</div>
                    <div className="text-xs text-(--color-text-muted)">
                      {isPasskeyAvailable
                        ? "Authenticate with your device's biometric sensor"
                        : "Use a security key or authenticator"}
                    </div>
                  </div>
                  {isPasskeyAvailable && (
                    <Badge variant="secondary" className="text-xs">
                      Recommended
                    </Badge>
                  )}
                </Button>

                {/* Device Icons */}
                <div className="flex justify-center gap-4 py-2">
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="flex flex-col items-center gap-1 text-(--color-text-muted)">
                        <Smartphone className="h-6 w-6" />
                        <span className="text-xs">Phone</span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Mobile authenticator apps</p>
                    </TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="flex flex-col items-center gap-1 text-(--color-text-muted)">
                        <Laptop className="h-6 w-6" />
                        <span className="text-xs">Laptop</span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Built-in biometric sensors</p>
                    </TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="flex flex-col items-center gap-1 text-(--color-text-muted)">
                        <Key className="h-6 w-6" />
                        <span className="text-xs">Security Key</span>
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Hardware security keys</p>
                    </TooltipContent>
                  </Tooltip>
                </div>

                {/* Remember Device Option */}
                {allowRememberDevice && (
                  <div className="flex items-start gap-3 p-3 rounded-lg bg-(--color-bg-secondary)">
                    <Checkbox
                      id="remember-device"
                      checked={rememberDevice}
                      onCheckedChange={(checked) =>
                        setRememberDevice(checked as boolean)
                      }
                    />
                    <div className="space-y-1">
                      <label
                        htmlFor="remember-device"
                        className="text-sm font-medium cursor-pointer"
                      >
                        Remember this device
                      </label>
                      <p className="text-xs text-(--color-text-muted)">
                        Skip authentication on this device for 30 days. Only use
                        on trusted, private devices.
                      </p>
                    </div>
                  </div>
                )}

                {/* Session Duration Info */}
                <div className="flex items-center gap-2 text-xs text-(--color-text-muted)">
                  <Clock className="h-3 w-3" />
                  <span>
                    Session valid for{" "}
                    {formatRemainingTime(sessionDurationMinutes)} after
                    authentication
                  </span>
                </div>

                {/* Error Message */}
                {error && (
                  <div className="p-3 rounded-lg bg-error/10 border border-error/20 flex items-start gap-2">
                    <X className="h-4 w-4 text-error shrink-0 mt-0.5" />
                    <span className="text-sm text-error">{error}</span>
                  </div>
                )}

                {/* Fallback Options */}
                <div className="pt-2 border-t border-(--border-subtle)">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full text-(--color-text-muted)"
                    onClick={handleFallbackAuth}
                  >
                    Use password instead
                  </Button>
                </div>
              </div>
            )}
          </DialogContent>
        </Dialog>
      </>
    </TooltipProvider>
  );
}
