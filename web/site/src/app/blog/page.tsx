import Link from 'next/link';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Blog - FunctionFly | Serverless Reliability Insights',
  description: 'Stay updated with the latest in serverless reliability, multi-cloud architectures, and best practices from the FunctionFly team.',
};

// Mock blog posts - in production, fetch from Sanity
const blogPosts = [
  {
    slug: 'introducing-functionfly',
    title: 'Introducing FunctionFly: Serverless Reliability Done Right',
    description: 'We\'re building the future of serverless reliability. Learn how FunctionFly eliminates downtime across AWS Lambda, Cloudflare Workers, and Vercel.',
    category: 'Announcements',
    author: { name: 'FunctionFly Team' },
    publishedAt: '2024-01-15',
  },
  {
    slug: 'multi-cloud-failover-guide',
    title: 'A Deep Dive into Multi-Cloud Failover Strategies',
    description: 'Learn the best practices for implementing multi-cloud failover in your serverless applications.',
    category: 'Technical',
    author: { name: 'Engineering Team' },
    publishedAt: '2024-02-01',
  },
  {
    slug: 'serverless-architecture-best-practices',
    title: 'Serverless Architecture Best Practices for 2024',
    description: 'Discover the latest best practices for building reliable serverless applications in production.',
    category: 'Guides',
    author: { name: 'Solutions Team' },
    publishedAt: '2024-02-15',
  },
];

export default function BlogIndexPage() {
  return (
    <main className="min-h-screen">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-br from-[var(--bg-primary)] via-[var(--bg-secondary)] to-[var(--bg-primary)] py-24">
        <div className="absolute inset-0 bg-[linear-gradient(rgba(99,102,241,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(99,102,241,0.03)_1px,transparent_1px)] bg-[size:32px_32px]" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-bold text-[var(--text-primary)] mb-6">
              FunctionFly
              <span className="block mt-2 gradient-text">Blog</span>
            </h1>
            <p className="text-lg md:text-xl text-[var(--text-secondary)] mb-8 max-w-3xl mx-auto">
              Insights, best practices, and deep dives into serverless reliability and multi-cloud architectures.
            </p>
          </div>
        </div>
      </section>

      {/* Blog Posts */}
      <section className="py-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
            {blogPosts.map((post) => (
              <article
                key={post.slug}
                className="bg-[var(--bg-secondary)] rounded-xl p-6 border border-[var(--border-subtle)] hover:border-[var(--color-brand-all duration-200-500)] transition"
              >
                <div className="mb-4">
                  <span className="bg-[var(--color-brand-500)]/10 text-[var(--color-brand-400)] px-3 py-1 rounded-full text-sm font-medium">
                    {post.category}
                  </span>
                </div>

                <h2 className="text-xl font-bold text-[var(--text-primary)] mb-3">
                  <Link
                    href={`/blog/${post.slug}`}
                    className="hover:text-[var(--color-brand-400)] transition-colors"
                  >
                    {post.title}
                  </Link>
                </h2>

                <p className="text-[var(--text-secondary)] mb-4">{post.description}</p>

                <div className="flex items-center justify-between text-sm text-[var(--text-muted)]">
                  <span>By {post.author.name}</span>
                  <time dateTime={post.publishedAt}>
                    {new Date(post.publishedAt).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </time>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
