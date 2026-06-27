// ReceiptSchemaViewer component unit tests.
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { Receipt } from '../types';
import { ReceiptSchemaViewer } from './ReceiptSchemaViewer';

const sampleReceipt: Receipt = {
  id: 'V1StGXR8_Z5jHi3B-myT',
  function: {
    name: 'test-fn',
    author: 'ada',
    runtime: 'python3.11',
    version: '1.0.0',
    visibility: 'public',
    description: 'A test function.',
    input_schema: {
      type: 'object',
      properties: {
        url: { type: 'string', description: 'URL to fetch' },
        max: { type: 'number', default: 3 },
      },
      required: ['url'],
    },
    output_schema: {
      type: 'object',
      properties: {
        summary: { type: 'string' },
      },
    },
  },
  execution: {
    input: { url: 'https://example.com' },
    output: { summary: 'Example domain.' },
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

describe('ReceiptSchemaViewer', () => {
  it('renders schema tabs when schemas are present', () => {
    render(<ReceiptSchemaViewer receipt={sampleReceipt} />);
    expect(screen.getByRole('tab', { name: /input/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /output/i })).toBeInTheDocument();
  });

  it('shows message when no schema declared', () => {
    const noSchemaReceipt = {
      ...sampleReceipt,
      function: { ...sampleReceipt.function, input_schema: undefined, output_schema: undefined },
    };
    render(<ReceiptSchemaViewer receipt={noSchemaReceipt} />);
    expect(
      screen.getByText(/The function author hasn't published a JSON Schema/i)
    ).toBeInTheDocument();
  });

  it('renders schema tree with required badge', () => {
    render(<ReceiptSchemaViewer receipt={sampleReceipt} />);
    // The url field is required, so required badge should appear.
    expect(screen.getByText('url')).toBeInTheDocument();
    expect(screen.getByText('required')).toBeInTheDocument();
  });

  it('shows type badges for schema fields', () => {
    render(<ReceiptSchemaViewer receipt={sampleReceipt} />);
    expect(screen.getAllByText('string').length).toBeGreaterThan(0);
  });
});
