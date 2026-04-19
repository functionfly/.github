import type { Meta, StoryObj } from '@storybook/react';
import { ProviderComparisonTooltip, ProviderComparisonTable } from '../components/ProviderComparisonTooltip';
import { PROVIDERS } from '@/lib/constants';
import { useState } from 'react';

const providerConfigs = Object.values(PROVIDERS);

const meta = {
  title: 'Providers/ProviderComparison',
  component: ProviderComparisonTooltip,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof ProviderComparisonTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper type for stories that use custom render functions (args not required)
type RenderStory = Omit<Story, 'args'> & { args?: Partial<Story['args']> };

export const TooltipDemo: RenderStory = {
  render: () => (
    <div className="flex flex-wrap gap-4 pt-24">
      <div className="relative">
        <ProviderComparisonTooltip provider={providerConfigs[0]}>
          <div className="p-4 border rounded-lg hover:bg-bg-secondary cursor-help transition-colors">
            Hover for Cloudflare Workers details
          </div>
        </ProviderComparisonTooltip>
      </div>
      <div className="relative">
        <ProviderComparisonTooltip provider={providerConfigs[1]}>
          <div className="p-4 border rounded-lg hover:bg-bg-secondary cursor-help transition-colors">
            Hover for Vercel details
          </div>
        </ProviderComparisonTooltip>
      </div>
      <div className="relative">
        <ProviderComparisonTooltip provider={providerConfigs[2]}>
          <div className="p-4 border rounded-lg hover:bg-bg-secondary cursor-help transition-colors">
            Hover for Fly.io details
          </div>
        </ProviderComparisonTooltip>
      </div>
    </div>
  ),
};

export const ComparisonTable: RenderStory = {
  render: () => (
    <div className="border rounded-lg p-4">
      <h3 className="text-lg font-semibold mb-4">Provider Comparison</h3>
      <ProviderComparisonTable providers={providerConfigs} />
    </div>
  ),
};

export const EmptyTable: RenderStory = {
  render: () => (
    <div className="border rounded-lg p-4">
      <h3 className="text-lg font-semibold mb-4">Provider Comparison</h3>
      <ProviderComparisonTable providers={[]} />
    </div>
  ),
};

export const SingleProvider: RenderStory = {
  render: () => (
    <div className="border rounded-lg p-4">
      <h3 className="text-lg font-semibold mb-4">Provider Comparison</h3>
      <ProviderComparisonTable providers={[providerConfigs[0]]} />
    </div>
  ),
};
