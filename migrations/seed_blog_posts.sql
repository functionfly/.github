-- Seed script for blog tables
-- Run: psql -h localhost -p 5432 -U postgres -d functionfly -f migrations/seed_blog_posts.sql

-- Insert default author
INSERT INTO blog_authors (id, name, slug, bio, role, active, created_at, updated_at)
VALUES (
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'FunctionFly Team',
  'functionfly-team',
  'Product and engineering notes from the FunctionFly team.',
  'Editorial',
  true,
  NOW(),
  NOW()
) ON CONFLICT (slug) DO NOTHING;

-- Insert categories
INSERT INTO blog_categories (id, title, slug, description, "order", created_at, updated_at) VALUES
  ('c1b2c3d4-e5f6-7890-abcd-ef1234567891', 'Announcements', 'announcements', 'Platform news, positioning, and what we are shipping.', 0, NOW(), NOW()),
  ('c1b2c3d4-e5f6-7890-abcd-ef1234567892', 'Trust & Safety', 'trust', 'Verification tiers, attestations, Trust API, and safe agent tooling.', 1, NOW(), NOW()),
  ('c1b2c3d4-e5f6-7890-abcd-ef1234567893', 'Security', 'security', 'Secrets, encryption, and zero-knowledge patterns.', 2, NOW(), NOW()),
  ('c1b2c3d4-e5f6-7890-abcd-ef1234567894', 'Architecture', 'architecture', 'Protocols, durable state, and verifiable execution.', 3, NOW(), NOW()),
  ('c1b2c3d4-e5f6-7890-abcd-ef1234567895', 'AI & Agents', 'ai-agents', 'Agent-native workflows, integration patterns, and the AI-era stack.', 4, NOW(), NOW()),
  ('c1b2c3d4-e5f6-7890-abcd-ef1234567896', 'Tutorials', 'tutorials', 'Step-by-step guides to ship on FunctionFly.', 5, NOW(), NOW()),
  ('c1b2c3d4-e5f6-7890-abcd-ef1234567897', 'Stories', 'stories', 'Builder narratives and patterns that worked.', 6, NOW(), NOW())
ON CONFLICT (slug) DO NOTHING;

-- Insert blog posts
INSERT INTO blog_posts (id, title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, updated_at, created_at, seo_title, seo_description, keywords, canonical_url)
VALUES (
  'd1e2f3a4-b5c6-7890-abcd-ef1234567891',
  'Welcome to FunctionFly: Serverless Infrastructure for the AI Era',
  'welcome-to-functionfly',
  'Introducing FunctionFly, the serverless platform purpose-built for AI applications with Flywheel Network, zero-knowledge secrets, and AI-first architecture.',
  '[{"type":"paragraph","children":[{"text":"Welcome to FunctionFly, the serverless platform designed for the AI era. We are building the infrastructure that enables developers to create, deploy, and monetize AI-powered applications with unprecedented ease and security."}]},{"type":"heading","level":2,"children":[{"text":"What Makes FunctionFly Different?"}]},{"type":"paragraph","children":[{"text":"FunctionFly is purpose-built for the AI-native world. Core innovations include the Flywheel Network for verifiable knowledge, Zero-Knowledge Secrets Vault for client-side encryption, and AI-First Architecture with State Fabric for durable memory."}]},{"type":"heading","level":2,"children":[{"text":"Our Mission"}]},{"type":"paragraph","children":[{"text":"To democratize AI development by providing infrastructure that makes building AI applications as simple as writing a function."}]},{"type":"paragraph","children":[{"text":"Whether you are building AI microservices, RAG applications, autonomous agents, or the next generation of AI-powered SaaS—we have designed FunctionFly to be your launchpad."}]}]',
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'c1b2c3d4-e5f6-7890-abcd-ef1234567891',
  '["functionfly", "serverless", "ai", "platform", "introduction"]'::jsonb,
  '{"url":"https://functionfly.com/og-default.svg","alt":"Welcome to FunctionFly"}'::jsonb,
  'published',
  NOW(),
  NOW(),
  NOW(),
  'Welcome to FunctionFly | AI-Native Serverless Platform',
  'FunctionFly is the serverless platform for AI era. Flywheel Network, zero-knowledge secrets, and infrastructure that treats AI as a first-class citizen.',
  '["functionfly", "serverless", "AI platform", "flywheel network", "secrets vault", "AI infrastructure"]',
  'https://functionfly.com/blog/welcome-to-functionfly'
) ON CONFLICT (slug) DO NOTHING;

