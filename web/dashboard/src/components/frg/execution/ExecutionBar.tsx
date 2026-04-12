/**
 * ExecutionBar Component
 * Controls: Run graph, Stop, Debug mode, Step execution
 */

import { useCallback, useState, useEffect } from 'react';
import { 
  Play, 
  Square, 
  Pause, 
  StepForward,
  Bug,
  Terminal,
  Activity,
  Zap,
  Clock,
  CheckCircle,
  AlertCircle,
  RotateCcw,
  Settings,
  ChevronUp,
  ChevronDown,
  LayoutTemplate,
  Maximize2,
  Minimize2,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Separator } from '@/components/ui/separator';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';

import { useFRGStore, selectExecutionStats } from '@/stores/frgStore';

interface ExecutionBarProps {
  className?: string;
  onRun?: () => void;
  onStop?: () => void;
  onStep?: () => void;
}

export function ExecutionBar({ className, onRun, onStop, onStep }: ExecutionBarProps) {
  const store = useFRGStore();
  const { 
    executionStatus, 
    executionProgress, 
    editorMode,
    setEditorMode,
    events,
    nodes,
  } = store;
  
  const stats = selectExecutionStats(store);
  const [isExpanded, setIsExpanded] = useState(false);
  const [debugMode, setDebugMode] = useState(false);
  const [showLogs, setShowLogs] = useState(true);

  const isRunning = executionStatus === 'running';
  const isPaused = executionStatus === 'paused';
  const isCompleted = executionStatus === 'completed' || executionStatus === 'failed';

  const handleRun = useCallback(() => {
    if (onRun) onRun();
  }, [onRun]);

  const handleStop = useCallback(() => {
    if (onStop) onStop();
  }, [onStop]);

  const handleStep = useCallback(() => {
    if (onStep) onStep();
  }, [onStep]);

  const toggleDebugMode = useCallback(() => {
    setDebugMode(!debugMode);
    setEditorMode(debugMode ? 'edit' : 'debug');
  }, [debugMode, setEditorMode]);

  // Get recent events
  const recentEvents = events.slice(-5).reverse();

  return (
    <div className={cn("border-t border-[var(--border-subtle)] bg-[var(--bg-secondary)]", className)}>
      {/* Main Bar */}
      <div className="flex items-center justify-between px-4 py-2">
        {/* Left: Controls */}
        <div className="flex items-center gap-2">
          {isRunning ? (
            <>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={handleRun}
                      className="h-8"
                    >
                      <Pause className="w-4 h-4 mr-1" />
                      Pause
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Pause execution</TooltipContent>
                </Tooltip>
              </TooltipProvider>
              
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="destructive" 
                      size="sm" 
                      onClick={handleStop}
                      className="h-8"
                    >
                      <Square className="w-4 h-4 mr-1" />
                      Stop
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Stop execution</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </>
          ) : isPaused ? (
            <>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="default" 
                      size="sm" 
                      onClick={handleRun}
                      className="h-8 bg-gradient-to-r from-green-500 to-emerald-500"
                    >
                      <Play className="w-4 h-4 mr-1" />
                      Resume
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Resume execution</TooltipContent>
                </Tooltip>
              </TooltipProvider>
              
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={handleStep}
                      className="h-8"
                    >
                      <StepForward className="w-4 h-4 mr-1" />
                      Step
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Execute next node</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </>
          ) : (
            <>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button 
                      variant="default" 
                      size="sm" 
                      onClick={handleRun}
                      className="h-8 bg-gradient-to-r from-green-500 to-emerald-500"
                    >
                      <Play className="w-4 h-4 mr-1" />
                      Run
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Run graph (Ctrl+Enter)</TooltipContent>
                </Tooltip>
              </TooltipProvider>
              
              {debugMode && (
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={handleStep}
                        className="h-8"
                      >
                        <StepForward className="w-4 h-4 mr-1" />
                        Step
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Debug: step execution</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
            </>
          )}

          <Separator orientation="vertical" className="h-6 mx-1" />

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button 
                  variant={debugMode ? 'default' : 'outline'}
                  size="sm"
                  onClick={toggleDebugMode}
                  className={cn("h-8", debugMode && "bg-purple-500")}
                >
                  <Bug className="w-4 h-4 mr-1" />
                  Debug
                </Button>
              </TooltipTrigger>
              <TooltipContent>Toggle debug mode</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        {/* Center: Status & Progress */}
        <div className="flex items-center gap-4 flex-1 max-w-md mx-4">
          {isRunning || isPaused ? (
            <div className="flex-1 space-y-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-[var(--text-secondary)]">
                  {stats.completedNodes}/{stats.totalNodes} nodes completed
                </span>
                <span className="text-[var(--text-secondary)]">
                  {executionProgress}%
                </span>
              </div>
              <Progress value={executionProgress} className="h-1.5" />
            </div>
          ) : (
            <div className="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
              <Zap className="w-4 h-4" />
              <span>{nodes.length} nodes ready</span>
            </div>
          )}
        </div>

        {/* Right: Stats & Actions */}
        <div className="flex items-center gap-3">
          {isRunning && (
            <div className="flex items-center gap-2 text-xs text-[var(--text-secondary)]">
              <Activity className="w-4 h-4 text-green-500 animate-pulse" />
              <span>Live</span>
            </div>
          )}

          {isCompleted && (
            <div className="flex items-center gap-2">
              {executionStatus === 'completed' ? (
                <Badge variant="default" className="text-xs">
                  <CheckCircle className="w-3 h-3 mr-1" />
                  Success
                </Badge>
              ) : (
                <Badge variant="destructive" className="text-xs">
                  <AlertCircle className="w-3 h-3 mr-1" />
                  Failed
                </Badge>
              )}
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleRun}>
                      <RotateCcw className="w-4 h-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Re-run</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          )}

          <Separator orientation="vertical" className="h-6 mx-1" />

          <div className="flex items-center gap-2">
            <Switch 
              checked={showLogs}
              onCheckedChange={setShowLogs}
              id="show-logs"
            />
            <Label htmlFor="show-logs" className="text-xs cursor-pointer">
              <Terminal className="w-3 h-3 inline mr-1" />
              Logs
            </Label>
          </div>

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button 
                  variant="ghost" 
                  size="icon" 
                  className="h-8 w-8"
                  onClick={() => setIsExpanded(!isExpanded)}
                >
                  {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronUp className="w-4 h-4" />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                {isExpanded ? <span>Collapse logs</span> : <span>Expand logs</span>}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>

      {/* Expanded Logs Panel */}
      {isExpanded && showLogs && (
        <div className="border-t border-[var(--border-subtle)] bg-[var(--bg-primary)] p-4">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <Terminal className="w-4 h-4 text-[var(--text-secondary)]" />
              <span className="text-sm font-medium text-[var(--text-primary)]">Execution Logs</span>
            </div>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="sm" className="h-7 text-xs">
                <RotateCcw className="w-3 h-3 mr-1" />
                Clear
              </Button>
              <Button variant="ghost" size="icon" className="h-7 w-7">
                <Settings className="w-3 h-3" />
              </Button>
            </div>
          </div>
          
          <div className="font-mono text-xs space-y-1 max-h-40 overflow-auto bg-[var(--bg-tertiary)] rounded-lg p-3">
            {recentEvents.length === 0 ? (
              <div className="text-[var(--text-muted)] italic">
                No execution events yet. Run the graph to see logs.
              </div>
            ) : (
              recentEvents.map((event) => (
                <div key={event.id} className="flex items-start gap-2">
                  <span className="text-[var(--text-muted)] shrink-0">
                    {new Date(event.timestamp).toLocaleTimeString()}
                  </span>
                  <span className={cn(
                    "uppercase w-16 shrink-0",
                    event.eventType === 'error' && "text-red-500",
                    event.eventType === 'complete' && "text-green-500",
                    event.eventType === 'stream' && "text-blue-500",
                  )}>
                    {event.eventType}
                  </span>
                  <span className="text-[var(--text-secondary)]">
                    {event.nodeId ? `Node ${event.nodeId}: ` : ''}
                    {event.payload ? JSON.stringify(event.payload).slice(0, 100) : ''}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default ExecutionBar;
