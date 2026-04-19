import { useState, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { VulnerabilityCard } from './VulnerabilityCard';
import { ScanSummaryCard, EmptyScanState, ScanErrorState } from './ScanSummary';
import { SeverityCounts } from './SeverityBadge';
import { 
  type SecurityScan, 
  type Vulnerability, 
  type SeverityLevel, 
  type VulnerabilityStatus,
  severityConfig,
  sortVulnerabilities,
  groupVulnerabilitiesBySeverity
} from './types';
import { 
  Search, 
  Filter, 
  RefreshCw, 
  Download, 
  Shield,
  Container,
  FileCode,
  Server,
  Lock,
  CheckCircle,
  XCircle,
  AlertTriangle
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface ContainerScanResultsProps {
  scan: SecurityScan | null;
  isLoading?: boolean;
  error?: string | null;
  onRefresh?: () => void;
  onExport?: () => void;
  onStatusChange?: (vulnId: string, status: VulnerabilityStatus) => void;
  className?: string;
}

export function ContainerScanResults({
  scan,
  isLoading,
  error,
  onRefresh,
  onExport,
  onStatusChange,
  className,
}: ContainerScanResultsProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [severityFilter, setSeverityFilter] = useState<SeverityLevel | 'all'>('all');
  const [statusFilter, setStatusFilter] = useState<VulnerabilityStatus | 'all'>('all');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');

  // Get unique categories
  const categories = useMemo(() => {
    if (!scan) return [];
    return Array.from(new Set(scan.vulnerabilities.map(v => v.category)));
  }, [scan]);

  // Filter vulnerabilities
  const filteredVulnerabilities = useMemo(() => {
    if (!scan) return [];
    
    return scan.vulnerabilities.filter(vuln => {
      const matchesSearch = 
        !searchQuery ||
        vuln.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        vuln.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        vuln.cve?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        vuln.location?.toLowerCase().includes(searchQuery.toLowerCase());
      
      const matchesSeverity = severityFilter === 'all' || vuln.severity === severityFilter;
      const matchesStatus = statusFilter === 'all' || vuln.status === statusFilter;
      const matchesCategory = categoryFilter === 'all' || vuln.category === categoryFilter;
      
      return matchesSearch && matchesSeverity && matchesStatus && matchesCategory;
    });
  }, [scan, searchQuery, severityFilter, statusFilter, categoryFilter]);

  // Sort by severity
  const sortedVulnerabilities = useMemo(() => {
    return sortVulnerabilities(filteredVulnerabilities);
  }, [filteredVulnerabilities]);

  // Group by severity for tabs
  const groupedVulns = useMemo(() => {
    return groupVulnerabilitiesBySeverity(sortedVulnerabilities);
  }, [sortedVulnerabilities]);

  if (isLoading) {
    return <ScanLoadingState className={className} />;
  }

  if (error) {
    return (
      <ScanErrorState 
        error={error} 
        onRetry={onRefresh}
        className={className}
      />
    );
  }

  if (!scan || scan.vulnerabilities.length === 0) {
    return (
      <EmptyScanState 
        message={scan?.vulnerabilities.length === 0 ? 
          'No vulnerabilities found in this scan. Your containers are secure!' :
          'No scan results available. Run a scan to check your container security.'
        }
        className={className}
      />
    );
  }

  const counts = {
    critical: scan.summary.critical_count,
    high: scan.summary.high_count,
    medium: scan.summary.medium_count,
    low: scan.summary.low_count,
    info: scan.summary.info_count,
  };

  return (
    <div className={cn('space-y-6', className)}>
      {/* Summary Card */}
      <ScanSummaryCard
        summary={scan.summary}
        scanType={scan.type}
        target={scan.target}
        startedAt={scan.started_at}
        completedAt={scan.completed_at}
        duration={scan.duration}
        status={scan.status}
      />

      {/* Filters */}
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex items-center gap-3">
              <Shield className="w-5 h-5 text-primary" />
              <CardTitle>Vulnerabilities</CardTitle>
              <Badge variant="secondary" className="ml-2">
                {filteredVulnerabilities.length} of {scan.summary.total_vulnerabilities}
              </Badge>
            </div>
            <div className="flex gap-2">
              {onRefresh && (
                <Button variant="outline" size="sm" onClick={onRefresh} className="gap-2">
                  <RefreshCw className="w-4 h-4" />
                  Refresh
                </Button>
              )}
              {onExport && (
                <Button variant="outline" size="sm" onClick={onExport} className="gap-2">
                  <Download className="w-4 h-4" />
                  Export
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col sm:flex-row gap-4">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="Search vulnerabilities..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>

            {/* Filters */}
            <div className="flex flex-wrap gap-2">
              <Select value={severityFilter} onValueChange={(v) => setSeverityFilter(v as SeverityLevel | 'all')}>
                <SelectTrigger className="w-[140px]">
                  <Filter className="w-4 h-4 mr-2" />
                  <SelectValue placeholder="Severity" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Severities</SelectItem>
                  <SelectItem value="critical">Critical</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="info">Info</SelectItem>
                </SelectContent>
              </Select>

              <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as VulnerabilityStatus | 'all')}>
                <SelectTrigger className="w-[130px]">
                  <CheckCircle className="w-4 h-4 mr-2" />
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="open">Open</SelectItem>
                  <SelectItem value="fixed">Fixed</SelectItem>
                  <SelectItem value="accepted">Accepted</SelectItem>
                  <SelectItem value="false_positive">False Positive</SelectItem>
                </SelectContent>
              </Select>

              {categories.length > 0 && (
                <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                  <SelectTrigger className="w-[140px]">
                    <Container className="w-4 h-4 mr-2" />
                    <SelectValue placeholder="Category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                    {categories.map(cat => (
                      <SelectItem key={cat} value={cat}>{cat}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </div>
          </div>

          {/* Severity Counts */}
          <div className="mt-4 pt-4 border-t">
            <SeverityCounts counts={counts} />
          </div>
        </CardContent>
      </Card>

      {/* Vulnerability Tabs */}
      <Tabs defaultValue="all" className="w-full">
        <TabsList className="flex flex-wrap h-auto gap-1">
          <TabsTrigger value="all" className="gap-2">
            All
            <Badge variant="secondary" className="ml-1">
              {sortedVulnerabilities.length}
            </Badge>
          </TabsTrigger>
          {(['critical', 'high', 'medium', 'low', 'info'] as SeverityLevel[]).map(sev => {
            const count = groupedVulns[sev].length;
            if (count === 0) return null;
            const config = severityConfig[sev];
            const Icon = config.icon;
            return (
              <TabsTrigger key={sev} value={sev} className={cn('gap-2', config.color)}>
                <Icon className="w-4 h-4" />
                {config.label}
                <Badge 
                  variant={sev === 'critical' || sev === 'high' ? 'destructive' : 'secondary'}
                  className="ml-1"
                >
                  {count}
                </Badge>
              </TabsTrigger>
            );
          })}
        </TabsList>

        <TabsContent value="all" className="mt-4 space-y-4">
          {sortedVulnerabilities.map(vuln => (
            <VulnerabilityCard
              key={vuln.id}
              vulnerability={vuln}
              onStatusChange={onStatusChange}
            />
          ))}
        </TabsContent>

        {(['critical', 'high', 'medium', 'low', 'info'] as SeverityLevel[]).map(sev => (
          <TabsContent key={sev} value={sev} className="mt-4 space-y-4">
            {groupedVulns[sev].length === 0 ? (
              <EmptyFilterState severity={sev} />
            ) : (
              groupedVulns[sev].map(vuln => (
                <VulnerabilityCard
                  key={vuln.id}
                  vulnerability={vuln}
                  onStatusChange={onStatusChange}
                />
              ))
            )}
          </TabsContent>
        ))}
      </Tabs>

      {/* No results from filters */}
      {filteredVulnerabilities.length === 0 && scan.vulnerabilities.length > 0 && (
        <Card className="py-12">
          <CardContent className="flex flex-col items-center justify-center text-center">
            <Filter className="w-12 h-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-semibold text-white mb-2">No matching vulnerabilities</h3>
            <p className="text-text-secondary">Try adjusting your filters to see more results</p>
            <Button 
              variant="outline" 
              className="mt-4"
              onClick={() => {
                setSearchQuery('');
                setSeverityFilter('all');
                setStatusFilter('all');
                setCategoryFilter('all');
              }}
            >
              Clear Filters
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function ScanLoadingState({ className }: { className?: string }) {
  return (
    <Card className={cn('py-12', className)}>
      <CardContent className="flex flex-col items-center justify-center text-center">
        <div className="relative w-16 h-16 mb-4">
          <div className="absolute inset-0 rounded-full border-4 border-primary/20" />
          <div className="absolute inset-0 rounded-full border-4 border-primary border-t-transparent animate-spin" />
          <Shield className="absolute inset-0 m-auto w-6 h-6 text-primary" />
        </div>
        <h3 className="text-lg font-semibold text-white mb-2">Scanning Containers...</h3>
        <p className="text-text-secondary">Analyzing Dockerfiles, images, and runtime security</p>
      </CardContent>
    </Card>
  );
}

function EmptyFilterState({ severity }: { severity: SeverityLevel }) {
  const config = severityConfig[severity];
  const Icon = config.icon;
  
  return (
    <Card className="py-12">
      <CardContent className="flex flex-col items-center justify-center text-center">
        <div className={cn('w-16 h-16 rounded-full flex items-center justify-center mb-4', config.bgColor)}>
          <Icon className={cn('w-8 h-8', config.color)} />
        </div>
        <h3 className="text-lg font-semibold text-white mb-2">No {config.label} Vulnerabilities</h3>
        <p className="text-text-secondary">
          Great! No vulnerabilities with {config.label.toLowerCase()} severity were found.
        </p>
      </CardContent>
    </Card>
  );
}
