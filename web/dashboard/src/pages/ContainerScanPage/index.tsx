import { PageLayout } from '@/components/layout/PageLayout';
import { ContainerScanResults } from '@/components/ContainerScan/ContainerScanResults';
import { useLatestContainerScan, useTriggerScan, useUpdateVulnerability, useExportScan } from '@/api/containerScan';
import { type VulnerabilityStatus } from '@/components/ContainerScan/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/components/ui/use-toast';
import { 
  Shield, 
  ScanLine, 
  History, 
  AlertTriangle,
  CheckCircle,
  FileText,
  Download,
  Clock
} from 'lucide-react';
import { useState } from 'react';

/**
 * Container Security Scan Results Page
 * Displays vulnerability findings from container, Dockerfile, and Docker Compose scans
 */
export function ContainerScanPage() {
  const { data: scan, isLoading, error, refetch } = useLatestContainerScan();
  const triggerScan = useTriggerScan();
  const updateVuln = useUpdateVulnerability();
  const exportScan = useExportScan();
  const { toast } = useToast();
  const [isExporting, setIsExporting] = useState(false);

  const handleRefresh = () => {
    refetch();
    toast({
      title: 'Scan Results Refreshed',
      description: 'Latest scan data has been loaded.',
    });
  };

  const handleTriggerScan = async () => {
    try {
      await triggerScan.mutateAsync({});
      toast({
        title: 'Scan Started',
        description: 'Container security scan is now running. Results will appear automatically.',
      });
    } catch (err) {
      toast({
        title: 'Failed to Start Scan',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const handleStatusChange = async (vulnId: string, status: VulnerabilityStatus) => {
    try {
      await updateVuln.mutateAsync({ vulnId, status });
      toast({
        title: 'Status Updated',
        description: `Vulnerability marked as ${status.replace('_', ' ')}`,
      });
    } catch (err) {
      toast({
        title: 'Update Failed',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    }
  };

  const handleExport = async () => {
    if (!scan?.id) return;
    
    try {
      setIsExporting(true);
      const downloadUrl = await exportScan.mutateAsync({ scanId: scan.id, format: 'json' });
      
      // Trigger download
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = `container-scan-${scan.id}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      toast({
        title: 'Export Complete',
        description: 'Scan results have been downloaded.',
      });
    } catch (err) {
      toast({
        title: 'Export Failed',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      });
    } finally {
      setIsExporting(false);
    }
  };

  const hasCritical = scan?.summary.critical_count && scan.summary.critical_count > 0;
  const hasHigh = scan?.summary.high_count && scan.summary.high_count > 0;

  return (
    <PageLayout 
      title="Container Security" 
      subtitle="Scan results for Dockerfiles, images, and runtime security"
    >
      {/* Header Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div className="flex items-center gap-3">
          <div className={`
            p-2 rounded-lg
            ${hasCritical ? 'bg-red-500/10' : hasHigh ? 'bg-orange-500/10' : 'bg-green-500/10'}
          `}>
            <Shield className={`
              w-5 h-5
              ${hasCritical ? 'text-red-500' : hasHigh ? 'text-orange-500' : 'text-green-500'}
            `} />
          </div>
          <div>
            <h1 className="text-xl font-semibold text-white">Container Security Scan</h1>
            <p className="text-sm text-text-secondary">
              {scan?.target || 'Scanning all container configurations'}
            </p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={isLoading}
            className="gap-2"
          >
            <History className="w-4 h-4" />
            Refresh
          </Button>
          
          <Button
            onClick={handleTriggerScan}
            disabled={triggerScan.isPending || scan?.status === 'running'}
            className="gap-2"
          >
            {scan?.status === 'running' ? (
              <>
                <Clock className="w-4 h-4 animate-spin" />
                Running...
              </>
            ) : (
              <>
                <ScanLine className="w-4 h-4" />
                {scan ? 'Rescan' : 'Start Scan'}
              </>
            )}
          </Button>
        </div>
      </div>

      {/* Quick Stats */}
      {scan && (
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-3 mb-6">
          <QuickStat 
            label="Total"
            value={scan.summary.total_vulnerabilities}
            color="text-white"
            bg="bg-muted"
          />
          <QuickStat 
            label="Critical"
            value={scan.summary.critical_count}
            color="text-red-500"
            bg="bg-red-500/10"
            warning={scan.summary.critical_count > 0}
          />
          <QuickStat 
            label="High"
            value={scan.summary.high_count}
            color="text-orange-500"
            bg="bg-orange-500/10"
            warning={scan.summary.high_count > 0}
          />
          <QuickStat 
            label="Medium"
            value={scan.summary.medium_count}
            color="text-yellow-500"
            bg="bg-yellow-500/10"
          />
          <QuickStat 
            label="Low"
            value={scan.summary.low_count}
            color="text-blue-500"
            bg="bg-blue-500/10"
          />
          <QuickStat 
            label="Info"
            value={scan.summary.info_count}
            color="text-green-500"
            bg="bg-green-500/10"
          />
        </div>
      )}

      {/* Scan Results */}
      <ContainerScanResults
        scan={scan}
        isLoading={isLoading}
        error={error instanceof Error ? error.message : error || null}
        onRefresh={handleRefresh}
        onExport={handleExport}
        onStatusChange={handleStatusChange}
      />

      {/* Info Card */}
      {!scan && !isLoading && (
        <Card className="mt-6 border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <div className="w-16 h-16 rounded-full bg-amber-500/10 flex items-center justify-center mb-4">
              <FileText className="w-8 h-8 text-amber-400" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">No Scan Results</h2>
            <p className="text-text-secondary mb-6 max-w-md">
              Container security scanning checks for vulnerabilities in Dockerfiles, 
              Docker Compose configurations, base images, and runtime security issues.
            </p>
            <div className="flex flex-wrap justify-center gap-2">
              <Badge variant="secondary" className="gap-1">
                <ScanLine className="w-3 h-3" />
                Dockerfile Analysis
              </Badge>
              <Badge variant="secondary" className="gap-1">
                <ScanLine className="w-3 h-3" />
                Image Vulnerabilities
              </Badge>
              <Badge variant="secondary" className="gap-1">
                <ScanLine className="w-3 h-3" />
                Runtime Security
              </Badge>
            </div>
          </CardContent>
        </Card>
      )}
    </PageLayout>
  );
}

interface QuickStatProps {
  label: string;
  value: number;
  color: string;
  bg: string;
  warning?: boolean;
}

function QuickStat({ label, value, color, bg, warning }: QuickStatProps) {
  return (
    <div className={`
      ${bg} rounded-lg p-3 flex flex-col items-center justify-center
      ${warning ? 'ring-1 ring-red-500/50' : ''}
    `}>
      <span className={`text-2xl font-bold ${color}`}>
        {value}
      </span>
      <span className="text-xs text-text-muted">{label}</span>
      {warning && (
        <AlertTriangle className="w-3 h-3 text-red-500 mt-1" />
      )}
    </div>
  );
}
