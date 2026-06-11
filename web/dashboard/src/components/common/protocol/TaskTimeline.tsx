/**
 * TaskTimeline — A2A task lifecycle viewer.
 * Renders state transitions: submitted → working → input-required → completed/failed/canceled.
 * Receives a receipt/task and overlays the state transitions.
 */

import { CheckCircle2, XCircle, Clock, AlertCircle, Ban, Loader2 } from 'lucide-react';

export type TaskState = 'submitted' | 'working' | 'input-required' | 'completed' | 'failed' | 'canceled';

interface TaskTimelineProps {
  currentState: TaskState;
  transitions?: { from: TaskState; to: TaskState; timestamp: string }[];
  className?: string;
}

const STATE_ORDER: TaskState[] = ['submitted', 'working', 'input-required', 'completed'];

const STATE_CONFIG: Record<TaskState, { icon: typeof CheckCircle2; label: string; color: string }> = {
  submitted: { icon: Clock, label: 'Submitted', color: 'text-blue-400 bg-blue-500/10 border-blue-500/20' },
  working: { icon: Loader2, label: 'Working', color: 'text-amber-400 bg-amber-500/10 border-amber-500/20' },
  'input-required': { icon: AlertCircle, label: 'Input Required', color: 'text-orange-400 bg-orange-500/10 border-orange-500/20' },
  completed: { icon: CheckCircle2, label: 'Completed', color: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20' },
  failed: { icon: XCircle, label: 'Failed', color: 'text-red-400 bg-red-500/10 border-red-500/20' },
  canceled: { icon: Ban, label: 'Canceled', color: 'text-zinc-400 bg-zinc-500/10 border-zinc-500/20' },
};

function isTerminal(state: TaskState): boolean {
  return state === 'completed' || state === 'failed' || state === 'canceled';
}

function stateIndex(state: TaskState): number {
  if (state === 'failed' || state === 'canceled') return 3; // terminal
  return STATE_ORDER.indexOf(state);
}

export function TaskTimeline({ currentState, transitions = [], className = '' }: TaskTimelineProps) {
  const currentIdx = stateIndex(currentState);
  const terminal = isTerminal(currentState);

  // Build the steps to display.
  const steps = STATE_ORDER.map((state, idx) => {
    const config = STATE_CONFIG[state];
    const Icon = config.icon;
    const reached = idx <= currentIdx || terminal;
    const isCurrent = state === currentState;
    const transition = transitions.find((t) => t.to === state);

    return (
      <div key={state} className="flex items-center gap-2">
        {/* Connector line */}
        {idx > 0 && (
          <div
            className={`w-8 h-0.5 ${
              idx <= currentIdx ? 'bg-brand-500/40' : 'bg-zinc-700/40'
            }`}
          />
        )}
        {/* Step */}
        <div
          className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full border text-xs font-medium transition-all ${
            isCurrent
              ? config.color + ' ring-1 ring-offset-1 ring-offset-bg-primary'
              : reached
              ? config.color
              : 'text-zinc-500 bg-zinc-800/40 border-zinc-700/30'
          }`}
        >
          <Icon className={`w-3.5 h-3.5 ${isCurrent && state === 'working' ? 'animate-spin' : ''}`} />
          {config.label}
          {transition && (
            <span className="text-[10px] opacity-60 ml-1">
              {new Date(transition.timestamp).toLocaleTimeString()}
            </span>
          )}
        </div>
      </div>
    );
  });

  // If the task ended in a terminal state not in the main flow, show it.
  const terminalConfig = STATE_CONFIG[currentState];
  if (terminal && !['completed'].includes(currentState)) {
    const TerminalIcon = terminalConfig.icon;
    steps.push(
      <div key="terminal" className="flex items-center gap-2">
        <div className="w-8 h-0.5 bg-red-500/40" />
        <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full border text-xs font-medium ${terminalConfig.color}`}>
          <TerminalIcon className="w-3.5 h-3.5" />
          {terminalConfig.label}
        </div>
      </div>
    );
  }

  return (
    <div className={`flex items-center flex-wrap gap-1 ${className}`}>
      {steps}
    </div>
  );
}
