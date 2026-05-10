import { useState } from 'react';
import { motion } from 'framer-motion';
import { RefreshCw, Unplug, Shield, Clock } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { cn, formatDistanceToNow } from '@/lib/utils';
import {
  useGitHubConnection,
  useGitHubDisconnect,
  useGitHubTokenRefresh,
} from '@/hooks/useGitHubConnection';
import type { GitHubConnection } from '@/types/github';

function getStatusColor(status: GitHubConnection['status']) {
  switch (status) {
    case 'active':
      return 'bg-emerald-500';
    case 'expired':
      return 'bg-red-500';
    case 'revoked':
      return 'bg-red-500';
    case 'error':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
}

function getStatusLabel(status: GitHubConnection['status']) {
  switch (status) {
    case 'active':
      return 'Active';
    case 'expired':
      return 'Expired';
    case 'revoked':
      return 'Revoked';
    case 'error':
      return 'Error';
    default:
      return 'Unknown';
  }
}

function getTokenExpiryColor(expiresAt: string | null): string {
  if (!expiresAt) return 'text-text-secondary';
  const daysUntilExpiry = (new Date(expiresAt).getTime() - Date.now()) / (1000 * 60 * 60 * 24);
  if (daysUntilExpiry < 0) return 'text-red-500';
  if (daysUntilExpiry < 7) return 'text-amber-500';
  return 'text-emerald-500';
}

interface GitHubConnectionCardProps {
  className?: string;
}

export function GitHubConnectionCard({ className }: GitHubConnectionCardProps) {
  const { data: connection, isLoading } = useGitHubConnection();
  const disconnectMutation = useGitHubDisconnect();
  const refreshMutation = useGitHubTokenRefresh();
  const [showDisconnectDialog, setShowDisconnectDialog] = useState(false);

  if (isLoading) {
    return (
      <Card className={cn('animate-pulse', className)}>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-muted" />
            <div className="space-y-2">
              <div className="h-4 w-32 rounded bg-muted" />
              <div className="h-3 w-24 rounded bg-muted" />
            </div>
          </div>
        </CardHeader>
      </Card>
    );
  }

  if (!connection) return null;

  const scopes = connection.token_scope?.split(',').map((s) => s.trim()).filter(Boolean) ?? [];
  const isExpiringSoon =
    connection.token_expires_at &&
    (new Date(connection.token_expires_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24) < 7;

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <Card className={cn('overflow-hidden', className)}>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="relative">
                <Avatar className="h-10 w-10 border-2 border-border-subtle">
                  <AvatarImage src={connection.github_avatar_url ?? undefined} alt={connection.github_username} />
                  <AvatarFallback>
                    <svg role="img" viewBox="0 0 24 24" className="h-5 w-5" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
                  </AvatarFallback>
                </Avatar>
                <span
                  className={cn(
                    'absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full border-2 border-card',
                    getStatusColor(connection.status)
                  )}
                  aria-label={`Status: ${getStatusLabel(connection.status)}`}
                />
              </div>
              <div>
                <CardTitle className="text-base flex items-center gap-2">
                  {connection.github_username}
                  <Badge variant="secondary" className="text-xs">
                    <svg role="img" viewBox="0 0 24 24" className="h-3 w-3 mr-1" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
                    GitHub
                  </Badge>
                </CardTitle>
                <p className="text-xs text-text-muted mt-0.5">
                  Connected {formatDistanceToNow(connection.connected_at, { addSuffix: true })}
                </p>
              </div>
            </div>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          {scopes.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {scopes.map((scope) => (
                <Badge key={scope} variant="outline" className="text-xs font-mono">
                  {scope}
                </Badge>
              ))}
            </div>
          )}

          {connection.token_expires_at && (
            <div className="flex items-center gap-2 text-xs">
              <Clock className="h-3.5 w-3.5 text-text-muted" />
              <span className="text-text-secondary">Token expires:</span>
              <span className={cn('font-medium', getTokenExpiryColor(connection.token_expires_at))}>
                {isExpiringSoon
                  ? formatDistanceToNow(connection.token_expires_at, { addSuffix: true })
                  : new Date(connection.token_expires_at).toLocaleDateString()}
              </span>
            </div>
          )}

          <div className="flex items-center gap-2 pt-2 border-t border-border-subtle">
            <Button
              variant="outline"
              size="sm"
              onClick={() => refreshMutation.mutate()}
              disabled={refreshMutation.isPending}
              aria-label="Refresh GitHub token"
            >
              <RefreshCw className={cn('h-3.5 w-3.5 mr-1.5', refreshMutation.isPending && 'animate-spin')} />
              Refresh Token
            </Button>

            <AlertDialog open={showDisconnectDialog} onOpenChange={setShowDisconnectDialog}>
              <AlertDialogTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-red-500 hover:text-red-600 hover:bg-red-500/10"
                  aria-label="Disconnect GitHub account"
                >
                  <Unplug className="h-3.5 w-3.5 mr-1.5" />
                  Disconnect
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Disconnect GitHub?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will revoke access to your GitHub account. Auto-sync for imported functions will stop
                    working. You can reconnect at any time.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    className="bg-red-500 hover:bg-red-600 text-white"
                    onClick={() => disconnectMutation.mutate()}
                  >
                    <Unplug className="h-3.5 w-3.5 mr-1.5" />
                    Disconnect
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>

            <div className="flex-1" />

            <div className="flex items-center gap-1 text-xs text-text-muted">
              <Shield className="h-3 w-3" />
              AES-256 encrypted
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
