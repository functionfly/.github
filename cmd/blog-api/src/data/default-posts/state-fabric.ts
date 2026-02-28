/**
 * Default blog post: Introducing State Fabric
 * Used by the NestJS blog API seed. Body is rich-text (Portable Text–style blocks).
 */
import { ContentStatus } from '../../modules/blog/dto/blog-post.dto';

export const slug = 'introducing-state-fabric';

/** Rich-text body: array of blocks with type and children */
const body = [
  {
    type: 'paragraph',
    children: [{ text: 'State Fabric is FunctionFly\'s composable durable state layer for stateless functions. It gives your serverless workloads a reliable, globally addressable memory layer—so agents, workflows, and functions can share state without losing it.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Why State Fabric?' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Today\'s functions are stateless by design. That\'s great for scale, but it makes coordination, agent memory, and multi-step workflows hard. State Fabric adds a first-class state layer: stores, snapshots, event logs, and triggers—all addressable via simple URIs like state://tenant/cart or memory://tenant/agent-1.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'What you get' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Durable storage (PostgreSQL + pgvector) as the source of truth, optional Redis for hot cache, deterministic replay, and snapshot management. State is bound to your fx:// function identity, so you get sticky infrastructure and storage-based value on top of usage-based compute.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Built for AI agents' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'State Fabric is built with AI agents in mind: working memory, long-term memory, and semantic search over state. Use it for session state, cart state, agent context, or any shared state your serverless stack needs.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'To try it, open State Fabric in the dashboard or read the architecture doc in the repo. We\'re shipping more features—replay, triggers, and tighter registry integration—over the next releases.' }],
  },
];

export const stateFabricPost = {
  title: 'Introducing State Fabric: Composable Durable State for Stateless Functions',
  slug,
  description: 'State Fabric is FunctionFly\'s durable state layer for serverless: globally addressable stores, snapshots, and memory for agents and workflows.',
  body,
  tags: ['state-fabric', 'serverless', 'agents', 'durable-state', 'functionfly'],
  status: ContentStatus.PUBLISHED,
  publishedAt: new Date().toISOString(),
  seoTitle: 'Introducing State Fabric | FunctionFly Blog',
  seoDescription: 'State Fabric adds composable durable state to stateless functions: stores, snapshots, and memory for AI agents and serverless workflows.',
  keywords: ['state fabric', 'serverless', 'durable state', 'AI agents', 'functionfly', 'serverless state'],
  canonicalUrl: 'https://functionfly.com/blog/introducing-state-fabric',
} as const;

export type StateFabricPostPayload = typeof stateFabricPost;
