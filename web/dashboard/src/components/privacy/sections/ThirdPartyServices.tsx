export function ThirdPartyServices() {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-2">Third-Party Services</h3>
      <p className="text-muted-foreground mb-3">
        We work with trusted third-party service providers to operate our business and deliver our services. These providers may have access to your personal information only to perform specific tasks on our behalf:
      </p>
      <div className="space-y-3">
        <div>
          <h4 className="font-medium mb-1">Analytics and Performance Monitoring</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>Google Analytics - Website usage analytics and performance monitoring</li>
            <li>Mixpanel - User behavior analytics and product insights</li>
            <li>Sentry - Error tracking and application monitoring</li>
          </ul>
        </div>
        <div>
          <h4 className="font-medium mb-1">Payment Processing</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>Stripe - Secure payment processing and transaction management</li>
            <li>PayPal - Alternative payment processing for user convenience</li>
          </ul>
        </div>
        <div>
          <h4 className="font-medium mb-1">Cloud Infrastructure and Hosting</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>Amazon Web Services (AWS) - Cloud hosting and data storage</li>
            <li>Cloudflare - Content delivery network and security services</li>
            <li>Vercel - Frontend deployment and hosting platform</li>
          </ul>
        </div>
        <div>
          <h4 className="font-medium mb-1">Communication Services</h4>
          <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
            <li>SendGrid/Mailgun - Email delivery and communication services</li>
            <li>Twilio - SMS and communication services (if applicable)</li>
          </ul>
        </div>
      </div>
      <p className="text-sm text-muted-foreground mt-3">
        All third-party providers are contractually obligated to maintain the confidentiality and security of your personal information and may only use it for the specific purposes we authorize.
      </p>
    </div>
  );
}