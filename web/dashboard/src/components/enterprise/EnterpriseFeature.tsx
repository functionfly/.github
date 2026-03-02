import { ReactNode } from 'react';
import { Lock } from 'lucide-react';
import { usePlan, type FeatureKey } from '@/hooks/usePlan';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';

interface EnterpriseFeatureProps {
  children: ReactNode;
  feature: FeatureKey;
  fallback?: 'hide' | 'upgrade' | 'blur';
  upgradeMessage?: string;
}

/**
 * Feature gate component that conditionally renders content based on plan
 * Shows upgrade prompt or hides content for users without access
 */
export function EnterpriseFeature({
  children,
  feature,
  fallback = 'upgrade',
  upgradeMessage = 'This feature is available on higher plans',
}: EnterpriseFeatureProps) {
  const { hasFeature } = usePlan();
  const hasAccess = hasFeature(feature);

  if (hasAccess) return <>{children}</>;

  if (fallback === 'hide') return null;

  if (fallback === 'blur') {
    return (
      <div className="relative">
        <div className="blur-sm pointer-events-none select-none">
          {children}
        </div>
        <div className="absolute inset-0 flex items-center justify-center">
          <UpgradePrompt message={upgradeMessage} />
        </div>
      </div>
    );
  }

  return <UpgradePrompt message={upgradeMessage} />;
}

function UpgradePrompt({ message }: { message: string }) {
  const navigate = useNavigate();

  return (
    <Card className="border-dashed border-white/20 bg-bg-secondary/50">
      <CardContent className="flex flex-col items-center justify-center py-8 text-center">
        <div className="w-12 h-12 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
          <Lock className="w-6 h-6 text-amber-400" />
        </div>
        <h3 className="font-medium text-white mb-1">Enterprise Feature</h3>
        <p className="text-sm text-text-secondary mb-4 max-w-xs">
          {message}
        </p>
        <Button
          variant="outline"
          onClick={() => navigate('/pricing')}
          className="border-amber-500/30 hover:border-amber-500/50"
        >
          View Plans
        </Button>
      </CardContent>
    </Card>
  );
}
