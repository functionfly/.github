/**
 * Single source of truth for default launch posts: seed + HTTP verification use the same slugs.
 */
import {
  aiAgentPost,
  slug as aiAgentSlug,
} from "../data/default-posts/ai-agent-integration";
import {
  successStoryPost,
  slug as successStorySlug,
} from "../data/default-posts/builder-success-story";
import {
  ccpPost,
  slug as ccpSlug,
} from "../data/default-posts/compute-capsules-protocol";
import {
  flywheelPost,
  slug as flywheelSlug,
} from "../data/default-posts/flywheel-network";
import {
  tutorialPost,
  slug as tutorialSlug,
} from "../data/default-posts/getting-started-tutorial";
import {
  secretsVaultPost,
  slug as secretsVaultSlug,
} from "../data/default-posts/secrets-vault";
import {
  stateFabricPost,
  slug as stateFabricSlug,
} from "../data/default-posts/state-fabric";
import {
  trustLayerPost,
  slug as trustLayerSlug,
} from "../data/default-posts/trust-layer-for-ai-agents";
import {
  welcomePost,
  slug as welcomeSlug,
} from "../data/default-posts/welcome-functionfly";

export type DefaultBlogCategorySlug =
  | "announcements"
  | "trust"
  | "security"
  | "architecture"
  | "ai-agents"
  | "tutorials"
  | "stories";

export interface DefaultBlogPostEntry {
  post: {
    title: string;
    slug: string;
    description: string;
    body: unknown;
    tags: readonly string[] | string[];
    publishedAt?: string;
    seoTitle?: string | null;
    seoDescription?: string | null;
    keywords?: readonly string[] | string[] | null;
    canonicalUrl?: string | null;
  };
  slug: string;
  categorySlug: DefaultBlogCategorySlug;
}

export const DEFAULT_HERO_IMAGE_URL = "https://functionfly.com/og-default.svg";

export const defaultBlogCategories: {
  slug: DefaultBlogCategorySlug;
  title: string;
  description: string;
  order: number;
}[] = [
  {
    slug: "announcements",
    title: "Announcements",
    description: "Platform news, positioning, and what we are shipping.",
    order: 0,
  },
  {
    slug: "trust",
    title: "Trust & safety",
    description:
      "Verification tiers, attestations, Trust API, and safe agent tooling.",
    order: 1,
  },
  {
    slug: "security",
    title: "Security",
    description: "Secrets, encryption, and zero-knowledge patterns.",
    order: 2,
  },
  {
    slug: "architecture",
    title: "Architecture",
    description: "Protocols, durable state, and verifiable execution.",
    order: 3,
  },
  {
    slug: "ai-agents",
    title: "AI & agents",
    description:
      "Agent-native workflows, integration patterns, and the AI-era stack.",
    order: 4,
  },
  {
    slug: "tutorials",
    title: "Tutorials",
    description: "Step-by-step guides to ship on FunctionFly.",
    order: 5,
  },
  {
    slug: "stories",
    title: "Stories",
    description: "Builder narratives and patterns that worked.",
    order: 6,
  },
];

export const defaultBlogAuthor = {
  slug: "functionfly-team",
  name: "FunctionFly Team",
  role: "Editorial",
  bio: "Product and engineering notes from the FunctionFly team.",
} as const;

export const defaultBlogPostEntries: DefaultBlogPostEntry[] = [
  { post: welcomePost, slug: welcomeSlug, categorySlug: "announcements" },
  { post: trustLayerPost, slug: trustLayerSlug, categorySlug: "trust" },
  { post: secretsVaultPost, slug: secretsVaultSlug, categorySlug: "security" },
  {
    post: stateFabricPost,
    slug: stateFabricSlug,
    categorySlug: "architecture",
  },
  { post: flywheelPost, slug: flywheelSlug, categorySlug: "architecture" },
  { post: ccpPost, slug: ccpSlug, categorySlug: "architecture" },
  { post: aiAgentPost, slug: aiAgentSlug, categorySlug: "ai-agents" },
  { post: tutorialPost, slug: tutorialSlug, categorySlug: "tutorials" },
  { post: successStoryPost, slug: successStorySlug, categorySlug: "stories" },
];

export const defaultBlogPostSlugs: string[] = defaultBlogPostEntries.map(
  (e) => e.slug,
);
