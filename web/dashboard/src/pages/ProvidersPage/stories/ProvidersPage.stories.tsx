import type { Meta, StoryObj } from '@storybook/react';
import { ProvidersPage } from '../index';

const meta = {
  title: 'Pages/ProvidersPage',
  component: ProvidersPage,
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: 'Main providers management page with all features integrated.',
      },
    },
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ProvidersPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  parameters: {
    docs: {
      description: {
        story: 'Default state of the ProvidersPage component.',
      },
    },
  },
};

export const Mobile: Story = {
  parameters: {
    viewport: {
      defaultViewport: 'mobile1',
    },
  },
};

export const Tablet: Story = {
  parameters: {
    viewport: {
      defaultViewport: 'tablet',
    },
  },
};
