import { useState, useEffect } from 'react';
import { Button } from "@/components/ui/button";
// import { FormError } from "@/components/ui/form-error"; // Not used in this component
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { OTPInput } from './OTPInput';
import { Shield, Smartphone } from 'lucide-react';

// Note: In production, generate TOTP secret on the server
// This is for demonstration purposes only
const generateTOTPSecret = () => {
  // Generate a random 32-character base32 secret
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  let secret = '';
  for (let i = 0; i < 32; i++) {
    secret += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return secret;
};

// const generateTOTPUrl = (secret: string, email: string, issuer: string = 'FunctionFly') => {
//   return `otpauth://totp/${issuer}:${email}?secret=${secret}&issuer=${issuer}`;
// };

interface MFASetupProps {
  email: string;
  onComplete: (secret: string) => void;
  onSkip: () => void;
}

export function MFASetup({ onComplete, onSkip }: MFASetupProps) {
  const [step, setStep] = useState<'setup' | 'verify'>('setup');
  const [secret, setSecret] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Generate secret on mount
  useEffect(() => {
    const newSecret = generateTOTPSecret();
    setSecret(newSecret);
  }, []);

  const handleSetupComplete = async () => {
    setStep('verify');
  };

  const handleOTPComplete = async (_otp: string) => {
    setIsLoading(true);
    setError(null);

    try {
      // TODO: Verify TOTP code with your backend
      // const response = await fetch('/api/auth/mfa/verify', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ secret, code: otp }),
      // });

      // if (!response.ok) throw new Error('Invalid code');

      onComplete(secret);

      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 1000));
    } catch (err) {
      setError('Invalid verification code. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

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
            {secret ? (
              <div className="w-48 h-48 border border-input rounded-md flex items-center justify-center bg-muted">
                <div className="text-center text-sm text-muted-foreground">
                  <Smartphone className="w-8 h-8 mx-auto mb-2" />
                  <p>QR Code will be displayed here</p>
                  <p className="text-xs mt-1">Use manual entry below</p>
                </div>
              </div>
            ) : (
              <div className="w-48 h-48 border border-input rounded-md flex items-center justify-center">
                <LoadingSpinner />
              </div>
            )}
          </div>

          <div className="text-center space-y-2">
            <p className="text-xs text-text-muted">Or enter this code manually:</p>
            <code className="text-xs bg-muted px-2 py-1 rounded font-mono break-all">
              {secret}
            </code>
          </div>
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