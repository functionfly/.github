import { PageLayout } from '@/components/layout/PageLayout';
import { useSupportChat } from '@/components/support';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { usePlan } from '@/hooks/usePlan';
import { motion } from 'framer-motion';
import {
  CheckCircle,
  ChevronLeft,
  Clock,
  Headphones,
  Mail,
  MessageCircle,
  Phone,
  X,
} from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

/**
 * Enterprise Support Page
 * Shows support options and ticket management for enterprise customers
 */
export function EnterpriseSupportPage() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();
  const { openChat } = useSupportChat();
  const [showContactModal, setShowContactModal] = useState(false);
  const [contactType, setContactType] = useState<'chat' | 'email' | 'phone' | null>(null);

  // Redirect non-enterprise users
  if (!isEnterprise) {
    return (
      <PageLayout title="Enterprise Support">
        <Card className="border-dashed border-border-default">
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
              <Headphones className="w-8 h-8 text-amber-500" />
            </div>
            <h2 className="text-xl font-semibold text-text-primary mb-2">Enterprise Feature</h2>
            <p className="text-text-secondary mb-6 max-w-md">
              Dedicated enterprise support is available exclusively for Enterprise plan customers.
              Upgrade to access priority support and dedicated account management.
            </p>
            <Button
              onClick={() => navigate('/pricing')}
              className="bg-gradient-to-r from-amber-500 to-yellow-500"
            >
              View Enterprise Plans
            </Button>
          </CardContent>
        </Card>
      </PageLayout>
    );
  }

  const handleContactClick = async (type: 'chat' | 'email' | 'phone') => {
    if (type === 'chat') {
      try {
        await openChat();
      } finally {
        setShowContactModal(false);
        setContactType(null);
      }
      return;
    }

    setContactType(type);
    setShowContactModal(true);
  };

  const handlePrimaryAction = () => {
    if (contactType === 'email') {
      window.location.href =
        'mailto:support@functionfly.com?subject=Enterprise%20Support%20Request';
    } else if (contactType === 'phone') {
      // Phone support is coming soon (no callback/dial-in yet).
      setShowContactModal(false);
    }
    setShowContactModal(false);
  };

  return (
    <PageLayout title="Enterprise Support">
      <Button variant="outline" className="mb-6" onClick={() => navigate(-1)}>
        <ChevronLeft className="w-4 h-4 mr-2" />
        Back
      </Button>
      <p className="text-text-secondary mb-6">
        Get priority support from our dedicated enterprise team
      </p>

      <div className="space-y-6">
        {/* Support Options */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <SupportCard
            title="Live Chat"
            description="Instant messaging with our support team"
            icon={MessageCircle}
            availability="24/7"
            action="Start Chat"
            onAction={() => handleContactClick('chat')}
          />
          <SupportCard
            title="Email Support"
            description="Detailed responses within 1 hour"
            icon={Mail}
            availability="24/7"
            action="Send Email"
            onAction={() => handleContactClick('email')}
          />
          <SupportCard
            title="Phone Support"
            description="Coming soon: direct line to senior engineers"
            icon={Phone}
            availability="Coming soon"
            action="Coming soon"
            onAction={() => handleContactClick('phone')}
          />
        </div>

        {/* Account Manager */}
        <Card className="border-amber-500/20 bg-gradient-to-br from-amber-500/5 to-transparent">
          <CardHeader>
            <CardTitle className="text-text-primary">Your Dedicated Account Manager</CardTitle>
            <CardDescription className="text-text-secondary">
              Personal support for your enterprise needs
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-start gap-4">
              <div className="w-16 h-16 rounded-full bg-amber-500/20 flex items-center justify-center">
                <Headphones className="w-8 h-8 text-amber-500" />
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-text-primary">Enterprise Support Team</h3>
                <p className="text-text-secondary mt-1">
                  Our enterprise support team is available to help you with any questions, from
                  technical implementation to billing inquiries.
                </p>
                <div className="flex flex-wrap gap-3 mt-4">
                  <Button
                    className="bg-gradient-to-r from-amber-500 to-yellow-500"
                    onClick={() => handleContactClick('email')}
                  >
                    <Mail className="w-4 h-4 mr-2" />
                    Contact Support
                  </Button>
                  <Button variant="outline" onClick={() => handleContactClick('phone')}>
                    <Clock className="w-4 h-4 mr-2" />
                    Schedule Call
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* SLA Guarantee */}
        <Card>
          <CardHeader>
            <CardTitle className="text-text-primary">Support SLA</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <SLAItem title="Critical" response="15 minutes" icon={CheckCircle} />
              <SLAItem title="High" response="1 hour" icon={CheckCircle} />
              <SLAItem title="Medium" response="4 hours" icon={CheckCircle} />
              <SLAItem title="Low" response="24 hours" icon={CheckCircle} />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Contact Modal */}
      {showContactModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-bg-primary border border-white/10 rounded-2xl p-6 max-w-md w-full mx-4 shadow-2xl"
          >
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-xl font-semibold text-white">
                {contactType === 'chat' && 'Live Chat'}
                {contactType === 'email' && 'Email Support'}
                {contactType === 'phone' && 'Phone Support'}
              </h3>
              <button
                onClick={() => setShowContactModal(false)}
                className="text-text-muted hover:text-white transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {contactType === 'chat' && (
              <div className="space-y-4">
                <p className="text-text-secondary">
                  Live chat is available to all Enterprise customers. Connect with our support team
                  instantly during business hours.
                </p>
                <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
                  <p className="text-sm text-text-secondary">
                    To access live chat, please contact us via email to receive your dedicated
                    support portal credentials.
                  </p>
                </div>
                <Button
                  className="w-full bg-gradient-to-r from-amber-500 to-yellow-500"
                  onClick={() => {
                    setContactType('email');
                  }}
                >
                  <Mail className="w-4 h-4 mr-2" />
                  Email Support Instead
                </Button>
              </div>
            )}

            {contactType === 'email' && (
              <div className="space-y-4">
                <p className="text-text-secondary">
                  Send an email to our enterprise support team. We respond to critical issues within
                  15 minutes, 24/7.
                </p>
                <div className="p-4 bg-bg-secondary border border-border-subtle rounded-lg">
                  <p className="font-mono text-sm text-white">support@functionfly.com</p>
                </div>
                <Button
                  className="w-full bg-gradient-to-r from-amber-500 to-yellow-500"
                  onClick={handlePrimaryAction}
                >
                  <Mail className="w-4 h-4 mr-2" />
                  Open Email Client
                </Button>
              </div>
            )}

            {contactType === 'phone' && (
              <div className="space-y-4">
                <div className="rounded-2xl border border-amber-500/25 bg-amber-500/10 px-4 py-3">
                  <div className="flex items-center gap-2">
                    <Phone className="w-4 h-4 text-amber-500" />
                    <span className="text-sm font-semibold text-amber-700 dark:text-amber-300">
                      Coming soon
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-text-secondary">
                    Phone support will be available in a future release. For now, Live Chat and
                    Email Support will route your request to the right engineers.
                  </p>
                </div>
                <p className="text-text-secondary">
                  Phone support is coming soon. In the meantime, please use Live Chat or Email
                  Support so we can route your request to the right team.
                </p>
                <div className="p-4 bg-bg-secondary border border-border-subtle rounded-lg">
                  <p className="font-mono text-lg text-white">Coming soon</p>
                  <p className="text-sm text-text-muted mt-1">
                    We will add a callback / dial-in option shortly.
                  </p>
                </div>
                <Button className="w-full bg-gradient-to-r from-amber-500 to-yellow-500" disabled>
                  <Phone className="w-4 h-4 mr-2" />
                  Coming soon
                </Button>
              </div>
            )}
          </motion.div>
        </div>
      )}
    </PageLayout>
  );
}

