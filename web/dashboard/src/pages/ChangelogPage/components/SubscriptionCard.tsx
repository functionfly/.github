import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Bell, ArrowRight, Shield, AlertCircle, Check } from 'lucide-react';
import { useState, FormEvent } from 'react';
import { useNewsletter } from '@/hooks/useNewsletter';
import { isValidEmail } from '@/lib/url-utils';

const SubscriptionCard = () => {
  const { subscribe, isLoading, isSuccess, error, reset } = useNewsletter();
  const [email, setEmail] = useState('');

  const handleSubscribe = async (e: FormEvent) => {
    e.preventDefault();
    if (!email.trim() || isLoading) return;

    if (!isValidEmail(email.trim())) {
      return;
    }

    const success = await subscribe({ email: email.trim() });
    if (success) {
      setEmail('');
      setTimeout(() => reset(), 5000);
    }
  };

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
        <form onSubmit={handleSubscribe} className="flex flex-col sm:flex-row gap-4">
          <Input
            type="email"
            placeholder="your.email@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={isLoading || isSuccess}
            className="flex-1 glass-card border-border-subtle/50 hover:border-brand-500/50 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 transition-all duration-300 text-(--input-foreground)"
          />
          <Button
            type="submit"
            disabled={isLoading || isSuccess || !email.trim() || !isValidEmail(email.trim())}
            className="btn-primary glow-lg hover:scale-105 transition-all duration-300 whitespace-nowrap"
          >
            {isLoading ? (
              <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : isSuccess ? (
              <Check className="h-4 w-4" />
            ) : (
              <>
                <span className="hidden sm:inline">Subscribe to Updates</span>
                <span className="sm:hidden">Subscribe</span>
                <ArrowRight className="h-4 w-4 ml-2 animate-bounce-subtle" />
              </>
            )}
          </Button>
        </form>
        {error && (
          <div className="flex items-center justify-center gap-2 text-error text-sm">
            <AlertCircle className="w-4 h-4" />
            <span>{error}</span>
          </div>
        )}
        {isSuccess && (
          <div className="flex items-center justify-center gap-2 text-success text-sm">
            <Check className="w-4 h-4" />
            <span>Successfully subscribed! Check your email for confirmation.</span>
          </div>
        )}
        <div className="flex items-center justify-center gap-2 text-sm text-text-muted">
          <Shield className="h-4 w-4 text-success animate-pulse-glow" />
          <span>We respect your privacy. Unsubscribe at any time.</span>
        </div>
      </CardContent>
    </Card>
  );
};

export default SubscriptionCard;