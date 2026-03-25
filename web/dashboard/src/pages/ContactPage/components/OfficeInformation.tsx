import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { MapPin } from 'lucide-react';

export function OfficeInformation() {
  return (
    <Card className="glass-card glow animate-float-delayed">
      <CardHeader>
        <CardTitle className="flex items-center gap-3 text-gradient text-xl">
          <MapPin className="h-6 w-6 animate-pulse-glow" />
          Office Information
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="glass-light rounded-lg p-4 border border-border-subtle">
          <h4 className="font-semibold mb-3 text-text-primary flex items-center gap-2">
            <div className="w-2 h-2 bg-brand-500 rounded-full animate-pulse"></div>
            Headquarters
          </h4>
          <div className="text-sm text-text-secondary space-y-1 leading-relaxed">
            <p className="font-medium text-text-primary">FunctionFly LLC (d/b/a FunctionFly)</p>
            <p className="text-text-muted text-xs">Wyoming limited liability company</p>
            <p className="pt-2 text-text-primary">Fort Worth, Texas, United States</p>
            <p className="text-text-muted text-xs">Principal operations</p>
          </div>
        </div>

        <div className="glass-light rounded-lg p-4 border border-border-subtle">
          <h4 className="font-semibold mb-3 text-text-primary flex items-center gap-2">
            <div className="w-2 h-2 bg-success rounded-full animate-pulse"></div>
            Business Hours
          </h4>
          <div className="text-sm text-text-secondary space-y-2">
            <div className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 bg-brand-500 rounded-full"></div>
              <p>Monday - Friday: 9:00 AM - 6:00 PM CT</p>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 bg-text-muted rounded-full"></div>
              <p>Saturday - Sunday: Closed</p>
            </div>
            <p className="text-xs mt-3 bg-info/10 text-info px-3 py-2 rounded-md border border-info/20">
              🚨 Emergency support available 24/7 for critical issues
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
