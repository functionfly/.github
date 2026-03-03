import { Headphones, MessageCircle, Mail, Phone, Clock, CheckCircle } from 'lucide-react';
import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { PageLayout } from '@/components/layout/PageLayout';
import { usePlan } from '@/hooks/usePlan';
import { useNavigate } from 'react-router-dom';

/**
 * Enterprise Support Page
 * Shows support options and ticket management for enterprise customers
 */
export function EnterpriseSupportPage() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();

  // Redirect non-enterprise users
  if (!isEnterprise) {
    return (
      <PageLayout title="Enterprise Support">
        <Card className="border-dashed border-border-default">
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
              <Headphones className="w-8 h-8 text-amber-500" />
            </div>
            <h2 className="text-xl font-semibold text-text-primary mb-2">
              Enterprise Feature
            </h2>
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

  return (
    <PageLayout title="Enterprise Support">
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
          />
          <SupportCard
            title="Email Support"
            description="Detailed responses within 1 hour"
            icon={Mail}
            availability="24/7"
            action="Send Email"
          />
          <SupportCard
            title="Phone Support"
            description="Direct line to senior engineers"
            icon={Phone}
            availability="Business hours"
            action="Call Now"
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
                  Our enterprise support team is available to help you with any questions,
                  from technical implementation to billing inquiries.
                </p>
                <div className="flex flex-wrap gap-3 mt-4">
                  <Button className="bg-gradient-to-r from-amber-500 to-yellow-500">
                    <Mail className="w-4 h-4 mr-2" />
                    Contact Support
                  </Button>
                  <Button variant="outline">
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
              <SLAItem
                title="Critical"
                response="15 minutes"
                icon={CheckCircle}
              />
              <SLAItem
                title="High"
                response="1 hour"
                icon={CheckCircle}
              />
              <SLAItem
                title="Medium"
                response="4 hours"
                icon={CheckCircle}
              />
              <SLAItem
                title="Low"
                response="24 hours"
                icon={CheckCircle}
              />
            </div>
          </CardContent>
        </Card>
      </div>
    </PageLayout>
  );
}

interface SupportCardProps {
  title: string;
  description: string;
  icon: typeof MessageCircle;
  availability: string;
  action: string;
}

function SupportCard({ title, description, icon: Icon, availability, action }: SupportCardProps) {
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
          <Button variant="outline" className="w-full">
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
