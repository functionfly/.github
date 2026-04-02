import type { SignupCaptchaPublic } from '@/api/signup-config';
import { Turnstile } from '@marsidev/react-turnstile';
import { useEffect, useRef } from 'react';
import { useGoogleReCaptcha } from 'react-google-recaptcha-v3';

function RecaptchaV3Binder({
  action,
  formValid,
  onToken,
}: {
  action: 'login' | 'signup';
  formValid: boolean;
  onToken: (token: string | null) => void;
}) {
  const { executeRecaptcha } = useGoogleReCaptcha();
  const onTokenRef = useRef(onToken);
  onTokenRef.current = onToken;

  useEffect(() => {
    if (!executeRecaptcha || !formValid) {
      onTokenRef.current(null);
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const t = await executeRecaptcha(action);
        if (!cancelled) onTokenRef.current(t ?? null);
      } catch {
        if (!cancelled) onTokenRef.current(null);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [executeRecaptcha, formValid, action]);

  return null;
}

/** Renders Turnstile or wires reCAPTCHA v3 when the API exposes a supported captcha config. */
export function AuthCaptchaWidgets({
  captcha,
  action,
  formValid,
  onToken,
}: {
  captcha: SignupCaptchaPublic | null | undefined;
  action: 'login' | 'signup';
  formValid: boolean;
  onToken: (token: string | null) => void;
}) {
  if (!captcha?.siteKey) return null;

  if (captcha.provider === 'turnstile') {
    return (
      <div className="flex justify-center py-1 min-h-[65px]">
        <Turnstile
          siteKey={captcha.siteKey}
          onSuccess={(token) => onToken(token)}
          onExpire={() => onToken(null)}
          onError={() => onToken(null)}
          options={{ appearance: 'interaction-only', action }}
        />
      </div>
    );
  }

  if (captcha.provider === 'recaptcha_v3') {
    return <RecaptchaV3Binder action={action} formValid={formValid} onToken={onToken} />;
  }

  return null;
}
