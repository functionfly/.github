-- Seed script for additional blog posts
-- Run: psql -h localhost -p 5432 -U postgres -d functionfly -f migrations/seed_blog_posts_extra.sql

-- State Fabric post
INSERT INTO blog_posts (id, title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, updated_at, created_at, seo_title, seo_description, keywords, canonical_url)
VALUES (
  'd1e2f3a4-b5c6-7890-abcd-ef1234567893',
  'State Fabric: The Memory Layer for AI Applications',
  'state-fabric',
  'State Fabric provides durable, consistent state management for distributed AI applications. Learn how it enables reliable agent memory and workflow state.',
  '[{"type":"paragraph","children":[{"text":"State Fabric is FunctionFly approach to durable state management for distributed AI applications. It provides a consistent, fault-tolerant state layer that enables agents to maintain context across restarts, failures, and scaling events."}]},{"type":"heading","level":2,"children":[{"text":"Why State Matters for AI"}]},{"type":"paragraph","children":[{"text":"Traditional stateless functions lose all context when they complete. For AI applications this means every invocation starts from scratch, rebuilding context from scratch. State Fabric solves this by providing durable, consistent state that persists across invocations."}]},{"type":"heading","level":2,"children":[{"text":"Key Features"}]},{"type":"bulleted-list","children":[{"text":"Automatic checkpointing at every decision point"},{"text":"Consistent reads across distributed replicas"},{"text":"Optimistic concurrency control for concurrent updates"},{"text":"TTL-based automatic cleanup for expired state"},{"text":"Event sourcing for complete audit trails"}]},{"type":"heading","level":2,"children":[{"text":"Architecture"}]},{"type":"paragraph","children":[{"text":"State Fabric uses a leader-follower replication model with Raft consensus for state updates. Reads can be served from any replica for scalability, while writes are always routed to the leader for consistency."}]}]',
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'c1b2c3d4-e5f6-7890-abcd-ef1234567894',
  '["state-fabric", "state-management", "distributed-systems", "ai", "architecture"]'::jsonb,
  '{"url":"https://functionfly.com/og-default.svg","alt":"State Fabric"}'::jsonb,
  'published',
  NOW(),
  NOW(),
  NOW(),
  'State Fabric: Durable State for AI Applications | FunctionFly',
  'State Fabric provides durable, consistent state management for distributed AI applications. Learn how it enables reliable agent memory and workflow state.',
  '["state fabric", "AI state management", "durable state", "agent memory", "distributed AI"]',
  'https://functionfly.com/blog/state-fabric'
) ON CONFLICT (slug) DO NOTHING;

-- Trust Layer post
INSERT INTO blog_posts (id, title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, updated_at, created_at, seo_title, seo_description, keywords, canonical_url)
VALUES (
  'd1e2f3a4-b5c6-7890-abcd-ef1234567894',
  'Trust Layer for AI Agents: Building Verifiable AI Workflows',
  'trust-layer-for-ai-agents',
  'The Trust Layer provides attestation and verification for AI agent actions. Learn how to build trustworthy AI systems with verifiable execution.',
  '[{"type":"paragraph","children":[{"text":"As AI agents become more autonomous, the need for trust and verification grows. The Trust Layer provides cryptographic attestation for AI agent actions, enabling verifiable execution trails."}]},{"type":"heading","level":2,"children":[{"text":"The Trust Problem"}]},{"type":"paragraph","children":[{"text":"When an AI agent executes actions on your behalf, how do you verify those actions were performed correctly? How do you audit what the agent did? The Trust Layer solves this by providing immutable, verifiable attestations."}]},{"type":"heading","level":2,"children":[{"text":"Verification Tiers"}]},{"type":"paragraph","children":[{"text":"The Trust Layer supports multiple verification tiers from basic execution logs to full cryptographic proofs. Higher tiers provide stronger guarantees but require more computational overhead."}]}]',
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'c1b2c3d4-e5f6-7890-abcd-ef1234567892',
  '["trust", "verification", "ai-agents", "security", "attestation"]'::jsonb,
  '{"url":"https://functionfly.com/og-default.svg","alt":"Trust Layer"}'::jsonb,
  'published',
  NOW(),
  NOW(),
  NOW(),
  'Trust Layer for AI Agents | FunctionFly',
  'The Trust Layer provides attestation and verification for AI agent actions. Learn how to build trustworthy AI systems.',
  '["trust layer", "AI verification", "agent attestation", "AI security", "verifiable AI"]',
  'https://functionfly.com/blog/trust-layer-for-ai-agents'
) ON CONFLICT (slug) DO NOTHING;

