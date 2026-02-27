import { useState, useEffect } from 'react';
import { Button } from "@/components/ui/button";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { OTPInput } from './OTPInput';
import { Shield, Smartphone } from 'lucide-react';
import { apiClient } from '@/api/client';

interface MFASetupProps {
  email: string;
  onComplete: (secret: string) => void;
  onSkip: () => void;
}

interface MFASetupData {
  secret: string;
  qr_code_url: string;
  backup_codes: string[];
}

export function MFASetup({ onComplete, onSkip }: MFASetupProps) {
  const [step, setStep] = useState<'setup' | 'verify'>('setup');
  const [setupData, setSetupData] = useState<MFASetupData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch TOTP secret from backend on mount — never generate secrets client-side
  useEffect(() => {
    const initMFA = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const data = await apiClient.post<MFASetupData>('/v1/auth/mfa/setup', {});
        setSetupData(data);
      } catch (err) {
        setError('Failed to initialize MFA setup. Please try again.');
      } finally {
        setIsLoading(false);
      }
    };

    initMFA();
  }, []);

  const handleSetupComplete = () => {
    setStep('verify');
  };

  const handleOTPComplete = async (otp: string) => {
    if (!setupData) return;

    setIsLoading(true);
    setError(null);

    try {
      await apiClient.post('/v1/auth/mfa/verify', { code: otp });
      await apiClient.post('/v1/auth/mfa/enable', {});
      onComplete(setupData.secret);
    } catch (err) {
      setError('Invalid verification code. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading && !setupData) {
    return (
      <div className="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
    );
  }

  if (step === 'setup') {
    return (
      <div className="space-y-6">
        <div className="text-center">
          <Shield className="w-12 h-12 mx-auto text-[#6366f1] mb-4" />
          <h3 className="text-lg font-semibold">Set Up Two-Factor Authentication</h3>
          <p className="text-sm text-text-muted mt-1">
            Add an extra layer of security to your account
          </p>
        </div>

        <div className="space-y-4">
          <div className="text-center">
            <Smartphone className="w-8 h-8 mx-auto text-text-secondary mb-2" />
            <p className="text-sm text-text-secondary">
              Scan this QR code with your authenticator app
            </p>
          </div>

          <div className="flex justify-center">
            {setupData?.qr_code_url ? (
              <img
                src={setupData.qr_code_url}
                alt="MFA QR code"
                className="w-48 h-48 border border-input rounded-md"
              />
            ) : (
              <div className="w-48 h-48 border border-input rounded-md flex items-center justify-center bg-muted">
                <div className="text-center text-sm text-muted-foreground">
                  <Smartphone className="w-8 h-8 mx-auto mb-2" />
                  <p>QR Code unavailable</p>
                  <p className="text-xs mt-1">Use manual entry below</p>
                </div>
              </div>
            )}
          </div>

          {setupData?.secret && (
            <div className="text-center space-y-2">
              <p className="text-xs text-text-muted">Or enter this code manually:</p>
              <code className="text-xs bg-muted px-2 py-1 rounded font-mono break-all">
                {setupData.secret}
              </code>
            </div>
          )}

          {error && (
            <p className="text-sm text-destructive text-center">{error}</p>
          )}
        </div>

        <div className="flex space-x-3">
          <Button
            type="button"
            variant="outline"
            onClick={onSkip}
            className="flex-1"
          >
            Skip for Now
          </Button>
          <Button
            type="button"
            onClick={handleSetupComplete}
            disabled={!setupData}
            className="flex-1"
          >
            Continue
          </Button>
        </div>
      </div>
    );
  }

  return (
    <OTPInput
      length={6}
      onComplete={handleOTPComplete}
      error={error || undefined}
      isLoading={isLoading}
      title="Verify Your Authenticator"
      description="Enter the 6-digit code from your authenticator app"
    />
  );
}
