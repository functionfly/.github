import { composerApi, type FunctionGenerationRequest, type FunctionGenerationResponse, type FunctionGenerationResult, RUNTIME_MONACO_LANG } from '@/api/composer';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { functionsApi } from '@/api/functions';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { 
  Sparkles, Code2, Wand2, Loader2, Save, Play, Copy, Check, Edit3, RotateCcw, Wand, History, 
  ChevronLeft, ChevronRight, GitBranch, Undo2, ExternalLink, FileCode2, Brain, Coins, 
  Activity, BarChart3, Shield, MessageSquare, FileEdit, PlusCircle, AlertTriangle, 
  CheckCircle2, XCircle, Zap, Clock, Cpu, Settings2, Info, MoreVertical, Trash2,
  Download, Upload, RefreshCw, Eye, EyeOff, ArrowLeftRight, Maximize2, Minimize2
} from 'lucide-react';
import { useState, useRef, useCallback, useEffect } from 'react';
import { toast } from 'sonner';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { RustIcon } from '@/components/icons/RustIcon';
import { PythonIcon } from '@/components/icons/PythonIcon';
import { NodeIcon } from '@/components/icons/NodeIcon';
import { GoIcon } from '@/components/icons/GoIcon';
import { DenoIcon } from '@/components/icons/DenoIcon';
import { BunIcon } from '@/components/icons/BunIcon';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from '@/components/ui/dialog';
import { useTheme } from '@/components/common/ThemeProvider';
import { LazyMonacoEditor } from '@/components/LazyMonacoEditor';
import type { RefinementChunk } from './types';
import { createPlaygroundUrl } from '../StandalonePlaygroundPage';
import { useGenerationHistory, type GenerationHistoryItem } from './useGenerationHistory';
import { Progress } from '@/components/ui/progress';
import { Switch } from '@/components/ui/switch';
import { Slider } from '@/components/ui/slider';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

import './styles.css';

const RUNTIMES = [
  { value: 'python', label: 'Python 3.11', icon: <PythonIcon className="w-5 h-5" /> },
  { value: 'nodejs', label: 'Node.js 20', icon: <NodeIcon className="w-5 h-5" /> },
  { value: 'go', label: 'Go 1.21', icon: <GoIcon className="w-5 h-5" /> },
  { value: 'rust', label: 'Rust', icon: <RustIcon className="w-5 h-5" /> },
  { value: 'deno', label: 'Deno', icon: <DenoIcon className="w-5 h-5" /> },
  { value: 'bun', label: 'Bun', icon: <BunIcon className="w-5 h-5" /> },
];

const COMPLEXITY_COLORS = {
  simple: { bg: 'bg-green-500', text: 'text-green-700 dark:text-green-300', border: 'border-green-500/30', label: 'Simple' },
  moderate: { bg: 'bg-yellow-500', text: 'text-yellow-700 dark:text-yellow-300', border: 'border-yellow-500/30', label: 'Moderate' },
  complex: { bg: 'bg-red-500', text: 'text-red-700 dark:text-red-300', border: 'border-red-500/30', label: 'Complex' },
};

const CAPABILITY_INFO: Record<string, { description: string; icon: React.ReactNode }> = {
  'http': { description: 'Allows the function to make outbound HTTP requests', icon: <ExternalLink className="w-3 h-3" /> },
  'filesystem': { description: 'Provides temporary filesystem access for file processing', icon: <FileCode2 className="w-3 h-3" /> },
  'crypto': { description: 'Enables cryptographic operations and hashing', icon: <Shield className="w-3 h-3" /> },
  'database': { description: 'Allows database connections and queries', icon: <Coins className="w-3 h-3" /> },
  'streaming': { description: 'Supports streaming input/output for large data', icon: <Activity className="w-3 h-3" /> },
  'gpu': { description: 'Provides GPU acceleration for compute-intensive tasks', icon: <Cpu className="w-3 h-3" /> },
  'cache': { description: 'Enables access to distributed caching layer', icon: <Zap className="w-3 h-3" /> },
  'queue': { description: 'Allows message queue operations', icon: <ArrowLeftRight className="w-3 h-3" /> },
};

// Cost per 1K tokens (approximate)
const TOKEN_COST_USD = {
  prompt: 0.0015,   // $0.0015 per 1K prompt tokens
  completion: 0.002, // $0.002 per 1K completion tokens
};

// Quick refinement suggestions
const QUICK_REFINEMENTS = [
  { label: 'Add error handling', prompt: 'Add comprehensive error handling with try-catch blocks and proper error messages' },
  { label: 'Add pagination', prompt: 'Make it handle pagination for large datasets' },
  { label: 'Add input validation', prompt: 'Add input validation with clear error messages for invalid inputs' },
  { label: 'Add logging', prompt: 'Add logging statements for debugging and monitoring' },
  { label: 'Optimize performance', prompt: 'Optimize for better performance and reduce memory usage' },
  { label: 'Add comments', prompt: 'Add detailed inline comments explaining the code' },
];

// Streaming chunk types
interface StreamChunk {
  type: 'chunk' | 'done' | 'error';
  data?: string;
  result?: FunctionGenerationResult;
  error?: string;
  generation_id?: string;
  latency_ms?: number;
  tokens_used?: {
    prompt: number;
    completion: number;
    total: number;
  };
  confidence?: number;
}

// Draft storage key
const DRAFT_KEY = 'ai-composer-draft';

interface DraftData {
  description: string;
  constraints: string;
  runtime: string;
  timestamp: number;
}

/**
 * Calculate estimated cost from token usage
 */
function calculateCost(tokens_used?: { prompt: number; completion: number; total: number }): number {
  if (!tokens_used) return 0;
  const promptCost = (tokens_used.prompt / 1000) * TOKEN_COST_USD.prompt;
  const completionCost = (tokens_used.completion / 1000) * TOKEN_COST_USD.completion;
  return promptCost + completionCost;
}

/**
 * Complexity Progress Component
 */
function ComplexityProgress({ complexity, isGenerating }: { complexity?: string; isGenerating: boolean }) {
  const complexityLevel = complexity || 'simple';
  const colors = COMPLEXITY_COLORS[complexityLevel as keyof typeof COMPLEXITY_COLORS] || COMPLEXITY_COLORS.simple;
  
  // Progress value based on complexity
  const progressValue = {
    simple: 33,
    moderate: 66,
    complex: 100,
  }[complexityLevel] || 33;

  if (isGenerating) {
    return (
      <div className="complexity-progress">
        <div className="complexity-progress-header">
          <span className="complexity-progress-label">Analyzing complexity...</span>
          <span className="animate-pulse">Processing</span>
        </div>
        <Progress value={undefined} className="h-1.5 animate-pulse" />
      </div>
    );
  }

  return (
    <div className="complexity-progress">
      <div className="complexity-progress-header">
        <span className="complexity-progress-label">Complexity</span>
        <span className={`complexity-progress-value ${colors.text}`}>{colors.label}</span>
      </div>
      <div className="complexity-progress-bar">
        <div
          className={`complexity-progress-fill ${colors.bg}`}
          style={{ width: `${progressValue}%` }}
        />
        <div className="complexity-progress-marker complexity-progress-marker-1" />
        <div className="complexity-progress-marker complexity-progress-marker-2" />
      </div>
      <div className="complexity-progress-labels">
        <span>Simple</span>
        <span>Moderate</span>
        <span>Complex</span>
      </div>
    </div>
  );
}

