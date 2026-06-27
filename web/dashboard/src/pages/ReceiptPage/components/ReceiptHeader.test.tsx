// ReceiptHeader component unit tests.
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import type { Receipt } from '../types';
import { ReceiptHeader } from './ReceiptHeader';

const sampleReceipt: Receipt = {
  id: 'V1StGXR8_Z5jHi3B-myT',
  function: {
    name: 'my-function',
    author: 'testauthor',
    runtime: 'python3.11',
    version: '1.4.2',
    visibility: 'public',
    description: 'A test function that does something useful.',
  },
  execution: {
    input: {},
    output: {},
    duration_ms: 142,
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

describe('ReceiptHeader', () => {
  it('renders function name and author', () => {
    render(
      <MemoryRouter>
        <ReceiptHeader receipt={sampleReceipt} />
      </MemoryRouter>
    );
    expect(screen.getByText(/testauthor\/my-function/)).toBeInTheDocument();
  });

  it('renders runtime badge', () => {
    render(
      <MemoryRouter>
        <ReceiptHeader receipt={sampleReceipt} />
      </MemoryRouter>
    );
    // Runtime label is derived from getRuntimeStyle.
    expect(screen.getByText(/python/i)).toBeInTheDocument();
  });

  it('renders version badge', () => {
    render(
      <MemoryRouter>
        <ReceiptHeader receipt={sampleReceipt} />
      </MemoryRouter>
    );
    expect(screen.getByText(/v1.4.2/)).toBeInTheDocument();
  });

  it('renders author link', () => {
    render(
      <MemoryRouter>
        <ReceiptHeader receipt={sampleReceipt} />
      </MemoryRouter>
    );
    expect(screen.getByRole('link', { name: /by testauthor/i })).toBeInTheDocument();
  });

  it('renders description when present', () => {
    render(
      <MemoryRouter>
        <ReceiptHeader receipt={sampleReceipt} />
      </MemoryRouter>
    );
    expect(screen.getByText(/A test function that does something useful/)).toBeInTheDocument();
  });

  it('does not render description when missing', () => {
    const noDesc = {
      ...sampleReceipt,
      function: { ...sampleReceipt.function, description: undefined },
    };
    render(
      <MemoryRouter>
        <ReceiptHeader receipt={noDesc} />
      </MemoryRouter>
    );
    expect(screen.queryByText(/test function/i)).not.toBeInTheDocument();
  });
});
