import type { Meta, StoryObj } from '@storybook/react';
import { ProviderUsage, generateMockUsageStats } from '../components/ProviderUsage';

const meta = {
  title: 'Providers/ProviderUsage',
  component: ProviderUsage,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    compact: {
      control: 'boolean',
      description: 'Show compact or full layout',
    },
  },
} satisfies Meta<typeof ProviderUsage>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

export const Compact: Story = {
  args: {
    stats: generateMockUsageStats(),
    compact: true,
  },
};

export const Full: Story = {
  args: {
    stats: generateMockUsageStats(),
    compact: false,
  },
};

export const HighUsage: Story = {
  args: {
    stats: {
      functionCount: 25,
      requestCount: 8500000,
      requestChangePercent: 45.2,
      avgLatency: 45,
      errorRate: 0.1,
    },
    compact: false,
  },
};

export const LowUsage: Story = {
  args: {
    stats: {
      functionCount: 1,
      requestCount: 150,
      requestChangePercent: -12.5,
      avgLatency: 120,
      errorRate: 2.5,
    },
    compact: false,
  },
};

export const WithErrorWarning: Story = {
  args: {
    stats: {
      functionCount: 5,
      requestCount: 50000,
      requestChangePercent: 15,
      avgLatency: 200,
      errorRate: 3.2,
    },
    compact: false,
  },
};

export const Comparison: RenderStory = {
  render: () => (
    <div className="space-y-4">
      <div>
        <span className="text-sm text-text-secondary block mb-2">High Traffic Provider</span>
        <ProviderUsage
          stats={{
            functionCount: 42,
            requestCount: 12500000,
            requestChangePercent: 38.5,
            avgLatency: 32,
            errorRate: 0.05,
          }}
        />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">New Provider</span>
        <ProviderUsage
          stats={{
            functionCount: 2,
            requestCount: 450,
            requestChangePercent: 150,
            avgLatency: 85,
            errorRate: 1.2,
          }}
        />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">Issue Detected</span>
        <ProviderUsage
          stats={{
            functionCount: 8,
            requestCount: 89000,
            requestChangePercent: -25,
            avgLatency: 450,
            errorRate: 5.8,
          }}
        />
      </div>
    </div>
  ),
};
