import type { Meta, StoryObj } from '@storybook/react';
import { ProviderCardSkeleton } from '../components/ProviderCardSkeleton';

const meta = {
  title: 'Providers/ProviderCardSkeleton',
  component: ProviderCardSkeleton,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ProviderCardSkeleton>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

export const Default: Story = {
  args: {},
};

export const MultipleCards: RenderStory = {
  render: () => (
    <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-5">
      <ProviderCardSkeleton />
      <ProviderCardSkeleton />
      <ProviderCardSkeleton />
    </div>
  ),
};
