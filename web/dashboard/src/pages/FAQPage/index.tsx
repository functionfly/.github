'use client';

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { HelpCircle, ChevronDown, Search, MessageSquare, ExternalLink } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { Navbar } from '@/components/common/Navbar';
import { Footer } from '@/pages/LandingPage/components/Footer';
import { MetaTags } from '@/components/seo/MetaTags';
import { FAQPageStructuredData } from '@/components/seo/StructuredData';
import { useWebVitals } from '@/hooks/useWebVitals';

interface FAQItemProps {
  question: string;
  answer: string;
  category?: string;
  isOpen: boolean;
  onToggle: () => void;
}

function FAQItem({ question, answer, category, isOpen, onToggle }: FAQItemProps) {
  return (
    <div className={`faq-item ${isOpen ? 'expanded' : ''}`}>
      <button
        onClick={onToggle}
        className="faq-item-button"
      >
        <div className="faq-item-content">
          <HelpCircle className="faq-item-icon" />
          <div className="faq-item-text">
            <h3 className="faq-item-title">{question}</h3>
            {category && (
              <span className="faq-item-category">{category}</span>
            )}
          </div>
        </div>
        <div className="faq-item-toggle">
          <ChevronDown className="h-4 w-4" />
        </div>
      </button>

      {isOpen && (
        <div className="faq-answer">
          <div className="faq-answer-content">
            {answer}
          </div>
        </div>
      )}
    </div>
  );
}

interface FAQSectionProps {
  title: string;
  icon: React.ReactNode;
  faqs: Array<{
    question: string;
    answer: string;
    category?: string;
  }>;
  openItems: Set<number>;
  onToggle: (index: number) => void;
}

function FAQSection({ title, icon, faqs, openItems, onToggle }: FAQSectionProps) {
  return (
    <div className="faq-section">
      <div className="faq-section-header">
        <h2 className="faq-section-title">
          {icon}
          {title}
        </h2>
      </div>
      <div className="p-6 space-y-0">
        {faqs.map((faq, index) => (
          <FAQItem
            key={index}
            question={faq.question}
            answer={faq.answer}
            category={faq.category}
            isOpen={openItems.has(index)}
            onToggle={() => onToggle(index)}
          />
        ))}
      </div>
    </div>
  );
}

