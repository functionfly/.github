import { JsonLd } from 'react-schemaorg';

export function LandingPageStructuredData() {
  return (
    <JsonLd
      item={{
        '@context': 'https://schema.org',
        '@graph': [
          {
            '@type': 'SoftwareApplication',
            'name': 'FunctionFly',
            'description': 'Deploy serverless functions to any cloud provider with zero-config deployments and instant scaling',
            'applicationCategory': 'DeveloperApplication',
            'operatingSystem': 'Cross-platform',
            'offers': {
              '@type': 'Offer',
              'price': '0',
              'priceCurrency': 'USD',
              'description': 'Free tier available, paid plans start from $9/month'
            },
            'author': {
              '@type': 'Organization',
              'name': 'FunctionFly'
            },
            'aggregateRating': {
              '@type': 'AggregateRating',
              'ratingValue': '4.8',
              'ratingCount': '1250'
            }
          },
          {
            '@type': 'WebSite',
            'name': 'FunctionFly',
            'url': 'https://functionfly.com',
            'description': 'Deploy serverless functions to any cloud provider',
            'potentialAction': {
              '@type': 'SearchAction',
              'target': 'https://functionfly.com/search?q={search_term_string}',
              'query-input': 'required name=search_term_string'
            }
          },
          {
            '@type': 'Organization',
            'name': 'FunctionFly',
            'url': 'https://functionfly.com',
            'logo': 'https://functionfly.com/logo.png',
            'sameAs': [
              'https://github.com/functionfly',
              'https://twitter.com/functionfly'
            ]
          }
        ]
      } as any}
    />
  );
}

export function PricingPageStructuredData() {
  return (
    <JsonLd
      item={{
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        'name': 'FunctionFly Pricing',
        'description': 'Choose the right plan for your serverless function deployments',
        'offers': [
          {
            '@type': 'Offer',
            'name': 'Free Plan',
            'price': '0',
            'priceCurrency': 'USD',
            'description': 'Perfect for getting started with up to 100 function invocations per month'
          },
          {
            '@type': 'Offer',
            'name': 'Pro Plan',
            'price': '29',
            'priceCurrency': 'USD',
            'description': 'Professional plan with 10,000 invocations and advanced features'
          },
          {
            '@type': 'Offer',
            'name': 'Enterprise Plan',
            'price': '99',
            'priceCurrency': 'USD',
            'description': 'Enterprise-grade deployments with unlimited invocations and priority support'
          }
        ]
      }}
    />
  );
}

export function FAQPageStructuredData() {
  return (
    <JsonLd
      item={{
        '@context': 'https://schema.org',
        '@type': 'FAQPage',
        'mainEntity': [
          {
            '@type': 'Question',
            'name': 'What is FunctionFly?',
            'acceptedAnswer': {
              '@type': 'Answer',
              'text': 'FunctionFly is an edge computing platform that allows you to deploy serverless functions globally. It provides instant deployment, automatic scaling, and edge execution across multiple cloud providers, giving you unprecedented control over your application\'s performance and reach.'
            }
          },
          {
            '@type': 'Question',
            'name': 'How do I get started with FunctionFly?',
            'acceptedAnswer': {
              '@type': 'Answer',
              'text': 'Getting started is simple! Sign up for a free account, connect your first cloud provider (we support AWS, Google Cloud, Cloudflare, Vercel, and Fly.io), and deploy your first function using our CLI or web dashboard. Check out our documentation for step-by-step guides.'
            }
          },
          {
            '@type': 'Question',
            'name': 'What programming languages are supported?',
            'acceptedAnswer': {
              '@type': 'Answer',
              'text': 'FunctionFly supports all major programming languages including JavaScript/TypeScript (Node.js), Python, Go, Rust, Java, PHP, Ruby, and .NET. You can also deploy containerized applications using Docker.'
            }
          },
          {
            '@type': 'Question',
            'name': 'How does FunctionFly pricing work?',
            'acceptedAnswer': {
              '@type': 'Answer',
              'text': 'We offer a generous free tier and flexible paid plans based on usage. You pay for compute time, bandwidth, and storage. Our pricing is transparent with no hidden fees, and you can monitor costs in real-time through our dashboard.'
            }
          },
          {
            '@type': 'Question',
            'name': 'How secure is FunctionFly?',
            'acceptedAnswer': {
              '@type': 'Answer',
              'text': 'Security is our top priority. We implement multiple layers of security including encryption at rest and in transit, regular security audits, compliance with SOC 2, GDPR, and CCPA, and advanced threat detection systems.'
            }
          },
          {
            '@type': 'Question',
            'name': 'What kind of support do you offer?',
            'acceptedAnswer': {
              '@type': 'Answer',
              'text': 'We offer multiple support channels including comprehensive documentation, community forums, email support, and priority support for paid customers. Enterprise customers get dedicated support managers and 24/7 phone support.'
            }
          }
        ]
      }}
    />
  );
}

