import type { Meta, StoryObj } from '@storybook/react';
import { ProviderCard } from '../components/ProviderCard';
import { ConnectDialog } from '../components/ConnectDialog';
import { ProviderCardSkeleton } from '../components/ProviderCardSkeleton';
import { PROVIDERS } from '@/lib/constants';
import { useState } from 'react';
import { generateMockHealthData } from '../components/ConnectionHealthSparkline';

const providerConfigs = Object.values(PROVIDERS);

const providerAccents: Record<string, { border: string; glow: string; text: string }> = {
  workers: { border: '#f48120', glow: 'rgba(244, 129, 32, 0.15)', text: '#f48120' },
  vercel: { border: '#171717', glow: 'rgba(23, 23, 23, 0.15)', text: '#171717' },
  fly: { border: '#7b68ee', glow: 'rgba(123, 104, 238, 0.15)', text: '#7b68ee' },
  deno: { border: '#0a0a0a', glow: 'rgba(10, 10, 10, 0.15)', text: '#3c3c3c' },
  'functionfly-edge': { border: '#6366f1', glow: 'rgba(99, 102, 241, 0.2)', text: '#6366f1' },
};

const meta = {
  title: 'Providers/ProviderCard',
  component: ProviderCard,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    connected: {
      control: 'boolean',
      description: 'Whether the provider is connected',
    },
    status: {
      control: 'select',
      options: ['online', 'offline', 'degraded', 'pending'],
      description: 'Connection status',
    },
    isDefault: {
      control: 'boolean',
      description: 'Whether this is the default provider',
    },
  },
} satisfies Meta<typeof ProviderCard>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

const InteractiveCard = ({
  providerId,
  initialConnected = false,
  initialStatus = 'pending',
}: {
  providerId: string;
  initialConnected?: boolean;
  initialStatus?: string;
}) => {
  const [connected, setConnected] = useState(initialConnected);
  const [status, setStatus] = useState(initialStatus);
  const [isDefault, setIsDefault] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null);

  const provider = providerConfigs.find((p) => p.id === providerId);
  if (!provider) return null;

  return (
    <div className="w-full max-w-md">
      <ProviderCard
        provider={provider}
        connected={connected}
        status={status}
        isDefault={isDefault}
        onSetDefault={connected ? () => setIsDefault(!isDefault) : undefined}
        onDisconnect={() => setConnected(false)}
        onConnect={async () => {
          await new Promise((resolve) => setTimeout(resolve, 1000));
          setConnected(true);
          setStatus('online');
        }}
        onTestConnection={
          connected
            ? async () => {
                setIsTesting(true);
                await new Promise((resolve) => setTimeout(resolve, 1000));
                setTestResult('success');
                setIsTesting(false);
                setTimeout(() => setTestResult(null), 3000);
              }
            : undefined
        }
        onRotateKey={connected ? () => console.log('Rotate key') : undefined}
        isDisconnecting={false}
        isTestingConnection={isTesting}
        isSettingDefault={false}
        lastUsedAt={connected ? new Date().toISOString() : undefined}
        isStale={status === 'degraded'}
        connectionTestResult={testResult}
        healthData={connected ? generateMockHealthData(24, status === 'online') : undefined}
        last24hUptime={connected ? 99.9 : undefined}
        functionCount={connected ? Math.floor(Math.random() * 20) + 1 : 0}
        accent={providerAccents[providerId]}
        connectDialog={
          <ConnectDialog
            provider={provider}
            accent={providerAccents[providerId]}
            onConnect={async () => {
              await new Promise((resolve) => setTimeout(resolve, 1000));
              setConnected(true);
              setStatus('online');
            }}
          />
        }
      />
    </div>
  );
};

export const NotConnected: RenderStory = {
  render: () => <InteractiveCard providerId="workers" initialConnected={false} />,
};

export const ConnectedOnline: RenderStory = {
  render: () => <InteractiveCard providerId="workers" initialConnected={true} initialStatus="online" />,
};

export const ConnectedDegraded: RenderStory = {
  render: () => <InteractiveCard providerId="workers" initialConnected={true} initialStatus="degraded" />,
};

export const ConnectedOffline: RenderStory = {
  render: () => <InteractiveCard providerId="workers" initialConnected={true} initialStatus="offline" />,
};

export const DefaultProvider: RenderStory = {
  render: () => {
    const [isDefault, setIsDefault] = useState(true);
    const provider = providerConfigs[0];

    return (
      <div className="w-full max-w-md">
        <ProviderCard
          provider={provider}
          connected={true}
          status="online"
          isDefault={isDefault}
          onSetDefault={() => setIsDefault(!isDefault)}
          onDisconnect={() => {}}
          onConnect={async () => {}}
          isDisconnecting={false}
          lastUsedAt={new Date().toISOString()}
          healthData={generateMockHealthData(24, true)}
          last24hUptime={99.9}
          functionCount={15}
          accent={providerAccents[provider.id]}
          connectDialog={
            <ConnectDialog provider={provider} accent={providerAccents[provider.id]} onConnect={async () => {}} />
          }
        />
      </div>
    );
  },
};

export const AllStates: RenderStory = {
  render: () => (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <InteractiveCard providerId="workers" initialConnected={false} />
      <InteractiveCard providerId="workers" initialConnected={true} initialStatus="online" />
      <InteractiveCard providerId="workers" initialConnected={true} initialStatus="degraded" />
      <InteractiveCard providerId="vercel" initialConnected={true} initialStatus="online" />
      <InteractiveCard providerId="fly" initialConnected={true} initialStatus="online" />
      <InteractiveCard providerId="functionfly-edge" initialConnected={true} initialStatus="online" />
    </div>
  ),
};

export const Skeleton: RenderStory = {
  render: () => (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <ProviderCardSkeleton />
      <ProviderCardSkeleton />
      <ProviderCardSkeleton />
    </div>
  ),
};
