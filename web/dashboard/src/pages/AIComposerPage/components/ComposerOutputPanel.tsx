import type { FunctionGenerationResponse, FunctionGenerationResult } from '@/api/composer';
import { RUNTIME_MONACO_LANG } from '@/api/composer';
import Editor, { type OnMount } from '@monaco-editor/react';
import {
  Activity,
  Brain,
  Check,
  Code2,
  Copy,
  Edit3,
  ExternalLink,
  Loader2,
  MessageSquare,
  MoreVertical,
  Play,
  Save,
  Shield,
  Sparkles,
  Wand,
  Wand2,
} from 'lucide-react';
import { ModelUsedBadge } from '@/components/ai/ModelUsedBadge';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
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
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { createPlaygroundUrl } from '../../StandalonePlaygroundPage';
import { COMPLEXITY_COLORS, QUICK_REFINEMENTS } from '../constants';
import type { RefinementHistoryItem } from '../types';
import { ComplexityProgress, ConfidenceDisplay, TokenUsageDisplay } from './ComposerMetrics';

interface ComposerOutputPanelProps {
  runtime: string;
  monacoTheme: string;
  monacoLanguage: string;
  hasResult: boolean;
  isGenerating: boolean;
  isStreaming: boolean;
  isRefining: boolean;
  copied: boolean;
  confidenceScore?: number;
  generatedFunction: FunctionGenerationResponse | null;
  streamingResult: Partial<FunctionGenerationResult> & { code: string };
  displayResult: Partial<FunctionGenerationResult> | undefined;
  refinementHistory: RefinementHistoryItem[];
  refinementDialogOpen: boolean;
  refinementRequest: string;
  explainPending: boolean;
  commentsPending: boolean;
  loggingPending: boolean;
  securityPending: boolean;
  savePending: boolean;
  onRefinementDialogOpenChange: (open: boolean) => void;
  onRefinementRequestChange: (value: string) => void;
  onQuickRefinement: (prompt: string) => void;
  onCustomRefinement: () => void;
  onCopy: () => void;
  onExplain: () => void;
  onAddComments: () => void;
  onAddLogging: () => void;
  onSecurityAudit: () => void;
  onEditorMount: OnMount;
  onSave: () => void;
}

