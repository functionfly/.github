import React, { useCallback } from "react";
import TurnstileWidget from "./TurnstileWidget";

interface ClientTurnstileProps {
  action?: string;
  theme?: "light" | "dark" | "auto";
  inputId?: string;
}

/**
 * Client-side wrapper for TurnstileWidget that handles the onVerify callback
 * internally. This avoids the serialization issue when passing functions from
 * Astro templates via client:only="react".
 */
export const ClientTurnstile: React.FC<ClientTurnstileProps> = ({
  action = "auth",
  theme = "dark",
  inputId = "turnstile-token",
}) => {
  const handleVerify = useCallback(
    (token: string) => {
      const hiddenInput = document.getElementById(inputId) as HTMLInputElement | null;
      if (hiddenInput) {
        hiddenInput.value = token;
      }
    },
    [inputId]
  );

  const handleError = useCallback(() => {
    // Clear the token on error
    const hiddenInput = document.getElementById(inputId) as HTMLInputElement | null;
    if (hiddenInput) {
      hiddenInput.value = "";
    }
  }, [inputId]);

  const handleExpire = useCallback(() => {
    // Clear the token when it expires
    const hiddenInput = document.getElementById(inputId) as HTMLInputElement | null;
    if (hiddenInput) {
      hiddenInput.value = "";
    }
  }, [inputId]);

  return (
    <TurnstileWidget
      onVerify={handleVerify}
      onError={handleError}
      onExpire={handleExpire}
      action={action}
      theme={theme}
    />
  );
};

export default ClientTurnstile;
