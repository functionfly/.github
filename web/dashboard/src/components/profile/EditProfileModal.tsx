import { useState, useEffect, useRef, useCallback } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import {
  User, MapPin, Building2, Briefcase, Globe, Camera, ImageIcon,
  Loader2, Check, AlertCircle, X,
} from "lucide-react";
import { Icon } from "@iconify/react";
import { cn } from "@/lib/utils";
import { SealedButton, FrameButton } from "@/components/containment";
import type { UserProfile } from "@/types";
import { usersApi, type UpdateProfileRequest, type LocationResult } from "@/api/users";

interface EditProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: UserProfile;
  onSave: (data: UpdateProfileRequest) => Promise<void>;
  isLoading?: boolean;
}

const MAX_BIO_LENGTH = 500;
const LOCATION_DEBOUNCE_MS = 300;
const LOCATION_MIN_CHARS = 2;
const LOCATION_CACHE_TTL_MS = 5 * 60 * 1000;
const LOCATION_CACHE_MAX_ENTRIES = 100;

interface CacheEntry {
  results: LocationResult[];
  timestamp: number;
}

const locationCache = new Map<string, CacheEntry>();

function getCachedLocations(query: string): LocationResult[] | null {
  const key = query.toLowerCase();
  const entry = locationCache.get(key);
  if (!entry) return null;
  if (Date.now() - entry.timestamp > LOCATION_CACHE_TTL_MS) {
    locationCache.delete(key);
    return null;
  }
  return entry.results;
}

function setCachedLocations(query: string, results: LocationResult[]) {
  const key = query.toLowerCase();
  if (locationCache.size >= LOCATION_CACHE_MAX_ENTRIES) {
    const oldest = locationCache.keys().next().value;
    if (oldest !== undefined) locationCache.delete(oldest);
  }
  locationCache.set(key, { results, timestamp: Date.now() });
}

const optionalUrl = z.string().optional().refine(
  (val) => !val || val === "" || /^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/.test(val),
  "Please enter a valid URL"
);

const optionalSocialHandle = z.string().optional().refine(
  (val) => !val || val === "" || /^[a-zA-Z0-9_.-]+$/.test(val),
  "Only letters, numbers, dots, hyphens, and underscores"
);

const editProfileSchema = z.object({
  name: z.string().min(1, "Name is required").transform((s) => s.trim())
    .refine((s) => s.length >= 2, "Name must be at least 2 characters")
    .refine((s) => s.length <= 50, "Name must be less than 50 characters"),
  username: z.string().min(1, "Username is required")
    .refine((s) => s.trim().length >= 3, "Username must be at least 3 characters")
    .refine((s) => s.trim().length <= 30, "Username must be less than 30 characters")
    .refine((s) => /^[a-z0-9_-]+$/.test(s.trim().toLowerCase()), "Username can only contain lowercase letters, numbers, hyphens, and underscores"),
  bio: z.string().max(MAX_BIO_LENGTH).optional(),
  location: z.string().optional(),
  company: z.string().optional(),
  jobTitle: z.string().optional(),
  website: optionalUrl,
  githubHandle: optionalSocialHandle,
  twitterHandle: optionalSocialHandle,
  linkedinHandle: optionalSocialHandle,
  coverImageUrl: z.union([z.string().url(), z.literal("")]).optional(),
});

type EditProfileFormValues = z.infer<typeof editProfileSchema>;

