import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { User, AtSign, Building2, Save, ExternalLink, FileText } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { useAuthStore } from "@/stores/authStore";
import { usersApi, type UpdateProfileRequest } from "@/api/users";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";

export function MyProfilePage() {
  const { t } = useTranslation();
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  // Fetch latest profile from server
  const { data: serverProfile, isLoading } = useQuery({
    queryKey: ["my-profile"],
    queryFn: () => usersApi.getMe(),
    staleTime: 60 * 1000,
  });

  const profile = serverProfile ?? {
    id: user?.id ?? "",
    name: user?.name ?? "",
    username: user?.username ?? "",
    companyName: user?.companyName ?? "",
    bio: user?.bio ?? "",
    email: user?.email ?? "",
    avatar: user?.avatar,
    updatedAt: user?.updatedAt ?? "",
  };

  const [name, setName] = useState(profile.name);
  const [username, setUsername] = useState(profile.username ?? "");
  const [companyName, setCompanyName] = useState(profile.companyName ?? "");
  const [bio, setBio] = useState(profile.bio ?? "");

  // Sync local state when server data arrives
  const [synced, setSynced] = useState(false);
  if (serverProfile && !synced) {
    setName(serverProfile.name ?? "");
    setUsername(serverProfile.username ?? "");
    setCompanyName(serverProfile.companyName ?? "");
    setBio(serverProfile.bio ?? "");
    setSynced(true);
  }

  const updateMutation = useMutation({
    mutationFn: (data: UpdateProfileRequest) => usersApi.updateMe(data),
    onSuccess: (res) => {
      toast.success(t("profile.updateSuccess"));
      queryClient.invalidateQueries({ queryKey: ["my-profile"] });
      // Also refresh auth store so Navbar reflects changes
      useAuthStore.getState().initialize();
    },
    onError: (err: Error) => {
      toast.error(err.message || t("profile.updateError"));
    },
  });

  const handleSave = () => {
    updateMutation.mutate({
      name: name.trim() || undefined,
      username: username.trim() || undefined,
      companyName: companyName.trim() || undefined,
      bio: bio.trim() || undefined,
    });
  };

  const isDirty =
    name !== (profile.name ?? "") ||
    username !== (profile.username ?? "") ||
    companyName !== (profile.companyName ?? "") ||
    bio !== (profile.bio ?? "");

  return (
    <div className="space-y-6 max-w-2xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">{t("profile.title")}</h1>
          <p className="text-text-secondary">{t("profile.subtitle")}</p>
        </div>
        {profile.username && (
          <Link to={`/u/${profile.username}`} target="_blank">
            <Button variant="outline" size="sm" className="gap-2">
              <ExternalLink className="w-4 h-4" />
              {t("profile.viewPublicProfile")}
            </Button>
          </Link>
        )}
      </div>

      {/* Avatar */}
      <Card>
        <CardHeader>
          <CardTitle>{t("profile.avatarTitle")}</CardTitle>
          <CardDescription className="text-text-secondary">
            {t("profile.avatarDescription")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-linear-to-br from-brand-500 to-brand-600 flex items-center justify-center text-white text-2xl font-bold overflow-hidden">
              {profile.avatar ? (
                <img
                  src={profile.avatar}
                  alt={profile.name || t("profile.avatarTitle")}
                  className="w-full h-full object-cover"
                />
              ) : (
                (profile.name || profile.username || profile.email).charAt(0).toUpperCase()
              )}
            </div>
            <div>
              <p className="text-sm text-text-secondary">
                {profile.avatar
                  ? t("profile.avatarSynced")
                  : t("profile.avatarNotSet")}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Profile Info */}
      <Card>
        <CardHeader>
          <CardTitle>{t("profile.profileInformation")}</CardTitle>
          <CardDescription className="text-text-secondary">
            {t("profile.profileInformationDescription")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Email (read-only) */}
          <div className="space-y-2">
            <Label htmlFor="email">{t("profile.email")}</Label>
            <Input
              id="email"
              type="email"
              value={profile.email}
              disabled
              className="opacity-60 cursor-not-allowed"
            />
            <p className="text-xs text-text-muted">{t("profile.emailCannotChange")}</p>
          </div>

          {/* Display Name */}
          <div className="space-y-2">
            <Label htmlFor="name" className="flex items-center gap-2">
              <User className="w-4 h-4" />
              {t("profile.displayName")}
            </Label>
            <Input
              id="name"
              type="text"
              placeholder={t("profile.displayNamePlaceholder")}
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isLoading}
            />
          </div>

          {/* Username */}
          <div className="space-y-2">
            <Label htmlFor="username" className="flex items-center gap-2">
              <AtSign className="w-4 h-4" />
              {t("profile.username")}
            </Label>
            <Input
              id="username"
              type="text"
              placeholder={t("profile.usernamePlaceholder")}
              value={username}
              onChange={(e) => setUsername(e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, ""))}
              disabled={isLoading}
            />
            <p className="text-xs text-text-muted">
              {t("profile.usernameHint")}
              {username && (
                <span className="ml-1 text-brand-400">
                  {t("profile.publicUrl", { username })}
                </span>
              )}
            </p>
          </div>

          {/* Company */}
          <div className="space-y-2">
            <Label htmlFor="companyName" className="flex items-center gap-2">
              <Building2 className="w-4 h-4" />
              {t("profile.company")} <span className="text-text-muted text-xs">{t("profile.optional")}</span>
            </Label>
            <Input
              id="companyName"
              type="text"
              placeholder={t("profile.companyPlaceholder")}
              value={companyName}
              onChange={(e) => setCompanyName(e.target.value)}
              disabled={isLoading}
            />
          </div>

          {/* Bio */}
          <div className="space-y-2">
            <Label htmlFor="bio" className="flex items-center gap-2">
              <FileText className="w-4 h-4" />
              {t("profile.bio")} <span className="text-text-muted text-xs">{t("profile.optional")}</span>
            </Label>
            <Textarea
              id="bio"
              placeholder={t("profile.bioPlaceholder")}
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              disabled={isLoading}
              rows={4}
              maxLength={500}
            />
            <p className="text-xs text-text-muted text-right">
              {t("profile.bioCharacterCount", { count: bio.length })}
            </p>
          </div>

          <Button
            onClick={handleSave}
            disabled={!isDirty || updateMutation.isPending || isLoading}
            className="gap-2"
          >
            <Save className="w-4 h-4" />
            {updateMutation.isPending ? t("profile.saving") : t("profile.saveChanges")}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
