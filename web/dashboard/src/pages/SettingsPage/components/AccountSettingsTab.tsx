import { usersApi } from '@/api/users';
import { MyFollowStats } from '@/components/follow';
import { UsernameChangeField } from '@/components/UsernameChangeField';
import { LanguagePicker } from '@/components/common/LanguagePicker';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useAuthStore } from '@/stores/authStore';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { Globe, Play, Users } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

export function AccountSettingsTab() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const nameParts = (user?.name || '').split(' ');
  const [firstName, setFirstName] = useState(nameParts[0] || '');
  const [lastName, setLastName] = useState(nameParts.slice(1).join(' ') || '');
  const [username, setUsername] = useState(user?.username || '');
  const [email] = useState(user?.email || '');
  const [dateOfBirth, setDateOfBirth] = useState('');
  const [isDobLocked, setIsDobLocked] = useState(false);
  const [isSavingProfile, setIsSavingProfile] = useState(false);

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isUpdatingPassword, setIsUpdatingPassword] = useState(false);

  useEffect(() => {
    let cancelled = false;
    usersApi
      .getMe()
      .then((me) => {
        if (cancelled) return;
        setDateOfBirth(me.dateOfBirth || '');
        setIsDobLocked(Boolean(me.dateOfBirth));
      })
      .catch(() => {
        // keep existing value if request fails
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const handleSaveProfile = async () => {
    setIsSavingProfile(true);
    try {
      const payload = {
        name: `${firstName} ${lastName}`.trim() || undefined,
        // Note: username is now handled separately via UsernameChangeField
        ...(isDobLocked ? {} : { dateOfBirth: dateOfBirth || null }),
      };

      await usersApi.updateMe(payload);
      await useAuthStore.getState().initialize();
      toast.success('Profile updated successfully');

      if (!isDobLocked && dateOfBirth) {
        setIsDobLocked(true);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update profile');
    } finally {
      setIsSavingProfile(false);
    }
  };

  const handleUpdatePassword = async () => {
    if (newPassword !== confirmPassword) {
      toast.error('New passwords do not match');
      return;
    }
    if (newPassword.length < 8) {
      toast.error('Password must be at least 8 characters');
      return;
    }
    setIsUpdatingPassword(true);
    try {
      await usersApi.changePassword({ currentPassword, newPassword });
      toast.success('Password updated successfully');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      toast.error(
        err instanceof Error
          ? err.message
          : 'Failed to update password. Check your current password.'
      );
    } finally {
      setIsUpdatingPassword(false);
    }
  };

  const { canResume, hasSkippedOnboarding, completedSteps } = useOnboardingStore();
  const showOnboardingResume = canResume() || (hasSkippedOnboarding && completedSteps.length === 0);

  const handleResumeOnboarding = () => {
    useOnboardingStore.setState({ hasSkippedOnboarding: false });
    navigate('/onboarding');
  };

  return (
    <div className="space-y-6">
      {showOnboardingResume && (
        <Card className="ff-card-velocity border-brand-500/30">
          <CardHeader>
            <CardTitle className="font-display flex items-center gap-2">
              <Play className="w-5 h-5 text-brand-500" />
              {t('settings.onboarding')}
            </CardTitle>
            <CardDescription className="text-text-secondary">
              {hasSkippedOnboarding
                ? t('settings.onboardingDescription')
                : `You've completed ${completedSteps.length} of 4 onboarding steps. Continue where you left off.`}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button onClick={handleResumeOnboarding} className="ff-btn-velocity gap-2">
              <Play className="w-4 h-4" />
              {completedSteps.length > 0 ? t('settings.resumeSetup') : t('settings.startSetup')}
            </Button>
          </CardContent>
        </Card>
      )}

      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display">{t('settings.profileInformation')}</CardTitle>
          <CardDescription className="text-text-secondary">
            {t('settings.profileDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="firstName">{t('settings.firstName')}</Label>
              <Input
                id="firstName"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="lastName">{t('settings.lastName')}</Label>
              <Input id="lastName" value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </div>
          </div>
          <UsernameChangeField
            value={username}
            onChange={(val) => setUsername(val)}
          />
          <div className="space-y-2">
            <Label htmlFor="email">{t('settings.email')}</Label>
            <Input
              id="email"
              type="email"
              value={email}
              disabled
              className="opacity-60 cursor-not-allowed"
            />
            <p className="text-xs text-text-muted">{t('settings.emailCannotChange')}</p>
          </div>
          <div className="space-y-2">
            <Label htmlFor="dateOfBirth">{t('settings.dateOfBirth')}</Label>
            <Input
              id="dateOfBirth"
              type="date"
              value={dateOfBirth}
              onChange={(e) => {
                if (isDobLocked) return;
                setDateOfBirth(e.target.value);
              }}
              max={new Date().toISOString().split('T')[0]}
              disabled={isDobLocked}
              className={isDobLocked ? 'opacity-60 cursor-not-allowed' : undefined}
            />
            {isDobLocked ? (
              <p className="text-xs text-text-muted">{t('settings.dobLocked')}</p>
            ) : (
              <p className="text-xs text-text-muted">
                {t('settings.dobDescription')}
              </p>
            )}
          </div>
          <Button onClick={handleSaveProfile} disabled={isSavingProfile} className="ff-btn-velocity">
            {isSavingProfile ? t('settings.saving') : t('settings.saveChanges')}
          </Button>
        </CardContent>
      </Card>

      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display">{t('settings.password')}</CardTitle>
          <CardDescription className="text-text-secondary">{t('settings.passwordDescription')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="currentPassword">{t('settings.currentPassword')}</Label>
            <Input
              id="currentPassword"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="newPassword">{t('settings.newPassword')}</Label>
            <Input
              id="newPassword"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmPassword">{t('settings.confirmPassword')}</Label>
            <Input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
            />
          </div>
          <Button onClick={handleUpdatePassword} disabled={isUpdatingPassword} className="ff-btn-velocity">
            {isUpdatingPassword ? t('settings.updating') : t('settings.updatePassword')}
          </Button>
        </CardContent>
      </Card>

      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Globe className="w-5 h-5 text-brand-500" />
            {t('settings.language')}
          </CardTitle>
          <CardDescription className="text-text-secondary">
            {t('settings.languageDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <LanguagePicker variant="default" showLabel={true} />
        </CardContent>
      </Card>

      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Users className="w-5 h-5 text-brand-500" />
            {t('settings.followStats')}
          </CardTitle>
          <CardDescription className="text-text-secondary">
            {t('settings.followStatsDescription')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <MyFollowStats />
        </CardContent>
      </Card>
    </div>
  );
}
