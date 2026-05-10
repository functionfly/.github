import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
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
import {
  useGitHubConnection,
  useGitHubConnect,
  useGitHubDisconnect,
} from '@/hooks/useGitHubConnection';
import { useGitHubStore } from '@/stores/githubStore';
import { motion } from 'framer-motion';
import {
  AlertTriangle,
  Building2,
  Globe,
  Key,
  Lock,
  RefreshCw,
  Settings,
  Shield,
  Unplug,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

export function GitHubSettingsPage() {
  const { t } = useTranslation();
  const { data: connection, isLoading } = useGitHubConnection();
  const connectMutation = useGitHubConnect();
  const disconnectMutation = useGitHubDisconnect();
  const setConnection = useGitHubStore((s) => s.setConnection);

  const [defaultVisibility, setDefaultVisibility] = useState('private');
  const [autoSync, setAutoSync] = useState(false);
  const [syncBranches, setSyncBranches] = useState('main');

  useEffect(() => {
    setConnection(connection ?? null);
  }, [connection, setConnection]);

  const isConnected = connection && connection.status === 'active';

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-40 rounded-lg" />
        <Skeleton className="h-32 rounded-lg" />
        <Skeleton className="h-32 rounded-lg" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Connection Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-bg-secondary border border-border-default flex items-center justify-center">
                <svg role="img" viewBox="0 0 24 24" className="w-5 h-5 text-text-primary" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
              </div>
              <div>
                <CardTitle className="text-base">GitHub Connection</CardTitle>
                <CardDescription>
                  {isConnected
                    ? `Connected as @${connection.github_username}`
                    : 'Not connected'}
                </CardDescription>
              </div>
            </div>
            {isConnected ? (
              <Badge variant="default" className="bg-green-500/20 text-green-500 border-green-500/30">
                Active
              </Badge>
            ) : (
              <Badge variant="outline">Disconnected</Badge>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {isConnected ? (
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                {connection.github_avatar_url && (
                  <img
                    src={connection.github_avatar_url}
                    alt={connection.github_username}
                    className="w-12 h-12 rounded-full border border-border-default"
                  />
                )}
                <div>
                  <p className="font-medium text-text-primary">{connection.github_username}</p>
                  <p className="text-xs text-text-muted">
                    Connected {new Date(connection.connected_at).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => connectMutation.mutate()}
                  disabled={connectMutation.isPending}
                >
                  <RefreshCw className="w-4 h-4 mr-2" />
                  Reconnect
                </Button>
              </div>
            </div>
          ) : (
            <div className="text-center py-4">
              <p className="text-text-secondary mb-4">
                Connect your GitHub account to import functions from repositories
              </p>
              <Button
                onClick={() => connectMutation.mutate()}
                disabled={connectMutation.isPending}
                className="gap-2"
              >
                <svg role="img" viewBox="0 0 24 24" className="w-4 h-4" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
                {connectMutation.isPending ? 'Connecting...' : 'Connect GitHub'}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* GitHub App (Enterprise) */}
      <Card className="opacity-60">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-bg-secondary border border-border-default flex items-center justify-center">
              <Building2 className="w-5 h-5 text-text-muted" />
            </div>
            <div>
              <CardTitle className="text-base flex items-center gap-2">
                GitHub App
                <Badge variant="outline" className="text-xs">
                  Coming Soon
                </Badge>
              </CardTitle>
              <CardDescription>
                Install the FunctionFly GitHub App for organization-wide access
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-text-muted">
            Enterprise GitHub App integration will allow organization admins to manage imports across
            all repositories with fine-grained permissions.
          </p>
        </CardContent>
      </Card>

      {/* Import Defaults */}
      {isConnected && (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-bg-secondary border border-border-default flex items-center justify-center">
                <Settings className="w-5 h-5 text-text-primary" />
              </div>
              <div>
                <CardTitle className="text-base">Import Defaults</CardTitle>
                <CardDescription>
                  Default settings for new GitHub imports
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>Default Visibility</Label>
              <Select value={defaultVisibility} onValueChange={setDefaultVisibility}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="public">
                    <div className="flex items-center gap-2">
                      <Globe className="w-4 h-4" />
                      Public
                    </div>
                  </SelectItem>
                  <SelectItem value="private">
                    <div className="flex items-center gap-2">
                      <Lock className="w-4 h-4" />
                      Private
                    </div>
                  </SelectItem>
                  <SelectItem value="unlisted">Unlisted</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Auto-sync by default</Label>
                <p className="text-xs text-text-muted">
                  Enable auto-sync for new imports
                </p>
              </div>
              <Switch checked={autoSync} onCheckedChange={setAutoSync} />
            </div>

            <div className="space-y-2">
              <Label>Default Sync Branches</Label>
              <Input
                placeholder="main"
                value={syncBranches}
                onChange={(e) => setSyncBranches(e.target.value)}
              />
              <p className="text-xs text-text-muted">
                Comma-separated list of branches to sync by default
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Danger Zone */}
      {isConnected && (
        <Card className="border-red-500/20">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-red-500/10 border border-red-500/20 flex items-center justify-center">
                <AlertTriangle className="w-5 h-5 text-red-500" />
              </div>
              <div>
                <CardTitle className="text-base text-red-500">Danger Zone</CardTitle>
                <CardDescription>
                  Irreversible actions for your GitHub connection
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive" size="sm" className="gap-2">
                  <Unplug className="w-4 h-4" />
                  Disconnect GitHub
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Disconnect GitHub?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will revoke FunctionFly's access to your GitHub account. Existing imports
                    will stop syncing, but deployed functions will remain active. You can reconnect
                    at any time.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    onClick={() => disconnectMutation.mutate()}
                    className="bg-red-500 hover:bg-red-600"
                  >
                    Disconnect
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
