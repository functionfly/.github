import { useGitHubConnection, useGitHubConnect, useGitHubDisconnect } from '@/hooks/useGitHubConnection';
import { useGitHubStore } from '@/stores/githubStore';
import { GitFork, Import, LayoutTemplate, Unlink } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, StatusPill, AnnotationTag,
} from '@/components/containment';
import { GitHubReposTab } from './GitHubReposTab';
import { GitHubImportsTab } from './GitHubImportsTab';
import { GitHubTemplatesTab } from './GitHubTemplatesTab';
import './styles.css';

const GitHubIcon = ({ className }: { className?: string }) => (
  <svg role="img" viewBox="0 0 24 24" className={className} xmlns="http://www.w3.org/2000/svg">
    <path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
  </svg>
);

function NoGitHubConnection() {
  const connectMutation = useGitHubConnect();

  return (
    <Chamber className="gh-no-connection">
      <div className="gh-no-connection__icon-wrap">
        <GitHubIcon className="gh-no-connection__icon" />
      </div>
      <h2 className="gh-no-connection__title">Connect GitHub</h2>
      <p className="gh-no-connection__desc">
        Link your GitHub account to import functions from your repositories, set up auto-sync, and
        manage deployments directly from your codebase.
      </p>
      <SealedButton onClick={() => connectMutation.mutate()} disabled={connectMutation.isPending}
        iconLeft={<GitHubIcon className="gh-icon-sm" />}>
        {connectMutation.isPending ? 'Connecting...' : 'Connect GitHub Account'}
      </SealedButton>
    </Chamber>
  );
}

export default function GitHubPage() {
  const { data: connection, isLoading } = useGitHubConnection();
  const disconnect = useGitHubDisconnect();
  const setConnection = useGitHubStore((s) => s.setConnection);
  const [searchParams] = useSearchParams();
  const [activeTab, setActiveTab] = useState(searchParams.get('tab') || 'repos');

  useEffect(() => { setConnection(connection ?? null); }, [connection, setConnection]);

  if (isLoading) {
    return (
      <div className="gh-page">
        <PageGrid />
        <Chamber className="gh-loading">
          <div className="gh-skeleton gh-skeleton--header" />
          <div className="gh-skeleton gh-skeleton--subtitle" />
          <div className="gh-skeleton gh-skeleton--content" />
        </Chamber>
      </div>
    );
  }

  const isConnected = connection && connection.status === 'active';

  if (!isConnected) {
    return (
      <div className="gh-page">
        <PageGrid />
        <NoGitHubConnection />
      </div>
    );
  }

  const tabs = [
    { value: 'repos', label: 'Repositories', icon: GitFork },
    { value: 'imports', label: 'Imports', icon: Import },
    { value: 'templates', label: 'Templates', icon: LayoutTemplate },
  ];

  return (
    <div className="gh-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="gh-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE GH-01" secondary="GitHub Integration" position="top-right" />

        <div className="gh-hero__header">
          <div className="gh-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="gh-hero__title">GitHub Integration</h1>
            <StatusPill status="live" label="Connected" />
          </div>
          <p className="gh-hero__subtitle">Import and sync functions from your GitHub repositories</p>
          <div className="gh-hero__actions">
            <FrameButton size="sm" onClick={() => disconnect.mutate()} disabled={disconnect.isPending}
              iconLeft={<Unlink className="gh-icon-sm" />}>
              {disconnect.isPending ? 'Disconnecting...' : 'Disconnect'}
            </FrameButton>
          </div>
        </div>
      </Chamber>

      {/* Tabs */}
      <div className="gh-tabs">
        {tabs.map((tab) => (
          <button key={tab.value} className={`gh-tab ${activeTab === tab.value ? 'gh-tab--active' : ''}`}
            onClick={() => setActiveTab(tab.value)}>
            <tab.icon className="gh-icon-sm" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="gh-tab-content">
        {activeTab === 'repos' && <GitHubReposTab />}
        {activeTab === 'imports' && <GitHubImportsTab />}
        {activeTab === 'templates' && <GitHubTemplatesTab />}
      </div>
    </div>
  );
}
