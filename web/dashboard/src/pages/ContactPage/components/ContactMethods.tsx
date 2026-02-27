import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Clock, Users, HeadphonesIcon } from 'lucide-react';

interface ContactMethod {
  icon: any;
  title: string;
  description: string;
  contact: string;
  responseTime: string;
  availability: string;
}

const contactMethods: ContactMethod[] = [
  {
    icon: ({ className }: { className?: string }) => <div className={className}>📧</div>,
    title: 'Email Support',
    description: 'Send us an email for detailed inquiries',
    contact: 'support@functionfly.com',
    responseTime: 'Within 24 hours',
    availability: '24/7',
  },
  {
    icon: ({ className }: { className?: string }) => <div className={className}>💬</div>,
    title: 'Live Chat',
    description: 'Chat with our support team in real-time',
    contact: 'Available on dashboard',
    responseTime: 'Instant (during business hours)',
    availability: 'Mon-Fri 9AM-6PM EST',
  },
];

export function ContactMethods() {
  return (
    <Card className="glass-card glow animate-float">
      <CardHeader>
        <CardTitle className="flex items-center gap-3 text-gradient text-xl">
          <HeadphonesIcon className="h-6 w-6 animate-pulse-glow" />
          Contact Methods
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {contactMethods.map((method, index) => (
          <div key={index} className={`glass-light border-border-subtle rounded-lg p-5 hover:glow-sm transition-all duration-300 hover:border-brand-500/50 animate-fade-in`} style={{animationDelay: `${index * 100}ms`}}>
            <div className="flex items-start gap-4">
              <method.icon className="h-6 w-6 text-brand-500 mt-1 animate-float" />
              <div className="flex-1">
                <h4 className="font-semibold mb-2 text-text-primary">{method.title}</h4>
                <p className="text-sm text-text-secondary mb-3 leading-relaxed">{method.description}</p>
                <div className="space-y-2 text-sm">
                  <div className="font-medium text-text-primary bg-brand-500/10 px-3 py-1 rounded-md inline-block">{method.contact}</div>
                  <div className="flex items-center gap-2 text-text-muted">
                    <Clock className="h-3 w-3 animate-spin-slow" />
                    <span>{method.responseTime}</span>
                  </div>
                  <div className="flex items-center gap-2 text-text-muted">
                    <Users className="h-3 w-3" />
                    <span>{method.availability}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}