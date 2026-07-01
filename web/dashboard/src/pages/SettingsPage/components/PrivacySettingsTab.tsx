/**
 * Settings Tab Component
 *
 * Displays profile settings for the user's own profile.
 * Includes visibility preferences, notification settings, and privacy controls.
 */

import { privacyApi, type PrivacySettings as PrivacyApiSettings } from '@/api/privacy';
import { usersApi, type UpdateProfileRequest } from '@/api/users';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { TooltipProvider } from '@/components/ui/tooltip';
import { CUSTOM_STATUS_OPTIONS, useCustomStatus } from '@/hooks/useCustomStatus';
import { tabContentVariants } from '@/pages/ProfilePage/animations';
import { useAuthStore } from '@/stores/authStore';
import type { UserProfile } from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import {
  AlertTriangle,
  Award,
  Bell,
  Check,
  CircleDot,
  Database,
  Download,
  Eye,
  FileWarning,
  Globe,
  HardDrive,
  Link as LinkIcon,
  Loader2,
  Lock,
  Save,
  Shield,
  Timer,
  Trash2,
  User,
  Users,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';

interface PrivacySettingsTabProps {
  profile?: UserProfile;
}

interface ProfileSettings {
  // Visibility settings
  profileVisibility: 'public' | 'followers' | 'private';
  showEmail: boolean;
  showLocation: boolean;
  showCompany: boolean;
  showActivity: boolean;
  showAnalytics: boolean;
  showFounderBadge: boolean;

  // Notification settings
  emailNotifications: boolean;
  pushNotifications: boolean;
  notifyOnFollow: boolean;
  notifyOnMention: boolean;
  notifyOnFunctionUsage: boolean;
  notifyOnReviews: boolean;
  weeklyDigest: boolean;

  // Privacy settings
  allowTagging: boolean;
  allowIndexing: boolean;
  showLastActive: boolean;
}

const defaultSettings: ProfileSettings = {
  profileVisibility: 'public',
  showEmail: false,
  showLocation: true,
  showCompany: true,
  showActivity: true,
  showAnalytics: true,
  showFounderBadge: true,
  emailNotifications: true,
  pushNotifications: false,
  notifyOnFollow: true,
  notifyOnMention: true,
  notifyOnFunctionUsage: true,
  notifyOnReviews: true,
  weeklyDigest: true,
  allowTagging: true,
  allowIndexing: true,
  showLastActive: true,
};

export function PrivacySettingsTab({ profile }: PrivacySettingsTabProps = {}) {
  const currentUser = useAuthStore((state) => state.user);
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const username = profile?.username || currentUser?.username || '';
  const { status: customStatus, isLoading: isLoadingCustomStatus, setStatus } = useCustomStatus();

  // Fetch settings from backend
  const { data: settingsData, isLoading: isLoadingSettings } = useQuery({
    queryKey: ['my-settings'],
    queryFn: async () => {
      const response = await usersApi.getMySettings();
      return response.settings as unknown as ProfileSettings;
    },
  });

  // Fetch privacy/data processing settings
  const { data: privacyData, isLoading: isLoadingPrivacy } = useQuery({
    queryKey: ['privacy-settings'],
    queryFn: async () => {
      const response = await privacyApi.getSettings();
      return response as unknown as PrivacyApiSettings;
    },
  });

  // Data processing local state
  const [dataProcessing, setDataProcessing] = useState({
    anonymize_ip: false,
    anonymize_user_agent: false,
    store_input_output: true,
    retention_days: 90,
    auto_delete_enabled: false,
  });

  // Export/deletion state
  const [exportRequestId, setExportRequestId] = useState<string | null>(null);
  const [deletionRequestId, setDeletionRequestId] = useState<string | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [isQuickExporting, setIsQuickExporting] = useState(false);

  // Local state for settings
  const [settings, setSettings] = useState<ProfileSettings>(defaultSettings);
  const [socialLinks, setSocialLinks] = useState({
    website: profile?.website || '',
    github: profile?.socialLinks?.github || '',
    twitter: profile?.socialLinks?.twitter || '',
    linkedin: profile?.socialLinks?.linkedin || '',
  });

  // Update local state when data is fetched
  useEffect(() => {
    if (settingsData) {
      setSettings((prev) => ({ ...prev, ...settingsData }));
    }
  }, [settingsData]);

  // Update social links when profile changes
  useEffect(() => {
    if (profile) {
      setSocialLinks({
        website: profile.website || '',
        github: profile.socialLinks?.github || '',
        twitter: profile.socialLinks?.twitter || '',
        linkedin: profile.socialLinks?.linkedin || '',
      });
    }
  }, [profile]);

  // Update data processing state when privacy data is fetched
  useEffect(() => {
    if (privacyData) {
      setDataProcessing({
        anonymize_ip: privacyData.anonymize_ip ?? false,
        anonymize_user_agent: privacyData.anonymize_user_agent ?? false,
        store_input_output: privacyData.store_input_output ?? true,
        retention_days: privacyData.retention_days ?? 90,
        auto_delete_enabled: privacyData.auto_delete_enabled ?? false,
      });
    }
  }, [privacyData]);

  // Mutations
  const updateVisibilityMutation = useMutation({
    mutationFn: (data: Partial<ProfileSettings>) => usersApi.updateMyVisibilitySettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-settings'] });
      queryClient.invalidateQueries({ queryKey: ['enhanced-profile', username] });
      toast.success(t('privacySettings.toastVisibilitySaved'));
    },
    onError: () => toast.error(t('privacySettings.toastVisibilityFailed')),
  });

  const updateNotificationsMutation = useMutation({
    mutationFn: (data: Record<string, boolean>) => usersApi.updateMyNotificationSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-settings'] });
      toast.success(t('privacySettings.toastNotificationsSaved'));
    },
    onError: () => toast.error(t('privacySettings.toastNotificationsFailed')),
  });

  const updatePrivacyMutation = useMutation({
    mutationFn: (data: Record<string, boolean>) => usersApi.updateMyPrivacySettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-settings'] });
      toast.success(t('privacySettings.toastPrivacySaved'));
    },
    onError: () => toast.error(t('privacySettings.toastPrivacyFailed')),
  });

  const updateSocialLinksMutation = useMutation({
    mutationFn: async () => {
      const socialUpdate: UpdateProfileRequest = {};
      if (socialLinks.website !== profile?.website) {
        socialUpdate.website = socialLinks.website;
      }
      if (socialLinks.github !== profile?.socialLinks?.github) {
        socialUpdate.githubUrl = socialLinks.github;
      }
      if (socialLinks.twitter !== profile?.socialLinks?.twitter) {
        socialUpdate.twitterUrl = socialLinks.twitter;
      }
      if (socialLinks.linkedin !== profile?.socialLinks?.linkedin) {
        socialUpdate.linkedinUrl = socialLinks.linkedin;
      }

      if (Object.keys(socialUpdate).length > 0) {
        await usersApi.updateMe(socialUpdate);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['enhanced-profile', username] });
      toast.success(t('privacySettings.toastSocialLinksSaved'));
    },
    onError: () => toast.error(t('privacySettings.toastSocialLinksFailed')),
  });

  const updateDataProcessingMutation = useMutation({
    mutationFn: (data: Partial<PrivacyApiSettings>) => privacyApi.updateSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['privacy-settings'] });
      toast.success(t('privacySettings.toastDataProcessingSaved'));
    },
    onError: () => toast.error(t('privacySettings.toastDataProcessingFailed')),
  });

  const requestExportMutation = useMutation({
    mutationFn: () => privacyApi.requestDataExport('full'),
    onSuccess: (data) => {
      const result = data as unknown as { id: string };
      setExportRequestId(result.id);
      toast.success(t('privacySettings.toastExportRequested'));
    },
    onError: () => toast.error(t('privacySettings.toastExportFailed')),
  });

  const requestDeletionMutation = useMutation({
    mutationFn: () => privacyApi.requestDataDeletion('full'),
    onSuccess: (data) => {
      const result = data as unknown as { id: string };
      setDeletionRequestId(result.id);
      setShowDeleteConfirm(false);
      toast.success(t('privacySettings.toastDeletionRequested'));
    },
    onError: () => toast.error(t('privacySettings.toastDeletionFailed')),
  });

  // Export status polling
  const { data: exportStatus } = useQuery({
    queryKey: ['export-status', exportRequestId],
    queryFn: async () => {
      if (!exportRequestId) return null;
      const response = await privacyApi.getExportStatus(exportRequestId);
      return response as unknown as { id: string; status: string; download_url?: string; download_token?: string; file_size?: number; error_message?: string };
    },
    enabled: !!exportRequestId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === 'completed' || status === 'failed') return false;
      return 3000;
    },
  });

  // Deletion status polling
  const { data: deletionStatus } = useQuery({
    queryKey: ['deletion-status', deletionRequestId],
    queryFn: async () => {
      if (!deletionRequestId) return null;
      const response = await privacyApi.getDeletionStatus(deletionRequestId);
      return response as unknown as { id: string; status: string; records_deleted?: number; error_message?: string };
    },
    enabled: !!deletionRequestId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === 'completed' || status === 'failed') return false;
      return 3000;
    },
  });

  const updateSetting = <K extends keyof ProfileSettings>(key: K, value: ProfileSettings[K]) => {
    setSettings((prev) => ({ ...prev, [key]: value }));
  };

  const updateSocialLink = (key: keyof typeof socialLinks, value: string) => {
    setSocialLinks((prev) => ({ ...prev, [key]: value }));
  };

  const handleSaveVisibility = () => {
    updateVisibilityMutation.mutate({
      profileVisibility: settings.profileVisibility,
      showEmail: settings.showEmail,
      showLocation: settings.showLocation,
      showCompany: settings.showCompany,
      showActivity: settings.showActivity,
      showAnalytics: settings.showAnalytics,
      showFounderBadge: settings.showFounderBadge,
    });
  };

  const handleSaveNotifications = () => {
    updateNotificationsMutation.mutate({
      emailNotifications: settings.emailNotifications,
      pushNotifications: settings.pushNotifications,
      notifyOnFollow: settings.notifyOnFollow,
      notifyOnMention: settings.notifyOnMention,
      notifyOnFunctionUsage: settings.notifyOnFunctionUsage,
      notifyOnReviews: settings.notifyOnReviews,
      weeklyDigest: settings.weeklyDigest,
    });
  };

  const handleSavePrivacy = () => {
    updatePrivacyMutation.mutate({
      allowTagging: settings.allowTagging,
      allowIndexing: settings.allowIndexing,
      showLastActive: settings.showLastActive,
    });
  };

  const handleSaveSocialLinks = () => {
    updateSocialLinksMutation.mutate();
  };

  const handleSaveDataProcessing = () => {
    updateDataProcessingMutation.mutate({
      anonymize_ip: dataProcessing.anonymize_ip,
      anonymize_user_agent: dataProcessing.anonymize_user_agent,
      store_input_output: dataProcessing.store_input_output,
      retention_days: dataProcessing.retention_days,
      auto_delete_enabled: dataProcessing.auto_delete_enabled,
    });
  };

  const handleQuickExport = async () => {
    setIsQuickExporting(true);
    try {
      const [me, mySettings, activeSessions] = await Promise.all([
        usersApi.getMe(),
        usersApi.getMySettings(),
        usersApi.listSessions(),
      ]);

      const payload = {
        exportedAt: new Date().toISOString(),
        account: me,
        settings: mySettings.settings ?? {},
        sessions: activeSessions.sessions ?? [],
      };

      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      const date = new Date().toISOString().slice(0, 10);
      anchor.href = url;
      anchor.download = `functionfly-data-export-${date}.json`;
      document.body.appendChild(anchor);
      anchor.click();
      document.body.removeChild(anchor);
      URL.revokeObjectURL(url);
      toast.success(t('privacySettings.toastQuickExportReady'));
    } catch {
      toast.error(t('privacySettings.toastQuickExportFailed'));
    } finally {
      setIsQuickExporting(false);
    }
  };

  const handleSaveAll = async () => {
    const results = await Promise.allSettled([
      updateVisibilityMutation.mutateAsync({
        profileVisibility: settings.profileVisibility,
        showEmail: settings.showEmail,
        showLocation: settings.showLocation,
        showCompany: settings.showCompany,
        showActivity: settings.showActivity,
        showAnalytics: settings.showAnalytics,
        showFounderBadge: settings.showFounderBadge,
      }),
      updateNotificationsMutation.mutateAsync({
        emailNotifications: settings.emailNotifications,
        pushNotifications: settings.pushNotifications,
        notifyOnFollow: settings.notifyOnFollow,
        notifyOnMention: settings.notifyOnMention,
        notifyOnFunctionUsage: settings.notifyOnFunctionUsage,
        notifyOnReviews: settings.notifyOnReviews,
        weeklyDigest: settings.weeklyDigest,
      }),
      updatePrivacyMutation.mutateAsync({
        allowTagging: settings.allowTagging,
        allowIndexing: settings.allowIndexing,
        showLastActive: settings.showLastActive,
      }),
      updateSocialLinksMutation.mutateAsync(),
      updateDataProcessingMutation.mutateAsync({
        anonymize_ip: dataProcessing.anonymize_ip,
        anonymize_user_agent: dataProcessing.anonymize_user_agent,
        store_input_output: dataProcessing.store_input_output,
        retention_days: dataProcessing.retention_days,
        auto_delete_enabled: dataProcessing.auto_delete_enabled,
      }),
    ]);

    const failed = results.filter((r) => r.status === 'rejected');
    if (failed.length === 0) {
      toast.success(t('privacySettings.toastAllSaved'));
    } else if (failed.length < results.length) {
      toast.warning(t('privacySettings.toastPartialSave', { failed: failed.length, total: results.length, defaultValue: `${results.length - failed.length} of ${results.length} settings saved. ${failed.length} failed.` }));
    } else {
      toast.error(t('privacySettings.toastAllFailed', { defaultValue: 'Failed to save settings' }));
    }
  };

  const getVisibilityIcon = (visibility: string) => {
    switch (visibility) {
      case 'public':
        return <Globe className="w-4 h-4" />;
      case 'followers':
        return <Users className="w-4 h-4" />;
      case 'private':
        return <Lock className="w-4 h-4" />;
      default:
        return <Globe className="w-4 h-4" />;
    }
  };

  const isSaving =
    updateVisibilityMutation.isPending ||
    updateNotificationsMutation.isPending ||
    updatePrivacyMutation.isPending ||
    updateSocialLinksMutation.isPending ||
    updateDataProcessingMutation.isPending;

  if (isLoadingSettings) {
    return (
      <motion.div
        variants={tabContentVariants}
        initial="hidden"
        animate="visible"
        exit="exit"
        className="settings-page space-y-6 px-4 md:px-8 pb-8"
      >
        <div className="animate-pulse space-y-4">
          <div className="h-32 bg-bg-secondary rounded-lg" />
          <div className="h-48 bg-bg-secondary rounded-lg" />
          <div className="h-48 bg-bg-secondary rounded-lg" />
        </div>
      </motion.div>
    );
  }

  return (
    <TooltipProvider>
      <motion.div
        variants={tabContentVariants}
        initial="hidden"
        animate="visible"
        exit="exit"
        className="settings-page space-y-6 px-4 md:px-8 pb-8"
      >
        {/* Header with save button */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="font-display text-xl font-semibold text-white">{t('privacySettings.title')}</h2>
            <p className="text-sm text-gray-400">
              {t('privacySettings.description')}
            </p>
          </div>
          <Button
            onClick={handleSaveAll}
            disabled={isSaving}
            style={{
              background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
              color: 'var(--text-on-light)',
              boxShadow: 'var(--shadow-btn-primary-rest)',
            }}
            className="min-w-[120px]"
          >
            {isSaving ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Save className="w-4 h-4 mr-2" />
            )}
            {isSaving ? t('privacySettings.saving') : t('privacySettings.saveAll')}
          </Button>
        </div>

        {/* Profile Visibility Section */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Eye className="w-5 h-5 text-brand-500" />
              {t('privacySettings.profileVisibility')}
            </CardTitle>
            <CardDescription>
              {t('privacySettings.profileVisibilityDesc')}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Visibility Level */}
            <div className="space-y-3">
              <Label className="text-sm font-medium">{t('privacySettings.visibilityLevel')}</Label>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                {(['public', 'followers', 'private'] as const).map((level) => (
                  <button
                    key={level}
                    onClick={() => updateSetting('profileVisibility', level)}
                    className={`flex items-center gap-3 p-3 rounded-lg border transition-all ${
                      settings.profileVisibility === level
                        ? 'border-brand-500 bg-brand-500/10'
                        : 'border-border-subtle hover:border-border-default bg-bg-secondary'
                    }`}
                  >
                    <div
                      className={`p-2 rounded-md ${
                        settings.profileVisibility === level
                          ? 'bg-brand-500 text-white'
                          : 'bg-bg-tertiary text-text-secondary'
                      }`}
                    >
                      {getVisibilityIcon(level)}
                    </div>
                    <div className="text-left">
                      <p
                        className={`font-medium capitalize ${
                          settings.profileVisibility === level
                            ? 'text-brand-400'
                            : 'text-text-primary'
                        }`}
                      >
                        {t(`privacySettings.${level}`)}
                      </p>
                      <p className="text-xs text-text-muted">
                        {level === 'public' && t('privacySettings.publicDesc')}
                        {level === 'followers' && t('privacySettings.followersDesc')}
                        {level === 'private' && t('privacySettings.privateDesc')}
                      </p>
                    </div>
                    {settings.profileVisibility === level && (
                      <Check className="w-4 h-4 text-brand-500 ml-auto" />
                    )}
                  </button>
                ))}
              </div>
            </div>

            <div className="ff-divider-flame" />

            {/* Field Visibility Toggles */}
            <div className="space-y-4">
              <Label className="text-sm font-medium">{t('privacySettings.fieldVisibility')}</Label>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <User className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{t('privacySettings.emailAddress')}</p>
                    <p className="text-sm text-text-muted">{t('privacySettings.emailDesc')}</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showEmail}
                  onCheckedChange={(checked) => updateSetting('showEmail', checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Globe className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{t('privacySettings.location')}</p>
                    <p className="text-sm text-text-muted">{t('privacySettings.locationDesc')}</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showLocation}
                  onCheckedChange={(checked) => updateSetting('showLocation', checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Users className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{t('privacySettings.company')}</p>
                    <p className="text-sm text-text-muted">{t('privacySettings.companyDesc')}</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showCompany}
                  onCheckedChange={(checked) => updateSetting('showCompany', checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Eye className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{t('privacySettings.activityStatus')}</p>
                    <p className="text-sm text-text-muted">{t('privacySettings.activityDesc')}</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showActivity}
                  onCheckedChange={(checked) => updateSetting('showActivity', checked)}
                />
              </div>

                <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Shield className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{t('privacySettings.analytics')}</p>
                    <p className="text-sm text-text-muted">{t('privacySettings.analyticsDesc')}</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showAnalytics}
                  onCheckedChange={(checked) => updateSetting('showAnalytics', checked)}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-md bg-bg-tertiary">
                    <Award className="w-4 h-4 text-text-secondary" />
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">{t('privacySettings.foundersBadge')}</p>
                    <p className="text-sm text-text-muted">{t('privacySettings.foundersBadgeDesc')}</p>
                  </div>
                </div>
                <Switch
                  checked={settings.showFounderBadge}
                  onCheckedChange={(checked) => updateSetting('showFounderBadge', checked)}
                />
              </div>
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSaveVisibility}
                disabled={updateVisibilityMutation.isPending}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                {updateVisibilityMutation.isPending && (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                )}
                {t('privacySettings.saveVisibility')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Social Links Section */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <LinkIcon className="w-5 h-5 text-brand-500" />
              {t('privacySettings.socialLinks')}
            </CardTitle>
            <CardDescription>{t('privacySettings.socialLinksDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            {/* Personal Website */}
            <div className="group">
              <Label htmlFor="website" className="text-sm font-medium mb-2 block">
                {t('privacySettings.personalWebsite')}
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-brand-500/10 group-hover:text-brand-500 transition-colors">
                  <Globe className="w-4 h-4" />
                </div>
                <Input
                  id="website"
                  type="url"
                  placeholder="https://yourwebsite.com"
                  value={socialLinks.website}
                  onChange={(e) => updateSocialLink('website', e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            {/* GitHub Profile */}
            <div className="group">
              <Label htmlFor="github" className="text-sm font-medium mb-2 block">
                {t('privacySettings.githubProfile')}
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-[#333] group-hover:text-white transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z" />
                  </svg>
                </div>
                <Input
                  id="github"
                  type="url"
                  placeholder="https://github.com/username"
                  value={socialLinks.github}
                  onChange={(e) => updateSocialLink('github', e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            {/* Twitter/X Profile */}
            <div className="group">
              <Label htmlFor="twitter" className="text-sm font-medium mb-2 block">
                {t('privacySettings.twitterProfile')}
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-black group-hover:text-white transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
                  </svg>
                </div>
                <Input
                  id="twitter"
                  type="url"
                  placeholder="https://twitter.com/username"
                  value={socialLinks.twitter}
                  onChange={(e) => updateSocialLink('twitter', e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            {/* LinkedIn Profile */}
            <div className="group">
              <Label htmlFor="linkedin" className="text-sm font-medium mb-2 block">
                {t('privacySettings.linkedinProfile')}
              </Label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 p-2 rounded-md bg-bg-tertiary text-text-secondary group-hover:bg-[#0077b5] group-hover:text-white transition-colors">
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z" />
                  </svg>
                </div>
                <Input
                  id="linkedin"
                  type="url"
                  placeholder="https://linkedin.com/in/username"
                  value={socialLinks.linkedin}
                  onChange={(e) => updateSocialLink('linkedin', e.target.value)}
                  className="pl-12 bg-bg-secondary transition-all focus:ring-2 focus:ring-brand-500/20"
                />
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <Button
                onClick={handleSaveSocialLinks}
                disabled={updateSocialLinksMutation.isPending}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                {updateSocialLinksMutation.isPending && (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                )}
                {updateSocialLinksMutation.isPending ? t('privacySettings.saving') : t('privacySettings.saveLinks')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Custom Status Section */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <CircleDot className="w-5 h-5 text-brand-500" />
              {t('privacySettings.presenceStatus')}
            </CardTitle>
            <CardDescription>
              {t('privacySettings.presenceStatusDesc')}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {isLoadingCustomStatus ? (
              <div className="flex items-center gap-2 text-text-muted">
                <Loader2 className="w-4 h-4 animate-spin" />
                <span className="text-sm">{t('privacySettings.loadingStatus')}</span>
              </div>
            ) : (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-2">
                  {CUSTOM_STATUS_OPTIONS.map((option) => (
                    <button
                      key={option.value}
                      onClick={() => setStatus(option.value, option.emoji)}
                      disabled={option.value === customStatus.customStatus}
                      className={`flex flex-col items-center gap-1.5 p-3 rounded-lg border transition-all ${
                        option.value === customStatus.customStatus
                          ? 'border-brand-500 bg-brand-500/10'
                          : 'border-border-subtle hover:border-border-default bg-bg-secondary disabled:opacity-50'
                      }`}
                    >
                      <span className="text-xl">{option.emoji}</span>
                      <span
                        className={`text-xs font-medium ${
                          option.value === customStatus.customStatus
                            ? 'text-brand-400'
                            : 'text-text-primary'
                        }`}
                      >
                        {option.label}
                      </span>
                    </button>
                  ))}
                </div>
                {customStatus.customStatusEmoji && (
                  <div className="flex items-center gap-2 text-sm text-text-muted">
                    <span>{t('privacySettings.currentStatus')}</span>
                    <span className="text-lg">{customStatus.customStatusEmoji}</span>
                    <span className="font-medium capitalize">{customStatus.customStatus}</span>
                  </div>
                )}
              </>
            )}
          </CardContent>
        </Card>

        {/* Notification Settings */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Bell className="w-5 h-5 text-brand-500" />
              {t('privacySettings.notificationPreferences')}
            </CardTitle>
            <CardDescription>
              {t('privacySettings.notificationPreferencesDesc')}
              {' '}
              <span className="text-text-muted">
                {t('privacySettings.forOperationalAlerts', 'For deployment and system alerts, see the Notifications tab.')}
              </span>
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Master Toggles */}
            <div className="flex flex-wrap gap-4">
              <div className="flex items-center gap-2">
                <Switch
                  checked={settings.emailNotifications}
                  onCheckedChange={(checked) => updateSetting('emailNotifications', checked)}
                />
                <Label className="text-sm font-medium cursor-pointer">{t('privacySettings.emailNotifications')}</Label>
              </div>
              <div className="flex items-center gap-2">
                <Switch
                  checked={settings.pushNotifications}
                  onCheckedChange={(checked) => updateSetting('pushNotifications', checked)}
                />
                <Label className="text-sm font-medium cursor-pointer">{t('privacySettings.pushNotifications')}</Label>
              </div>
            </div>

            <div className="ff-divider-flame" />

            {/* Notification Types */}
            <div className="space-y-4">
              <Label className="text-sm font-medium">{t('privacySettings.notificationTypes')}</Label>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.newFollowers')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.newFollowersDesc')}</p>
                </div>
                <Switch
                  checked={settings.notifyOnFollow}
                  onCheckedChange={(checked) => updateSetting('notifyOnFollow', checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.mentions')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.mentionsDesc')}</p>
                </div>
                <Switch
                  checked={settings.notifyOnMention}
                  onCheckedChange={(checked) => updateSetting('notifyOnMention', checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.functionUsage')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.functionUsageDesc')}</p>
                </div>
                <Switch
                  checked={settings.notifyOnFunctionUsage}
                  onCheckedChange={(checked) => updateSetting('notifyOnFunctionUsage', checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.reviews')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.reviewsDesc')}</p>
                </div>
                <Switch
                  checked={settings.notifyOnReviews}
                  onCheckedChange={(checked) => updateSetting('notifyOnReviews', checked)}
                  disabled={!settings.emailNotifications && !settings.pushNotifications}
                />
              </div>

              <div className="flex items-center justify-between py-2">
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.weeklyDigest')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.weeklyDigestDesc')}</p>
                </div>
                <Switch
                  checked={settings.weeklyDigest}
                  onCheckedChange={(checked) => updateSetting('weeklyDigest', checked)}
                  disabled={!settings.emailNotifications}
                />
              </div>
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSaveNotifications}
                disabled={updateNotificationsMutation.isPending}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                {updateNotificationsMutation.isPending && (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                )}
                {t('privacySettings.saveNotifications')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Privacy Settings */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Shield className="w-5 h-5 text-brand-500" />
              {t('privacySettings.privacySecurity')}
            </CardTitle>
            <CardDescription>{t('privacySettings.privacySecurityDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Users className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.allowTagging')}</p>
                  <p className="text-sm text-text-muted">
                    {t('privacySettings.allowTaggingDesc')}
                  </p>
                </div>
              </div>
              <Switch
                checked={settings.allowTagging}
                onCheckedChange={(checked) => updateSetting('allowTagging', checked)}
              />
            </div>

            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Globe className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.searchEngineIndexing')}</p>
                  <p className="text-sm text-text-muted">
                    {t('privacySettings.searchEngineIndexingDesc')}
                  </p>
                </div>
              </div>
              <Switch
                checked={settings.allowIndexing}
                onCheckedChange={(checked) => updateSetting('allowIndexing', checked)}
              />
            </div>

            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Eye className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.lastActiveStatus')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.lastActiveStatusDesc')}</p>
                </div>
              </div>
              <Switch
                checked={settings.showLastActive}
                onCheckedChange={(checked) => updateSetting('showLastActive', checked)}
              />
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSavePrivacy}
                disabled={updatePrivacyMutation.isPending}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                {updatePrivacyMutation.isPending && (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                )}
                {t('privacySettings.savePrivacy')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Data Processing Section */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Database className="w-5 h-5 text-brand-500" />
              {t('privacySettings.dataProcessing')}
            </CardTitle>
            <CardDescription>{t('privacySettings.dataProcessingDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Lock className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.anonymizeIp')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.anonymizeIpDesc')}</p>
                </div>
              </div>
              <Switch
                checked={dataProcessing.anonymize_ip}
                onCheckedChange={(checked) =>
                  setDataProcessing((prev) => ({ ...prev, anonymize_ip: checked }))
                }
              />
            </div>

            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Shield className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.anonymizeUA')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.anonymizeUADesc')}</p>
                </div>
              </div>
              <Switch
                checked={dataProcessing.anonymize_user_agent}
                onCheckedChange={(checked) =>
                  setDataProcessing((prev) => ({ ...prev, anonymize_user_agent: checked }))
                }
              />
            </div>

            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <HardDrive className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.storeInputOutput')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.storeInputOutputDesc')}</p>
                </div>
              </div>
              <Switch
                checked={dataProcessing.store_input_output}
                onCheckedChange={(checked) =>
                  setDataProcessing((prev) => ({ ...prev, store_input_output: checked }))
                }
              />
            </div>

            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Trash2 className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.autoDelete')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.autoDeleteDesc')}</p>
                </div>
              </div>
              <Switch
                checked={dataProcessing.auto_delete_enabled}
                onCheckedChange={(checked) =>
                  setDataProcessing((prev) => ({ ...prev, auto_delete_enabled: checked }))
                }
              />
            </div>

            <div className="ff-divider-flame" />

            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Timer className="w-4 h-4 text-text-secondary" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.dataRetention')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.dataRetentionDesc')}</p>
                </div>
              </div>
              <div className="pl-12 space-y-2">
                <div className="flex items-center gap-4">
                  <input
                    type="range"
                    min={30}
                    max={365}
                    step={30}
                    value={dataProcessing.retention_days}
                    onChange={(e) =>
                      setDataProcessing((prev) => ({
                        ...prev,
                        retention_days: Number(e.target.value),
                      }))
                    }
                    className="flex-1 accent-brand-500"
                  />
                  <span className="text-sm font-medium text-text-primary min-w-[80px] text-right">
                    {dataProcessing.retention_days} {t('privacySettings.days')}
                  </span>
                </div>
                <div className="flex justify-between text-xs text-text-muted">
                  <span>30</span>
                  <span>90</span>
                  <span>180</span>
                  <span>365</span>
                </div>
              </div>
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSaveDataProcessing}
                disabled={updateDataProcessingMutation.isPending}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                {updateDataProcessingMutation.isPending && (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                )}
                {t('privacySettings.saveDataProcessing')}
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Your Data Section (GDPR) */}
        <Card className="settings-panel border-border-subtle">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2">
              <Download className="w-5 h-5 text-brand-500" />
              {t('privacySettings.yourData')}
            </CardTitle>
            <CardDescription>{t('privacySettings.yourDataDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Data Export */}
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <Download className="w-4 h-4 text-text-secondary" />
                </div>
                <div className="flex-1">
                  <p className="font-medium text-text-primary">{t('privacySettings.exportData')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.exportDataDesc')}</p>
                </div>
              </div>
              <div className="pl-12 space-y-3">
                <Button
                  variant="outline"
                  onClick={() => requestExportMutation.mutate()}
                  disabled={requestExportMutation.isPending}
                  className="w-full sm:w-auto"
                >
                  {requestExportMutation.isPending ? (
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  ) : (
                    <Download className="w-4 h-4 mr-2" />
                  )}
                  {t('privacySettings.requestExport')}
                </Button>

                {exportStatus && (
                  <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3 space-y-2">
                    <div className="flex items-center gap-2">
                      <div
                        className={`w-2 h-2 rounded-full ${
                          exportStatus.status === 'completed'
                            ? 'bg-green-500'
                            : exportStatus.status === 'failed'
                            ? 'bg-red-500'
                            : 'bg-yellow-500 animate-pulse'
                        }`}
                      />
                      <span className="text-sm font-medium capitalize">{exportStatus.status}</span>
                    </div>
                    {exportStatus.status === 'completed' && exportStatus.download_url && (
                      <a
                        href={`/v1/privacy/export/${exportStatus.id}/download?token=${exportStatus.download_token}`}
                        className="inline-flex items-center gap-1 text-sm text-brand-400 hover:text-brand-300 underline"
                      >
                        <Download className="w-3 h-3" />
                        {t('privacySettings.downloadExport')}
                        {exportStatus.file_size && (
                          <span className="text-text-muted">
                            ({(exportStatus.file_size / 1024 / 1024).toFixed(1)} MB)
                          </span>
                        )}
                      </a>
                    )}
                    {exportStatus.status === 'failed' && exportStatus.error_message && (
                      <p className="text-sm text-red-400">{exportStatus.error_message}</p>
                    )}
                  </div>
                )}
              </div>
            </div>

            <div className="ff-divider-flame" />

            {/* Quick Export (client-side) */}
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-bg-tertiary">
                  <FileWarning className="w-4 h-4 text-text-secondary" />
                </div>
                <div className="flex-1">
                  <p className="font-medium text-text-primary">{t('privacySettings.quickExport')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.quickExportDesc')}</p>
                </div>
              </div>
              <div className="pl-12">
                <Button
                  variant="outline"
                  onClick={handleQuickExport}
                  disabled={isQuickExporting}
                  className="w-full sm:w-auto"
                >
                  {isQuickExporting ? (
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  ) : (
                    <Download className="w-4 h-4 mr-2" />
                  )}
                  {isQuickExporting ? t('privacySettings.preparingExport') : t('privacySettings.quickExportButton')}
                </Button>
              </div>
            </div>

          </CardContent>
        </Card>

        {/* Danger Zone */}
        <Card className="settings-panel border-red-500/30">
          <CardHeader>
            <CardTitle className="font-display text-lg flex items-center gap-2 text-red-400">
              <AlertTriangle className="w-5 h-5" />
              {t('privacySettings.dangerZoneTitle')}
            </CardTitle>
            <CardDescription>{t('privacySettings.dangerZoneDesc')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between py-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-red-500/10">
                  <Trash2 className="w-4 h-4 text-red-400" />
                </div>
                <div>
                  <p className="font-medium text-text-primary">{t('privacySettings.deleteAccount')}</p>
                  <p className="text-sm text-text-muted">{t('privacySettings.deleteAccountDesc')}</p>
                </div>
              </div>
            </div>

            <div className="space-y-3">
              {!showDeleteConfirm && !deletionRequestId && (
                <Button
                  variant="destructive"
                  onClick={() => setShowDeleteConfirm(true)}
                  style={{
                    background: 'var(--status-revoked)',
                    borderColor: 'var(--status-revoked)',
                  }}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  {t('privacySettings.requestDeletion')}
                </Button>
              )}

              {showDeleteConfirm && (
                <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-4 space-y-3">
                  <div className="flex items-start gap-2">
                    <AlertTriangle className="w-5 h-5 text-red-400 mt-0.5 shrink-0" />
                    <div>
                      <p className="font-medium text-red-300">{t('privacySettings.deleteConfirmTitle')}</p>
                      <p className="text-sm text-text-muted mt-1">
                        {t('privacySettings.deleteConfirmDesc')}
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      onClick={() => requestDeletionMutation.mutate()}
                      disabled={requestDeletionMutation.isPending}
                      className="border-red-500/50 bg-red-500/10 text-red-300 hover:bg-red-500/20"
                    >
                      {requestDeletionMutation.isPending ? (
                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      ) : (
                        <Trash2 className="w-4 h-4 mr-2" />
                      )}
                      {t('privacySettings.confirmDelete')}
                    </Button>
                    <Button
                      variant="ghost"
                      onClick={() => setShowDeleteConfirm(false)}
                    >
                      {t('privacySettings.cancel')}
                    </Button>
                  </div>
                </div>
              )}

              {deletionStatus && (
                <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3 space-y-2">
                  <div className="flex items-center gap-2">
                    <div
                      className={`w-2 h-2 rounded-full ${
                        deletionStatus.status === 'completed'
                          ? 'bg-green-500'
                          : deletionStatus.status === 'failed'
                          ? 'bg-red-500'
                          : 'bg-yellow-500 animate-pulse'
                      }`}
                    />
                    <span className="text-sm font-medium capitalize">{deletionStatus.status}</span>
                  </div>
                  {deletionStatus.status === 'completed' && (
                    <p className="text-sm text-text-muted">
                      {t('privacySettings.deletionComplete', {
                        count: deletionStatus.records_deleted ?? 0,
                        defaultValue: `${deletionStatus.records_deleted ?? 0} records deleted`,
                      })}
                    </p>
                  )}
                  {deletionStatus.status === 'failed' && deletionStatus.error_message && (
                    <p className="text-sm text-red-400">{deletionStatus.error_message}</p>
                  )}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Save Button at bottom */}
        <div className="flex items-center justify-end gap-4">
          <Button
            onClick={handleSaveAll}
            disabled={isSaving}
            size="lg"
            style={{
              background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
              color: 'var(--text-on-light)',
              boxShadow: 'var(--shadow-btn-primary-rest)',
            }}
            className="min-w-[160px]"
          >
            {isSaving ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Save className="w-4 h-4 mr-2" />
            )}
            {isSaving ? t('privacySettings.saving') : t('privacySettings.saveAllChanges')}
          </Button>
        </div>
      </motion.div>
    </TooltipProvider>
  );
}

export default PrivacySettingsTab;
