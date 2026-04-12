/**
 * Seed script: insert default blog posts into Neon functionfly database.
 * Run with: npx ts-node --esm seed-to-neon.ts
 */
import { drizzle } from 'drizzle-orm/node-postgres';
import * as pg from 'pg';
import { eq } from 'drizzle-orm';

const DATABASE_URL = 'postgresql://neondb_owner:npg_4PrLqSc2CMKQ@ep-patient-art-aizindo4-pooler.c-4.us-east-1.aws.neon.tech/functionfly?sslmode=require';

// ContentStatus enum
const ContentStatus = {
  DRAFT: 'draft',
  IN_REVIEW: 'in_review',
  APPROVED: 'approved',
  SCHEDULED: 'scheduled',
  PUBLISHED: 'published',
} as const;

type ContentStatus = typeof ContentStatus[keyof typeof ContentStatus];

interface BlogPost {
  title: string;
  slug: string;
  description: string;
  body: unknown[];
  tags: string[];
  status: ContentStatus;
  publishedAt: string;
  seoTitle?: string;
  seoDescription?: string;
  keywords?: string[];
  canonicalUrl?: string;
}

const authorId = 'b1111111-1111-1111-1111-111111111111';
const categories: Record<string, string> = {
  'announcements': 'a1111111-1111-1111-1111-111111111111',
  'trust': 'a2222222-2222-2222-2222-222222222222',
  'security': 'a3333333-3333-3333-3333-333333333333',
  'architecture': 'a4444444-4444-4444-4444-444444444444',
  'ai-agents': 'a5555555-5555-5555-5555-555555555555',
  'tutorials': 'a6666666-6666-6666-6666-666666666666',
  'stories': 'a7777777-7777-7777-7777-777777777777',
};

const DEFAULT_HERO_IMAGE_URL = 'https://functionfly.com/og-default.svg';

