/**
 * AIAssistant Panel
 * Features: "Build me a workflow", Suggest nodes, Fix errors
 */

import { useState, useCallback, useRef, useEffect } from 'react';
import { 
  Sparkles, 
  Send, 
  Wand2, 
  Zap, 
  CheckCircle, 
  AlertCircle,
  Plus,
  RefreshCw,
  Lightbulb,
  Code,
  ArrowRight,
  Bot,
  User,
  Loader2,
  X,
  ThumbsUp,
  ThumbsDown,
} from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Card } from '@/components/ui/card';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

import { useFRGStore } from '@/stores/frgStore';
import type { AISuggestion } from '@/types/frg';

// Mock suggestions - in production, from API
const mockSuggestions: AISuggestion[] = [
  {
    id: 's1',
    type: 'add_node',
    description: 'Add error handling to your data processing chain',
    confidence: 0.92,
    affectedNodes: ['node-1', 'node-2'],
    action: {
      type: 'insert_node',
      payload: { nodeType: 'error-handler', target: 'node-2' },
    },
    explanation: 'Based on the current flow, adding an error handler after the data transformer will improve reliability',
  },
  {
    id: 's2',
    type: 'optimize',
    description: 'Parallelize independent operations',
    confidence: 0.88,
    affectedNodes: ['node-3', 'node-4'],
    action: {
      type: 'parallelize',
      payload: { nodes: ['node-3', 'node-4'] },
    },
    explanation: 'These two nodes don\'t depend on each other and can run simultaneously for 40% speedup',
  },
  {
    id: 's3',
    type: 'fix',
    description: 'Fix data type mismatch in edge connection',
    confidence: 0.95,
    affectedNodes: ['node-1', 'node-2'],
    action: {
      type: 'add_transform',
      payload: { edge: 'e-1-2', transform: 'map' },
    },
    explanation: 'The output from node-1 is an array but node-2 expects an object. Add a mapper to fix this.',
  },
];

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  suggestions?: AISuggestion[];
  isLoading?: boolean;
}

