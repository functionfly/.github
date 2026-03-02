import { TrendingUp, Clock, AlertCircle, CheckCircle } from 'lucide-react';
import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { PageLayout } from '@/components/layout/PageLayout';
import { usePlan } from '@/hooks/usePlan';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';

/**
 * Enterprise SLA Dashboard Page
 * Shows uptime metrics, incident history, and SLA compliance
 */
export function EnterpriseSLAPage() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();

  // Redirect non-enterprise users
  if (!isEnterprise) {
    return (
      <PageLayout title="SLA Dashboard">
        <Card className="border-dashed border-white/20">
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
              <TrendingUp className="w-8 h-8 text-amber-400" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">
              Enterprise Feature
            </h2>
            <p className="text-text-secondary mb-6 max-w-md">
              The SLA Dashboard is available exclusively for Enterprise plan customers.
              Upgrade to access detailed uptime metrics and SLA compliance reports.
            </p>
            <Button
              onClick={() => navigate('/pricing')}
              className="bg-gradient-to-r from-amber-500 to-yellow-500"
            >
              View Enterprise Plans
            </Button>
          </CardContent>
        </Card>
      </PageLayout>
    );
  }

  return (
    <PageLayout title="SLA Dashboard">
      <p className="text-text-secondary mb-6">
        Monitor your service level agreements and uptime metrics
      </p>
      <div className="space-y-6">
        {/* SLA Overview Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <SLACard
            title="Current Uptime"
            value="99.99%"
            subtitle="Last 30 days"
            icon={TrendingUp}
            status="success"
          />
          <SLACard
            title="SLA Target"
            value="99.99%"
            subtitle="Guaranteed uptime"
            icon={CheckCircle}
            status="success"
          />
          <SLACard
            title="Incidents"
            value="0"
            subtitle="Last 30 days"
            icon={AlertCircle}
            status="success"
          />
        </div>

        {/* Uptime Graph Placeholder */}
        <Card>
          <CardHeader>
            <CardTitle className="text-white">Uptime History</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-64 flex items-center justify-center bg-bg-secondary rounded-lg border border-white/8">
              <p className="text-text-secondary">
                Uptime metrics visualization coming soon
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Recent Incidents */}
        <Card>
          <CardHeader>
            <CardTitle className="text-white">Recent Incidents</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="p-4 rounded-lg bg-bg-secondary border border-white/8">
                <div className="flex items-center gap-3">
                  <CheckCircle className="w-5 h-5 text-green-400" />
                  <div>
                    <p className="text-white font-medium">No incidents reported</p>
                    <p className="text-sm text-text-secondary">
                      Your services have been running smoothly
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </PageLayout>
  );
}

interface SLACardProps {
  title: string;
  value: string;
  subtitle: string;
  icon: typeof TrendingUp;
  status: 'success' | 'warning' | 'error';
}

function SLACard({ title, value, subtitle, icon: Icon, status }: SLACardProps) {
  const statusColors = {
    success: 'text-green-400 bg-green-400/10',
    warning: 'text-amber-400 bg-amber-400/10',
    error: 'text-red-400 bg-red-400/10',
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <Card className="border-amber-500/20">
        <CardContent className="p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-text-secondary">{title}</p>
              <p className="text-3xl font-bold text-white mt-1">{value}</p>
              <p className="text-xs text-text-muted mt-1">{subtitle}</p>
            </div>
            <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${statusColors[status]}`}>
              <Icon className="w-5 h-5" />
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
