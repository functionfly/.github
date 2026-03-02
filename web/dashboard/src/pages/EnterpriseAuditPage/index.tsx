import { FileText, Search, Filter, Download } from 'lucide-react';
import { motion } from 'framer-motion';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { PageLayout } from '@/components/layout/PageLayout';
import { usePlan } from '@/hooks/usePlan';
import { useNavigate } from 'react-router-dom';

/**
 * Enterprise Audit Logs Page
 * Shows audit trail of all actions in the tenant
 */
export function EnterpriseAuditPage() {
  const { isEnterprise } = usePlan();
  const navigate = useNavigate();

  // Redirect non-enterprise users
  if (!isEnterprise) {
    return (
      <PageLayout title="Audit Logs">
        <Card className="border-dashed border-white/20">
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
              <FileText className="w-8 h-8 text-amber-400" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">
              Enterprise Feature
            </h2>
            <p className="text-text-secondary mb-6 max-w-md">
              Audit logs are available exclusively for Enterprise plan customers.
              Upgrade to access detailed audit trails and compliance reporting.
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
    <PageLayout title="Audit Logs">
      <p className="text-text-secondary mb-6">
        View and export audit trails for compliance and security monitoring
      </p>

      <div className="space-y-6">
        {/* Search and Filter Bar */}
        <Card>
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  placeholder="Search audit logs..."
                  className="pl-10"
                />
              </div>
              <div className="flex gap-2">
                <Button variant="outline" className="gap-2">
                  <Filter className="w-4 h-4" />
                  Filter
                </Button>
                <Button variant="outline" className="gap-2">
                  <Download className="w-4 h-4" />
                  Export
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Audit Log Table */}
        <Card>
          <CardHeader>
            <CardTitle className="text-white">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-white/10">
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Timestamp
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      User
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Action
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Resource
                    </th>
                    <th className="text-left py-3 px-4 text-sm font-medium text-text-secondary">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-white/5">
                    <td className="py-3 px-4 text-sm text-text-secondary">
                      2024-03-01 14:32:00
                    </td>
                    <td className="py-3 px-4 text-sm text-white">
                      admin@example.com
                    </td>
                    <td className="py-3 px-4 text-sm text-white">
                      function.deploy
                    </td>
                    <td className="py-3 px-4 text-sm text-text-secondary">
                      my-function
                    </td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-1 rounded-full text-xs bg-green-500/20 text-green-400">
                        Success
                      </span>
                    </td>
                  </tr>
                  <tr className="border-b border-white/5">
                    <td className="py-3 px-4 text-sm text-text-secondary">
                      2024-03-01 14:30:00
                    </td>
                    <td className="py-3 px-4 text-sm text-white">
                      user@example.com
                    </td>
                    <td className="py-3 px-4 text-sm text-white">
                      api_key.create
                    </td>
                    <td className="py-3 px-4 text-sm text-text-secondary">
                      production-key
                    </td>
                    <td className="py-3 px-4">
                      <span className="px-2 py-1 rounded-full text-xs bg-green-500/20 text-green-400">
                        Success
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td colSpan={5} className="py-8 text-center text-text-secondary">
                      <FileText className="w-8 h-8 mx-auto mb-2 opacity-50" />
                      <p>Audit log integration coming soon</p>
                      <p className="text-sm text-text-muted">
                        Connect your audit log provider to view detailed activity
                      </p>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </div>
    </PageLayout>
  );
}
