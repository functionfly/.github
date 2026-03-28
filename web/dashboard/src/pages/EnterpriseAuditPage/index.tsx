import { PageLayout } from '@/components/layout/PageLayout';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { usePlan } from '@/hooks/usePlan';
import { Download, FileText, Filter, Mail, Search } from 'lucide-react';
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
            <h2 className="text-xl font-semibold text-white mb-2">Enterprise Feature</h2>
            <p className="text-text-secondary mb-6 max-w-md">
              Audit logs are available exclusively for Enterprise plan customers. Upgrade to access
              detailed audit trails and compliance reporting.
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
                <Input placeholder="Search audit logs..." className="pl-10" />
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
                  <tr>
                    <td colSpan={5} className="py-16 text-center">
                      <div className="flex flex-col items-center justify-center">
                        <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
                          <FileText className="w-8 h-8 text-amber-400" />
                        </div>
                        <h3 className="text-lg font-semibold text-white mb-2">
                          Audit Logs Coming Soon
                        </h3>
                        <p className="text-text-secondary max-w-md mb-4">
                          Comprehensive audit logging for compliance and security monitoring is
                          currently under development. This feature will provide:
                        </p>
                        <ul className="text-left text-sm text-text-muted space-y-2 mb-6 max-w-md">
                          <li className="flex items-start gap-2">
                            <span className="text-green-400 mt-0.5">✓</span>
                            <span>Complete activity trail for all workspace actions</span>
                          </li>
                          <li className="flex items-start gap-2">
                            <span className="text-green-400 mt-0.5">✓</span>
                            <span>Exportable logs for compliance reporting</span>
                          </li>
                          <li className="flex items-start gap-2">
                            <span className="text-green-400 mt-0.5">✓</span>
                            <span>Integration with SIEM tools</span>
                          </li>
                          <li className="flex items-start gap-2">
                            <span className="text-green-400 mt-0.5">✓</span>
                            <span>Real-time alerts for security events</span>
                          </li>
                        </ul>
                        <p className="text-sm text-text-muted mb-4">
                          For immediate security needs, please contact our support team.
                        </p>
                        <Button
                          variant="outline"
                          className="gap-2"
                          onClick={() =>
                            (window.location.href =
                              'mailto:support@functionfly.com?subject=Enterprise%20Audit%20Log%20Inquiry')
                          }
                        >
                          <Mail className="w-4 h-4" />
                          Contact Support
                        </Button>
                      </div>
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