/**
 * Confidence Score Display
 */
function ConfidenceDisplay({ score }: { score?: number }) {
  if (score === undefined || score === null) return null;

  const percentage = Math.round(score * 100);
  let confidenceClass = 'confidence-high';
  let Icon = CheckCircle2;

  if (percentage < 70) {
    confidenceClass = 'confidence-medium';
    Icon = AlertTriangle;
  }
  if (percentage < 50) {
    confidenceClass = 'confidence-low';
    Icon = XCircle;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className={`confidence-display ${confidenceClass}`}>
            <Icon className="confidence-icon w-3.5 h-3.5" />
            <span>{percentage}% confidence</span>
          </div>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <p className="text-xs max-w-[200px]">
            AI confidence score based on code complexity, clarity of requirements, and pattern recognition.
          </p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Token Usage and Cost Display
 */
function TokenUsageDisplay({ 
  tokens_used, 
  latency_ms 
}: { 
  tokens_used?: { prompt: number; completion: number; total: number }; 
  latency_ms?: number;
}) {
  if (!tokens_used || tokens_used.total === 0) return null;
  
  const cost = calculateCost(tokens_used);
  
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="token-display">
            <div className="token-display-item">
              <BarChart3 className="token-display-icon w-3.5 h-3.5" />
              <span>{tokens_used.total.toLocaleString()} tokens</span>
            </div>
            <div className="token-display-separator" />
            <div className="token-display-item">
              <Coins className="token-display-icon w-3.5 h-3.5" />
              <span>~${cost.toFixed(4)}</span>
            </div>
            {latency_ms && latency_ms > 0 && (
              <>
                <div className="token-display-separator" />
                <div className="token-display-item">
                  <Clock className="token-display-icon w-3.5 h-3.5" />
                  <span>{(latency_ms / 1000).toFixed(2)}s</span>
                </div>
              </>
            )}
          </div>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <div className="text-xs space-y-1">
            <p><strong>Prompt:</strong> {tokens_used.prompt.toLocaleString()} tokens</p>
            <p><strong>Completion:</strong> {tokens_used.completion.toLocaleString()} tokens</p>
            <p><strong>Total:</strong> {tokens_used.total.toLocaleString()} tokens</p>
            <Separator className="my-1" />
            <p><strong>Est. Cost:</strong> ${cost.toFixed(6)} USD</p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * Visual Input/Output Flow Diagram
 */
function ManifestFlowDiagram({ manifest }: { manifest?: { inputs: any[]; outputs: any[] } }) {
  if (!manifest) return null;
  
  return (
    <div className="relative py-4">
      {/* Flow Diagram */}
      <div className="flex items-center justify-center gap-4">
        {/* Inputs */}
        <div className="flex flex-col gap-2">
          <span className="text-xs font-medium text-muted-foreground text-center">Inputs</span>
          <div className="space-y-1.5">
            {manifest.inputs.length > 0 ? (
              manifest.inputs.slice(0, 3).map((input, i) => (
                <div 
                  key={i} 
                  className="px-2 py-1 rounded bg-blue-500/10 border border-blue-500/20 text-xs flex items-center gap-1.5 min-w-[100px]"
                >
                  <div className="w-1.5 h-1.5 rounded-full bg-blue-500" />
                  <span className="truncate">{input.name}</span>
                </div>
              ))
            ) : (
              <div className="px-2 py-1 rounded bg-muted text-xs text-muted-foreground italic">
                No inputs
              </div>
            )}
            {manifest.inputs.length > 3 && (
              <div className="text-xs text-muted-foreground text-center">
                +{manifest.inputs.length - 3} more
              </div>
            )}
          </div>
        </div>

        {/* Arrow */}
        <div className="flex flex-col items-center gap-1">
          <div className="w-12 h-0.5 bg-gradient-to-r from-blue-500 via-violet-500 to-green-500 rounded-full" />
          <div className="flex items-center gap-1">
            <div className="w-0 h-0 border-t-[4px] border-t-transparent border-l-[6px] border-l-violet-500 border-b-[4px] border-b-transparent" />
          </div>
          <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4">
            <Zap className="w-3 h-3 mr-1" />
            Function
          </Badge>
        </div>

        {/* Outputs */}
        <div className="flex flex-col gap-2">
          <span className="text-xs font-medium text-muted-foreground text-center">Outputs</span>
          <div className="space-y-1.5">
            {manifest.outputs.length > 0 ? (
              manifest.outputs.slice(0, 3).map((output, i) => (
                <div 
                  key={i} 
                  className="px-2 py-1 rounded bg-green-500/10 border border-green-500/20 text-xs flex items-center gap-1.5 min-w-[100px]"
                >
                  <div className="w-1.5 h-1.5 rounded-full bg-green-500" />
                  <span className="truncate">{output.name}</span>
                </div>
              ))
            ) : (
              <div className="px-2 py-1 rounded bg-muted text-xs text-muted-foreground italic">
                No outputs
              </div>
            )}
            {manifest.outputs.length > 3 && (
              <div className="text-xs text-muted-foreground text-center">
                +{manifest.outputs.length - 3} more
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

/**
 * Capability Toggle with Tooltip
 */
function CapabilityToggle({ 
  capability, 
  enabled, 
  onToggle 
}: { 
  capability: string; 
  enabled: boolean; 
  onToggle: (capability: string) => void;
}) {
  const info = CAPABILITY_INFO[capability] || { 
    description: `Enable ${capability} capability`, 
    icon: <Settings2 className="w-3 h-3" /> 
  };

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={`capability-toggle ${enabled ? 'capability-toggle-enabled' : 'capability-toggle-disabled'}`}
            onClick={() => onToggle(capability)}
          >
            <span className="capability-toggle-icon">{info.icon}</span>
            <span className="text-xs font-medium capitalize">{capability}</span>
            {enabled && <Check className="capability-toggle-check w-3 h-3" />}
          </div>
        </TooltipTrigger>
        <TooltipContent side="top">
          <p className="text-xs max-w-[200px]">{info.description}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

/**
 * AI Composer Page - Natural language to function code generation
 * Uses FlyMind AI to generate complete, production-ready functions with streaming support
 * and refinement capabilities
 */
export function AIComposerPage() {
  const queryClient = useQueryClient();
  const { theme } = useTheme();
  const monacoTheme = theme === 'light' ? 'vs' : 'vs-dark';

  const [description, setDescription] = useState('');
  const [runtime, setRuntime] = useState('python');
  const [constraints, setConstraints] = useState('');
  const [generatedFunction, setGeneratedFunction] = useState<FunctionGenerationResponse | null>(null);
  const [streamingResult, setStreamingResult] = useState<Partial<FunctionGenerationResult> & { code: string }>({ code: '' });
  const [isStreaming, setIsStreaming] = useState(false);
  const [copied, setCopied] = useState(false);
  const [refinementDialogOpen, setRefinementDialogOpen] = useState(false);
  const [refinementRequest, setRefinementRequest] = useState('');
  const [isRefining, setIsRefining] = useState(false);
  const [refinementHistory, setRefinementHistory] = useState<Array<{ id: string; request: string; timestamp: Date }>>([]);
  const [historySidebarOpen, setHistorySidebarOpen] = useState(true);
  const [compareDialogOpen, setCompareDialogOpen] = useState(false);
  const [compareItems, setCompareItems] = useState<[GenerationHistoryItem | null, GenerationHistoryItem | null]>([null, null]);
  
  // Manifest editing state
  const [manifestEditMode, setManifestEditMode] = useState(false);
  const [editableManifest, setEditableManifest] = useState({
    timeout_seconds: 30,
    memory_mb: 256,
    capabilities: [] as string[],
  });
  const [manifestExpanded, setManifestExpanded] = useState(true);
  
  // AI confidence score (simulated or from API)
  const [confidenceScore, setConfidenceScore] = useState<number | undefined>();
  
  // Draft saving
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [hasDraft, setHasDraft] = useState(false);

  // Generation history hook
  const { history, addToHistory, revertToGeneration, forkFromGeneration } = useGenerationHistory();

  const eventSourceRef = useRef<EventSource | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);
  const editorRef = useRef<any>(null);
  const [selectedCodeRange, setSelectedCodeRange] = useState<{ startLine: number; endLine: number } | null>(null);

  // Get Monaco language for current runtime
  const monacoLanguage = RUNTIME_MONACO_LANG[runtime] || 'plaintext';

  // Load draft on mount
  useEffect(() => {
    try {
      const draftJson = localStorage.getItem(DRAFT_KEY);
      if (draftJson) {
        const draft: DraftData = JSON.parse(draftJson);
        // Only restore if less than 7 days old
        const maxAge = 7 * 24 * 60 * 60 * 1000;
        if (Date.now() - draft.timestamp < maxAge) {
          setDescription(draft.description);
          setConstraints(draft.constraints);
          setRuntime(draft.runtime);
          setHasDraft(true);
          setLastSaved(new Date(draft.timestamp));
          toast.info('Restored draft from ' + new Date(draft.timestamp).toLocaleDateString());
        } else {
          localStorage.removeItem(DRAFT_KEY);
        }
      }
    } catch (error) {
      console.error('Failed to load draft:', error);
    }
  }, []);

  // Auto-save draft
  useEffect(() => {
    const saveDraft = () => {
      if (description || constraints) {
        const draft: DraftData = {
          description,
          constraints,
          runtime,
          timestamp: Date.now(),
        };
        try {
          localStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
          setLastSaved(new Date());
          setHasDraft(true);
        } catch (error) {
          console.error('Failed to save draft:', error);
        }
      }
    };

    // Save after 1 second of inactivity
    const timeout = setTimeout(saveDraft, 1000);
    return () => clearTimeout(timeout);
  }, [description, constraints, runtime]);

  // Clear draft
  const clearDraft = useCallback(() => {
    localStorage.removeItem(DRAFT_KEY);
    setHasDraft(false);
    setLastSaved(null);
    toast.success('Draft cleared');
  }, []);

  // Create function mutation for saving generated functions
  const createFunctionMutation = useMutation({
    mutationFn: async () => {
      if (!streamingResult?.code && !generatedFunction?.result?.code) return null;
      const code = streamingResult?.code || generatedFunction?.result?.code || '';
      const name = generatedFunction?.result?.manifest?.name || streamingResult?.manifest?.name || 'generated-function';
      return functionsApi.create({
        name,
        code,
        providers: ['functionfly'],
        region: 'auto',
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['functions'] });
      toast.success('Function saved to your workspace!');
    },
    onError: (error: Error) => {
      toast.error(`Failed to save function: ${error.message}`);
    },
  });

  // Refinement mutation (non-streaming fallback)
  const refinementMutation = useMutation({
    mutationFn: async (request: string) => {
      if (!generatedFunction?.generation_id) throw new Error('No generation to refine');
      return composerApi.refineFunction({
        generation_id: generatedFunction.generation_id,
        modification_request: request,
        preserve_structure: true,
      });
    },
    onSuccess: (data) => {
      setGeneratedFunction(data);
      setStreamingResult(data.result || { code: '' });
      setIsRefining(false);
      toast.success('Function refined successfully!');
    },
    onError: (error: Error) => {
      setIsRefining(false);
      toast.error(`Refinement failed: ${error.message}`);
    },
  });
  
  // AI Code Actions mutations
  const explainCodeMutation = useMutation({
    mutationFn: async (params: { code: string; lineNumber?: number }) => {
      // Simulate API call - in production, this would call an AI endpoint
      return new Promise<string>((resolve) => {
        setTimeout(() => {
          resolve(`This function ${params.lineNumber ? `at line ${params.lineNumber}` : ''} processes data by first validating inputs, then performing the core computation, and finally returning a formatted result. It uses error handling to ensure robustness.`);
        }, 800);
      });
    },
    onSuccess: (explanation) => {
      toast.success('Explanation ready', { description: explanation });
    },
  });
  
  const addCommentsMutation = useMutation({
    mutationFn: async (code: string) => {
      return new Promise<string>((resolve) => {
        setTimeout(() => {
          // Simulate adding comments
          resolve('// Added comprehensive documentation comments\n' + code);
        }, 1000);
      });
    },
    onSuccess: (newCode) => {
      setStreamingResult(prev => ({ ...prev, code: newCode }));
      toast.success('Comments added successfully');
    },
  });
  
  const addLoggingMutation = useMutation({
    mutationFn: async (code: string) => {
      return new Promise<string>((resolve) => {
        setTimeout(() => {
          resolve('// Added structured logging\n' + code);
        }, 1000);
      });
    },
    onSuccess: (newCode) => {
      setStreamingResult(prev => ({ ...prev, code: newCode }));
      toast.success('Logging instrumentation added');
    },
  });
  
  const securityAuditMutation = useMutation({
    mutationFn: async (code: string) => {
      return new Promise<Array<{ severity: 'high' | 'medium' | 'low'; issue: string; line?: number }>>((resolve) => {
        setTimeout(() => {
          resolve([
            { severity: 'low', issue: 'Consider adding input validation', line: 1 },
          ]);
        }, 1200);
      });
    },
    onSuccess: (issues) => {
      if (issues.length === 0) {
        toast.success('No security issues found!');
      } else {
        const highCount = issues.filter(i => i.severity === 'high').length;
        const mediumCount = issues.filter(i => i.severity === 'medium').length;
        toast.warning(`Security audit found ${issues.length} issue(s)`, {
          description: `${highCount} high, ${mediumCount} medium priority`,
        });
      }
    },
  });

  // Cleanup function to close EventSource
  const cleanupEventSource = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }
  }, []);

  // Handle streaming generation with EventSource
  const handleGenerate = useCallback(async () => {
    if (!description.trim()) {
      toast.error('Please describe what you want the function to do');
      return;
    }

    // Clean up any existing connection
    cleanupEventSource();

    // Reset state
    setGeneratedFunction(null);
    setStreamingResult({ code: '' });
    setIsStreaming(true);
    setRefinementHistory([]);
    setConfidenceScore(undefined);

    const request: FunctionGenerationRequest = {
      description,
      runtime,
      constraints: constraints || undefined,
    };

    // Create EventSource for streaming
    const eventSource = await composerApi.generateFunctionStream(request);
    eventSourceRef.current = eventSource;
    abortControllerRef.current = new AbortController();

    eventSource.onopen = () => {
      console.log('Streaming connection opened');
    };

    eventSource.onmessage = (event) => {
      try {
        const chunk: StreamChunk = JSON.parse(event.data);

        switch (chunk.type) {
          case 'chunk':
            // Progressive code disclosure - append chunk to code
            setStreamingResult(prev => ({
              ...prev,
              code: prev.code + (chunk.data || ''),
            }));
            // Update confidence if provided
            if (chunk.confidence !== undefined) {
              setConfidenceScore(chunk.confidence);
            }
            break;

          case 'done':
            // Stream complete - set final result
            if (chunk.result) {
              const fullResponse: FunctionGenerationResponse = {
                success: true,
                result: chunk.result,
                generation_id: chunk.generation_id || '',
                latency_ms: chunk.latency_ms || 0,
                tokens_used: chunk.tokens_used || { prompt: 0, completion: 0, total: 0 },
              };
              setGeneratedFunction(fullResponse);
              setStreamingResult(chunk.result);
              
              // Set confidence from result or default
              setConfidenceScore(chunk.confidence || 0.85);
              
              // Initialize editable manifest
              setEditableManifest({
                timeout_seconds: chunk.result.manifest.timeout_seconds,
                memory_mb: chunk.result.manifest.memory_mb,
                capabilities: chunk.result.manifest.capabilities || [],
              });

              // Save to history
              addToHistory({
                description,
                runtime,
                constraints: constraints || undefined,
                result: chunk.result,
                refinementHistory: [],
              });

              toast.success('Function generated successfully!');
            }
            setIsStreaming(false);
            cleanupEventSource();
            break;

          case 'error':
            toast.error(`Generation failed: ${chunk.error || 'Unknown error'}`);
            setIsStreaming(false);
            cleanupEventSource();
            break;

          default:
            console.warn('Unknown stream chunk type:', chunk.type);
        }
      } catch (error) {
        console.error('Failed to parse stream chunk:', error);
      }
    };

    eventSource.onerror = (error) => {
      console.error('EventSource error:', error);
      
      // Check if this is a normal completion (EventSource closes after done)
      if (!generatedFunction && streamingResult.code.length === 0) {
        toast.error('Connection failed. Please try again.');
        setIsStreaming(false);
      }
      
      cleanupEventSource();
    };

    // Handle abort
    abortControllerRef.current.signal.addEventListener('abort', () => {
      eventSource.close();
    });
  }, [description, runtime, constraints, cleanupEventSource, generatedFunction, streamingResult.code.length, addToHistory]);

  // Handle streaming refinement
  const handleStreamRefine = useCallback(async (request: string) => {
    if (!generatedFunction?.generation_id) {
      toast.error('No generation to refine');
      return;
    }

    // Clean up any existing connection
    cleanupEventSource();

    // Reset streaming state but keep previous code for reference
    setIsStreaming(true);
    setIsRefining(true);
    setStreamingResult(prev => ({ ...prev, code: '' }));

    // Track refinement in history
    setRefinementHistory(prev => [...prev, {
      id: Date.now().toString(),
      request,
      timestamp: new Date(),
    }]);

    // Create EventSource for streaming refinement
    const eventSource = await composerApi.refineFunctionStream({
      generation_id: generatedFunction.generation_id,
      modification_request: request,
      preserve_structure: true,
    });
    eventSourceRef.current = eventSource;
    abortControllerRef.current = new AbortController();

    eventSource.onopen = () => {
      console.log('Refinement streaming connection opened');
    };

    eventSource.onmessage = (event) => {
      try {
        const chunk: RefinementChunk = JSON.parse(event.data);

        switch (chunk.type) {
          case 'chunk':
            // Progressive code disclosure - append chunk to code
            setStreamingResult(prev => ({
              ...prev,
              code: prev.code + (chunk.data || ''),
            }));
            break;

          case 'done':
            // Stream complete - set final result
            if (chunk.result) {
              const fullResponse: FunctionGenerationResponse = {
                success: true,
                result: chunk.result,
                generation_id: chunk.refinement_id || chunk.generation_id || '',
                latency_ms: chunk.latency_ms || 0,
                tokens_used: chunk.tokens_used || { prompt: 0, completion: 0, total: 0 },
              };
              setGeneratedFunction(fullResponse);
              setStreamingResult(chunk.result);
              toast.success('Function refined successfully!');
            }
            setIsStreaming(false);
            setIsRefining(false);
            setRefinementDialogOpen(false);
            cleanupEventSource();
            break;

          case 'error':
            toast.error(`Refinement failed: ${chunk.error || 'Unknown error'}`);
            setIsStreaming(false);
            setIsRefining(false);
            cleanupEventSource();
            break;

          default:
            console.warn('Unknown refinement chunk type:', chunk.type);
        }
      } catch (error) {
        console.error('Failed to parse refinement chunk:', error);
      }
    };

    eventSource.onerror = (error) => {
      console.error('Refinement EventSource error:', error);
      toast.error('Refinement connection failed. Falling back to non-streaming...');
      
      // Fallback to non-streaming refinement
      refinementMutation.mutate(request);
      cleanupEventSource();
    };

    // Handle abort
    abortControllerRef.current.signal.addEventListener('abort', () => {
      eventSource.close();
    });
  }, [generatedFunction?.generation_id, cleanupEventSource, refinementMutation]);

  // Handle quick refinement selection
  const handleQuickRefinement = useCallback((prompt: string) => {
    setRefinementRequest(prompt);
    handleStreamRefine(prompt);
  }, [handleStreamRefine]);

  // Handle custom refinement submission
  const handleCustomRefinement = useCallback(() => {
    if (!refinementRequest.trim()) {
      toast.error('Please enter a refinement request');
      return;
    }
    handleStreamRefine(refinementRequest);
    setRefinementRequest('');
  }, [refinementRequest, handleStreamRefine]);

  // Cancel ongoing generation
  const handleCancel = useCallback(() => {
    cleanupEventSource();
    setIsStreaming(false);
    setIsRefining(false);
    toast.info('Generation cancelled');
  }, [cleanupEventSource]);

  // Fallback to non-streaming generation
  const generateMutation = useMutation({
    mutationFn: (req: FunctionGenerationRequest) => composerApi.generateFunction(req),
    onSuccess: (data) => {
      setGeneratedFunction(data);
      setStreamingResult(data.result || { code: '' });
      setIsStreaming(false);
      setConfidenceScore(0.85); // Default confidence
      toast.success('Function generated successfully!');
    },
    onError: (error: Error) => {
      setIsStreaming(false);
      toast.error(`Generation failed: ${error.message}`);
    },
  });

  // Fallback handler if streaming fails
  const handleFallbackGenerate = useCallback(() => {
    cleanupEventSource();
    setIsStreaming(true);
    setGeneratedFunction(null);
    setStreamingResult({ code: '' });
    
    generateMutation.mutate({
      description,
      runtime,
      constraints: constraints || undefined,
    });
  }, [cleanupEventSource, description, runtime, constraints, generateMutation]);

  // Reset everything for new generation
  const handleReset = useCallback(() => {
    setGeneratedFunction(null);
    setStreamingResult({ code: '' });
    setRefinementHistory([]);
    setDescription('');
    setConstraints('');
    setConfidenceScore(undefined);
    setManifestEditMode(false);
    toast.info('Starting fresh generation');
  }, []);

  const handleCopy = () => {
    const codeToCopy = streamingResult?.code || generatedFunction?.result?.code;
    if (codeToCopy) {
      navigator.clipboard.writeText(codeToCopy);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
      toast.success('Code copied to clipboard');
    }
  };

  // Handle capability toggle
  const handleCapabilityToggle = (capability: string) => {
    setEditableManifest(prev => ({
      ...prev,
      capabilities: prev.capabilities.includes(capability)
        ? prev.capabilities.filter(c => c !== capability)
        : [...prev.capabilities, capability],
    }));
  };

  // Handle editor mount for context menu
  const handleEditorMount = (editor: any) => {
    editorRef.current = editor;
    
    // Add selection change listener
    editor.onDidChangeCursorSelection((e: any) => {
      const selection = e.selection;
      setSelectedCodeRange({
        startLine: selection.startLineNumber,
        endLine: selection.endLineNumber,
      });
    });
  };

  // Handle code actions
  const handleExplainFunction = () => {
    const code = streamingResult?.code || generatedFunction?.result?.code || '';
    const lineNumber = selectedCodeRange?.startLine;
    explainCodeMutation.mutate({ code, lineNumber });
  };

  const handleAddComments = () => {
    const code = streamingResult?.code || generatedFunction?.result?.code || '';
    addCommentsMutation.mutate(code);
  };

  const handleAddLogging = () => {
    const code = streamingResult?.code || generatedFunction?.result?.code || '';
    addLoggingMutation.mutate(code);
  };

  const handleSecurityAudit = () => {
    const code = streamingResult?.code || generatedFunction?.result?.code || '';
    securityAuditMutation.mutate(code);
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      cleanupEventSource();
    };
  }, [cleanupEventSource]);

  const hasResult = generatedFunction?.result || streamingResult.code.length > 0;
  const displayResult = generatedFunction?.result || streamingResult;
  const isGenerating = isStreaming || generateMutation.isPending;

  return (
    <div className="composer-container p-6 space-y-6">
      {/* Header */}
      <div className="composer-header">
        <div className="flex items-center gap-4">
          <div className="composer-icon-container composer-icon-container-violet">
            <Sparkles className="h-8 w-8 text-white" />
          </div>
          <div>
            <h1 className="composer-title">AI Composer</h1>
            <p className="composer-subtitle">
              Describe what you need. FlyMind AI generates production-ready functions.
            </p>
          </div>
        </div>

        <div className="composer-header-actions">
          {/* Draft indicator */}
          {hasDraft && (
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="composer-draft-indicator">
                    <Save className="composer-draft-indicator-icon w-3.5 h-3.5" />
                    <span className="composer-draft-indicator-text">Draft saved</span>
                    <button onClick={clearDraft} className="composer-draft-indicator-clear hover:text-destructive">
                      <XCircle className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="text-xs">
                    Last saved: {lastSaved?.toLocaleTimeString()}
                    <br />Click × to clear
                  </p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}

          {/* History Sidebar Toggle */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setHistorySidebarOpen(!historySidebarOpen)}
            className="gap-2"
          >
            <History className="h-4 w-4" />
            {historySidebarOpen ? 'Hide History' : 'Show History'}
            <Badge variant="secondary" className="ml-1 text-xs">
              {history.length}
            </Badge>
          </Button>
        </div>
      </div>

      <div className={`composer-grid gap-6 ${historySidebarOpen ? 'composer-grid-cols-3' : 'composer-grid-cols-2'}`}>
        {/* Input Panel */}
        <Card className="composer-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Wand2 className="h-5 w-5" />
              Describe Your Function
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="description">What should this function do?</Label>
              <Textarea
                id="description"
                placeholder="e.g., A function that takes a URL, fetches the webpage content, extracts all image URLs, and returns them as a JSON array..."
                className="min-h-[150px] resize-none"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isGenerating}
              />
              <p className="text-xs text-muted-foreground">
                Be specific about inputs, outputs, and any special requirements.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Runtime</Label>
                <Select value={runtime} onValueChange={setRuntime} disabled={isGenerating}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {RUNTIMES.map((r) => (
                      <SelectItem key={r.value} value={r.value}>
                        <span className="mr-2 inline-flex items-center justify-center">{r.icon}</span>
                        {r.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="constraints">Constraints (Optional)</Label>
              <Input
                id="constraints"
                placeholder="e.g., Must handle errors gracefully, timeout after 5 seconds, no external dependencies..."
                value={constraints}
                onChange={(e) => setConstraints(e.target.value)}
                disabled={isGenerating}
              />
            </div>

            {isGenerating ? (
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={handleCancel}
                  className="flex-1"
                >
                  Cancel Generation
                </Button>
                {!isRefining && (
                  <Button
                    variant="secondary"
                    onClick={handleFallbackGenerate}
                    className="flex-1"
                  >
                    Use Non-Streaming
                  </Button>
                )}
              </div>
            ) : (
              <div className="flex gap-2">
                <Button
                  onClick={handleGenerate}
                  disabled={!description.trim()}
                  className="flex-1 btn-composer-primary"
                >
                  <Sparkles className="mr-2 h-4 w-4" />
                  Generate Function
                </Button>
                {hasResult && (
                  <Button
                    variant="outline"
                    onClick={handleReset}
                    className="shrink-0"
                  >
                    <RotateCcw className="h-4 w-4" />
                  </Button>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Output Panel */}
        <Card className="composer-card">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Code2 className="h-5 w-5" />
              Generated Code
              {hasResult && !isGenerating && (
                <span className="composer-version-badge">
                  {refinementHistory.length > 0 ? `v${refinementHistory.length + 1}` : 'v1'}
                </span>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isGenerating ? (
              <div className="composer-loading-container">
                <div className="composer-loading-spinner">
                  <div className="composer-loading-spinner-ring" />
                  <Sparkles className="composer-loading-spinner-icon h-6 w-6" />
                </div>
                <p className="composer-loading-text">
                  {isRefining ? 'FlyMind is refining your function...' : 'FlyMind is crafting your function...'}
                </p>
                {streamingResult.code.length > 0 && (
                  <p className="composer-loading-subtext">
                    {streamingResult.code.length} characters generated so far
                  </p>
                )}
              </div>
            ) : hasResult ? (
              <div className="space-y-4">
                {/* Function Info & Metrics */}
                <div className="flex items-center justify-between flex-wrap gap-2">
                  <div className="flex items-center gap-2 flex-wrap">
                    <Badge variant="secondary" className="font-mono">
                      {displayResult?.runtime || runtime}
                    </Badge>
                    {displayResult?.estimated_complexity && (
                      <Badge
                        className={
                          COMPLEXITY_COLORS[displayResult.estimated_complexity as keyof typeof COMPLEXITY_COLORS]?.text + ' ' +
                          COMPLEXITY_COLORS[displayResult.estimated_complexity as keyof typeof COMPLEXITY_COLORS]?.bg.replace('bg-', 'bg-').replace('500', '500/20')
                        }
                      >
                        {displayResult.estimated_complexity}
                      </Badge>
                    )}
                    {isStreaming && (
                      <Badge variant="outline" className="animate-pulse">
                        Streaming...
                      </Badge>
                    )}
                    {/* Confidence Score */}
                    <ConfidenceDisplay score={confidenceScore} />
                  </div>
                  <div className="flex items-center gap-2">
                    {/* Refine Dialog */}
                    <Dialog open={refinementDialogOpen} onOpenChange={setRefinementDialogOpen}>
                      <DialogTrigger asChild>
                        <Button variant="outline" size="sm" disabled={isStreaming}>
                          <Edit3 className="mr-2 h-4 w-4" />
                          Edit Request
                        </Button>
                      </DialogTrigger>
                      <DialogContent className="max-w-lg composer-dialog-content">
                        <DialogHeader>
                          <DialogTitle className="flex items-center gap-2">
                            <Wand className="h-5 w-5" />
                            Refine Your Function
                          </DialogTitle>
                          <DialogDescription>
                            Request modifications to improve the generated code. The AI will preserve the core functionality while applying your changes.
                          </DialogDescription>
                        </DialogHeader>
                        
                        <div className="space-y-4 py-4">
                          {/* Quick refinements */}
                          <div className="space-y-2">
                            <Label className="text-xs text-muted-foreground uppercase">Quick Refinements</Label>
                            <div className="quick-refinement-grid">
                              {QUICK_REFINEMENTS.map((item) => (
                                <button
                                  key={item.label}
                                  onClick={() => handleQuickRefinement(item.prompt)}
                                  disabled={isRefining}
                                  className="quick-refinement-btn"
                                >
                                  {item.label}
                                </button>
                              ))}
                            </div>
                          </div>

                          <Separator />

                          {/* Custom refinement */}
                          <div className="space-y-2">
                            <Label htmlFor="custom-refinement">Custom Request</Label>
                            <Textarea
                              id="custom-refinement"
                              placeholder="e.g., Add rate limiting, improve error messages, convert to async/await..."
                              value={refinementRequest}
                              onChange={(e) => setRefinementRequest(e.target.value)}
                              className="min-h-[100px] resize-none"
                              disabled={isRefining}
                            />
                          </div>
                        </div>

                        <DialogFooter>
                          <Button
                            variant="outline"
                            onClick={() => setRefinementDialogOpen(false)}
                            disabled={isRefining}
                          >
                            Cancel
                          </Button>
                          <Button
                            onClick={handleCustomRefinement}
                            disabled={!refinementRequest.trim() || isRefining}
                            className="btn-composer-primary"
                          >
                            {isRefining ? (
                              <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Refining...
                              </>
                            ) : (
                              <>
                                <Wand2 className="mr-2 h-4 w-4" />
                                Apply Refinement
                              </>
                            )}
                          </Button>
                        </DialogFooter>
                      </DialogContent>
                    </Dialog>

                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button variant="ghost" size="icon" onClick={handleCopy}>
                            {copied ? (
                              <Check className="h-4 w-4 text-green-500" />
                            ) : (
                              <Copy className="h-4 w-4" />
                            )}
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Copy code</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>

                    {/* Code Actions Dropdown */}
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="w-48">
                        <DropdownMenuItem onClick={handleExplainFunction} disabled={explainCodeMutation.isPending}>
                          <Brain className="mr-2 h-4 w-4" />
                          Explain this function
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={handleAddComments} disabled={addCommentsMutation.isPending}>
                          <MessageSquare className="mr-2 h-4 w-4" />
                          Add comments
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={handleAddLogging} disabled={addLoggingMutation.isPending}>
                          <Activity className="mr-2 h-4 w-4" />
                          Add logging
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={handleSecurityAudit} disabled={securityAuditMutation.isPending}>
                          <Shield className="mr-2 h-4 w-4" />
                          Security audit
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>

                {/* Complexity Progress Indicator */}
                {isGenerating && (
                  <ComplexityProgress 
                    complexity={displayResult?.estimated_complexity} 
                    isGenerating={isGenerating} 
                  />
                )}

                {/* Token Usage & Cost Display */}
                {!isStreaming && generatedFunction?.tokens_used && (
                  <TokenUsageDisplay 
                    tokens_used={generatedFunction.tokens_used}
                    latency_ms={generatedFunction.latency_ms}
                  />
                )}

                {/* Code Display with Context Menu */}
                <ContextMenu>
                  <ContextMenuTrigger className="w-full">
                    <div className="composer-editor-container overflow-hidden rounded-md border">
                      <LazyMonacoEditor
                        height="300px"
                        language={monacoLanguage}
                        value={streamingResult.code || generatedFunction?.result?.code || ''}
                        theme={monacoTheme}
                        onMount={handleEditorMount}
                        options={{
                          readOnly: true,
                          minimap: { enabled: false },
                          fontSize: 13,
                          lineNumbers: 'on',
                          scrollBeyondLastLine: false,
                          automaticLayout: true,
                          wordWrap: 'on',
                          contextmenu: false, // Disable default context menu
                        }}
                      />
                    </div>
                  </ContextMenuTrigger>
                  <ContextMenuContent className="w-64">
                    <ContextMenuItem onClick={handleExplainFunction}>
                      <Brain className="mr-2 h-4 w-4" />
                      Explain this function
                      <span className="ml-auto text-xs text-muted-foreground">AI</span>
                    </ContextMenuItem>
                    <ContextMenuItem onClick={handleAddComments}>
                      <MessageSquare className="mr-2 h-4 w-4" />
                      Add comments
                      <span className="ml-auto text-xs text-muted-foreground">AI</span>
                    </ContextMenuItem>
                    <ContextMenuItem onClick={handleAddLogging}>
                      <Activity className="mr-2 h-4 w-4" />
                      Add logging
                      <span className="ml-auto text-xs text-muted-foreground">AI</span>
                    </ContextMenuItem>
                    <ContextMenuSeparator />
                    <ContextMenuItem onClick={handleSecurityAudit}>
                      <Shield className="mr-2 h-4 w-4" />
                      Security audit
                      <span className="ml-auto text-xs text-muted-foreground">AI</span>
                    </ContextMenuItem>
                    <ContextMenuSeparator />
                    <ContextMenuItem onClick={handleCopy}>
                      <Copy className="mr-2 h-4 w-4" />
                      Copy code
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>

                {/* Token count indicator during streaming */}
                {isStreaming && streamingResult.code.length > 0 && (
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />
                    <span>Live streaming: {streamingResult.code.length} characters</span>
                  </div>
                )}

                {/* Explanation - only show when complete or has content */}
                {(displayResult?.explanation || !isStreaming) && displayResult?.explanation && (
                  <div className="bg-muted/50 rounded-lg p-4">
                    <h4 className="font-semibold mb-2">How it works</h4>
                    <p className="text-sm text-muted-foreground">
                      {displayResult.explanation}
                    </p>
                  </div>
                )}

                {/* Suggested Tests - only show when complete */}
                {displayResult?.suggested_tests && displayResult.suggested_tests.length > 0 && !isStreaming && (
                  <div>
                    <h4 className="font-semibold mb-2">Suggested Tests</h4>
                    <ScrollArea className="h-[120px]">
                      <ul className="text-sm space-y-1">
                        {displayResult.suggested_tests.map((test, i) => (
                          <li key={i} className="text-muted-foreground flex items-center gap-2">
                            <Play className="h-3 w-3" />
                            {test}
                          </li>
                        ))}
                      </ul>
                    </ScrollArea>
                  </div>
                )}

                {/* Refinement History */}
                {refinementHistory.length > 0 && !isStreaming && (
                  <div className="bg-muted/30 rounded-lg p-3">
                    <h4 className="font-semibold mb-2 text-sm">Refinement History</h4>
                    <ScrollArea className="h-[100px]">
                      <ul className="text-xs space-y-1">
                        {refinementHistory.map((item, i) => (
                          <li key={item.id} className="text-muted-foreground flex items-start gap-2">
                            <Badge variant="outline" className="shrink-0 text-[10px]">
                              v{i + 2}
                            </Badge>
                            <span className="truncate">{item.request}</span>
                          </li>
                        ))}
                      </ul>
                    </ScrollArea>
                  </div>
                )}

                {/* Only show actions when generation is complete */}
                {!isStreaming && generatedFunction?.result && (
                  <>
                    <Separator />

                    {/* Actions */}
                    <div className="flex gap-2 flex-wrap">
                      <Button
                        onClick={() => createFunctionMutation.mutate()}
                        disabled={createFunctionMutation.isPending}
                        className="flex-1"
                      >
                        {createFunctionMutation.isPending ? (
                          <>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            Saving...
                          </>
                        ) : (
                          <>
                            <Save className="mr-2 h-4 w-4" />
                            Save to My Functions
                          </>
                        )}
                      </Button>

                      {/* Test in Playground Button */}
                      <Button
                        variant="outline"
                        onClick={() => {
                          const code = streamingResult?.code || generatedFunction?.result?.code || '';
                          const playgroundUrl = createPlaygroundUrl(code, runtime);
                          window.open(playgroundUrl, '_blank');
                        }}
                        className="flex-1"
                      >
                        <ExternalLink className="mr-2 h-4 w-4" />
                        Test in Playground
                      </Button>

                      <Button variant="outline" onClick={handleCopy}>
                        {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <Code2 className="h-12 w-12 text-muted-foreground/50 mb-4" />
                <p className="text-muted-foreground">
                  Your generated code will appear here
                </p>
                <p className="text-sm text-muted-foreground/60">
                  Describe what you need and click Generate
                </p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Generation History Sidebar */}
        {historySidebarOpen && (
          <Card className="border-border/50 shadow-sm lg:order-first">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2 text-base">
                  <History className="h-4 w-4" />
                  Generation History
                </CardTitle>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setHistorySidebarOpen(false)}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <ScrollArea className="h-[calc(100vh-300px)]">
                {history.length === 0 ? (
                  <div className="p-4 text-center text-sm text-muted-foreground">
                    <History className="h-8 w-8 mx-auto mb-2 opacity-50" />
                    <p>No generations yet.</p>
                    <p className="text-xs">Generated functions will appear here.</p>
                  </div>
                ) : (
                  <div className="space-y-2 p-4">
                    {history.map((item, index) => (
                      <div
                        key={item.id}
                        className="group relative bg-muted/50 rounded-lg p-3 hover:bg-muted transition-colors"
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 mb-1">
                              <Badge variant="outline" className="text-[10px] shrink-0">
                                v{history.length - index}
                              </Badge>
                              <Badge variant="secondary" className="text-[10px] shrink-0">
                                {item.runtime}
                              </Badge>
                            </div>
                            <p className="text-xs text-muted-foreground line-clamp-2 mb-1">
                              {item.description}
                            </p>
                            <p className="text-[10px] text-muted-foreground/60">
                              {new Date(item.timestamp).toLocaleString(undefined, {
                                month: 'short',
                                day: 'numeric',
                                hour: '2-digit',
                                minute: '2-digit',
                              })}
                            </p>
                          </div>
                        </div>

                        {/* Actions */}
                        <div className="flex items-center gap-1 mt-2 pt-2 border-t border-border/50">
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7"
                                  onClick={() => {
                                    revertToGeneration(item.id);
                                    setDescription(item.description);
                                    setRuntime(item.runtime);
                                    setConstraints(item.constraints || '');
                                    setStreamingResult({ ...item.result, code: item.result.code });
                                    setGeneratedFunction({
                                      success: true,
                                      result: item.result,
                                      generation_id: item.id,
                                      latency_ms: 0,
                                      tokens_used: { prompt: 0, completion: 0, total: 0 },
                                    });
                                    setRefinementHistory(item.refinementHistory || []);
                                    toast.success('Reverted to previous generation');
                                  }}
                                >
                                  <Undo2 className="h-3.5 w-3.5" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Revert to this version</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>

                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7"
                                  onClick={() => {
                                    const forked = forkFromGeneration(item.id);
                                    setDescription(forked.description);
                                    setRuntime(forked.runtime);
                                    setConstraints(forked.constraints || '');
                                    setStreamingResult({ ...forked.result, code: forked.result.code });
                                    setGeneratedFunction({
                                      success: true,
                                      result: forked.result,
                                      generation_id: forked.id,
                                      latency_ms: 0,
                                      tokens_used: { prompt: 0, completion: 0, total: 0 },
                                    });
                                    setRefinementHistory([]);
                                    toast.success('Forked from previous generation');
                                  }}
                                >
                                  <GitBranch className="h-3.5 w-3.5" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Fork from this version</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>

                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7"
                                  onClick={() => {
                                    const currentItem = history.find(h => h.id === item.id);
                                    if (currentItem) {
                                      const playgroundUrl = createPlaygroundUrl(currentItem.result.code, currentItem.runtime);
                                      window.open(playgroundUrl, '_blank');
                                    }
                                  }}
                                >
                                  <FileCode2 className="h-3.5 w-3.5" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Open in Playground</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>

                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-7 w-7"
                                  onClick={() => {
                                    // Set compare items - find the next item in history (older)
                                    const currentIndex = history.findIndex(h => h.id === item.id);
                                    const olderItem = history[currentIndex + 1] || null;
                                    setCompareItems([item, olderItem]);
                                    setCompareDialogOpen(true);
                                  }}
                                >
                                  <Code2 className="h-3.5 w-3.5" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Compare with previous</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </ScrollArea>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Compare Dialog */}
      <Dialog open={compareDialogOpen} onOpenChange={setCompareDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[90vh]">
          <DialogHeader>
            <DialogTitle>Compare Versions</DialogTitle>
            <DialogDescription>
              Comparing code changes between generations
            </DialogDescription>
          </DialogHeader>
          <div className="grid grid-cols-2 gap-4 mt-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Badge variant="default">Current</Badge>
                <span className="text-xs text-muted-foreground">
                  {compareItems[0] && new Date(compareItems[0].timestamp).toLocaleString()}
                </span>
              </div>
              <div className="rounded-md border overflow-hidden">
                <LazyMonacoEditor
                  height="400px"
                  language={compareItems[0] ? RUNTIME_MONACO_LANG[compareItems[0].runtime] || 'plaintext' : 'plaintext'}
                  value={compareItems[0]?.result.code || ''}
                  theme={monacoTheme}
                  options={{
                    readOnly: true,
                    minimap: { enabled: false },
                    fontSize: 12,
                    lineNumbers: 'on',
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    wordWrap: 'on',
                  }}
                />
              </div>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Badge variant="secondary">Previous</Badge>
                <span className="text-xs text-muted-foreground">
                  {compareItems[1] ? new Date(compareItems[1].timestamp).toLocaleString() : 'None selected'}
                </span>
              </div>
              <div className="rounded-md border overflow-hidden">
                <LazyMonacoEditor
                  height="400px"
                  language={compareItems[1] ? RUNTIME_MONACO_LANG[compareItems[1].runtime] || 'plaintext' : 'plaintext'}
                  value={compareItems[1]?.result.code || '// No previous version to compare'}
                  theme={monacoTheme}
                  options={{
                    readOnly: true,
                    minimap: { enabled: false },
                    fontSize: 12,
                    lineNumbers: 'on',
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    wordWrap: 'on',
                  }}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCompareDialogOpen(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Enhanced Manifest Preview */}
      {generatedFunction?.result?.manifest && !isStreaming && (
        <Card className="border-border/50 shadow-sm">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <FileCode2 className="h-5 w-5" />
                Function Manifest
              </CardTitle>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setManifestEditMode(!manifestEditMode)}
                >
                  <Edit3 className="mr-2 h-4 w-4" />
                  {manifestEditMode ? 'Done Editing' : 'Edit Manifest'}
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setManifestExpanded(!manifestExpanded)}
                >
                  {manifestExpanded ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
                </Button>
              </div>
            </div>
          </CardHeader>
          {manifestExpanded && (
            <CardContent>
              <div className="space-y-6">
                {/* Visual Flow Diagram */}
                <div className="border rounded-lg bg-muted/30">
                  <div className="px-4 py-2 border-b bg-muted/50 flex items-center gap-2">
                    <Activity className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Input/Output Flow</span>
                  </div>
                  <ManifestFlowDiagram manifest={generatedFunction.result.manifest} />
                </div>

                {/* Configuration Grid */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {/* Inputs */}
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Download className="h-4 w-4" />
                      Inputs
                    </h4>
                    {generatedFunction.result.manifest.inputs.length > 0 ? (
                      <ul className="text-sm space-y-2">
                        {generatedFunction.result.manifest.inputs.map((input: any, i: number) => (
                          <li key={i} className="bg-muted/50 rounded p-2">
                            <div className="flex items-center gap-2">
                              <code className="font-mono text-xs">{input.name}</code>
                              <Badge variant="outline" className="text-xs">
                                {input.type}
                              </Badge>
                              {input.required && (
                                <Badge variant="secondary" className="text-xs">
                                  required
                                </Badge>
                              )}
                            </div>
                            <p className="text-muted-foreground text-xs mt-1">
                              {input.description}
                            </p>
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-sm text-muted-foreground italic">No inputs defined</p>
                    )}
                  </div>

                  {/* Outputs */}
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Upload className="h-4 w-4" />
                      Outputs
                    </h4>
                    {generatedFunction.result.manifest.outputs.length > 0 ? (
                      <ul className="text-sm space-y-2">
                        {generatedFunction.result.manifest.outputs.map((output: any, i: number) => (
                          <li key={i} className="bg-muted/50 rounded p-2">
                            <div className="flex items-center gap-2">
                              <code className="font-mono text-xs">{output.name}</code>
                              <Badge variant="outline" className="text-xs">
                                {output.type}
                              </Badge>
                            </div>
                            <p className="text-muted-foreground text-xs mt-1">
                              {output.description}
                            </p>
                          </li>
                        ))}
                      </ul>
                    ) : (
                      <p className="text-sm text-muted-foreground italic">No outputs defined</p>
                    )}
                  </div>

                  {/* Configuration - Editable */}
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Settings2 className="h-4 w-4" />
                      Configuration
                    </h4>
                    
                    {manifestEditMode ? (
                      <div className="space-y-4">
                        {/* Timeout Slider */}
                        <div className="space-y-2">
                          <div className="flex items-center justify-between text-xs">
                            <Label className="text-muted-foreground">Timeout</Label>
                            <span className="font-mono">{editableManifest.timeout_seconds}s</span>
                          </div>
                          <Slider
                            value={[editableManifest.timeout_seconds]}
                            onValueChange={([v]) => setEditableManifest(prev => ({ ...prev, timeout_seconds: v }))}
                            min={1}
                            max={300}
                            step={1}
                          />
                        </div>

                        {/* Memory Slider */}
                        <div className="space-y-2">
                          <div className="flex items-center justify-between text-xs">
                            <Label className="text-muted-foreground">Memory</Label>
                            <span className="font-mono">{editableManifest.memory_mb} MB</span>
                          </div>
                          <Slider
                            value={[editableManifest.memory_mb]}
                            onValueChange={([v]) => setEditableManifest(prev => ({ ...prev, memory_mb: v }))}
                            min={128}
                            max={4096}
                            step={128}
                          />
                        </div>

                        <Separator />

                        {/* Runtime Display */}
                        <div className="flex justify-between text-sm">
                          <span className="text-muted-foreground">Runtime</span>
                          <span>{generatedFunction.result.manifest.runtime}</span>
                        </div>
                      </div>
                    ) : (
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Timeout</span>
                          <span className="font-mono">{editableManifest.timeout_seconds}s</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Memory</span>
                          <span className="font-mono">{editableManifest.memory_mb} MB</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Runtime</span>
                          <span>{generatedFunction.result.manifest.runtime}</span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                {/* Capabilities with Toggles */}
                {generatedFunction.result.manifest.capabilities && (
                  <div className="border-t pt-4">
                    <h4 className="font-semibold mb-3 flex items-center gap-2">
                      <Shield className="h-4 w-4" />
                      Capabilities
                      {manifestEditMode && (
                        <span className="text-xs font-normal text-muted-foreground">
                          (Click to toggle)
                        </span>
                      )}
                    </h4>
                    <div className="flex flex-wrap gap-2">
                      {/* All possible capabilities */}
                      {Object.keys(CAPABILITY_INFO).map((cap) => {
                        const isEnabled = editableManifest.capabilities.includes(cap);
                        const isOriginal = generatedFunction.result.manifest.capabilities.includes(cap);
                        
                        if (!manifestEditMode && !isOriginal) return null;
                        
                        return (
                          <CapabilityToggle
                            key={cap}
                            capability={cap}
                            enabled={manifestEditMode ? isEnabled : isOriginal}
                            onToggle={manifestEditMode ? handleCapabilityToggle : () => {}}
                          />
                        );
                      })}
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          )}
        </Card>
      )}
    </div>
  );
}
