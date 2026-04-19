import type { Meta, StoryObj } from '@storybook/react';
import { ConnectionAuditLog, generateMockAuditLog } from '../components/ConnectionAuditLog';

const meta = {
  title: 'Providers/ConnectionAuditLog',
  component: ConnectionAuditLog,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    entries: {
      control: 'object',
      description: 'Array of audit log entries',
    },
    maxHeight: {
      control: 'number',
      description: 'Maximum height for scroll area',
    },
  },
} satisfies Meta<typeof ConnectionAuditLog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithData: Story = {
  args: {
    entries: generateMockAuditLog(),
    maxHeight: 300,
  },
};

export const Empty: Story = {
  args: {
    entries: [],
    maxHeight: 200,
  },
};

export const SingleProvider: Story = {
  args: {
    entries: [
      {
        id: '1',
        timestamp: new Date().toISOString(),
        action: 'connected',
        providerName: 'Cloudflare Workers',
        providerId: 'workers',
        actor: 'admin@example.com',
      },
      {
        id: '2',
        timestamp: new Date(Date.now() - 3600000).toISOString(),
        action: 'tested',
        providerName: 'Cloudflare Workers',
        providerId: 'workers',
        details: 'Connection test successful',
      },
    ],
    maxHeight: 200,
  },
};

export const WithErrors: Story = {
  args: {
    entries: [
      {
        id: '1',
        timestamp: new Date().toISOString(),
        action: 'failed',
        providerName: 'Vercel',
        providerId: 'vercel',
        details: 'Authentication failed - invalid API key',
      },
      {
        id: '2',
        timestamp: new Date(Date.now() - 1800000).toISOString(),
        action: 'rotated',
        providerName: 'Vercel',
        providerId: 'vercel',
        details: 'API key rotated successfully',
      },
      {
        id: '3',
        timestamp: new Date(Date.now() - 3600000).toISOString(),
        action: 'tested',
        providerName: 'Fly.io',
        providerId: 'fly',
        details: 'Connection test failed - timeout',
      },
    ],
    maxHeight: 250,
  },
};

export const FullHistory: Story = {
  args: {
    entries: [
      ...generateMockAuditLog(),
      ...generateMockAuditLog(),
      ...generateMockAuditLog(),
    ].slice(0, 20),
    maxHeight: 400,
  },
};
