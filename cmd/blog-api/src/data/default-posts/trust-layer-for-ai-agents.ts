/**
 * Default blog post: Trust layer for AI agents (aligned with web/site trust.astro).
 */
import { ContentStatus } from '../../modules/blog/dto/blog-post.dto';

export const slug = 'trust-layer-for-ai-agents';

const body = [
  {
    type: 'paragraph',
    children: [
      {
        text: 'Agents do not need more tools—they need tools they can trust. FunctionFly turns functions into verified, signed, auditable building blocks with execution-backed trust scores, a zero-knowledge vault, and a Trust API you can query from your own policies.',
      },
    ],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Verification levels' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Each level adds deeper assurance before a function can be treated as trusted in your stack.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '**L1 — Format checks:** Validate manifest structure and I/O schema so agents only call well-described tools.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '**L2 — Security scans:** Scan for risky behaviors and validate capability constraints before exposure to production agents.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '**L3 — Code review:** Manual and automated review of safety-relevant aspects for higher-assurance workloads.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '**L4 — Platform verified:** Signed and reviewed as official or recommended tooling—the strongest default for agent marketplaces.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Signing, attestations, and revocation' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Trust is portable when verification becomes an immutable record. Artifacts carry platform signatures; attestations capture what was checked and when; if a tool is flagged, it can be downgraded or removed from trusted pools so agents stop selecting it.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'What agents see: a trust score and policy-relevant metadata—capabilities, constraints, and verification level—so your runtime can enforce least privilege instead of guessing.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Execution-backed trust' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Trust scores compound from real execution history. “It ran” and “it ran safely” both matter. That pairs naturally with deterministic execution models (like Compute Capsules) and networks that treat proven runs as knowledge—see our posts on CCP and Flywheel for the full picture.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Zero-knowledge vault and Trust API' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Secrets are encrypted client-side; FunctionFly stores ciphertext only, so agent tools can use credentials without the platform ever seeing plaintext. The Trust API exposes attestations, trust scores, and revocation state on a usage-based model so dynamic agent workloads can adapt trust as they run.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Ready to build? Open the dashboard to publish functions under explicit trust policies, or read the dedicated posts on Secrets Vault and AI agent integration for next steps.' }],
  },
];

export const trustLayerPost = {
  title: 'The Trust Layer for AI Agents: Verification, Signing, and Safe Tooling',
  slug,
  description:
    'How FunctionFly combines verification tiers, signed attestations, execution-backed trust scores, a zero-knowledge vault, and the Trust API so agents pick tools on policy—not vibes.',
  body,
  tags: [
    'trust',
    'verification',
    'ai-agents',
    'security',
    'functionfly',
    'trust-api',
    'attestations',
  ],
  status: ContentStatus.PUBLISHED,
  publishedAt: new Date().toISOString(),
  seoTitle: 'Trust Layer for AI Agents | FunctionFly',
  seoDescription:
    'Verification levels L1–L4, signing and revocation, execution-backed trust scores, zero-knowledge vault, and Trust API—why FunctionFly is built for safe agent tooling.',
  keywords: [
    'AI agent trust',
    'tool verification',
    'signed functions',
    'Trust API',
    'zero-knowledge secrets',
    'FunctionFly',
  ],
  canonicalUrl: 'https://functionfly.com/blog/trust-layer-for-ai-agents',
} as const;

export type TrustLayerPostPayload = typeof trustLayerPost;
