import { ComponentType } from 'react';
import {
  AwsIcon,
  GoogleCloudIcon,
  VercelIcon,
  NetlifyIcon,
  FlyIoIcon,
  RailwayIcon,
  RenderIcon,
  NextJsIcon,
  ExpressIcon,
  FastifyIcon,
  DjangoIcon,
  PostgreSQLIcon,
  MongoDBIcon,
  RedisIcon,
  StripeIcon,
  GitHubIcon,
  SlackIcon,
  DatadogIcon,
  SentryIcon,
  NewRelicIcon,
  CloudflareIcon
} from '@/pages/LandingPage/components/icons';

export interface Integration {
  id: string;
  name: string;
  category: string;
  description: string;
  icon: ComponentType<{ className?: string }>;
  features: string[];
  setupTime: string;
  documentation: {
    guide: string;
    api: string;
    examples: string;
  };
  status: 'available' | 'coming-soon' | 'beta';
  popularity: 'high' | 'medium' | 'low';
}

export interface IntegrationCategory {
  name: string;
  description: string;
  color: string;
  icon: string;
}

export const integrationCategories: IntegrationCategory[] = [
  {
    name: "Cloud Providers",
    description: "Deploy to multiple cloud platforms simultaneously with automatic failover",
    color: "#3b82f6",
    icon: "cloud"
  },
  {
    name: "Frameworks",
    description: "Native support for popular web frameworks and runtimes",
    color: "#10b981",
    icon: "code"
  },
  {
    name: "Deployment Platforms",
    description: "One-click deployment to your preferred hosting platforms",
    color: "#8b5cf6",
    icon: "rocket"
  },
  {
    name: "Databases",
    description: "Connect to popular databases and data stores",
    color: "#f59e0b",
    icon: "database"
  },
  {
    name: "APIs & Services",
    description: "Integrate with third-party APIs and microservices",
    color: "#ef4444",
    icon: "api"
  },
  {
    name: "Monitoring & Analytics",
    description: "Track performance, errors, and user behavior",
    color: "#06b6d4",
    icon: "bar-chart"
  }
];

