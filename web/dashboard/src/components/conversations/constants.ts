import type { ConversationType } from '@/api/conversations';

export const CONVERSATION_TYPES: { value: ConversationType; label: string }[] = [
  { value: 'dm', label: 'Direct message' },
  { value: 'function_thread', label: 'Function thread' },
  { value: 'issue_thread', label: 'Issue thread' },
  { value: 'fix_mode', label: 'Fix mode' },
  { value: 'bounty_thread', label: 'Bounty thread' },
  { value: 'org_thread', label: 'Org thread' },
  { value: 'security_disclosure', label: 'Security disclosure' },
];

export const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export function normalizeParticipantHandle(s: string): string {
  const t = s.trim();
  if (t.startsWith('@')) return t.slice(1).trim();
  return t;
}

export function participantSegmentAtCaret(full: string, caret: number): string {
  const left = full.slice(0, Math.min(Math.max(caret, 0), full.length));
  const lastComma = left.lastIndexOf(',');
  const raw = lastComma === -1 ? left : left.slice(lastComma + 1);
  return normalizeParticipantHandle(raw);
}

export function applyParticipantUsernamePick(
  full: string,
  caret: number,
  username: string
): { value: string; caret: number } {
  const left = full.slice(0, caret);
  const lastComma = left.lastIndexOf(',');
  const prefix = lastComma === -1 ? '' : left.slice(0, lastComma + 1);
  const suffix = full.slice(caret);
  let head: string;
  if (!prefix) {
    head = `${username}, `;
  } else {
    const needsSpace = !/,\s*$/.test(prefix);
    head = (needsSpace ? `${prefix} ` : prefix) + `${username}, `;
  }
  return { value: head + suffix, caret: head.length };
}

export function formatParticipantLine(
  participantIds: string[] | undefined,
  currentUserId: string | undefined,
  displayFor: (id: string) => string
): string {
  if (!participantIds?.length) return 'Conversation';
  const self = currentUserId?.toLowerCase();
  const others = participantIds.filter((id) => id.toLowerCase() !== self);
  const idsToShow = others.length > 0 ? others : participantIds;
  const labels = idsToShow.map(displayFor);
  if (labels.every((l) => l === '…')) {
    return `${participantIds.length} participant(s)`;
  }
  if (labels.length <= 4) return labels.join(' · ');
  return `${labels.slice(0, 4).join(' · ')} +${labels.length - 4}`;
}
