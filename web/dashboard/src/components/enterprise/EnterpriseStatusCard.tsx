import { Crown, TrendingUp, Shield } from 'lucide-react';
import { motion } from 'framer-motion';
import { useQuery } from '@tanstack/react-query';
import { usePlan } from '@/hooks/usePlan';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useNavigate } from 'react-router-dom';
import { enterpriseSlaApi } from '@/api/enterprise';

/**
 * Enterprise status card for the dashboard
 * Shows SLA status (from API), security compliance, and quick actions
 */
export function EnterpriseStatusCard() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();
  const { data: slaOverview } = useQuery({
    queryKey: ['enterprise', 'sla', 'overview'],
    queryFn: () => enterpriseSlaApi.getOverview(30),
    enabled: !!isEnterprise,
    staleTime: 60_000,
  });

  if (!isEnterprise) return null;

  const slaValue =
    slaOverview != null
      ? `${slaOverview.current_uptime_percent.toFixed(2)}%`
      : '—';
  const slaSubtext = slaOverview != null ? `Last ${slaOverview.period_days} days` : 'Last 30 days';

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <Card className="border-amber-500/20 overflow-hidden relative">
        {/* Background gradient */}
        <div className="absolute inset-0 bg-gradient-to-br from-amber-500/5 via-transparent to-yellow-500/5" />

        <CardHeader className="relative">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-amber-500 to-yellow-500
                              flex items-center justify-center">
                <Crown className="w-4 h-4 text-white" />
              </div>
              <CardTitle className="text-white text-base">Enterprise Status</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full
                                 rounded-full bg-green-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500" />
              </span>
              <span className="text-xs text-green-400 font-medium">Active</span>
            </div>
          </div>
        </CardHeader>

        <CardContent className="relative space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <StatusItem
              icon={TrendingUp}
              label="SLA Status"
              value={slaValue}
              subtext={slaSubtext}
            />
            <StatusItem
              icon={Shield}
              label="Security"
              value="Compliant"
              subtext="SOC2 Type II"
            />
          </div>

          <div className="flex gap-2 pt-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => navigate('/enterprise/sla')}
              className="flex-1 border-amber-500/30 hover:bg-amber-500/10"
            >
              View SLA
            </Button>
            <Button
              size="sm"
              onClick={() => navigate('/enterprise/support')}
              className="flex-1 bg-gradient-to-r from-amber-500 to-yellow-500
                         hover:from-amber-600 hover:to-yellow-600"
            >
              Get Support
            </Button>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

function StatusItem({
  icon: Icon,
  label,
  value,
  subtext,
}: {
  icon: typeof Crown;
  label: string;
  value: string;
  subtext: string;
}) {
  return (
    <div className="flex items-start gap-3">
      <div className="w-8 h-8 rounded-lg bg-amber-500/10 flex items-center justify-center flex-shrink-0">
        <Icon className="w-4 h-4 text-amber-400" />
      </div>
      <div>
        <p className="text-sm text-text-secondary">{label}</p>
        <p className="text-white font-medium">{value}</p>
        <p className="text-xs text-text-muted">{subtext}</p>
      </div>
    </div>
  );
}