const blogPosts: { post: BlogPost; slug: string; categorySlug: string }[] = [
  {
    post: {
      title: 'Welcome to FunctionFly: Serverless Infrastructure for the AI Era',
      slug: 'welcome-to-functionfly',
      description: 'Introducing FunctionFly, the serverless platform purpose-built for AI applications with Flywheel Network, zero-knowledge secrets, and AI-first architecture.',
      body: [
        { type: 'paragraph', children: [{ text: 'Welcome to FunctionFly, the serverless platform designed for the AI era. We\'re building the infrastructure that enables developers to create, deploy, and monetize AI-powered applications with unprecedented ease and security.' }] },
        { type: 'heading', level: 2, children: [{ text: 'What Makes FunctionFly Different?' }] },
        { type: 'paragraph', children: [{ text: 'FunctionFly isn\'t just another serverless platform—it\'s purpose-built for the AI-native world. Our core innovations address the fundamental challenges of building AI applications at scale:' }] },
        { type: 'paragraph', children: [{ text: '🔄 **Flywheel Network™**: A proof-of-execution knowledge network where every function execution becomes verifiable, composable knowledge. Problems are structured, solutions are executable, and AI agents collaborate in open debates.' }] },
        { type: 'paragraph', children: [{ text: '🔐 **Zero-Knowledge Secrets Vault**: Client-side encrypted secrets that scale from free to enterprise-grade without compromising security. Your data never touches our servers in plaintext.' }] },
        { type: 'paragraph', children: [{ text: '🧠 **AI-First Architecture**: Built for AI agents, RAG systems, and autonomous workflows with features like State Fabric for durable memory and CCP (Compute Capsules Protocol) for verifiable execution.' }] },
        { type: 'heading', level: 2, children: [{ text: 'Our Mission' }] },
        { type: 'paragraph', children: [{ text: 'To democratize AI development by providing the infrastructure that makes building AI applications as simple as writing a function. We believe the future belongs to platforms that treat AI as a first-class citizen, not an afterthought.' }] },
        { type: 'paragraph', children: [{ text: 'Whether you\'re building AI microservices, RAG applications, autonomous agents, or the next generation of AI-powered SaaS—we\'ve designed FunctionFly to be your launchpad.' }] },
        { type: 'heading', level: 2, children: [{ text: 'What\'s Next' }] },
        { type: 'paragraph', children: [{ text: 'This blog will be your guide to building on FunctionFly. We\'ll dive deep into our core technologies, share tutorials for common patterns, showcase builder success stories, and keep you updated on our roadmap.' }] },
        { type: 'paragraph', children: [{ text: 'Ready to start building? Head to our dashboard and deploy your first function. The future of AI development starts here.' }] },
      ],
      tags: ['functionfly', 'serverless', 'ai', 'platform', 'introduction', 'flywheel-network', 'secrets-vault'],
      status: ContentStatus.PUBLISHED,
      publishedAt: new Date().toISOString(),
      seoTitle: 'Welcome to FunctionFly | AI-Native Serverless Platform',
      seoDescription: 'FunctionFly is the serverless platform for AI era. Flywheel Network, zero-knowledge secrets, and infrastructure that treats AI as a first-class citizen.',
      keywords: ['functionfly', 'serverless', 'AI platform', 'flywheel network', 'secrets vault', 'AI infrastructure'],
      canonicalUrl: 'https://functionfly.com/blog/welcome-to-functionfly',
    },
    slug: 'welcome-to-functionfly',
    categorySlug: 'announcements',
  },
  {
    post: {
      title: 'Trust Layer for AI Agents: Verification & Attestations',
      slug: 'trust-layer-for-ai-agents',
      description: 'How FunctionFly\'s Trust Layer enables AI agents to verify function execution, build reputation, and operate safely in multi-agent workflows.',
      body: [
        { type: 'paragraph', children: [{ text: 'AI agents need to trust what they execute. FunctionFly\'s Trust Layer provides cryptographic verification of every function execution, enabling AI agents to build verifiable reputation and operate safely.' }] },
        { type: 'heading', level: 2, children: [{ text: 'The Trust Problem' }] },
        { type: 'paragraph', children: [{ text: 'When an AI agent executes a function, how does it know the execution was honest? Traditional platforms offer no verification—agents must trust the platform implicitly.' }] },
        { type: 'heading', level: 2, children: [{ text: 'Our Solution: Verifiable Execution' }] },
        { type: 'paragraph', children: [{ text: 'FunctionFly\'s Trust Layer uses CCP (Compute Capsules Protocol) to create deterministic, verifiable execution receipts. Every function execution generates a cryptographic proof that can be verified by any agent.' }] },
        { type: 'paragraph', children: [{ text: 'This enables trust networks where agents can:- Verify execution integrity- Build reputation over time- Participate in multi-agent debates with confidence- Delegate to other agents with provable guarantees' }] },
      ],
      tags: ['trust', 'ai-agents', 'verification', 'security', 'attestations'],
      status: ContentStatus.PUBLISHED,
      publishedAt: new Date().toISOString(),
      seoTitle: 'Trust Layer for AI Agents | FunctionFly',
      seoDescription: 'Verify AI agent function execution with FunctionFly\'s Trust Layer. Cryptographic proofs, attestations, and reputation for multi-agent workflows.',
    },
    slug: 'trust-layer-for-ai-agents',
    categorySlug: 'trust',
  },
  {
    post: {
      title: 'Zero-Knowledge Secrets Vault: Security That Scales',
      slug: 'secrets-vault',
      description: 'Client-side encrypted secrets management that grows from free tier to enterprise without compromising security or requiring trust in our servers.',
      body: [
        { type: 'paragraph', children: [{ text: 'Your secrets should never touch our servers in plaintext. The Zero-Knowledge Secrets Vault ensures end-to-end encryption with client-side key management.' }] },
        { type: 'heading', level: 2, children: [{ text: 'How It Works' }] },
        { type: 'paragraph', children: [{ text: '1. You generate a passphrase locally—never transmitted to our servers2. Your passphrase derives an AES-256-GCM encryption key3. Secrets are encrypted client-side before storage4. Decryption happens only in your browser or function runtime' }] },
        { type: 'heading', level: 2, children: [{ text: 'Enterprise Features' }] },
        { type: 'paragraph', children: [{ text: '• HSM-backed key management for compliance\n• Audit logs for secret access\n• Role-based access controls\n• Automatic secret rotation\n• Integration with your existing KMS' }] },
      ],
      tags: ['security', 'secrets', 'zero-knowledge', 'encryption'],
      status: ContentStatus.PUBLISHED,
      publishedAt: new Date().toISOString(),
      seoTitle: 'Zero-Knowledge Secrets Vault | FunctionFly',
      seoDescription: 'Client-side encrypted secrets management that scales from free to enterprise. AES-256-GCM encryption with no server-side plaintext.',
    },
    slug: 'secrets-vault',
    categorySlug: 'security',
  },
  {
    post: {
      title: 'State Fabric: Durable State for Serverless Functions',
      slug: 'state-fabric',
      description: 'Build stateful serverless functions with State Fabric—durable state management that survives function invocations and enables complex workflows.',
      body: [
        { type: 'paragraph', children: [{ text: 'Serverless functions are stateless by nature, but real applications need state. State Fabric provides durable state storage that survives function invocations.' }] },
        { type: 'heading', level: 2, children: [{ text: 'Key Features' }] },
        { type: 'paragraph', children: [{ text: '• Atomic operations with optimistic locking\n• Event sourcing for audit trails\n• Sub-millisecond read/write latencies\n• Automatic encryption at rest\n• Built-in TTLs and expiration policies' }] },
        { type: 'heading', level: 2, children: [{ text: 'Use Cases' }] },
        { type: 'paragraph', children: [{ text: 'State Fabric enables patterns like:\n- Shopping carts that persist across sessions\n- Workflow state machines with checkpointing\n- Rate limiting with sliding windows\n- Caching with invalidation patterns\n- Multi-agent coordination state' }] },
      ],
      tags: ['state', 'architecture', 'serverless', 'durable-state'],
      status: ContentStatus.PUBLISHED,
      publishedAt: new Date().toISOString(),
      seoTitle: 'State Fabric | FunctionFly',
      seoDescription: 'Durable state management for serverless functions. Build stateful workflows with sub-millisecond reads and automatic encryption.',
    },
    slug: 'state-fabric',
    categorySlug: 'architecture',
  },
  {
    post: {
      title: 'Getting Started: Deploy Your First FunctionFly Function',
      slug: 'getting-started-deploy-your-first-functionfly-function',
      description: 'Learn to deploy your first serverless function on FunctionFly in under 10 minutes. Build a sentiment analysis API with step-by-step guidance.',
      body: [
        { type: 'paragraph', children: [{ text: 'Ready to start building on FunctionFly? This tutorial will walk you through deploying your first serverless function in under 10 minutes.' }] },
        { type: 'heading', level: 2, children: [{ text: 'Prerequisites' }] },
        { type: 'paragraph', children: [{ text: '• A FunctionFly account\n• Go 1.25+ or downloaded fly binary\n• Basic familiarity with JavaScript/TypeScript' }] },
        { type: 'heading', level: 2, children: [{ text: 'Step 1: Install the CLI' }] },
        { type: 'paragraph', children: [{ text: '```bash\ngo install github.com/functionfly/functionfly/cmd/fly@latest\nfly --version\n```' }] },
        { type: 'heading', level: 2, children: [{ text: 'Step 2: Authenticate' }] },
        { type: 'paragraph', children: [{ text: '```bash\nfly auth login\n```' }] },
        { type: 'heading', level: 2, children: [{ text: 'Step 3: Create and Deploy' }] },
        { type: 'paragraph', children: [{ text: '```bash\nmkdir my-function && cd my-function\nfly init\nfly deploy\n```' }] },
        { type: 'paragraph', children: [{ text: 'Congratulations! Your first function is live. Check out our documentation for more advanced patterns.' }] },
      ],
      tags: ['tutorial', 'getting-started', 'javascript', 'serverless', 'api'],
      status: ContentStatus.PUBLISHED,
      publishedAt: new Date().toISOString(),
      seoTitle: 'Getting Started with FunctionFly | First Function Tutorial',
      seoDescription: 'Deploy your first serverless function on FunctionFly in under 10 minutes. Step-by-step tutorial for building APIs.',
    },
    slug: 'getting-started-deploy-your-first-functionfly-function',
    categorySlug: 'tutorials',
  },
];

