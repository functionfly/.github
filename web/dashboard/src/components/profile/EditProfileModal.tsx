/**
 * EditProfileModal Component
 *
 * A comprehensive modal for editing user profile information.
 * Includes all editable fields with validation and loading states.
 *
 * @example
 * <EditProfileModal
 *   isOpen={isOpen}
 *   onClose={() => setIsOpen(false)}
 *   profile={profile}
 *   onSave={handleSave}
 * />
 */

import { useState, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  X,
  User,
  MapPin,
  Building2,
  Briefcase,
  Globe,
  Github,
  Twitter,
  Linkedin,
  Camera,
  ImageIcon,
  Loader2,
  Check,
  AlertCircle,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import type { UserProfile } from "@/types";
import type { UpdateProfileRequest } from "@/api/users";

interface EditProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: UserProfile;
  onSave: (data: UpdateProfileRequest) => Promise<void>;
  isLoading?: boolean;
}

interface FormErrors {
  name?: string;
  username?: string;
  website?: string;
  twitterUrl?: string;
  githubUrl?: string;
  linkedinUrl?: string;
}

const MAX_BIO_LENGTH = 500;
const MAX_SKILLS = 20;

// URL validation regex
const URL_REGEX = /^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/;
const USERNAME_REGEX = /^[a-z0-9_-]+$/;

