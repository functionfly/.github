import { type ConversationMessage, conversationsApi } from '@/api/conversations';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { useQuery } from '@tanstack/react-query';
import { formatDistanceToNow } from 'date-fns';
import { Loader2, Search, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

export interface MessageSearchProps {
  username: string;
  conversationId?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function MessageSearch({ username, conversationId, open, onOpenChange }: MessageSearchProps) {
  const [query, setQuery] = useState('');
  const [submittedQuery, setSubmittedQuery] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (open) {
      setQuery('');
      setSubmittedQuery('');
      requestAnimationFrame(() => inputRef.current?.focus());
    }
  }, [open]);

  const { data, isFetching } = useQuery({
    queryKey: ['message-search', username, submittedQuery, conversationId],
    queryFn: () =>
      conversationsApi.searchMessages(username, {
        q: submittedQuery,
        conversation_id: conversationId,
        limit: 30,
      }),
    enabled: submittedQuery.length >= 2,
    staleTime: 10_000,
  });

  const results = data?.messages ?? [];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = query.trim();
    if (trimmed.length >= 2) setSubmittedQuery(trimmed);
  };

  const handleResultClick = (msg: ConversationMessage) => {
    onOpenChange(false);
    navigate(`/u/${username}/conversations/${msg.conversation_id}`);
  };

  const highlightMatch = (content: string, q: string) => {
    if (!q) return content;
    const idx = content.toLowerCase().indexOf(q.toLowerCase());
    if (idx === -1) return content.slice(0, 200);
    const start = Math.max(0, idx - 40);
    const end = Math.min(content.length, idx + q.length + 80);
    const snippet = (start > 0 ? '...' : '') + content.slice(start, end) + (end < content.length ? '...' : '');
    return (
      <>
        {snippet.slice(0, idx - start + (start > 0 ? 3 : 0))}
        <mark className="bg-brand-500/30 rounded px-0.5">
          {snippet.slice(idx - start + (start > 0 ? 3 : 0), idx - start + (start > 0 ? 3 : 0) + q.length)}
        </mark>
        {snippet.slice(idx - start + (start > 0 ? 3 : 0) + q.length)}
      </>
    );
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] bg-black/50">
      <div className="w-full max-w-lg mx-4 bg-card border border-border rounded-lg shadow-xl overflow-hidden">
        <form onSubmit={handleSubmit} className="flex items-center gap-2 p-3 border-b border-border">
          <Search className="h-4 w-4 text-muted-foreground shrink-0" />
          <Input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={conversationId ? 'Search in conversation...' : 'Search all messages...'}
            className="flex-1 border-0 shadow-none focus-visible:ring-0 h-8"
          />
          {isFetching && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={() => onOpenChange(false)}
          >
            <X className="h-4 w-4" />
          </Button>
        </form>

        {submittedQuery.length >= 2 && !isFetching && results.length === 0 && (
          <div className="p-6 text-center text-sm text-muted-foreground">
            No messages found for &ldquo;{submittedQuery}&rdquo;
          </div>
        )}

        {results.length > 0 && (
          <ScrollArea className="max-h-[50vh]">
            <ul className="py-1">
              {results.map((msg) => (
                <li key={msg.id}>
                  <button
                    className={cn(
                      'w-full text-left px-4 py-3 hover:bg-muted/60 transition-colors',
                      'focus:bg-muted/60 focus:outline-none'
                    )}
                    onClick={() => handleResultClick(msg)}
                  >
                    <div className="flex items-center justify-between gap-2 mb-1">
                      <span className="text-xs font-medium text-muted-foreground">
                        {msg.author_id.slice(0, 8)}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(new Date(msg.created_at), { addSuffix: true })}
                      </span>
                    </div>
                    <p className="text-sm line-clamp-2">
                      {highlightMatch(msg.content, submittedQuery)}
                    </p>
                  </button>
                </li>
              ))}
            </ul>
          </ScrollArea>
        )}

        {!submittedQuery && (
          <div className="p-6 text-center text-sm text-muted-foreground">
            Type at least 2 characters and press Enter to search
          </div>
        )}
      </div>
    </div>
  );
}
