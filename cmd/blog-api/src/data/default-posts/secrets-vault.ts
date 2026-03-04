/**
 * Default blog post: Introducing Secrets Vault
 * Zero-knowledge secrets management overview
 */
import { ContentStatus } from '../../modules/blog/dto/blog-post.dto';

export const slug = 'introducing-secrets-vault';

const body = [
  {
    type: 'paragraph',
    children: [{ text: 'Secrets Vault is FunctionFly\'s zero-knowledge secrets management system. It starts completely free with client-side encryption and scales to enterprise-grade security without compromising the core security model. Your secrets never touch our servers in plaintext.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'The Security Problem' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Traditional secrets management puts you in an impossible position. Either you trust a third-party service with your plaintext secrets, or you manage everything yourself with expensive infrastructure. Database breaches, insider threats, and compliance requirements make this a constant source of anxiety.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Secrets Vault solves this with zero-knowledge architecture. Your secrets are encrypted on your device before they ever reach our servers. We can help you manage and share secrets without ever seeing them.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'How Zero-Knowledge Works' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'The magic happens in your browser using the Web Crypto API:' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '1. **Client-Side Encryption**: Your passphrase is converted to a cryptographic key using PBKDF2 with 100,000 iterations' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '2. **AES-256-GCM**: Secrets are encrypted with authenticated encryption before transmission' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '3. **Zero Server Access**: Our servers only store encrypted blobs—we can\'t decrypt them even if we wanted to' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '4. **Token-Based Access**: Time-limited access tokens allow sharing without revealing your passphrase' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Cost-Effective Scaling' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Unlike other secrets managers that start expensive, Secrets Vault scales with your revenue:' }],
  },
  {
    type: 'heading',
    level: 3,
    children: [{ text: 'Phase 1: Free ($0)' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Web Crypto API encryption' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• PostgreSQL storage (you already pay for this)' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Basic token management' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Audit logging for compliance' }],
  },
  {
    type: 'heading',
    level: 3,
    children: [{ text: 'Phase 2: Growth ($50-100/month)' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• AWS KMS integration for key hierarchy' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Redis caching for high-volume token validation' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• S3 audit log exports' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Advanced policy engine' }],
  },
  {
    type: 'heading',
    level: 3,
    children: [{ text: 'Phase 3: Enterprise ($1000+/month)' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Hardware Security Modules (HSM)' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Nitro Enclaves for attested decryption' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Dynamic secrets with auto-rotation' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• Blockchain-anchored audit trails' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Security Benefits' }],
  },
  {
    type: 'heading',
    level: 3,
    children: [{ text: 'No Single Point of Failure' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Your passphrase never leaves your device. Even if our entire infrastructure is compromised, attackers get only encrypted blobs. The encryption key exists only in your browser session.' }],
  },
  {
    type: 'heading',
    level: 3,
    children: [{ text: 'Perfect Forward Secrecy' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Each encryption uses a unique salt and initialization vector. Compromising one secret doesn\'t help attackers decrypt others.' }],
  },
  {
    type: 'heading',
    level: 3,
    children: [{ text: 'Compliance Ready' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Built-in audit logging, SOC2 compliance templates, and enterprise features when you need them. Start with self-audit capabilities that scale to full compliance reporting.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Use Cases' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Secrets Vault works for any application that needs secure secret management:' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• **API Keys**: Database credentials, third-party service keys, payment processor secrets' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• **Environment Variables**: Production configuration, connection strings, service URLs' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• **Team Secrets**: Shared credentials for development and deployment' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '• **Function Configuration**: Runtime secrets for serverless functions' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'Getting Started' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Getting started is simple:' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '1. **Create a Vault**: Choose a strong passphrase (we recommend a password manager)' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '2. **Add Secrets**: Encrypt and store your sensitive data' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '3. **Generate Tokens**: Create time-limited access for your functions' }],
  },
  {
    type: 'paragraph',
    children: [{ text: '4. **Use in Functions**: Access secrets securely in your FunctionFly deployments' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'The entire process happens in your browser. Your secrets never leave your device unencrypted.' }],
  },
  {
    type: 'heading',
    level: 2,
    children: [{ text: 'The Future of Secrets Management' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'Secrets Vault represents a new approach to secrets management. Instead of expensive infrastructure from day one, you get enterprise-grade security that grows with your business. The zero-knowledge architecture ensures your security model never weakens as you scale.' }],
  },
  {
    type: 'paragraph',
    children: [{ text: 'This is more than a secrets manager—it\'s a demonstration that security and cost-effectiveness can go hand-in-hand.' }],
  },
];

export const secretsVaultPost = {
  title: 'Introducing Secrets Vault: Zero-Knowledge Secrets That Scale',
  slug,
  description: 'Client-side encrypted secrets that start free and scale to enterprise-grade. Zero-knowledge architecture means your data never touches our servers in plaintext.',
  body,
  tags: ['secrets-vault', 'security', 'zero-knowledge', 'encryption', 'compliance', 'functionfly'],
  status: ContentStatus.PUBLISHED,
  publishedAt: new Date().toISOString(),
  seoTitle: 'Secrets Vault | Zero-Knowledge Secrets Management',
  seoDescription: 'Client-side encrypted secrets that start free and scale to enterprise. Zero-knowledge architecture with AES-256-GCM encryption and token-based access.',
  keywords: ['secrets vault', 'zero knowledge', 'client side encryption', 'AES-256-GCM', 'secrets management', 'functionfly'],
  canonicalUrl: 'https://functionfly.com/blog/introducing-secrets-vault',
} as const;

export type SecretsVaultPostPayload = typeof secretsVaultPost;