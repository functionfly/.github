// ReceiptForkCTA component unit tests — mocks the useReceiptFork hook.
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import type { Receipt } from '../types';

// Mock the hook before importing the component.
vi.mock('../hooks/useReceiptFork', () => ({
  useReceiptFork: vi.fn(() => ({
    buildForkLink: () => ({
      target: 'editor',
      href: '/editor?fork=abc',
      sourceBytes: 1024,
    }),
    isLoading: false,
  })),
}));

import { ReceiptForkCTA } from './ReceiptForkCTA';

const sampleReceipt: Receipt = {
  id: 'V1StGXR8_Z5jHi3B-myT',
  function: {
    name: 'test-fn',
    author: 'ada',
    runtime: 'python3.11',
    version: '1.0.0',
    visibility: 'public',
    description: 'A test function.',
  },
  execution: {
    input: {},
    output: {},
    duration_ms: 42,
    cached: false,
    created_at: '2026-06-01T18:42:11Z',
  },
  share: {
    url: 'https://functionfly.com/r/V1StGXR8_Z5jHi3B-myT',
    embed_url: 'https://functionfly.com/r/V1StGXR8_Z5jHi3B-myT/embed',
    tweet_intent_url: 'https://twitter.com/intent/tweet?text=hi',
    og_meta: { title: 't', description: 'd', image: 'i' },
  },
  can_run: true,
  is_paid: false,
  price_per_call_usd: 0,
};

describe('ReceiptForkCTA', () => {
  it('renders the fork CTA card', () => {
    render(
      <MemoryRouter>
        <ReceiptForkCTA receipt={sampleReceipt} isAuthenticated={false} />
      </MemoryRouter>
    );
    expect(screen.getByTestId('receipt-fork-cta')).toBeInTheDocument();
    expect(screen.getByText(/deploy your own function/i)).toBeInTheDocument();
  });

  it('shows source bytes when link is available', () => {
    render(
      <MemoryRouter>
        <ReceiptForkCTA receipt={sampleReceipt} isAuthenticated={false} />
      </MemoryRouter>
    );
    expect(screen.getByText(/1,024 bytes/i)).toBeInTheDocument();
  });

  it('renders the 3 step explainer', () => {
    render(
      <MemoryRouter>
        <ReceiptForkCTA receipt={sampleReceipt} isAuthenticated={false} />
      </MemoryRouter>
    );
    expect(screen.getByText('Fork')).toBeInTheDocument();
    expect(screen.getByText('Edit')).toBeInTheDocument();
    expect(screen.getByText('Deploy')).toBeInTheDocument();
  });
});
