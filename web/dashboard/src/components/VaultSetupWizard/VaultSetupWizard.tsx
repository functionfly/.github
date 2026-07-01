import { useState, useCallback } from 'react';
import { Shield, Lock, KeyRound, CheckCircle2, Eye, EyeOff, ArrowRight, ArrowLeft, ExternalLink, Loader2, Shuffle, Copy, Check, Download } from 'lucide-react';
import { setupVaultPassphrase } from '@/services/vault-api-key-storage';
import { PasswordStrengthIndicator } from '@/components/common/PasswordStrengthIndicator';
import { getPublicDocsSiteOrigin } from '@/lib/constants';
import { toast } from 'sonner';
import './styles.css';

type WizardStep = 'welcome' | 'passphrase' | 'done';

interface VaultSetupWizardProps {
  open: boolean;
  onComplete: () => void;
}

const PASSPHRASE_CHARS = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789';
const PASSPHRASE_SYMBOLS = '!@#$%^&*+-=?';

function generatePassphrase(length: number = 24): string {
  const allChars = PASSPHRASE_CHARS + PASSPHRASE_SYMBOLS;
  const randomValues = crypto.getRandomValues(new Uint8Array(length));
  let result = '';
  for (let i = 0; i < length; i++) {
    result += allChars[randomValues[i] % allChars.length];
  }
  const classes = [
    'ABCDEFGHJKLMNPQRSTUVWXYZ',
    'abcdefghjkmnpqrstuvwxyz',
    '23456789',
    PASSPHRASE_SYMBOLS,
  ];
  const rng = crypto.getRandomValues(new Uint8Array(classes.length));
  classes.forEach((chars, i) => {
    const pos = (rng[i] & 0x7f) % result.length;
    const replacement = chars[rng[i] % chars.length];
    result = result.substring(0, pos) + replacement + result.substring(pos + 1);
  });
  return result;
}

const DOCS_PATH = '/guides/secrets-vault-guide/';

const stepMeta: Record<WizardStep, { step: number; label: string }> = {
  welcome: { step: 1, label: 'Welcome' },
  passphrase: { step: 2, label: 'Create Key' },
  done: { step: 3, label: 'All Set' },
};

