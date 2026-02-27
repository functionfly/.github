import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Bell, ArrowRight, Shield } from 'lucide-react';

const SubscriptionCard = () => {
  return (
    <Card className="glass-card glow-lg animate-float border-brand-500/30 bg-gradient-to-br from-brand-500/5 via-purple-500/5 to-pink-500/5 hover:from-brand-500/10 hover:via-purple-500/10 hover:to-pink-500/10 transition-all duration-500 group">
      <CardHeader className="text-center pb-6">
        <div className="flex justify-center mb-4">
          <div className="p-4 bg-gradient-to-br from-brand-500/20 to-purple-500/20 rounded-2xl border border-brand-500/30 group-hover:from-brand-500/30 group-hover:to-purple-500/30 transition-all duration-300">
            <Bell className="h-8 w-8 text-brand-500 animate-pulse-glow" />
          </div>
        </div>
        <CardTitle className="text-2xl lg:text-3xl text-gradient group-hover:scale-105 transition-transform duration-300">
          Stay in the Loop
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <p className="text-text-secondary text-lg text-center leading-relaxed">
          Get notified about new releases, exciting features, and important updates directly in your inbox.
          Never miss a beat in our journey of innovation.
        </p>
        <div className="flex flex-col sm:flex-row gap-4">
          <Input
            type="email"
            placeholder="your.email@example.com"
            className="flex-1 glass-card border-border-subtle/50 hover:border-brand-500/50 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 transition-all duration-300 text-(--input-foreground)"
          />
          <Button className="btn-primary glow-lg hover:scale-105 transition-all duration-300 whitespace-nowrap">
            <span className="hidden sm:inline">Subscribe to Updates</span>
            <span className="sm:hidden">Subscribe</span>
            <ArrowRight className="h-4 w-4 ml-2 animate-bounce-subtle" />
          </Button>
        </div>
        <div className="flex items-center justify-center gap-2 text-sm text-text-muted">
          <Shield className="h-4 w-4 text-success animate-pulse-glow" />
          <span>We respect your privacy. Unsubscribe at any time.</span>
        </div>
      </CardContent>
    </Card>
  );
};

export default SubscriptionCard;