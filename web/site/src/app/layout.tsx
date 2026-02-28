import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  metadataBase: new URL('https://functionfly.com'),
  title: {
    default: 'FunctionFly - Serverless Reliability Platform',
    template: '%s | FunctionFly',
  },
  description: 'Serverless reliability platform with multi-cloud failover for AWS Lambda, Cloudflare Workers, and Vercel Functions.',
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://functionfly.com',
    siteName: 'FunctionFly',
    title: 'FunctionFly - Serverless Reliability Platform',
    description: 'Serverless reliability platform with multi-cloud failover for AWS Lambda, Cloudflare Workers, and Vercel Functions.',
    images: [
      {
        url: '/og.png',
        width: 1200,
        height: 630,
        alt: 'FunctionFly',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'FunctionFly - Serverless Reliability Platform',
    description: 'Serverless reliability platform with multi-cloud failover for AWS Lambda, Cloudflare Workers, and Vercel Functions.',
    images: ['/og.png'],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  verification: {
    google: 'google-site-verification-code',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.svg" />
        <link rel="canonical" href="https://functionfly.com" />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              '@context': 'https://schema.org',
              '@type': 'Organization',
              name: 'FunctionFly',
              url: 'https://functionfly.com',
              logo: 'https://functionfly.com/logo/logo-full.svg',
              description: 'Serverless reliability platform with multi-cloud failover for AWS Lambda, Cloudflare Workers, and Vercel Functions.',
              foundingDate: '2024',
              contactPoint: {
                '@type': 'ContactPoint',
                email: 'hello@functionfly.com',
                contactType: 'customer service',
              },
              sameAs: [
                'https://github.com/functionfly',
                'https://twitter.com/functionfly',
                'https://linkedin.com/company/functionfly',
              ],
            }),
          }}
        />
      </head>
      <body className="bg-[var(--bg-primary)] text-[var(--text-primary)] min-h-screen">
        {children}
      </body>
    </html>
  );
}
