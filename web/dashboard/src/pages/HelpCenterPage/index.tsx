'use client';

import { Navbar } from '@/components/common/Navbar';
import { MetaTags } from '@/components/seo/MetaTags';
import { reportWebVitalsBatch, useWebVitals } from '@/hooks/useWebVitals';
import { DOCS_SITE_URL } from '@/lib/constants';
import { Footer } from '@/pages/LandingPage/components/Footer';
import {
  BookOpen,
  Bug,
  ChevronRight,
  CreditCard,
  ExternalLink,
  FileText,
  HelpCircle,
  Key,
  Layout,
  LifeBuoy,
  Mail,
  MessageSquare,
  Play,
  Rocket,
  Search,
  Server,
  Shield,
  Users,
} from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import './HelpCenter.css';

interface HelpCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
  links: Array<{
    label: string;
    href: string;
    external?: boolean;
  }>;
}

function HelpCard({ icon, title, description, links }: HelpCardProps) {
  return (
    <div className="help-card">
      <div className="help-card-header">
        <div className="help-card-icon">{icon}</div>
        <h3 className="help-card-title">{title}</h3>
      </div>
      <p className="help-card-description">{description}</p>
      <ul className="help-card-links">
        {links.map((link, index) => (
          <li key={index}>
            {link.external ? (
              <a
                href={link.href}
                target="_blank"
                rel="noopener noreferrer"
                className="help-card-link"
              >
                <ExternalLink className="h-3 w-3" />
                {link.label}
              </a>
            ) : (
              <Link to={link.href} className="help-card-link">
                <ChevronRight className="h-3 w-3" />
                {link.label}
              </Link>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}

interface QuickLinkProps {
  icon: React.ReactNode;
  title: string;
  href: string;
  external?: boolean;
}

function QuickLink({ icon, title, href, external }: QuickLinkProps) {
  if (external) {
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" className="quick-link">
        <div className="quick-link-icon">{icon}</div>
        <span className="quick-link-title">{title}</span>
        <ExternalLink className="quick-link-arrow" />
      </a>
    );
  }
  return (
    <Link to={href} className="quick-link">
      <div className="quick-link-icon">{icon}</div>
      <span className="quick-link-title">{title}</span>
      <ChevronRight className="quick-link-arrow" />
    </Link>
  );
}

export function HelpCenterPage() {
  const [searchQuery, setSearchQuery] = useState('');

  useWebVitals((metrics) => {
    if (import.meta.env.DEV) {
      // eslint-disable-next-line no-console
      console.log('Web Vitals:', metrics);
    } else {
      reportWebVitalsBatch(metrics, { page: 'help-center' });
    }
  });

  const helpCategories: HelpCardProps[] = [
    {
      icon: <Rocket className="h-6 w-6" />,
      title: 'Getting Started',
      description: 'New to FunctionFly? Learn the basics and deploy your first function.',
      links: [
        { label: 'Quick Start Guide', href: `${DOCS_SITE_URL}/quickstart`, external: true },
        { label: 'Platform Overview', href: '/features' },
        {
          label: 'Deploy Your First Function',
          href: `${DOCS_SITE_URL}/deploy/first-function`,
          external: true,
        },
        { label: 'Onboarding Walkthrough', href: '/onboarding' },
      ],
    },
    {
      icon: <BookOpen className="h-6 w-6" />,
      title: 'Documentation',
      description: 'Comprehensive guides, API references, and examples.',
      links: [
        { label: 'Full Documentation', href: DOCS_SITE_URL, external: true },
        { label: 'API Reference', href: `${DOCS_SITE_URL}/api`, external: true },
        { label: 'SDK Integrations', href: '/sdk-integrations' },
        { label: 'Code Samples', href: `${DOCS_SITE_URL}/examples`, external: true },
      ],
    },
    {
      icon: <Layout className="h-6 w-6" />,
      title: 'Building & Deploying',
      description: 'Learn how to create, deploy, and manage functions.',
      links: [
        { label: 'Function Basics', href: `${DOCS_SITE_URL}/functions/basics`, external: true },
        { label: 'State Fabric Guide', href: `${DOCS_SITE_URL}/state-fabric`, external: true },
        { label: 'Secrets Vault', href: `${DOCS_SITE_URL}/secrets`, external: true },
        { label: 'Version Control', href: `${DOCS_SITE_URL}/functions/versions`, external: true },
      ],
    },
    {
      icon: <CreditCard className="h-6 w-6" />,
      title: 'Billing & Plans',
      description: 'Manage your subscription, usage, and billing details.',
      links: [
        { label: 'Pricing Information', href: '/pricing' },
        { label: 'Billing Settings', href: '/settings' },
        { label: 'Usage Tracking', href: '/usage' },
        { label: 'Enterprise Plans', href: '/enterprise' },
      ],
    },
    {
      icon: <Key className="h-6 w-6" />,
      title: 'API Keys & Authentication',
      description: 'Generate keys, manage access, and secure your functions.',
      links: [
        { label: 'API Keys Dashboard', href: '/api-keys' },
        { label: 'Authentication Guide', href: `${DOCS_SITE_URL}/auth`, external: true },
        { label: 'Security Best Practices', href: '/security' },
        { label: 'Provider Connections', href: '/providers' },
      ],
    },
    {
      icon: <Users className="h-6 w-6" />,
      title: 'Teams & Collaboration',
      description: 'Work together with team members and manage access.',
      links: [
        { label: 'Team Management', href: '/teams' },
        { label: 'Member Roles', href: `${DOCS_SITE_URL}/teams/roles`, external: true },
        { label: 'Shared Functions', href: '/functions' },
        { label: 'Organization Setup', href: `${DOCS_SITE_URL}/teams`, external: true },
      ],
    },
    {
      icon: <Bug className="h-6 w-6" />,
      title: 'Troubleshooting',
      description: 'Debug issues, find solutions, and get unstuck.',
      links: [
        { label: 'Common Issues', href: '/faq' },
        { label: 'Error Reference', href: `${DOCS_SITE_URL}/errors`, external: true },
        { label: 'Debugging Guide', href: `${DOCS_SITE_URL}/debugging`, external: true },
        { label: 'Status Page', href: 'https://status.functionfly.com', external: true },
      ],
    },
    {
      icon: <LifeBuoy className="h-6 w-6" />,
      title: 'Support Channels',
      description: 'Get help from our team and community.',
      links: [
        { label: 'Contact Support', href: '/contact' },
        { label: 'Community Forum', href: '/community' },
        { label: 'Enterprise Support', href: '/enterprise/support' },
        { label: 'Feature Requests', href: '/feedback' },
      ],
    },
  ];

  const quickLinks: QuickLinkProps[] = [
    {
      icon: <Play className="h-5 w-5" />,
      title: 'Video Tutorials',
      href: `${DOCS_SITE_URL}/tutorials`,
      external: true,
    },
    { icon: <FileText className="h-5 w-5" />, title: 'Changelog', href: '/changelog' },
    {
      icon: <Server className="h-5 w-5" />,
      title: 'System Status',
      href: 'https://status.functionfly.com',
      external: true,
    },
    { icon: <Shield className="h-5 w-5" />, title: 'Security', href: '/security' },
    { icon: <Search className="h-5 w-5" />, title: 'FAQ', href: '/faq' },
    { icon: <Mail className="h-5 w-5" />, title: 'Contact Us', href: '/contact' },
  ];

  const filteredCategories = searchQuery
    ? helpCategories.filter(
        (cat) =>
          cat.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
          cat.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
          cat.links.some((link) => link.label.toLowerCase().includes(searchQuery.toLowerCase()))
      )
    : helpCategories;

  return (
    <div className="min-h-screen bg-bg-primary" style={{ backgroundColor: 'var(--bg-primary)' }}>
      <MetaTags
        title="Help Center | FunctionFly"
        description="Get help with FunctionFly. Browse documentation, guides, FAQs, and support resources."
      />

      <Navbar />

      {/* Hero Section */}
      <div className="help-hero">
        <div className="max-w-4xl mx-auto px-4 lg:px-6 text-center">
          <div className="help-hero-badge">
            <HelpCircle className="h-4 w-4" />
            Help Center
          </div>
          <h1 className="help-hero-title">How can we help you?</h1>
          <p className="help-hero-description">
            Find answers, explore documentation, and get support for your FunctionFly journey.
          </p>

          {/* Search Bar */}
          <div className="help-search-container">
            <div className="help-search-input-wrapper">
              <Search className="help-search-icon" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search for topics, guides, or answers..."
                className="help-search-input"
              />
              {searchQuery && (
                <button onClick={() => setSearchQuery('')} className="help-search-clear">
                  Clear
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Quick Links */}
      <div className="help-quick-links-section">
        <div className="max-w-7xl mx-auto px-4 lg:px-6">
          <h2 className="help-section-title">Quick Links</h2>
          <div className="quick-links-grid">
            {quickLinks.map((link, index) => (
              <QuickLink key={index} {...link} />
            ))}
          </div>
        </div>
      </div>

      {/* Help Categories */}
      <div className="help-categories-section">
        <div className="max-w-7xl mx-auto px-4 lg:px-6">
          <h2 className="help-section-title">Browse by Topic</h2>

          {filteredCategories.length === 0 ? (
            <div className="help-no-results">
              <HelpCircle className="help-no-results-icon" />
              <h3 className="help-no-results-title">No results found</h3>
              <p className="help-no-results-description">
                Try adjusting your search terms or browse all categories below.
              </p>
              <button onClick={() => setSearchQuery('')} className="help-clear-search-button">
                Clear Search
              </button>
            </div>
          ) : (
            <div className="help-categories-grid">
              {filteredCategories.map((category, index) => (
                <HelpCard key={index} {...category} />
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Contact Support CTA */}
      <div className="help-support-cta">
        <div className="max-w-4xl mx-auto px-4 lg:px-6">
          <div className="help-support-card">
            <div className="help-support-content">
              <div className="help-support-icon">
                <MessageSquare className="h-8 w-8" />
              </div>
              <div className="help-support-text">
                <h3 className="help-support-title">Still need help?</h3>
                <p className="help-support-description">
                  Can&apos;t find what you&apos;re looking for? Our support team is ready to assist
                  you.
                </p>
              </div>
              <div className="help-support-actions">
                <Link to="/contact" className="help-support-button primary">
                  <Mail className="h-4 w-4" />
                  Contact Support
                </Link>
                <a
                  href={DOCS_SITE_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="help-support-button secondary"
                >
                  <BookOpen className="h-4 w-4" />
                  Documentation
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Footer />
    </div>
  );
}
