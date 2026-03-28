import { authApi, type MFASetupResponse, type MFAStatusResponse } from '@/api/auth';
import { usersApi, type SessionItem } from '@/api/users';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuthStore } from '@/stores/authStore';
import { Icon } from '@iconify/react';
import {
  AlertTriangle,
  CheckCircle2,
  Download,
  KeyRound,
  Monitor,
  ShieldAlert,
  ShieldCheck,
  Smartphone,
  Trash2,
} from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';
import { useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';

function detectCurrentDeviceLabel(): string {
  if (typeof navigator === 'undefined') return 'Current browser';
  const ua = navigator.userAgent;
  if (ua.includes('Edg/')) return 'Edge (this device)';
  if (ua.includes('Chrome/') && !ua.includes('Edg/')) return 'Chrome (this device)';
  if (ua.includes('Firefox/')) return 'Firefox (this device)';
  if (ua.includes('Safari/') && !ua.includes('Chrome/')) return 'Safari (this device)';
  if (ua.includes('Windows')) return 'Desktop (Windows)';
  if (ua.includes('Mac')) return 'Desktop (macOS)';
  if (ua.includes('Linux')) return 'Desktop (Linux)';
  if (ua.includes('Android') || ua.includes('iPhone')) return 'Mobile (this device)';
  return 'Current browser';
}

function getSessionSimpleIcon(deviceLabel: string): string | null {
  const normalized = deviceLabel.toLowerCase();
  if (normalized.includes('chrome')) return 'simple-icons:googlechrome';
  if (normalized.includes('firefox')) return 'simple-icons:firefoxbrowser';
  if (normalized.includes('safari')) return 'simple-icons:safari';
  if (normalized.includes('edge')) return 'simple-icons:microsoftedge';
  if (normalized.includes('opera')) return 'simple-icons:opera';
  if (normalized.includes('brave')) return 'simple-icons:brave';
  if (
    normalized.includes('mobile') ||
    normalized.includes('iphone') ||
    normalized.includes('android')
  ) {
    return null;
  }
  return null;
}

export function SecuritySettingsTab() {
  const [status, setStatus] = useState<MFAStatusResponse | null>(null);
  const [loadingStatus, setLoadingStatus] = useState(true);
  const [working, setWorking] = useState(false);

  const [setupData, setSetupData] = useState<MFASetupResponse | null>(null);
  const [verificationCode, setVerificationCode] = useState('');
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
  const [sessions, setSessions] = useState<SessionItem[]>([]);
  const [loadingSessions, setLoadingSessions] = useState(true);
  const [revokingSessionId, setRevokingSessionId] = useState<string | null>(null);
  const [revokingOthers, setRevokingOthers] = useState(false);
  const [exportingData, setExportingData] = useState(false);
  const [deleteConfirmation, setDeleteConfirmation] = useState('');
  const [deletingAccount, setDeletingAccount] = useState(false);
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  const [disablePassword, setDisablePassword] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [showMfaSecret, setShowMfaSecret] = useState(false);

  const qrCodeValue = useMemo(() => {
    return setupData?.qr_code_url ?? '';
  }, [setupData?.qr_code_url]);

  const loadStatus = async () => {
    setLoadingStatus(true);
    try {
      const nextStatus = await authApi.getMFAStatus();
      setStatus(nextStatus);
    } catch {
      toast.error('Failed to load MFA status');
    } finally {
      setLoadingStatus(false);
    }
  };

  useEffect(() => {
    void loadStatus();
  }, []);

  const loadSessions = async () => {
    setLoadingSessions(true);
    try {
      const data = await usersApi.listSessions();
      setSessions(data.sessions ?? []);
    } catch {
      toast.error('Failed to load active sessions');
    } finally {
      setLoadingSessions(false);
    }
  };

  useEffect(() => {
    void loadSessions();
  }, []);

  const handleStartSetup = async () => {
    setWorking(true);
    try {
      const data = await authApi.setupMFA();
      setSetupData(data);
      setRecoveryCodes(data.backup_codes ?? []);
      setVerificationCode('');
      setShowMfaSecret(false);
      toast.success('MFA setup started. Verify with your authenticator code.');
    } catch {
      toast.error('Failed to start MFA setup');
    } finally {
      setWorking(false);
    }
  };

  const handleEnableMFA = async () => {
    if (!verificationCode.trim()) {
      toast.error('Enter the 6-digit code from your authenticator app');
      return;
    }
    setWorking(true);
    try {
      const verify = await authApi.verifyMFASetupCode(verificationCode.trim());
      if (!verify.verified) {
        toast.error('That verification code is invalid');
        return;
      }
      await authApi.enableMFA();
      toast.success('MFA enabled');
      setSetupData(null);
      setVerificationCode('');
      await loadStatus();
    } catch {
      toast.error('Failed to enable MFA');
    } finally {
      setWorking(false);
    }
  };

  const handleDisableMFA = async () => {
    if (!disablePassword.trim() || !disableCode.trim()) {
      toast.error('Password and MFA code are required');
      return;
    }
    setWorking(true);
    try {
      await authApi.disableMFA(disablePassword.trim(), disableCode.trim());
      toast.success('MFA disabled');
      setDisablePassword('');
      setDisableCode('');
      await loadStatus();
    } catch {
      toast.error('Failed to disable MFA. Check your password and authenticator code.');
    } finally {
      setWorking(false);
    }
  };

  const handleRevokeSession = async (sessionId: string) => {
    setRevokingSessionId(sessionId);
    try {
      await usersApi.revokeSession(sessionId);
      toast.success('Session revoked');
      await loadSessions();
    } catch {
      toast.error('Failed to revoke session');
    } finally {
      setRevokingSessionId(null);
    }
  };

  const handleRevokeOthers = async () => {
    setRevokingOthers(true);
    try {
      await usersApi.revokeOtherSessions();
      toast.success('Other sessions revoked');
      await loadSessions();
    } catch {
      toast.error('Failed to revoke other sessions');
    } finally {
      setRevokingOthers(false);
    }
  };

  const handleExportData = async () => {
    setExportingData(true);
    try {
      const [me, settings, activeSessions] = await Promise.all([
        usersApi.getMe(),
        usersApi.getMySettings(),
        usersApi.listSessions(),
      ]);

      const payload = {
        exportedAt: new Date().toISOString(),
        account: me,
        settings: settings.settings ?? {},
        sessions: activeSessions.sessions ?? [],
      };

      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      const date = new Date().toISOString().slice(0, 10);
      const username = user?.username || 'account';
      anchor.href = url;
      anchor.download = `functionfly-${username}-data-export-${date}.json`;
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);
      URL.revokeObjectURL(url);
      toast.success('Your data export is ready');
    } catch {
      toast.error('Failed to export data');
    } finally {
      setExportingData(false);
    }
  };

  const handleDeleteAccount = async () => {
    if (deleteConfirmation !== 'DELETE') {
      toast.error('Type DELETE to confirm');
      return;
    }

    setDeletingAccount(true);
    try {
      await usersApi.deleteMe();
      toast.success('Account deleted');
      await logout();
      window.location.href = '/';
      return;
    } catch {
      const subject = encodeURIComponent('Account deletion request');
      const body = encodeURIComponent(
        `Please delete my FunctionFly account.\n\nUsername: ${user?.username || 'unknown'}\nEmail: ${
          user?.email || 'unknown'
        }\n\nI understand this action is permanent.`
      );
      window.location.href = `mailto:support@functionfly.com?subject=${subject}&body=${body}`;
      toast.info('Direct delete is unavailable. We opened a deletion request email to support.');
    } finally {
      setDeletingAccount(false);
      setDeleteConfirmation('');
    }
  };

  const handleDownloadRecoveryCodes = async () => {
    if (recoveryCodes.length === 0) return;

    const username = user?.username || 'account';
    const date = new Date().toISOString().slice(0, 10);
    const content = [
      'FunctionFly MFA Recovery Codes',
      '',
      'Each code can be used once.',
      `GeneratedAt: ${new Date().toISOString()}`,
      `Username: ${username}`,
      '',
      recoveryCodes.join('\n'),
      '',
    ].join('\n');

    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `functionfly-${username}-mfa-recovery-codes-${date}.txt`;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);

    toast.success('Recovery codes downloaded');
  };

  const displaySessions: SessionItem[] = useMemo(() => {
    if (sessions.length > 0) return sessions;
    return [
      {
        id: 'current-browser-fallback',
        device: detectCurrentDeviceLabel(),
        ip: 'Current network',
        location: '',
        lastActive: 'Now',
        currentSession: true,
      },
    ];
  }, [sessions]);

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-brand-500" />
            Multi-Factor Authentication
          </CardTitle>
          <CardDescription className="text-text-secondary">
            Add an authenticator app as a second factor to protect your account.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {loadingStatus ? (
            <p className="text-sm text-text-muted">Loading security status...</p>
          ) : (
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant={status?.enabled ? 'default' : 'secondary'}>
                {status?.enabled ? 'MFA enabled' : 'MFA disabled'}
              </Badge>
              {status?.required ? (
                <Badge
                  variant="outline"
                  className="border-amber-500/50 text-amber-600 dark:text-amber-400"
                >
                  Required for your role
                </Badge>
              ) : null}
              {typeof status?.backup_codes_remaining === 'number' ? (
                <Badge variant="outline">Backup codes: {status.backup_codes_remaining}</Badge>
              ) : null}
            </div>
          )}

          {!status?.enabled ? (
            <div className="rounded-lg border border-border-default bg-bg-secondary p-4 space-y-3">
              <p className="text-sm text-text-secondary">
                Enable MFA to require a one-time code from your authenticator app at login.
              </p>
              {!setupData ? (
                <Button onClick={handleStartSetup} disabled={working || loadingStatus}>
                  {working ? 'Starting...' : 'Set Up Authenticator App'}
                </Button>
              ) : (
                <div className="space-y-4">
                  <div className="flex flex-col gap-4 md:flex-row md:items-start">
                    <div className="rounded-lg border border-border-default bg-bg-primary p-3">
                      {qrCodeValue ? (
                        <QRCodeSVG
                          value={qrCodeValue}
                          size={180}
                          bgColor="transparent"
                          fgColor="var(--text-primary)"
                          level="M"
                        />
                      ) : (
                        <div className="h-[180px] w-[180px] flex items-center justify-center text-xs text-text-muted">
                          Loading QR...
                        </div>
                      )}
                    </div>
                    <div className="space-y-2">
                      <p className="text-sm text-text-secondary">
                        Scan this QR code with Google Authenticator, 1Password, Authy, or similar.
                      </p>
                      <div className="flex items-center gap-2 flex-wrap">
                        <p className="text-xs text-text-muted">Manual key</p>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => setShowMfaSecret((v) => !v)}
                          className="h-7 px-2 text-xs"
                        >
                          {showMfaSecret ? 'Hide' : 'Show'}
                        </Button>
                      </div>
                      {showMfaSecret ? (
                        <code className="inline-block break-all rounded bg-bg-primary px-2 py-1 text-xs">
                          {setupData.secret}
                        </code>
                      ) : (
                        <code className="inline-block break-all rounded bg-bg-primary px-2 py-1 text-xs text-text-muted">
                          Secret hidden
                        </code>
                      )}
                    </div>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mfa-verify-code">Verification code</Label>
                    <Input
                      id="mfa-verify-code"
                      placeholder="123456"
                      inputMode="numeric"
                      value={verificationCode}
                      onChange={(e) =>
                        setVerificationCode(e.target.value.replace(/[^\d]/g, '').slice(0, 6))
                      }
                    />
                  </div>
                  <div className="flex gap-2">
                    <Button onClick={handleEnableMFA} disabled={working}>
                      {working ? 'Enabling...' : 'Enable MFA'}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => {
                        setSetupData(null);
                        setVerificationCode('');
                      }}
                      disabled={working}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="rounded-lg border border-border-default bg-bg-secondary p-4 space-y-4">
              <div className="flex items-start gap-2 text-sm text-text-secondary">
                <CheckCircle2 className="mt-0.5 h-4 w-4 text-green-500" />
                <p>Your account is protected with MFA.</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="disable-password">Current password</Label>
                <Input
                  id="disable-password"
                  type="password"
                  value={disablePassword}
                  onChange={(e) => setDisablePassword(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="disable-code">Authenticator code</Label>
                <Input
                  id="disable-code"
                  inputMode="numeric"
                  placeholder="123456"
                  value={disableCode}
                  onChange={(e) => setDisableCode(e.target.value.replace(/[^\d]/g, '').slice(0, 6))}
                />
              </div>
              <Button
                variant="destructive"
                onClick={handleDisableMFA}
                disabled={working || status?.required}
              >
                {working ? 'Disabling...' : 'Disable MFA'}
              </Button>
              {status?.required ? (
                <p className="flex items-center gap-2 text-xs text-amber-600 dark:text-amber-400">
                  <AlertTriangle className="h-3.5 w-3.5" />
                  MFA is required for your role and cannot be disabled.
                </p>
              ) : null}
            </div>
          )}
        </CardContent>
      </Card>

      {recoveryCodes.length > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-brand-500" />
              Recovery Codes
            </CardTitle>
            <CardDescription className="text-text-secondary">
              Save these one-time backup codes somewhere safe. Each code can be used once.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="mb-3 flex items-center justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => void handleDownloadRecoveryCodes()}
              >
                Download Codes
              </Button>
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              {recoveryCodes.map((code) => (
                <code
                  key={code}
                  className="rounded border border-border-default bg-bg-secondary px-3 py-2 text-xs"
                >
                  {code}
                </code>
              ))}
            </div>
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Smartphone className="h-5 w-5 text-brand-500" />
            Session Security
          </CardTitle>
          <CardDescription className="text-text-secondary">
            Review active devices and revoke sessions you do not recognize.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              onClick={() => void loadSessions()}
              disabled={loadingSessions}
            >
              {loadingSessions ? 'Refreshing...' : 'Refresh Sessions'}
            </Button>
            <Button
              variant="destructive"
              onClick={handleRevokeOthers}
              disabled={revokingOthers || loadingSessions}
            >
              {revokingOthers ? 'Signing out...' : 'Sign Out Other Devices'}
            </Button>
          </div>

          {loadingSessions ? (
            <p className="text-sm text-text-muted">Loading active sessions...</p>
          ) : (
            <div className="space-y-2">
              {displaySessions.map((session) => (
                <div
                  key={session.id}
                  className="flex flex-col gap-3 rounded-lg border border-border-default bg-bg-secondary p-3 md:flex-row md:items-center md:justify-between"
                >
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      {(() => {
                        const iconName = getSessionSimpleIcon(session.device);
                        if (iconName) {
                          return <Icon icon={iconName} className="h-4 w-4 shrink-0" />;
                        }
                        const normalized = session.device.toLowerCase();
                        if (
                          normalized.includes('desktop') ||
                          normalized.includes('mac') ||
                          normalized.includes('windows') ||
                          normalized.includes('linux')
                        ) {
                          return <Monitor className="h-4 w-4 text-text-muted shrink-0" />;
                        }
                        return <Smartphone className="h-4 w-4 text-text-muted shrink-0" />;
                      })()}
                      <p className="text-sm font-medium text-text-primary truncate">
                        {session.device}
                      </p>
                      {session.currentSession ? (
                        <Badge
                          variant="outline"
                          className="text-green-600 dark:text-green-400 border-green-500/40"
                        >
                          Current
                        </Badge>
                      ) : null}
                    </div>
                    <p className="text-xs text-text-muted truncate">
                      {session.ip} {session.location ? `- ${session.location}` : ''}
                    </p>
                    <p className="text-xs text-text-muted">Last active: {session.lastActive}</p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleRevokeSession(session.id)}
                    disabled={
                      session.currentSession ||
                      session.id === 'current-browser-fallback' ||
                      revokingSessionId === session.id
                    }
                  >
                    {revokingSessionId === session.id ? 'Revoking...' : 'Revoke'}
                  </Button>
                </div>
              ))}
            </div>
          )}
          {!loadingSessions && sessions.length === 0 ? (
            <p className="text-xs text-text-muted">
              We are showing your current browser session. Historical sessions may not be available
              in this environment.
            </p>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Download className="h-5 w-5 text-brand-500" />
            Data Export
          </CardTitle>
          <CardDescription className="text-text-secondary">
            Download a JSON archive of your account profile, settings, and active sessions.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-text-secondary">
            This export includes your account metadata and preferences for backup and compliance
            requests.
          </p>
          <Button onClick={handleExportData} disabled={exportingData}>
            {exportingData ? 'Preparing export...' : 'Export My Data'}
          </Button>
        </CardContent>
      </Card>

      <Card className="border-red-500/30">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-red-600 dark:text-red-400">
            <ShieldAlert className="h-5 w-5" />
            Danger Zone
          </CardTitle>
          <CardDescription className="text-text-secondary">
            Permanent and destructive actions for your account.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="text-sm text-text-secondary">
            Deleting your account removes access and cannot be undone.
          </p>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button variant="destructive">
                <Trash2 className="mr-2 h-4 w-4" />
                Delete Account
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete your account?</AlertDialogTitle>
                <AlertDialogDescription>
                  This action is permanent. Type <strong>DELETE</strong> to confirm.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <div className="space-y-2">
                <Label htmlFor="delete-confirmation">Confirmation</Label>
                <Input
                  id="delete-confirmation"
                  placeholder="Type DELETE"
                  value={deleteConfirmation}
                  onChange={(e) => setDeleteConfirmation(e.target.value)}
                />
              </div>
              <AlertDialogFooter>
                <AlertDialogCancel onClick={() => setDeleteConfirmation('')}>
                  Cancel
                </AlertDialogCancel>
                <AlertDialogAction
                  className="bg-red-600 hover:bg-red-700"
                  onClick={(e) => {
                    e.preventDefault();
                    void handleDeleteAccount();
                  }}
                  disabled={deletingAccount || deleteConfirmation !== 'DELETE'}
                >
                  {deletingAccount ? 'Deleting...' : 'Permanently Delete'}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </CardContent>
      </Card>
    </div>
  );
}
