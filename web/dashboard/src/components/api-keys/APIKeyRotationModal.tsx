import { useState } from "react";
import { Copy, Check, AlertTriangle, RefreshCw } from "lucide-react";
import { Modal, SealedButton, FrameButton } from "@/components/containment";
import {
  APIKey,
  RotationReason,
  ROTATION_REASON_LABELS,
} from "@/types/api-key";
import { apiKeysService } from "@/services/api-keys";
import styles from "./APIKeyRotationModal.module.css";

interface APIKeyRotationModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  apiKey: APIKey | null;
  onSuccess?: (newKey: string) => void;
}

export function APIKeyRotationModal({
  open,
  onOpenChange,
  apiKey,
  onSuccess,
}: APIKeyRotationModalProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [showKey, setShowKey] = useState<string | null>(null);
  const [reason, setReason] = useState<RotationReason>("manual");
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleRotate = async () => {
    if (!apiKey) return;

    setError(null);
    setIsLoading(true);
    try {
      const response = await apiKeysService.rotateKey(apiKey.id, { reason });
      setShowKey(response.plaintext);
      onSuccess?.(response.plaintext);
    } catch (err) {
      console.error("Failed to rotate API key:", err);
      setError(err instanceof Error ? err.message : "Failed to rotate API key");
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setShowKey(null);
    setReason("manual");
    setError(null);
    setCopied(false);
    onOpenChange(false);
  };

  const copyToClipboard = async () => {
    if (showKey) {
      await navigator.clipboard.writeText(showKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (showKey) {
    return (
      <Modal open={open} onClose={handleClose} title="API Key Rotated">
        <div className={styles.content}>
          <div className={styles.keyDisplay}>
            <div className={styles.keyHeader}>
              <span className={styles.keyLabel}>New API Key</span>
              <button
                type="button"
                className={styles.copyButton}
                onClick={copyToClipboard}
                aria-label="Copy API key"
              >
                {copied ? (
                  <Check size={14} className={styles.copySuccess} />
                ) : (
                  <Copy size={14} />
                )}
                <span>{copied ? "Copied" : "Copy"}</span>
              </button>
            </div>
            <code className={styles.keyValue}>{showKey}</code>
          </div>

          <div className={styles.warningBox}>
            <AlertTriangle size={16} className={styles.warningIcon} />
            <p className={styles.warningText}>
              <strong>Important:</strong> The old key is now invalid. Update any
              applications using the old key with the new one.
            </p>
          </div>

          <div className={styles.footer}>
            <SealedButton onClick={handleClose}>Done</SealedButton>
          </div>
        </div>
      </Modal>
    );
  }

  return (
    <Modal open={open} onClose={handleClose} title="Rotate API Key">
      <div className={styles.content}>
        <p className={styles.description}>
          Rotate the API key <strong>"{apiKey?.name}"</strong>. This will
          invalidate the current key and create a new one.
        </p>

        {error && (
          <div className={styles.errorBox} role="alert">
            <p>{error}</p>
          </div>
        )}

        <div className={styles.field}>
          <label htmlFor="reason" className={styles.label}>
            Reason for rotation
          </label>
          <div className={styles.selectWrapper}>
            <select
              id="reason"
              value={reason}
              onChange={(e) => setReason(e.target.value as RotationReason)}
              className={styles.select}
            >
              <option value="manual">Manual</option>
              <option value="automatic">Automatic</option>
              <option value="compromised">Compromised</option>
            </select>
            <div className={styles.selectChevron}>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="m6 9 6 6 6-6" />
              </svg>
            </div>
          </div>
          <p className={styles.hint}>{ROTATION_REASON_LABELS[reason]}</p>
        </div>

        <div className={styles.warningBox}>
          <AlertTriangle size={16} className={styles.warningIcon} />
          <p className={styles.warningText}>
            <strong>Warning:</strong> Any applications using the current key will
            stop working until updated with the new key.
          </p>
        </div>

        <div className={styles.footer}>
          <FrameButton onClick={() => onOpenChange(false)}>Cancel</FrameButton>
          <SealedButton
            onClick={handleRotate}
            loading={isLoading}
            iconLeft={<RefreshCw size={14} />}
          >
            {isLoading ? "Rotating..." : "Rotate Key"}
          </SealedButton>
        </div>
      </div>
    </Modal>
  );
}
