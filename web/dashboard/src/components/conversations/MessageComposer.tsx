import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Loader2, Play, Send } from 'lucide-react';
import { useEffect, useRef } from 'react';

export interface MessageComposerProps {
  messageDraft: string;
  onDraftChange: (value: string) => void;
  onSend: () => void;
  sendPending: boolean;
  showRunPanel: boolean;
  onToggleRunPanel: () => void;
  onTyping?: (typing: boolean) => void;
}

export function MessageComposer({
  messageDraft,
  onDraftChange,
  onSend,
  sendPending,
  showRunPanel,
  onToggleRunPanel,
  onTyping,
}: MessageComposerProps) {
  const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isTypingRef = useRef(false);

  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    };
  }, []);

  const handleChange = (value: string) => {
    onDraftChange(value);

    if (!onTyping) return;

    if (value.length > 0 && !isTypingRef.current) {
      isTypingRef.current = true;
      onTyping(true);
    } else if (value.length === 0 && isTypingRef.current) {
      isTypingRef.current = false;
      onTyping(false);
    }

    if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
    if (value.length > 0) {
      typingTimeoutRef.current = setTimeout(() => {
        isTypingRef.current = false;
        onTyping(false);
      }, 3000);
    }
  };
  return (
    <div className="border-t border-border p-3 flex gap-2">
      <Button
        variant="ghost"
        size="icon"
        className="shrink-0"
        onClick={onToggleRunPanel}
        title={showRunPanel ? 'Hide Run panel' : 'Run in thread'}
      >
        <Play className="h-4 w-4" />
      </Button>
      <Input
        placeholder="Type a message…"
        value={messageDraft}
        onChange={(e) => handleChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
            isTypingRef.current = false;
            onTyping?.(false);
            onSend();
          }
        }}
        className="flex-1"
      />
      <Button
        size="icon"
        onClick={onSend}
        disabled={!messageDraft.trim() || sendPending}
      >
        {sendPending ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Send className="h-4 w-4" />
        )}
      </Button>
    </div>
  );
}
