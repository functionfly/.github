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
