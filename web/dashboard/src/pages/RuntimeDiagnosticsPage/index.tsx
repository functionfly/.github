import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Progress } from '@/components/ui/progress';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  Cpu,
  Database,
  Download,
  ExternalLink,
  FileJson,
  Globe,
  HardDrive,
  History,
  LineChart,
  Loader2,
  MemoryStick,
  Play,
  RefreshCw,
  Server,
  Settings,
  Shield,
  Terminal,
  Upload,
  XCircle,
  Zap,
} from 'lucide-react';

interface DiagnosticResult {
  category: string;
  items: DiagnosticItem[];
}

interface DiagnosticItem {
  name: string;
  status: 'pass' | 'warning' | 'error' | 'info';
  value: string;
  details?: string;
}

interface RuntimeDiagnosticsPageProps {
  functionId?: string;
}

export function RuntimeDiagnosticsPage({ functionId }: RuntimeDiagnosticsPageProps) {
  const [selectedFunction, setSelectedFunction] = useState<string>(functionId || '');
  const [isLoading, setIsLoading] = useState(false);
  const [results, setResults] = useState<DiagnosticResult[]>([]);

  const runDiagnostics = async () => {
    if (!selectedFunction) return;
    
    setIsLoading(true);
    
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    setResults([
      {
        category: 'Runtime Environment',
        items: [
          { name: 'Node.js Version', status: 'pass', value: '20.11.0', details: 'LTS version' },
          { name: 'V8 Engine', status: 'pass', value: '12.0.267.53' },
          { name: 'Platform', status: 'pass', value: 'linux/amd64' },
          { name: 'Architecture', status: 'pass', value: 'x86_64' },
        ],
      },
      {
        category: 'Resource Usage',
        items: [
          { name: 'Memory Limit', status: 'pass', value: '128 MB' },
          { name: 'Memory Used', status: 'info', value: '45 MB', details: '35% of limit' },
          { name: 'CPU Limit', status: 'pass', value: '1 vCPU' },
          { name: 'Timeout', status: 'pass', value: '30s' },
        ],
      },
      {
        category: 'Network & Connectivity',
        items: [
          { name: 'Network Access', status: 'warning', value: 'Enabled', details: 'May incur additional costs' },
          { name: 'DNS Resolution', status: 'pass', value: 'Working' },
          { name: 'External APIs', status: 'pass', value: '3 requests/min' },
        ],
      },
      {
        category: 'Security',
        items: [
          { name: 'Sandbox Mode', status: 'pass', value: 'Enabled' },
          { name: 'Module Restrictions', status: 'pass', value: 'Active' },
          { name: 'Blocked Modules', status: 'info', value: '12', details: 'child_process, fs, net, etc.' },
          { name: 'Environment Variables', status: 'pass', value: '5 secrets' },
        ],
      },
      {
        category: 'Performance',
        items: [
          { name: 'Cold Start (avg)', status: 'warning', value: '450ms', details: 'Above 300ms threshold' },
          { name: 'Warm Execution (avg)', status: 'pass', value: '12ms' },
          { name: 'Code Cache', status: 'pass', value: 'Enabled' },
          { name: 'Concurrent Limit', status: 'pass', value: '10' },
        ],
      },
    ]);
    
    setIsLoading(false);
  };

  const getStatusIcon = (status: DiagnosticItem['status']) => {
    switch (status) {
      case 'pass':
        return <CheckCircle2 className="h-4 w-4 text-green-500" />;
      case 'warning':
        return <AlertCircle className="h-4 w-4 text-yellow-500" />;
      case 'error':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'info':
        return <Badge variant="outline" className="text-xs">INFO</Badge>;
    }
  };

  const getStatusColor = (status: DiagnosticItem['status']) => {
    switch (status) {
      case 'pass':
        return 'border-green-500/20 bg-green-500/5';
      case 'warning':
        return 'border-yellow-500/20 bg-yellow-500/5';
      case 'error':
        return 'border-red-500/20 bg-red-500/5';
      default:
        return 'border-gray-500/20';
    }
  };

  return (
    <div className="container mx-auto py-8 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Runtime Diagnostics</h1>
          <p className="text-muted-foreground mt-1">
            Debug and analyze your function's runtime behavior
          </p>
        </div>
        <Button onClick={runDiagnostics} disabled={isLoading || !selectedFunction}>
          {isLoading ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <RefreshCw className="mr-2 h-4 w-4" />
          )}
          Run Diagnostics
        </Button>
      </div>

      {/* Function Selection */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5" />
            Select Function
          </CardTitle>
          <CardDescription>
            Choose a function to diagnose
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Select value={selectedFunction} onValueChange={setSelectedFunction}>
            <SelectTrigger className="w-full max-w-md">
              <SelectValue placeholder="Select a function" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="func-1">slugify</SelectItem>
              <SelectItem value="func-2">base64-decode</SelectItem>
              <SelectItem value="func-3">json-minify</SelectItem>
              <SelectItem value="func-4">hash-sha256</SelectItem>
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      {results.length === 0 && !isLoading && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <Terminal className="h-16 w-16 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium mb-2">No Diagnostics Run</h3>
            <p className="text-muted-foreground text-center max-w-md">
              Select a function and click "Run Diagnostics" to analyze its runtime behavior,
              resource usage, and performance metrics.
            </p>
          </CardContent>
        </Card>
      )}

      {results.map((category, index) => (
        <Card key={index} className="overflow-hidden">
          <CardHeader className="bg-muted/30">
            <CardTitle className="flex items-center gap-2 text-lg">
              {category.category === 'Runtime Environment' && <Server className="h-5 w-5" />}
              {category.category === 'Resource Usage' && <Cpu className="h-5 w-5" />}
              {category.category === 'Network & Connectivity' && <Globe className="h-5 w-5" />}
              {category.category === 'Security' && <Shield className="h-5 w-5" />}
              {category.category === 'Performance' && <Zap className="h-5 w-5" />}
              {category.category}
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y">
              {category.items.map((item, itemIndex) => (
                <div
                  key={itemIndex}
                  className={`p-4 flex items-center justify-between ${getStatusColor(item.status)}`}
                >
                  <div className="flex items-center gap-3">
                    {getStatusIcon(item.status)}
                    <div>
                      <p className="font-medium">{item.name}</p>
                      {item.details && (
                        <p className="text-sm text-muted-foreground">{item.details}</p>
                      )}
                    </div>
                  </div>
                  <Badge variant="outline">{item.value}</Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}

      {results.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <LineChart className="h-5 w-5" />
              Summary
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="p-4 rounded-lg bg-green-500/10 border border-green-500/20">
                <CheckCircle2 className="h-8 w-8 text-green-500 mb-2" />
                <p className="text-2xl font-bold">
                  {results.reduce((acc, cat) => 
                    acc + cat.items.filter(i => i.status === 'pass').length, 0)}
                </p>
                <p className="text-sm text-muted-foreground">Passed</p>
              </div>
              <div className="p-4 rounded-lg bg-yellow-500/10 border border-yellow-500/20">
                <AlertCircle className="h-8 w-8 text-yellow-500 mb-2" />
                <p className="text-2xl font-bold">
                  {results.reduce((acc, cat) => 
                    acc + cat.items.filter(i => i.status === 'warning').length, 0)}
                </p>
                <p className="text-sm text-muted-foreground">Warnings</p>
              </div>
              <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20">
                <XCircle className="h-8 w-8 text-red-500 mb-2" />
                <p className="text-2xl font-bold">
                  {results.reduce((acc, cat) => 
                    acc + cat.items.filter(i => i.status === 'error').length, 0)}
                </p>
                <p className="text-sm text-muted-foreground">Errors</p>
              </div>
              <div className="p-4 rounded-lg bg-blue-500/10 border border-blue-500/20">
                <Clock className="h-8 w-8 text-blue-500 mb-2" />
                <p className="text-2xl font-bold">
                  {results.reduce((acc, cat) => 
                    acc + cat.items.filter(i => i.status === 'info').length, 0)}
                </p>
                <p className="text-sm text-muted-foreground">Info</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

export default RuntimeDiagnosticsPage;
