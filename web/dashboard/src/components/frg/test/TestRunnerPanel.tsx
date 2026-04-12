/**
 * TestRunner Panel
 * Features: Input JSON, Output preview, Logs
 */

import { useState, useCallback } from 'react';
import { 
  Play, 
  Plus, 
  Save, 
  Trash2, 
  CheckCircle, 
  AlertCircle,
  Clock,
  Terminal,
  Code,
  FileJson,
  Zap,
  Copy,
  MoreHorizontal,
  ChevronDown,
  ChevronRight,
  RefreshCw,
  Settings,
  Download,
  Upload,
  Filter,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';

import { useFRGStore } from '@/stores/frgStore';
import type { TestCase } from '@/types/frg';

// Mock test cases
const mockTestCases: TestCase[] = [
  {
    id: 'test-1',
    name: 'Basic data processing',
    input: { data: [1, 2, 3, 4, 5] },
    expectedOutput: { sum: 15, count: 5 },
    actualOutput: { sum: 15, count: 5 },
    status: 'passed',
    durationMs: 45,
    logs: ['[INFO] Test started', '[DEBUG] Processing 5 items', '[INFO] Test completed'],
  },
  {
    id: 'test-2',
    name: 'Empty array handling',
    input: { data: [] },
    expectedOutput: { sum: 0, count: 0 },
    actualOutput: { sum: 0, count: 0 },
    status: 'passed',
    durationMs: 23,
    logs: ['[INFO] Test started', '[DEBUG] Empty array detected', '[INFO] Test completed'],
  },
  {
    id: 'test-3',
    name: 'Large dataset',
    input: { data: Array.from({ length: 1000 }, (_, i) => i) },
    status: 'pending',
  },
];

export function TestRunnerPanel() {
  const store = useFRGStore();
  const { testCases, activeTestCase, setTestCases, updateTestCase } = store;

  const [activeTab, setActiveTab] = useState('cases');
  const [selectedTestId, setSelectedTestId] = useState<string | null>('test-1');
  const [showNewTestDialog, setShowNewTestDialog] = useState(false);
  const [newTestName, setNewTestName] = useState('');
  const [jsonInput, setJsonInput] = useState(JSON.stringify({ data: [] }, null, 2));
  const [jsonOutput, setJsonOutput] = useState('');
  const [logs, setLogs] = useState<string[]>([]);
  const [isRunning, setIsRunning] = useState(false);
  const [autoValidate, setAutoValidate] = useState(true);

  const selectedTest = mockTestCases.find(t => t.id === selectedTestId);
  const passedCount = mockTestCases.filter(t => t.status === 'passed').length;
  const failedCount = mockTestCases.filter(t => t.status === 'failed').length;

  const handleRunTest = useCallback(() => {
    setIsRunning(true);
    setLogs(['[INFO] Test starting...']);
    
    // Simulate test execution
    setTimeout(() => {
      setLogs(prev => [...prev, '[DEBUG] Processing input...', '[DEBUG] Validating schema...']);
      setTimeout(() => {
        setLogs(prev => [...prev, '[INFO] Test completed successfully', '[INFO] Duration: 42ms']);
        setJsonOutput(JSON.stringify({ result: 'success', processed: true }, null, 2));
        setIsRunning(false);
      }, 800);
    }, 400);
  }, []);

  const handleRunAll = useCallback(() => {
    mockTestCases.forEach(test => {
      updateTestCase(test.id, { status: 'running' });
    });
  }, [updateTestCase]);

  const handleCreateTest = useCallback(() => {
    if (!newTestName.trim()) return;
    
    const newTest: TestCase = {
      id: `test-${Date.now()}`,
      name: newTestName,
      input: {},
      status: 'pending',
    };
    
    // In production: add to store
    setNewTestName('');
    setShowNewTestDialog(false);
    setSelectedTestId(newTest.id);
  }, [newTestName]);

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-[var(--border-subtle)]">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-[var(--bg-tertiary)] flex items-center justify-center">
              <Zap className="w-4 h-4 text-[var(--text-secondary)]" />
            </div>
            <div>
              <h3 className="font-semibold text-sm text-[var(--text-primary)]">Test Runner</h3>
              <p className="text-[10px] text-[var(--text-secondary)]">
                {passedCount} passed, {failedCount} failed, {mockTestCases.length - passedCount - failedCount} pending
              </p>
            </div>
          </div>
          <div className="flex gap-1">
            <Button variant="outline" size="sm" className="h-7 text-xs" onClick={handleRunAll}>
              <Play className="w-3 h-3 mr-1" />
              Run All
            </Button>
            <Button size="sm" className="h-7 text-xs" onClick={() => setShowNewTestDialog(true)}>
              <Plus className="w-3 h-3 mr-1" />
              New Test
            </Button>
          </div>
        </div>

        {/* Stats */}
        <div className="flex gap-2">
          <Badge variant="default" className="text-[10px]">
            <CheckCircle className="w-3 h-3 mr-1" />
            {passedCount} Passed
          </Badge>
          <Badge variant="destructive" className="text-[10px]">
            <AlertCircle className="w-3 h-3 mr-1" />
            {failedCount} Failed
          </Badge>
          <Badge variant="secondary" className="text-[10px]">
            <Clock className="w-3 h-3 mr-1" />
            {mockTestCases.length - passedCount - failedCount} Pending
          </Badge>
        </div>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col min-h-0">
        <TabsList className="w-full rounded-none border-b border-[var(--border-subtle)] bg-transparent p-0 h-9 px-3">
          <TabsTrigger 
            value="cases" 
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            Test Cases
          </TabsTrigger>
          <TabsTrigger 
            value="runner"
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            Run Test
          </TabsTrigger>
          <TabsTrigger 
            value="history"
            className="rounded-none data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-brand-500 text-xs"
          >
            History
          </TabsTrigger>
        </TabsList>

        <ScrollArea className="flex-1">
          <TabsContent value="cases" className="m-0 p-0 divide-y divide-[var(--border-subtle)]">
            {mockTestCases.map((test) => (
              <div 
                key={test.id}
                className={cn(
                  "p-3 hover:bg-[var(--bg-tertiary)] cursor-pointer transition-colors",
                  selectedTestId === test.id && "bg-brand-500/10"
                )}
                onClick={() => setSelectedTestId(test.id)}
              >
                <div className="flex items-start gap-2">
                  <div className={cn(
                    "w-5 h-5 rounded-full flex items-center justify-center shrink-0 mt-0.5",
                    test.status === 'passed' && "bg-green-500/20",
                    test.status === 'failed' && "bg-red-500/20",
                    test.status === 'pending' && "bg-[var(--bg-tertiary)]",
                    test.status === 'running' && "bg-blue-500/20"
                  )}>
                    {test.status === 'passed' ? (
                      <CheckCircle className="w-3 h-3 text-green-500" />
                    ) : test.status === 'failed' ? (
                      <AlertCircle className="w-3 h-3 text-red-500" />
                    ) : test.status === 'running' ? (
                      <RefreshCw className="w-3 h-3 text-blue-500 animate-spin" />
                    ) : (
                      <Clock className="w-3 h-3 text-[var(--text-muted)]" />
                    )}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-[var(--text-primary)] truncate">
                        {test.name}
                      </span>
                    </div>
                    
                    {test.durationMs && (
                      <div className="flex items-center gap-2 mt-1 text-[10px] text-[var(--text-muted)]">
                        <Clock className="w-3 h-3" />
                        {test.durationMs}ms
                      </div>
                    )}
                  </div>

                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button variant="ghost" size="icon" className="h-6 w-6">
                        <MoreHorizontal className="w-3 h-3" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem>
                        <Play className="w-4 h-4 mr-2" />
                        Run
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Copy className="w-4 h-4 mr-2" />
                        Duplicate
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-red-500">
                        <Trash2 className="w-4 h-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            ))}
          </TabsContent>

          <TabsContent value="runner" className="m-0 p-3 space-y-3">
            {/* Selected Test Header */}
            {selectedTest && (
              <div className="flex items-center justify-between">
                <div>
                  <span className="text-xs text-[var(--text-muted)]">Running test:</span>
                  <p className="text-sm font-medium text-[var(--text-primary)]">{selectedTest.name}</p>
                </div>
                <Button 
                  size="sm" 
                  className="bg-green-500 hover:bg-green-600"
                  disabled={isRunning}
                  onClick={handleRunTest}
                >
                  {isRunning ? (
                    <>
                      <RefreshCw className="w-4 h-4 mr-1 animate-spin" />
                      Running...
                    </>
                  ) : (
                    <>
                      <Play className="w-4 h-4 mr-1" />
                      Run Test
                    </>
                  )}
                </Button>
              </div>
            )}

            <Separator />

            {/* Input */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label className="text-xs flex items-center gap-1">
                  <FileJson className="w-3 h-3" />
                  Input JSON
                </Label>
                <div className="flex items-center gap-2">
                  <Switch 
                    checked={autoValidate}
                    onCheckedChange={setAutoValidate}
                    id="auto-validate"
                  />
                  <Label htmlFor="auto-validate" className="text-[10px] cursor-pointer">
                    Auto-validate
                  </Label>
                </div>
              </div>
              <Textarea
                value={jsonInput}
                onChange={(e) => setJsonInput(e.target.value)}
                className="min-h-[120px] font-mono text-xs"
                placeholder="Enter JSON input..."
              />
            </div>

            {/* Output */}
            <div className="space-y-2">
              <Label className="text-xs flex items-center gap-1">
                <Code className="w-3 h-3" />
                Output
              </Label>
              <div className="min-h-[100px] bg-[var(--bg-tertiary)] rounded-lg p-3 font-mono text-xs overflow-auto">
                {jsonOutput ? (
                  <pre className="text-[var(--text-secondary)]">{jsonOutput}</pre>
                ) : (
                  <span className="text-[var(--text-muted)] italic">Output will appear here after running the test</span>
                )}
              </div>
            </div>

            {/* Logs */}
            <Accordion type="single" collapsible defaultValue="logs">
              <AccordionItem value="logs" className="border-0">
                <AccordionTrigger className="text-xs py-2 hover:no-underline">
                  <div className="flex items-center gap-1">
                    <Terminal className="w-3 h-3" />
                    Execution Logs ({logs.length})
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="bg-[var(--bg-tertiary)] rounded-lg p-3 font-mono text-xs max-h-32 overflow-auto">
                    {logs.length === 0 ? (
                      <span className="text-[var(--text-muted)] italic">No logs yet</span>
                    ) : (
                      logs.map((log, i) => (
                        <div key={i} className="text-[var(--text-secondary)]">
                          {log.includes('[INFO]') ? (
                            <span className="text-green-500">{log.split(']')[0]}]</span>
                          ) : log.includes('[DEBUG]') ? (
                            <span className="text-blue-500">{log.split(']')[0]}]</span>
                          ) : (
                            <span className="text-yellow-500">{log.split(']')[0]}]</span>
                          )}
                          <span className="ml-1">{log.split(']').slice(1).join(']')}</span>
                        </div>
                      ))
                    )}
                  </div>
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          </TabsContent>

          <TabsContent value="history" className="m-0 p-4">
            <div className="text-center py-8">
              <Clock className="w-12 h-12 text-[var(--text-muted)] mx-auto mb-3" />
              <p className="text-sm text-[var(--text-secondary)]">No test history yet</p>
              <p className="text-xs text-[var(--text-muted)] mt-1">
                Run tests to see execution history
              </p>
            </div>
          </TabsContent>
        </ScrollArea>
      </Tabs>

      {/* Footer */}
      <div className="p-3 border-t border-[var(--border-subtle)]">
        <div className="flex gap-2">
          <Button variant="outline" size="sm" className="flex-1">
            <Upload className="w-3 h-3 mr-1" />
            Import
          </Button>
          <Button variant="outline" size="sm" className="flex-1">
            <Download className="w-3 h-3 mr-1" />
            Export
          </Button>
        </div>
      </div>

      {/* New Test Dialog */}
      <Dialog open={showNewTestDialog} onOpenChange={setShowNewTestDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Test Case</DialogTitle>
            <DialogDescription>
              Enter a name for your new test case
            </DialogDescription>
          </DialogHeader>
          <Input
            value={newTestName}
            onChange={(e) => setNewTestName(e.target.value)}
            placeholder="Test case name..."
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleCreateTest();
            }}
          />
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="outline" onClick={() => setShowNewTestDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateTest} disabled={!newTestName.trim()}>
              Create
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default TestRunnerPanel;
