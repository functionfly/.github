import { ChatInput } from '@/components/ui/chat-input';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Coins, MoreHorizontal, Play, Search, CheckCircle } from 'lucide-react';
import { RunInThreadPanel } from './RunInThreadPanel';
import { parseSlashCommand } from './conversation-ui';

export interface ConversationThreadComposerProps {
  value: string;
  onChange: (value: string) => void;
  onSend: (content: string) => void;
  onTyping?: (typing: boolean) => void;
  pending?: boolean;
  disabled?: boolean;
  username: string;
  conversationId: string;
  showRunPanel: boolean;
  onToggleRunPanel: () => void;
  onSearch: () => void;
  onResolve: () => void;
  onBounty: () => void;
  resolvePending?: boolean;
  isResolved?: boolean;
  defaultAuthor?: string;
  defaultName?: string;
  onSlashCommand?: (command: string, rest: string) => boolean;
}

export function ConversationThreadComposer({
  value,
  onChange,
  onSend,
  onTyping,
  pending,
  disabled,
  username,
  conversationId,
  showRunPanel,
  onToggleRunPanel,
  onSearch,
  onResolve,
  onBounty,
  resolvePending,
  isResolved,
  defaultAuthor,
  defaultName,
  onSlashCommand,
}: ConversationThreadComposerProps) {
  const handleSend = () => {
    const trimmed = value.trim();
    if (!trimmed) return;

    const { command, rest } = parseSlashCommand(trimmed);
    if (command && onSlashCommand?.(command, rest)) {
      onChange('');
      return;
    }

    onSend(trimmed);
  };

  return (
    <div className="conv-composer-bar">
      {showRunPanel && (
        <div className="mb-2 rounded-lg border border-border bg-muted/20 p-3">
          <RunInThreadPanel
            username={username}
            conversationId={conversationId}
            defaultAuthor={defaultAuthor}
            defaultName={defaultName}
            onSnippetAdded={() => onToggleRunPanel()}
          />
        </div>
      )}
      <div className="conv-composer-toolbar">
        <Button
          type="button"
          variant={showRunPanel ? 'secondary' : 'ghost'}
          size="sm"
          className="h-7 px-2 text-xs"
          onClick={onToggleRunPanel}
        >
          <Play className="h-3.5 w-3.5 mr-1" />
          Run
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-xs"
          onClick={onBounty}
        >
          <Coins className="h-3.5 w-3.5 mr-1" />
          Bounty
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-xs"
          onClick={onSearch}
        >
          <Search className="h-3.5 w-3.5 mr-1" />
          Search
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button type="button" variant="ghost" size="sm" className="h-7 w-7 p-0 ml-auto">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {!isResolved && (
              <DropdownMenuItem onClick={onResolve} disabled={resolvePending}>
                <CheckCircle className="h-4 w-4 mr-2" />
                Resolve thread
              </DropdownMenuItem>
            )}
            <DropdownMenuItem onClick={onSearch}>
              <Search className="h-4 w-4 mr-2" />
              Search messages
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <ChatInput
        value={value}
        onChange={onChange}
        onSend={handleSend}
        onTyping={onTyping}
        pending={pending}
        disabled={disabled || isResolved}
        placeholder={
          isResolved
            ? 'Thread resolved — start a new conversation to continue'
            : 'Message… (/run, /bounty, /resolve, /search)'
        }
        showMarkdownPreview
      />
    </div>
  );
}