interface BlogPostStructuredDataProps {
  post: {
    title: string;
    excerpt: string;
    content: string;
    author: {
      name: string;
      bio?: string;
      avatar?: string;
    };
    publishedAt: string;
    updatedAt?: string;
    tags?: string[];
    slug: string;
    readingTime?: number;
    image?: string;
  };
  url: string;
}

export function BlogPostStructuredData({ post, url }: BlogPostStructuredDataProps) {
  const articleBody = post.content.replace(/<[^>]*>/g, '').substring(0, 5000); // Strip HTML and limit length

  return (
    <JsonLd
      item={{
        '@context': 'https://schema.org',
        '@type': 'BlogPosting',
        'headline': post.title,
        'description': post.excerpt,
        'articleBody': articleBody,
        'author': {
          '@type': 'Person',
          'name': post.author.name,
          'description': post.author.bio,
          'image': post.author.avatar
        },
        'publisher': {
          '@type': 'Organization',
          'name': 'FunctionFly',
          'logo': {
            '@type': 'ImageObject',
            'url': 'https://functionfly.com/logo.png'
          }
        },
        'datePublished': post.publishedAt,
        'dateModified': post.updatedAt || post.publishedAt,
        'mainEntityOfPage': {
          '@type': 'WebPage',
          '@id': url
        },
        'image': post.image ? [post.image] : ['https://functionfly.com/og-image.jpg'],
        'keywords': post.tags?.join(', '),
        'timeRequired': post.readingTime ? `PT${Math.ceil(post.readingTime)}M` : undefined,
        'wordCount': articleBody.split(/\s+/).length,
        'inLanguage': 'en-US',
        'isPartOf': {
          '@type': 'Blog',
          'name': 'FunctionFly Blog',
          'url': 'https://functionfly.com/blog'
        }
      }}
    />
  );
}

interface BreadcrumbStructuredDataProps {
  items: Array<{
    name: string;
    url: string;
  }>;
}

export function BreadcrumbStructuredData({ items }: BreadcrumbStructuredDataProps) {
  return (
    <JsonLd
      item={{
        '@context': 'https://schema.org',
        '@type': 'BreadcrumbList',
        'itemListElement': items.map((item, index) => ({
          '@type': 'ListItem',
          'position': index + 1,
          'name': item.name,
          'item': item.url
        }))
      }}
    />
  );
}

interface BlogPageStructuredDataProps {
  posts?: Array<{
    title: string;
    excerpt: string;
    author: {
      name: string;
    };
    publishedAt: string;
    slug: string;
    tags?: string[];
  }>;
}

export function BlogPageStructuredData({ posts }: BlogPageStructuredDataProps) {
  return (
    <JsonLd
      item={{
        '@context': 'https://schema.org',
        '@graph': [
          {
            '@type': 'Blog',
            'name': 'FunctionFly Blog',
            'description': 'Insights, tutorials, and best practices for serverless functions, edge computing, and AI agents',
            'url': 'https://functionfly.dev/blog',
            'publisher': {
              '@type': 'Organization',
              'name': 'FunctionFly',
              'logo': {
                '@type': 'ImageObject',
                'url': 'https://functionfly.dev/logo.png'
              }
            },
            'mainEntityOfPage': {
              '@type': 'WebPage',
              '@id': 'https://functionfly.dev/blog'
            }
          },
          ...(posts ? posts.slice(0, 10).map((post, index) => ({
            '@type': 'BlogPosting',
            'headline': post.title,
            'description': post.excerpt,
            'author': {
              '@type': 'Person',
              'name': post.author.name
            },
            'datePublished': post.publishedAt,
            'url': `https://functionfly.dev/blog/${post.slug}`,
            'keywords': post.tags?.join(', '),
            'position': index + 1,
            'isPartOf': {
              '@type': 'Blog',
              'name': 'FunctionFly Blog',
              'url': 'https://functionfly.dev/blog'
            }
          })) : [])
        ]
      }}
    />
  );
}
