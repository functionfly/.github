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
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();

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
      title: t('helpCenter.gettingStarted'),
      description: t('helpCenter.gettingStartedDesc'),
      links: [
        { label: t('helpCenter.quickStartGuide'), href: `${DOCS_SITE_URL}/quickstart`, external: true },
        { label: t('helpCenter.platformOverview'), href: '/features' },
        {
          label: t('helpCenter.deployFirstFunction'),
          href: `${DOCS_SITE_URL}/deploy/first-function`,
          external: true,
        },
        { label: t('helpCenter.onboardingWalkthrough'), href: '/onboarding' },
      ],
    },
    {
      icon: <BookOpen className="h-6 w-6" />,
      title: t('helpCenter.documentation'),
      description: t('helpCenter.documentationDesc'),
      links: [
        { label: t('helpCenter.fullDocumentation'), href: DOCS_SITE_URL, external: true },
        { label: t('helpCenter.apiReference'), href: `${DOCS_SITE_URL}/api`, external: true },
        { label: t('helpCenter.sdkIntegrations'), href: '/sdk-integrations' },
        { label: t('helpCenter.codeSamples'), href: `${DOCS_SITE_URL}/examples`, external: true },
      ],
    },
    {
      icon: <Layout className="h-6 w-6" />,
      title: t('helpCenter.buildingDeploying'),
      description: t('helpCenter.buildingDeployingDesc'),
      links: [
        { label: t('helpCenter.functionBasics'), href: `${DOCS_SITE_URL}/functions/basics`, external: true },
        { label: t('helpCenter.stateFabricGuide'), href: `${DOCS_SITE_URL}/state-fabric`, external: true },
        { label: t('helpCenter.secretsVault'), href: `${DOCS_SITE_URL}/secrets`, external: true },
        { label: t('helpCenter.versionControl'), href: `${DOCS_SITE_URL}/functions/versions`, external: true },
      ],
    },
    {
      icon: <CreditCard className="h-6 w-6" />,
      title: t('helpCenter.billingPlans'),
      description: t('helpCenter.billingPlansDesc'),
      links: [
        { label: t('helpCenter.pricingInformation'), href: '/pricing' },
        { label: t('helpCenter.billingSettings'), href: '/settings' },
        { label: t('helpCenter.usageTracking'), href: '/usage' },
        { label: t('helpCenter.enterprisePlans'), href: '/enterprise' },
      ],
    },
    {
      icon: <Key className="h-6 w-6" />,
      title: t('helpCenter.apiKeysAuth'),
      description: t('helpCenter.apiKeysAuthDesc'),
      links: [
        { label: t('helpCenter.apiKeysDashboard'), href: '/api-keys' },
        { label: t('helpCenter.authGuide'), href: `${DOCS_SITE_URL}/auth`, external: true },
        { label: t('helpCenter.securityBestPractices'), href: '/security' },
        { label: t('helpCenter.providerConnections'), href: '/providers' },
      ],
    },
    {
      icon: <Users className="h-6 w-6" />,
      title: t('helpCenter.teamsCollaboration'),
      description: t('helpCenter.teamsCollaborationDesc'),
      links: [
        { label: t('helpCenter.teamManagement'), href: '/teams' },
        { label: t('helpCenter.memberRoles'), href: `${DOCS_SITE_URL}/teams/roles`, external: true },
        { label: t('helpCenter.sharedFunctions'), href: '/functions' },
        { label: t('helpCenter.organizationSetup'), href: `${DOCS_SITE_URL}/teams`, external: true },
      ],
    },
    {
      icon: <Bug className="h-6 w-6" />,
      title: t('helpCenter.troubleshooting'),
      description: t('helpCenter.troubleshootingDesc'),
      links: [
        { label: t('helpCenter.commonIssues'), href: '/faq' },
        { label: t('helpCenter.errorReference'), href: `${DOCS_SITE_URL}/errors`, external: true },
        { label: t('helpCenter.debuggingGuide'), href: `${DOCS_SITE_URL}/debugging`, external: true },
        { label: t('helpCenter.statusPage'), href: 'https://status.functionfly.com', external: true },
      ],
    },
    {
      icon: <LifeBuoy className="h-6 w-6" />,
      title: t('helpCenter.supportChannels'),
      description: t('helpCenter.supportChannelsDesc'),
      links: [
        { label: t('helpCenter.contactSupport'), href: '/contact' },
        { label: t('helpCenter.communityForum'), href: '/community' },
        { label: t('helpCenter.enterpriseSupport'), href: '/enterprise/support' },
        { label: t('helpCenter.featureRequests'), href: '/feedback' },
      ],
    },
  ];

  const quickLinks: QuickLinkProps[] = [
    {
      icon: <Play className="h-5 w-5" />,
      title: t('helpCenter.videoTutorials'),
      href: `${DOCS_SITE_URL}/tutorials`,
      external: true,
    },
    { icon: <FileText className="h-5 w-5" />, title: t('helpCenter.changelog'), href: '/changelog' },
    {
      icon: <Server className="h-5 w-5" />,
      title: t('helpCenter.systemStatus'),
      href: 'https://status.functionfly.com',
      external: true,
    },
    { icon: <Shield className="h-5 w-5" />, title: t('helpCenter.security'), href: '/security' },
    { icon: <Search className="h-5 w-5" />, title: t('helpCenter.faq'), href: '/faq' },
    { icon: <Mail className="h-5 w-5" />, title: t('helpCenter.contactUs'), href: '/contact' },
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
        title={t('helpCenter.metaTitle')}
        description={t('helpCenter.metaDescription')}
      />

      <Navbar />

      {/* Hero Section */}
      <div className="help-hero">
        <div className="max-w-4xl mx-auto px-4 lg:px-6 text-center">
          <div className="help-hero-badge">
            <HelpCircle className="h-4 w-4" />
            {t('helpCenter.helpCenter')}
          </div>
          <h1 className="help-hero-title">{t('helpCenter.heroTitle')}</h1>
          <p className="help-hero-description">
            {t('helpCenter.heroDescription')}
          </p>

          {/* Search Bar */}
          <div className="help-search-container">
            <div className="help-search-input-wrapper">
              <Search className="help-search-icon" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t('helpCenter.searchPlaceholder')}
                className="help-search-input"
              />
              {searchQuery && (
                <button onClick={() => setSearchQuery('')} className="help-search-clear">
                  {t('helpCenter.clear')}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Quick Links */}
      <div className="help-quick-links-section">
        <div className="max-w-7xl mx-auto px-4 lg:px-6">
          <h2 className="help-section-title">{t('helpCenter.quickLinks')}</h2>
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
          <h2 className="help-section-title">{t('helpCenter.browseByTopic')}</h2>

          {filteredCategories.length === 0 ? (
            <div className="help-no-results">
              <HelpCircle className="help-no-results-icon" />
              <h3 className="help-no-results-title">{t('helpCenter.noResultsFound')}</h3>
              <p className="help-no-results-description">
                {t('helpCenter.noResultsDescription')}
              </p>
              <button onClick={() => setSearchQuery('')} className="help-clear-search-button">
                {t('helpCenter.clearSearch')}
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
                <h3 className="help-support-title">{t('helpCenter.stillNeedHelp')}</h3>
                <p className="help-support-description">
                  {t('helpCenter.stillNeedHelpDescription')}
                </p>
              </div>
              <div className="help-support-actions">
                <Link to="/contact" className="help-support-button primary">
                  <Mail className="h-4 w-4" />
                  {t('helpCenter.contactSupport')}
                </Link>
                <a
                  href={DOCS_SITE_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="help-support-button secondary"
                >
                  <BookOpen className="h-4 w-4" />
                  {t('helpCenter.documentation')}
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