async function seed() {
  console.log('🌱 Starting seed to Neon...');
  
  const pool = new pg.Pool({ connectionString: DATABASE_URL });
  const db = drizzle(pool);

  const now = new Date();
  let inserted = 0;

  for (const entry of blogPosts) {
    const categoryId = categories[entry.categorySlug];
    if (!categoryId) {
      console.error(`Unknown category: ${entry.categorySlug}`);
      continue;
    }

    const heroImage = { url: DEFAULT_HERO_IMAGE_URL, alt: entry.post.title };
    
    // Check if post exists
    const existing = await pool.query(
      'SELECT id FROM blog_api_posts WHERE slug = $1',
      [entry.slug]
    );

    if (existing.rows.length === 0) {
      await pool.query(
        `INSERT INTO blog_api_posts 
         (title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, created_at, updated_at, seo_title, seo_description, keywords, canonical_url)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
        [
          entry.post.title,
          entry.post.slug,
          entry.post.description,
          JSON.stringify(entry.post.body),
          authorId,
          categoryId,
          JSON.stringify(entry.post.tags),
          JSON.stringify(heroImage),
          entry.post.status,
          entry.post.publishedAt,
          now,
          now,
          entry.post.seoTitle || null,
          entry.post.seoDescription || null,
          JSON.stringify(entry.post.keywords || []),
          entry.post.canonicalUrl || null,
        ]
      );
      console.log(`✓ Inserted: ${entry.post.title}`);
      inserted++;
    } else {
      console.log(`↻ Exists: ${entry.post.title}`);
    }
  }

  console.log(`\n✅ Seed complete! Inserted ${inserted} new posts.`);
  await pool.end();
}

seed().catch(err => {
  console.error('Seed failed:', err);
  process.exit(1);
});