export function VaultSetupWizard({ open, onComplete }: VaultSetupWizardProps) {
  const [step, setStep] = useState<WizardStep>('welcome');
  const [passphrase, setPassphrase] = useState('');
  const [confirmPassphrase, setConfirmPassphrase] = useState('');
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const docsUrl = `${getPublicDocsSiteOrigin()}${DOCS_PATH}`;
  const currentMeta = stepMeta[step];

  const resetForm = useCallback(() => {
    setPassphrase('');
    setConfirmPassphrase('');
    setShowPassphrase(false);
    setCopied(false);
    setError(null);
    setIsLoading(false);
  }, []);

  const handleGenerate = useCallback(() => {
    const generated = generatePassphrase(24);
    setPassphrase(generated);
    setConfirmPassphrase(generated);
    setShowPassphrase(true);
    setCopied(false);
    setError(null);
  }, []);

  const handleCopy = useCallback(async () => {
    if (!passphrase) return;
    try {
      await navigator.clipboard.writeText(passphrase);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // clipboard may be blocked
    }
  }, [passphrase]);

  const handleDownload = useCallback(() => {
    if (!passphrase) return;
    const blob = new Blob(
      ['FunctionFly Vault Secret Key\n', '================================\n', `${passphrase}\n`, '\nKeep this file in a secure location.\nDo NOT share it or store it in an unencrypted place.\n'],
      { type: 'text/plain' },
    );
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'functionfly-vault-key.txt';
    a.click();
    URL.revokeObjectURL(url);
  }, [passphrase]);

  const goTo = useCallback((next: WizardStep) => {
    setError(null);
    setStep(next);
  }, []);

  const handleCreatePassphrase = useCallback(async () => {
    setError(null);

    if (!passphrase) {
      setError('Please enter a passphrase');
      return;
    }
    if (passphrase.length < 12) {
      setError('Must be at least 12 characters');
      return;
    }
    if (!/[A-Z]/.test(passphrase) || !/[a-z]/.test(passphrase) || !/[0-9]/.test(passphrase)) {
      setError('Include uppercase, lowercase, and a number');
      return;
    }
    if (passphrase !== confirmPassphrase) {
      setError('Passphrases do not match');
      return;
    }

    setIsLoading(true);
    try {
      const result = await setupVaultPassphrase(passphrase);
      if (result.success) {
        toast.success('Vault is ready');
        goTo('done');
      } else {
        setError(result.error || 'Something went wrong');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setIsLoading(false);
    }
  }, [passphrase, confirmPassphrase, goTo]);

  const handleFinish = useCallback(() => {
    resetForm();
    onComplete();
  }, [resetForm, onComplete]);

  if (!open) return null;

  return (
    <div className="vault-wizard-overlay">
      <div className="vault-wizard" role="dialog" aria-label="Set up your Vault">
        {/* Step indicator */}
        <div className="vault-wizard__steps" aria-label="Setup progress">
          {(['welcome', 'passphrase', 'done'] as WizardStep[]).map((s, i) => {
            const meta = stepMeta[s];
            const isActive = s === step;
            const isPast = meta.step < currentMeta.step;
            return (
              <div key={s} className={`vault-wizard__step-dot ${isActive ? 'vault-wizard__step-dot--active' : ''} ${isPast ? 'vault-wizard__step-dot--done' : ''}`}>
                <span className="vault-wizard__step-number">{isPast ? '✓' : meta.step}</span>
                <span className="vault-wizard__step-label">{meta.label}</span>
              </div>
            );
          })}
        </div>

        {/* Step 1: Welcome */}
        {step === 'welcome' && (
          <div className="vault-wizard__body">
            <div className="vault-wizard__icon-ring">
              <Shield className="vault-wizard__icon" />
            </div>
            <h2 className="vault-wizard__title">Welcome to your Vault</h2>
            <p className="vault-wizard__desc">
              Your vault keeps passwords, API keys, and secrets safe using encryption that only <strong>you</strong> can unlock.
            </p>
            <div className="vault-wizard__card">
              <div className="vault-wizard__card-row">
                <Lock className="vault-wizard__card-icon" />
                <div>
                  <p className="vault-wizard__card-title">Locked by you</p>
                  <p className="vault-wizard__card-text">Only your secret key can open the vault. Not even we can peek inside.</p>
                </div>
              </div>
              <div className="vault-wizard__card-row">
                <KeyRound className="vault-wizard__card-icon" />
                <div>
                  <p className="vault-wizard__card-title">One key, one vault</p>
                  <p className="vault-wizard__card-text">You will create a passphrase next. Keep it somewhere safe — it cannot be recovered.</p>
                </div>
              </div>
            </div>
            <a
              href={docsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="vault-wizard__link"
            >
              How does zero-knowledge encryption work?
              <ExternalLink className="vault-wizard__link-icon" />
            </a>
          </div>
        )}

        {/* Step 2: Create Passphrase */}
        {step === 'passphrase' && (
          <div className="vault-wizard__body">
            <div className="vault-wizard__icon-ring">
              <KeyRound className="vault-wizard__icon" />
            </div>
            <h2 className="vault-wizard__title">Create your secret key</h2>
            <p className="vault-wizard__desc">
              This key locks and unlocks your vault. Choose something only you would know.
            </p>

            {error && (
              <div className="vault-wizard__error" role="alert">{error}</div>
            )}

            <div className="vault-wizard__field">
              <label htmlFor="vw-passphrase" className="vault-wizard__label">
                Secret key
              </label>
              <div className="vault-wizard__input-wrap">
                <input
                  id="vw-passphrase"
                  type={showPassphrase ? 'text' : 'password'}
                  autoComplete="new-password"
                  placeholder="At least 12 characters"
                  value={passphrase}
                  onChange={(e) => setPassphrase(e.target.value)}
                  className="vault-wizard__input"
                  disabled={isLoading}
                />
                <div className="vault-wizard__input-actions">
                  <button
                    type="button"
                    onClick={handleGenerate}
                    className="vault-wizard__action-btn"
                    aria-label="Generate strong key"
                    title="Generate strong key"
                    disabled={isLoading}
                  >
                    <Shuffle className="h-4 w-4" />
                  </button>
                  {passphrase && (
                    <>
                      <button
                        type="button"
                        onClick={handleCopy}
                        className="vault-wizard__action-btn"
                        aria-label={copied ? 'Copied' : 'Copy key'}
                        title={copied ? 'Copied' : 'Copy key'}
                      >
                        {copied ? <Check className="h-4 w-4 vault-wizard__action-check" /> : <Copy className="h-4 w-4" />}
                      </button>
                      <button
                        type="button"
                        onClick={handleDownload}
                        className="vault-wizard__action-btn"
                        aria-label="Download key"
                        title="Download key as file"
                      >
                        <Download className="h-4 w-4" />
                      </button>
                    </>
                  )}
                  <button
                    type="button"
                    onClick={() => setShowPassphrase(!showPassphrase)}
                    className="vault-wizard__action-btn"
                    aria-label={showPassphrase ? 'Hide key' : 'Show key'}
                  >
                    {showPassphrase ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>
              {passphrase && <PasswordStrengthIndicator password={passphrase} className="vault-wizard__strength" />}
            </div>

            <div className="vault-wizard__field">
              <label htmlFor="vw-confirm" className="vault-wizard__label">
                Confirm secret key
              </label>
              <input
                id="vw-confirm"
                type={showPassphrase ? 'text' : 'password'}
                autoComplete="new-password"
                placeholder="Type it again"
                value={confirmPassphrase}
                onChange={(e) => setConfirmPassphrase(e.target.value)}
                className="vault-wizard__input"
                disabled={isLoading}
              />
            </div>

            <div className="vault-wizard__warning">
              <Lock className="vault-wizard__warning-icon" />
              <p className="vault-wizard__warning-text">
                Save this key in a password manager. If you lose it, your secrets are gone forever.
              </p>
            </div>
          </div>
        )}

        {/* Step 3: Done */}
        {step === 'done' && (
          <div className="vault-wizard__body vault-wizard__body--center">
            <div className="vault-wizard__icon-ring vault-wizard__icon-ring--success">
              <CheckCircle2 className="vault-wizard__icon" />
            </div>
            <h2 className="vault-wizard__title">You're all set!</h2>
            <p className="vault-wizard__desc">
              Your vault is locked and ready. You can now store secrets securely.
            </p>
            <div className="vault-wizard__checklist">
              <div className="vault-wizard__check-item">
                <CheckCircle2 className="vault-wizard__check-icon" />
                <span>Secret key created</span>
              </div>
              <div className="vault-wizard__check-item">
                <CheckCircle2 className="vault-wizard__check-icon" />
                <span>Vault is locked</span>
              </div>
              <div className="vault-wizard__check-item">
                <CheckCircle2 className="vault-wizard__check-icon" />
                <span>Ready to store secrets</span>
              </div>
            </div>
            <a
              href={docsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="vault-wizard__link"
            >
              Learn more about vault security
              <ExternalLink className="vault-wizard__link-icon" />
            </a>
          </div>
        )}

        {/* Footer */}
        <div className="vault-wizard__footer">
          {step === 'welcome' && (
            <button className="vault-wizard__btn vault-wizard__btn--primary" onClick={() => goTo('passphrase')}>
              Get Started
              <ArrowRight className="vault-wizard__btn-icon" />
            </button>
          )}

          {step === 'passphrase' && (
            <>
              <button className="vault-wizard__btn vault-wizard__btn--ghost" onClick={() => goTo('welcome')} disabled={isLoading}>
                <ArrowLeft className="vault-wizard__btn-icon" />
                Back
              </button>
              <button className="vault-wizard__btn vault-wizard__btn--primary" onClick={handleCreatePassphrase} disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="vault-wizard__btn-icon vault-wizard__btn-icon--spin" />
                    Setting up...
                  </>
                ) : (
                  <>
                    Create Key
                    <Lock className="vault-wizard__btn-icon" />
                  </>
                )}
              </button>
            </>
          )}

          {step === 'done' && (
            <button className="vault-wizard__btn vault-wizard__btn--primary vault-wizard__btn--wide" onClick={handleFinish}>
              Open Vault
              <ArrowRight className="vault-wizard__btn-icon" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
