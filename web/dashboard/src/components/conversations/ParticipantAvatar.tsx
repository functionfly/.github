import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';

export interface UserLookupEntry {
  id: string;
  username?: string;
  name?: string;
  avatar?: string;
}

export interface ParticipantAvatarProps {
  user?: UserLookupEntry;
  userId?: string;
  displayName?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const SIZE_CLASS = {
  sm: 'h-7 w-7 text-[10px]',
  md: 'h-9 w-9 text-xs',
  lg: 'h-11 w-11 text-sm',
};

export function ParticipantAvatar({
  user,
  userId,
  displayName,
  size = 'md',
  className,
}: ParticipantAvatarProps) {
  const label =
    displayName?.replace(/^@/, '') ||
    user?.name ||
    user?.username ||
    userId?.slice(0, 2) ||
    '?';
  const initials = label.slice(0, 2).toUpperCase();

  return (
    <Avatar className={cn(SIZE_CLASS[size], className)}>
      {user?.avatar ? <AvatarImage src={user.avatar} alt={label} /> : null}
      <AvatarFallback className="bg-muted font-medium">{initials}</AvatarFallback>
    </Avatar>
  );
}
