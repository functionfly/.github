/**
 * Default blog post: Welcome to FunctionFly
 * Platform introduction and overview
 */
import { ContentStatus } from '../../modules/blog/dto/blog-post.dto';

export const slug = 'welcome-to-functionfly';

const body = [
  {
    type: 'paragraph',
    children: [{ text: 'Welcome to FunctionFly, the serverless platform designed for the AI era. We\'re building the infrastructure that enables developers to create, deploy, and monetize AI-powered applications with unprecedented ease and security.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'What Makes FunctionFly Different?' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'FunctionFly isn\'t just another serverless platform—it\'s purpose-built for the AI-native world. Our core innovations address the fundamental challenges of building AI applications at scale:' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '🔄 **Flywheel Network™**: A proof-of-execution knowledge network where every function execution becomes verifiable, composable knowledge. Problems are structured, solutions are executable, and AI agents collaborate in open debates.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '🔐 **Zero-Knowledge Secrets Vault**: Client-side encrypted secrets that scale from free to enterprise-grade without compromising security. Your data never touches our servers in plaintext.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '🧠 **AI-First Architecture**: Built for AI agents, RAG systems, and autonomous workflows with features like State Fabric for durable memory and CCP (Compute Capsules Protocol) for verifiable execution.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Our Mission' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'To democratize AI development by providing the infrastructure that makes building AI applications as simple as writing a function. We believe the future belongs to platforms that treat AI as a first-class citizen, not an afterthought.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Whether you\'re building AI microservices, RAG applications, autonomous agents, or the next generation of AI-powered SaaS—we\'ve designed FunctionFly to be your launchpad.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'What\'s Next' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'This blog will be your guide to building on FunctionFly. We\'ll dive deep into our core technologies, share tutorials for common patterns, showcase builder success stories, and keep you updated on our roadmap.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Ready to start building? Head to our dashboard and deploy your first function. The future of AI development starts here.' }],
  },
];

export const welcomePost = {
  title: 'Welcome to FunctionFly: Serverless Infrastructure for the AI Era',
  slug,
  description: 'Introducing FunctionFly, the serverless platform purpose-built for AI applications with Flywheel Network, zero-knowledge secrets, and AI-first architecture.',
  body,
  tags: ['functionfly', 'serverless', 'ai', 'platform', 'introduction', 'flywheel-network', 'secrets-vault'],
  status: ContentStatus.PUBLISHED,
  publishedAt: new Date().toISOString(),
  seoTitle: 'Welcome to FunctionFly | AI-Native Serverless Platform',
  seoDescription: 'FunctionFly is the serverless platform for AI era. Flywheel Network, zero-knowledge secrets, and infrastructure that treats AI as a first-class citizen.',
  keywords: ['functionfly', 'serverless', 'AI platform', 'flywheel network', 'secrets vault', 'AI infrastructure'],
  canonicalUrl: 'https://functionfly.com/blog/welcome-to-functionfly',
} as const;

export type WelcomePostPayload = typeof welcomePost;