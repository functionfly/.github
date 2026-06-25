import { motion } from 'framer-motion';
import { Check, Globe, Loader2, AlertCircle, ExternalLink, Sparkles } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useState } from 'react';
import { toast } from 'sonner';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { hasCustomDomainsFeature, getCustomDomainsLimit, type PlanTier } from '@/lib/plan-gating';

export function CustomDomainStep() {
  const {
    customDomain,
    domainVerified,
    setCustomDomain,
    selectedPlan = 'free',
  } = useOnboardingStore();

  const [domain, setDomain] = useState(customDomain || '');
  const [isVerifying, setIsVerifying] = useState(false);
  const [verificationStep, setVerificationStep] = useState<'idle' | 'dns' | 'ssl' | 'complete'>('idle');

  const hasFeature = hasCustomDomainsFeature(selectedPlan as PlanTier);
  const limit = getCustomDomainsLimit(selectedPlan as PlanTier);
  const isUnlimited = limit === Infinity;

  const handleVerify = async () => {
    if (!domain.trim()) {
      toast.error('Please enter a domain name');
      return;
    }

    // Basic domain validation
    const domainRegex = /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;
    if (!domainRegex.test(domain.trim())) {
      toast.error('Please enter a valid domain name (e.g., example.com)');
      return;
    }

    setIsVerifying(true);
    setVerificationStep('dns');

    // Simulate DNS verification
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setVerificationStep('ssl');

    // Simulate SSL provisioning
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setVerificationStep('complete');
    setIsVerifying(false);

    setCustomDomain(domain.trim(), true);
    toast.success('Domain verified and configured successfully!');
  };

  if (!hasFeature) {
    return (
      <div className="space-y-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center space-y-4"
        >
          <div className="w-16 h-16 bg-aviation-amber/20 rounded-full flex items-center justify-center mx-auto">
            <Sparkles className="w-8 h-8 text-aviation-amber" />
          </div>
          <h3 className="text-xl font-mono font-bold text-aviation-text-primary">
            Custom Domains
          </h3>
          <p className="text-aviation-text-secondary font-mono max-w-xl mx-auto">
            Add your own domain to your FunctionFly functions for a branded experience.
          </p>
        </motion.div>

        <Card className="onboarding-step-card p-6 border-aviation-amber/30">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 rounded-lg bg-aviation-amber/20 flex items-center justify-center flex-shrink-0">
              <Globe className="w-6 h-6 text-aviation-amber" />
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-2">
                <h4 className="font-mono font-semibold text-aviation-text-primary">
                  Unlock Custom Domains
                </h4>
                <span className="text-xs font-mono bg-aviation-amber/20 text-aviation-amber px-2 py-0.5 rounded">
                  STARTER+
                </span>
              </div>
              <p className="text-sm text-aviation-text-secondary font-mono mb-4">
                Custom domains are available on Starter ($24/mo), Professional ($79/mo), and Enterprise ($299/mo) plans.
                Your current Free plan doesn't include this feature.
              </p>
              <div className="space-y-2 mb-4">
                <div className="flex items-center gap-2 text-sm text-aviation-text-secondary font-mono">
                  <Check className="w-4 h-4 text-aviation-green" />
                  Starter: 1 custom domain
                </div>
                <div className="flex items-center gap-2 text-sm text-aviation-text-secondary font-mono">
                  <Check className="w-4 h-4 text-aviation-green" />
                  Professional: 5 custom domains
                </div>
                <div className="flex items-center gap-2 text-sm text-aviation-text-secondary font-mono">
                  <Check className="w-4 h-4 text-aviation-green" />
                  Enterprise: Unlimited custom domains
                </div>
              </div>
              <Button
                onClick={() => window.open('/pricing', '_blank')}
                className="font-mono"
              >
                <ExternalLink className="w-4 h-4 mr-2" />
                View Plans & Upgrade
              </Button>
            </div>
          </div>
        </Card>

        <div className="bg-aviation-bg-tertiary rounded-lg p-4">
          <h4 className="font-mono font-medium text-aviation-text-primary mb-2">
            What are Custom Domains?
          </h4>
          <p className="text-sm text-aviation-text-secondary font-mono">
            Custom domains let you serve your FunctionFly functions on your own domain
            (e.g., api.yourcompany.com) instead of the default *.functionfly.app domain.
            This helps with branding, SEO, and providing a more professional experience.
          </p>
        </div>
      </div>
    );
  }

  if (customDomain && domainVerified) {
    return (
      <div className="space-y-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center space-y-4"
        >
          <div className="w-16 h-16 bg-aviation-green/20 rounded-full flex items-center justify-center mx-auto">
            <Check className="w-8 h-8 text-aviation-green" />
          </div>
          <h3 className="text-xl font-mono font-bold text-aviation-text-primary">
            Custom Domain Configured!
          </h3>
          <p className="text-aviation-text-secondary font-mono">
            Your domain <span className="text-aviation-cyan">{customDomain}</span> is ready.
          </p>
        </motion.div>

        <Card className="onboarding-step-card p-4 border-aviation-green/30">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-lg bg-aviation-green/20 flex items-center justify-center">
              <Globe className="w-6 h-6 text-aviation-green" />
            </div>
            <div className="flex-1">
              <h4 className="font-mono font-semibold text-aviation-text-primary">
                {customDomain}
              </h4>
              <p className="text-sm text-aviation-green font-mono flex items-center gap-1">
                <Check className="w-4 h-4" />
                Verified & SSL Active
              </p>
            </div>
          </div>
        </Card>

        <Card className="onboarding-step-card p-4">
          <h4 className="font-mono font-semibold text-aviation-text-primary mb-3">
            DNS Configuration
          </h4>
          <div className="bg-aviation-bg-tertiary rounded-lg p-3 font-mono text-sm space-y-2">
            <div className="flex justify-between">
              <span className="text-aviation-text-muted">Type</span>
              <span className="text-aviation-text-primary">CNAME</span>
            </div>
            <div className="flex justify-between">
              <span className="text-aviation-text-muted">Name</span>
              <span className="text-aviation-text-primary">www</span>
            </div>
            <div className="flex justify-between">
              <span className="text-aviation-text-muted">Value</span>
              <span className="text-aviation-cyan">cname.functionfly.app</span>
            </div>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="onboarding-step-icon w-16 h-16 rounded-2xl flex items-center justify-center mx-auto bg-emerald-500/20">
          <Globe className="w-8 h-8 text-emerald-400" />
        </div>
        <p className="text-lg text-aviation-text-secondary font-mono max-w-xl mx-auto">
          Add a custom domain to serve your functions on your own domain.
        </p>
        {!isUnlimited && (
          <span className="inline-block text-sm font-mono bg-aviation-bg-tertiary text-aviation-text-muted px-3 py-1 rounded">
            Limit: {limit} domain{limit !== 1 ? 's' : ''} on your plan
          </span>
        )}
      </motion.div>

      <Card className="onboarding-step-card p-6">
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="domain" className="flex items-center gap-2 font-mono text-aviation-text-secondary">
              Custom Domain
              <HelpTooltip content="Enter your domain name (e.g., api.example.com or example.com)" />
            </Label>
            <div className="flex gap-2">
              <Input
                id="domain"
                value={domain}
                onChange={(e) => setDomain(e.target.value)}
                placeholder="e.g., api.yourcompany.com"
                className="font-mono flex-1"
                disabled={isVerifying}
              />
              <Button
                onClick={handleVerify}
                disabled={isVerifying || !domain.trim()}
                className="font-mono"
              >
                {isVerifying ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Verifying...
                  </>
                ) : (
                  'Verify'
                )}
              </Button>
            </div>
          </div>

          {isVerifying && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              className="space-y-2"
            >
              <div className="flex items-center gap-2 text-sm">
                {verificationStep === 'dns' && (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin text-aviation-cyan" />
                    <span className="text-aviation-text-secondary font-mono">Verifying DNS configuration...</span>
                  </>
                )}
                {verificationStep === 'ssl' && (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin text-aviation-amber" />
                    <span className="text-aviation-text-secondary font-mono">Provisioning SSL certificate...</span>
                  </>
                )}
              </div>
            </motion.div>
          )}

          {verificationStep === 'complete' && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="flex items-center gap-2 text-aviation-green"
            >
              <Check className="w-4 h-4" />
              <span className="font-mono">Domain verified and SSL configured!</span>
            </motion.div>
          )}
        </div>
      </Card>

      <Card className="onboarding-step-card p-4">
        <h4 className="font-mono font-semibold text-aviation-text-primary mb-3">
          DNS Setup Instructions
        </h4>
        <p className="text-sm text-aviation-text-secondary font-mono mb-4">
          Add the following DNS record to your domain's settings:
        </p>
        <div className="bg-aviation-bg-tertiary rounded-lg p-3 font-mono text-sm space-y-2">
          <div className="flex justify-between">
            <span className="text-aviation-text-muted">Record Type</span>
            <span className="text-aviation-text-primary">CNAME</span>
          </div>
          <div className="flex justify-between">
            <span className="text-aviation-text-muted">Host/Name</span>
            <span className="text-aviation-text-primary">www (or @)</span>
          </div>
          <div className="flex justify-between">
            <span className="text-aviation-text-muted">Value/Target</span>
            <span className="text-aviation-cyan">cname.functionfly.app</span>
          </div>
        </div>
        <div className="flex items-start gap-2 mt-4 text-aviation-amber/80">
          <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
          <p className="text-xs font-mono">
            DNS changes may take up to 48 hours to propagate globally, though it usually happens much faster.
          </p>
        </div>
      </Card>

      <div className="flex justify-center">
        <Button
          variant="outline"
          onClick={() => {/* Skip for now */}}
          className="font-mono"
        >
          Skip for now
        </Button>
      </div>
    </div>
  );
}
