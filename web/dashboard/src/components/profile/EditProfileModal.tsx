/**
 * EditProfileModal Component
 *
 * A comprehensive modal for editing user profile information.
 * Uses react-hook-form, zod validation, sonner toasts, and framer-motion.
 *
 * @example
 * <EditProfileModal
 *   isOpen={isOpen}
 *   onClose={() => setIsOpen(false)}
 *   profile={profile}
 *   onSave={handleSave}
 * />
 */

import { useState, useEffect, useRef } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion, AnimatePresence } from "framer-motion";
import { toast } from "sonner";
import {
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

const MAX_BIO_LENGTH = 500;

// Nominatim (OpenStreetMap) - free, no API key. Policy: 1 req/sec, set User-Agent.
const NOMINATIM_URL = "https://nominatim.openstreetmap.org/search";
const LOCATION_DEBOUNCE_MS = 1000;
const LOCATION_MIN_CHARS = 2;
const LOCATION_MAX_RESULTS = 5;

interface NominatimAddress {
  city?: string;
  town?: string;
  village?: string;
  municipality?: string;
  state?: string;
  state_district?: string;
  country?: string;
  country_code?: string;
}

interface NominatimResult {
  place_id: number;
  display_name: string;
  address?: NominatimAddress;
}

function formatLocationLabel(r: NominatimResult): string {
  const a = r.address;
  if (!a) return r.display_name;
  const city = a.city || a.town || a.village || a.municipality;
  const state = a.state || a.state_district;
  const country = a.country;
  const parts = [city, state, country].filter(Boolean);
  return parts.length > 0 ? parts.join(", ") : r.display_name;
}

async function fetchLocationSuggestions(query: string): Promise<NominatimResult[]> {
  if (!query || query.length < LOCATION_MIN_CHARS) return [];
  const params = new URLSearchParams({
    q: query,
    format: "json",
    limit: String(LOCATION_MAX_RESULTS),
    addressdetails: "1",
  });
  const res = await fetch(`${NOMINATIM_URL}?${params}`, {
    headers: { "Accept-Language": "en", "User-Agent": "FunctionFly-Profile-Edit/1.0" },
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

const optionalUrl = z
  .string()
  .optional()
  .refine(
    (val) => !val || val === "" || /^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/.test(val),
    "Please enter a valid URL"
  );

const editProfileSchema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .transform((s) => s.trim())
    .refine((s) => s.length >= 2, "Name must be at least 2 characters")
    .refine((s) => s.length <= 50, "Name must be less than 50 characters"),
  username: z
    .string()
    .min(1, "Username is required")
    .refine((s) => s.trim().length >= 3, "Username must be at least 3 characters")
    .refine((s) => s.trim().length <= 30, "Username must be less than 30 characters")
    .refine((s) => /^[a-z0-9_-]+$/.test(s.trim().toLowerCase()), "Username can only contain lowercase letters, numbers, hyphens, and underscores"),
  bio: z.string().max(MAX_BIO_LENGTH).optional(),
  location: z.string().optional(),
  company: z.string().optional(),
  jobTitle: z.string().optional(),
  website: optionalUrl,
  twitterUrl: optionalUrl,
  githubUrl: optionalUrl,
  linkedinUrl: optionalUrl,
  coverImageUrl: z.union([z.string().url(), z.literal("")]).optional(),
});

type EditProfileFormValues = z.infer<typeof editProfileSchema>;

function getDefaultValues(profile: UserProfile): EditProfileFormValues {
  return {
    name: profile.name || "",
    username: profile.username || "",
    bio: profile.bio || "",
    location: profile.location || "",
    company: profile.company || "",
    jobTitle: profile.jobTitle || "",
    website: profile.website || "",
    twitterUrl: profile.socialLinks?.twitter || "",
    githubUrl: profile.socialLinks?.github || "",
    linkedinUrl: profile.socialLinks?.linkedin || "",
    coverImageUrl: profile.coverImage || "",
  };
}

export function EditProfileModal({
  isOpen,
  onClose,
  profile,
  onSave,
  isLoading = false,
}: EditProfileModalProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isDirty },
    reset,
    watch,
    setValue,
  } = useForm<EditProfileFormValues>({
    resolver: zodResolver(editProfileSchema),
    defaultValues: getDefaultValues(profile),
    mode: "onBlur",
  });

  const location = watch("location");
  const bio = watch("bio");
  const coverImageUrl = watch("coverImageUrl");

  // Location autocomplete (Nominatim)
  const [locationSuggestions, setLocationSuggestions] = useState<NominatimResult[]>([]);
  const [locationSuggestionsLoading, setLocationSuggestionsLoading] = useState(false);
  const [locationSuggestionsOpen, setLocationSuggestionsOpen] = useState(false);
  const locationAutocompleteRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const trimmed = (location ?? "").trim();
    if (trimmed.length < LOCATION_MIN_CHARS) {
      setLocationSuggestions([]);
      setLocationSuggestionsOpen(false);
      return;
    }
    const t = setTimeout(() => {
      setLocationSuggestionsLoading(true);
      fetchLocationSuggestions(trimmed)
        .then((results) => {
          setLocationSuggestions(results);
          setLocationSuggestionsOpen(results.length > 0);
        })
        .catch(() => {
          setLocationSuggestions([]);
          setLocationSuggestionsOpen(false);
        })
        .finally(() => setLocationSuggestionsLoading(false));
    }, LOCATION_DEBOUNCE_MS);
    return () => clearTimeout(t);
  }, [location]);

  useEffect(() => {
    if (isOpen) {
      reset(getDefaultValues(profile));
      setLocationSuggestions([]);
      setLocationSuggestionsOpen(false);
    }
  }, [isOpen, profile, reset]);

  const onSubmit = async (data: EditProfileFormValues) => {
    try {
      const payload: UpdateProfileRequest = {
        name: data.name.trim() || undefined,
        username: data.username.trim().toLowerCase() || undefined,
        bio: (data.bio ?? "").trim() || undefined,
        location: (data.location ?? "").trim() || undefined,
        companyName: (data.company ?? "").trim() || undefined,
        jobTitle: (data.jobTitle ?? "").trim() || undefined,
        website: (data.website ?? "").trim() || undefined,
        twitterUrl: (data.twitterUrl ?? "").trim() || undefined,
        githubUrl: (data.githubUrl ?? "").trim() || undefined,
        linkedinUrl: (data.linkedinUrl ?? "").trim() || undefined,
      };
      await onSave(payload);
      toast.success("Profile updated successfully");
      onClose();
    } catch (err) {
      toast.error("Failed to save profile. Please try again.");
      throw err;
    }
  };

  const name = watch("name");

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto p-0 gap-0">
        <motion.div
          initial={{ opacity: 0, scale: 0.96 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.96 }}
          transition={{ duration: 0.2, ease: "easeOut" }}
          className="p-6 pb-4"
        >
          <DialogHeader>
            <DialogTitle className="text-xl font-semibold text-text-primary flex items-center gap-2">
              <User className="w-5 h-5 text-brand-500" />
              Edit Profile
            </DialogTitle>
            <DialogDescription className="text-text-secondary">
              Update your public profile information. Changes will be visible to other users.
            </DialogDescription>
          </DialogHeader>
        </motion.div>

        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col">
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.05, duration: 0.2 }}
            className="space-y-6 py-4 px-6 overflow-y-auto flex-1"
          >
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
                      (name || profile.name || "").charAt(0).toUpperCase()
                    )}
                  </div>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <button
                          type="button"
                          className="absolute -bottom-1 -right-1 w-7 h-7 bg-brand-500 hover:bg-brand-600 rounded-full flex items-center justify-center text-white shadow-lg transition-colors"
                          aria-label="Avatar is synced from your social login"
                        >
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

            <div className="h-6" />

            {/* Basic Info */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
                Basic Information
              </h3>

              <div className="space-y-2">
                <Label htmlFor="name" className="flex items-center gap-2 text-text-secondary">
                  <User className="w-4 h-4" />
                  Display Name *
                </Label>
                <Input
                  id="name"
                  {...register("name")}
                  placeholder="Your full name"
                  className={cn(errors.name && "border-error focus:border-error focus:ring-error/20")}
                  disabled={isLoading}
                />
                {errors.name && (
                  <p className="text-xs text-error flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    {errors.name.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="username" className="flex items-center gap-2 text-text-secondary">
                  <Globe className="w-4 h-4" />
                  Username *
                </Label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted text-sm">@</span>
                  <Input
                    id="username"
                    {...register("username")}
                    placeholder="username"
                    className={cn(
                      "pl-7",
                      errors.username && "border-error focus:border-error focus:ring-error/20"
                    )}
                    disabled={isLoading}
                  />
                </div>
                {errors.username ? (
                  <p className="text-xs text-error flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    {errors.username.message}
                  </p>
                ) : (
                  <p className="text-xs text-text-muted">
                    Your public profile URL: /u/{watch("username") || "username"}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="bio" className="text-text-secondary">Bio</Label>
                  <span className={cn("text-xs", (bio?.length ?? 0) > MAX_BIO_LENGTH ? "text-error" : "text-text-muted")}>
                    {bio?.length ?? 0}/{MAX_BIO_LENGTH}
                  </span>
                </div>
                <Textarea
                  id="bio"
                  {...register("bio")}
                  placeholder="Tell others about yourself..."
                  rows={4}
                  disabled={isLoading}
                  onChange={(e) => {
                    const v = e.target.value.slice(0, MAX_BIO_LENGTH);
                    setValue("bio", v, { shouldValidate: true });
                  }}
                />
              </div>
            </div>

            {/* Location & Work */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
                Location & Work
              </h3>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2 relative" ref={locationAutocompleteRef}>
                  <Label htmlFor="location" className="flex items-center gap-2 text-text-secondary">
                    <MapPin className="w-4 h-4" />
                    Location
                  </Label>
                  <div className="relative">
                    <Input
                      id="location"
                      {...register("location")}
                      onFocus={() => locationSuggestions.length > 0 && setLocationSuggestionsOpen(true)}
                      onBlur={() => setTimeout(() => setLocationSuggestionsOpen(false), 200)}
                      placeholder="City, Country"
                      disabled={isLoading}
                      autoComplete="off"
                      className="pr-9"
                    />
                    {locationSuggestionsLoading && (
                      <span className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none">
                        <Loader2 className="w-4 h-4 animate-spin text-text-muted" />
                      </span>
                    )}
                  </div>
                  <AnimatePresence>
                    {locationSuggestionsOpen && locationSuggestions.length > 0 && (
                      <motion.ul
                        initial={{ opacity: 0, y: -4 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -4 }}
                        transition={{ duration: 0.15 }}
                        className="absolute z-50 w-full mt-1 py-1 bg-bg-elevated border border-border rounded-md shadow-lg max-h-48 overflow-auto"
                      >
                        {locationSuggestions.map((item) => (
                          <li key={item.place_id}>
                            <button
                              type="button"
                              className="w-full px-3 py-2 text-left text-sm text-text-primary hover:bg-bg-hover focus:bg-bg-hover focus:outline-none"
                              onMouseDown={(e) => {
                                e.preventDefault();
                                setValue("location", formatLocationLabel(item), { shouldValidate: true });
                                setLocationSuggestionsOpen(false);
                              }}
                            >
                              {formatLocationLabel(item)}
                            </button>
                          </li>
                        ))}
                      </motion.ul>
                    )}
                  </AnimatePresence>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="company" className="flex items-center gap-2 text-text-secondary">
                    <Building2 className="w-4 h-4" />
                    Company
                  </Label>
                  <Input
                    id="company"
                    {...register("company")}
                    placeholder="Company or organization"
                    disabled={isLoading}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="jobTitle" className="flex items-center gap-2 text-text-secondary">
                  <Briefcase className="w-4 h-4" />
                  Job Title
                </Label>
                <Input
                  id="jobTitle"
                  {...register("jobTitle")}
                  placeholder="e.g., Senior Software Engineer"
                  disabled={isLoading}
                />
              </div>
            </div>

            {/* Online Presence */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-text-primary uppercase tracking-wider">
                Online Presence
              </h3>

              <div className="space-y-2">
                <Label htmlFor="website" className="flex items-center gap-2 text-text-secondary">
                  <Globe className="w-4 h-4" />
                  Personal Website
                </Label>
                <Input
                  id="website"
                  {...register("website")}
                  placeholder="https://yourwebsite.com"
                  className={cn(errors.website && "border-error focus:border-error focus:ring-error/20")}
                  disabled={isLoading}
                />
                {errors.website && (
                  <p className="text-xs text-error flex items-center gap-1">
                    <AlertCircle className="w-3 h-3" />
                    {errors.website.message}
                  </p>
                )}
              </div>

              <div className="space-y-3">
                <Label className="text-text-secondary">Social Profiles</Label>

                <div className="space-y-2">
                  <div className="relative">
                    <Github className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                    <Input
                      {...register("githubUrl")}
                      placeholder="https://github.com/username"
                      className={cn("pl-10", errors.githubUrl && "border-error focus:border-error focus:ring-error/20")}
                      disabled={isLoading}
                    />
                  </div>
                  {errors.githubUrl && (
                    <p className="text-xs text-error flex items-center gap-1">
                      <AlertCircle className="w-3 h-3" />
                      {errors.githubUrl.message}
                    </p>
                  )}
                </div>

                <div className="space-y-2">
                  <div className="relative">
                    <Twitter className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                    <Input
                      {...register("twitterUrl")}
                      placeholder="https://twitter.com/username"
                      className={cn("pl-10", errors.twitterUrl && "border-error focus:border-error focus:ring-error/20")}
                      disabled={isLoading}
                    />
                  </div>
                  {errors.twitterUrl && (
                    <p className="text-xs text-error flex items-center gap-1">
                      <AlertCircle className="w-3 h-3" />
                      {errors.twitterUrl.message}
                    </p>
                  )}
                </div>

                <div className="space-y-2">
                  <div className="relative">
                    <Linkedin className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                    <Input
                      {...register("linkedinUrl")}
                      placeholder="https://linkedin.com/in/username"
                      className={cn("pl-10", errors.linkedinUrl && "border-error focus:border-error focus:ring-error/20")}
                      disabled={isLoading}
                    />
                  </div>
                  {errors.linkedinUrl && (
                    <p className="text-xs text-error flex items-center gap-1">
                      <AlertCircle className="w-3 h-3" />
                      {errors.linkedinUrl.message}
                    </p>
                  )}
                </div>
              </div>
            </div>

            {/* Cover Image */}
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
                  {...register("coverImageUrl")}
                  placeholder="https://example.com/cover-image.jpg"
                  disabled={isLoading}
                />
                <p className="text-xs text-text-muted">
                  Recommended size: 1500x500 pixels. Leave empty for a gradient background.
                </p>
              </div>
            </div>
          </motion.div>

          {/* Footer */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.1 }}
            className="flex flex-col-reverse sm:flex-row sm:justify-end gap-3 pt-4 px-6 pb-6 border-t border-border-subtle"
          >
            <Button type="button" variant="outline" onClick={onClose} disabled={isLoading} className="w-full sm:w-auto">
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isLoading || !isDirty}
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
          </motion.div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default EditProfileModal;
