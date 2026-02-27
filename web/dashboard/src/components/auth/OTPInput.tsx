import React, { useState, useRef, useEffect } from 'react';
import { Button } from "@/components/ui/button";
import { FormError } from "@/components/ui/form-error";
import { LoadingSpinner } from "@/components/ui/loading-spinner";

interface OTPInputProps {
  length?: number;
  onComplete: (otp: string) => void;
  onResend?: () => void;
  error?: string;
  isLoading?: boolean;
  resendCooldown?: number;
  title?: string;
  description?: string;
}

export function OTPInput({
  length = 6,
  onComplete,
  onResend,
  error,
  isLoading = false,
  resendCooldown = 60,
  title = "Enter Verification Code",
  description = "We've sent a 6-digit code to your email"
}: OTPInputProps) {
  const [otp, setOtp] = useState<string[]>(new Array(length).fill(''));
  const [cooldown, setCooldown] = useState(0);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  // Handle cooldown timer
  useEffect(() => {
    if (cooldown > 0) {
      const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [cooldown]);

  // Handle resend with cooldown
  const handleResend = () => {
    if (cooldown === 0 && onResend) {
      onResend();
      setCooldown(resendCooldown);
    }
  };

  // Handle input change
  const handleChange = (index: number, value: string) => {
    if (isNaN(Number(value))) return;

    const newOtp = [...otp];
    newOtp[index] = value;
    setOtp(newOtp);

    // Auto-focus next input
    if (value && index < length - 1) {
      inputRefs.current[index + 1]?.focus();
    }

    // Check if OTP is complete
    const completeOtp = newOtp.join('');
    if (completeOtp.length === length) {
      onComplete(completeOtp);
    }
  };

  // Handle backspace
  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !otp[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  // Handle paste
  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pasteData = e.clipboardData.getData('text').slice(0, length);
    const newOtp = [...otp];

    for (let i = 0; i < pasteData.length; i++) {
      if (!isNaN(Number(pasteData[i]))) {
        newOtp[i] = pasteData[i];
      }
    }

    setOtp(newOtp);

    // Focus last filled input or next empty one
    const lastFilledIndex = newOtp.reduce((lastIndex, digit, index) => digit !== '' ? index : lastIndex, -1);
    const focusIndex = lastFilledIndex < length - 1 ? lastFilledIndex + 1 : lastFilledIndex;
    inputRefs.current[focusIndex]?.focus();

    // Check if OTP is complete after paste
    const completeOtp = newOtp.join('');
    if (completeOtp.length === length) {
      onComplete(completeOtp);
    }
  };

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h3 className="text-lg font-semibold">{title}</h3>
        <p className="text-sm text-text-muted mt-1">{description}</p>
      </div>

      <FormError error={error} />

      <div className="flex justify-center space-x-2">
        {otp.map((digit, index) => (
          <input
            key={index}
            ref={(el) => { inputRefs.current[index] = el; }}
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            maxLength={1}
            value={digit}
            onChange={(e) => handleChange(index, e.target.value)}
            onKeyDown={(e) => handleKeyDown(index, e)}
            onPaste={handlePaste}
            className="w-12 h-12 text-center text-xl font-semibold border border-input rounded-md focus:border-primary focus:ring-1 focus:ring-primary bg-background"
            disabled={isLoading}
          />
        ))}
      </div>

      <div className="text-center space-y-4">
        <div className="text-sm text-text-muted">
          Didn't receive the code?
        </div>

        <Button
          type="button"
          variant="outline"
          onClick={handleResend}
          disabled={cooldown > 0 || isLoading}
          className="text-sm"
        >
          {isLoading ? (
            <LoadingSpinner text="Sending..." />
          ) : cooldown > 0 ? (
            `Resend in ${cooldown}s`
          ) : (
            'Resend Code'
          )}
        </Button>
      </div>
    </div>
  );
}