export function AIAssistantPanel() {
  const store = useFRGStore();
  const { 
    aiSuggestions, 
    isAiLoading, 
    setAiSuggestions, 
    dismissAiSuggestion, 
    applyAiSuggestion,
    setAiLoading,
    nodes,
    edges,
  } = store;

  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content: 'I\'m your AI assistant for building workflows. I can suggest nodes, optimize your graph, or help fix errors. What would you like to do?',
      timestamp: new Date(),
    },
  ]);
  const [input, setInput] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  // Scroll to bottom on new messages
  useEffect(() => {
    scrollRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const sendMessage = useCallback(() => {
    if (!input.trim()) return;

    const userMessage: ChatMessage = {
      id: `msg-${Date.now()}`,
      role: 'user',
      content: input,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setAiLoading(true);

    // Simulate AI response
    setTimeout(() => {
      const response: ChatMessage = {
        id: `msg-${Date.now()}`,
        role: 'assistant',
        content: 'I analyzed your workflow and found some opportunities:',
        timestamp: new Date(),
        suggestions: mockSuggestions,
      };
      setMessages(prev => [...prev, response]);
      setAiSuggestions(mockSuggestions);
      setAiLoading(false);
    }, 1500);
  }, [input, setAiLoading, setAiSuggestions]);

  const handleSuggestionAction = useCallback((suggestionId: string, action: 'apply' | 'dismiss') => {
    if (action === 'apply') {
      applyAiSuggestion(suggestionId);
    } else {
      dismissAiSuggestion(suggestionId);
    }
  }, [applyAiSuggestion, dismissAiSuggestion]);

  const quickActions = [
    { icon: Plus, label: 'Build workflow', prompt: 'Build me a workflow that...' },
    { icon: Zap, label: 'Optimize', prompt: 'Optimize my current workflow' },
    { icon: Wand2, label: 'Fix errors', prompt: 'Find and fix errors in my graph' },
    { icon: Lightbulb, label: 'Suggest nodes', prompt: 'Suggest nodes for my workflow' },
  ];

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-[var(--border-subtle)]">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-500 to-purple-600 flex items-center justify-center">
            <Sparkles className="w-4 h-4 text-white" />
          </div>
          <div>
            <h3 className="font-semibold text-sm text-[var(--text-primary)]">AI Assistant</h3>
            <p className="text-[10px] text-[var(--text-secondary)]">Powered by FlyMind AI</p>
          </div>
        </div>
      </div>

      {/* Suggestions Panel */}
      {aiSuggestions.length > 0 && (
        <div className="p-3 border-b border-[var(--border-subtle)] bg-brand-500/5">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-medium text-[var(--text-primary)]">
              Suggestions ({aiSuggestions.length})
            </span>
            <Button 
              variant="ghost" 
              size="sm" 
              className="h-6 text-[10px]"
              onClick={() => setAiSuggestions([])}
            >
              <X className="w-3 h-3 mr-1" />
              Clear
            </Button>
          </div>
          <div className="space-y-2">
            {aiSuggestions.map((suggestion) => (
              <Card 
                key={suggestion.id} 
                className="p-2 border-brand-500/20 bg-brand-500/5 hover:bg-brand-500/10 transition-colors"
              >
                <div className="flex items-start gap-2">
                  <div className="w-6 h-6 rounded bg-brand-500/20 flex items-center justify-center shrink-0 mt-0.5">
                    {suggestion.type === 'add_node' ? <Plus className="w-3 h-3 text-brand-500" /> :
                     suggestion.type === 'optimize' ? <Zap className="w-3 h-3 text-brand-500" /> :
                     suggestion.type === 'fix' ? <CheckCircle className="w-3 h-3 text-green-500" /> :
                     <Wand2 className="w-3 h-3 text-brand-500" />}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-medium text-[var(--text-primary)]">
                      {suggestion.description}
                    </p>
                    <p className="text-[10px] text-[var(--text-secondary)] mt-1">
                      {suggestion.explanation}
                    </p>
                    <div className="flex items-center gap-2 mt-2">
                      <Badge variant="secondary" className="text-[9px]">
                        {Math.round(suggestion.confidence * 100)}% confidence
                      </Badge>
                      <div className="flex gap-1 ml-auto">
                        <Button 
                          variant="ghost" 
                          size="icon" 
                          className="h-6 w-6"
                          onClick={() => handleSuggestionAction(suggestion.id, 'dismiss')}
                        >
                          <X className="w-3 h-3" />
                        </Button>
                        <Button 
                          size="sm" 
                          className="h-6 text-[10px]"
                          onClick={() => handleSuggestionAction(suggestion.id, 'apply')}
                        >
                          <CheckCircle className="w-3 h-3 mr-1" />
                          Apply
                        </Button>
                      </div>
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Chat */}
      <ScrollArea className="flex-1 p-3">
        <div className="space-y-4">
          {messages.map((msg) => (
            <div 
              key={msg.id}
              className={cn(
                "flex gap-2",
                msg.role === 'user' && "flex-row-reverse"
              )}
            >
              <div className={cn(
                "w-6 h-6 rounded-full flex items-center justify-center shrink-0",
                msg.role === 'user' 
                  ? "bg-[var(--bg-tertiary)]" 
                  : "bg-gradient-to-br from-brand-500 to-purple-600"
              )}>
                {msg.role === 'user' ? (
                  <User className="w-3 h-3 text-[var(--text-secondary)]" />
                ) : (
                  <Bot className="w-3 h-3 text-white" />
                )}
              </div>
              <div className={cn(
                "max-w-[80%] rounded-lg p-2 text-sm",
                msg.role === 'user' 
                  ? "bg-brand-500 text-white" 
                  : "bg-[var(--bg-tertiary)] text-[var(--text-primary)]"
              )}>
                {msg.content}
                
                {msg.suggestions && (
                  <div className="mt-2 space-y-1">
                    {msg.suggestions.map((s, i) => (
                      <div 
                        key={s.id} 
                        className="flex items-center gap-1 text-[10px] text-[var(--text-secondary)]"
                      >
                        <ArrowRight className="w-3 h-3" />
                        {s.description}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
          
          {isAiLoading && (
            <div className="flex gap-2">
              <div className="w-6 h-6 rounded-full bg-gradient-to-br from-brand-500 to-purple-600 flex items-center justify-center">
                <Bot className="w-3 h-3 text-white" />
              </div>
              <div className="bg-[var(--bg-tertiary)] rounded-lg p-3">
                <Loader2 className="w-4 h-4 animate-spin text-brand-500" />
              </div>
            </div>
          )}
          
          <div ref={scrollRef} />
        </div>

        {/* Quick Actions */}
        {messages.length === 1 && (
          <div className="mt-4">
            <p className="text-[10px] text-[var(--text-muted)] mb-2">Quick actions</p>
            <div className="grid grid-cols-2 gap-2">
              {quickActions.map((action) => (
                <Button
                  key={action.label}
                  variant="outline"
                  size="sm"
                  className="h-auto py-2 justify-start text-xs"
                  onClick={() => setInput(action.prompt)}
                >
                  <action.icon className="w-3 h-3 mr-2 text-brand-500" />
                  {action.label}
                </Button>
              ))}
            </div>
          </div>
        )}
      </ScrollArea>

      {/* Input */}
      <div className="p-3 border-t border-[var(--border-subtle)]">
        <div className="flex gap-2">
          <Textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ask me to build, optimize, or fix your workflow..."
            className="min-h-[60px] text-sm resize-none"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
              }
            }}
          />
          <Button 
            size="icon" 
            className="shrink-0"
            disabled={!input.trim() || isAiLoading}
            onClick={sendMessage}
          >
            {isAiLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Send className="w-4 h-4" />
            )}
          </Button>
        </div>
        <p className="text-[9px] text-[var(--text-muted)] mt-2 text-center">
          AI suggestions are suggestions only. Review before applying.
        </p>
      </div>
    </div>
  );
}

export default AIAssistantPanel;
