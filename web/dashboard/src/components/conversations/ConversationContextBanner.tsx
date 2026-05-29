import type { Conversation } from '@/api/conversations';
import { Button } from '@/components/ui/button';
import { ExternalLink, FunctionSquare } from 'lucide-react';
import { Link } from 'react-router-dom';
import { getFunctionContextFromConversation } from './conversation-ui';

interface ConversationContextBannerProps {
  conversation: Conversation;
  executionCount?: number;
}

export function ConversationContextBanner({
  conversation,
  executionCount = 0,
}: ConversationContextBannerProps) {
  const fn = getFunctionContextFromConversation(conversation);
  if (!fn?.author || !fn?.name) return null;

  return (
    <div className="conv-context-banner">
      <FunctionSquare className="h-4 w-4 shrink-0 text-primary" />
      <span className="font-mono text-sm truncate">
        {fn.author}/{fn.name}
      </span>
      {executionCount > 0 && (
        <span className="text-muted-foreground text-xs shrink-0">
          {executionCount} execution{executionCount === 1 ? '' : 's'} in thread
        </span>
      )}
      <Button variant="ghost" size="sm" className="ml-auto h-7 text-xs shrink-0" asChild>
        <Link to={`/fx/${fn.author}/${fn.name}`}>
          View function
          <ExternalLink className="h-3 w-3 ml-1" />
        </Link>
      </Button>
    </div>
  );
}
