import { Badge } from '@/components/ui/badge';
import { useCookieConsent } from '@/hooks/useCookieConsent';

export function CookieDetails() {
  const { categories } = useCookieConsent();

  const cookieCategories = [
    {
      name: 'Necessary',
      description: 'Essential cookies required for the website to function properly. These cannot be disabled.',
      purpose: 'Session management, authentication, security features',
      duration: 'Session / 1 year',
      status: 'Always Active',
      required: true,
    },
    {
      name: 'Analytics',
      description: 'Help us understand how visitors interact with our website by collecting anonymous information.',
      purpose: 'Usage analytics, performance monitoring, error tracking',
      duration: '2 years',
      status: categories.analytics ? 'Active' : 'Inactive',
      required: false,
    },
    {
      name: 'Marketing',
      description: 'Used to deliver personalized advertisements and track campaign effectiveness.',
      purpose: 'Targeted advertising, retargeting, social media integration',
      duration: '1 year',
      status: categories.marketing ? 'Active' : 'Inactive',
      required: false,
    },
    {
      name: 'Functionality',
      description: 'Enable enhanced functionality and personalization features.',
      purpose: 'Language preferences, UI customization, feature preferences',
      duration: '1 year',
      status: categories.functionality ? 'Active' : 'Inactive',
      required: false,
    },
  ];

  return (
    <div className="privacy-card">
      <div className="space-y-6">
        <div>
          <h3 className="text-xl font-bold text-white">Cookie Details</h3>
        </div>
        <div className="space-y-4">
          {cookieCategories.map((category) => (
            <div key={category.name} className="privacy-cookie-category">
              <div className="flex items-center justify-between mb-2">
                <h4 className="font-semibold">{category.name} Cookies</h4>
                <Badge variant={category.status === 'Active' ? 'default' : 'secondary'}>
                  {category.status}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground mb-2">{category.description}</p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm">
                <div>
                  <span className="font-medium">Purpose:</span> {category.purpose}
                </div>
                <div>
                  <span className="font-medium">Duration:</span> {category.duration}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}