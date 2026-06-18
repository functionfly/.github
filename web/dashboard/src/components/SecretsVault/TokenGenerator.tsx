/**
 * TokenGenerator - Generate access tokens for secrets
 * Creates time-limited tokens with scoped permissions
 */

import { useState, useCallback, useMemo } from "react";
import {
  X,
  Copy,
  Check,
  Clock,
  Shield,
  AlertTriangle,
  Key,
  Loader2,
} from "lucide-react";
import { format, addHours } from "date-fns";
import { cn } from "@/lib/utils";
import { useGenerateToken, useSecretTokens } from "@/hooks/useVault";
import { useAuthStore } from "@/stores/authStore";
import { getTokensPerSecretLimit } from "@/lib/plan-utils";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// Available scopes for tokens
const AVAILABLE_SCOPES = [
  { value: "read", label: "Read", description: "Can read secret value" },
  { value: "write", label: "Write", description: "Can update secret metadata" },
  { value: "admin", label: "Admin", description: "Full access" },
];

// Preset expiration options
const EXPIRATION_OPTIONS = [
  { value: "1", label: "1 hour" },
  { value: "24", label: "24 hours" },
  { value: "168", label: "7 days" },
  { value: "720", label: "30 days" },
  { value: "8760", label: "1 year" },
];

export interface TokenGeneratorProps {
  secretId: string;
  onClose: () => void;
  onGenerated?: () => void;
}

interface GeneratedToken {
  token_id: string;
  token: string;
  expires_at: string;
  name?: string;
  scopes?: string[];
}

