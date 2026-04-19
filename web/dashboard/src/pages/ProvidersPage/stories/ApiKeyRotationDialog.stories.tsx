import type { Meta, StoryObj } from '@storybook/react';
import { ApiKeyRotationDialog } from '../components/ApiKeyRotationDialog';
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
  title: 'Providers/ApiKeyRotationDialog',
  component: ApiKeyRotationDialog,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ApiKeyRotationDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

const InteractiveRotation = ({ providerId }: { providerId: string }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [isRotating, setIsRotating] = useState(false);
  const provider = Object.values(PROVIDERS).find((p) => p.id === providerId);

  if (!provider) return null;

  return (
    <div className="p-4">
      <button
        onClick={() => setIsOpen(true)}
        className="px-4 py-2 bg-amber-600 text-white rounded-md hover:bg-amber-700"
      >
        Open Rotation Dialog for {provider?.name}
      </button>
      <ApiKeyRotationDialog
        provider={provider}
        accent={providerAccents[providerId]}
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onRotate={async () => {
          setIsRotating(true);
          await new Promise((resolve) => setTimeout(resolve, 1500));
          setIsRotating(false);
          setIsOpen(false);
        }}
        isRotating={isRotating}
      />
    </div>
  );
};

export const CloudflareWorkers: RenderStory = {
  render: () => <InteractiveRotation providerId="workers" />,
};

export const Vercel: RenderStory = {
  render: () => <InteractiveRotation providerId="vercel" />,
};

export const FlyIo: RenderStory = {
  render: () => <InteractiveRotation providerId="fly" />,
};

export const WithLongProviderName: RenderStory = {
  render: () => <InteractiveRotation providerId="functionfly-edge" />,
};
