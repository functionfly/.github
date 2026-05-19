import { type Extension } from "@/api/marketplace";

export interface DefaultPlugin {
  id: string;
  name: string;
  version: string;
  description: string;
  category: string;
  creator_id: string;
  verified: boolean;
  install_count: number;
  rating_average: number;
  rating_count: number;
  trust_score: number;
  security_score: number;
  sandbox_score: number;
  runtime_score: number;
  changelog?: string;
  tags: string[];
  icon_emoji: string;
  gradient_from: string;
  gradient_to: string;
}

export const FUNCTIONFLY_TEAM_PLUGINS: DefaultPlugin[] = [
  {
    id: "ff-github",
    name: "GitHub Integration",
    version: "2.1.0",
    description: "Connect your FunctionFly workflows to GitHub repositories. Sync commits, create issues, manage pull requests, and trigger deployments from GitHub Actions webhooks. The official GitHub integration from the FunctionFly team.",
    category: "integration",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 15420,
    rating_average: 4.8,
    rating_count: 342,
    trust_score: 98,
    security_score: 95,
    sandbox_score: 92,
    runtime_score: 94,
    changelog: `## v2.1.0
- Added support for GitHub Enterprise
- Improved webhook reliability
- Bug fixes for rate limiting

## v2.0.0
- Complete rewrite for v2 API
- Added pull request reviews support
- New action: Sync repository settings

## v1.5.0
- Added GitHub Actions workflow triggers
- Improved error handling`,
    tags: ["github", "vcs", "ci/cd", "automation", "webhooks"],
    icon_emoji: "🐙",
    gradient_from: "from-gray-600/30",
    gradient_to: "to-gray-800/30",
  },
  {
    id: "ff-slack",
    name: "Slack Alerts",
    version: "1.8.0",
    description: "Send notifications and alerts to Slack channels. Get real-time updates on workflow executions, errors, and deployment events. Supports rich formatting, interactive blocks, and thread replies.",
    category: "integration",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 12850,
    rating_average: 4.7,
    rating_count: 289,
    trust_score: 97,
    security_score: 94,
    sandbox_score: 91,
    runtime_score: 93,
    changelog: `## v1.8.0
- Added interactive message buttons
- Support for Slack workflows
- Improved message formatting

## v1.7.0
- Added channel picker UI
- Bug fixes`,
    tags: ["slack", "notifications", "alerts", "messaging"],
    icon_emoji: "💬",
    gradient_from: "from-purple-500/30",
    gradient_to: "to-purple-700/30",
  },
  {
    id: "ff-stripe",
    name: "Stripe Billing",
    version: "1.5.0",
    description: "Integrate Stripe payments into your workflows. Handle subscriptions, one-time payments, invoices, and customer management. Automatically sync payment events with your FunctionFly billing data.",
    category: "integration",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 8340,
    rating_average: 4.9,
    rating_count: 178,
    trust_score: 99,
    security_score: 98,
    sandbox_score: 94,
    runtime_score: 91,
    changelog: `## v1.5.0
- Added subscription management
- Support for Stripe Billing Portal
- Tax handling improvements

## v1.4.0
- Invoice generation
- Payment retry logic`,
    tags: ["stripe", "payments", "billing", "subscriptions", "commerce"],
    icon_emoji: "💳",
    gradient_from: "from-blue-500/30",
    gradient_to: "to-indigo-500/30",
  },
  {
    id: "ff-vercel",
    name: "Vercel Deployments",
    version: "1.3.0",
    description: "Deploy your applications to Vercel directly from FunctionFly workflows. Trigger builds, monitor deployments, manage environments, and roll back failed deployments—all from a unified dashboard.",
    category: "infrastructure",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 6780,
    rating_average: 4.6,
    rating_count: 156,
    trust_score: 96,
    security_score: 93,
    sandbox_score: 90,
    runtime_score: 95,
    changelog: `## v1.3.0
- Added preview deployments support
- Environment management
- Deployment rollback actions

## v1.2.0
- Initial release`,
    tags: ["vercel", "deployment", "infrastructure", "hosting", "serverless"],
    icon_emoji: "▲",
    gradient_from: "from-gray-600/30",
    gradient_to: "to-black/30",
  },
  {
    id: "ff-datadog",
    name: "Datadog Monitoring",
    version: "1.4.0",
    description: "Send metrics, traces, and logs to Datadog. Monitor your FunctionFly workflows in real-time with custom dashboards, set up alerts for workflow failures, and correlate events across your infrastructure.",
    category: "ai_tool",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 4560,
    rating_average: 4.5,
    rating_count: 98,
    trust_score: 95,
    security_score: 92,
    sandbox_score: 89,
    runtime_score: 88,
    changelog: `## v1.4.0
- Added log streaming
- Custom metric support
- Dashboard sync

## v1.3.0
- Trace correlation
- Alert integration`,
    tags: ["datadog", "monitoring", "metrics", "observability", "logs"],
    icon_emoji: "📊",
    gradient_from: "from-purple-500/30",
    gradient_to: "to-yellow-500/30",
  },
  {
    id: "ff-sendgrid",
    name: "SendGrid Email",
    version: "1.2.0",
    description: "Send transactional emails through SendGrid. Supports templates, dynamic content, attachments, and scheduled delivery. Track opens and clicks with detailed analytics.",
    category: "integration",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 5890,
    rating_average: 4.4,
    rating_count: 134,
    trust_score: 94,
    security_score: 96,
    sandbox_score: 91,
    runtime_score: 90,
    changelog: `## v1.2.0
- Added template management
- Scheduled sending
- Attachment support

## v1.1.0
- Initial release`,
    tags: ["sendgrid", "email", "notifications", "transactional"],
    icon_emoji: "✉️",
    gradient_from: "from-blue-400/30",
    gradient_to: "to-blue-600/30",
  },
  {
    id: "ff-aws-s3",
    name: "AWS S3 Storage",
    version: "1.1.0",
    description: "Interact with AWS S3 buckets from your workflows. Upload and download files, manage object lifecycle, generate signed URLs, and trigger workflows on S3 events.",
    category: "infrastructure",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 7230,
    rating_average: 4.7,
    rating_count: 201,
    trust_score: 96,
    security_score: 97,
    sandbox_score: 93,
    runtime_score: 91,
    changelog: `## v1.1.0
- Added S3 event triggers
- Signed URL generation
- Lifecycle policy management

## v1.0.0
- Initial release`,
    tags: ["aws", "s3", "storage", "infrastructure", "cloud"],
    icon_emoji: "🪣",
    gradient_from: "from-orange-500/30",
    gradient_to: "to-yellow-500/30",
  },
  {
    id: "ff-postgres",
    name: "PostgreSQL Connector",
    version: "1.6.0",
    description: "Connect to PostgreSQL databases directly from your workflows. Execute queries, manage transactions, and sync data between services. Supports connection pooling and query caching.",
    category: "infrastructure",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 11200,
    rating_average: 4.8,
    rating_count: 445,
    trust_score: 98,
    security_score: 96,
    sandbox_score: 94,
    runtime_score: 92,
    changelog: `## v1.6.0
- Connection pooling improvements
- Added query builder
- Transaction support

## v1.5.0
- Query caching
- Batch operations`,
    tags: ["postgresql", "database", "sql", "infrastructure", "data"],
    icon_emoji: "🐘",
    gradient_from: "from-blue-500/30",
    gradient_to: "to-blue-700/30",
  },
  {
    id: "ff-ai-context",
    name: "AI Context Manager",
    version: "2.0.0",
    description: "Enhanced AI context management for LLM-powered workflows. Provides smarter context window optimization, conversation history management, and semantic caching for reduced token costs.",
    category: "ai_tool",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 9870,
    rating_average: 4.9,
    rating_count: 312,
    trust_score: 99,
    security_score: 97,
    sandbox_score: 95,
    runtime_score: 96,
    changelog: `## v2.0.0
- Semantic caching for 40% token savings
- Dynamic context window sizing
- Multi-modal context support

## v1.5.0
- Conversation threading
- Context compression`,
    tags: ["ai", "llm", "context", "openai", "anthropic", "gpt"],
    icon_emoji: "🧠",
    gradient_from: "from-violet-500/30",
    gradient_to: "to-purple-500/30",
  },
  {
    id: "ff-webhook",
    name: "Webhook Builder",
    version: "1.4.0",
    description: "Build and manage custom webhooks with visual tools. Define request/response transformations, set up authentication, and monitor webhook activity with built-in logging.",
    category: "ui",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 14560,
    rating_average: 4.6,
    rating_count: 523,
    trust_score: 97,
    security_score: 98,
    sandbox_score: 94,
    runtime_score: 93,
    changelog: `## v1.4.0
- Visual transformation builder
- HMAC signature verification
- Response templating

## v1.3.0
- Activity logging
- Retry logic`,
    tags: ["webhook", "api", "integration", "http", "automation"],
    icon_emoji: "🪝",
    gradient_from: "from-green-500/30",
    gradient_to: "to-teal-500/30",
  },
  {
    id: "ff-scheduler",
    name: "Workflow Scheduler",
    version: "1.3.0",
    description: "Schedule workflows to run at specific times or intervals. Supports cron expressions, calendar-based scheduling, and complex recurrence patterns. Includes timezone support and daylight saving handling.",
    category: "runtime",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 16800,
    rating_average: 4.8,
    rating_count: 612,
    trust_score: 98,
    security_score: 95,
    sandbox_score: 94,
    runtime_score: 97,
    changelog: `## v1.3.0
- Calendar-based scheduling
- Timezone improvements
- Conflict detection

## v1.2.0
- Cron expression builder UI
- DAG scheduling`,
    tags: ["scheduler", "cron", "automation", "scheduling", "time"],
    icon_emoji: "⏰",
    gradient_from: "from-cyan-500/30",
    gradient_to: "to-blue-500/30",
  },
  {
    id: "ff-rate-limiter",
    name: "Rate Limiter",
    version: "1.1.0",
    description: "Add rate limiting to your workflows with configurable limits per endpoint, IP, or user. Supports sliding window, token bucket, and fixed window algorithms.",
    category: "infrastructure",
    creator_id: "FunctionFly Team",
    verified: true,
    install_count: 5340,
    rating_average: 4.5,
    rating_count: 145,
    trust_score: 94,
    security_score: 96,
    sandbox_score: 91,
    runtime_score: 93,
    changelog: `## v1.1.0
- Initial release with token bucket algorithm
- Redis-backed distributed rate limiting`,
    tags: ["rate-limiting", "security", "infrastructure", "throttling"],
    icon_emoji: "🚦",
    gradient_from: "from-red-500/30",
    gradient_to: "to-orange-500/30",
  },
];

export function toExtension(plugin: DefaultPlugin): Extension {
  return {
    id: plugin.id,
    creator_id: plugin.creator_id,
    name: plugin.name,
    version: plugin.version,
    description: plugin.description,
    category: plugin.category,
    verified: plugin.verified,
    install_count: plugin.install_count,
    rating_average: plugin.rating_average,
    rating_count: plugin.rating_count,
    trust_score: plugin.trust_score,
    security_score: plugin.security_score,
    sandbox_score: plugin.sandbox_score,
    runtime_score: plugin.runtime_score,
    changelog: plugin.changelog,
    tags: plugin.tags,
    featured: true,
    status: "published",
    manifest: { name: plugin.name, version: plugin.version },
  };
}

export const ALL_FUNCTIONFLY_PLUGINS = FUNCTIONFLY_TEAM_PLUGINS.map(toExtension);