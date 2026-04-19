import type { Meta, StoryObj } from '@storybook/react';
import { ConnectDialog } from '../components/ConnectDialog';
import { PROVIDERS } from '@/lib/constants';
import { useState } from 'react';

const providerAccents: Record<string, { border: string; glow: string; text: string }> = {
  workers: { border: '#f48120', glow: 'rgba(244, 129, 32, 0.15)', text: '#f48120' },
  vercel: { border: '#171717', glow: 'rgba(23, 23, 23, 0.15)', text: '#171717' },
  fly: { border: '#7b68ee', glow: 'rgba(123, 104, 238, 0.15)', text: '#7b68ee' },
  deno: { border: '#0a0a0a', glow: 'rgba(10, 10, 10, 0.15)', text: '#3c3c3c' },
  'functionfly-edge': { border: '#6366f1', glow: 'rgba(99, 102, 241, 0.2)', text: '#6366f1' },
};

const meta = {
  title: 'Providers/ConnectDialog',
  component: ConnectDialog,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ConnectDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

const InteractiveConnect = ({ providerId }: { providerId: string }) => {
  const [isOpen, setIsOpen] = useState(false);
  const provider = Object.values(PROVIDERS).find((p) => p.id === providerId);

  if (!provider) return null;

  return (
    <div className="p-4">
      <button
        onClick={() => setIsOpen(true)}
        className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
      >
        Open Connect Dialog for {provider.name}
      </button>
      {isOpen && (
        <ConnectDialog
          provider={provider}
          accent={providerAccents[providerId]}
          onConnect={async () => {
            await new Promise((resolve) => setTimeout(resolve, 1000));
            setIsOpen(false);
          }}
        />
      )}
    </div>
  );
};

export const CloudflareWorkers: RenderStory = {
  render: () => <InteractiveConnect providerId="workers" />,
};

export const Vercel: RenderStory = {
  render: () => <InteractiveConnect providerId="vercel" />,
};

export const FunctionFlyEdge: RenderStory = {
  render: () => <InteractiveConnect providerId="functionfly-edge" />,
};

export const AllProviders: RenderStory = {
  render: () => (
    <div className="space-y-4">
      {Object.values(PROVIDERS).map((provider) => (
        <div key={provider.id} className="flex items-center gap-4">
          <span className="w-32 text-sm text-text-secondary">{provider.name}</span>
          <ConnectDialog
            provider={provider}
            accent={providerAccents[provider.id]}
            onConnect={async () => {
              await new Promise((resolve) => setTimeout(resolve, 500));
            }}
          />
        </div>
      ))}
    </div>
  ),
};