export function TokenGenerator({
  secretId,
  onClose,
  onGenerated,
}: TokenGeneratorProps) {
  const generateToken = useGenerateToken(secretId);
  const { data: tokensResponse } = useSecretTokens(secretId);
  const tokens = tokensResponse?.tokens ?? [];
  const user = useAuthStore((state) => state.user);
  const userPlan = user?.plan;
  const tokenLimit = getTokensPerSecretLimit(userPlan);
  const currentTokenCount = tokens?.length ?? 0;
  const hasReachedTokenLimit = tokenLimit > 0 && currentTokenCount >= tokenLimit;
  const canCreateTokens = tokenLimit > 0;

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [expiresInHours, setExpiresInHours] = useState("24");
  const [selectedScopes, setSelectedScopes] = useState<string[]>(["read"]);
  const [generatedToken, setGeneratedToken] = useState<GeneratedToken | null>(
    null
  );
  const [copied, setCopied] = useState(false);
  const [showToken, setShowToken] = useState(false);

  // Toggle scope selection
  const toggleScope = useCallback((scope: string) => {
    setSelectedScopes((prev) => {
      if (prev.includes(scope)) {
        return prev.filter((s) => s !== scope);
      }
      return [...prev, scope];
    });
  }, []);

  // Handle token generation
  const handleGenerate = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const response = await generateToken.mutateAsync({
        name: name || undefined,
        scopes: selectedScopes,
        expires_in_hours: parseInt(expiresInHours, 10),
      });

      setGeneratedToken(response);
      onGenerated?.();
    } catch (err) {
      console.error("Failed to generate token:", err);
    }
  };

  // Copy token to clipboard
  const handleCopy = useCallback(async () => {
    if (!generatedToken?.token) return;

    try {
      await navigator.clipboard.writeText(generatedToken.token);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  }, [generatedToken]);

  // Calculate expiration time
  const expirationTime = generatedToken
    ? new Date(generatedToken.expires_at)
    : addHours(new Date(), parseInt(expiresInHours, 10));

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Key className="h-5 w-5 text-brand-500" />
            Generate Access Token
          </DialogTitle>
          <DialogDescription>
            Create a time-limited token to access this secret programmatically.
          </DialogDescription>
          {/* Token limit indicator */}
          {canCreateTokens && (
            <div className={cn(
              "mt-2 flex items-center gap-2 text-sm",
              hasReachedTokenLimit ? "text-warning" : "text-muted-foreground"
            )}>
              <span>
                {currentTokenCount} of {tokenLimit} tokens used
              </span>
              {hasReachedTokenLimit && (
                <span className="text-warning">- Limit reached</span>
              )}
            </div>
          )}
        </DialogHeader>

        {!generatedToken ? (
          <form onSubmit={handleGenerate} className="space-y-5">
            {/* Token Name */}
            <div className="space-y-2">
              <Label htmlFor="token-name">Token Name (Optional)</Label>
              <Input
                id="token-name"
                placeholder="e.g., Production API Server"
                value={name}
                onChange={(e) => setName(e.target.value)}
                maxLength={100}
              />
              <p className="text-xs text-text-muted">
                A descriptive name to help identify this token later
              </p>
            </div>

            {/* Expiration */}
            <div className="space-y-2">
              <Label htmlFor="expiration">Expires In</Label>
              <Select value={expiresInHours} onValueChange={setExpiresInHours}>
                <SelectTrigger id="expiration">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {EXPIRATION_OPTIONS.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-text-muted flex items-center gap-1">
                <Clock className="h-3 w-3" />
                Token will expire on {format(expirationTime, "MMM d, yyyy HH:mm")}
              </p>
            </div>

            {/* Scopes */}
            <div className="space-y-2">
              <Label>Scopes</Label>
              <div className="flex flex-wrap gap-2">
                {AVAILABLE_SCOPES.map((scope) => (
                  <button
                    key={scope.value}
                    type="button"
                    onClick={() => toggleScope(scope.value)}
                    className={cn(
                      "inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-all duration-200",
                      selectedScopes.includes(scope.value)
                        ? "bg-brand-500/20 text-brand-600 border border-brand-500/30"
                        : "bg-bg-tertiary text-text-muted border border-border-subtle hover:border-border-default"
                    )}
                  >
                    {scope.label}
                    {selectedScopes.includes(scope.value) && (
                      <span className="text-xs">✓</span>
                    )}
                  </button>
                ))}
              </div>
              <p className="text-xs text-text-muted">
                Select the permissions this token will have
              </p>
            </div>

            {/* Actions */}
            <div className="flex justify-end gap-3 pt-4 border-t border-border-subtle">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={
                  generateToken.isPending ||
                  selectedScopes.length === 0 ||
                  hasReachedTokenLimit ||
                  !canCreateTokens
                }
              >
                {generateToken.isPending ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Generating...
                  </>
                ) : hasReachedTokenLimit ? (
                  <>
                    <AlertTriangle className="h-4 w-4 mr-2" />
                    Limit Reached
                  </>
                ) : !canCreateTokens ? (
                  <>
                    <Shield className="h-4 w-4 mr-2" />
                    Not Available
                  </>
                ) : (
                  <>
                    <Key className="h-4 w-4 mr-2" />
                    Generate Token
                  </>
                )}
              </Button>
            </div>
          </form>
        ) : (
          <div className="space-y-5">
            {/* Token Display */}
            <div className="rounded-lg border border-warning/30 bg-warning-glow p-4 space-y-3">
              <div className="flex items-start gap-3">
                <AlertTriangle className="h-5 w-5 text-warning shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold text-warning">
                    Copy this token now!
                  </h4>
                  <p className="text-sm text-warning/80">
                    This is the only time the token will be displayed. It cannot
                    be retrieved later.
                  </p>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label className="text-warning">Your Token</Label>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => setShowToken(!showToken)}
                      className="text-warning/70 hover:text-warning p-1"
                      title={showToken ? "Hide" : "Show"}
                    >
                      {showToken ? (
                        <span className="sr-only">Hide token</span>
                      ) : (
                        <span className="sr-only">Show token</span>
                      )}
                      {showToken ? (
                        <span className="text-xs">Hide</span>
                      ) : (
                        <span className="text-xs">Show</span>
                      )}
                    </button>
                    <button
                      onClick={handleCopy}
                      className="text-warning/70 hover:text-warning p-1 flex items-center gap-1"
                      title="Copy to clipboard"
                    >
                      {copied ? (
                        <>
                          <Check className="h-4 w-4" />
                          <span className="text-xs">Copied!</span>
                        </>
                      ) : (
                        <>
                          <Copy className="h-4 w-4" />
                          <span className="text-xs">Copy</span>
                        </>
                      )}
                    </button>
                  </div>
                </div>
                <div className="relative">
                  <code
                    className={cn(
                      "block w-full p-3 rounded-lg bg-black/20 font-mono text-sm break-all",
                      showToken ? "text-warning" : "text-warning/50"
                    )}
                  >
                    {showToken
                      ? generatedToken.token
                      : "•".repeat(Math.min(generatedToken.token.length, 40))}
                  </code>
                </div>
              </div>

              <div className="flex items-center gap-2 text-sm text-warning/80">
                <Clock className="h-4 w-4" />
                <span>
                  Expires: {format(new Date(generatedToken.expires_at), "PPpp")}
                </span>
              </div>
            </div>

            {/* Token Details */}
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-text-muted">Token ID</span>
                <span className="font-mono text-card-foreground">
                  {generatedToken.token_id}
                </span>
              </div>
              {generatedToken.name && (
                <div className="flex justify-between">
                  <span className="text-text-muted">Name</span>
                  <span className="text-card-foreground">
                    {generatedToken.name}
                  </span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-text-muted">Scopes</span>
                <div className="flex gap-1">
                  {generatedToken.scopes?.map((scope) => (
                    <Badge key={scope} variant="outline" className="text-xs">
                      {scope}
                    </Badge>
                  ))}
                </div>
              </div>
            </div>

            {/* Close Button */}
            <div className="pt-4 border-t border-border-subtle">
              <Button onClick={onClose} className="w-full">
                I've Copied My Token
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
