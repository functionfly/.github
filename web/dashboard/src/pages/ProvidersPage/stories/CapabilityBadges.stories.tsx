import type { Meta, StoryObj } from '@storybook/react';
import { CapabilityBadges } from '../components/CapabilityBadges';

const meta = {
  title: 'Providers/CapabilityBadges',
  component: CapabilityBadges,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    providerId: {
      control: 'select',
      options: ['workers', 'vercel', 'fly', 'deno', 'functionfly-edge'],
      description: 'Provider ID to show capabilities for',
    },
    showAll: {
      control: 'boolean',
      description: 'Show all capabilities or truncate',
    },
    maxDisplay: {
      control: 'number',
      description: 'Maximum number of badges to display',
    },
  },
} satisfies Meta<typeof CapabilityBadges>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

export const CloudflareWorkers: Story = {
  args: {
    providerId: 'workers',
  },
};

export const Vercel: Story = {
  args: {
    providerId: 'vercel',
  },
};

export const FlyIo: Story = {
  args: {
    providerId: 'fly',
  },
};

export const FunctionFlyEdge: Story = {
  args: {
    providerId: 'functionfly-edge',
  },
};

export const ShowAll: Story = {
  args: {
    providerId: 'workers',
    showAll: true,
  },
};

export const LimitedDisplay: Story = {
  args: {
    providerId: 'workers',
    maxDisplay: 2,
  },
};

export const AllProviders: RenderStory = {
  render: () => (
    <div className="space-y-4">
      <div>
        <span className="text-sm text-text-secondary block mb-2">Cloudflare Workers</span>
        <CapabilityBadges providerId="workers" showAll />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">Vercel</span>
        <CapabilityBadges providerId="vercel" showAll />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">Fly.io</span>
        <CapabilityBadges providerId="fly" showAll />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">Deno Deploy</span>
        <CapabilityBadges providerId="deno" showAll />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">FunctionFly Edge</span>
        <CapabilityBadges providerId="functionfly-edge" showAll />
      </div>
    </div>
  ),
};