export const integrations: Integration[] = [
  // Cloud Providers
  {
    id: "aws-lambda",
    name: "AWS Lambda",
    category: "Cloud Providers",
    description: "Deploy serverless functions to AWS Lambda with automatic scaling, monitoring, and cost optimization.",
    icon: AwsIcon,
    features: [
      "Automatic scaling",
      "99.9% uptime SLA",
      "Pay per execution",
      "Built-in monitoring",
      "VPC integration",
      "Custom runtimes"
    ],
    setupTime: "5 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/aws-lambda",
      api: "https://docs.functionfly.com/api/integrations/aws-lambda",
      examples: "https://docs.functionfly.com/examples/aws-lambda"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "google-cloud-functions",
    name: "Google Cloud Functions",
    category: "Cloud Providers",
    description: "Run your functions on Google's global infrastructure with Firebase integration.",
    icon: GoogleCloudIcon,
    features: [
      "Global edge network",
      "Firebase integration",
      "99.5% uptime SLA",
      "No cold starts",
      "Pub/Sub triggers",
      "HTTP/HTTPS endpoints"
    ],
    setupTime: "3 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/google-cloud-functions",
      api: "https://docs.functionfly.com/api/integrations/google-cloud-functions",
      examples: "https://docs.functionfly.com/examples/google-cloud-functions"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "vercel-functions",
    name: "Vercel Functions",
    category: "Cloud Providers",
    description: "Deploy edge functions globally with Vercel's serverless platform.",
    icon: VercelIcon,
    features: [
      "200+ edge locations",
      "Zero cold starts",
      "Automatic scaling",
      "Preview deployments",
      "Environment variables",
      "Custom domains"
    ],
    setupTime: "2 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/vercel-functions",
      api: "https://docs.functionfly.com/api/integrations/vercel-functions",
      examples: "https://docs.functionfly.com/examples/vercel-functions"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "netlify-functions",
    name: "Netlify Functions",
    category: "Cloud Providers",
    description: "Build and deploy serverless functions with Netlify's developer-friendly platform.",
    icon: NetlifyIcon,
    features: [
      "175+ edge locations",
      "Build hooks",
      "Form handling",
      "Identity service",
      "Large media support",
      "Function logs"
    ],
    setupTime: "3 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/netlify-functions",
      api: "https://docs.functionfly.com/api/integrations/netlify-functions",
      examples: "https://docs.functionfly.com/examples/netlify-functions"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "cloudflare-workers",
    name: "Cloudflare Workers",
    category: "Cloud Providers",
    description: "Run functions at the edge with Cloudflare's global network.",
    icon: CloudflareIcon,
    features: [
      "300+ data centers",
      "Durable Objects",
      "KV storage",
      "WebSockets",
      "Cron triggers",
      "Smart routing"
    ],
    setupTime: "4 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/cloudflare-workers",
      api: "https://docs.functionfly.com/api/integrations/cloudflare-workers",
      examples: "https://docs.functionfly.com/examples/cloudflare-workers"
    },
    status: "available",
    popularity: "medium"
  },

  // Frameworks
  {
    id: "nextjs",
    name: "Next.js",
    category: "Frameworks",
    description: "Full-stack React framework with API routes and server-side rendering support.",
    icon: NextJsIcon,
    features: [
      "API routes",
      "Server-side rendering",
      "Static generation",
      "Image optimization",
      "Internationalization",
      "Middleware"
    ],
    setupTime: "1 minute",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/nextjs",
      api: "https://docs.functionfly.com/api/integrations/nextjs",
      examples: "https://docs.functionfly.com/examples/nextjs"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "express",
    name: "Express.js",
    category: "Frameworks",
    description: "Fast, unopinionated web framework for Node.js with middleware support.",
    icon: ExpressIcon,
    features: [
      "Middleware support",
      "Routing system",
      "Template engines",
      "Error handling",
      "Security features",
      "Community ecosystem"
    ],
    setupTime: "1 minute",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/express",
      api: "https://docs.functionfly.com/api/integrations/express",
      examples: "https://docs.functionfly.com/examples/express"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "fastify",
    name: "Fastify",
    category: "Frameworks",
    description: "Fast and low overhead web framework for Node.js with built-in validation.",
    icon: FastifyIcon,
    features: [
      "High performance",
      "Built-in validation",
      "Plugin system",
      "TypeScript support",
      "Logging",
      "Decorators"
    ],
    setupTime: "2 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/fastify",
      api: "https://docs.functionfly.com/api/integrations/fastify",
      examples: "https://docs.functionfly.com/examples/fastify"
    },
    status: "available",
    popularity: "medium"
  },
  {
    id: "django",
    name: "Django",
    category: "Frameworks",
    description: "High-level Python web framework with batteries included.",
    icon: DjangoIcon,
    features: [
      "ORM included",
      "Admin interface",
      "Security features",
      "URL routing",
      "Template system",
      "Authentication"
    ],
    setupTime: "3 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/django",
      api: "https://docs.functionfly.com/api/integrations/django",
      examples: "https://docs.functionfly.com/examples/django"
    },
    status: "available",
    popularity: "medium"
  },

  // Deployment Platforms
  {
    id: "railway",
    name: "Railway",
    category: "Deployment Platforms",
    description: "Infrastructure platform that makes deployment effortless with databases and more.",
    icon: RailwayIcon,
    features: [
      "Database hosting",
      "Environment variables",
      "Custom domains",
      "Logs & monitoring",
      "Team collaboration",
      "Auto-scaling"
    ],
    setupTime: "2 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/railway",
      api: "https://docs.functionfly.com/api/integrations/railway",
      examples: "https://docs.functionfly.com/examples/railway"
    },
    status: "available",
    popularity: "medium"
  },
  {
    id: "render",
    name: "Render",
    category: "Deployment Platforms",
    description: "Unified cloud platform for static sites, web services, and databases.",
    icon: RenderIcon,
    features: [
      "Static sites",
      "Web services",
      "Background workers",
      "Managed databases",
      "Private networks",
      "Auto-scaling"
    ],
    setupTime: "3 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/render",
      api: "https://docs.functionfly.com/api/integrations/render",
      examples: "https://docs.functionfly.com/examples/render"
    },
    status: "available",
    popularity: "medium"
  },
  {
    id: "fly-io",
    name: "Fly.io",
    category: "Deployment Platforms",
    description: "Deploy apps and databases close to your users with global distribution.",
    icon: FlyIoIcon,
    features: [
      "Global distribution",
      "GPU support",
      "Persistent volumes",
      "Private networking",
      "Custom domains",
      "Logs & metrics"
    ],
    setupTime: "4 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/fly-io",
      api: "https://docs.functionfly.com/api/integrations/fly-io",
      examples: "https://docs.functionfly.com/examples/fly-io"
    },
    status: "available",
    popularity: "medium"
  },

  // Databases
  {
    id: "postgresql",
    name: "PostgreSQL",
    category: "Databases",
    description: "Advanced open source relational database with JSON support and extensions.",
    icon: PostgreSQLIcon,
    features: [
      "ACID compliance",
      "JSON support",
      "Extensions",
      "Full-text search",
      "Geospatial data",
      "Advanced indexing"
    ],
    setupTime: "5 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/postgresql",
      api: "https://docs.functionfly.com/api/integrations/postgresql",
      examples: "https://docs.functionfly.com/examples/postgresql"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "mongodb",
    name: "MongoDB",
    category: "Databases",
    description: "Document database with flexible schema and powerful query capabilities.",
    icon: MongoDBIcon,
    features: [
      "Document model",
      "Flexible schema",
      "Aggregation pipeline",
      "Geospatial queries",
      "Text search",
      "Change streams"
    ],
    setupTime: "4 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/mongodb",
      api: "https://docs.functionfly.com/api/integrations/mongodb",
      examples: "https://docs.functionfly.com/examples/mongodb"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "redis",
    name: "Redis",
    category: "Databases",
    description: "In-memory data structure store used as database, cache, and message broker.",
    icon: RedisIcon,
    features: [
      "In-memory storage",
      "Data structures",
      "Pub/Sub messaging",
      "Caching",
      "Sessions",
      "Rate limiting"
    ],
    setupTime: "3 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/redis",
      api: "https://docs.functionfly.com/api/integrations/redis",
      examples: "https://docs.functionfly.com/examples/redis"
    },
    status: "available",
    popularity: "high"
  },

  // APIs & Services
  {
    id: "stripe",
    name: "Stripe",
    category: "APIs & Services",
    description: "Payment processing and financial services for internet businesses.",
    icon: StripeIcon,
    features: [
      "Payment processing",
      "Subscriptions",
      "Connect platform",
      "Radar fraud detection",
      "Billing management",
      "Webhooks"
    ],
    setupTime: "5 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/stripe",
      api: "https://docs.functionfly.com/api/integrations/stripe",
      examples: "https://docs.functionfly.com/examples/stripe"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "github",
    name: "GitHub",
    category: "APIs & Services",
    description: "Version control and collaboration platform with powerful APIs.",
    icon: GitHubIcon,
    features: [
      "Webhooks",
      "GitHub Actions",
      "Issues & PRs",
      "Repository management",
      "Code scanning",
      "Deployments"
    ],
    setupTime: "3 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/github",
      api: "https://docs.functionfly.com/api/integrations/github",
      examples: "https://docs.functionfly.com/examples/github"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "slack",
    name: "Slack",
    category: "APIs & Services",
    description: "Team communication platform with extensive API capabilities.",
    icon: SlackIcon,
    features: [
      "Real-time messaging",
      "Bot users",
      "Interactive components",
      "File uploads",
      "Workflow automation",
      "App directories"
    ],
    setupTime: "4 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/slack",
      api: "https://docs.functionfly.com/api/integrations/slack",
      examples: "https://docs.functionfly.com/examples/slack"
    },
    status: "available",
    popularity: "high"
  },

  // Monitoring & Analytics
  {
    id: "datadog",
    name: "Datadog",
    category: "Monitoring & Analytics",
    description: "Monitoring, security, and analytics platform for cloud applications.",
    icon: DatadogIcon,
    features: [
      "Infrastructure monitoring",
      "APM & tracing",
      "Log management",
      "Real user monitoring",
      "Security monitoring",
      "Dashboards"
    ],
    setupTime: "7 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/datadog",
      api: "https://docs.functionfly.com/api/integrations/datadog",
      examples: "https://docs.functionfly.com/examples/datadog"
    },
    status: "available",
    popularity: "medium"
  },
  {
    id: "sentry",
    name: "Sentry",
    category: "Monitoring & Analytics",
    description: "Application monitoring and error tracking for production applications.",
    icon: SentryIcon,
    features: [
      "Error tracking",
      "Performance monitoring",
      "Release tracking",
      "User feedback",
      "Custom integrations",
      "Real-time alerts"
    ],
    setupTime: "4 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/sentry",
      api: "https://docs.functionfly.com/api/integrations/sentry",
      examples: "https://docs.functionfly.com/examples/sentry"
    },
    status: "available",
    popularity: "high"
  },
  {
    id: "new-relic",
    name: "New Relic",
    category: "Monitoring & Analytics",
    description: "Observability platform for application and infrastructure monitoring.",
    icon: NewRelicIcon,
    features: [
      "Application monitoring",
      "Infrastructure monitoring",
      "Synthetic monitoring",
      "Log management",
      "Distributed tracing",
      "Alerting"
    ],
    setupTime: "6 minutes",
    documentation: {
      guide: "https://docs.functionfly.com/integrations/new-relic",
      api: "https://docs.functionfly.com/api/integrations/new-relic",
      examples: "https://docs.functionfly.com/examples/new-relic"
    },
    status: "available",
    popularity: "medium"
  }
];

export const getIntegrationsByCategory = (category: string): Integration[] => {
  return integrations.filter(integration => integration.category === category);
};

export const getPopularIntegrations = (): Integration[] => {
  return integrations.filter(integration => integration.popularity === 'high');
};

export const getAvailableIntegrations = (): Integration[] => {
  return integrations.filter(integration => integration.status === 'available');
};