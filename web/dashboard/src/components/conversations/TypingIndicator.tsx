import { TypingIndicator as UiTypingIndicator } from '@/components/ui/typing-indicator';

export interface TypingIndicatorProps {
  typingUsers: Record<string, boolean>;
  displayForParticipantId: (id: string) => string;
}

export function TypingIndicator({
  typingUsers,
  displayForParticipantId,
}: TypingIndicatorProps) {
  const userIds = Object.keys(typingUsers);
  if (userIds.length === 0) return null;

  const names = userIds.map(displayForParticipantId);
  return <UiTypingIndicator users={names} />;
}
