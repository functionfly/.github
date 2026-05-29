import type { Conversation, ConversationType } from '@/api/conversations';
import type { FixModeMetadata } from './FixModeLayout';
import {
  format,
  isToday,
  isYesterday,
  isSameDay,
} from 'date-fns';

export type InboxFilter = 'all' | 'unread' | 'dm' | 'functions' | 'issues' | 'bounties' | 'resolved';

export const INBOX_FILTERS: { id: InboxFilter; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'unread', label: 'Unread' },
  { id: 'dm', label: 'DMs' },
  { id: 'functions', label: 'Functions' },
  { id: 'issues', label: 'Issues' },
  { id: 'bounties', label: 'Bounties' },
  { id: 'resolved', label: 'Resolved' },
];

export interface TypeBadgeMeta {
  label: string;
  className: string;
}

export const TYPE_BADGES: Record<ConversationType, TypeBadgeMeta> = {
  dm: { label: 'DM', className: 'conv-badge-dm' },
  function_thread: { label: 'Function', className: 'conv-badge-function' },
  issue_thread: { label: 'Issue', className: 'conv-badge-issue' },
  fix_mode: { label: 'Fix', className: 'conv-badge-fix' },
  bounty_thread: { label: 'Bounty', className: 'conv-badge-bounty' },
  org_thread: { label: 'Org', className: 'conv-badge-org' },
  security_disclosure: { label: 'Security', className: 'conv-badge-security' },
};

export function getTypeBadge(type: ConversationType): TypeBadgeMeta {
  return TYPE_BADGES[type] ?? { label: type, className: 'conv-badge-dm' };
}

export function getConversationTitle(
  conversation: Conversation,
  displayForParticipantId: (id: string) => string,
  currentUserId?: string
): string {
  const meta = (conversation.metadata ?? {}) as Record<string, unknown> & FixModeMetadata;

  if (conversation.type === 'function_thread') {
    const fn = meta.function_ref as { author?: string; name?: string } | undefined;
    if (fn?.author && fn?.name) return `${fn.author}/${fn.name}`;
    if (typeof meta.function === 'string') return meta.function;
  }

  if (conversation.type === 'fix_mode' && meta.problem_statement) {
    const short = meta.problem_statement.trim().slice(0, 60);
    return short.length < meta.problem_statement.trim().length ? `${short}…` : short;
  }

  if (conversation.type === 'issue_thread' && typeof meta.title === 'string') {
    return meta.title;
  }

  const ids = conversation.participant_ids ?? [];
  const self = currentUserId?.toLowerCase();
  const others = ids.filter((id) => id.toLowerCase() !== self);
  const labels = (others.length > 0 ? others : ids).map(displayForParticipantId);
  if (labels.every((l) => l === '…')) return `${ids.length} participant(s)`;
  if (labels.length <= 3) return labels.join(', ');
  return `${labels.slice(0, 2).join(', ')} +${labels.length - 2}`;
}

export function filterConversations(
  conversations: Conversation[],
  filter: InboxFilter,
  query: string,
  displayForParticipantId: (id: string) => string,
  currentUserId?: string
): { active: Conversation[]; resolved: Conversation[] } {
  const q = query.trim().toLowerCase();
  let list = conversations;

  switch (filter) {
    case 'unread':
      list = list.filter((c) => (c.unread_count ?? 0) > 0 && !c.resolved_at);
      break;
    case 'dm':
      list = list.filter((c) => c.type === 'dm' && !c.resolved_at);
      break;
    case 'functions':
      list = list.filter((c) => c.type === 'function_thread' && !c.resolved_at);
      break;
    case 'issues':
      list = list.filter((c) => (c.type === 'issue_thread' || c.type === 'fix_mode') && !c.resolved_at);
      break;
    case 'bounties':
      list = list.filter((c) => c.type === 'bounty_thread' && !c.resolved_at);
      break;
    case 'resolved':
      return { active: [], resolved: list.filter((c) => c.resolved_at) };
    case 'all':
    default:
      list = list.filter((c) => !c.resolved_at);
      break;
  }

  if (q) {
    list = list.filter((c) => {
      const title = getConversationTitle(c, displayForParticipantId, currentUserId).toLowerCase();
      const preview = (c.last_message_preview ?? '').toLowerCase();
      const type = c.type.replace(/_/g, ' ');
      return title.includes(q) || preview.includes(q) || type.includes(q);
    });
  }

  const resolved =
    filter === 'all' || filter === 'resolved'
      ? conversations.filter((c) => c.resolved_at)
      : [];

  if (filter === 'resolved') {
    if (q) {
      const filtered = resolved.filter((c) => {
        const title = getConversationTitle(c, displayForParticipantId, currentUserId).toLowerCase();
        return title.includes(q) || (c.last_message_preview ?? '').toLowerCase().includes(q);
      });
      return { active: [], resolved: filtered };
    }
    return { active: [], resolved };
  }

  return { active: list, resolved: filter === 'all' ? resolved : [] };
}

export function formatMessageDateLabel(date: Date): string {
  if (isToday(date)) return 'Today';
  if (isYesterday(date)) return 'Yesterday';
  return format(date, 'MMMM d, yyyy');
}

export type MessageGroup = { label: string; messages: { id: string; created_at: string }[] };

export function groupMessagesByDate<T extends { id: string; created_at: string }>(
  messages: T[]
): { label: string; messages: T[] }[] {
  const groups: { label: string; messages: T[] }[] = [];
  let currentLabel = '';
  for (const m of messages) {
    const d = new Date(m.created_at);
    const label = formatMessageDateLabel(d);
    if (label !== currentLabel) {
      groups.push({ label, messages: [m] });
      currentLabel = label;
    } else {
      groups[groups.length - 1].messages.push(m);
    }
  }
  return groups;
}

export function shouldShowDateSeparator<T extends { created_at: string }>(
  messages: T[],
  index: number
): string | null {
  if (index === 0) return formatMessageDateLabel(new Date(messages[0].created_at));
  const prev = new Date(messages[index - 1].created_at);
  const curr = new Date(messages[index].created_at);
  if (!isSameDay(prev, curr)) return formatMessageDateLabel(curr);
  return null;
}

export function parseSlashCommand(input: string): {
  command: string | null;
  rest: string;
} {
  const trimmed = input.trim();
  if (!trimmed.startsWith('/')) return { command: null, rest: input };
  const space = trimmed.indexOf(' ');
  if (space === -1) return { command: trimmed.slice(1).toLowerCase(), rest: '' };
  return {
    command: trimmed.slice(1, space).toLowerCase(),
    rest: trimmed.slice(space + 1),
  };
}

export function getPrimaryParticipantId(
  participantIds: string[] | undefined,
  currentUserId?: string
): string | undefined {
  if (!participantIds?.length) return undefined;
  const self = currentUserId?.toLowerCase();
  const other = participantIds.find((id) => id.toLowerCase() !== self);
  return other ?? participantIds[0];
}

export function getFunctionContextFromConversation(conversation: Conversation): {
  author?: string;
  name?: string;
} | null {
  const meta = (conversation.metadata ?? {}) as Record<string, unknown>;
  const fn = meta.function_ref as { author?: string; name?: string } | undefined;
  if (fn?.author && fn?.name) return fn;
  if (typeof meta.function === 'string' && meta.function.includes('/')) {
    const [author, name] = meta.function.split('/');
    if (author && name) return { author, name };
  }
  return null;
}
