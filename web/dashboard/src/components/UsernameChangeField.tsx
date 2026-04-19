import { usersApi, type UsernameChangeEligibility } from '@/api/users';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuthStore } from '@/stores/authStore';
import { useQuery } from '@tanstack/react-query';
import { AlertCircle, Check, Clock, DollarSign, History, Info, Loader2, User } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';

interface UsernameChangeFieldProps {
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}

// Username validation rules (must match backend)
const USERNAME_REGEX = /^[a-z0-9_-]+$/;
const MIN_USERNAME_LENGTH = 3;
const MAX_USERNAME_LENGTH = 30;

function validateUsername(username: string): string | null {
  if (!username) {
    return 'Username is required';
  }
  if (username.length < MIN_USERNAME_LENGTH) {
    return `Username must be at least ${MIN_USERNAME_LENGTH} characters`;
  }
  if (username.length > MAX_USERNAME_LENGTH) {
    return `Username must be ${MAX_USERNAME_LENGTH} characters or fewer`;
  }
  if (!USERNAME_REGEX.test(username)) {
    return 'Username can only contain lowercase letters, numbers, hyphens, and underscores';
  }
  if (/^-|-$/.test(username)) {
    return 'Username cannot start or end with a hyphen';
  }
  return null;
}

function formatCurrency(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

function formatDate(dateString?: string): string {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

export function UsernameChangeField({ value, onChange, disabled }: UsernameChangeFieldProps) {
  const user = useAuthStore((state) => state.user);
  const currentUsername = user?.username || '';
  const [newUsername, setNewUsername] = useState(value);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [isChanging, setIsChanging] = useState(false);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [showHistoryDialog, setShowHistoryDialog] = useState(false);

  // Fetch eligibility on mount
  const {
    data: eligibility,
    isLoading: isLoadingEligibility,
    refetch: refetchEligibility,
  } = useQuery({
    queryKey: ['username-change-eligibility'],
    queryFn: usersApi.getUsernameChangeEligibility,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Fetch history
  const { data: historyData, isLoading: isLoadingHistory } = useQuery({
    queryKey: ['username-change-history'],
    queryFn: usersApi.getUsernameChangeHistory,
    enabled: showHistoryDialog,
  });

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawValue = e.target.value.toLowerCase();
    // Auto-clean invalid characters as user types
    const cleaned = rawValue.replace(/[^a-z0-9_-]/g, '');
    setNewUsername(cleaned);
    onChange(cleaned);

    // Validate in real-time
    const error = validateUsername(cleaned);
    setValidationError(error);
  };

  const handleInitiateChange = () => {
    const error = validateUsername(newUsername);
    if (error) {
      setValidationError(error);
      toast.error(error);
      return;
    }

    if (newUsername === currentUsername) {
      toast.info('This is already your current username');
      return;
    }

    setShowConfirmDialog(true);
  };

  const handleConfirmChange = async (payFee: boolean = false) => {
    setIsChanging(true);
    try {
      // If paying a fee, create checkout session and redirect to Stripe
      if (payFee && eligibility && !eligibility.canChangeFreely) {
        const checkoutResponse = await usersApi.createUsernameChangeCheckout({
          new_username: newUsername,
          success_url: `${window.location.origin}/settings?usernameChange=success`,
          cancel_url: `${window.location.origin}/settings?usernameChange=cancel`,
        });

        // Redirect to Stripe checkout
        if (checkoutResponse.url) {
          toast.info('Redirecting to payment...');
          window.location.href = checkoutResponse.url;
          return;
        }
      }

      // Free change - proceed directly
      const response = await usersApi.changeUsername({
        new_username: newUsername,
      });

      if (response.success) {
        toast.success(response.message);
        // Refresh eligibility data
        await refetchEligibility();
        // Update auth store
        await useAuthStore.getState().initialize();
        setShowConfirmDialog(false);
      } else {
        toast.error(response.message || 'Failed to change username');
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to change username';
      toast.error(message);
    } finally {
      setIsChanging(false);
    }
  };

  const canShowChangeButton =
    eligibility &&
    newUsername !== currentUsername &&
    !validationError &&
    newUsername.length >= MIN_USERNAME_LENGTH;

  const getStatusBadge = () => {
    if (!eligibility) return null;

    if (eligibility.canChangeFreely && eligibility.changesRemaining > 0) {
      return (
        <Badge variant="default" className="bg-green-600 hover:bg-green-700">
          <Check className="w-3 h-3 mr-1" />
          {eligibility.changesRemaining} free change{eligibility.changesRemaining !== 1 ? 's' : ''} left
        </Badge>
      );
    }

    if (eligibility.canChangeWithFee && !eligibility.canChangeFreely) {
      return (
        <Badge variant="outline" className="border-amber-500 text-amber-700 dark:text-amber-400">
          <DollarSign className="w-3 h-3 mr-1" />
          Fee required ({formatCurrency(eligibility.earlyChangeFeeCents)})
        </Badge>
      );
    }

    if (eligibility.nextFreeChangeDate) {
      return (
        <Badge variant="outline" className="border-gray-400">
          <Clock className="w-3 h-3 mr-1" />
          Free on {formatDate(eligibility.nextFreeChangeDate)}
        </Badge>
      );
    }

    return (
      <Badge variant="outline" className="border-gray-400">
        <AlertCircle className="w-3 h-3 mr-1" />
        Limit reached
      </Badge>
    );
  };

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="username" className="flex items-center gap-2">
            <User className="w-4 h-4 text-text-secondary" />
            Username
          </Label>
          <div className="flex items-center gap-2">
            {isLoadingEligibility ? (
              <Loader2 className="w-4 h-4 animate-spin text-text-muted" />
            ) : (
              getStatusBadge()
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowHistoryDialog(true)}
              className="text-text-muted hover:text-text-primary"
            >
              <History className="w-4 h-4 mr-1" />
              History
            </Button>
          </div>
        </div>

        <div className="flex gap-2">
          <Input
            id="username"
            type="text"
            placeholder="yourhandle"
            value={newUsername}
            onChange={handleInputChange}
            disabled={disabled || isChanging}
            className={validationError ? 'border-error' : ''}
          />
          {canShowChangeButton && (
            <Button
              onClick={handleInitiateChange}
              disabled={isChanging || isLoadingEligibility}
              variant="default"
            >
              {isChanging ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : eligibility?.canChangeFreely ? (
                'Change'
              ) : eligibility?.canChangeWithFee ? (
                <>
                  <DollarSign className="w-4 h-4 mr-1" />
                  Change
                </>
              ) : (
                'Change'
              )}
            </Button>
          )}
        </div>

        {validationError && (
          <p className="text-sm text-error">{validationError}</p>
        )}

        <p className="text-xs text-text-muted">
          Lowercase letters, numbers, hyphens and underscores only. 3-30 characters.
          {newUsername && !validationError && (
            <span className="ml-1 text-brand-400">Public URL: /u/{newUsername}</span>
          )}
        </p>

        {eligibility && !eligibility.canChangeFreely && !eligibility.canChangeWithFee && (
          <Alert variant="warning" className="mt-4">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Username change limit reached</AlertTitle>
            <AlertDescription>
              {eligibility.message}
            </AlertDescription>
          </Alert>
        )}

        {eligibility && eligibility.canChangeWithFee && !eligibility.canChangeFreely && (
          <Alert variant="default" className="mt-4 border-amber-500/50 bg-amber-500/5">
            <DollarSign className="h-4 w-4 text-amber-600" />
            <AlertTitle>Early change fee applies</AlertTitle>
            <AlertDescription>
              {eligibility.message} You can pay {formatCurrency(eligibility.earlyChangeFeeCents)} to change now.
            </AlertDescription>
          </Alert>
        )}
      </div>

      {/* Confirmation Dialog */}
      <Dialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <User className="w-5 h-5" />
              Change Username
            </DialogTitle>
            <DialogDescription>
              You are about to change your username from <strong>{currentUsername}</strong> to{' '}
              <strong>{newUsername}</strong>.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {eligibility?.canChangeFreely ? (
              <Alert className="border-green-500/50 bg-green-500/5">
                <Check className="h-4 w-4 text-green-600" />
                <AlertTitle>Free change available</AlertTitle>
                <AlertDescription>
                  You have {eligibility.changesRemaining} free change{eligibility.changesRemaining !== 1 ? 's' : ''}{' '}
                  remaining this year.
                </AlertDescription>
              </Alert>
            ) : eligibility?.canChangeWithFee ? (
              <>
                <Alert variant="warning">
                  <DollarSign className="h-4 w-4" />
                  <AlertTitle>Fee required</AlertTitle>
                  <AlertDescription>
                    You have used your 2 free changes. A fee of{' '}
                    {formatCurrency(eligibility.earlyChangeFeeCents)} will be charged to change now.
                  </AlertDescription>
                </Alert>
                <div className="rounded-lg border border-border-default p-4">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Early change fee</span>
                    <span className="font-medium">{formatCurrency(eligibility.earlyChangeFeeCents)}</span>
                  </div>
                  <p className="text-xs text-text-muted mt-2">
                    Payment processing will be handled via Stripe. You will be redirected to complete payment.
                  </p>
                </div>
              </>
            ) : (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Cannot change username</AlertTitle>
                <AlertDescription>{eligibility?.message}</AlertDescription>
              </Alert>
            )}

            <div className="rounded-lg border border-border-default p-4 bg-bg-secondary/50">
              <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                <Info className="w-4 h-4" />
                Important Notes
              </h4>
              <ul className="text-sm text-text-secondary space-y-1 list-disc list-inside">
                <li>Your old username will be reserved for 30 days</li>
                <li>External links to your profile may break</li>
                <li>This action is logged for security purposes</li>
                <li>You can change your username 2 times per year for free</li>
              </ul>
            </div>
          </div>

          <DialogFooter className="flex-col sm:flex-row gap-2">
            <Button variant="outline" onClick={() => setShowConfirmDialog(false)} disabled={isChanging}>
              Cancel
            </Button>
            {eligibility?.canChangeFreely ? (
              <Button onClick={() => handleConfirmChange(false)} disabled={isChanging}>
                {isChanging ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Changing...
                  </>
                ) : (
                  'Confirm Free Change'
                )}
              </Button>
            ) : eligibility?.canChangeWithFee ? (
              <Button onClick={() => handleConfirmChange(true)} disabled={isChanging} variant="default">
                {isChanging ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Processing...
                  </>
                ) : (
                  <>
                    <DollarSign className="w-4 h-4 mr-2" />
                    Pay & Change
                  </>
                )}
              </Button>
            ) : null}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* History Dialog */}
      <Dialog open={showHistoryDialog} onOpenChange={setShowHistoryDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <History className="w-5 h-5" />
              Username Change History
            </DialogTitle>
            <DialogDescription>Your previous username changes</DialogDescription>
          </DialogHeader>

          <div className="py-4">
            {isLoadingHistory ? (
              <div className="flex justify-center py-8">
                <Loader2 className="w-6 h-6 animate-spin text-text-muted" />
              </div>
            ) : historyData?.history && historyData.history.length > 0 ? (
              <div className="space-y-3">
                {historyData.history.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center justify-between p-3 rounded-lg border border-border-default bg-bg-secondary/50"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">{item.old_username}</span>
                        <span className="text-text-muted">→</span>
                        <span className="text-sm font-medium text-brand-400">{item.new_username}</span>
                      </div>
                      <p className="text-xs text-text-muted">
                        {new Date(item.changed_at).toLocaleDateString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          year: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </p>
                    </div>
                    {item.was_early_change && (
                      <Badge variant="outline" className="border-amber-500/50 text-amber-600">
                        <DollarSign className="w-3 h-3 mr-1" />
                        {formatCurrency(item.fee_paid_cents)}
                      </Badge>
                    )}
                    {!item.was_early_change && (
                      <Badge variant="outline" className="border-green-500/50 text-green-600">
                        <Check className="w-3 h-3 mr-1" />
                        Free
                      </Badge>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-text-muted">
                <User className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>No username changes yet</p>
                <p className="text-sm">Your username change history will appear here</p>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowHistoryDialog(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