-- Secrets Vault post
INSERT INTO blog_posts (id, title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, updated_at, created_at, seo_title, seo_description, keywords, canonical_url)
VALUES (
  'd1e2f3a4-b5c6-7890-abcd-ef1234567895',
  'Zero-Knowledge Secrets Vault: Client-Side Encryption for AI Applications',
  'zero-knowledge-secrets-vault',
  'The Secrets Vault uses zero-knowledge encryption to protect sensitive data. Learn how client-side encryption keeps your secrets safe.',
  '[{"type":"paragraph","children":[{"text":"When running AI agents that handle sensitive data, security is paramount. The Zero-Knowledge Secrets Vault ensures that even FunctionFly cannot access your secrets—encryption happens entirely client-side."}]},{"type":"heading","level":2,"children":[{"text":"How It Works"}]},{"type":"paragraph","children":[{"text":"Your application encrypts secrets using a key only you control before sending to FunctionFly. We store the encrypted blob without ever seeing the plaintext. Decryption also happens client-side."}]},{"type":"heading","level":2,"children":[{"text":"Use Cases"}]},{"type":"bulleted-list","children":[{"text":"API keys for external services"},{"text":"Database credentials"},{"text":"User authentication tokens"},{"text":"Encryption keys for data at rest"}]}]',
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'c1b2c3d4-e5f6-7890-abcd-ef1234567893',
  '["security", "secrets", "encryption", "zero-knowledge", "vault"]'::jsonb,
  '{"url":"https://functionfly.com/og-default.svg","alt":"Secrets Vault"}'::jsonb,
  'published',
  NOW(),
  NOW(),
  NOW(),
  'Zero-Knowledge Secrets Vault | FunctionFly',
  'The Secrets Vault uses zero-knowledge encryption to protect sensitive data with client-side encryption.',
  '["zero knowledge secrets", "client encryption", "secret management", "AI security", "encryption vault"]',
  'https://functionfly.com/blog/zero-knowledge-secrets-vault'
) ON CONFLICT (slug) DO NOTHING;

-- Flywheel post
INSERT INTO blog_posts (id, title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, updated_at, created_at, seo_title, seo_description, keywords, canonical_url)
VALUES (
  'd1e2f3a4-b5c6-7890-abcd-ef1234567896',
  'Flywheel Network: Building AI Knowledge Graphs',
  'flywheel-network',
  'Flywheel Network enables AI agents to share and verify knowledge. Learn how knowledge graphs improve AI reliability and reduce hallucinations.',
  '[{"type":"paragraph","children":[{"text":"AI hallucinations are a major challenge for production AI applications. Flywheel Network addresses this by creating a verifiable knowledge graph that AI agents can reference and update."}]},{"type":"heading","level":2,"children":[{"text":"Knowledge Graph Architecture"}]},{"type":"paragraph","children":[{"text":"Flywheel stores facts as edges in a distributed knowledge graph. Each fact is cryptographically signed by the agent that added it, creating an immutable audit trail."}]},{"type":"heading","level":2,"children":[{"text":"Verification"}]},{"type":"paragraph","children":[{"text":"Before using any knowledge, agents can verify facts against the Flywheel Network. This creates a trust chain from source to current claim."}]}]',
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'c1b2c3d4-e5f6-7890-abcd-ef1234567891',
  '["flywheel", "knowledge-graph", "ai", "rag", "hallucination"]'::jsonb,
  '{"url":"https://functionfly.com/og-default.svg","alt":"Flywheel Network"}'::jsonb,
  'published',
  NOW(),
  NOW(),
  NOW(),
  'Flywheel Network: AI Knowledge Graphs | FunctionFly',
  'Flywheel Network enables AI agents to share and verify knowledge using distributed knowledge graphs.',
  '["flywheel network", "knowledge graph", "AI RAG", "hallucination reduction", "AI reliability"]',
  'https://functionfly.com/blog/flywheel-network'
) ON CONFLICT (slug) DO NOTHING;

-- AI Agent Platform post
INSERT INTO blog_posts (id, title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, updated_at, created_at, seo_title, seo_description, keywords, canonical_url)
VALUES (
  'd1e2f3a4-b5c6-7890-abcd-ef1234567897',
  'Building Your First AI Agent on FunctionFly',
  'building-first-ai-agent',
  'A step-by-step tutorial for building and deploying your first AI agent on FunctionFly. Includes example code and best practices.',
  '[{"type":"paragraph","children":[{"text":"This tutorial walks you through building an AI agent that can answer questions about your codebase using RAG (Retrieval Augmented Generation)."}]},{"type":"heading","level":2,"children":[{"text":"Prerequisites"}]},{"type":"bulleted-list","children":[{"text":"A FunctionFly account"},{"text":"OpenAI API key or similar LLM provider"},{"text":"A codebase to query"}]},{"type":"heading","level":2,"children":[{"text":"Step 1: Create the Agent Function"}]},{"type":"paragraph","children":[{"text":"Start by creating a new agent function in the dashboard. Our SDK makes it easy to define tool calls and context management."}]},{"type":"heading","level":2,"children":[{"text":"Step 2: Add Tools"}]},{"type":"paragraph","children":[{"text":"Tools extend what your agent can do. Add a code search tool, a vector search tool for RAG, and any other APIs your agent needs."}]},{"type":"heading","level":2,"children":[{"text":"Step 3: Deploy and Test"}]},{"type":"paragraph","children":[{"text":"Deploy your agent with a single command. Use the playground to test interactions before going live."}]}]',
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'c1b2c3d4-e5f6-7890-abcd-ef1234567896',
  '["tutorial", "ai-agents", "getting-started", "rag", "code-search"]'::jsonb,
  '{"url":"https://functionfly.com/og-default.svg","alt":"AI Agent Tutorial"}'::jsonb,
  'published',
  NOW(),
  NOW(),
  NOW(),
  'Building Your First AI Agent | FunctionFly Tutorial',
  'A step-by-step tutorial for building and deploying your first AI agent on FunctionFly with RAG and tool calls.',
  '["AI agent tutorial", "building AI agents", "RAG tutorial", "FunctionFly getting started", "AI agents tutorial"]',
  'https://functionfly.com/blog/building-first-ai-agent'
) ON CONFLICT (slug) DO NOTHING;

\echo 'Additional blog posts seeded successfully!'