function extractSocialHandle(url: string | undefined, prefix: string): string {
  if (!url) return "";
  try {
    const u = url.startsWith("http") ? url : `https://${url}`;
    const parsed = new URL(u);
    const path = parsed.pathname.replace(/^\/|\/$/g, "");
    if (prefix && parsed.hostname.includes(prefix)) return path;
    return path || "";
  } catch {
    return url.replace(/^https?:\/\/[^/]+\//, "").replace(/\/$/, "");
  }
}

function getDefaultValues(profile: UserProfile): EditProfileFormValues {
  return {
    name: profile.name || "",
    username: profile.username || "",
    bio: profile.bio || "",
    location: profile.location || "",
    company: profile.company || "",
    jobTitle: profile.jobTitle || "",
    website: profile.website || "",
    githubHandle: extractSocialHandle(profile.socialLinks?.github, "github.com"),
    twitterHandle: extractSocialHandle(profile.socialLinks?.twitter, "twitter.com") || extractSocialHandle(profile.socialLinks?.twitter, "x.com"),
    linkedinHandle: extractSocialHandle(profile.socialLinks?.linkedin, "linkedin.com"),
    coverImageUrl: profile.coverImage || "",
  };
}

export function EditProfileModal({ isOpen, onClose, profile, onSave, isLoading = false }: EditProfileModalProps) {
  const { register, handleSubmit, formState: { errors, isDirty }, reset, watch, setValue } = useForm<EditProfileFormValues>({
    resolver: zodResolver(editProfileSchema),
    defaultValues: getDefaultValues(profile),
    mode: "onBlur",
  });

  const location = watch("location");
  const bio = watch("bio");
  const coverImageUrl = watch("coverImageUrl");
  const name = watch("name");

  const [locationSuggestions, setLocationSuggestions] = useState<LocationResult[]>([]);
  const [locationSuggestionsLoading, setLocationSuggestionsLoading] = useState(false);
  const [locationSuggestionsOpen, setLocationSuggestionsOpen] = useState(false);
  const locationAutocompleteRef = useRef<HTMLDivElement>(null);
  const locationAbortRef = useRef<AbortController | null>(null);

  const fetchLocations = useCallback(async (query: string) => {
    const cached = getCachedLocations(query);
    if (cached) {
      setLocationSuggestions(cached);
      setLocationSuggestionsOpen(cached.length > 0);
      return;
    }

    if (locationAbortRef.current) {
      locationAbortRef.current.abort();
    }
    const controller = new AbortController();
    locationAbortRef.current = controller;

    setLocationSuggestionsLoading(true);
    try {
      const res = await usersApi.searchLocations(query);
      if (controller.signal.aborted) return;
      const results = res.locations ?? [];
      setCachedLocations(query, results);
      setLocationSuggestions(results);
      setLocationSuggestionsOpen(results.length > 0);
    } catch {
      if (!controller.signal.aborted) {
        setLocationSuggestions([]);
        setLocationSuggestionsOpen(false);
      }
    } finally {
      if (!controller.signal.aborted) {
        setLocationSuggestionsLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    const trimmed = (location ?? "").trim();
    if (trimmed.length < LOCATION_MIN_CHARS) {
      setLocationSuggestions([]);
      setLocationSuggestionsOpen(false);
      return;
    }
    const t = setTimeout(() => fetchLocations(trimmed), LOCATION_DEBOUNCE_MS);
    return () => clearTimeout(t);
  }, [location, fetchLocations]);

  useEffect(() => {
    if (isOpen) {
      reset(getDefaultValues(profile));
      setLocationSuggestions([]);
      setLocationSuggestionsOpen(false);
    } else if (locationAbortRef.current) {
      locationAbortRef.current.abort();
    }
  }, [isOpen, profile, reset]);

  const onSubmit = async (data: EditProfileFormValues) => {
    try {
      const handleToUrl = (handle: string | undefined, base: string) => {
        const h = (handle ?? "").trim();
        return h ? `${base}/${h}` : undefined;
      };
      const payload: UpdateProfileRequest = {
        name: data.name.trim() || undefined,
        username: data.username.trim().toLowerCase() || undefined,
        bio: (data.bio ?? "").trim() || undefined,
        location: (data.location ?? "").trim() || undefined,
        companyName: (data.company ?? "").trim() || undefined,
        jobTitle: (data.jobTitle ?? "").trim() || undefined,
        website: (data.website ?? "").trim() || undefined,
        twitterUrl: handleToUrl(data.twitterHandle, "https://x.com"),
        githubUrl: handleToUrl(data.githubHandle, "https://github.com"),
        linkedinUrl: handleToUrl(data.linkedinHandle, "https://linkedin.com/in"),
      };
      await onSave(payload);
      toast.success("Profile updated successfully");
      onClose();
    } catch (err) {
      toast.error("Failed to save profile. Please try again.");
      throw err;
    }
  };

  if (!isOpen) return null;

  return (
    <div className="epm-overlay" onClick={onClose}>
      <div className="epm-modal" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="epm-header">
          <div>
            <h2 className="epm-title"><User className="epm-icon-sm epm-icon-accent" /> Edit Profile</h2>
            <p className="epm-desc">Update your public profile information. Changes will be visible to other users.</p>
          </div>
          <button className="epm-close" onClick={onClose} aria-label="Close"><X className="epm-icon-sm" /></button>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="epm-form">
          {/* Avatar & Cover */}
          <div className="epm-cover">
            <div className="epm-cover__bg">
              {coverImageUrl ? <img src={coverImageUrl} alt="Cover" className="epm-cover__img" /> : <div className="epm-cover__gradient" />}
              <div className="epm-cover__overlay" />
            </div>
            <div className="epm-avatar-wrap">
              <div className="epm-avatar">
                {profile.avatar ? <img src={profile.avatar} alt={profile.name} className="epm-avatar__img" /> : (name || profile.name || "").charAt(0).toUpperCase()}
              </div>
              <button type="button" className="epm-avatar-btn" aria-label="Avatar is synced from your social login">
                <Camera className="epm-icon-xs" />
              </button>
            </div>
          </div>

          <div className="epm-body">
            {/* Basic Info */}
            <div className="epm-section">
              <h3 className="epm-section-title">Basic Information</h3>
              <div className="epm-field">
                <label className="epm-label"><User className="epm-icon-xs" /> Display Name *</label>
                <input className={cn("epm-input", errors.name && "epm-input--error")} {...register("name")} placeholder="Your full name" disabled={isLoading} />
                {errors.name && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.name.message}</p>}
              </div>
              <div className="epm-field">
                <label className="epm-label"><Globe className="epm-icon-xs" /> Username *</label>
                <div className="epm-input-wrap">
                  <span className="epm-input-prefix">@</span>
                  <input className={cn("epm-input epm-input--prefixed", errors.username && "epm-input--error")} {...register("username")} placeholder="username" disabled={isLoading} />
                </div>
                {errors.username ? <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.username.message}</p> : <p className="epm-hint">Your public profile URL: /u/{watch("username") || "username"}</p>}
              </div>
              <div className="epm-field">
                <div className="epm-field-row">
                  <label className="epm-label">Bio</label>
                  <span className={cn("epm-counter", (bio?.length ?? 0) > MAX_BIO_LENGTH && "epm-counter--error")}>{bio?.length ?? 0}/{MAX_BIO_LENGTH}</span>
                </div>
                <textarea className="epm-textarea" {...register("bio")} placeholder="Tell others about yourself..." rows={4} disabled={isLoading}
                  onChange={(e) => setValue("bio", e.target.value.slice(0, MAX_BIO_LENGTH), { shouldValidate: true, shouldDirty: true })} />
              </div>
            </div>

            {/* Location & Work */}
            <div className="epm-section">
              <h3 className="epm-section-title">Location & Work</h3>
              <div className="epm-grid-2">
                <div className="epm-field" ref={locationAutocompleteRef}>
                  <label className="epm-label"><MapPin className="epm-icon-xs" /> Location</label>
                  <div className="epm-input-wrap">
                    <input className="epm-input" {...register("location")} placeholder="City, Country" disabled={isLoading} autoComplete="off"
                      onFocus={() => locationSuggestions.length > 0 && setLocationSuggestionsOpen(true)}
                      onBlur={() => setTimeout(() => setLocationSuggestionsOpen(false), 200)} />
                    {locationSuggestionsLoading && <Loader2 className="epm-icon-xs epm-spin epm-input-spinner" />}
                  </div>
                  {locationSuggestionsOpen && locationSuggestions.length > 0 && (
                    <ul className="epm-suggestions">
                      {locationSuggestions.map((item) => (
                        <li key={item.placeId}>
                          <button type="button" className="epm-suggestion" onMouseDown={(e) => { e.preventDefault(); setValue("location", item.label, { shouldValidate: true, shouldDirty: true }); setLocationSuggestionsOpen(false); }}>
                            {item.label}
                          </button>
                        </li>
                      ))}
                      <li className="epm-suggestions__attribution">
                        <span>&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener noreferrer">OpenStreetMap</a> contributors</span>
                      </li>
                    </ul>
                  )}
                </div>
                <div className="epm-field">
                  <label className="epm-label"><Building2 className="epm-icon-xs" /> Company</label>
                  <input className="epm-input" {...register("company")} placeholder="Company or organization" disabled={isLoading} />
                </div>
              </div>
              <div className="epm-field">
                <label className="epm-label"><Briefcase className="epm-icon-xs" /> Job Title</label>
                <input className="epm-input" {...register("jobTitle")} placeholder="e.g., Senior Software Engineer" disabled={isLoading} />
              </div>
            </div>

            {/* Online Presence */}
            <div className="epm-section">
              <h3 className="epm-section-title">Online Presence</h3>
              <div className="epm-field">
                <label className="epm-label"><Globe className="epm-icon-xs" /> Personal Website</label>
                <input className={cn("epm-input", errors.website && "epm-input--error")} {...register("website")} placeholder="https://yourwebsite.com" disabled={isLoading} />
                {errors.website && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.website.message}</p>}
              </div>
              <label className="epm-label">Social Profiles</label>
              <div className="epm-field">
                <div className="epm-input-wrap">
                  <Icon icon="simple-icons:github" className="epm-icon-xs epm-input-icon" />
                  <span className="epm-input-prefix epm-input-prefix--social">github.com/</span>
                  <input className={cn("epm-input epm-input--social", errors.githubHandle && "epm-input--error")} {...register("githubHandle")} placeholder="username" disabled={isLoading} />
                </div>
                {errors.githubHandle && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.githubHandle.message}</p>}
              </div>
              <div className="epm-field">
                <div className="epm-input-wrap">
                  <Icon icon="simple-icons:x" className="epm-icon-xs epm-input-icon" />
                  <span className="epm-input-prefix epm-input-prefix--social">x.com/</span>
                  <input className={cn("epm-input epm-input--social", errors.twitterHandle && "epm-input--error")} {...register("twitterHandle")} placeholder="username" disabled={isLoading} />
                </div>
                {errors.twitterHandle && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.twitterHandle.message}</p>}
              </div>
              <div className="epm-field">
                <div className="epm-input-wrap">
                  <Icon icon="simple-icons:linkedin" className="epm-icon-xs epm-input-icon" />
                  <span className="epm-input-prefix epm-input-prefix--social">linkedin.com/in/</span>
                  <input className={cn("epm-input epm-input--social", errors.linkedinHandle && "epm-input--error")} {...register("linkedinHandle")} placeholder="username" disabled={isLoading} />
                </div>
                {errors.linkedinHandle && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.linkedinHandle.message}</p>}
              </div>
            </div>

            {/* Cover Image */}
            <div className="epm-section">
              <h3 className="epm-section-title">Profile Customization</h3>
              <div className="epm-field">
                <label className="epm-label"><ImageIcon className="epm-icon-xs" /> Cover Image URL</label>
                <input className="epm-input" {...register("coverImageUrl")} placeholder="https://example.com/cover-image.jpg" disabled={isLoading} />
                <p className="epm-hint">Recommended size: 1500x500 pixels. Leave empty for a gradient background.</p>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="epm-footer">
            <FrameButton type="button" onClick={onClose} disabled={isLoading}>Cancel</FrameButton>
            <SealedButton type="submit" disabled={isLoading || !isDirty} loading={isLoading} iconLeft={isLoading ? <Loader2 className="epm-icon-sm epm-spin" /> : <Check className="epm-icon-sm" />}>
              {isLoading ? "Saving..." : "Save Changes"}
            </SealedButton>
          </div>
        </form>
      </div>
    </div>
  );
}

export default EditProfileModal;
