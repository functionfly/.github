import * as React from 'react';
import { cn } from '@/lib/utils';
import { EmptyState, type EmptyStateProps } from './empty-state';
import {
  MessageSquare,
  Bug,
  Wrench,
  DollarSign,
  Shield,
  Building2,
  FunctionSquare,
} from 'lucide-react';
import type { ConversationType } from '@/api/conversations';

const TYPE_ICONS: Record<ConversationType, React.ComponentType<{ className?: string }>> = {
  dm: MessageSquare,
  function_thread: FunctionSquare,
  issue_thread: Bug,
  fix_mode: Wrench,
  bounty_thread: DollarSign,
  org_thread: Building2,
  security_disclosure: Shield,
};

const TYPE_TITLES: Record<ConversationType, string> = {
  dm: 'No messages yet',
  function_thread: 'Start a function discussion',
  issue_thread: 'No issues reported yet',
  fix_mode: 'Describe the problem to begin',
  bounty_thread: 'No bounty activity yet',
  org_thread: 'Organization thread is quiet',
  security_disclosure: 'No disclosures yet',
};

const TYPE_DESCRIPTIONS: Record<ConversationType, string> = {
  dm: 'Start the conversation by sending a message below.',
  function_thread: 'Discuss this function with collaborators, share ideas, and track progress.',
  issue_thread: 'Report an issue to get help from the community.',
  fix_mode: 'Describe what\u2019s broken and the AI will help you find a fix.',
  bounty_thread: 'Offer a bounty to incentivize community solutions.',
  org_thread: 'Coordinate with your organization members here.',
  security_disclosure: 'Report security findings responsibly.',
};

export interface EmptyChatProps extends Omit<EmptyStateProps, 'icon' | 'title'> {
  title?: string;
  icon?: React.ReactNode;
  conversationType?: ConversationType;
}

export function EmptyChat({
  title,
  icon,
  description,
  conversationType,
  ...props
}: EmptyChatProps) {
  const TypeIcon = conversationType ? TYPE_ICONS[conversationType] : MessageSquare;
  const resolvedTitle = title ?? (conversationType ? TYPE_TITLES[conversationType] : 'No messages yet');
  const resolvedDescription = description ?? (conversationType ? TYPE_DESCRIPTIONS[conversationType] : 'Start the conversation by sending a message below.');

  return (
    <EmptyState
      icon={icon ?? <TypeIcon className="h-8 w-8" />}
      title={resolvedTitle}
      description={resolvedDescription}
      variant="ghost"
      size="full"
      {...props}
    />
  );
}
