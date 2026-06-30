import { useState, useEffect, useRef } from "react";
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
import type { UpdateProfileRequest } from "@/api/users";

interface EditProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  profile: UserProfile;
  onSave: (data: UpdateProfileRequest) => Promise<void>;
  isLoading?: boolean;
}

const MAX_BIO_LENGTH = 500;
const NOMINATIM_URL = "https://nominatim.openstreetmap.org/search";
const LOCATION_DEBOUNCE_MS = 1000;
const LOCATION_MIN_CHARS = 2;
const LOCATION_MAX_RESULTS = 5;

interface NominatimAddress {
  city?: string; town?: string; village?: string; municipality?: string;
  state?: string; state_district?: string; country?: string; country_code?: string;
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
  const params = new URLSearchParams({ q: query, format: "json", limit: String(LOCATION_MAX_RESULTS), addressdetails: "1" });
  const res = await fetch(`${NOMINATIM_URL}?${params}`, {
    headers: { "Accept-Language": "en", "User-Agent": "FunctionFly-Profile-Edit/1.0" },
  });
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

const optionalUrl = z.string().optional().refine(
  (val) => !val || val === "" || /^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/.test(val),
  "Please enter a valid URL"
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

  const [locationSuggestions, setLocationSuggestions] = useState<NominatimResult[]>([]);
  const [locationSuggestionsLoading, setLocationSuggestionsLoading] = useState(false);
  const [locationSuggestionsOpen, setLocationSuggestionsOpen] = useState(false);
  const locationAutocompleteRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const trimmed = (location ?? "").trim();
    if (trimmed.length < LOCATION_MIN_CHARS) { setLocationSuggestions([]); setLocationSuggestionsOpen(false); return; }
    const t = setTimeout(() => {
      setLocationSuggestionsLoading(true);
      fetchLocationSuggestions(trimmed)
        .then((results) => { setLocationSuggestions(results); setLocationSuggestionsOpen(results.length > 0); })
        .catch(() => { setLocationSuggestions([]); setLocationSuggestionsOpen(false); })
        .finally(() => setLocationSuggestionsLoading(false));
    }, LOCATION_DEBOUNCE_MS);
    return () => clearTimeout(t);
  }, [location]);

  useEffect(() => {
    if (isOpen) { reset(getDefaultValues(profile)); setLocationSuggestions([]); setLocationSuggestionsOpen(false); }
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
                  onChange={(e) => setValue("bio", e.target.value.slice(0, MAX_BIO_LENGTH), { shouldValidate: true })} />
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
                        <li key={item.place_id}>
                          <button type="button" className="epm-suggestion" onMouseDown={(e) => { e.preventDefault(); setValue("location", formatLocationLabel(item), { shouldValidate: true }); setLocationSuggestionsOpen(false); }}>
                            {formatLocationLabel(item)}
                          </button>
                        </li>
                      ))}
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
                  <input className={cn("epm-input epm-input--prefixed", errors.githubUrl && "epm-input--error")} {...register("githubUrl")} placeholder="https://github.com/username" disabled={isLoading} />
                </div>
                {errors.githubUrl && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.githubUrl.message}</p>}
              </div>
              <div className="epm-field">
                <div className="epm-input-wrap">
                  <Icon icon="simple-icons:x" className="epm-icon-xs epm-input-icon" />
                  <input className={cn("epm-input epm-input--prefixed", errors.twitterUrl && "epm-input--error")} {...register("twitterUrl")} placeholder="https://twitter.com/username" disabled={isLoading} />
                </div>
                {errors.twitterUrl && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.twitterUrl.message}</p>}
              </div>
              <div className="epm-field">
                <div className="epm-input-wrap">
                  <Icon icon="simple-icons:linkedin" className="epm-icon-xs epm-input-icon" />
                  <input className={cn("epm-input epm-input--prefixed", errors.linkedinUrl && "epm-input--error")} {...register("linkedinUrl")} placeholder="https://linkedin.com/in/username" disabled={isLoading} />
                </div>
                {errors.linkedinUrl && <p className="epm-error"><AlertCircle className="epm-icon-xs" /> {errors.linkedinUrl.message}</p>}
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
