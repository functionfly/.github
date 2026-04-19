import type { Meta, StoryObj } from '@storybook/react';
import { RegionBadges, RegionList } from '../components/RegionBadges';

const meta = {
  title: 'Providers/RegionBadges',
  component: RegionBadges,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    regions: {
      control: 'object',
      description: 'Array of region codes',
    },
    providerId: {
      control: 'select',
      options: ['workers', 'vercel', 'fly', 'deno', 'functionfly-edge'],
      description: 'Provider ID for region mapping',
    },
  },
} satisfies Meta<typeof RegionBadges>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

export const CloudflareWorkers: Story = {
  args: {
    regions: ['iad', 'lhr', 'sin', 'syd', 'fra', 'hkg', 'tyo'],
    providerId: 'workers',
  },
};

export const Vercel: Story = {
  args: {
    regions: ['iad1', 'sfo1', 'lhr1', 'fra1'],
    providerId: 'vercel',
  },
};

export const FlyIo: Story = {
  args: {
    regions: ['iad', 'lax', 'ord', 'lhr', 'fra', 'sin', 'syd', 'nrt'],
    providerId: 'fly',
  },
};

export const FunctionFlyEdge: Story = {
  args: {
    regions: ['us-east-1', 'us-west-2', 'eu-central-1'],
    providerId: 'functionfly-edge',
  },
};

export const AllProviders: RenderStory = {
  render: () => (
    <div className="space-y-4">
      <div>
        <span className="text-sm text-text-secondary block mb-2">Cloudflare Workers</span>
        <RegionBadges regions={['iad', 'lhr', 'sin', 'syd', 'fra', 'hkg', 'tyo']} providerId="workers" />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">Vercel</span>
        <RegionBadges regions={['iad1', 'sfo1', 'lhr1', 'fra1']} providerId="vercel" />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">Fly.io</span>
        <RegionBadges regions={['iad', 'lax', 'ord', 'lhr', 'fra', 'sin', 'syd', 'nrt']} providerId="fly" />
      </div>
    </div>
  ),
};

// RegionList stories
export const ListView: RenderStory = {
  render: () => (
    <div className="space-y-2">
      <RegionList regions={['iad', 'lhr', 'sin', 'syd', 'fra', 'hkg', 'tyo']} providerId="workers" />
      <RegionList regions={['iad1', 'sfo1', 'lhr1', 'fra1']} providerId="vercel" maxDisplay={2} />
    </div>
  ),
};
