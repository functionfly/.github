import type { Meta, StoryObj } from '@storybook/react';
import { AutoFailoverDialog } from '../components/AutoFailoverDialog';
import { PROVIDERS } from '@/lib/constants';
import { useState } from 'react';
import type { FailoverConfig } from '../components/AutoFailoverDialog';

const meta = {
  title: 'Providers/AutoFailoverDialog',
  component: AutoFailoverDialog,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof AutoFailoverDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

const InteractiveFailover = ({
  initialConnectedIds,
}: {
  initialConnectedIds: string[];
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [config, setConfig] = useState<FailoverConfig>({
    enabled: false,
    primaryProviderId: null,
    fallbackProviderId: null,
    autoSwitchThreshold: 10,
    switchbackDelay: 15,
  });

  const providers = Object.values(PROVIDERS);

  return (
    <div className="p-4">
      <button
        onClick={() => setIsOpen(true)}
        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
      >
        Open Failover Configuration
      </button>
      <div className="mt-4 p-4 bg-bg-secondary rounded-lg">
        <h4 className="font-medium text-text-primary mb-2">Current Config:</h4>
        <pre className="text-sm text-text-secondary overflow-auto">
          {JSON.stringify(config, null, 2)}
        </pre>
      </div>
      <AutoFailoverDialog
        providers={providers}
        connectedProviderIds={initialConnectedIds}
        currentConfig={config}
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onSave={async (newConfig) => {
          setIsSaving(true);
          await new Promise((resolve) => setTimeout(resolve, 800));
          setConfig(newConfig);
          setIsSaving(false);
          setIsOpen(false);
        }}
        isSaving={isSaving}
      />
    </div>
  );
};

export const NoProviders: RenderStory = {
  render: () => <InteractiveFailover initialConnectedIds={[]} />,
};

export const OneProvider: RenderStory = {
  render: () => <InteractiveFailover initialConnectedIds={['workers']} />,
};

export const TwoProviders: RenderStory = {
  render: () => <InteractiveFailover initialConnectedIds={['workers', 'vercel']} />,
};

export const AllProviders: RenderStory = {
  render: () =>
    <InteractiveFailover initialConnectedIds={Object.values(PROVIDERS).map((p) => p.id)} />,
};

export const PreConfigured: RenderStory = {
  render: () => {
    const [isOpen, setIsOpen] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [config, setConfig] = useState<FailoverConfig>({
      enabled: true,
      primaryProviderId: 'workers',
      fallbackProviderId: 'vercel',
      autoSwitchThreshold: 5,
      switchbackDelay: 30,
    });

    const providers = Object.values(PROVIDERS);

    return (
      <div className="p-4">
        <button
          onClick={() => setIsOpen(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Edit Existing Failover Config
        </button>
        <AutoFailoverDialog
          providers={providers}
          connectedProviderIds={['workers', 'vercel', 'fly']}
          currentConfig={config}
          isOpen={isOpen}
          onClose={() => setIsOpen(false)}
          onSave={async (newConfig) => {
            setIsSaving(true);
            await new Promise((resolve) => setTimeout(resolve, 800));
            setConfig(newConfig);
            setIsSaving(false);
            setIsOpen(false);
          }}
          isSaving={isSaving}
        />
      </div>
    );
  },
};
