import type { Meta, StoryObj } from '@storybook/react';
import { DisconnectConfirmationDialog } from '../components/DisconnectConfirmationDialog';
import { useState } from 'react';

const meta = {
  title: 'Providers/DisconnectConfirmationDialog',
  component: DisconnectConfirmationDialog,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    providerName: {
      control: 'text',
      description: 'Name of the provider to disconnect',
    },
    isDisconnecting: {
      control: 'boolean',
      description: 'Whether disconnection is in progress',
    },
  },
} satisfies Meta<typeof DisconnectConfirmationDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

const InteractiveDisconnect = ({ providerName }: { providerName: string }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isDisconnecting, setIsDisconnecting] = useState(false);

  return (
    <div className="p-4">
      <button
        onClick={() => setIsOpen(true)}
        className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700"
      >
        Disconnect {providerName}
      </button>
      <DisconnectConfirmationDialog
        providerName={providerName}
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onConfirm={async () => {
          setIsDisconnecting(true);
          await new Promise((resolve) => setTimeout(resolve, 1500));
          setIsDisconnecting(false);
          setIsOpen(false);
        }}
        isDisconnecting={isDisconnecting}
      />
    </div>
  );
};

export const CloudflareWorkers: RenderStory = {
  render: () => <InteractiveDisconnect providerName="Cloudflare Workers" />,
};

export const Vercel: RenderStory = {
  render: () => <InteractiveDisconnect providerName="Vercel" />,
};

export const FunctionFlyEdge: RenderStory = {
  render: () => <InteractiveDisconnect providerName="FunctionFly Edge" />,
};

export const DisconnectingState: Story = {
  args: {
    providerName: 'Cloudflare Workers',
    isOpen: true,
    onClose: () => {},
    onConfirm: () => {},
    isDisconnecting: true,
  },
};
