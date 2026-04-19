import type { Meta, StoryObj } from '@storybook/react';
import { ConnectionHealthSparkline, generateMockHealthData } from '../components/ConnectionHealthSparkline';

const meta = {
  title: 'Providers/ConnectionHealthSparkline',
  component: ConnectionHealthSparkline,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    data: {
      control: 'object',
      description: 'Array of health data points',
    },
    height: {
      control: 'number',
      description: 'Height of the sparkline in pixels',
    },
    width: {
      control: 'number',
      description: 'Width of the sparkline in pixels',
    },
  },
} satisfies Meta<typeof ConnectionHealthSparkline>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

export const Healthy: Story = {
  args: {
    data: generateMockHealthData(24, true),
    height: 24,
    width: 100,
  },
};

export const Degraded: Story = {
  args: {
    data: [
      ...generateMockHealthData(12, true),
      ...generateMockHealthData(12, true).map(d => ({ ...d, status: 'degraded' as const, latency: d.latency ? d.latency + 200 : 200 })),
    ],
    height: 24,
    width: 100,
  },
};

export const Offline: Story = {
  args: {
    data: generateMockHealthData(24, false),
    height: 24,
    width: 100,
  },
};

export const Empty: Story = {
  args: {
    data: [],
    height: 24,
    width: 100,
  },
};

export const CustomSize: Story = {
  args: {
    data: generateMockHealthData(48, true),
    height: 40,
    width: 200,
  },
};

export const MultipleTimeframes: RenderStory = {
  render: () => (
    <div className="space-y-4">
      <div>
        <span className="text-sm text-text-secondary block mb-2">24h Healthy</span>
        <ConnectionHealthSparkline data={generateMockHealthData(24, true)} height={24} width={150} />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">48h with issues</span>
        <ConnectionHealthSparkline
          data={[
            ...generateMockHealthData(24, true),
            ...generateMockHealthData(24, false).map(d => ({ ...d, status: 'degraded' as const })),
          ]}
          height={24}
          width={150}
        />
      </div>
      <div>
        <span className="text-sm text-text-secondary block mb-2">12h degraded</span>
        <ConnectionHealthSparkline data={generateMockHealthData(12, false)} height={24} width={150} />
      </div>
    </div>
  ),
};
