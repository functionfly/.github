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

export function MyProfilePage() {
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
      toast.success("Profile updated successfully");
      queryClient.invalidateQueries({ queryKey: ["my-profile"] });
      // Also refresh auth store so Navbar reflects changes
      useAuthStore.getState().initialize();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to update profile");
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
          <h1 className="text-2xl font-bold text-white">My Profile</h1>
          <p className="text-text-secondary">Manage your public profile information</p>
        </div>
        {profile.username && (
          <Link to={`/u/${profile.username}`} target="_blank">
            <Button variant="outline" size="sm" className="gap-2">
              <ExternalLink className="w-4 h-4" />
              View Public Profile
            </Button>
          </Link>
        )}
      </div>

      {/* Avatar */}
      <Card>
        <CardHeader>
          <CardTitle>Avatar</CardTitle>
          <CardDescription className="text-text-secondary">
            Your profile picture (set via social login)
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-linear-to-br from-brand-500 to-brand-600 flex items-center justify-center text-white text-2xl font-bold overflow-hidden">
              {profile.avatar ? (
                <img
                  src={profile.avatar}
                  alt={profile.name || "Avatar"}
                  className="w-full h-full object-cover"
                />
              ) : (
                (profile.name || profile.username || profile.email).charAt(0).toUpperCase()
              )}
            </div>
            <div>
              <p className="text-sm text-text-secondary">
                {profile.avatar
                  ? "Avatar synced from your social login provider."
                  : "No avatar set. Connect a social account to get one."}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Profile Info */}
      <Card>
        <CardHeader>
          <CardTitle>Profile Information</CardTitle>
          <CardDescription className="text-text-secondary">
            Update your display name, username, and company
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Email (read-only) */}
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={profile.email}
              disabled
              className="opacity-60 cursor-not-allowed"
            />
            <p className="text-xs text-text-muted">Email cannot be changed here.</p>
          </div>

          {/* Display Name */}
          <div className="space-y-2">
            <Label htmlFor="name" className="flex items-center gap-2">
              <User className="w-4 h-4" />
              Display Name
            </Label>
            <Input
              id="name"
              type="text"
              placeholder="Your full name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isLoading}
            />
          </div>

          {/* Username */}
          <div className="space-y-2">
            <Label htmlFor="username" className="flex items-center gap-2">
              <AtSign className="w-4 h-4" />
              Username
            </Label>
            <Input
              id="username"
              type="text"
              placeholder="yourhandle"
              value={username}
              onChange={(e) => setUsername(e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, ""))}
              disabled={isLoading}
            />
            <p className="text-xs text-text-muted">
              Lowercase letters, numbers, hyphens and underscores only.
              {username && (
                <span className="ml-1 text-brand-400">
                  Public URL: /u/{username}
                </span>
              )}
            </p>
          </div>

          {/* Company */}
          <div className="space-y-2">
            <Label htmlFor="companyName" className="flex items-center gap-2">
              <Building2 className="w-4 h-4" />
              Company <span className="text-text-muted text-xs">(optional)</span>
            </Label>
            <Input
              id="companyName"
              type="text"
              placeholder="Acme Inc"
              value={companyName}
              onChange={(e) => setCompanyName(e.target.value)}
              disabled={isLoading}
            />
          </div>

          {/* Bio */}
          <div className="space-y-2">
            <Label htmlFor="bio" className="flex items-center gap-2">
              <FileText className="w-4 h-4" />
              Bio <span className="text-text-muted text-xs">(optional)</span>
            </Label>
            <Textarea
              id="bio"
              placeholder="Tell us about yourself, your expertise, or what you're building..."
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              disabled={isLoading}
              rows={4}
              maxLength={500}
            />
            <p className="text-xs text-text-muted text-right">
              {bio.length}/500 characters
            </p>
          </div>

          <Button
            onClick={handleSave}
            disabled={!isDirty || updateMutation.isPending || isLoading}
            className="gap-2"
          >
            <Save className="w-4 h-4" />
            {updateMutation.isPending ? "Saving..." : "Save Changes"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