export function ComposerOutputPanel({
  runtime,
  monacoTheme,
  monacoLanguage,
  hasResult,
  isGenerating,
  isStreaming,
  isRefining,
  copied,
  confidenceScore,
  generatedFunction,
  streamingResult,
  displayResult,
  refinementHistory,
  refinementDialogOpen,
  refinementRequest,
  explainPending,
  commentsPending,
  loggingPending,
  securityPending,
  savePending,
  onRefinementDialogOpenChange,
  onRefinementRequestChange,
  onQuickRefinement,
  onCustomRefinement,
  onCopy,
  onExplain,
  onAddComments,
  onAddLogging,
  onSecurityAudit,
  onEditorMount,
  onSave,
}: ComposerOutputPanelProps) {
  return (
    <Card className="border-border/50 shadow-sm">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Code2 className="h-5 w-5" />
          Generated Code
          {hasResult && !isGenerating && (
            <Badge variant="outline" className="ml-2 text-xs">
              {refinementHistory.length > 0 ? `v${refinementHistory.length + 1}` : 'v1'}
            </Badge>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {isGenerating ? (
          <div className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="relative">
              <div className="h-16 w-16 rounded-full border-4 border-violet-200 border-t-violet-500 animate-spin" />
              <Sparkles className="h-6 w-6 text-violet-500 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2" />
            </div>
            <p className="text-muted-foreground animate-pulse">
              {isRefining
                ? 'FlyMind is refining your function...'
                : 'FlyMind is crafting your function...'}
            </p>
            {streamingResult.code.length > 0 && (
              <p className="text-xs text-muted-foreground">
                {streamingResult.code.length} characters generated so far
              </p>
            )}
          </div>
        ) : hasResult ? (
          <div className="space-y-4">
            <div className="flex items-center justify-between flex-wrap gap-2">
              <div className="flex items-center gap-2 flex-wrap">
                <Badge variant="secondary" className="font-mono">
                  {displayResult?.runtime || runtime}
                </Badge>
                {displayResult?.estimated_complexity && (
                  <Badge
                    className={
                      COMPLEXITY_COLORS[
                        displayResult.estimated_complexity as keyof typeof COMPLEXITY_COLORS
                      ]?.text +
                      ' ' +
                      COMPLEXITY_COLORS[
                        displayResult.estimated_complexity as keyof typeof COMPLEXITY_COLORS
                      ]?.bg.replace('bg-', 'bg-').replace('500', '500/20')
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
                <ConfidenceDisplay score={confidenceScore} />
                <ModelUsedBadge
                  modelUsed={generatedFunction?.model_used}
                  costUsd={generatedFunction?.cost_usd}
                />
              </div>
              <div className="flex items-center gap-2">
                <Dialog open={refinementDialogOpen} onOpenChange={onRefinementDialogOpenChange}>
                  <DialogTrigger asChild>
                    <Button variant="outline" size="sm" disabled={isStreaming}>
                      <Edit3 className="mr-2 h-4 w-4" />
                      Edit Request
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="max-w-lg">
                    <DialogHeader>
                      <DialogTitle className="flex items-center gap-2">
                        <Wand className="h-5 w-5" />
                        Refine Your Function
                      </DialogTitle>
                      <DialogDescription>
                        Request modifications to improve the generated code. The AI will preserve
                        the core functionality while applying your changes.
                      </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4 py-4">
                      <div className="space-y-2">
                        <Label className="text-xs text-muted-foreground uppercase">
                          Quick Refinements
                        </Label>
                        <div className="flex flex-wrap gap-2">
                          {QUICK_REFINEMENTS.map((item) => (
                            <Button
                              key={item.label}
                              variant="outline"
                              size="sm"
                              onClick={() => onQuickRefinement(item.prompt)}
                              disabled={isRefining}
                              className="text-xs"
                            >
                              {item.label}
                            </Button>
                          ))}
                        </div>
                      </div>

                      <Separator />

                      <div className="space-y-2">
                        <Label htmlFor="custom-refinement">Custom Request</Label>
                        <Textarea
                          id="custom-refinement"
                          placeholder="e.g., Add rate limiting, improve error messages, convert to async/await..."
                          value={refinementRequest}
                          onChange={(e) => onRefinementRequestChange(e.target.value)}
                          className="min-h-[100px] resize-none"
                          disabled={isRefining}
                        />
                      </div>
                    </div>

                    <DialogFooter>
                      <Button
                        variant="outline"
                        onClick={() => onRefinementDialogOpenChange(false)}
                        disabled={isRefining}
                      >
                        Cancel
                      </Button>
                      <Button
                        onClick={onCustomRefinement}
                        disabled={!refinementRequest.trim() || isRefining}
                        className="bg-gradient-to-r from-violet-500 to-purple-600"
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
                      <Button variant="ghost" size="icon" onClick={onCopy}>
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

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-48">
                    <DropdownMenuItem onClick={onExplain} disabled={explainPending}>
                      <Brain className="mr-2 h-4 w-4" />
                      Explain this function
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={onAddComments} disabled={commentsPending}>
                      <MessageSquare className="mr-2 h-4 w-4" />
                      Add comments
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={onAddLogging} disabled={loggingPending}>
                      <Activity className="mr-2 h-4 w-4" />
                      Add logging
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={onSecurityAudit} disabled={securityPending}>
                      <Shield className="mr-2 h-4 w-4" />
                      Security audit
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>

            {isGenerating && (
              <ComplexityProgress
                complexity={displayResult?.estimated_complexity}
                isGenerating={isGenerating}
              />
            )}

            {!isStreaming && generatedFunction?.tokens_used && (
              <TokenUsageDisplay
                tokens_used={generatedFunction.tokens_used}
                latency_ms={generatedFunction.latency_ms}
              />
            )}

            <ContextMenu>
              <ContextMenuTrigger className="w-full">
                <div className="rounded-md border bg-muted/50 overflow-hidden">
                  <Editor
                    height="300px"
                    language={monacoLanguage}
                    value={streamingResult.code || generatedFunction?.result?.code || ''}
                    theme={monacoTheme}
                    onMount={onEditorMount}
                    options={{
                      readOnly: true,
                      minimap: { enabled: false },
                      fontSize: 13,
                      lineNumbers: 'on',
                      scrollBeyondLastLine: false,
                      automaticLayout: true,
                      wordWrap: 'on',
                      contextmenu: false,
                    }}
                  />
                </div>
              </ContextMenuTrigger>
              <ContextMenuContent className="w-64">
                <ContextMenuItem onClick={onExplain}>
                  <Brain className="mr-2 h-4 w-4" />
                  Explain this function
                  <span className="ml-auto text-xs text-muted-foreground">AI</span>
                </ContextMenuItem>
                <ContextMenuItem onClick={onAddComments}>
                  <MessageSquare className="mr-2 h-4 w-4" />
                  Add comments
                  <span className="ml-auto text-xs text-muted-foreground">AI</span>
                </ContextMenuItem>
                <ContextMenuItem onClick={onAddLogging}>
                  <Activity className="mr-2 h-4 w-4" />
                  Add logging
                  <span className="ml-auto text-xs text-muted-foreground">AI</span>
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={onSecurityAudit}>
                  <Shield className="mr-2 h-4 w-4" />
                  Security audit
                  <span className="ml-auto text-xs text-muted-foreground">AI</span>
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={onCopy}>
                  <Copy className="mr-2 h-4 w-4" />
                  Copy code
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>

            {isStreaming && streamingResult.code.length > 0 && (
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />
                <span>Live streaming: {streamingResult.code.length} characters</span>
              </div>
            )}

            {(displayResult?.explanation || !isStreaming) && displayResult?.explanation && (
              <div className="bg-muted/50 rounded-lg p-4">
                <h4 className="font-semibold mb-2">How it works</h4>
                <p className="text-sm text-muted-foreground">{displayResult.explanation}</p>
              </div>
            )}

            {displayResult?.suggested_tests &&
              displayResult.suggested_tests.length > 0 &&
              !isStreaming && (
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

            {!isStreaming && generatedFunction?.result && (
              <>
                <Separator />
                <div className="flex gap-2 flex-wrap">
                  <Button onClick={onSave} disabled={savePending} className="flex-1">
                    {savePending ? (
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

                  <Button
                    variant="outline"
                    onClick={() => {
                      const code =
                        streamingResult?.code || generatedFunction?.result?.code || '';
                      const playgroundUrl = createPlaygroundUrl(code, runtime);
                      window.open(playgroundUrl, '_blank');
                    }}
                    className="flex-1"
                  >
                    <ExternalLink className="mr-2 h-4 w-4" />
                    Test in Playground
                  </Button>

                  <Button variant="outline" onClick={onCopy}>
                    {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                  </Button>
                </div>
              </>
            )}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Code2 className="h-12 w-12 text-muted-foreground/50 mb-4" />
            <p className="text-muted-foreground">Your generated code will appear here</p>
            <p className="text-sm text-muted-foreground/60">
              Describe what you need and click Generate
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Re-export for monaco language lookup used elsewhere
export { RUNTIME_MONACO_LANG };
