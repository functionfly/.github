import { useEffect, useRef, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
  MessageCircle, 
  Send,
  Loader2,
  Pause,
  Play,
  Wifi,
  WifiOff,
  ArrowRight
} from 'lucide-react';
import { agentApi } from '@/api/agent';
import { toast } from 'sonner';

interface Message {
  id: string;
  fromAgentId: string;
  toAgentId: string;
  messageType: string;
  payload?: Record<string, unknown>;
  status: 'pending' | 'delivered' | 'read' | 'failed' | 'expired';
  createdAt: string;
  direction: 'incoming' | 'outgoing';
}

interface Agent {
  id: string;
  name: string;
  swarmRole?: string;
}

interface RealtimeMessageFlowProps {
  agentId: string;
  agentName: string;
  connectedAgents?: Agent[];
  maxMessages?: number;
  autoRefresh?: boolean;
}

const statusIcons = {
  pending: '⏳',
  delivered: '✓',
  read: '✓✓',
  failed: '✗',
  expired: '○',
};

const statusColors = {
  pending: 'bg-yellow-500',
  delivered: 'bg-blue-500',
  read: 'bg-green-500',
  failed: 'bg-red-500',
  expired: 'bg-gray-500',
};

export function RealtimeMessageFlow({
  agentId,
  agentName,
  connectedAgents = [],
  maxMessages = 100,
  autoRefresh = true,
}: RealtimeMessageFlowProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [isPaused, setIsPaused] = useState(false);
  const [selectedRecipient, setSelectedRecipient] = useState<string>('');
  const [composePayload, setComposePayload] = useState('');
  const [sending, setSending] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  // Load messages
  const loadMessages = useCallback(async () => {
    if (isPaused) return;

    try {
      const { messages: inboxMessages } = await agentApi.getInbox(agentId);
      
      // Transform and dedupe messages
      const transformed: Message[] = inboxMessages.map(m => ({
        id: m.id,
        fromAgentId: m.fromAgentId,
        toAgentId: m.toAgentId,
        messageType: m.messageType,
        payload: m.payload,
        status: (m.status === 'pending' || m.status === 'delivered' || m.status === 'read' || m.status === 'failed' || m.status === 'expired'
          ? m.status
          : 'pending') as Message['status'],
        createdAt: m.createdAt,
        direction: m.fromAgentId === agentId ? 'outgoing' : 'incoming',
      }));

      setMessages(prev => {
        const merged = [...transformed, ...prev.filter(p => 
          !transformed.some(t => t.id === p.id)
        )];
        return merged.slice(0, maxMessages);
      });
    } catch (err) {
      console.error('Failed to load messages:', err);
    } finally {
      setLoading(false);
    }
  }, [agentId, isPaused, maxMessages]);

  // Initial load and polling
  useEffect(() => {
    loadMessages();

    if (autoRefresh && !isPaused) {
      intervalRef.current = setInterval(loadMessages, 3000);
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [loadMessages, autoRefresh, isPaused]);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const handleSendMessage = async () => {
    if (!selectedRecipient || !composePayload.trim()) return;

    setSending(true);
    try {
      let payload: Record<string, unknown> = {};
      try {
        payload = JSON.parse(composePayload);
      } catch {
        payload = { message: composePayload };
      }

      await agentApi.sendMessage(agentId, {
        to_agent_id: selectedRecipient,
        message_type: 'direct_message',
        payload,
      });

      toast.success('Message sent');
      setComposePayload('');
      loadMessages(); // Refresh
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to send message');
    } finally {
      setSending(false);
    }
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const getAgentName = (id: string) => {
    if (id === agentId) return agentName;
    const agent = connectedAgents.find(a => a.id === id);
    return agent?.name || id.slice(0, 8) + '...';
  };

  const pendingCount = messages.filter(m => m.status === 'pending').length;
  const failedCount = messages.filter(m => m.status === 'failed').length;

  return (
    <Card className="h-[600px] flex flex-col">
      <CardHeader className="pb-3 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`p-2 rounded-full ${isPaused ? 'bg-yellow-100' : 'bg-green-100'}`}>
              {isPaused ? (
                <Pause className="h-4 w-4 text-yellow-600" />
              ) : (
                <Wifi className="h-4 w-4 text-green-600" />
              )}
            </div>
            <div>
              <CardTitle className="text-base flex items-center gap-2">
                Real-time Message Flow
                {pendingCount > 0 && (
                  <Badge variant="secondary" className="text-xs">
                    {pendingCount} pending
                  </Badge>
                )}
                {failedCount > 0 && (
                  <Badge variant="destructive" className="text-xs">
                    {failedCount} failed
                  </Badge>
                )}
              </CardTitle>
              <CardDescription className="text-xs">
                {messages.length} messages · {connectedAgents.length} connected agents
              </CardDescription>
            </div>
          </div>
          <Button
            variant="outline"
            size="icon"
            onClick={() => setIsPaused(!isPaused)}
          >
            {isPaused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col min-h-0 px-4">
        {/* Message list */}
        <ScrollArea ref={scrollRef} className="flex-1 -mx-2 px-2">
          {loading && messages.length === 0 ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : messages.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <MessageCircle className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">No messages yet</p>
            </div>
          ) : (
            <div className="space-y-3 py-2">
              {messages.map((msg, index) => {
                const showTime = index === 0 || 
                  new Date(msg.createdAt).getMinutes() !== 
                  new Date(messages[index - 1].createdAt).getMinutes();

                return (
                  <div key={msg.id}>
                    {showTime && (
                      <div className="text-center my-2">
                        <span className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                          {new Date(msg.createdAt).toLocaleDateString([], { 
                            month: 'short', 
                            day: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit'
                          })}
                        </span>
                      </div>
                    )}
                    
                    <div className={`flex ${msg.direction === 'outgoing' ? 'justify-end' : 'justify-start'}`}>
                      <div className={`max-w-[80%] rounded-lg px-3 py-2 text-sm ${
                        msg.direction === 'outgoing' 
                          ? 'bg-primary text-primary-foreground' 
                          : 'bg-muted'
                      }`}>
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-xs font-medium opacity-80">
                            {msg.direction === 'outgoing' ? 'You' : getAgentName(msg.fromAgentId)}
                          </span>
                          <span className="text-xs opacity-60">
                            {formatTime(msg.createdAt)}
                          </span>
                          <span className="text-xs opacity-60">
                            {statusIcons[msg.status]}
                          </span>
                        </div>
                        
                        <p className="text-xs opacity-90 mb-1">
                          {msg.messageType}
                        </p>
                        
                        {msg.payload && (
                          <pre className="text-xs opacity-75 overflow-x-auto max-w-full">
                            {JSON.stringify(msg.payload, null, 2).slice(0, 200)}
                            {JSON.stringify(msg.payload).length > 200 && '...'}
                          </pre>
                        )}
                        
                        {msg.direction === 'outgoing' && (
                          <div className="flex items-center gap-1 mt-1 text-xs opacity-60">
                            <ArrowRight className="h-3 w-3" />
                            {getAgentName(msg.toAgentId)}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </ScrollArea>

        {/* Compose message */}
        <div className="flex-shrink-0 border-t pt-3 mt-3 space-y-2">
          <div className="flex gap-2">
            <select
              value={selectedRecipient}
              onChange={(e) => setSelectedRecipient(e.target.value)}
              className="flex-1 h-9 rounded-md border border-input bg-background px-3 text-sm"
            >
              <option value="">Select recipient...</option>
              {connectedAgents.map(agent => (
                <option key={agent.id} value={agent.id}>
                  {agent.name} ({agent.swarmRole || 'agent'})
                </option>
              ))}
            </select>
          </div>
          
          <div className="flex gap-2">
            <textarea
              value={composePayload}
              onChange={(e) => setComposePayload(e.target.value)}
              placeholder='Enter JSON payload or plain message... e.g., {"task": "analyze", "data": {...}}'
              className="flex-1 min-h-[60px] max-h-[120px] rounded-md border border-input bg-background px-3 py-2 text-sm resize-y"
              onKeyDown={(e) => {
                if (e.key === 'Enter' && e.metaKey) {
                  handleSendMessage();
                }
              }}
            />
            <Button
              onClick={handleSendMessage}
              disabled={!selectedRecipient || !composePayload.trim() || sending}
              className="self-end"
            >
              {sending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
            </Button>
          </div>
          
          <p className="text-xs text-muted-foreground">
            Cmd+Enter to send · Messages auto-deliver to agent inbox
          </p>
        </div>
      </CardContent>
    </Card>
  );
}

export default RealtimeMessageFlow;
