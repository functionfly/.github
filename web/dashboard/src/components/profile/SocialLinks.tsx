/**
 * SocialLinks Component
 *
 * Displays social media links with icons. Supports compact (icons only) and
 * expanded (icons + text) variants. Links open in new tabs.
 *
 * @example
 * <SocialLinks
 *   links={{
 *     github: "https://github.com/username",
 *     twitter: "https://twitter.com/username",
 *     linkedin: "https://linkedin.com/in/username",
 *     website: "https://example.com"
 *   }}
 *   variant="expanded"
 * />
 *
 * @example
 * <SocialLinks links={socialLinks} variant="compact" />
 */

import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { SocialLinks as SocialLinksType } from "@/types";
import {
  Globe,
  MessageCircle,
  ExternalLink,
  type LucideIcon,
} from "lucide-react";
import { Icon } from '@iconify/react';

interface SocialLinksProps {
  links: SocialLinksType;
  variant?: "compact" | "expanded";
  className?: string;
  iconClassName?: string;
  showLabels?: boolean;
}

interface SocialLinkConfig {
  key: keyof SocialLinksType;
  icon: LucideIcon | string;
  iconType: 'lucide' | 'iconify';
  label: string;
  color: string;
  hoverColor: string;
  domain?: string;
}

const SOCIAL_CONFIGS: SocialLinkConfig[] = [
  {
    key: "github",
    icon: "simple-icons:github",
    iconType: 'iconify',
    label: "GitHub",
    color: "text-text-secondary",
    hoverColor: "hover:text-white hover:bg-gray-800",
    domain: "github.com",
  },
  {
    key: "twitter",
    icon: "simple-icons:x",
    iconType: 'iconify',
    label: "Twitter",
    color: "text-text-secondary",
    hoverColor: "hover:text-white hover:bg-sky-500",
    domain: "twitter.com",
  },
  {
    key: "linkedin",
    icon: "simple-icons:linkedin",
    iconType: 'iconify',
    label: "LinkedIn",
    color: "text-text-secondary",
    hoverColor: "hover:text-white hover:bg-blue-600",
    domain: "linkedin.com",
  },
  {
    key: "website",
    icon: Globe,
    iconType: 'lucide',
    label: "Website",
    color: "text-text-secondary",
    hoverColor: "hover:text-white hover:bg-brand-500",
  },
  {
    key: "discord",
    icon: MessageCircle,
    iconType: 'lucide',
    label: "Discord",
    color: "text-text-secondary",
    hoverColor: "hover:text-white hover:bg-indigo-500",
    domain: "discord.com",
  },
  {
    key: "devto",
    icon: ExternalLink,
    iconType: 'lucide',
    label: "Dev.to",
    color: "text-text-secondary",
    hoverColor: "hover:text-white hover:bg-black",
    domain: "dev.to",
  },
  {
    key: "medium",
    icon: ExternalLink,
    iconType: 'lucide',
    label: "Medium",
    color: "text-text-secondary",
    hoverColor: "hover:text-white hover:bg-green-600",
    domain: "medium.com",
  },
];

// Helper to normalize URL
function normalizeUrl(url: string, config: SocialLinkConfig): string {
  if (!url) return "";

  // If URL already has protocol, return as-is
  if (url.startsWith("http://") || url.startsWith("https://")) {
    return url;
  }

  // If it's just a username/path, construct full URL
  if (config.domain) {
    // Remove leading @ or / if present
    const cleanPath = url.replace(/^[@/]/, "");
    return `https://${config.domain}/${cleanPath}`;
  }

  // For website, add https if missing
  return `https://${url}`;
}

// Helper to extract display text from URL
function getDisplayText(url: string, config: SocialLinkConfig): string {
  if (!url) return "";

  try {
    const urlObj = new URL(normalizeUrl(url, config));

    // For social profiles, show the path (username)
    if (config.domain && config.key !== "website") {
      const path = urlObj.pathname.replace(/^\//, "");
      return path || config.label;
    }

    // For website, show the hostname
    if (config.key === "website") {
      return urlObj.hostname.replace(/^www\./, "");
    }

    return config.label;
  } catch {
    // If URL parsing fails, return the original
    return url.replace(/^https?:\/\//, "").replace(/^www\./, "");
  }
}

export function SocialLinks({
  links,
  variant = "compact",
  className,
  iconClassName,
  showLabels = true,
}: SocialLinksProps) {
  // Filter to only show links that have values
  const activeLinks = SOCIAL_CONFIGS.filter((config) => {
    const value = links[config.key];
    return value && value.trim().length > 0;
  });

  if (activeLinks.length === 0) {
    return null;
  }

  const isCompact = variant === "compact";

  return (
    <TooltipProvider>
      <div
        className={cn(
          "flex items-center",
          isCompact ? "gap-1" : "gap-2 flex-wrap",
          className
        )}
      >
        {activeLinks.map((config) => {
          const url = links[config.key];
          if (!url) return null;

          const normalizedUrl = normalizeUrl(url, config);
          const displayText = getDisplayText(url, config);

          const renderIcon = () => {
            if (config.iconType === 'iconify') {
              return <Icon icon={config.icon as string} className={cn(
                "w-4 h-4 transition-colors",
                !isCompact && config.color,
                !isCompact && "group-hover:text-brand-400"
              )} />;
            } else {
              const LucideIconComponent = config.icon as LucideIcon;
              return <LucideIconComponent className={cn(
                "w-4 h-4 transition-colors",
                !isCompact && config.color,
                !isCompact && "group-hover:text-brand-400"
              )} />;
            }
          };

          const linkContent = (
            <a
              href={normalizedUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={cn(
                "inline-flex items-center gap-2 transition-all duration-200",
                isCompact
                  ? cn(
                      "w-9 h-9 rounded-lg justify-center",
                      "bg-bg-tertiary border border-border-subtle",
                      config.color,
                      "hover:border-border-focus hover:shadow-md hover:scale-105",
                      config.hoverColor
                    )
                  : cn(
                      "px-3 py-2 rounded-lg",
                      "bg-bg-tertiary border border-border-subtle",
                      "hover:border-border-focus hover:shadow-md",
                      "group"
                    ),
                iconClassName
              )}
            >
              {renderIcon()}
              {!isCompact && showLabels && (
                <div className="flex flex-col items-start min-w-0">
                  <span className="text-xs text-text-muted">{config.label}</span>
                  <span className="text-sm text-text-primary font-medium truncate max-w-[150px]">
                    {displayText}
                  </span>
                </div>
              )}
            </a>
          );

          return (
            <Tooltip key={config.key}>
              <TooltipTrigger asChild>
                {linkContent}
              </TooltipTrigger>
              <TooltipContent side="bottom" className="max-w-xs">
                <div className="flex items-center gap-2">
                  {renderIcon()}
                  <span>
                    {config.label}: {displayText}
                  </span>
                </div>
              </TooltipContent>
            </Tooltip>
          );
        })}
      </div>
    </TooltipProvider>
  );
}

// Compact variant shorthand
export function SocialLinksCompact(props: Omit<SocialLinksProps, "variant">) {
  return <SocialLinks {...props} variant="compact" />;
}

// Expanded variant shorthand
export function SocialLinksExpanded(props: Omit<SocialLinksProps, "variant">) {
  return <SocialLinks {...props} variant="expanded" />;
}

export default SocialLinks;