export function EditProfileModal({
  isOpen,
  onClose,
  profile,
  onSave,
  isLoading = false,
}: EditProfileModalProps) {
  // Form state
  const [name, setName] = useState(profile.name || "");
  const [username, setUsername] = useState(profile.username || "");
  const [bio, setBio] = useState(profile.bio || "");
  const [location, setLocation] = useState(profile.location || "");
  const [company, setCompany] = useState(profile.company || "");
  const [jobTitle, setJobTitle] = useState(profile.jobTitle || "");
  const [website, setWebsite] = useState(profile.website || "");
  const [twitterUrl, setTwitterUrl] = useState(profile.socialLinks?.twitter || "");
  const [githubUrl, setGithubUrl] = useState(profile.socialLinks?.github || "");
  const [linkedinUrl, setLinkedinUrl] = useState(profile.socialLinks?.linkedin || "");
  const [coverImageUrl, setCoverImageUrl] = useState(profile.coverImage || "");

  // Validation state
  const [errors, setErrors] = useState<FormErrors>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  // Reset form when profile changes or modal opens
  useEffect(() => {
    if (isOpen) {
      setName(profile.name || "");
      setUsername(profile.username || "");
      setBio(profile.bio || "");
      setLocation(profile.location || "");
      setCompany(profile.company || "");
      setJobTitle(profile.jobTitle || "");
      setWebsite(profile.website || "");
      setTwitterUrl(profile.socialLinks?.twitter || "");
      setGithubUrl(profile.socialLinks?.github || "");
      setLinkedinUrl(profile.socialLinks?.linkedin || "");
      setCoverImageUrl(profile.coverImage || "");
      setErrors({});
      setTouched({});
    }
  }, [isOpen, profile]);

  // Validation functions
  const validateUrl = useCallback((url: string, field: string): string | undefined => {
    if (!url) return undefined;
    if (!URL_REGEX.test(url) && !url.startsWith("http")) {
      return `Please enter a valid URL`;
    }
    return undefined;
  }, []);

  const validateUsername = useCallback((value: string): string | undefined => {
    if (!value) return "Username is required";
    if (value.length < 3) return "Username must be at least 3 characters";
    if (value.length > 30) return "Username must be less than 30 characters";
    if (!USERNAME_REGEX.test(value)) {
      return "Username can only contain lowercase letters, numbers, hyphens, and underscores";
    }
    return undefined;
  }, []);

  const validateName = useCallback((value: string): string | undefined => {
    if (!value.trim()) return "Name is required";
    if (value.trim().length < 2) return "Name must be at least 2 characters";
    if (value.trim().length > 50) return "Name must be less than 50 characters";
    return undefined;
  }, []);

  // Validate field on blur
  const handleBlur = useCallback(
    (field: string) => {
      setTouched((prev) => ({ ...prev, [field]: true }));

      let error: string | undefined;
      switch (field) {
        case "name":
          error = validateName(name);
          break;
        case "username":
          error = validateUsername(username);
          break;
        case "website":
          error = validateUrl(website, "website");
          break;
        case "twitterUrl":
          error = validateUrl(twitterUrl, "twitter");
          break;
        case "githubUrl":
          error = validateUrl(githubUrl, "github");
          break;
        case "linkedinUrl":
          error = validateUrl(linkedinUrl, "linkedin");
          break;
      }

      setErrors((prev) => ({ ...prev, [field]: error }));
    },
    [name, username, website, twitterUrl, githubUrl, linkedinUrl, validateName, validateUsername, validateUrl]
  );

  // Check if form is valid
  const isFormValid = useCallback(() => {
    const nameError = validateName(name);
    const usernameError = validateUsername(username);
    const websiteError = validateUrl(website, "website");
    const twitterError = validateUrl(twitterUrl, "twitter");
    const githubError = validateUrl(githubUrl, "github");
    const linkedinError = validateUrl(linkedinUrl, "linkedin");

    return !nameError && !usernameError && !websiteError && !twitterError && !githubError && !linkedinError;
  }, [name, username, website, twitterUrl, githubUrl, linkedinUrl, validateName, validateUsername, validateUrl]);

  // Handle save
  const handleSave = async () => {
    // Validate all fields
    const newErrors: FormErrors = {
      name: validateName(name),
      username: validateUsername(username),
      website: validateUrl(website, "website"),
      twitterUrl: validateUrl(twitterUrl, "twitter"),
      githubUrl: validateUrl(githubUrl, "github"),
      linkedinUrl: validateUrl(linkedinUrl, "linkedin"),
    };

    setErrors(newErrors);
    setTouched({
      name: true,
      username: true,
      website: true,
      twitterUrl: true,
      githubUrl: true,
      linkedinUrl: true,
    });

    // Check if there are any errors
    if (Object.values(newErrors).some((error) => error !== undefined)) {
      return;
    }

    const data: UpdateProfileRequest = {
      name: name.trim() || undefined,
      username: username.trim() || undefined,
      bio: bio.trim() || undefined,
      location: location.trim() || undefined,
      companyName: company.trim() || undefined,
      jobTitle: jobTitle.trim() || undefined,
      website: website.trim() || undefined,
      twitterUrl: twitterUrl.trim() || undefined,
      githubUrl: githubUrl.trim() || undefined,
      linkedinUrl: linkedinUrl.trim() || undefined,
    };

    await onSave(data);
    onClose();
  };

  // Handle cancel
  const handleCancel = () => {
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-semibold text-text-primary flex items-center gap-2">
            <User className="w-5 h-5 text-brand-500" />
            Edit Profile
          </DialogTitle>
          <DialogDescription className="text-text-secondary">
            Update your public profile information. Changes will be visible to other users.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Avatar & Cover Preview */}
          <div className="relative">
            <div className="h-32 rounded-lg overflow-hidden bg-gradient-to-br from-brand-500/20 to-purple-500/20 relative">
              {coverImageUrl ? (
                <img
                  src={coverImageUrl}
                  alt="Cover"
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full bg-gradient-to-br from-brand-500 via-brand-600 to-indigo-700 opacity-50" />
              )}
              <div className="absolute inset-0 bg-black/20" />
            </div>
            <div className="absolute -bottom-8 left-6 flex items-end gap-4">
              <div className="relative">
                <div className="w-20 h-20 rounded-full border-4 border-card bg-gradient-to-br from-brand-500 to-purple-500 flex items-center justify-center text-white text-2xl font-bold overflow-hidden">
                  {profile.avatar ? (
                    <img
                      src={profile.avatar}
                      alt={profile.name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    name.charAt(0).toUpperCase()
                  )}
                </div>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button className="absolute -bottom-1 -right-1 w-7 h-7 bg-brand-500 hover:bg-brand-600 rounded-full flex items-center justify-center text-white shadow-lg transition-colors">
                        <Camera className="w-3.5 h-3.5" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Avatar is synced from your social login</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
            </div>
          </div>

          {/* Spacer for avatar overlap */}
          <div className="h-6" />

          {/* Basic Info Section */}
          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
              Basic Information
            </h3>

            {/* Display Name */}
            <div className="space-y-2">
              <Label htmlFor="name" className="flex items-center gap-2 text-text-secondary">
                <User className="w-4 h-4" />
                Display Name *
              </Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                onBlur={() => handleBlur("name")}
                placeholder="Your full name"
                className={cn(
                  touched.name && errors.name && "border-error focus:border-error focus:ring-error/20"
                )}
                disabled={isLoading}
              />
              {touched.name && errors.name && (
                <p className="text-xs text-error flex items-center gap-1">
                  <AlertCircle className="w-3 h-3" />
                  {errors.name}
                </p>
              )}
            </div>

            {/* Username */}
            <div className="space-y-2">
              <Label htmlFor="username" className="flex items-center gap-2 text-text-secondary">
                <Globe className="w-4 h-4" />
                Username *
              </Label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-sm">
                  @
                </span>
                <Input
                  id="username"
                  value={username}
                  onChange={(e) =>
                    setUsername(e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, ""))
                  }
                  onBlur={() => handleBlur("username")}
                  placeholder="username"
                  className={cn(
                    "pl-7",
                    touched.username && errors.username && "border-error focus:border-error focus:ring-error/20"
                  )}
                  disabled={isLoading}
                />
              </div>
              {touched.username && errors.username ? (
                <p className="text-xs text-error flex items-center gap-1">
                  <AlertCircle className="w-3 h-3" />
                  {errors.username}
                </p>
              ) : (
                <p className="text-xs text-text-muted">
                  Your public profile URL: /u/{username || "username"}
                </p>
              )}
            </div>

            {/* Bio */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="bio" className="text-text-secondary">
                  Bio
                </Label>
                <span
                  className={cn(
                    "text-xs",
                    bio.length > MAX_BIO_LENGTH ? "text-error" : "text-text-muted"
                  )}
                >
                  {bio.length}/{MAX_BIO_LENGTH}
                </span>
              </div>
              <Textarea
                id="bio"
                value={bio}
                onChange={(e) => setBio(e.target.value.slice(0, MAX_BIO_LENGTH))}
                placeholder="Tell others about yourself, your interests, and what you build..."
                rows={4}
                disabled={isLoading}
              />
            </div>
          </div>

          {/* Location & Work Section */}
          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
              Location & Work
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Location */}
              <div className="space-y-2">
                <Label htmlFor="location" className="flex items-center gap-2 text-text-secondary">
                  <MapPin className="w-4 h-4" />
                  Location
                </Label>
                <Input
                  id="location"
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  placeholder="City, Country"
                  disabled={isLoading}
                />
              </div>

              {/* Company */}
              <div className="space-y-2">
                <Label htmlFor="company" className="flex items-center gap-2 text-text-secondary">
                  <Building2 className="w-4 h-4" />
                  Company
                </Label>
                <Input
                  id="company"
                  value={company}
                  onChange={(e) => setCompany(e.target.value)}
                  placeholder="Company or organization"
                  disabled={isLoading}
                />
              </div>
            </div>

            {/* Job Title */}
            <div className="space-y-2">
              <Label htmlFor="jobTitle" className="flex items-center gap-2 text-text-secondary">
                <Briefcase className="w-4 h-4" />
                Job Title
              </Label>
              <Input
                id="jobTitle"
                value={jobTitle}
                onChange={(e) => setJobTitle(e.target.value)}
                placeholder="e.g., Senior Software Engineer"
                disabled={isLoading}
              />
            </div>
          </div>

          {/* Online Presence Section */}
          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
              Online Presence
            </h3>

            {/* Website */}
            <div className="space-y-2">
              <Label htmlFor="website" className="flex items-center gap-2 text-text-secondary">
                <Globe className="w-4 h-4" />
                Personal Website
              </Label>
              <Input
                id="website"
                value={website}
                onChange={(e) => setWebsite(e.target.value)}
                onBlur={() => handleBlur("website")}
                placeholder="https://yourwebsite.com"
                className={cn(
                  touched.website && errors.website && "border-error focus:border-error focus:ring-error/20"
                )}
                disabled={isLoading}
              />
              {touched.website && errors.website && (
                <p className="text-xs text-error flex items-center gap-1">
                  <AlertCircle className="w-3 h-3" />
                  {errors.website}
                </p>
              )}
            </div>

            {/* Social Links */}
            <div className="space-y-3">
              <Label className="text-text-secondary">Social Profiles</Label>

              {/* GitHub */}
              <div className="space-y-2">
                <div className="relative">
                  <Github className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                  <Input
                    value={githubUrl}
                    onChange={(e) => setGithubUrl(e.target.value)}
                    onBlur={() => handleBlur("githubUrl")}
                    placeholder="https://github.com/username"
                    className={cn(
                      "pl-10",
                      touched.githubUrl && errors.githubUrl && "border-error focus:border-error focus:ring-error/20"
                    )}
                    disabled={isLoading}
                  />
                </div>
                {touched.githubUrl && errors.githubUrl && (
                  <p className="text-xs text-error flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    {errors.githubUrl}
                  </p>
                )}
              </div>

              {/* Twitter/X */}
              <div className="space-y-2">
                <div className="relative">
                  <Twitter className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                  <Input
                    value={twitterUrl}
                    onChange={(e) => setTwitterUrl(e.target.value)}
                    onBlur={() => handleBlur("twitterUrl")}
                    placeholder="https://twitter.com/username"
                    className={cn(
                      "pl-10",
                      touched.twitterUrl && errors.twitterUrl && "border-error focus:border-error focus:ring-error/20"
                    )}
                    disabled={isLoading}
                  />
                </div>
                {touched.twitterUrl && errors.twitterUrl && (
                  <p className="text-xs text-error flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    {errors.twitterUrl}
                  </p>
                )}
              </div>

              {/* LinkedIn */}
              <div className="space-y-2">
                <div className="relative">
                  <Linkedin className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                  <Input
                    value={linkedinUrl}
                    onChange={(e) => setLinkedinUrl(e.target.value)}
                    onBlur={() => handleBlur("linkedinUrl")}
                    placeholder="https://linkedin.com/in/username"
                    className={cn(
                      "pl-10",
                      touched.linkedinUrl && errors.linkedinUrl && "border-error focus:border-error focus:ring-error/20"
                    )}
                    disabled={isLoading}
                  />
                </div>
                {touched.linkedinUrl && errors.linkedinUrl && (
                  <p className="text-xs text-error flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    {errors.linkedinUrl}
                  </p>
                )}
              </div>
            </div>
          </div>

          {/* Cover Image Section */}
          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
              Profile Customization
            </h3>

            <div className="space-y-2">
              <Label htmlFor="coverImage" className="flex items-center gap-2 text-text-secondary">
                <ImageIcon className="w-4 h-4" />
                Cover Image URL
              </Label>
              <Input
                id="coverImage"
                value={coverImageUrl}
                onChange={(e) => setCoverImageUrl(e.target.value)}
                placeholder="https://example.com/cover-image.jpg"
                disabled={isLoading}
              />
              <p className="text-xs text-text-muted">
                Recommended size: 1500x500 pixels. Leave empty for a gradient background.
              </p>
            </div>
          </div>
        </div>

        {/* Footer Actions */}
        <div className="flex flex-col-reverse sm:flex-row sm:justify-end gap-3 pt-4 border-t border-border-subtle">
          <Button
            variant="outline"
            onClick={handleCancel}
            disabled={isLoading}
            className="w-full sm:w-auto"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={isLoading || !isFormValid()}
            className="w-full sm:w-auto gap-2"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Check className="w-4 h-4" />
                Save Changes
              </>
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export default EditProfileModal;
