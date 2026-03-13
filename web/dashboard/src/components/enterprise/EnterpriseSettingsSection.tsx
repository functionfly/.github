import { Crown, Shield, Headphones, FileText, TrendingUp } from 'lucide-react';
import { usePlan } from '@/hooks/usePlan';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useNavigate } from 'react-router-dom';

/**
 * Enterprise settings section for the Settings page
 * Shows enterprise plan details, feature highlights, and quick actions
 */
export function EnterpriseSettingsSection() {
  const { isEnterprise, plan } = usePlan();
  const navigate = useNavigate();

  if (!isEnterprise) return null;

  return (
    <Card className="settings-enterprise-section border border-border-default dark:border-amber-500/20 bg-linear-to-br from-amber-500/5 to-transparent">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-linear-to-br from-amber-500 to-yellow-500
                            flex items-center justify-center">
              <Crown className="w-5 h-5 text-white" />
            </div>
            <div>
              <CardTitle className="text-text-primary">Enterprise Plan</CardTitle>
              <p className="text-sm text-text-secondary">
                Active since March 2024
              </p>
            </div>
          </div>
          <Badge className="bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40 dark:border-amber-500/30">
            Active
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Feature Grid */}
        <div className="grid grid-cols-3 gap-4">
          <FeatureCard
            icon={TrendingUp}
            value="99.99%"
            label="SLA Uptime"
          />
          <FeatureCard
            icon={Shield}
            value="∞"
            label="Unlimited"
          />
          <FeatureCard
            icon={Headphones}
            value="24/7"
            label="Support"
          />
        </div>

        {/* Quick Actions */}
        <div className="flex flex-wrap gap-3">
          <Button
            variant="outline"
            onClick={() => navigate('/enterprise/sla')}
            className="border-border-strong hover:bg-amber-500/10 hover:border-amber-500/30 dark:border-amber-500/30"
          >
            <TrendingUp className="w-4 h-4 mr-2" />
            SLA Dashboard
          </Button>
          <Button
            variant="outline"
            onClick={() => navigate('/enterprise/audit')}
            className="border-border-strong hover:bg-amber-500/10 hover:border-amber-500/30 dark:border-amber-500/30"
          >
            <FileText className="w-4 h-4 mr-2" />
            Audit Logs
          </Button>
          <Button
            variant="outline"
            onClick={() => navigate('/enterprise/support')}
            className="border-border-strong hover:bg-amber-500/10 hover:border-amber-500/30 dark:border-amber-500/30"
          >
            <Headphones className="w-4 h-4 mr-2" />
            Contact Support
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function FeatureCard({
  icon: Icon,
  value,
  label,
}: {
  icon: typeof Crown;
  value: string;
  label: string;
}) {
  return (
    <div className="p-4 rounded-lg bg-bg-elevated border border-border-default shadow-sm text-center">
      <Icon className="w-5 h-5 text-amber-500 dark:text-amber-400 mx-auto mb-2" />
      <p className="text-lg font-semibold text-text-primary">{value}</p>
      <p className="text-xs text-text-secondary">{label}</p>
    </div>
  );
}