INSERT INTO blog_posts (id, title, slug, description, body, author_id, category_id, tags, hero_image, status, published_at, updated_at, created_at, seo_title, seo_description, keywords, canonical_url)
VALUES (
  'd1e2f3a4-b5c6-7890-abcd-ef1234567892',
  'Durable Execution for AI Agents: Building Reliable AI Workflows',
  'durable-execution-for-ai-agents',
  'Learn how durable execution enables AI agents to recover from failures seamlessly. Explore State Fabric for checkpointing, replay, and production-grade reliability.',
  '[{"type":"paragraph","children":[{"text":"AI agents are only as reliable as their ability to recover from failures. When an autonomous agent is halfway through executing a multi-step task and the process crashes, everything is lost in traditional serverless architectures. For AI agents handling critical workflows, this is unacceptable."}]},{"type":"heading","level":2,"children":[{"text":"The Problem with Stateless AI Workflows"}]},{"type":"paragraph","children":[{"text":"Most serverless platforms treat functions as fire-and-forget invocations. You call a function, it runs to completion, and that is it. For AI agents, this creates a fundamental tension: agents are inherently stateful -- maintaining context across multiple steps -- yet the infrastructure they run on is often stateless by design."}]},{"type":"paragraph","children":[{"text":"The consequences include lost progress on failure, inconsistent state when external calls succeed but are not recorded, idempotency issues on retry, and debugging nightmares since production issues cannot be reproduced."}]},{"type":"heading","level":2,"children":[{"text":"What Is Durable Execution?"}]},{"type":"paragraph","children":[{"text":"Durable execution automatically checkpoints workflow state at safe points and resumes from those checkpoints after any failure -- process crashes, machine reboots, or code deployments mid-execution. Think of it as an automatic undo system for distributed workflows."}]},{"type":"heading","level":2,"children":[{"text":"State Fabric: Durable Memory for AI Agents"}]},{"type":"paragraph","children":[{"text":"FunctionFly State Fabric is purpose-built infrastructure for durable AI agent execution. It provides automatic checkpointing without manual instrumentation, instant recovery with full context intact, deterministic replay ensuring idempotency, and distributed durability across availability zones."}]},{"type":"heading","level":2,"children":[{"text":"How State Fabric Works"}]},{"type":"paragraph","children":[{"text":"State Fabric models workflows as directed acyclic graphs (DAGs). Each step inputs and outputs are captured, along with dependencies. On failure, it traverses the DAG to find the last completed step and resumes from there -- returning cached outputs for completed steps and queuing pending ones."}]},{"type":"heading","level":2,"children":[{"text":"Real-World Example: Invoice Processing Agent"}]},{"type":"paragraph","children":[{"text":"Consider an AI agent that fetches invoices, extracts line items via vision model, validates against contracts, routes for approval, posts to accounting, and sends confirmations. Without durable execution, a crash at step 5 means re-running expensive vision model calls. With State Fabric, the agent resumes at step 5 with all previous results intact."}]},{"type":"heading","level":2,"children":[{"text":"The Business Impact"}]},{"type":"paragraph","children":[{"text":"Durable execution delivers reduced compute costs (no wasted GPU cycles), faster recovery (milliseconds vs minutes), predictable SLAs for customers, and automatic audit compliance with complete execution history for regulators."}]},{"type":"heading","level":2,"children":[{"text":"Beyond Checkpointing: Event Sourcing for AI"}]},{"type":"paragraph","children":[{"text":"State Fabric records every step as an immutable event, creating a complete audit trail. For compliance-heavy industries, regulators can replay exactly what happened. Debugging becomes querying the event log rather than reproducing failures."}]},{"type":"heading","level":2,"children":[{"text":"Technical Architecture"}]},{"type":"paragraph","children":[{"text":"State Fabric builds on PostgreSQL with the following key architectural decisions. First, every workflow execution gets a unique identifier used to namespace all state. Second, checkpoints are written transactionally alongside business logic, ensuring atomicity. Third, the DAG structure allows parallel execution of independent steps while maintaining strict ordering where required."}]},{"type":"paragraph","children":[{"text":"The system uses optimistic concurrency control to handle race conditions when multiple agent instances attempt to update the same workflow state. Each checkpoint includes a version number; if a conflict is detected, the system automatically retries from the last known good state."}]},{"type":"heading","level":2,"children":[{"text":"Comparison with Traditional Approaches"}]},{"type":"paragraph","children":[{"text":"Before State Fabric, developers built durable workflows using external databases, message queues, or specialized orchestration engines. These approaches work but introduce significant complexity. External databases require custom state management logic in every function. Message queues add operational overhead and monitoring burden. Orchestration engines like Temporal or Conductor are powerful but require learning a new programming model and deployment infrastructure."}]},{"type":"paragraph","children":[{"text":"State Fabric integrates directly into the FunctionFly runtime, making durable execution as simple as deploying a serverless function. There is no additional infrastructure to manage, no SDK to integrate, and no new programming model to learn."}]},{"type":"heading","level":2,"children":[{"text":"Getting Started"}]},{"type":"paragraph","children":[{"text":"FunctionFly enables durable execution as a first-class primitive -- State Fabric is enabled by default when you deploy an agent. Create a free account at functionfly.com, deploy your first agent function, configure retry policies, and monitor execution history in the dashboard."}]},{"type":"paragraph","children":[{"text":"Start with a simple multi-step workflow -- fetch data, process it, and store results. Deploy it, trigger a failure mid-execution, and watch State Fabric automatically resume from where it left off. Once you see it in action, you will understand why durable execution is essential for production AI agents."}]},{"type":"heading","level":2,"children":[{"text":"Conclusion"}]},{"type":"paragraph","children":[{"text":"Durable execution transforms AI agents from best-effort assistants into reliable automation. State Fabric brings enterprise-grade reliability to AI-native workflows -- whether building customer service agents, data pipelines, or autonomous decision systems, durable execution is the foundation for production-grade AI."}]}]',
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'c1b2c3d4-e5f6-7890-abcd-ef1234567895',
  '["ai-agents", "durable-execution", "state-fabric", "reliability", "serverless", "ai-workflows"]'::jsonb,
  '{"url":"https://functionfly.com/og-default.svg","alt":"Durable Execution for AI Agents"}'::jsonb,
  'published',
  NOW(),
  NOW(),
  NOW(),
  'Durable Execution for AI Agents | State Fabric by FunctionFly',
  'Discover how durable execution enables AI agents to recover from failures seamlessly. Learn about State Fabric, checkpointing, event sourcing, and building reliable AI workflows.',
  '["durable execution for AI agents", "AI agent reliability", "State Fabric", "serverless AI", "agentic AI workflows", "autonomous agent failures"]',
  'https://functionfly.com/blog/durable-execution-for-ai-agents'
) ON CONFLICT (slug) DO NOTHING;

\echo 'Blog posts seeded successfully!'