interface SupportCardProps {
  title: string;
  description: string;
  icon: typeof MessageCircle;
  availability: string;
  action: string;
  onAction: () => void;
}

function SupportCard({
  title,
  description,
  icon: Icon,
  availability,
  action,
  onAction,
}: SupportCardProps) {
  const isComingSoon = action === 'Coming soon' || availability === 'Coming soon';

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <Card className="h-full">
        <CardContent className="p-6">
          <div className="w-12 h-12 rounded-lg bg-amber-500/10 flex items-center justify-center mb-4">
            <Icon className="w-6 h-6 text-amber-500" />
          </div>
          <h3 className="text-lg font-semibold text-text-primary mb-1">{title}</h3>
          <p className="text-sm text-text-secondary mb-4">{description}</p>
          <div className="flex items-center gap-2 mb-4">
            <Clock className="w-4 h-4 text-text-muted" />
            <span className="text-xs text-text-muted">{availability}</span>
          </div>
          {isComingSoon && (
            <div className="mb-4 rounded-xl border border-amber-500/25 bg-amber-500/10 px-4 py-3">
              <div className="flex items-center gap-2">
                <Clock className="w-4 h-4 text-amber-500" />
                <span className="text-xs font-semibold text-amber-600 dark:text-amber-400">
                  Coming soon
                </span>
              </div>
              <p className="mt-1 text-xs text-text-secondary">
                We are rolling this out soon. Use Live Chat or Email Support in the meantime.
              </p>
            </div>
          )}
          <Button variant="outline" className="w-full" onClick={onAction}>
            {action}
          </Button>
        </CardContent>
      </Card>
    </motion.div>
  );
}

interface SLAItemProps {
  title: string;
  response: string;
  icon: typeof CheckCircle;
}

function SLAItem({ title, response, icon: Icon }: SLAItemProps) {
  return (
    <div className="p-4 rounded-lg bg-bg-secondary border border-border-subtle">
      <div className="flex items-center gap-2 mb-2">
        <Icon className="w-4 h-4 text-green-600 dark:text-green-400" />
        <span className="text-sm font-medium text-text-primary">{title}</span>
      </div>
      <p className="text-lg font-semibold text-amber-600 dark:text-amber-400">{response}</p>
      <p className="text-xs text-text-muted">Response time</p>
    </div>
  );
}
