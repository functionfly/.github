import Link from 'next/link';
import { Metadata } from 'next';

interface Props {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;

  return {
    title: `${slug.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase())} | FunctionFly Blog`,
    description: 'Read more about serverless reliability on the FunctionFly blog.',
  };
}

// Mock post data - in production, fetch from Sanity
const posts: Record<string, {
  title: string;
  description: string;
  content: string;
  category: string;
  author: { name: string };
  publishedAt: string;
}> = {
  'introducing-functionfly': {
    title: 'Introducing FunctionFly: Serverless Reliability Done Right',
    description: 'We\'re building the future of serverless reliability.',
    content: 'Full article content would be rendered here from Sanity...',
    category: 'Announcements',
    author: { name: 'FunctionFly Team' },
    publishedAt: '2024-01-15',
  },
};

export async function generateStaticParams() {
  return Object.keys(posts).map((slug) => ({
    slug,
  }));
}

export default async function BlogPostPage({ params }: Props) {
  const { slug } = await params;
  const post = posts[slug] || {
    title: slug.replace(/-/g, ' ').replace(/\b\w/g, c => c.toUpperCase()),
    description: 'Article not found',
    content: 'The article you\'re looking for doesn\'t exist or is being written.',
    category: 'Uncategorized',
    author: { name: 'Unknown' },
    publishedAt: new Date().toISOString(),
  };

  return (
    <main className="min-h-screen">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-br from-[var(--bg-primary)] via-[var(--bg-secondary)] to-[var(--bg-primary)] py-16">
        <div className="absolute inset-0 bg-[linear-gradient(rgba(99,102,241,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(99,102,241,0.03)_1px,transparent_1px)] bg-[size:32px_32px]" />

        <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <Link
            href="/blog"
            className="inline-flex items-center text-[var(--text-muted)] hover:text-[var(--color-brand-400)] transition-colors mb-8"
          >
            <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Blog
          </Link>

          <div className="mb-4">
            <span className="bg-[var(--color-brand-500)]/10 text-[var(--color-brand-400)] px-3 py-1 rounded-full text-sm font-medium">
              {post.category}
            </span>
          </div>

          <h1 className="text-3xl md:text-5xl font-bold text-[var(--text-primary)] mb-6">
            {post.title}
          </h1>

          <div className="flex items-center gap-4 text-[var(--text-muted)]">
            <span>By {post.author.name}</span>
            <span>•</span>
            <time dateTime={post.publishedAt}>
              {new Date(post.publishedAt).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'long',
                day: 'numeric',
              })}
            </time>
          </div>
        </div>
      </section>

      {/* Content */}
      <section className="py-16">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <article className="prose prose-invert max-w-none">
            <div className="text-[var(--text-secondary)] text-lg leading-relaxed whitespace-pre-line">
              {post.content}
            </div>
          </article>
        </div>
      </section>
    </main>
  );
}
