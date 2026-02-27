import Link from 'next/link';

export default function HomePage() {
  return (
    <main className="min-h-screen">
      {/* Hero Section */}
      <section className="relative overflow-hidden bg-gradient-to-br from-[var(--bg-primary)] via-[var(--bg-secondary)] to-[var(--bg-primary)]">
        <div className="absolute inset-0 bg-[linear-gradient(rgba(99,102,241,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(99,102,241,0.03)_1px,transparent_1px)] bg-[size:32px_32px]" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 md:py-32">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold text-[var(--text-primary)] mb-6 tracking-tight">
              Serverless
              <span className="block mt-2">
                <span className="gradient-text">Reliability</span>
              </span>
              Done Right
            </h1>

            <p className="text-lg md:text-xl text-[var(--text-secondary)] mb-8 max-w-3xl mx-auto leading-relaxed">
              Eliminate serverless downtime with intelligent multi-cloud failover.
              FunctionFly automatically routes traffic to healthy functions across AWS Lambda,
              Cloudflare Workers, and Vercel.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
              <Link
                href="/docs"
                className="inline-flex items-center justify-center px-8 py-4 text-base font-medium text-white bg-[var(--color-brand-500)] hover:bg-[var(--color-brand-600)] rounded-lg transition-all duration-200 hover:shadow-lg hover:shadow-[var(--color-brand-500)]/25"
              >
                Get Started
                <svg className="ml-2 -mr-1 w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
              </Link>

              <Link
                href="/pricing"
                className="inline-flex items-center justify-center px-8 py-4 text-base font-medium text-[var(--text-primary)] bg-[var(--bg-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-hover)] rounded-lg transition-all duration-200"
              >
                View Pricing
              </Link>
            </div>
          </div>

          {/* Stats Section */}
          <div className="mt-20 grid grid-cols-2 md:grid-cols-4 gap-8">
            {[
              { label: 'Uptime SLA', value: '99.9%' },
              { label: 'Failover Time', value: '<30ms' },
              { label: 'Supported Providers', value: '4+' },
              { label: 'Active Regions', value: '30+' },
            ].map((stat) => (
              <div key={stat.label} className="text-center">
                <div className="text-3xl md:text-4xl font-bold text-[var(--color-brand-400)] mb-2">
                  {stat.value}
                </div>
                <div className="text-sm text-[var(--text-muted)]">
                  {stat.label}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-[var(--bg-secondary)]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-[var(--text-primary)] mb-4">
              Why FunctionFly?
            </h2>
            <p className="text-lg text-[var(--text-secondary)] max-w-2xl mx-auto">
              Built for developers who demand reliability without sacrificing flexibility.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
            {[
              {
                title: 'Multi-Cloud Failover',
                description: 'Automatically route traffic to healthy backends across AWS Lambda, Cloudflare Workers, Vercel, and more.',
                icon: (
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064" />
                  </svg>
                ),
              },
              {
                title: 'Real-Time Health Checks',
                description: 'Continuous monitoring with intelligent circuit breakers to detect and respond to issues instantly.',
                icon: (
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                ),
              },
              {
                title: 'Instant Failover',
                description: 'Sub-30ms failover times ensure your users never experience downtime or latency spikes.',
                icon: (
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                ),
              },
              {
                title: 'Zero Lock-In',
                description: 'Bring your own cloud accounts. We never lock you into our infrastructure.',
                icon: (
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                ),
              },
              {
                title: 'Smart Traffic Routing',
                description: 'Route based on latency, geography, or custom rules for optimal performance.',
                icon: (
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" />
                  </svg>
                ),
              },
              {
                title: 'Developer First',
                description: 'Simple API, comprehensive docs, and SDKs for your favorite languages.',
                icon: (
                  <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                  </svg>
                ),
              },
            ].map((feature) => (
              <div
                key={feature.title}
                className="p-6 rounded-xl bg-[var(--bg-tertiary)] border border-[var(--border-subtle)] hover:border-[var(--border-default)] transition-all duration-200"
              >
                <div className="w-12 h-12 rounded-lg bg-[var(--color-brand-500)]/10 flex items-center justify-center text-[var(--color-brand-400)] mb-4">
                  {feature.icon}
                </div>
                <h3 className="text-xl font-semibold text-[var(--text-primary)] mb-2">
                  {feature.title}
                </h3>
                <p className="text-[var(--text-secondary)]">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-[var(--text-primary)] mb-6">
            Ready to eliminate serverless downtime?
          </h2>
          <p className="text-lg text-[var(--text-secondary)] mb-8">
            Start with our free tier. No credit card required.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/docs"
              className="inline-flex items-center justify-center px-8 py-4 text-base font-medium text-white bg-[var(--color-brand-500)] hover:bg-[var(--color-brand-600)] rounded-lg transition-all duration-200"
            >
              Get Started Free
            </Link>
            <Link
              href="/contact"
              className="inline-flex items-center justify-center px-8 py-4 text-base font-medium text-[var(--text-primary)] bg-[var(--bg-secondary)] border border-[var(--border-default)] hover:bg-[var(--bg-hover)] rounded-lg transition-all duration-200"
            >
              Contact Sales
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 bg-[var(--bg-secondary)] border-t border-[var(--border-subtle)]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid md:grid-cols-4 gap-8">
            <div>
              <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-4">FunctionFly</h3>
              <p className="text-sm text-[var(--text-muted)]">
                Serverless reliability platform with multi-cloud failover.
              </p>
            </div>
            <div>
              <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-4">Product</h4>
              <ul className="space-y-2 text-sm text-[var(--text-muted)]">
                <li><Link href="/docs" className="hover:text-[var(--text-primary)] transition-colors">Documentation</Link></li>
                <li><Link href="/pricing" className="hover:text-[var(--text-primary)] transition-colors">Pricing</Link></li>
                <li><Link href="/blog" className="hover:text-[var(--text-primary)] transition-colors">Blog</Link></li>
              </ul>
            </div>
            <div>
              <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-4">Company</h4>
              <ul className="space-y-2 text-sm text-[var(--text-muted)]">
                <li><Link href="/about" className="hover:text-[var(--text-primary)] transition-colors">About</Link></li>
                <li><Link href="/careers" className="hover:text-[var(--text-primary)] transition-colors">Careers</Link></li>
                <li><Link href="/contact" className="hover:text-[var(--text-primary)] transition-colors">Contact</Link></li>
              </ul>
            </div>
            <div>
              <h4 className="text-sm font-semibold text-[var(--text-primary)] mb-4">Legal</h4>
              <ul className="space-y-2 text-sm text-[var(--text-muted)]">
                <li><Link href="/privacy" className="hover:text-[var(--text-primary)] transition-colors">Privacy Policy</Link></li>
                <li><Link href="/terms" className="hover:text-[var(--text-primary)] transition-colors">Terms of Service</Link></li>
              </ul>
            </div>
          </div>
          <div className="mt-8 pt-8 border-t border-[var(--border-subtle)] text-center text-sm text-[var(--text-muted)]">
            © {new Date().getFullYear()} FunctionFly. All rights reserved.
          </div>
        </div>
      </footer>
    </main>
  );
}