export function FAQPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [openItems, setOpenItems] = useState<Map<string, Set<number>>>(new Map());

  // Monitor Core Web Vitals
  useWebVitals((metrics) => {
    // In production, send to your analytics service
    // For now, only log in development
    if (import.meta.env.DEV) {
      // eslint-disable-next-line no-console
      console.log('Web Vitals:', metrics);
    }
  });

  const categories = [
    { id: 'all', label: 'All Questions', icon: <HelpCircle className="h-4 w-4" /> },
    { id: 'getting-started', label: 'Getting Started', icon: <ExternalLink className="h-4 w-4" /> },
    { id: 'deployment', label: 'Deployment', icon: <MessageSquare className="h-4 w-4" /> },
    { id: 'pricing', label: 'Pricing & Billing', icon: <HelpCircle className="h-4 w-4" /> },
    { id: 'security', label: 'Security', icon: <HelpCircle className="h-4 w-4" /> },
    { id: 'support', label: 'Support', icon: <HelpCircle className="h-4 w-4" /> },
  ];

  const faqData = {
    'getting-started': [
      {
        question: 'What is FunctionFly?',
        answer: 'FunctionFly is an edge computing platform that allows you to deploy serverless functions globally. It provides instant deployment, automatic scaling, and edge execution across multiple cloud providers, giving you unprecedented control over your application\'s performance and reach.',
        category: 'Platform'
      },
      {
        question: 'How do I get started with FunctionFly?',
        answer: 'Getting started is simple! Sign up for a free account, connect your first cloud provider (we support AWS, Google Cloud, Cloudflare, Vercel, and Fly.io), and deploy your first function using our CLI or web dashboard. Check out our documentation for step-by-step guides.',
        category: 'Onboarding'
      },
      {
        question: 'What programming languages are supported?',
        answer: 'FunctionFly supports all major programming languages including JavaScript/TypeScript (Node.js), Python, Go, Rust, Java, PHP, Ruby, and .NET. You can also deploy containerized applications using Docker.',
        category: 'Languages'
      },
      {
        question: 'Do I need to change my existing code?',
        answer: 'No! FunctionFly is designed to work with your existing applications. Most functions can be deployed without any code changes. We provide adapters for popular frameworks and deployment patterns.',
        category: 'Migration'
      }
    ],
    'deployment': [
      {
        question: 'How fast is deployment?',
        answer: 'Deployments typically take 10-30 seconds for cold starts and under 5 seconds for warm deployments. Our global edge network ensures your functions are available worldwide within seconds of deployment.',
        category: 'Performance'
      },
      {
        question: 'Can I deploy to multiple regions simultaneously?',
        answer: 'Absolutely! FunctionFly allows you to deploy to multiple cloud providers and regions simultaneously. You can configure geo-distribution rules to optimize performance and ensure high availability across different geographic locations.',
        category: 'Multi-region'
      },
      {
        question: 'What happens during traffic spikes?',
        answer: 'FunctionFly automatically scales your functions based on incoming traffic. Our intelligent scaling system can handle sudden traffic spikes by distributing load across multiple instances and regions, ensuring your application stays responsive.',
        category: 'Scaling'
      },
      {
        question: 'How do I monitor my deployments?',
        answer: 'Our dashboard provides real-time monitoring, logs, metrics, and alerting. You can also integrate with popular monitoring tools like DataDog, New Relic, and Sentry for comprehensive observability.',
        category: 'Monitoring'
      },
      {
        question: 'Can I use custom domains?',
        answer: 'Yes! You can connect custom domains to your FunctionFly deployments. We provide SSL certificates automatically and support advanced routing rules for complex domain configurations.',
        category: 'Domains'
      }
    ],
    'pricing': [
      {
        question: 'How does FunctionFly pricing work?',
        answer: 'We offer a generous free tier and flexible paid plans based on usage. You pay for compute time, bandwidth, and storage. Our pricing is transparent with no hidden fees, and you can monitor costs in real-time through our dashboard.',
        category: 'Billing'
      },
      {
        question: 'What\'s included in the free tier?',
        answer: 'The free tier includes 100,000 function invocations, 100GB bandwidth, and 1GB storage per month. Perfect for development, testing, and small production applications.',
        category: 'Free Tier'
      },
      {
        question: 'Do you offer enterprise pricing?',
        answer: 'Yes! We offer custom enterprise plans with dedicated support, custom SLAs, advanced security features, and volume discounts. Contact our sales team for a personalized quote.',
        category: 'Enterprise'
      },
      {
        question: 'How do you calculate compute costs?',
        answer: 'Compute costs are based on GB-seconds (memory allocated × execution time in seconds). Different memory configurations have different rates, allowing you to optimize costs based on your function\'s requirements.',
        category: 'Compute'
      }
    ],
    'security': [
      {
        question: 'How secure is FunctionFly?',
        answer: 'Security is our top priority. We implement multiple layers of security including encryption at rest and in transit, regular security audits, compliance with SOC 2, GDPR, and CCPA, and advanced threat detection systems.',
        category: 'Platform Security'
      },
      {
        question: 'Are my functions isolated?',
        answer: 'Yes, each function runs in its own isolated environment with no access to other functions or the underlying infrastructure. We use advanced containerization and sandboxing technologies to ensure complete isolation.',
        category: 'Isolation'
      },
      {
        question: 'What compliance standards do you meet?',
        answer: 'FunctionFly is SOC 2 Type II compliant and meets GDPR, CCPA, and HIPAA requirements. We undergo regular third-party security audits and maintain comprehensive security documentation.',
        category: 'Compliance'
      },
      {
        question: 'How do you handle data encryption?',
        answer: 'All data is encrypted at rest using AES-256 encryption and in transit using TLS 1.3. We also provide client-side encryption options for sensitive data and support customer-managed encryption keys.',
        category: 'Encryption'
      }
    ],
    'support': [
      {
        question: 'What kind of support do you offer?',
        answer: 'We offer multiple support channels including comprehensive documentation, community forums, email support, and priority support for paid customers. Enterprise customers get dedicated support managers and 24/7 phone support.',
        category: 'Support Options'
      },
      {
        question: 'Do you offer training or consulting?',
        answer: 'Yes! We provide training workshops, architectural reviews, performance optimization consulting, and migration assistance. Our team of experts can help you get the most out of FunctionFly.',
        category: 'Professional Services'
      },
      {
        question: 'How do I report security issues?',
        answer: 'We have a dedicated security team and responsible disclosure program. If you discover a security vulnerability, please email security@functionfly.com with detailed information. We offer bounties for valid security reports.',
        category: 'Security Issues'
      },
      {
        question: 'Where can I find API documentation?',
        answer: 'Our comprehensive API documentation is available at docs.functionfly.com. It includes interactive examples, SDK documentation, and integration guides for all supported languages and frameworks.',
        category: 'Documentation'
      }
    ]
  };

  const toggleItem = (sectionId: string, index: number) => {
    setOpenItems(prev => {
      const newMap = new Map(prev);
      const sectionSet = newMap.get(sectionId) || new Set<number>();

      if (sectionSet.has(index)) {
        sectionSet.delete(index);
      } else {
        sectionSet.add(index);
      }

      newMap.set(sectionId, sectionSet);
      return newMap;
    });
  };

  const filteredFaqs = Object.entries(faqData).reduce((acc, [sectionId, faqs]) => {
    if (selectedCategory !== 'all' && selectedCategory !== sectionId) return acc;

    const filtered = faqs.filter(faq =>
      faq.question.toLowerCase().includes(searchQuery.toLowerCase()) ||
      faq.answer.toLowerCase().includes(searchQuery.toLowerCase())
    );

    if (filtered.length > 0) {
      acc[sectionId] = filtered;
    }

    return acc;
  }, {} as Record<string, typeof faqData[keyof typeof faqData]>);

  return (
    <div className="min-h-screen bg-background">
      {/* SEO Meta Tags */}
      <MetaTags
        title="FunctionFly FAQ - Frequently Asked Questions | Serverless Platform"
        description="Find answers to common questions about FunctionFly's serverless platform. Learn about deployment, pricing, security, and getting started with our edge computing solution."
        keywords={["FAQ", "serverless", "functions", "deployment", "pricing", "security", "support", "getting started", "cloud platform"]}
        url="/faq"
      />

      {/* Structured Data */}
      <FAQPageStructuredData />

      <Navbar variant="landing" />

      {/* Header */}
      <div className="faq-header">
        <div className="icon-container animate-float">
          <HelpCircle className="h-8 w-8" />
        </div>
        <h1>Frequently Asked Questions</h1>
        <p>
          Find answers to common questions about FunctionFly. Can't find what you're looking for?
          <Link to="/contact" className="text-primary hover:underline ml-1">
            Contact our support team
          </Link>.
        </p>
      </div>

      <div className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto space-y-8">

          {/* Search and Filters */}
          <div className="faq-search-container">
            <div className="faq-search-input">
              <Search className="h-5 w-5" />
              <input
                type="text"
                placeholder="Search FAQs..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>

            <div className="faq-filter-buttons">
              {categories.map((category) => (
                <button
                  key={category.id}
                  onClick={() => setSelectedCategory(category.id)}
                  className={`faq-filter-button ${selectedCategory === category.id ? 'active' : ''}`}
                >
                  {category.icon}
                  {category.label}
                </button>
              ))}
            </div>
          </div>

          {/* FAQ Sections */}
          {Object.keys(filteredFaqs).length === 0 ? (
            <div className="faq-section">
              <div className="p-12 text-center">
                <HelpCircle className="h-12 w-12 text-text-secondary mx-auto mb-4" />
                <h3 className="text-xl font-semibold mb-2 text-white">No results found</h3>
                <p className="text-text-secondary">
                  Try adjusting your search terms or browse all categories.
                </p>
                <button
                  onClick={() => {
                    setSearchQuery('');
                    setSelectedCategory('all');
                  }}
                  className="faq-contact-button secondary mt-4"
                >
                  Clear filters
                </button>
              </div>
            </div>
          ) : (
            <div className="space-y-8">
              {Object.entries(filteredFaqs).map(([sectionId, faqs]) => {
                const category = categories.find(cat => cat.id === sectionId);
                if (!category) return null;

                return (
                  <FAQSection
                    key={sectionId}
                    title={category.label}
                    icon={category.icon}
                    faqs={faqs}
                    openItems={openItems.get(sectionId) || new Set()}
                    onToggle={(index) => toggleItem(sectionId, index)}
                  />
                );
              })}
            </div>
          )}

          {/* Contact Support */}
          <div className="faq-contact-section">
            <MessageSquare className="faq-contact-icon" />
            <h3 className="faq-contact-title">Still need help?</h3>
            <p className="faq-contact-description">
              Can't find the answer you're looking for? Our support team is here to help.
            </p>
            <div className="faq-contact-buttons">
              <Link to="/contact" className="faq-contact-button primary">
                <MessageSquare className="h-4 w-4" />
                Contact Support
              </Link>
              <a href="https://docs.functionfly.com" target="_blank" rel="noopener noreferrer" className="faq-contact-button secondary">
                <ExternalLink className="h-4 w-4" />
                View Documentation
              </a>
            </div>
          </div>

          {/* Back to Home */}
          <div className="text-center">
            <Link to="/" className="inline-block">
              <button className="faq-contact-button secondary">
                <HelpCircle className="h-4 w-4 mr-2" />
                Back to Home
              </button>
            </Link>
          </div>
        </div>
      </div>

      <Footer />
    </div>
  );
}