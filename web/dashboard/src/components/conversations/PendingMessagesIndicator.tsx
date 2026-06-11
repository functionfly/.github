import { cn } from '@/lib/utils';
import { useOfflineMessage, type PendingMessageIndicator } from '@/hooks/useOfflineMessage';
import { AlertCircle, Loader2, RefreshCw, X } from 'lucide-react';

export interface PendingMessagesIndicatorProps {
  className?: string;
}

export function PendingMessagesIndicator({ className }: PendingMessagesIndicatorProps) {
  const { pendingMessages, retryMessage, dismissFailedMessage, isOnline } = useOfflineMessage();

  if (pendingMessages.length === 0) return null;

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      {!isOnline && (
        <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/50 px-3 py-2 rounded-lg">
          <AlertCircle className="h-3 w-3" />
          <span>You are offline</span>
        </div>
      )}

      {pendingMessages.map((message) => (
        <PendingMessageItem
          key={message.id}
          message={message}
          onRetry={() => retryMessage(message.id)}
          onDismiss={() => dismissFailedMessage(message.id)}
        />
      ))}
    </div>
  );
}

interface PendingMessageItemProps {
  message: PendingMessageIndicator;
  onRetry: () => void;
  onDismiss: () => void;
}

function PendingMessageItem({ message, onRetry, onDismiss }: PendingMessageItemProps) {
  return (
    <div
      className={cn(
        'flex items-center gap-2 px-3 py-2 rounded-lg border text-sm',
        message.status === 'pending' && 'bg-muted/50 border-border/60',
        message.status === 'sending' && 'bg-muted/50 border-border/60',
        message.status === 'failed' && 'bg-destructive/10 border-destructive/30'
      )}
    >
      {message.status === 'pending' && (
        <>
          <div className="h-2 w-2 rounded-full bg-muted-foreground animate-pulse" />
          <span className="flex-1 truncate text-muted-foreground">{message.content}</span>
          <span className="text-xs text-muted-foreground">Pending</span>
        </>
      )}

      {message.status === 'sending' && (
        <>
          <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />
          <span className="flex-1 truncate text-muted-foreground">{message.content}</span>
          <span className="text-xs text-muted-foreground">Sending...</span>
        </>
      )}

      {message.status === 'failed' && (
        <>
          <AlertCircle className="h-3 w-3 text-destructive" />
          <span className="flex-1 truncate">{message.content}</span>
          <button
            onClick={onRetry}
            className="p-1 hover:bg-accent rounded"
            title="Retry"
          >
            <RefreshCw className="h-3 w-3" />
          </button>
          <button
            onClick={onDismiss}
            className="p-1 hover:bg-accent rounded"
            title="Dismiss"
          >
            <X className="h-3 w-3" />
          </button>
        </>
      )}
    </div>
  );
}