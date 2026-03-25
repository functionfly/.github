import { DOCS_SITE_URL } from '@/lib/constants';
import {
  ArrowRightLeft,
  BarChart3,
  Brain,
  Code2,
  Globe,
  Network,
  Settings,
  Users,
} from 'lucide-react';
import { GlobalNetworkIllustration, PredictiveRoutingIllustration } from '../illustrations';
import { Feature, categoryColors } from '../types';

export const features: Feature[] = [
  {
    icon: ArrowRightLeft,
    title: 'Fast Failover',
    description:
      'Sub-second failover between providers ensures your users never experience downtime. Our intelligent routing automatically redirects traffic to healthy endpoints. Available in all plans including Free tier.',
    category: 'Reliability',
    highlights: [
      '<100ms failover time',
      'Automatic health monitoring',
      'Zero-downtime deployments',
    ],
    cta: {
      text: 'Test Failover Demo',
      action: 'demo',
      link: '#interactive-demos',
    },
    detailedContent: {
      overview:
        'Our fast failover system uses multiple layers of health checking to ensure your applications stay online.',
      technicalDetails: [
        'Multi-layer health checks (HTTP, TCP, DNS)',
        'Circuit breaker pattern implementation',
        'Automated rollback capabilities',
        'Real-time provider status monitoring',
      ],
      benefits: [
        'Zero downtime during provider outages',
        'Automatic recovery without manual intervention',
        'Seamless user experience during failures',
        'Detailed incident reporting and analytics',
      ],
      implementation:
        'Failover decisions are made in <50ms using distributed consensus algorithms across our global control plane.',
    },
    ...categoryColors['Reliability'],
  },
  {
    icon: Network,
    title: 'Multi-Provider Deployment',
    description:
      'Deploy your functions to multiple edge providers simultaneously. No vendor lock-in, maximum flexibility, and global distribution. Pro plan required for production deployments.',
    category: 'Deployment',
    highlights: [
      'Vercel, Netlify, Fly.io, Cloudflare',
      'Load balancing across providers',
      'Geographic distribution',
    ],
    cta: {
      text: 'Try Multi-Provider Deployment',
      action: 'demo',
      link: '#interactive-demos',
    },
    detailedContent: {
      overview: 'Deploy to multiple providers simultaneously with our unified deployment pipeline.',
      technicalDetails: [
        'Unified deployment API across all providers',
        'Provider-specific optimization layers',
        'Atomic deployment transactions',
        'Rollback capabilities across providers',
      ],
      benefits: [
        'No vendor lock-in - switch providers anytime',
        'Redundant deployment for high availability',
        'Cost optimization through provider selection',
        'Geographic coverage expansion',
      ],
      implementation:
        'Our orchestration layer manages deployments using Kubernetes-style controllers with provider-specific adapters.',
    },
    ...categoryColors['Deployment'],
  },
  {
    icon: Brain,
    title: 'Predictive Routing',
    description:
      'AI-powered traffic routing analyzes real-time health metrics to predict and prevent issues before they impact users. Pro plan feature with advanced ML algorithms.',
    category: 'Intelligence',
    highlights: [
      'Machine learning algorithms',
      'Real-time health monitoring',
      'Proactive issue prevention',
    ],
    illustration: PredictiveRoutingIllustration,
    cta: {
      text: 'Experience AI Routing',
      action: 'demo',
      link: '#interactive-demos',
    },
    detailedContent: {
      overview:
        'Machine learning algorithms predict and prevent issues before they impact your users.',
      technicalDetails: [
        'LSTM neural networks for time-series prediction',
        'Real-time anomaly detection using statistical models',
        'Multi-armed bandit algorithms for optimal routing',
        'Historical data analysis with 99.9% uptime tracking',
      ],
      benefits: [
        'Proactive issue prevention reduces incidents by 80%',
        'Intelligent load balancing improves performance',
        'Predictive scaling prevents resource exhaustion',
        'Automated incident response and recovery',
      ],
      implementation:
        'Our AI models are trained on millions of deployment events and continuously updated using reinforcement learning.',
    },
    ...categoryColors['Intelligence'],
  },
  {
    icon: Code2,
    title: 'Developer Experience',
    description:
      'Seamless deployment with Git integration, CLI tools, and comprehensive APIs. Deploy in minutes, not hours. Core features available in Free tier.',
    category: 'Developer Tools',
    highlights: ['Git integration', 'CLI and API support', 'TypeScript/JavaScript support'],
    cta: {
      text: 'Get Started with CLI',
      action: 'link',
      link: `${DOCS_SITE_URL}/docs/cli`,
    },
    ...categoryColors['Developer Tools'],
  },
  {
    icon: Users,
    title: 'Team Collaboration',
    description:
      'Invite team members, manage permissions, and collaborate on deployments with role-based access control. Up to 3 team members in Free tier.',
    category: 'Collaboration',
    highlights: ['Role-based permissions', 'Team member invites', 'Audit logs and tracking'],
    cta: {
      text: 'Invite Team Members',
      action: 'link',
      link: '/dashboard/team',
    },
    ...categoryColors['Collaboration'],
  },
  {
    icon: BarChart3,
    title: 'Advanced Analytics',
    description:
      'Comprehensive analytics with custom dashboards, performance metrics, and detailed insights into your deployments. Basic metrics in Free tier, advanced analytics in Pro plan.',
    category: 'Monitoring',
    highlights: ['Real-time metrics', 'Custom dashboards', 'Performance insights'],
    cta: {
      text: 'View Analytics Dashboard',
      action: 'link',
      link: '/dashboard/analytics',
    },
    detailedContent: {
      overview:
        'Get deep insights into your deployments with comprehensive analytics and custom dashboards.',
      technicalDetails: [
        'Real-time metrics collection with <1s latency',
        'Custom dashboard builder with drag-and-drop interface',
        'Advanced querying with SQL-like syntax',
        'Integration with external monitoring tools',
      ],
      benefits: [
        'Data-driven optimization decisions',
        'Performance bottleneck identification',
        'Cost analysis and optimization insights',
        'Compliance reporting and audit trails',
      ],
      implementation:
        'Built on ClickHouse for high-performance analytics with real-time data ingestion and processing.',
    },
    ...categoryColors['Monitoring'],
  },
  {
    icon: Settings,
    title: 'Flexible Configuration',
    description:
      'Configure environments, custom domains, environment variables, and deployment settings with granular control. Custom domains require Pro plan.',
    category: 'Configuration',
    highlights: ['Environment management', 'Custom domains', 'Advanced settings'],
    cta: {
      text: 'Configure Environment',
      action: 'link',
      link: '/settings',
    },
    detailedContent: {
      overview:
        'Fine-tune your deployments with granular configuration options and environment management.',
      technicalDetails: [
        'Environment-specific configuration inheritance',
        'Secret management with encryption at rest',
        'Custom domain SSL certificate automation',
        'Advanced routing rules and middleware configuration',
      ],
      benefits: [
        'Secure environment variable management',
        'Custom domain support with automatic SSL',
        'Granular access control per environment',
        'Configuration as code with Git integration',
      ],
      implementation:
        'Configuration is stored encrypted and distributed globally using our edge network for low-latency access.',
    },
    ...categoryColors['Configuration'],
  },
  {
    icon: Globe,
    title: 'Global Edge Network',
    description:
      'Deploy to 200+ edge locations worldwide for optimal performance and reduced latency for your users. Full global coverage in Pro plan.',
    category: 'Infrastructure',
    highlights: ['200+ edge locations', 'Global CDN', 'Optimal performance'],
    illustration: GlobalNetworkIllustration,
    cta: {
      text: 'Explore Edge Locations',
      action: 'demo',
      link: '#interactive-demos',
    },
    detailedContent: {
      overview:
        'Leverage our global edge network with 200+ locations for optimal performance worldwide.',
      technicalDetails: [
        'Anycast routing for optimal path selection',
        'Edge-side compute with WebAssembly runtime',
        'Global database replication with RPO <1s',
        'Content delivery network with edge caching',
      ],
      benefits: [
        'Sub-100ms latency worldwide',
        'Automatic traffic optimization',
        'DDoS protection at the edge',
        'Compliance with regional data regulations',
      ],
      implementation:
        'Our edge network uses BGP anycast routing and runs on bare metal in 200+ data centers worldwide.',
    },
    ...categoryColors['Infrastructure'],
  },
];

export const categories = [...new Set(features.map((f) => f.category))];
