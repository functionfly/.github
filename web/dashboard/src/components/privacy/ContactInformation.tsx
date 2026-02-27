import { Mail, Phone } from 'lucide-react';

export function ContactInformation() {
  return (
    <div className="privacy-card">
      <div className="space-y-6">
        <div>
          <h3 className="text-xl font-bold text-white">Contact Us</h3>
        </div>
        <div className="space-y-4">
          <p className="text-muted-foreground">
            If you have any questions about this Privacy Policy or our cookie practices,
            please contact us:
          </p>
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex items-center gap-2">
              <Mail className="h-4 w-4" />
              <span>privacy@functionfly.com</span>
            </div>
            <div className="flex items-center gap-2">
              <Phone className="h-4 w-4" />
              <span>+1 (555) 123-4567</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}