import type { Meta, StoryObj } from '@storybook/react';
import { ConnectionStatus } from '../components/ConnectionStatus';

const meta = {
  title: 'Providers/ConnectionStatus',
  component: ConnectionStatus,
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
      description: 'Current connection status',
    },
  },
} satisfies Meta<typeof ConnectionStatus>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

export const ConnectedOnline: Story = {
  args: {
    connected: true,
    status: 'online',
  },
};

export const ConnectedDegraded: Story = {
  args: {
    connected: true,
    status: 'degraded',
  },
};

export const ConnectedOffline: Story = {
  args: {
    connected: true,
    status: 'offline',
  },
};

export const NotConnected: Story = {
  args: {
    connected: false,
  },
};

export const AllVariants: RenderStory = {
  render: () => (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <span className="text-sm text-text-secondary w-24">Connected</span>
        <ConnectionStatus connected={true} status="online" />
      </div>
      <div className="flex items-center gap-4">
        <span className="text-sm text-text-secondary w-24">Degraded</span>
        <ConnectionStatus connected={true} status="degraded" />
      </div>
      <div className="flex items-center gap-4">
        <span className="text-sm text-text-secondary w-24">Offline</span>
        <ConnectionStatus connected={true} status="offline" />
      </div>
      <div className="flex items-center gap-4">
        <span className="text-sm text-text-secondary w-24">Not Connected</span>
        <ConnectionStatus connected={false} />
      </div>
    </div>
  ),
};
