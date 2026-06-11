import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Skeleton } from '@/components/ui/skeleton';
import { PageHeader } from '@/components/layout/PageHeader';
import { PageLayout } from '@/components/layout/PageLayout';
import { useGitHubConnection, useGitHubConnect, useGitHubDisconnect } from '@/hooks/useGitHubConnection';
import { useGitHubStore } from '@/stores/githubStore';
import { motion } from 'framer-motion';
import { GitFork, Import, LayoutTemplate, RefreshCw, Unlink } from 'lucide-react';
import { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { GitHubReposTab } from './GitHubReposTab';
import { GitHubImportsTab } from './GitHubImportsTab';
import { GitHubTemplatesTab } from './GitHubTemplatesTab';
import './styles.css';

function NoGitHubConnection() {
  const connectMutation = useGitHubConnect();

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="github-no-connection"
    >
      <div className="github-icon-container">
        <svg role="img" viewBox="0 0 24 24" className="github-icon" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
      </div>
      <h2 className="github-no-connection-title">Connect GitHub</h2>
      <p className="github-no-connection-description">
        Link your GitHub account to import functions from your repositories, set up auto-sync, and
        manage deployments directly from your codebase.
      </p>
      <Button
        onClick={() => connectMutation.mutate()}
        disabled={connectMutation.isPending}
        size="lg"
        className="github-connect-btn"
      >
        <svg role="img" viewBox="0 0 24 24" className="github-connect-icon" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
        {connectMutation.isPending ? 'Connecting...' : 'Connect GitHub Account'}
      </Button>
    </motion.div>
  );
}

function GitHubDisconnectButton() {
  const disconnect = useGitHubDisconnect();

  return (
    <Button
      variant="outline"
      size="sm"
      onClick={() => disconnect.mutate()}
      disabled={disconnect.isPending}
      className="github-disconnect-btn"
    >
      <Unlink className="h-4 w-4" />
      {disconnect.isPending ? 'Disconnecting...' : 'Disconnect'}
    </Button>
  );
}

export default function GitHubPage() {
  const { data: connection, isLoading } = useGitHubConnection();
  const disconnect = useGitHubDisconnect();
  const setConnection = useGitHubStore((s) => s.setConnection);
  const [searchParams] = useSearchParams();

  const defaultTab = searchParams.get('tab') || 'repos';

  useEffect(() => {
    setConnection(connection ?? null);
  }, [connection, setConnection]);

  if (isLoading) {
    return (
      <PageLayout>
        <div className="github-loading-container space-y-6">
          <Skeleton className="github-skeleton github-skeleton-header" />
          <Skeleton className="github-skeleton github-skeleton-subtitle" />
          <Skeleton className="github-skeleton github-skeleton-content" />
        </div>
      </PageLayout>
    );
  }

  const isConnected = connection && connection.status === 'active';

  if (!isConnected) {
    return (
      <PageLayout>
        <NoGitHubConnection />
      </PageLayout>
    );
  }

  return (
    <PageLayout>
      <PageHeader
        title="GitHub Integration"
        subtitle="Import and sync functions from your GitHub repositories"
        badges={[{ label: 'Connected', variant: 'new', className: 'github-badge github-badge-connected' }]}
        actions={[{ label: 'Disconnect', onClick: () => disconnect.mutate(), variant: 'outline', size: 'sm', className: 'github-disconnect-btn' }]}
      />

      <Tabs defaultValue={defaultTab} className="github-tabs space-y-6">
        <TabsList className="github-tabs-list">
          <TabsTrigger value="repos" className="github-tabs-trigger">
            <GitFork className="github-tabs-trigger-icon" />
            Repositories
          </TabsTrigger>
          <TabsTrigger value="imports" className="github-tabs-trigger">
            <Import className="github-tabs-trigger-icon" />
            Imports
          </TabsTrigger>
          <TabsTrigger value="templates" className="github-tabs-trigger">
            <LayoutTemplate className="github-tabs-trigger-icon" />
            Templates
          </TabsTrigger>
        </TabsList>

        <TabsContent value="repos" className="github-tabs-content">
          <GitHubReposTab />
        </TabsContent>

        <TabsContent value="imports" className="github-tabs-content">
          <GitHubImportsTab />
        </TabsContent>

        <TabsContent value="templates" className="github-tabs-content">
          <GitHubTemplatesTab />
        </TabsContent>
      </Tabs>
    </PageLayout>
  );
}
